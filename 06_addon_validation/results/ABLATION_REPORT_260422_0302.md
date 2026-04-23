# Controlled Annotation Ablation Report

> Generated: 260422_0302
> Runs: 5 | n_ctx: 16384 | temp: 0.1
> Filenames: P01.go~P15.go (neutral, seed=20260422)
> Vuln: 9 | Safe: 6

## Confound Control

| Variable | Value | Status |
|:---------|:------|:-------|
| n_ctx | 16384 | Matched |
| prompt | P1 zero-shot | Matched |
| temperature | 0.1 | Matched |
| runs | 5 | Matched |
| filenames | P01~P15 (neutral) | Controlled |
| **inline comments** | **ann vs abl** | **Experimental variable** |

## File Mapping

| Neutral | Original | Ground Truth |
|:--------|:---------|:-------------|
| P01.go | safe_01_logging.go | safe |
| P02.go | vuln_01_b_interprocedural.go | vulnerable |
| P03.go | safe_05_deterministic_time.go | safe |
| P04.go | safe_03_map_read.go | safe |
| P05.go | vuln_05_phantom.go | vulnerable |
| P06.go | vuln_04_map_iter.go | vulnerable |
| P07.go | vuln_06_iterator_leak.go | vulnerable |
| P08.go | vuln_01_time.go | vulnerable |
| P09.go | safe_06_external_lib.go | safe |
| P10.go | vuln_06_b_conditional_leak.go | vulnerable |
| P11.go | vuln_04_b_nested_map.go | vulnerable |
| P12.go | safe_04_math_rand.go | safe |
| P13.go | vuln_03_goroutine.go | vulnerable |
| P14.go | safe_02_local_var.go | safe |
| P15.go | vuln_02_global.go | vulnerable |

## LLAMA

### Annotated (comments)

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 9 | 6 | 0 | 0 | 1.00 | 0.00 |
| 2 | 9 | 6 | 0 | 0 | 1.00 | 0.00 |
| 3 | 9 | 6 | 0 | 0 | 1.00 | 0.00 |
| 4 | 9 | 6 | 0 | 0 | 1.00 | 0.00 |
| 5 | 9 | 5 | 0 | 1 | 1.00 | 0.17 |

- **TPR mean**: 1.0000 (range: 1.00-1.00)
- **TNR mean**: 0.0333 (range: 0.00-0.17)

### Ablated (no comments)

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 8 | 6 | 1 | 0 | 0.89 | 0.00 |
| 2 | 6 | 6 | 3 | 0 | 0.67 | 0.00 |
| 3 | 9 | 6 | 0 | 0 | 1.00 | 0.00 |
| 4 | 8 | 6 | 1 | 0 | 0.89 | 0.00 |
| 5 | 9 | 5 | 0 | 1 | 1.00 | 0.17 |

- **TPR mean**: 0.8889 (range: 0.67-1.00)
- **TNR mean**: 0.0333 (range: 0.00-0.17)

### Annotation Effect (llama)

- **TPR diff (ann - abl)**: +0.1111
- **TNR diff (ann - abl)**: +0.0000
- Interpretation: positive = comments help, negative = comments hurt


## QWEN

### Annotated (comments)

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 8 | 0 | 1 | 6 | 0.89 | 1.00 |
| 2 | 8 | 0 | 1 | 6 | 0.89 | 1.00 |
| 3 | 8 | 0 | 1 | 6 | 0.89 | 1.00 |
| 4 | 4 | 0 | 5 | 6 | 0.44 | 1.00 |
| 5 | 7 | 0 | 2 | 6 | 0.78 | 1.00 |

- **TPR mean**: 0.7778 (range: 0.44-0.89)
- **TNR mean**: 1.0000 (range: 1.00-1.00)

### Ablated (no comments)

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 4 | 3 | 5 | 3 | 0.44 | 0.50 |
| 2 | 6 | 2 | 3 | 4 | 0.67 | 0.67 |
| 3 | 4 | 3 | 5 | 3 | 0.44 | 0.50 |
| 4 | 3 | 3 | 6 | 3 | 0.33 | 0.50 |
| 5 | 6 | 4 | 3 | 2 | 0.67 | 0.33 |

- **TPR mean**: 0.5111 (range: 0.33-0.67)
- **TNR mean**: 0.5000 (range: 0.33-0.67)

### Annotation Effect (qwen)

- **TPR diff (ann - abl)**: +0.2667
- **TNR diff (ann - abl)**: +0.5000
- Interpretation: positive = comments help, negative = comments hurt

---

## Methodology Notes

- All confounds from prior D1-vs-D1clean comparison are eliminated:
  - n_ctx unified (was 4096 vs 16384)
  - prompt strategy unified (was 3-way vs zero-shot)
  - run count unified (was 1 vs 5)
  - filename leakage removed (was safe_/vuln_ prefix, now P##.go)
- The ONLY difference between conditions is presence/absence of inline comments.
- Classifier v2 (identical to scripts 08/20).
