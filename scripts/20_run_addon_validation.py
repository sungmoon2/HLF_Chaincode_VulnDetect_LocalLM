"""
20_run_addon_validation.py  (v1.0 — 2026-04-21)
Add-on External Validation: 17 public HLF chaincode files

- Phase A: Semgrep (auto + security-audit + hlf_consensus.yml) on 17 files
- Phase B: Qwen2.5-Coder-7B zero_shot on 17 files × 5 runs
- Phase C: Llama-3.1-8B zero_shot on 17 files × 5 runs
- Classifier v2 (same as 08_run_golisa_validation.py)
- Ground truth comparison + TPR/TNR/CI calculation
- Output: CSV + meta.json + ADDON_VALIDATION_REPORT.md

Usage:
    python scripts/20_run_addon_validation.py
    python scripts/20_run_addon_validation.py --skip-semgrep
    python scripts/20_run_addon_validation.py --skip-llm
    python scripts/20_run_addon_validation.py --runs 1
"""

import argparse
import csv
import hashlib
import json
import math
import os
import platform
import shutil
import subprocess
import sys
import time
from collections import Counter, defaultdict
from datetime import datetime
from pathlib import Path

from colorama import init, Fore, Style
from tqdm import tqdm

init(autoreset=True)

# ══════════════════════════════════════════════════════════════════════
# Path configuration
# ══════════════════════════════════════════════════════════════════════
PROJECT_ROOT = Path(__file__).resolve().parent.parent
ADDON_DIR = PROJECT_ROOT / "06_addon_validation"
DATASET_DIR = ADDON_DIR / "dataset"
RESULTS_DIR = ADDON_DIR / "results"
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"
HLF_RULES = PROJECT_ROOT / "rules" / "hlf_consensus.yml"

MODEL_FILES = {
    "qwen": "qwen2.5-coder-7b-instruct-q4_k_m.gguf",
    "llama": "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
}

# ══════════════════════════════════════════════════════════════════════
# Inference parameters (GoLiSA-consistent)
# ══════════════════════════════════════════════════════════════════════
INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 16384,
    "temperature": 0.1,
    "max_tokens": 2048,
}

# ══════════════════════════════════════════════════════════════════════
# Prompt (identical to 02_run_audit_v3.py P1 zero_shot)
# ══════════════════════════════════════════════════════════════════════
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

# ══════════════════════════════════════════════════════════════════════
# Ground truth labels
# ══════════════════════════════════════════════════════════════════════
GROUND_TRUTH = {
    "U01_LandRegistry.go": "safe",
    "U02_ethtxcc.go": "safe",
    "U03_election_code.go": "vulnerable",
    "U07_ProductDetails.go": "safe",
    "U08_smartcontract.go": "safe",
    "U09_charity.go": "safe",
    "U10_carcert.go": "safe",
    "U11_sharebook.go": "safe",
    "U12_local_model_chaincode.go": "safe",
    "U13_security_manager.go": "vulnerable",
    "U14_smartaicc.go": "safe",
    "U17_realty_chaincode.go": "safe",
    "U18_donation_chaincode.go": "vulnerable",
    "U20_maintenance.go": "vulnerable",
    "U21_movies.go": "safe",
    "U22_voting.go": "safe",
    "U23_private_blockchain.go": "safe",
}

VULN_CLASSES = {
    "U03_election_code.go": "timestamp",
    "U13_security_manager.go": "phantom_read",
    "U18_donation_chaincode.go": "timestamp",
    "U20_maintenance.go": "timestamp",
}

# Sensitivity analysis: U03=safe scenario
GROUND_TRUTH_U03_SAFE = {**GROUND_TRUTH, "U03_election_code.go": "safe"}

# ══════════════════════════════════════════════════════════════════════
# Semgrep configuration
# ══════════════════════════════════════════════════════════════════════
SEMGREP_CONFIGS = [
    {"label": "auto", "config": "auto"},
    {"label": "security-audit", "config": "p/security-audit"},
]


# ══════════════════════════════════════════════════════════════════════
# Classifier v2 (identical to 08_run_golisa_validation.py)
# ══════════════════════════════════════════════════════════════════════
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


def extract_vuln_types(response: str) -> list[str]:
    if not response:
        return []
    resp_lower = response.lower()
    found = []
    patterns = {
        "timestamp": ["time.now", "timestamp", "non-deterministic time", "nondeterministic time"],
        "global_var": ["global variable", "global state", "mutable global"],
        "goroutine": ["goroutine", "race condition", "concurrent"],
        "map_iter": ["map iteration", "iteration order", "range over map"],
        "phantom_read": ["phantom read", "read-after-write", "mvcc"],
        "iterator_leak": ["iterator", "close()", "resource leak", "getstatebyrange"],
    }
    for vtype, keywords in patterns.items():
        if any(kw in resp_lower for kw in keywords):
            found.append(vtype)
    return found


# ══════════════════════════════════════════════════════════════════════
# Clopper-Pearson exact 95% CI
# ══════════════════════════════════════════════════════════════════════
def clopper_pearson(k: int, n: int, alpha: float = 0.05) -> tuple[float, float]:
    if n == 0:
        return (0.0, 1.0)
    if k == 0:
        lo = 0.0
    else:
        lo = _beta_ppf(alpha / 2, k, n - k + 1)
    if k == n:
        hi = 1.0
    else:
        hi = _beta_ppf(1 - alpha / 2, k + 1, n - k)
    return (round(lo, 4), round(hi, 4))


def _beta_ppf(p: float, a: float, b: float) -> float:
    from scipy.stats import beta as beta_dist
    return beta_dist.ppf(p, a, b)


# ══════════════════════════════════════════════════════════════════════
# Semgrep runner
# ══════════════════════════════════════════════════════════════════════
def check_semgrep() -> str | None:
    path = shutil.which("semgrep")
    if path:
        return path
    try:
        result = subprocess.run(["semgrep", "--version"], capture_output=True, text=True, timeout=15)
        if result.returncode == 0:
            return "semgrep"
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass
    return None


def run_semgrep_on_file(semgrep_path: str, go_file: Path, config: str) -> dict:
    try:
        result = subprocess.run(
            [semgrep_path, "--config", config, "--json", "--no-git-ignore", str(go_file)],
            capture_output=True, text=True, timeout=120,
        )
        if result.returncode in (0, 1):
            data = json.loads(result.stdout) if result.stdout.strip() else {"results": []}
            findings = data.get("results", [])
            return {"findings": len(findings), "details": findings, "error": None}
        return {"findings": 0, "details": [], "error": f"returncode={result.returncode}"}
    except Exception as e:
        return {"findings": 0, "details": [], "error": str(e)}


def run_semgrep_hlf_on_file(semgrep_path: str, go_file: Path) -> dict:
    if not HLF_RULES.exists():
        return {"findings": 0, "details": [], "error": "hlf_consensus.yml not found"}
    try:
        result = subprocess.run(
            [semgrep_path, "--config", str(HLF_RULES), "--json", "--no-git-ignore", str(go_file)],
            capture_output=True, text=True, timeout=120,
        )
        if result.returncode in (0, 1):
            data = json.loads(result.stdout) if result.stdout.strip() else {"results": []}
            findings = data.get("results", [])
            return {"findings": len(findings), "details": findings, "error": None}
        return {"findings": 0, "details": [], "error": f"returncode={result.returncode}"}
    except Exception as e:
        return {"findings": 0, "details": [], "error": str(e)}


# ══════════════════════════════════════════════════════════════════════
# LLM runner
# ══════════════════════════════════════════════════════════════════════
def load_model(model_path: Path):
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


# ══════════════════════════════════════════════════════════════════════
# Metrics calculation
# ══════════════════════════════════════════════════════════════════════
def calculate_metrics(predictions: dict[str, str], ground_truth: dict[str, str]) -> dict:
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
    tpr_ci = clopper_pearson(tp, n_pos)
    tnr_ci = clopper_pearson(tn, n_neg)

    return {
        "TP": tp, "FP": fp, "FN": fn, "TN": tn,
        "TPR": round(tpr, 4), "TNR": round(tnr, 4),
        "TPR_CI_95": tpr_ci, "TNR_CI_95": tnr_ci,
        "n_pos": n_pos, "n_neg": n_neg,
    }


# ══════════════════════════════════════════════════════════════════════
# Main
# ══════════════════════════════════════════════════════════════════════
def parse_args():
    parser = argparse.ArgumentParser(description="Add-on External Validation")
    parser.add_argument("--runs", type=int, default=5, help="Number of LLM runs (default: 5)")
    parser.add_argument("--skip-semgrep", action="store_true", help="Skip Semgrep phase")
    parser.add_argument("--skip-llm", action="store_true", help="Skip LLM phase")
    parser.add_argument("--models", nargs="+", choices=["qwen", "llama"], default=["qwen", "llama"])
    return parser.parse_args()


def main():
    args = parse_args()
    run_start = datetime.now()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] No .go files in {DATASET_DIR}")
        sys.exit(1)

    print(f"{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN} Add-on External Validation")
    print(f"{Fore.GREEN} Files: {len(go_files)} | Runs: {args.runs} | Models: {args.models}")
    print(f"{Fore.GREEN}{'='*60}\n")

    timestamp = run_start.strftime("%y%m%d_%H%M")
    all_results = {}

    # ── Phase A: Semgrep ──────────────────────────────────────────────
    if not args.skip_semgrep:
        semgrep_path = check_semgrep()
        if not semgrep_path:
            print(f"{Fore.RED}[Error] semgrep not found. Install: pip install semgrep")
            print(f"{Fore.YELLOW}[Info] Skipping Semgrep phase.")
        else:
            print(f"\n{Fore.YELLOW}{'='*60}")
            print(f"{Fore.YELLOW} Phase A: Semgrep Analysis")
            print(f"{Fore.YELLOW}{'='*60}\n")

            semgrep_csv = RESULTS_DIR / f"addon_semgrep_{timestamp}.csv"
            semgrep_records = []

            for go_file in tqdm(go_files, desc="Semgrep"):
                fname = go_file.name

                for cfg in SEMGREP_CONFIGS:
                    result = run_semgrep_on_file(semgrep_path, go_file, cfg["config"])
                    classification = "vulnerable" if result["findings"] > 0 else "safe"
                    semgrep_records.append({
                        "file": fname,
                        "tool": f"semgrep_{cfg['label']}",
                        "findings": result["findings"],
                        "classification": classification,
                        "ground_truth": GROUND_TRUTH.get(fname, "unknown"),
                        "correct": classification == GROUND_TRUTH.get(fname, "unknown"),
                        "error": result["error"] or "",
                    })

                hlf_result = run_semgrep_hlf_on_file(semgrep_path, go_file)
                hlf_class = "vulnerable" if hlf_result["findings"] > 0 else "safe"
                semgrep_records.append({
                    "file": fname,
                    "tool": "semgrep_hlf",
                    "findings": hlf_result["findings"],
                    "classification": hlf_class,
                    "ground_truth": GROUND_TRUTH.get(fname, "unknown"),
                    "correct": hlf_class == GROUND_TRUTH.get(fname, "unknown"),
                    "error": hlf_result["error"] or "",
                })

            fields = ["file", "tool", "findings", "classification", "ground_truth", "correct", "error"]
            with open(semgrep_csv, "w", newline="", encoding="utf-8") as f:
                writer = csv.DictWriter(f, fieldnames=fields)
                writer.writeheader()
                writer.writerows(semgrep_records)

            print(f"{Fore.GREEN}[Semgrep] Results: {semgrep_csv.name} ({len(semgrep_records)} records)")

            for tool_label in ["semgrep_auto", "semgrep_security-audit", "semgrep_hlf"]:
                preds = {r["file"]: r["classification"] for r in semgrep_records if r["tool"] == tool_label}
                metrics = calculate_metrics(preds, GROUND_TRUTH)
                print(f"  {tool_label}: TP={metrics['TP']} FP={metrics['FP']} FN={metrics['FN']} TN={metrics['TN']} "
                      f"TPR={metrics['TPR']:.2f} TNR={metrics['TNR']:.2f}")
                all_results[tool_label] = {"metrics": metrics, "predictions": preds}

    # ── Phase B/C: LLM Models ────────────────────────────────────────
    if not args.skip_llm:
        for model_key in args.models:
            model_file = MODEL_FILES[model_key]
            model_path = MODELS_DIR / model_file
            if not model_path.exists():
                print(f"{Fore.RED}[Skip] Model not found: {model_file}")
                continue

            print(f"\n{Fore.YELLOW}{'='*60}")
            print(f"{Fore.YELLOW} Phase {'B' if model_key == 'qwen' else 'C'}: {model_key.upper()} P1 zero-shot × {args.runs} runs")
            print(f"{Fore.YELLOW}{'='*60}\n")

            llm = load_model(model_path)

            llm_csv = RESULTS_DIR / f"addon_{model_key}_{timestamp}.csv"
            fields = [
                "run", "file", "code_chars", "elapsed_s",
                "classification", "vuln_types", "ground_truth", "correct",
                "result",
            ]

            run_predictions = defaultdict(dict)  # run_num -> {file: classification}

            with open(llm_csv, "w", newline="", encoding="utf-8") as f:
                writer = csv.DictWriter(f, fieldnames=fields)
                writer.writeheader()

                for run_num in range(1, args.runs + 1):
                    print(f"\n{Fore.CYAN}--- Run {run_num}/{args.runs} ---")
                    for go_file in tqdm(go_files, desc=f"{model_key} R{run_num}"):
                        fname = go_file.name
                        code = go_file.read_text(encoding="utf-8")

                        start_t = time.time()
                        try:
                            result = audit_chaincode(llm, code, fname)
                        except Exception as e:
                            result = f"ERROR: {e}"
                        elapsed = round(time.time() - start_t, 3)

                        classification = classify_response(result)
                        vtypes = extract_vuln_types(result)
                        gt = GROUND_TRUTH.get(fname, "unknown")

                        run_predictions[run_num][fname] = classification

                        writer.writerow({
                            "run": run_num,
                            "file": fname,
                            "code_chars": len(code),
                            "elapsed_s": elapsed,
                            "classification": classification,
                            "vuln_types": "|".join(vtypes),
                            "ground_truth": gt,
                            "correct": classification == gt,
                            "result": result[:2000],
                        })
                        f.flush()

            del llm
            print(f"{Fore.GREEN}[{model_key}] Results: {llm_csv.name}")

            per_run_metrics = []
            for run_num in range(1, args.runs + 1):
                m = calculate_metrics(run_predictions[run_num], GROUND_TRUTH)
                per_run_metrics.append(m)
                print(f"  Run {run_num}: TP={m['TP']} FP={m['FP']} FN={m['FN']} TN={m['TN']} "
                      f"TPR={m['TPR']:.2f} TNR={m['TNR']:.2f}")

            tprs = [m["TPR"] for m in per_run_metrics]
            tnrs = [m["TNR"] for m in per_run_metrics]
            all_results[model_key] = {
                "per_run": per_run_metrics,
                "tpr_range": (min(tprs), max(tprs)),
                "tnr_range": (min(tnrs), max(tnrs)),
                "tpr_mean": round(sum(tprs) / len(tprs), 4),
                "tnr_mean": round(sum(tnrs) / len(tnrs), 4),
                "predictions_all_runs": dict(run_predictions),
            }

            # Sensitivity: U03=safe
            sens_metrics = []
            for run_num in range(1, args.runs + 1):
                sm = calculate_metrics(run_predictions[run_num], GROUND_TRUTH_U03_SAFE)
                sens_metrics.append(sm)
            all_results[f"{model_key}_u03safe"] = {
                "per_run": sens_metrics,
                "tpr_mean": round(sum(m["TPR"] for m in sens_metrics) / len(sens_metrics), 4),
                "tnr_mean": round(sum(m["TNR"] for m in sens_metrics) / len(sens_metrics), 4),
            }

    # ── Report generation ─────────────────────────────────────────────
    report_path = RESULTS_DIR / f"ADDON_VALIDATION_REPORT_{timestamp}.md"
    generate_report(report_path, all_results, args, run_start)

    # ── Meta JSON ─────────────────────────────────────────────────────
    meta = {
        "script": "20_run_addon_validation.py",
        "version": "1.0",
        "timestamp": run_start.isoformat(),
        "dataset_dir": str(DATASET_DIR),
        "n_files": len(go_files),
        "n_runs": args.runs,
        "models": args.models,
        "n_ctx": INFERENCE_PARAMS["n_ctx"],
        "temperature": INFERENCE_PARAMS["temperature"],
        "max_tokens": INFERENCE_PARAMS["max_tokens"],
        "ground_truth": GROUND_TRUTH,
        "ground_truth_u03_safe": GROUND_TRUTH_U03_SAFE,
        "platform": platform.platform(),
    }
    meta_path = RESULTS_DIR / f"addon_meta_{timestamp}.json"
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2, ensure_ascii=False)

    print(f"\n{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN} Validation complete!")
    print(f"{Fore.GREEN} Report: {report_path.name}")
    print(f"{Fore.GREEN} Meta: {meta_path.name}")
    print(f"{Fore.GREEN}{'='*60}")


def generate_report(path: Path, results: dict, args, run_start: datetime):
    lines = [
        "# Add-on External Validation Report",
        "",
        f"> Generated: {run_start.strftime('%Y-%m-%d %H:%M')}",
        f"> Files: 17 | Runs: {args.runs} | Models: {', '.join(args.models)}",
        f"> n_ctx: {INFERENCE_PARAMS['n_ctx']} | temp: {INFERENCE_PARAMS['temperature']}",
        "",
        "---",
        "",
        "## Ground Truth",
        "",
        f"- Vulnerable: 4 (U03, U13, U18, U20)",
        f"- Safe: 13",
        f"- Total: 17",
        "",
    ]

    if any(k.startswith("semgrep") for k in results):
        lines.extend([
            "## Semgrep Results",
            "",
            "| Tool | TP | FP | FN | TN | TPR | TNR |",
            "|:-----|:---|:---|:---|:---|:----|:----|",
        ])
        for tool in ["semgrep_auto", "semgrep_security-audit", "semgrep_hlf"]:
            if tool in results:
                m = results[tool]["metrics"]
                lines.append(
                    f"| {tool} | {m['TP']} | {m['FP']} | {m['FN']} | {m['TN']} | "
                    f"{m['TPR']:.2f} ({m['TPR_CI_95'][0]:.2f}-{m['TPR_CI_95'][1]:.2f}) | "
                    f"{m['TNR']:.2f} ({m['TNR_CI_95'][0]:.2f}-{m['TNR_CI_95'][1]:.2f}) |"
                )
        lines.extend(["", ""])

    for model_key in ["qwen", "llama"]:
        if model_key not in results:
            continue
        r = results[model_key]
        lines.extend([
            f"## {model_key.upper()} P1 Zero-Shot Results",
            "",
            f"| Run | TP | FP | FN | TN | TPR | TNR |",
            f"|:----|:---|:---|:---|:---|:----|:----|",
        ])
        for i, m in enumerate(r["per_run"], 1):
            lines.append(
                f"| {i} | {m['TP']} | {m['FP']} | {m['FN']} | {m['TN']} | "
                f"{m['TPR']:.2f} | {m['TNR']:.2f} |"
            )

        lines.extend([
            "",
            f"- **TPR mean**: {r['tpr_mean']:.2f} (range: {r['tpr_range'][0]:.2f}-{r['tpr_range'][1]:.2f})",
            f"- **TNR mean**: {r['tnr_mean']:.2f} (range: {r['tnr_range'][0]:.2f}-{r['tnr_range'][1]:.2f})",
        ])

        best_m = r["per_run"][0]
        lines.extend([
            f"- **TPR 95% CI** (Run 1): {best_m['TPR_CI_95'][0]:.2f}-{best_m['TPR_CI_95'][1]:.2f}",
            f"- **TNR 95% CI** (Run 1): {best_m['TNR_CI_95'][0]:.2f}-{best_m['TNR_CI_95'][1]:.2f}",
            "",
        ])

        sens_key = f"{model_key}_u03safe"
        if sens_key in results:
            sr = results[sens_key]
            lines.extend([
                f"### Sensitivity Analysis (U03=safe)",
                f"- TPR mean: {sr['tpr_mean']:.2f}",
                f"- TNR mean: {sr['tnr_mean']:.2f}",
                "",
            ])

    lines.extend([
        "---",
        "",
        "## Notes",
        "",
        "- Classifier: v2 (identical to 08_run_golisa_validation.py)",
        "- This is a **limited external validation** with small positive class (V=4).",
        "- Wide TPR CIs reflect positive class scarcity, not model failure.",
        "- Results should be framed as descriptive evidence, not precise TPR estimation.",
    ])

    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")


if __name__ == "__main__":
    main()
