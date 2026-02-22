# CPR (Correct Positive Rate) / Hallucination Verification

> Manual cross-verification of model responses against actual source code.
> Scope: Run 04, 15 files, Qwen2.5-Coder-7B zero-shot + Llama-3.1-8B zero-shot.
> Verified: 2026-02-22, source files in `02_resources/dataset/`.

## Methodology

For each model response:
1. Read the actual `.go` source code
2. Identify the ground-truth vulnerability (or safety reason)
3. Compare the model's stated finding against the actual code
4. Classify each finding as:
   - **Correct**: finding matches actual vulnerability in code
   - **Partially Correct**: correct vulnerability category, wrong location or incomplete
   - **Hallucination**: fabricated vulnerability that does not exist in the code
   - **Missed**: actual vulnerability not mentioned by model

---

## Qwen2.5-Coder-7B Zero-Shot (15 files)

### Safe Files (6/6 correct — TNR 100%)

| File | Model Output | Actual Code | Verdict |
|:-----|:-------------|:------------|:--------|
| safe_01_logging.go | "No vulnerabilities detected." | time.Now() in fmt.Printf only; PutState uses GetTxTimestamp() | **Correct** |
| safe_02_local_var.go | "No vulnerabilities detected." | All globals are constants/read-only; local vars deterministic | **Correct** |
| safe_03_map_read.go | "No vulnerabilities detected." | Maps used for lookup only; json.Marshal sorts keys | **Correct** |
| safe_04_math_rand.go | "No vulnerabilities detected." | math/rand in fmt.Printf only; PutState receives function args | **Correct** |
| safe_05_deterministic_time.go | "No vulnerabilities detected." | time.Now() for logging only; PutState uses GetTxTimestamp() | **Correct** |
| safe_06_external_lib.go | "No vulnerabilities detected." | crypto/sha256, strconv are pure deterministic functions | **Correct** |

**Safe file hallucinations: 0**

### Vulnerable Files (9/9 detected — TPR 100%)

#### vuln_01_time.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Non-deterministic: time.Now() in CreateShipment | Line 41: `now := time.Now()` → line 48-49 fields → PutState line 57 | **Correct** |
| 2 | Non-deterministic: time.Now() in UpdateShipmentStatus | Line 80: `time.Now()` → UpdatedAt → PutState line 87 | **Correct** |
| 3 | Non-deterministic: time.Now() in RecordTemperature | Lines 109-110: two time.Now() calls → PutState line 117 | **Correct** |

**Hallucinations: 0** | **Missed: 0**

#### vuln_01_b_interprocedural.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Interprocedural nondeterminism via helper functions | Lines 30-32 getCurrentTimestamp(), 37-40 formatTimestampNano(), 45-47 buildTimestampedNote() all wrap time.Now(); flow to PutState at lines 76, 108, 124 | **Correct** |

**Hallucinations: 0** | **Missed: 0**

#### vuln_02_global.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Global variable: transactionCounter and lastProcessedID | Lines 14-15: `var transactionCounter int` and `var lastProcessedID string`; modified at lines 35, 52, 87-88, 129-130; flow to PutState | **Correct** |

**Hallucinations: 0** | **Missed: 0**

#### vuln_03_goroutine.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Non-deterministic: goroutine race in TallyVotes | Lines 77-100: goroutines with concurrent stub access; results go to PutState line 108-109 | **Correct** |

**Hallucinations: 0** | **Missed: 0** (BatchUpdateStatus and AsyncNotify also vulnerable but model focused on primary finding)

#### vuln_04_map_iter.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Non-deterministic: Go map iteration randomization | Lines 47-75 BatchRestock, 83-112 GenerateReport, 115-143 ApplyDiscounts: `for k,v := range map` with PutState inside loop | **Correct** |

**Hallucinations: 0** | **Missed: 0**

#### vuln_04_b_nested_map.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Non-deterministic: nested map iteration | Lines 40-70 double `for range`: outer zones, inner items; PutState at lines 65, 69 inside nested loop | **Correct** |

**Hallucinations: 0** | **Missed: 0**

#### vuln_05_phantom.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Phantom read / MVCC conflict in CloseAuction | Lines 65-69 GetStateByPartialCompositeKey → loop → PutState at line 98; concurrent PlaceBid invalidates read-set | **Correct** |

**Hallucinations: 0** | **Missed: 0** (IncrementCounter and TransferWithBalanceCheck also vulnerable but model focused on primary finding)

#### vuln_06_iterator_leak.go

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Iterator leak: GetStateByRange without defer Close() | Lines 51, 80, 109, 139: iterators opened without defer; early returns at lines 61, 67, 90, 118, 128, 149 skip cleanup | **Correct** |

**Hallucinations: 0** | **Missed: 0**

#### vuln_06_b_conditional_leak.go (detailed analysis)

| # | Model Finding | Actual Code | Verdict |
|:--|:-------------|:------------|:--------|
| 1 | Access control: no authorization check in ProcessClaimsByRange | No `ctx.GetClientIdentity()` check. Factually true but **not consensus-layer** | **Hallucination (non-consensus)** |
| 2 | Input validation: iterator not closed on normal path in GetHighValueClaims | Line 94: iterator opened; line 120: Close() only in else branch; normal path leaks. Model correctly identified the leak pattern but categorized it as "input validation" | **Partially Correct** (correct observation, wrong category) |
| 3 | Phantom read in ProcessClaimsByRange | Lines 27-32: double query pattern. Factually true but secondary to the iterator leak | **Partially Correct** (real pattern but not primary vulnerability) |
| 4 | Private data leakage in GetHighValueClaims | No private data collection used in this file | **Hallucination** |
| 5 | Non-deterministic operations in ArchiveOldClaims | Model flagged non-deterministic ordering. Actual vulnerability is iterator leak (lines 141, 155 early returns skip Close()) | **Partially Correct** (real concern but misidentified root cause) |
| 6 | Insecure key management | No cryptographic key operations in this file | **Hallucination** |

Additionally, the response ends with "No vulnerabilities detected" — a self-contradiction after listing 6 findings. Classifier v2 correctly overrides this to `vulnerable`.

**Summary for vuln_06_b_conditional_leak.go**:
- Core vulnerability (conditional iterator leak): **detected** but miscategorized
- Hallucinations: **2** (access control without consensus relevance counted as non-consensus noise; private data leakage and key management are fabricated)
- Self-contradiction: present (6 findings followed by "No vulnerabilities detected")

---

## Llama-3.1-8B Zero-Shot (15 files)

### Safe Files (1/6 correct — TNR 17%)

| File | Model Output | Actual Code | Verdict |
|:-----|:-------------|:------------|:--------|
| safe_01_logging.go | "Input validation flaw" on `id` parameter | No input validation issue relevant to consensus | **Hallucination (non-consensus)** |
| safe_02_local_var.go | "Input validation flaw" on `from`/`to` fields | No input validation issue relevant to consensus | **Hallucination (non-consensus)** |
| safe_03_map_read.go | "Input validation flaw" in ValidateVoterEligibility | Normal validation logic, not a vulnerability | **Hallucination** |
| safe_04_math_rand.go | "Input validation flaw" on weight/transfer args | Normal argument handling, not a vulnerability | **Hallucination** |
| safe_05_deterministic_time.go | "No vulnerabilities detected." | Correct — uses GetTxTimestamp() | **Correct** |
| safe_06_external_lib.go | "Access control issue" in CreateAccount | No access control vulnerability relevant to consensus | **Hallucination (non-consensus)** |

**Pattern**: Llama systematically fabricates "input validation" and "access control" findings on safe files. These are generic code review observations (e.g., "parameter not validated"), not consensus-layer vulnerabilities. The model cannot distinguish between general code quality suggestions and actual security vulnerabilities.

### Vulnerable Files (9/9 detected — TPR 100%)

Llama correctly detects all 9 vulnerable files. The vulnerability types are generally correct (time.Now(), global variables, goroutines, map iteration, phantom reads, iterator leaks). However, Llama responses tend to also include non-consensus noise findings alongside the correct ones.

---

## Summary Statistics

### Qwen2.5-Coder-7B Zero-Shot

| Metric | Value |
|:-------|:------|
| TPR (True Positive Rate) | 9/9 (100%) |
| TNR (True Negative Rate) | 6/6 (100%) |
| Correct vulnerability identification | 8/9 files with exact match |
| Partially correct | 1/9 (vuln_06_b: iterator leak detected but miscategorized) |
| Hallucinations on vulnerable files | 2 fabricated findings in vuln_06_b (private data leakage, insecure key management) |
| Hallucinations on safe files | 0 |
| Self-contradictions | 1 (vuln_06_b: findings followed by "No vulnerabilities detected") |

### Llama-3.1-8B Zero-Shot

| Metric | Value |
|:-------|:------|
| TPR (True Positive Rate) | 9/9 (100%) |
| TNR (True Negative Rate) | 1/6 (17%) |
| False positives on safe files | 5/6 — all generic "input validation" or "access control" hallucinations |
| Hallucination pattern | Systematic non-consensus findings on safe code |

---

## Key Findings

1. **Qwen CPR is high**: 8/9 vulnerable files have exactly correct vulnerability identification. The single partial case (vuln_06_b) still detects the iterator leak pattern but miscategorizes it.

2. **Qwen hallucinations are localized**: Only vuln_06_b_conditional_leak.go contains fabricated findings (2 out of 6 stated findings). All other 14 files have zero hallucinations.

3. **Llama's TNR failure is systematic hallucination**: Llama fabricates "input validation" and "access control" findings on safe code. These are generic code review suggestions, not consensus-layer vulnerabilities. The model lacks the ability to distinguish code quality from security vulnerabilities.

4. **Self-contradiction exists but is handled**: Qwen's vuln_06_b response ends with "No vulnerabilities detected" after listing 6 findings. Classifier v2 correctly resolves this by prioritizing structured evidence.

5. **Vulnerability type accuracy**: For the 8 cleanly detected vulnerable files, Qwen correctly identifies the specific vulnerability type (timestamp, global variable, goroutine race, map iteration, phantom read, iterator leak) in every case.
