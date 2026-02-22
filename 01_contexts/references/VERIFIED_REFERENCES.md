# 검증된 참고문헌 레지스트리 (Verified References Registry)

> **최종 갱신**: 2026-02-09
> **검증 방법**: (1) arXiv 페이지 직접 조회 (2) Gemini Deep Research 6단계 독립 검증
> **원칙**: 이 파일에 등록된 문헌만 논문에 인용 가능
> **현재 논문**: 260209_2140_전체논문_v6_불필요참고문헌삭제.tex (15개 참고문헌)

---

## 신뢰도 등급

| 등급 | 설명 | 인용 가능 여부 |
|:-----|:-----|:--------------|
| **A+ (이중검증)** | arXiv 직접 확인 + Gemini DR 독립 검증 완료 | 인용 가능 |
| **A (확인됨)** | arXiv/DOI로 원문 접근 확인, 제목/저자/연도 일치 | 인용 가능 |

---

## 논문 v6 참고문헌 전체 (15개)

### [1] androulaki2018fabric — A+ (이중검증)
- **제목**: Hyperledger Fabric: A Distributed Operating System for Permissioned Blockchains
- **저자**: Elli Androulaki et al. (20인)
- **학회**: EuroSys 2018
- **DOI**: 10.1145/3190508.3190538
- **arXiv**: 1801.10228
- **검증**: arXiv 직접 확인 + Gemini DR 6단계 (제목/저자/DOI 일치)
- **논문 내 용도**: HLF XOV 아키텍처 설명, 합의 레이어 취약점의 기술적 기반

### [2] hui2024qwen25coder — A+ (이중검증)
- **제목**: Qwen2.5-Coder Technical Report
- **저자**: Binyuan Hui, Jian Yang, Zeyu Cui et al.
- **arXiv**: 2409.12186
- **연도**: 2024
- **검증**: arXiv 직접 확인 + Gemini DR 4단계 (88.4% HumanEval, 84.1% Llama-3-70B 비교 수치 일치)
- **핵심 수치**: HumanEval Pass@1 88.4%, 5.5T 토큰 코퍼스, 128K ctx
- **논문 내 용도**: Qwen2.5-Coder 모델 성능 및 아키텍처 설명

### [3] grattafiori2024llama3 — A+ (이중검증)
- **제목**: The Llama 3 Herd of Models
- **저자**: Aaron Grattafiori et al. (500+ co-authors)
- **arXiv**: 2407.21783
- **연도**: 2024
- **검증**: arXiv 직접 확인 + Gemini DR 6단계 (제목/arXiv 일치. "Dubey et al."도 통용)
- **논문 내 용도**: Llama-3.1-8B 모델 아키텍처 설명

### [4] khare2023understanding — A+ (이중검증)
- **제목**: Understanding the Effectiveness of Large Language Models in Detecting Security Vulnerabilities
- **저자**: Avishree Khare, Saikat Dutta, Ziyang Li, Alaia Solko-Breslin, Rajeev Alur, Mayur Naik
- **arXiv**: 2311.16169
- **연도**: 2023
- **검증**: arXiv 직접 확인 + Gemini DR 2단계
- **핵심 수치**: 16 LLMs, 5,000 코드 샘플 (5개 데이터셋 × 1,000), 평균 정확도 62.8%, F1 0.71
- **DR 추가 확인**: 25 CWE, 실세계 C/C++에서 성능 저하 (global context 부재)
- **논문 내 용도**: LLM 기반 코드 감사의 현재 한계 인용

### [5] ullah2023llms — A+ (이중검증)
- **제목**: LLMs Cannot Reliably Identify and Reason About Security Vulnerabilities (Yet?)
- **저자**: Saad Ullah, Mingji Han, Saurabh Pujar, Hammond Pearce, Ayse Coskun, Gianluca Stringhini
- **arXiv**: 2312.12575
- **연도**: 2023
- **검증**: arXiv 직접 확인 + Gemini DR 2단계
- **핵심 수치**: SecLLMHolmes 프레임워크, 변수명 변경 시 GPT-4 17% 오답률 증가, PaLM2 26%
- **DR 추가 확인**: "unfaithful reasoning" — 올바른 탐지도 잘못된 설명으로 정당화
- **논문 내 용도**: 난독화 실험의 선행 연구 근거, LLM의 명명 단서 의존성

### [6] david2023manual (v5까지 [7]) — A+ (이중검증)
- **제목**: Do you still need a manual smart contract audit?
- **저자**: Isaac David, Liyi Zhou, Kaihua Qin, Dawn Song, Lorenzo Cavallaro, Arthur Gervais
- **arXiv**: 2306.12338
- **연도**: 2023
- **검증**: arXiv 직접 확인 + Gemini DR 5단계
- **핵심 수치**: GPT-4 + Claude v1.3 사용, 취약점 유형 식별 정확도 40% (52개 DeFi 해킹 컨트랙트)
- **DR 추가 확인**: mutation testing에서 GPT-4-32k TPR 78.7%, Ethereum/Solidity 전용
- **논문 내 용도**: 클라우드 LLM의 스마트 컨트랙트 감사 성능 선행 연구

### [7] cheshkov2023evaluation (v5까지 [8]) — A+ (이중검증)
- **제목**: Evaluation of ChatGPT Model for Vulnerability Detection
- **저자**: Anton Cheshkov, Pavel Zadorozhny, Rodion Levichev
- **arXiv**: 2304.07232
- **연도**: 2023
- **검증**: arXiv 직접 확인 + Gemini DR 2단계
- **핵심 수치**: AUC 0.51 vs dummy 0.50 (통계적 동등), Recall 0.99 + 낮은 Precision
- **DR 추가 확인**: 거의 모든 것을 vulnerable로 분류하는 편향. "Fixed Code" anomaly 발견
- **논문 내 용도**: LLM 기반 취약점 탐지의 한계 인용

### [8] ding2024vulnerability (v5까지 [9]) — A+ (이중검증)
- **제목**: Vulnerability Detection with Code Language Models: How Far Are We?
- **저자**: Yangruibo Ding, Yanjun Fu, Omniyyah Ibrahim et al.
- **arXiv**: 2403.18624
- **연도**: 2024
- **검증**: arXiv 직접 확인 + Gemini DR 2단계
- **핵심 수치**: StarCoder2 7B — BigVul F1 68.26% → PrimeVul F1 3.09%
- **DR 추가 확인**: commit tangling, 18.9% train-test 중복, PrimeVul chronological split + VD-S 메트릭
- **논문 내 용도**: 데이터셋 품질에 따른 성능 과대평가 문제 인용

### [9] feist2019slither (v5까지 [10]) — A+ (이중검증)
- **제목**: Slither: A Static Analysis Framework For Smart Contracts
- **저자**: Josselin Feist, Gustavo Grieco, Alex Groce
- **학회**: WETSEB 2019 (ICSE co-located)
- **DOI**: 10.1109/WETSEB.2019.00008
- **arXiv**: 1908.09878
- **검증**: arXiv 직접 확인 + Gemini DR 6단계 (DOI/학회/저자 일치)
- **논문 내 용도**: 전통 정적 분석 도구의 대표 사례, Solidity 중심 한계 설명

### [10] hao2023ev (v5까지 [12]) — A+ (이중검증)
- **제목**: E&V: Prompting Large Language Models to Perform Static Analysis by Pseudo-code Execution and Verification
- **저자**: Yu Hao, Weiteng Chen, Ziqiao Zhou, Weidong Cui
- **arXiv**: 2312.08477
- **연도**: 2023
- **검증**: arXiv 직접 확인 + Gemini DR 6단계 + arXiv abstract 직접 확인
- **핵심 수치**: 170개 Linux 커널 버그, blamed function 식별 정확도 81.2% (베이스라인 28.2% 대비)
- **논문 내 용도**: LLM을 정적 분석에 활용하는 접근법의 선행 연구

### [11] semgrep (v5까지 [13]) — A (확인됨)
- **제목**: Semgrep: Lightweight static analysis for many languages
- **출처**: https://semgrep.dev
- **검증**: 공식 웹사이트 확인 + 도구 직접 설치/실행 (v1.151.0)
- **논문 내 용도**: 전통 도구 베이스라인

### [12] olivieri2023golisa (v5까지 [14]) — A+ (이중검증)
- **제목**: Information Flow Analysis for Detecting Non-Determinism in Blockchain
- **저자**: Luca Olivieri, Luca Negrini, Vincenzo Arceri, Fabio Tagliaferro, Pietro Ferrara, Agostino Cortesi, Fausto Spoto
- **학회**: ECOOP 2023 (37th, Seattle, July 17-21, 2023)
- **DOI**: 10.4230/LIPIcs.ECOOP.2023.23
- **검증**: Gemini DR 1단계 (DOI 실존, abstract interpretation 일치, GitHub 600+ 코퍼스 일치)
- **DR 확인**: LiSA 프레임워크, CFG fixpoint iteration, taint analysis + non-interference checking
- **논문 내 용도**: HLF 전용 정적 분석 도구 (SOTA 비교), GoLiSA 벤치마크 데이터셋 출처

### [13] li2025vulfinder (v5까지 [15]) — A+ (이중검증)
- **제목**: VulFinder: Exploring Chaincode Vulnerabilities More Effectively and Efficiently Using Knowledge Graph Based Defect Pattern Matching
- **저자**: **Bixin Li**, Tianyuan Hu, Xiangfei Xu, Lulu Wang
- **학회**: IEEE Trans. Softw. Eng., Vol. 51, No. 12, December 2025
- **DOI**: 10.1109/TSE.2025.3605379
- **검증**: Gemini DR 1단계 (DOI 실존, IEEE TSE 게재 확인, KG 기술 확인, 22개 카테고리 확인)
- **DR 확인**: 온톨로지 + SPARQL 쿼리, recall 98.87%, Online Early Access Sep 2025
- **수정 이력**: 저자명 `Y.~Li` → `B.~Li` 수정 (2026-02-09, Gemini DR에서 발견)
- **논문 내 용도**: HLF 체인코드 보안 분석의 SOTA 비교군

### [14] zahan2024llm (v5까지 [16]) — A+ (이중검증)
- **제목**: Leveraging Large Language Models to Detect npm Malicious Packages
- **저자**: Nusrat Zahan (NCSU), Philipp Burckhardt, Mikola Lysenko, Feross Aboukhadijeh, Laurie Williams
- **학회**: ICSE 2025 (47th, Research Track, Ottawa, April 27-May 3, 2025)
- **arXiv**: 2403.12196
- **검증**: Gemini DR 3단계 (ICSE 2025 게재 확정, 수치 일치)
- **핵심 수치**: GPT-4 precision 0.99, 3 FP, recall 0.95, F1 0.97, MalwareBench 5,115 packages
- **DR 추가 확인**: CodeQL 대비 16% precision 향상, 9% F1 향상, 하이브리드 파이프라인 76.1% 비용 절감
- **논문 내 용도**: LLM + 전통 도구 통합의 SOTA 사례

### [15] li2025iris (v5까지 [17]) — A+ (이중검증)
- **제목**: IRIS: LLM-Assisted Static Analysis for Detecting Security Vulnerabilities
- **저자**: Ziyang Li (UPenn), Saikat Dutta (Cornell), Mayur Naik (UPenn)
- **학회**: ICLR 2025 (Conference Paper, Poster, Published Jan 22, 2025)
- **arXiv**: 2405.17238
- **검증**: Gemini DR 3단계 (ICLR 2025 게재 확정, 수치 일치)
- **핵심 수치**: CodeQL 27 → IRIS 55 취약점 탐지 (+103.7%), CWE-Bench-Java 120 CVE, 4개 zero-day
- **DR 확인**: neuro-symbolic (GPT-4 → CodeQL source-sink 스펙 추론 → 재분석 루프)
- **논문 내 용도**: LLM + CodeQL 통합의 neuro-symbolic 접근법

---

## 삭제된 참고문헌

### chen2023chatgpt (v5까지 [6], v6에서 삭제)
- **제목**: When ChatGPT Meets Smart Contract Vulnerability Detection: How Far Are We?
- **저자**: Chong Chen, Jianzhong Su, Jiachi Chen et al.
- **arXiv**: 2309.05520
- **이전 등급**: A+ (arXiv 직접 확인 + Gemini DR 5단계)
- **삭제 사유**: 본문에서 구체적 수치 인용 없이 한 줄 언급에 그침. 리뷰어 관점에서 불필요
- **삭제일**: 2026-02-09 (v6)

### zhang2023prompt (v5까지 [11], v6에서 삭제)
- **제목**: Prompt-Enhanced Software Vulnerability Detection Using ChatGPT
- **저자**: Chenyuan Zhang, Hao Liu, Jiutian Zeng, Kejing Yang, Yuhong Li, Hui Li
- **arXiv**: 2308.12697
- **이전 등급**: A+ (arXiv 직접 확인 + Gemini DR 6단계)
- **삭제 사유**: hao2023ev와 병렬 인용된 한 줄에서 구체적 기여 없음. hao2023ev 단독으로 문장 성립
- **삭제일**: 2026-02-09 (v6)

### zhu2024deepseek (v5까지 [18], v6에서 삭제)
- **제목**: DeepSeek-Coder-V2: Breaking the Barrier of Closed-Source Models in Code Intelligence
- **저자**: Qihao Zhu et al.
- **arXiv**: 2406.11931
- **이전 등급**: A+ (Gemini DR 4단계)
- **삭제 사유**: 논문에서 실험하지 않은 모델. hui2024qwen25coder만으로 코드 전문화 트렌드 논거 충분
- **삭제일**: 2026-02-09 (v6)

### zhang2023surveyse (이전 [13], v5에서 삭제)
- **제목**: A Survey on Large Language Models for Software Engineering
- **저자**: Quanjun Zhang, Chunrong Fang et al.
- **arXiv**: 2312.15223
- **이전 등급**: A (arXiv 직접 확인)
- **삭제 사유**: 논문 v5 본문에서 `\cite{zhang2023surveyse}` 미사용 (죽은 참고문헌)
- **삭제일**: 2026-02-09
- **Gemini DR 6단계**: 제목/저자/arXiv 자체는 실존 확인됨 (논문 자체 문제 아닌 미인용 문제)

### bhatt2023cyberseceval (v5 미포함)
- **제목**: Purple Llama CyberSecEval: A Secure Coding Benchmark for Language Models
- **저자**: Manish Bhatt, Sahana Chennabasappa et al.
- **arXiv**: 2312.04724
- **이전 등급**: A (arXiv 직접 확인)
- **비고**: v5에서 제거됨 — 논문 본문에서 인용하지 않음

---

## 검증 이력

| 일자 | 작업 | 결과 |
|:-----|:-----|:-----|
| 2026-02-09 (초기) | arXiv 페이지 직접 조회 (13개) | A등급 13개 확정 |
| 2026-02-09 (DR 1단계) | GoLiSA, VulFinder 검증 | B→A+ 승격, VulFinder 저자명 오류 발견 |
| 2026-02-09 (DR 2단계) | 통계 수치 [4][5][8][9] | 모든 수치 일치 확인 |
| 2026-02-09 (DR 3단계) | 2025 논문 [16][17] | ICSE/ICLR 게재 확정, 수치 일치 |
| 2026-02-09 (DR 4단계) | 모델 벤치마크 [2][18] | 88.4%/90.2%/236B/21B 수치 일치 |
| 2026-02-09 (DR 5단계) | 이더리움 연구 [6][7] | Ethereum 전용 확인, 40% 수치 일치 |
| 2026-02-09 (DR 6단계) | 서지 정보 [1][3][10][11][12] | 전체 일치, [12] 81.2% arXiv 직접 확인 |
| 2026-02-09 (수정) | 논문 v5 수정 | [16] 저자명 B.~Li, [13] 삭제, 18개로 축소 |
| 2026-02-09 (v6) | Gemini DR v6 재검증 + 비판적 리뷰 | [6] chen2023chatgpt, [11] zhang2023prompt, [18] zhu2024deepseek 삭제 → 15개 |
| 2026-02-09 (v6) | IEEE Xplore 직접 확인 | li2025vulfinder vol.51 no.12 Dec 2025 pp.3247-3266 확정 |
| 2026-02-09 (v6) | 15개 전체 최종 확인 | 할루시네이션 0건 — Gemini DR 6건 + IEEE Xplore 1건 교차 검증 완료 |
