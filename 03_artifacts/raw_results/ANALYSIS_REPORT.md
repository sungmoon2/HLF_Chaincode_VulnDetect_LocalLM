# 실험 결과 비판적 분석 보고서
**일시**: 2026-02-09 (Gemini 결과 추가: 2026-02-09 16:10)
**분석 대상**: 240건 감사 결과 (로컬 sLM 120건 + Claude API 45건 + Gemini API 45건 + semgrep 30건)
**데이터셋**: 15개 Go 체인코드 (vuln 9 + safe 6)
**분류 기준**: severity 키워드 기반 robust classifier (단순 문자열 매칭 아님)

---

## 1. 통합 결과표 (실측, 수정된 분류기)

### 1.1 원본 데이터셋 (15 files)

| 모델 | 프롬프트 | TPR (vuln 9) | TNR (safe 6) | FN 파일 | FP 파일 |
|:------|:---------|:-------------|:-------------|:--------|:--------|
| **Qwen2.5-Coder-7B** | zero_shot | 9/9 (100%) | 6/6 (100%) | - | - |
| **Qwen2.5-Coder-7B** | few_shot | 9/9 (100%) | 6/6 (100%) | - | - |
| **Qwen2.5-Coder-7B** | cot | 9/9 (100%) | 6/6 (100%) | - | - |
| Llama-3.1-8B | zero_shot | 9/9 (100%) | 1/6 (17%) | - | safe_01,02,03,04,06 |
| Llama-3.1-8B | few_shot | 9/9 (100%) | 1/6 (17%) | - | safe_01,02,03,04,05 |
| Llama-3.1-8B | cot | 9/9 (100%) | 1/6 (17%) | - | safe_01,02,04,05,06 |
| Claude Haiku 4.5 | zero_shot | 9/9 (100%) | 5/6 (83%) | - | safe_03 |
| Claude Sonnet 4.5 | zero_shot | 9/9 (100%) | 2/6 (33%) | - | safe_01,02,03,06 |
| Claude Opus 4.5 | zero_shot | 9/9 (100%) | 5/6 (83%) | - | safe_02 |
| Gemini 2.5 Pro | zero_shot | 9/9 (100%) | 0/6 (0%) | - | 전부 |
| Gemini 2.5 Flash | zero_shot | 9/9 (100%) | 0/6 (0%) | - | 전부 |
| Gemini 2.5 Flash Lite | zero_shot | 9/9 (100%) | 2/6 (33%) | - | safe_01,02,03,06 |

### 1.2 난독화 데이터셋 (15 files, zero-shot only)

| 모델 | TPR (vuln 9) | TNR (safe 6) | FN 파일 | FP 파일 |
|:------|:-------------|:-------------|:--------|:--------|
| **Qwen2.5-Coder-7B** | 7/9 (78%) | 4/6 (67%) | vuln_01_b, vuln_02 | safe_01, safe_03 |
| Llama-3.1-8B | 9/9 (100%) | 0/6 (0%) | - | 전부 |

### 1.3 전통 도구 (semgrep, security-audit config)

| 도구 | TPR (vuln 9) | TNR (safe 6) | 합의 취약점 탐지 |
|:------|:-------------|:-------------|:----------------|
| semgrep | 0/9 (0%) | 6/6 (100%) | **0건** |

---

## 2. 비판적 관찰 (Critical Observations)

### 2.1 Qwen의 강점은 실재하나, 암기 의존 증거가 있다

**원본 데이터에서 Qwen은 완벽하다.** 3가지 프롬프트 전략 모두에서 TPR 100%, TNR 100%. 프롬프트 변경에 무관하게 일관된 성능은 모델이 프롬프트에 민감하지 않음을 보여준다.

**그러나 난독화 시 성능이 크게 저하된다:**
- TPR: 100% → 78% (vuln_01_b_interprocedural, vuln_02_global 미탐)
- TNR: 100% → 67% (safe_01_logging, safe_03_map_read 오탐)

놓친 파일 분석:
- `vuln_01_b_interprocedural`: `getCurrentTimestamp()` → `FuncE()`로 치환되면 `time.Now()`와의 연결을 추적 실패. 원본에서는 함수명이 힌트로 작용.
- `vuln_02_global`: `transactionCounter` → `G1`로 치환되면 전역 변수 변이 패턴을 인식 실패.

**해석**: Qwen은 코드 구조(API 호출 흐름)와 명명 단서(naming cues) 모두를 사용하여 판단하며, 명명 단서가 제거되면 일부 취약점을 놓친다. 이는 순수 추론이 아닌 **부분적 패턴 매칭 + 부분적 추론**의 하이브리드 전략을 시사한다.

### 2.2 Llama의 문제는 아키텍처적이며 프롬프트로 해결 불가

Llama-3.1-8B는:
- TPR: 3가지 프롬프트 모두 100% (취약점 탐지 자체는 문제없음)
- TNR: 3가지 프롬프트 모두 17% (safe_05만 일관되게 통과)
- 난독화: TPR 100% 유지, TNR 0%로 더 악화

**Few-shot과 CoT가 Llama의 TNR을 개선하지 못한다.** 예시를 보여주거나 단계별 추론을 요구해도, Llama는 safe 파일에서 "의심스러운" 키워드(`time.Now`, `var`, `map`, `rand`)를 보면 무조건 경고를 발생시킨다. 이는 Llama가 **토큰 수준 패턴 매칭**에 의존하며 데이터 흐름 분석 능력이 부재함을 확인시켜 준다.

safe_05_deterministic_time.go만 통과하는 이유는 추가 조사 필요 (zero_shot에서만 — few_shot/cot에서는 실패).

### 2.3 Claude 4.5 모델 간 성능 역전: Sonnet < Haiku

| Claude 모델 | 파라미터 규모 | TPR | TNR |
|:------------|:-------------|:----|:----|
| Haiku 4.5 | 소형 | 100% | 83% |
| Sonnet 4.5 | 중형 | 100% | **33%** |
| Opus 4.5 | 대형 | 100% | 83% |

**Sonnet 4.5가 Haiku 4.5보다 TNR이 50pp 낮다.** 이는 "모델이 클수록 성능이 좋다"는 가정에 반한다. Sonnet은 4개 safe 파일(safe_01, 02, 03, 06)에서 오탐하는데, 이는 더 긴 출력을 생성하는 경향 때문일 수 있다 — 더 많은 분석을 시도할수록 FP가 증가하는 "과잉 분석(over-analysis)" 패턴.

### 2.4 Gemini 2.5 모델: 과잉 분석의 극단적 사례

Gemini 2.5 모델 3종(Pro, Flash, Flash Lite)은 모두 TPR 100%를 달성하나, TNR이 극히 낮다.

| Gemini 모델 | TPR | TNR | 주요 FP 원인 |
|:------------|:----|:----|:-------------|
| 2.5 Pro | 100% | 0% (0/6) | 전 safe 파일에서 access control/input validation 취약점 보고 |
| 2.5 Flash | 100% | 0% (0/6) | 동일 패턴 |
| 2.5 Flash Lite | 100% | 33% (2/6) | safe_04, safe_05만 정확. 나머지 4개 FP |

**핵심 관찰**: Gemini 모델은 합의 레이어 취약점이 아닌 **일반 보안 이슈(access control 부재, input validation 미흡)**를 보고한다. 프롬프트가 "access control issues, input validation flaws"를 포함하고 있어 기술적으로 오답은 아니나, 실험의 ground-truth 라벨은 합의 레이어 취약점 기준이므로 FP로 분류된다.

**해석**: Gemini는 합의 레이어 문맥을 고려하지 않고 일반적인 보안 감사를 수행하는 경향이 있다. 이는 Claude Sonnet의 "over-analysis" 패턴보다 더 극단적인 형태로, 도메인 특화 지식(HLF 합의 의미론)의 부재를 시사한다.

**Flash Lite의 예외적 성공 (safe_04, safe_05)**:
- safe_04_math_rand.go: "No vulnerabilities detected" — math/rand가 로깅 전용임을 정확히 인식 (533 chars, 0.9s)
- safe_05_deterministic_time.go: "No vulnerabilities detected" — time.Now()가 GetTxTimestamp()로 덮어쓰기됨을 정확히 인식 (3,446 chars, 3.6s)
- Flash Lite가 Pro/Flash보다 더 보수적인 판단을 보이는 것은 소형 모델의 "과잉 분석 억제" 효과일 수 있다.

**추론 시간 (zero-shot, 원본, 실측)**:
| Gemini 모델 | Total (15 files) | Avg/file |
|:------------|:-----------------|:---------|
| 2.5 Pro | 294.5s | 19.6s |
| 2.5 Flash | 170.0s | 11.3s |
| 2.5 Flash Lite | 95.8s | 6.4s |

### 2.5 safe_03_map_read.go는 범용적 함정

safe_03은 가장 많은 모델을 속이는 파일이다:
- Llama: zero_shot FP, few_shot FP
- Claude Haiku: FP
- Claude Sonnet: FP
- Gemini 2.5 Pro: FP
- Gemini 2.5 Flash Lite: FP
- Qwen 난독화: FP
- **Qwen 원본과 Gemini 2.5 Flash만 정확하게 CLEAN 판정** (단, Flash는 다른 safe 파일에서 FP)

이 파일은 `map[string]int`을 `range`로 순회하지만 `json.Marshal`이 키를 정렬하므로 결정론적이다. 이 지식(`json.Marshal sorts map keys`)이 없으면 대부분의 모델이 map 순회를 비결정론적이라고 판단한다.

### 2.5 전통 도구(semgrep)는 합의 취약점에 완전히 무력하다

semgrep security-audit 룰셋은:
- 취약 파일 9개에서 합의 관련 경고 **0건**
- safe_04_math_rand.go에서 `math/rand` 사용 경고 1건 (일반적 crypto 권고, 합의 무관)

이는 "Traditional tools are blind to consensus-layer vulnerabilities"라는 논문의 핵심 주장을 **실증적으로 입증**한다.

### 2.6 비용-프라이버시 트레이드오프

| 모델 | TPR | TNR | 비용 | 프라이버시 | 추론 시간/파일 |
|:------|:----|:----|:-----|:----------|:--------------|
| Qwen (local) | 100% | 100% | $0 | 완전 보장 | 3.9s |
| Claude Haiku (cloud) | 100% | 83% | API 과금 | 클라우드 전송 | 12.9s |
| Claude Opus (cloud) | 100% | 83% | 고비용 | 클라우드 전송 | 22.5s |
| Gemini 2.5 Flash Lite (cloud) | 100% | 33% | API 과금 | 클라우드 전송 | 6.4s |
| Gemini 2.5 Pro (cloud) | 100% | 0% | 고비용 | 클라우드 전송 | 19.6s |
| Gemini 2.5 Flash (cloud) | 100% | 0% | API 과금 | 클라우드 전송 | 11.3s |
| Llama (local) | 100% | 17% | $0 | 완전 보장 | 10.1s |
| semgrep | 0% | 100% | $0 | 완전 보장 | - |

**로컬 Qwen이 모든 클라우드 모델(Claude + Gemini)보다 TNR에서 우월하다.** Gemini 모델은 합의 레이어 특화 분석보다 일반 보안 감사를 수행하여 Claude보다도 낮은 TNR을 기록했다. 이는 논문의 핵심 가치 제안(프라이버시 보장 + 동등 이상 성능)을 더욱 강력하게 지지한다.

---

## 3. 논문 서사에 미치는 영향

### 3.1 강화되는 주장
- "Code-specialist sLM이 generalist보다 우수하다" — **3가지 프롬프트 모두에서 확인**
- "전통 도구는 합의 취약점에 무력하다" — **semgrep 0건으로 실증**
- "프롬프트 전략은 아키텍처 격차를 보완하지 못한다" — **Llama TNR 17% 고정**
- "로컬 sLM이 클라우드 SOTA와 동등 이상이다" — **Qwen TNR 100% > Claude 최고 83% > Gemini 최고 33%**
- "도메인 특화 지식이 모델 규모보다 중요하다" — **Gemini 2.5 Pro(대형, TNR 0%)가 Qwen 7B(소형, TNR 100%)에 대폭 열위**

### 3.2 약화 또는 수정이 필요한 주장
- ~~"Qwen은 순수 semantic reasoning을 수행한다"~~ → **난독화 시 TPR 78%, TNR 67%로 하락. 부분적 패턴 매칭 의존 인정 필요**
- "모델이 클수록 성능이 좋다" → **Claude Sonnet > Haiku 가정 깨짐. 과잉 분석 경향 논의 필요**

### 3.3 신규 발견 (논문에 추가 가능)
- **Over-analysis 패턴**: 더 많이 분석하는 모델(Sonnet, Llama, Gemini)이 더 많은 FP를 생성. Gemini가 이 패턴의 극단적 사례 (TNR 0%)
- **일반 보안 vs 합의 보안 혼동**: Gemini 모델은 합의 레이어가 아닌 일반 보안 이슈(access control, input validation)를 보고. 프롬프트에 두 종류 모두 포함되어 있어 도메인 컨텍스트 이해도의 차이를 노출
- **Naming cue dependency**: 코드-전문 모델의 추론이 명명 단서에 부분적으로 의존
- **safe_03의 보편적 함정**: json.Marshal의 키 정렬 지식이 정탐/오탐을 가르는 리트머스 테스트. Gemini Pro, Flash Lite도 추가로 FP
- **소형 모델의 보수적 이점**: Gemini Flash Lite(소형)가 Pro(대형)보다 TNR 우수 (33% vs 0%). Claude에서도 Haiku > Sonnet. 소형 모델이 과잉 분석을 억제하는 경향

---

## 4. GoLiSA 외부 검증 결과 (2026-02-10)

**데이터셋**: GoLiSA ECOOP 2023 Benchmark, 657개 .go, 326개 GitHub 저장소, 5,438,685 bytes
**실험 일시**: 2026-02-09 20:21 ~ 23:02 (본 실험), 2026-02-10 00:50 (보완 실험)

### 4.1 657개 파일 결과

| 도구 | Consensus-Layer 탐지 |
|:-----|:--------------------|
| Qwen zero_shot (classifier v1) | 380/657 (57.8%) flagged |
| Qwen zero_shot (classifier v2) | 477/657 (72.6%) flagged |
| Semgrep (auto + security-audit) | 0건 |

- Qwen 추론 시간: 5252.7s (87.5min), 7.995s/file 평균, 에러 0건
- Semgrep: 12건 일반 경고 (consensus 무관), 11개 파일에서 발생

### 4.2 Running_Examples 프롬프트 전략 비교

| File (chars) | zero_shot | few_shot | cot | json_mode |
|:-------------|:---------:|:--------:|:---:|:---------:|
| Channel.go (401) | X | O | O | O |
| GlobalVariable.go (350) | O | O | X | O |
| GoRoutines.go (458) | O | O | O | O |
| MapIteration.go (262) | X | O | X | O |
| MethodFunction.go (216) | X | O | O | O |
| **정확도** | **2/5** | **5/5** | **3/5** | **5/5** |

- zero_shot 결과는 classifier v2 적용 기준 (v1 기준 1/5)
- few_shot과 json_mode 모두 5/5 달성
- micro-benchmark(4,291~5,336 chars)에서는 zero_shot으로 동일 패턴 모두 탐지 → 코드 크기가 결정적 요인

### 4.3 Classifier v1 vs v2

| Classifier | 657개 vulnerable | 657개 safe | Running_Examples |
|:-----------|:----------------|:----------|:----------------|
| v1 (original) | 380 | 277 | 1/5 |
| v2 (improved) | 477 | 180 | 3/5 |
| 변경 건수 | +97 | -97 | +2 |

v2 설계 근거: LLM이 상세한 취약점 분석(severity, recommended fix 포함)을 작성한 후 "No vulnerabilities detected"를 추가하는 자기 모순 응답을 감지.

### 4.4 Context Injection Ablation (Channel.go)

| 변형 | 분류 | 시간 |
|:-----|:----|:-----|
| 원본 (빈 함수 본문) | safe | 0.1s |
| 수정 (c <- "hello" 추가) | safe | 1.8s |
| 수정 + CoT | safe | 3.6s |

채널 비결정성은 zero_shot/CoT로 탐지 불가. few_shot/json_mode에서만 탐지 성공.

### 4.5 GoLiSA 결과의 논문 기여

- **Semgrep 0건 (657개)**: micro-benchmark(15개)의 관찰을 대규모로 확인
- **프롬프트 전략 효과**: 최소 코드에서 few_shot/json_mode가 zero_shot 한계를 보상 (2/5 → 5/5)
- **자기 모순 응답**: LLM 기반 보안 도구의 실용적 과제 발견 (97건)
- **한계 인정**: 657개 파일의 per-file ground truth 없음; precision 검증 불가

---

## 5. 미해결 분석 (수동 검증 필요)

- [ ] Correct Prioritization Rate (CPR): 취약 파일에서 ground-truth 취약점이 최고 우선순위로 보고되었는지 수동 확인
- [ ] 할루시네이션 카운트: 존재하지 않는 API, 잘못된 함수 분류 등 수동 검증
- [ ] Qwen 난독화 FN 2건의 출력 내용 상세 분석
- [ ] Claude Sonnet FP 4건의 출력 내용 — 어떤 종류의 오탐인지 분류
- [ ] Gemini FP 분석: 보고된 access control/input validation 이슈가 실제 코드에 존재하는지 수동 확인 (합의 무관 FP vs 실제 취약점 혼재)
- [ ] Gemini Flash Lite safe_04/safe_05 성공 사례의 추론 경로 분석
