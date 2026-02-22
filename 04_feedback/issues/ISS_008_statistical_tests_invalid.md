# ISS_008: 통계적 검정 부적절 (N=15)

**Status**: open
**Priority**: high
**Source Feedback**: FB_260209_1600_gemini
**Target Sections**: IV, V
**Created**: 2026-02-09

## 문제 설명
N=15인 상황에서 t-test, p-value 등의 통계적 검정 수행은 수학적으로 무의미하다. "Llama 대비 유의미한 성능 향상"이라는 표현을 통계적 근거 없이 사용할 경우 방법론적 결함으로 Reject 사유가 될 수 있다.

## 현재 논문 상태
- v2에서는 t-test/p-value를 사용하지 않고 있음 (비율 직접 보고)
- 다만 "outperforms", "achieves the highest" 등의 표현이 암묵적으로 통계적 우월성을 시사

## 권장 조치
- 통계적 검정을 시도하지 않음을 Methodology에 명시
- "outperforms" → "achieves higher rates in our controlled test" 등 범위 한정 표현 사용
- 비교 문구에 "in this 15-file evaluation" 한정어 추가

## 합의된 통계 전략 (FB_260209_1700_gemini_strategy)

### N=15 실험(RQ1): 질적 분석으로 대체
- 통계 수치 대신 질적 분석(Qualitative Analysis) + 실패 원인 분석(Error Analysis)에 집중
- "확률적 차이"가 아닌 모델 아키텍처에 따른 "결정론적 차이(Deterministic Difference)"임을 강조
- 예: "Llama는 json.Marshal의 키 정렬 기능을 이해하지 못해 오탐" / "Qwen은 time.Now 값이 PutState로 흘러가는지 데이터 흐름 추적 성공"

### Fisher's Exact Test (비상 무기, Rebuttal용)
- 범주형 데이터(성공/실패)에 사용 가능, 소표본에서도 유효
- Qwen(6성공/0실패) vs Llama(1성공/5실패) TNR → p ~ 0.008
- 본문에는 사용하지 않음 — 리뷰어가 통계를 요구할 경우 반박 답변서(Rebuttal Letter)에서만 사용

### GoLiSA 실험(RQ2): 통계 검정 가능
- N=651에서는 t-test 등 통계적 검정 수행 가능
- ISS_013 (GoLiSA 외부 검증) 실행 시 본 이슈의 비판이 실질적으로 해소됨

## 관련 이슈
- ISS_013: GoLiSA 외부 검증 (통계 검정 가능한 규모)
- ISS_014: Core Thesis 재정립 (질적 연구 전환과 직결)
