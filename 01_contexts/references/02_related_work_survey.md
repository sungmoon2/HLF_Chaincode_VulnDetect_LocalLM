# 02. 관련 연구 조사 (Deep Research 1: Related Work Survey)
> 추출 원본: Gemini 딥리서치 결과 (2026-02-09)
> 제목: "프라이버시 보존형 온프레미스 sLM을 활용한 Hyperledger Fabric 체인코드 취약점 탐지 연구: 신규성 검증 및 심층 문헌 분석 보고서"
> 검색 범위: 2023년~2026년 논문(Conference, Journal, Pre-print)

## 1. 검색 전략 (키워드 조합)

- **Group A (Target Domain):** "Hyperledger Fabric", "Permissioned Blockchain", "Private Blockchain", "Chaincode", "Go chaincode"
- **Group B (Task):** "Vulnerability Detection", "Security Audit", "Bug Finding", "Static Analysis", "Code Analysis"
- **Group C (Methodology):** "Large Language Model", "LLM", "sLM", "Small Language Model", "Local Model", "Privacy-preserving", "Llama", "Mistral", "Qwen", "Fine-tuning"

## 2. 핵심 질문 및 답변

### 2.1 Direct Conflict Check (치명적 중복 확인)
**Q:** HLF 체인코드 취약점 탐지를 위해 '로컬 sLM'을 적용하고 성능을 비교한 연구가 있는가?

**A:** 존재하지 않음 (No Direct Conflict).
- EVuLLM: 로컬 sLM 사용했으나 이더리움 대상
- VulFinder: HLF 대상이나 지식 그래프 기반
- LegiCode: HLF와 LLM 사용했으나 코드 생성이 목적
- **결론:** "HLF 체인코드 대상 로컬 sLM 기반 취약점 탐지"는 완전한 공백(White Space)

### 2.2 Baseline Analysis (비교군 확인)
**Q:** HLF 취약점 탐지에 적용된 최신 ML/DL 연구는?

**A:** 지식 그래프 및 정적 분석이 주류, 딥러닝 적용 미미.
- HLF 분야 최신 베이스라인(SOTA): VulFinder (2025)
- HLF/Go 체인코드를 효과적으로 학습시킨 딥러닝 모델은 거의 미보고

### 2.3 Trend Analysis (경향성 파악)
**Q:** LLM 프라이버시 문제를 지적하며 로컬 모델 필요성을 언급한 논문이 있는가?

**A:** 다수 존재.
- EVuLLM: "로컬 하드웨어에 배포 가능한 경량 모델은 데이터 프라이버시 강화 필수"
- Sachan et al.: "민감한 정보 포함 프롬프트가 모델 학습에 재사용/유출될 위험"

## 3. 선행 연구 상세 분석

### 3.1 Group A: HLF 취약점 탐지 (Non-LLM Approaches)

#### 논문 1: VulFinder
| 항목 | 내용 |
|:-----|:-----|
| 제목 | VulFinder: Exploring Chaincode Vulnerabilities More Effectively and Efficiently Using Knowledge Graph Based Defect Pattern Matching |
| 저자 | Li et al. |
| 연도 | 2025 |
| 학회 | IEEE Transactions on Software Engineering (TSE) |
| 핵심 요약 | 소스 코드 → 지식 그래프(KG) 구축 → SPARQL 쿼리로 22가지 결함 패턴 매칭 |
| 성과 | 자체 데이터셋에서 98.87% Recall |
| 한계점 (Gap) | 사전 정의된 패턴만 탐지, 변종 공격/복잡한 비즈니스 로직 오류 추론 불가, 수정 방안 제안 능력 없음, LLM 미사용 |

#### 논문 2: PDChecker
| 항목 | 내용 |
|:-----|:-----|
| 제목 | Understanding and Detecting Privacy Leakage Vulnerabilities in Hyperledger Fabric Chaincodes |
| 저자 | Chen et al. |
| 연도 | 2024 |
| 학회 | IEEE ISSRE |
| 핵심 요약 | PDC(Private Data Collection) 오용 탐지, 데이터 흐름 분석으로 프라이빗 데이터 유출 경로 추적 |
| 성과 | 956개 체인코드 분석, 67.78%에서 취약점 발견, 10개 Zero-day 식별 |
| 한계점 (Gap) | '데이터 프라이버시 유출'만 다룸, 일반 보안 위협 미포함, 규칙 기반 탐지 한계 |

#### 논문 3: GoLiSA 관련 연구
| 항목 | 내용 |
|:-----|:-----|
| 제목 | Smart Contract Vulnerability Detection Techniques for Hyperledger Fabric |
| 연도 | 2023 |
| 학회 | IEEE Conference |
| 핵심 요약 | GoLiSA 정적 분석 도구, 추상 해석(Abstract Interpretation) 기법으로 Go 언어 의미론적 분석 |
| 한계점 (Gap) | 2023년 연구, 최신 AI 기술(Transformer 등) 미적용, 분석 규칙 수동 설계 |

### 3.2 Group B: AI/LLM 기반 취약점 탐지 (이더리움 중심)

#### 논문 4: EVuLLM
| 항목 | 내용 |
|:-----|:-----|
| 제목 | EVuLLM: Ethereum Smart Contract Vulnerability Detection Using Large Language Models |
| 연도 | 2025 |
| 학회 | MDPI Electronics |
| 핵심 요약 | Llama-3, CodeGemma 등을 QLoRA로 Fine-tuning, 소비자용 하드웨어에서 구동 가능한 로컬 환경 구축 |
| 성과 | GPT-4에 필적하는 94.78% 정확도, 데이터 외부 유출 없이 보안 감사 가능 입증 |
| 한계점 (Gap) | 오직 이더리움/Solidity만 대상, HLF/Go에 대한 적용 가능성 미검증 |

#### 논문 5: SmartGuard
| 항목 | 내용 |
|:-----|:-----|
| 제목 | SmartGuard: An LLM-enhanced Framework for Smart Contract Vulnerability Detection |
| 연도 | 2025 |
| 학회 | Expert Systems with Applications |
| 핵심 요약 | DeBERTa + BiLSTM + CNN 결합 하이브리드 모델 |
| 성과 | 0.91 F1-score |
| 한계점 (Gap) | 복잡도 높음, 생성형 AI의 설명 가능성/수정 제안보다 단순 분류에 초점 |

#### 논문 6: LegiCode
| 항목 | 내용 |
|:-----|:-----|
| 제목 | LegiCode: A Blockchain-Legal LLM Framework for Real-Time Compliance |
| 연도 | 2025 |
| 학회 | Empirical Software Engineering |
| 핵심 요약 | 법률 텍스트를 HLF 체인코드로 변환하는 LLM(LegiLM) 개발 |
| 한계점 (Gap) | 취약점 '탐지'가 아닌 코드 '생성'에 초점 |

### 3.3 Group C: 프라이버시 및 로컬 모델 트렌드

#### 논문 7: Sachan et al.
| 항목 | 내용 |
|:-----|:-----|
| 제목 | Responsible LLM Deployment for High-Stake Decisions (PrivChatGPT 등) |
| 저자 | Sachan et al. |
| 연도 | 2025 |
| 학회 | arXiv, IEEE 등 |
| 핵심 요약 | 금융/헬스케어 분야에서 LLM 도입 시 데이터 프라이버시가 최대 장벽, 로컬 LLM/연합 학습/블록체인 활용 감사 로그 기록 제안 |
| 활용 방안 | "기업형 블록체인인 HLF의 도입 목적(프라이버시)과 로컬 sLM의 도입 목적(데이터 보호)은 일맥상통" 논리 전개 근거 |

## 4. 차별화 전략

### 4.1 연구의 차별점 (Unique Selling Proposition)
1. **최초의 도메인 융합**: 이더리움 편중된 '로컬 LLM 보안 감사' 방법론을 HLF/Go 도메인으로 최초 이식
2. **Go 언어 특화 튜닝**: Solidity와 달리 범용 언어 Go의 특성(Pointer, Struct, Concurrency)을 이해하는 Code-LLM의 효율성 검증
3. **데이터 주권 해결**: 기업형 블록체인의 핵심 요구사항 '데이터 프라이버시'를 On-premise sLM으로 해결

### 4.2 극복 과제
- **데이터셋 구축**: HLF 체인코드 취약점 데이터셋은 공개된 것이 거의 없음 (VulFinder 저자들도 직접 구축)
- **환각(Hallucination) 제어**: Go는 문맥이 길고 복잡하여 sLM 환각 가능성 존재

## 5. 인용 문헌 요약 (이 보고서에서 식별된 논문)

| 분류 | 논문 | 핵심 내용 | Gap |
|:-----|:-----|:----------|:----|
| 비교군 | VulFinder | KG 기반 HLF Go 취약점 탐지, 98% Recall | 패턴 매칭 한계, LLM 추론 능력 부재 |
| 방법론 | EVuLLM | 이더리움 대상 로컬 sLM + QLoRA, GPT-4 수준 성능 | HLF/Go 미적용, 데이터셋 Solidity 한정 |
| 타겟 | PDChecker | HLF PDC 오용 탐지 (정적 분석) | 특정 취약점(프라이버시)에만 국한 |
| 논리 | Sachan et al. | 고위험 결정 시스템에서 로컬 LLM/블록체인 활용 | HLF 체인코드 '감사' 구체적 실험 부재 |
