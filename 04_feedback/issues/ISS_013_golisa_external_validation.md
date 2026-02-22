# ISS_013: GoLiSA 외부 검증 실험

**Status**: open
**Priority**: high
**Source Feedback**: FB_260209_1700_gemini_strategy
**Target Sections**: III, IV, V
**Created**: 2026-02-09

## 문제 설명
GoLiSA 데이터셋(~651개)의 존재가 확인되었으므로, 이를 활용한 외부 검증(External Validation) 실험을 수행하여 N=15 한계를 보완하고 일반화 가능성(Generalization)을 증명해야 한다.

## 제안된 실험 설계 (Gemini 논의에서 합의)

### RQ 구조화
| 실험 세트 | 데이터셋 | 개수 (N) | 사용 모델 | 목적 |
|:----------|:---------|:---------|:----------|:-----|
| RQ1 | Internal Micro-benchmark | 15 | All Models (Local + Cloud) | 추론 능력 검증 (Adversarial Traps) |
| RQ2 | External Benchmark (GoLiSA) | ~651 | Local Only (Qwen, Llama, Semgrep) | 일반화 검증 (통계 검정 가능) |
| RQ3 | External Sample | 30~50 | All Models (Local + Cloud) | 비용 효율성 검증 |

### 실행 방법
1. GoLiSA 다운로드: Zenodo (DOI: 10.5281/zenodo.7896323) 또는 lisa-analyzer/go-lisa GitHub
2. 전수 조사 (RQ2): 로컬 모델(Qwen, Llama) + Semgrep → 651개 전체 (비용 0원, GPU만 사용)
3. 표본 조사 (RQ3, 선택): 클라우드 모델(Claude, Gemini) → 30~50개 랜덤 샘플링 (API 비용 소액)
4. N=651에서는 t-test 등 통계적 검정 수행 가능

### 예상 효과
- N=15 한계 비판(ISS_006)에 대한 결정적 방어
- "우리 모델이 야생(Wild) 코드에서도 작동한다"는 일반화 증명
- 100%가 아닌 현실적 수치가 나올 경우 오히려 신뢰도 상승

## 전제조건 (독립 검증 완료 2026-02-09)

### GoLiSA 데이터셋 접근성 — 중대 제약 확인
- **Zenodo DOI는 부정확**: 10.5281/zenodo.7896323 → 관련 없는 논문으로 리다이렉트 (딥리서치 오류)
- **GitHub repo**: 도구 코드만 포함. 651개 체인코드 미포함
- **DARTS Artifact**: 4.98 GB OVA VM 이미지 내에 데이터셋 포함 가능성 있음
- **결론**: GoLiSA 데이터셋은 단순 다운로드 불가. OVA VM을 VirtualBox 등으로 열어 내부에서 .go 파일을 추출하거나, 저자에게 직접 연락하여 데이터셋을 요청해야 함
- **"651개" 숫자**: 원문 PDF에서 직접 검증 필요

### 기존 스크립트 호환성 (코드 분석 완료)
- **02_run_audit_v3.py (로컬 LLM)**: `--dataset-dir` CLI 인자로 경로 변경 가능. 코드 수정 불필요. 파일명 의존성 없음. .go 파일만 있으면 실행 가능.
  - 예상 소요: 651 files x 2 models x 3 prompts = 3,906건, avg ~10s/건 = ~11시간+
- **05_run_traditional_tools.py (Semgrep)**: 경로가 하드코딩됨(코드 수정 필요). `vuln_`/`safe_` 파일명 prefix 의존성 있음.
  - 경로 하드코딩 수정 + prefix 처리 로직 변경 필요
  - 예상 소요: 651 files x 2 configs = 1,302건, avg ~3-5s = ~1.5-2시간

### 실행 가능성 판단
- **현재 상태**: GoLiSA 데이터셋을 확보하지 못한 상태. OVA 다운로드(4.98GB) + VM 마운트 + 파일 추출 절차 필요
- **대안**: GoLiSA 저자(Olivieri, Arceri 등)에게 이메일로 .go 파일셋 직접 요청

## 관련 이슈
- ISS_006: N=15 일반화 한계 (본 실험이 해결책)
- ISS_010: GoLiSA 인용 (본 실험 결과가 Related Work 기술에 영향)
- ISS_012: Deep Research 참고문헌 (GoLiSA 원문 PDF 확보 필요)
