# GoLiSA Full-Corpus Benchmark — Freeze Record
> Version: 1.0 | Freeze Date: 2026-04-23 | Status: **FROZEN**

## Composition

| Category | Count | Description |
|:---------|:------|:------------|
| VULNERABLE | 32 | Verified by code-level review (Agent + human 2nd pass) |
| SAFE | 432 | Content-hash deduplicated, includes hard negatives |
| **Total** | **464** | Unique files after SHA-256 dedup |
| EXCLUDE | 4 | Non-chaincode or partial (not in benchmark) |
| Dev excluded | 31 | 10 repos reserved for sanity checking |
| Content duplicates removed | 154 | 57 groups, 0 VULNERABLE lost (1 V deduped: einai=ashokcin1) |

## Source

- **Corpus**: GoLiSA public benchmark (Casado-Vara et al., 2024)
- **Raw files**: 652 .go files across 325 repos
- **After dev exclusion**: 621 files
- **After labeling**: 622 files (621 + 1 Running_Examples)
- **After EXCLUDE removal**: 618 usable files
- **After content-hash dedup**: 464 unique files

## Labeling

| Item | Value |
|:-----|:------|
| Model | Claude Opus 4.5 (claude-opus-4-5@20251101) via Vertex AI |
| Temperature | 0.0 |
| Prompt | Rubric v2.0 (GOLISA_TAXONOMY_RUBRIC_v2.md) |
| Input | Comment-stripped Go source (strip_go_comments.exe) |
| Output | VERDICT / PRIMARY_CLASS / SECONDARY_CLASS / AUXILIARY_C5 / EVIDENCE_LINES / RATIONALE |
| Run | run_260422_2142/ |
| Errors | 0 |
| Strip failures | 0 |

## Verification

| Scope | Method | Result |
|:------|:-------|:-------|
| V 46 files (all labeled VULNERABLE) | 3+1 Agent code review + human 2nd pass | 33 TRUE V, 13 FP |
| Positive→SAFE 70 files | 4 Agent code review + human 2nd pass | 70 SAFE confirmed, 0 FN |
| 40 HIGH RISK (C4+Write same file) | Same-function API co-location analysis | 2 cases checked, both SAFE |
| FP 13 files | Reclassified V→S in per_file JSON | Documented in verification/false_positives.json |

## Vulnerability Family Distribution

| Class | Count | Reportability |
|:------|:------|:-------------|
| C1 TIME_NOW | 22 | Inferential (n≥5) |
| C6 GLOBAL_MUTABLE_STATE | 4 | Descriptive (n=3~4) |
| C3 MAP_ITERATION | 3 | Descriptive (n=3~4) |
| C4 NON_REVALIDATED_QUERY | 2 | Anecdotal (n<3) |
| C2 GOROUTINE | 1 | Anecdotal (n<3) |

C1 dominance (69%) reflects GoLiSA corpus composition, not vulnerability prevalence in the wild.

## Token Length

- Max input tokens: 12,310
- n_ctx limit: 16,384
- Truncation: **0 files** (all fit within context window)

## Integrity

- Content dedup: SHA-256 first 16 hex chars
- V dedup: einai/main.go removed (identical content to ashokcin1/main.go, both C1 VULNERABLE)
- No VULNERABLE content was lost — both copies had same verdict

## Files

- `BENCHMARK_FREEZE.json`: Machine-readable freeze with per-file details
- `BENCHMARK_FREEZE.md`: This document (human-readable)
- `../labeling/run_260422_2142/per_file/`: Per-file labeling results (622 JSON)
- `../labeling/verification/`: Verification records

## Freeze Declaration

**This benchmark is FROZEN as of 2026-04-23.**
No additions, removals, or relabeling permitted after this point.
Any corrections must be documented as errata with justification.
