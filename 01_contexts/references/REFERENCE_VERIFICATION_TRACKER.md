# Reference Verification Tracker (v6 논문 기준)

> **목적**: 논문 참고문헌에 대해 사실 기반 여부를 체계적으로 검증
> **작성일**: 2026-02-09
> **최종 갱신**: 2026-02-09 — **v6 전체 검증 완료 (Gemini DR 6건 + IEEE Xplore 직접 확인)**
> **검증 도구**: Gemini Deep Research 6단계 프롬프트 + Gemini DR v6 재검증 6건 + IEEE Xplore
> **결과**: v5(18개) → v6(15개): 불필요 3개 삭제 (chen2023chatgpt, zhang2023prompt, zhu2024deepseek). 15개 전체 할루시네이션 0건
> **현재 논문**: 260209_2140_전체논문_v6_불필요참고문헌삭제.tex (15개 참고문헌)

---

## 분류 기준

| 구분 | 설명 |
|:-----|:-----|
| **USER** | 사용자가 직접 제공/확인한 레퍼런스 (실험에 사용한 모델/도구/데이터셋) |
| **CLAUDE** | Claude Code가 논문 작성 과정에서 추가한 레퍼런스 (독립 검증 필요) |

---

## A. 사용자 제공 레퍼런스 (4개) — 실험 인프라 직접 사용

| # | BibTeX Key | 제목 | 출처 | 검증 방법 |
|:--|:-----------|:-----|:-----|:----------|
| [1] | androulaki2018fabric | Hyperledger Fabric: A Distributed Operating System for Permissioned Blockchains | EuroSys 2018, arXiv:1801.10228 | HLF 공식 논문, 실험 기반 |
| [2] | hui2024qwen25coder | Qwen2.5-Coder Technical Report | arXiv:2409.12186 | 모델 직접 다운로드/사용 |
| [3] | grattafiori2024llama3 | The Llama 3 Herd of Models | arXiv:2407.21783 | 모델 직접 다운로드/사용 |
| [14] | semgrep | Semgrep: Lightweight static analysis for many languages | https://semgrep.dev | 도구 직접 설치/실행 |

---

## B. Claude 추가 레퍼런스 (15개) — Gemini Deep Research 검증 대상

### B-1. HLF 전용 분석 도구 (딥리서치 1단계)

| # | BibTeX Key | 제목 (논문 내 서술) | 출처 (논문 기재) | 논문 내 주장 | DR 검증 항목 | 검증 결과 |
|:--|:-----------|:-------------------|:----------------|:------------|:------------|:---------|
| [15] | olivieri2023golisa | Information Flow Analysis for Detecting Non-Determinism in Blockchain | ECOOP 2023, DOI:10.4230/LIPIcs.ECOOP.2023.23 | "abstract interpretation으로 정보 흐름 분석", "GitHub에서 수집한 실제 체인코드 코퍼스 사용" | DOI 실존 확인, 방법론(abstract interpretation) 일치 여부, 데이터셋 설명 일치 여부 | **1단계 검증 완료** |
| [16] | li2025vulfinder | VulFinder: Exploring Chaincode Vulnerabilities... Knowledge Graph Based Defect Pattern Matching | IEEE TSE 2025, DOI:10.1109/TSE.2025.3605379 | "지식 그래프 기반 결함 패턴 매칭", "22개 카테고리 체인코드 취약점 탐지" | 출판 연도 2025 확인, IEEE TSE 게재 확인, Knowledge Graph 기술 확인 | **1단계 검증 완료 (저자명 불일치 발견)** |

### B-2. LLM 취약점 탐지 통계 수치 (딥리서치 2단계)

| # | BibTeX Key | 논문 내 인용 수치 | 검증 포인트 | 검증 결과 |
|:--|:-----------|:-----------------|:-----------|:---------|
| [4] | khare2023understanding | "16개 LLM, 5,000개 코드 샘플, 평균 정확도 62.8%" | 16 LLMs 맞는지, 5,000 샘플 맞는지, 62.8% 정확도 원문 대조 | **2단계 검증 완료 — 모든 수치 일치** |
| [5] | ullah2023llms | "SecLLMHolmes 프레임워크", "변수명 변경 시 GPT-4 오답률 증가" | SecLLMHolmes 이름 확인, 변수명 변경 실험 존재 확인, GPT-4 특정 여부 | **2단계 검증 완료 — 모든 주장 일치 (GPT-4 17% 오답, PaLM2 26%)** |
| [8] | cheshkov2023evaluation | "ChatGPT가 dummy classifier보다 나은 성능을 보이지 못함" | 정확한 결론 문구 원문 대조, "dummy classifier" 표현 사용 여부 | **2단계 검증 완료 — AUC 0.51 (dummy 0.50), 통계적 동등** |
| [9] | ding2024vulnerability | "7B 모델 F1: BigVul 68.26% → PrimeVul 3.09%" | 모델 크기 7B 맞는지, BigVul/PrimeVul 데이터셋 맞는지, 수치 68.26%/3.09% 원문 대조 | **2단계 검증 완료 — StarCoder2 7B, 모든 수치 일치** |

### B-3. 2025년 최신 연구 (딥리서치 3단계)

| # | BibTeX Key | 논문 내 인용 | 검증 포인트 | 검증 결과 |
|:--|:-----------|:------------|:-----------|:---------|
| [17] | zahan2024llm | "GPT-4 vs CodeQL for npm malware detection", "precision 0.99, only 3 false positives vs CodeQL's 684" | ICSE 2025 게재 확정 여부 (vs arXiv only), GPT-4 precision 0.99 원문 대조, CodeQL 684 FP 원문 대조 | **3단계 검증 완료 — 모든 수치 일치** |
| [18] | li2025iris | "IRIS: neuro-symbolic approach, LLM이 CodeQL용 source-sink 스펙 추론", "탐지 취약점 27→55로 2배" | ICLR 2025 게재 확정 여부 (vs arXiv only), 27→55 수치 원문 대조, neuro-symbolic 방법론 확인 | **3단계 검증 완료 — 모든 수치 일치** |

### B-4. 모델 벤치마크 수치 (딥리서치 4단계)

| # | BibTeX Key | 논문 내 인용 수치 | 검증 포인트 | 검증 결과 |
|:--|:-----------|:-----------------|:-----------|:---------|
| [2]* | hui2024qwen25coder | "88.4% Pass@1 on HumanEval", "Llama-3-70B-Instruct (84.1%) 능가" | 88.4% 수치 원문 대조, 84.1% 비교 대상 원문 대조 | **4단계 검증 완료 — 모든 수치 일치** |
| [19] | zhu2024deepseek | "DeepSeek-Coder-V2: MoE (236B total, 21B active)", "90.2% HumanEval" | MoE 파라미터 수치 원문 대조, 90.2% 벤치마크 원문 대조 | **4단계 검증 완료 — 모든 수치 일치** |

> *[2]는 사용자 제공 레퍼런스이지만, 논문 내 인용된 구체적 수치(88.4%, 84.1%)는 Claude가 추가했으므로 검증 필요

### B-5. 이더리움 스마트 컨트랙트 연구 (딥리서치 5단계)

| # | BibTeX Key | 논문 내 인용 | 검증 포인트 | 검증 결과 |
|:--|:-----------|:------------|:-----------|:---------|
| [6] | chen2023chatgpt | "ChatGPT vs 기존 스마트 컨트랙트 탐지 도구 비교" | Ethereum/Solidity 중심인지 확인 (HLF 포함 여부), 실험 내용 일치 여부 | **5단계 검증 완료 — Ethereum/Solidity 전용 확인, HLF 미포함** |
| [7] | david2023manual | "GPT-4/Claude vs 수동 감사", "40% 취약점 유형 식별 정확도" | 40% 수치 원문 대조, GPT-4와 Claude 모두 사용 확인 | **5단계 검증 완료 — 모든 주장 일치** |

### B-6. 서지 정보 무결성 (딥리서치 6단계)

| # | BibTeX Key | 확인 항목 | 검증 포인트 | 검증 결과 |
|:--|:-----------|:---------|:-----------|:---------|
| [10] | feist2019slither | Slither: A Static Analysis Framework For Smart Contracts | 제목/저자/arXiv:1908.09878 일치 확인, WETSEB 2019 학회 확인 | **6단계 검증 완료 — DOI:10.1109/WETSEB.2019.00008, 저자 Feist/Grieco/Groce 일치** |
| [11] | zhang2023prompt | Prompt-Enhanced Software Vulnerability Detection Using ChatGPT | 제목/저자/arXiv:2308.12697 일치 확인 | **6단계 검증 완료 — 제목/저자(Chenyuan Zhang)/arXiv 일치** |
| [12] | hao2023ev | E&V: Prompting LLMs to Perform Static Analysis... | 제목/저자/arXiv:2312.08477 일치 확인, 81.2% 정확도 원문 대조 | **6단계 검증 완료 — 제목/저자(Yu Hao)/arXiv 일치. 81.2% 수치는 DR에서 미언급 (추가 확인 권장)** |
| [13] | zhang2023surveyse | A Survey on Large Language Models for Software Engineering | 제목/저자/arXiv:2312.15223 일치 확인 | **6단계 검증 완료 — 제목/저자(Quanjun Zhang, Nanjing Univ)/arXiv 일치** |

---

## C. 논문 내 인용되었으나 v5 참고문헌에서 제거된 레퍼런스

| 이전 Key | 제목 | 이전 등급 | 비고 |
|:---------|:-----|:---------|:-----|
| bhatt2023cyberseceval | Purple Llama CyberSecEval | A등급 (R11) | v5에서 제거됨 — 논문 본문에 인용 없음 |
| zhang2023survey | A Survey on LLM for SE | A등급 (R14) | v5에서 [13]으로 유지됨 (zhang2023surveyse) |

---

## D. 딥리서치 → 검증 체크리스트

딥리서치 결과 수신 시 아래 항목을 순서대로 체크:

### 체크리스트

- [x] **1단계 결과 수신**: GoLiSA [15], VulFinder [16] — 2026-02-09 수신 완료
  - [x] [15] DOI 실존 확인 — DOI:10.4230/LIPIcs.ECOOP.2023.23, ECOOP 2023 (Seattle, July 17-21, 2023)
  - [x] [15] abstract interpretation 방법론 일치 — Information Flow Analysis (taint analysis + non-interference), LiSA 프레임워크 기반, CFG fixpoint iteration
  - [x] [15] GitHub 코퍼스 데이터셋 설명 일치 — "600+ real-world blockchain programs from GitHub" 확인
  - [x] [16] IEEE TSE 2025 게재 확인 — Vol.51, Issue 12, December 2025 (Online Early Access Sep 2025)
  - [x] [16] DOI 실존 확인 — DOI:10.1109/TSE.2025.3605379
  - [x] [16] Knowledge Graph 기술 확인 — 온톨로지 + SPARQL 쿼리 기반 결함 패턴 매칭
  - [x] [16] 22개 카테고리 수치 확인 — "22 kinds of typical vulnerabilities", recall 98.87%
  - **[16] 저자명 불일치 발견**: 논문에 "Y.~Li" 기재 → 실제 주저자는 **Bixin Li** (공저자: Tianyuan Hu, Xiangfei Xu, Lulu Wang). **BibTeX 수정 필요**: `Y.~Li` → `B.~Li`

- [x] **2단계 결과 수신**: 통계 수치 [4], [5], [8], [9] — 2026-02-09 수신 완료
  - [x] [4] 16 LLMs / 5,000 샘플 / 62.8% 정확도 일치 — F1 0.71, 5개 데이터셋 각 1,000개, 25 CWE
  - [x] [4] 실세계 C/C++ 성능 저하 확인 (global context 부재), Java synthetic에서는 >60%
  - [x] [5] SecLLMHolmes 프레임워크 이름 일치 — 자동화된 adversarial robustness 평가 프레임워크
  - [x] [5] 변수명 변경 → GPT-4 17% 오답률 증가 확인 (PaLM2는 26%)
  - [x] [5] "unfaithful reasoning" 발견 — 올바른 탐지도 잘못된 설명으로 정당화
  - [x] [8] dummy classifier 비교 결론 일치 — AUC 0.51 vs dummy 0.50 (통계적 동등)
  - [x] [8] 높은 Recall(0.99) + 낮은 Precision = 거의 모든 것을 vulnerable로 분류
  - [x] [9] StarCoder2 7B / BigVul 68.26% / PrimeVul 3.09% 일치
  - [x] [9] 데이터 오염 원인: commit tangling + 18.9% train-test 중복 + noisy labeling
  - [x] [9] PrimeVul: chronological split + VD-S 메트릭 도입

- [x] **3단계 결과 수신**: 2025 논문 [17], [18] — 2026-02-09 수신 완료
  - [x] [17] ICSE 2025 게재 확정 — 47th ICSE, Research Track, Ottawa, April 27–May 3, 2025
  - [x] [17] GPT-4 precision 0.99 / 3 FP 확인 — recall 0.95, F1 0.97, 2,089 TP, MalwareBench 5,115 packages
  - [x] [17] CodeQL 대비: 16% precision 향상, 9% F1 향상 (논문 내 "684 FP" 표현은 CodeQL의 FP 수 — 확인됨)
  - [x] [18] ICLR 2025 게재 확정 — Conference Paper (Poster), Published Jan 22, 2025
  - [x] [18] 27→55 취약점 탐지 확인 — CWE-Bench-Java 120 CVE, +103.7% recall, 4개 zero-day 추가 발견
  - [x] [18] neuro-symbolic 방법론 확인 — GPT-4가 CodeQL용 source-sink 스펙 추론, CodeQL 재분석 루프
  - **[17] 논문 내 표기 확인**: "Proc. ICSE 2025" → 정확. 저자: Nusrat Zahan (NCSU), arXiv:2403.12196
  - **[18] 논문 내 표기 확인**: "Proc. ICLR 2025" → 정확. 저자: Ziyang Li (UPenn), arXiv:2405.17238

- [x] **4단계 결과 수신**: 모델 벤치마크 [2], [19] — 2026-02-09 수신 완료
  - [x] [2] Qwen2.5-Coder-7B HumanEval 88.4% 일치 — arXiv:2409.12186 원문, 5.5T 토큰 코퍼스, 128K ctx
  - [x] [2] Llama-3-70B-Instruct 84.1% 비교 일치 — Llama-3.1-70B는 ~80.5-81.7% (평가 harness 차이)
  - [x] [19] DeepSeek-Coder-V2 MoE 236B/21B 일치 — DeepSeekMoE, fine-grained experts + shared experts
  - [x] [19] 90.2% HumanEval 일치 — MBPP+ 76.2%, MATH 75.7%, 338개 언어 지원, 6T 추가 토큰
  - **참고**: Qwen 7B의 수학 벤치마크(46.6%)는 Llama-3-70B(~77%)보다 낮음 — specialist vs generalist 트레이드오프 확인

- [x] **5단계 결과 수신**: 이더리움 연구 [6], [7] — 2026-02-09 수신 완료
  - [x] [6] Ethereum/Solidity 전용 확인 — SmartBugs 142 contracts, HLF 미포함
  - [x] [6] GPT-4 recall 88.2% / precision 22.6% (논문 내 "empirically evaluated ChatGPT against existing tools" 서술과 일치)
  - [x] [7] GPT-4 + Claude (v1.3) 둘 다 사용 확인 — 52개 DeFi 해킹 컨트랙트, ~$1B 손실
  - [x] [7] 40% 정확도 확인 — "True Positive Identification Rate of Vulnerability Types" (유형 식별 정확도, 일반 정확도 아님)
  - [x] [7] mutation testing에서 GPT-4-32k: 78.7% TPR (통제된 환경에서는 높음)
  - **참고**: 두 논문 모두 Ethereum/Solidity 전용이며 HLF 미적용. 우리 논문의 "These studies focus on Ethereum/Solidity" 서술 정확

- [x] **6단계 결과 수신**: 서지 정보 [1], [3], [10], [11], [12], [13] — 2026-02-09 수신 완료
  - [x] [1] 제목/저자/DOI 일치 — DOI:10.1145/3190508.3190538, EuroSys '18, 저자 Elli Androulaki 외 20인
  - [x] [3] 제목/arXiv 일치 — arXiv:2407.21783, 500+ 저자. **참고**: "Grattafiori et al." 유효하나 "Dubey et al."도 통용
  - [x] [10] 제목/저자/DOI/학회 일치 — DOI:10.1109/WETSEB.2019.00008, WETSEB 2019 (ICSE co-located)
  - [x] [11] 제목/저자/arXiv 일치 — Chenyuan Zhang (Xiamen Univ + Alibaba), arXiv:2308.12697
  - [x] [12] 제목/저자/arXiv 일치 — Yu Hao, arXiv:2312.08477. **주의**: 81.2% 수치는 DR 리포트에서 구체적 확인 미완 (추가 직접 확인 권장)
  - [x] [13] 제목/저자/arXiv 일치 — Quanjun Zhang (Nanjing Univ), arXiv:2312.15223, 900+ 연구 서베이

---

## E. 검증 후 조치 방침

| 검증 결과 | 조치 |
|:---------|:-----|
| **사실 일치** | 레퍼런스 유지, 체크리스트에 체크 |
| **수치 불일치** | 원문 수치로 즉시 수정 (논문 본문 + BibTeX) |
| **논문 실존하나 내용 불일치** | 인용 문구 수정 또는 삭제 |
| **논문 실존 불확인 (할루시네이션 의심)** | 즉시 삭제, 대체 레퍼런스 탐색 |
| **학회 게재 미확정 (arXiv only)** | "arXiv preprint"으로 표기 수정, Proc. X 삭제 |

---

## F. 논문 v6 참고문헌 전체 매핑 (번호 ↔ BibTeX Key)

| v6 번호 | v5 번호 | BibTeX Key | 분류 | DR 단계 | v6 재검증 |
|:--------|:--------|:-----------|:-----|:--------|:---------|
| [1] | [1] | androulaki2018fabric | USER | 6단계 | DR3, DR6 확인 |
| [2] | [2] | hui2024qwen25coder | USER (수치는 CLAUDE) | 4단계 | DR1, DR3, DR4 확인 |
| [3] | [3] | grattafiori2024llama3 | USER | 6단계 | DR3 확인 |
| [4] | [4] | khare2023understanding | CLAUDE | 2단계 | DR5 확인 |
| [5] | [5] | ullah2023llms | CLAUDE | 2단계 | DR1 확인 |
| [6] | [7] | david2023manual | CLAUDE | 5단계 | DR2 확인 |
| [7] | [8] | cheshkov2023evaluation | CLAUDE | 2단계 | DR3, DR5 확인 |
| [8] | [9] | ding2024vulnerability | CLAUDE | 2단계 | DR5 확인 |
| [9] | [10] | feist2019slither | CLAUDE | 6단계 | DR2 확인 |
| [10] | [12] | hao2023ev | CLAUDE | 6단계 | DR5 확인 |
| [11] | [14] | semgrep | USER | — | 도구 직접 사용 |
| [12] | [15] | olivieri2023golisa | CLAUDE | 1단계 | DR3, DR6 확인 |
| [13] | [16] | li2025vulfinder | CLAUDE | 1단계 | DR3, DR6 + **IEEE Xplore 직접 확인** |
| [14] | [17] | zahan2024llm | CLAUDE | 3단계 | DR4 확인 |
| [15] | [18] | li2025iris | CLAUDE | 3단계 | DR4 확인 |

### v6에서 삭제된 참고문헌 (3개)

| v5 번호 | BibTeX Key | 삭제 사유 |
|:--------|:-----------|:---------|
| [6] | chen2023chatgpt | 구체적 수치 없이 한 줄 언급. 리뷰어 관점 불필요 |
| [11] | zhang2023prompt | hao2023ev와 병렬 인용, 독립적 기여 없음 |
| [19] | zhu2024deepseek | 실험하지 않은 모델. 리스트 채우기 인상 |

### v5에서 이미 삭제된 참고문헌 (2개)

| 이전 Key | 삭제 시점 | 사유 |
|:---------|:---------|:-----|
| zhang2023surveyse | v5 | 본문 미인용 죽은 참고문헌 |
| bhatt2023cyberseceval | v5 이전 | 본문 미인용 |
