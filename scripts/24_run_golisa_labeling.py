"""
24_run_golisa_labeling.py  (v2.0 — 2026-04-22)
GoLiSA benchmark 1차 라벨링 — Claude Opus 4.5 via Vertex AI.

사용법:
  python 24_run_golisa_labeling.py                # 전체 실행
  python 24_run_golisa_labeling.py --smoke 3      # 스모크 테스트 (3개만)
  python 24_run_golisa_labeling.py --resume       # 이전 실행 이어서

출력 구조:
  06_addon_validation/labeling/run_YYMMDD_HHMM/
  ├── summary.csv                  # 전체 결과 요약 (1행 = 1파일)
  ├── summary.meta.json            # 실행 메타데이터
  ├── per_file/                    # 개별 파일별 상세 결과
  │   ├── 0001_reponame_filename.json
  │   ├── 0002_reponame_filename.json
  │   └── ...
  └── progress.json                # 재시작용 진행 상태
"""

import argparse
import csv
import hashlib
import json
import os
import re
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

from anthropic import AnthropicVertex
from colorama import Fore, init

init(autoreset=True)

# ── 경로 설정 ────────────────────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
MINING_DIR = PROJECT_ROOT / "06_addon_validation" / "golisa_mining"
LABELING_DIR = PROJECT_ROOT / "06_addon_validation" / "labeling"
STRIPPER = PROJECT_ROOT / "scripts" / "strip_go_comments.exe"

POSITIVE_FILE = MINING_DIR / "candidates_260422_1747.json"
SAFE_FILE = MINING_DIR / "safe_candidates_260422_1747.json"
HARD_NEG_FILE = MINING_DIR / "hard_negative_candidates_260422_1747.json"
BENCHMARK_DIR = PROJECT_ROOT / "02_resources" / "golisa_benchmark" / "Benchmark"

# ── 모델 설정 ────────────────────────────────────────────────────────
MODEL = "claude-opus-4-5@20251101"
TEMPERATURE = 0.0
MAX_TOKENS = 2048
MAX_RETRIES = 3
RETRY_BASE_DELAY = 2

# ── Dev exclusion repos ──────────────────────────────────────────────
DEV_REPOS = {
    "bluezd", "xuehuiit", "nitesh7sid", "RAntonio09",
    "cactusfluo", "RakhiSoni", "lutianYan", "joseprados",
    "ewerter", "pankajcheema"
}

# ── 라벨링 system prompt (rubric v2.0) ───────────────────────────────
SYSTEM_PROMPT = """You are a Hyperledger Fabric chaincode security analyst performing file-level labeling for consensus-critical nondeterminism. Follow the rubric below EXACTLY.

## Taxonomy (5 Core Classes + 1 Auxiliary)

Core Classes (determine V/S):
| ID | Class | Source Pattern | Sink Requirement |
|C1|TIME_NOW|time.Now(), time.Since(), time.Until()|Must influence ledger write or proposal response|
|C2|GOROUTINE|go func(), go methodCall() (goroutine-only; channel <- is out of scope)|Must influence ledger write or response|
|C3|MAP_ITERATION|Any for...range over map-typed expression|Must influence ledger write or response|
|C4|NON_REVALIDATED_QUERY|GetQueryResult, GetPrivateDataQueryResult, GetQueryResultWithPagination, GetHistoryForKey|Query result must determine write decision|
|C6|GLOBAL_MUTABLE_STATE|var globalVar = ... at package level|Must be read in tx logic AND influence write/response|

Auxiliary (does NOT determine V/S):
|C5|ITERATOR_LEAK|Iterator created but Close() not called on all paths|

## Sinks
Ledger: PutState, DelState, PutPrivateData, DelPrivateData, PurgePrivateData, SetStateValidationParameter, SetPrivateDataValidationParameter
Response: shim.Success/Error, SetEvent, contractapi transaction function return value/error
Implicit-flow: tainted guard determining whether a sink executes counts as source->sink
API equivalence: stub.X(...) = ctx.GetStub().X(...) for all sinks

## Out of Scope (IGNORE these patterns)
math/rand, crypto/rand, system/environment APIs, file/IO APIs, external process APIs, channel receives (<-), InvokeChaincode, access control, input validation, private data leakage, key management, generic code quality

## Decision Tree
1. HLF transaction entrypoint? (shim Init/Invoke OR contractapi public method with TransactionContextInterface/*TransactionContext/custom context) -> NO=EXCLUDE
2. C1-C4 or C6 source pattern? -> NO=SAFE
3. Concrete intra-file evidence source reaches sink? -> NO=SAFE
4. Usage in logging/test/dead code only? -> YES=SAFE -> NO=VULNERABLE

## Scope Rule
STRICTLY intra-file evidence only. Do NOT chase calls into other files/packages/vendor code.

## SAFE Exceptions
- GetTxTimestamp() -> SAFE for C1
- encoding/json.Marshal(map) -> SAFE for C3 (Go std lib v1 sorts keys). Does NOT extend to v2 or other serializers.
- yaml.Marshal(map), fmt.Sprintf(map), custom serializer -> SAFE only if deterministic ordering documented
- Map range sorted before write -> SAFE for C3
- Iterator Close() in defer -> SAFE for C5. Flag only if no reachable Close().
- GetStateByRange, GetStateByPartialCompositeKey alone -> SAFE for C4
- Global var read-only, deterministic init, no later mutation -> SAFE for C6
- Package-level logger for diagnostics only -> SAFE for C6
- init() nondeterministic API -> V only if value reaches in-file tx sink

## NOT-SAFE
- GetHistoryForKey deciding writes -> C4
- time.Now() reaching PutState via any intra-file path -> C1
- go func() result influencing ledger write -> C2

## Important Rules
- Concrete evidence required. Mere co-location insufficient.
- Comments are stripped. Do not speculate about removed comments.
- No cross-file reasoning. C5 alone does NOT make VULNERABLE.
- Multiple tx functions, one has issue -> V (file-level).
- *_test.go or package xxx_test -> EXCLUDE. Partial code -> EXCLUDE.

## Output Format (MANDATORY)
VERDICT: VULNERABLE / SAFE / EXCLUDE
PRIMARY_CLASS: C1 / C2 / C3 / C4 / C6 / NONE
SECONDARY_CLASS: C1 / C2 / C3 / C4 / C6 / NONE
AUXILIARY_C5: YES / NO
EVIDENCE_LINES: L[start]-L[end], L[start]-L[end]
RATIONALE: [1-3 sentences]"""


def resolve_vertex_config():
    project_id = os.environ.get("GOOGLE_CLOUD_PROJECT") or os.environ.get(
        "ANTHROPIC_VERTEX_PROJECT_ID"
    )
    region = os.environ.get("CLOUD_ML_REGION", "us-east5")
    if not project_id:
        print(f"{Fore.RED}[Error] GOOGLE_CLOUD_PROJECT 또는 ANTHROPIC_VERTEX_PROJECT_ID 환경변수 필요")
        sys.exit(1)
    return project_id, region


def strip_comments(filepath: str) -> str | None:
    """strip_go_comments.exe는 dir→dir 방식. 임시 디렉터리로 단일 파일 처리."""
    import shutil, tempfile
    if not STRIPPER.exists():
        print(f"{Fore.RED}[Error] {STRIPPER} not found")
        sys.exit(1)
    with tempfile.TemporaryDirectory() as tmpdir:
        inp = Path(tmpdir) / "input"
        out = Path(tmpdir) / "output"
        inp.mkdir()
        out.mkdir()
        shutil.copy2(filepath, inp / Path(filepath).name)
        result = subprocess.run(
            [str(STRIPPER), str(inp), str(out)],
            capture_output=True, text=True, encoding="utf-8"
        )
        if result.returncode != 0:
            return None
        out_file = out / Path(filepath).name
        if out_file.exists():
            return out_file.read_text(encoding="utf-8")
        return None


def parse_response(text: str) -> dict:
    fields = {}
    for key in ["VERDICT", "PRIMARY_CLASS", "SECONDARY_CLASS", "AUXILIARY_C5",
                 "EVIDENCE_LINES", "RATIONALE"]:
        match = re.search(rf"{key}:\s*(.+?)(?:\n|$)", text)
        fields[key] = match.group(1).strip() if match else ""
    return fields


def label_file(client, code: str) -> tuple[str, dict]:
    user_prompt = (
        "Analyze this comment-stripped Hyperledger Fabric chaincode "
        "for consensus-critical nondeterminism:\n\n```go\n" + code + "\n```"
    )
    last_error = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            response = client.messages.create(
                model=MODEL,
                max_tokens=MAX_TOKENS,
                temperature=TEMPERATURE,
                system=SYSTEM_PROMPT,
                messages=[{"role": "user", "content": user_prompt}],
            )
            text = "".join(b.text for b in response.content if hasattr(b, "text"))
            usage = {
                "input_tokens": response.usage.input_tokens,
                "output_tokens": response.usage.output_tokens,
            }
            return text, usage
        except Exception as e:
            last_error = e
            if attempt < MAX_RETRIES:
                delay = RETRY_BASE_DELAY ** attempt
                print(f"{Fore.YELLOW}  [Retry] attempt {attempt}/{MAX_RETRIES}: {e}")
                time.sleep(delay)
    return f"ERROR: {last_error}", {}


def get_repo_from_path(filepath: str) -> str:
    parts = filepath.replace("\\", "/").split("/")
    try:
        idx = parts.index("Benchmark")
        return parts[idx + 1]
    except (ValueError, IndexError):
        return ""


def load_candidates():
    positives = json.load(open(POSITIVE_FILE, encoding="utf-8"))
    safes = json.load(open(SAFE_FILE, encoding="utf-8"))
    hard_negs = json.load(open(HARD_NEG_FILE, encoding="utf-8"))
    hard_neg_paths = {c.get("filepath", "").replace("\\", "/") for c in hard_negs}

    all_candidates = []
    for c in positives:
        fp = c.get("filepath", "")
        repo = get_repo_from_path(fp)
        all_candidates.append({
            "filepath": fp,
            "repo": repo,
            "filename": Path(fp).name,
            "candidate_type": "positive",
            "is_hard_negative": fp.replace("\\", "/") in hard_neg_paths,
            "dev_excluded": repo in DEV_REPOS,
            "mining_families": c.get("families", []),
        })
    for c in safes:
        fp = c.get("filepath", "")
        repo = get_repo_from_path(fp)
        all_candidates.append({
            "filepath": fp,
            "repo": repo,
            "filename": Path(fp).name,
            "candidate_type": "safe",
            "is_hard_negative": fp.replace("\\", "/") in hard_neg_paths,
            "dev_excluded": repo in DEV_REPOS,
            "mining_families": c.get("families", []),
        })
    safe_fps = {c.get("filepath", "").replace("\\", "/") for c in safes}
    for c in hard_negs:
        fp = c.get("filepath", "")
        if fp.replace("\\", "/") in safe_fps:
            continue
        repo = get_repo_from_path(fp)
        all_candidates.append({
            "filepath": fp,
            "repo": repo,
            "filename": Path(fp).name,
            "candidate_type": "safe",
            "is_hard_negative": True,
            "dev_excluded": repo in DEV_REPOS,
            "mining_families": c.get("possible_families", []),
        })
    return all_candidates, len(positives), len(safes), len(hard_negs)


def load_full_corpus():
    positives = json.load(open(POSITIVE_FILE, encoding="utf-8"))
    safes = json.load(open(SAFE_FILE, encoding="utf-8"))
    hard_negs = json.load(open(HARD_NEG_FILE, encoding="utf-8"))
    hard_neg_paths = {c.get("filepath", "").replace("\\", "/") for c in hard_negs}
    pos_paths = {c.get("filepath", "").replace("\\", "/") for c in positives}
    safe_paths = {c.get("filepath", "").replace("\\", "/") for c in safes}

    all_candidates = []
    for repo_name in sorted(BENCHMARK_DIR.iterdir()):
        if not repo_name.is_dir():
            continue
        repo = repo_name.name
        for go_file in sorted(repo_name.glob("*.go")):
            fp = str(go_file)
            fp_norm = fp.replace("\\", "/")
            if fp_norm in pos_paths:
                ctype = "positive"
            elif fp_norm in safe_paths:
                ctype = "safe"
            elif fp_norm in hard_neg_paths:
                ctype = "hard_negative"
            else:
                ctype = "unlabeled"
            all_candidates.append({
                "filepath": fp,
                "repo": repo,
                "filename": go_file.name,
                "candidate_type": ctype,
                "is_hard_negative": fp_norm in hard_neg_paths,
                "dev_excluded": repo in DEV_REPOS,
                "mining_families": [],
            })
    n_total = len(all_candidates)
    n_pos = sum(1 for c in all_candidates if c["candidate_type"] == "positive")
    n_safe = sum(1 for c in all_candidates if c["candidate_type"] == "safe")
    n_hn = sum(1 for c in all_candidates if c["candidate_type"] == "hard_negative")
    n_new = sum(1 for c in all_candidates if c["candidate_type"] == "unlabeled")
    return all_candidates, n_pos, n_safe, n_hn, n_new


def main():
    parser = argparse.ArgumentParser(description="GoLiSA 1차 라벨링")
    parser.add_argument("--smoke", type=int, default=0,
                        help="스모크 테스트: N개만 실행 (0=전체)")
    parser.add_argument("--resume", action="store_true",
                        help="이전 실행 이어서 (progress.json 기반)")
    parser.add_argument("--run-dir", type=str, default="",
                        help="재시작할 run 디렉터리 이름 (--resume와 함께)")
    parser.add_argument("--full-corpus", action="store_true",
                        help="GoLiSA 전수 라벨링 (Benchmark/ 전체 스캔)")
    args = parser.parse_args()

    run_start = datetime.now()
    timestamp = run_start.strftime("%y%m%d_%H%M")

    # ── Vertex AI ────────────────────────────────────────────────────
    project_id, region = resolve_vertex_config()
    print(f"{Fore.CYAN}[Config] Project: {project_id}, Region: {region}, Model: {MODEL}")
    client = AnthropicVertex(region=region, project_id=project_id)
    print(f"{Fore.GREEN}[Auth] OK")

    # ── 후보 로드 ────────────────────────────────────────────────────
    if args.full_corpus:
        all_candidates, n_pos, n_safe, n_hard, n_new = load_full_corpus()
        active = [c for c in all_candidates if not c["dev_excluded"]]
        n_excluded = len(all_candidates) - len(active)
        print(f"{Fore.GREEN}[Full Corpus] total={len(all_candidates)}, pos={n_pos}, safe={n_safe}, hard_neg={n_hard}, new={n_new}")
        print(f"{Fore.GREEN}[Full Corpus] active={len(active)}, dev_excluded={n_excluded}")
    else:
        all_candidates, n_pos, n_safe, n_hard = load_candidates()
        active = [c for c in all_candidates if not c["dev_excluded"]]
        n_excluded = len(all_candidates) - len(active)
        print(f"{Fore.GREEN}[Pool] positive={n_pos}, safe={n_safe}, hard_neg={n_hard}")
        print(f"{Fore.GREEN}[Pool] active={len(active)}, dev_excluded={n_excluded}")

    if args.smoke > 0:
        active = active[:args.smoke]
        print(f"{Fore.YELLOW}[Smoke] 스모크 테스트: {args.smoke}개만 실행")

    # ── 출력 디렉터리 ────────────────────────────────────────────────
    if args.resume and args.run_dir:
        run_dir = LABELING_DIR / args.run_dir
    elif args.resume:
        dirs = sorted(LABELING_DIR.glob("run_*"))
        if not dirs:
            print(f"{Fore.RED}[Error] 재시작할 run 디렉터리 없음")
            sys.exit(1)
        run_dir = dirs[-1]
    else:
        prefix = "smoke" if args.smoke > 0 else "run"
        run_dir = LABELING_DIR / f"{prefix}_{timestamp}"

    run_dir.mkdir(parents=True, exist_ok=True)
    per_file_dir = run_dir / "per_file"
    per_file_dir.mkdir(exist_ok=True)

    progress_path = run_dir / "progress.json"
    csv_path = run_dir / "summary.csv"
    meta_path = run_dir / "summary.meta.json"

    print(f"{Fore.CYAN}[Output] {run_dir.name}/")

    # ── 재시작: 완료된 파일 건너뛰기 ─────────────────────────────────
    completed = set()
    completed_files = set()
    if args.resume and progress_path.exists():
        progress = json.load(open(progress_path, encoding="utf-8"))
        completed = set(progress.get("completed_repos", []))
        completed_files = set(progress.get("completed_files", []))
        if not completed_files:
            for jf in per_file_dir.glob("*.json"):
                data = json.load(open(jf, encoding="utf-8"))
                r, fn = data.get("repo", ""), data.get("filename", "")
                if r and fn:
                    completed_files.add(f"{r}/{fn}")
        print(f"{Fore.YELLOW}[Resume] {len(completed_files)}개 완료 건너뛰기")

    # ── CSV 준비 ─────────────────────────────────────────────────────
    csv_fields = [
        "idx", "repo", "filename", "candidate_type", "is_hard_negative",
        "code_chars", "stripped_chars", "elapsed_s", "input_tokens", "output_tokens",
        "verdict", "primary_class", "secondary_class", "auxiliary_c5",
        "evidence_lines", "rationale",
    ]

    csv_mode = "a" if args.resume and csv_path.exists() else "w"
    csvfile = open(csv_path, csv_mode, newline="", encoding="utf-8")
    writer = csv.DictWriter(csvfile, fieldnames=csv_fields)
    if csv_mode == "w":
        writer.writeheader()

    # ── 실행 ─────────────────────────────────────────────────────────
    results = []
    error_count = 0
    strip_fail = 0
    total_elapsed = 0
    total_input_tokens = 0
    total_output_tokens = 0

    pending = [c for c in active
               if f"{c['repo']}/{c['filename']}" not in completed_files]
    print(f"{Fore.GREEN}[Run] {len(pending)}개 라벨링 시작\n")

    for idx, cand in enumerate(pending):
        filepath = cand["filepath"]
        filename = cand["filename"]
        repo = cand["repo"]
        file_id = f"{idx+1+len(completed):04d}_{repo}_{filename.replace('.go','')}"

        print(f"[{idx+1}/{len(pending)}] {repo}/{filename} ", end="", flush=True)

        # 코드 읽기
        try:
            raw_code = Path(filepath).read_text(encoding="utf-8")
        except Exception as e:
            print(f"{Fore.RED}READ_ERROR: {e}")
            error_count += 1
            continue

        # 주석 제거
        stripped = strip_comments(filepath)
        if stripped is None:
            print(f"{Fore.YELLOW}(strip fail, using raw) ", end="")
            stripped = raw_code
            strip_fail += 1

        # API 호출
        t0 = time.time()
        raw_response, usage = label_file(client, stripped)
        elapsed = time.time() - t0
        total_elapsed += elapsed
        total_input_tokens += usage.get("input_tokens", 0)
        total_output_tokens += usage.get("output_tokens", 0)

        # 파싱
        if raw_response.startswith("ERROR:"):
            parsed = {k: "" for k in ["VERDICT", "PRIMARY_CLASS", "SECONDARY_CLASS",
                                       "AUXILIARY_C5", "EVIDENCE_LINES", "RATIONALE"]}
            parsed["VERDICT"] = "ERROR"
            error_count += 1
            print(f"{Fore.RED}ERROR ({elapsed:.1f}s)")
        else:
            parsed = parse_response(raw_response)
            v = parsed.get("VERDICT", "?")
            pc = parsed.get("PRIMARY_CLASS", "?")
            color = Fore.RED if "VUL" in v.upper() else (Fore.GREEN if "SAFE" in v.upper() else Fore.YELLOW)
            print(f"{color}{v} [{pc}] ({elapsed:.1f}s)")

        # 개별 파일 JSON 저장
        per_file_data = {
            "file_id": file_id,
            "repo": repo,
            "filename": filename,
            "filepath": filepath,
            "candidate_type": cand["candidate_type"],
            "is_hard_negative": cand["is_hard_negative"],
            "mining_families": cand["mining_families"],
            "code_chars": len(raw_code),
            "stripped_chars": len(stripped),
            "code_hash": hashlib.sha256(raw_code.encode()).hexdigest()[:16],
            "stripped_hash": hashlib.sha256(stripped.encode()).hexdigest()[:16],
            "elapsed_s": round(elapsed, 2),
            "input_tokens": usage.get("input_tokens", 0),
            "output_tokens": usage.get("output_tokens", 0),
            "verdict": parsed.get("VERDICT", ""),
            "primary_class": parsed.get("PRIMARY_CLASS", ""),
            "secondary_class": parsed.get("SECONDARY_CLASS", ""),
            "auxiliary_c5": parsed.get("AUXILIARY_C5", ""),
            "evidence_lines": parsed.get("EVIDENCE_LINES", ""),
            "rationale": parsed.get("RATIONALE", ""),
            "raw_response": raw_response,
            "model": MODEL,
            "temperature": TEMPERATURE,
            "timestamp": datetime.now().isoformat(),
        }
        per_file_path = per_file_dir / f"{file_id}.json"
        with open(per_file_path, "w", encoding="utf-8") as f:
            json.dump(per_file_data, f, indent=2, ensure_ascii=False)

        # CSV 행 저장
        row = {
            "idx": idx + len(completed),
            "repo": repo,
            "filename": filename,
            "candidate_type": cand["candidate_type"],
            "is_hard_negative": cand["is_hard_negative"],
            "code_chars": len(raw_code),
            "stripped_chars": len(stripped),
            "elapsed_s": round(elapsed, 2),
            "input_tokens": usage.get("input_tokens", 0),
            "output_tokens": usage.get("output_tokens", 0),
            "verdict": parsed.get("VERDICT", ""),
            "primary_class": parsed.get("PRIMARY_CLASS", ""),
            "secondary_class": parsed.get("SECONDARY_CLASS", ""),
            "auxiliary_c5": parsed.get("AUXILIARY_C5", ""),
            "evidence_lines": parsed.get("EVIDENCE_LINES", ""),
            "rationale": parsed.get("RATIONALE", ""),
        }
        writer.writerow(row)
        csvfile.flush()
        results.append(row)

        # progress 업데이트
        completed.add(repo)
        completed_files.add(f"{repo}/{filename}")
        with open(progress_path, "w", encoding="utf-8") as f:
            json.dump({
                "completed_repos": list(completed),
                "completed_files": sorted(completed_files),
                "total_done": len(completed_files),
                "total_target": len(active),
                "last_updated": datetime.now().isoformat(),
                "errors": error_count,
            }, f, indent=2)

    csvfile.close()

    # ── 통계 ─────────────────────────────────────────────────────────
    verdicts = {}
    classes = {}
    c5_count = 0
    for r in results:
        v = r["verdict"].upper()
        verdicts[v] = verdicts.get(v, 0) + 1
        pc = r["primary_class"].upper()
        if pc and pc != "NONE":
            classes[pc] = classes.get(pc, 0) + 1
        if r.get("auxiliary_c5", "").upper() == "YES":
            c5_count += 1

    print(f"\n{'='*60}")
    print(f"[Done] Labeling complete")
    print(f"  Total labeled: {len(results)}")
    print(f"  Verdicts: {verdicts}")
    print(f"  Primary classes: {classes}")
    print(f"  Auxiliary C5: {c5_count}")
    print(f"  Errors: {error_count}, Strip failures: {strip_fail}")
    print(f"  Time: {total_elapsed:.1f}s (avg {total_elapsed/max(len(results),1):.1f}s/file)")
    print(f"  Tokens: input={total_input_tokens:,}, output={total_output_tokens:,}")
    print(f"  Output: {run_dir.name}/")

    # ── 메타데이터 ───────────────────────────────────────────────────
    meta = {
        "script": "24_run_golisa_labeling.py",
        "version": "2.0",
        "run_dir": run_dir.name,
        "is_smoke": args.smoke > 0,
        "smoke_count": args.smoke,
        "timestamp_start": run_start.isoformat(),
        "timestamp_end": datetime.now().isoformat(),
        "duration_s": round(total_elapsed, 2),
        "model": MODEL,
        "temperature": TEMPERATURE,
        "max_tokens": MAX_TOKENS,
        "backend": "Vertex AI (AnthropicVertex)",
        "vertex_project_id": project_id,
        "vertex_region": region,
        "rubric_version": "GOLISA_TAXONOMY_RUBRIC_v2.md",
        "candidates": {"positive": n_pos, "safe": n_safe, "hard_negative": n_hard},
        "dev_excluded": n_excluded,
        "active": len(active),
        "labeled": len(results),
        "errors": error_count,
        "strip_failures": strip_fail,
        "tokens": {"input": total_input_tokens, "output": total_output_tokens},
        "verdicts": verdicts,
        "primary_classes": classes,
        "auxiliary_c5_count": c5_count,
    }
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2, ensure_ascii=False)


if __name__ == "__main__":
    main()
