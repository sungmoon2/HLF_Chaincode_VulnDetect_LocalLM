"""
13_run_golisa_re_llama.py  (v1.0 — 2026-02-10)
- GoLiSA Running_Examples 5파일 x Llama-3.1-8B (zero-shot)
- B2 보완 실험: Llama 기초 판별 확인 (마이크로벤치마크 TNR 1/6 모델)
- 결과: 03_artifacts/raw_results/golisa_re_llama_YYMMDD_HHMM.csv
"""

import hashlib
import json
import sys
import csv
import time
from pathlib import Path
from datetime import datetime

from llama_cpp import Llama
from colorama import init, Fore
from tqdm import tqdm

init(autoreset=True)

# ── 경로 설정 ──────────────────────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
RUNNING_EXAMPLES_DIR = (
    PROJECT_ROOT / "02_resources" / "golisa_benchmark" / "Benchmark" / "Running_Examples"
)
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

MODEL_FILE = "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf"

# ── 추론 파라미터 ──────────────────────────────────────────────────────
INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 4096,
    "temperature": 0.1,
    "max_tokens": 2048,
}

# ── 시스템 프롬프트 (기존 실험과 동일) ─────────────────────────────────
SYSTEM_PROMPT = (
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


def collect_go_files(base_dir: Path) -> list[Path]:
    return sorted(base_dir.rglob("*.go"))


def generate_output_path(results_dir: Path, run_start: datetime) -> Path:
    base = f"golisa_re_llama_{run_start.strftime('%y%m%d_%H%M')}"
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
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    go_files = collect_go_files(RUNNING_EXAMPLES_DIR)
    if not go_files:
        print(f"{Fore.RED}[Error] {RUNNING_EXAMPLES_DIR}에 .go 파일이 없습니다.")
        sys.exit(1)

    print(f"{Fore.GREEN}[Info] Running_Examples 파일: {len(go_files)}개")
    for gf in go_files:
        print(f"  {gf.relative_to(RUNNING_EXAMPLES_DIR)}")

    model_path = MODELS_DIR / MODEL_FILE
    if not model_path.exists():
        print(f"{Fore.RED}[Error] 모델 파일 없음: {model_path}")
        sys.exit(1)

    print(f"{Fore.CYAN}[Model] 로딩 중: {MODEL_FILE}")
    llm = Llama(
        model_path=str(model_path),
        n_gpu_layers=INFERENCE_PARAMS["n_gpu_layers"],
        n_ctx=INFERENCE_PARAMS["n_ctx"],
        verbose=False,
    )
    print(f"{Fore.GREEN}[Model] 로드 완료: {MODEL_FILE}")

    output_csv = generate_output_path(RESULTS_DIR, run_start)
    csv_fields = ["timestamp", "model", "prompt_strategy", "file", "code_chars", "elapsed_s", "result"]

    total_records = 0
    timing_list = []

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for go_file in tqdm(go_files, desc=f"Llama zero-shot"):
            code = go_file.read_text(encoding="utf-8")
            code_chars = len(code)
            rel_name = str(go_file.relative_to(RUNNING_EXAMPLES_DIR)).replace("\\", "/")

            user_prompt = (
                f"Analyze this Hyperledger Fabric chaincode file '{rel_name}' "
                f"for security vulnerabilities:\n\n```go\n{code}\n```"
            )

            start = time.time()
            try:
                response = llm.create_chat_completion(
                    messages=[
                        {"role": "system", "content": SYSTEM_PROMPT},
                        {"role": "user", "content": user_prompt},
                    ],
                    max_tokens=INFERENCE_PARAMS["max_tokens"],
                    temperature=INFERENCE_PARAMS["temperature"],
                )
                result = response["choices"][0]["message"]["content"]
            except Exception as e:
                result = f"ERROR: {e}"
                print(f"{Fore.RED}[Error] {rel_name}: {e}")
            elapsed = round(time.time() - start, 3)

            timing_list.append(elapsed)
            print(f"{Fore.GREEN}[Done] {rel_name} ({elapsed:.1f}s)")

            writer.writerow({
                "timestamp": datetime.now().isoformat(),
                "model": MODEL_FILE,
                "prompt_strategy": "zero_shot",
                "file": rel_name,
                "code_chars": code_chars,
                "elapsed_s": elapsed,
                "result": result,
            })
            f.flush()
            total_records += 1

    del llm
    run_end = datetime.now()

    meta_path = output_csv.with_suffix(".meta.json")
    meta = {
        "script_version": "1.0",
        "script": "13_run_golisa_re_llama.py",
        "experiment": "B2 - GoLiSA Running_Examples x Llama-3.1-8B (zero-shot)",
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "run_duration_s": round((run_end - run_start).total_seconds(), 1),
        "output_csv": output_csv.name,
        "output_csv_bytes": output_csv.stat().st_size,
        "total_records": total_records,
        "dataset_dir": str(RUNNING_EXAMPLES_DIR),
        "dataset_files": [str(f.relative_to(RUNNING_EXAMPLES_DIR)).replace("\\", "/") for f in go_files],
        "dataset_file_count": len(go_files),
        "model": MODEL_FILE,
        "inference_params": INFERENCE_PARAMS,
        "system_prompt": SYSTEM_PROMPT,
        "timing": {
            "total_s": round(sum(timing_list), 1),
            "avg_s": round(sum(timing_list) / len(timing_list), 3) if timing_list else 0,
            "per_file": [round(t, 3) for t in timing_list],
        },
    }
    run_id = hashlib.sha256(f"{run_start.isoformat()}|golisa_re_llama".encode()).hexdigest()[:8]
    meta["run_id"] = run_id

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] GoLiSA Running_Examples Llama 감사 완료")
    print(f"{Fore.GREEN}  결과 CSV  : {output_csv}")
    print(f"{Fore.GREEN}  총 레코드 : {total_records}건")
    print(f"{Fore.GREEN}  소요 시간 : {meta['run_duration_s']}s")
    print(f"{Fore.GREEN}{'='*60}")

    for i, gf in enumerate(go_files):
        rel = str(gf.relative_to(RUNNING_EXAMPLES_DIR)).replace("\\", "/")
        print(f"{Fore.CYAN}  [{rel}] {timing_list[i]:.1f}s")


if __name__ == "__main__":
    main()
