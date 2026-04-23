# WORK-001: Fabric Vulnerability Detection (Local sLM)
> Hardware: RTX 3090 Ti (24564 MiB) | Status: **AANN 2026 Major Revision — Phase 0-8 완료, Phase 9 진행 중** | Last Update: 2026-04-23

## 세션 정보
- 종료 시간: 2026-04-23 09:43 KST
- 작업 디렉토리: C:\Users\qkrtj\Desktop\02_논문\[AANN_2026]-China_Qingdao
- 세션 ID: S260423-phase7-8-v54
- Claude Code 모델: claude-opus-4-6 (Opus 4.6, 1M context)

## 이번 세션 작업 내용 (S260423-phase7-8-v54)

### Phase 7: Robustness Run (완료)
1. 27_run_robustness.py v1.0 실행 → 12개 이슈 발견 (raw response 미저장, CSV 완료 후 생성, resume 불능 등)
2. v1.0 프로세스 강제 종료 (PID 61144, 21/463)
3. 27_run_robustness.py v2.0 전면 재작성 (12개 이슈 수정):
   - Raw LLM response 전수 저장 (per_file/ JSON)
   - Incremental CSV (매 파일 flush)
   - finish_reason + token count 기록
   - primary_class 추적
   - Running_Examples 하위 디렉토리 경로 해결 (rglob)
   - 기능적 --resume (latest run dir 탐색)
   - meta.json 시작 시 생성
   - BENCHMARK_FREEZE SHA-256 무결성 검증
   - run.log 구조화 로그
   - Truncation 탐지 (finish_reason=length)
   - 시간 계산 버그 수정 (line 202)
   - Classifier Phase 6 동일 유지
4. Smoke test (2건 → 4건 resume) 통과
5. 본 실행: robustness_260423_0341/ (PID 950, nohup)
6. 30분 간격 자동 모니터링 (CronCreate, Job 91da12eb) × 10회
7. 464/464 완료 (errors=0, 08:06 KST)
8. 총 소요: 264.6분 (15,877초), avg 34.2s/file
9. v1.0 결과 (robustness_260423_0235/) 보존 (raw response 없음, 참고용)

### Phase 7 결과 검증 (완료)
10. 13개 무결성 검증 전수 PASS:
    - CHECK 1: BENCHMARK_FREEZE SHA-256 일치
    - CHECK 2: 464 coverage (missing=0, extra=0)
    - CHECK 3: Ground truth 일치 (mismatch=0)
    - CHECK 4: CSV vs per_file JSON (3,248 comparisons, mismatch=0)
    - CHECK 5: Raw response 무결성 (2,320 seeds, empty=0, zero_tokens=0, errors=0)
    - CHECK 6: Phase 6 vs Phase 7 교차 검증 (463 common, 82 disagreements)
    - CHECK 7: Stability/majority 재계산 (오차=0)
    - CHECK 8: 중복 키 없음 (dupes=0)
    - CHECK 9: Elapsed time 범위 정상 (neg=0, >300s=0)
    - CHECK 10: progress.json 일치 (464==464)
    - CHECK 11: Running_Examples 경로 해결 확인
    - CHECK 12: P6 vs P7 불일치 분석 (P6 정답 45, P7 정답 37)
    - CHECK 13: Majority vs individual seed 비교

### Phase 8: Analysis (완료)
11. Wilson 95% CI 산출: P6 Qwen, Semgrep, P7 Majority 전부
12. McNemar test 3쌍: P6↔P7 (p=0.44, 비유의), Qwen↔Semgrep (p<0.0001, 유의), seed_1↔seed_2 (p=0.017, 유의)
13. Family breakdown (C1-C6): C3/C4 Qwen 전담, C2 양쪽 실패
14. Complementarity: Union 30/31 (96.8%), Qwen-only 6, Semgrep-only 1, Neither 1
15. Error analysis: FN 3건 (juniorug C1, Running_Examples C2, ryu-sato C3), FP 234건 (72 perfect stability)
16. Truncation impact: 107/464 (23.1%), FPR +21pp (70.4% vs 49.4%)
17. Seed sensitivity: seed_1 구조적 보수 편향 (TPR 68.8%, TNR 56.5%, acc 57.3%)
18. Majority vote: TPR↑(90.6%) but 총 정답↓(227/464 vs seed_1 266/464)

### Phase 9: 논문 v54 (진행 중)
19. v54 전면 재구성: GoLiSA 464 메인, D1 압축 보조, D2 삭제
20. pdflatex 컴파일 성공: 5페이지, errors=0, overfull=0
21. 전수 검증 103건 수치 대조: 102 OK + 1 반올림 차이 (McNemar chi2 5.70 vs 5.69)
22. GPT 최종 자문 프롬프트 작성: GPT_v54_final_review.md

## 다음 세션 필수 작업
- [ ] GPT 자문 결과 반영 (2세션: 원문 저장 + 분석 요약 + 교차 분석)
- [ ] v54 보강 (현재 5p, 1p 여유): GPT 피드백 기반 + C5 참고문헌 추가
- [ ] ISS_023 잔여: GitHub 레포 업데이트 (464 benchmark 메타데이터 + 신규 스크립트)
- [ ] 최종 제출 준비 (AIS 플랫폼, 마감 2026-05-22)

## Current Status Summary
| Item | Status |
| :--- | :--- |
| Project Phase | **AANN 2026 Major Revision — Phase 0-8 완료, Phase 9 진행 중** |
| Models Disk Usage | 9.0G (du -sh 실측) |
| GoLiSA Benchmark | 464개 .go (32V + 432S, FROZEN) |
| Benchmark IAA | **κ=0.766 (Substantial)**, raw=96.5% (445/461) |
| Adjudication | **16/16 해결** (전부 1st 확정) |
| Phase 6 Qwen TPR | **80.6%** [63.7, 90.8] (25/31) |
| Phase 6 Qwen TNR | **48.6%** [43.9, 53.3] (210/432) |
| Phase 6 Semgrep TPR | **77.4%** [60.2, 88.6] (24/31) |
| Phase 6 Semgrep TNR | **99.3%** [98.0, 99.8] (429/432) |
| Phase 7 Majority TPR | **90.6%** [75.8, 96.8] (29/32) |
| Phase 7 Majority TNR | **45.8%** [41.2, 50.5] (198/432) |
| Phase 7 Stability | avg=0.809, perfect=171/464 (36.9%) |
| Union TPR | **96.8%** (30/31), Qwen-only 6, Sg-only 1 |
| McNemar P6↔P7 | p=0.44 (비유의) |
| McNemar Qwen↔Semgrep | p<0.0001 (유의) |
| Phase 7 Robustness | **완료** (464/464, errors=0, 264.6min) |
| Phase 8 Analysis | **완료** (CI + McNemar + family + error) |
| Truncation | 107/464 (23.1%), FPR +21pp |
| Paper Version | **v54** (5p, 103건 수치 검증 PASS) |
| Running_Examples 경로 | **해결** (v2.0 rglob) |
| Execution Checklist | **v2.0** (Phase 0-8 완료) |
| GPT Advisory | 12세션 완료, 대기 0건, v54 최종 자문 프롬프트 준비 |
| Active Issues | ISS_023 잔여 (GitHub 업데이트) |
| AANN Full Paper 마감 | 2026-05-22 (D-29) |
