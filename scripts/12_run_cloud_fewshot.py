"""
12_run_cloud_fewshot.py  (v1.0 — 2026-02-10)
- 마이크로벤치마크 15파일 x Claude 3모델 + Gemini 3모델 (few-shot)
- B3 보완 실험: 클라우드 모델 프롬프트 공정성 검증
- 결과: 03_artifacts/raw_results/cloud_fewshot_YYMMDD_HHMM.csv
- 인증: Google Cloud Vertex AI (ADC + AnthropicVertex)
"""

import hashlib
import json
import os
import sys
import csv
import time
from pathlib import Path
from datetime import datetime

import google.auth
import google.auth.transport.requests
from anthropic import AnthropicVertex
from openai import OpenAI
from colorama import init, Fore
from tqdm import tqdm

init(autoreset=True)

# ── 경로 설정 ──────────────────────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

# ── 모델 목록 ──────────────────────────────────────────────────────────
CLAUDE_MODELS = [
    "claude-haiku-4-5@20251001",
    "claude-sonnet-4-5@20250929",
    "claude-opus-4-5@20251101",
]
GEMINI_MODELS = [
    "google/gemini-2.5-pro",
    "google/gemini-2.5-flash",
    "google/gemini-2.5-flash-lite",
]

# ── 추론 파라미터 ──────────────────────────────────────────────────────
INFERENCE_PARAMS = {"temperature": 0.1, "max_tokens": 2048}
MAX_RETRIES = 3
RETRY_BASE_DELAY = 2

# ── Few-shot 프롬프트 (02_run_audit_v3.py와 동일) ─────────────────────
SYSTEM_PROMPT_FEW_SHOT = (
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


def resolve_vertex_config() -> tuple[str, str]:
    project_id = os.environ.get("GOOGLE_CLOUD_PROJECT") or os.environ.get(
        "ANTHROPIC_VERTEX_PROJECT_ID"
    )
    region = os.environ.get("CLOUD_ML_REGION", "us-east5")
    if not project_id:
        print(f"{Fore.RED}[Error] GOOGLE_CLOUD_PROJECT 환경 변수가 설정되지 않았습니다.")
        sys.exit(1)
    return project_id, region


def get_access_token() -> str:
    credentials, _ = google.auth.default(
        scopes=["https://www.googleapis.com/auth/cloud-platform"]
    )
    auth_request = google.auth.transport.requests.Request()
    credentials.refresh(auth_request)
    return credentials.token


def audit_claude(client: AnthropicVertex, model: str, code: str, filename: str) -> str:
    user_prompt = (
        f"Analyze this Hyperledger Fabric chaincode file '{filename}' "
        f"for security vulnerabilities:\n\n```go\n{code}\n```"
    )
    last_error = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            response = client.messages.create(
                model=model,
                max_tokens=INFERENCE_PARAMS["max_tokens"],
                temperature=INFERENCE_PARAMS["temperature"],
                system=SYSTEM_PROMPT_FEW_SHOT,
                messages=[{"role": "user", "content": user_prompt}],
            )
            return "".join(b.text for b in response.content if hasattr(b, "text"))
        except Exception as e:
            last_error = e
            if attempt < MAX_RETRIES:
                delay = RETRY_BASE_DELAY ** attempt
                print(f"{Fore.YELLOW}[Retry] {model}/{filename} attempt {attempt}: {e}")
                time.sleep(delay)
    return f"ERROR: {last_error}"


def audit_gemini(client: OpenAI, model: str, code: str, filename: str) -> str:
    user_prompt = (
        f"Analyze this Hyperledger Fabric chaincode file '{filename}' "
        f"for security vulnerabilities:\n\n```go\n{code}\n```"
    )
    last_error = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            response = client.chat.completions.create(
                model=model,
                max_tokens=INFERENCE_PARAMS["max_tokens"],
                temperature=INFERENCE_PARAMS["temperature"],
                messages=[
                    {"role": "system", "content": SYSTEM_PROMPT_FEW_SHOT},
                    {"role": "user", "content": user_prompt},
                ],
            )
            return response.choices[0].message.content
        except Exception as e:
            last_error = e
            if attempt < MAX_RETRIES:
                delay = RETRY_BASE_DELAY ** attempt
                print(f"{Fore.YELLOW}[Retry] {model}/{filename} attempt {attempt}: {e}")
                time.sleep(delay)
                if "401" in str(e) or "403" in str(e):
                    token = get_access_token()
                    client.api_key = token
    return f"ERROR: {last_error}"


def generate_output_path(results_dir: Path, run_start: datetime) -> Path:
    base = f"cloud_fewshot_{run_start.strftime('%y%m%d_%H%M')}"
    candidate = results_dir / f"{base}.csv"
    if not candidate.exists():
        return candidate
    seq = 2
    while True:
        candidate = results_dir / f"{base}_{seq}.csv"
        if not candidate.exists():
            return candidate
        seq += 1


def main():
    run_start = datetime.now()
    project_id, region = resolve_vertex_config()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] {DATASET_DIR}에 .go 파일이 없습니다.")
        sys.exit(1)

    print(f"{Fore.GREEN}[Info] 마이크로벤치마크 파일: {len(go_files)}개")

    # ── 클라이언트 생성 ──────────────────────────────────────────────
    claude_client = AnthropicVertex(region=region, project_id=project_id)
    print(f"{Fore.GREEN}[Auth] Claude Vertex AI 클라이언트 초기화 완료")

    token = get_access_token()
    gemini_base_url = (
        f"https://{region}-aiplatform.googleapis.com/v1beta1/"
        f"projects/{project_id}/locations/{region}/endpoints/openapi"
    )
    gemini_client = OpenAI(base_url=gemini_base_url, api_key=token)
    print(f"{Fore.GREEN}[Auth] Gemini OpenAI 호환 클라이언트 초기화 완료")

    all_models = [(m, "claude") for m in CLAUDE_MODELS] + [(m, "gemini") for m in GEMINI_MODELS]
    total_runs = len(all_models) * len(go_files)
    print(f"{Fore.GREEN}[Info] 총 감사 예정: {total_runs}건 ({len(all_models)}모델 x {len(go_files)}파일, few-shot)")

    output_csv = generate_output_path(RESULTS_DIR, run_start)
    csv_fields = ["timestamp", "model", "prompt_strategy", "file", "code_chars", "elapsed_s", "result"]

    total_records = 0
    error_count = 0
    timing_summary = {}

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for model_name, backend in all_models:
            print(f"\n{Fore.YELLOW}{'='*60}")
            print(f"{Fore.YELLOW}[Audit] {backend.upper()} few-shot: {model_name}")
            print(f"{Fore.YELLOW}{'='*60}")

            model_timings = []
            short_name = model_name.split("@")[0] if "@" in model_name else model_name.split("/")[-1]

            for go_file in tqdm(go_files, desc=f"{short_name}|few_shot"):
                code = go_file.read_text(encoding="utf-8")
                code_chars = len(code)

                start = time.time()
                if backend == "claude":
                    result = audit_claude(claude_client, model_name, code, go_file.name)
                else:
                    result = audit_gemini(gemini_client, model_name, code, go_file.name)
                elapsed = round(time.time() - start, 3)

                model_timings.append(elapsed)
                if result.startswith("ERROR:"):
                    error_count += 1
                    print(f"{Fore.RED}[Fail] {go_file.name} ({elapsed:.1f}s)")
                else:
                    print(f"{Fore.GREEN}[Done] {go_file.name} ({elapsed:.1f}s)")

                writer.writerow({
                    "timestamp": datetime.now().isoformat(),
                    "model": model_name,
                    "prompt_strategy": "few_shot",
                    "file": go_file.name,
                    "code_chars": code_chars,
                    "elapsed_s": elapsed,
                    "result": result,
                })
                f.flush()
                total_records += 1

            timing_summary[model_name] = model_timings

    run_end = datetime.now()

    meta_path = output_csv.with_suffix(".meta.json")
    meta = {
        "script_version": "1.0",
        "script": "12_run_cloud_fewshot.py",
        "experiment": "B3 - Micro-benchmark x Cloud Models (few-shot)",
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "run_duration_s": round((run_end - run_start).total_seconds(), 1),
        "output_csv": output_csv.name,
        "output_csv_bytes": output_csv.stat().st_size,
        "total_records": total_records,
        "error_count": error_count,
        "dataset_dir": str(DATASET_DIR),
        "dataset_files": [f.name for f in go_files],
        "dataset_file_count": len(go_files),
        "claude_models": CLAUDE_MODELS,
        "gemini_models": GEMINI_MODELS,
        "inference_params": INFERENCE_PARAMS,
        "system_prompt": "PROMPT_FEW_SHOT (same as 02_run_audit_v3.py)",
        "timing_summary": {
            model: {
                "total_s": round(sum(times), 1),
                "avg_s": round(sum(times) / len(times), 3) if times else 0,
                "per_file": [round(t, 3) for t in times],
            }
            for model, times in timing_summary.items()
        },
    }
    run_id = hashlib.sha256(f"{run_start.isoformat()}|cloud_fewshot".encode()).hexdigest()[:8]
    meta["run_id"] = run_id

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] 클라우드 모델 Few-shot 감사 완료")
    print(f"{Fore.GREEN}  결과 CSV  : {output_csv}")
    print(f"{Fore.GREEN}  총 레코드 : {total_records}건")
    if error_count > 0:
        print(f"{Fore.RED}  오류 건수 : {error_count}건")
    print(f"{Fore.GREEN}  소요 시간 : {meta['run_duration_s']}s")
    print(f"{Fore.GREEN}{'='*60}")

    for model, info in meta["timing_summary"].items():
        print(f"{Fore.CYAN}  [{model}] total={info['total_s']}s, avg={info['avg_s']}s/file")


if __name__ == "__main__":
    main()
