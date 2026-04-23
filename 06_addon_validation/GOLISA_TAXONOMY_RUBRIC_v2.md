# GoLiSA-Derived, Scope-Restricted, Fabric-Focused Subset: Taxonomy & Labeling Rubric
> Version: 2.0 | Frozen: 2026-04-22 | Status: FROZEN — DO NOT MODIFY AFTER LABELING BEGINS
> Derived from: v1.0 (2026-04-22) + GPT 5.4 Pro 9-session advisory + Claude critical cross-analysis

---

## 1. Task Definition

- **Task**: Binary file-level detection of consensus-critical nondeterminism
- **Scope**: Hyperledger Fabric chaincode (Go)
- **Unit**: Single .go source file
- **Output**: VULNERABLE or SAFE (binary)
- **Corpus naming**: "GoLiSA-derived, scope-restricted, Fabric-focused subset"

This benchmark uses a narrowed, benchmark-specific subset of GoLiSA/Fabric nondeterminism families. It is inspired by GoLiSA, but it is not a full reproduction of GoLiSA's official source taxonomy. C4 is a Fabric-specific extension; C5 is an auxiliary resource-hygiene label rather than a primary consensus-critical nondeterminism class.

---

## 2. Taxonomy

### 2.1 Core Classes (5 Classes — determine V/S assignment)

| ID | Class Name | Definition | Source Pattern | Sink Requirement |
|:---|:-----------|:-----------|:---------------|:-----------------|
| C1 | TIME_NOW | Use of current wall-clock time (time.Now, time.Since, time.Until) in transaction logic | `time.Now()`, `time.Since()`, `time.Until()` | Must influence ledger write or proposal response |
| C2 | GOROUTINE | Goroutine or async concurrent work whose scheduling/results affect transaction outcome. GoLiSA concurrency family, goroutine-only subset (go statements only); channel receives (<-) and channel-order nondeterminism are excluded from this frozen subset. | `go func()`, `go methodCall()` | Must influence ledger write or proposal response |
| C3 | MAP_ITERATION | Iteration over a map-typed expression whose unspecified order influences transaction outcome | Any `for ... range x` where `x` is inferred to have map type | Must influence ledger write or proposal response |
| C4 | NON_REVALIDATED_QUERY (legacy alias: PHANTOM_READ) | Query API whose results are not re-validated at commit time, used to decide ledger writes. Fabric-specific extension, not a GoLiSA source class. | `GetQueryResult`, `GetPrivateDataQueryResult`, `GetQueryResultWithPagination`, `GetHistoryForKey` | Query result must determine write decision |
| C6 | GLOBAL_MUTABLE_STATE | Package-level variable whose cross-invocation value influences transaction outcome. C6 follows GoLiSA's global-variable source family at package scope. | `var globalVar = ...` at package level | Must be read in transaction logic AND influence ledger write or response |

**C6 Benchmark Precision Heuristic**: Ignore canonical logger-handle globals when they are used solely for logging and there is no concrete intra-file evidence that the global value influences a core sink. This exclusion is a benchmark house rule, not part of the original GoLiSA definition.

### 2.2 Auxiliary Label (does NOT determine V/S assignment)

| ID | Class Name | Definition | Source Pattern | Note |
|:---|:-----------|:-----------|:---------------|:-----|
| C5 | ITERATOR_LEAK | Ledger/query iterator created but not closed on all execution paths | `GetStateByRange`, `GetQueryResult`, `GetStateByPartialCompositeKey`, etc. → missing `.Close()` on some path | Auxiliary resource-hygiene tag. Does not determine the positive label in the core V/S task. Retained for traceability, overlap analysis, hard-negative construction, and error analysis. |

### 2.3 Explicitly OUT OF SCOPE

| Pattern | Reason |
|:--------|:-------|
| `math/rand`, `crypto/rand` | GoLiSA official source family, but excluded from this frozen subset to avoid reopening D1/P0/D2/Semgrep rules |
| System/environment APIs (`os.Getenv`, `os.Hostname`, `net.Interfaces`, etc.) | GoLiSA official source family, excluded from this frozen subset |
| File system & I/O APIs (`os.Open`, `net.Listen`, `io/ioutil.ReadFile`, etc.) | GoLiSA official source family, excluded from this frozen subset |
| External process APIs (`os.StartProcess`, `os/exec.Command`) | GoLiSA official source family, excluded from this frozen subset |
| Channel-based nondeterminism (`<-`, receive-order/scheduling effects) | GoLiSA concurrency family includes both `go` and `<-`; channel read is excluded from this frozen subset |
| `InvokeChaincode` | Out of scope as an indirect inter-chaincode sink boundary. This frozen subset does not chase effects across chaincode boundaries. |
| Access control issues | Task scope 외 |
| Input validation | Task scope 외 |
| Private data leakage | Task scope 외 |
| Key management | Task scope 외 |
| Generic code quality | Task scope 외 |

---

## 3. Sink Definition (Consensus-Critical)

A finding counts as VULNERABLE **only if** the nondeterministic source reaches one of these sinks:

### 3.1 Ledger / Metadata Write Sinks
- `PutState(key, value)`
- `DelState(key)`
- `PutPrivateData(collection, key, value)`
- `DelPrivateData(collection, key)`
- `PurgePrivateData(collection, key)`
- `SetStateValidationParameter(key, ep)`
- `SetPrivateDataValidationParameter(collection, key, ep)`

### 3.2 Response Sinks
- `shim.Success(payload)` / `shim.Error(msg)` — return value of Invoke/Init
- `ctx.GetStub().SetEvent(name, payload)`
- Any non-error value returned from a contractapi transaction function
- Any error return that maps to `shim.Error`

### 3.3 Implicit-Flow Rule
A branch is not itself a sink. However, if a source taints the guard of an `if`/`switch`/`select` and that guard determines whether a ledger-write sink or response sink executes, treat this as a confirmed source→sink path via implicit flow. The real sink is still the guarded write/response operation, not the branch node itself.

### 3.4 API Equivalence
- **shim-style**: `stub.X(...)`
- **contractapi-style**: `ctx.GetStub().X(...)`
- The same equivalence applies to all sink APIs: `PutState`, `DelState`, `PutPrivateData`, `DelPrivateData`, `PurgePrivateData`, `SetStateValidationParameter`, `SetPrivateDataValidationParameter`, and `SetEvent`.

---

## 4. Labeling Rubric

### 4.1 Label Values
- **V** (VULNERABLE): At least one C1-C4 or C6 core class present with confirmed source→sink path
- **S** (SAFE): No core class present, OR core class present but no consensus-critical sink reachable
- **EXCLUDE**: File is not a chaincode transaction handler (test, utility, config, main.go with no entrypoint, etc.)

### 4.2 Multi-Label
- A single file can have multiple vulnerability classes (e.g., TIME_NOW + NON_REVALIDATED_QUERY)
- Record **primary_family** (most significant) and **secondary_family** (if applicable)
- For binary metric: any core V label → VULNERABLE
- C5 (ITERATOR_LEAK) is tagged as auxiliary and does **not** contribute to the V/S assignment

### 4.3 Source→Sink Rubric (Decision Tree)

```
1. Does the file contain a recognizable Hyperledger Fabric transaction entrypoint?

   Recognizable entrypoints include:
   (a) shim-style entrypoints: Init(...) or Invoke(...); or
   (b) contractapi-style callable contract methods, including public methods
       that are dispatched by ContractChaincode.
       Sufficient indicators include:
       - a public method whose first parameter is
         contractapi.TransactionContextInterface,
         *contractapi.TransactionContext, or a custom transaction-context
         type/interface accepted by contractapi; or
       - a public method on a contract type (e.g., a type embedding
         contractapi.Contract); or
       - local evidence that the contract type is passed to
         contractapi.NewChaincode(...).

   Note: for contractapi code, do NOT require literal method names Init or
   Invoke. For contractapi, the transaction context is optional; if present,
   it must be the first parameter.

   Rationale: in Fabric Contract API, public contract methods are callable
   through ContractChaincode dispatch, and their return values are serialized
   into shim.Success/shim.Error.

   → NO → EXCLUDE
   → YES → continue

2. Does the file contain any C1-C4 or C6 source pattern?
   → NO → SAFE
   → YES → continue

3. Does the source pattern reach a consensus-critical sink (§3)?
   → NO → SAFE (pattern present but not consensus-relevant)
   → YES → continue

4. Is the usage in:
   - logging/metrics only? → SAFE
   - test/dead/demo code only? → SAFE
   - comments/filenames only? → SAFE
   - constant/read-only config? → SAFE (for GLOBAL_MUTABLE_STATE, see C6 heuristic)
   - user-provided timestamp (not time.Now)? → SAFE (for TIME_NOW)
   - slice/array iteration (not map)? → SAFE (for MAP_ITERATION)
   - GetStateByRange or GetStateByPartialCompositeKey alone? → SAFE (for NON_REVALIDATED_QUERY)
   - Iterator with Close() on ALL paths? → SAFE (for ITERATOR_LEAK auxiliary)
   → YES to any → SAFE
   → NO to all → VULNERABLE, record core class(es)
```

### 4.4 Scope Rule

Labels are assigned strictly at single-file granularity. Use only code present in the current file, including helper functions defined in the same file, to establish explicit or implicit source→sink paths. Do not chase calls into other files, other packages, generated code, vendored dependencies, or imported libraries.

A file is VULNERABLE only if a recognizable transaction entrypoint and a supporting source→sink path are both evidenced within that same file. Helper-only files with no recognizable entrypoint are EXCLUDE.

A file that defines a contractapi/shim transaction method counts as entrypoint-containing even if the receiver type or main function is declared elsewhere in the package.

### 4.5 Ambiguous Cases

| Scenario | Decision |
|:---------|:---------|
| time.Now() stored but no concrete intra-file influence on sink is visible | SAFE — mere co-location of a source and a sink in the same file, or speculation about helper behavior not shown in the file, is insufficient. Label V only when there is concrete intra-file evidence that a source reaches a sink, or that a tainted condition guards sink execution. |
| `GetTxTimestamp()` used in transaction logic | SAFE for C1 — deterministic across all endorsers (value from ChannelHeader) |
| Global var exists but only read, never written after init | SAFE only if its initializer is deterministic and no in-scope write occurs later |
| Iterator Close() in `defer` | SAFE with respect to normal returns and panics; Go executes deferred calls when the surrounding function returns or panics. Flag only if no reachable Close() exists (e.g., early return before defer, conditional omission, or non-return termination such as os.Exit). |
| Map range but result sorted before write | SAFE — nondeterminism neutralized |
| `encoding/json.Marshal(map)` from Go standard library (v1) | SAFE for C3 — map keys are sorted during marshaling. This exception does not extend to `encoding/json/v2` or to other serializers unless deterministic key ordering is explicitly documented. |
| `yaml.Marshal(map)`, `fmt.Sprintf(map)`, custom serializer | SAFE only if deterministic ordering is guaranteed by the specific library/version or explicit key sorting is visible in code; otherwise do not auto-exempt |
| Multiple transaction functions, only one has issue | V — file-level label |
| Partial code / snippet (no package declaration) | EXCLUDE |
| `*_test.go` or `package xxx_test` | EXCLUDE |
| `//go:build`-restricted file | Include/exclude according to the benchmark build context, not whether the file is standalone-compilable |
| `init()` / package-level initializer uses nondeterministic API | V only if the initialized value is later read by transaction logic in the same file and influences a sink; otherwise SAFE/EXCLUDE |
| `GetHistoryForKey` result used to decide writes | NON_REVALIDATED_QUERY (C4) — not merely ITERATOR_LEAK |

---

## 5. NON_REVALIDATED_QUERY Boundary (Critical Distinction)

### 5.1 NON_REVALIDATED_QUERY (Flag as C4)
Query APIs whose results are not re-validated at commit time:
- `GetQueryResult(queryString)` — CouchDB rich query, NOT re-validated at commit
- `GetPrivateDataQueryResult(collection, queryString)` — same issue
- `GetQueryResultWithPagination(queryString, pageSize, bookmark)` — read-only transaction only
- `GetHistoryForKey(key)` — NOT re-validated at commit; official docs state it should not be used in update transactions

Result used to decide PutState/DelState → VULNERABLE

### 5.2 NOT NON_REVALIDATED_QUERY (Do NOT flag as C4)
- `GetStateByRange(startKey, endKey)` alone — has key-level MVCC validation with phantom protection
- `GetStateByPartialCompositeKey(objectType, keys)` — key-based, MVCC protected
- These may still trigger C5 auxiliary tag (ITERATOR_LEAK) if not properly closed

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
- **Statistics**: 95% CI (Wilson for rates, bootstrap for composites), raw confusion counts, exact McNemar for paired detector comparison

---

## 7. Benchmark Construction Rules

| Rule | Value |
|:-----|:------|
| Target size | 30V + 30S (fallback: 24V + 24S) |
| One-file-per-repo | Mandatory (no exceptions even for fallback) |
| TIME_NOW cap | ≤ 40% of total positives |
| Dominant-family max cap | Set per family to prevent any single family from dominating |
| Family-level metric threshold | n ≥ 3 positives per family (descriptive); n ≥ 5 for inferential |
| GOROUTINE if n < 3 | Exclude from family-level analysis, include in overall |
| Safe selection | Hard negatives mandatory inclusion + stratified random easy safes |
| Hard-negative reporting | TNR_easy / TNR_hard reported separately |
| Reserve files | 10 (for preprocess failure or label revision) |
| Near-duplicate check | Required (normalized text hash + Jaccard similarity) |
| Labeling basis | Comment-stripped source files (primary view) |
| Model input | Comment-stripped + no filename/path |
| Sampling method | Stratified random with fixed seed (no manual "representative" selection) |
| V/S assignment | Core classes C1-C4, C6 only. C5 is auxiliary and does not determine V/S. |
