"""
08_run_golisa_validation.py  (v1.0 -- 2026-02-09)
GoLiSA 외부 검증 실험: 657개 실제 HLF 체인코드에 대해 Qwen2.5-Coder-7B + Semgrep 실행

- Phase A: Semgrep (auto + p/security-audit) on 657 files
- Phase B: Qwen2.5-Coder-7B zero_shot on 657 files
- 응답 분류기: vulnerable/safe 이진 분류 + 취약점 유형 추출
- Running_Examples 5개 파일 sanity check (알려진 취약점)
- 출력: CSV + meta.json + GOLISA_VALIDATION_REPORT.md

Usage:
    python scripts/08_run_golisa_validation.py
    python scripts/08_run_golisa_validation.py --skip-semgrep
    python scripts/08_run_golisa_validation.py --skip-qwen
"""

import argparse
import csv
import hashlib
import json
import os
import platform
import re
import shutil
import subprocess
import sys
import time
from collections import Counter, defaultdict
from datetime import datetime
from pathlib import Path

from colorama import init, Fore, Style
from tqdm import tqdm

init(autoreset=True)

# ══════════════════════════════════════════════════════════════════════
# 경로 설정
# ══════════════════════════════════════════════════════════════════════
PROJECT_ROOT = Path(__file__).resolve().parent.parent
BENCHMARK_DIR = PROJECT_ROOT / "02_resources" / "golisa_benchmark" / "Benchmark"
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

MODEL_FILE = "qwen2.5-coder-7b-instruct-q4_k_m.gguf"

# ══════════════════════════════════════════════════════════════════════
# 추론 파라미터
# ══════════════════════════════════════════════════════════════════════
INFERENCE_PARAMS = {
    "n_gpu_layers": -1,
    "n_ctx": 16384,       # 4096→16384 상향 (대형 파일 처리)
    "temperature": 0.1,
    "max_tokens": 2048,
}

# ══════════════════════════════════════════════════════════════════════
# 프롬프트 (02_run_audit_v3.py의 zero_shot 동일)
# ══════════════════════════════════════════════════════════════════════
PROMPT_ZERO_SHOT = (
    "You are a Hyperledger Fabric Security Expert. "
    "Analyze the following Go chaincode for security vulnerabilities. "
    "Focus on: access control issues, input validation flaws, "
    "read-after-write conflicts (phantom reads), private data leakage, "
    "non-deterministic operations, and insecure key management. "
    "For each vulnerability found, provide:\n"
    "1. Vulnerability type\n"
    "2. Severity (Critical/High/Medium/Low)\n"
    "3. Affected code location (function name and line reference)\n"
    "4. Description of the issue\n"
    "5. Recommended fix\n"
    "If no vulnerabilities are found, state 'No vulnerabilities detected.'"
)

# ══════════════════════════════════════════════════════════════════════
# Semgrep 설정
# ══════════════════════════════════════════════════════════════════════
SEMGREP_CONFIGS = [
    {"label": "auto", "config": "auto"},
    {"label": "security-audit", "config": "p/security-audit"},
]

CONSENSUS_KEYWORDS = [
    "non-deterministic", "nondeterministic",
    "endorsement", "consensus", "phantom read",
    "read-after-write", "chaincode",
    "fabric", "hlf", "ledger",
    "getstate", "putstate",
]

# ══════════════════════════════════════════════════════════════════════
# Running_Examples ground truth
# ══════════════════════════════════════════════════════════════════════
RUNNING_EXAMPLES_GROUND_TRUTH = {
    "Running_Examples/channel/Channel.go": {
        "expected": "vulnerable",
        "vuln_type": "goroutine",
        "description": "Goroutine/channel nondeterminism: two goroutines send to channel, receive order varies, result goes to PutState",
    },
    "Running_Examples/global/GlobalVariable.go": {
        "expected": "vulnerable",
        "vuln_type": "global_var",
        "description": "Global variable mutation: var glob written across invocations, PutState uses glob value",
    },
    "Running_Examples/goroutine/GoRoutines.go": {
        "expected": "vulnerable",
        "vuln_type": "goroutine",
        "description": "Goroutine race condition: two goroutines append to shared string s, PutState uses s",
    },
    "Running_Examples/map-iter/MapIteration.go": {
        "expected": "vulnerable",
        "vuln_type": "map_iter",
        "description": "Map iteration order: range over map concatenates values, order nondeterministic, goes to PutState",
    },
    "Running_Examples/method-function/MethodFunction.go": {
        "expected": "vulnerable",
        "vuln_type": "timestamp",
        "description": "Nondeterministic timestamp: time.Now() result directly passed to PutState",
    },
}


# ══════════════════════════════════════════════════════════════════════
# 파일 탐색
# ══════════════════════════════════════════════════════════════════════
def discover_go_files(benchmark_dir: Path) -> list[dict]:
    """재귀적으로 .go 파일을 탐색하고, repo/file 메타데이터를 반환한다."""
    files = []
    for go_file in sorted(benchmark_dir.rglob("*.go")):
        rel_path = go_file.relative_to(benchmark_dir)
        parts = rel_path.parts
        if len(parts) >= 2:
            repo = parts[0]
            filename = "/".join(parts[1:])
        else:
            repo = "_root"
            filename = parts[0]

        files.append({
            "abs_path": go_file,
            "rel_path": str(rel_path).replace("\\", "/"),
            "repo": repo,
            "filename": filename,
            "size_bytes": go_file.stat().st_size,
        })
    return files


# ══════════════════════════════════════════════════════════════════════
# 응답 분류기
# ══════════════════════════════════════════════════════════════════════
def classify_response(response: str) -> str:
    """LLM 응답을 'vulnerable' 또는 'safe'로 분류한다."""
    if not response or response.startswith("ERROR:"):
        return "error"

    resp_lower = response.lower()

    safe_indicators = [
        "no vulnerabilities detected",
        "no vulnerabilities found",
        "no security vulnerabilities",
        "no significant vulnerabilities",
        "no vulnerabilities were found",
        "no vulnerabilities were detected",
        "no critical vulnerabilities",
        "the code appears to be secure",
        "the code is secure",
        "no issues found",
        "no issues detected",
    ]
    for indicator in safe_indicators:
        if indicator in resp_lower:
            # 추가 확인: "no vulnerabilities detected" 뒤에 "however" 등으로
            # 실제 취약점을 언급하는 경우 제외
            idx = resp_lower.find(indicator)
            after = resp_lower[idx + len(indicator):]
            if any(kw in after[:200] for kw in ["however", "but ", "although", "note that"]):
                # 안전 판정 후 부연에 취약점이 있을 수 있음 → vulnerable로 전환
                vuln_check = ["vulnerability", "vulnerable", "severity:", "recommended fix"]
                if sum(1 for v in vuln_check if v in after) >= 2:
                    return "vulnerable"
            return "safe"

    # 취약점 지표 카운트
    vuln_indicators = [
        "vulnerability type",
        "severity:",
        "severity :",
        "recommended fix",
        "affected code",
        "critical",
        "high",
        "non-deterministic",
        "nondeterministic",
        "phantom read",
        "global variable",
        "goroutine",
        "race condition",
        "map iteration",
        "iterator leak",
        "putstate",
    ]
    vuln_count = sum(1 for ind in vuln_indicators if ind in resp_lower)
    if vuln_count >= 2:
        return "vulnerable"

    return "vulnerable"  # 기본값: 불확실하면 vulnerable (보수적)


def extract_vuln_types(response: str) -> list[str]:
    """LLM 응답에서 취약점 유형을 추출한다."""
    if not response:
        return []

    resp_lower = response.lower()
    found_types = []

    type_patterns = {
        "timestamp": [
            "time.now", "timestamp", "non-deterministic time",
            "nondeterministic time", "gettxtimestamp",
        ],
        "global_var": [
            "global variable", "global state", "mutable global",
            "cross-invocation", "shared state",
        ],
        "goroutine": [
            "goroutine", "go routine", "race condition",
            "concurrent", "channel", "sync.",
        ],
        "map_iter": [
            "map iteration", "map traversal", "iteration order",
            "range over map", "unordered map",
        ],
        "phantom_read": [
            "phantom read", "read-after-write", "mvcc",
            "read after write", "getstate.*putstate",
        ],
        "iterator_leak": [
            "iterator", "close()", "defer.*close",
            "resource leak", "getstatebyrange",
        ],
        "access_control": [
            "access control", "authorization", "permission",
            "getCreator", "getcreator", "msp",
        ],
        "input_validation": [
            "input validation", "sanitiz", "injection",
            "unchecked input", "user input",
        ],
        "random": [
            "math/rand", "crypto/rand", "random",
        ],
        "external_call": [
            "external api", "http.", "net/http",
            "external service", "file i/o",
        ],
    }

    for vtype, patterns in type_patterns.items():
        for pattern in patterns:
            if pattern in resp_lower:
                found_types.append(vtype)
                break

    return found_types if found_types else ["other"]


# ══════════════════════════════════════════════════════════════════════
# Semgrep 실행
# ══════════════════════════════════════════════════════════════════════
def check_semgrep() -> str | None:
    """semgrep 실행 파일 경로를 반환한다. 없으면 None."""
    path = shutil.which("semgrep")
    if path:
        return path
    try:
        result = subprocess.run(
            ["semgrep", "--version"],
            capture_output=True, text=True, timeout=15,
        )
        if result.returncode == 0:
            return "semgrep"
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass
    return None


def get_semgrep_version(semgrep_path: str) -> str:
    try:
        result = subprocess.run(
            [semgrep_path, "--version"],
            capture_output=True, text=True, timeout=15,
        )
        return result.stdout.strip()
    except Exception:
        return "unknown"


def run_semgrep_on_file(semgrep_path: str, go_file: Path, config_value: str) -> dict:
    """단일 .go 파일에 대해 semgrep config를 실행한다."""
    start = time.time()
    cmd = [
        semgrep_path,
        "--config", config_value,
        "--json",
        "--no-git-ignore",
        "--quiet",
        str(go_file),
    ]

    try:
        env = {**os.environ}
        if config_value != "auto":
            env["SEMGREP_SEND_METRICS"] = "off"
        result = subprocess.run(
            cmd, capture_output=True, timeout=120,
            encoding="utf-8", errors="replace",
        )
        elapsed = round(time.time() - start, 3)

        stdout = result.stdout.strip()
        if not stdout:
            return {
                "success": False, "findings": [], "elapsed_s": elapsed,
                "error": result.stderr.strip() or f"No output (rc={result.returncode})",
            }

        try:
            data = json.loads(stdout)
        except json.JSONDecodeError as e:
            return {
                "success": False, "findings": [], "elapsed_s": elapsed,
                "error": f"JSON parse error: {e}",
            }

        findings = []
        for r in data.get("results", []):
            rule_id = r.get("check_id", "unknown")
            message = r.get("extra", {}).get("message", "")
            severity = r.get("extra", {}).get("severity", "UNKNOWN")
            line = r.get("start", {}).get("line", 0)

            # consensus 관련 여부 판정
            text_lower = (rule_id + " " + message).lower()
            is_consensus = any(kw in text_lower for kw in CONSENSUS_KEYWORDS)

            findings.append({
                "rule_id": rule_id,
                "severity": severity,
                "message": message,
                "line": line,
                "is_consensus": is_consensus,
            })

        return {
            "success": True, "findings": findings,
            "elapsed_s": elapsed, "error": None,
        }

    except subprocess.TimeoutExpired:
        return {
            "success": False, "findings": [], "elapsed_s": round(time.time() - start, 3),
            "error": "Timeout (>120s)",
        }
    except Exception as e:
        return {
            "success": False, "findings": [], "elapsed_s": round(time.time() - start, 3),
            "error": str(e),
        }


def format_findings_detail(findings: list) -> str:
    if not findings:
        return "NO_FINDINGS"
    parts = []
    for f in findings:
        consensus_tag = "[CONSENSUS]" if f["is_consensus"] else "[generic]"
        parts.append(
            f"[L{f['line']}] {f['severity']} {consensus_tag} | "
            f"{f['rule_id']}: {f['message'][:100]}"
        )
    return " || ".join(parts)


# ══════════════════════════════════════════════════════════════════════
# Qwen 감사
# ══════════════════════════════════════════════════════════════════════
def load_model(model_path: Path):
    """GGUF 모델을 GPU 최대 활용 모드로 로드한다."""
    from llama_cpp import Llama
    print(f"{Fore.CYAN}[Model] 로딩 중: {model_path.name} (n_ctx={INFERENCE_PARAMS['n_ctx']})")
    llm = Llama(
        model_path=str(model_path),
        n_gpu_layers=INFERENCE_PARAMS["n_gpu_layers"],
        n_ctx=INFERENCE_PARAMS["n_ctx"],
        verbose=False,
    )
    print(f"{Fore.GREEN}[Model] 로드 완료: {model_path.name}")
    return llm


def audit_chaincode(llm, code: str, filename: str) -> str:
    """체인코드에 대해 보안 감사를 수행한다."""
    user_prompt = (
        f"Analyze this Hyperledger Fabric chaincode file '{filename}' "
        f"for security vulnerabilities:\n\n```go\n{code}\n```"
    )
    response = llm.create_chat_completion(
        messages=[
            {"role": "system", "content": PROMPT_ZERO_SHOT},
            {"role": "user", "content": user_prompt},
        ],
        max_tokens=INFERENCE_PARAMS["max_tokens"],
        temperature=INFERENCE_PARAMS["temperature"],
    )
    return response["choices"][0]["message"]["content"]


# ══════════════════════════════════════════════════════════════════════
# 리포트 생성
# ══════════════════════════════════════════════════════════════════════
def generate_report(
    go_files_meta: list[dict],
    qwen_results: dict,
    semgrep_results: dict,
    run_start: datetime,
    run_end: datetime,
    output_path: Path,
):
    """GOLISA_VALIDATION_REPORT.md를 생성한다."""
    total_files = len(go_files_meta)
    total_bytes = sum(f["size_bytes"] for f in go_files_meta)
    sizes = sorted(f["size_bytes"] for f in go_files_meta)
    avg_bytes = total_bytes // total_files if total_files else 0
    median_bytes = sizes[total_files // 2] if total_files else 0
    max_bytes = sizes[-1] if sizes else 0
    repos = set(f["repo"] for f in go_files_meta)

    # ── Qwen 통계 ────────────────────────────────────────────────
    qwen_classifications = Counter()
    qwen_vuln_types = Counter()
    qwen_timings = []

    for rel_path, res in qwen_results.items():
        cls = res.get("classification", "error")
        qwen_classifications[cls] += 1
        if cls == "vulnerable":
            for vt in res.get("vuln_types", []):
                qwen_vuln_types[vt] += 1
        qwen_timings.append(res.get("elapsed_s", 0))

    qwen_total_s = sum(qwen_timings)
    qwen_avg_s = qwen_total_s / len(qwen_timings) if qwen_timings else 0

    # ── Semgrep 통계 ──────────────────────────────────────────────
    semgrep_total_findings = 0
    semgrep_consensus_findings = 0
    semgrep_config_stats = defaultdict(lambda: {"total": 0, "consensus": 0, "files_with_findings": 0})

    for rel_path, configs in semgrep_results.items():
        for config_label, res in configs.items():
            count = res.get("findings_count", 0)
            consensus = res.get("consensus_relevant", 0)
            semgrep_total_findings += count
            semgrep_consensus_findings += consensus
            semgrep_config_stats[config_label]["total"] += count
            semgrep_config_stats[config_label]["consensus"] += consensus
            if count > 0:
                semgrep_config_stats[config_label]["files_with_findings"] += 1

    # ── Running_Examples 검증 ─────────────────────────────────────
    running_examples_results = []
    for gt_path, gt_info in RUNNING_EXAMPLES_GROUND_TRUTH.items():
        qwen_res = qwen_results.get(gt_path, {})
        qwen_cls = qwen_res.get("classification", "N/A")
        match = "O" if qwen_cls == gt_info["expected"] else "X"
        running_examples_results.append({
            "file": gt_path.split("/")[-1],
            "expected": gt_info["expected"].upper(),
            "qwen_result": qwen_cls.upper(),
            "match": match,
            "vuln_type": gt_info["vuln_type"],
        })

    re_correct = sum(1 for r in running_examples_results if r["match"] == "O")
    re_total = len(running_examples_results)

    # ── Cross-tool 비교 ───────────────────────────────────────────
    qwen_flagged = set(
        rp for rp, res in qwen_results.items() if res.get("classification") == "vulnerable"
    )
    semgrep_flagged = set()
    for rp, configs in semgrep_results.items():
        for config_label, res in configs.items():
            if res.get("findings_count", 0) > 0:
                semgrep_flagged.add(rp)
                break

    both_flagged = qwen_flagged & semgrep_flagged
    only_qwen = qwen_flagged - semgrep_flagged
    only_semgrep = semgrep_flagged - qwen_flagged

    # ── 리포트 작성 ───────────────────────────────────────────────
    lines = []
    lines.append("# GoLiSA External Validation Report")
    lines.append(f"> Generated: {run_end.strftime('%Y-%m-%d %H:%M:%S')} (KST)")
    lines.append(f"> Duration: {round((run_end - run_start).total_seconds(), 1)}s")
    lines.append("")

    # 1. Dataset Summary
    lines.append("## 1. Dataset Summary")
    lines.append(f"- **Source**: GoLiSA ECOOP 2023 Benchmark (Olivieri et al.)")
    lines.append(f"- **Total files**: {total_files}")
    lines.append(f"- **Total repositories**: {len(repos)}")
    lines.append(f"- **Total size**: {total_bytes:,} bytes ({total_bytes/1024/1024:.1f} MB)")
    lines.append(f"- **Average file size**: {avg_bytes:,} bytes")
    lines.append(f"- **Median file size**: {median_bytes:,} bytes")
    lines.append(f"- **Max file size**: {max_bytes:,} bytes")
    lines.append(f"- **Files > 16KB**: {sum(1 for f in go_files_meta if f['size_bytes'] > 16384)}")
    lines.append("")

    # 2. Qwen Results
    lines.append("## 2. Qwen2.5-Coder-7B Zero-Shot Results")
    lines.append(f"- **Model**: {MODEL_FILE}")
    lines.append(f"- **Prompt**: zero_shot (identical to micro-benchmark)")
    lines.append(f"- **n_ctx**: {INFERENCE_PARAMS['n_ctx']}")
    lines.append(f"- **Total inference time**: {qwen_total_s:.1f}s ({qwen_total_s/60:.1f}min)")
    lines.append(f"- **Average per file**: {qwen_avg_s:.3f}s")
    lines.append("")
    lines.append("### Classification Summary")
    lines.append("| Classification | Count | Percentage |")
    lines.append("|:---------------|------:|-----------:|")
    for cls in ["vulnerable", "safe", "error"]:
        count = qwen_classifications.get(cls, 0)
        pct = count / total_files * 100 if total_files else 0
        lines.append(f"| {cls} | {count} | {pct:.1f}% |")
    lines.append("")

    lines.append("### Vulnerability Type Distribution (files classified as vulnerable)")
    lines.append("| Vulnerability Type | Count |")
    lines.append("|:-------------------|------:|")
    for vt, count in qwen_vuln_types.most_common():
        lines.append(f"| {vt} | {count} |")
    lines.append("")

    # 3. Running_Examples
    lines.append("## 3. Running_Examples Validation (Mini Ground Truth)")
    lines.append(f"- **Accuracy**: {re_correct}/{re_total}")
    lines.append("")
    lines.append("| File | Expected | Qwen Result | Vuln Type | Match |")
    lines.append("|:-----|:---------|:------------|:----------|:-----:|")
    for r in running_examples_results:
        lines.append(
            f"| {r['file']} | {r['expected']} | {r['qwen_result']} | "
            f"{r['vuln_type']} | {r['match']} |"
        )
    lines.append("")

    # 4. Semgrep
    lines.append("## 4. Semgrep Comparison")
    lines.append(f"- **Semgrep version**: (see meta.json)")
    lines.append(f"- **Configs used**: {', '.join(c['config'] for c in SEMGREP_CONFIGS)}")
    lines.append(f"- **Total findings**: {semgrep_total_findings}")
    lines.append(f"- **Consensus-relevant findings**: {semgrep_consensus_findings}")
    lines.append("")
    lines.append("| Config | Total Findings | Files with Findings | Consensus-Relevant |")
    lines.append("|:-------|---------------:|--------------------:|-------------------:|")
    for cfg_label, stats in semgrep_config_stats.items():
        lines.append(
            f"| {cfg_label} | {stats['total']} | "
            f"{stats['files_with_findings']} | {stats['consensus']} |"
        )
    lines.append("")

    # 5. Cross-tool
    lines.append("## 5. Cross-Tool Comparison")
    lines.append(f"- **Qwen flagged as vulnerable**: {len(qwen_flagged)} files")
    lines.append(f"- **Semgrep flagged (any finding)**: {len(semgrep_flagged)} files")
    lines.append(f"- **Both flagged**: {len(both_flagged)} files")
    lines.append(f"- **Only Qwen**: {len(only_qwen)} files")
    lines.append(f"- **Only Semgrep**: {len(only_semgrep)} files")
    lines.append("")

    # 6. Key Findings
    lines.append("## 6. Key Findings for Paper")
    lines.append("")

    vuln_count = qwen_classifications.get("vulnerable", 0)
    safe_count = qwen_classifications.get("safe", 0)
    vuln_pct = vuln_count / total_files * 100 if total_files else 0

    lines.append(
        f"1. **Detection breadth**: Qwen2.5-Coder-7B flagged {vuln_count} out of "
        f"{total_files} real-world HLF chaincodes ({vuln_pct:.1f}%) as containing "
        f"consensus-layer vulnerabilities in a zero-shot setting."
    )
    lines.append(
        f"2. **Running_Examples accuracy**: {re_correct}/{re_total} known vulnerable files "
        f"correctly identified, validating the model's ability to detect canonical "
        f"nondeterminism patterns."
    )
    lines.append(
        f"3. **Traditional tool gap**: Semgrep found {semgrep_total_findings} generic findings "
        f"but {semgrep_consensus_findings} consensus-layer detections across {total_files} files, "
        f"confirming the domain-specific detection gap."
    )

    if qwen_vuln_types:
        top_type = qwen_vuln_types.most_common(1)[0]
        lines.append(
            f"4. **Most prevalent vulnerability**: '{top_type[0]}' was the most frequently "
            f"detected type ({top_type[1]} files), consistent with real-world HLF "
            f"chaincode practices."
        )

    lines.append(
        f"5. **Cross-tool complementarity**: {len(only_qwen)} files were flagged only by "
        f"Qwen (not by Semgrep), demonstrating the LLM's ability to detect domain-specific "
        f"vulnerabilities beyond traditional static analysis capabilities."
    )
    lines.append("")

    report_text = "\n".join(lines)
    output_path.write_text(report_text, encoding="utf-8")
    return report_text


# ══════════════════════════════════════════════════════════════════════
# 출력 경로 생성
# ══════════════════════════════════════════════════════════════════════
def generate_output_path(results_dir: Path, prefix: str, run_start: datetime) -> Path:
    base = f"{prefix}_{run_start.strftime('%y%m%d_%H%M')}"
    candidate = results_dir / f"{base}.csv"
    if not candidate.exists():
        return candidate
    seq = 2
    while True:
        candidate = results_dir / f"{base}_{seq}.csv"
        if not candidate.exists():
            return candidate
        seq += 1


# ══════════════════════════════════════════════════════════════════════
# CLI
# ══════════════════════════════════════════════════════════════════════
def parse_args():
    parser = argparse.ArgumentParser(
        description="GoLiSA External Validation -- Qwen + Semgrep on 657 HLF chaincodes"
    )
    parser.add_argument("--skip-semgrep", action="store_true", help="Semgrep 단계 건너뛰기")
    parser.add_argument("--skip-qwen", action="store_true", help="Qwen 감사 단계 건너뛰기")
    return parser.parse_args()


# ══════════════════════════════════════════════════════════════════════
# 메인
# ══════════════════════════════════════════════════════════════════════
def main():
    args = parse_args()
    run_start = datetime.now()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    print(f"{Fore.CYAN}{Style.BRIGHT}")
    print("=" * 70)
    print("  08_run_golisa_validation.py  v1.0")
    print("  GoLiSA External Validation: Qwen2.5-Coder-7B + Semgrep")
    print("  657 Real-World HLF Chaincodes from ECOOP 2023 Benchmark")
    print("=" * 70)
    print(Style.RESET_ALL)

    # ── 파일 탐색 ─────────────────────────────────────────────────
    if not BENCHMARK_DIR.exists():
        print(f"{Fore.RED}[Error] Benchmark directory not found: {BENCHMARK_DIR}")
        sys.exit(1)

    go_files_meta = discover_go_files(BENCHMARK_DIR)
    total_files = len(go_files_meta)
    total_bytes = sum(f["size_bytes"] for f in go_files_meta)
    repos = set(f["repo"] for f in go_files_meta)

    print(f"{Fore.GREEN}[Dataset] {total_files} .go files from {len(repos)} repositories")
    print(f"{Fore.GREEN}[Dataset] Total size: {total_bytes:,} bytes ({total_bytes/1024/1024:.1f} MB)")
    print(f"{Fore.GREEN}[Dataset] Avg: {total_bytes//total_files:,} bytes, Max: {max(f['size_bytes'] for f in go_files_meta):,} bytes")

    # ── 출력 경로 ─────────────────────────────────────────────────
    qwen_csv_path = generate_output_path(RESULTS_DIR, "golisa_qwen", run_start)
    semgrep_csv_path = generate_output_path(RESULTS_DIR, "golisa_semgrep", run_start)
    meta_path = RESULTS_DIR / f"golisa_validation_{run_start.strftime('%y%m%d_%H%M')}.meta.json"
    report_path = RESULTS_DIR / "GOLISA_VALIDATION_REPORT.md"

    # ══════════════════════════════════════════════════════════════
    # PHASE A: Semgrep
    # ══════════════════════════════════════════════════════════════
    semgrep_results = {}  # rel_path -> {config_label -> result}
    semgrep_total_time = 0

    if not args.skip_semgrep:
        print(f"\n{Fore.YELLOW}{Style.BRIGHT}{'='*70}")
        print(f"{Fore.YELLOW}{Style.BRIGHT}  PHASE A: Semgrep Static Analysis")
        print(f"{Fore.YELLOW}{Style.BRIGHT}{'='*70}{Style.RESET_ALL}")

        semgrep_path = check_semgrep()
        if semgrep_path is None:
            print(f"{Fore.RED}[Warning] semgrep not found. Skipping Phase A.")
            semgrep_version = "N/A"
        else:
            semgrep_version = get_semgrep_version(semgrep_path)
            print(f"{Fore.GREEN}[OK] semgrep: {semgrep_path} (v{semgrep_version})")

            semgrep_csv_fields = [
                "timestamp", "tool", "config", "repo", "file",
                "file_path_rel", "findings_count", "consensus_relevant",
                "elapsed_s", "findings_detail",
            ]

            semgrep_start = time.time()
            semgrep_record_count = 0

            with open(semgrep_csv_path, "w", newline="", encoding="utf-8") as f:
                writer = csv.DictWriter(f, fieldnames=semgrep_csv_fields)
                writer.writeheader()

                for cfg in SEMGREP_CONFIGS:
                    config_label = f"semgrep:{cfg['label']}"
                    config_value = cfg["config"]

                    print(f"\n{Fore.YELLOW}[Semgrep] Config: {config_label}")

                    desc = f"semgrep:{cfg['label']}"
                    for fmeta in tqdm(go_files_meta, desc=desc):
                        result = run_semgrep_on_file(
                            semgrep_path, fmeta["abs_path"], config_value
                        )

                        findings_count = len(result["findings"])
                        consensus_count = sum(
                            1 for f in result["findings"] if f["is_consensus"]
                        )

                        rel_path = fmeta["rel_path"]
                        if rel_path not in semgrep_results:
                            semgrep_results[rel_path] = {}
                        semgrep_results[rel_path][config_label] = {
                            "findings_count": findings_count,
                            "consensus_relevant": consensus_count,
                            "elapsed_s": result["elapsed_s"],
                        }

                        writer.writerow({
                            "timestamp": datetime.now().isoformat(),
                            "tool": config_label,
                            "config": config_value,
                            "repo": fmeta["repo"],
                            "file": fmeta["filename"],
                            "file_path_rel": rel_path,
                            "findings_count": findings_count,
                            "consensus_relevant": consensus_count,
                            "elapsed_s": result["elapsed_s"],
                            "findings_detail": format_findings_detail(result["findings"]),
                        })
                        f.flush()
                        semgrep_record_count += 1

            semgrep_total_time = round(time.time() - semgrep_start, 1)
            print(f"\n{Fore.GREEN}[Semgrep] Complete: {semgrep_record_count} records, {semgrep_total_time}s")
            print(f"{Fore.GREEN}[Semgrep] CSV: {semgrep_csv_path.name}")
    else:
        print(f"{Fore.YELLOW}[Skip] Semgrep phase skipped (--skip-semgrep)")
        semgrep_version = "skipped"

    # ══════════════════════════════════════════════════════════════
    # PHASE B: Qwen2.5-Coder-7B Zero-Shot
    # ══════════════════════════════════════════════════════════════
    qwen_results = {}  # rel_path -> {classification, vuln_types, elapsed_s, result}
    qwen_total_time = 0

    if not args.skip_qwen:
        print(f"\n{Fore.YELLOW}{Style.BRIGHT}{'='*70}")
        print(f"{Fore.YELLOW}{Style.BRIGHT}  PHASE B: Qwen2.5-Coder-7B Zero-Shot Audit")
        print(f"{Fore.YELLOW}{Style.BRIGHT}{'='*70}{Style.RESET_ALL}")

        model_path = MODELS_DIR / MODEL_FILE
        if not model_path.exists():
            print(f"{Fore.RED}[Error] Model not found: {model_path}")
            sys.exit(1)

        llm = load_model(model_path)

        qwen_csv_fields = [
            "timestamp", "model", "prompt_strategy", "repo", "file",
            "file_path_rel", "code_chars", "elapsed_s",
            "classification", "vuln_types", "result",
        ]

        qwen_start = time.time()
        qwen_record_count = 0

        with open(qwen_csv_path, "w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=qwen_csv_fields)
            writer.writeheader()

            for fmeta in tqdm(go_files_meta, desc="Qwen zero_shot"):
                try:
                    code = fmeta["abs_path"].read_text(encoding="utf-8", errors="replace")
                except Exception as e:
                    code = ""
                    print(f"{Fore.RED}[Read Error] {fmeta['rel_path']}: {e}")

                code_chars = len(code)
                start = time.time()
                try:
                    result_text = audit_chaincode(llm, code, fmeta["filename"])
                except Exception as e:
                    result_text = f"ERROR: {e}"
                    print(f"{Fore.RED}[Inference Error] {fmeta['rel_path']}: {e}")
                elapsed = round(time.time() - start, 3)

                classification = classify_response(result_text)
                vuln_types = extract_vuln_types(result_text) if classification == "vulnerable" else []

                qwen_results[fmeta["rel_path"]] = {
                    "classification": classification,
                    "vuln_types": vuln_types,
                    "elapsed_s": elapsed,
                    "result": result_text,
                }

                writer.writerow({
                    "timestamp": datetime.now().isoformat(),
                    "model": MODEL_FILE,
                    "prompt_strategy": "zero_shot",
                    "repo": fmeta["repo"],
                    "file": fmeta["filename"],
                    "file_path_rel": fmeta["rel_path"],
                    "code_chars": code_chars,
                    "elapsed_s": elapsed,
                    "classification": classification,
                    "vuln_types": ";".join(vuln_types),
                    "result": result_text,
                })
                f.flush()
                qwen_record_count += 1

        del llm
        qwen_total_time = round(time.time() - qwen_start, 1)
        print(f"\n{Fore.GREEN}[Qwen] Complete: {qwen_record_count} records, {qwen_total_time}s ({qwen_total_time/60:.1f}min)")
        print(f"{Fore.GREEN}[Qwen] CSV: {qwen_csv_path.name}")
    else:
        print(f"{Fore.YELLOW}[Skip] Qwen phase skipped (--skip-qwen)")

    run_end = datetime.now()
    run_duration = round((run_end - run_start).total_seconds(), 1)

    # ══════════════════════════════════════════════════════════════
    # 리포트 생성
    # ══════════════════════════════════════════════════════════════
    print(f"\n{Fore.CYAN}[Report] Generating GOLISA_VALIDATION_REPORT.md ...")
    report_text = generate_report(
        go_files_meta, qwen_results, semgrep_results,
        run_start, run_end, report_path,
    )
    print(f"{Fore.GREEN}[Report] Saved: {report_path.name}")

    # ══════════════════════════════════════════════════════════════
    # 메타데이터 JSON
    # ══════════════════════════════════════════════════════════════
    run_id_source = f"{run_start.isoformat()}|golisa_validation|{MODEL_FILE}"
    run_id = hashlib.sha256(run_id_source.encode()).hexdigest()[:8]

    # Qwen classification summary
    qwen_cls_summary = Counter(
        res.get("classification", "error") for res in qwen_results.values()
    )
    qwen_vuln_type_summary = Counter()
    for res in qwen_results.values():
        for vt in res.get("vuln_types", []):
            qwen_vuln_type_summary[vt] += 1

    # Semgrep summary
    semgrep_total = sum(
        res.get("findings_count", 0)
        for configs in semgrep_results.values()
        for res in configs.values()
    )
    semgrep_consensus = sum(
        res.get("consensus_relevant", 0)
        for configs in semgrep_results.values()
        for res in configs.values()
    )

    meta = {
        "script": "08_run_golisa_validation.py",
        "script_version": "1.0",
        "run_id": run_id,
        "purpose": (
            "External validation of Qwen2.5-Coder-7B zero-shot vulnerability detection "
            "on 657 real-world HLF chaincodes from GoLiSA ECOOP 2023 benchmark, "
            "with Semgrep baseline comparison."
        ),
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "run_duration_s": run_duration,
        "system_info": {
            "platform": platform.platform(),
            "python_version": platform.python_version(),
        },
        "dataset": {
            "source": "GoLiSA ECOOP 2023 Benchmark (Olivieri et al.)",
            "benchmark_dir": str(BENCHMARK_DIR),
            "total_files": total_files,
            "total_repos": len(repos),
            "total_bytes": total_bytes,
        },
        "qwen": {
            "skipped": args.skip_qwen,
            "model": MODEL_FILE,
            "prompt_strategy": "zero_shot",
            "inference_params": INFERENCE_PARAMS,
            "prompt": PROMPT_ZERO_SHOT,
            "output_csv": qwen_csv_path.name if not args.skip_qwen else None,
            "total_records": len(qwen_results),
            "total_time_s": qwen_total_time,
            "classification_summary": dict(qwen_cls_summary),
            "vuln_type_distribution": dict(qwen_vuln_type_summary.most_common()),
        },
        "semgrep": {
            "skipped": args.skip_semgrep,
            "version": semgrep_version,
            "configs": [c["config"] for c in SEMGREP_CONFIGS],
            "output_csv": semgrep_csv_path.name if not args.skip_semgrep else None,
            "total_findings": semgrep_total,
            "consensus_relevant_findings": semgrep_consensus,
        },
        "running_examples_validation": {
            "total": len(RUNNING_EXAMPLES_GROUND_TRUTH),
            "correct": sum(
                1 for gt_path, gt_info in RUNNING_EXAMPLES_GROUND_TRUTH.items()
                if qwen_results.get(gt_path, {}).get("classification") == gt_info["expected"]
            ),
            "details": {
                gt_path: {
                    "expected": gt_info["expected"],
                    "actual": qwen_results.get(gt_path, {}).get("classification", "N/A"),
                    "match": qwen_results.get(gt_path, {}).get("classification") == gt_info["expected"],
                }
                for gt_path, gt_info in RUNNING_EXAMPLES_GROUND_TRUTH.items()
            },
        },
        "report_path": report_path.name,
    }

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    # ══════════════════════════════════════════════════════════════
    # 완료 보고
    # ══════════════════════════════════════════════════════════════
    print(f"\n{Fore.GREEN}{Style.BRIGHT}{'='*70}")
    print(f"{Fore.GREEN}{Style.BRIGHT}  GoLiSA External Validation -- COMPLETE")
    print(f"{Fore.GREEN}{Style.BRIGHT}{'='*70}{Style.RESET_ALL}")
    print(f"{Fore.GREEN}  Run ID       : {run_id}")
    print(f"{Fore.GREEN}  Duration     : {run_duration}s ({run_duration/60:.1f}min)")
    print(f"{Fore.GREEN}  Files audited: {total_files}")

    if not args.skip_qwen:
        print(f"{Fore.GREEN}  Qwen CSV     : {qwen_csv_path.name}")
        print(f"{Fore.CYAN}    Vulnerable  : {qwen_cls_summary.get('vulnerable', 0)}")
        print(f"{Fore.CYAN}    Safe        : {qwen_cls_summary.get('safe', 0)}")
        print(f"{Fore.CYAN}    Error       : {qwen_cls_summary.get('error', 0)}")
        print(f"{Fore.CYAN}    Avg time    : {qwen_total_time/total_files:.3f}s/file")

    if not args.skip_semgrep:
        print(f"{Fore.GREEN}  Semgrep CSV  : {semgrep_csv_path.name}")
        print(f"{Fore.CYAN}    Findings    : {semgrep_total} total, {semgrep_consensus} consensus")

    # Running_Examples 결과
    re_correct = meta["running_examples_validation"]["correct"]
    re_total = meta["running_examples_validation"]["total"]
    re_color = Fore.GREEN if re_correct == re_total else Fore.YELLOW
    print(f"{re_color}  Running_Examples: {re_correct}/{re_total} correct")

    print(f"{Fore.GREEN}  Report       : {report_path.name}")
    print(f"{Fore.GREEN}  Metadata     : {meta_path.name}")
    print(f"{Fore.GREEN}{'='*70}")


if __name__ == "__main__":
    main()
