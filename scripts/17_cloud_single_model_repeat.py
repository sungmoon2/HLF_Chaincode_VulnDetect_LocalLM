"""
17_cloud_single_model_repeat.py  (v1.0 — 2026-02-10)
- 단일 클라우드 모델 x 3 prompts x 15 files x 5 runs
- CLI: python 17_cloud_single_model_repeat.py --model <model_name> --backend <claude|gemini>
- 6개 인스턴스를 병렬 실행하여 전체 시간 단축
"""

import argparse
import hashlib
import json
import os
import sys
import csv
import time
from pathlib import Path
from datetime import datetime

from colorama import init, Fore
from tqdm import tqdm

init(autoreset=True)

PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

INFERENCE_PARAMS = {"temperature": 0.1, "max_tokens": 2048}
MAX_RETRIES = 3
RETRY_BASE_DELAY = 2
N_RUNS = 5
TOKEN_REFRESH_INTERVAL = 1800

# ── 프롬프트 (02_run_audit_v3.py 동일) ────────────────────────────────
PROMPT_ZERO_SHOT = (
    "You are a Hyperledger Fabric Security Expert. "
    "Analyze the following Go chaincode for security vulnerabilities. "
    "Focus on: access control issues, input validation flaws, "
    "read-after-write conflicts (phantom reads), private data leakage, "
    "non-deterministic operations, and insecure key management. "
    "For each vulnerability found, provide:\n"
    "1. Vulnerability type\n"
    "2. Severity (Critical/High/Medium/Low)\n"
    "3. Affected code location (function name and line reference)\n"
    "4. Description of the issue\n"
    "5. Recommended fix\n"
    "If no vulnerabilities are found, state 'No vulnerabilities detected.'"
)

PROMPT_FEW_SHOT = (
    "You are a Hyperledger Fabric Security Expert. "
    "Analyze the following Go chaincode for security vulnerabilities. "
    "Focus on: access control issues, input validation flaws, "
    "read-after-write conflicts (phantom reads), private data leakage, "
    "non-deterministic operations, and insecure key management. "
    "For each vulnerability found, provide:\n"
    "1. Vulnerability type\n"
    "2. Severity (Critical/High/Medium/Low)\n"
    "3. Affected code location (function name and line reference)\n"
    "4. Description of the issue\n"
    "5. Recommended fix\n"
    "If no vulnerabilities are found, state 'No vulnerabilities detected.'\n\n"
    "--- Example 1 (VULNERABLE) ---\n"
    "Code snippet:\n"
    "```go\n"
    "func (s *Contract) StoreEvent(ctx contractapi.TransactionContextInterface, id string) error {\n"
    "    event := Event{ID: id, Timestamp: time.Now().Format(time.RFC3339)}\n"
    "    eventJSON, _ := json.Marshal(event)\n"
    "    return ctx.GetStub().PutState(id, eventJSON)\n"
    "}\n"
    "```\n"
    "Analysis: **Vulnerable.** `time.Now()` produces a different value on each endorsing peer. "
    "The resulting `eventJSON` differs across peers, causing endorsement mismatch. "
    "Severity: High. Fix: Use `ctx.GetStub().GetTxTimestamp()` instead.\n\n"
    "--- Example 2 (SAFE) ---\n"
    "Code snippet:\n"
    "```go\n"
    "func (s *Contract) LogAndStore(ctx contractapi.TransactionContextInterface, id string, value string) error {\n"
    "    fmt.Printf(\"[%s] Storing %s\\n\", time.Now().Format(time.RFC3339), id)\n"
    "    return ctx.GetStub().PutState(id, []byte(value))\n"
    "}\n"
    "```\n"
    "Analysis: **No vulnerabilities detected.** `time.Now()` is used only in `fmt.Printf` "
    "for local console logging. The value written to the ledger via `PutState` is the "
    "deterministic `value` argument. The write set is identical across all peers.\n\n"
    "--- Now analyze the following chaincode ---"
)

PROMPT_COT = (
    "You are a Hyperledger Fabric Security Expert. "
    "Analyze the following Go chaincode for security vulnerabilities.\n\n"
    "IMPORTANT: Before stating your conclusions, you MUST reason step-by-step:\n"
    "Step 1: Identify ALL state-modifying operations (PutState, DelState) in the code.\n"
    "Step 2: For each PutState/DelState call, trace the data backward to its source. "
    "Determine whether each value written to the ledger is deterministic "
    "(same on all endorsing peers) or nondeterministic (varies per peer).\n"
    "Step 3: Check if any nondeterministic source (time.Now(), math/rand, map iteration order, "
    "goroutine race, external API call, file I/O) flows into a PutState value or key.\n"
    "Step 4: Check for resource leaks (iterators from GetStateByRange not closed with defer).\n"
    "Step 5: Check for read-after-write (phantom read) patterns where GetState is followed "
    "by PutState on overlapping key ranges without proper MVCC handling.\n"
    "Step 6: Only flag a finding as a vulnerability if you confirmed in Steps 2-5 that "
    "nondeterministic data actually reaches the ledger or resources are leaked.\n\n"
    "For each vulnerability found, provide:\n"
    "1. Vulnerability type\n"
    "2. Severity (Critical/High/Medium/Low)\n"
    "3. Affected code location (function name and line reference)\n"
    "4. Description of the issue\n"
    "5. Recommended fix\n"
    "If no vulnerabilities are found, state 'No vulnerabilities detected.'"
)

PROMPTS = {"zero_shot": PROMPT_ZERO_SHOT, "few_shot": PROMPT_FEW_SHOT, "cot": PROMPT_COT}


def resolve_vertex_config():
    project_id = os.environ.get("GOOGLE_CLOUD_PROJECT") or os.environ.get("ANTHROPIC_VERTEX_PROJECT_ID")
    region = os.environ.get("CLOUD_ML_REGION", "us-east5")
    if not project_id:
        print(f"{Fore.RED}[Error] GOOGLE_CLOUD_PROJECT not set.")
        sys.exit(1)
    return project_id, region


def get_access_token():
    import google.auth
    import google.auth.transport.requests
    creds, _ = google.auth.default(scopes=["https://www.googleapis.com/auth/cloud-platform"])
    creds.refresh(google.auth.transport.requests.Request())
    return creds.token


def make_claude_client(region, project_id):
    from anthropic import AnthropicVertex
    return AnthropicVertex(region=region, project_id=project_id)


def make_gemini_client(region, project_id):
    from openai import OpenAI
    token = get_access_token()
    base_url = (f"https://{region}-aiplatform.googleapis.com/v1beta1/"
                f"projects/{project_id}/locations/{region}/endpoints/openapi")
    return OpenAI(base_url=base_url, api_key=token), token


def audit_claude(client, model, code, filename, system_prompt):
    user_prompt = f"Analyze this Hyperledger Fabric chaincode file '{filename}' for security vulnerabilities:\n\n```go\n{code}\n```"
    last_error = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            resp = client.messages.create(
                model=model, max_tokens=INFERENCE_PARAMS["max_tokens"],
                temperature=INFERENCE_PARAMS["temperature"],
                system=system_prompt, messages=[{"role": "user", "content": user_prompt}])
            return "".join(b.text for b in resp.content if hasattr(b, "text"))
        except Exception as e:
            last_error = e
            if attempt < MAX_RETRIES:
                time.sleep(RETRY_BASE_DELAY ** attempt)
                print(f"{Fore.YELLOW}[Retry] {model} attempt {attempt}: {e}")
    return f"ERROR: {last_error}"


def audit_gemini(client, model, code, filename, system_prompt):
    user_prompt = f"Analyze this Hyperledger Fabric chaincode file '{filename}' for security vulnerabilities:\n\n```go\n{code}\n```"
    last_error = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            resp = client.chat.completions.create(
                model=model, max_tokens=INFERENCE_PARAMS["max_tokens"],
                temperature=INFERENCE_PARAMS["temperature"],
                messages=[{"role": "system", "content": system_prompt},
                          {"role": "user", "content": user_prompt}])
            return resp.choices[0].message.content
        except Exception as e:
            last_error = e
            if attempt < MAX_RETRIES:
                time.sleep(RETRY_BASE_DELAY ** attempt)
                if "401" in str(e) or "403" in str(e):
                    client.api_key = get_access_token()
    return f"ERROR: {last_error}"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", required=True)
    parser.add_argument("--backend", required=True, choices=["claude", "gemini"])
    args = parser.parse_args()

    run_start = datetime.now()
    project_id, region = resolve_vertex_config()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] No .go files"); sys.exit(1)

    # Client
    if args.backend == "claude":
        client = make_claude_client(region, project_id)
        audit_fn = lambda code, fn, sp: audit_claude(client, args.model, code, fn, sp)
    else:
        client, token = make_gemini_client(region, project_id)
        last_refresh = [time.time()]
        def _audit_gemini(code, fn, sp):
            if time.time() - last_refresh[0] > TOKEN_REFRESH_INTERVAL:
                client.api_key = get_access_token()
                last_refresh[0] = time.time()
            return audit_gemini(client, args.model, code, fn, sp)
        audit_fn = _audit_gemini

    strategies = list(PROMPTS.keys())
    total = len(strategies) * len(go_files) * N_RUNS
    short = args.model.split("@")[0].split("/")[-1]

    print(f"{Fore.GREEN}[Info] Model: {args.model} ({args.backend})")
    print(f"{Fore.GREEN}[Info] Total: {total} calls (3 prompts x 15 files x 5 runs)")

    ts = run_start.strftime("%y%m%d_%H%M")
    safe_name = short.replace(".", "_")
    output_csv = RESULTS_DIR / f"repeat_{safe_name}_{ts}.csv"
    seq = 2
    while output_csv.exists():
        output_csv = RESULTS_DIR / f"repeat_{safe_name}_{ts}_{seq}.csv"
        seq += 1

    csv_fields = ["timestamp", "run_number", "model", "prompt_strategy",
                  "file", "code_chars", "elapsed_s", "result"]
    total_records = 0
    error_count = 0

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for run_num in range(1, N_RUNS + 1):
            for strategy in strategies:
                system_prompt = PROMPTS[strategy]
                desc = f"R{run_num}|{short}|{strategy}"
                print(f"\n{Fore.YELLOW}[Run {run_num}/{N_RUNS}] {short} | {strategy}")

                for go_file in tqdm(go_files, desc=desc):
                    code = go_file.read_text(encoding="utf-8")
                    start = time.time()
                    result = audit_fn(code, go_file.name, system_prompt)
                    elapsed = round(time.time() - start, 3)

                    if result.startswith("ERROR:"):
                        error_count += 1

                    writer.writerow({
                        "timestamp": datetime.now().isoformat(),
                        "run_number": run_num,
                        "model": args.model,
                        "prompt_strategy": strategy,
                        "file": go_file.name,
                        "code_chars": len(code),
                        "elapsed_s": elapsed,
                        "result": result,
                    })
                    f.flush()
                    total_records += 1

    run_end = datetime.now()
    meta = {
        "script": "17_cloud_single_model_repeat.py",
        "model": args.model, "backend": args.backend,
        "n_runs": N_RUNS, "total_records": total_records, "error_count": error_count,
        "run_start": run_start.isoformat(), "run_end": run_end.isoformat(),
        "run_duration_s": round((run_end - run_start).total_seconds(), 1),
        "output_csv": output_csv.name,
        "inference_params": INFERENCE_PARAMS,
        "run_id": hashlib.sha256(f"{run_start.isoformat()}|{args.model}".encode()).hexdigest()[:8],
    }
    with open(output_csv.with_suffix(".meta.json"), "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    print(f"\n{Fore.GREEN}[Complete] {short}: {total_records} records, {error_count} errors, {meta['run_duration_s']}s")


if __name__ == "__main__":
    main()
