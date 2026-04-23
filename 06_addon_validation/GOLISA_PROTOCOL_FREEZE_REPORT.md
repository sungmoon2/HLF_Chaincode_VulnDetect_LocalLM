# GoLiSA-Derived Benchmark: Protocol Freeze Report
> 생성일: 2026-04-22 05:20 KST
> Status: **ALL PRE-LABELING STEPS COMPLETE — READY FOR GPT LABELING**

---

## 1. 완료 단계 요약

| Step | 작업 | 결과 | 산출물 |
|:-----|:-----|:-----|:-------|
| **Step 0** | Taxonomy/rubric freeze | 6-class 확정, rubric 문서화 | `GOLISA_TAXONOMY_RUBRIC.md` |
| **Step 2** | Go comment stripper 구현 | 빌드 성공 | `scripts/strip_go_comments.go` → `strip_go_comments.exe` |
| **Step 3** | GoLiSA 657개 parsability dry run | **657/657 (100%) 성공, 실패 0** | `golisa_mining/parsability_dryrun.json` |
| **Step 4** | Candidate mining (6-family grep/regex) | 146 positive candidates + 184 safe + 14 hard negatives (one-per-repo) | `golisa_mining/` 6개 JSON |
| **Step 1** | Prompt dev sanity check (10 files × 2 runs) | **포맷 100%, determinism 100%** | `golisa_mining/dev_sanity_check_260422_0515.json` |

---

## 2. 실측 결과 (사실 기반, 추정 없음)

### 2.1 Parsability (실측)
- 전체 파일: 657개 .go
- go/parser 파싱 성공: **657/657 (100.0%)**
- 파싱 실패: **0개**
- comment stripping 후 재파싱 검증: dry run 모드로 전수 확인

### 2.2 Candidate Mining (실측)

**전체 분포:**
| 항목 | 실측 값 |
|:-----|:--------|
| Total .go files | 657 |
| Chaincode files (shim/contractapi indicator 有) | 656 |
| Non-chaincode files | 1 |
| Files with any family hit | 266 |
| Files with consensus-critical sink | 570 |
| Chaincode + hit + sink | 251 |
| Unique repos | 326 |
| Repos with hits | 157 |
| **One-per-repo positive candidates** | **146** |
| Near-duplicate groups | 59 |
| Context budget 초과 (prompt+code > 65536 chars) | **0** |

**Family 분포 (chaincode only, 실측):**
| Family | 파일 수 |
|:-------|:--------|
| TIME_NOW | 33 |
| GOROUTINE | 3 |
| MAP_ITERATION | 10 |
| PHANTOM_READ | **111** |
| ITERATOR_LEAK | 218 |
| GLOBAL_MUTABLE_STATE | 34 |

**주의**: 위 수치는 grep/regex surface pattern match 수. 실제 source→sink 확인은 라벨링에서 수행.
특히 PHANTOM_READ 111건과 ITERATOR_LEAK 218건은 상당수가 라벨링 시 SAFE로 분류될 수 있음 (close 유무, sink 연결 등).

**Safe pool (one-per-repo):**
- Pure safe (chaincode, has sink, no family hit): **184개**
- Hard negatives (chaincode, has family hit, no sink): **14개**

**파일 크기 분포:**
| 통계 | Bytes | Lines |
|:-----|:------|:------|
| Min | 216 | 19 |
| Max | 42,808 | 1,234 |
| Mean | 8,264 | 255 |
| P90 | 24,910 | 640 |

### 2.3 Prompt Dev Sanity Check (실측)

**실험 설정:**
- Model: qwen2.5-coder-7b-instruct-q4_k_m.gguf
- n_ctx: 16384
- Temperature: 0.0
- Dev files: 10개 (positive 6 + safe 4)
- Runs per file: 2 (determinism 확인용)

**결과:**
| 항목 | 실측 |
|:-----|:-----|
| 포맷 안정성 (LABEL/CLASSES/EVIDENCE 3필드) | **20/20 (100%)** |
| Determinism (temp=0.0, 2회 동일 출력) | **10/10 (100%)** |
| LABEL-CLASSES 의미 충돌 | 4/20 (2파일 × 2runs) |
| 충돌 유형 | VULNERABLE + CLASSES=NONE → CLASSES 우선 규칙으로 SAFE 처리 |
| 최종 parse 가능률 (CLASSES 우선 적용 후) | **20/20 (100%)** |
| VULNERABLE 출력 | 12/20 (6파일) |
| SAFE 출력 | 8/20 (4파일, 충돌 해결 포함) |
| 평균 추론 시간 | ~1.0초/파일 |

**Parser 규칙 확인:**
- CLASSES 우선 규칙 작동 확인: LABEL=VULNERABLE + CLASSES=NONE → SAFE 처리 (2건)
- 허용 외 class 출현: 0건
- Extra lines: 0건

---

## 3. FROZEN Artifacts

### 3.1 확정 Prompt (FROZEN)
```
You are auditing a single Go source file from a Hyperledger Fabric chaincode project.

This is a CLOSED-SET classification task.
Consider ONLY the six consensus-critical nondeterminism classes listed below.
Ignore all other bug types, security issues, or code quality concerns.

A file is VULNERABLE only if at least one class below can influence:
- ledger writes (PutState, DelState, PutPrivateData, DelPrivateData), or
- the returned proposal response, or
- control flow that determines such writes or responses.

Treat shim-style calls (stub.PutState) and contractapi-style calls (ctx.GetStub().PutState) as equivalent sinks.

Do NOT report: access control, input validation, key management, logging-only uses, comments, filenames, test/dead code, constant globals, slice/array iteration, user-provided timestamps, or any issue outside the six classes.

Allowed classes:
1. TIME_NOW — time.Now/time.Since/time.Until influencing a sink
2. GOROUTINE — go statement or concurrent work influencing a sink
3. MAP_ITERATION — iteration over a Go map whose order influences a sink (NOT slices/arrays)
4. PHANTOM_READ — rich-query results (GetQueryResult, GetPrivateDataQueryResult, GetQueryResultWithPagination) used to decide writes (NOT GetStateByRange alone)
5. ITERATOR_LEAK — ledger/query iterator not closed on all paths
6. GLOBAL_MUTABLE_STATE — mutable package-level state influencing a sink

Return EXACTLY these three lines and nothing else:
LABEL: VULNERABLE | SAFE
CLASSES: NONE | <comma-separated from allowed set>
EVIDENCE: <one line, max 30 words>

Go source:
```go
{CODE}
```
```

### 3.2 확정 Parser 규칙 (FROZEN)
1. LABEL line → binary decision
2. CLASSES line → allowed set 검증
3. LABEL-CLASSES 충돌 → **CLASSES 우선**
   - VULNERABLE + CLASSES=NONE → SAFE
   - SAFE + CLASSES 有 → VULNERABLE
4. 허용 외 class → 무시
5. parse 완전 실패 → 1회 재시도 후 parse_error

### 3.3 확정 Taxonomy (FROZEN)
6개 class: TIME_NOW, GOROUTINE, MAP_ITERATION, PHANTOM_READ, ITERATOR_LEAK, GLOBAL_MUTABLE_STATE
out-of-scope: math/rand, access control, input validation, etc.

### 3.4 확정 Comment Stripper (FROZEN)
- 방식: go/parser (ParseComments 미사용) + format.Node 재출력
- 실행 파일: scripts/strip_go_comments.exe
- 파싱 실패 시: 교체 (regex fallback 금지)
- 재파싱 검증: 필수

---

## 4. 산출 파일 목록

| 파일 | 위치 | 내용 |
|:-----|:-----|:-----|
| `GOLISA_TAXONOMY_RUBRIC.md` | `06_addon_validation/` | Taxonomy + labeling rubric (FROZEN) |
| `GOLISA_PROTOCOL_FREEZE_REPORT.md` | `06_addon_validation/` | 이 문서 |
| `strip_go_comments.go` | `scripts/` | Go comment stripper 소스 |
| `strip_go_comments.exe` | `scripts/` | Go comment stripper 바이너리 |
| `22_mine_golisa_candidates.py` | `scripts/` | Candidate mining script (22번째) |
| `23_prompt_dev_sanity_check.py` | `scripts/` | Prompt dev sanity check script (23번째) |
| `parsability_dryrun.json` | `06_addon_validation/golisa_mining/` | 657개 파싱 결과 |
| `mining_full_260422_0513.json` | `06_addon_validation/golisa_mining/` | 전체 mining 결과 (657 records) |
| `mining_summary_260422_0513.json` | `06_addon_validation/golisa_mining/` | Mining 통계 요약 |
| `labeling_candidates_260422_0513.json` | `06_addon_validation/golisa_mining/` | Positive 후보 146개 (one-per-repo) |
| `safe_candidates_260422_0513.json` | `06_addon_validation/golisa_mining/` | Safe 후보 184개 (one-per-repo) |
| `hard_negative_candidates_260422_0513.json` | `06_addon_validation/golisa_mining/` | Hard negative 14개 (one-per-repo) |
| `dev_files.json` | `06_addon_validation/golisa_mining/` | Dev sanity check 파일 목록 (10개) |
| `dev_sanity_check_260422_0515.json` | `06_addon_validation/golisa_mining/` | Dev sanity check 결과 (20 runs) |

---

## 5. 다음 단계: GPT 라벨링

### 5.1 라벨링 대상
- **Positive 후보**: 146개 (labeling_candidates_260422_0513.json)
- **Safe 후보 + hard negatives**: 184 + 14 = 198개 중 safe 선정용으로 일부 라벨링
- **라벨링 방식**: GPT (source→sink rubric 기반)
- **라벨링 기준 파일**: raw source (comment-stripped 아님)

### 5.2 라벨링 프로토콜
1. 146 positive 후보 전수 GPT 라벨링 (V / S / EXCLUDE)
2. Primary/secondary family 기록
3. Safe 후보에서 matched hard negatives 선정
4. 30V + 30S 목표 (fallback: 24V+24S)
5. Family-capped: TIME_NOW ≤ 12
6. One-file-per-repo 유지
7. Reserve 10개 확보
8. Second-annotator audit: positive 전수 + safe 20-25%

### 5.3 라벨링 후 프로세스
1. Benchmark freeze (파일 목록 + gold labels)
2. Comment stripping (model input 생성)
3. Main experiment (Qwen temp=0.0 + Semgrep)
4. Robustness (Qwen temp=0.1 × 5-run)
5. D2 hygiene rerun
6. Analysis + v54 작성
