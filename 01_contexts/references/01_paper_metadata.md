# 01. 논문 메타데이터 (Paper Metadata)
> 추출 원본: Gemini 대화 (2026-02-09)
> 추출 기준: 사실 기반, 할루시네이션/추정 없음

## 1. 타겟 학회

| 항목 | 내용 |
|:-----|:-----|
| 학회명 | AMLDS 2026 (The 2nd International Conference on Advanced Machine Learning and Data Science) |
| 장소 | Kansai University, Osaka, Japan |
| 제출 마감 | 2026-02-10 |
| 학회 성격 | Machine Learning & Data Science (블록체인 자체보다 "블록체인에 ML 적용" 또는 "ML을 위해 블록체인 활용"이 핵심) |

## 2. 선정된 주제

- **주제 1** (선택됨): 로컬 sLM을 활용한 프라이빗 체인코드 보안 취약점 탐지 비교 연구

## 3. 제목 후보 (Gemini 대화에서 제시됨)

1. `Privacy-Preserving Vulnerability Detection in Hyperledger Fabric Chaincode: A Comparative Study of Local sLMs (Qwen2.5 vs. Llama 3)`
2. `Optimizing Privacy-Preserving Smart Contract Auditing: A Performance Analysis of Local sLMs in Hyperledger Fabric Environments`

## 4. 미선택 주제 (참고용)

- **주제 2**: 멀티모달(VLM)을 활용한 블록체인 네트워크 이상 징후 탐지
  - 제목(가제): `Multimodal Anomaly Detection in Permissioned Blockchains: Synergizing System Logs and Visual Dashboard Metrics via VLMs`
- **주제 3**: 바이오데이터 무결성을 위한 온체인 데이터 관리의 성능 병목 분석
  - 제목(가제): `Performance Bottleneck Analysis of On-Chain Genomic Data Verification: Balancing Integrity and Throughput in Healthcare Blockchain`

## 5. 핵심 문제 제기 (Motivation)

- 블록체인(특히 기업용 Fabric)의 체인코드는 기업의 기밀 로직을 담고 있어 외부 유출이 치명적
- ChatGPT, Claude 같은 API 기반 모델은 코드를 외부 서버로 전송해야 하므로 보안 규정상 사용 불가능
- 폐쇄망(On-premise)에서 돌아가는 경량화 모델(sLM)이 Go 언어 체인코드의 취약점을 얼마나 정확하게 찾아내는지 검증 필요

## 6. 논문 포지셔닝

- "Runtime Safety(안 죽는가?)" 가 아니라 "Consensus Safety(합의가 깨지는가?)" 관점
- "완전 자동화 도구"가 아닌 "인간 감사자(Human Auditor)를 보조하는 도구"
- Pre-deployment Audit (배포 전 감사) 단계에서의 정적 분석(Static Analysis)

## 7. 논문 구조 (Gemini 제안)

1. **Introduction**: 최근 AI가 블록체인 개발/운영을 돕고 있으나 프라이버시 문제 미해결
2. **Related Work**: 기존 연구는 이더리움(Solidity) 중심, HLF(Go) 연구 부족, GPT-4 성능만 강조하고 프라이버시 간과
3. **Methodology**: 모델/툴 버전 명시, 데이터셋 구축, 실험 설계
4. **Results**: 비교표, "로컬 모델도 탐지 가능" 주장
5. **Discussion**: Fuzzing vs Static Analysis 정당성, Limitations
6. **Conclusion**: 로컬 sLM은 비용 0원, 데이터 외부 유출 없음, 향후 Fine-tuning 방향
