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
import subprocess
import sys
from pathlib import Path

root = Path(sys.argv[1]).expanduser().resolve()

schema_candidates = (root / "AGENTS.md", root / "CLAUDE.md")
existing_schema_paths = [path for path in schema_candidates if path.exists()]
if len(existing_schema_paths) > 1:
    print("Scale status: INVALID")
    print(
        "Multiple schema files: "
        + ", ".join(path.name for path in existing_schema_paths)
    )
    raise SystemExit(1)
if not existing_schema_paths:
    print("Scale status: INVALID")
    print("Missing required paths: AGENTS.md or CLAUDE.md")
    raise SystemExit(1)

schema_path = existing_schema_paths[0]
if not schema_path.is_file():
    print("Scale status: INVALID")
    print("Schema path is not a file: " + schema_path.name)
    raise SystemExit(1)

required_paths = (
    schema_path,
    root / "raw" / "inbox",
    root / "raw" / "sources",
    root / "wiki",
    root / "wiki" / "pages",
)
missing_paths = [str(path.relative_to(root)) for path in required_paths if not path.exists()]
if missing_paths:
    print("Scale status: INVALID")
    print("Missing required paths: " + ", ".join(missing_paths))
    raise SystemExit(1)


def files_under(path):
    if not path.is_dir():
        return []
    return [item for item in path.rglob("*") if item.is_file()]


def size_of(path):
    try:
        return path.stat().st_size
    except OSError:
        return 0


def format_bytes(value):
    units = ("B", "KiB", "MiB", "GiB")
    number = float(value)
    for unit in units:
        if number < 1024 or unit == units[-1]:
            if unit == "B":
                return f"{int(number)} {unit}"
            return f"{number:.1f} {unit}"
        number /= 1024


source_root = root / "raw" / "sources"
inbox_root = root / "raw" / "inbox"
wiki_root = root / "wiki"
pages_root = root / "wiki" / "pages"
syntheses_root = root / "wiki" / "syntheses"
source_records = sorted(item for item in source_root.iterdir() if item.is_dir())
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
missing_extraction = []
for record in source_records:
    current_files = sorted(
        path for path in record.iterdir()
        if path.is_file() and path.name.startswith("source.")
    )
    source_files.extend(current_files)
    extraction = record / "extracted.md"
    if not current_files:
        missing_current_source.append(record.name)
    if not extraction.is_file():
        missing_extraction.append(record.name)
    else:
        extraction_files.append(extraction)
    assets_root = record / "assets"
    asset_files.extend(files_under(assets_root))

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
largest_source_bytes = size_of(largest_source_file) if largest_source_file else 0
largest_wiki_bytes = size_of(largest_wiki_file) if largest_wiki_file else 0
largest_synthesis_bytes = size_of(largest_synthesis_file) if largest_synthesis_file else 0

# These thresholds mirror docs/scalability.md.
if (
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


def git_value(*args):
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


git_available = git_value("rev-parse", "--is-inside-work-tree") == "true"
git_shallow = git_value("rev-parse", "--is-shallow-repository")
git_changes = None
if git_available:
    git_changes = git_value("status", "--short", "--untracked-files=all", "--", "raw/sources", "wiki")
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
    relative_largest = largest_source_file.relative_to(root)
    print(
        "Largest source-layer file: "
        f"{format_bytes(largest_source_bytes)} ({relative_largest})"
    )
else:
    print("Largest source-layer file: 0 B")
if largest_wiki_file:
    relative_largest = largest_wiki_file.relative_to(root)
    print(
        "Largest wiki Markdown file: "
        f"{format_bytes(largest_wiki_bytes)} ({relative_largest})"
    )
else:
    print("Largest wiki Markdown file: 0 B")
if largest_synthesis_file:
    relative_largest = largest_synthesis_file.relative_to(root)
    print(
        "Largest synthesis Markdown file: "
        f"{format_bytes(largest_synthesis_bytes)} ({relative_largest})"
    )
else:
    print("Largest synthesis Markdown file: 0 B")
print(f"Sources missing current original: {len(missing_current_source)}")
print(f"Sources missing extracted.md: {len(missing_extraction)}")
if missing_current_source:
    print("Missing current originals: " + ", ".join(missing_current_source))
if missing_extraction:
    print("Missing extractions: " + ", ".join(missing_extraction))
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
else:
    recommendation = (
        "Evaluate a local rebuildable full-text index; "
        "do not change the canonical Markdown/Git source of truth."
    )
print(f"Recommendation: {recommendation}")
PY
