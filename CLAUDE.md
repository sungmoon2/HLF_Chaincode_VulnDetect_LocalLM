# CLAUDE.md - HLF Chaincode Vulnerability Detection Project

## Project Overview
Local sLM(Small Language Model)을 활용한 Hyperledger Fabric 체인코드(Go) 보안 취약점 자동 탐지 시스템.
AMLDS 2026 (Japan, Osaka) 논문 실험 환경.

## Working Directory
- **Claude Code 실행 위치**: `[AMLDS_2026]Japan_Osaka/`
- **프로젝트 서브디렉토리**: `Exp_HLF_Chaincode_VulnDetect_LocalLM/`
- **`.claude/` 위치**: `[AMLDS_2026]Japan_Osaka/.claude/` (상위 디렉토리)

## Hardware (실측)
- GPU: NVIDIA GeForce RTX 3090 Ti (24564 MiB VRAM, Compute 8.6)
- GPU Driver: 581.29
- CUDA Toolkit: 13.0 (V13.0.88)
- OS: Windows 11 (Build 10.0.26100.7623)

## Key Commands

### /export-main
현재 세션을 백업하고 진행 상황을 기록합니다.
- `Exp_HLF_Chaincode_VulnDetect_LocalLM/01_contexts/current/WORK_STATUS.md` 업데이트
- 세션 JSONL 파일을 `Exp_HLF_Chaincode_VulnDetect_LocalLM/01_contexts/exports/`로 복사
- AI 스냅샷 자동 생성 (`Exp_HLF_Chaincode_VulnDetect_LocalLM/01_contexts/snapshots/`)
- `Exp_HLF_Chaincode_VulnDetect_LocalLM/01_contexts/CHAIN_INDEX.json` 업데이트

### /read-main
이전 세션의 컨텍스트를 복원합니다.
- 최신 `WORK_STATUS.md` 읽기
- `CHAIN_INDEX.json` 해시 체인 확인
- 최근 스냅샷 3개 요약
- 모델/데이터셋/결과 상태 확인

## Directory Structure
```
[AMLDS_2026]Japan_Osaka/                      <-- Claude Code 작업 디렉토리
├── .claude/                                   <-- 슬래시 명령 및 스크립트
│   ├── commands/
│   │   ├── export-main.md
│   │   └── read-main.md
│   ├── scripts/
│   │   ├── generate_ai_snapshot.py
│   │   └── export-main-session.ps1
│   ├── hooks/
│   ├── logs/
│   └── settings.json
│
└── Exp_HLF_Chaincode_VulnDetect_LocalLM/     <-- 프로젝트 데이터
    ├── 01_contexts/
    │   ├── current/WORK_STATUS.md
    │   ├── exports/
    │   ├── snapshots/
    │   ├── archive/
    │   └── CHAIN_INDEX.json
    ├── 02_resources/
    │   ├── dataset/                           (15 .go files, 82,151 bytes)
    │   ├── dataset_obfuscated/                (15 .go files, 44,658 bytes)
    │   └── models/                            (2 .gguf files, ~9.0GB)
    ├── 03_artifacts/
    │   ├── raw_results/                       (CSV 4개 + meta.json + ANALYSIS_REPORT.md)
    │   └── figures/
    ├── 04_feedback/                           (FEEDBACK_INDEX.json + 안건 추적)
    ├── scripts/
    │   ├── 01_download_models.py
    │   ├── 02_run_audit.py                    (v2.0)
    │   ├── 02_run_audit_v3.py                 (v3.0, 다중 프롬프트)
    │   ├── 03_obfuscate_dataset.py
    │   ├── 04_run_claude_audit.py
    │   └── 05_run_traditional_tools.py
    ├── CLAUDE.md                              (이 파일)
    ├── PIPELINE_WORKFLOW.md
    └── requirements.txt
```

## Data Integrity Rules
1. 할루시네이션 금지: 실제 작업한 내용만 기록
2. 정보 축소/확장 금지: 있는 그대로 기록
3. 예측/추정 금지: 확인된 사실만 기록
4. 시간 추정 금지

## Session Workflow
```
1. /read-main  -> 이전 컨텍스트 복원
2. 실험 작업    -> 모델 다운로드, 감사 실행, 분석
3. /export-main -> 진행 상황 저장 + AI 스냅샷
```
