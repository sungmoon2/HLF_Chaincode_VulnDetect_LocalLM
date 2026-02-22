# ISS_009: Semgrep Strawman Baseline 방어 필요

**Status**: open
**Priority**: high
**Source Feedback**: FB_260209_1600_gemini
**Target Sections**: III, IV, V
**Created**: 2026-02-09

## 문제 설명
Semgrep 탐지율 0%는 도구가 무능해서가 아니라, 저자가 HLF 합의 전용 Custom Rules를 작성하지 않았기 때문이라는 비판이 가능하다. "의도적으로 약한 상대를 골라 이기려 했다(Strawman Argument)"로 오해될 수 있다.

## 현재 논문 상태
- v2에서 Semgrep을 "p/security-audit" 기본 config로 실행
- "out-of-the-box experience" 프레이밍이 불충분

## 권장 조치
- Methodology에 "default ruleset without custom HLF-specific rules" 명시
- "일반 개발자가 추가 설정 없이 사용하는 시나리오"임을 명확히 방어
- Discussion에서 Custom Rules 작성 시 성능 향상 가능성 인정
- Threats to Validity에 "Baseline tool configuration" 항목 추가

## 합의된 방어 논리 (FB_260209_1700_gemini_strategy)

### "일반 개발자(Generalist)" 관점 프레이밍
- Semgrep을 기본 룰셋(Default Ruleset, `p/security-audit`)으로 제한한 것은 의도적
- 이유: 보안 전문 지식이 없는 일반 개발자가 "설치하자마자(Out-of-the-box)" 사용하는 시나리오를 시뮬레이션
- 비교 대상: "이론적 최대 성능(Theoretical Maximum)"이 아닌 "현실적 사용 경험(Raw Utility)"

### 논문 삽입 위치
- Methodology (Section III-F): "We intentionally restricted Semgrep to its default p/security-audit ruleset to simulate the real-world experience of a generalist developer who lacks the expertise to write custom security rules."
- Discussion: Custom Rules 작성 시 성능 향상 가능성 인정 (Threats to Validity)

## 관련 이슈
- ISS_014: Core Thesis 재정립 ("Out-of-the-box" 논리와 일관)
