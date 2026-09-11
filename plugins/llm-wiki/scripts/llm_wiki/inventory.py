from __future__ import annotations

from pathlib import Path

from .models import SourceRecord, relative
from .vault import REQUIRED_DIRECTORIES, git_available, run_git, schema_paths


def _markdown_files(path: Path) -> list[Path]:
    if not path.is_dir():
        return []
    return sorted(item for item in path.rglob("*.md") if item.is_file())


def _warning(root: Path, code: str, message: str, path: Path | None = None, **details) -> dict:
    item = {"code": code, "message": message}
    if path is not None:
        item["path"] = relative(path, root)
    if details:
        item["details"] = details
    return item


def build_inventory(root: Path) -> dict:
    warnings: list[dict] = []
    schemas = schema_paths(root)
    if not schemas:
        warnings.append(_warning(root, "INV-SCHEMA-MISSING", "Missing AGENTS.md or CLAUDE.md"))
    elif len(schemas) > 1:
        warnings.append(
            _warning(
                root,
                "INV-SCHEMA-MULTIPLE",
                "Multiple schema files found",
                schemas=[relative(path, root) for path in schemas],
            )
        )
    for required in REQUIRED_DIRECTORIES:
        path = root / required
        if not path.exists():
            warnings.append(_warning(root, "INV-DIRECTORY-MISSING", "Missing required directory", path))
        elif not path.is_dir():
            warnings.append(
                _warning(root, "INV-DIRECTORY-TYPE", "Required path is not a directory", path)
            )

    source_root = root / "raw/sources"
    records: list[SourceRecord] = []
    if source_root.is_dir():
        for item in sorted(source_root.iterdir()):
            if item.name != ".gitkeep" and not item.is_dir():
                warnings.append(
                    _warning(root, "INV-SOURCE-ENTRY-TYPE", "Source root entry is not a directory", item)
                )

        for record in sorted(item for item in source_root.iterdir() if item.is_dir()):
            sources = sorted(
                item for item in record.iterdir() if item.is_file() and item.name.startswith("source.")
            )
            extraction = record / "extracted.md"
            if not sources:
                warnings.append(
                    _warning(root, "INV-SOURCE-MISSING", "Source record has no current source.*", record)
                )
            elif len(sources) > 1:
                warnings.append(
                    _warning(
                        root,
                        "INV-SOURCE-MULTIPLE",
                        "Source record has multiple current source.* files",
                        record,
                        sources=[item.name for item in sources],
                    )
                )
            if not extraction.is_file():
                warnings.append(
                    _warning(root, "INV-EXTRACTION-MISSING", "Source record is missing extracted.md", record)
                )
            unexpected = sorted(
                item.name
                for item in record.iterdir()
                if item.name != "extracted.md"
                and item.name != "assets"
                and not (item.is_file() and item.name.startswith("source."))
            )
            if unexpected:
                warnings.append(
                    _warning(
                        root,
                        "INV-SOURCE-UNEXPECTED",
                        "Source record contains unexpected entries",
                        record,
                        entries=unexpected,
                    )
                )
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
    index = root / "wiki/index.md"
    if not index.is_file():
        warnings.append(_warning(root, "INV-INDEX-MISSING", "Missing wiki/index.md", index))

    git = {"available": git_available(root), "state": "unavailable", "shallow": None}
    if git["available"]:
        status = run_git(root, "status", "--short", "--untracked-files=all", "--", "raw", "wiki")
        shallow = run_git(root, "rev-parse", "--is-shallow-repository")
        git["state"] = "uncommitted" if status and status.stdout.strip() else "clean"
        git["shallow"] = bool(shallow and shallow.stdout.strip() == "true")

    warnings.sort(key=lambda item: (item.get("code", ""), item.get("path", ""), item["message"]))
    return {
        "vault": str(root),
        "schemas": [relative(path, root) for path in schemas],
        "source_records": [record.to_dict() for record in records],
        "canonical_pages": pages,
        "syntheses": syntheses,
        "pending_inbox": pending,
        "warnings": warnings,
        "git": git,
        "counts": {
            "schemas": len(schemas),
            "source_records": len(records),
            "canonical_pages": len(pages),
            "syntheses": len(syntheses),
            "pending_inbox": len(pending),
        },
    }
