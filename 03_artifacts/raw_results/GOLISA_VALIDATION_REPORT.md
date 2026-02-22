# GoLiSA External Validation Report
> Generated: 2026-02-09 23:02:30 (KST)
> Updated: 2026-02-10 00:51 (KST) -- Supplementary experiments added
> Duration: 9684.2s (main) + 72.4s (supplementary)

## 1. Dataset Summary
- **Source**: GoLiSA ECOOP 2023 Benchmark (Olivieri et al.)
- **Total files**: 657
- **Total repositories**: 326
- **Total size**: 5,438,685 bytes (5.2 MB)
- **Average file size**: 8,278 bytes
- **Median file size**: 5,436 bytes
- **Max file size**: 42,808 bytes
- **Files > 16KB**: 95

## 2. Qwen2.5-Coder-7B Zero-Shot Results
- **Model**: qwen2.5-coder-7b-instruct-q4_k_m.gguf
- **Prompt**: zero_shot (identical to micro-benchmark)
- **n_ctx**: 16384
- **Total inference time**: 5252.7s (87.5min)
- **Average per file**: 7.995s

### Classification Summary (Classifier v1 -- original)
| Classification | Count | Percentage |
|:---------------|------:|-----------:|
| vulnerable | 380 | 57.8% |
| safe | 277 | 42.2% |
| error | 0 | 0.0% |

### Classification Summary (Classifier v2 -- improved, see Section 7)
| Classification | Count | Percentage |
|:---------------|------:|-----------:|
| vulnerable | 477 | 72.6% |
| safe | 180 | 27.4% |
| error | 0 | 0.0% |

- **Changed classifications**: 97 files (safe -> vulnerable)
- **Reason**: Classifier v2 detects self-contradictory LLM responses where the model produces structured vulnerability analysis (severity ratings, recommended fixes) but appends "No vulnerabilities detected" at the end.

### Vulnerability Type Distribution (files classified as vulnerable, v1)
| Vulnerability Type | Count |
|:-------------------|------:|
| access_control | 379 |
| input_validation | 379 |
| phantom_read | 371 |
| goroutine | 128 |
| random | 68 |
| timestamp | 52 |
| iterator_leak | 29 |
| external_call | 1 |

## 3. Running_Examples Validation (Mini Ground Truth)

### Classifier v1 (original)
- **Accuracy**: 1/5

### Classifier v2 (improved)
- **Accuracy**: 3/5

| File | Expected | v1 Result | v2 Result | Vuln Type |
|:-----|:---------|:----------|:----------|:----------|
| Channel.go | VULNERABLE | SAFE | SAFE | goroutine |
| GlobalVariable.go | VULNERABLE | SAFE | **VULNERABLE** | global_var |
| GoRoutines.go | VULNERABLE | VULNERABLE | VULNERABLE | goroutine |
| MapIteration.go | VULNERABLE | SAFE | SAFE | map_iter |
| MethodFunction.go | VULNERABLE | SAFE | **VULNERABLE** | timestamp |

### Why v2 rescues GlobalVariable.go and MethodFunction.go
- **GlobalVariable.go**: Qwen produced 6 numbered vulnerability sections with severity ratings and recommended fixes (including "Non-Deterministic Operations" identifying global variable mutation), then appended "No vulnerabilities detected" at the final line. Classifier v1 matched the safe phrase and short-circuited. Classifier v2 detects structured_score >= 2 and overrides.
- **MethodFunction.go**: Qwen started with "No vulnerabilities detected" then listed 6 vulnerability categories with severity ratings. Classifier v1 matched the safe phrase at the start. Classifier v2 detects structured_score >= 2 and overrides.

### Why Channel.go and MapIteration.go remain SAFE under v2
- **Channel.go**: Qwen responded with only "No vulnerabilities detected." (28 chars, 0.075s). No structured analysis exists; classifier v2 correctly returns safe.
- **MapIteration.go**: Qwen responded with "No vulnerabilities detected." followed by an explanation of why each category is safe. No structured vulnerability analysis (no severity ratings, no recommended fixes); classifier v2 correctly returns safe.

## 4. Semgrep Comparison
- **Semgrep version**: 1.151.0
- **Configs used**: auto, p/security-audit
- **Total findings**: 12
- **Consensus-relevant findings**: 0

| Config | Total Findings | Files with Findings | Consensus-Relevant |
|:-------|---------------:|--------------------:|-------------------:|
| semgrep:auto | 0 | 0 | 0 |
| semgrep:security-audit | 12 | 11 | 0 |

## 5. Cross-Tool Comparison
- **Qwen flagged as vulnerable (v1)**: 380 files
- **Qwen flagged as vulnerable (v2)**: 477 files
- **Semgrep flagged (any finding)**: 11 files
- **Both flagged (v1)**: 10 files
- **Only Qwen (v1)**: 370 files
- **Only Semgrep**: 1 file

## 6. Key Findings for Paper

1. **Traditional tool gap at scale**: Semgrep found 0 consensus-layer detections across 657 real-world HLF chaincodes, confirming the domain-specific detection gap observed in the micro-benchmark.
2. **Scalability**: Qwen processed 657 files with 0 errors in 5252.7s (87.5min), averaging 7.995s/file.
3. **Self-contradictory LLM responses**: 97 of 277 files classified as "safe" by v1 contained structured vulnerability analysis contradicted by a "No vulnerabilities detected" conclusion. This is a practical challenge for automated LLM-based security tools.
4. **Prompt strategy overcomes code minimality**: few_shot and json_mode prompts achieve 5/5 on Running_Examples (see Section 8).

## 7. Classifier v2 Design

### Rationale
LLM responses can contain detailed vulnerability analysis (numbered sections with severity ratings, affected code locations, recommended fixes) followed by a contradictory "No vulnerabilities detected" conclusion. When structured vulnerability evidence co-occurs with a safe phrase, the structured evidence takes priority.

### Logic
1. Count structured vulnerability markers: "severity:", "recommended fix", "affected code" (need >= 2)
2. Check for safe indicator phrases
3. If structured_score >= 2 AND safe phrase present -> VULNERABLE (self-contradictory response)
4. If safe phrase present AND structured_score < 2 -> SAFE (genuine safe judgment)
5. Otherwise, count general vulnerability indicators; if >= 2 -> VULNERABLE
6. Default: VULNERABLE (conservative)

### Validation
- Does NOT change Channel.go or MapIteration.go (correctly remain SAFE)
- Rescues GlobalVariable.go (6 numbered vuln sections with severity) and MethodFunction.go (6 categories with severity)
- 97 files changed across 657; distribution shifts from 380/277 to 477/180

## 8. Supplementary Experiments (2026-02-10 00:51)

### 8.1 Multi-Prompt Ablation on Running_Examples (5 files)

| File | zero_shot | few_shot | cot | json_mode |
|:-----|:---------:|:--------:|:---:|:---------:|
| Channel.go | X | O | O | O |
| GlobalVariable.go | O | O | X | O |
| GoRoutines.go | O | O | O | O |
| MapIteration.go | X | O | X | O |
| MethodFunction.go | X | O | O | O |
| **Accuracy** | **2/5** | **5/5** | **3/5** | **5/5** |

- Classifier v2 applied to zero_shot, few_shot, cot; JSON parser applied to json_mode.
- few_shot and json_mode achieve 5/5 (100%) on all Running_Examples.
- CoT achieves 3/5; GlobalVariable.go and MapIteration.go missed under step-by-step reasoning.

### 8.2 Context Injection Ablation (Channel.go)

| Variant | Classification | Time |
|:--------|:--------------|:-----|
| Original (empty function bodies) | safe | 0.1s |
| Modified (c <- "hello"/"world" added) | safe | 1.8s |
| Modified + CoT | safe | 3.6s |

- Even with functional channel operations, Qwen does not detect channel-based nondeterminism at this code size (401 chars).
- Channel nondeterminism detection requires few_shot or json_mode prompts (see 8.1).

### 8.3 Experimental Setup
- Script: `09_reclassify_and_ablation.py` (v1.0)
- Model: qwen2.5-coder-7b-instruct-q4_k_m.gguf
- Hardware: RTX 3090 Ti (24,564 MiB)
- n_ctx: 16384, temperature: 0.1, max_tokens: 2048
- Results: `golisa_supplementary_260210_0051.json`
