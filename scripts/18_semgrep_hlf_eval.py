"""Evaluate HLF-specific Semgrep custom rules on micro-benchmark (15 files)."""
import json
import subprocess
import sys
import os

DATASET_DIR = os.path.join(os.path.dirname(__file__), "..", "02_resources", "dataset")
RULES_FILE = os.path.join(os.path.dirname(__file__), "..", "rules", "hlf_consensus.yml")

VULN_FILES = [
    "vuln_01_time.go", "vuln_01_b_interprocedural.go", "vuln_02_global.go",
    "vuln_03_goroutine.go", "vuln_04_map_iter.go", "vuln_04_b_nested_map.go",
    "vuln_05_phantom.go", "vuln_06_iterator_leak.go", "vuln_06_b_conditional_leak.go",
]
SAFE_FILES = [
    "safe_01_logging.go", "safe_02_local_var.go", "safe_03_map_read.go",
    "safe_04_math_rand.go", "safe_05_deterministic_time.go", "safe_06_external_lib.go",
]

def main():
    cmd = ["semgrep", "scan", "--config", RULES_FILE, DATASET_DIR, "--json"]
    result = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    data = json.loads(result.stdout.decode("utf-8", errors="replace"))
    findings = data.get("results", [])

    file_hits = {}
    for r in findings:
        fname = os.path.basename(r["path"])
        rule = r["check_id"].split(".")[-1]
        file_hits.setdefault(fname, set()).add(rule)

    print("=== VULNERABLE FILES ===")
    tp = 0
    for f in VULN_FILES:
        if f in file_hits:
            tp += 1
            print(f"  [HIT]  {f}: {sorted(file_hits[f])}")
        else:
            print(f"  [MISS] {f}")

    print("\n=== SAFE FILES ===")
    tn = 0
    fp_list = []
    for f in SAFE_FILES:
        if f in file_hits:
            fp_list.append(f)
            print(f"  [FP]   {f}: {sorted(file_hits[f])}")
        else:
            tn += 1
            print(f"  [TN]   {f}")

    print(f"\n--- Summary ---")
    print(f"TPR: {tp}/9")
    print(f"TNR: {tn}/6")
    print(f"Total findings: {len(findings)}")
    print(f"Missed vulns: {[f for f in VULN_FILES if f not in file_hits]}")
    print(f"False positives: {fp_list}")

if __name__ == "__main__":
    main()
