# ISS_010: GoLiSA 데이터셋 Related Work 인용 필수

**Status**: open
**Priority**: critical
**Source Feedback**: FB_260209_1700_gemini_strategy
**Target Sections**: II, III, V
**Created**: 2026-02-09

## 문제 설명
딥리서치 결과 GoLiSA Artifact (ECOOP 2023, ~651개 HLF 체인코드)의 존재가 확인되었다. 이 데이터셋을 인용하지 않으면 리뷰어에게 "관련 연구 조사 부족(Poor Literature Review)"으로 Reject 사유가 된다.

## 딥리서치에서 확인된 사실 (독립 검증 완료 2026-02-09)
- **GoLiSA Artifact (ECOOP 2023)**: 딥리서치 보고서에서 "~651개 실제 HLF 체인코드"라고 주장 (원문 PDF에서 미검증)
- **Dagstuhl PDF**: https://drops.dagstuhl.de/storage/00lipics/lipics-vol263-ecoop2023/LIPIcs.ECOOP.2023.23/LIPIcs.ECOOP.2023.23.pdf (Open Access)
- **GitHub**: https://github.com/lisa-analyzer/go-lisa (HTTP 200 OK 확인, 6 stars, Java 기반 도구)
- **DARTS Artifact**: 4.98 GB OVA VM 이미지 (https://drops.dagstuhl.de/entities/document/10.4230/DARTS.9.2.23)
- **[오류 수정] Zenodo DOI 10.5281/zenodo.7896323**: 부정확 — 관련 없는 우즈벡어 논문으로 리다이렉트. 딥리서치 보고서의 DOI 정보는 잘못됨
- **GitHub repo 내 데이터셋 부재**: go-lisa repo에는 도구 코드 + 소규모 테스트 케이스(non-det 3개 디렉토리)만 포함. 651개 체인코드 미포함
- **VulFinder (IEEE TSE 2025)**: 합성/주입 기반, 22종 취약점 분류
- **zm-stack/Chaincode**: PDC(Private Data Collection) 프라이버시 누출 전용, GitHub 공개
- **Revive-CC**: 20개 파일, Chaincode Scanner와 비교 평가용, GitHub 공개
- **Hyperledger Labs Chaincode Analyzer**: 공식 실험 프로젝트 (archived), 테스트 fixture 포함

## 차별화 논리 (합의됨)
- GoLiSA: GitHub에서 수집(Mined)된 Raw Corpus. 정적 분석 도구의 재현율(Recall) 측정에 유리.
- 우리 15개: 의도적으로 설계(Engineered)된 Adversarial Set. LLM의 정밀도(Precision)와 추론 능력(Reasoning) 검증에 특화. 의미론적 함정(Benign Traps) 포함.
- 관계: 경쟁이 아닌 상호 보완(Complementary).

## 권장 조치
- Related Work에 GoLiSA를 인용하고 그 의의를 인정
- 동시에 수집형 데이터셋이 LLM 추론 정밀도 테스트에 갖는 한계를 지적
- "데이터셋이 전무하다"는 기존 주장을 삭제
- "기존 대규모 코퍼스(GoLiSA 등)는 정적 분석 도구의 재현율 검증에 유리하나, LLM의 문맥 기반 추론 정밀도를 측정하기 위한 적대적 함정(Adversarial Traps)이 부재하다"로 차별화

## 관련 이슈
- ISS_006: N=15 일반화 한계 (데이터셋 명칭 변경과 연결)
- ISS_013: GoLiSA 외부 검증 실험
- ISS_012: Deep Research 참고문헌 확보 (GoLiSA 원문 PDF 필요)
