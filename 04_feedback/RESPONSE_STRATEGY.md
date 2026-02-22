# 피드백 대응 전략서

**작성일**: 2026-02-09
**대상 피드백**: FB_260209_1500_reviewer
**제약 조건**: AMLDS 2026 제출 마감 2026-02-10
**하드웨어**: RTX 3090 Ti (24GB VRAM), 2 모델 로컬 보유

---

## 전략 프레임워크: 두 트랙 병행

| 트랙 | 설명 | 적용 대상 |
|:-----|:-----|:----------|
| **Track A: 실험 보강** | 추가 실험 실행 → 새 데이터 → Table 추가 | ISS_002, ISS_003, ISS_004 |
| **Track B: 논문 리프레이밍** | 주장 범위 조정 + Threats to Validity 강화 | ISS_001, ISS_005 |

---

## ISS_001: 데이터셋 규모 부족 (N=9) — critical

### 진단
리뷰어 비평의 핵심: N=9로 일반화하는 것은 통계적으로 무의미하다.

### 대응: Track B (리프레이밍)

**데이터셋 확장(50~100개)은 마감 전 불가능하다.** 실제 체인코드의 ground-truth 라벨링에는 도메인 전문가의 수동 검증이 필수적이며, 자동 라벨링은 새로운 타당성 위협을 만든다. 따라서 논문의 **주장 범위를 축소**하고, N=9가 정당한 이유를 적극 논증한다.

#### 조치 항목

**1) 논문 스코프를 "Feasibility Study"로 명시적 전환**
- Title 수정: "A Comparative Study" → "A Feasibility Study"
- Abstract에서 "demonstrate" → "explore whether", "provide preliminary evidence"
- Contribution 문구를 "we establish a methodology and present initial evidence" 톤으로 수정

**2) N=9의 설계 정당성을 Methodology에 추가 (신규 서브섹션)**
- 각 파일이 독립적인 취약점 클래스를 대표함 (단순 표본이 아닌 범주 대표)
- 6개 취약점 × 정탐/미탐 × 우선순위 = 파일당 3개 평가 축, 유효 관측 27회
- 3개 benign trap은 "adversarial probe" 역할 — discrimination test의 필수 요소
- 대표적 선행 연구에서도 유사 규모 데이터셋 사용 사례 인용 (있을 경우)

**3) Threats to Validity 강화**
- 현재 5줄짜리 bullet → 전용 서브섹션으로 확대
- "Dataset scale"을 첫 번째 항목으로 배치
- 문구: "Our dataset is designed for controlled evaluation of discrimination capability, not for statistical generalization. Each file represents a distinct vulnerability class..."
- Future Work에 "large-scale evaluation on real-world repositories" 명시

**4) Conclusion 톤 조정**
- "demonstrates" → "provides preliminary evidence"
- "argues for adoption" → "suggests potential for"
- Future Work 문장 강화

### 영향
논문의 학술적 정직성이 강화됨. "작은 데이터셋이지만 우리가 뭘 보여주려는지 명확하다"는 프레이밍.

---

## ISS_002: 비교군(Baseline) 부재 — critical

### 진단
두 가지 비교군이 모두 누락됨: (A) 전통 정적 분석 도구 (lower bound), (B) SOTA 클라우드 모델 (upper bound).

### 대응: Track A (실험 보강) + Track B (프레이밍)

#### A. 전통 도구 비교 — 실행 가능, 고효과

**go vet + staticcheck + semgrep를 9개 파일에 대해 실행한다.**

| 도구 | 설치/실행 난이도 | 예상 결과 |
|:-----|:----------------|:----------|
| `go vet` | Go 기본 내장 | HLF 합의 취약점 0건 탐지 (언어 수준 검사만 수행) |
| `staticcheck` | `go install` 1줄 | HLF 합의 취약점 0건 탐지 (Go 관용구 검사) |
| `semgrep` | pip install | HLF 전용 룰셋 부재 → 0건 또는 무관한 경고만 |

**예상 결과가 0건인 것 자체가 논문의 핵심 주장을 실증한다**: "Traditional tools are blind to consensus-layer vulnerabilities" — 이제 증거가 있다.

실행 절차:
1. 9개 .go 파일을 Go 모듈 구조로 래핑 (go.mod 생성)
2. `go vet ./...` 실행 → 출력 기록
3. `staticcheck ./...` 실행 → 출력 기록
4. `semgrep --config=auto *.go` 실행 → 출력 기록
5. 결과를 Table로 정리 (Tool × File × Findings)

논문 반영:
- Table 추가: "Traditional Tool Comparison" (go vet, staticcheck, semgrep 열)
- Introduction의 "largely blind" 주장에 "as empirically confirmed in Section IV-F" 추가
- Results에 신규 서브섹션 "Baseline Comparison with Traditional Tools" 추가

#### B. SOTA 클라우드 모델 비교 — 선택적

GPT-4o / Claude API 접근 가능 여부에 따라:

| 조건 | 전략 |
|:-----|:-----|
| API 접근 가능 + 비용 허용 | 동일 프롬프트로 9개 파일 감사 → Table에 열 추가 |
| API 접근 불가 또는 비용 부담 | Threats to Validity + Future Work에서 정당화 |

API 접근 불가 시 정당화 문구:
- "Comparison with cloud-based SOTA models (GPT-4o, Claude) is deferred to future work. Our primary research question concerns the feasibility of **local, offline** inference; cloud API comparison, while informative, does not affect the core privacy-preserving deployment scenario."
- 이 프레이밍은 논문의 핵심 가치 제안(프라이버시 + 오프라인)과 일관됨

### 영향
전통 도구 비교(Track A)만으로도 비평의 가장 치명적인 부분("증거 없이 주장만")을 해소할 수 있다.

---

## ISS_003: 프롬프트 실험 단순함 — high

### 진단
Zero-shot 단일 프롬프트만으로 모델 성능을 단정하면 "프롬프트 탓인지 모델 탓인지" 분리 불가.

### 대응: Track A (실험 보강)

**3가지 프롬프트 전략 × 2 모델 × 9 파일 = 54건 감사를 실행한다.**

#### 프롬프트 설계

| 전략 | 설명 | 시스템 프롬프트 수정 방식 |
|:-----|:-----|:-------------------------|
| **P1: Zero-shot** (기존) | 현행 프롬프트 그대로 | 변경 없음 (이미 Run 03에서 완료) |
| **P2: Few-shot** | 취약/안전 예시 각 1개를 프롬프트에 포함 | 프롬프트 말미에 "Example 1 (Vulnerable): ... Example 2 (Safe): ..." 추가 |
| **P3: Chain-of-Thought** | 단계별 추론을 명시적으로 요구 | "Before concluding, reason step-by-step: (1) Identify all state-modifying operations, (2) Trace each value to PutState or response, (3) Determine if the value is deterministic across all endorsing peers, (4) Only flag as vulnerable if nondeterministic data reaches PutState." 추가 |

추가 실행량: P2 + P3 = 2 전략 × 2 모델 × 9 파일 = **36건**
예상 소요: Run 03 기준 평균 ~17s/건 → 약 36 × 17 = 612초 ≈ **10분**

#### 결과 활용

어떤 결과가 나오든 논문에 유리하다:

| Llama 결과 | 해석 | 논문 반영 |
|:-----------|:-----|:----------|
| Few-shot/CoT로 Llama 성능 대폭 개선 | 프롬프트 민감성 존재, 구조적 열위는 아님 | "Prompt engineering can partially compensate for architectural gaps" |
| Few-shot/CoT에도 Llama 성능 미개선 | 아키텍처 차이가 근본 원인 | "The specialist-generalist gap is robust across prompting strategies" |
| Qwen도 CoT로 FP 감소 | CoT가 범용적으로 유효 | "CoT benefits both models, but Qwen's baseline already near-optimal" |

논문 반영:
- Results에 신규 서브섹션 "Prompt Strategy Comparison" 추가
- Table 추가: Prompt × Model × (TPR, TNR, CPR, FPC)
- Discussion에서 프롬프트 민감성 분석 추가

#### 스크립트 수정
02_run_audit.py의 `SYSTEM_PROMPT`를 파라미터화하여 3가지 프롬프트를 순차 실행하도록 수정. 또는 프롬프트별로 3회 별도 실행.

### 영향
"실험 설계의 게으름" 비판을 완전히 해소. 추가 GPU 시간 약 10분으로 논문 완성도가 크게 향상됨.

---

## ISS_004: 데이터 오염/암기 가능성 — high

### 진단
Qwen의 100% 성능이 진정한 추론인지 학습 데이터 암기인지 구분할 수 없다.

### 대응: Track A (실험 보강) + Track B (논의 강화)

#### A. 난독화(Obfuscation) 실험

취약/안전 파일 9개의 **변수명, 함수명, 구조체명을 무의미한 이름으로 치환**한 버전을 생성하여 재감사한다.

난독화 규칙:
- 함수명: `InitLedger` → `FuncA`, `TransferAsset` → `FuncB`
- 변수명: `ownerKey` → `v1`, `amount` → `v2`
- 구조체명: `AssetContract` → `StructX`
- **HLF API 호출은 보존**: `stub.PutState`, `stub.GetState` 등은 변경하지 않음 (이것까지 바꾸면 문맥 자체가 파괴됨)
- 주석 전부 제거

추가 실행량: 2 모델 × 9 파일 = **18건**
예상 소요: 약 18 × 17 = 306초 ≈ **5분**

결과 해석:

| 난독화 후 Qwen 결과 | 의미 |
|:---------------------|:-----|
| 여전히 정탐/우선순위 유지 | **추론 기반 탐지 증거** — 변수명이 아닌 API 호출 흐름 + 데이터 경로를 기반으로 판단 |
| 성능 저하 | **암기 의존 증거** — 하지만 "어느 정도" 저하인지에 따라 해석 상이 |

논문 반영:
- Results에 신규 서브섹션 "Robustness to Code Obfuscation" 추가
- Table 추가: Original vs Obfuscated × Model × (TPR, TNR, CPR)

#### B. Discussion에서 Data Contamination 명시적 논의

- 신규 서브섹션 "Data Contamination Considerations" 추가
- Qwen2.5-Coder의 학습 데이터에 HLF 문서 포함 가능성을 인정
- 난독화 실험 결과로 부분적 반박 또는 한계 인정
- "Even if partial memorization exists, the model's ability to correctly discriminate safe from unsafe patterns under obfuscated naming suggests a degree of structural reasoning beyond surface-level memorization."

### 영향
이 실험은 준비 5분 + GPU 5분으로 논문의 학술적 깊이를 크게 높임.

---

## ISS_005: "Semantic Analysis" 용어 과대포장 — medium

### 진단
LLM은 CFG/DFG를 구성하지 않으므로 "Semantic Static Analysis"는 과장.

### 대응: Track B (리프레이밍)

#### 조치 항목

**1) 용어 교체**

| 현재 | 변경안 |
|:-----|:-------|
| "Semantic Static Analysis" (제목) | "LLM-Assisted Vulnerability Detection" |
| "semantic static analysis" (본문) | "context-aware vulnerability detection" |
| "semantic understanding" | "contextual awareness" |
| "semantic judgment" | "context-dependent judgment" |

수정 대상: Title, Abstract, Methodology (§III), Discussion (§V), Conclusion (§VI)

**2) 용어 정의 추가 (Methodology §III)**

신규 단락:
> "We use the term *context-aware* to describe the model's ability to evaluate a code construct (e.g., `time.Now()`) differently depending on its downstream usage (e.g., `PutState` vs. `fmt.Printf`). This is distinct from traditional static analysis, which constructs explicit control-flow or data-flow graphs. The LLM operates on token-level patterns but, as we demonstrate, achieves functionally equivalent discrimination in our controlled test cases."

**3) Qualitative Evidence 추출 (Case Study)**

audit_log.csv에서 Qwen의 출력 텍스트 중 **추론 경로가 명시적으로 보이는 사례**를 추출:
- safe_01에서 "No vulnerabilities detected" 판정 근거
- vuln_01에서 `time.Now()` → `PutState` 경로를 명시한 부분
- 이를 Figure 또는 인용 블록으로 제시

### 영향
용어 교체는 순수 텍스트 작업. Case Study 추출은 기존 audit_log.csv에서 발췌만 하면 됨.

---

## 실행 우선순위 및 의존성

```
[Phase 1] 독립 실행 가능 — 즉시 착수
├── ISS_002A: go vet + staticcheck + semgrep 실행 (전통 도구 비교)
├── ISS_004A: 난독화 파일 9개 생성
└── ISS_005:  용어 교체 + Case Study 추출 (audit_log.csv 발췌)

[Phase 2] Phase 1 결과 필요
├── ISS_004A: 난독화 파일 감사 실행 (18건, ~5분)
├── ISS_003:  Few-shot + CoT 프롬프트 설계 → 감사 실행 (36건, ~10분)
└── ISS_002A: 전통 도구 결과 Table 작성

[Phase 3] 모든 실험 완료 후
├── ISS_001:  Title/Abstract/Conclusion 리프레이밍
├── ISS_002B: SOTA 비교 여부 결정 → Threats to Validity 반영
└── 전체:     Results/Discussion 섹션 재편성
```

## 추가 실험 총량

| 실험 | 감사 건수 | 예상 GPU 시간 |
|:-----|:---------|:-------------|
| ISS_002A: 전통 도구 3종 × 9 파일 | 27건 (GPU 불필요) | 0분 (CPU) |
| ISS_003: 2 프롬프트 × 2 모델 × 9 파일 | 36건 | ~10분 |
| ISS_004A: 2 모델 × 9 난독화 파일 | 18건 | ~5분 |
| **합계** | **81건** | **~15분 GPU + α** |

## 논문 구조 변경 요약 (Track A+B 기준, v2까지)

| 섹션 | 변경 내용 |
|:-----|:----------|
| **Title** | "Semantic Static Analysis" → "LLM-Assisted Vulnerability Detection" + "Feasibility Study" |
| **Abstract** | 주장 톤 완화, "preliminary evidence" 프레이밍 |
| **§I Introduction** | "largely blind" 주장에 실증 참조 추가 |
| **§III Methodology** | N=9 정당성 서브섹션, 프롬프트 전략 3종 명세, "context-aware" 정의 |
| **§IV Results** | 3개 서브섹션 신설: Baseline Comparison, Prompt Strategy, Obfuscation Robustness |
| **§V Discussion** | Data Contamination 서브섹션, Case Study 추가, 용어 정리 |
| **§V Threats** | 대폭 확장 (N=9 정당화, SOTA 미비교 정당화, 단일 하드웨어) |
| **§VI Conclusion** | 톤 완화, Future Work 강화 |

---

## Track C: 전략적 방향 전환 (FB_260209_1700_gemini_strategy 기반)

**추가일**: 2026-02-09T17:30:00
**트리거**: Gemini와의 심층 전략 논의 + 딥리서치 결과 (GoLiSA 발견)
**적용 대상**: ISS_010, ISS_011, ISS_012, ISS_013, ISS_014 + ISS_006~009 방어 전략 강화

### 핵심 전략 변경점

#### C-1. GoLiSA 인용 및 차별화 (ISS_010)
- **딥리서치 결과**: GoLiSA Artifact (ECOOP 2023, ~651개 HLF 체인코드) 존재 확인
- **조치**: Related Work에 GoLiSA 인용. 수집형(Mined) vs 설계형(Engineered) 차별화
- **논리**: "GoLiSA는 정적 분석 도구의 재현율(Recall) 측정에 유리. 우리 15개는 LLM의 정밀도(Precision)와 추론 능력 검증에 특화. 상호 보완 관계."

#### C-2. Privacy Paradox 신규 논거 (ISS_011)
- **논리 흐름**: HLF = 기업용 → 코드 기밀 → 클라우드 LLM 사용 불가 → 로컬 sLM 필수 → Qwen이 보안+성능 양립
- **삽입 위치**: Introduction (동기), Discussion (비용-프라이버시 분석), Conclusion (실용적 함의)

#### C-3. Core Thesis 3대 기둥 수립 (ISS_014)
1. **Silent Failure**: HLF 버그는 Crash가 아닌 합의 불일치 → 기존 도구 탐지 0%
2. **Context over Keywords**: 키워드가 아닌 문맥 파악이 핵심 → Qwen(TNR 100%) vs Llama(TNR 17%)
3. **Privacy Paradox**: 기업 환경에서 로컬 sLM이 유일한 해결책

#### C-4. 논문 톤 전면 수정 (ISS_014)
| 현재 | 변경 |
|:-----|:-----|
| "Superior Performance" | "Competitive Accuracy" / "Feasibility Study" |
| "Outperforms" | "demonstrates comparable reasoning" |
| "Standard Benchmark" | "Curated Micro-benchmark" |
| 100% 강조 | "통제된 환경에서 의도된 패턴 식별 성공" |

#### C-5. 딥리서치 참고문헌 확보 (ISS_012) — 완료
- 4개 카테고리: (1) HLF XOV 아키텍처, (2) GoLiSA/VulFinder 원문, (3) LLM vs 정적 분석, (4) 코드 특화 sLM 성능
- **검증 완료**: 7개 논문 arXiv/PDF 링크 HTTP 200 OK 확인. VulFinder(IEEE TSE 2025)만 paywalled.
- **팩트 체크 경고**: 딥리서치 보고서의 Zenodo DOI (10.5281/zenodo.7896323)는 부정확 — 관련 없는 논문으로 리다이렉트. GoLiSA artifact는 DARTS(Dagstuhl)에서 OVA VM으로 배포.
- **"651개" 미검증**: 원문 PDF에서 직접 확인 필요

#### C-6. GoLiSA 외부 검증 실험 (ISS_013) — 진행 중
- **데이터셋 확보**: DARTS OVA (4.99GB) 다운로드 진행 중. OVA 내부에서 .go 파일 추출 필요 (07_extract_golisa.py 준비 완료)
- **스크립트 호환성** (실측 분석 완료):
  - 02_run_audit_v3.py: `--dataset-dir` CLI 지원, 코드 수정 불필요. 추정 ~11시간+ (651 x 2 models x 3 prompts)
  - 05_run_traditional_tools.py: 경로 하드코딩 + `vuln_`/`safe_` prefix 의존성. 코드 수정 필요. 추정 ~1.5-2시간
- RQ2: GoLiSA로 로컬 모델(Qwen, Llama) + Semgrep 전수 조사 (비용 0원)
- RQ3: 클라우드 모델은 30~50개 샘플링 (선택적)
- 효과: N=15 한계 비판(ISS_006) 해소, 통계적 검정 가능, 100% 역설(ISS_007) 해소

### Track C 논문 구조 추가 변경

| 섹션 | 추가 변경 내용 |
|:-----|:--------------|
| **§I Introduction** | Privacy Paradox 동기 추가, Core Thesis 3대 기둥 기반 재구성 |
| **§II Related Work** | GoLiSA (ECOOP 2023) 인용 + 차별화, Solidity 논문과 거리두기 |
| **§III Methodology** | "Curated Micro-benchmark" 정의, RQ 구조화 (RQ1 Internal + RQ2 External) |
| **§IV Results** | GoLiSA 외부 검증 결과 (RQ2, 실행 시) |
| **§V Discussion** | 질적 Error Analysis 심화, Privacy+Performance 트레이드오프 분석 |
| **§V Threats** | N=15 + 통계 미수행 근거 명시, GoLiSA로 보완 언급 |
| **meta** | 톤 다운 (Title/Abstract), 참고문헌 갱신 |
