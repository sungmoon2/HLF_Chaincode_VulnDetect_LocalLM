# Dev Set Exclusion List
> Created: 2026-04-22 | Purpose: sanity check/dev 파일의 final 60 / reserve / D2 재유입 방지
> 근거: Experiment Design 자문 3/3 합의 — "dev/sanity에 쓴 파일은 final 60, reserve, D2에서 완전 제외"

## Dev Files (sanity check에 사용된 10개 파일)

| # | Repo | Filename | Dev에서의 용도 | Candidate Pool 상태 |
|:--|:-----|:---------|:-------------|:------------------|
| 1 | bluezd | integralTrace.go | sanity check | **Positive candidate — 제외** |
| 2 | xuehuiit | example_cc.go | sanity check | **Positive candidate — 제외** |
| 3 | nitesh7sid | marbles_chaincode.go | sanity check | 후보 탈락 (해당 없음) |
| 4 | RAntonio09 | ccConsortium.go | sanity check | **Positive candidate — 제외** |
| 5 | cactusfluo | marbles_chaincode.go | sanity check | 후보 탈락 (해당 없음) |
| 6 | RakhiSoni | marbles_chaincode.go | sanity check | 후보 탈락 (해당 없음) |
| 7 | lutianYan | myHospital.go | sanity check | **Safe candidate — 제외** |
| 8 | joseprados | ReadingAsset.go | sanity check | **Safe candidate — 제외** |
| 9 | ewerter | customerloyalty.go | sanity check | **Safe candidate — 제외** |
| 10 | pankajcheema | dummyuser.go | sanity check | **Safe candidate — 제외** |

## Same-Repo Different-File (보수적 제외)

| Repo | Dev 파일 | Candidate 파일 | Candidate 유형 | 조치 |
|:-----|:--------|:-------------|:-------------|:-----|
| cactusfluo | marbles_chaincode.go | invoke.go | Hard negative | **제외** — 같은 repo에서 dev 파일을 사용했으므로 tuning leakage 위험 |

## 제외 영향 요약

| 항목 | 제외 전 | 제외 후 | 감소 |
|:-----|:-------|:-------|:-----|
| Positive candidates | 101 | **98** | -3 |
| Safe candidates | 112 | **108** | -4 |
| Hard negatives | 13 | **12** | -1 |

## 적용 규칙

- 위 11개 파일(10 dev + 1 same-repo)은 **final 60, reserve 10, D2 holdout**에서 완전 제외
- GPT 1차 라벨링은 전체 pool(101+112)에 대해 실행하되, benchmark freeze 단계에서 제외 적용
- 라벨링 결과 자체는 보존 (사후 참고용)
