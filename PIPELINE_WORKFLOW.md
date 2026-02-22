# Pipeline Workflow - HLF Chaincode Vulnerability Detection

## System Architecture

```
Claude Code Session (JSONL)
    |
/export-main command
    |-> 1. WORK_STATUS.md update
    |-> 2. Session file extraction (PowerShell)
    |-> 3. AI Snapshot generation (Claude Opus 4.5)
    |-> 4. CHAIN_INDEX.json update
```

## Workflow Steps

### NOTE: Path Convention
- Claude Code 실행 위치: `[AMLDS_2026]Japan_Osaka/`
- `.claude/` 디렉토리: 상위 디렉토리 (`[AMLDS_2026]Japan_Osaka/.claude/`)
- 프로젝트 데이터: `Exp_HLF_Chaincode_VulnDetect_LocalLM/` 서브디렉토리
- 아래 경로들은 프로젝트 서브디렉토리 기준 상대경로

### STEP 1: Session Start (/read-main)
1. Read `Exp_HLF_Chaincode_VulnDetect_LocalLM/01_contexts/current/WORK_STATUS.md`
2. Verify hash chain in `Exp_HLF_Chaincode_VulnDetect_LocalLM/01_contexts/CHAIN_INDEX.json`
3. Analyze recent 3 AI snapshots
4. Check model download status (`02_resources/models/`)
5. Check dataset status (`02_resources/dataset/`)
6. Check audit results (`03_artifacts/raw_results/`)
7. Display AI-recommended next actions

### STEP 2: Work Execution
- Environment setup, model download, dataset preparation
- Audit execution, result analysis
- Claude Code auto-records to JSONL

### STEP 3: Session End (/export-main)

**3.1 Progress Update**
- Session start/end time
- Experiment phase progress
- Actual work performed
- Unresolved issues

**3.2 Session File Extraction (PowerShell)**
```
export-main-session.ps1:
1. Convert project name (special chars -> -)
2. Search: %LOCALAPPDATA%\Temp\claude\*
3. SHA256 duplicate check
4. Copy to exports/main_YYYYMMDD_HHMMSS.jsonl
```

**3.3 AI Snapshot (Python + Claude Opus 4.5)**
```
generate_ai_snapshot.py:
1. Parse JSONL (8,000 char limit)
2. Claude Opus 4.5 analysis
3. Output: JSON + Markdown snapshot
4. Fallback: pattern matching if AI unavailable
```

**3.4 Chain Index Update**
```json
{
  "hash": "current_sha256[:16]",
  "previous_hash": "last_session_hash",
  "session_type": "setup|experiment|analysis|debugging",
  "quality_score": 85,
  "importance": "HIGH"
}
```

## Data Flow

```
Input: Claude Code Session (JSONL)
  |-- user messages
  |-- assistant responses
  |-- tool calls
  |-- timestamps

Processing:
  PowerShell extraction -> Python parsing -> Claude Opus 4.5 analysis
  -> JSON structuring -> Markdown rendering

Output: 01_contexts/
  |-- current/WORK_STATUS.md (progress)
  |-- exports/main_YYYYMMDD_HHMMSS.jsonl (raw backup)
  |-- snapshots/ai_snapshot_*.json (AI analysis)
  |-- snapshots/ai_snapshot_*.md (human readable)
  |-- CHAIN_INDEX.json (connectivity)
```

## Experiment Pipeline

```
[Phase 1: Setup]
  Python + llama-cpp-python (CUDA) + dependencies
      |
[Phase 2: Models]
  Qwen2.5-Coder-7B-Instruct (Q4_K_M)
  Meta-Llama-3.1-8B-Instruct (Q4_K_M)
      |
[Phase 3: Dataset]
  5 Go chaincode files with known vulnerabilities
  Categories: access_control, input_validation,
              phantom_read, data_leakage,
              nondeterminism, key_management
      |
[Phase 4: Audit]
  02_run_audit.py -> n_gpu_layers=-1 (full GPU offload)
  System prompt: HLF Security Expert
  Output: audit_log.csv
      |
[Phase 5: Analysis]
  Compare model outputs
  Generate figures
  Write paper section
```

## Cost Analysis

**Claude Opus 4.5 (AI Snapshot)**:
- Input: ~2,000 tokens/session
- Output: ~500 tokens/session
- Cost: ~$0.015/session
- Monthly (100 sessions): ~$1.50

## Error Handling

| Error | Action |
|:------|:-------|
| AI service failure | Fallback to pattern matching |
| No session file | Generate placeholder |
| Chain integrity broken | Log warning + recovery |
| Model not found | Skip model in audit |
| CUDA OOM | Reduce n_ctx or batch size |
