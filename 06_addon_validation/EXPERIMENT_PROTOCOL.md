# Add-on External Validation — 실험 프로토콜

> 작성: 2026-04-21
> 목적: C3 (N=15 데이터셋 규모) 리뷰어 지적 대응
> 근거: GPT 5.4 Pro 3세션 만장일치 — "추가 실험 필요"
> 기존 ORCHESTRATION.md 전략 Y 재개

---

## 실험 구조

```
D1 = 기존 15-file controlled diagnostic benchmark (유지, 변경 없음)
D2 = 독립 public-chaincode external validation set (신규)
```

두 세트는 **독립 보고**. 합산 N=32로 단일 벤치마크 취급하지 않음.

---

## Phase 1: 데이터셋 확정

### 1.1 후보 현황 (Phase 2 완료 상태)

| Status | 개수 | 파일 |
|:-------|:-----|:-----|
| KEEP | 17 | U01~U03, U07~U14, U17~U18, U20~U23 |
| FLAG | 4 | U05 (SAMPLE_EXTENDED), U16 (NO_PUTSTATE), U24 (SMALL), U25 (TOO_LARGE) |
| DROP | 3 | U04 (TOO_SHORT), U06 (SAMPLE_BENCHMARK), U15 (NOT_HLF) |

### 1.2 적격 기준 (GPT 3세션 수렴)

- [x] HLF chaincode Go 파일
- [x] vendored/generated/test/mock/tutorial 제외
- [x] repo당 1파일 우선 (중복 제거)
- [x] file-level label 독립 판정 가능
- [x] HLF endorsement-nondeterminism 관련 API/패턴 잠재적 존재
- [x] n_ctx=16,384 내 수용 가능 (prompt 포함)

### 1.3 n_ctx 적격성 검증

n_ctx=16,384 ≈ ~12,000 tokens 가용 (prompt+few-shot 제외 시).
1 token ≈ 3-4 bytes (Go code). 안전 상한 ≈ 36,000 bytes.

| 파일 | bytes | NCLOC | n_ctx 적격 |
|:-----|:------|:------|:----------|
| U25_supplychain.go | 41,825 | 1,180 | **초과 위험** |
| U13_security_manager.go | 23,938 | 495 | OK (여유) |
| U22_voting.go | 19,537 | 430 | OK |
| 나머지 15개 | ≤16,191 | ≤469 | OK |

→ **U25 제외** (n_ctx 초과 + FLAG:TOO_LARGE). 나머지 17 KEEP 전부 적격.

### 1.4 FLAG 보충 검토

vulnerable/safe 한쪽이 5개 미만일 경우에만 FLAG에서 보충.

| FLAG | 보충 가능성 | 비고 |
|:-----|:----------|:-----|
| U05 fabcar | 낮음 | 75% fabric-samples 유사, self-authored 비판과 유사한 문제 |
| U16 hlf_time_oracle | 보통 | PutState=0이라 file-level "safe" 판정 가능 |
| U24 fabric_chaincode | 보통 | 80 NCLOC borderline이나 적격 기준 충족 |

→ 라벨링 결과에 따라 결정. 사전 포함하지 않음.

### 1.5 최종 데이터셋 Freeze

**17 KEEP 파일 전부 사전 고정.**
모델 실행 전 변경 금지. 라벨링 후 ambiguous/out-of-scope 파일만 제외.

---

## Phase 2: 라벨링

### 2.1 라벨러
- 라벨러 A: 저자 1 (Park)
- 라벨러 B: 저자 2 (Yang) 또는 대리 전문가

### 2.2 프로토콜
1. 모델 출력 미열람 상태에서 독립 라벨링
2. Label scope: 6개 HLF endorsement/validation nondeterminism class
   - nondeterministic timestamps
   - global variable mutation
   - goroutine concurrency hazards
   - map iteration randomness
   - phantom reads
   - iterator resource leaks
3. 판정 기준:
   - **vulnerable**: 6개 class 중 하나가 endorsement-relevant path에 존재 + line-level evidence
   - **safe**: targeted class 없음 (suspicious construct 있어도 ledger-write path 미도달이면 safe)
   - **exclude**: 판정 불가 / ambiguous / out-of-scope
4. 각 파일 기록:
   - repo / commit / file path
   - NCLOC / bytes
   - final label
   - vulnerability class (해당 시)
   - evidence line(s)
   - rationale (1-3문장)
   - ambiguity flag

### 2.3 불일치 처리
- disagreement → 합의 adjudication 1회
- 여전히 불일치 → exclude 처리
- raw agreement + Cohen's κ 기록

### 2.4 라벨 밸런스 확인
- vulnerable ≥ 5, safe ≥ 5 목표
- 한쪽 5개 미만 → FLAG 파일 보충 검토

---

## Phase 3: 모델 실행

### 3.1 실행 범위 (GPT 수렴안)

| 모델 | 프롬프트 | 우선순위 | runs |
|:-----|:---------|:---------|:-----|
| Qwen2.5-Coder-7B | P1 zero-shot | **필수** | 5회 |
| Semgrep default | — | **필수** | 1회 |
| Semgrep HLF-specific | — | **필수** | 1회 |
| Llama-3.1-8B | P1 zero-shot | **권장** | 5회 |
| Qwen2.5-Coder-7B | P2 few-shot | **선택** (fallback) | 5회 |

### 3.2 프로토콜 (기존과 동일)
- temperature = 0.1
- max_tokens = 2048
- classifier = v2
- few-shot examples = 기존 held-out 2개 동일
- n_ctx = 16,384 (GoLiSA 설정 재사용)
- hardware: RTX 3090 Ti, Q4_K_M quantization
- script: 02_run_audit_v3.py, 05_run_traditional_tools.py

### 3.3 결과 확인 후 수정 금지
- prompt 수정 금지
- Semgrep rule 수정 금지
- classifier 수정 금지
- 파일 추가/제거 금지

---

## Phase 4: 결과 보고

### 4.1 보고 지표
- TP / FP / FN / TN (count)
- TPR / TNR (proportion)
- exact 95% Clopper-Pearson CI
- 5-run variability (range)

### 4.2 결과 시나리오별 논문 전략

| 시나리오 | 논문 대응 |
|:---------|:---------|
| 유지 (Qwen 우위) | "diagnostic benchmark 결과가 independent public code에서 재현" |
| 약간 하락 | "specialist advantage 유지, transfer gap 존재" |
| 크게 하락 | "diagnostic benchmark isolates capability; public code reveals transfer gap" |

### 4.3 논문 반영 위치
- Methodology: 1개 문단 "Independent Public-Chaincode Validation"
- Results: 1개 문단 + compact table
- Threats: 1-2문장 수정
- Conclusion: 1문장 보강

---

## 3대 원칙 (ORCHESTRATION 계승)

1. **Label-before-run**: 라벨 확정 후 모델 실행
2. **Protocol consistency**: 기존 실험과 동일 설정
3. **결과 무관 보고**: 좋으면 "trend preserves", 나쁘면 "transfer gap reveals"
