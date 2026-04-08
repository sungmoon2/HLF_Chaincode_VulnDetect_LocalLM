# WORK-001: Fabric Vulnerability Detection (Local sLM)
> Hardware: RTX 3090 Ti (24564 MiB) | Status: AANN 2026 투고 확정. v43 최종본 완성 (GPT v39 검토 1/2 + 2/2 교차분석 반영, 미세수정 20건 + 용어통일 + AANN선례 4편 + GenAI disclosure + 비판적 전수검증 PASS). | Last Update: 2026-04-07 23:30:00 (KST)

## 세션 정보
- 종료 시간: 2026-04-07 23:30:00 (KST)
- 작업 디렉토리: C:\Users\qkrtj\Desktop\02_논문\[AANN_2026]-China_Qingdao
- 세션 ID: S260407-2330
- Claude Code 모델: claude-opus-4-6 (Opus 4.6, 1M context)

## Current Status Summary
| Item | Status |
| :--- | :--- |
| Project Phase | AMLDS 2026 (S2700) 철회 확정 (2026-04-07). AANN 2026 (칭다오, 8/7-9, IEEE CIS) 후보. v37 tex 수정 없이 AANN 제출 가능 (동일 IEEEtran.cls conference mode, 6p). Full Paper 마감 2026-05-22. 비용은 ACK 과제(SOCFAI, 한-EU KIAT P0026190)로 충당 예정 |
| Models Disk Usage | 9.0G (du -sh 실측) |
| Dataset (Run 04) | 15개 .go (vuln 9 + safe 6), 82,151 bytes |
| Dataset Obfuscated | 15개 .go, 44,658 bytes, 459개 식별자 치환 |
| GoLiSA Benchmark | 657개 .go, 326개 저장소, 5,438,685 bytes (02_resources/golisa_benchmark/Benchmark/) |
| GoLiSA API Research | 7개 JSON (GitHub API 4 + Zenodo API 3), 99,240 bytes (02_resources/golisa_benchmark/api_research/) |
| audit_v3_original_260209_1422.csv | 306,962 bytes, 90 records (2 models x 3 prompts x 15 files) |
| audit_v3_obfuscated_260209_1445_2.csv | 110,290 bytes, 30 records (2 models x 1 prompt x 15 files) |
| claude_audit_260209_1503.csv | 272,744 bytes, 45 records (3 Claude 4.5 models x 15 files) |
| traditional_tools_260209_1422.csv | 3,437 bytes, 30 records (semgrep 2 configs x 15 files) |
| gemini_audit_260209_1556.csv | 137,883 bytes, 45 records (3 Gemini 2.5 models x 15 files) |
| golisa_semgrep_260209_2021.csv | 173,559 bytes, GoLiSA Semgrep 완료 (657/657 x 2 configs) |
| golisa_qwen_260209_2021.csv | 2,461,013 bytes, 657 records (Qwen zero_shot x 657 files) |
| golisa_supplementary_260210_0051.json | 보완 실험 결과 (classifier v2 + multi-prompt + context injection) |
| json_mode_microbenchmark_260210_1253.csv | 11,328 bytes, 15 records (Qwen json_mode x 15 files) |
| golisa_re_cloud_260210_1336.csv | 104,849 bytes, 30 records (6 cloud models x 5 Running_Examples, zero-shot) |
| golisa_re_llama_260210_1336.csv | 33,656 bytes, 5 records (Llama x 5 Running_Examples, zero-shot) |
| cloud_fewshot_260210_1336.csv | 379,718 bytes, 90 records (6 cloud models x 15 files, few-shot) |
| 논문 v37 영문 (최종) | 260331_v37_카메라레디.tex (v31→v37: GPT 자문 8회 + 실측 전수 검증, PDF 6p, 0err 0overfull) |
| 논문 v31 영문 (교정) | 260211_v31_GPT52교정.tex (GPT 5.2 C1~C16 반영, 61건 수정, PDF 6p, 0err 0overfull) |
| 논문 v31 한글 | 260211_v31_GPT52교정_한글.tex (v31 영문 1:1 번역, 12항목 전수검증 PASS, PDF 7p, 0err 0overfull) |
| 논문 v30 영문 (제출) | 260210_v30_서브섹션병합.tex (33,488 bytes, 13개 참고문헌, PDF 6p, AMLDS 2026 제출) |
| 논문 v26 한글 | 260210_v26_H배치_최종_한글.tex (PDF 7p, [H]강제배치, Overfull 0건) |
| 논문 풀이 | 260211_v31_논문전체풀이_AtoZ.md (v31 기준 A-to-Z 전체 해설) |
| 저자 | 4인: Park, Jeong, Choi, Yang(교신, 마지막) — v27에서 Kim 삭제 |
| ACK | IITP (IITP-2026-RS-2024-00436773, 2025→2026) + KIAT (P0026190) |
| Figure 디렉토리 | figure/ (Fig1/, Fig2/, Fig3/, fig4_audit_pitfall/) — v27에서 graphicspath 수정, Fig1/Fig3 참조 교정 |
| References | 15개 파일 (01_contexts/references/) |
| VERIFIED_REFERENCES.md | v20 기준 13개 참고문헌 (고아 2건 삭제), 할루시네이션 0건, IEEE Xplore 저자순서 확인 |
| Scripts | 20개 .py (01~17 기존 + 18_semgrep_hlf_eval + 19_measure_vram) |
| Semgrep HLF Rules | rules/hlf_consensus.yml (4 active rules: time.Now, goroutine, map-iter-putstate, iterator-close) |
| vram_measurement.json | Qwen 13,407 MiB peak, Llama 14,000 MiB peak (03_artifacts/raw_results/) |
| 카메라레디 v37 (최종) | 260331_v37_카메라레디.tex/pdf (6p, 0err, 0overfull) — archive_AMLDS_2026/카메라레디_피드백/LaTeX/ |
| 오버리프 최종 PDF | AMLDS_2026_Japan_Osaka.pdf — v37 tex와 완전 일치 확인 (archive_AMLDS_2026/) |
| 카메라레디 v32~v36 | 중간 버전 (archive_AMLDS_2026/카메라레디_피드백/LaTeX/) |
| GPT 자문 | 8회 자문 + 7회 분석 + 1회 교차대조 (archive_AMLDS_2026/카메라레디_피드백/reference/) |
| 카메라레디 리포트 | 카메라레디_수정_리포트.html (archive_AMLDS_2026/카메라레디_피드백/) |
| AANN 2026 정보 | AANN_2026_INFO.md (루트, 양식/형식/일정/등록비/토픽/호환성/비용출처) |
| AANN 템플릿 | 논문_템플릿_AANN/ (LaTeX + Word + Checking list + 투고 안내) |
| AMLDS 아카이브 | archive_AMLDS_2026/ (템플릿, 마스터가이드, 출장, 카메라레디, 저자양식, PDF 2건) |
| 04_feedback | FEEDBACK_INDEX.json v1.4 + 6 피드백 + 23 안건 (19 resolved, 2 in_progress, 2 open) + RESPONSE_STRATEGY.md |
| Figures | Fig1.png (난독화, figure/Fig1/), Fig2.png (모델비교, figure/Fig2/), Fig3.png (프롬프트전략, figure/Fig3/) |
| local_repeat_260210_2042.csv | 완료, 450건, 0에러, 1,607,677 bytes (14_local_repeat.py) |
| repeat_claude-haiku-4-5_260210_2110.csv | 완료, 225건, 0에러, 1,278,433 bytes (17 병렬) |
| repeat_claude-sonnet-4-5_260210_2110.csv | 완료, 225건, 0에러, 1,342,052 bytes, 6207.3s (17 병렬) |
| repeat_claude-opus-4-5_260210_2110.csv | 완료, 225건, 0에러, 1,244,395 bytes, 5749.1s (17 병렬) |
| repeat_gemini-2_5-pro_260210_2110.csv | 완료, 225건, 0에러, 123,000 bytes, 4340.8s (17 병렬) |
| repeat_gemini-2_5-flash_260210_2110.csv | 완료, 225건, 0에러, 211,998 bytes (17 병렬) |
| repeat_gemini-2_5-flash-lite_260210_2110.csv | 완료, 225건, 0에러, 1,548,849 bytes (17 병렬) |
| AMLDS2026_SUBMISSION_FORM.md | 제출 양식 데이터 (4저자, 제목, 초록, 키워드 8개), 2,520 bytes |
| MiKTeX 25.12 | winget 설치 완료, pdflatex 컴파일 가능 |
| AMLDS Paper ID | S2700 (제출 확인 메일 수신 2026-02-12, 저자정보 양식 제출 2026-02-14) — **철회 확정 2026-04-07** |
| AANN 2026 | **투고 확정**. 칭다오, 8/7-9. IEEE+CIS. Full Paper 마감 2026-05-22. $550/편 (4~6p). iicaann@163.com |
| 논문 v43 (최종본) | AANN_2026_v43.tex (39,642 bytes, 6p, 0err, 0overfull) — 논문_초안/ |
| 논문 v42 | AANN_2026_v42.tex (39,642 bytes, CoT약어+ACK병합) — 논문_초안/ |
| 논문 v41 | AANN_2026_v41.tex (39,503 bytes, Fig제거) — 논문_초안/ |
| 논문 v40 | AANN_2026_v40.tex (41,294 bytes, 미세수정20건+선례+disclosure+Fig교체) — 논문_초안/ |
| 논문 v39 (AANN 재프레이밍) | AANN_2026_v39_part2_final.tex (38,906 bytes, 6p, 0err, 0overfull) — 논문_초안/ |
| 논문 v38 (Part 1/2) | AANN_2026_v38_reframing.tex (38,912 bytes, 6p, 0err, 0overfull) — 논문_초안/ |
| GPT v39 검토 S4 (1/2) | 논문_초안/reference/raw/GPT54_session4_v39_review_part1.md (12,750 bytes) |
| GPT v39 검토 S5 (2/2) | 논문_초안/reference/raw/GPT54_session5_v39_review_part2.md (7,363 bytes) |
| GPT v39 교차 분석 | 논문_초안/reference/analysis/GPT54_v39_review_cross_analysis.md (11,263 bytes) |
| GPT 자문 Session 4 (Part 1/2) | raw/GPT54_session4_modification_review_part1.md (10,520 bytes) |
| GPT 자문 Session 5 (Part 2/2) | raw/GPT54_session5_modification_review_part2.md (10,076 bytes) |
| GPT 교차 분석 | analysis/GPT54_modification_review_cross_analysis.md (10,337 bytes) |
| GPT v39 검토 프롬프트 | prompts/GPT54_v39_review_and_next_steps.md (7,666 bytes) — 두 세션에 발송 완료, 응답 대기 중 |
| 논문_초안 아카이브 | archive_AMLDS_2026/논문_초안_v1-v31/ (283MB, v1~v31 tex/pdf 95개+) |
| 와처 (AANN) | PID 202120, AANN_2026-China_Qingdao 슬러그. JSONL 파싱 실패 중 (스냅샷 미생성) |
| GitHub Repository | https://github.com/sungmoon2/HLF_Chaincode_VulnDetect_LocalLM (Public, 828+3 files) |
| Reproducibility Docs | PROMPTS.md + CLASSIFIER.md + LABELING_CRITERIA.md (교차검증 완료, 불일치 0건) |
| Active Issues | 1 open (ISS_013 GoLiSA 외부검증), 1 in_progress (ISS_006 N=15일반화 — W1 claim calibration으로 부분 대응), 21 resolved (ISS_007 + ISS_023 해결) |
| AMLDS Review | Score 2 (Accept). Originality Fair, Technical Good, Clarity Good, Relevance Good, Significance Fair — **철회 확정** |
| 카메라레디 마감 | AMLDS: 2026-04-12 (철회 확정, 미제출). AANN: Full Paper 2026-05-22 / Final Paper 2026-07-05 |

## Verified Environment
| Component | Value |
|:----------|:------|
| Python | 3.11.9 |
| llama-cpp-python | 0.3.16 (CUDA 13.0 + Ninja + MSVC 14.44.35207 소스 빌드) |
| GPU | NVIDIA GeForce RTX 3090 Ti (24564 MiB, Compute 8.6) |
| GPU Driver | 581.29 |
| CUDA Toolkit | 13.0 (V13.0.88) |
| GPU Offload | True |
| anthropic | 0.75.0 (vertex 지원) |
| semgrep | 1.151.0 |
| huggingface_hub | 0.35.3 |
| pandas | 2.3.3 |
| tqdm | 4.67.1 |
| colorama | 0.4.6 |
| OS | Windows 11 (Build 10.0.26100.7623) |

## 실험 단계별 진행도
| 단계 | 진행률 | 주요 작업 | 미해결 사항 |
|------|--------|-----------|------------|
| Environment Setup | 100% | llama-cpp-python CUDA 13.0 빌드, VS Build Tools | None |
| Model Download | 100% | Qwen2.5-Coder-7B (4,683,073,536 bytes), Llama-3.1-8B (4,920,739,232 bytes) | None |
| Dataset Preparation | 100% | Run 04: 15개 .go (vuln 9 + safe 6), 82,151 bytes. Run 01~03 archived/잔존. | None |
| Audit Execution | 100% | Run 04: 380건 총 감사 (기존 255 + B1 30 + B2 5 + B3 90) | None |
| Result Analysis | 100% | Run 04 비판적 분석 완료 + B1/B2/B3 보완 실험 분석 | None |
| Paper Drafting | 100% | v31 GPT52교정 (C1~C16 반영, 61건 수정, 6p) | None |
| Paper Final Edit | v43 최종 확정 | v37→v39 재프레이밍 + v39→v43 미세수정(과장완화20건+용어통일+AANN선례4편+GenAI disclosure+CoT약어+ACK병합+TableIII footnote) + 비판적 전수검증 PASS | AANN Full Paper 마감 2026-05-22 |
| Camera-Ready Review | AMLDS 철회 | Score 2 Accept → 철회 확정 (2026-04-07). v37 tex AANN 제출 가능 확인 | AANN Full Paper 마감 2026-05-22 |
| Semgrep Custom Rules | 완료 | HLF-specific 4룰 작성, TPR 8/9, TNR 4/6, RE 2/5 | None |
| VRAM Measurement | 완료 | Qwen 13,407 MiB, Llama 14,000 MiB (nvidia-smi 실측) | None |
| Llama Error Audit | 완료 | classifier v2로 90건 전수 분석: 78/90 FP (86.7%) | Table II 교정 완료 |
| Reference Verification | 100% | v20 기준 13개 (고아2건 삭제, \cite{semgrep} 추가). 딥리서치 13건 전수 검증 완료. 할루시네이션 0건. VulFinder 저자순서 IEEE Xplore 확인 | None |
| GoLiSA Acquisition | 100% | OVA 재다운로드 → VBoxManage import → RAW 변환 → WSL 마운트 → 657개 .go 추출 | 손상 OVA + RAW 이미지 정리 미완 |
| GoLiSA External Validation | 100% | Semgrep 657/657, Qwen 657/657, 보완 실험 완료, B1(cloud RE) + B2(llama RE) 완료 | None |
| GoLiSA Supplementary | 100% | Classifier v2 재분류 (97건 변경), Multi-prompt ablation, Context injection | None |
| JSON Mode Micro-benchmark | 100% | Qwen json_mode x 15 files: TPR 9/9, TNR 6/6, 23.5s total (1.568s/file) | None |
| B1 Cloud Running_Examples | 100% | 6 cloud models x 5 RE files (zero-shot): Claude 3모델 5/5, Gemini Pro/Flash 5/5, FL 2/5 | None |
| B2 Llama Running_Examples | 100% | Llama x 5 RE files (zero-shot): 5/5 (비합의 발견 다수 포함), 85.1s | None |
| B3 Cloud Few-shot | 100% | 6 cloud models x 15 files (few-shot): Claude H/O TNR 6/6, S 5/6, Gemini P/F 0/6, FL 3/6 | None |
| Repeat + CoT Experiment | 100% | 7/7 완료: 1,800건 총, 0에러. v2 classifier 분석 완료. TPR 9/9 전수 일관. CoT: Claude 3모델+FL 6/6 일관, Pro 5/6(median), Flash 1/6(median) | None |
| Visualization | 100% | figure/ 디렉토리 4개 (fig1~fig4). v12: Fig1+Fig2+Fig3 사용. Fig2.png 생성 중 | Fig2.png 이미지 대기 |

## Checklist
- [x] Environment Setup (Libs & Dirs)
- [x] Download Models (Qwen2.5 & Llama-3.1)
- [x] Generate Dataset (Run 01~03 archived, Run 04: 15 files = vuln 9 + safe 6)
- [x] Run Audit Script (Run 04: 380 records total)
- [x] Deep Research 3건 추출 및 구조화 (references/ 15개 파일)
- [x] 실측 검증 (오류 3건 수정, 누락 4건 보완)
- [x] Analyze Results (Run 04: 비판적 분석 보고서 작성)
- [x] 피드백 시스템 구축 (04_feedback/, 5개 안건 추적)
- [x] 피드백 대응 전략 수립 (RESPONSE_STRATEGY.md)
- [x] 데이터셋 확장 (9 → 15 files: vuln 3 + safe 3 추가)
- [x] 난독화 데이터셋 생성 (03_obfuscate_dataset.py, 459개 식별자 치환)
- [x] 다중 프롬프트 실험 (Zero-shot + Few-shot + CoT, 02_run_audit_v3.py)
- [x] Claude API 비교 실험 (Haiku/Sonnet/Opus 4.5, 04_run_claude_audit.py)
- [x] 전통 도구 베이스라인 (Semgrep, 05_run_traditional_tools.py)
- [x] Gemini 2.5 감사 실험 (3 모델 x 15 파일 = 45건)
- [x] 관련연구 조사 및 참고문헌 확정 (VERIFIED_REFERENCES.md, 15개)
- [x] GoLiSA OVA 다운로드 + 추출 완료 (657개 .go, 326개 저장소)
- [x] GoLiSA 외부 검증 실험 완료 — Qwen 657/657, Semgrep 657/657 x 2
- [x] Classifier v2 설계 + 657개 재분류 (380/277 → 477/180, 97건 변경)
- [x] Running_Examples multi-prompt ablation (few_shot 5/5, json_mode 5/5)
- [x] 논문 v7~v9 영문+한글 작성
- [x] 시각화 디렉토리 구조화 (figure/fig1~fig4, Gemini 코드 v1/v2)
- [x] Figure 용어 오류 수정 (TNR Precision→Specificity)
- [x] 저자/ACK 정보 추출 및 AUTHOR_ACK_INFO.tex 작성
- [x] JSON mode 마이크로벤치마크 보완 실험 (TPR 9/9, TNR 6/6, 23.5s)
- [x] B1: GoLiSA Running_Examples x Cloud 6모델 (zero-shot) — 30건, 462.2s
- [x] B2: GoLiSA Running_Examples x Llama (zero-shot) — 5건, 85.1s
- [x] B3: Micro-benchmark x Cloud 6모델 (few-shot) — 90건, 1490.1s
- [x] 논문 v10~v12 영문+한글 작성
- [x] Gemini 리뷰 피드백 비판적 검토
- [x] v13~v22 영문+한글 작성 (다수 수정 반영)
- [x] 참고문헌 딥리서치 전수 검증 (13/13 확인, 할루시네이션 0건)
- [x] 반복/CoT 실험 7/7 완료 (1,800건, 0에러)
- [x] v23~v26 float 배치 최적화 + [H]강제배치
- [x] v27 CoT 전모델 + 저자4인 + ACK 2026
- [x] v28~v30 최종 편집 + AMLDS 2026 제출
- [x] v31 GPT52 교정본 (C1~C16 반영, 61건 수정)
- [x] v31 한글 버전 생성 (12항목 전수검증 PASS, pdflatex 7p 0err 0overfull)
- [x] 논문 전체 풀이 AtoZ 작성 (260211_v31_논문전체풀이_AtoZ.md)
- [x] 오랄 발표 슬라이드 비판적 분석 (12~14장 구조, Q&A 방어전략)
- [x] AMLDS 2026 저자정보 양식(Author Information Form) 작성 및 이메일 제출 (Paper ID: S2700)
- [x] 바탕화면 산재 파일 7개 프로젝트 디렉토리 구조화 이동 (api_research/)
- [x] GitHub Public 레포 생성 + 초기 커밋 (sungmoon2/HLF_Chaincode_VulnDetect_LocalLM, 828 files)
- [x] ISS_023 재현성 보강 3/4: PROMPTS.md + CLASSIFIER.md + LABELING_CRITERIA.md (교차검증 완료)
- [x] README.md 작성 (영문+한글 병기)
- [x] CPR/할루시네이션 수동 검증 (Run 04) — Qwen+Llama 15파일 전수 교차 대조 완료 (CPR_VERIFICATION.md)
- [ ] 손상 OVA + RAW 이미지 정리 (~18GB)
- [x] 카메라레디 리뷰어 피드백 대응 (v37 최종 확정, GPT 4건 [충분] 판정)
- [x] AMLDS 2026 철회 확정 (2026-04-07)
- [x] AANN 2026 후보 학회 조사 (템플릿 호환성 확인, 양식/형식/일정/등록비 문서화)
- [x] AMLDS 관련 파일 아카이브 (archive_AMLDS_2026/)
- [x] AANN 2026 정보 문서 생성 (AANN_2026_INFO.md)
- [x] AANN 템플릿 배치 (논문_템플릿_AANN/)
- [x] AANN 투고 확정 (교수님 결정)
- [x] 논문_초안 v1~v31 → archive_AMLDS_2026/논문_초안_v1-v31/ 아카이브 이동
- [x] 논문_초안/ AANN용 재구성 (v37 base + IEEEtran.cls)
- [x] GPT 자문 Part 1/2 수신 → 원문 저장 + 분석 요약본 작성
- [x] v38 작성 (Part 1/2 반영: 제목/초록/키워드/intro/related work/conclusion)
- [x] GPT 자문 Part 2/2 수신 → 원문 저장 + 교차 분석
- [x] v39 작성 (Part 1/2 + Part 2/2 반영: 키워드 보수화, audit procedure, caveat 재배치, 용어 통일)
- [x] GPT v39 검토 프롬프트 작성 → 두 세션에 발송
- [x] 와처 교체: 옛 AMLDS 와처(PID 205056, 종료) → 신규 AANN 와처(PID 202120)
- [x] GPT v39 검토 응답 수신 (S4: 1/2, S5: 2/2) → 교차 분석 완료 → 채택안 18건 확정
- [x] v40 작성: 미세수정 20건 + AANN 선례 4편 + GenAI disclosure + Table V 제거 + Discussion 축약 + Fig.1 pipeline 교체
- [x] v41 작성: Fig.1 제거 (float 배치 문제 해소)
- [x] v42 작성: CoT 약어 정의 + ACK 공백/문장 병합
- [x] v43 작성: Table III footnote (single-run vs median 명시) — **최종본**
- [x] v43 비판적 전수검증: 수치 전수 일치, 과장 0건, 용어불일치 0건, 참고문헌 17/17, blocking 0건
- [x] AANN 선례 4편 DOI verified 인용 (Li2025, Liu2024cae, Liu2024xgboost, Zhu2024hyperband)
- [x] GenAI disclosure 작성 (GPT-5.4 Pro + Claude Opus 4.5/4.6, language editing/structural revision only)
- [ ] AMLDS 철회 회신 확인
- [ ] git commit + push (v35~v43, 논문_초안/reference/)
- [ ] 오버리프 tex 최종 교체 (v43)
- [ ] 오랄 발표 슬라이드 + 대본 작성 (학회 확정 후)
- [x] Table II median 정합성 수정 (A안: 4셀 + footnote + 연쇄 본문 11곳)
- [x] 5-run repeat 실험 전수 재분류 (classifier v2, 78/90 확인, 13/24 확인)
- [x] safe_03 per-file claim 실측 수정 (FL→Opus)
- [x] 오버리프 PDF 전수 대조 검증 (v37 tex와 완전 일치)

## 이번 세션 작업 내용 (시간순, 사실 기반) — S260222-1931
1. /read-main 실행: 이전 세션(S260214-1917) 컨텍스트 복원
2. WORK_STATUS.md + CHAIN_INDEX.json + 스냅샷 + 모델/데이터셋/결과 상태 전수 확인
3. 바탕화면 산재 파일 7개 식별 및 내용 분석
   - GitHub API 응답 4개: golisa_contents.json, golisa_sub.json, golisa_testcases.json, golisa_nondet.json
   - Zenodo API 응답 3개: zenodo_golisa.json (false positive), zenodo_search.json, zenodo_search2.json
4. 디렉토리 구조 생성: 02_resources/golisa_benchmark/api_research/{github/, zenodo/}
5. 7개 파일 이동 완료 (바탕화면 → 프로젝트 디렉토리)
6. INDEX.json 생성: 각 파일별 API endpoint, 설명, 크기, 비고 기록
7. GitHub CLI 설치 (winget, gh 2.87.2) + 인증 (sungmoon2)
8. .gitignore 작성 (모델 9GB, OVA/RAW 22GB, 세션 JSONL, snapshots, __pycache__ 제외)
9. git init + GitHub Public 레포 생성: sungmoon2/HLF_Chaincode_VulnDetect_LocalLM
10. 초기 커밋 (828 files, 462,626 lines) + push (커밋 d834a74)
11. ISS_023 재현성 보강: 원본 스크립트 실측 읽기 → 3개 문서 작성
    - PROMPTS.md: P1~P4 프롬프트 전문 (소스 파일/행번호 참조)
    - CLASSIFIER.md: v1/v2/JSON classifier 전체 코드 + 키워드 리스트 4종
    - LABELING_CRITERIA.md: consensus-relevant 정의 + 15개 ground truth + 5개 Running_Examples
12. 교차 검증: 30+ 항목 소스 코드 전수 대조, 불일치 0건
13. 재현성 문서 커밋 + push (커밋 864e7de)
14. ISS_023 상태 open → in_progress 변경 (3/4 완료, 잔여: 베이스라인 강화 + 용어 통일)
15. README.md 작성 (영문+한글), WORK_STATUS.md + CHAIN_INDEX.json 최신화, 커밋 + push

## 이전 세션 작업 내용 (시간순, 사실 기반) — S260214-1917
1~8. (CHAIN_INDEX.json 참조 — AMLDS 저자정보 양식 작성 및 이메일 제출)

## Key Experimental Results (Run 04 — 15 Files, 380 Runs)

### 원본 데이터셋 (robust classifier 기준)
| 모델 | 프롬프트 | TPR (vuln 9) | TNR (safe 6) |
|:------|:---------|:-------------|:-------------|
| Qwen2.5-Coder-7B | zero_shot | 9/9 (100%) | 6/6 (100%) |
| Qwen2.5-Coder-7B | few_shot | 9/9 (100%) | 6/6 (100%) |
| Qwen2.5-Coder-7B | cot | 9/9 (100%) | 6/6 (100%) |
| Qwen2.5-Coder-7B | json_mode | 9/9 (100%) | 6/6 (100%) |
| Llama-3.1-8B | zero_shot | 9/9 (100%) | 1/6 (17%) |
| Llama-3.1-8B | few_shot | 9/9 (100%) | 1/6 (17%) |
| Llama-3.1-8B | cot | 9/9 (100%) | 1/6 (17%) |
| Claude Haiku 4.5 | zero_shot | 9/9 (100%) | 5/6 (83%) |
| Claude Haiku 4.5 | few_shot | 9/9 (100%) | 6/6 (100%) |
| Claude Sonnet 4.5 | zero_shot | 9/9 (100%) | 2/6 (33%) |
| Claude Sonnet 4.5 | few_shot | 9/9 (100%) | 5/6 (83%) |
| Claude Opus 4.5 | zero_shot | 9/9 (100%) | 5/6 (83%) |
| Claude Opus 4.5 | few_shot | 9/9 (100%) | 6/6 (100%) |
| Gemini 2.5 Pro | zero_shot | 9/9 (100%) | 0/6 (0%) |
| Gemini 2.5 Pro | few_shot | 9/9 (100%) | 0/6 (0%) |
| Gemini 2.5 Flash | zero_shot | 9/9 (100%) | 0/6 (0%) |
| Gemini 2.5 Flash | few_shot | 9/9 (100%) | 0/6 (0%) |
| Gemini 2.5 Flash Lite | zero_shot | 9/9 (100%) | 2/6 (33%) |
| Gemini 2.5 Flash Lite | few_shot | 9/9 (100%) | 3/6 (50%) |

### 난독화 데이터셋 (zero-shot only)
| 모델 | TPR (vuln 9) | TNR (safe 6) |
|:------|:-------------|:-------------|
| Qwen2.5-Coder-7B | 7/9 (78%) | 4/6 (67%) |
| Llama-3.1-8B | 9/9 (100%) | 0/6 (0%) |

### 전통 도구 (Semgrep)
| 도구 | TPR (vuln 9) | 합의 취약점 탐지 | 일반 경고 |
|:------|:-------------|:----------------|:---------|
| Semgrep 1.151.0 | 0/9 (0%) | 0건 | 1건 (math/rand, safe_04) |

### GoLiSA 외부 검증 (657 Files)
| 도구 | Consensus-Layer 탐지 |
|:-----|:--------------------|
| Qwen zero_shot (classifier v1) | 380/657 flagged |
| Qwen zero_shot (classifier v2) | 477/657 flagged |
| Semgrep (auto + security-audit) | 0건 |

### GoLiSA Running_Examples (5 known-vulnerable files)
| 모델 | 프롬프트 | 탐지 |
|:------|:---------|:------|
| Qwen2.5-Coder-7B | zero_shot (classifier v2) | 2/5 |
| Qwen2.5-Coder-7B | few_shot | 5/5 |
| Qwen2.5-Coder-7B | cot | 3/5 |
| Qwen2.5-Coder-7B | json_mode | 5/5 |
| Claude Haiku 4.5 | zero_shot | 5/5 |
| Claude Sonnet 4.5 | zero_shot | 5/5 |
| Claude Opus 4.5 | zero_shot | 5/5 |
| Gemini 2.5 Pro | zero_shot | 5/5 |
| Gemini 2.5 Flash | zero_shot | 5/5 |
| Gemini 2.5 Flash Lite | zero_shot | 2/5 |
| Llama-3.1-8B | zero_shot | 5/5 (비합의 발견 다수) |

### JSON Mode Micro-benchmark (15 files)
| 프롬프트 | TPR (vuln 9) | TNR (safe 6) | 총 시간 | 평균/파일 |
|:---------|:-------------|:-------------|:--------|:---------|
| json_mode | 9/9 (100%) | 6/6 (100%) | 23.5s | 1.568s |

### 추론 시간 (zero-shot, 원본, 실측)
| 모델 | Total (15 files) | Avg/file |
|:------|:-----------------|:---------|
| Qwen (local) | 59.1s | 3.941s |
| Gemini 2.5 Flash Lite | 95.8s | 6.387s |
| Llama (local) | 151.3s | 10.088s |
| Gemini 2.5 Flash | 170.0s | 11.331s |
| Claude Haiku 4.5 | 193.4s | 12.894s |
| Gemini 2.5 Pro | 294.5s | 19.630s |
| Claude Opus 4.5 | 337.3s | 22.489s |
| Claude Sonnet 4.5 | 418.1s | 27.875s |

## Context Management System (v3.0)
| Component | File | Version | Status |
|:----------|:-----|:--------|:-------|
| /export-main | .claude/commands/export-main.md | - | 동작 확인 |
| /read-main | .claude/commands/read-main.md | - | 동작 확인 |
| AI Snapshot | .claude/scripts/generate_ai_snapshot.py | v3.0 | 13섹션 R2-D2 표준, Opus 4.5 |
| Session Extract | .claude/scripts/export-main-session.ps1 | v1.4 | ~/.claude/projects/ 경로 |
| Watcher Daemon | .claude/scripts/session_watcher.py | v2.0 | PID 78064, 5-retry, supervisor |
| Auto-start | Startup 폴더 .lnk | - | 등록 완료 |
| Settings | .claude/settings.json | - | 생성 완료 |

## 이번 세션 작업 내용 (시간순, 사실 기반) — S260331-2100
1. /read-main 실행: 이전 세션(S260330-1430) 컨텍스트 복원
2. GPT 5.4 Pro 최종검증 결과 수신 → 원문 저장 + 할루시네이션 검증 (0건) + 분석 요약본 작성
3. v35 작성: GPT 최종검증 반영 16건 (supplementary evaluation 전역교체 8곳, Clopper-Pearson CI, Threats 병합, Semgrep/Llama/Error Analysis 소제목 강화, footnote 참조, 측정 protocol, bibitem 인용 보강)
4. v35 pdflatex 컴파일: 6p, 0err, 0overfull
5. v35 전수 검증 에이전트 실행: 10/10 항목 PASS, 불일치 0건
6. GPT v35 검증 프롬프트 2개 작성 (기존 세션 + 새 세션)
7. GPT 기존세션 v35 검증 결과 수신 → 원문 저장 + 분석 (6건 이슈, 할루시네이션 0건)
8. GPT 새세션 v35 검증 결과 수신 → 원문 저장 + 교차 대조 분석
9. 실측 데이터 전수 검증:
   - 78/90 FP: classifier v2 (09_reclassify_and_ablation.py 원본 로직) 적용 → 78/90 (86.7%) 실측 확인
   - 13/24 consistent: 24쌍 5-run TNR 전수 재분류 → 13/24 실측 확인
   - Table II median 불일치 4건 발견: Llama ZS 1→2, Sonnet ZS 2→3, Opus ZS 5→4, FL FS 3→6
   - Table II footnote 변천사 추적: v30(정확) → v31(범위확대) → v34(불일치 발생)
10. GPT Table II 자문 프롬프트 2개 작성 (기존 + 새 세션)
11. GPT 기존세션 TableII 자문 결과 수신 → 원문 저장 (A안 추천, 연쇄 수정 10~12곳)
12. GPT 새세션 TableII 자문 결과 수신 → 원문 저장 (A안 추천, Table III 분리 추가 권장)
13. v36 작성: A안 반영 24건 (Table II 4셀 + ‡제거 + footnote + 연쇄 본문 11곳 + Table III 분리 2곳 + GPT잔여 4곳)
14. v36 전수 검증 에이전트: 10/10 PASS
15. GPT v36 최종검증 프롬프트 작성 + 결과 수신 → 원문 저장 + 분석 (4건 [충분], blocking 0건)
16. v37 작성: GPT 최종검증 2건 (270행 median 명확화, 302행 Fig.1 single-run) + safe_03 per-file 실측 수정 (FL→Opus) + 374행 headroom 수치화
17. v37 AI표현/오타/수치 전수 대조: 이상 없음
18. 오버리프 PDF (AMLDS_2026_Japan_Osaka.pdf) 전수 대조: v37 tex와 완전 일치
19. Registration form 검토 + 작성 정보 정리 (Students $450, 동행자 없음)
20. 제출 보류 결정: 타 학회 규정 확인 대기. 논문 철회 시 미출판 상태로 타 학회 투고 가능.

## 이전 세션 작업 내용 (시간순, 사실 기반) — S260330-1430
1. /read-main 실행: 이전 세션(S260222-2030) 컨텍스트 복원
2. AMLDS 2026 리뷰 결과 수신: Score 2 (Accept), 4건 지적
3. 리뷰 파일 3개 다운로드 → 카메라레디_피드백/ 디렉토리 생성 및 구조화
   - Acceptance Letter-S2700.pdf (→ 오사카_출장/)
   - Review Form-S2700.pdf (→ 카메라레디_피드백/)
   - registration_form_full paper.docx (→ 카메라레디_피드백/)
4. GPT 5.4 Pro 자문 프롬프트 작성 (카메라레디_피드백/reference/prompts/)
5. GPT 5.4 Pro 초기 자문 수신 → 원문 저장 + 분석 요약본 작성 (할루시네이션 0건)
6. GitHub repo 재현성 인프라 강화:
   - requirements.txt 버전 고정 (8개 의존성, 실측 버전)
   - LICENSE (MIT), CITATION.cff 생성
   - REPRODUCTION.md 8단계 재현 가이드 작성 (스크립트 경로 전수 검증)
   - README.md 업데이트 (submitted→accepted, 뱃지, Citation, Reproduction)
   - GitHub Topics 8개 설정
   - git commit + push (f53917a)
7. 카메라레디 추가 실험:
   - CP1: Semgrep HLF 커스텀 룰 4개 YAML 작성 (rules/hlf_consensus.yml)
   - CP2: 15파일 + 5 RE 실행 → TPR 8/9, TNR 4/6, RE 2/5
   - CP3: nvidia-smi VRAM 실측 → Qwen 13,407 MiB, Llama 14,000 MiB (scripts/19_measure_vram.py)
   - CP4: Llama safe 90건 로그 분석 (에이전트 분석: 73/90)
8. v32 LaTeX 작성: GPT 자문 16건 수정 + Fig.2/3 삭제 + Table V 삽입 (6p, 0err)
9. GPT 기존 세션 재평가 수신 → 원문 저장 + 분석 → 잔여 5건 식별
10. v33 LaTeX 작성: 기존 세션 재평가 5건 + 보완 1건 반영 (6p, 0err)
11. GPT 새 세션 검증 수신 → 원문 저장 + 분석 → Llama CoT 27/30 수치 충돌 발견
12. 실측 데이터 검증: 논문의 classifier v2로 Llama 90건 재분류
    - 에이전트 73/90 → classifier v2 실측 78/90 (86.7%)
    - Table II few_shot TNR 1/6 → 실측 0/6 (5회 모두 0/6)
    - Table II 주석: Llama zero-shot/CoT에서도 변동 확인
13. v34 LaTeX 작성: 실측 기반 3건 교정 (6p, 0err, 0overfull) — 최종
14. 카메라레디_수정_리포트.html 작성 (리뷰어 지적별 before/after 시각화)
15. GPT 최종검증 프롬프트 작성 (Part 1: 리뷰어 대응 완결성, Part 2: 논문 자체 교정)

## 이번 세션 작업 내용 (시간순, 사실 기반) — S260407-cont
1. 이전 세션(S260407-2308) 연결성 추적: WORK_STATUS.md + CHAIN_INDEX.json + 메모리 + 스냅샷 전수 확인
2. 이전 AMLDS 디렉토리 메모리 확인: `00_논문/[AMLDS_2026]-Japan_Osaka`에는 memory 폴더 없음 (auto memory 미사용 시기). 복사할 메모리 없음 확인.
3. 와처 교체:
   - 옛 AMLDS 와처 PID 205056 → 이미 종료 상태. PID 파일 정리.
   - 옛 와처 로그: 21개 JSONL 전부 5/5 PERMANENTLY FAILED (디렉토리명 변경 후 경로 불일치)
   - 신규 AANN 와처 시작: PID 202120, 슬러그 AANN_2026-China_Qingdao, 루트 와처(PID 52312)와 충돌 없음
   - 신규 와처 JSONL 파싱 실패 중 (generate_ai_snapshot.py가 JSONL에서 0 messages 추출 — 별도 조사 필요)
4. 논문_초안 디렉토리 재구성:
   - 논문_초안/ (v1~v31, 283MB, 95개+ 파일) → archive_AMLDS_2026/논문_초안_v1-v31/ 이동
   - 신규 논문_초안/ 생성: AANN_2026_base_v37.tex + IEEEtran.cls 복사
5. GPT 자문 Part 1/2 수신:
   - 원문 저장: raw/GPT54_session4_modification_review_part1.md (10,520 bytes)
   - 분석 요약본: analysis/GPT54_modification_review_analysis.md (5,573 bytes)
   - 핵심: B안(재프레이밍) 권고, 제목/초록/키워드/contribution/related work/conclusion 수정
6. v38 작성 (Part 1/2 반영):
   - 제목: "Privacy-Preserving Anomaly Detection in HLF Chaincode Using Compact Local Transformer Models" (76자)
   - 초록/키워드/intro/related work positioning/conclusion 전면 재프레이밍
   - 컴파일: 6p, 0err, 0overfull, 12 underfull (v37: 10 underfull)
7. GPT 자문 Part 2/2 수신:
   - 원문 저장: raw/GPT54_session5_modification_review_part2.md (10,076 bytes)
   - 교차 분석: analysis/GPT54_modification_review_cross_analysis.md (10,337 bytes)
   - 일치도 90%+, 엇갈림 5건 식별 → 채택 판단 기록
8. v39 작성 (Part 1/2 + 2/2 반영, 엇갈림 절충):
   - 키워드 보수화: "few-shot learning"→"few-shot prompting", "lightweight"→"resource-efficient"
   - Positioning: "few-shot learning applied to cybersecurity" 삭제
   - Dataset caveat 축소 → Threats 참조로 이동
   - Prompt strategies → "reproducible auditing procedure" + classifier pipeline 통합
   - "consensus-layer" 용어 → "endorsement nondeterminism" 통일
   - 컴파일: 6p, 0err, 0overfull, 13 underfull
9. GPT v39 검토 프롬프트 작성 + 두 세션에 발송:
   - prompts/GPT54_v39_review_and_next_steps.md (7,666 bytes)
   - 요청 A: v39 본문 검토 (재프레이밍 성공 여부, 과장 체크, 누락 체크, 1:1 수정안)
   - 요청 B1~B4: Pipeline figure, 외적 타당성 실험 설계, AANN 선례 인용, GenAI disclosure
10. WORK_STATUS.md + CHAIN_INDEX.json + 메모리 전수 업데이트

## 이번 세션 작업 내용 (시간순, 사실 기반) — S260407-2330
1. /read-main 실행: 이전 세션(S260407-1845) 컨텍스트 복원
2. GPT v39 검토 Part 1/2 (S4) 수신:
   - 원문 저장: 논문_초안/reference/raw/GPT54_session4_v39_review_part1.md (12,750 bytes)
   - 분석 요약본: 논문_초안/reference/analysis/GPT54_v39_review_part1_analysis.md (7,242 bytes)
   - 핵심: 재프레이밍 성공 [LIKELY], 미세수정 11건 + 용어정리 5건, Fig.1→pipeline figure 교체 권고
3. GPT v39 검토 Part 2/2 (S5) 수신:
   - 원문 저장: 논문_초안/reference/raw/GPT54_session5_v39_review_part2.md (7,363 bytes)
   - 분석 요약본: 논문_초안/reference/analysis/GPT54_v39_review_part2_analysis.md (4,983 bytes)
   - 핵심: 재프레이밍 성공 [LIKELY], 미세수정 8건, public HLF sample 우선 권고
4. 교차 분석 작성: 논문_초안/reference/analysis/GPT54_v39_review_cross_analysis.md (11,263 bytes)
   - 일치도 95%+, 엇갈림 2건 (L139 처리방식, 외적타당성 우선순위) → 채택안 결정
   - 최종 수정 목록 18건 확정
5. v40 작성 (v39 기반):
   - 미세수정 20건 (과장완화 8건 + 용어통일 12건, 데이터 변경 없음)
   - AANN 선례 4편 bibitem + Related Work 1문장 (Li2025, Liu2024cae, Liu2024xgboost, Zhu2024hyperband)
   - GenAI disclosure (ACK 뒤, GPT-5.4 Pro + Claude Opus 4.5/4.6)
   - Table V (Local Inference Cost) 제거 (prose에 동일 수치 보존)
   - Discussion V-A 축약 (Results와 중복 수치 제거, 정보 손실 없음)
   - Fig.1 obfuscation bar chart → TikZ pipeline figure 교체
   - 컴파일: 6p, 0err, 0overfull
6. v40 오버리프 PDF 대조: 6p, Fig.1 렌더링 정상, 수치 일치 확인
7. Fig.1 TikZ 디자인 반복 (3회):
   - fbox → TikZ 4-stage color → Obfuscated variant 화살표 조정
   - float 배치 문제 발생: Fig.1 + Table IV 같은 컬럼에 몰림 → 본문 가로지르기
8. v41 작성: Fig.1 완전 제거 결정 (float 충돌 해소, 텍스트만으로 충분)
   - tikz 패키지 제거, Fig 참조 제거
   - 컴파일: 6p, 0err, 0overfull
9. v41 비판적 전수검증 (에이전트): CRITICAL 0, SHOULD-FIX 2, MINOR 4, COSMETIC 4
10. v42 작성: SHOULD-FIX 2건 반영
    - S1: ACK "(MSIT)(IITP-2026..." → "(MSIT) (IITP-2026..." + 2문장 병합
    - S2: "P3 (Chain-of-Thought)" → "P3 (Chain-of-Thought, CoT)" 약어 정의
11. MINOR/COSMETIC 순차 검토 (사용자 판단):
    - M1 Llama 모델명 혼동: PASS (이미 "comparable size"로 설명됨)
    - M2 ACK 반복: v42에서 해결 완료
    - M3 Claude Code 도구/모델 구분: PASS (분야에서 무방)
    - M4 Table II vs III 수치 차이: FIX (footnote 추가)
    - C1 "strong": PASS (sLM 타당성 동기에 필요)
    - C2 "7.995 s/file": PASS (실측값 유지)
    - C3 "intersection": PASS (학술 관용구)
    - C4 "promising option": PASS (적절한 hedging)
12. v43 작성: M4 반영 — Table III에 "Original values reflect single-run results and may differ from the five-run median reported in Table II." footnote 추가
13. v43 최종 오버리프 PDF 대조: 6p, 0err, 0overfull, 전항목 정상 확인
14. WORK_STATUS.md + CHAIN_INDEX.json 업데이트

## 이전 세션 작업 내용 (시간순, 사실 기반) — S260407-cont
1~10. (상기 참조 — AMLDS 철회, AANN 후보 조사, 디렉토리 재편, GPT 자문 Part 1/2 + 2/2, v38-v39 작성)

## 다음 세션 필수 작업
- [ ] AMLDS 철회 회신 확인
- [ ] 오버리프 tex 최종 교체 (v43)
- [ ] git commit + push (v35~v43, 논문_초안/reference/)
- [ ] 교수님 v43 최종 확인 → 제출 결정
- [ ] 와처 JSONL 파싱 실패 원인 조사 + 수정
- [ ] 손상 OVA + RAW 정리 (~18GB)
- [ ] 오랄 발표 슬라이드 + 대본 (Accept 후)

## Connectivity Reference
- **Previous Session:** S260407-1845 (Chain Hash: n6p71845)
- **Current Session:** S260407-2330
- **Chain Hash:** o7q82330
