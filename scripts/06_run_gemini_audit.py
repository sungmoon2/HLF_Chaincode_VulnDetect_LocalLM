"""
06_run_gemini_audit.py  (v1.0 — 2026-02-09)
- 02_resources/dataset/ 내 .go 체인코드 파일을 순회하며
  Gemini API (Vertex AI OpenAI 호환 엔드포인트)를 이용해 보안 취약점 감사를 수행한다.
- 사용 모델:
    google/gemini-2.5-pro
    google/gemini-2.5-flash
    google/gemini-2.5-flash-lite
- 결과: 03_artifacts/raw_results/gemini_audit_YYMMDD_HHMM.csv
- 메타: 03_artifacts/raw_results/gemini_audit_YYMMDD_HHMM.meta.json
- 인증: Google Cloud ADC (gcloud auth application-default login)
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

import google.auth
import google.auth.transport.requests
from openai import OpenAI
from colorama import init, Fore, Style
from tqdm import tqdm

init(autoreset=True)

# ── 프로젝트 루트 경로 설정 ──────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

# ── 사용할 Gemini 모델 목록 ──────────────────────────────────────────
GEMINI_MODELS = [
    "google/gemini-2.5-pro",
    "google/gemini-2.5-flash",
    "google/gemini-2.5-flash-lite",
]

# ── 추론 파라미터 (로컬 모델과 동일) ─────────────────────────────────
INFERENCE_PARAMS = {
    "temperature": 0.1,
    "max_tokens": 2048,
}

# ── 재시도 설정 ──────────────────────────────────────────────────────
MAX_RETRIES = 3
RETRY_BASE_DELAY = 2  # seconds, exponential backoff base

# ── 시스템 프롬프트 (다른 스크립트와 동일) ────────────────────────────
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
    """Vertex AI 프로젝트 ID와 리전을 환경 변수에서 읽어온다."""
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
        sys.exit(1)

    return project_id, region


def get_access_token() -> str:
    """Google Cloud ADC에서 액세스 토큰을 획득한다."""
    credentials, _ = google.auth.default(
        scopes=["https://www.googleapis.com/auth/cloud-platform"]
    )
    auth_request = google.auth.transport.requests.Request()
    credentials.refresh(auth_request)
    return credentials.token


def create_openai_client(project_id: str, region: str) -> OpenAI:
    """Vertex AI OpenAI 호환 엔드포인트용 OpenAI 클라이언트를 생성한다."""
    print(f"{Fore.CYAN}[Auth] Vertex AI OpenAI 호환 클라이언트 초기화 중...")
    print(f"{Fore.CYAN}  Project : {project_id}")
    print(f"{Fore.CYAN}  Region  : {region}")

    token = get_access_token()
    base_url = (
        f"https://{region}-aiplatform.googleapis.com/v1beta1/"
        f"projects/{project_id}/locations/{region}/endpoints/openapi"
    )

    client = OpenAI(
        base_url=base_url,
        api_key=token,  # Bearer token as api_key
    )
    print(f"{Fore.GREEN}[Auth] 클라이언트 초기화 완료")
    return client


def generate_output_path(results_dir: Path, run_start: datetime) -> Path:
    """타임스탬프 기반 고유 CSV 경로를 생성한다."""
    base = f"gemini_audit_{run_start.strftime('%y%m%d_%H%M')}"
    candidate = results_dir / f"{base}.csv"
    if not candidate.exists():
        return candidate
    seq = 2
    while True:
        candidate = results_dir / f"{base}_{seq}.csv"
        if not candidate.exists():
            return candidate
        seq += 1


def audit_chaincode_gemini(
    client: OpenAI,
    model: str,
    code: str,
    filename: str,
    project_id: str,
    region: str,
) -> str:
    """Gemini API를 사용하여 체인코드 보안 감사를 수행한다. 최대 3회 재시도."""
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
                    {"role": "system", "content": SYSTEM_PROMPT},
                    {"role": "user", "content": user_prompt},
                ],
            )
            return response.choices[0].message.content

        except Exception as e:
            last_error = e
            if attempt < MAX_RETRIES:
                delay = RETRY_BASE_DELAY ** attempt
                print(
                    f"{Fore.YELLOW}[Retry] {model} / {filename} — "
                    f"attempt {attempt}/{MAX_RETRIES} failed: {e}"
                )
                print(f"{Fore.YELLOW}[Retry] {delay}s 후 재시도...")
                time.sleep(delay)

                # 토큰 만료 가능성 대비 — 클라이언트 재생성
                if "401" in str(e) or "403" in str(e) or "Unauthorized" in str(e):
                    print(f"{Fore.YELLOW}[Retry] 토큰 갱신 중...")
                    token = get_access_token()
                    client.api_key = token
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
        sys.exit(1)

    print(f"{Fore.GREEN}[Info] 발견된 체인코드 파일: {len(go_files)}개")

    # ── Vertex AI 클라이언트 생성 ────────────────────────────────────
    client = create_openai_client(project_id, region)

    # ── 타임스탬프 기반 고유 파일명 생성 ─────────────────────────────
    output_csv = generate_output_path(RESULTS_DIR, run_start)
    print(f"{Fore.GREEN}[Info] 결과 저장 경로: {output_csv.name}")
    print(f"{Fore.GREEN}[Info] 모델: {len(GEMINI_MODELS)}개 - {', '.join(GEMINI_MODELS)}")
    print(f"{Fore.GREEN}[Info] 총 감사 예정: {len(GEMINI_MODELS) * len(go_files)}건")

    # ── CSV 컬럼 정의 ────────────────────────────────────────────────
    csv_fields = ["timestamp", "model", "prompt_strategy", "file", "code_chars", "elapsed_s", "result"]

    total_records = 0
    error_count = 0
    timing_summary = {}

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for model_name in GEMINI_MODELS:
            print(f"\n{Fore.YELLOW}{'='*60}")
            print(f"{Fore.YELLOW}[Audit] 모델: {model_name}")
            print(f"{Fore.YELLOW}{'='*60}\n")

            model_timings = []

            for go_file in tqdm(go_files, desc=f"Auditing with {model_name}"):
                code = go_file.read_text(encoding="utf-8")
                code_chars = len(code)
                print(f"{Fore.CYAN}[Audit] {go_file.name} ({code_chars} chars)")

                start = time.time()
                result = audit_chaincode_gemini(
                    client, model_name, code, go_file.name,
                    project_id, region,
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

    # ── 메타데이터 JSON 생성 ─────────────────────────────────────────
    meta_path = output_csv.with_suffix(".meta.json")
    meta = {
        "script_version": "1.0",
        "script": "06_run_gemini_audit.py",
        "backend": "Vertex AI (OpenAI Compatible API)",
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
        "models": GEMINI_MODELS,
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
    run_id_source = f"{run_start.isoformat()}|gemini|{','.join(GEMINI_MODELS)}"
    run_id = hashlib.sha256(run_id_source.encode()).hexdigest()[:8]
    meta["run_id"] = run_id

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    # ── gemini_audit_latest.csv 편의 복사본 ──────────────────────────
    latest_csv = RESULTS_DIR / "gemini_audit_latest.csv"
    shutil.copy2(output_csv, latest_csv)

    # ── 완료 보고 ────────────────────────────────────────────────────
    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] Gemini 감사 완료")
    print(f"{Fore.GREEN}  결과 CSV  : {output_csv}")
    print(f"{Fore.GREEN}  메타데이터: {meta_path}")
    print(f"{Fore.GREEN}  최신 복사 : {latest_csv}")
    print(f"{Fore.GREEN}  총 레코드 : {total_records}건")
    if error_count > 0:
        print(f"{Fore.RED}  오류 건수 : {error_count}건")
    print(f"{Fore.GREEN}  소요 시간 : {meta['run_duration_s']}s")
    print(f"{Fore.GREEN}  Run ID    : {run_id}")
    print(f"{Fore.GREEN}{'='*60}")

    for model, info in meta["timing_summary"].items():
        print(f"{Fore.CYAN}  [{model}] total={info['total_s']}s, avg={info['avg_s']}s/file")


if __name__ == "__main__":
    main()
