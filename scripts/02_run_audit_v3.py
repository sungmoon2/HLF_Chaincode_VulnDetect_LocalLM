"""
02_run_audit_v3.py  (v3.0 — 2026-02-09)
- 다중 프롬프트 전략(zero-shot, few-shot, chain-of-thought) 지원
- 다중 데이터셋 디렉토리(original, obfuscated) 지원
- CLI 인자: --prompts, --dataset-dir, --tag
- 결과: 03_artifacts/raw_results/audit_v3_{tag}_YYMMDD_HHMM.csv
- 변경 이력:
    v1.0  초기 버전
    v2.0  타임스탬프 파일명, elapsed_s/code_chars 컬럼 추가
    v3.0  다중 프롬프트, 다중 데이터셋, CLI 인자, prompt_strategy 컬럼 추가
"""

import argparse
import hashlib
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
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

# ── 사용할 모델 목록 (GGUF 파일명) ───────────────────────────────────
MODEL_FILES = [
    "qwen2.5-coder-7b-instruct-q4_k_m.gguf",
    "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
]

# ── 추론 파라미터 ─────────────────────────────────────────────────────
INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 4096,
    "temperature": 0.1,
    "max_tokens": 2048,
}

# ═══════════════════════════════════════════════════════════════════════
# 프롬프트 전략 정의
# ═══════════════════════════════════════════════════════════════════════

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


def generate_output_path(results_dir: Path, tag: str, run_start: datetime) -> Path:
    """태그 + 타임스탬프 기반 고유 CSV 경로를 생성한다."""
    base = f"audit_v3_{tag}_{run_start.strftime('%y%m%d_%H%M')}"
    candidate = results_dir / f"{base}.csv"
    if not candidate.exists():
        return candidate
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


def audit_chaincode(llm: Llama, code: str, filename: str, system_prompt: str) -> str:
    """체인코드에 대해 보안 감사를 수행하고 결과를 반환한다."""
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


def parse_args():
    parser = argparse.ArgumentParser(description="HLF Chaincode Audit v3 — Multi-prompt, Multi-dataset")
    parser.add_argument(
        "--prompts",
        nargs="+",
        choices=["zero_shot", "few_shot", "cot"],
        default=["zero_shot", "few_shot", "cot"],
        help="프롬프트 전략 선택 (기본: 3가지 모두)",
    )
    parser.add_argument(
        "--dataset-dir",
        type=str,
        default=None,
        help="데이터셋 디렉토리 경로 (기본: 02_resources/dataset/)",
    )
    parser.add_argument(
        "--tag",
        type=str,
        default="full",
        help="실행 태그 (파일명에 반영, 기본: full)",
    )
    parser.add_argument(
        "--models",
        nargs="+",
        default=None,
        help="사용할 모델 파일명 (기본: 전체)",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    run_start = datetime.now()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    # ── 데이터셋 디렉토리 결정 ────────────────────────────────────────
    if args.dataset_dir:
        dataset_dir = Path(args.dataset_dir)
        if not dataset_dir.is_absolute():
            dataset_dir = PROJECT_ROOT / args.dataset_dir
    else:
        dataset_dir = PROJECT_ROOT / "02_resources" / "dataset"

    # ── .go 파일 수집 ────────────────────────────────────────────────
    go_files = sorted(dataset_dir.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] {dataset_dir} 에 .go 파일이 없습니다.")
        sys.exit(1)

    # ── 모델 목록 ────────────────────────────────────────────────────
    model_files = args.models if args.models else MODEL_FILES

    # ── 프롬프트 전략 목록 ───────────────────────────────────────────
    prompt_strategies = args.prompts

    total_runs = len(model_files) * len(go_files) * len(prompt_strategies)
    print(f"{Fore.GREEN}[Info] 데이터셋: {dataset_dir.name} ({len(go_files)}개 파일)")
    print(f"{Fore.GREEN}[Info] 모델: {len(model_files)}개")
    print(f"{Fore.GREEN}[Info] 프롬프트 전략: {prompt_strategies}")
    print(f"{Fore.GREEN}[Info] 총 감사 예정: {total_runs}건")

    # ── 출력 경로 생성 ───────────────────────────────────────────────
    output_csv = generate_output_path(RESULTS_DIR, args.tag, run_start)
    print(f"{Fore.GREEN}[Info] 결과 저장 경로: {output_csv.name}")

    # ── CSV 컬럼 정의 ────────────────────────────────────────────────
    csv_fields = ["timestamp", "model", "prompt_strategy", "file", "code_chars", "elapsed_s", "result"]

    total_records = 0
    timing_summary = {}  # "model|prompt" -> [elapsed, ...]

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for model_file in model_files:
            model_path = MODELS_DIR / model_file
            if not model_path.exists():
                print(f"{Fore.RED}[Skip] 모델 파일 없음: {model_file}")
                continue

            llm = load_model(model_path)

            for strategy in prompt_strategies:
                system_prompt = PROMPTS[strategy]
                run_key = f"{model_file}|{strategy}"
                timing_summary[run_key] = []

                print(f"\n{Fore.YELLOW}{'='*60}")
                print(f"{Fore.YELLOW}[Audit] 모델: {model_file}")
                print(f"{Fore.YELLOW}[Audit] 프롬프트: {strategy}")
                print(f"{Fore.YELLOW}{'='*60}\n")

                desc = f"{Path(model_file).stem[:15]}|{strategy}"
                for go_file in tqdm(go_files, desc=desc):
                    code = go_file.read_text(encoding="utf-8")
                    code_chars = len(code)
                    print(f"{Fore.CYAN}[Audit] {go_file.name} ({code_chars} chars) [{strategy}]")

                    start = time.time()
                    try:
                        result = audit_chaincode(llm, code, go_file.name, system_prompt)
                    except Exception as e:
                        result = f"ERROR: {e}"
                        print(f"{Fore.RED}[Error] {go_file.name}: {e}")
                    elapsed = round(time.time() - start, 3)

                    timing_summary[run_key].append(elapsed)
                    print(f"{Fore.GREEN}[Done] {go_file.name} ({elapsed:.1f}s) [{strategy}]")

                    writer.writerow({
                        "timestamp": datetime.now().isoformat(),
                        "model": model_file,
                        "prompt_strategy": strategy,
                        "file": go_file.name,
                        "code_chars": code_chars,
                        "elapsed_s": elapsed,
                        "result": result,
                    })
                    f.flush()
                    total_records += 1

            # 모델 메모리 해제
            del llm

    run_end = datetime.now()

    # ── 메타데이터 JSON 생성 ──────────────────────────────────────────
    meta_path = output_csv.with_suffix(".meta.json")
    meta = {
        "script_version": "3.0",
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "run_duration_s": round((run_end - run_start).total_seconds(), 1),
        "output_csv": output_csv.name,
        "output_csv_bytes": output_csv.stat().st_size,
        "total_records": total_records,
        "dataset_dir": str(dataset_dir),
        "dataset_files": [f.name for f in go_files],
        "dataset_file_count": len(go_files),
        "models": model_files,
        "prompt_strategies": prompt_strategies,
        "prompts": {k: v for k, v in PROMPTS.items() if k in prompt_strategies},
        "inference_params": INFERENCE_PARAMS,
        "tag": args.tag,
        "timing_summary": {
            key: {
                "total_s": round(sum(times), 1),
                "avg_s": round(sum(times) / len(times), 3) if times else 0,
                "per_file": [round(t, 3) for t in times],
            }
            for key, times in timing_summary.items()
        },
    }
    # ── run_id 생성 (실험 세트 연결용) ──────────────────────────────
    run_id_source = f"{run_start.isoformat()}|{args.tag}|{','.join(prompt_strategies)}|{','.join(model_files)}"
    run_id = hashlib.sha256(run_id_source.encode()).hexdigest()[:8]
    meta["run_id"] = run_id

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    # ── audit_v3_latest.csv 편의 복사본 ──────────────────────────────
    latest_csv = RESULTS_DIR / f"audit_v3_{args.tag}_latest.csv"
    shutil.copy2(output_csv, latest_csv)

    # ── 완료 보고 ────────────────────────────────────────────────────
    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] 감사 완료")
    print(f"{Fore.GREEN}  결과 CSV  : {output_csv}")
    print(f"{Fore.GREEN}  메타데이터: {meta_path}")
    print(f"{Fore.GREEN}  최신 복사 : {latest_csv}")
    print(f"{Fore.GREEN}  총 레코드 : {total_records}건")
    print(f"{Fore.GREEN}  소요 시간 : {meta['run_duration_s']}s")
    print(f"{Fore.GREEN}  Run ID    : {run_id}")
    print(f"{Fore.GREEN}{'='*60}")

    for key, info in meta["timing_summary"].items():
        print(f"{Fore.CYAN}  [{key}] total={info['total_s']}s, avg={info['avg_s']}s/file")


if __name__ == "__main__":
    main()
