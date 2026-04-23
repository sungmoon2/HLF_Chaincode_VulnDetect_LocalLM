# Add-on External Validation — 라벨링 시트

> 라벨러: __________ (Park / Yang)
> 작성일: 2026-04-__
> 모델 출력 열람 여부: **미열람** (라벨링 완료 전 모델 실행 결과를 보지 않음)

## 라벨 기준

**Vulnerable**: 아래 6개 HLF endorsement/validation nondeterminism class 중 하나가 endorsement-relevant path(PutState, ChaincodeStubInterface write)에 도달하는 경우.

1. Nondeterministic timestamps (time.Now() → ledger write)
2. Global variable mutation (package-level mutable state across invocations)
3. Goroutine concurrency hazards (go statement within chaincode)
4. Map iteration randomness (range over map → PutState without serialization)
5. Phantom reads (GetStateByRange/GetStateByPartialCompositeKey read-write conflict)
6. Iterator resource leaks (iterator not closed after range query)

**Safe**: 위 6개 class에 해당하는 targeted anomaly 없음. Suspicious construct(time.Now, map, iterator 등)가 있어도 endorsement-relevant path에 도달하지 않으면 safe.

**Exclude**: 판정 불가 / ambiguous / out-of-scope (전체 프로젝트 컨텍스트 필요 등).

---

## 라벨링 결과

| # | File | NCLOC | Bytes | Label | Class | Evidence Lines | Rationale | Ambiguity |
|:--|:-----|:------|:------|:------|:------|:---------------|:----------|:----------|
| 1 | U01_LandRegistry.go | 118 | 4,911 | | | | | |
| 2 | U02_ethtxcc.go | 83 | 4,147 | | | | | |
| 3 | U03_election_code.go | 192 | 7,731 | | | | | |
| 4 | U07_ProductDetails.go | 158 | 5,258 | | | | | |
| 5 | U08_smartcontract.go | 170 | 6,306 | | | | | |
| 6 | U09_charity.go | 83 | 2,737 | | | | | |
| 7 | U10_carcert.go | 148 | 6,354 | | | | | |
| 8 | U11_sharebook.go | 499 | 16,026 | | | | | |
| 9 | U12_local_model_chaincode.go | 100 | 4,191 | | | | | |
| 10 | U13_security_manager.go | 495 | 23,938 | | | | | |
| 11 | U14_smartaicc.go | 318 | 9,474 | | | | | |
| 12 | U17_realty_chaincode.go | 422 | 15,525 | | | | | |
| 13 | U18_donation_chaincode.go | 92 | 3,351 | | | | | |
| 14 | U20_maintenance.go | 469 | 16,191 | | | | | |
| 15 | U21_movies.go | 415 | 15,595 | | | | | |
| 16 | U22_voting.go | 430 | 19,537 | | | | | |
| 17 | U23_private_blockchain.go | 170 | 7,347 | | | | | |

## 요약

- Total: 17
- Vulnerable: __
- Safe: __
- Excluded: __
- Initial Agreement: __/17
- Cohen's κ: __
