# Add-on External Validation — Final Labels

> Finalized: 2026-04-21
> Basis: GPT 5.4 Pro 3-session cross-analysis (labeling 3 sessions + U03 adjudication 3 sessions)
> Protocol: EXPERIMENT_PROTOCOL.md Phase 2

---

## Label Summary

| # | File | Label | Class | Evidence | Rationale |
|:--|:-----|:------|:------|:---------|:----------|
| U01 | LandRegistry.go | **safe** | — | — | Pure CRUD, no nondeterministic source |
| U02 | ethtxcc.go | **safe** | — | — | Deterministic key-value operations only |
| U03 | election_code.go | **vulnerable** | timestamp | L499 `time.Now()` → L508 control branch → L557,L569 `PutState` | time.Now() in control dependency of write path; startDate≤endDate invariant not enforced in code; malformed date interval makes branch reachable |
| U07 | ProductDetails.go | **safe** | — | — | Standard asset CRUD |
| U08 | smartcontract.go | **safe** | — | — | Deterministic operations only |
| U09 | charity.go | **safe** | — | — | Simple PutState with deterministic inputs |
| U10 | carcert.go | **safe** | — | — | Certificate CRUD, no nondeterministic source |
| U11 | sharebook.go | **safe** | — | — | Book sharing logic, deterministic |
| U12 | local_model_chaincode.go | **safe** | — | — | Model metadata CRUD |
| U13 | security_manager.go | **vulnerable** | phantom_read | GetState→PutState on overlapping keys without MVCC guard | Read-after-write pattern with concurrent access risk |
| U14 | smartaicc.go | **safe** | — | — | AI model metadata, deterministic |
| U17 | realty_chaincode.go | **safe** | — | — | Real estate CRUD, deterministic |
| U18 | donation_chaincode.go | **vulnerable** | timestamp | `time.Now()` value flows to PutState | Nondeterministic timestamp written to ledger |
| U20 | maintenance.go | **vulnerable** | timestamp | `time.Now()` value flows to PutState | Nondeterministic timestamp written to ledger |
| U21 | movies.go | **safe** | — | — | Movie catalog CRUD |
| U22 | voting.go | **safe** | — | — | Voting logic, deterministic operations |
| U23 | private_blockchain.go | **safe** | — | — | Private data operations, deterministic |

---

## Counts

| Label | Count | Files |
|:------|:------|:------|
| Vulnerable | **4** | U03, U13, U18, U20 |
| Safe | **13** | U01, U02, U07, U08, U09, U10, U11, U12, U14, U17, U21, U22, U23 |
| **Total** | **17** | |

## Class Distribution (Vulnerable)

| Class | Count | Files |
|:------|:------|:------|
| timestamp | 3 | U03, U18, U20 |
| phantom_read | 1 | U13 |

## U03 Adjudication Record

- Previous analysis: "L188 `&&` always false → dead code → safe"
- GPT 3-session adjudication: **3/3 vulnerable**
- Key reasoning: `startDate <= endDate` invariant not enforced in code; malformed date interval (endDate < now < startDate) makes branch reachable; time.Now() remains in control dependency of write path
- Label annotation: "vulnerable, but activation depends on malformed date interval due to an apparent logic bug"
- Sensitivity row: paper will report both U03=V and U03=S scenarios

## FLAG Files (Not Included)

| FLAG | File | Reason for Exclusion |
|:-----|:-----|:--------------------|
| U05 | fabcar.go | SAMPLE_EXTENDED (75% fabric-samples similar) |
| U16 | hlf_time_oracle.go | NO_PUTSTATE (no ledger write) |
| U24 | fabric_chaincode.go | SMALL (80 NCLOC borderline) |
| U25 | supplychain.go | TOO_LARGE (41,825 bytes, n_ctx overflow) |

## Statistical Notes

- Prevalence: 4/17 = 23.5% (exact 95% CI: 6.8–49.9%)
- TPR CI if 4/4 detected: 39.8–100.0%
- TPR CI if 3/4 detected: 9.4–77.5%
- Framing: "descriptive evidence of real-world prevalence, not precise TPR estimation"
