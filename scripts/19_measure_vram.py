"""Measure peak GPU VRAM usage during local model inference.

Runs one warm-up file then 15-file micro-benchmark under P1 (zero-shot).
Samples nvidia-smi every 200ms in a background thread.
Reports peak VRAM for each model.
"""
import subprocess
import threading
import time
import os
import glob
import json

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_DIR = os.path.normpath(os.path.join(SCRIPT_DIR, os.pardir))
MODEL_DIR = os.path.join(PROJECT_DIR, "02_resources", "models")
DATASET_DIR = os.path.join(PROJECT_DIR, "02_resources", "dataset")
RESULTS_DIR = os.path.join(PROJECT_DIR, "03_artifacts", "raw_results")

MODELS = {
    "Qwen2.5-Coder-7B": "qwen2.5-coder-7b-instruct-q4_k_m.gguf",
    "Llama-3.1-8B": "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
}

SYSTEM_PROMPT = """You are a Hyperledger Fabric chaincode security auditor.
Analyze the following Go chaincode for consensus-layer vulnerabilities that could cause endorsement mismatch.
Focus on these 6 categories:
1. Nondeterministic timestamps (time.Now())
2. Global variable mutation across invocations
3. Goroutine concurrency hazards
4. Map iteration order nondeterminism
5. Phantom reads / MVCC conflicts
6. Iterator resource leaks (missing Close())

For each finding, state the vulnerability type, affected function, and relevant code.
End with an overall verdict: VULNERABLE or SAFE."""


def sample_vram(interval_s, stop_event, samples):
    """Background thread: sample GPU memory.used via nvidia-smi."""
    while not stop_event.is_set():
        try:
            out = subprocess.check_output(
                ["nvidia-smi", "--query-gpu=memory.used", "--format=csv,noheader,nounits"],
                text=True,
            ).strip()
            samples.append(int(out))
        except Exception:
            pass
        time.sleep(interval_s)


def run_model(model_name, model_file):
    from llama_cpp import Llama

    model_path = os.path.join(MODEL_DIR, model_file)
    if not os.path.exists(model_path):
        print(f"  [SKIP] {model_name}: model file not found at {model_path}")
        return None

    go_files = sorted([
        os.path.join(DATASET_DIR, f)
        for f in os.listdir(DATASET_DIR)
        if f.endswith(".go")
    ])
    if not go_files:
        print("  [ERROR] No .go files found in dataset directory")
        return None

    # Start VRAM sampling
    samples = []
    stop_event = threading.Event()
    sampler = threading.Thread(target=sample_vram, args=(0.2, stop_event, samples))
    sampler.start()

    print(f"  Loading {model_name}...")
    llm = Llama(
        model_path=model_path,
        n_gpu_layers=-1,
        n_ctx=4096,
        verbose=False,
    )

    # Warm-up: 1 file
    print(f"  Warm-up inference...")
    with open(go_files[0], "r", encoding="utf-8") as f:
        warmup_code = f.read()
    llm.create_chat_completion(
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": warmup_code},
        ],
        temperature=0.1,
        max_tokens=2048,
    )

    # Benchmark: 15 files
    print(f"  Running 15-file benchmark...")
    t0 = time.time()
    for gf in go_files:
        with open(gf, "r", encoding="utf-8") as f:
            code = f.read()
        llm.create_chat_completion(
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": code},
            ],
            temperature=0.1,
            max_tokens=2048,
        )
    elapsed = time.time() - t0

    # Stop sampling
    stop_event.set()
    sampler.join()

    # Cleanup
    del llm

    peak_vram = max(samples) if samples else 0
    avg_per_file = elapsed / len(go_files)

    print(f"  {model_name}: {elapsed:.1f}s total, {avg_per_file:.2f}s/file, peak VRAM {peak_vram} MiB")
    return {
        "model": model_name,
        "total_s": round(elapsed, 1),
        "avg_s_per_file": round(avg_per_file, 2),
        "peak_vram_mib": peak_vram,
        "n_samples": len(samples),
        "files": len(go_files),
    }


def main():
    results = []
    for name, fname in MODELS.items():
        print(f"\n=== {name} ===")
        r = run_model(name, fname)
        if r:
            results.append(r)
        # Wait for GPU memory to settle
        time.sleep(5)

    print("\n=== RESULTS ===")
    for r in results:
        print(f"  {r['model']}: {r['total_s']}s total, {r['avg_s_per_file']}s/file, peak VRAM {r['peak_vram_mib']} MiB")

    # Save to JSON
    out_path = os.path.join(RESULTS_DIR, "vram_measurement.json")
    with open(out_path, "w") as f:
        json.dump(results, f, indent=2)
    print(f"\nSaved to {out_path}")


if __name__ == "__main__":
    main()
