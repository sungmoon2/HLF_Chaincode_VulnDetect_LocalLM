# ISS_005: "Semantic Analysis" 용어 과대포장

## 메타데이터
- **안건 ID**: ISS_005
- **상태**: open
- **우선순위**: medium
- **출처 피드백**: FB_260209_1500_reviewer
- **대상 섹션**: meta (Title/Abstract), I (Introduction), III (Methodology)
- **대상 파일**: 초안/260209_1245_실험결과_토론섹션_초안작성_v1.tex
- **생성일**: 2026-02-09
- **해결일**:

## 내용
"Semantic Static Analysis"라는 용어가 LLM의 실제 동작과 괴리가 있다:

1. **LLM ≠ 정적 분석기**: LLM은 CFG/DFG를 구성하지 않으며, 확률적 텍스트 생성에 기반
2. **Qwen 정답 근거 불명**: Qwen이 safe_03의 json.Marshal 정렬 특성을 "이해"해서 맞혔는지, 단순히 map iteration을 덜 위험하게 판단하도록 튜닝된 것인지 불명확
3. **증명 수단 부재**: Attention Map 분석이나 심층 질적 분석 없이 "semantic understanding"을 주장

## 조치 사항
- [ ] 용어 순화: "Semantic Static Analysis" → "LLM-driven Vulnerability Detection" 또는 "LLM-assisted Code Audit"
- [ ] 또는 LLM이 실제 문맥을 본다는 증거를 Case Study로 보강 (출력 텍스트에서 추론 경로 추출)
- [ ] Title, Abstract, Introduction, Methodology 전반의 용어 일관 수정
- [ ] Discussion에서 LLM 기반 분석의 한계를 정적 분석과 비교하여 명시

## 연결성
- **선행 안건**: ISS_004 (데이터 오염 여부가 불확실하면 "semantic understanding" 주장이 더 약해짐)
- **후행 안건**: 없음
- **관련 마스터가이드**: 논문_마스터가이드/01_핵심주장/, 논문_마스터가이드/09_용어정의/
