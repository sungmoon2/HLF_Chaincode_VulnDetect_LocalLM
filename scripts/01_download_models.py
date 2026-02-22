"""
01_download_models.py
- Qwen2.5-Coder-7B-Instruct (Q4_K_M) 및 Meta-Llama-3.1-8B-Instruct (Q4_K_M) GGUF 모델 다운로드
- 저장 경로: 02_resources/models/
"""

import os
import sys
from pathlib import Path
from huggingface_hub import hf_hub_download
from tqdm import tqdm
from colorama import init, Fore, Style

init(autoreset=True)

# 프로젝트 루트 경로 설정
PROJECT_ROOT = Path(__file__).resolve().parent.parent
MODELS_DIR = PROJECT_ROOT / "02_resources" / "models"

# 다운로드 대상 모델 정의
MODELS = [
    {
        "name": "Qwen2.5-Coder-7B-Instruct (Q4_K_M)",
        "repo_id": "Qwen/Qwen2.5-Coder-7B-Instruct-GGUF",
        "filename": "qwen2.5-coder-7b-instruct-q4_k_m.gguf",
    },
    {
        "name": "Meta-Llama-3.1-8B-Instruct (Q4_K_M)",
        "repo_id": "bartowski/Meta-Llama-3.1-8B-Instruct-GGUF",
        "filename": "Meta-Llama-3.1-8B-Instruct-Q4_K_M.gguf",
    },
]


def download_model(repo_id: str, filename: str, local_dir: Path) -> Path:
    """HuggingFace Hub에서 GGUF 모델 파일을 다운로드한다."""
    print(f"{Fore.CYAN}[Download] {repo_id} -> {filename}")
    path = hf_hub_download(
        repo_id=repo_id,
        filename=filename,
        local_dir=str(local_dir),
        local_dir_use_symlinks=False,
    )
    return Path(path)


def main():
    MODELS_DIR.mkdir(parents=True, exist_ok=True)
    print(f"{Fore.GREEN}[Info] 모델 저장 경로: {MODELS_DIR}")
    print(f"{Fore.GREEN}[Info] 총 {len(MODELS)}개 모델 다운로드 시작\n")

    for i, model in enumerate(MODELS, 1):
        print(f"{Fore.YELLOW}{'='*60}")
        print(f"{Fore.YELLOW}[{i}/{len(MODELS)}] {model['name']}")
        print(f"{Fore.YELLOW}{'='*60}")

        target_path = MODELS_DIR / model["filename"]
        if target_path.exists():
            size_gb = target_path.stat().st_size / (1024**3)
            print(f"{Fore.GREEN}[Skip] 이미 존재함: {target_path.name} ({size_gb:.2f} GB)\n")
            continue

        try:
            downloaded = download_model(model["repo_id"], model["filename"], MODELS_DIR)
            size_gb = downloaded.stat().st_size / (1024**3)
            print(f"{Fore.GREEN}[Done] {downloaded.name} ({size_gb:.2f} GB)\n")
        except Exception as e:
            print(f"{Fore.RED}[Error] {model['name']} 다운로드 실패: {e}\n")
            sys.exit(1)

    print(f"{Fore.GREEN}{'='*60}")
    print(f"{Fore.GREEN}[Complete] 모든 모델 다운로드 완료!")
    print(f"{Fore.GREEN}{'='*60}")


if __name__ == "__main__":
    main()
