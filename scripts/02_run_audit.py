"""
02_run_audit.py  (v2.0 — 2026-02-09)
- 02_resources/dataset/ 내 .go 체인코드 파일을 순회하며
  로컬 sLM(GGUF)을 이용해 보안 취약점 감사를 수행한다.
- 결과: 03_artifacts/raw_results/audit_log_YYMMDD_HHMM.csv  (타임스탬프 기반, 덮어쓰기 없음)
- 변경 이력:
    v1.0  초기 버전, audit_log.csv 고정 파일명, elapsed 미저장
    v2.0  타임스탬프 파일명, elapsed_s/code_chars 컬럼 추가, 덮어쓰기 방지, 메타데이터 JSON 동시 생성
"""

import json
import os
import shutil
import sys
import csv
import time
from pathlib import Path
from datetime import datetime

from llama_cpp import Llama
from colorama import init, Fore, Style
from tqdm import tqdm

init(autoreset=True)

# ── 프로젝트 루트 경로 설정 ──────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

# ── 사용할 모델 목록 (GGUF 파일명) ───────────────────────────────────
MODEL_FILES = [
    "qwen2.5-coder-7b-instruct-q4_k_m.gguf",
    "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
]

# ── 추론 파라미터 (메타데이터 기록용) ─────────────────────────────────
INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 4096,
    "temperature": 0.1,
    "max_tokens": 2048,
}

# ── 시스템 프롬프트 ──────────────────────────────────────────────────
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


def generate_output_path(results_dir: Path, run_start: datetime) -> Path:
    """타임스탬프 기반 고유 CSV 경로를 생성한다. 동일 분 내 재실행 시 _2, _3 ... 접미사를 붙인다."""
    base = f"audit_log_{run_start.strftime('%y%m%d_%H%M')}"
    candidate = results_dir / f"{base}.csv"
    if not candidate.exists():
        return candidate
    # 동일 분 내 충돌 — 순번 접미사
    seq = 2
    while True:
        candidate = results_dir / f"{base}_{seq}.csv"
        if not candidate.exists():
            return candidate
        seq += 1


def load_model(model_path: str) -> Llama:
    """GGUF 모델을 GPU 최대 활용 모드로 로드한다."""
    print(f"{Fore.CYAN}[Model] 로딩 중: {Path(model_path).name}")
    llm = Llama(
        model_path=str(model_path),
        n_gpu_layers=INFERENCE_PARAMS["n_gpu_layers"],
        n_ctx=INFERENCE_PARAMS["n_ctx"],
        verbose=False,
    )
    print(f"{Fore.GREEN}[Model] 로드 완료: {Path(model_path).name}")
    return llm


def audit_chaincode(llm: Llama, code: str, filename: str) -> str:
    """체인코드에 대해 보안 감사를 수행하고 결과를 반환한다."""
    user_prompt = (
        f"Analyze this Hyperledger Fabric chaincode file '{filename}' "
        f"for security vulnerabilities:\n\n```go\n{code}\n```"
    )

    response = llm.create_chat_completion(
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": user_prompt},
        ],
        max_tokens=INFERENCE_PARAMS["max_tokens"],
        temperature=INFERENCE_PARAMS["temperature"],
    )

    return response["choices"][0]["message"]["content"]


def main():
    run_start = datetime.now()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    # ── .go 파일 수집 ────────────────────────────────────────────────
    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] {DATASET_DIR} 에 .go 파일이 없습니다.")
        print(f"{Fore.YELLOW}[Hint] 02_resources/dataset/ 에 Go 체인코드 파일을 추가하세요.")
        sys.exit(1)

    print(f"{Fore.GREEN}[Info] 발견된 체인코드 파일: {len(go_files)}개")

    # ── 타임스탬프 기반 고유 파일명 생성 (덮어쓰기 방지) ──────────────
    output_csv = generate_output_path(RESULTS_DIR, run_start)
    print(f"{Fore.GREEN}[Info] 결과 저장 경로: {output_csv.name}")

    # ── CSV 컬럼 정의 ────────────────────────────────────────────────
    csv_fields = ["timestamp", "model", "file", "code_chars", "elapsed_s", "result"]

    total_records = 0
    timing_summary = {}  # model -> [elapsed, ...]

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for model_file in MODEL_FILES:
            model_path = MODELS_DIR / model_file
            if not model_path.exists():
                print(f"{Fore.RED}[Skip] 모델 파일 없음: {model_file}")
                print(f"{Fore.YELLOW}[Hint] scripts/01_download_models.py 를 먼저 실행하세요.")
                continue

            llm = load_model(model_path)
            print(f"\n{Fore.YELLOW}{'='*60}")
            print(f"{Fore.YELLOW}[Audit] 모델: {model_file}")
            print(f"{Fore.YELLOW}{'='*60}\n")

            model_timings = []

            for go_file in tqdm(go_files, desc=f"Auditing with {model_file[:20]}..."):
                code = go_file.read_text(encoding="utf-8")
                code_chars = len(code)
                print(f"{Fore.CYAN}[Audit] {go_file.name} ({code_chars} chars)")

                start = time.time()
                try:
                    result = audit_chaincode(llm, code, go_file.name)
                except Exception as e:
                    result = f"ERROR: {e}"
                    print(f"{Fore.RED}[Error] {go_file.name}: {e}")
                elapsed = round(time.time() - start, 3)

                model_timings.append(elapsed)
                print(f"{Fore.GREEN}[Done] {go_file.name} ({elapsed:.1f}s)")

                writer.writerow({
                    "timestamp": datetime.now().isoformat(),
                    "model": model_file,
                    "file": go_file.name,
                    "code_chars": code_chars,
                    "elapsed_s": elapsed,
                    "result": result,
                })
                f.flush()
                total_records += 1

            timing_summary[model_file] = model_timings

            # 모델 메모리 해제
            del llm

    run_end = datetime.now()

    # ── 메타데이터 JSON 생성 (CSV와 동일 경로, 동일 접두사) ──────────
    meta_path = output_csv.with_suffix(".meta.json")
    meta = {
        "script_version": "2.0",
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "run_duration_s": round((run_end - run_start).total_seconds(), 1),
        "output_csv": output_csv.name,
        "output_csv_bytes": output_csv.stat().st_size,
        "total_records": total_records,
        "dataset_dir": str(DATASET_DIR),
        "dataset_files": [f.name for f in go_files],
        "dataset_file_count": len(go_files),
        "models": MODEL_FILES,
        "inference_params": INFERENCE_PARAMS,
        "system_prompt": SYSTEM_PROMPT,
        "timing_summary": {
            model: {
                "total_s": round(sum(times), 1),
                "avg_s": round(sum(times) / len(times), 3) if times else 0,
                "per_file": [round(t, 3) for t in times],
            }
            for model, times in timing_summary.items()
        },
    }
    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    # ── audit_log_latest.csv 편의 복사본 (최신 실행 결과 빠른 참조용) ─
    latest_csv = RESULTS_DIR / "audit_log_latest.csv"
    shutil.copy2(output_csv, latest_csv)

    # ── 완료 보고 ────────────────────────────────────────────────────
    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] 감사 완료")
    print(f"{Fore.GREEN}  결과 CSV  : {output_csv}")
    print(f"{Fore.GREEN}  메타데이터: {meta_path}")
    print(f"{Fore.GREEN}  최신 복사 : {latest_csv}")
    print(f"{Fore.GREEN}  총 레코드 : {total_records}건")
    print(f"{Fore.GREEN}  소요 시간 : {meta['run_duration_s']}s")
    print(f"{Fore.GREEN}{'='*60}")

    # ── 타이밍 요약 출력 ─────────────────────────────────────────────
    for model, info in meta["timing_summary"].items():
        short_name = Path(model).stem[:25]
        print(f"{Fore.CYAN}  [{short_name}] total={info['total_s']}s, avg={info['avg_s']}s/file")


if __name__ == "__main__":
    main()
