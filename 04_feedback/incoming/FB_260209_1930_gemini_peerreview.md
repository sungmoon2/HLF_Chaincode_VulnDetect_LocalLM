# FB_260209_1930_gemini_peerreview

## Metadata
- **Source**: gemini (Peer Review Simulation)
- **Date**: 2026-02-09T19:30:00
- **Target Draft**: 260209_전체논문_v4.tex
- **Simulated Venue**: Top-Tier AI/SE Conference (e.g., ASE 2026)
- **Overall Score**: 5/10 (Borderline - Weak Reject)
- **Confidence**: 4/5 (Expert)

## Feedback Summary

### 1. [Critical] N=15 Dataset Scale (ISS_006 escalation)
- **Reviewer concern**: N=15 is "toy-level"; 100% TPR may be luck or overfitting; samples < 30 cannot claim statistical significance.
- **Attack vector**: Qwen's 100% may reflect trivial tasks or data contamination, not model capability.
- **Defense strategy (best)**: Include GoLiSA (651 files) preliminary result.
- **Defense strategy (fallback)**: Add "Data Contamination Analysis" subsection in Discussion. Acknowledge contamination possibility, argue obfuscation experiment provides partial counter-evidence.

### 2. [Methodology] Semgrep Strawman Baseline (ISS_009 escalation)
- **Reviewer concern**: No custom rules authored for Semgrep; unfair comparison. LLMs got prompt tuning but Semgrep got zero configuration.
- **Attack vector**: At minimum, a taint analysis rule (source: time.Now -> sink: PutState) should have been written.
- **Defense strategy**: Justify target audience is "generalist chaincode developer" who cannot write complex AST-based rules. Out-of-the-box comparison is more realistic and practically meaningful.

### 3. [Validity] 100% Detection Rate Suspicion (ISS_007 escalation)
- **Reviewer concern**: 100% in AI papers is not credible; indicates synthetic/trivial test cases.
- **Attack vector**: Cannot prove whether Qwen's correct classification of safe_03 is "reasoning" or "memorization".
- **Defense strategy**: Tone down "achieved 100%" → "demonstrated complete coverage on this specific micro-benchmark". Emphasize obfuscation failures (2 missed vulns) to show model is imperfect.

### 4. [Significance] Why HLF? Market Share Problem
- **Reviewer concern**: Ethereum/Solidity is mainstream; HLF has fewer users; impact is limited.
- **Attack vector**: HLF research has less impact than Ethereum reentrancy detection.
- **Defense strategy**: Reinforce Privacy Paradox — Ethereum code is public (cloud LLM OK), HLF code is proprietary (local sLM is the only viable option). HLF's enterprise nature makes this research MORE valuable, not less.

## Derived Issues
- ISS_015: Title reframing — add "Qualitative" to title
- ISS_016: Data Contamination Analysis — new Discussion subsection
- ISS_017: Model version footnotes — add access date footnotes
- ISS_018: Detailed obfuscation failure analysis — expand failure cases

## Recommended v5 Changes
1. Title: "A Feasibility Study" → "A Qualitative Feasibility Study"
2. Abstract: Tone down coverage claims
3. Introduction: Strengthen Privacy Paradox with Ethereum contrast
4. Methodology Baseline: Add generalist developer justification
5. Results: Use "complete coverage on this micro-benchmark" instead of "100%"
6. Discussion: Add Data Contamination Analysis subsection
7. Discussion: Strengthen GoLiSA acknowledgment
8. Discussion: Expand Privacy-Preserving Deployment with privacy paradox
9. Obfuscation: Detailed failure analysis of 2 missed vulns
10. Conclusion: Add enterprise-specific value proposition
11. Footnotes: Model version access dates
