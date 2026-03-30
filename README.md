# HLF Chaincode Vulnerability Detection with Local sLM

**Local Small Language Model-based Security Vulnerability Detection for Hyperledger Fabric Chaincode (Go)**

**Accepted** at **AMLDS 2026** (2nd International Conference on Advanced Machine Learning and Data Science, Osaka, Japan, July 21-23, 2026) | Paper ID: S2700

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Conference](https://img.shields.io/badge/AMLDS%202026-Accepted-brightgreen.svg)](https://amlds.site)

---

## Overview

This repository contains the experiment artifacts for detecting **consensus-layer non-determinism vulnerabilities** in Hyperledger Fabric (HLF) Go chaincode using locally deployed small language models (sLMs). Traditional static analysis tools (e.g., Semgrep) cannot detect HLF-specific semantic vulnerabilities such as non-deterministic operations that cause endorsement mismatch. We demonstrate that a 7B-parameter model running on a local GPU achieves detection accuracy comparable to cloud-based large language models, while preserving source code privacy and incurring zero API cost.

## Key Results

| Model | Type | TPR (9 vuln) | TNR (6 safe) | Avg Time/File |
|:------|:-----|:-------------|:-------------|:--------------|
| **Qwen2.5-Coder-7B** | Local (7B) | 9/9 (100%) | 6/6 (100%) | 3.94s |
| Llama-3.1-8B | Local (8B) | 9/9 (100%) | 1/6 (17%) | 10.09s |
| Claude Haiku 4.5 | Cloud | 9/9 (100%) | 5/6 (83%) | 12.89s |
| Claude Opus 4.5 | Cloud | 9/9 (100%) | 5/6 (83%) | 22.49s |
| Gemini 2.5 Pro | Cloud | 9/9 (100%) | 0/6 (0%) | 19.63s |
| Semgrep 1.151.0 | Static | 0/9 (0%) | — | — |

*Zero-shot prompt, original dataset, Classifier v2. Full results in `01_contexts/current/WORK_STATUS.md`.*

## Repository Structure

```
.
├── scripts/                        # 18 Python experiment scripts
│   ├── 01_download_models.py       # Model download (HuggingFace)
│   ├── 02_run_audit_v3.py          # Main audit: multi-prompt, multi-model
│   ├── 03_obfuscate_dataset.py     # Identifier obfuscation (459 replacements)
│   ├── 04_run_claude_audit.py      # Claude API audit
│   ├── 05_run_traditional_tools.py # Semgrep baseline
│   ├── 06_run_gemini_audit.py      # Gemini API audit
│   ├── 08_run_golisa_validation.py # GoLiSA 657-file external validation
│   ├── 09_reclassify_and_ablation.py # Classifier v2 + ablation studies
│   ├── 10_run_json_mode_microbenchmark.py # JSON mode evaluation
│   ├── 11~13_*.py                  # Cloud/Llama Running_Examples experiments
│   └── 14~17_*.py                  # Repeat + CoT experiments (1,800 runs)
│
├── 02_resources/
│   ├── dataset/                    # 15 Go chaincodes (vuln 9 + safe 6)
│   ├── dataset_obfuscated/         # 15 obfuscated Go files
│   ├── models/                     # .gguf files (excluded via .gitignore)
│   └── golisa_benchmark/           # 657 Go files from 326 GitHub repos
│       ├── Benchmark/              # Extracted .go files
│       └── api_research/           # GitHub/Zenodo API exploration data
│
├── 03_artifacts/
│   └── raw_results/                # CSV audit results + meta.json
│
├── 04_feedback/                    # Issue tracking (23 issues)
│
├── 01_contexts/                    # Session tracking, references
│
├── PROMPTS.md                      # All prompt templates (P1-P4) verbatim
├── CLASSIFIER.md                   # Classifier v1/v2/JSON logic + keyword lists
├── LABELING_CRITERIA.md            # Ground truth labels + consensus-relevant definition
├── PIPELINE_WORKFLOW.md            # Experiment pipeline description
├── requirements.txt                # Python dependencies (version-pinned)
├── REPRODUCTION.md                 # Step-by-step reproduction guide
├── CITATION.cff                    # Citation metadata
├── LICENSE                         # MIT License
└── .gitignore                      # Excludes models (9GB), VM images (22GB)
```

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

### Ground Truth & Labeling

Labeling criteria are documented in [`LABELING_CRITERIA.md`](LABELING_CRITERIA.md):
- 15 micro-benchmark files with manually assigned labels
- 5 GoLiSA Running_Examples with known vulnerability types
- 12 consensus-relevant keywords for Semgrep finding classification

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

See [`REPRODUCTION.md`](REPRODUCTION.md) for a step-by-step guide to reproduce all experiments.

## Citation

If you use this code or dataset, please cite:

```bibtex
@inproceedings{park2026local,
  title={Local Small Language Models for Consensus-Layer Vulnerability Detection in Hyperledger Fabric Chaincode: A Qualitative Feasibility Study},
  author={Park, Sungmoon and Jeong, Misook and Choi, Hoansuk and Yang, Jinhong},
  booktitle={Proceedings of the 2nd International Conference on Advanced Machine Learning and Data Science (AMLDS 2026)},
  year={2026},
  publisher={IEEE},
  address={Osaka, Japan}
}
```

## License

This project is licensed under the [MIT License](LICENSE). The GoLiSA benchmark files in `02_resources/golisa_benchmark/` are sourced from the GoLiSA project (Olivieri et al., ECOOP 2023) and retain their original licensing.

## Acknowledgments

This work was supported by the Institute of Information & Communications Technology Planning & Evaluation (IITP) grant (No. IITP-2026-RS-2024-00436773) and the Korea Institute for Advancement of Technology (KIAT) grant (No. P0026190).

---

# HLF 체인코드 취약점 탐지 (로컬 sLM)

**로컬 소형 언어모델(sLM) 기반 Hyperledger Fabric 체인코드(Go) 보안 취약점 자동 탐지**

**AMLDS 2026** (국제 응용 머신러닝 및 데이터 사이언스 학회, 일본 오사카, 2026년 7월 21-23일) **채택(Accepted)** | Paper ID: S2700

---

## 개요

본 저장소는 로컬 GPU에서 소형 언어모델(sLM)을 활용하여 Hyperledger Fabric(HLF) Go 체인코드의 **합의 계층 비결정성 취약점**을 탐지하는 실험 산출물을 포함합니다. 기존 정적 분석 도구(Semgrep 등)는 보증 불일치를 유발하는 HLF 특화 의미론적 취약점을 탐지하지 못합니다. 본 연구는 7B 파라미터 로컬 모델이 클라우드 대형 모델과 대등한 탐지 정확도를 달성하면서, 소스코드 프라이버시 보호와 API 비용 제로의 이점을 제공함을 실증합니다.

## 주요 결과

| 모델 | 유형 | TPR (취약 9개) | TNR (안전 6개) | 평균 시간/파일 |
|:------|:-----|:-------------|:-------------|:--------------|
| **Qwen2.5-Coder-7B** | 로컬 (7B) | 9/9 (100%) | 6/6 (100%) | 3.94초 |
| Llama-3.1-8B | 로컬 (8B) | 9/9 (100%) | 1/6 (17%) | 10.09초 |
| Claude Haiku 4.5 | 클라우드 | 9/9 (100%) | 5/6 (83%) | 12.89초 |
| Claude Opus 4.5 | 클라우드 | 9/9 (100%) | 5/6 (83%) | 22.49초 |
| Gemini 2.5 Pro | 클라우드 | 9/9 (100%) | 0/6 (0%) | 19.63초 |
| Semgrep 1.151.0 | 정적 도구 | 0/9 (0%) | — | — |

*Zero-shot 프롬프트, 원본 데이터셋, Classifier v2 기준. 전체 결과는 `01_contexts/current/WORK_STATUS.md` 참조.*

## 저장소 구조

```
.
├── scripts/                        # 18개 Python 실험 스크립트
│   ├── 01_download_models.py       # 모델 다운로드 (HuggingFace)
│   ├── 02_run_audit_v3.py          # 메인 감사: 다중 프롬프트, 다중 모델
│   ├── 03_obfuscate_dataset.py     # 식별자 난독화 (459개 치환)
│   ├── 04_run_claude_audit.py      # Claude API 감사
│   ├── 05_run_traditional_tools.py # Semgrep 베이스라인
│   ├── 06_run_gemini_audit.py      # Gemini API 감사
│   ├── 08_run_golisa_validation.py # GoLiSA 657개 파일 외부 검증
│   ├── 09_reclassify_and_ablation.py # Classifier v2 + 절제 연구
│   ├── 10_run_json_mode_microbenchmark.py # JSON 모드 평가
│   ├── 11~13_*.py                  # 클라우드/Llama Running_Examples 실험
│   └── 14~17_*.py                  # 반복 + CoT 실험 (1,800건)
│
├── 02_resources/
│   ├── dataset/                    # 15개 Go 체인코드 (취약 9 + 안전 6)
│   ├── dataset_obfuscated/         # 15개 난독화 Go 파일
│   ├── models/                     # .gguf 파일 (.gitignore로 제외)
│   └── golisa_benchmark/           # 326개 GitHub 저장소에서 추출한 657개 Go 파일
│       ├── Benchmark/              # 추출된 .go 파일
│       └── api_research/           # GitHub/Zenodo API 탐색 데이터
│
├── 03_artifacts/
│   └── raw_results/                # CSV 감사 결과 + meta.json
│
├── 04_feedback/                    # 이슈 추적 (23개 안건)
│
├── 01_contexts/                    # 세션 추적, 참고문헌
│
├── PROMPTS.md                      # 모든 프롬프트 템플릿 (P1-P4) 전문
├── CLASSIFIER.md                   # Classifier v1/v2/JSON 로직 + 키워드 리스트
├── LABELING_CRITERIA.md            # Ground truth 라벨 + consensus-relevant 정의
├── PIPELINE_WORKFLOW.md            # 실험 파이프라인 설명
├── requirements.txt                # Python 의존성 (버전 고정)
├── REPRODUCTION.md                 # 단계별 재현 가이드
├── CITATION.cff                    # 인용 메타데이터
├── LICENSE                         # MIT 라이선스
└── .gitignore                      # 모델(9GB), VM 이미지(22GB) 제외
```

## 재현성

### 프롬프트 전략

4가지 프롬프트 전략이 [`PROMPTS.md`](PROMPTS.md)에 전문 공개되어 있습니다:

| 프롬프트 | 설명 |
|:---------|:-----|
| P1: Zero-shot | 6개 취약점 카테고리, 구조화 출력 |
| P2: Few-shot | P1 + 예시 2개 (취약 vs 안전한 `time.Now()` 사용) |
| P3: Chain-of-Thought | 6단계 추론: PutState 역추적 |
| P4: JSON mode | `is_vulnerable` boolean 포함 구조화 JSON 출력 |

### 분류 로직

3가지 분류기가 [`CLASSIFIER.md`](CLASSIFIER.md)에 전체 코드와 함께 공개되어 있습니다:

| 분류기 | 핵심 기능 |
|:-------|:---------|
| v1 (원본) | Safe 문구 조기 반환 + 모순 검사 |
| v2 (개선) | 자기모순 감지: 구조적 증거가 safe 문구보다 우선 |
| JSON 파서 | `is_vulnerable` 필드 파싱, 실패 시 v2 폴백 |

### Ground Truth 및 라벨링

라벨링 기준이 [`LABELING_CRITERIA.md`](LABELING_CRITERIA.md)에 공개되어 있습니다:
- 15개 마이크로벤치마크 파일의 수동 라벨
- 5개 GoLiSA Running_Examples의 취약점 유형
- 12개 consensus-relevant 키워드 (Semgrep 결과 분류용)

## 하드웨어

| 구성 요소 | 사양 |
|:----------|:-----|
| GPU | NVIDIA GeForce RTX 3090 Ti (24564 MiB VRAM) |
| CUDA | 13.0 (V13.0.88) |
| Python | 3.11.9 |
| llama-cpp-python | 0.3.16 (CUDA 빌드) |
| Semgrep | 1.151.0 |

## 모델 (저장소에 미포함)

| 모델 | 파일 | 크기 | 출처 |
|:------|:-----|:-----|:-----|
| Qwen2.5-Coder-7B-Instruct | Q4_K_M.gguf | 4.4 GB | HuggingFace |
| Meta-Llama-3.1-8B-Instruct | Q4_K_M.gguf | 4.6 GB | HuggingFace |

`scripts/01_download_models.py`로 다운로드 가능합니다.

## 재현

모든 실험의 단계별 재현 가이드는 [`REPRODUCTION.md`](REPRODUCTION.md)를 참조하세요.

## 인용

본 코드 또는 데이터셋을 사용하시는 경우, 위 영문 섹션의 BibTeX를 인용해 주세요.

## 라이선스

본 프로젝트는 [MIT 라이선스](LICENSE)로 공개됩니다. `02_resources/golisa_benchmark/`의 GoLiSA 벤치마크 파일은 GoLiSA 프로젝트(Olivieri et al., ECOOP 2023)에서 가져온 것이며, 원본 라이선스를 따릅니다.

## 감사의 글

본 연구는 정보통신기획평가원(IITP) 지원사업(No. IITP-2026-RS-2024-00436773)과 한국산업기술진흥원(KIAT) 지원사업(No. P0026190)의 지원을 받았습니다.
