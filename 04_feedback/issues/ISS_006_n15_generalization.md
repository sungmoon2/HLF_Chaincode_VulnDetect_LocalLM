# ISS_006: N=15 통계적 일반화 불가

**Status**: open
**Priority**: critical
**Source Feedback**: FB_260209_1600_gemini
**Target Sections**: III, IV, V, VI, meta
**Created**: 2026-02-09

## 문제 설명
N=15(취약 9, 정상 6)는 통계적으로 유의미한 결론을 내리기에 부족하며, "일반적인 HLF 체인코드"에 적용 가능한 일반화(Generalization)가 불가능하다. 이 논문은 "성능 평가(Performance Evaluation)" 논문 요건을 충족하지 못한다.

## ISS_001과의 관계
ISS_001은 N=9 → N=15 확장 + Feasibility Study 리프레이밍으로 resolved 처리되었으나, 본 이슈는 N=15에서도 여전히 통계적 일반화 불가를 지적한다. ISS_001의 해결이 부분적이었음을 확인.

## 권장 조치
- 논문을 "Qualitative Case Study"로 포지셔닝
- 수치 비교보다 원인 분석(Error Analysis)에 집중
- "왜 Llama는 틀렸고 Qwen은 맞았는가"에 대한 질적 분석 강화
- Threats to Validity에서 N=15 한계를 첫 번째로 명시

## 합의된 방어 전략 (FB_260209_1700_gemini_strategy)

### 데이터셋 명칭 변경
- "Standard Benchmark" 사용 금지 → "Curated Micro-benchmark" 또는 "Diagnostic Unit-Test Suite"
- "Standard"라고 자칭하면 "건방지다(Arrogant)"는 인상을 줄 위험

### 커버리지 기반 방어
- 15개는 HLF 공식 문서의 6가지 합의 오류 유형(6 Deadly Sins) + 3가지 구조적 변형(Structural Variants) + 6가지 함정(Benign Traps)으로 구성
- "양(Quantity)"이 아닌 "조합적 커버리지(Combinatorial Coverage)"로 정당화

### GoLiSA 차별화
- 딥리서치에서 GoLiSA (ECOOP 2023, ~651개)의 존재 확인됨
- GoLiSA는 수집형(Mined) Raw Corpus로 재현율(Recall) 측정에 유리
- 우리 15개는 설계형(Engineered) Adversarial Set으로 정밀도(Precision) 및 추론 능력(Reasoning) 검증에 특화
- 두 데이터셋은 상호 보완(Complementary) 관계

### 외부 검증 실험 (ISS_013 참조)
- GoLiSA 651개로 추가 실험 시 N=651에서 통계적 검정 가능
- 이 실험이 수행되면 본 이슈의 비판을 실질적으로 해소

## 관련 이슈
- ISS_010: GoLiSA 인용
- ISS_013: GoLiSA 외부 검증 실험
- ISS_014: Core Thesis 재정립
