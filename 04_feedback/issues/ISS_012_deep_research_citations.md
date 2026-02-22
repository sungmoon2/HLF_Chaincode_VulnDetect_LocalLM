# ISS_012: Deep Research 참고문헌 확보 (4개 카테고리)

**Status**: open (딥리서치 결과 수신, 검증 완료, 논문 반영 대기)
**Priority**: critical
**Source Feedback**: FB_260209_1700_gemini_strategy
**Target Sections**: I, II, III
**Created**: 2026-02-09
**Last Updated**: 2026-02-09T17:45:00

## 문제 설명
논문의 Introduction과 Related Work를 학술적으로 탄탄하게 구성하기 위해 4개 카테고리의 핵심 참고문헌을 확보해야 한다. 각 카테고리당 1개의 대표 논문(Representative Paper)을 무료 접근 가능한(Open Access/arXiv) 형태로 확보하는 것이 목표이다.

## 딥리서치 결과 및 독립 검증 (2026-02-09 완료)

### 카테고리 1: HLF 아키텍처/XOV — VERIFIED
- **논문**: "Hyperledger Fabric: A Distributed Operating System for Permissioned Blockchains"
- **저자**: Elli Androulaki, Artem Barger, Vita Bortnikov, Christian Cachin, et al.
- **학회**: EuroSys 2018 (13th European Conference on Computer Systems)
- **arXiv PDF**: https://arxiv.org/pdf/1801.10228 — **HTTP 200 OK 확인 (2026-02-09)**
- **무료 접근**: Open Access (arXiv)
- **핵심 인용구** (딥리서치 보고서에서 추출, 논문 Section 3.3):
  > "For each transaction, it compares the versions of the keys in the readset with those currently in the ledger. If the versions do not match, the transaction is marked as invalid and its effects are ignored."
- **논문 활용**: Introduction에서 XOV 아키텍처와 비결정론 위험성의 이론적 근거로 인용
- **주의**: 인용구는 딥리서치 보고서에서 추출한 것이며, 원문 PDF에서 직접 재확인 필요

### 카테고리 2: 정적 분석/기존 벤치마크 — VERIFIED (접근성 제한 확인)

#### GoLiSA (ECOOP 2023) — Open Access 확인
- **논문**: "Information Flow Analysis for Detecting Non-Determinism in Blockchain"
- **저자**: Luca Olivieri (U. Verona), Luca Negrini (Corvallis Srl), Vincenzo Arceri (U. Parma), Fabio Tagliaferro (CYS4 Srl), Pietro Ferrara (Ca' Foscari U.), Agostino Cortesi (Ca' Foscari U.), Fausto Spoto (U. Verona)
- **학회**: ECOOP 2023 (37th European Conference on Object-Oriented Programming)
- **Dagstuhl PDF**: https://drops.dagstuhl.de/storage/00lipics/lipics-vol263-ecoop2023/LIPIcs.ECOOP.2023.23/LIPIcs.ECOOP.2023.23.pdf — **Open Access 확인**
- **대학 repo PDF**: https://vincenzoarceri.github.io/papers/ecoop2023.pdf
- **방법론**: Abstract Interpretation 기반 Information Flow Analysis (Taint + Non-interference)
- **GitHub**: https://github.com/lisa-analyzer/go-lisa — **HTTP 200 OK 확인, 6 stars, Java**
- **DARTS Artifact**: https://drops.dagstuhl.de/entities/document/10.4230/DARTS.9.2.23 — **4.98 GB OVA (VM 이미지)**

##### 데이터셋 크기 "651개" 검증 상태: **미검증**
- 딥리서치 보고서는 "651 real-world Hyperledger Fabric smart contracts"라고 주장
- **Zenodo DOI 10.5281/zenodo.7896323**: 302 → record 7896324로 리다이렉트되며, 이는 **전혀 다른 우즈벡어 논문**임. **딥리서치의 Zenodo DOI는 부정확(오류)**
- GoLiSA GitHub repo (`lisa-analyzer/go-lisa`): 도구 코드 + 소규모 테스트 케이스(non-det: channel, goroutines, map-iter 3개 디렉토리)만 포함. **651개 체인코드 미포함**
- DARTS Artifact: 4.98 GB OVA VM 이미지 내에 데이터셋이 포함되어 있을 가능성 있음
- **결론**: "651개"라는 숫자와 데이터셋 접근 방법은 원문 PDF에서 직접 확인 필요. 간단히 다운로드 가능한 .go 파일 세트 형태로는 존재하지 않음

#### VulFinder (IEEE TSE 2025) — Paywalled
- **논문**: "VulFinder: Exploring Chaincode Vulnerabilities More Effectively and Efficiently Using Knowledge Graph Based Defect Pattern Matching"
- **DOI**: 10.1109/TSE.2025.3605379
- **접근성**: IEEE Xplore (유료). ResearchGate에서 full-text 요청 가능
- **방법론**: Knowledge Graph + SPARQL 패턴 매칭, 22종 취약점 분류
- **성능**: Recall 98.87% (딥리서치 보고서 기준)
- **무료 PDF 미확보** — 논문에 인용 시 DOI만 사용, 상세 인용은 제한적

### 카테고리 3: LLM vs 정적 분석 — VERIFIED

#### 주 논문: Zahan et al. (ICSE 2025)
- **논문**: "Leveraging Large Language Models to Detect npm Malicious Packages"
- **arXiv**: https://arxiv.org/abs/2403.12196 — **HTTP 200 OK 확인**
- **학회**: ICSE 2025
- **핵심 수치** (딥리서치 보고서에서 추출):
  - CodeQL: Precision 0.75, Recall 0.97, FP 684건
  - GPT-4: Precision 0.99, Recall 0.95, FP 3건
  - GPT-3: Precision 0.91, Recall 0.97, FP 195건
- **주의**: npm 생태계 malware 탐지에 초점. HLF 체인코드와는 도메인이 다름. 논문에서 인용 시 "LLM이 정적 분석 대비 FP를 크게 줄인 사례"로 일반화하여 참조

#### 보조 논문: IRIS (ICLR 2025)
- **논문**: "IRIS: LLM-Assisted Static Analysis for Detecting Security Vulnerabilities"
- **arXiv**: https://arxiv.org/abs/2405.17238 — **HTTP 200 OK 확인**
- **핵심**: Neuro-symbolic 접근. LLM이 CodeQL의 Source/Sink 스펙을 추론 → CodeQL 탐지율 27 → 55건으로 2배 이상 향상

### 카테고리 4: 코드 특화 sLM 성능 — VERIFIED

#### 주 논문: Qwen2.5-Coder Technical Report
- **논문**: "Qwen2.5-Coder Technical Report"
- **저자**: Binyuan Hui et al. (Alibaba Cloud)
- **arXiv**: https://arxiv.org/abs/2409.12186 — **HTTP 200 OK 확인**
- **핵심 수치** (딥리서치 보고서에서 추출):
  - Qwen2.5-Coder-7B-Instruct: HumanEval 88.4% Pass@1
  - Llama-3-70B-Instruct: ~84.1%
  - GPT-4o: ~90.2%
- **의의**: 7B 파라미터 코드 특화 모델이 70B 범용 모델을 HumanEval에서 능가
- **학습 데이터**: 5.5조 토큰, 실행 검증(execution-verified) 합성 데이터 포함

#### 보조 논문: DeepSeek-Coder-V2
- **arXiv**: https://arxiv.org/abs/2406.11931 — **HTTP 200 OK 확인**
- **핵심**: 236B 총 / 21B 활성 (MoE), HumanEval 90.2%

#### 보조 논문: StarCoder2
- **arXiv**: https://arxiv.org/abs/2402.19173 — **HTTP 200 OK 확인**
- **핵심**: 15B, ~67.7% HumanEval, 윤리적 데이터 수집 (The Stack v2)

## 검증 요약

| 카테고리 | 논문 | arXiv/PDF | 접근성 | 상태 |
|:---------|:-----|:----------|:------|:-----|
| 1. HLF XOV | Androulaki et al. (EuroSys 2018) | arXiv:1801.10228 | Open Access | VERIFIED |
| 2a. GoLiSA | Olivieri et al. (ECOOP 2023) | Dagstuhl LIPIcs | Open Access (CC-BY) | VERIFIED |
| 2b. VulFinder | Li et al. (IEEE TSE 2025) | IEEE Xplore | Paywalled | DOI만 확보 |
| 3. LLM vs SAST | Zahan et al. (ICSE 2025) | arXiv:2403.12196 | Open Access | VERIFIED |
| 3b. IRIS | Li et al. (ICLR 2025) | arXiv:2405.17238 | Open Access | VERIFIED |
| 4. Qwen2.5-Coder | Hui et al. (2024) | arXiv:2409.12186 | Open Access | VERIFIED |
| 4b. DeepSeek-V2 | Zhu et al. (2024) | arXiv:2406.11931 | Open Access | VERIFIED |
| 4c. StarCoder2 | Lozhkov et al. (2024) | arXiv:2402.19173 | Open Access | VERIFIED |

## 팩트 체크 경고

### Zenodo DOI 오류
- 딥리서치 보고서에서 제공한 "DOI: 10.5281/zenodo.7896323"은 **부정확**
- 해당 DOI는 관련 없는 우즈벡어 논문으로 리다이렉트됨
- GoLiSA artifact는 Zenodo가 아닌 **DARTS (Dagstuhl)**에서 OVA VM 이미지로 배포

### "651개 체인코드" 검증 미완
- 딥리서치 보고서는 "651 real-world chaincodes"라고 주장
- 원문 PDF에서 이 숫자를 직접 확인하지 못함 (PDF 바이너리 파싱 불가)
- 논문 인용 시 "~651"이 아닌, 원문에서 확인된 정확한 숫자를 사용해야 함

## 관련 이슈
- ISS_010: GoLiSA 인용 (본 카테고리 2 결과에 의존)
- ISS_013: GoLiSA 외부 검증 실험 (데이터셋 접근성 제한 확인 — OVA VM 이미지)
- ISS_014: Core Thesis 재정립 (카테고리 1, 4 결과 활용)
