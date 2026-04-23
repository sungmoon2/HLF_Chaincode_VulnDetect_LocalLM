# Inference Contract — Frozen Manifest
> Version: 1.0 | Date: 2026-04-23 | Status: **FROZEN**

## 1. Qwen Inference

### Model
| Item | Value |
|:-----|:------|
| Model | Qwen2.5-Coder-7B-Instruct |
| Quantization | Q4_K_M |
| File | qwen2.5-coder-7b-instruct-q4_k_m.gguf |
| SHA-256 | `509287f78cb4d4cf6b3843734733b914b2c158e43e22a7f4bf5e963800894d3c` |
| Runtime | llama-cpp-python 0.3.16 |
| n_ctx | 16384 |
| n_gpu_layers | -1 (all layers on GPU) |
| GPU | NVIDIA GeForce RTX 3090 Ti (24564 MiB) |
| Driver | 591.86 |

### Prompt Template (FROZEN)
```
System: You are a Hyperledger Fabric Security Expert. Analyze the following Go chaincode for security vulnerabilities. Focus on: access control issues, input validation flaws, read-after-write conflicts (phantom reads), private data leakage, non-deterministic operations, and insecure key management. For each vulnerability found, provide:
1. Vulnerability type
2. Severity (Critical/High/Medium/Low)
3. Affected code location (function name and line reference)
4. Description of the issue
5. Recommended fix
If no vulnerabilities are found, state 'No vulnerabilities detected.'

User: Analyze this Hyperledger Fabric chaincode file '{filename}' for security vulnerabilities:

```go
{code}
```
```

### Inference Parameters
| Parameter | Deterministic (Phase 6) | Robustness (Phase 7) |
|:----------|:-----------------------|:---------------------|
| temperature | 0.0 | 0.1 |
| max_tokens | 2048 | 2048 |
| seed | (implicit via temp=0.0) | {1, 2, 3, 4, 5} |
| per-file session | fresh (no carry-over) | fresh (no carry-over) |

### Input
- Comment-stripped Go source code (via strip_go_comments.exe)
- Same stripped input as labeling (consistency)
- No filename in code body (filename only in prompt framing)

### Parser Logic (FROZEN)
```python
def classify_response(response: str) -> str:
    resp_lower = response.lower()
    
    # Step 1: Check explicit safe indicators
    safe_indicators = [
        "no vulnerabilities detected", "no vulnerabilities found",
        "no security vulnerabilities", "no significant vulnerabilities",
        "no vulnerabilities were found", "no vulnerabilities were detected",
    ]
    for indicator in safe_indicators:
        if indicator in resp_lower:
            # Check for "however" reversal
            idx = resp_lower.find(indicator)
            after = resp_lower[idx + len(indicator):]
            if any(kw in after[:200] for kw in ["however", "but ", "although"]):
                vuln_check = ["vulnerability", "vulnerable", "severity:", "recommended fix"]
                if sum(1 for v in vuln_check if v in after) >= 2:
                    return "vulnerable"
            return "safe"
    
    # Step 2: Count vulnerability indicators (≥2 → vulnerable)
    vuln_indicators = [
        "vulnerability type", "severity:", "recommended fix",
        "non-deterministic", "nondeterministic", "phantom read",
        "global variable", "goroutine", "race condition",
        "map iteration", "iterator leak", "putstate",
    ]
    if sum(1 for ind in vuln_indicators if ind in resp_lower) >= 2:
        return "vulnerable"
    
    return "safe"
```

### Scoring Rule
```
prediction := classify_response(qwen_output)
ground_truth := benchmark_freeze[file].verdict  # "VULNERABLE" or "SAFE"
TP := prediction == "vulnerable" AND ground_truth == "VULNERABLE"
FP := prediction == "vulnerable" AND ground_truth == "SAFE"
FN := prediction == "safe" AND ground_truth == "VULNERABLE"
TN := prediction == "safe" AND ground_truth == "SAFE"
```

## 2. Semgrep Inference

### Tool
| Item | Value |
|:-----|:------|
| Semgrep | 1.151.0 |
| Edition | OSS |
| Rules | rules/hlf_consensus.yml (custom) |

### Rules (FROZEN)
| Rule ID | Target Pattern |
|:--------|:--------------|
| hlf-time-now-in-chaincode | `time.Now()` |
| hlf-goroutine-in-chaincode | `go func(...) { ... }(...)` |
| hlf-map-iteration-putstate | `for $K, $V := range $MAP { ... PutState(...) ... }` |
| hlf-iterator-leak | Iterator without defer Close() |
| hlf-rich-query-in-chaincode | `GetQueryResult(...)` |

### File-Level Reduction Rule
```
prediction := "vulnerable" if ANY frozen rule ID hits the file
prediction := "safe" if NO frozen rule hits
```

### Input
- Same stripped/neutralized .go files as Qwen input (identical benchmark files)

## 3. Execution Protocol

1. **Per-file fresh session**: Each file is processed independently. No context carry-over.
2. **Identical input**: Both Qwen and Semgrep receive the same stripped source files.
3. **Blind evaluation**: Neither tool sees the other's output or the ground truth labels.
4. **Deterministic first**: Phase 6 uses temp=0.0. Phase 7 (robustness) uses temp=0.1 × 5 runs.
5. **Runtime recording**: Wall-clock time per file, peak VRAM, total elapsed.

## 4. Freeze Declaration

This inference contract is **FROZEN as of 2026-04-23**.
No modifications to prompt, parser, scoring rule, model, or Semgrep rules permitted.
Any deviations must be documented as protocol amendments with justification.
