# Phase 2 Labeling Verification Summary
> Date: 2026-04-22~23 | Verifier: Claude Opus 4.6 (code-level review)
> Source: run_260422_2142/ (622 files, Opus 4.5 via Vertex AI)
> Scope: GoLiSA full corpus (621 non-dev files + 1 Running_Examples)

## Final Numbers

| Item | Count |
|:-----|:------|
| Total labeled | 622 |
| GoLiSA non-dev coverage | 621/621 (100%) |
| TRUE VULNERABLE | 33 |
| TRUE SAFE | 585 |
| EXCLUDE | 4 |
| FALSE POSITIVES (V→S reclassified) | 13 |
| FALSE NEGATIVES (S→V reclassified) | 0 |

## Labeling Pipeline

```
GoLiSA 652 .go files
  → dev exclusion (10 repos, 31 files)
  → 621 non-dev files + 1 Running_Examples = 622 labeled
  → Opus 4.5 via Vertex AI (temp=0.0, rubric v2.0)
  → 46 VULNERABLE, 572 SAFE, 4 EXCLUDE
  → code-level verification (all 46 V + 70 positive→SAFE)
  → 13 FP reclassified → 33 TRUE V, 585 TRUE S, 4 EXCLUDE
```

## Verification Coverage

| Target | Count | Method | Result |
|:-------|:------|:-------|:-------|
| VULNERABLE (original 30) | 30 | 3 agents + human 2nd pass | 24 CORRECT, 6 FP |
| VULNERABLE (new 16) | 16 | 1 agent + human 2nd pass | 9 CORRECT, 7 FP |
| Positive→SAFE | 70 | 4 agents + human 2nd pass + same-function analysis | 70 SAFE CORRECT |
| All VULNERABLE total | 46 | full code review | 33 TRUE V, 13 FP |

## False Positive Analysis (13 files)

See: `false_positives.json`

### Original batch (6)
| File | Class | FP Reason |
|:-----|:------|:----------|
| Noonefang/chaincode.go | C3 | Iterates slices, not maps |
| richardfelkl/main.go | C3 | Iteration over StateQueryIterator, not map |
| netkiller/gasoline.go | C3 | Map field direct assignment, no for..range |
| soodakshay/mycc.go | C4 | Query gates entry, PutState unconditional |
| yoonjk/account.go | C4 | Read-only query, no write decision |
| linqd1/chaincode.go | C6 | No PutState in invoke (hard negative) |

### New batch (7)
| File | Class | FP Reason |
|:-----|:------|:----------|
| 2016Nishi/chaincode_example01.go | C6 | PutState commented out |
| abegpatel/marbles_chaincode.go | C4 | Query result ignored, PutState unconditional |
| fourbroad/chaincode_example01.go | C6 | PutState commented out |
| koakh/lscc.go | C3 | GetMSPs() cross-package, scope violation |
| lcy1317/chaincode.go | C6 | time.Local from time package, not file-level |
| qubing/relay_adapter.go | C4 | Cross-package function calls, scope violation |
| whatisoop/chaincode_example01.go | C6 | PutState commented out |

## FP Pattern Summary
- C6 "PutState commented out": 3 files (chaincode_example01 clones)
- C4 "query doesn't determine write": 3 files
- C3 "not actually map iteration": 3 files
- C6 "scope violation (external package var)": 1 file
- C4 "scope violation (cross-package functions)": 2 files
- C6 "no PutState exists": 1 file

## True VULNERABLE Family Distribution (33 files)

| Class | Count | % |
|:------|:------|:--|
| C1 TIME_NOW | 21 | 63.6% |
| C6 GLOBAL_MUTABLE_STATE | 5 | 15.2% |
| C3 MAP_ITERATION | 4 | 12.1% |
| C4 NON_REVALIDATED_QUERY | 2 | 6.1% |
| C2 GOROUTINE | 1 | 3.0% |

## Data Integrity Checks (ALL PASS)

1. per_file JSONs: 622, no missing fields, no invalid verdicts
2. summary.csv: 622 rows, matches per_file exactly
3. FP reclassification: all 13 FPs updated in per_file (verdict=SAFE, verified_reclassified=true)
4. verification covers all 46 ever-VULNERABLE files
5. GoLiSA coverage: 621/621 (100.0%)
6. Duplicates: 0
7. VULNERABLE without class: 0
8. SAFE with non-NONE class (excl reclassified): 0

## File Manifest

```
labeling/
├── run_260422_2142/
│   ├── per_file/              (622 JSON files, 1 per chaincode)
│   ├── summary.csv            (622 rows, synced with per_file)
│   ├── summary.meta.json      (run metadata)
│   └── progress.json          (completion tracking)
├── verification/
│   ├── VERIFICATION_SUMMARY.md         (this file)
│   ├── false_positives.json            (13 FP entries with reasons)
│   ├── vulnerable_verification.json    (46 entries: 33 correct + 13 FP)
│   └── positive_safe_verification.json (70 positive→SAFE confirmations)
└── smoke_*/                   (smoke test runs, archival)
```
