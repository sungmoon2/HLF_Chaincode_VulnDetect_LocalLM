# ISS_020: 과장/추정 표현 일괄 완화 (7건)

## 메타데이터
- **안건 ID**: ISS_020
- **상태**: resolved
- **우선순위**: critical
- **출처 피드백**: FB_260211_advisor_gpt52
- **대상 섹션**: I, III, IV, V, VI
- **대상 파일**: 260211_v31_GPT52교정.tex
- **생성일**: 2026-02-11
- **해결일**: 2026-02-11

## 내용
GPT 5.2 리뷰에서 "과장/추정으로 읽힐 수 있어 문장 톤을 조정하는 편이 안전" 지적. 총 7건의 1:1 수정 제안.

관련 1:1 수정 제안 7건:
- **C3** (p.1, Section I): "zero marginal cost" → "low marginal cost per analysis"
- **C4** (p.1, Section I): "Traditional static analysis tools operate on syntactic patterns and lack the domain knowledge" → "General-purpose static analysis tools typically operate on syntactic patterns and may not capture HLF endorsement semantics without Fabric-specific rules or analyses"
- **C7** (p.2, Section III-D): "structurally prevents self-contradictory responses" → "constrain responses to a fixed schema and reduce label inconsistency"
- **C12** (p.4, Section IV-F): "the fastest among all models tested" 삭제 (클라우드 모델 지연시간은 네트워크/쿼터/서버상태에 좌우)
- **C13** (p.4, Section V-A): "appears to have internalized Go API semantics sufficient to trace" → "outputs are consistent with tracing whether nondeterministic values reach PutState, suggesting stronger data-flow sensitivity than keyword-triggered heuristics"
- **C14** (p.5, Section V-D): "Qwen's training data likely includes" → "Qwen's pretraining data may include HLF documentation and public chaincode examples, but we do not have direct evidence of dataset overlap"
- **C16** (p.6, Section VI): "A code-specialist sLM achieves discrimination parity with prompt-engineered cloud models" → "Within our evaluation, a code-specialist sLM matches the best-performing prompt-engineered cloud models on the micro-benchmark"

## 조치 사항
- [x] C3: "zero marginal cost"→"low marginal cost per analysis" — v31 반영
- [x] C4: "Traditional"→"General-purpose", "lack"→"may not capture" — v31 반영
- [x] C7: "structurally prevents"→"reduces label inconsistency" — v31 반영
- [x] C12: "the fastest among all models tested" 삭제 — v31 반영
- [x] C13: "internalized"→관측 기반 "outputs are consistent with" — v31 반영
- [x] C14: "likely includes"→"may include...no direct evidence" — v31 반영
- [x] C16: "parity"→"Within our evaluation...matches...suggesting" — v31 반영

## 해결 내역
v31 (260211_v31_GPT52교정.tex)에서 7건 전수 적용.

## 연결성
- **선행 안건**: ISS_019 (용어 정합성 — 톤 조정 방향과 연동)
- **후행 안건**: 없음
- **관련 기존 안건**: ISS_005 (용어 과대포장 — resolved), ISS_007 (100% 정확도 역설 — in_progress)
