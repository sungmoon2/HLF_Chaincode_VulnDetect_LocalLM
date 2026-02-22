# [원문 전문] Deep Research 1: 관련 연구 조사
> 원본 제목: "프라이버시 보존형 온프레미스 sLM을 활용한 Hyperledger Fabric 체인코드 취약점 탐지 연구: 신규성 검증 및 심층 문헌 분석 보고서"
> 생성일: 2026-02-09
> 도구: Gemini Deep Research
> 분석 요약: `02_related_work_survey.md`

---

## 1. 서론 (Introduction)

### 1.1 연구 배경 및 필요성

블록체인 기술은 탈중앙화된 신뢰 환경을 구축하는 핵심 인프라로 자리 잡았으며, 특히 기업형(Enterprise) 환경에서는 허가형 블록체인(Permissioned Blockchain)인 **Hyperledger Fabric (HLF)**이 사실상의 표준으로 통용되고 있다. 금융, 공급망 관리(SCM), 헬스케어 등 고도의 보안과 프라이버시가 요구되는 산업군에서 HLF의 도입이 가속화됨에 따라, 시스템의 핵심 로직을 담당하는 스마트 컨트랙트(이하 체인코드, Chaincode)의 보안성은 전체 네트워크의 신뢰성을 결정짓는 척도가 되었다.

그러나 체인코드의 보안 취약점은 기존의 이더리움(Ethereum) 기반 스마트 컨트랙트(Smart Contract)와는 다른 양상을 보인다. 이더리움이 Solidity라는 도메인 특화 언어(DSL)를 사용하는 것과 달리, HLF는 Go, Java, Node.js와 같은 범용 프로그래밍 언어를 사용하여 체인코드를 작성한다. 이는 개발의 유연성을 제공하는 동시에, 범용 언어가 내포한 비결정적(Non-deterministic) 실행 오류, 동시성 문제(Concurrency Issues), 그리고 HLF 특유의 아키텍처인 실행-순서화-검증(Execute-Order-Validate) 모델에서 기인한 MVCC(Multi-Version Concurrency Control) 충돌 등의 고유한 취약점을 야기한다.

최근 인공지능(AI), 특히 대규모 언어 모델(LLM)의 비약적인 발전은 코드 보안 감사 분야에 혁신을 가져왔다. GPT-4와 같은 상용 모델은 코드의 문맥을 이해하고 논리적 오류를 탐지하는 데 탁월한 성능을 입증하였다. 그러나 기업 환경에서 이러한 외부 API 기반의 LLM을 사용하는 것은 데이터 주권(Data Sovereignty) 및 프라이버시 침해(Privacy Leakage) 문제를 야기한다. 기업의 핵심 비즈니스 로직이 담긴 체인코드를 외부 서버로 전송하는 것은 보안 정책상 허용되지 않는 경우가 빈번하며, 이는 AI 기반 보안 감사 도구의 현장 도입을 저해하는 주요 장벽으로 작용하고 있다.

### 1.2 연구 목표 및 범위

본 보고서는 "Hyperledger Fabric 체인코드(Go/Java)의 보안 취약점을 탐지하기 위해, 외부 API가 아닌 **로컬 sLM(Small Language Models, 예: Llama-3, Qwen)**을 활용하는 연구"의 학술적 신규성(Novelty)을 검증하고, 선행 연구의 한계점을 분석하여 연구의 방향성을 확립하는 것을 목적으로 한다.

이를 위해 2023년부터 2026년까지 발표된 최신 문헌을 대상으로 (1) Hyperledger Fabric 보안 연구의 현황, (2) 스마트 컨트랙트 취약점 탐지에서의 AI 적용 동향, (3) 프라이버시 보존형 로컬 모델의 효용성을 중점적으로 분석하였다. 특히, 이더리움/Solidity 중심의 주류 연구 흐름과 구별되는 HLF/Go 기반 연구의 희소성을 입증하고, 기업형 블록체인 환경에서 로컬 sLM 도입의 필연성을 논리적으로 구축하고자 한다.

## 2. 블록체인 보안 감사 기술의 진화와 현황

### 2.1 정적/동적 분석에서 인공지능으로의 패러다임 전환

초기 스마트 컨트랙트 보안 감사는 정해진 규칙(Rule-based)에 따라 코드를 검사하는 정적 분석(Static Analysis)과 코드를 실행하며 오류를 찾는 동적 분석(Dynamic Analysis)이 주류를 이루었다. 이더리움 생태계에서는 Slither, Mythril 등의 도구가 표준으로 자리 잡았으며, HLF 생태계에서는 Chaincode Scanner, GoLiSA, HFContractFuzzer 등이 개발되었다.

그러나 이러한 전통적 방법론은 다음과 같은 한계에 직면해 있다:

- 높은 오탐율(High False Positive Rate): 문맥을 이해하지 못하고 단순 패턴 매칭에 의존하여 실제로는 안전한 코드를 취약점으로 경고하는 경우가 많다.
- 복잡한 논리 오류 탐지 불가: 비즈니스 로직에 깊숙이 내재된 취약점은 단순한 코드 패턴으로는 식별이 불가능하다.
- 확장성 부족: 새로운 유형의 취약점이 발견될 때마다 수동으로 탐지 규칙을 업데이트해야 한다.

이러한 한계를 극복하기 위해 딥러닝(DL) 및 그래프 신경망(GNN)을 활용한 연구가 2023-2024년 사이 활발히 진행되었으며, 2025년에 이르러서는 LLM을 활용한 생성형 감사(Generative Auditing) 기술이 급부상하고 있다.

### 2.2 Hyperledger Fabric 보안 연구의 특수성

이더리움과 달리 HLF 체인코드는 범용 언어(Go, Java)로 작성되므로, 기존의 소프트웨어 취약점(예: 메모리 누수, 포인터 오류)과 블록체인 특화 취약점(예: 원장 불일치, 비결정적 명령 사용)이 혼재되어 나타난다. 문헌 조사 결과, HLF 대상 보안 연구는 이더리움 대비 절대적으로 부족한 실정이다.

- 언어적 복잡성: Solidity는 튜링 완전하지만 스마트 컨트랙트 전용으로 설계되어 상대적으로 분석이 용이한 반면, Go 언어는 고루틴(Goroutine), 채널(Channel) 등 복잡한 동시성 제어 기능을 포함하고 있어 취약점 탐지 모델링이 훨씬 난해하다.
- 데이터셋의 부재: 이더리움은 퍼블릭 블록체인 특성상 SmartBugs와 같은 대규모 오픈 데이터셋이 존재하나, HLF는 프라이빗 블록체인 특성상 공개된 체인코드 데이터셋이 극히 드물다. 이는 데이터 중심의 AI 모델 학습에 치명적인 제약이 되어 왔다.

## 3. 선행 연구 심층 분석 (Comprehensive Related Work)

### 3.1 Group A: Hyperledger Fabric 취약점 탐지 (Non-LLM Approaches)

이 그룹의 연구들은 대부분 정적 분석, 퍼징(Fuzzing), 또는 머신러닝/지식그래프(Knowledge Graph) 기법을 사용하며, LLM을 적용한 사례는 발견되지 않았다.

1. 논문 제목: VulFinder: Exploring Chaincode Vulnerabilities More Effectively and Efficiently Using Knowledge Graph Based Defect Pattern Matching / 저자: Li et al. / 연도: 2025 / 학회: IEEE TSE / 핵심 요약: 소스 코드로부터 지식 그래프를 구축, SPARQL 쿼리로 22가지 결함 패턴 매칭. 성과: 98.87% Recall. 한계점: 사전 정의된 패턴 의존, LLM 미사용.

2. 논문 제목: Understanding and Detecting Privacy Leakage Vulnerabilities in Hyperledger Fabric Chaincodes / 저자: Chen et al. / 연도: 2024 / 학회: IEEE ISSRE / 핵심 요약: PDChecker, 데이터 흐름 분석. 성과: 956개 체인코드 분석, 67.78%에서 취약점, 10개 Zero-day. 한계점: 프라이버시 유출만 다룸.

3. 논문 제목: Smart Contract Vulnerability Detection Techniques for Hyperledger Fabric / 연도: 2023 / 학회: IEEE Conference / 핵심 요약: GoLiSA, 추상 해석. 한계점: 2023년 연구, 분석 규칙 수동 설계.

### 3.2 Group B: AI/LLM 기반 취약점 탐지 (이더리움 중심)

1. EVuLLM / 2025 / MDPI Electronics / Llama-3, CodeGemma, QLoRA Fine-tuning, 94.78% 정확도. 한계점: 이더리움/Solidity만 대상.

2. SmartGuard / 2025 / Expert Systems with Applications / DeBERTa+BiLSTM+CNN, 0.91 F1-score. 한계점: 단순 분류 초점.

3. LegiCode / 2025 / Empirical Software Engineering / 법률→체인코드 변환 LLM. 한계점: 코드 생성 목적.

### 3.3 Group C: 프라이버시 및 로컬 모델 트렌드

1. Sachan et al. / 2025 / arXiv, IEEE / 고위험 분야 LLM 도입 시 프라이버시 장벽, 로컬 LLM 필요성.

## 4. 연구 질문에 대한 상세 검증

### 4.1 Direct Conflict Check
A: 존재하지 않음 (No Direct Conflict). "HLF 체인코드 대상 로컬 sLM 기반 취약점 탐지"는 완전한 공백(White Space).

### 4.2 Baseline Analysis
A: HLF 분야 최신 SOTA는 VulFinder (2025). HLF/Go 체인코드 딥러닝 모델은 거의 미보고.

### 4.3 Trend Analysis
A: 다수 존재. EVuLLM, Sachan et al. 등이 로컬 모델 필요성 명시.

## 5. 연구의 신규성 및 차별화 전략

### 5.1 차별점
1. 최초의 도메인 융합 (이더리움→HLF/Go)
2. Go 언어 특화 Code-LLM 효율성 검증
3. 데이터 주권 해결 (On-premise sLM)

### 5.2 극복 과제
- 데이터셋 구축 (공개 데이터셋 부재)
- 환각 제어 (RAG 기법 고려)

## 6. 결론

해당 주제는 치명적인 중복 연구가 존재하지 않는 블루오션. (1) HLF 특화 취약점 데이터셋 구축, (2) QLoRA 등 효율적 로컬 모델 튜닝, (3) VulFinder 등 기존 도구와의 비교 평가를 통해 우수성 입증 가능.

### 주요 인용 문헌 요약

| 분류 | 논문 | 핵심 내용 | Gap |
|:-----|:-----|:---------|:----|
| 비교군 | VulFinder | KG 기반 HLF Go 취약점 탐지, 98% Recall | 패턴 매칭 한계, LLM 추론 능력 부재 |
| 방법론 | EVuLLM | 이더리움 대상 로컬 sLM + QLoRA, GPT-4 수준 | HLF/Go 미적용, 데이터셋 Solidity 한정 |
| 타겟 | PDChecker | HLF PDC 오용 탐지 (정적 분석) | 특정 취약점에만 국한 |
| 논리 | Sachan et al. | 고위험 시스템 로컬 LLM/블록체인 활용 | HLF 체인코드 감사 구체적 실험 부재 |
