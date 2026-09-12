#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "FAIL: $*" >&2
  exit 2
}

if [ "$#" -ne 1 ]; then
  fail "usage: bash tests/assess-vault-scale.sh <vault-root>"
fi

vault_root="$1"
test -d "$vault_root" || fail "vault root does not exist: $vault_root"

command -v python3 >/dev/null 2>&1 || fail "python3 is required"

python3 - "$vault_root" <<'PY'
from __future__ import annotations

import os
import re
import stat
import subprocess
import sys
from pathlib import Path


root = Path(sys.argv[1]).expanduser().resolve()
SLUG_RE = re.compile(
    r"^[a-z0-9]+(?:-[a-z0-9]+)*(?:/[a-z0-9]+(?:-[a-z0-9]+)*)*$"
)
inaccessible_paths: set[Path] = set()


def relative(path: Path) -> str:
    return path.relative_to(root).as_posix()


def path_state(path: Path) -> str:
    try:
        info = path.stat()
    except FileNotFoundError:
        return "missing"
    except OSError:
        inaccessible_paths.add(path)
        return "inaccessible"
    if stat.S_ISDIR(info.st_mode):
        return "directory"
    if stat.S_ISREG(info.st_mode):
        return "file"
    return "other"


def safe_entries(path: Path) -> list[Path]:
    try:
        return sorted(path.iterdir(), key=lambda item: item.name)
    except OSError:
        inaccessible_paths.add(path)
        return []


def files_under(path: Path) -> list[Path]:
    if path_state(path) != "directory":
        return []

    found: list[Path] = []

    def onerror(error: OSError) -> None:
        inaccessible_paths.add(Path(error.filename) if error.filename else path)

    for current, directories, files in os.walk(path, onerror=onerror, followlinks=False):
        directories.sort()
        files.sort()
        for name in files:
            candidate = Path(current) / name
            if path_state(candidate) == "file":
                found.append(candidate)
    return sorted(found)


def size_of(path: Path | None) -> int:
    if path is None:
        return 0
    try:
        return path.stat().st_size
    except OSError:
        inaccessible_paths.add(path)
        return 0


def format_bytes(value: int) -> str:
    units = ("B", "KiB", "MiB", "GiB")
    number = float(value)
    for unit in units:
        if number < 1024 or unit == units[-1]:
            if unit == "B":
                return f"{int(number)} {unit}"
            return f"{number:.1f} {unit}"
        number /= 1024
    raise AssertionError("unreachable")


def git_value(*args: str):
    try:
        result = subprocess.run(
            ["git", "-C", str(root), *args],
            check=False,
            capture_output=True,
            text=True,
        )
    except OSError:
        return None
    if result.returncode != 0:
        return None
    return result.stdout.strip()


def invalid(message: str) -> None:
    print("Scale status: INVALID")
    print(message)
    raise SystemExit(1)


schema_candidates = (root / "AGENTS.md", root / "CLAUDE.md")
existing_schema_paths = [
    path for path in schema_candidates if path_state(path) != "missing"
]
if len(existing_schema_paths) > 1:
    invalid(
        "Multiple schema files: "
        + ", ".join(path.name for path in existing_schema_paths)
    )
if not existing_schema_paths:
    invalid("Missing required paths: AGENTS.md or CLAUDE.md")

schema_path = existing_schema_paths[0]
if path_state(schema_path) != "file":
    invalid(f"Invalid path type: {schema_path.name} must be a regular file")

required_directories = (
    root / "raw" / "inbox",
    root / "raw" / "sources",
    root / "wiki",
    root / "wiki" / "pages",
)
missing_paths = [
    relative(path) for path in required_directories if path_state(path) == "missing"
]
if missing_paths:
    invalid("Missing required paths: " + ", ".join(missing_paths))

invalid_types = [
    relative(path)
    for path in required_directories
    if path_state(path) not in {"directory", "inaccessible"}
]
if invalid_types:
    print("Scale status: INVALID")
    for path in invalid_types:
        print(f"Invalid path type: {path} must be a directory")
    raise SystemExit(1)

source_root = root / "raw" / "sources"
inbox_root = root / "raw" / "inbox"
wiki_root = root / "wiki"
pages_root = root / "wiki" / "pages"
syntheses_root = root / "wiki" / "syntheses"

source_directories = sorted(
    (path for path in source_root.rglob("*") if path_state(path) == "directory"),
    key=lambda path: path.as_posix(),
)
support_asset_dirs = set()
for item in source_directories:
    if item.name != "assets" or item.parent == source_root:
        continue
    entries = safe_entries(item.parent)
    namespace_children = [
        entry
        for entry in entries
        if entry.name != "assets" and path_state(entry) == "directory"
    ]
    if not namespace_children:
        support_asset_dirs.add(item)

source_records = []
namespace_files = []
for item in source_directories:
    if any(asset == item or asset in item.parents for asset in support_asset_dirs):
        continue
    entries = safe_entries(item)
    namespace_children = [
        entry
        for entry in entries
        if entry.name != "assets" and path_state(entry) == "directory"
    ]
    direct_files = [
        entry
        for entry in entries
        if entry.name != ".gitkeep" and path_state(entry) == "file"
    ]
    if namespace_children and direct_files:
        namespace_files.extend(direct_files)
    if not namespace_children:
        source_records.append(item)
unexpected_root_entries = [
    relative(item)
    for item in safe_entries(source_root)
    if item.name != ".gitkeep" and path_state(item) != "directory"
]

pending_files = [
    path for path in files_under(inbox_root) if path.name != ".gitkeep"
]
canonical_pages = [
    path for path in files_under(pages_root) if path.suffix.lower() == ".md"
]
syntheses = [
    path for path in files_under(syntheses_root) if path.suffix.lower() == ".md"
]
wiki_markdown_files = [
    path for path in files_under(wiki_root) if path.suffix.lower() == ".md"
]

source_files = []
extraction_files = []
asset_files = []
missing_current_source = []
multiple_current_source = []
missing_extraction = []
invalid_slugs = []
unexpected_source_entries = []

for record in source_records:
    slug = record.relative_to(source_root).as_posix()
    if not SLUG_RE.fullmatch(slug):
        invalid_slugs.append(slug)

    entries = safe_entries(record)
    current_files = sorted(
        item
        for item in entries
        if item.name.startswith("source.") and path_state(item) == "file"
    )
    source_files.extend(current_files)
    if not current_files:
        missing_current_source.append(slug)
    elif len(current_files) > 1:
        multiple_current_source.append(
            (slug, [item.name for item in current_files])
        )

    extraction = record / "extracted.md"
    if path_state(extraction) != "file":
        missing_extraction.append(slug)
    else:
        extraction_files.append(extraction)

    assets_root = record / "assets"
    if path_state(assets_root) == "directory":
        asset_files.extend(files_under(assets_root))

    for entry in entries:
        entry_state = path_state(entry)
        if entry.name.startswith("source.") and entry_state == "file":
            continue
        if entry.name == "extracted.md" and entry_state == "file":
            continue
        if entry.name == "assets" and entry_state == "directory":
            continue
        if entry_state != "inaccessible":
            unexpected_source_entries.append(relative(entry))

source_layer_files = [
    path for path in files_under(source_root) if path.name != ".gitkeep"
]
largest_source_file = max(source_layer_files, key=size_of, default=None)
largest_wiki_file = max(wiki_markdown_files, key=size_of, default=None)
largest_synthesis_file = max(syntheses, key=size_of, default=None)
source_bytes = sum(size_of(path) for path in source_files)
extraction_bytes = sum(size_of(path) for path in extraction_files)
asset_bytes = sum(size_of(path) for path in asset_files)
wiki_markdown_bytes = sum(size_of(path) for path in wiki_markdown_files)
synthesis_bytes = sum(size_of(path) for path in syntheses)
largest_source_bytes = size_of(largest_source_file)
largest_wiki_bytes = size_of(largest_wiki_file)
largest_synthesis_bytes = size_of(largest_synthesis_file)

structural_issues = bool(
    missing_current_source
    or multiple_current_source
    or invalid_slugs
    or unexpected_root_entries
    or namespace_files
    or unexpected_source_entries
    or missing_extraction
    or inaccessible_paths
)

# These thresholds mirror docs/scalability.md.
if structural_issues:
    scale_status = "INVALID"
elif (
    len(source_records) > 100
    or len(canonical_pages) > 300
    or wiki_markdown_bytes >= 250 * 1024 * 1024
):
    scale_status = "DERIVED-SEARCH-CANDIDATE"
elif (
    len(source_records) >= 80
    or len(canonical_pages) >= 200
    or largest_source_bytes >= 5 * 1024 * 1024
    or largest_wiki_bytes >= 5 * 1024 * 1024
    or wiki_markdown_bytes >= 50 * 1024 * 1024
):
    scale_status = "WATCH"
else:
    scale_status = "GREEN"

git_available = git_value("rev-parse", "--is-inside-work-tree") == "true"
git_shallow = git_value("rev-parse", "--is-shallow-repository")
git_changes = None
if git_available:
    git_changes = git_value(
        "status", "--short", "--untracked-files=all", "--", "raw/sources", "wiki"
    )
git_state = (
    "unavailable"
    if not git_available
    else "uncommitted"
    if git_changes
    else "clean"
)

print(f"Vault: {root}")
print(f"Schema: {schema_path.name}")
print(f"Scale status: {scale_status}")
print(f"Current source records: {len(source_records)}")
print(f"Canonical pages: {len(canonical_pages)}")
print(f"Syntheses: {len(syntheses)}")
print(f"Pending inbox files: {len(pending_files)}")
print(f"Current source bytes: {format_bytes(source_bytes)}")
print(f"Extraction bytes: {format_bytes(extraction_bytes)}")
print(f"Asset bytes: {format_bytes(asset_bytes)}")
print(f"Wiki Markdown bytes: {format_bytes(wiki_markdown_bytes)}")
print(f"Synthesis Markdown bytes: {format_bytes(synthesis_bytes)}")
if largest_source_file:
    print(
        "Largest source-layer file: "
        f"{format_bytes(largest_source_bytes)} ({relative(largest_source_file)})"
    )
else:
    print("Largest source-layer file: 0 B")
if largest_wiki_file:
    print(
        "Largest wiki Markdown file: "
        f"{format_bytes(largest_wiki_bytes)} ({relative(largest_wiki_file)})"
    )
else:
    print("Largest wiki Markdown file: 0 B")
if largest_synthesis_file:
    print(
        "Largest synthesis Markdown file: "
        f"{format_bytes(largest_synthesis_bytes)} ({relative(largest_synthesis_file)})"
    )
else:
    print("Largest synthesis Markdown file: 0 B")
print(f"Sources missing current original: {len(missing_current_source)}")
print(f"Sources with multiple current originals: {len(multiple_current_source)}")
print(f"Sources missing extracted.md: {len(missing_extraction)}")
if missing_current_source:
    print("Missing current originals: " + ", ".join(sorted(missing_current_source)))
if multiple_current_source:
    for slug, names in multiple_current_source:
        print(f"Multiple current originals: {slug} ({', '.join(names)})")
if missing_extraction:
    print("Missing extractions: " + ", ".join(sorted(missing_extraction)))
if invalid_slugs:
    print("Invalid source slugs: " + ", ".join(sorted(invalid_slugs)))
if unexpected_root_entries:
    print("Unexpected source-root entries: " + ", ".join(sorted(unexpected_root_entries)))
if namespace_files:
    print(
        "Unexpected namespace source entries: "
        + ", ".join(sorted(relative(path) for path in namespace_files))
    )
if unexpected_source_entries:
    print("Unexpected source entries: " + ", ".join(sorted(unexpected_source_entries)))
for path in sorted(inaccessible_paths, key=relative):
    print("Inaccessible directory: " + relative(path))
if git_available:
    history = "shallow" if git_shallow == "true" else "full-or-unknown"
    print(f"Git: {git_state} ({history})")
else:
    print("Git: unavailable")

if scale_status == "GREEN":
    recommendation = (
        "Keep targeted/index-first retrieval with Markdown and Git; "
        "measure real query behavior before adding a derived layer."
    )
elif scale_status == "WATCH":
    recommendation = (
        "Measure retrieval latency and coverage on real workloads; "
        "review a derived full-text layer only if the experience degrades."
    )
elif scale_status == "DERIVED-SEARCH-CANDIDATE":
    recommendation = (
        "Evaluate a local rebuildable full-text index; "
        "do not change the canonical Markdown/Git source of truth."
    )
else:
    recommendation = "Repair structural issues before using the scale recommendation."
print(f"Recommendation: {recommendation}")
if scale_status == "INVALID":
    raise SystemExit(1)
PY
