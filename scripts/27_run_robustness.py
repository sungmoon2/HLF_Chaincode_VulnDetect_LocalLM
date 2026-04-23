"""
27_run_robustness.py — Phase 7: Robustness Run (v2.0)
Qwen2.5-Coder-7B (temp=0.1) x 5 seeds on frozen benchmark (464 files).
Follows INFERENCE_CONTRACT.md Section 1 (Phase 7 column).

v2.0 changes:
  - Raw LLM responses saved per file per seed (per_file/ JSONs)
  - Incremental CSV with flush after every row
  - Per-seed elapsed time + finish_reason + token counts
  - primary_class from BENCHMARK_FREEZE tracked
  - Running_Examples subdirectory path resolution
  - Functional --resume (finds latest existing run directory)
  - meta.json written at start, updated at end
  - BENCHMARK_FREEZE SHA-256 integrity check
  - Classifier identical to Phase 6 (26_run_main_experiment.py)
"""

import argparse, csv, hashlib, json, os, re, subprocess, sys, time
from datetime import datetime
from pathlib import Path
from collections import Counter
from colorama import Fore, init

init(autoreset=True)

PROJECT_ROOT = Path(__file__).resolve().parent.parent
BENCHMARK_DIR = PROJECT_ROOT / "02_resources" / "golisa_benchmark" / "Benchmark"
BENCHMARK_FREEZE = PROJECT_ROOT / "06_addon_validation" / "benchmark" / "BENCHMARK_FREEZE.json"
RESULTS_DIR = PROJECT_ROOT / "06_addon_validation" / "experiment"
MODEL_PATH = PROJECT_ROOT / "02_resources" / "models" / "qwen2.5-coder-7b-instruct-q4_k_m.gguf"
STRIPPER = PROJECT_ROOT / "scripts" / "strip_go_comments.exe"

SEEDS = [1, 2, 3, 4, 5]
TEMPERATURE = 0.1

INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 16384,
    "max_tokens": 2048,
    "temperature": TEMPERATURE,
}

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


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    h.update(path.read_bytes())
    return h.hexdigest()


def strip_comments(filepath: str):
    import shutil, tempfile
    with tempfile.TemporaryDirectory() as tmpdir:
        inp = Path(tmpdir) / "input"
        out = Path(tmpdir) / "output"
        inp.mkdir(); out.mkdir()
        shutil.copy2(filepath, inp / Path(filepath).name)
        result = subprocess.run(
            [str(STRIPPER), str(inp), str(out)],
            capture_output=True, text=True, encoding="utf-8",
        )
        if result.returncode != 0:
            return None
        out_file = out / Path(filepath).name
        return out_file.read_text(encoding="utf-8") if out_file.exists() else None


def classify_response(response: str) -> str:
    """Classifier identical to Phase 6 (26_run_main_experiment.py lines 60-89)."""
    if not response or response.startswith("ERROR:"):
        return "error"
    resp_lower = response.lower()
    safe_indicators = [
        "no vulnerabilities detected", "no vulnerabilities found",
        "no security vulnerabilities", "no significant vulnerabilities",
        "no vulnerabilities were found", "no vulnerabilities were detected",
        "no critical vulnerabilities", "the code appears to be secure",
        "the code is secure", "no issues found", "no issues detected",
    ]
    for indicator in safe_indicators:
        if indicator in resp_lower:
            idx = resp_lower.find(indicator)
            after = resp_lower[idx + len(indicator):]
            if any(kw in after[:200] for kw in ["however", "but ", "although", "note that"]):
                vuln_check = ["vulnerability", "vulnerable", "severity:", "recommended fix"]
                if sum(1 for v in vuln_check if v in after) >= 2:
                    return "vulnerable"
            return "safe"
    vuln_indicators = [
        "vulnerability type", "severity:", "severity :",
        "recommended fix", "affected code", "critical", "high",
        "non-deterministic", "nondeterministic", "phantom read",
        "global variable", "goroutine", "race condition",
        "map iteration", "iterator leak", "putstate",
    ]
    if sum(1 for ind in vuln_indicators if ind in resp_lower) >= 2:
        return "vulnerable"
    return "safe"


def resolve_filepath(repo: str, filename: str) -> Path | None:
    direct = BENCHMARK_DIR / repo / filename
    if direct.exists():
        return direct
    repo_dir = BENCHMARK_DIR / repo
    if repo_dir.exists():
        for go_file in repo_dir.rglob(filename):
            return go_file
    return None


def find_latest_run_dir(prefix: str = "robustness_") -> Path | None:
    if not RESULTS_DIR.exists():
        return None
    dirs = sorted(
        [d for d in RESULTS_DIR.iterdir() if d.is_dir() and d.name.startswith(prefix)],
        key=lambda d: d.stat().st_mtime,
        reverse=True,
    )
    return dirs[0] if dirs else None


def main():
    parser = argparse.ArgumentParser(description="Phase 7: Robustness Run (v2.0)")
    parser.add_argument("--smoke", type=int, default=0, help="Smoke test: N files only")
    parser.add_argument("--seeds", type=str, default="1,2,3,4,5", help="Comma-separated seeds")
    parser.add_argument("--resume", action="store_true", help="Resume latest robustness run")
    parser.add_argument("--run-dir", type=str, default=None, help="Explicit run directory to resume")
    args = parser.parse_args()

    seeds = [int(s) for s in args.seeds.split(",")]
    run_start = datetime.now()
    timestamp = run_start.strftime("%y%m%d_%H%M")

    # --- Run directory resolution ---
    if args.resume:
        if args.run_dir:
            run_dir = Path(args.run_dir)
        else:
            run_dir = find_latest_run_dir()
        if run_dir is None or not run_dir.exists():
            print(f"{Fore.RED}[Error] No previous run directory found to resume.")
            sys.exit(1)
        print(f"{Fore.YELLOW}[Resume] Using {run_dir.name}")
    else:
        run_dir = RESULTS_DIR / f"robustness_{timestamp}"
        run_dir.mkdir(parents=True, exist_ok=True)

    per_file_dir = run_dir / "per_file"
    per_file_dir.mkdir(exist_ok=True)

    # --- BENCHMARK_FREEZE integrity ---
    freeze_hash = sha256_file(BENCHMARK_FREEZE)
    with open(BENCHMARK_FREEZE, encoding="utf-8") as f:
        freeze = json.load(f)
    bench_files = freeze["files"]
    print(f"{Fore.CYAN}[Benchmark] {len(bench_files)} files ({freeze['vulnerable']}V + {freeze['safe']}S)")
    print(f"{Fore.CYAN}[Freeze SHA-256] {freeze_hash[:16]}...")
    print(f"{Fore.CYAN}[Config] temp={TEMPERATURE}, seeds={seeds}, runs={len(seeds)}")

    # --- Write meta.json at start ---
    meta_path = run_dir / "meta.json"
    meta = {
        "version": "2.0",
        "phase": "Phase 7 Robustness",
        "run_start": run_start.isoformat(),
        "run_end": None,
        "temperature": TEMPERATURE,
        "seeds": seeds,
        "model": MODEL_PATH.name,
        "inference_params": INFERENCE_PARAMS,
        "benchmark_freeze_sha256": freeze_hash,
        "benchmark_total": len(bench_files),
        "benchmark_vulnerable": freeze["vulnerable"],
        "benchmark_safe": freeze["safe"],
        "classifier_note": "Identical to Phase 6 (26_run_main_experiment.py). Superset of INFERENCE_CONTRACT simplified spec.",
        "total_files": 0,
        "avg_stability": None,
        "perfect_stability": None,
        "total_inference_s": None,
        "errors": 0,
    }
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2, ensure_ascii=False)

    # --- Load model ---
    from llama_cpp import Llama
    print(f"{Fore.CYAN}[Model] Loading {MODEL_PATH.name}...")
    llm = Llama(
        model_path=str(MODEL_PATH),
        n_gpu_layers=INFERENCE_PARAMS["n_gpu_layers"],
        n_ctx=INFERENCE_PARAMS["n_ctx"],
        verbose=False,
    )
    print(f"{Fore.GREEN}[Model] Loaded")

    # --- Build candidate list with path resolution ---
    candidates = []
    path_issues = []
    for bf in bench_files:
        resolved = resolve_filepath(bf["repo"], bf["filename"])
        if resolved:
            candidates.append({
                "repo": bf["repo"],
                "filename": bf["filename"],
                "filepath": str(resolved),
                "ground_truth": bf["verdict"],
                "primary_class": bf.get("primary_class", ""),
                "content_hash": bf.get("content_hash", ""),
            })
        else:
            path_issues.append(f"{bf['repo']}/{bf['filename']}")

    if path_issues:
        print(f"{Fore.RED}[Warning] {len(path_issues)} file(s) not found:")
        for p in path_issues:
            print(f"  - {p}")

    if args.smoke > 0:
        candidates = candidates[:args.smoke]
        print(f"{Fore.YELLOW}[Smoke] {args.smoke} files only")

    # --- Load existing progress for resume ---
    progress_path = run_dir / "progress.json"
    completed_keys = set()
    if args.resume and progress_path.exists():
        with open(progress_path, encoding="utf-8") as f:
            prog = json.load(f)
        completed_keys = set(prog.get("completed_files", []))
        print(f"{Fore.YELLOW}[Resume] {len(completed_keys)} already completed")

    # --- CSV setup (incremental) ---
    csv_fields = [
        "idx", "repo", "filename", "ground_truth", "primary_class",
    ] + [f"seed_{s}" for s in seeds] + [
        "stability", "majority_vote", "avg_elapsed_s", "total_elapsed_s",
        "any_truncated",
    ]
    csv_path = run_dir / "results.csv"

    if args.resume and csv_path.exists() and completed_keys:
        csv_mode = "a"
    else:
        csv_mode = "w"

    csvfile = open(csv_path, csv_mode, newline="", encoding="utf-8")
    writer = csv.DictWriter(csvfile, fieldnames=csv_fields)
    if csv_mode == "w":
        writer.writeheader()
        csvfile.flush()

    # --- Log file ---
    log_path = run_dir / "run.log"
    logfile = open(log_path, "a", encoding="utf-8")
    logfile.write(f"\n{'='*60}\n")
    logfile.write(f"[{run_start.isoformat()}] {'Resume' if args.resume else 'Start'}\n")
    logfile.write(f"seeds={seeds} temp={TEMPERATURE} candidates={len(candidates)}\n")
    logfile.flush()

    total_inference_time = 0
    error_count = 0
    all_stabilities = []
    processed_count = 0

    for idx, cand in enumerate(candidates):
        repo, filename = cand["repo"], cand["filename"]
        file_key = f"{repo}/{filename}"

        if file_key in completed_keys:
            continue

        processed_count += 1
        print(f"[{idx+1}/{len(candidates)}] {file_key} ", end="", flush=True)

        stripped = strip_comments(cand["filepath"])
        if stripped is None:
            stripped = Path(cand["filepath"]).read_text(encoding="utf-8", errors="ignore")

        user_prompt = (
            f"Analyze this Hyperledger Fabric chaincode file '{filename}' "
            f"for security vulnerabilities:\n\n```go\n{stripped}\n```"
        )

        seed_data = {}
        file_elapsed_total = 0
        any_truncated = False

        for seed in seeds:
            t0 = time.time()
            try:
                response = llm.create_chat_completion(
                    messages=[
                        {"role": "system", "content": PROMPT_ZERO_SHOT},
                        {"role": "user", "content": user_prompt},
                    ],
                    max_tokens=INFERENCE_PARAMS["max_tokens"],
                    temperature=TEMPERATURE,
                    seed=seed,
                )
                raw = response["choices"][0]["message"]["content"]
                finish_reason = response["choices"][0].get("finish_reason", "unknown")
                usage = response.get("usage", {})
                prompt_tokens = usage.get("prompt_tokens", 0)
                completion_tokens = usage.get("completion_tokens", 0)
                pred = classify_response(raw)
                if finish_reason == "length":
                    any_truncated = True
            except Exception as e:
                raw = f"ERROR: {e}"
                pred = "error"
                finish_reason = "error"
                prompt_tokens = 0
                completion_tokens = 0
                error_count += 1

            elapsed = time.time() - t0
            file_elapsed_total += elapsed
            total_inference_time += elapsed

            seed_data[str(seed)] = {
                "raw_response": raw,
                "prediction": pred,
                "elapsed_s": round(elapsed, 2),
                "finish_reason": finish_reason,
                "prompt_tokens": prompt_tokens,
                "completion_tokens": completion_tokens,
            }

        preds = [seed_data[str(s)]["prediction"] for s in seeds]
        vuln_count = preds.count("vulnerable")
        safe_count = preds.count("safe")
        majority = "vulnerable" if vuln_count > safe_count else "safe"
        stability = max(vuln_count, safe_count) / len(preds)
        avg_elapsed = file_elapsed_total / len(seeds)
        all_stabilities.append(stability)

        labels = "".join("V" if p == "vulnerable" else ("S" if p == "safe" else "E") for p in preds)
        trunc_flag = " [TRUNC]" if any_truncated else ""
        print(f"{labels} stab={stability:.0%} ({file_elapsed_total:.1f}s){trunc_flag}")

        logfile.write(f"[{idx+1}/{len(candidates)}] {file_key} {labels} stab={stability:.2f} elapsed={file_elapsed_total:.1f}s{trunc_flag}\n")
        logfile.flush()

        # --- Save per-file JSON (with all raw responses) ---
        safe_repo = repo.replace("/", "_").replace("\\", "_")
        safe_fn = filename.replace(".go", "")
        per_file_json = {
            "repo": repo,
            "filename": filename,
            "ground_truth": cand["ground_truth"],
            "primary_class": cand["primary_class"],
            "content_hash": cand["content_hash"],
            "seeds": seed_data,
            "stability": round(stability, 4),
            "majority_vote": majority,
            "avg_elapsed_s": round(avg_elapsed, 2),
            "total_elapsed_s": round(file_elapsed_total, 2),
            "any_truncated": any_truncated,
            "timestamp": datetime.now().isoformat(),
        }
        with open(per_file_dir / f"{safe_repo}_{safe_fn}.json", "w", encoding="utf-8") as f:
            json.dump(per_file_json, f, indent=2, ensure_ascii=False)

        # --- Write CSV row (incremental) ---
        csv_row = {
            "idx": idx,
            "repo": repo,
            "filename": filename,
            "ground_truth": cand["ground_truth"],
            "primary_class": cand["primary_class"],
            "stability": round(stability, 4),
            "majority_vote": majority,
            "avg_elapsed_s": round(avg_elapsed, 2),
            "total_elapsed_s": round(file_elapsed_total, 2),
            "any_truncated": any_truncated,
        }
        for s in seeds:
            csv_row[f"seed_{s}"] = seed_data[str(s)]["prediction"]
        writer.writerow(csv_row)
        csvfile.flush()

        # --- Update progress.json ---
        completed_keys.add(file_key)
        with open(progress_path, "w", encoding="utf-8") as f:
            json.dump({
                "completed_files": sorted(completed_keys),
                "total_done": len(completed_keys),
                "total_target": len(candidates),
                "errors": error_count,
                "last_updated": datetime.now().isoformat(),
            }, f, indent=2)

    csvfile.close()
    run_end = datetime.now()

    # --- Final summary ---
    avg_stab = sum(all_stabilities) / len(all_stabilities) if all_stabilities else 0
    perfect = sum(1 for s in all_stabilities if s == 1.0)

    print(f"\n{'='*60}")
    print(f"[Done] Robustness Run Complete")
    print(f"  Files processed: {processed_count}/{len(candidates)}")
    print(f"  Skipped (resumed): {len(candidates) - processed_count - len(path_issues)}")
    print(f"  Path issues: {len(path_issues)}")
    print(f"  Errors: {error_count}")
    print(f"  Avg stability: {avg_stab:.3f}")
    print(f"  Perfect stability (5/5): {perfect}/{len(all_stabilities)}")
    print(f"  Total inference time: {total_inference_time:.0f}s ({total_inference_time/60:.1f}min)")
    print(f"  Output: {run_dir.name}/")

    # --- Update meta.json ---
    meta["run_end"] = run_end.isoformat()
    meta["total_files"] = len(candidates)
    meta["processed_this_session"] = processed_count
    meta["path_issues"] = path_issues
    meta["errors"] = error_count
    meta["avg_stability"] = round(avg_stab, 4) if all_stabilities else None
    meta["perfect_stability"] = perfect
    meta["total_inference_s"] = round(total_inference_time, 1)
    meta["avg_per_file_s"] = round(total_inference_time / max(processed_count, 1), 1)
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2, ensure_ascii=False)

    logfile.write(f"[{run_end.isoformat()}] Done. processed={processed_count} errors={error_count} avg_stab={avg_stab:.3f}\n")
    logfile.close()


if __name__ == "__main__":
    main()
