# FB_260421 — AANN 2026 Expert Review 분석

- **분석 기준**: v51 tex 원문 (AANN_2026_v51.tex, 438줄) 대조
- **분석 원칙**: 사실 기반, 누락/할루시네이션 없음

---

## 리뷰 톤 판정

**긍정적 + 조건부 수정 요구**. 결론부에서 "promising(고무적)"이라 평가하면서 "further revision(추가 수정)" 필요를 명시. Reject가 아닌 수정 요청 성격.

---

## 코멘트별 분석

### C1. 아키텍처 혁신/미세조정 부재

> "relies on existing pre-trained models and standard prompting techniques without proposing any architectural innovations or specialized fine-tuning"

| 항목 | 내용 |
|:-----|:-----|
| **사실 여부** | **정확**. 논문은 Qwen2.5-Coder-7B, Llama-3.1-8B 등 기존 사전학습 모델을 그대로 사용. 미세조정(fine-tuning) 없음. |
| **논문 내 기존 방어** | Section II-D Positioning (L117-119): "Unlike studies that propose new model architectures or training procedures, we evaluate how pre-trained transformer models perform on a domain-specific anomaly detection task under varied prompting regimes---without any task-specific fine-tuning." |
| **기존 ISS** | ISS_014 (Core Thesis 재정립) — resolved. 이미 "evaluation-oriented" 연구로 포지셔닝 완료. |
| **심각도** | **Low**. 이미 논문이 이 점을 명시적으로 인정하고 기여점을 evaluation study로 정의함. 새로운 방어가 필요하지 않으나, Positioning 섹션 문구 보강 검토 가능. |
| **대응 방향** | Rebuttal에서 "our contribution is empirical evaluation, not architectural innovation" 재확인. Positioning 문구가 이미 이 방어를 담고 있으므로 추가 수정 최소화 가능. |

---

### C2. 파일 수준 분류 vs. 줄 단위 현지화

> "assumes that file-level classification is sufficient for auditing when industrial applications require precise line-level localization"

| 항목 | 내용 |
|:-----|:-----|
| **사실 여부** | **부분 정확**. 논문은 file-level binary verdict를 사용함 (맞음). 그러나 "감사가 가능하다고 가정"한다는 표현은 과해석 — 논문은 file-level을 충분하다고 주장하지 않음. |
| **논문 내 기존 방어** | Section V-D Threats to Validity (L359-361): "Our evaluation assigns binary file-level verdicts; line-level localization and vulnerability-type ranking are not assessed, so a model may receive credit for a correct file-level classification without precisely identifying the issue's location within the file." |
| **기존 ISS** | 직접 대응 ISS 없음. 새로운 이슈로 등록 필요. |
| **심각도** | **Medium**. Threats에 이미 인정되어 있으나, Discussion이나 Future Work에 line-level localization을 명시적 향후 과제로 추가하면 방어력 강화 가능. |
| **대응 방향** | (1) Rebuttal: Threats L359-361 인용하여 "이미 한계로 명시" 설명. (2) 수정: Conclusion의 future work에 "line-level localization" 1문장 추가 검토. (3) 실측: 기존 모델 출력(raw CSV)에 줄 번호 정보가 포함되어 있는지 확인 → 포함 시 보충 분석 가능. |

---

### C3. 15개 파일 데이터셋 규모

> "bases its primary conclusions on a very small dataset of fifteen files which lacks the statistical power"

| 항목 | 내용 |
|:-----|:-----|
| **사실 여부** | **정확**. N=15는 통계적 검정력 부족. 논문 스스로 인정. |
| **논문 내 기존 방어** | (1) Section III-A (L137): Clopper-Pearson 95% CI 보고 (TPR 66.4-100.0%, TNR 54.1-100.0%). (2) Section III-D (L191-199): GoLiSA 657파일 보충 평가. (3) Section V-D (L327-330): "controlled diagnostic set, not a statistically representative sample." (4) 용어 선택: "curated micro-benchmark", "diagnostic benchmark" 등 규모 한계를 반영한 프레이밍. |
| **기존 ISS** | ISS_006 (N=15 일반화) — in_progress. GoLiSA 657개 + B1/B2 Running_Examples로 부분 대응 중. |
| **심각도** | **High** (반복 지적). 사전 시뮬레이션(FB_260209, FB_260209_1600, FB_260209_1930)에서도 동일 지적. 이번이 실제 리뷰어의 공식 지적이므로 무게감 있음. |
| **대응 방향** | (1) Rebuttal: 기존 방어선(CI, GoLiSA 657, diagnostic benchmark 프레이밍) 종합 제시. (2) 수정: Abstract/Conclusion에서 "15-file diagnostic benchmark"를 더 강조하고, "broader validation remains future work" 문구 보강. (3) 추가 실험: 시간/자원 허용 시 Add-on dataset(24개 후보 중 17개 keep) 실험 실행으로 N 확대 가능 (reserve package). |

---

### C4. Gemini 모델 명명법 표준화

> "should standardize the naming of the Gemini models to match official versioning and ensure consistency across all tables and textual descriptions"

| 항목 | 내용 |
|:-----|:-----|
| **사실 여부** | **부분 타당**. Table I에서는 "Gemini 2.5 Pro / Flash / Flash Lite" 사용. 본문에서는 "Gemini Pro", "Flash", "Flash Lite" 등 약칭 혼용. 공식 모델 ID와의 대조 필요. |
| **논문 내 현황 (실측)** | Table I (L162-164): `Gemini 2.5 Pro`, `Gemini 2.5 Flash`, `Gemini 2.5 Flash Lite` — 풀네임 사용. 본문 약칭: "Gemini Pro and Flash" (L228), "Flash Lite" (L229), "Gemini Pro" (L230), "Flash" (L231), "Gemini Flash Lite" (L336). Table II (L269): "Gemini 2.5 Pro/Flash" — 슬래시 약칭. Fig 1 캡션: "eight models" — 개별 명칭 미기재. |
| **기존 ISS** | 직접 대응 ISS 없음. ISS_023 (용어 통일)과 부분 관련. |
| **심각도** | **Low-Medium**. 수정 용이. 본문 약칭을 Table I과 일치시키면 해결. |
| **대응 방향** | (1) 본문의 "Gemini Pro" → "Gemini 2.5 Pro", "Flash" → "Gemini 2.5 Flash" 등 전수 통일. (2) Google 공식 모델 ID 확인하여 footnote b의 정확성 재검증. (3) Fig 1/Fig 2 축 레이블에서도 동일 명칭 사용 확인. |

---

### C5. Go 보안/정적 분석 LLM 관련 참고문헌 추가

> "should expand the references to include recent studies on large language models applied specifically to Go language security and static analysis"

| 항목 | 내용 |
|:-----|:-----|
| **사실 여부** | **타당**. 현재 Go 특화 참고문헌은 GoLiSA [11]과 VulFinder [12]뿐. LLM을 Go 보안에 적용한 연구는 별도로 인용하지 않음. 나머지 LLM 취약점 탐지 연구는 주로 Python/C/Solidity 대상. |
| **논문 내 현황** | Related Work (Section II): LLM vulnerability detection 7편 (Solidity/C/Python/npm 중심) + HLF tools 3편 (GoLiSA, VulFinder, Slither) + Code-specialist 1편 (Qwen technical report) + AANN venue 3편. 총 17편. Go + LLM 교차 연구 = 0편 명시 인용. |
| **기존 ISS** | ISS_012 (Deep Research citations) — resolved (v4에서 5편 추가). 그러나 "Go + LLM" 교차 영역 특화 논문은 추가되지 않았음. |
| **심각도** | **Medium**. 문헌 조사 + 2-3편 추가 인용 필요. |
| **대응 방향** | (1) 문헌 조사: "LLM Go vulnerability", "LLM Go static analysis", "LLM smart contract Go" 등 키워드로 2024-2026 최신 논문 탐색. (2) Related Work Section II-A 또는 II-C에 1-2문장 + 2-3편 인용 추가. (3) Positioning에서 "no prior work applies LLMs specifically to Go/HLF nondeterminism detection" 차별화 문구 강화. |

---

## 종합 대조표

| Comment | 논문 기존 방어 | 기존 ISS | 심각도 | 수정 필요도 |
|:--------|:-------------|:---------|:-------|:-----------|
| C1. 아키텍처 혁신 없음 | Positioning L117-119 | ISS_014 (resolved) | Low | 최소 (Rebuttal 중심) |
| C2. 파일 수준 분류 한계 | Threats L359-361 | 신규 필요 | Medium | 소규모 (future work 1문장) |
| C3. N=15 규모 부족 | CI + GoLiSA + framing | ISS_006 (in_progress) | High | 중규모 (문구 보강 + 선택적 추가 실험) |
| C4. Gemini 명명법 통일 | Table I 풀네임 존재 | 신규 필요 | Low-Medium | 소규모 (검색-치환) |
| C5. Go+LLM 참고문헌 확충 | 없음 | ISS_012 확장 | Medium | 중규모 (문헌 조사 + 2-3편 추가) |

---

## 신규 ISS 후보

| ID 후보 | 제목 | 출처 | 우선순위 |
|:--------|:-----|:-----|:---------|
| ISS_024 | 파일 수준 분류 한계 인정 + line-level future work 명시 | C2 | medium |
| ISS_025 | Gemini 모델 명칭 전수 통일 | C4 | medium |
| ISS_026 | Go+LLM 보안/정적 분석 참고문헌 추가 | C5 | medium |

---

## 결론

- **리뷰 성격**: 수정 요청 (Minor-to-Moderate Revision). Reject 성격 아님.
- **5건 중 3건(C1, C2, C3)은 논문이 이미 인지하고 방어선을 갖춘 known limitation**.
- **2건(C4, C5)은 actionable한 수정 사항** — 둘 다 수정 난이도 낮음~중간.
- **Showstopper 없음**. 모든 코멘트에 대응 가능.
