# ISS_019: "consensus-layer" 용어 정합성 + 평가 지표 정의 명확화

## 메타데이터
- **안건 ID**: ISS_019
- **상태**: resolved
- **우선순위**: critical
- **출처 피드백**: FB_260211_advisor_gpt52
- **대상 섹션**: meta, I, III
- **대상 파일**: 260211_v31_GPT52교정.tex
- **생성일**: 2026-02-11
- **해결일**: 2026-02-11

## 내용
GPT 5.2 리뷰에서 가장 강하게 지적한 용어 문제. Fabric에서 체인코드는 ordering service의 합의 레이어에서 실행되지 않으며, 문제의 핵심은 endorsement/validation 단계의 비결정성. "consensus-layer vulnerability"는 블록체인 연구 커뮤니티에서 "합의 프로토콜(PBFT/PoW 등) 취약점"으로 오해될 소지가 큼.

관련 1:1 수정 제안 4건:
- **C1** (p.1, Section I, 1문단 2번째 문장): "consensus-layer vulnerabilities" → "endorsement/validation nondeterminism vulnerabilities (sometimes loosely referred to as 'consensus-layer' issues in HLF)"
- **C2** (p.1, Abstract, 1문장): "consensus-layer vulnerabilities...that cause" → "endorsement/validation nondeterminism vulnerabilities...that can cause...without explicit runtime errors"
- **C9** (p.3, Section III-H, TPR 정의): "fraction of vulnerable files with at least one consensus-relevant finding reported" → "proportion of vulnerable files for which the model reports at least one finding mapped to the targeted consensus-layer classes"
- **C10** (p.3, Section III-H, TNR 정의): "fraction of benign files correctly identified as safe" → "proportion of benign-trap files for which the model outputs a final 'safe' verdict under our consensus-only labeling"

추가 구조적 권고:
- 논문 전체에서 "consensus-layer"를 유지할지, "endorsement/validation nondeterminism"로 전환할지 결정 후 일관 적용

## 조치 사항
- [x] C1 적용: Intro에 용어 정의 3문장 신설, Abstract 용어 교체 — v31 반영
- [x] C2 적용: Abstract "cause"→"can lead to", "explicit" 추가 — v31 반영
- [x] C9 적용: TPR 정의 "proportion...mapped to targeted consensus-layer classes" — v31 반영
- [x] C10 적용: TNR 정의 "benign-trap...consensus-only labeling" — v31 반영
- [x] 용어 통일: Intro에서 약칭 정의 선언, 이후 본문에서 shorthand 사용으로 결정

## 해결 내역
v31 (260211_v31_GPT52교정.tex)에서 C1, C2, C9, C10 전수 적용. Intro L85에 "we use the term 'consensus-layer vulnerabilities' as shorthand" 정의 선언. 이후 본문 잔존 "consensus-layer" 사용은 약칭으로 정당화.

## 연결성
- **선행 안건**: 없음
- **후행 안건**: ISS_020 (과장 표현 완화와 톤 일관성 연동)
- **관련 기존 안건**: ISS_005 (용어 과대포장 — resolved), ISS_014 (Core Thesis 재정립 — resolved)
