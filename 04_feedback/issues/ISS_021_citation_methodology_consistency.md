# ISS_021: 서지 오류 + 방법론 서술 일관성 (4건)

## 메타데이터
- **안건 ID**: ISS_021
- **상태**: resolved
- **우선순위**: critical
- **출처 피드백**: FB_260211_advisor_gpt52
- **대상 섹션**: III, IV
- **대상 파일**: 260211_v31_GPT52교정.tex
- **생성일**: 2026-02-11
- **해결일**: 2026-02-11

## 내용
GPT 5.2 리뷰에서 서지 오류(명백한 인용 오류), 방법론 서술 불일치(context window), Table 각주 불일치, Table 오타를 지적. 4건의 1:1 수정 제안.

관련 1:1 수정 제안 4건:
- **C5** (p.2, Section III-C): llama-cpp-python을 [2],[3](모델 기술보고서)로 인용하는 것은 명백한 서지 오류. "[2], [3]" 인용 제거, 별도 참고문헌 항목 추가 권장
- **C6** (p.2, Section III-C): "All models use temperature 0.1, max tokens 2,048, context window 4,096" → "Unless otherwise stated, we use temperature 0.1 and max tokens 2,048 with n_ctx=4,096 (we increase n_ctx to 16,384 for the GoLiSA corpus evaluation in Section III-G)"
- **C11** (p.3, Table II 하단 각주): "Zero-shot and few-shot values are from the initial experiment" → "All prompt strategies were repeated five times; for cloud models we report the median across runs when variation was observed (‡)" (Section III-H의 5회 반복 서술과 충돌)
- **C17** (p.2, Table I): "Gemini 2.5 Prob" → "Gemini 2.5 Pro" (오타)

## 조치 사항
- [x] C5: llama-cpp-python \cite{} [2],[3] 인용 제거 (3곳: III-B, V-A 2곳) — v31 반영
- [x] C6: "All models use...4,096"→"Unless otherwise stated...increase to 16,384" — v31 반영
- [x] C11: Table II 각주 "initial experiment"→"repeated five times" — v31 반영
- [x] C17: Table I "Gemini 2.5 Pro" — v30 소스에서 이미 정상. 해당없음 확인

## 해결 내역
v31 (260211_v31_GPT52교정.tex)에서 C5, C6, C11 적용. C17은 v30에서 이미 정상(GPT PDF 오독).

## 연결성
- **선행 안건**: 없음
- **후행 안건**: 없음
- **관련 기존 안건**: ISS_012 (참고문헌 확보 — resolved)
