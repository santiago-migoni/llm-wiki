from __future__ import annotations

import hashlib
import re
from pathlib import Path

from .models import relative
from .vault import git_available, run_git


SOURCE_PATH_RE = re.compile(r"^raw/sources/([^/]+)/source\.[^/]+$")
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def current_hashes(root: Path) -> list[dict]:
    result = []
    source_root = root / "raw/sources"
    if not source_root.is_dir():
        return result
    for path in sorted(source_root.glob("*/source.*")):
        if path.is_file():
            result.append(
                {
                    "slug": path.parent.name,
                    "path": relative(path, root),
                    "sha256": sha256_file(path),
                }
            )
    return result


def historical_hashes(root: Path) -> tuple[list[dict], list[str]]:
    warnings: list[str] = []
    if not git_available(root):
        return [], ["Git history is unavailable"]

    prefix_result = run_git(root, "rev-parse", "--show-prefix")
    prefix = prefix_result.stdout.strip() if prefix_result and prefix_result.returncode == 0 else ""
    log = run_git(
        root,
        "log",
        "--all",
        "--format=commit:%H",
        "--name-only",
        "--",
        "raw/sources",
    )
    if not log or log.returncode != 0:
        return [], ["Unable to enumerate Git history"]

    commit: str | None = None
    candidates: set[tuple[str, str, str]] = set()
    for raw in log.stdout.splitlines():
        if raw.startswith("commit:"):
            commit = raw.removeprefix("commit:")
        elif commit:
            vault_path = raw.removeprefix(prefix) if not prefix or raw.startswith(prefix) else ""
            if SOURCE_PATH_RE.fullmatch(vault_path):
                candidates.add((commit, raw, vault_path))

    results = []
    for commit, repository_path, vault_path in sorted(candidates):
        blob = run_git(root, "show", f"{commit}:{repository_path}", binary=True)
        if not blob or blob.returncode != 0:
            warnings.append(f"Unable to read {commit[:12]}:{vault_path}")
            continue
        match = SOURCE_PATH_RE.fullmatch(vault_path)
        results.append(
            {
                "slug": match.group(1) if match else None,
                "path": vault_path,
                "commit": commit,
                "sha256": hashlib.sha256(blob.stdout).hexdigest(),
            }
        )
    return results, warnings


def build_hash_report(
    root: Path,
    *,
    include_history: bool = False,
    input_path: Path | None = None,
    revert_to: str | None = None,
) -> dict:
    current = current_hashes(root)
    historical: list[dict] = []
    warnings: list[str] = []
    if include_history or revert_to:
        historical, warnings = historical_hashes(root)

    current_by_hash: dict[str, list[dict]] = {}
    historical_by_hash: dict[str, list[dict]] = {}
    for item in current:
        current_by_hash.setdefault(item["sha256"], []).append(item)
    for item in historical:
        historical_by_hash.setdefault(item["sha256"], []).append(item)

    by_hash: dict[str, list[dict]] = {}
    for item in [*current, *historical]:
        location = {key: value for key, value in item.items() if key != "sha256"}
        by_hash.setdefault(item["sha256"], []).append(location)

    duplicates = []
    for digest in sorted(by_hash):
        current_locations = [
            {key: value for key, value in item.items() if key != "sha256"}
            for item in current_by_hash.get(digest, [])
        ]
        historical_locations = [
            {key: value for key, value in item.items() if key != "sha256"}
            for item in historical_by_hash.get(digest, [])
        ]
        locations = [*current_locations, *historical_locations]
        if len(locations) < 2:
            continue
        if len(current_locations) > 1 and historical_locations:
            kind = "current-and-historical"
        elif len(current_locations) > 1:
            kind = "current"
        elif len(historical_locations) > 1:
            kind = "historical"
        else:
            kind = "current-and-historical"
        duplicates.append(
            {
                "sha256": digest,
                "kind": kind,
                "locations": locations,
                "current_locations": current_locations,
                "historical_locations": historical_locations,
            }
        )

    input_record = None
    matches = []
    input_match_type = "not-requested"
    if input_path is not None:
        resolved_input = input_path.expanduser().resolve()
        if not resolved_input.is_file():
            raise ValueError(f"input is not a readable file: {resolved_input}")
        input_record = {"path": str(resolved_input), "sha256": sha256_file(resolved_input)}
        matches = by_hash.get(input_record["sha256"], [])
        has_current = bool(current_by_hash.get(input_record["sha256"]))
        has_historical = bool(historical_by_hash.get(input_record["sha256"]))
        if has_current and has_historical:
            input_match_type = "current-and-historical-duplicate"
        elif has_current:
            input_match_type = "current-duplicate"
        elif has_historical:
            input_match_type = "historical-duplicate"
        else:
            input_match_type = "new"
        input_record["match_type"] = input_match_type

    reversion = {
        "requested": bool(revert_to),
        "sha256": revert_to,
        "status": "not-requested" if not revert_to else "not-found",
        "kind": None,
        "matches": [],
    }
    if revert_to:
        if not SHA256_RE.fullmatch(revert_to):
            raise ValueError("revert-to must be a 64-character lowercase SHA-256 digest")
        current_matches = current_by_hash.get(revert_to, [])
        historical_matches = historical_by_hash.get(revert_to, [])
        reversion["matches"] = [
            {key: value for key, value in item.items() if key != "sha256"}
            for item in [*current_matches, *historical_matches]
        ]
        if current_matches and historical_matches:
            reversion["kind"] = "current-and-historical"
        elif current_matches:
            reversion["kind"] = "current"
        elif historical_matches:
            reversion["kind"] = "historical"
        if reversion["matches"]:
            reversion["status"] = "requested"

    return {
        "vault": str(root),
        "input": input_record,
        "matches": matches,
        "input_match_type": input_match_type,
        "current": current,
        "historical": historical,
        "duplicates": duplicates,
        "reversion": reversion,
        "warnings": warnings,
    }
