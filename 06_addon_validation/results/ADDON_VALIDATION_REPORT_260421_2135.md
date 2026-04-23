# Add-on External Validation Report

> Generated: 2026-04-21 21:35
> Files: 17 | Runs: 5 | Models: qwen, llama
> n_ctx: 16384 | temp: 0.1

---

## Ground Truth

- Vulnerable: 4 (U03, U13, U18, U20)
- Safe: 13
- Total: 17

## Semgrep Results

| Tool | TP | FP | FN | TN | TPR | TNR |
|:-----|:---|:---|:---|:---|:----|:----|
| semgrep_auto | 0 | 0 | 4 | 13 | 0.00 (0.00-0.60) | 1.00 (0.75-1.00) |
| semgrep_security-audit | 0 | 0 | 4 | 13 | 0.00 (0.00-0.60) | 1.00 (0.75-1.00) |
| semgrep_hlf | 0 | 0 | 4 | 13 | 0.00 (0.00-0.60) | 1.00 (0.75-1.00) |


## QWEN P1 Zero-Shot Results

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 3 | 8 | 1 | 5 | 0.75 | 0.38 |
| 2 | 2 | 6 | 2 | 7 | 0.50 | 0.54 |
| 3 | 3 | 9 | 1 | 4 | 0.75 | 0.31 |
| 4 | 2 | 8 | 2 | 5 | 0.50 | 0.38 |
| 5 | 2 | 9 | 2 | 4 | 0.50 | 0.31 |

- **TPR mean**: 0.60 (range: 0.50-0.75)
- **TNR mean**: 0.38 (range: 0.31-0.54)
- **TPR 95% CI** (Run 1): 0.19-0.99
- **TNR 95% CI** (Run 1): 0.14-0.68

### Sensitivity Analysis (U03=safe)
- TPR mean: 0.60
- TNR mean: 0.39

## LLAMA P1 Zero-Shot Results

| Run | TP | FP | FN | TN | TPR | TNR |
|:----|:---|:---|:---|:---|:----|:----|
| 1 | 4 | 13 | 0 | 0 | 1.00 | 0.00 |
| 2 | 4 | 12 | 0 | 1 | 1.00 | 0.08 |
| 3 | 3 | 11 | 1 | 2 | 0.75 | 0.15 |
| 4 | 3 | 12 | 1 | 1 | 0.75 | 0.08 |
| 5 | 4 | 12 | 0 | 1 | 1.00 | 0.08 |

- **TPR mean**: 0.90 (range: 0.75-1.00)
- **TNR mean**: 0.08 (range: 0.00-0.15)
- **TPR 95% CI** (Run 1): 0.40-1.00
- **TNR 95% CI** (Run 1): 0.00-0.25

### Sensitivity Analysis (U03=safe)
- TPR mean: 0.87
- TNR mean: 0.07

---

## Notes

- Classifier: v2 (identical to 08_run_golisa_validation.py)
- This is a **limited external validation** with small positive class (V=4).
- Wide TPR CIs reflect positive class scarcity, not model failure.
- Results should be framed as descriptive evidence, not precise TPR estimation.
