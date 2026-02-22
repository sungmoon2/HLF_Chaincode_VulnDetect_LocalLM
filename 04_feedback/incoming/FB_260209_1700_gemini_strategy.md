# FB_260209_1700_gemini_strategy

**Source**: gemini (Gemini 모델 기반 전략적 논문 방향 논의)
**Date**: 2026-02-09T17:00:00
**Target Draft**: 260209_1620_전체논문_v3.tex
**Target Sections**: I, II, III, IV, V, VI, meta
**Severity**: major

---

## 배경

FB_260209_1600_gemini (ISS_006~009)의 비판 포인트를 수용한 후, 구체적인 방어 전략과 논문 방향 전환에 대해 Gemini와 심층 전략 논의를 진행함. 딥리서치(Deep Research)를 통해 기존 데이터셋(GoLiSA 등)의 존재를 확인하고, 이에 맞춰 논문의 포지셔닝을 재설정함.

---

## 핵심 논의 결과 (사실 기반)

### 1. GoLiSA 데이터셋 발견 (딥리서치 결과)
- **사실**: GoLiSA Artifact (ECOOP 2023) — ~651개 실제 HLF 체인코드, 비결정론 주석
- **출처**: Zenodo (DOI: 10.5281/zenodo.7896323), lisa-analyzer/go-lisa GitHub
- **구성**: GitHub에서 수집(Mined), 정적 분석 도구 검증용
- **결론**: "HLF 합의 오류 데이터셋이 전무하다"는 주장은 기각됨
- **추가 발견**: VulFinder (IEEE TSE 2025), zm-stack/Chaincode (PDC 프라이버시), Revive-CC (20개 파일)

### 2. 데이터셋 포지셔닝 변경
- **변경 전**: "Standard Benchmark" (위험 — 건방진 인상)
- **변경 후**: "Curated Micro-benchmark" 또는 "Diagnostic Unit-Test Suite"
- **차별화 논리**: GoLiSA는 수집형(Mined) Raw Corpus로 재현율(Recall) 측정에 유리 / 우리 15개는 설계형(Engineered) Adversarial Set으로 정밀도(Precision) 및 추론 능력(Reasoning) 검증에 특화
- **관계**: 경쟁이 아닌 상호 보완(Complementary)

### 3. 통계적 검증 전략
- **합의**: N=15에서 t-test, p-value 등 통계적 검정을 수행하지 않음
- **대체**: 질적 분석(Qualitative Analysis) + 실패 원인 분석(Error Analysis)
- **논거**: 100% vs 0% (Semgrep), 100% vs 17% (Llama)의 차이는 결정론적(Deterministic)이며 통계 검증이 무의미할 정도로 자명
- **비상 무기(Rebuttal용)**: Fisher's Exact Test (Qwen vs Llama TNR → p ~ 0.008) — 본문이 아닌 반박 답변서에만 사용

### 4. Privacy Paradox (신규 논거)
- **논리**: HLF = 기업용 프라이빗 블록체인 → 체인코드에 영업 비밀 포함 → 클라우드 LLM API에 코드 전송 시 데이터 주권 위험 + 규정 위반 → 로컬 sLM이 필수
- **결론**: 코드 특화 sLM(Qwen)이 보안(Privacy)과 성능(Performance)의 트레이드오프를 해결

### 5. 핵심 주장(Core Thesis) 재정립
- **한 문장**: "HLF의 치명적인 '합의 불일치(Consensus Divergence)' 오류는 기존 도구로는 탐지할 수 없으나, 로컬에 배포된 '코드 특화 소형 모델(Specialist sLM)'을 활용하면 클라우드 비용이나 프라이버시 침해 없이 의미론적(Semantic)으로 정확히 탐지할 수 있다."
- **3대 기둥**:
  1. 문제의 특수성: HLF 버그는 Crash가 아닌 Silent Failure (합의 불일치) → 기존 도구 탐지 0%
  2. 해결책 검증: 키워드가 아닌 문맥(Context) 이해가 핵심 → Llama(TNR 17%) vs Qwen(TNR 100%)
  3. 현실적 제약 해결: Privacy Paradox → 로컬 sLM이 보안과 성능 양립

### 6. 논문 톤 수정
- "Superior Performance" → "Competitive Accuracy" 또는 "Feasibility Study"
- "Outperforms" → "demonstrates comparable reasoning"
- "Standard Benchmark" → "Curated Micro-benchmark"
- 논문 성격: "Quantitative Benchmark" → "Qualitative Feasibility Study on Semantic Reasoning"

### 7. 100% 탐지율 해석 주의
- "완벽하다" 자랑 금지
- "통제된 실험 환경(Controlled Setting) 내에서 의도된 패턴을 식별하는 데 성공했다"로 건조하게 기술
- 난독화 실험 결과(78%)를 들어 "변수명에 의존하는 경향이 있음"을 자체 비판

### 8. GoLiSA 외부 검증 실험 제안
- GoLiSA 651개 다운로드 (Zenodo/GitHub)
- 로컬 모델(Qwen, Llama) + Semgrep으로 전수 조사 (비용 0원)
- 클라우드 모델은 30~50개 샘플링 테스트 (비용 절감)
- N=651에서는 통계적 검정도 가능해짐

### 9. 참고문헌 조사 방향 (딥리서치 프롬프트 4개, 진행 중)
- 카테고리 1: HLF 아키텍처/XOV — 합의 불일치의 이론적 근거 (EuroSys 2018 등)
- 카테고리 2: 정적 분석/기존 벤치마크 — GoLiSA/VulFinder 팩트 체크 (ECOOP 2023 등)
- 카테고리 3: LLM vs 정적 분석 — LLM이 FPR에서 우위라는 근거 (ICSE/FSE/ASE 등)
- 카테고리 4: 코드 특화 sLM 성능 — 소형 모델이 대형 모델과 대등 (Qwen2.5-Coder TR 등)
- 제약조건: 무료 접근 가능(Open Access/arXiv), PDF 원문 텍스트 추출 가능

### 10. Claude 4.5 모델 버전 확인
- **팩트**: 현재 시점(2026년 2월) Claude 4.5 (Haiku/Sonnet/Opus)는 실존
- Gemini가 한때 "Claude 4.5는 없다"고 잘못 발언 → 이후 직접 수정함
- 논문에서 Claude 4.5 버전을 절대 변경하지 않음

---

## Derived Issues
- ISS_010: GoLiSA 데이터셋 Related Work 인용 필수
- ISS_011: Privacy Paradox 논거 추가
- ISS_012: Deep Research 참고문헌 확보 (4개 카테고리)
- ISS_013: GoLiSA 외부 검증 실험
- ISS_014: Core Thesis 재정립 및 논문 톤 수정
