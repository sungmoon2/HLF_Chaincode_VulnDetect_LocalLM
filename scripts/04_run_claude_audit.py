"""
04_run_claude_audit.py  (v1.0 — 2026-02-09)
- 02_resources/dataset/ 내 .go 체인코드 파일을 순회하며
  Claude API (Vertex AI 백엔드)를 이용해 보안 취약점 감사를 수행한다.
- 사용 모델:
    claude-3-haiku@20240307
    claude-sonnet-4-20250514
    claude-opus-4-20250514
- 결과: 03_artifacts/raw_results/claude_audit_log_YYMMDD_HHMM.csv
- 메타: 03_artifacts/raw_results/claude_audit_log_YYMMDD_HHMM.meta.json
- 인증: Google Cloud Vertex AI (GOOGLE_CLOUD_PROJECT / CLOUD_ML_REGION 환경 변수 필요)
"""

import hashlib
import json
import os
import shutil
import sys
import csv
import time
from pathlib import Path
from datetime import datetime

from anthropic import AnthropicVertex
from colorama import init, Fore, Style
from tqdm import tqdm

init(autoreset=True)

# ── 프로젝트 루트 경로 설정 ──────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

# ── 사용할 Claude 모델 목록 ──────────────────────────────────────────
CLAUDE_MODELS = [
    "claude-haiku-4-5@20251001",
    "claude-sonnet-4-5@20250929",
    "claude-opus-4-5@20251101",
]

# ── 추론 파라미터 (로컬 모델과 동일) ─────────────────────────────────
INFERENCE_PARAMS = {
    "temperature": 0.1,
    "max_tokens": 2048,
}

# ── 재시도 설정 ──────────────────────────────────────────────────────
MAX_RETRIES = 3
RETRY_BASE_DELAY = 2  # seconds, exponential backoff base

# ── 시스템 프롬프트 (02_run_audit.py 와 동일) ────────────────────────
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


def resolve_vertex_config() -> tuple[str, str]:
    """Vertex AI 프로젝트 ID와 리전을 환경 변수에서 읽어온다.

    Returns:
        (project_id, region) 튜플

    Raises:
        SystemExit: 필수 환경 변수가 설정되지 않은 경우
    """
    project_id = os.environ.get("GOOGLE_CLOUD_PROJECT") or os.environ.get(
        "ANTHROPIC_VERTEX_PROJECT_ID"
    )
    region = os.environ.get("CLOUD_ML_REGION", "us-east5")

    if not project_id:
        print(f"{Fore.RED}{'='*60}")
        print(f"{Fore.RED}[Error] Google Cloud 프로젝트 ID가 설정되지 않았습니다.")
        print(f"{Fore.RED}{'='*60}")
        print()
        print(f"{Fore.YELLOW}다음 환경 변수 중 하나를 설정하세요:")
        print()
        print(f"{Fore.CYAN}  [Windows PowerShell]")
        print(f"{Fore.WHITE}  $env:GOOGLE_CLOUD_PROJECT = \"your-project-id\"")
        print()
        print(f"{Fore.CYAN}  [Windows CMD]")
        print(f"{Fore.WHITE}  set GOOGLE_CLOUD_PROJECT=your-project-id")
        print()
        print(f"{Fore.CYAN}  [Linux / macOS]")
        print(f"{Fore.WHITE}  export GOOGLE_CLOUD_PROJECT=\"your-project-id\"")
        print()
        print(f"{Fore.YELLOW}또는 ANTHROPIC_VERTEX_PROJECT_ID 를 사용할 수도 있습니다.")
        print()
        print(f"{Fore.YELLOW}참고: gcloud auth application-default login 을 통해")
        print(f"{Fore.YELLOW}      ADC(Application Default Credentials) 인증도 필요합니다.")
        sys.exit(1)

    return project_id, region


def generate_output_path(results_dir: Path, run_start: datetime) -> Path:
    """타임스탬프 기반 고유 CSV 경로를 생성한다. 동일 분 내 재실행 시 _2, _3 ... 접미사를 붙인다."""
    base = f"claude_audit_{run_start.strftime('%y%m%d_%H%M')}"
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


def create_client(project_id: str, region: str) -> AnthropicVertex:
    """AnthropicVertex 클라이언트를 생성한다."""
    print(f"{Fore.CYAN}[Auth] Vertex AI 클라이언트 초기화 중...")
    print(f"{Fore.CYAN}  Project : {project_id}")
    print(f"{Fore.CYAN}  Region  : {region}")
    client = AnthropicVertex(region=region, project_id=project_id)
    print(f"{Fore.GREEN}[Auth] 클라이언트 초기화 완료")
    return client


def audit_chaincode_claude(
    client: AnthropicVertex, model: str, code: str, filename: str
) -> str:
    """Claude API를 사용하여 체인코드 보안 감사를 수행한다. 최대 3회 재시도."""
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
                system=SYSTEM_PROMPT,
                messages=[
                    {"role": "user", "content": user_prompt},
                ],
            )
            # 응답 텍스트 추출
            result_text = ""
            for block in response.content:
                if hasattr(block, "text"):
                    result_text += block.text
            return result_text

        except Exception as e:
            last_error = e
            if attempt < MAX_RETRIES:
                delay = RETRY_BASE_DELAY ** attempt  # 2s, 4s, 8s
                print(
                    f"{Fore.YELLOW}[Retry] {model} / {filename} — "
                    f"attempt {attempt}/{MAX_RETRIES} failed: {e}"
                )
                print(f"{Fore.YELLOW}[Retry] {delay}s 후 재시도...")
                time.sleep(delay)
            else:
                print(
                    f"{Fore.RED}[Error] {model} / {filename} — "
                    f"최대 재시도 횟수 초과 ({MAX_RETRIES}회): {e}"
                )

    return f"ERROR: {last_error}"


def main():
    run_start = datetime.now()

    # ── Vertex AI 설정 확인 ──────────────────────────────────────────
    project_id, region = resolve_vertex_config()

    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    # ── .go 파일 수집 ────────────────────────────────────────────────
    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] {DATASET_DIR} 에 .go 파일이 없습니다.")
        print(f"{Fore.YELLOW}[Hint] 02_resources/dataset/ 에 Go 체인코드 파일을 추가하세요.")
        sys.exit(1)

    print(f"{Fore.GREEN}[Info] 발견된 체인코드 파일: {len(go_files)}개")

    # ── Vertex AI 클라이언트 생성 ────────────────────────────────────
    client = create_client(project_id, region)

    # ── 타임스탬프 기반 고유 파일명 생성 (덮어쓰기 방지) ──────────────
    output_csv = generate_output_path(RESULTS_DIR, run_start)
    print(f"{Fore.GREEN}[Info] 결과 저장 경로: {output_csv.name}")

    # ── CSV 컬럼 정의 ────────────────────────────────────────────────
    csv_fields = ["timestamp", "model", "prompt_strategy", "file", "code_chars", "elapsed_s", "result"]

    total_records = 0
    error_count = 0
    timing_summary = {}  # model -> [elapsed, ...]

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for model_name in CLAUDE_MODELS:
            print(f"\n{Fore.YELLOW}{'='*60}")
            print(f"{Fore.YELLOW}[Audit] 모델: {model_name}")
            print(f"{Fore.YELLOW}{'='*60}\n")

            model_timings = []

            for go_file in tqdm(go_files, desc=f"Auditing with {model_name}"):
                code = go_file.read_text(encoding="utf-8")
                code_chars = len(code)
                print(f"{Fore.CYAN}[Audit] {go_file.name} ({code_chars} chars)")

                start = time.time()
                result = audit_chaincode_claude(
                    client, model_name, code, go_file.name
                )
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
                    "prompt_strategy": "zero_shot",
                    "file": go_file.name,
                    "code_chars": code_chars,
                    "elapsed_s": elapsed,
                    "result": result,
                })
                f.flush()
                total_records += 1

            timing_summary[model_name] = model_timings

    run_end = datetime.now()

    # ── 메타데이터 JSON 생성 (CSV와 동일 경로, 동일 접두사) ──────────
    meta_path = output_csv.with_suffix(".meta.json")
    meta = {
        "script_version": "1.0",
        "script": "04_run_claude_audit.py",
        "backend": "Vertex AI (AnthropicVertex)",
        "vertex_project_id": project_id,
        "vertex_region": region,
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
        "models": CLAUDE_MODELS,
        "inference_params": INFERENCE_PARAMS,
        "retry_config": {
            "max_retries": MAX_RETRIES,
            "retry_base_delay_s": RETRY_BASE_DELAY,
        },
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
    # ── run_id 생성 (실험 세트 연결용) ──────────────────────────────
    run_id_source = f"{run_start.isoformat()}|claude|{','.join(CLAUDE_MODELS)}"
    run_id = hashlib.sha256(run_id_source.encode()).hexdigest()[:8]
    meta["run_id"] = run_id

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    # ── claude_audit_latest.csv 편의 복사본 ────────────────────────
    latest_csv = RESULTS_DIR / "claude_audit_latest.csv"
    shutil.copy2(output_csv, latest_csv)

    # ── 완료 보고 ────────────────────────────────────────────────────
    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] Claude 감사 완료")
    print(f"{Fore.GREEN}  결과 CSV  : {output_csv}")
    print(f"{Fore.GREEN}  메타데이터: {meta_path}")
    print(f"{Fore.GREEN}  최신 복사 : {latest_csv}")
    print(f"{Fore.GREEN}  총 레코드 : {total_records}건")
    if error_count > 0:
        print(f"{Fore.RED}  오류 건수 : {error_count}건")
    print(f"{Fore.GREEN}  소요 시간 : {meta['run_duration_s']}s")
    print(f"{Fore.GREEN}{'='*60}")

    # ── 타이밍 요약 출력 ─────────────────────────────────────────────
    for model, info in meta["timing_summary"].items():
        print(f"{Fore.CYAN}  [{model}] total={info['total_s']}s, avg={info['avg_s']}s/file")


if __name__ == "__main__":
    main()
