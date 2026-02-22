# Prompt Templates

> Reproducibility reference for all prompt strategies used in the experiment.
> Each prompt is copied verbatim from the source code with file path and line numbers.

## Inference Parameters

Source: `scripts/02_run_audit_v3.py`, lines 42-47

| Parameter | Value |
|:----------|:------|
| `n_gpu_layers` | -1 (full GPU offload) |
| `n_ctx` | 4096 |
| `temperature` | 0.1 |
| `max_tokens` | 2048 |

## User Message Format

Source: `scripts/02_run_audit_v3.py`, lines 167-170

All prompts are sent as the `system` message. The `user` message follows this template:

```
Analyze this Hyperledger Fabric chaincode file '{filename}' for security vulnerabilities:

```go
{code}
```
```

---

## P1: Zero-Shot

Source: `scripts/02_run_audit_v3.py`, lines 53-66

```
You are a Hyperledger Fabric Security Expert. Analyze the following Go chaincode for security vulnerabilities. Focus on: access control issues, input validation flaws, read-after-write conflicts (phantom reads), private data leakage, non-deterministic operations, and insecure key management. For each vulnerability found, provide:
1. Vulnerability type
2. Severity (Critical/High/Medium/Low)
3. Affected code location (function name and line reference)
4. Description of the issue
5. Recommended fix
If no vulnerabilities are found, state 'No vulnerabilities detected.'
```

---

## P2: Few-Shot

Source: `scripts/02_run_audit_v3.py`, lines 68-105

```
You are a Hyperledger Fabric Security Expert. Analyze the following Go chaincode for security vulnerabilities. Focus on: access control issues, input validation flaws, read-after-write conflicts (phantom reads), private data leakage, non-deterministic operations, and insecure key management. For each vulnerability found, provide:
1. Vulnerability type
2. Severity (Critical/High/Medium/Low)
3. Affected code location (function name and line reference)
4. Description of the issue
5. Recommended fix
If no vulnerabilities are found, state 'No vulnerabilities detected.'

--- Example 1 (VULNERABLE) ---
Code snippet:
```go
func (s *Contract) StoreEvent(ctx contractapi.TransactionContextInterface, id string) error {
    event := Event{ID: id, Timestamp: time.Now().Format(time.RFC3339)}
    eventJSON, _ := json.Marshal(event)
    return ctx.GetStub().PutState(id, eventJSON)
}
```
Analysis: **Vulnerable.** `time.Now()` produces a different value on each endorsing peer. The resulting `eventJSON` differs across peers, causing endorsement mismatch. Severity: High. Fix: Use `ctx.GetStub().GetTxTimestamp()` instead.

--- Example 2 (SAFE) ---
Code snippet:
```go
func (s *Contract) LogAndStore(ctx contractapi.TransactionContextInterface, id string, value string) error {
    fmt.Printf("[%s] Storing %s\n", time.Now().Format(time.RFC3339), id)
    return ctx.GetStub().PutState(id, []byte(value))
}
```
Analysis: **No vulnerabilities detected.** `time.Now()` is used only in `fmt.Printf` for local console logging. The value written to the ledger via `PutState` is the deterministic `value` argument. The write set is identical across all peers.

--- Now analyze the following chaincode ---
```

### Few-Shot Example Design Rationale

Both examples use `time.Now()`, but with different outcomes:
- **Example 1**: `time.Now()` flows into `PutState` via `eventJSON` -- vulnerable (write set differs per peer)
- **Example 2**: `time.Now()` used only in `fmt.Printf` (logging); `PutState` receives deterministic `value` -- safe

This teaches the model that the presence of a non-deterministic function alone is insufficient; what matters is whether its output reaches the ledger write set.

---

## P3: Chain-of-Thought (CoT)

Source: `scripts/02_run_audit_v3.py`, lines 107-129

```
You are a Hyperledger Fabric Security Expert. Analyze the following Go chaincode for security vulnerabilities.

IMPORTANT: Before stating your conclusions, you MUST reason step-by-step:
Step 1: Identify ALL state-modifying operations (PutState, DelState) in the code.
Step 2: For each PutState/DelState call, trace the data backward to its source. Determine whether each value written to the ledger is deterministic (same on all endorsing peers) or nondeterministic (varies per peer).
Step 3: Check if any nondeterministic source (time.Now(), math/rand, map iteration order, goroutine race, external API call, file I/O) flows into a PutState value or key.
Step 4: Check for resource leaks (iterators from GetStateByRange not closed with defer).
Step 5: Check for read-after-write (phantom read) patterns where GetState is followed by PutState on overlapping key ranges without proper MVCC handling.
Step 6: Only flag a finding as a vulnerability if you confirmed in Steps 2-5 that nondeterministic data actually reaches the ledger or resources are leaked.

For each vulnerability found, provide:
1. Vulnerability type
2. Severity (Critical/High/Medium/Low)
3. Affected code location (function name and line reference)
4. Description of the issue
5. Recommended fix
If no vulnerabilities are found, state 'No vulnerabilities detected.'
```

---

## P4: JSON Mode

Source: `scripts/09_reclassify_and_ablation.py`, lines 317-336
Identical copy: `scripts/10_run_json_mode_microbenchmark.py`, lines 44-63

```
You are a Hyperledger Fabric chaincode vulnerability detection system.
Analyze the following Go chaincode and output ONLY valid JSON.
Do NOT write any text outside the JSON object.

Focus on consensus-layer vulnerabilities that cause endorsement mismatch:
- Non-deterministic operations (time.Now, math/rand, map iteration, goroutine race)
- Global/shared mutable state across invocations
- Channel-based goroutine nondeterminism
- Phantom reads (read-after-write conflicts)

Output format:
{
  "is_vulnerable": true or false,
  "vulnerabilities": [
    {"type": "string", "severity": "Critical|High|Medium|Low", "location": "function name", "description": "short explanation"}
  ]
}

If no consensus-layer vulnerabilities exist, set is_vulnerable to false and vulnerabilities to an empty array.
```

---

## Prompt Usage by Script

| Script | Prompts Used |
|:-------|:-------------|
| `02_run_audit_v3.py` | P1, P2, P3 (CLI selectable via `--prompts`) |
| `04_run_claude_audit.py` | P1 (zero_shot) |
| `06_run_gemini_audit.py` | P1 (zero_shot) |
| `08_run_golisa_validation.py` | P1 (zero_shot) |
| `09_reclassify_and_ablation.py` | P1, P2, P3, P4 (Stage 2 multi-prompt ablation) |
| `10_run_json_mode_microbenchmark.py` | P4 (json_mode) |
| `11_run_golisa_re_cloud.py` | P1 (zero_shot) |
| `12_run_cloud_fewshot.py` | P2 (few_shot) |
| `13_run_golisa_re_llama.py` | P1 (zero_shot) |
| `14_local_repeat.py` | P1, P2, P3 |
| `15_claude_repeat_cot.py` | P1, P2, P3 |
| `16_gemini_repeat_cot.py` | P1, P2, P3 |
| `17_cloud_single_model_repeat.py` | P1, P2, P3 |
