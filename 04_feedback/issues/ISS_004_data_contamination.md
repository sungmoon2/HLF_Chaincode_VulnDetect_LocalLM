# ISS_004: 데이터 오염/암기 가능성 (Data Contamination)

## 메타데이터
- **안건 ID**: ISS_004
- **상태**: open
- **우선순위**: high
- **출처 피드백**: FB_260209_1500_reviewer
- **대상 섹션**: IV (Results), V (Discussion)
- **대상 파일**: 초안/260209_1245_실험결과_토론섹션_초안작성_v1.tex, 02_resources/dataset/*.go
- **생성일**: 2026-02-09
- **해결일**:

## 내용
Qwen2.5-Coder의 100% 탐지율이 진정한 "추론(Reasoning)"인지, 학습 데이터 암기(Memorization)인지 구분할 수 없다:

1. **학습 데이터 포함 가능성**: Qwen2.5-Coder는 코딩 특화 모델로, HLF 공식 문서/샘플 코드가 학습 데이터에 포함되었을 가능성 높음
2. **Canonical Anti-patterns 유사성**: 저자가 만든 취약 코드가 공식 문서 예제와 구조적으로 유사할 경우, 패턴 매칭에 불과할 수 있음
3. **반박 근거 부재**: 변형된(Obfuscated) 코드나 학습 데이터에 없을 법한 패턴으로의 테스트가 없음

## 조치 사항
- [ ] 취약 코드 변형 실험: 변수명/함수명 난독화(Obfuscation) 후 재감사
- [ ] 또는 HLF 공식 예제와의 코드 유사도(Code Similarity) 측정하여 차별성 입증
- [ ] Discussion에서 Data Contamination을 Threat to Validity로 명시적 논의
- [ ] Qwen의 출력에서 "추론 과정"이 보이는 사례를 Qualitative Case Study로 제시

## 연결성
- **선행 안건**: ISS_001 (데이터셋이 확장되면 오염 우려도 희석됨)
- **후행 안건**: ISS_005 (추론 여부 불확실 → "Semantic Analysis" 용어 과대포장과 연결)
- **관련 마스터가이드**: 논문_마스터가이드/06_분석프레임워크/, 논문_마스터가이드/07_제약조건_타당성위협/
