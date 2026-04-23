# GoLiSA-Derived Benchmark: Taxonomy & Labeling Rubric
> Version: 1.0 | Frozen: 2026-04-22 | Status: FROZEN — DO NOT MODIFY AFTER LABELING BEGINS

---

## 1. Task Definition

- **Task**: Binary file-level detection of consensus-critical nondeterminism
- **Scope**: Hyperledger Fabric chaincode (Go)
- **Unit**: Single .go source file
- **Output**: VULNERABLE or SAFE (binary)
- **Corpus naming**: "GoLiSA-derived real-world benchmark"

---

## 2. Taxonomy (6 Classes — FROZEN)

| ID | Class Name | Definition | Source Pattern | Sink Requirement |
|:---|:-----------|:-----------|:---------------|:-----------------|
| C1 | TIME_NOW | Use of current wall-clock time (time.Now, time.Since, time.Until) in transaction logic | `time.Now()`, `time.Since()`, `time.Until()` | Must influence ledger write or proposal response |
| C2 | GOROUTINE | Goroutine or async concurrent work whose scheduling/results affect transaction outcome | `go func()`, `go methodCall()` | Must influence ledger write or proposal response |
| C3 | MAP_ITERATION | Iteration over a Go map whose nondeterministic order influences transaction outcome | `for k, v := range mapVar` (where mapVar is `map[...]`) | Must influence ledger write or proposal response |
| C4 | PHANTOM_READ | Rich-query results used to decide ledger writes without safe re-validation | `GetQueryResult`, `GetPrivateDataQueryResult`, `GetQueryResultWithPagination` | Query result must determine write decision |
| C5 | ITERATOR_LEAK | Ledger/query iterator created but not closed on all execution paths | `GetStateByRange`, `GetQueryResult`, `GetHistoryForKey`, etc. → missing `.Close()` on some path | Iterator resource leak |
| C6 | GLOBAL_MUTABLE_STATE | Mutable package-level/global variable whose cross-invocation value influences transaction outcome | `var globalVar = ...` (non-const, non-logger) at package level | Must be read in transaction logic AND influence ledger write or response |

### 2.1 Explicitly OUT OF SCOPE

| Pattern | Reason |
|:--------|:-------|
| `math/rand` | Ontology 확장 시 D1/P0/D2/Semgrep 규칙 전부 재개방 필요. 이번 논문에서는 exploratory only |
| Access control issues | Task scope 외 |
| Input validation | Task scope 외 |
| Private data leakage | Task scope 외 |
| Key management | Task scope 외 |
| Generic code quality | Task scope 외 |

---

## 3. Sink Definition (Consensus-Critical)

A finding counts as VULNERABLE **only if** the nondeterministic source reaches one of these sinks:

### 3.1 Ledger Write Sinks
- `PutState(key, value)`
- `DelState(key)`
- `PutPrivateData(collection, key, value)`
- `DelPrivateData(collection, key)`
- `PurgePrivateData(collection, key)`
- `SetStateValidationParameter(key, ep)`

### 3.2 Response Sinks
- `shim.Success(payload)` / `shim.Error(msg)` — return value of Invoke/Init
- `ctx.GetStub().SetEvent(name, payload)`
- Any `return` from a transaction function that carries computed data

### 3.3 Control Flow Sinks
- Conditional branch (`if`/`switch`/`select`) that determines whether a ledger write or response occurs

### 3.4 API Equivalence
- **shim-style**: `stub.PutState(...)` — direct ChaincodeStubInterface call
- **contractapi-style**: `ctx.GetStub().PutState(...)` — wrapper through TransactionContextInterface
- These are **equivalent sinks**. Both are treated identically.

---

## 4. Labeling Rubric

### 4.1 Label Values
- **V** (VULNERABLE): At least one C1-C6 class present with confirmed source→sink path
- **S** (SAFE): No C1-C6 class present, OR class present but no consensus-critical sink reachable
- **EXCLUDE**: File is not a chaincode transaction handler (test, utility, config, main.go with no Invoke, etc.)

### 4.2 Multi-Label
- A single file can have multiple vulnerability classes (e.g., TIME_NOW + ITERATOR_LEAK)
- Record **primary_family** (most significant) and **secondary_family** (if applicable)
- For binary metric: any V label → VULNERABLE

### 4.3 Source→Sink Rubric (Decision Tree)

```
1. Does the file contain a recognizable HLF transaction handler?
   (Invoke, Init, or contractapi transaction function)
   → NO → EXCLUDE
   → YES → continue

2. Does the file contain any C1-C6 source pattern?
   → NO → SAFE
   → YES → continue

3. Does the source pattern reach a consensus-critical sink (§3)?
   → NO → SAFE (pattern present but not consensus-relevant)
   → YES → continue

4. Is the usage in:
   - logging/metrics only? → SAFE
   - test/dead/demo code only? → SAFE
   - comments/filenames only? → SAFE
   - constant/read-only config? → SAFE (for GLOBAL_MUTABLE_STATE)
   - user-provided timestamp (not time.Now)? → SAFE (for TIME_NOW)
   - slice/array iteration (not map)? → SAFE (for MAP_ITERATION)
   - GetStateByRange alone (not rich query)? → SAFE (for PHANTOM_READ)
   - Iterator with Close() on ALL paths? → SAFE (for ITERATOR_LEAK)
   → YES to any → SAFE
   → NO to all → VULNERABLE, record class(es)
```

### 4.4 Ambiguous Cases

| Scenario | Decision |
|:---------|:---------|
| time.Now() stored but unclear if it reaches PutState | Label as V if any plausible execution path exists |
| Global var exists but only read, never written after init | SAFE |
| Iterator Close() in defer but function has panic path | V (ITERATOR_LEAK) — defer may not execute on all paths in all Go versions |
| Map range but result sorted before write | SAFE — nondeterminism neutralized |
| Multiple transaction functions, only one has issue | V — file-level label |
| Partial code / snippet (no package declaration) | EXCLUDE |
| Build tag restricted code | EXCLUDE if not compilable as standalone |

---

## 5. PHANTOM_READ Boundary (Critical Distinction)

### 5.1 PHANTOM_READ (Flag as C4)
- `GetQueryResult(queryString)` — CouchDB rich query, NOT re-validated at commit
- `GetPrivateDataQueryResult(collection, queryString)` — same issue
- `GetQueryResultWithPagination(queryString, pageSize, bookmark)` — same
- Result used to decide PutState/DelState → VULNERABLE

### 5.2 NOT PHANTOM_READ (Do NOT flag as C4)
- `GetStateByRange(startKey, endKey)` alone — has key-level MVCC validation
- `GetStateByPartialCompositeKey(objectType, keys)` — key-based, MVCC protected
- `GetHistoryForKey(key)` — read-only history, not phantom-relevant
- These may still be ITERATOR_LEAK (C5) if not properly closed

---

## 6. Primary Metrics

| Metric | Formula | Role |
|:-------|:--------|:-----|
| TPR (Sensitivity/Recall) | TP / (TP + FN) | Main |
| TNR (Specificity) | TN / (TN + FP) | Main |
| BA (Balanced Accuracy) | (TPR + TNR) / 2 | Main |
| Precision | TP / (TP + FP) | Secondary |
| F1 | 2 × (Precision × Recall) / (Precision + Recall) | Secondary |

### 6.1 Reporting Structure
- **Main table**: temp=0.0 single deterministic run (if determinism confirmed)
- **Robustness**: temp=0.1 × 5-run per-file stability (secondary)
- **Family breakdown**: per-class TPR for families with n ≥ 3 positives

---

## 7. Benchmark Construction Rules

| Rule | Value |
|:-----|:------|
| Target size | 30V + 30S (fallback: 24V + 24S) |
| One-file-per-repo | Mandatory (no exceptions even for fallback) |
| TIME_NOW cap | ≤ 12 positives (≤ 40% of total positives) |
| Family-level metric threshold | n ≥ 3 positives per family |
| GOROUTINE if n < 3 | Exclude from family-level analysis, include in overall |
| Safe selection | Matched hard negatives preferred |
| Reserve files | 10 (for preprocess failure or label revision) |
| Near-duplicate check | Required (normalized text hash) |
| Labeling basis | Raw source files |
| Model input | Comment-stripped + no filename/path |
