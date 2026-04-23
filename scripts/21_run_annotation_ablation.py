"""
21_run_annotation_ablation.py  (v1.0 — 2026-04-22)
Controlled Annotation Ablation: 2-condition factorial on D1 pilot set

Design:
  Condition A (ann): D1 original (inline comments preserved) + neutral filenames
  Condition B (abl): D1-clean  (inline comments removed)  + neutral filenames

Controls (all matched to D1-clean / D2 settings):
  n_ctx      = 16384
  prompt     = P1 zero-shot (identical to 20_run_addon_validation.py)
  temperature= 0.1
  max_tokens = 2048
  runs       = 5
  filenames  = P01.go .. P15.go (neutral, randomized mapping)

Confounds eliminated vs prior D1-vs-D1clean comparison:
  1. n_ctx mismatch (was 4096 vs 16384) -> now both 16384
  2. Prompt strategy (was 3-way vs zero-shot) -> now both zero-shot
  3. Run count (was 1 vs 5) -> now both 5
  4. Filename leakage (was safe_/vuln_ prefix) -> now P##.go neutral

Output:
  06_addon_validation/results/ablation_ann_YYMMDD_HHMM.csv
  06_addon_validation/results/ablation_abl_YYMMDD_HHMM.csv
  06_addon_validation/results/ablation_meta_YYMMDD_HHMM.json
  06_addon_validation/results/ABLATION_REPORT_YYMMDD_HHMM.md

Usage:
  python scripts/21_run_annotation_ablation.py
  python scripts/21_run_annotation_ablation.py --runs 1          # pilot
  python scripts/21_run_annotation_ablation.py --models qwen     # single model
  python scripts/21_run_annotation_ablation.py --skip-setup       # reuse existing neutral dirs
"""

import argparse
import csv
import json
import os
import platform
import random
import shutil
import sys
import time
from collections import defaultdict
from datetime import datetime
from pathlib import Path

from colorama import init, Fore
from tqdm import tqdm

init(autoreset=True)

PROJECT_ROOT = Path(__file__).resolve().parent.parent
ADDON_DIR = PROJECT_ROOT / "06_addon_validation"
RESULTS_DIR = ADDON_DIR / "results"
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"

D1_ORIGINAL_DIR = PROJECT_ROOT / "02_resources" / "dataset"
D1_CLEAN_DIR = ADDON_DIR / "dataset_d1_clean"

NEUTRAL_ANN_DIR = ADDON_DIR / "dataset_ablation_ann"
NEUTRAL_ABL_DIR = ADDON_DIR / "dataset_ablation_abl"
MAPPING_FILE = ADDON_DIR / "ablation_filename_mapping.json"

MODEL_FILES = {
    "qwen": "qwen2.5-coder-7b-instruct-q4_k_m.gguf",
    "llama": "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
}

INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 16384,
    "temperature": 0.1,
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

GROUND_TRUTH_BY_ORIGINAL = {
    "safe_01_logging.go": "safe",
    "safe_02_local_var.go": "safe",
    "safe_03_map_read.go": "safe",
    "safe_04_math_rand.go": "safe",
    "safe_05_deterministic_time.go": "safe",
    "safe_06_external_lib.go": "safe",
    "vuln_01_time.go": "vulnerable",
    "vuln_01_b_interprocedural.go": "vulnerable",
    "vuln_02_global.go": "vulnerable",
    "vuln_03_goroutine.go": "vulnerable",
    "vuln_04_map_iter.go": "vulnerable",
    "vuln_04_b_nested_map.go": "vulnerable",
    "vuln_05_phantom.go": "vulnerable",
    "vuln_06_iterator_leak.go": "vulnerable",
    "vuln_06_b_conditional_leak.go": "vulnerable",
}

RANDOM_SEED = 20260422


def create_neutral_datasets():
    """Create P01~P15 neutral-filename copies with randomized mapping."""
    original_files = sorted(GROUND_TRUTH_BY_ORIGINAL.keys())
    assert len(original_files) == 15, f"Expected 15 files, got {len(original_files)}"

    rng = random.Random(RANDOM_SEED)
    shuffled = list(original_files)
    rng.shuffle(shuffled)

    mapping = {}
    ground_truth_neutral = {}
    for i, orig_name in enumerate(shuffled, 1):
        neutral = f"P{i:02d}.go"
        mapping[neutral] = orig_name
        ground_truth_neutral[neutral] = GROUND_TRUTH_BY_ORIGINAL[orig_name]

    for d in [NEUTRAL_ANN_DIR, NEUTRAL_ABL_DIR]:
        d.mkdir(parents=True, exist_ok=True)
        for f in d.glob("*.go"):
            f.unlink()

    for neutral, orig in mapping.items():
        src_ann = D1_ORIGINAL_DIR / orig
        src_abl = D1_CLEAN_DIR / orig
        if not src_ann.exists():
            print(f"{Fore.RED}[Error] Missing: {src_ann}")
            sys.exit(1)
        if not src_abl.exists():
            print(f"{Fore.RED}[Error] Missing: {src_abl}")
            sys.exit(1)
        shutil.copy2(src_ann, NEUTRAL_ANN_DIR / neutral)
        shutil.copy2(src_abl, NEUTRAL_ABL_DIR / neutral)

    mapping_data = {
        "seed": RANDOM_SEED,
        "created": datetime.now().isoformat(),
        "mapping": mapping,
        "ground_truth": ground_truth_neutral,
        "vuln_count": sum(1 for v in ground_truth_neutral.values() if v == "vulnerable"),
        "safe_count": sum(1 for v in ground_truth_neutral.values() if v == "safe"),
    }
    with open(MAPPING_FILE, "w", encoding="utf-8") as f:
        json.dump(mapping_data, f, indent=2, ensure_ascii=False)

    print(f"{Fore.GREEN}[Setup] Neutral datasets created:")
    print(f"  Ann: {NEUTRAL_ANN_DIR} ({len(list(NEUTRAL_ANN_DIR.glob('*.go')))} files)")
    print(f"  Abl: {NEUTRAL_ABL_DIR} ({len(list(NEUTRAL_ABL_DIR.glob('*.go')))} files)")
    print(f"  Mapping: {MAPPING_FILE}")

    for neutral in sorted(mapping.keys()):
        gt = ground_truth_neutral[neutral]
        print(f"    {neutral} <- {mapping[neutral]} [{gt}]")

    return ground_truth_neutral


def load_mapping():
    with open(MAPPING_FILE, "r", encoding="utf-8") as f:
        data = json.load(f)
    return data["ground_truth"]


def classify_response(response: str) -> str:
    if not response or response.startswith("ERROR:"):
        return "error"
    resp_lower = response.lower()
    safe_indicators = [
        "no vulnerabilities detected",
        "no vulnerabilities found",
        "no security vulnerabilities",
        "no significant vulnerabilities",
        "no vulnerabilities were found",
        "no vulnerabilities were detected",
        "no critical vulnerabilities",
        "the code appears to be secure",
        "the code is secure",
        "no issues found",
        "no issues detected",
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
    vuln_count = sum(1 for ind in vuln_indicators if ind in resp_lower)
    if vuln_count >= 2:
        return "vulnerable"
    return "vulnerable"


def clopper_pearson(k: int, n: int, alpha: float = 0.05):
    if n == 0:
        return (0.0, 1.0)
    from scipy.stats import beta as beta_dist
    lo = 0.0 if k == 0 else beta_dist.ppf(alpha / 2, k, n - k + 1)
    hi = 1.0 if k == n else beta_dist.ppf(1 - alpha / 2, k + 1, n - k)
    return (round(lo, 4), round(hi, 4))


def calculate_metrics(predictions, ground_truth):
    tp = fp = fn = tn = 0
    for fname, gt in ground_truth.items():
        pred = predictions.get(fname, "safe")
        if gt == "vulnerable" and pred == "vulnerable":
            tp += 1
        elif gt == "safe" and pred == "vulnerable":
            fp += 1
        elif gt == "vulnerable" and pred == "safe":
            fn += 1
        elif gt == "safe" and pred == "safe":
            tn += 1
    n_pos = tp + fn
    n_neg = tn + fp
    tpr = tp / n_pos if n_pos > 0 else 0.0
    tnr = tn / n_neg if n_neg > 0 else 0.0
    return {
        "TP": tp, "FP": fp, "FN": fn, "TN": tn,
        "TPR": round(tpr, 4), "TNR": round(tnr, 4),
        "TPR_CI": clopper_pearson(tp, n_pos),
        "TNR_CI": clopper_pearson(tn, n_neg),
        "n_pos": n_pos, "n_neg": n_neg,
    }


def load_model(model_path):
    from llama_cpp import Llama
    print(f"{Fore.CYAN}[Model] Loading: {model_path.name} (n_ctx={INFERENCE_PARAMS['n_ctx']})")
    llm = Llama(
        model_path=str(model_path),
        n_gpu_layers=INFERENCE_PARAMS["n_gpu_layers"],
        n_ctx=INFERENCE_PARAMS["n_ctx"],
        verbose=False,
    )
    print(f"{Fore.GREEN}[Model] Loaded: {model_path.name}")
    return llm


def audit_chaincode(llm, code: str, filename: str) -> str:
    user_prompt = (
        f"Analyze this Hyperledger Fabric chaincode file '{filename}' "
        f"for security vulnerabilities:\n\n```go\n{code}\n```"
    )
    response = llm.create_chat_completion(
        messages=[
            {"role": "system", "content": PROMPT_ZERO_SHOT},
            {"role": "user", "content": user_prompt},
        ],
        max_tokens=INFERENCE_PARAMS["max_tokens"],
        temperature=INFERENCE_PARAMS["temperature"],
    )
    return response["choices"][0]["message"]["content"]


def run_condition(llm, dataset_dir, ground_truth, n_runs, csv_path):
    """Run one condition (ann or abl) and write CSV."""
    go_files = sorted(dataset_dir.glob("*.go"))
    fields = ["run", "file", "code_chars", "elapsed_s",
              "classification", "ground_truth", "correct", "result"]

    run_predictions = defaultdict(dict)

    with open(csv_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fields)
        writer.writeheader()

        for run_num in range(1, n_runs + 1):
            print(f"\n{Fore.CYAN}--- Run {run_num}/{n_runs} ---")
            for go_file in tqdm(go_files, desc=f"R{run_num}"):
                fname = go_file.name
                code = go_file.read_text(encoding="utf-8")

                start_t = time.time()
                try:
                    result = audit_chaincode(llm, code, fname)
                except Exception as e:
                    result = f"ERROR: {e}"
                elapsed = round(time.time() - start_t, 3)

                classification = classify_response(result)
                gt = ground_truth.get(fname, "unknown")
                run_predictions[run_num][fname] = classification

                writer.writerow({
                    "run": run_num,
                    "file": fname,
                    "code_chars": len(code),
                    "elapsed_s": elapsed,
                    "classification": classification,
                    "ground_truth": gt,
                    "correct": classification == gt,
                    "result": result[:2000],
                })
                f.flush()

    return dict(run_predictions)


def compute_summary(run_predictions, ground_truth, n_runs):
    per_run = []
    for run_num in range(1, n_runs + 1):
        m = calculate_metrics(run_predictions[run_num], ground_truth)
        per_run.append(m)
    tprs = [m["TPR"] for m in per_run]
    tnrs = [m["TNR"] for m in per_run]
    return {
        "per_run": per_run,
        "tpr_mean": round(sum(tprs) / len(tprs), 4),
        "tnr_mean": round(sum(tnrs) / len(tnrs), 4),
        "tpr_range": (min(tprs), max(tprs)),
        "tnr_range": (min(tnrs), max(tnrs)),
    }


def generate_report(path, results, timestamp, n_runs, mapping_data):
    lines = [
        "# Controlled Annotation Ablation Report",
        "",
        f"> Generated: {timestamp}",
        f"> Runs: {n_runs} | n_ctx: {INFERENCE_PARAMS['n_ctx']} | temp: {INFERENCE_PARAMS['temperature']}",
        f"> Filenames: P01.go~P15.go (neutral, seed={RANDOM_SEED})",
        f"> Vuln: {mapping_data.get('vuln_count', 9)} | Safe: {mapping_data.get('safe_count', 6)}",
        "",
        "## Confound Control",
        "",
        "| Variable | Value | Status |",
        "|:---------|:------|:-------|",
        f"| n_ctx | {INFERENCE_PARAMS['n_ctx']} | Matched |",
        "| prompt | P1 zero-shot | Matched |",
        f"| temperature | {INFERENCE_PARAMS['temperature']} | Matched |",
        f"| runs | {n_runs} | Matched |",
        "| filenames | P01~P15 (neutral) | Controlled |",
        "| **inline comments** | **ann vs abl** | **Experimental variable** |",
        "",
        "## File Mapping",
        "",
        "| Neutral | Original | Ground Truth |",
        "|:--------|:---------|:-------------|",
    ]
    mapping = mapping_data.get("mapping", {})
    gt = mapping_data.get("ground_truth", {})
    for neutral in sorted(mapping.keys()):
        lines.append(f"| {neutral} | {mapping[neutral]} | {gt[neutral]} |")

    for model_key in sorted(results.keys()):
        model_results = results[model_key]
        lines.extend(["", f"## {model_key.upper()}", ""])

        for cond in ["ann", "abl"]:
            if cond not in model_results:
                continue
            s = model_results[cond]
            label = "Annotated (comments)" if cond == "ann" else "Ablated (no comments)"
            lines.extend([
                f"### {label}",
                "",
                "| Run | TP | FP | FN | TN | TPR | TNR |",
                "|:----|:---|:---|:---|:---|:----|:----|",
            ])
            for i, m in enumerate(s["per_run"], 1):
                lines.append(
                    f"| {i} | {m['TP']} | {m['FP']} | {m['FN']} | {m['TN']} | "
                    f"{m['TPR']:.2f} | {m['TNR']:.2f} |"
                )
            lines.extend([
                "",
                f"- **TPR mean**: {s['tpr_mean']:.4f} (range: {s['tpr_range'][0]:.2f}-{s['tpr_range'][1]:.2f})",
                f"- **TNR mean**: {s['tnr_mean']:.4f} (range: {s['tnr_range'][0]:.2f}-{s['tnr_range'][1]:.2f})",
                "",
            ])

        if "ann" in model_results and "abl" in model_results:
            ann_s = model_results["ann"]
            abl_s = model_results["abl"]
            tpr_diff = ann_s["tpr_mean"] - abl_s["tpr_mean"]
            tnr_diff = ann_s["tnr_mean"] - abl_s["tnr_mean"]
            lines.extend([
                f"### Annotation Effect ({model_key})",
                "",
                f"- **TPR diff (ann - abl)**: {tpr_diff:+.4f}",
                f"- **TNR diff (ann - abl)**: {tnr_diff:+.4f}",
                f"- Interpretation: positive = comments help, negative = comments hurt",
                "",
            ])

    lines.extend([
        "---",
        "",
        "## Methodology Notes",
        "",
        "- All confounds from prior D1-vs-D1clean comparison are eliminated:",
        "  - n_ctx unified (was 4096 vs 16384)",
        "  - prompt strategy unified (was 3-way vs zero-shot)",
        "  - run count unified (was 1 vs 5)",
        "  - filename leakage removed (was safe_/vuln_ prefix, now P##.go)",
        "- The ONLY difference between conditions is presence/absence of inline comments.",
        "- Classifier v2 (identical to scripts 08/20).",
    ])

    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")


def parse_args():
    p = argparse.ArgumentParser(description="Controlled Annotation Ablation")
    p.add_argument("--runs", type=int, default=5)
    p.add_argument("--models", nargs="+", choices=["qwen", "llama"], default=["qwen", "llama"])
    p.add_argument("--skip-setup", action="store_true",
                   help="Skip dataset creation (reuse existing neutral dirs)")
    return p.parse_args()


def main():
    args = parse_args()
    run_start = datetime.now()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = run_start.strftime("%y%m%d_%H%M")

    if not args.skip_setup:
        ground_truth = create_neutral_datasets()
    else:
        ground_truth = load_mapping()

    with open(MAPPING_FILE, "r", encoding="utf-8") as f:
        mapping_data = json.load(f)

    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN} Controlled Annotation Ablation")
    print(f"{Fore.GREEN} Conditions: ann (comments) vs abl (no comments)")
    print(f"{Fore.GREEN} Files: 15 (neutral P01~P15) | Runs: {args.runs} | Models: {args.models}")
    print(f"{Fore.GREEN} n_ctx: {INFERENCE_PARAMS['n_ctx']} | temp: {INFERENCE_PARAMS['temperature']}")
    print(f"{Fore.GREEN}{'='*60}\n")

    all_results = {}

    for model_key in args.models:
        model_path = MODELS_DIR / MODEL_FILES[model_key]
        if not model_path.exists():
            print(f"{Fore.RED}[Skip] Model not found: {MODEL_FILES[model_key]}")
            continue

        llm = load_model(model_path)
        all_results[model_key] = {}

        for cond, dataset_dir in [("ann", NEUTRAL_ANN_DIR), ("abl", NEUTRAL_ABL_DIR)]:
            label = "Annotated" if cond == "ann" else "Ablated"
            print(f"\n{Fore.YELLOW}{'='*60}")
            print(f"{Fore.YELLOW} {model_key.upper()} - {label} - {args.runs} runs")
            print(f"{Fore.YELLOW}{'='*60}")

            csv_path = RESULTS_DIR / f"ablation_{cond}_{model_key}_{timestamp}.csv"
            preds = run_condition(llm, dataset_dir, ground_truth, args.runs, csv_path)
            summary = compute_summary(preds, ground_truth, args.runs)
            all_results[model_key][cond] = summary

            print(f"\n{Fore.GREEN}[{model_key}/{cond}] TPR={summary['tpr_mean']:.4f} "
                  f"TNR={summary['tnr_mean']:.4f}")
            for i, m in enumerate(summary["per_run"], 1):
                print(f"  Run {i}: TP={m['TP']} FP={m['FP']} FN={m['FN']} TN={m['TN']} "
                      f"TPR={m['TPR']:.2f} TNR={m['TNR']:.2f}")

        del llm

    report_path = RESULTS_DIR / f"ABLATION_REPORT_{timestamp}.md"
    generate_report(report_path, all_results, timestamp, args.runs, mapping_data)

    meta = {
        "script": "21_run_annotation_ablation.py",
        "version": "1.0",
        "timestamp": run_start.isoformat(),
        "duration_s": round(time.time() - run_start.timestamp(), 1),
        "conditions": ["ann", "abl"],
        "n_files": 15,
        "n_runs": args.runs,
        "models": args.models,
        "inference_params": INFERENCE_PARAMS,
        "random_seed": RANDOM_SEED,
        "mapping_file": str(MAPPING_FILE),
        "confounds_controlled": [
            "n_ctx unified at 16384",
            "prompt strategy: zero-shot only",
            "run count: matched",
            "filenames: neutral P01-P15",
        ],
        "experimental_variable": "inline comments (ann=present, abl=removed)",
        "platform": platform.platform(),
        "results_summary": {
            model: {
                cond: {"tpr_mean": s["tpr_mean"], "tnr_mean": s["tnr_mean"]}
                for cond, s in conds.items()
            }
            for model, conds in all_results.items()
        },
    }
    meta_path = RESULTS_DIR / f"ablation_meta_{timestamp}.json"
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2, ensure_ascii=False)

    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN} Ablation complete!")
    print(f"{Fore.GREEN} Report: {report_path.name}")
    print(f"{Fore.GREEN} Meta: {meta_path.name}")

    for model_key in sorted(all_results.keys()):
        mr = all_results[model_key]
        if "ann" in mr and "abl" in mr:
            d_tpr = mr["ann"]["tpr_mean"] - mr["abl"]["tpr_mean"]
            d_tnr = mr["ann"]["tnr_mean"] - mr["abl"]["tnr_mean"]
            print(f"{Fore.YELLOW} {model_key}: TPR diff={d_tpr:+.4f}  TNR diff={d_tnr:+.4f}")

    print(f"{Fore.GREEN}{'='*60}")


if __name__ == "__main__":
    main()
