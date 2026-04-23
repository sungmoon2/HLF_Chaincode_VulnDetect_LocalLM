# GoLiSA Main Experiment — 실행 체크리스트
> Version: 2.0 | Created: 2026-04-22 | Last Updated: 2026-04-23 | D1-D5 확정: 2026-04-22
> 근거: GPT 5.4 Pro 자문 6세션 (rubric 3 + experiment design 3) + Claude 비판적 교차 분석
> 마감: 2026-05-22 (AANN full paper)
> Phase 0-1: 이전 세션(S260422-rubric-labeling)에서 완료, GOLISA_TAXONOMY_RUBRIC_v2.md FROZEN

---

## Phase 0: Rubric v2.0 수정 + 재동결

### 0-1. Taxonomy 명칭/구조
- [x] §1 제목을 "GoLiSA-Derived, Scope-Restricted, Fabric-Focused Subset Taxonomy"로 변경
- [x] §1 Corpus naming을 "GoLiSA-derived, scope-restricted, Fabric-focused subset"으로 변경
- [x] §2 taxonomy를 5+1 구조로 변경: C1-C4,C6 = core nondeterminism / C5 = auxiliary resource-hygiene label
- [x] C5 ITERATOR_LEAK를 binary V/S 판정에서 제외. auxiliary label로만 태깅

### 0-2. Class 정의 수정
- [x] C2 GOROUTINE: 정의에 "goroutine-only subset of GoLiSA concurrency family" 명시
- [x] C3 MAP_ITERATION: source pattern을 "any `for...range` over a map-typed expression"으로 일반화
- [x] C3: json.Marshal(map) SAFE note 추가 ("encoding/json v1 standard library, deterministic key sorting")
- [x] C3: 제3자 serializer는 "documented deterministic ordering이 있을 때만 SAFE" 명시
- [x] C4: 명칭을 **NON_REVALIDATED_QUERY**로 변경 (legacy alias: PHANTOM_READ). 정의: "all query APIs whose results are not re-validated at commit time"
- [x] C4: GetHistoryForKey를 §5 "NOT PHANTOM_READ"에서 제거 → C4 범위로 이동
- [x] C4: "Fabric-specific extension, not a GoLiSA source class" 표기
- [x] C4: 스크립트/내부 문서에서는 PHANTOM_READ 맵핑 유지 (downstream 호환)
- [x] C6: 정의는 GoLiSA의 "package-level var"를 유지 (축소하지 않음). logger 제외를 **benchmark-specific precision heuristic**으로 별도 분리 표기
- [x] C6: heuristic 문구 — "As a benchmark-specific precision heuristic, declarations whose sole role is logging infrastructure are excluded. This heuristic is a documented deviation from GoLiSA, not part of the core family definition."

### 0-3. OUT OF SCOPE 확장
- [x] crypto/rand 추가
- [x] system/environment APIs (os.Getenv 등) 추가
- [x] file/network/I/O APIs 추가
- [x] external process APIs 추가
- [x] channel receives (<-) 추가 + "GoLiSA's full concurrency family includes both go and <- patterns" 주석
- [x] InvokeChaincode를 OOS에 각주로 명시 (single-file scope 외)

### 0-4. Sink 완전성
- [x] §3.1에 SetPrivateDataValidationParameter 추가
- [x] §3.2에 contractapi return value/error 명시
- [x] §3.3을 "Implicit-Flow Rule"로 rename + rewrite
- [x] §3.4에 전체 sink API의 shim/contractapi equivalence 열거

### 0-5. Decision Tree
- [x] Step 1을 "recognizable HLF transaction entrypoint"로 변경
- [x] contractapi entrypoint 추가: TransactionContextInterface / *TransactionContext / custom context
- [x] "Do NOT require literal Init/Invoke in contractapi files" 명시
- [x] Rationale 문구 추가

### 0-6. Ambiguous Cases + Scope
- [x] L4 defer 문구 수정: "SAFE with respect to normal returns and panics. Flag only if no reachable Close() exists."
- [x] §4.x Scope Rule 신설: "strictly intra-file evidence only"
- [x] GetTxTimestamp() → SAFE for C1 추가
- [x] init()/package initializer 규칙 추가
- [x] _test.go → EXCLUDE 추가
- [x] build-tag → benchmark build context 기준 추가
- [x] "plausible execution path" → "concrete intra-file source→sink evidence" 수정
- [x] global var heuristic: "deterministic init + no later mutation → SAFE" 조건부로 수정
- [x] time.Now() co-location rule 추가: "time.Now() appears in file but no concrete intra-file influence on sink → SAFE; mere co-location is insufficient"

### 0-7. 재동결
- [x] 수정 완료된 rubric을 GOLISA_TAXONOMY_RUBRIC_v2.md로 저장
- [x] v2.0 FROZEN 마킹
- [x] (권장, 비필수) pilot double-check: 혼합 패턴 GoLiSA 파일 3-5개에 rubric v2.0 적용 → 판정 일관성 확인

---

## Phase 1: GPT 라벨링 준비

### 1-1. 라벨링 프롬프트 설계
- [x] rubric v2.0을 그대로 포함하는 라벨링 프롬프트 작성
- [x] 입력: **stripped code** (comment-stripped, no filename/path)
- [x] 출력 형식: `VERDICT | PRIMARY_CLASS | SECONDARY_CLASS | EVIDENCE_LINES | SHORT_RATIONALE`
- [x] CLASSES = NONE이면 SAFE, CLASSES ≠ NONE이면 VULNERABLE (LABEL 필드 불사용)
- [x] C5는 auxiliary label로 별도 태깅 (binary V/S에 불포함)

### 1-2. 라벨링 뷰 규칙 (annotation-view mismatch 방지)
- [x] **Primary view: stripped code** — 모든 판정은 stripped 코드 기준
- [x] Raw source(주석 포함)는 syntax 확인/provenance 확인용 secondary reference로만 허용
- [x] **Comments를 판정 근거로 사용 금지** — 명시적으로 프롬프트에 포함

### 1-3. Dev set 격리
- [x] sanity check에 사용한 10개 파일 목록 확정
- [x] 해당 10개 파일을 final 60 / reserve / D2에서 **완전 제외**
- [x] 격리 목록을 DEV_EXCLUSION_LIST.md에 기록

---

## Phase 2: GPT 1차 라벨링 실행

> **[변경] Full corpus 전환 (2026-04-22~23)**
> 원래 계획: positive 101 + safe 112 + hard neg 13 = 226개 대상
> 변경 후: GoLiSA 전수 621개 (non-dev) + Running_Examples 1개 = 622개
> 변경 근거: N=15 리뷰어 지적 대응, "full-corpus evaluation" 주장 확보, mining miss 탐지

### 2-1. ~~Positive candidates (101개)~~ → Full Corpus (622개)
- [x] stripped code 기준으로 Opus 4.5 via Vertex AI 라벨링 (622/622)
- [x] C5 auxiliary label 별도 태깅 (C5=YES: 10건)
- [x] evidence line spans 기록 (전수 기입, 빈 값 0건)
- [x] 결과를 per_file/ JSON + summary.csv로 저장
- [x] hard negative 13개 → 12개 active 라벨링 완료 (스크립트 버그 수정 후)

### 2-2. 스크립트 버그 수정 (세션 중 발견)
- [x] `load_candidates()`: hard neg pool 미포함 → 추가 루프로 수정
- [x] resume 로직: repo 단위 → (repo/filename) 단위로 변경
- [x] `--full-corpus` 모드 추가 (Benchmark/ 전체 스캔)

### 2-3. 1차 라벨링 품질 체크
- [x] CLASSES/VERDICT conflict: **0건**
- [x] parse failure rate: **0건**
- [x] invalid verdict 정규화: 2건 `SAFE**` → `SAFE`, 1건 `** NO` → `NO`
- [x] family 분포: C1:23, C6:4, C3:3, C4:2, C2:1 (V=46 기준, 재분류 전)

### 2-4. 코드 레벨 검증 (추가 — 체크리스트 원본에 없음)
- [x] VULNERABLE 46건 전수 코드 리뷰 (Agent + human 2nd pass)
- [x] 13건 FALSE POSITIVE 발견 → SAFE 재분류
- [x] positive→SAFE 70건 전수 코드 리뷰 (4 Agent + human 2nd pass)
- [x] FALSE NEGATIVE: **0건** 확인
- [x] 40건 HIGH RISK same-function C4+Write 분석 → 추가 FN 0건
- [x] 검증 결과 verification/ 디렉토리에 구조화 저장
- [x] 최종: 33 TRUE V + 585 TRUE S + 4 EXCLUDE

---

## Phase 3: 2차 검증 (Second Annotator)

> **[변경] full corpus 464개 전수 dual review**
> 원래: final 60개 전수 dual review
> 변경: benchmark 464개 전수 dual review (비용 ~$15, 충분히 감당 가능)
>
> **[변경] 2nd annotator 모델 결정 (2026-04-23)**
> 초기 시도: Haiku 4.5 → frontier 아님, 검증력 부족 판단
> 확정: **Opus 4.5 + 다른 프롬프트** (같은 모델이지만 프롬프트 상이 → 유효한 2nd annotator)
> Haiku 결과: second_haiku_supplementary/에 보조 참고용 보존
>
> 프롬프트 차이:
> - 1차: "security analyst performing file-level labeling" + 테이블 taxonomy + SAFE exceptions 14개
> - 2차: "independent reviewer auditing" + 불릿 taxonomy + Decision Process 4단계

### 3-1. 2차 검증 설계
- [x] 2차 검증자: **Opus 4.5 + 다른 프롬프트** + **blinded** (1차 라벨 비공개)
- [x] "같은 GPT 같은 prompt 재실행 ≠ second annotator" 원칙 준수 (프롬프트 상이 확인)
- [x] 2차 프롬프트에도 rubric v2.0 포함, stripped code 입력
- [x] 스크립트: 25_run_second_annotation.py

### 3-2. 2차 검증 범위
- [x] **benchmark 464개 전수** dual review (진행 중)
- [ ] D2 17개 검증 (Phase 6 후 별도)

### 3-3. Adjudication
- [x] 1차 vs 2차 disagreement 목록 생성
- [x] **인간 저자가 disagreement 케이스만 adjudication**
- [x] adjudication log 기록
- [x] raw agreement 계산
- [x] Cohen's κ 계산

---

## Phase 4: Benchmark Freeze

> **[변경] Full corpus → sampling 불필요, dedup 후 전수 사용**
> 원래: 30V+30S=60 stratified sampling
> 변경: 32V+432S=464 (content-hash dedup 후 전수)
> 변경 근거: full corpus evaluation으로 cherry-pick 의심 제거, N=464 >> N=60

### 4-1. Eligibility Filter
- [x] self-contained: intra-file evidence로 판정 (rubric scope rule 적용)
- [x] dedup: **content-hash (SHA-256) dedup** → 618→464 (154건 중복 제거)
- [x] dev exclusion: 10 repos (31 files) 제외 확인
- [x] token-length check: max 12,310 tokens, **truncation 0건** (n_ctx=16384 이내)
- [x] EXCLUDE 4건 제외 (비체인코드/불완전)

### 4-2. Family Distribution (cap 미적용)
> **[변경] C1 cap 미적용 — 전수 corpus에서 인위적 제거는 부적절**
> Threats에 "C1 dominance (69%)" 명시로 대체

- [x] family 분포 보고: C1:22, C6:4, C3:3, C4:2, C2:1
- [x] reportable: C1 inferential (n=22≥5), C6/C3 descriptive (n=3~4), C4/C2 anecdotal (n≤2)
- [ ] Threats에 C1 dominance + family imbalance 명시 (Phase 9)

### 4-3. Safe Set 구성
- [x] **전수 편입** (sampling 불필요)
- [x] hard negatives: 라벨링 완료, 11 SAFE + 1 V(→FP→SAFE 재분류) = 12건 포함
- [x] TNR_easy / TNR_hard 분리 보고 설계 유지 (Phase 8에서 실행)

### 4-4. ~~Stratified Random Sampling~~ → 불필요
- [x] full corpus이므로 sampling 불필요 — 전수 사용

### 4-5. Freeze 확정
- [x] BENCHMARK_FREEZE.json 저장 (06_addon_validation/benchmark/)
- [x] BENCHMARK_FREEZE.md 작성 (인간 가독용) — 2026-04-23
- [x] freeze 시점 기록: 2026-04-23
- [x] freeze 이후 benchmark 변경 금지 선언

---

## Phase 5: Inference Contract Freeze

### 5-1. Frozen Manifest 작성
- [x] exact prompt template (전문) — PROMPT_ZERO_SHOT (generic security audit)
- [x] parser 로직 — classify_response() keyword-based (safe indicators + vuln indicators ≥2)
- [x] scoring rule — prediction := classify_response(output), TP/FP/FN/TN vs benchmark
- [x] seed — temp=0.0 implicit / temp=0.1 seeds={1,2,3,4,5}
- [x] quantization: Q4_K_M
- [x] model file SHA256: 509287f78cb4d4cf6b3843734733b914b2c158e43e22a7f4bf5e963800894d3c
- [x] llama-cpp-python: 0.3.16
- [x] n_ctx = 16384
- [x] n_gpu_layers = -1 (all GPU)
- [x] Semgrep: 1.151.0 OSS, rules/hlf_consensus.yml
- [x] Semgrep file-level reduction: any frozen rule hit → VULNERABLE
- [x] **per-file fresh session** 명시
- [x] INFERENCE_CONTRACT.md 저장 (06_addon_validation/benchmark/)

### 5-2. Freeze 확인
- [x] manifest 작성 완료 (2026-04-23)
- [ ] manifest 파일 commit hash 기록 (git commit 후)

---

## Phase 6: Main Experiment 실행

> **[변경] benchmark 464개 전수 실행 (원래 60개)**
> 스크립트: 26_run_main_experiment.py
> INFERENCE_CONTRACT.md의 frozen manifest에 따라 구현

### 6-1. Qwen Deterministic Run
- [x] temp=0.0 (INFERENCE_CONTRACT 준수)
- [x] benchmark 464개 전수 실행 ~~(원래 60개)~~
- [x] per-file fresh session (llama_cpp per-call, carry-over 없음)
- [x] truncation 0건 확인 (max 12,310 tokens < n_ctx 16,384)
- [x] 결과 CSV + per_file JSON 저장

### 6-2. Semgrep Run
- [x] frozen rule pack (rules/hlf_consensus.yml) 사용
- [x] 동일 benchmark 파일 사용 (stripped input)
- [x] file-level reduction: any rule hit → VULNERABLE
- [x] 결과 CSV 저장 (Qwen과 동일 CSV에 병합)

### 6-3. Runtime 기록
- [x] 평균 초/파일
- [x] 총 소요시간
- [ ] peak VRAM (nvidia-smi)
- [x] meta.json 저장

### 6-4. D2 Hygiene Rerun
- [ ] D2 17개에 동일 프로토콜 적용
- [x] 결과 CSV 저장

---

## Phase 7: Robustness Run

- [ ] Qwen temp=0.1 × 5-run (seed={1,2,3,4,5})
- [ ] per-file stability 계산
- [ ] majority-vote 결과는 **stability 설명용**, main score로 올리지 않음
- [ ] optional: Llama 1-row (공간 남을 때만)
- [ ] optional: deterministic rerun hash-match (temp=0.0 2회 실행 → output hash 비교)

---

## Phase 8: Analysis

### 8-1. 필수 통계
- [ ] confusion matrix (raw counts): Qwen 2×2, Semgrep 2×2
- [ ] TPR / TNR / BA / Precision / F1 (점추정)
- [ ] **95% CI**: TPR/TNR = Wilson CI, BA/F1 = bootstrap CI
- [ ] **exact McNemar test**: Qwen vs Semgrep (전체 60개 paired comparison)
- [ ] **hard-negative breakdown**: TNR_easy / TNR_hard 분리 보고

### 8-2. Complementarity 분석
- [ ] Qwen-only TP / Semgrep-only TP / both TP 표
- [ ] Qwen-only FP / Semgrep-only FP 표
- [ ] union = "OR-ensemble", intersection = "agreement subset" 표현 사용
- [ ] **"upper bound" 표현 금지**

### 8-3. Family-level 분석
- [ ] family별 TPR (n≥3인 family만)
- [ ] n≥5: quantitative comparison 허용
- [ ] n=3~4: **descriptive/exploratory only** 한정 명시
- [ ] family imbalance를 Threats에 명시

### 8-4. Error Analysis
- [ ] FP/FN을 family별 1-2문장 분석
- [ ] representative false positive / false negative 예시

### 8-5. Agreement 보고
- [ ] 1차 vs 2차 annotator raw agreement
- [ ] Cohen's κ (가능하면 보고 — 결과에 따라 adjudication 절차 강화 여부 판단)
- [ ] adjudication 건수

### 8-6. 강력 권장 분석
- [ ] token-length stats: max / median / truncation 0건 statement
- [ ] regex-negative audit: mining에 걸리지 않은 파일 중 20-30개 blind 검토 → miss-rate 보고 또는 Limitations에 "regex-based mining may miss indirect patterns" 명시
- [ ] localization audit: positive 30개 EVIDENCE hit/miss
- [ ] per-file stability (robustness에서)

---

## Phase 9: 논문 작성 (v54)

### 9-1. 구성: C-lite
- [ ] **Results 중심**: GoLiSA main N=60
- [ ] **보조 결과**: D2 holdout
- [ ] **D1/P0**: Discussion 1문단 diagnostic ("Initial pilot discarded after leakage diagnosis")
- [ ] D1 수치표 본문에서 삭제

### 9-2. 프레이밍
- [ ] claim: **"privacy-preserving complementary triage signal under human-in-the-loop review"**
- [ ] N=60: "curated, deduplicated public benchmark" (일반화 주장 금지)
- [ ] D2: "auxiliary disjoint holdout" (independent validation 아님)
- [ ] privacy: **cloud LLM 대비로 한정** (Semgrep도 local tool)
- [ ] Semgrep 비교: "practical utility comparison on identical file-level task"

### 9-3. GoLiSA 수치 화해
- [ ] "We extracted 657 individual .go files from GoLiSA's public artifact; the original paper reports 651 chaincode-level entries across 954 repositories, reflecting a different unit of analysis."

### 9-4. GPT Labeling 공시 (AANN 규정)
- [ ] Methodology 또는 Acknowledgement에 tool name, version, functionality, specific application 명시
- [ ] "Labels were assigned by GPT-assisted first pass, verified by independent blinded second annotation, with human adjudication of disagreements."

### 9-5. Threats to Validity (필수 포함)
- [ ] public GitHub pretraining contamination
- [ ] balanced 30/30 ≠ real-world prevalence
- [ ] file-level target (line-level 미지원)
- [ ] family imbalance
- [ ] GPT-assisted annotation
- [ ] single prompt / single main model
- [ ] mining heuristic (regex) spectrum bias
- [ ] D2 tiny positive count (V=4)
- [ ] quantized local inference / backend specificity
- [ ] sample/tutorial memorization (public GitHub tutorial/sample chaincode와 pretraining data 중첩 가능성)

### 9-6. 추가 포함
- [ ] runtime/practicality table (초/파일, VRAM)
- [ ] artifact release statement
- [ ] 2025 논문 인용 (실존 검증 후: NDSS, DSN, ISSTA, MSR)
- [ ] C3 json.Marshal note에 "encoding/json v1 (Go standard library)" 버전 고정 문구

---

## 진행 상태 요약

| Phase | 상태 | 완료일 | 비고 |
|:------|:-----|:------|:-----|
| Phase 0: Rubric v2.0 | **완료** | 2026-04-22 | v2.0 FROZEN, 감사 통과 |
| Phase 1: 라벨링 준비 | **완료** | 2026-04-22 | 프롬프트 v2.0 + DEV_EXCLUSION_LIST |
| Phase 2: GPT 1차 라벨링 | **완료** | 2026-04-23 | full corpus 622개, 검증 완료 (33V+585S+4EX) |
| Phase 4: Benchmark Freeze | **FROZEN** | 2026-04-23 | 32V+432S=464, FREEZE.json/md |
| Phase 5: Inference Contract | **FROZEN** | 2026-04-23 | INFERENCE_CONTRACT.md |
| Phase 3: 2차 검증 | **진행 중** | | Opus 4.5 + 다른 프롬프트, 464개 전수 |
| Phase 6: Main Experiment | **진행 중** | | Qwen+Semgrep 464개, ~51/464 |
| Phase 7: Robustness | 미시작 | | |
| Phase 8: Analysis | 미시작 | | |
| Phase 9: 논문 v54 | 미시작 | | |

### 주요 변경 이력
| 일시 | 변경 | 근거 |
|:-----|:-----|:-----|
| 2026-04-22 | Phase 2 범위 확대: mining candidates → full corpus | N=15 리뷰어 지적, cherry-pick 방지 |
| 2026-04-23 | Phase 4 sampling → 전수: 30V+30S → 32V+432S | full corpus이므로 sampling 불필요 |
| 2026-04-23 | C1 cap 미적용 | 전수 corpus에서 인위적 제거 부적절, Threats 명시로 대체 |
| 2026-04-23 | Phase 3 모델: Haiku→Opus+다른프롬프트 | Haiku는 frontier 아님, 검증력 부족. Opus+다른프롬프트가 유효한 2nd annotator |
| 2026-04-23 | Phase 5 scoring: CLASSES→keyword parser | Qwen 7B에 구조화 출력 강제 시 성능 저하. 기존 v51과 일관성 유지 |
| 2026-04-23 | Phase 6 benchmark: 60→464 | full corpus 전수 실험으로 변경 |
