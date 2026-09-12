"""Shared vault path and slug rules.

Source records may be addressed by a flat slug (``policy``) or by a
hierarchical slug (``fundamentos/actividades``).  Keeping the rules here
prevents inventory, validation, hashes, and provenance from interpreting the
same path differently.
"""

from __future__ import annotations

import re
from pathlib import Path, PurePosixPath


SLUG_SEGMENT_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
SLUG_RE = re.compile(
    r"^[a-z0-9]+(?:-[a-z0-9]+)*(?:/[a-z0-9]+(?:-[a-z0-9]+)*)*$"
)
SOURCE_NAME_RE = re.compile(r"^source\.[^/]+$")


def is_valid_slug(value: object) -> bool:
    """Return whether *value* is a safe flat or hierarchical slug."""

    return isinstance(value, str) and bool(SLUG_RE.fullmatch(value))


def _source_directories(source_root: Path) -> list[Path]:
    return sorted(item for item in source_root.rglob("*") if item.is_dir())


def _support_asset_dirs(source_root: Path, directories: list[Path]) -> set[Path]:
    """Return ``assets`` directories that belong to source-record leaves."""

    support_assets: set[Path] = set()
    for path in directories:
        if path.name != "assets" or path.parent == source_root:
            continue
        try:
            siblings = list(path.parent.iterdir())
        except OSError:
            continue
        namespace_children = [
            item
            for item in siblings
            if item.is_dir() and item.name != "assets"
        ]
        if not namespace_children:
            support_assets.add(path)
    return support_assets


def _under_support_assets(path: Path, support_assets: set[Path]) -> bool:
    return any(asset == path or asset in path.parents for asset in support_assets)


def source_record_dirs(root: Path) -> list[Path]:
    """Enumerate source-record leaf directories below ``raw/sources``.

    A directory containing a child other than ``assets`` is considered a
    namespace.  A directory with no namespace children is a record candidate,
    including malformed/empty leaves so callers can report structural errors.
    """

    source_root = root / "raw/sources"
    if not source_root.is_dir():
        return []

    directories = _source_directories(source_root)
    support_assets = _support_asset_dirs(source_root, directories)
    records: list[Path] = []
    for path in directories:
        if _under_support_assets(path, support_assets):
            continue
        try:
            children = list(path.iterdir())
        except OSError:
            # The caller cannot inspect this directory reliably, but retaining
            # it as a candidate lets validation surface the path.
            records.append(path)
            continue
        namespace_children = [
            item for item in children if item.is_dir() and item.name != "assets"
        ]
        if not namespace_children:
            records.append(path)
    return records


def source_slug(root: Path, record: Path) -> str:
    """Return the complete slash-separated slug for a source record path."""

    source_root = root / "raw/sources"
    return record.relative_to(source_root).as_posix()


def source_namespace_files(root: Path) -> list[tuple[Path, list[Path]]]:
    """Find files incorrectly placed directly in namespace directories."""

    source_root = root / "raw/sources"
    if not source_root.is_dir():
        return []

    findings: list[tuple[Path, list[Path]]] = []
    directories = _source_directories(source_root)
    support_assets = _support_asset_dirs(source_root, directories)
    for path in directories:
        if _under_support_assets(path, support_assets):
            continue
        try:
            children = list(path.iterdir())
        except OSError:
            continue
        has_namespace_child = any(
            item.is_dir() and item.name != "assets" for item in children
        )
        if not has_namespace_child:
            continue
        files = [
            item for item in children if item.is_file() and item.name != ".gitkeep"
        ]
        if files:
            findings.append((path, sorted(files)))
    return findings


def source_slug_from_vault_path(value: object) -> str | None:
    """Extract a validated source slug from ``raw/sources/.../source.*``."""

    if not isinstance(value, str) or not value or "\\" in value:
        return None
    path = PurePosixPath(value)
    if path.is_absolute() or path.as_posix() != value:
        return None
    parts = path.parts
    if len(parts) < 4 or parts[:2] != ("raw", "sources"):
        return None
    if not SOURCE_NAME_RE.fullmatch(path.name):
        return None
    slug = "/".join(parts[2:-1])
    return slug if is_valid_slug(slug) else None


def canonical_page_slug(root: Path, page: Path) -> str:
    """Return a canonical page's full relative slug, without ``.md``."""

    pages_root = root / "wiki/pages"
    return page.relative_to(pages_root).with_suffix("").as_posix()
