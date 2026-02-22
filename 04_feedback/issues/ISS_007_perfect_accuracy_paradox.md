# ISS_007: 100% 정확도 역설 (Overfitting Signal)

**Status**: open
**Priority**: high
**Source Feedback**: FB_260209_1600_gemini
**Target Sections**: IV, V
**Created**: 2026-02-09

## 문제 설명
ML 연구에서 테스트 셋 정확도 100%는 긍정적 신호가 아니라 "데이터 오염(Data Contamination)" 또는 "문제가 지나치게 단순함(Trivial Task)"을 시사한다. Qwen이 학습 과정에서 HLF 공식 문서/예제 코드를 암기(Memorization)했을 가능성을 배제할 수 없다.

## 기존 방어 기제
- 난독화 실험에서 Qwen TPR 100% → 78%, TNR 100% → 67%로 하락 (ISS_004 참조)
- 이는 부분적 패턴 매칭 의존을 확인하나, 암기 의심을 일부 해소

## 권장 조치
- "100% 달성" 강조 대신 난독화 결과와 함께 맥락화
- Discussion에서 Data Contamination 가능성을 더 강하게 인정
- 100%가 아닌 "왜 모델이 맞았는가"에 대한 원인 분석 집중

## 합의된 해석 가이드 (FB_260209_1700_gemini_strategy)

### 건조한 기술 원칙
- "완벽하다" 자랑 금지
- "통제된 실험 환경(Controlled Setting) 내에서 의도된 패턴을 식별하는 데 성공했다"로 건조하게 기술
- 난독화 실험 결과(78%)를 들어 "변수명에 의존하는 경향이 있음"을 자체 비판 → 리뷰어 신뢰도 상승

### GoLiSA 외부 검증 효과
- GoLiSA 651개 테스트 시 100%가 나오지 않을 가능성이 높음
- 100%가 아닌 현실적 수치가 나올 경우 오히려 과적합/암기 의심을 해소하는 근거가 됨

## 관련 이슈
- ISS_013: GoLiSA 외부 검증 (100% 역설 해소 가능)
- ISS_014: Core Thesis 재정립 (톤 수정과 직결)
