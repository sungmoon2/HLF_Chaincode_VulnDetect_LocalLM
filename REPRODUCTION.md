# Reproduction Guide

Step-by-step instructions to reproduce all experiments from the paper.

## Prerequisites

| Component | Version | Notes |
|:----------|:--------|:------|
| Python | 3.11.9 | Other 3.11.x may work |
| NVIDIA GPU | RTX 3090 Ti (24 GB VRAM) | Any GPU with >= 8 GB VRAM should work for Q4_K_M models |
| CUDA Toolkit | 13.0 | Required for llama-cpp-python GPU build |
| OS | Windows 11 | Linux should work with minor path adjustments |
| Semgrep | 1.151.0 | Install via `pip install semgrep` |

### Cloud API Keys (for cloud model experiments only)

- **Claude**: Anthropic API key or Google Cloud credentials with Vertex AI access
- **Gemini**: Google Cloud credentials with Vertex AI access

Local-only experiments (Steps 1-5) require no API keys.

## Setup

```bash
# Clone repository
git clone https://github.com/sungmoon2/HLF_Chaincode_VulnDetect_LocalLM.git
cd HLF_Chaincode_VulnDetect_LocalLM

# Install dependencies
pip install -r requirements.txt
```

### llama-cpp-python CUDA Build (Windows)

The pip package does not include CUDA support by default. Build from source:

```bash
# Requires: Visual Studio Build Tools (MSVC 14.x), CMake, Ninja
set CMAKE_ARGS=-DGGML_CUDA=on
set FORCE_CMAKE=1
pip install llama-cpp-python==0.3.16 --no-binary :all:
```

Verify GPU offload:
```python
from llama_cpp import Llama
m = Llama("02_resources/models/qwen2.5-coder-7b-instruct-q4_k_m.gguf", n_gpu_layers=-1, verbose=True)
# Should show "offloaded X/Y layers to GPU"
```

## Experiment Steps

### Step 1: Download Models

```bash
python scripts/01_download_models.py
```

Downloads two GGUF models to `02_resources/models/`:
- `qwen2.5-coder-7b-instruct-q4_k_m.gguf` (4.4 GB)
- `Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf` (4.6 GB)

### Step 2: Dataset Verification

The 15 Go chaincode files are already in `02_resources/dataset/`:
- 9 vulnerable files (`vuln_01_time.go` ... `vuln_06_b_conditional_leak.go`)
- 6 benign trap files (`safe_01_logging.go` ... `safe_06_external_lib.go`)

Ground truth labels: see [`LABELING_CRITERIA.md`](LABELING_CRITERIA.md)

### Step 3: Run Local Model Audit (Table II in paper)

```bash
python scripts/02_run_audit_v3.py
```

Runs Qwen2.5-Coder-7B and Llama-3.1-8B under three prompting strategies (zero-shot, few-shot, CoT) on all 15 files.

- **Output**: `03_artifacts/raw_results/audit_v3_original_*.csv`
- **Expected time**: ~5 minutes (local GPU)
- **Prompt templates**: see [`PROMPTS.md`](PROMPTS.md) (P1, P2, P3)
- **Classifier logic**: see [`CLASSIFIER.md`](CLASSIFIER.md)

### Step 4: Obfuscation Experiment (Table III in paper)

```bash
# Generate obfuscated dataset
python scripts/03_obfuscate_dataset.py

# Run audit on obfuscated files
python scripts/02_run_audit_v3.py --dataset-dir 02_resources/dataset_obfuscated --tag obfuscated
```

- **Output**: `02_resources/dataset_obfuscated/` (15 files, 459 identifiers renamed)
- **Output**: `03_artifacts/raw_results/audit_v3_obfuscated_*.csv`

### Step 5: Semgrep Baseline (Section IV-B in paper)

```bash
python scripts/05_run_traditional_tools.py
```

Runs Semgrep with `p/security-audit` ruleset on all 15 files.

- **Output**: `03_artifacts/raw_results/traditional_tools_*.csv`
- **Expected result**: 0 consensus-layer findings, 1 generic `math-random-used` warning

### Step 6: Cloud Model Experiments (requires API keys)

```bash
# Claude 4.5 (Haiku, Sonnet, Opus)
python scripts/04_run_claude_audit.py

# Gemini 2.5 (Pro, Flash, Flash Lite)
python scripts/06_run_gemini_audit.py

# Cloud models with few-shot prompting
python scripts/12_run_cloud_fewshot.py
```

### Step 7: GoLiSA External Validation (Section IV-E in paper)

```bash
# Qwen on 657 GoLiSA files (zero-shot)
python scripts/08_run_golisa_validation.py

# Classifier v2 reclassification + ablation
python scripts/09_reclassify_and_ablation.py

# JSON mode micro-benchmark
python scripts/10_run_json_mode_microbenchmark.py

# Running_Examples: cloud models
python scripts/11_run_golisa_re_cloud.py

# Running_Examples: Llama
python scripts/13_run_golisa_re_llama.py
```

- **GoLiSA corpus**: 657 Go files in `02_resources/golisa_benchmark/Benchmark/`
- **Expected time**: ~87 minutes for Qwen on 657 files

### Step 8: Repeat Experiments (5 runs, Section IV-A in paper)

```bash
# Local models (Qwen + Llama) x 3 strategies x 15 files x 5 runs = 450
python scripts/14_local_repeat.py

# Claude 3 models x 3 strategies x 15 files x 5 runs = 675
python scripts/15_claude_repeat_cot.py

# Gemini 3 models x 3 strategies x 15 files x 5 runs = 675
python scripts/16_gemini_repeat_cot.py
```

- **Total**: 1,800 evaluations
- **Expected time**: Local ~57 min, Cloud ~3-4 hours total

## Output Files

All results are saved in `03_artifacts/raw_results/` as CSV files with accompanying `meta.json` files containing runtime metadata (total time, per-file times, errors).

## Verification

Compare your results against the paper's Table II (main results), Table III (obfuscation), and Table IV (GoLiSA Running_Examples). Due to LLM stochasticity at temperature 0.1, minor variations (±1 file TNR) may occur in cloud model results.

Local model results (Qwen TPR 9/9, TNR 6/6; Llama TPR 9/9, TNR 1/6) should be fully reproducible given identical model weights and inference parameters.
