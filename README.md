# HLF Chaincode Vulnerability Detection with Local sLM

**Privacy-Preserving Anomaly Detection in Hyperledger Fabric Chaincode Using Compact Local Transformer Models**

Submitted to **AANN 2026** (6th International Conference on Advanced Algorithms and Neural Networks, Qingdao, China, August 7-9, 2026) | Paper No: M7VNBFDSWP

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Overview

This repository contains experiment artifacts for detecting **endorsement-nondeterminism vulnerabilities** in Hyperledger Fabric (HLF) Go chaincode using a locally deployed compact transformer model (Qwen2.5-Coder-7B, 4-bit quantized) alongside custom Semgrep rules. The 464-file benchmark is derived from the GoLiSA corpus with content-hash deduplication and dual-annotator verification.

## Key Results (464-File GoLiSA Benchmark)

| Method | TPR | TNR | Prec. | F1 |
|:-------|:----|:----|:------|:---|
| Qwen2.5-Coder-7B (det.) | 80.6% (25/31) | 48.6% (210/432) | 10.1% | 18.0% |
| Semgrep (5 custom rules) | 77.4% (24/31) | 99.3% (429/432) | 88.9% | 82.8% |
| Majority vote (5 seeds) | 90.6% (29/32) | 45.8% (198/432) | 11.0% | 19.7% |
| OR-Union | 96.8% (30/31) | 48.4% (209/432) | 11.9% | 21.1% |

*Deterministic results on 463 common files (31V, 432S). Majority on 464 files (32V, 432S). Union on 463 common files.*

### Supplementary: 15-File Diagnostic Benchmark

| Model | Type | TPR (9 vuln) | TNR (6 safe) | Avg Time/File |
|:------|:-----|:-------------|:-------------|:--------------|
| **Qwen2.5-Coder-7B** | Local (7B) | 9/9 (100%) | 6/6 (100%) | 3.94s |
| Llama-3.1-8B | Local (8B) | 9/9 (100%) | 1/6 (17%) | 10.09s |
| Claude Haiku 4.5 | Cloud | 9/9 (100%) | 5/6 (83%) | 12.89s |
| Gemini 2.5 Pro | Cloud | 9/9 (100%) | 0/6 (0%) | 19.63s |

## Repository Structure

```
.
├── scripts/                        # 29 experiment scripts
│   ├── 01_download_models.py       # Model download (HuggingFace)
│   ├── 02_run_audit_v3.py          # Multi-prompt, multi-model audit
│   ├── 03_obfuscate_dataset.py     # Identifier obfuscation (459 replacements)
│   ├── 04~06_*.py                  # Cloud API audits (Claude, Gemini)
│   ├── 07~09_*.py                  # GoLiSA validation, reclassification
│   ├── 10~19_*.py                  # Microbenchmark, repeat, CoT experiments
│   ├── 20_run_addon_validation.py  # Addon dataset validation (D1/D2)
│   ├── 21_run_annotation_ablation.py # Annotation ablation study
│   ├── 22_mine_golisa_candidates.py  # GoLiSA positive/negative mining
│   ├── 23_prompt_dev_sanity_check.py # Prompt development sanity check
│   ├── 24_run_golisa_labeling.py   # 464-file annotation pipeline
│   ├── 25_run_second_annotation.py # Independent second annotation
│   ├── 26_run_main_experiment.py   # Phase 6: deterministic evaluation
│   ├── 27_run_robustness.py        # Phase 7: 5-seed robustness (v2.0)
│   └── strip_go_comments.go        # Go comment stripper (source)
│
├── 02_resources/
│   ├── dataset/                    # 15 Go chaincodes (vuln 9 + safe 6)
│   ├── dataset_obfuscated/         # 15 obfuscated Go files
│   ├── models/                     # .gguf files (excluded via .gitignore)
│   └── golisa_benchmark/           # 657 Go files from 326 GitHub repos
│
├── 06_addon_validation/            # 464-file benchmark pipeline
│   ├── benchmark/                  # BENCHMARK_FREEZE.json (ground truth)
│   │                               # INFERENCE_CONTRACT.md (parameters)
│   ├── dataset/                    # 17 addon .go files
│   ├── dataset_d1_clean/           # 15 annotation-stripped .go files
│   ├── dataset_ablation_*/         # Ablation datasets (ann/abl)
│   ├── golisa_mining/              # Candidate mining data
│   ├── labeling/                   # Primary + secondary annotation data
│   │   ├── run_260422_2142/        # Primary annotation (per-file JSON)
│   │   ├── second_260423_*/        # Second annotation runs
│   │   └── verification/           # Manual verification results
│   ├── experiment/
│   │   ├── main_260423_0047/       # Phase 6 deterministic (463 per-file)
│   │   └── robustness_260423_0341/ # Phase 7 robustness (464 x 5 seeds)
│   └── results/                    # Summary CSVs and reports
│
├── 03_artifacts/raw_results/       # CSV audit results + meta.json
├── 04_feedback/                    # Issue tracking
├── 01_contexts/                    # Session tracking, references
│
├── rules/hlf_consensus.yml         # 5 custom Semgrep rules for HLF
├── PROMPTS.md                      # Prompt templates (P1-P4) verbatim
├── CLASSIFIER.md                   # Classifier v1/v2/JSON logic
├── LABELING_CRITERIA.md            # Ground truth labels + criteria
├── PIPELINE_WORKFLOW.md            # Experiment pipeline description
├── REPRODUCTION.md                 # Step-by-step reproduction guide
├── requirements.txt                # Python dependencies (version-pinned)
├── CITATION.cff                    # Citation metadata
├── LICENSE                         # MIT License
└── .gitignore                      # Excludes models (9GB), VM images (22GB)
```

## Benchmark Construction (464-File)

The benchmark is derived from the GoLiSA corpus (657 files, 326 repos):
1. Remove files with insufficient chaincode structure: 657 -> 618
2. Content-hash (SHA-256) deduplication: 618 -> 464 (154 duplicates removed)
3. Primary annotation (Claude Sonnet 4.5): 46 initial positives -> 33 after manual verification
4. Second annotation (Claude Opus 4.5): Cohen's kappa = 0.766 (substantial)
5. Final benchmark: 32 vulnerable, 432 safe

Ground truth labels: [`06_addon_validation/benchmark/BENCHMARK_FREEZE.json`](06_addon_validation/benchmark/BENCHMARK_FREEZE.json)

## Reproducibility

### Prompt Strategies

Four prompt strategies are documented in [`PROMPTS.md`](PROMPTS.md):

| Prompt | Description |
|:-------|:------------|
| P1: Zero-shot | 6 vulnerability categories, structured output |
| P2: Few-shot | P1 + 2 examples (vulnerable vs. safe `time.Now()` usage) |
| P3: Chain-of-Thought | 6-step reasoning: PutState backward tracing |
| P4: JSON mode | Structured JSON output with `is_vulnerable` boolean |

### Classification Logic

Three classifiers are documented in [`CLASSIFIER.md`](CLASSIFIER.md):

| Classifier | Key Feature |
|:-----------|:------------|
| v1 (original) | Safe-phrase early return with contradiction check |
| v2 (improved) | Self-contradiction detection: structured evidence overrides safe phrase |
| JSON parser | Parses `is_vulnerable` field, falls back to v2 |

### Six Targeted Vulnerability Classes

| Class | Description |
|:------|:------------|
| C1 | Nondeterministic timestamps (`time.Now()`) |
| C2 | Goroutine concurrency |
| C3 | Map-iteration randomness |
| C4 | Phantom reads (`GetQueryResult`) |
| C5 | Iterator resource leaks (auxiliary) |
| C6 | Global mutable state |

## Hardware

| Component | Specification |
|:----------|:-------------|
| GPU | NVIDIA GeForce RTX 3090 Ti (24564 MiB VRAM) |
| CUDA | 13.0 (V13.0.88) |
| Python | 3.11.9 |
| llama-cpp-python | 0.3.16 (CUDA build) |
| Semgrep | 1.151.0 |

## Models (not included in repo)

| Model | File | Size | Source |
|:------|:-----|:-----|:-------|
| Qwen2.5-Coder-7B-Instruct | Q4_K_M.gguf | 4.4 GB | HuggingFace |
| Meta-Llama-3.1-8B-Instruct | Q4_K_M.gguf | 4.6 GB | HuggingFace |

Download via `scripts/01_download_models.py`.

## Reproduction

See [`REPRODUCTION.md`](REPRODUCTION.md) for a step-by-step guide.

## Citation

```bibtex
@inproceedings{park2026privacy,
  title={Privacy-Preserving Anomaly Detection in Hyperledger Fabric Chaincode Using Compact Local Transformer Models},
  author={Park, Sungmoon and Yang, Jinhong},
  booktitle={Proceedings of the 6th International Conference on Advanced Algorithms and Neural Networks (AANN 2026)},
  year={2026},
  publisher={IEEE},
  address={Qingdao, China}
}
```

## License

This project is licensed under the [MIT License](LICENSE). The GoLiSA benchmark files in `02_resources/golisa_benchmark/` are sourced from the GoLiSA project (Olivieri et al., ECOOP 2023) and retain their original licensing.

## Acknowledgments

This work was supported by the Korea Institute for Advancement of Technology (KIAT) grant funded by the Korea government (Ministry of Trade, Industry and Energy) through the International Cooperation in Industrial Technology program (Project Number: P0026190).
