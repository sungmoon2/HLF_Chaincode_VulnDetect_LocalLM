#!/usr/bin/env python3
"""
GoLiSA-derived benchmark candidate mining script.
6-family grep/regex mining + phantom_read miner + file metadata.
Semgrep is NOT used (baseline leakage prevention).
"""

import os
import re
import json
import hashlib
from pathlib import Path
from datetime import datetime
from itertools import combinations

BENCHMARK_DIR = Path(__file__).parent.parent / "02_resources" / "golisa_benchmark" / "Benchmark"
OUTPUT_DIR = Path(__file__).parent.parent / "06_addon_validation" / "golisa_mining"
OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

# === Family regex patterns (grep/regex only, NO Semgrep) ===

PATTERNS = {
    "TIME_NOW": [
        re.compile(r'\btime\.Now\s*\('),
        re.compile(r'\btime\.Since\s*\('),
        re.compile(r'\btime\.Until\s*\('),
    ],
    "GOROUTINE": [
        re.compile(r'\bgo\s+func\s*\('),
        re.compile(r'\bgo\s+\w+\s*\('),
    ],
    "MAP_ITERATION": [
        re.compile(r'\bfor\s+\w+\s*,\s*\w+\s*:?=\s*range\b'),
        re.compile(r'\bfor\s+\w+\s*:?=\s*range\b'),
    ],
    "PHANTOM_READ": [
        re.compile(r'\bGetQueryResult\s*\('),
        re.compile(r'\bGetPrivateDataQueryResult\s*\('),
        re.compile(r'\bGetQueryResultWithPagination\s*\('),
    ],
    "ITERATOR_LEAK": [
        re.compile(r'\bGetStateByRange\s*\('),
        re.compile(r'\bGetStateByPartialCompositeKey\s*\('),
        re.compile(r'\bGetHistoryForKey\s*\('),
        re.compile(r'\bGetPrivateDataByRange\s*\('),
    ],
    "GLOBAL_MUTABLE_STATE": [
        re.compile(r'^var\s+\w+', re.MULTILINE),
    ],
}

# Sink patterns for source→sink verification
SINK_PATTERNS = [
    re.compile(r'\bPutState\s*\('),
    re.compile(r'\bDelState\s*\('),
    re.compile(r'\bPutPrivateData\s*\('),
    re.compile(r'\bDelPrivateData\s*\('),
    re.compile(r'\bPurgePrivateData\s*\('),
    re.compile(r'\bSetStateValidationParameter\s*\('),
    re.compile(r'\bSetEvent\s*\('),
]

# Close pattern for iterator leak check
CLOSE_PATTERN = re.compile(r'\.Close\s*\(')

# Chaincode indicator patterns
CHAINCODE_INDICATORS = [
    re.compile(r'\bshim\.ChaincodeStubInterface\b'),
    re.compile(r'\bshim\.Chaincode\b'),
    re.compile(r'\bcontractapi\.ContractInterface\b'),
    re.compile(r'\bcontractapi\.TransactionContextInterface\b'),
    re.compile(r'\bfunc\s*\(\s*\w+\s+\*?\w+\s*\)\s*Invoke\s*\('),
    re.compile(r'\bfunc\s*\(\s*\w+\s+\*?\w+\s*\)\s*Init\s*\('),
    re.compile(r'shim\.Start\s*\('),
    re.compile(r'contractapi\.NewChaincode\s*\('),
]

def get_repo_id(filepath):
    """Extract repo directory name as repo_id."""
    parts = Path(filepath).relative_to(BENCHMARK_DIR).parts
    return parts[0] if parts else "unknown"

def compute_file_hash(content):
    """Normalized hash for exact-duplicate detection."""
    normalized = re.sub(r'\s+', ' ', content.strip())
    return hashlib.sha256(normalized.encode('utf-8')).hexdigest()[:16]

def tokenize_for_similarity(content):
    """Tokenize Go source into a set of 3-grams for Jaccard similarity."""
    tokens = re.findall(r'[a-zA-Z_]\w*|[{}()\[\];,.<>=!+\-*/&|^%:]', content)
    if len(tokens) < 3:
        return set()
    return set(tuple(tokens[i:i+3]) for i in range(len(tokens) - 2))

def jaccard_similarity(ngrams_a, ngrams_b):
    """Compute Jaccard similarity between two n-gram sets."""
    if not ngrams_a or not ngrams_b:
        return 0.0
    intersection = len(ngrams_a & ngrams_b)
    union = len(ngrams_a | ngrams_b)
    return intersection / union if union > 0 else 0.0

def deduplicate_by_similarity(candidates, threshold=0.85):
    """Remove near-duplicates using pairwise Jaccard similarity on token 3-grams."""
    contents = {}
    for c in candidates:
        try:
            with open(c["filepath"], 'r', encoding='utf-8', errors='replace') as f:
                contents[c["filepath"]] = f.read()
        except Exception:
            contents[c["filepath"]] = ""

    ngrams = {fp: tokenize_for_similarity(ct) for fp, ct in contents.items()}

    # Union-Find for grouping similar files
    parent = {c["filepath"]: c["filepath"] for c in candidates}

    def find(x):
        while parent[x] != x:
            parent[x] = parent[parent[x]]
            x = parent[x]
        return x

    def union(a, b):
        ra, rb = find(a), find(b)
        if ra != rb:
            parent[ra] = rb

    paths = [c["filepath"] for c in candidates]
    for i, j in combinations(range(len(paths)), 2):
        if find(paths[i]) == find(paths[j]):
            continue
        sim = jaccard_similarity(ngrams[paths[i]], ngrams[paths[j]])
        if sim >= threshold:
            union(paths[i], paths[j])

    # Group by root and keep the first (alphabetical by repo_id) per group
    groups = {}
    path_to_candidate = {c["filepath"]: c for c in candidates}
    for c in candidates:
        root = find(c["filepath"])
        if root not in groups:
            groups[root] = []
        groups[root].append(c)

    kept = []
    sim_groups = []
    for root, members in groups.items():
        members.sort(key=lambda x: x["repo_id"])
        kept.append(members[0])
        if len(members) > 1:
            sim_groups.append([m["repo_id"] + "/" + m["filename"] for m in members])

    return kept, sim_groups

def has_close_on_all_paths(content, iterator_line_num):
    """Simple heuristic: check if .Close() exists after the iterator creation."""
    lines = content.split('\n')
    remaining = '\n'.join(lines[iterator_line_num:])
    return bool(CLOSE_PATTERN.search(remaining))

def check_map_type(content, match_line):
    """Check if the range variable is actually a map (heuristic)."""
    range_match = re.search(r'range\s+(\w+)', match_line)
    if not range_match:
        return False
    var_name = range_match.group(1)
    map_decl = re.compile(rf'\b{re.escape(var_name)}\b.*\bmap\[')
    make_map = re.compile(rf'{re.escape(var_name)}\s*[:=]+\s*make\s*\(\s*map\[')
    map_literal = re.compile(rf'{re.escape(var_name)}\s*[:=]+\s*map\[')
    return bool(map_decl.search(content) or make_map.search(content) or map_literal.search(content))

def mine_file(filepath):
    """Mine a single .go file for candidate vulnerabilities."""
    try:
        with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
            content = f.read()
    except Exception as e:
        return None

    file_size = len(content.encode('utf-8'))
    line_count = content.count('\n') + 1
    repo_id = get_repo_id(filepath)
    file_hash = compute_file_hash(content)

    # Check if it's a chaincode file
    is_chaincode = any(p.search(content) for p in CHAINCODE_INDICATORS)

    # Check for sinks
    has_sink = any(p.search(content) for p in SINK_PATTERNS)

    # Mine each family
    hits = {}
    for family, patterns in PATTERNS.items():
        family_hits = []
        for pattern in patterns:
            for match in pattern.finditer(content):
                line_num = content[:match.start()].count('\n') + 1
                line_text = content.split('\n')[line_num - 1].strip()
                family_hits.append({
                    "line": line_num,
                    "match": match.group().strip(),
                    "context": line_text[:120],
                })
        if family_hits:
            hits[family] = family_hits

    # Refine GOROUTINE: filter out "go " in comments/strings (rough heuristic)
    if "GOROUTINE" in hits:
        refined = []
        for h in hits["GOROUTINE"]:
            line = h["context"]
            if not line.strip().startswith("//") and not line.strip().startswith("/*"):
                refined.append(h)
        if refined:
            hits["GOROUTINE"] = refined
        else:
            del hits["GOROUTINE"]

    # Refine MAP_ITERATION: check if range variable is actually a map
    if "MAP_ITERATION" in hits:
        refined = []
        for h in hits["MAP_ITERATION"]:
            if check_map_type(content, h["context"]):
                refined.append(h)
        if refined:
            hits["MAP_ITERATION"] = refined
        else:
            del hits["MAP_ITERATION"]

    # Refine ITERATOR_LEAK: check if .Close() exists after each iterator
    if "ITERATOR_LEAK" in hits:
        has_close = bool(CLOSE_PATTERN.search(content))
        if has_close:
            # Has at least one Close() — may still leak on some paths
            # but reduce confidence
            for h in hits["ITERATOR_LEAK"]:
                h["close_found"] = True
        else:
            for h in hits["ITERATOR_LEAK"]:
                h["close_found"] = False

    # Refine GLOBAL_MUTABLE_STATE: filter constants, loggers, types
    if "GLOBAL_MUTABLE_STATE" in hits:
        refined = []
        for h in hits["GLOBAL_MUTABLE_STATE"]:
            line = h["context"]
            # Skip: const, type, func, logger/log patterns, empty
            if any(line.strip().startswith(kw) for kw in ["const ", "type ", "func ", "//"]):
                continue
            # Skip common safe globals
            safe_globals = ["logger", "log", "Logger", "Log", "ErrNil", "err"]
            if any(sg in line for sg in safe_globals):
                continue
            refined.append(h)
        if refined:
            hits["GLOBAL_MUTABLE_STATE"] = refined
        else:
            del hits["GLOBAL_MUTABLE_STATE"]

    # Determine possible families
    possible_families = list(hits.keys())

    return {
        "filepath": str(filepath),
        "repo_id": repo_id,
        "filename": Path(filepath).name,
        "file_size": file_size,
        "line_count": line_count,
        "file_hash": file_hash,
        "is_chaincode": is_chaincode,
        "has_sink": has_sink,
        "possible_families": possible_families,
        "hits": {k: v for k, v in hits.items()},
        "hit_count": sum(len(v) for v in hits.values()),
    }


def main():
    print(f"Mining GoLiSA benchmark: {BENCHMARK_DIR}")
    print(f"Output directory: {OUTPUT_DIR}")

    # Collect all .go files
    go_files = sorted(BENCHMARK_DIR.rglob("*.go"))
    print(f"Total .go files: {len(go_files)}")

    # Mine all files
    results = []
    for f in go_files:
        r = mine_file(f)
        if r:
            results.append(r)

    # === Statistics ===
    total = len(results)
    chaincode_files = [r for r in results if r["is_chaincode"]]
    non_chaincode = [r for r in results if not r["is_chaincode"]]
    with_hits = [r for r in results if r["possible_families"]]
    with_sink = [r for r in results if r["has_sink"]]
    chaincode_with_hits = [r for r in chaincode_files if r["possible_families"]]
    chaincode_with_sink = [r for r in chaincode_files if r["has_sink"]]

    # Family distribution
    family_counts = {}
    for r in results:
        for fam in r["possible_families"]:
            family_counts[fam] = family_counts.get(fam, 0) + 1

    # Family distribution (chaincode only)
    cc_family_counts = {}
    for r in chaincode_files:
        for fam in r["possible_families"]:
            cc_family_counts[fam] = cc_family_counts.get(fam, 0) + 1

    # File size distribution
    sizes = [r["file_size"] for r in results]
    line_counts = [r["line_count"] for r in results]

    # Near-duplicate detection
    hash_groups = {}
    for r in results:
        h = r["file_hash"]
        if h not in hash_groups:
            hash_groups[h] = []
        hash_groups[h].append(r["filepath"])
    duplicates = {h: fps for h, fps in hash_groups.items() if len(fps) > 1}

    # Repo distribution
    repos = set(r["repo_id"] for r in results)
    repos_with_hits = set(r["repo_id"] for r in with_hits)

    # === Candidate pool: chaincode files with hits AND sinks ===
    candidates = [r for r in chaincode_files if r["possible_families"] and r["has_sink"]]

    # One-file-per-repo for candidates
    repo_best = {}
    for c in candidates:
        rid = c["repo_id"]
        if rid not in repo_best or c["hit_count"] > repo_best[rid]["hit_count"]:
            repo_best[rid] = c
    per_repo_candidates = list(repo_best.values())

    # One-file-per-content: deduplicate by file_hash (keep first by repo_id alphabetical)
    hash_best = {}
    for c in sorted(per_repo_candidates, key=lambda x: x["repo_id"]):
        h = c["file_hash"]
        if h not in hash_best:
            hash_best[h] = c
    hash_deduped_candidates = list(hash_best.values())
    content_dupes_removed = len(per_repo_candidates) - len(hash_deduped_candidates)

    # Near-duplicate removal by Jaccard similarity on token 3-grams
    unique_candidates, sim_duplicate_groups = deduplicate_by_similarity(hash_deduped_candidates, threshold=0.85)
    sim_dupes_removed = len(hash_deduped_candidates) - len(unique_candidates)

    # === Print summary ===
    print(f"\n{'='*60}")
    print(f"MINING RESULTS SUMMARY")
    print(f"{'='*60}")
    print(f"Total .go files:            {total}")
    print(f"Chaincode files:            {len(chaincode_files)}")
    print(f"Non-chaincode files:        {len(non_chaincode)}")
    print(f"Files with any hit:         {len(with_hits)}")
    print(f"Files with sink:            {len(with_sink)}")
    print(f"Chaincode + hits:           {len(chaincode_with_hits)}")
    print(f"Chaincode + hits + sink:    {len(candidates)}")
    print(f"Unique repos:               {len(repos)}")
    print(f"Repos with hits:            {len(repos_with_hits)}")
    print(f"One-per-repo candidates:    {len(per_repo_candidates)}")
    print(f"Content-dedup candidates:   {len(hash_deduped_candidates)} (removed {content_dupes_removed})")
    print(f"Similarity-dedup candidates:{len(unique_candidates)} (removed {sim_dupes_removed}, threshold=0.85)")
    print(f"Similarity-duplicate groups:{len(sim_duplicate_groups)}")
    print(f"Near-duplicate groups:      {len(duplicates)}")

    print(f"\nFamily distribution (all files):")
    for fam in ["TIME_NOW", "GOROUTINE", "MAP_ITERATION", "PHANTOM_READ", "ITERATOR_LEAK", "GLOBAL_MUTABLE_STATE"]:
        print(f"  {fam:30s}: {family_counts.get(fam, 0)}")

    print(f"\nFamily distribution (chaincode only):")
    for fam in ["TIME_NOW", "GOROUTINE", "MAP_ITERATION", "PHANTOM_READ", "ITERATOR_LEAK", "GLOBAL_MUTABLE_STATE"]:
        print(f"  {fam:30s}: {cc_family_counts.get(fam, 0)}")

    print(f"\nFile size distribution:")
    print(f"  Min: {min(sizes)} bytes ({min(line_counts)} lines)")
    print(f"  Max: {max(sizes)} bytes ({max(line_counts)} lines)")
    print(f"  Mean: {sum(sizes)//len(sizes)} bytes ({sum(line_counts)//len(line_counts)} lines)")
    p90_size = sorted(sizes)[int(len(sizes)*0.9)]
    p90_lines = sorted(line_counts)[int(len(line_counts)*0.9)]
    print(f"  P90: {p90_size} bytes ({p90_lines} lines)")

    # Context budget check (prompt ~800 tokens ≈ 3200 chars + code)
    PROMPT_CHARS = 3200
    CONTEXT_LIMIT = 16384 * 4  # ~4 chars per token, rough estimate
    over_budget = [r for r in results if r["file_size"] + PROMPT_CHARS > CONTEXT_LIMIT]
    print(f"\n  Files exceeding context budget (>{CONTEXT_LIMIT} chars): {len(over_budget)}")
    if over_budget:
        for ob in over_budget[:5]:
            print(f"    {ob['repo_id']}/{ob['filename']}: {ob['file_size']} bytes")

    if duplicates:
        print(f"\nExact-duplicates ({len(duplicates)} groups):")
        for h, fps in list(duplicates.items())[:5]:
            print(f"  Hash {h}:")
            for fp in fps[:3]:
                print(f"    {fp}")
            if len(fps) > 3:
                print(f"    ... +{len(fps)-3} more")

    if sim_duplicate_groups:
        print(f"\nSimilarity-duplicate groups (Jaccard >= 0.85):")
        for g in sim_duplicate_groups[:10]:
            print(f"  Group ({len(g)} files): {', '.join(g[:3])}")
            if len(g) > 3:
                print(f"    ... +{len(g)-3} more")

    # === Save outputs ===
    timestamp = datetime.now().strftime("%y%m%d_%H%M")

    # 1. Full mining results
    full_path = OUTPUT_DIR / f"mining_full_{timestamp}.json"
    with open(full_path, 'w', encoding='utf-8') as f:
        json.dump(results, f, indent=2, ensure_ascii=False)
    print(f"\nFull results: {full_path} ({len(results)} records)")

    # 2. Candidate list (chaincode + hits + sink, one-per-repo)
    candidate_path = OUTPUT_DIR / f"candidates_{timestamp}.json"
    with open(candidate_path, 'w', encoding='utf-8') as f:
        json.dump(unique_candidates, f, indent=2, ensure_ascii=False)
    print(f"Candidates (one-per-repo): {candidate_path} ({len(unique_candidates)} records)")

    # 3. Summary statistics
    summary = {
        "timestamp": timestamp,
        "benchmark_dir": str(BENCHMARK_DIR),
        "total_go_files": total,
        "chaincode_files": len(chaincode_files),
        "non_chaincode_files": len(non_chaincode),
        "files_with_hits": len(with_hits),
        "files_with_sink": len(with_sink),
        "chaincode_with_hits_and_sink": len(candidates),
        "unique_repos": len(repos),
        "repos_with_hits": len(repos_with_hits),
        "one_per_repo_candidates": len(unique_candidates),
        "near_duplicate_groups": len(duplicates),
        "family_distribution_all": family_counts,
        "family_distribution_chaincode": cc_family_counts,
        "file_size_stats": {
            "min_bytes": min(sizes),
            "max_bytes": max(sizes),
            "mean_bytes": sum(sizes) // len(sizes),
            "p90_bytes": p90_size,
            "min_lines": min(line_counts),
            "max_lines": max(line_counts),
            "mean_lines": sum(line_counts) // len(line_counts),
            "p90_lines": p90_lines,
        },
        "context_budget_exceeded": len(over_budget),
        "near_duplicates": {h: fps for h, fps in duplicates.items()},
    }
    summary_path = OUTPUT_DIR / f"mining_summary_{timestamp}.json"
    with open(summary_path, 'w', encoding='utf-8') as f:
        json.dump(summary, f, indent=2, ensure_ascii=False)
    print(f"Summary: {summary_path}")

    # 4. Candidate list for GPT labeling (simplified)
    labeling_list = []
    for c in sorted(unique_candidates, key=lambda x: (-len(x["possible_families"]), x["repo_id"])):
        labeling_list.append({
            "repo_id": c["repo_id"],
            "filename": c["filename"],
            "filepath": c["filepath"],
            "file_size": c["file_size"],
            "line_count": c["line_count"],
            "possible_families": c["possible_families"],
            "hit_count": c["hit_count"],
            "file_hash": c["file_hash"],
        })
    labeling_path = OUTPUT_DIR / f"labeling_candidates_{timestamp}.json"
    with open(labeling_path, 'w', encoding='utf-8') as f:
        json.dump(labeling_list, f, indent=2, ensure_ascii=False)
    print(f"Labeling candidates: {labeling_path} ({len(labeling_list)} records)")

    # 5. Safe pool candidates (chaincode, has sink, NO hits)
    safe_pool = [r for r in chaincode_files if not r["possible_families"] and r["has_sink"]]
    # One-per-repo
    safe_repo_best = {}
    for s in safe_pool:
        rid = s["repo_id"]
        if rid not in safe_repo_best:
            safe_repo_best[rid] = s
    per_repo_safe = list(safe_repo_best.values())

    # One-per-content for safe
    safe_hash_best = {}
    for s in sorted(per_repo_safe, key=lambda x: x["repo_id"]):
        h = s["file_hash"]
        if h not in safe_hash_best:
            safe_hash_best[h] = s
    safe_hash_deduped = list(safe_hash_best.values())
    safe_content_dupes_removed = len(per_repo_safe) - len(safe_hash_deduped)

    unique_safe, safe_sim_groups = deduplicate_by_similarity(safe_hash_deduped, threshold=0.85)
    safe_sim_dupes_removed = len(safe_hash_deduped) - len(unique_safe)

    safe_path = OUTPUT_DIR / f"safe_candidates_{timestamp}.json"
    with open(safe_path, 'w', encoding='utf-8') as f:
        json.dump([{
            "repo_id": s["repo_id"],
            "filename": s["filename"],
            "filepath": s["filepath"],
            "file_size": s["file_size"],
            "line_count": s["line_count"],
            "file_hash": s["file_hash"],
        } for s in unique_safe], f, indent=2, ensure_ascii=False)
    print(f"Safe candidates (one-per-repo): {safe_path} ({len(unique_safe)} records)")

    # 6. Hard negatives: chaincode files with family surface pattern but no sink
    hard_neg = [r for r in chaincode_files if r["possible_families"] and not r["has_sink"]]
    hn_repo_best = {}
    for h in hard_neg:
        rid = h["repo_id"]
        if rid not in hn_repo_best:
            hn_repo_best[rid] = h
    per_repo_hn = list(hn_repo_best.values())

    # One-per-content for hard negatives
    hn_hash_best = {}
    for h in sorted(per_repo_hn, key=lambda x: x["repo_id"]):
        hh = h["file_hash"]
        if hh not in hn_hash_best:
            hn_hash_best[hh] = h
    unique_hn = list(hn_hash_best.values())
    hn_content_dupes_removed = len(per_repo_hn) - len(unique_hn)

    hn_path = OUTPUT_DIR / f"hard_negative_candidates_{timestamp}.json"
    with open(hn_path, 'w', encoding='utf-8') as f:
        json.dump([{
            "repo_id": h["repo_id"],
            "filename": h["filename"],
            "filepath": h["filepath"],
            "file_size": h["file_size"],
            "line_count": h["line_count"],
            "possible_families": h["possible_families"],
            "file_hash": h["file_hash"],
        } for h in unique_hn], f, indent=2, ensure_ascii=False)
    print(f"Hard negatives (one-per-repo): {hn_path} ({len(unique_hn)} records)")

    print(f"\n{'='*60}")
    print(f"MINING COMPLETE")
    print(f"{'='*60}")


if __name__ == "__main__":
    main()
