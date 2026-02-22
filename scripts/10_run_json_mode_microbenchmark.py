"""
10_run_json_mode_microbenchmark.py  (v1.0 -- 2026-02-10)

Run JSON mode prompt on 15 micro-benchmark files (Qwen2.5-Coder-7B).
Fills the missing data point: JSON mode was previously tested only on
GoLiSA Running_Examples, not on the micro-benchmark.

Usage:
  python scripts/10_run_json_mode_microbenchmark.py
"""

import csv
import json
import sys
import time
from datetime import datetime
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"
MODEL_FILE = "qwen2.5-coder-7b-instruct-q4_k_m.gguf"

# Ground truth
GROUND_TRUTH = {
    "vuln_01_time.go": "vulnerable",
    "vuln_02_global.go": "vulnerable",
    "vuln_03_goroutine.go": "vulnerable",
    "vuln_04_map_iter.go": "vulnerable",
    "vuln_05_phantom.go": "vulnerable",
    "vuln_06_iterator_leak.go": "vulnerable",
    "vuln_01_b_interprocedural.go": "vulnerable",
    "vuln_04_b_nested_map.go": "vulnerable",
    "vuln_06_b_conditional_leak.go": "vulnerable",
    "safe_01_logging.go": "safe",
    "safe_02_local_var.go": "safe",
    "safe_03_map_read.go": "safe",
    "safe_04_math_rand.go": "safe",
    "safe_05_deterministic_time.go": "safe",
    "safe_06_external_lib.go": "safe",
}

PROMPT_JSON = (
    "You are a Hyperledger Fabric chaincode vulnerability detection system.\n"
    "Analyze the following Go chaincode and output ONLY valid JSON.\n"
    "Do NOT write any text outside the JSON object.\n\n"
    "Focus on consensus-layer vulnerabilities that cause endorsement mismatch:\n"
    "- Non-deterministic operations (time.Now, math/rand, map iteration, goroutine race)\n"
    "- Global/shared mutable state across invocations\n"
    "- Channel-based goroutine nondeterminism\n"
    "- Phantom reads (read-after-write conflicts)\n\n"
    "Output format:\n"
    "{\n"
    '  "is_vulnerable": true or false,\n'
    '  "vulnerabilities": [\n'
    '    {"type": "string", "severity": "Critical|High|Medium|Low", '
    '"location": "function name", "description": "short explanation"}\n'
    "  ]\n"
    "}\n\n"
    "If no consensus-layer vulnerabilities exist, set is_vulnerable to false "
    "and vulnerabilities to an empty array."
)


def classify_json_response(response):
    if not response or response.startswith("ERROR:"):
        return "error"
    try:
        resp = response.strip()
        start = resp.find("{")
        end = resp.rfind("}")
        if start >= 0 and end > start:
            json_str = resp[start:end+1]
            data = json.loads(json_str)
            if data.get("is_vulnerable", False):
                return "vulnerable"
            else:
                return "safe"
    except (json.JSONDecodeError, KeyError, TypeError):
        pass
    return "unknown"


def main():
    from llama_cpp import Llama

    timestamp = datetime.now().strftime("%y%m%d_%H%M")
    print(f"\n{'='*70}")
    print(f"  JSON Mode on Micro-benchmark (15 files)")
    print(f"  Qwen2.5-Coder-7B, {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"{'='*70}")

    model_path = MODELS_DIR / MODEL_FILE
    print(f"  Loading model: {MODEL_FILE}")
    llm = Llama(
        model_path=str(model_path),
        n_gpu_layers=-1,
        n_ctx=4096,
        verbose=False,
    )
    print(f"  Model loaded.\n")

    go_files = sorted(DATASET_DIR.glob("*.go"))
    print(f"  Dataset: {len(go_files)} files\n")

    results = []
    total_start = time.time()

    for go_file in go_files:
        fname = go_file.name
        code = go_file.read_text(encoding="utf-8")
        expected = GROUND_TRUTH.get(fname, "unknown")

        sys.stdout.write(f"  {fname:<40} ")
        sys.stdout.flush()

        messages = [
            {"role": "system", "content": PROMPT_JSON},
            {"role": "user", "content": f"```go\n{code}\n```"},
        ]
        t0 = time.time()
        try:
            output = llm.create_chat_completion(
                messages=messages,
                temperature=0.1,
                max_tokens=2048,
            )
            response = output["choices"][0]["message"]["content"].strip()
        except Exception as e:
            response = f"ERROR: {e}"
        elapsed = time.time() - t0

        classification = classify_json_response(response)
        match = classification == expected
        mark = "O" if match else "X"

        sys.stdout.write(f"expected={expected:<12} got={classification:<12} [{mark}] ({elapsed:.1f}s)\n")

        results.append({
            "file": fname,
            "expected": expected,
            "classification": classification,
            "match": match,
            "elapsed_s": round(elapsed, 3),
            "response": response,
        })

    total_elapsed = time.time() - total_start

    # Summary
    vuln_files = [r for r in results if r["expected"] == "vulnerable"]
    safe_files = [r for r in results if r["expected"] == "safe"]
    tp = sum(1 for r in vuln_files if r["classification"] == "vulnerable")
    tn = sum(1 for r in safe_files if r["classification"] == "safe")

    print(f"\n{'='*70}")
    print(f"  RESULTS SUMMARY")
    print(f"{'='*70}")
    print(f"  TPR: {tp}/{len(vuln_files)} ({tp/len(vuln_files)*100:.1f}%)")
    print(f"  TNR: {tn}/{len(safe_files)} ({tn/len(safe_files)*100:.1f}%)")
    print(f"  Total time: {total_elapsed:.1f}s ({total_elapsed/len(go_files):.3f}s/file)")

    # Save CSV
    csv_path = RESULTS_DIR / f"json_mode_microbenchmark_{timestamp}.csv"
    with open(csv_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=["file", "expected", "classification", "match", "elapsed_s", "response"])
        writer.writeheader()
        writer.writerows(results)
    print(f"\n  CSV saved: {csv_path.name}")

    # Save meta
    meta = {
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "model": MODEL_FILE,
        "prompt": "json_mode",
        "dataset": "micro-benchmark (15 files)",
        "n_ctx": 4096,
        "temperature": 0.1,
        "tpr": f"{tp}/{len(vuln_files)}",
        "tnr": f"{tn}/{len(safe_files)}",
        "total_time_s": round(total_elapsed, 1),
        "avg_time_s": round(total_elapsed / len(go_files), 3),
    }
    meta_path = RESULTS_DIR / f"json_mode_microbenchmark_{timestamp}.meta.json"
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2, ensure_ascii=False)
    print(f"  Meta saved: {meta_path.name}")


if __name__ == "__main__":
    main()
