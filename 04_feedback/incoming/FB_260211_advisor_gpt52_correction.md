# FB_260211_advisor_gpt52_correction

## Metadata
- **Source**: advisor (GPT 5.2 활용 교정본)
- **Date**: 2026-02-11
- **Target Draft**: 260210_v30_서브섹션병합.tex (제출본, 260210_2346_최종제출)
- **Tool**: GPT 5.2 (교수님 컨텍스트 윈도우)
- **Type**: 교정본 (C1~C17 반영, 섹션 단위 대체용 문단)
- **Parent Feedback**: FB_260211_advisor_gpt52

## 피드백 원문

아래는 업로드하신 논문
260210_2346_최종제출

 에서 제가 앞서 제시한 "크리티컬 수정(C1~C17)"을 문맥 단위로 반영해, 해당 섹션 전체 문단을 "교정본(대체용)" 스타일로 재작성한 것입니다.
(각 블록은 해당 섹션의 문단 전체를 그대로 교체할 수 있도록 구성했습니다.)

### Abstract (PDF 1쪽) — 교정본 대체문

Abstract—Hyperledger Fabric (HLF) chaincode can exhibit endorsement/validation nondeterminism (sometimes loosely referred to as "consensus-layer" issues in HLF)—such as nondeterministic timestamps, global-variable mutation, and phantom reads—that can lead to silent endorsement failures without explicit runtime errors. This paper investigates whether locally deployed small language models (sLMs) can detect such issues while distinguishing them from safe coding patterns. Using 15 Go chaincode files (9 vulnerable, 6 benign traps), we compare two quantized 7–8B local models under three prompting strategies against six cloud LLMs and Semgrep. Qwen2.5-Coder-7B correctly classifies all 15 files across all prompting strategies; cloud models achieve a True Positive Rate (TPR) of 9/9, but their True Negative Rate (TNR) under zero-shot ranges from 0/6 to 5/6. Under few-shot prompting, Claude models improve to TNR 5/6–6/6, matching Qwen, while Gemini models show limited or no improvement. Identifier obfuscation reduces Qwen's performance to 7/9 TPR and 4/6 TNR. External validation on the GoLiSA benchmark (657 Go files) shows that few-shot prompting detects 5/5 known-vulnerable "Running Examples" files where zero-shot detects only 2/5. These results are empirical observations within a controlled evaluation and do not claim statistical generalizability.

### Index Terms (PDF 1쪽) — 교정본 대체문(권장)

Index Terms는 "문단"은 아니지만, C1(용어 오해)와 같이 움직이는 것이 안전합니다.

Index Terms—Hyperledger Fabric, chaincode, smart contract security, small language model, vulnerability detection, endorsement nondeterminism, local inference, privacy-preserving

### I. INTRODUCTION (PDF 1쪽) — 교정본 대체문(해당 섹션 전체 문단)

Hyperledger Fabric (HLF) employs an execute–order–validate architecture in which endorsing peers independently execute chaincode and must produce byte-identical read–write sets for a transaction to be committed [1]. Consequently, nondeterministic behaviors in chaincode—e.g., reading wall-clock time, relying on randomized map iteration, or triggering phantom reads—may not surface as conventional runtime failures. Instead, they manifest as endorsement mismatches when peers compare their read–write sets during validation. Although the root cause is not the ordering-service consensus protocol itself, we use the term "consensus-layer vulnerabilities" as shorthand for vulnerabilities that break endorsement agreement in HLF.

General-purpose static analysis tools typically operate on syntactic patterns and may not capture HLF endorsement semantics without Fabric-specific rules or analyses. Dedicated tools such as GoLiSA [11] and VulFinder [12] exist but require specialized deployment or target broader defect classes. Cloud-based LLMs offer strong code-reasoning capabilities, but transmitting proprietary chaincode to external APIs can conflict with the confidentiality expectations that motivate permissioned platforms. In contrast, locally deployed sLMs on consumer GPUs can operate offline at low marginal cost per analysis while preserving code confidentiality—an important property when chaincode embodies proprietary business logic.

This paper asks: Can a locally deployed 7B-parameter language model detect HLF consensus-layer vulnerabilities (i.e., endorsement/validation nondeterminism) while correctly distinguishing safe patterns, and how does it compare to cloud LLMs and traditional tools?

We make four contributions: (1) a curated micro-benchmark of 15 Go files designed to test both detection and discrimination; (2) a multi-axis comparison of two local sLMs, six cloud LLMs, and Semgrep under multiple prompting strategies; (3) an obfuscation experiment quantifying naming-cue dependence; and (4) a qualitative error analysis of distinct failure modes across model families.

### III. METHODOLOGY – A. Threat Model and Scope (PDF 1쪽) — 교정본 대체문

We target six classes of HLF endorsement/validation nondeterminism (referred to as "consensus-layer vulnerabilities" for brevity): nondeterministic timestamps, global variable mutation, goroutine concurrency hazards, map iteration randomness, phantom reads, and iterator resource leaks.

### III. METHODOLOGY – C. Models (PDF 1쪽) — 교정본 대체문

Table I lists all evaluated models. For local inference, we run quantized models using llama-cpp-python 0.3.16 with full GPU offload on an NVIDIA RTX 3090 Ti (24,564 MiB VRAM). Unless otherwise stated, we use temperature 0.1 and a maximum of 2,048 generated tokens with n_ctx=4,096; for the GoLiSA corpus evaluation in Section III-G we increase n_ctx to 16,384 to accommodate longer inputs.

(참고) Table I 내 "Gemini 2.5 Prob" 표기 수정(C17)
위치: PDF 2쪽(Table I)
교정: Gemini 2.5 Prob → Gemini 2.5 Pro

### III. METHODOLOGY – D. Prompt Strategies (PDF 2쪽) — 교정본 대체문

We evaluate three prompting strategies. P1 (Zero-shot) uses a system prompt that enumerates the six targeted vulnerability classes and requests a structured output followed by an overall verdict. P2 (Few-shot) augments P1 with two annotated examples (one vulnerable, one safe) to ground the notion of endorsement/validation nondeterminism. P3 (Chain-of-Thought) adds step-by-step reasoning instructions. All models are evaluated under P1, P2, and P3.

For local models, we additionally test a JSON-mode variant implemented via grammar-constrained decoding in llama-cpp-python. JSON mode constrains responses to a fixed schema and reduces label inconsistency (e.g., cases where vulnerability indicators are described but the final verdict is "safe"). Because structured-output support differs across cloud providers and endpoints, JSON-mode evaluation is limited to local models.

### III. METHODOLOGY – F. Traditional Tool Baseline (PDF 2쪽) — 교정본 대체문

As a general-purpose static analysis baseline, we run Semgrep 1.151.0 [10] with the default p/security-audit ruleset on all 15 files without any HLF-specific custom rules. This configuration approximates what a generalist developer might apply out of the box when auditing Go code.

### III. METHODOLOGY – G. External Validation on GoLiSA Benchmark (PDF 2쪽) — 교정본 대체문(권장)

III-C에서 n_ctx 예외를 명시했더라도, III-G 문단 자체도 매끈하게 다듬어 두는 게 좋습니다.

We run Qwen2.5-Coder-7B on the GoLiSA benchmark corpus [11]: 657 Go files from 326 GitHub repositories (5,438,685 bytes). For this corpus evaluation we use n_ctx=16,384. The corpus includes five Running Examples files with known ground truth (all vulnerable, 216–458 characters each). On these five files, we compare four strategies: zero-shot, few-shot, Chain-of-Thought, and JSON mode. Llama-3.1-8B is excluded from the large-scale validation given its TNR of 1/6 on the micro-benchmark, indicating insufficient discrimination under our consensus-only labeling.

Cloud models and Llama-3.1-8B are additionally evaluated on the five Running Examples files under P1 for cross-model comparison. Semgrep is also run on all 657 files.

To handle self-contradictory outputs—where a response describes concrete vulnerability indicators but ends with a "safe" verdict—we implement two classifier versions: v1 (keyword-based) and v2 (prioritizing vulnerability indicators over the final label).

### III. METHODOLOGY – H. Evaluation Metrics (PDF 2쪽) — 교정본 대체문

We report file-level detection and discrimination using two metrics. TPR is the proportion of vulnerable files for which the model reports at least one finding mapped to the targeted consensus-layer classes (i.e., endorsement/validation nondeterminism). TNR is the proportion of benign-trap files for which the model outputs a final "safe" verdict under our consensus-only labeling.

Given N=15, we do not apply statistical significance tests; we report observed rates as qualitative evidence within this evaluation. All experiments were repeated five times at temperature 0.1 to validate result consistency (Section IV).

### TABLE II 각주 (PDF 3쪽) — 교정본 대체문

All prompt strategies were repeated five times at temperature 0.1. For cloud models, we report the median when run-to-run variation was observed (‡). ‡Run-to-run variation observed under CoT: Gemini Pro 5/6–6/6; Gemini Flash 0/6–3/6. All other CoT results were identical across five runs.

### IV. RESULTS – B. Traditional Tool Baseline (PDF 3쪽) — 교정본 대체문(권장)

Semgrep reports no findings mapped to our targeted consensus-layer (endorsement/validation nondeterminism) classes across all 15 files. Its only output is a generic math-random-used warning on safe_04, which is unrelated to endorsement nondeterminism.

### IV. RESULTS – F. External Validation Results (PDF 4쪽) — 교정본 대체문

Qwen processed all 657 GoLiSA files with zero errors in 5,252.7 seconds (7.995 s/file). Semgrep produced 12 generic findings and 0 consensus-layer findings across all 657 files.

On the Running Examples files (Table IV), all Claude models and Gemini Pro/Flash achieved 5/5 under zero-shot, while Gemini Flash Lite achieved 2/5. Llama-3.1-8B flagged all five files but reported extensive non-consensus findings alongside consensus-layer issues. Qwen's zero-shot rate of 2/5 on these minimal-context files (216–458 characters) is lower than most cloud models; few-shot prompting restores Qwen to 5/5.

On the full corpus, classifier v1 categorizes 380 files as vulnerable and 277 as safe; classifier v2 reclassifies 97 self-contradictory responses, yielding 477 vulnerable and 180 safe. On the 15-file micro-benchmark, Qwen completes a full audit in 59.1 seconds (3.9 s/file) on local hardware with no network dependency.

### V. DISCUSSION – A. Specialist vs. Generalist Discrimination (PDF 4–5쪽) — 교정본 대체문(해당 소절 전체 문단)

The central finding is not that models achieve high TPR—all do—but that TNR varies from 0/6 to 6/6. In our micro-benchmark, Qwen2.5-Coder-7B's outputs are consistent with tracing whether nondeterministic sources influence ledger writes (e.g., whether a nondeterministic value reaches PutState), suggesting stronger data-flow sensitivity than keyword-triggered heuristics. For example, only Qwen and Gemini Flash Lite correctly identify safe_03, which iterates a map but serializes it via json.Marshal (deterministic key sorting). Llama-3.1-8B flags many files based on the presence of tokens such as time.Now or map without tracing downstream effects, and this behavior persists across prompting strategies—consistent with an architectural rather than prompt-dependent limitation for this task.

Under few-shot prompting (Table II), Claude Haiku and Opus each reach TNR 6/6, matching Qwen's zero-shot performance on the micro-benchmark. Sonnet improves from 2/6 to 5/6. This indicates that the zero-shot TNR gap for Claude models is prompt-dependent rather than architectural.

Under Chain-of-Thought prompting, all three Claude models converge to TNR 6/6, with Sonnet gaining the additional file that few-shot missed. Gemini Pro reverses from TNR 0/6 (zero-shot and few-shot) to 5/6 under CoT, indicating that step-by-step reasoning instructions can steer it toward endorsement/validation semantics rather than general security concerns. Flash Lite reaches TNR 6/6 under CoT, consistent across five runs. Flash remains at TNR 1/6, suggesting that its scope-expansion pattern is relatively resistant to prompt-based correction in this evaluation.

Fig. 2 illustrates this convergence across prompting strategies. The practical distinction is that Qwen achieves full discrimination on the micro-benchmark without requiring example code or reasoning instructions in the prompt and operates offline, preserving code confidentiality.

Within the Claude family, Sonnet (mid-tier) achieves lower TNR (2/6) than both Haiku and Opus (5/6 each) under zero-shot, and Gemini Flash Lite (2/6) outperforms Pro and Flash (0/6), suggesting that model scale does not monotonically improve discrimination in this evaluation. Flash Lite is the sole Gemini model whose TNR improves under few-shot prompting (2/6 to 3/6); Pro and Flash remain at 0/6 under few-shot but diverge under CoT (Pro 5/6, Flash 1/6).

### V. DISCUSSION – D. Data Contamination Analysis (PDF 5쪽) — 교정본 대체문

Qwen's pretraining data may include HLF documentation and public chaincode examples, but we do not have direct evidence of dataset overlap. The obfuscation experiment provides partial evidence against pure memorization: Qwen still correctly classifies 11/15 files under identifier renaming. The model also correctly distinguishes safe_05, where time.Now() is overwritten by GetTxTimestamp() before PutState—requiring data-flow tracking through variable reassignment rather than mere keyword detection. We characterize the analysis mechanism as a hybrid of structural cues and learned associations, without claiming that either operates in isolation.

### VI. CONCLUSION (PDF 6쪽) — 교정본 대체문(해당 섹션 전체 문단)

This feasibility study evaluates local sLMs for detecting HLF endorsement/validation nondeterminism ("consensus-layer") vulnerabilities on a 15-file micro-benchmark and validates the approach on 657 GoLiSA benchmark files. Qwen2.5-Coder-7B correctly classifies all 15 files across three prompting strategies while operating offline on consumer hardware. Cloud models achieve TPR 9/9 but lower TNR under zero-shot (0/6 to 5/6); few-shot prompting improves Claude models to TNR 5/6–6/6, and CoT prompting achieves TNR 6/6 for all Claude models and Gemini Flash Lite, with Gemini Pro reaching 5/6. With the default p/security-audit ruleset, Semgrep detects zero consensus-layer vulnerabilities on both datasets. On minimal code fragments (five GoLiSA Running Examples), prompt strategy becomes decisive: few-shot and JSON mode both achieve 5/5 where zero-shot detects 2/5. Obfuscation reduces Qwen's rates to 7/9 TPR and 4/6 TNR, confirming partial naming-cue dependence. Within our evaluation, a code-specialist sLM matches the best-performing prompt-engineered cloud models on discrimination while preserving code confidentiality, suggesting a practical alternative for privacy-sensitive HLF environments. Future work will evaluate larger corpora with per-file ground truth and investigate fine-tuning for improved obfuscation robustness, including resilience to adversarial identifier manipulation.

### 추가 제안

원하시면 다음도 같이 해드릴 수 있습니다(질문 없이 제가 바로 진행 가능):

위 교정본을 기준으로 논문 전체에서 "consensus-layer" 용어를 어떻게 일관되게 유지/대체할지(전역 치환 전략) 제시

"crucial claim(과장 가능)" 문장들을 전체 스캔해서 톤 다운 교정본 일괄 작성

Table/Caption/Footnote까지 포함한 IEEE 스타일 전체 교정본(섹션 단위) 정리

## 파생 안건

없음 (FB_260211_advisor_gpt52의 C1~C17 반영 교정본이므로, 기존 ISS_019~ISS_023에 귀속)

## 비고
- FB_260211_advisor_gpt52 의견 피드백에 대한 교정본(correction) 후속 자료
- 총 15개 섹션/소절 단위 대체용 문단 + Table I 오타 수정 1건 + Table II 각주 대체 1건
- 대상 섹션: Abstract, Index Terms, I, III-A, III-C, III-D, III-F, III-G, III-H, Table II 각주, IV-B, IV-F, V-A, V-D, VI
