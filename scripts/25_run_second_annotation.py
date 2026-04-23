"""
25_run_second_annotation.py — Phase 3: Blinded 2nd Annotation
Different model (Sonnet) + different prompt framing + blinded (no 1st labels).
Reads BENCHMARK_FREEZE.json, labels each file, saves to separate run directory.
"""

import argparse, csv, hashlib, json, os, re, subprocess, sys, time
from datetime import datetime
from pathlib import Path
from anthropic import AnthropicVertex
from colorama import Fore, init

init(autoreset=True)

PROJECT_ROOT = Path(__file__).resolve().parent.parent
BENCHMARK_DIR = PROJECT_ROOT / "02_resources" / "golisa_benchmark" / "Benchmark"
LABELING_DIR = PROJECT_ROOT / "06_addon_validation" / "labeling"
BENCHMARK_FREEZE = PROJECT_ROOT / "06_addon_validation" / "benchmark" / "BENCHMARK_FREEZE.json"
STRIPPER = PROJECT_ROOT / "scripts" / "strip_go_comments.exe"

MODEL = "claude-opus-4-5@20251101"
TEMPERATURE = 0.0
MAX_TOKENS = 2048
MAX_RETRIES = 3
RETRY_BASE_DELAY = 2

DEV_REPOS = {
    "bluezd", "xuehuiit", "nitesh7sid", "RAntonio09",
    "cactusfluo", "RakhiSoni", "lutianYan", "joseprados",
    "ewerter", "pankajcheema"
}

SYSTEM_PROMPT_2ND = """You are an independent reviewer auditing Hyperledger Fabric Go chaincode files for consensus-critical nondeterminism. Your task is to determine whether each file contains patterns that could cause endorsement divergence.

Apply the following taxonomy EXACTLY:

### Core Vulnerability Classes (determine VULNERABLE vs SAFE)
- **C1 TIME_NOW**: time.Now(), time.Since(), time.Until() that influences a ledger write (PutState, DelState, PutPrivateData, DelPrivateData) or proposal response (shim.Success/Error, SetEvent, return value). GetTxTimestamp() is SAFE.
- **C2 GOROUTINE**: go func() or go methodCall() whose result influences a ledger write or response. Channel receives (<-) are out of scope.
- **C3 MAP_ITERATION**: Any for...range over a map-typed expression whose iteration order influences a ledger write or response. encoding/json.Marshal(map) is SAFE (deterministic key sorting in Go std lib v1).
- **C4 NON_REVALIDATED_QUERY**: GetQueryResult, GetPrivateDataQueryResult, GetQueryResultWithPagination, or GetHistoryForKey whose result determines a write decision. Read-only queries returned via shim.Success are NOT C4. GetStateByPartialCompositeKey and GetStateByRange are NOT C4 sources.
- **C6 GLOBAL_MUTABLE_STATE**: Package-level var declared in THIS file, read in transaction logic AND influencing a write or response. Logger-only globals and read-only/deterministic-init globals are SAFE.

### Auxiliary Label (does NOT affect VULNERABLE/SAFE verdict)
- **C5 ITERATOR_LEAK**: Iterator created but Close() not called on all paths.

### Decision Process
1. Does the file contain a recognizable HLF transaction entrypoint (shim Init/Invoke OR contractapi public method)? If NO → EXCLUDE.
2. Does any C1–C4 or C6 source pattern exist? If NO → SAFE.
3. Is there concrete intra-file evidence that the source reaches a sink? If NO → SAFE.
4. Is the usage only in logging, test, or dead code? If YES → SAFE. Otherwise → VULNERABLE.

### Scope
- Strictly intra-file evidence only. Do NOT reason about cross-file calls.
- Comments have been stripped. Do not speculate about removed comments.
- C5 alone does NOT make a file VULNERABLE.
- If multiple transaction functions exist and one has an issue → VULNERABLE (file-level).

### Output (MANDATORY format)
VERDICT: VULNERABLE / SAFE / EXCLUDE
PRIMARY_CLASS: C1 / C2 / C3 / C4 / C6 / NONE
SECONDARY_CLASS: C1 / C2 / C3 / C4 / C6 / NONE
AUXILIARY_C5: YES / NO
EVIDENCE_LINES: L[start]-L[end], ...
RATIONALE: [1-3 sentences explaining the decision]"""


def resolve_vertex_config():
    project_id = os.environ.get("GOOGLE_CLOUD_PROJECT") or os.environ.get("ANTHROPIC_VERTEX_PROJECT_ID")
    region = os.environ.get("CLOUD_ML_REGION", "us-east5")
    if not project_id:
        print(f"{Fore.RED}[Error] GOOGLE_CLOUD_PROJECT required")
        sys.exit(1)
    return project_id, region


def strip_comments(filepath: str):
    import shutil, tempfile
    if not STRIPPER.exists():
        print(f"{Fore.RED}[Error] {STRIPPER} not found")
        sys.exit(1)
    with tempfile.TemporaryDirectory() as tmpdir:
        inp = Path(tmpdir) / "input"
        out = Path(tmpdir) / "output"
        inp.mkdir(); out.mkdir()
        shutil.copy2(filepath, inp / Path(filepath).name)
        result = subprocess.run([str(STRIPPER), str(inp), str(out)], capture_output=True, text=True, encoding="utf-8")
        if result.returncode != 0:
            return None
        out_file = out / Path(filepath).name
        return out_file.read_text(encoding="utf-8") if out_file.exists() else None


def parse_response(text):
    fields = {}
    for key in ["VERDICT", "PRIMARY_CLASS", "SECONDARY_CLASS", "AUXILIARY_C5", "EVIDENCE_LINES", "RATIONALE"]:
        match = re.search(rf"{key}:\s*(.+?)(?:\n|$)", text)
        fields[key] = match.group(1).strip() if match else ""
    return fields


def label_file(client, code):
    user_prompt = "Audit this comment-stripped Hyperledger Fabric chaincode for consensus-critical nondeterminism:\n\n```go\n" + code + "\n```"
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            response = client.messages.create(
                model=MODEL, max_tokens=MAX_TOKENS, temperature=TEMPERATURE,
                system=SYSTEM_PROMPT_2ND,
                messages=[{"role": "user", "content": user_prompt}],
            )
            text = "".join(b.text for b in response.content if hasattr(b, "text"))
            usage = {"input_tokens": response.usage.input_tokens, "output_tokens": response.usage.output_tokens}
            return text, usage
        except Exception as e:
            if attempt < MAX_RETRIES:
                time.sleep(RETRY_BASE_DELAY ** attempt)
                print(f"{Fore.YELLOW}  [Retry {attempt}] {e}")
            else:
                return f"ERROR: {e}", {}


def main():
    parser = argparse.ArgumentParser(description="Phase 3: 2nd Annotation (Blinded)")
    parser.add_argument("--smoke", type=int, default=0)
    parser.add_argument("--resume", action="store_true")
    parser.add_argument("--run-dir", type=str, default="")
    args = parser.parse_args()

    run_start = datetime.now()
    timestamp = run_start.strftime("%y%m%d_%H%M")

    project_id, region = resolve_vertex_config()
    print(f"{Fore.CYAN}[2nd Annotation] Model: {MODEL}, Project: {project_id}")
    client = AnthropicVertex(region=region, project_id=project_id)
    print(f"{Fore.GREEN}[Auth] OK")

    with open(BENCHMARK_FREEZE, encoding="utf-8") as f:
        freeze = json.load(f)
    bench_files = freeze["files"]
    print(f"{Fore.GREEN}[Benchmark] {len(bench_files)} files ({freeze['vulnerable']}V + {freeze['safe']}S)")

    if args.resume and args.run_dir:
        run_dir = LABELING_DIR / args.run_dir
    elif args.resume:
        dirs = sorted(LABELING_DIR.glob("second_*"))
        if not dirs:
            print(f"{Fore.RED}[Error] No second_* directory found")
            sys.exit(1)
        run_dir = dirs[-1]
    else:
        run_dir = LABELING_DIR / f"second_{timestamp}"

    run_dir.mkdir(parents=True, exist_ok=True)
    per_file_dir = run_dir / "per_file"
    per_file_dir.mkdir(exist_ok=True)
    progress_path = run_dir / "progress.json"
    csv_path = run_dir / "summary.csv"

    print(f"{Fore.CYAN}[Output] {run_dir.name}/")

    completed_files = set()
    if args.resume and progress_path.exists():
        progress = json.load(open(progress_path, encoding="utf-8"))
        completed_files = set(progress.get("completed_files", []))
        if not completed_files:
            for jf in per_file_dir.glob("*.json"):
                data = json.load(open(jf, encoding="utf-8"))
                r, fn = data.get("repo", ""), data.get("filename", "")
                if r and fn:
                    completed_files.add(f"{r}/{fn}")
        print(f"{Fore.YELLOW}[Resume] {len(completed_files)} completed, skipping")

    candidates = []
    for bf in bench_files:
        filepath = BENCHMARK_DIR / bf["repo"] / bf["filename"]
        if not filepath.exists():
            continue
        candidates.append({"repo": bf["repo"], "filename": bf["filename"], "filepath": str(filepath)})

    if args.smoke > 0:
        candidates = candidates[:args.smoke]

    pending = [c for c in candidates if f"{c['repo']}/{c['filename']}" not in completed_files]
    print(f"{Fore.GREEN}[Run] {len(pending)} files to label\n")

    csv_fields = ["idx", "repo", "filename", "elapsed_s", "input_tokens", "output_tokens",
                  "verdict", "primary_class", "secondary_class", "auxiliary_c5", "evidence_lines", "rationale"]
    csv_mode = "a" if args.resume and csv_path.exists() else "w"
    csvfile = open(csv_path, csv_mode, newline="", encoding="utf-8")
    writer = csv.DictWriter(csvfile, fieldnames=csv_fields)
    if csv_mode == "w":
        writer.writeheader()

    error_count = 0
    results = []

    for idx, cand in enumerate(pending):
        repo, filename, filepath = cand["repo"], cand["filename"], cand["filepath"]
        file_key = f"{repo}/{filename}"
        print(f"[{idx+1}/{len(pending)}] {file_key} ", end="", flush=True)

        stripped = strip_comments(filepath)
        if stripped is None:
            raw = Path(filepath).read_text(encoding="utf-8", errors="ignore")
            stripped = raw

        t0 = time.time()
        raw_resp, usage = label_file(client, stripped)
        elapsed = time.time() - t0
        parsed = parse_response(raw_resp)

        verdict = parsed.get("VERDICT", "").upper().rstrip("*")
        if verdict not in ("SAFE", "VULNERABLE", "EXCLUDE"):
            verdict = "SAFE"
            error_count += 1

        print(f"{verdict} [{parsed.get('PRIMARY_CLASS','')}] ({elapsed:.1f}s)")

        file_id = f"{idx+1+len(completed_files):04d}_{repo}_{filename.replace('.go','')}"
        detail = {
            "file_id": file_id, "repo": repo, "filename": filename,
            "elapsed_s": round(elapsed, 2),
            "input_tokens": usage.get("input_tokens", 0),
            "output_tokens": usage.get("output_tokens", 0),
            "verdict": verdict,
            "primary_class": parsed.get("PRIMARY_CLASS", ""),
            "secondary_class": parsed.get("SECONDARY_CLASS", ""),
            "auxiliary_c5": parsed.get("AUXILIARY_C5", ""),
            "evidence_lines": parsed.get("EVIDENCE_LINES", ""),
            "rationale": parsed.get("RATIONALE", ""),
            "raw_response": raw_resp,
            "model": MODEL, "temperature": TEMPERATURE,
            "timestamp": datetime.now().isoformat(),
            "annotator": "2nd",
        }
        with open(per_file_dir / f"{file_id}.json", "w", encoding="utf-8") as f:
            json.dump(detail, f, indent=2, ensure_ascii=False)

        row = {field: detail.get(field, "") for field in csv_fields}
        row["idx"] = idx + len(completed_files)
        writer.writerow(row)
        csvfile.flush()
        results.append(row)

        completed_files.add(file_key)
        with open(progress_path, "w", encoding="utf-8") as f:
            json.dump({
                "completed_files": sorted(completed_files),
                "total_done": len(completed_files),
                "total_target": len(candidates),
                "last_updated": datetime.now().isoformat(),
                "errors": error_count,
            }, f, indent=2)

    csvfile.close()

    verdicts = {}
    classes = {}
    for r in results:
        v = r["verdict"]
        verdicts[v] = verdicts.get(v, 0) + 1
        pc = r["primary_class"]
        if pc and pc != "NONE":
            classes[pc] = classes.get(pc, 0) + 1

    print(f"\n{'='*60}")
    print(f"[Done] 2nd annotation complete")
    print(f"  Labeled: {len(results)}")
    print(f"  Verdicts: {verdicts}")
    print(f"  Classes: {classes}")
    print(f"  Errors: {error_count}")
    print(f"  Output: {run_dir.name}/")


if __name__ == "__main__":
    main()
