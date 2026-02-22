# Classifier Logic

> Reproducibility reference for all response classifiers used in the experiment.
> Each classifier is copied verbatim from the source code with file path and line numbers.

## Overview

Three classifiers are used to convert free-text LLM responses into binary labels (`vulnerable` / `safe`):

| Classifier | Source | Purpose |
|:-----------|:-------|:--------|
| v1 (original) | `scripts/09_reclassify_and_ablation.py`, lines 43-76 | Initial classification used in `08_run_golisa_validation.py` |
| v2 (improved) | `scripts/09_reclassify_and_ablation.py`, lines 82-151 | Self-contradiction detection; used for all reported results |
| JSON parser | `scripts/09_reclassify_and_ablation.py`, lines 348-368 | Parses structured JSON responses from P4 (json_mode) |

---

## Classifier v1 (Original)

Source: `scripts/09_reclassify_and_ablation.py`, lines 43-76
Also deployed in: `scripts/08_run_golisa_validation.py`, lines 152-209

```python
def classify_v1(response: str) -> str:
    """Original classifier -- safe phrase triggers early return."""
    if not response or response.startswith("ERROR:"):
        return "error"
    resp_lower = response.lower()

    safe_indicators = [
        "no vulnerabilities detected", "no vulnerabilities found",
        "no security vulnerabilities", "no significant vulnerabilities",
        "no vulnerabilities were found", "no vulnerabilities were detected",
        "no critical vulnerabilities", "the code appears to be secure",
        "the code is secure", "no issues found", "no issues detected",
    ]
    for indicator in safe_indicators:
        if indicator in resp_lower:
            idx = resp_lower.find(indicator)
            after = resp_lower[idx + len(indicator):]
            if any(kw in after[:200] for kw in ["however", "but ", "although", "note that"]):
                vuln_check = ["vulnerability", "vulnerable", "severity:", "recommended fix"]
                if sum(1 for v in vuln_check if v in after) >= 2:
                    return "vulnerable"
            return "safe"

    vuln_indicators = [
        "vulnerability type", "severity:", "severity :",
        "recommended fix", "affected code", "critical", "high",
        "non-deterministic", "nondeterministic", "phantom read",
        "global variable", "goroutine", "race condition",
        "map iteration", "iterator leak", "putstate",
    ]
    vuln_count = sum(1 for ind in vuln_indicators if ind in resp_lower)
    if vuln_count >= 2:
        return "vulnerable"
    return "vulnerable"
```

### v1 Decision Flow

1. Empty or `ERROR:` prefix -> `error`
2. Scan for 11 safe phrases (case-insensitive substring match)
   - If found: check next 200 chars for hedge words ("however", "but ", "although", "note that")
     - If hedge + 2 or more vulnerability evidence keywords -> `vulnerable` (contradiction)
     - Otherwise -> `safe`
3. Count 16 vulnerability indicator keywords
   - If 2 or more found -> `vulnerable`
4. Default -> `vulnerable` (conservative)

### v1 Limitation

When an LLM response contains detailed vulnerability analysis (severity ratings, recommended fixes) **followed by** a generic "No vulnerabilities detected" conclusion, v1 classifies it as `safe` because the safe phrase is matched first. This is the self-contradiction problem that v2 addresses.

---

## Classifier v2 (Improved)

Source: `scripts/09_reclassify_and_ablation.py`, lines 82-151

```python
def classify_v2(response: str) -> str:
    """Improved classifier -- detects self-contradictory responses.

    Rationale (for paper):
      LLM responses can contain detailed vulnerability analysis followed
      by a contradictory 'No vulnerabilities detected' conclusion. This
      is a known LLM consistency issue. When structured vulnerability
      evidence (severity ratings, recommended fixes, affected code
      locations) co-occurs with a safe phrase, the structured evidence
      takes priority, as it reflects the model's analytical output
      rather than a boilerplate concluding statement.
    """
    if not response or response.startswith("ERROR:"):
        return "error"
    resp_lower = response.lower()

    # --- Step 1: Count structured vulnerability evidence ---
    structured_markers = [
        "severity:", "severity :",
        "recommended fix", "suggested fix",
        "affected code", "code location",
    ]
    struct_count = sum(1 for m in structured_markers if m in resp_lower)

    # --- Step 2: Count HLF consensus-layer nondeterminism keywords ---
    hlf_nondeterminism = [
        "non-deterministic", "nondeterministic",
        "global variable", "mutable global", "shared state",
        "goroutine", "race condition", "concurrent",
        "map iteration", "iteration order",
        "time.now", "gettxtimestamp",
        "channel",  # Go channel in goroutine context
    ]
    hlf_count = sum(1 for kw in hlf_nondeterminism if kw in resp_lower)

    # --- Step 3: Check safe indicators ---
    safe_indicators = [
        "no vulnerabilities detected", "no vulnerabilities found",
        "no security vulnerabilities", "no significant vulnerabilities",
        "no vulnerabilities were found", "no vulnerabilities were detected",
        "no critical vulnerabilities", "the code appears to be secure",
        "the code is secure", "no issues found", "no issues detected",
    ]
    has_safe_phrase = any(ind in resp_lower for ind in safe_indicators)

    # --- Step 4: Decision logic ---
    # Case A: Structured vuln analysis EXISTS + safe phrase EXISTS
    #   -> Self-contradictory response -> trust the analysis
    if struct_count >= 2 and has_safe_phrase:
        return "vulnerable"

    # Case B: Safe phrase only, no structured analysis
    #   -> Genuine safe judgment
    if has_safe_phrase and struct_count < 2:
        return "safe"

    # Case C: No safe phrase, check vuln indicators
    vuln_indicators = [
        "vulnerability type", "severity:", "severity :",
        "recommended fix", "affected code", "critical", "high",
        "non-deterministic", "nondeterministic", "phantom read",
        "global variable", "goroutine", "race condition",
        "map iteration", "iterator leak", "putstate",
    ]
    vuln_count = sum(1 for ind in vuln_indicators if ind in resp_lower)
    if vuln_count >= 2:
        return "vulnerable"

    # Default: conservative
    return "vulnerable"
```

### v2 Decision Flow

1. Empty or `ERROR:` prefix -> `error`
2. Count structured vulnerability evidence (6 markers)
3. Count HLF consensus-layer keywords (13 keywords, counted but used for future reference; `hlf_count` is not directly used in decision logic in this version)
4. Check for safe phrase presence (11 phrases)
5. Decision:
   - **Case A**: structured evidence >= 2 AND safe phrase present -> `vulnerable` (self-contradiction resolved)
   - **Case B**: safe phrase present AND structured evidence < 2 -> `safe` (genuine safe)
   - **Case C**: no safe phrase, vulnerability indicators >= 2 -> `vulnerable`
6. Default -> `vulnerable` (conservative)

### v1 vs v2 Reclassification Impact

Source: Stage 1 output of `scripts/09_reclassify_and_ablation.py`

Applied to GoLiSA Qwen 657-file CSV:
- v1: 380 vulnerable / 277 safe
- v2: 477 vulnerable / 180 safe
- Changed classifications: 97 (all safe->vulnerable, self-contradiction cases)

---

## JSON Classifier

Source: `scripts/09_reclassify_and_ablation.py`, lines 348-368
Also: `scripts/10_run_json_mode_microbenchmark.py`, lines 66-82

```python
def classify_json_response(response: str) -> str:
    """Parse JSON response and classify."""
    if not response or response.startswith("ERROR:"):
        return "error"
    try:
        resp = response.strip()
        start = resp.find("{")
        end = resp.rfind("}")
        if start >= 0 and end > start:
            json_str = resp[start:end+1]
            data = json.loads(json_str)
            if data.get("is_vulnerable", False):
                return "vulnerable"
            else:
                return "safe"
    except (json.JSONDecodeError, KeyError, TypeError):
        pass
    # Fallback to v2 classifier
    return classify_v2(response)
```

### JSON Classifier Decision Flow

1. Empty or `ERROR:` prefix -> `error`
2. Extract substring from first `{` to last `}` in response
3. Parse as JSON, read `is_vulnerable` field
   - `true` -> `vulnerable`
   - `false` -> `safe`
4. If JSON parsing fails -> fallback to Classifier v2

Note: The version in `10_run_json_mode_microbenchmark.py` returns `"unknown"` instead of falling back to v2. In practice, all 15 micro-benchmark responses parsed successfully as JSON, so the fallback was never triggered.

---

## Keyword Lists Summary

### Safe Phrases (11, shared across v1/v2)

```
no vulnerabilities detected
no vulnerabilities found
no security vulnerabilities
no significant vulnerabilities
no vulnerabilities were found
no vulnerabilities were detected
no critical vulnerabilities
the code appears to be secure
the code is secure
no issues found
no issues detected
```

### Vulnerability Indicators (16, shared across v1/v2 Case C)

```
vulnerability type
severity:
severity :
recommended fix
affected code
critical
high
non-deterministic
nondeterministic
phantom read
global variable
goroutine
race condition
map iteration
iterator leak
putstate
```

### Structured Markers (6, v2 only)

```
severity:
severity :
recommended fix
suggested fix
affected code
code location
```

### HLF Nondeterminism Keywords (13, v2 only)

```
non-deterministic
nondeterministic
global variable
mutable global
shared state
goroutine
race condition
concurrent
map iteration
iteration order
time.now
gettxtimestamp
channel
```
