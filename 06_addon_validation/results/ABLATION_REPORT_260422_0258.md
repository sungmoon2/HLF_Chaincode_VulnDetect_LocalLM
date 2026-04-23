# Controlled Annotation Ablation Report

> Generated: 260422_0258
> Runs: 1 | n_ctx: 16384 | temp: 0.1
> Filenames: P01.go~P15.go (neutral, seed=20260422)
> Vuln: 9 | Safe: 6

## Confound Control

| Variable | Value | Status |
|:---------|:------|:-------|
| n_ctx | 16384 | Matched |
| prompt | P1 zero-shot | Matched |
| temperature | 0.1 | Matched |
| runs | 1 | Matched |
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

## QWEN

### Annotated (comments)

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 8 | 0 | 1 | 6 | 0.89 | 1.00 |

- **TPR mean**: 0.8889 (range: 0.89-0.89)
- **TNR mean**: 1.0000 (range: 1.00-1.00)

### Ablated (no comments)

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 6 | 3 | 3 | 3 | 0.67 | 0.50 |

- **TPR mean**: 0.6667 (range: 0.67-0.67)
- **TNR mean**: 0.5000 (range: 0.50-0.50)

### Annotation Effect (qwen)

- **TPR diff (ann - abl)**: +0.2222
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
