#!/usr/bin/env python3
"""
Prompt dev sanity check for GoLiSA main benchmark.
Runs Qwen on 10 dev files × 2 runs (temp=0.0) to verify:
1. Output format stability (3-line LABEL/CLASSES/EVIDENCE)
2. Determinism at temp=0.0
3. Parser compatibility
"""

import json
import os
import re
import time
from pathlib import Path
from datetime import datetime

PROJ = Path(__file__).parent.parent
MODEL_PATH = PROJ / "02_resources" / "models" / "qwen2.5-coder-7b-instruct-q4_k_m.gguf"
DEV_FILES_PATH = PROJ / "06_addon_validation" / "golisa_mining" / "dev_files.json"
OUTPUT_DIR = PROJ / "06_addon_validation" / "golisa_mining"

PROMPT_TEMPLATE = """You are auditing a single Go source file from a Hyperledger Fabric chaincode project.

This is a CLOSED-SET classification task.
Consider ONLY the six consensus-critical nondeterminism classes listed below.
Ignore all other bug types, security issues, or code quality concerns.

A file is VULNERABLE only if at least one class below can influence:
- ledger writes (PutState, DelState, PutPrivateData, DelPrivateData), or
- the returned proposal response, or
- control flow that determines such writes or responses.

Treat shim-style calls (stub.PutState) and contractapi-style calls (ctx.GetStub().PutState) as equivalent sinks.

Do NOT report: access control, input validation, key management, logging-only uses, comments, filenames, test/dead code, constant globals, slice/array iteration, user-provided timestamps, or any issue outside the six classes.

Allowed classes:
1. TIME_NOW — time.Now/time.Since/time.Until influencing a sink
2. GOROUTINE — go statement or concurrent work influencing a sink
3. MAP_ITERATION — iteration over a Go map whose order influences a sink (NOT slices/arrays)
4. PHANTOM_READ — rich-query results (GetQueryResult, GetPrivateDataQueryResult, GetQueryResultWithPagination) used to decide writes (NOT GetStateByRange alone)
5. ITERATOR_LEAK — ledger/query iterator not closed on all paths
6. GLOBAL_MUTABLE_STATE — mutable package-level state influencing a sink

Return EXACTLY these three lines and nothing else:
LABEL: VULNERABLE | SAFE
CLASSES: NONE | <comma-separated from allowed set>
EVIDENCE: <one line, max 30 words>

Go source:
```go
{CODE}
```"""

ALLOWED_CLASSES = {
    "TIME_NOW", "GOROUTINE", "MAP_ITERATION",
    "PHANTOM_READ", "ITERATOR_LEAK", "GLOBAL_MUTABLE_STATE"
}


def parse_response(text):
    """Parse the 3-line response format."""
    result = {
        "raw": text.strip(),
        "label": None,
        "classes": None,
        "evidence": None,
        "parse_ok": False,
        "errors": [],
    }

    lines = [l.strip() for l in text.strip().split('\n') if l.strip()]

    # Find LABEL line
    label_line = None
    classes_line = None
    evidence_line = None

    for line in lines:
        if line.upper().startswith("LABEL:"):
            label_line = line
        elif line.upper().startswith("CLASSES:"):
            classes_line = line
        elif line.upper().startswith("EVIDENCE:"):
            evidence_line = line

    if not label_line:
        result["errors"].append("LABEL line not found")
    else:
        val = label_line.split(":", 1)[1].strip().upper()
        if val in ("VULNERABLE", "SAFE"):
            result["label"] = val
        else:
            result["errors"].append(f"Invalid LABEL value: {val}")

    if not classes_line:
        result["errors"].append("CLASSES line not found")
    else:
        val = classes_line.split(":", 1)[1].strip()
        if val.upper() == "NONE":
            result["classes"] = []
        else:
            parsed = [c.strip().upper() for c in val.split(",")]
            invalid = [c for c in parsed if c not in ALLOWED_CLASSES]
            if invalid:
                result["errors"].append(f"Invalid classes: {invalid}")
            result["classes"] = [c for c in parsed if c in ALLOWED_CLASSES]

    if not evidence_line:
        result["errors"].append("EVIDENCE line not found")
    else:
        result["evidence"] = evidence_line.split(":", 1)[1].strip()

    # Cross-validation: CLASSES priority rule
    if result["label"] and result["classes"] is not None:
        if result["label"] == "VULNERABLE" and len(result["classes"]) == 0:
            result["warnings"] = result.get("warnings", [])
            result["warnings"].append("VULNERABLE but CLASSES=NONE → override to SAFE (CLASSES priority)")
            result["label"] = "SAFE"  # CLASSES priority rule
        if result["label"] == "SAFE" and len(result["classes"]) > 0:
            result["warnings"] = result.get("warnings", [])
            result["warnings"].append(f"SAFE but CLASSES={result['classes']} → override to VULNERABLE (CLASSES priority)")
            result["label"] = "VULNERABLE"  # CLASSES priority rule

    if not result["errors"]:
        result["parse_ok"] = True

    # Count extra lines
    extra_lines = len(lines) - 3
    if extra_lines > 0:
        result["errors"].append(f"{extra_lines} extra lines in output")

    return result


def main():
    from llama_cpp import Llama

    dev_files = json.load(open(DEV_FILES_PATH, encoding='utf-8'))
    print(f"Dev files: {len(dev_files)}")
    print(f"Model: {MODEL_PATH.name}")

    # Load model
    print("Loading Qwen model...")
    llm = Llama(
        model_path=str(MODEL_PATH),
        n_ctx=16384,
        n_gpu_layers=-1,
        verbose=False,
    )
    print("Model loaded.")

    results = []
    N_RUNS = 2  # 2 runs at temp=0.0 for determinism check

    for file_idx, filepath in enumerate(dev_files):
        filename = Path(filepath).name
        repo = Path(filepath).parent.name

        with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
            code = f.read()

        prompt = PROMPT_TEMPLATE.replace("{CODE}", code)
        print(f"\n[{file_idx+1}/{len(dev_files)}] {repo}/{filename} ({len(code)} chars)")

        file_results = []
        for run in range(N_RUNS):
            t0 = time.time()
            response = llm.create_chat_completion(
                messages=[{"role": "user", "content": prompt}],
                temperature=0.0,
                max_tokens=256,
                top_p=1.0,
            )
            elapsed = time.time() - t0

            raw_text = response["choices"][0]["message"]["content"]
            parsed = parse_response(raw_text)
            parsed["run"] = run + 1
            parsed["elapsed_s"] = round(elapsed, 2)
            parsed["repo"] = repo
            parsed["filename"] = filename
            parsed["code_chars"] = len(code)

            status = "OK" if parsed["parse_ok"] else f"ERR: {parsed['errors']}"
            print(f"  Run {run+1}: LABEL={parsed['label']} CLASSES={parsed['classes']} [{elapsed:.1f}s] {status}")
            file_results.append(parsed)

        # Determinism check
        if N_RUNS == 2:
            r1, r2 = file_results[0], file_results[1]
            deterministic = (r1["raw"] == r2["raw"])
            print(f"  Determinism: {'YES' if deterministic else 'NO (outputs differ)'}")
            for fr in file_results:
                fr["deterministic"] = deterministic

        results.extend(file_results)

    # === Summary ===
    print(f"\n{'='*60}")
    print("SANITY CHECK SUMMARY")
    print(f"{'='*60}")

    total_runs = len(results)
    parse_ok = sum(1 for r in results if r["parse_ok"])
    parse_fail = total_runs - parse_ok
    deterministic_files = sum(1 for i in range(0, len(results), N_RUNS)
                             if results[i].get("deterministic", False))

    print(f"Total runs: {total_runs}")
    print(f"Parse OK: {parse_ok}/{total_runs}")
    print(f"Parse FAIL: {parse_fail}/{total_runs}")
    print(f"Deterministic files: {deterministic_files}/{len(dev_files)}")

    if parse_fail > 0:
        print(f"\nParse failures:")
        for r in results:
            if not r["parse_ok"]:
                print(f"  {r['repo']}/{r['filename']} Run{r['run']}: {r['errors']}")
                print(f"    Raw: {r['raw'][:200]}")

    # Errors summary
    all_errors = []
    for r in results:
        all_errors.extend(r["errors"])
    if all_errors:
        print(f"\nAll errors:")
        for e in set(all_errors):
            count = all_errors.count(e)
            print(f"  [{count}x] {e}")

    # Label distribution
    labels = [r["label"] for r in results if r["label"]]
    print(f"\nLabel distribution: VULNERABLE={labels.count('VULNERABLE')}, SAFE={labels.count('SAFE')}")

    # Save results
    timestamp = datetime.now().strftime("%y%m%d_%H%M")
    report_path = OUTPUT_DIR / f"dev_sanity_check_{timestamp}.json"
    with open(report_path, 'w', encoding='utf-8') as f:
        json.dump({
            "timestamp": timestamp,
            "model": MODEL_PATH.name,
            "n_ctx": 16384,
            "temperature": 0.0,
            "n_runs": N_RUNS,
            "dev_files": len(dev_files),
            "total_runs": total_runs,
            "parse_ok": parse_ok,
            "parse_fail": parse_fail,
            "deterministic_files": deterministic_files,
            "results": results,
        }, f, indent=2, ensure_ascii=False)
    print(f"\nReport saved: {report_path}")

    # Final verdict
    if parse_fail == 0 and deterministic_files == len(dev_files):
        print(f"\n*** VERDICT: PROMPT FORMAT STABLE + DETERMINISTIC → READY TO FREEZE ***")
    elif parse_fail == 0:
        print(f"\n*** VERDICT: PROMPT FORMAT STABLE, DETERMINISM PARTIAL → CHECK NON-DETERMINISTIC FILES ***")
    else:
        print(f"\n*** VERDICT: PARSE FAILURES DETECTED → PROMPT NEEDS REVISION ***")


if __name__ == "__main__":
    main()
