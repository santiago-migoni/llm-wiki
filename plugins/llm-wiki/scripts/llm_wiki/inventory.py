from __future__ import annotations

from pathlib import Path

from .models import SourceRecord, relative
from .vault import git_available, run_git, schema_paths


def _markdown_files(path: Path) -> list[Path]:
    if not path.is_dir():
        return []
    return sorted(item for item in path.rglob("*.md") if item.is_file())


def build_inventory(root: Path) -> dict:
    schemas = schema_paths(root)
    source_root = root / "raw/sources"
    records: list[SourceRecord] = []
    if source_root.is_dir():
        for record in sorted(item for item in source_root.iterdir() if item.is_dir()):
            sources = sorted(
                item for item in record.iterdir() if item.is_file() and item.name.startswith("source.")
            )
            extraction = record / "extracted.md"
            records.append(
                SourceRecord(
                    slug=record.name,
                    path=relative(record, root),
                    sources=[relative(item, root) for item in sources],
                    extraction=relative(extraction, root) if extraction.is_file() else None,
                )
            )

    inbox = root / "raw/inbox"
    pending = []
    if inbox.is_dir():
        pending = sorted(
            relative(item, root)
            for item in inbox.rglob("*")
            if item.is_file() and item.name != ".gitkeep"
        )

    pages = [relative(path, root) for path in _markdown_files(root / "wiki/pages")]
    syntheses = [relative(path, root) for path in _markdown_files(root / "wiki/syntheses")]

    git = {"available": git_available(root), "state": "unavailable", "shallow": None}
    if git["available"]:
        status = run_git(root, "status", "--short", "--untracked-files=all", "--", "raw", "wiki")
        shallow = run_git(root, "rev-parse", "--is-shallow-repository")
        git["state"] = "uncommitted" if status and status.stdout.strip() else "clean"
        git["shallow"] = bool(shallow and shallow.stdout.strip() == "true")

    return {
        "vault": str(root),
        "schemas": [relative(path, root) for path in schemas],
        "source_records": [record.to_dict() for record in records],
        "canonical_pages": pages,
        "syntheses": syntheses,
        "pending_inbox": pending,
        "git": git,
        "counts": {
            "schemas": len(schemas),
            "source_records": len(records),
            "canonical_pages": len(pages),
            "syntheses": len(syntheses),
            "pending_inbox": len(pending),
        },
    }
