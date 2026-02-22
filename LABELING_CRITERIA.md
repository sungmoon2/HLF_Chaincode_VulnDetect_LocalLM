# Labeling Criteria

> Reproducibility reference for how findings are labeled as "consensus-relevant" and how ground truth is defined.

## 1. Consensus-Relevant Finding Definition

A finding is labeled **consensus-relevant** if it describes a vulnerability that can cause endorsement mismatch in Hyperledger Fabric's execute-order-validate architecture. Specifically, any non-deterministic behavior that causes different endorsing peers to produce different read-write sets for the same transaction.

### Keyword-Based Detection

Source: `scripts/08_run_golisa_validation.py`, lines 83-89

```python
CONSENSUS_KEYWORDS = [
    "non-deterministic", "nondeterministic",
    "endorsement", "consensus", "phantom read",
    "read-after-write", "chaincode",
    "fabric", "hlf", "ledger",
    "getstate", "putstate",
]
```

**Matching method** (from `scripts/08_run_golisa_validation.py`):

```python
text_lower = (rule_id + " " + message).lower()
is_consensus = any(kw in text_lower for kw in CONSENSUS_KEYWORDS)
```

- Input: concatenation of `rule_id` and `message` from a finding
- Transform: lowercase
- Match: case-insensitive substring match against 12 keywords
- Output: boolean (`True` if any keyword matches)

This method is used to classify Semgrep findings. For LLM responses, Classifier v1/v2 (see CLASSIFIER.md) is used instead.

---

## 2. Micro-Benchmark Ground Truth

Source: `scripts/10_run_json_mode_microbenchmark.py`, lines 26-42

15 Go chaincode files with manually assigned labels:

### Vulnerable Files (9)

| File | Vulnerability Type | Description |
|:-----|:-------------------|:------------|
| `vuln_01_time.go` | timestamp | `time.Now()` result flows to PutState |
| `vuln_02_global.go` | global variable | Global variable mutated across invocations, value written to PutState |
| `vuln_03_goroutine.go` | goroutine race | Goroutines with shared state, result written to PutState |
| `vuln_04_map_iter.go` | map iteration | Go map range iteration order is non-deterministic, concatenated result goes to PutState |
| `vuln_05_phantom.go` | phantom read | GetState followed by PutState on overlapping keys without MVCC handling |
| `vuln_06_iterator_leak.go` | iterator leak | GetStateByRange iterator not closed with defer |
| `vuln_01_b_interprocedural.go` | timestamp (interprocedural) | `time.Now()` passed through helper function to PutState |
| `vuln_04_b_nested_map.go` | map iteration (nested) | Nested map iteration, concatenated result goes to PutState |
| `vuln_06_b_conditional_leak.go` | iterator leak (conditional) | Iterator closed only on one branch of a conditional |

### Safe Files (6)

| File | Why Safe |
|:-----|:---------|
| `safe_01_logging.go` | `time.Now()` used only in logging (fmt.Printf), not in PutState |
| `safe_02_local_var.go` | All variables are local and deterministic |
| `safe_03_map_read.go` | Map used for read-only lookup, no iteration over map for PutState value |
| `safe_04_math_rand.go` | `math/rand` used only in non-ledger context |
| `safe_05_deterministic_time.go` | Uses `ctx.GetStub().GetTxTimestamp()` (deterministic) instead of `time.Now()` |
| `safe_06_external_lib.go` | External library call result does not flow to PutState |

### Labeling Principle

A file is labeled `vulnerable` if and only if:
- A non-deterministic source (time.Now, math/rand, map iteration, goroutine race, external API, file I/O) produces a value that **flows into a PutState/DelState call** (directly or transitively), OR
- A resource leak exists (iterator not closed), OR
- A phantom read pattern exists (GetState followed by PutState on overlapping keys)

A file is labeled `safe` if:
- Non-deterministic functions may be present, but their output **does not reach** the ledger write set (PutState/DelState)

---

## 3. GoLiSA Running_Examples Ground Truth

Source: `scripts/08_run_golisa_validation.py`, lines 94-120
Also: `scripts/09_reclassify_and_ablation.py`, lines 32-38

5 known-vulnerable Go chaincode files from the GoLiSA benchmark's `Running_Examples/` directory:

| File Path | Vulnerability Type | Description |
|:----------|:-------------------|:------------|
| `Running_Examples/channel/Channel.go` | goroutine | Two goroutines send to channel, receive order varies, result goes to PutState |
| `Running_Examples/global/GlobalVariable.go` | global_var | `var glob` written across invocations, PutState uses glob value |
| `Running_Examples/goroutine/GoRoutines.go` | goroutine | Two goroutines append to shared string `s`, PutState uses `s` |
| `Running_Examples/map-iter/MapIteration.go` | map_iter | Range over map concatenates values, order nondeterministic, goes to PutState |
| `Running_Examples/method-function/MethodFunction.go` | timestamp | `time.Now()` result directly passed to PutState |

All 5 files are expected to be classified as `vulnerable`.

These files are authored by the GoLiSA project (Olivieri et al., ECOOP 2023) and serve as external validation against independently constructed test cases.

---

## 4. GoLiSA Full Benchmark (657 Files)

Source: `02_resources/golisa_benchmark/Benchmark/`

- 657 Go chaincode files extracted from 326 GitHub repositories
- No per-file ground truth labels available (unlike the 5 Running_Examples)
- Used for distribution analysis, not per-file accuracy measurement
- Classifier v1 result: 380 vulnerable / 277 safe
- Classifier v2 result: 477 vulnerable / 180 safe (97 reclassified from safe to vulnerable)

---

## 5. Semgrep Baseline Configuration

Source: `scripts/05_run_traditional_tools.py`, lines 40-43

```python
SEMGREP_CONFIGS = [
    {"label": "auto",           "config": "auto"},
    {"label": "security-audit", "config": "p/security-audit"},
]
```

- `auto`: Semgrep registry automatic rule selection
- `p/security-audit`: Security audit-specific ruleset

Result on micro-benchmark (15 files):
- Consensus-relevant findings: 0
- Generic findings: 1 (math/rand usage in `safe_04_math_rand.go`)

Excluded tools (documented in meta.json):
- `go vet`: Requires Go compiler (not installed)
- `staticcheck`: Requires Go compiler
- `golangci-lint`: Requires Go compiler
