"""
07_extract_golisa.py — GoLiSA OVA Artifact에서 .go 체인코드 파일 추출

사용법:
  python scripts/07_extract_golisa.py

파이프라인:
  1. OVA (tar) → VMDK 추출 (tar)
  2. VMDK → 7z로 파일 시스템 탐색 → .go 파일 추출
  3. 추출된 .go 파일을 02_resources/golisa_dataset/에 복사

필요 도구: tar, 7z (7-zip)
"""

import subprocess
import os
import sys
import shutil
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent
ARTIFACT_DIR = PROJECT_ROOT / "02_resources" / "golisa_artifact"
OVA_FILE = ARTIFACT_DIR / "golisa_ecoop2023.ova"
EXTRACT_DIR = ARTIFACT_DIR / "extracted"
OUTPUT_DIR = PROJECT_ROOT / "02_resources" / "golisa_dataset"

SEVEN_ZIP = r"C:\Program Files\7-Zip\7z.exe"


def check_prereqs():
    if not OVA_FILE.exists():
        print(f"[ERROR] OVA file not found: {OVA_FILE}")
        print(f"  Expected size: ~5.35 GB")
        sys.exit(1)

    ova_size = OVA_FILE.stat().st_size
    expected = 5355389952
    if ova_size < expected * 0.99:
        print(f"[WARNING] OVA file may be incomplete: {ova_size:,} bytes (expected {expected:,})")
        print(f"  Progress: {ova_size/expected*100:.1f}%")
        sys.exit(1)

    if not Path(SEVEN_ZIP).exists():
        print(f"[ERROR] 7-Zip not found at {SEVEN_ZIP}")
        print("  Install: winget install 7zip.7zip")
        sys.exit(1)

    print(f"[OK] OVA: {ova_size:,} bytes")
    print(f"[OK] 7-Zip: {SEVEN_ZIP}")


def step1_extract_ova():
    """OVA (tar archive) → VMDK + OVF 추출"""
    print("\n=== Step 1: OVA → VMDK ===")
    EXTRACT_DIR.mkdir(parents=True, exist_ok=True)

    cmd = ["tar", "xf", str(OVA_FILE), "-C", str(EXTRACT_DIR)]
    print(f"  Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=600)

    if result.returncode != 0:
        print(f"  [ERROR] tar failed: {result.stderr}")
        sys.exit(1)

    files = list(EXTRACT_DIR.iterdir())
    print(f"  Extracted {len(files)} files:")
    for f in files:
        size_mb = f.stat().st_size / (1024 * 1024)
        print(f"    {f.name} ({size_mb:.1f} MB)")

    vmdk_files = [f for f in files if f.suffix.lower() == '.vmdk']
    if not vmdk_files:
        print("  [ERROR] No .vmdk file found in OVA")
        print("  Available files:", [f.name for f in files])
        sys.exit(1)

    return vmdk_files[0]


def step2_list_vmdk_contents(vmdk_path):
    """7z로 VMDK 내부 파일 시스템 탐색"""
    print(f"\n=== Step 2: VMDK 내부 탐색 ===")
    print(f"  VMDK: {vmdk_path.name}")

    cmd = [SEVEN_ZIP, "l", str(vmdk_path), "-r", "*.go"]
    print(f"  Running: 7z l {vmdk_path.name} -r *.go")
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)

    if result.returncode != 0:
        print(f"  [WARNING] 7z list failed, trying nested extraction...")
        return step2_nested_extraction(vmdk_path)

    lines = result.stdout.strip().split('\n')
    go_files = [l for l in lines if '.go' in l.lower()]
    print(f"  Found {len(go_files)} .go references")
    for line in go_files[:20]:
        print(f"    {line.strip()}")
    if len(go_files) > 20:
        print(f"    ... and {len(go_files) - 20} more")

    return go_files


def step2_nested_extraction(vmdk_path):
    """VMDK를 단계적으로 추출 (7z가 직접 탐색 실패 시)"""
    print("  Attempting nested extraction...")

    nested_dir = EXTRACT_DIR / "vmdk_contents"
    nested_dir.mkdir(parents=True, exist_ok=True)

    cmd = [SEVEN_ZIP, "x", str(vmdk_path), "-o" + str(nested_dir), "-y"]
    print(f"  Running: 7z x {vmdk_path.name}")
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=600)

    if result.returncode != 0:
        print(f"  [ERROR] 7z extraction failed: {result.stderr[:500]}")
        return []

    # 추출된 디렉토리에서 .go 파일 탐색
    go_files = list(nested_dir.rglob("*.go"))
    print(f"  Found {len(go_files)} .go files after extraction")
    return go_files


def step3_extract_go_files(vmdk_path):
    """VMDK에서 .go 파일만 추출"""
    print(f"\n=== Step 3: .go 파일 추출 ===")
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    # 먼저 전체 추출 시도
    nested_dir = EXTRACT_DIR / "vmdk_contents"
    if not nested_dir.exists():
        nested_dir.mkdir(parents=True, exist_ok=True)
        cmd = [SEVEN_ZIP, "x", str(vmdk_path), "-o" + str(nested_dir), "-y"]
        print(f"  Extracting VMDK contents...")
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=1200)
        if result.returncode != 0:
            print(f"  [ERROR] Extraction failed")
            return 0

    # .go 파일 검색 및 복사
    go_files = list(nested_dir.rglob("*.go"))
    print(f"  Found {len(go_files)} .go files total")

    # 체인코드 관련 .go 파일 필터링 (fabric/chaincode 관련)
    chaincode_files = []
    other_go_files = []
    for f in go_files:
        content = ""
        try:
            content = f.read_text(encoding='utf-8', errors='ignore')[:2000]
        except Exception:
            pass

        if any(kw in content for kw in [
            'shim.ChaincodeStubInterface',
            'peer.Response',
            'stub.PutState',
            'stub.GetState',
            'chaincode',
            'fabric'
        ]):
            chaincode_files.append(f)
        else:
            other_go_files.append(f)

    print(f"  Chaincode-related: {len(chaincode_files)}")
    print(f"  Other Go files: {len(other_go_files)}")

    # 체인코드 파일 복사
    copied = 0
    for src in chaincode_files:
        dst_name = src.name
        # 파일명 충돌 방지
        counter = 1
        dst = OUTPUT_DIR / dst_name
        while dst.exists():
            stem = src.stem
            dst = OUTPUT_DIR / f"{stem}_{counter}.go"
            counter += 1

        shutil.copy2(src, dst)
        copied += 1

    print(f"\n  Copied {copied} chaincode .go files to {OUTPUT_DIR}")
    return copied


def step4_summary():
    """추출 결과 요약"""
    print(f"\n=== Summary ===")
    if not OUTPUT_DIR.exists():
        print("  No output directory found")
        return

    go_files = sorted(OUTPUT_DIR.glob("*.go"))
    total_size = sum(f.stat().st_size for f in go_files)
    print(f"  Output: {OUTPUT_DIR}")
    print(f"  Files: {len(go_files)} .go")
    print(f"  Total size: {total_size:,} bytes ({total_size/(1024*1024):.1f} MB)")

    if len(go_files) > 0:
        print(f"\n  First 10 files:")
        for f in go_files[:10]:
            print(f"    {f.name} ({f.stat().st_size:,} bytes)")
        if len(go_files) > 10:
            print(f"    ... and {len(go_files) - 10} more")


if __name__ == "__main__":
    print("=" * 60)
    print("GoLiSA OVA Artifact → .go Chaincode Extraction")
    print("=" * 60)

    check_prereqs()
    vmdk_path = step1_extract_ova()
    step2_list_vmdk_contents(vmdk_path)
    count = step3_extract_go_files(vmdk_path)
    step4_summary()

    if count > 0:
        print(f"\n[DONE] {count} files ready for audit.")
        print(f"  Run: python scripts/02_run_audit_v3.py --dataset-dir 02_resources/golisa_dataset --tag golisa")
    else:
        print("\n[WARNING] No chaincode files extracted. Manual inspection of VM may be needed.")
