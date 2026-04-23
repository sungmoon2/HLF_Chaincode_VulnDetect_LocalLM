"""
26_run_main_experiment.py — Phase 6: Main Experiment
Qwen2.5-Coder-7B (temp=0.0) + Semgrep on frozen benchmark (464 files).
Follows INFERENCE_CONTRACT.md exactly.
"""

import argparse, csv, json, os, re, subprocess, sys, time
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
SEMGREP_RULES = PROJECT_ROOT / "rules" / "hlf_consensus.yml"

INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 16384,
    "temperature": 0.0,
    "max_tokens": 2048,
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


def strip_comments(filepath: str):
    import shutil, tempfile
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


def classify_response(response: str) -> str:
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


def run_semgrep_on_file(filepath: str) -> dict:
    if not SEMGREP_RULES.exists():
        return {"error": "rules file not found", "findings": [], "hit": False}
    cmd = ["semgrep", "scan", "--config", str(SEMGREP_RULES), filepath, "--json", "--quiet"]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", timeout=30)
        data = json.loads(result.stdout) if result.stdout else {"results": []}
        findings = data.get("results", [])
        rule_ids = list(set(r.get("check_id", "").split(".")[-1] for r in findings))
        return {"findings_count": len(findings), "rule_ids": rule_ids, "hit": len(findings) > 0}
    except Exception as e:
        return {"error": str(e), "findings_count": 0, "rule_ids": [], "hit": False}


def main():
    parser = argparse.ArgumentParser(description="Phase 6: Main Experiment")
    parser.add_argument("--smoke", type=int, default=0)
    parser.add_argument("--resume", action="store_true")
    parser.add_argument("--skip-qwen", action="store_true", help="Semgrep only")
    parser.add_argument("--skip-semgrep", action="store_true", help="Qwen only")
    args = parser.parse_args()

    run_start = datetime.now()
    timestamp = run_start.strftime("%y%m%d_%H%M")
    run_dir = RESULTS_DIR / f"main_{timestamp}"
    run_dir.mkdir(parents=True, exist_ok=True)

    with open(BENCHMARK_FREEZE, encoding="utf-8") as f:
        freeze = json.load(f)
    bench_files = freeze["files"]
    print(f"{Fore.CYAN}[Benchmark] {len(bench_files)} files ({freeze['vulnerable']}V + {freeze['safe']}S)")

    # Load Qwen model
    llm = None
    if not args.skip_qwen:
        from llama_cpp import Llama
        print(f"{Fore.CYAN}[Model] Loading {MODEL_PATH.name}...")
        llm = Llama(
            model_path=str(MODEL_PATH),
            n_gpu_layers=INFERENCE_PARAMS["n_gpu_layers"],
            n_ctx=INFERENCE_PARAMS["n_ctx"],
            verbose=False,
        )
        print(f"{Fore.GREEN}[Model] Loaded")

    # Progress tracking
    progress_path = run_dir / "progress.json"
    completed = set()
    if args.resume and progress_path.exists():
        with open(progress_path, encoding="utf-8") as f:
            prog = json.load(f)
        completed = set(prog.get("completed_files", []))
        print(f"{Fore.YELLOW}[Resume] {len(completed)} completed")

    candidates = []
    for bf in bench_files:
        filepath = BENCHMARK_DIR / bf["repo"] / bf["filename"]
        if filepath.exists():
            candidates.append({
                "repo": bf["repo"], "filename": bf["filename"],
                "filepath": str(filepath), "ground_truth": bf["verdict"],
                "primary_class": bf.get("primary_class", ""),
            })

    if args.smoke > 0:
        candidates = candidates[:args.smoke]

    pending = [c for c in candidates if f"{c['repo']}/{c['filename']}" not in completed]
    print(f"{Fore.GREEN}[Run] {len(pending)} files pending\n")

    # CSV setup
    csv_fields = ["idx", "repo", "filename", "ground_truth", "ground_truth_class",
                  "qwen_prediction", "qwen_elapsed_s", "qwen_raw_length",
                  "semgrep_prediction", "semgrep_rules_hit", "semgrep_findings_count"]
    csv_path = run_dir / "results.csv"
    csv_mode = "a" if args.resume and csv_path.exists() else "w"
    csvfile = open(csv_path, csv_mode, newline="", encoding="utf-8")
    writer = csv.DictWriter(csvfile, fieldnames=csv_fields)
    if csv_mode == "w":
        writer.writeheader()

    error_count = 0
    total_qwen_time = 0

    for idx, cand in enumerate(pending):
        repo, filename = cand["repo"], cand["filename"]
        filepath = cand["filepath"]
        file_key = f"{repo}/{filename}"
        gt = cand["ground_truth"]
        gt_class = cand["primary_class"]

        print(f"[{idx+1}/{len(pending)}] {file_key} (GT={gt}) ", end="", flush=True)

        # Strip comments
        stripped = strip_comments(filepath)
        if stripped is None:
            stripped = Path(filepath).read_text(encoding="utf-8", errors="ignore")

        # Qwen inference
        qwen_pred = "skip"
        qwen_elapsed = 0
        qwen_raw_len = 0
        if llm is not None:
            user_prompt = f"Analyze this Hyperledger Fabric chaincode file '{filename}' for security vulnerabilities:\n\n```go\n{stripped}\n```"
            t0 = time.time()
            try:
                response = llm.create_chat_completion(
                    messages=[
                        {"role": "system", "content": PROMPT_ZERO_SHOT},
                        {"role": "user", "content": user_prompt},
                    ],
                    max_tokens=INFERENCE_PARAMS["max_tokens"],
                    temperature=INFERENCE_PARAMS["temperature"],
                )
                raw = response["choices"][0]["message"]["content"]
                qwen_pred = classify_response(raw)
                qwen_raw_len = len(raw)
            except Exception as e:
                qwen_pred = "error"
                raw = f"ERROR: {e}"
                error_count += 1
            qwen_elapsed = time.time() - t0
            total_qwen_time += qwen_elapsed

        # Semgrep
        semgrep_pred = "skip"
        semgrep_rules = ""
        semgrep_count = 0
        if not args.skip_semgrep:
            sg = run_semgrep_on_file(filepath)
            semgrep_pred = "vulnerable" if sg.get("hit") else "safe"
            semgrep_rules = ",".join(sg.get("rule_ids", []))
            semgrep_count = sg.get("findings_count", 0)

        # Print result
        q_mark = "V" if qwen_pred == "vulnerable" else ("S" if qwen_pred == "safe" else "E")
        s_mark = "V" if semgrep_pred == "vulnerable" else ("S" if semgrep_pred == "safe" else "-")
        print(f"Q={q_mark} SG={s_mark} ({qwen_elapsed:.1f}s)")

        # Save per-file detail
        detail = {
            "repo": repo, "filename": filename,
            "ground_truth": gt, "ground_truth_class": gt_class,
            "qwen_prediction": qwen_pred, "qwen_elapsed_s": round(qwen_elapsed, 2),
            "qwen_raw_response": raw if llm else "",
            "semgrep_prediction": semgrep_pred, "semgrep_rules_hit": semgrep_rules,
            "semgrep_findings_count": semgrep_count,
            "timestamp": datetime.now().isoformat(),
        }
        per_file_dir = run_dir / "per_file"
        per_file_dir.mkdir(exist_ok=True)
        with open(per_file_dir / f"{repo}_{filename.replace('.go','')}.json", "w", encoding="utf-8") as f:
            json.dump(detail, f, indent=2, ensure_ascii=False)

        row = {field: detail.get(field, "") for field in csv_fields}
        row["idx"] = idx + len(completed)
        writer.writerow(row)
        csvfile.flush()

        completed.add(file_key)
        with open(progress_path, "w", encoding="utf-8") as f:
            json.dump({
                "completed_files": sorted(completed),
                "total_done": len(completed),
                "total_target": len(candidates),
                "last_updated": datetime.now().isoformat(),
                "errors": error_count,
            }, f, indent=2)

    csvfile.close()
    run_end = datetime.now()

    # Summary
    print(f"\n{'='*60}")
    print(f"[Done] Main Experiment Complete")
    print(f"  Files: {len(pending)}")
    print(f"  Errors: {error_count}")
    print(f"  Qwen total time: {total_qwen_time:.1f}s (avg {total_qwen_time/max(len(pending),1):.1f}s/file)")
    print(f"  Output: {run_dir.name}/")

    # Save meta
    meta = {
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "total_files": len(candidates),
        "processed": len(pending),
        "errors": error_count,
        "qwen_total_s": round(total_qwen_time, 1),
        "model": str(MODEL_PATH.name),
        "inference_params": INFERENCE_PARAMS,
        "semgrep_version": "1.151.0",
        "semgrep_rules": str(SEMGREP_RULES),
    }
    with open(run_dir / "meta.json", "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2, ensure_ascii=False)


if __name__ == "__main__":
    main()
