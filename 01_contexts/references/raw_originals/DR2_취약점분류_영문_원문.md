# [원문 전문] Deep Research 2: Vulnerability Taxonomy (English)
> Title: "Comprehensive Analysis of Deterministic Execution and Security Anti-Patterns in Hyperledger Fabric Smart Contracts"
> 생성일: 2026-02-09
> 도구: Gemini Deep Research
> 분석 요약: 04_vulnerability_taxonomy_en.md

(이 파일의 전문 내용은 분석 요약본 04_vulnerability_taxonomy_en.md에 contractapi 기반 코드 블록을 포함하여 원문에 가깝게 구조화되어 있습니다. 04 파일이 사실상 원문 전문 역할을 수행합니다.)

## 원문 구조
- 1. Introduction: The Imperative of Determinism in Permissioned Ledgers
  - 1.1 The Execute-Order-Validate Architecture
  - 1.2 The Role of Static Analysis
- 2. Vulnerability Case Study 1: Non-Deterministic Execution
  - 2.1 Theoretical Mechanism (System Time, Randomness, Map Iteration)
  - 2.2 Impact Analysis (Endorsement Failure, Ledger Divergence)
  - 2.3 Code Example: TimeBasedAsset (contractapi)
  - 2.4 Remediation Strategy (GetTxTimestamp)
- 3. Vulnerability Case Study 2: Global Variable Usage
  - 3.1 Mechanism (Isolated, Ephemeral, Independent containers)
  - 3.2 Impact (Split Brain, Data Loss, Concurrency Conflicts)
  - 3.3 Code Example: GlobalStateStore (contractapi)
  - 3.4 Remediation (PutState/GetState)
- 4. Vulnerability Case Study 3: Phantom Reads
  - 4.1 Mechanism (Range Query Hash Validation, Check-Then-Act)
  - 4.2 Impact (Availability Denial, Data Inconsistency with Rich Queries)
  - 4.3 Code Example: PhantomAsset (contractapi)
  - 4.4 Remediation (Off-chain indexer, Delta Updates, Pagination)
- 5. Vulnerability Case Study 4: Unbounded Range Queries
  - 5.1 Mechanism (Execution Timeout 30s, gRPC 100MB, OOM)
  - 5.2 Impact (DoS, System Instability)
  - 5.3 Code Example: UnboundedIterator (contractapi)
  - 5.4 Remediation (GetStateByRangeWithPagination)
- 6. Vulnerability Case Study 5: Concurrency Hazards (Goroutines)
  - 6.1 Mechanism (Race Condition, Non-Deterministic Order, Silent Failures)
  - 6.2 Impact (Data Integrity Loss, Endorsement Mismatches, Runtime Panics)
  - 6.3 Code Example: ConcurrentWriter (contractapi)
  - 6.4 Remediation (Sequential execution)
- 7. Implications for Static Analysis Tools
  - 7.1 Detection Heuristics (Context-Aware AST, Forbidden Packages, Global Variable Taint, API Patterns)
  - 7.2 Dataset Utility (Ground Truth, False Negative/Positive rates)
- 8. Conclusion

각 Case에 contractapi 기반 Go 코드 예제 + VULNERABILITY 주석 + Remediation Strategy 포함.
비교 분석표: Dynamic Fuzzing vs Formal SA vs Semantic SA (6행) 포함.
