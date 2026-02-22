# ISS_011: Privacy Paradox 논거 추가

**Status**: open
**Priority**: high
**Source Feedback**: FB_260209_1700_gemini_strategy
**Target Sections**: I, V, VI
**Created**: 2026-02-09

## 문제 설명
현재 논문에는 "왜 로컬 sLM을 써야 하는가"에 대한 현실적 동기(Practical Motivation)가 부족하다. "Privacy Paradox"를 명시적으로 서술하여 로컬 sLM 사용의 핵심 명분을 확보해야 한다.

## Privacy Paradox 논리 흐름 (합의됨)
1. HLF = 기업용(Enterprise) 프라이빗 블록체인 → 체인코드에 영업 비밀/기밀 비즈니스 로직 포함
2. 클라우드 LLM(GPT/Claude/Gemini) API에 코드 전송 → 데이터 주권(Data Sovereignty) 위험 + 규정 위반
3. 기존 로컬 모델(Llama 등 범용 모델) → 안전하지만 성능 부족 (TNR 17%)
4. 우리의 발견: 코드 특화 sLM(Qwen)은 로컬(에어갭, Air-gapped) 환경에서 클라우드 SOTA와 대등한 탐지율 달성 → "보안과 성능의 양립"

## 딥리서치에서 확인된 근거
- 딥리서치 보고서 Section 6 (Gap Analysis)에서 "Privacy Paradox" 개념 확인: "The primary gap in HLF research is that the system is designed for privacy... High-value chaincodes exist inside private consortia and are never exposed to public repositories."
- 이 사실은 (1) 공개 데이터셋이 실제 기업 환경을 대변하지 못하는 이유이자, (2) 클라우드 모델을 사용할 수 없는 현실적 제약의 근거

## 권장 조치
- Introduction에 Privacy Paradox를 연구 동기(Motivation)로 추가
- Discussion에 비용-프라이버시 트레이드오프 분석 (ANALYSIS_REPORT.md Section 2.6 데이터 활용)
- Conclusion에서 실용적 함의(Practical Implication)로 강조

## 관련 이슈
- ISS_014: Core Thesis 재정립 (Privacy Paradox가 3대 기둥 중 하나)
