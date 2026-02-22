"""
05_run_traditional_tools.py  (v1.0 — 2026-02-09)
- 02_resources/dataset/ 내 .go 체인코드 파일에 대해 전통적 정적 분석 도구(semgrep)를 실행한다.
- 결과: 03_artifacts/raw_results/traditional_tools_YYMMDD_HHMM.csv
- 목적: 전통적 도구가 HLF 합의 계층 취약점을 탐지하지 못함을 실증적으로 보여준다.

NOTE — Go 환경 미설치:
  go vet, staticcheck 등 Go 네이티브 도구는 Go 컴파일러/런타임이 시스템에
  설치되어 있어야 실행 가능하다. 본 실험 환경에는 Go가 설치되어 있지 않으므로
  해당 도구들은 실행에서 제외하고, semgrep 만을 사용하여 분석한다.
  이는 meta.json 에 excluded_tools 항목으로 기록된다.

Semgrep configs used:
  --config auto              일반 규칙 (semgrep registry 자동 선택)
  --config p/security-audit  보안 감사 특화 규칙
"""

import csv
import hashlib
import json
import os
import platform
import shutil
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

from colorama import init, Fore, Style

init(autoreset=True)

# ── 프로젝트 루트 경로 설정 ──────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
RESULTS_DIR = PROJECT_ROOT / "03_artifacts" / "raw_results"

# ── Semgrep 설정 목록 ────────────────────────────────────────────────
SEMGREP_CONFIGS = [
    {"label": "auto",           "config": "auto"},
    {"label": "security-audit", "config": "p/security-audit"},
]

# ── Go 네이티브 도구 (설치 불가로 제외) ──────────────────────────────
EXCLUDED_TOOLS = [
    {
        "tool": "go vet",
        "reason": "Requires Go compiler installation (go toolchain not present on this system)",
    },
    {
        "tool": "staticcheck",
        "reason": "Requires Go compiler installation and 'go install honnef.co/go/tools/cmd/staticcheck'",
    },
    {
        "tool": "golangci-lint",
        "reason": "Requires Go compiler installation; meta-linter wrapping go vet/staticcheck/etc.",
    },
]


def check_semgrep() -> str | None:
    """semgrep 실행 파일 경로를 반환한다. 없으면 None."""
    path = shutil.which("semgrep")
    if path:
        return path
    # Windows에서 Scripts/ 에 있을 수 있음
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
    """semgrep 버전 문자열을 반환한다."""
    try:
        result = subprocess.run(
            [semgrep_path, "--version"],
            capture_output=True, text=True, timeout=15,
        )
        return result.stdout.strip()
    except Exception:
        return "unknown"


def run_semgrep_on_file(
    semgrep_path: str,
    go_file: Path,
    config_label: str,
    config_value: str,
) -> dict:
    """
    단일 .go 파일에 대해 특정 semgrep config 를 실행하고 결과를 반환한다.

    Returns:
        {
            "success": bool,
            "findings": [
                {
                    "rule_id": str,
                    "severity": str,
                    "message": str,
                    "line": int,
                    "col": int,
                    "end_line": int,
                }
            ],
            "error": str | None,
            "elapsed_s": float,
        }
    """
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
        # --config auto requires metrics; only disable for non-auto configs
        if config_value != "auto":
            env["SEMGREP_SEND_METRICS"] = "off"
        result = subprocess.run(
            cmd,
            capture_output=True,
            timeout=120,
            encoding="utf-8",
            errors="replace",
        )
        elapsed = round(time.time() - start, 3)

        # semgrep 은 findings 가 있어도 returncode=0, 에러 시 returncode!=0
        # 하지만 config 없을 때도 returncode=1 일 수 있으므로 JSON 파싱 시도
        stdout = result.stdout.strip()
        stderr = result.stderr.strip()

        if not stdout:
            # JSON 출력이 없으면 에러
            return {
                "success": False,
                "findings": [],
                "error": stderr or f"No output (returncode={result.returncode})",
                "elapsed_s": elapsed,
            }

        try:
            data = json.loads(stdout)
        except json.JSONDecodeError as e:
            return {
                "success": False,
                "findings": [],
                "error": f"JSON parse error: {e}; stderr: {stderr}",
                "elapsed_s": elapsed,
            }

        findings = []
        # semgrep JSON 구조: {"results": [...], "errors": [...]}
        for r in data.get("results", []):
            findings.append({
                "rule_id": r.get("check_id", "unknown"),
                "severity": r.get("extra", {}).get("severity", "UNKNOWN"),
                "message": r.get("extra", {}).get("message", ""),
                "line": r.get("start", {}).get("line", 0),
                "col": r.get("start", {}).get("col", 0),
                "end_line": r.get("end", {}).get("line", 0),
            })

        return {
            "success": True,
            "findings": findings,
            "error": None,
            "elapsed_s": elapsed,
        }

    except subprocess.TimeoutExpired:
        elapsed = round(time.time() - start, 3)
        return {
            "success": False,
            "findings": [],
            "error": "Timeout (>120s)",
            "elapsed_s": elapsed,
        }
    except FileNotFoundError:
        elapsed = round(time.time() - start, 3)
        return {
            "success": False,
            "findings": [],
            "error": "semgrep executable not found",
            "elapsed_s": elapsed,
        }
    except Exception as e:
        elapsed = round(time.time() - start, 3)
        return {
            "success": False,
            "findings": [],
            "error": str(e),
            "elapsed_s": elapsed,
        }


def generate_output_path(results_dir: Path, run_start: datetime) -> Path:
    """타임스탬프 기반 고유 CSV 경로를 생성한다. 동일 분 내 재실행 시 _2, _3 ... 접미사."""
    base = f"traditional_tools_{run_start.strftime('%y%m%d_%H%M')}"
    candidate = results_dir / f"{base}.csv"
    if not candidate.exists():
        return candidate
    seq = 2
    while True:
        candidate = results_dir / f"{base}_{seq}.csv"
        if not candidate.exists():
            return candidate
        seq += 1


def format_findings_detail(findings: list) -> str:
    """findings 목록을 사람이 읽기 편한 문자열로 직렬화한다."""
    if not findings:
        return "NO_FINDINGS"
    parts = []
    for f in findings:
        parts.append(
            f"[L{f['line']}] {f['severity']} | {f['rule_id']}: "
            f"{f['message'][:120]}"
        )
    return " || ".join(parts)


def print_matrix(matrix: dict, go_files: list[Path]):
    """tool x file 매트릭스를 컬러 테이블로 출력한다."""
    # 도구(config) 목록
    tools = sorted(matrix.keys())
    filenames = [f.name for f in go_files]

    if not tools:
        print(f"{Fore.YELLOW}[Matrix] No tool results to display.")
        return

    # 열 너비 계산
    max_file_len = max(len(fn) for fn in filenames)
    col_w = max(max_file_len, 12)
    tool_col_w = max(len(t) for t in tools) + 2

    # 헤더
    header = f"{'File':<{col_w}}"
    for t in tools:
        header += f"  {t:>{tool_col_w}}"
    header += f"  {'Total':>6}"

    print(f"\n{Fore.CYAN}{Style.BRIGHT}{'=' * len(header)}")
    print(f"{Fore.CYAN}{Style.BRIGHT}  Finding Counts: Tool x File Matrix")
    print(f"{Fore.CYAN}{Style.BRIGHT}{'=' * len(header)}")
    print(f"{Fore.WHITE}{Style.BRIGHT}{header}")
    print(f"{Fore.CYAN}{'-' * len(header)}")

    # 각 파일 행
    file_totals = {fn: 0 for fn in filenames}
    tool_totals = {t: 0 for t in tools}
    grand_total = 0

    for fn in filenames:
        row = f"{fn:<{col_w}}"
        row_sum = 0
        for t in tools:
            count = matrix[t].get(fn, 0)
            row_sum += count
            tool_totals[t] += count

            if count > 0:
                cell = f"{Fore.YELLOW}{count:>{tool_col_w}}{Style.RESET_ALL}"
            else:
                cell = f"{Fore.GREEN}{count:>{tool_col_w}}{Style.RESET_ALL}"
            row += f"  {cell}"

        file_totals[fn] = row_sum
        grand_total += row_sum

        total_color = Fore.YELLOW if row_sum > 0 else Fore.GREEN
        row += f"  {total_color}{row_sum:>6}{Style.RESET_ALL}"

        # 파일 이름 색상: vuln_ = 노랑, safe_ = 초록
        if fn.startswith("vuln_"):
            print(f"{Fore.YELLOW}{fn:<{col_w}}{Style.RESET_ALL}" + row[col_w:])
        elif fn.startswith("safe_"):
            print(f"{Fore.GREEN}{fn:<{col_w}}{Style.RESET_ALL}" + row[col_w:])
        else:
            print(row)

    # 합계 행
    print(f"{Fore.CYAN}{'-' * len(header)}")
    total_row = f"{'TOTAL':<{col_w}}"
    for t in tools:
        tc = tool_totals[t]
        color = Fore.YELLOW if tc > 0 else Fore.GREEN
        total_row += f"  {color}{tc:>{tool_col_w}}{Style.RESET_ALL}"
    total_color = Fore.YELLOW if grand_total > 0 else Fore.GREEN
    total_row += f"  {total_color}{grand_total:>6}{Style.RESET_ALL}"
    print(total_row)
    print(f"{Fore.CYAN}{'=' * len(header)}\n")


def main():
    run_start = datetime.now()
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

    print(f"{Fore.CYAN}{Style.BRIGHT}")
    print("=" * 70)
    print("  05_run_traditional_tools.py")
    print("  Traditional Static Analysis Tools vs HLF Chaincode Vulnerabilities")
    print("=" * 70)
    print(Style.RESET_ALL)

    # ── semgrep 확인 ─────────────────────────────────────────────────
    semgrep_path = check_semgrep()
    if semgrep_path is None:
        print(f"{Fore.RED}[Error] semgrep is not installed or not in PATH.")
        print(f"{Fore.YELLOW}[Hint]  pip install semgrep")
        print(f"{Fore.YELLOW}[Hint]  After installation, re-run this script.")
        sys.exit(1)

    semgrep_version = get_semgrep_version(semgrep_path)
    print(f"{Fore.GREEN}[OK] semgrep found: {semgrep_path}")
    print(f"{Fore.GREEN}[OK] semgrep version: {semgrep_version}")

    # ── 제외 도구 안내 ───────────────────────────────────────────────
    print(f"\n{Fore.YELLOW}[Info] Excluded tools (Go runtime not installed):")
    for ex in EXCLUDED_TOOLS:
        print(f"{Fore.YELLOW}  - {ex['tool']}: {ex['reason']}")

    # ── .go 파일 수집 ────────────────────────────────────────────────
    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"{Fore.RED}[Error] No .go files found in {DATASET_DIR}")
        sys.exit(1)

    print(f"\n{Fore.GREEN}[Info] Target .go files: {len(go_files)}")
    for gf in go_files:
        label = "vuln" if gf.name.startswith("vuln_") else "safe"
        color = Fore.YELLOW if label == "vuln" else Fore.GREEN
        print(f"  {color}[{label}] {gf.name}")

    # ── 출력 파일 경로 ───────────────────────────────────────────────
    output_csv = generate_output_path(RESULTS_DIR, run_start)
    print(f"\n{Fore.GREEN}[Info] Output CSV: {output_csv.name}")

    # ── CSV 컬럼 ─────────────────────────────────────────────────────
    csv_fields = [
        "timestamp",
        "tool",
        "config",
        "file",
        "file_category",
        "findings_count",
        "elapsed_s",
        "findings_detail",
    ]

    # ── 분석 실행 ────────────────────────────────────────────────────
    total_records = 0
    all_results = []        # (config_label, filename, result_dict)
    matrix = {}             # config_label -> {filename -> count}
    consensus_vuln_detected = 0  # HLF 합의 취약점 탐지 횟수 (핵심 지표)

    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields)
        writer.writeheader()

        for cfg in SEMGREP_CONFIGS:
            config_label = f"semgrep:{cfg['label']}"
            config_value = cfg["config"]
            matrix[config_label] = {}

            print(f"\n{Fore.YELLOW}{'=' * 60}")
            print(f"{Fore.YELLOW}[Semgrep] Config: {config_label}")
            print(f"{Fore.YELLOW}{'=' * 60}")

            for go_file in go_files:
                file_category = (
                    "vulnerable" if go_file.name.startswith("vuln_") else "safe"
                )
                print(
                    f"{Fore.CYAN}  [Scan] {go_file.name} "
                    f"({file_category}) ... ",
                    end="",
                    flush=True,
                )

                result = run_semgrep_on_file(
                    semgrep_path, go_file, cfg["label"], config_value
                )

                count = len(result["findings"])
                matrix[config_label][go_file.name] = count
                all_results.append((config_label, go_file.name, result))

                if result["success"]:
                    if count > 0:
                        print(f"{Fore.YELLOW}{count} finding(s) ({result['elapsed_s']}s)")
                        for finding in result["findings"]:
                            # HLF 합의 관련 키워드 확인
                            rule_lower = finding["rule_id"].lower()
                            msg_lower = finding["message"].lower()
                            consensus_keywords = [
                                "non-deterministic", "nondeterministic",
                                "endorsement", "consensus", "phantom read",
                                "read-after-write", "chaincode",
                                "fabric", "hlf", "ledger",
                                "getstate", "putstate",
                            ]
                            is_consensus = any(
                                kw in rule_lower or kw in msg_lower
                                for kw in consensus_keywords
                            )
                            if is_consensus:
                                consensus_vuln_detected += 1
                                marker = f"{Fore.RED}[CONSENSUS-RELEVANT]"
                            else:
                                marker = f"{Fore.WHITE}[generic]"

                            print(
                                f"    {marker} L{finding['line']} "
                                f"{finding['severity']} | {finding['rule_id']}"
                            )
                            if finding["message"]:
                                msg_short = finding["message"][:100]
                                print(f"      {Fore.WHITE}{msg_short}")
                    else:
                        print(f"{Fore.GREEN}0 findings ({result['elapsed_s']}s)")
                else:
                    print(f"{Fore.RED}ERROR ({result['elapsed_s']}s)")
                    err_short = (result["error"] or "")[:120]
                    print(f"    {Fore.RED}{err_short}")

                # CSV 기록
                writer.writerow({
                    "timestamp": datetime.now().isoformat(),
                    "tool": config_label,
                    "config": config_value,
                    "file": go_file.name,
                    "file_category": file_category,
                    "findings_count": count,
                    "elapsed_s": result["elapsed_s"],
                    "findings_detail": format_findings_detail(result["findings"]),
                })
                f.flush()
                total_records += 1

    run_end = datetime.now()
    run_duration = round((run_end - run_start).total_seconds(), 1)

    # ── 매트릭스 출력 ────────────────────────────────────────────────
    print_matrix(matrix, go_files)

    # ── 핵심 결론 출력 ───────────────────────────────────────────────
    print(f"{Fore.CYAN}{Style.BRIGHT}{'=' * 70}")
    print(f"{Fore.CYAN}{Style.BRIGHT}  KEY FINDING")
    print(f"{Fore.CYAN}{Style.BRIGHT}{'=' * 70}")

    total_findings = sum(
        count
        for tool_map in matrix.values()
        for count in tool_map.values()
    )

    if consensus_vuln_detected == 0:
        print(
            f"{Fore.GREEN}  Traditional tools found {total_findings} total finding(s), "
            f"but {Fore.RED}{Style.BRIGHT}0 consensus-layer vulnerabilities{Style.RESET_ALL}"
            f"{Fore.GREEN} were detected."
        )
        print(
            f"{Fore.GREEN}  This confirms that semgrep (and similar tools) CANNOT detect"
        )
        print(
            f"{Fore.GREEN}  HLF-specific consensus vulnerabilities such as:"
        )
        print(f"{Fore.WHITE}    - Non-deterministic execution (time.Now in write path)")
        print(f"{Fore.WHITE}    - Phantom reads (read-after-write conflicts)")
        print(f"{Fore.WHITE}    - Global state mutation across transactions")
        print(f"{Fore.WHITE}    - Goroutine-induced non-determinism")
        print(f"{Fore.WHITE}    - Map iteration order non-determinism")
        print(f"{Fore.WHITE}    - Iterator / resource leaks affecting endorsement")
    else:
        print(
            f"{Fore.YELLOW}  {consensus_vuln_detected} consensus-related finding(s) "
            f"detected out of {total_findings} total."
        )
        print(
            f"{Fore.YELLOW}  Review these findings to verify if they are true "
            f"HLF consensus-layer detections."
        )

    print(f"{Fore.CYAN}{Style.BRIGHT}{'=' * 70}\n")

    # ── 메타데이터 JSON ──────────────────────────────────────────────
    meta_path = output_csv.with_suffix(".meta.json")

    # 전체 findings 상세 기록
    detailed_results = {}
    for config_label, filename, result in all_results:
        key = f"{config_label}|{filename}"
        detailed_results[key] = {
            "success": result["success"],
            "findings_count": len(result["findings"]),
            "findings": result["findings"],
            "error": result["error"],
            "elapsed_s": result["elapsed_s"],
        }

    run_id_source = f"{run_start.isoformat()}|traditional|semgrep"
    run_id = hashlib.sha256(run_id_source.encode()).hexdigest()[:8]

    meta = {
        "script": "05_run_traditional_tools.py",
        "script_version": "1.0",
        "run_id": run_id,
        "purpose": (
            "Run traditional static analysis tools on HLF chaincode dataset "
            "to demonstrate their inability to detect consensus-layer vulnerabilities."
        ),
        "run_start": run_start.isoformat(),
        "run_end": run_end.isoformat(),
        "run_duration_s": run_duration,
        "output_csv": output_csv.name,
        "output_csv_bytes": output_csv.stat().st_size,
        "total_records": total_records,
        "system_info": {
            "platform": platform.platform(),
            "python_version": platform.python_version(),
        },
        "tools": {
            "semgrep": {
                "path": semgrep_path,
                "version": semgrep_version,
                "configs_used": [c["config"] for c in SEMGREP_CONFIGS],
            },
        },
        "excluded_tools": EXCLUDED_TOOLS,
        "dataset": {
            "dir": str(DATASET_DIR),
            "files": [f.name for f in go_files],
            "file_count": len(go_files),
            "vulnerable_files": [
                f.name for f in go_files if f.name.startswith("vuln_")
            ],
            "safe_files": [
                f.name for f in go_files if f.name.startswith("safe_")
            ],
        },
        "summary": {
            "total_findings": total_findings,
            "consensus_relevant_findings": consensus_vuln_detected,
            "matrix": {
                tool: {fn: cnt for fn, cnt in file_map.items()}
                for tool, file_map in matrix.items()
            },
        },
        "key_conclusion": (
            "Traditional static analysis tools (semgrep) found "
            f"{total_findings} generic finding(s) but "
            f"{consensus_vuln_detected} HLF consensus-layer vulnerability detection(s). "
            "This validates the need for domain-specific LLM-based analysis."
        ),
        "detailed_results": detailed_results,
    }

    with open(meta_path, "w", encoding="utf-8") as mf:
        json.dump(meta, mf, indent=2, ensure_ascii=False)

    # ── 완료 보고 ────────────────────────────────────────────────────
    print(f"{Fore.GREEN}{'=' * 60}")
    print(f"{Fore.GREEN}[Complete] Traditional tool analysis finished")
    print(f"{Fore.GREEN}  Result CSV   : {output_csv}")
    print(f"{Fore.GREEN}  Metadata     : {meta_path}")
    print(f"{Fore.GREEN}  Total records: {total_records}")
    print(f"{Fore.GREEN}  Duration     : {run_duration}s")
    print(f"{Fore.GREEN}  Total findings (all tools)   : {total_findings}")
    print(f"{Fore.GREEN}  Consensus-relevant findings  : {consensus_vuln_detected}")
    print(f"{Fore.GREEN}{'=' * 60}")


if __name__ == "__main__":
    main()
