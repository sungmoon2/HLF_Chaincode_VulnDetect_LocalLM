# ISS_014: Core Thesis 재정립 및 논문 톤 수정

**Status**: open
**Priority**: critical
**Source Feedback**: FB_260209_1700_gemini_strategy
**Target Sections**: I, II, III, IV, V, VI, meta
**Created**: 2026-02-09

## 문제 설명
현재 논문(v3)은 여전히 "성능 벤치마크" 톤이 잔존하며, Gemini 전략 논의에서 합의된 "Qualitative Feasibility Study" 방향과 불일치한다. 핵심 주장(Core Thesis)을 재정립하고 전체 논문의 톤을 통일해야 한다.

## 합의된 Core Thesis (한 문장)
"HLF의 치명적인 '합의 불일치(Consensus Divergence)' 오류는 기존 도구로는 탐지할 수 없으나, 로컬에 배포된 '코드 특화 소형 모델(Specialist sLM)'을 활용하면 클라우드 비용이나 프라이버시 침해 없이 의미론적(Semantic)으로 정확히 탐지할 수 있다."

## 합의된 3대 논리적 기둥
1. **문제의 특수성 (Silent Failure)**: HLF 버그는 Crash가 아닌 노드 간 결과 불일치. 기존 Fuzzing/Linter(Semgrep) 탐지율 0%.
2. **해결책 검증 (Context over Keywords)**: time.Now()의 유무가 아닌 "원장에 기록되는가"라는 문맥(Context) 파악이 핵심. Llama(키워드 의존, TNR 17%) vs Qwen(데이터 흐름 추적, TNR 100%).
3. **현실적 제약 해결 (Privacy Paradox)**: 기업용 코드는 클라우드 전송 불가 → 로컬 sLM이 보안(Privacy)과 성능(Performance)의 트레이드오프 해결.

## 톤 수정 사항 (합의됨)

### 삭제/변경할 표현
| 현재 (위험) | 변경 (안전) |
|:------------|:-----------|
| "Superior Performance" | "Competitive Accuracy" 또는 "Feasibility Study" |
| "Outperforms" | "demonstrates comparable reasoning" |
| "Standard Benchmark" | "Curated Micro-benchmark" |
| "demonstrates" (절대적) | "provides preliminary evidence" |
| "100% 달성" 강조 | "통제된 환경 내에서 의도된 패턴을 식별하는 데 성공" |

### 논문 성격 전환
- **변경 전**: Quantitative Benchmark (누가 더 성능이 뛰어난가)
- **변경 후**: Qualitative Feasibility Study on Semantic Reasoning (sLM이 HLF 합의 오류의 문맥을 이해할 수 있는가)

### 100% 탐지율 해석 가이드
- "완벽하다" 자랑 금지
- "통제된 실험 환경(Controlled Setting) 내에서 의도된 패턴을 식별하는 데 성공했다"로 건조하게 기술
- 난독화 실험 결과(78%)를 들어 "변수명에 의존하는 경향이 있음"을 자체 비판 → 신뢰도 상승

## 권장 조치
- Title/Abstract: 톤 다운 적용
- Introduction: Core Thesis 3대 기둥 기반으로 재구성
- Methodology: "Curated Micro-benchmark" 정의 삽입
- Discussion: 질적 Error Analysis 강화, 100% 건조한 해석
- Threats to Validity: N=15 한계 + 통계 검정 미수행 근거 명시
- Conclusion: "provides preliminary evidence", Future Work 강화

## 관련 이슈
- ISS_006: N=15 일반화 (Feasibility Study 전환과 직결)
- ISS_007: 100% 역설 (해석 가이드와 직결)
- ISS_008: 통계적 검정 (질적 분석 전환과 직결)
- ISS_009: Semgrep 방어 ("Out-of-the-box" 논리와 일관)
- ISS_010: GoLiSA 인용 (차별화 논리와 연결)
- ISS_011: Privacy Paradox (3대 기둥 중 하나)
