# FB_260211_advisor_gpt52

## Metadata
- **Source**: advisor (GPT 5.2 활용 리뷰)
- **Date**: 2026-02-11
- **Target Draft**: 260210_v30_서브섹션병합.tex (제출본, 260210_2346_최종제출)
- **Tool**: GPT 5.2 (교수님 컨텍스트 윈도우)
- **Prompt**: "본 논문을 리뷰어 입장에서 검토하고 평가한다. 그리고 크리티컬한 수정 부분이 있다면 그 이유를 설명하고, 기존 문장을 1:1로 대체할 수 있도록 수정 문장을 제시한다. 위치와 함께."
- **Overall Recommendation**: Major revision (수정 후 재심)
- **Severity**: major
- **Action**: 의견 확인만, 교정본 나오면 그 때 수정 포인트 검토 (교수님 지시)

## 피드백 원문

### 1) 리뷰어 관점 요약 평가

#### 논문 요약

본 논문은 Hyperledger Fabric(HLF) 체인코드에서 **피어 간 실행 결과 불일치(endorsement mismatch)**를 유발하는 비결정성 요인(예: time.Now(), map iteration, phantom read 등)을 대상으로, 로컬에서 구동 가능한 7–8B급 sLM이 이를 얼마나 잘 탐지·변별하는지(특히 FP 억제) 실험적으로 탐색합니다. 15개 마이크로 벤치마크(취약 9, benign trap 6)와 GoLiSA 벤치마크 일부(657개 파일 중 5개 GT 파일)를 사용하고, 프롬프트 전략(Zero-shot/Few-shot/CoT/JSON mode), 식별자 난독화, 클라우드 LLM/도구(Semgrep) 비교를 수행합니다.

#### 강점(Strengths)

문제 설정이 실제적: 퍼미션드 환경에서 체인코드 외부 전송이 어렵다는 동기(기밀성)는 설득력이 큼.

FP(오탐) 억제 중심의 평가 설계: benign trap을 넣어 "탐지율(TPR)"뿐 아니라 "변별(TNR)"을 전면에 둔 점이 좋음.

다각도 실험축: (i) 프롬프트 전략, (ii) 난독화, (iii) 외부 코퍼스(GoLiSA) 일부 검증 등 "feasibility" 논문으로서 구성은 탄탄함.

결과 해석을 과도하게 일반화하지 않으려는 노력: N=15 한계를 논문 내에서 명시.

#### 주요 약점/우려(Major Concerns)

용어 정확성(매우 중요)
본문 전반에서 "consensus-layer vulnerability"라는 표현을 쓰는데, Fabric에서 체인코드는 ordering service의 합의 레이어에서 실행되지 않으며, 문제의 핵심은 endorsement/validation 단계의 비결정성입니다. 이 용어는 블록체인 연구 커뮤니티에서 "합의 프로토콜(PBFT/PoW 등) 취약점"으로 오해될 소지가 커서 정확한 정의/표현 조정이 필요합니다(크리티컬).

재현성 정보 부족(중요)
마이크로 벤치마크 15개 파일, 프롬프트(P1~P3), 판정 규칙(v1/v2) 등은 결과에 결정적이나, 논문만으로 완전 재현이 어렵습니다(부록/아티팩트 공개/프롬프트 전문 제공 권장).

실험 설정의 일관성 문제(중요)
"모든 모델 context window 4096"이라고 해놓고, 이후 GoLiSA 실험에서 n_ctx=16384를 사용합니다. 예외를 명시하지 않으면 방법론 서술의 신뢰성이 떨어집니다(크리티컬).

과한/근거 약한 서술(중요)

"zero marginal cost", "fastest among all models tested", "internalized Go API semantics" 등은 과장/추정으로 읽힐 수 있어 문장 톤을 조정하는 편이 안전합니다(크리티컬).

인용 오류(중요)
llama-cpp-python 0.3.16 [2], [3]처럼 도구를 모델 논문으로 인용한 형태는 명백한 서지 오류로 보입니다(크리티컬).

베이스라인의 설득력(보완 필요)
Semgrep default ruleset만으로 "전통 도구는 0개 탐지" 결론을 강하게 끌면 반발이 있을 수 있습니다. (논문은 "default"임을 말하지만 결론부 문장이 더 단정적으로 읽힘)

#### 종합 의견(Recommendation)

결론: "Feasibility study"로서 방향과 실험 구성은 좋으나, 용어/방법론 일관성/인용/과한 주장은 리뷰 단계에서 치명적 지적을 받을 가능성이 큽니다.

추천: Major revision(수정 후 재심) 성격이 강합니다.

신뢰도: 중간 이상(논문 내 근거/표의 수치가 명확하나, 서술 일관성과 용어는 수정 필요)

### 2) 크리티컬 수정 제안 (1:1 문장 대체 / 위치 포함)

#### [C1] "consensus-layer" 용어 오해 방지(정의 강화)

위치: p.1, Section I. INTRODUCTION, 1문단 2번째 문장

수정 이유: Fabric에서 체인코드 이슈는 엄밀히 합의(ordering) 레이어가 아니라 endorsement/validation 비결정성 이슈입니다. 오해를 차단해야 함.

원문

This design renders consensus-layer vulnerabilities—such as calling time.Now() or iterating Go maps—invisible to conventional testing, since divergence manifests only when peers compare read-write sets.

대체 문장

This design renders endorsement/validation nondeterminism vulnerabilities (sometimes loosely referred to as "consensus-layer" issues in HLF)—such as calling time.Now() or iterating Go maps—invisible to conventional testing, since divergence manifests only when peers compare read-write sets.

#### [C2] 초록 첫 문장: 인과("cause") 단정 완화 + 용어 정합

위치: p.1, Abstract, 1문장

수정 이유: "cause"는 단정이 강함. 또한 [C1]과 용어 정합 필요.

원문

Abstract—Hyperledger Fabric (HLF) chaincode is susceptible to consensus-layer vulnerabilities—nondeterministic timestamps, global variable mutation, phantom reads—that cause silent endorsement failures without runtime errors.

대체 문장

Abstract—Hyperledger Fabric (HLF) chaincode is susceptible to endorsement/validation nondeterminism vulnerabilities—nondeterministic timestamps, global variable mutation, phantom reads—that can cause silent endorsement failures without explicit runtime errors.

#### [C3] "zero marginal cost" 과장 표현 완화

위치: p.1, Section I. INTRODUCTION, 중간 문장(로컬 sLM 장점 서술)

수정 이유: GPU 전력/시간 비용이 존재하므로 "zero marginal cost"는 과장으로 지적될 수 있음.

원문

Locally deployed sLMs on consumer GPUs can operate offline at zero marginal cost, preserving code confidentiality—a requirement specific to permissioned environments where chaincode embodies proprietary business logic.

대체 문장

Locally deployed sLMs on consumer GPUs can operate offline at low marginal cost per analysis while preserving code confidentiality—a requirement specific to permissioned environments where chaincode embodies proprietary business logic.

#### [C4] "Traditional static analysis tools…" 과도한 일반화 완화

위치: p.1, Section I. INTRODUCTION, "Traditional static analysis tools…" 문장

수정 이유: "전통 정적 분석은 도메인 지식이 없다"는 일반화로 반박 가능(특화 도구 존재). "General-purpose"로 한정하는 게 안전.

원문

Traditional static analysis tools operate on syntactic patterns and lack the domain knowledge to reason about HLF's endorsement semantics.

대체 문장

General-purpose static analysis tools typically operate on syntactic patterns and may not capture HLF endorsement semantics without Fabric-specific rules or analyses.

#### [C5] llama-cpp-python 인용 오류 수정(서지 오류)

위치: p.2, Section III-C. Models, "Local models run via…" 문장

수정 이유: llama-cpp-python을 [2],[3](모델 기술보고서)로 인용하는 것은 명백한 인용 오류로 보임.

원문

Local models run via llama-cpp-python 0.3.16 [2], [3] with full GPU offload on an NVIDIA RTX 3090 Ti (24,564 MiB VRAM).

대체 문장(최소 수정: 잘못된 인용 제거)

Local models run via llama-cpp-python 0.3.16 with full GPU offload on an NVIDIA RTX 3090 Ti (24,564 MiB VRAM).

(권장 추가 조치) 가능하면 참고문헌에 llama-cpp-python 항목을 별도로 추가하는 편이 학술적으로 더 깔끔합니다. URL이 필요하면 참고문헌에 넣되, 본문 작성 환경에 맞춰 처리하세요.

#### [C6] Context window 서술 일관성 확보(4096 vs 16384)

위치: p.2, Section III-C. Models, "All models use…" 문장

수정 이유: 뒤에서 GoLiSA 실험에 n_ctx=16,384를 사용하므로 "All models … 4,096"은 불일치로 읽힘.

원문

All models use temperature 0.1, max tokens 2,048, context window 4,096.

대체 문장

Unless otherwise stated, we use temperature 0.1 and max tokens 2,048 with n_ctx=4,096 (we increase n_ctx to 16,384 for the GoLiSA corpus evaluation in Section III-G).

#### [C7] JSON mode 효과 과장("structurally prevents") 완화

위치: p.2, Section III-D. Prompt Strategies, "JSON mode uses…" 문장

수정 이유: 출력 스키마 강제는 "형식 위반"은 줄이지만, "논리적 자기모순"을 구조적으로 완전 방지한다고 단정하긴 어려움. 표현을 "reduce"로 완화 권장.

원문

JSON mode uses grammar-based output enforcement in llama-cpp-python, which structurally prevents self-contradictory responses; cloud APIs offer architecturally different structured output mechanisms, so JSON mode evaluation is limited to local models.

대체 문장

JSON mode uses grammar-based output enforcement in llama-cpp-python to constrain responses to a fixed schema and reduce label inconsistency; because structured-output support differs across cloud providers/endpoints, we evaluate JSON mode only for local models.

#### [C8] Semgrep baseline 문장 가독성/정확성 강화

위치: p.2, Section III-F. Traditional Tool Baseline, 1문장

수정 이유: 수동태/구문이 어색하고, "default ruleset"임을 더 명확히 하는 편이 좋음.

원문

Semgrep 1.151.0 [10] with the default p/security-audit ruleset is run on all 15 files without custom HLF-specific rules.

대체 문장

We run Semgrep 1.151.0 [10] with the default p/security-audit ruleset on all 15 files without any HLF-specific custom rules.

#### [C9] 평가 지표(TPR/TNR) 정의의 모호성 완화

위치: p.3, Section III-H. Evaluation Metrics, TPR 정의 문장

수정 이유: "consensus-relevant finding"이 무엇인지 더 명확한 문구가 필요(라벨링 기준 오해 방지).

원문

TPR: fraction of vulnerable files with at least one consensus-relevant finding reported.

대체 문장

TPR: proportion of vulnerable files for which the model reports at least one finding mapped to the targeted consensus-layer classes.

#### [C10] 평가 지표(TNR)도 동일하게 명확화

위치: p.3, Section III-H. Evaluation Metrics, TNR 정의 문장

수정 이유: "benign files"가 일반적 안전을 의미하는지, "benign trap(합의/endorsement 관점에서만 benign)"인지 혼동 가능.

원문

TNR: fraction of benign files correctly identified as safe.

대체 문장

TNR: proportion of benign-trap files for which the model outputs a final "safe" verdict under our consensus-only labeling.

#### [C11] Table II 각주: "initial experiment" 표현이 본문(5회 반복)과 충돌

위치: p.3, Table II 하단 각주, 첫 문장

수정 이유: Section III-H에서 "All experiments were repeated five times…"라고 했는데, Table II는 zero/few-shot이 "initial experiment"라고 되어 있어 일관성 문제가 생김.

원문

Zero-shot and few-shot values are from the initial experiment; CoT values for cloud models represent the median of five independent runs.

대체 문장

All prompt strategies were repeated five times; for cloud models we report the median across runs when variation was observed (‡).

#### [C12] "fastest among all models tested" 근거 부족(측정 범위 불명확)

위치: p.4, Section IV-F. External Validation Results 말미 문장

수정 이유: 클라우드 모델 지연시간은 네트워크/쿼터/서버상태에 좌우되어 "fastest" 단정은 공격 포인트가 됩니다.

원문

Qwen completes the 15-file micro-benchmark audit in 59.1 seconds (3.9 s/file) on local hardware with no network dependency, the fastest among all models tested.

대체 문장

Qwen completes the 15-file micro-benchmark audit in 59.1 seconds (3.9 s/file) on local hardware with no network dependency.

#### [C13] "internalized Go API semantics" 추정 표현 완화

위치: p.4, Section V-A. Specialist vs. Generalist Discrimination, 관련 문장

수정 이유: 내부 메커니즘("internalized") 단정은 근거 요구를 받기 쉬움. 관측 기반 표현으로 바꾸는 것이 안전.

원문

Qwen2.5-Coder-7B, trained on a code-heavy corpus [2], appears to have internalized Go API semantics sufficient to trace whether nondeterministic values reach PutState.

대체 문장

In our micro-benchmark, Qwen2.5-Coder-7B's outputs are consistent with tracing whether nondeterministic values reach PutState, suggesting stronger data-flow sensitivity than keyword-triggered heuristics.

#### [C14] 데이터/학습코퍼스 추정("likely includes") 완화

위치: p.5, Section V-D. Data Contamination Analysis, 1문장

수정 이유: "likely includes"는 근거 요구를 받기 쉬운 주장. "may include / we do not have direct evidence"로 방어력 강화.

원문

Qwen's training data likely includes HLF documentation and public chaincode examples.

대체 문장

Qwen's pretraining data may include HLF documentation and public chaincode examples, but we do not have direct evidence of dataset overlap.

#### [C15] 결론의 Semgrep 문장: "default ruleset" 조건을 재명시

위치: p.6, Section VI. CONCLUSION, "Semgrep detects…" 문장

수정 이유: 결론에서 조건이 빠지면 "Semgrep은 못한다"로 과잉 일반화로 읽힐 수 있음.

원문

Semgrep detects zero consensus-layer vulnerabilities on both datasets.

대체 문장

With the default p/security-audit ruleset, Semgrep detects zero consensus-layer vulnerabilities on both datasets.

#### [C16] 결론 "parity" 단정 완화(평가 범위 명시)

위치: p.6, Section VI. CONCLUSION, "A code-specialist…" 문장

수정 이유: "parity"는 범위를 명시해야 안전(본 논문도 "feasibility" 성격). "Within our evaluation" 추가 권장.

원문

A code-specialist sLM achieves discrimination parity with prompt-engineered cloud models while preserving code confidentiality, offering a practical alternative for privacy-sensitive HLF environments.

대체 문장

Within our evaluation, a code-specialist sLM matches the best-performing prompt-engineered cloud models on the micro-benchmark while preserving code confidentiality, suggesting a practical alternative for privacy-sensitive HLF environments.

#### [C17] (표 텍스트 1:1) Table I 오타 수정

위치: p.2, Table I (MODEL CONFIGURATIONS), 모델명 셀

수정 이유: "Prob"는 오타로 보이며, 본문/표2에서는 "Pro" 사용. 일관성 필요.

원문(표 셀)

Gemini 2.5 Prob

대체(표 셀)

Gemini 2.5 Pro

### 3) 추가로 "수정 권고(문장 치환 외)" — 리뷰어가 요구할 가능성이 큰 항목

문장 1:1 치환을 넘어서는 구조적 보완이지만, 채택 가능성을 크게 올릴 요소들입니다.

프롬프트(P1~P3) 전문, benign trap 설계 원칙, 판정 규칙(v1/v2) 공개(부록/레포지토리).

비교 베이스라인 강화: 가능하면 GoLiSA/VulFinder 같은 HLF 특화 도구를 직접 실행하거나, "본 논문 범위에서 제외한 이유/한계"를 더 강하게 명시.

라벨링 기준 명문화: "consensus-relevant finding"의 매핑 규칙(키워드/정규식/수동판정 여부)과 예시 2–3개 제시.

용어 통일: 논문 전체에서 "consensus-layer"를 유지할지, "endorsement/validation nondeterminism"로 전환할지 결정 후 일관 적용.

## 파생 안건

| 안건 ID | 요약 | 우선순위 | 대상 섹션 | 관련 C항목 |
|:--------|:-----|:---------|:----------|:-----------|
| ISS_019 | "consensus-layer" 용어 정합성 + 평가 지표 정의 명확화 | critical | meta, I, III | C1, C2, C9, C10 |
| ISS_020 | 과장/추정 표현 일괄 완화 (7건) | critical | I, III, IV, V, VI | C3, C4, C7, C12, C13, C14, C16 |
| ISS_021 | 서지 오류 + 방법론 서술 일관성 (4건) | critical | III, IV | C5, C6, C11, C17 |
| ISS_022 | Semgrep/결론부 조건 명시 | high | III, VI | C8, C15 |
| ISS_023 | 재현성 정보 보강 (구조적 보완 권고) | high | III, IV, V | 추가 권고 4건 |

## 비고
- 교수님 지시: "일단 의견만 확인하고 교정본 나오면 그 때 수정할 포인트를 검토하고 챙기라고 하셨어"
- 전체 17개 1:1 수정 제안(C1~C17) + 4개 구조적 보완 권고
- 종합 판정: Major revision
- 모든 안건 상태: open (교정본 대기)
