"""
14_local_repeat.py  (v1.0 — 2026-02-10)
- 로컬 모델 반복 실험: Qwen2.5-Coder-7B + Llama-3.1-8B
- 3 prompts (zero_shot, few_shot, cot) x 15 files x 5 runs
- temperature=0.1 (원래 실험과 동일 조건)
- 결과: 03_artifacts/raw_results/local_repeat_YYMMDD_HHMM.csv
- GPU 순차 실행 (모델 1개씩 로드)
"""

import hashlib
import json
import csv
import time
import sys
from pathlib import Path
from datetime import datetime

from llama_cpp import Llama
from colorama import init, Fore
from tqdm import tqdm

init(autoreset=True)

# ── 경로 설정 ──────────────────────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

# ── 모델 목록 ──────────────────────────────────────────────────────────
MODEL_FILES = [
    "qwen2.5-coder-7b-instruct-q4_k_m.gguf",
    "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
]

# ── 추론 파라미터 ──────────────────────────────────────────────────────
INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 4096,
    "temperature": 0.1,
    "max_tokens": 2048,
}

N_RUNS = 5

# ── 프롬프트 정의 (02_run_audit_v3.py와 동일) ─────────────────────────
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

PROMPTS = {
    "zero_shot": PROMPT_ZERO_SHOT,
    "few_shot": PROMPT_FEW_SHOT,
    "cot": PROMPT_COT,
}


def load_model(model_path: Path) -> Llama:
    print(f"{Fore.CYAN}[Model] Loading: {model_path.name}")
    llm = Llama(
        model_path=str(model_path),
        n_gpu_layers=INFERENCE_PARAMS["n_gpu_layers"],
        n_ctx=INFERENCE_PARAMS["n_ctx"],
        verbose=False,
    )
    print(f"{Fore.GREEN}[Model] Loaded: {model_path.name}")
    return llm


def audit(llm: Llama, code: str, filename: str, system_prompt: str) -> str:
    user_prompt = (
        f"Analyze this Hyperledger Fabric chaincode file '{filename}' "
        f"for security vulnerabilities:\n\n```go\n{code}\n```"
    )
    response = llm.create_chat_completion(
        messages=[
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ],
        max_tokens=INFERENCE_PARAMS["max_tokens"],
        temperature=INFERENCE_PARAMS["temperature"],
    )
    return response["choices"][0]["message"]["content"]


def main():
    run_start = datetime.now()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] No .go files in {DATASET_DIR}")
        sys.exit(1)

    strategies = list(PROMPTS.keys())
    total = len(MODEL_FILES) * len(strategies) * len(go_files) * N_RUNS
    print(f"{Fore.GREEN}[Info] Local Repeat Experiment")
    print(f"{Fore.GREEN}  Models: {len(MODEL_FILES)}")
    print(f"{Fore.GREEN}  Prompts: {strategies}")
    print(f"{Fore.GREEN}  Files: {len(go_files)}")
    print(f"{Fore.GREEN}  Runs: {N_RUNS}")
    print(f"{Fore.GREEN}  Total inferences: {total}")

    ts = run_start.strftime("%y%m%d_%H%M")
    output_csv = RESULTS_DIR / f"local_repeat_{ts}.csv"
    seq = 2
    while output_csv.exists():
        output_csv = RESULTS_DIR / f"local_repeat_{ts}_{seq}.csv"
        seq += 1

    csv_fields = [
        "timestamp", "run_number", "model", "prompt_strategy",
        "file", "code_chars", "elapsed_s", "result",
    ]

    total_records = 0
    error_count = 0
    timing_data = {}  # "model|strategy|run" -> [elapsed, ...]

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for model_file in MODEL_FILES:
            model_path = MODELS_DIR / model_file
            if not model_path.exists():
                print(f"{Fore.RED}[Skip] Model not found: {model_file}")
                continue

            llm = load_model(model_path)
            short_name = model_file.split(".")[0][:20]

            for run_num in range(1, N_RUNS + 1):
                for strategy in strategies:
                    system_prompt = PROMPTS[strategy]
                    run_key = f"{model_file}|{strategy}|run{run_num}"
                    timing_data[run_key] = []

                    desc = f"{short_name}|{strategy}|R{run_num}"
                    print(f"\n{Fore.YELLOW}{'='*60}")
                    print(f"{Fore.YELLOW}[Run {run_num}/{N_RUNS}] {model_file} | {strategy}")
                    print(f"{Fore.YELLOW}{'='*60}")

                    for go_file in tqdm(go_files, desc=desc):
                        code = go_file.read_text(encoding="utf-8")
                        code_chars = len(code)

                        start = time.time()
                        try:
                            result = audit(llm, code, go_file.name, system_prompt)
                        except Exception as e:
                            result = f"ERROR: {e}"
                            error_count += 1
                        elapsed = round(time.time() - start, 3)

                        timing_data[run_key].append(elapsed)

                        writer.writerow({
                            "timestamp": datetime.now().isoformat(),
                            "run_number": run_num,
                            "model": model_file,
                            "prompt_strategy": strategy,
                            "file": go_file.name,
                            "code_chars": code_chars,
                            "elapsed_s": elapsed,
                            "result": result,
                        })
                        f.flush()
                        total_records += 1

            del llm
            print(f"{Fore.GREEN}[Model] Released: {model_file}")

    run_end = datetime.now()

    # ── Meta JSON ──────────────────────────────────────────────────────
    meta_path = output_csv.with_suffix(".meta.json")
    meta = {
        "script": "14_local_repeat.py",
        "script_version": "1.0",
        "experiment": "Local model repeat experiment (reproducibility)",
        "n_runs": N_RUNS,
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "run_duration_s": round((run_end - run_start).total_seconds(), 1),
        "output_csv": output_csv.name,
        "output_csv_bytes": output_csv.stat().st_size,
        "total_records": total_records,
        "error_count": error_count,
        "models": MODEL_FILES,
        "prompt_strategies": strategies,
        "inference_params": INFERENCE_PARAMS,
        "dataset_dir": str(DATASET_DIR),
        "dataset_files": [f.name for f in go_files],
        "dataset_file_count": len(go_files),
        "run_id": hashlib.sha256(
            f"{run_start.isoformat()}|local_repeat".encode()
        ).hexdigest()[:8],
    }

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] Local Repeat Experiment")
    print(f"{Fore.GREEN}  CSV: {output_csv.name}")
    print(f"{Fore.GREEN}  Records: {total_records}")
    print(f"{Fore.GREEN}  Errors: {error_count}")
    print(f"{Fore.GREEN}  Duration: {meta['run_duration_s']}s")
    print(f"{Fore.GREEN}{'='*60}")


if __name__ == "__main__":
    main()
