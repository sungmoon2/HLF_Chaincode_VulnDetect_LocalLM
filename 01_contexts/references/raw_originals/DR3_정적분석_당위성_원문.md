# [원문 전문] Deep Research 3: 정적 분석 당위성 보고서
> Title: "Deterministic Assurance in Non-Deterministic Environments: A Case for Local sLM-Based Static Analysis in Hyperledger Fabric Chaincode Security"
> 생성일: 2026-02-09
> 도구: Gemini Deep Research
> 분석 요약: 05_정적분석_당위성_보고서.md

(이 파일의 전문 분석 내용은 05_정적분석_당위성_보고서.md에 구조화되어 있습니다. 본 파일은 원문의 섹션 구조와 인용 목록을 보존합니다.)

## 원문 구조 (10개 섹션)
1. Abstract
2. Introduction: The Divergence of Enterprise Blockchain Security
   - 2.1 The Problem of Dynamic Analysis in Permissioned Systems
   - 2.2 The Static Analysis and sLM Alternative
3. The Consensus vs. Crash Argument: Why Fuzzing Fails in HLF
   - 3.1 The "Execute-Order-Validate" Architecture and Silent Failures (The Fuzzing Blind Spot)
   - 3.2 Ineffectiveness of Fuzzing on Non-Deterministic Execution (Yang et al. 2021, Differential Fuzzing cost)
   - 3.3 Phantom Reads and Optimistic Concurrency (Olivieri et al. 2024)
4. The Harness Complexity Barrier
   - 4.1 Structural Differences: EVM vs HLF
   - 4.2 The Failure of MockStub (Rich Query Limitations, Concurrency Blindness, Harness Engineering Overhead)
   - 4.3 Static Analysis as the Practical Alternative
5. Precedents of Static-Only Research (2020-2026)
   - 5.1 The Dominance of Slither in Ethereum (Feist et al. 2019)
   - 5.2 GoLiSA and the HLF Static Renaissance (Olivieri et al., Commercio.network)
   - 5.3 Emerging Trends 2025-2026 (GoDetect, Cross-Channel Risk Detection)
6. LLM as a Semantic Static Analyzer
   - 6.1 Logic Bugs vs. Structural Bugs (Semantic Gap)
   - 6.2 Evidence of LLM Efficacy (SADA 2024, Forge 2026, Reentrancy Detection 2025)
   - 6.3 The Case for Local sLMs in Enterprise (Data Sovereignty, Resource Efficiency, Adversarial Resilience)
7. Comparative Analysis (Table 1: 6 rows x 4 columns)
   - 7.2 Proposed Methodology: The Hybrid Engine (Stage 1 Hard Filter + Stage 2 Soft Filter)
8. Conclusion
9. Key Arguments Summary (4 arguments)
10. Citation List (29 citations)

## Citation List (29건 원문 그대로)

1. Zhuang, Y., et al. (2025). "TPH-Fuzz: A Two-Phase Hybrid Fuzzing Framework for Smart Contract Vulnerability Detection." Electronics.
2. Hacken. (2024). "Fuzzing for Blockchain: Limitations and Challenges." Hacken Blog.
3. Yang, Y., et al. (2021). "Finding Consensus Bugs in Ethereum via Multi-transaction Differential Fuzzing." OSDI '21.
4. Yang, Y., et al. (2021). "Fluffy: Consensus Bug Detection." USENIX OSDI.
5. Ding, M., et al. (2021). "HFContractFuzzer: Fuzzing Hyperledger Fabric Smart Contracts." EASE 2021.
6. Ding, M., et al. (2021). "Limitations of Go-Fuzz for Hyperledger Fabric." arXiv.
7. "Hyperledger Fabric Transaction Flow: Execute-Order-Validate." arXiv 2509.07425.
8. Olivieri, L., et al. (2023). "Information Flow Analysis for Detecting Non-Determinism in Blockchain." GoLiSA Project.
9. Olivieri, L., et al. (2023). "GoLiSA: Semantics-based Static Analysis for Go." ResearchGate.
10. Olivieri, L., et al. (2024). "Detection of Phantom Reads in Hyperledger Fabric." IEEE Access.
11. Olivieri, L., et al. (2024). "Phantom Read Detection Limitations in MockStub." IEEE Access.
12. Feist, J., et al. (2019). "Slither: A Static Analysis Framework for Smart Contracts." WETSEB.
13. Feist, J. (2019). "Slither: Design and Implementation." IEEE.
14. Oracle. (2024). "Hyperledger Fabric MockStub Limitations." Oracle Blockchain Documentation.
15. Oracle. (2021). "Using Oracle Blockchain Platform Enterprise Edition." Oracle Docs.
16. Fu, X., et al. (2021). "On Private Data Collection of Hyperledger Fabric." ICDCS.
17. Chen, X., et al. (2025). "GoDetect: Precise Detection of Go Smart Contract Vulnerabilities." ICICC 2025.
18. "Forge: An LLM-driven Framework for Smart Contract Vulnerability Dataset Construction." ICSE 2026.
19. AncilarTech. (2024). "AI Code: The Bright Future of LLMs in Smart Contract Development." Medium.
20. "An Initial Exploration of Fine-tuning Small Language Models for Smart Contract Reentrancy Detection." arXiv 2505.19059.
21. LeewayHertz. (2025). "Small Language Models vs Large Language Models." LeewayHertz Insights.
22. "Potential Risks of Hyperledger Fabric Smart Contracts." EasyChair Preprint.
23. "Performance Benchmarking and Optimizing Hyperledger Fabric." Linux Foundation.
24. "Validation System Chaincode (VSCC) Checks." University of Klagenfurt.
25. Chen, X., et al. (2025). "GoDetect: Precise Detection of Go Smart Contract Vulnerabilities Based on Static Single Assignment (SSA)." IEEE.
26. Olivieri, L., et al. (2025). "Static Detection of Untrusted Cross-Contract Invocations in Go Smart Contracts." ResearchGate.
27. "LLM-generated Smart Contracts Security Analysis." arXiv 2602.04039.
28. "SADA: Static and Dynamic Analyzer using Multiple LLM Agents." SCIRP.
29. "SADA: Novel Approach to Smart Contract Vulnerability Detection." SCIRP.
