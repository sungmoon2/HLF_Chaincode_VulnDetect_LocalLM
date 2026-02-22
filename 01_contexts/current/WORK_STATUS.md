# WORK-001: Fabric Vulnerability Detection (Local sLM)
> Hardware: RTX 3090 Ti (24564 MiB) | Status: AMLDS 저자정보 양식 제출 (리뷰 대기) | Last Update: 2026-02-22 19:31:00 (KST)

## 세션 정보
- 종료 시간: 2026-02-22 19:31:00 (KST)
- 작업 디렉토리: /c/Users/qkrtj/Desktop/00_논문모음/[AMLDS_2026]Japan_Osaka
- 세션 ID: S260222-1931
- Claude Code 모델: claude-opus-4-6 (Opus 4.6)

## Current Status Summary
| Item | Status |
| :--- | :--- |
| Project Phase | AMLDS 2026 저자정보 양식 제출 완료 (Paper ID: S2700, 리뷰 결과 대기) |
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
| Scripts | 18개 .py (01~13 + 14_local_repeat + 15_claude_repeat_cot + 16_gemini_repeat_cot + 17_cloud_single_model_repeat) |
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
| AMLDS Paper ID | S2700 (제출 확인 메일 수신 2026-02-12, 저자정보 양식 제출 2026-02-14) |
| GitHub Repository | https://github.com/sungmoon2/HLF_Chaincode_VulnDetect_LocalLM (Public, 828+3 files) |
| Reproducibility Docs | PROMPTS.md + CLASSIFIER.md + LABELING_CRITERIA.md (교차검증 완료, 불일치 0건) |
| Active Issues | 1 open (ISS_013 GoLiSA 외부검증), 3 in_progress (ISS_006 N=15일반화, ISS_007 100%역설, ISS_023 재현성보강 3/4완료), 19 resolved |

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
| Paper Final Edit | v31 교정+한글 | v31 교정본 + v31 한글버전 생성 | 카메라레디 피드백 대기 |
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
- [ ] 손상 OVA + RAW 이미지 정리 (~18GB)
- [ ] CPR/할루시네이션 수동 검증 (Run 04) — 제출 후 보완 가능
- [ ] 카메라레디 리뷰어 피드백 대응
- [ ] 오랄 발표 슬라이드 + 대본 작성 (카메라레디 확정 후)

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

## 다음 세션 필수 작업
- [ ] 리뷰 결과 대기 (Paper ID: S2700)
- [ ] 카메라레디 리뷰어 피드백 대응
- [ ] Fig2.png CoT 반영 이미지 업데이트 (사용자 작업중)
- [ ] 오랄 발표 슬라이드 + 대본 작성 (카메라레디 확정 후)
- [ ] 제출 사이트 Abstract/Keywords v31 기준 갱신 (AMLDS2026_SUBMISSION_FORM.md에 준비 완료)
- [ ] ISS_023 잔여 2건: 베이스라인 강화 + 용어 통일 (리뷰어 피드백 후)
- [ ] CPR/할루시네이션 수동 검증 (Run 04) — 제출 후 보완 가능

## Connectivity Reference
- **Previous Session:** S260214-1917 (Chain Hash: i1k31917)
- **Current Session:** S260222-1931
- **Chain Hash:** j2l41931
