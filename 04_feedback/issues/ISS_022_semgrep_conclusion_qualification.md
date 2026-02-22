# ISS_022: Semgrep/결론부 조건 명시

## 메타데이터
- **안건 ID**: ISS_022
- **상태**: resolved
- **우선순위**: high
- **출처 피드백**: FB_260211_advisor_gpt52
- **대상 섹션**: III, VI
- **대상 파일**: 260211_v31_GPT52교정.tex
- **생성일**: 2026-02-11
- **해결일**: 2026-02-11

## 내용
GPT 5.2 리뷰에서 Semgrep baseline 서술의 가독성/정확성 강화 및 결론부에서 "default ruleset" 조건 재명시 필요 지적. 결론에서 조건이 빠지면 "Semgrep은 못한다"로 과잉 일반화로 읽힐 수 있음.

관련 1:1 수정 제안 2건:
- **C8** (p.2, Section III-F): 수동태 → 능동태 전환 + "default ruleset" 명확화. "Semgrep 1.151.0 [10] with the default p/security-audit ruleset is run on all 15 files without custom HLF-specific rules." → "We run Semgrep 1.151.0 [10] with the default p/security-audit ruleset on all 15 files without any HLF-specific custom rules."
- **C15** (p.6, Section VI): "Semgrep detects zero consensus-layer vulnerabilities on both datasets." → "With the default p/security-audit ruleset, Semgrep detects zero consensus-layer vulnerabilities on both datasets."

## 조치 사항
- [x] C8: Semgrep 능동태 전환 + "As a general-purpose static analysis baseline" — v31 반영
- [x] C15: 결론 "With the default p/security-audit ruleset" 조건 추가 — v31 반영

## 해결 내역
v31 (260211_v31_GPT52교정.tex)에서 C8, C15 적용.

## 연결성
- **선행 안건**: 없음
- **후행 안건**: 없음
- **관련 기존 안건**: ISS_009 (Semgrep Strawman Baseline 방어 — resolved)
