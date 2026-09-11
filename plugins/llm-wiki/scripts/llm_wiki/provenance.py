"""Structured provenance for canonical pages and durable syntheses.

The page contract accepts the historical singular ``source``/``extracted``/
``sha256`` fields, but new provenance is represented as an ordered list of
source records. This module keeps validation and the safe migration in one
place so the CLI and wiki-lint cannot drift apart.
"""

from __future__ import annotations

import difflib
import os
import re
import tempfile
from pathlib import Path, PurePosixPath
from typing import Any

from .frontmatter import FrontmatterError, parse_file
from .hashes import sha256_file


SLUG_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")
SOURCE_FIELDS = ("slug", "source", "extracted", "sha256", "role")
LEGACY_FIELDS = ("source", "extracted", "sha256")
ALLOWED_ROLES = ("primary", "supporting", "context", "counterpoint")
ROLE_ORDER = {role: position for position, role in enumerate(ALLOWED_ROLES)}
CITATION_LINE_RE = re.compile(
    r"^\s*-\s+(Evidence|Support|Source|Sources):\s*(.*?)\s*$",
    re.IGNORECASE,
)
CODE_SPAN_RE = re.compile(r"`([^`]+)`")
SOURCE_NAME_RE = re.compile(r"^source\.[^/]+$")


def _issue(
    code: str,
    severity: str,
    message: str,
    **details: Any,
) -> dict[str, Any]:
    return {
        "code": code,
        "severity": severity,
        "message": message,
        "details": details,
    }


def _path_value(value: Any) -> PurePosixPath | None:
    if not isinstance(value, str) or not value or "\\" in value:
        return None
    path = PurePosixPath(value)
    if path.is_absolute() or ".." in path.parts or "." in path.parts:
        return None
    if path.as_posix() != value:
        return None
    return path


def _source_record_prefix(slug: str) -> str:
    return f"raw/sources/{slug}/"


def _entry_issues(
    root: Path,
    entry: Any,
    index: int,
    *,
    require_hash: bool = True,
) -> tuple[list[dict[str, Any]], dict[str, Any] | None]:
    issues: list[dict[str, Any]] = []
    if not isinstance(entry, dict):
        return [
            _issue(
                "PROV-ENTRY-TYPE",
                "high",
                "Provenance entry must be a mapping",
                index=index,
            )
        ], None

    missing = [field for field in SOURCE_FIELDS if field not in entry]
    if missing:
        issues.append(
            _issue(
                "PROV-REQUIRED-FIELD",
                "high",
                "Provenance entry is missing required fields",
                index=index,
                fields=missing,
            )
        )

    slug = entry.get("slug")
    if not isinstance(slug, str) or not SLUG_RE.fullmatch(slug):
        issues.append(
            _issue(
                "PROV-SLUG",
                "high",
                "Provenance source slug is not lowercase kebab-case",
                index=index,
                slug=slug,
            )
        )

    source = entry.get("source")
    source_path = _path_value(source)
    if source_path is None:
        issues.append(
            _issue(
                "PROV-SOURCE-PATH",
                "high",
                "Provenance source path must be a normalized vault-relative path",
                index=index,
                source=source,
            )
        )
    elif isinstance(slug, str) and SLUG_RE.fullmatch(slug):
        prefix = _source_record_prefix(slug)
        if not source.startswith(prefix) or not SOURCE_NAME_RE.fullmatch(source_path.name):
            issues.append(
                _issue(
                    "PROV-SOURCE-RELATION",
                    "high",
                    "Provenance source path does not belong to its source slug",
                    index=index,
                    slug=slug,
                    source=source,
                )
            )

    extracted = entry.get("extracted")
    extracted_path = _path_value(extracted)
    if extracted_path is None:
        issues.append(
            _issue(
                "PROV-EXTRACTED-PATH",
                "high",
                "Provenance extraction path must be a normalized vault-relative path",
                index=index,
                extracted=extracted,
            )
        )
    elif isinstance(slug, str) and SLUG_RE.fullmatch(slug):
        expected = f"{_source_record_prefix(slug)}extracted.md"
        if extracted != expected:
            issues.append(
                _issue(
                    "PROV-EXTRACTED-RELATION",
                    "high",
                    "Provenance extraction path does not belong to its source slug",
                    index=index,
                    slug=slug,
                    extracted=extracted,
                    expected=expected,
                )
            )

    digest = entry.get("sha256")
    if require_hash and (not isinstance(digest, str) or not SHA256_RE.fullmatch(digest)):
        issues.append(
            _issue(
                "PROV-HASH-FORMAT",
                "high",
                "Provenance sha256 must be a lowercase 64-character hexadecimal hash",
                index=index,
                sha256=digest,
            )
        )
    elif digest is not None and (
        not isinstance(digest, str) or not SHA256_RE.fullmatch(digest)
    ):
        issues.append(
            _issue(
                "PROV-HASH-FORMAT",
                "high",
                "Provenance sha256 must be a lowercase 64-character hexadecimal hash",
                index=index,
                sha256=digest,
            )
        )

    role = entry.get("role")
    if role not in ALLOWED_ROLES:
        issues.append(
            _issue(
                "PROV-ROLE",
                "high",
                "Provenance role is not allowed",
                index=index,
                role=role,
                allowed=list(ALLOWED_ROLES),
            )
        )

    if not isinstance(slug, str) or not SLUG_RE.fullmatch(slug):
        return issues, None

    normalized = {
        "slug": slug,
        "source": source,
        "extracted": extracted,
        "sha256": digest,
        "role": role,
    }
    if source_path is not None and SOURCE_NAME_RE.fullmatch(source_path.name):
        source_file = root / source
        if not source_file.is_file():
            issues.append(
                _issue(
                    "PROV-SOURCE-MISSING",
                    "high",
                    "Provenance source file does not exist",
                    index=index,
                    source=source,
                )
            )
        elif isinstance(digest, str) and SHA256_RE.fullmatch(digest):
            actual = sha256_file(source_file)
            if digest != actual:
                issues.append(
                    _issue(
                        "PROV-HASH-STALE",
                        "high",
                        "Provenance SHA-256 does not match the current source",
                        index=index,
                        source=source,
                        expected=actual,
                        actual=digest,
                    )
                )

    if extracted_path is not None and not (root / extracted).is_file():
        issues.append(
            _issue(
                "PROV-EXTRACTED-MISSING",
                "high",
                "Provenance extraction file does not exist",
                index=index,
                extracted=extracted,
            )
        )

    return issues, normalized


def validate_source_entries(
    root: Path,
    entries: Any,
    *,
    require_hash: bool = True,
) -> list[dict[str, Any]]:
    """Validate every structured source entry and its current file bytes."""

    if not isinstance(entries, list):
        return [
            _issue(
                "PROV-SOURCES-TYPE",
                "high",
                "sources metadata must be a list of mappings",
            )
        ]

    issues: list[dict[str, Any]] = []
    seen_slugs: set[str] = set()
    seen_paths: set[str] = set()
    previous_role: int | None = None
    previous_slug: str | None = None

    for index, entry in enumerate(entries):
        entry_issues, normalized = _entry_issues(
            root, entry, index, require_hash=require_hash
        )
        issues.extend(entry_issues)
        if normalized is None:
            continue

        slug = normalized["slug"]
        source = normalized["source"]
        if slug in seen_slugs:
            issues.append(
                _issue(
                    "PROV-DUPLICATE-SLUG",
                    "high",
                    "Provenance sources contain a duplicate slug",
                    index=index,
                    slug=slug,
                )
            )
        seen_slugs.add(slug)
        if source in seen_paths:
            issues.append(
                _issue(
                    "PROV-DUPLICATE-PATH",
                    "high",
                    "Provenance sources contain a duplicate source path",
                    index=index,
                    source=source,
                )
            )
        seen_paths.add(source)

        role = normalized["role"]
        role_position = ROLE_ORDER.get(role)
        if role_position is not None and previous_role is not None:
            if role_position < previous_role or (
                role_position == previous_role
                and previous_slug is not None
                and slug < previous_slug
            ):
                issues.append(
                    _issue(
                        "PROV-ORDER",
                        "medium",
                        "Provenance sources are not in stable role/slug order",
                        index=index,
                        previous_slug=previous_slug,
                        slug=slug,
                    )
                )
        if role_position is not None:
            previous_role = role_position
            previous_slug = slug

    return issues


def _legacy_slug(source: Any) -> str | None:
    if not isinstance(source, str):
        return None
    path = _path_value(source)
    if path is None or len(path.parts) < 4:
        return None
    slug = path.parts[2]
    return slug if SLUG_RE.fullmatch(slug) else None


def legacy_entry(metadata: dict[str, Any]) -> dict[str, Any] | None:
    """Return the compatibility entry represented by singular metadata."""

    source = metadata.get("source")
    slug = _legacy_slug(source)
    if slug is None:
        return None
    return {
        "slug": slug,
        "source": source,
        "extracted": metadata.get("extracted"),
        "sha256": metadata.get("sha256"),
        "role": "primary",
    }


def extract_citations(body: str) -> list[dict[str, Any]]:
    """Extract the explicit Evidence/Support citation lines from a page."""

    citations = []
    for line_number, line in enumerate(body.splitlines(), start=1):
        match = CITATION_LINE_RE.match(line)
        if not match:
            continue
        slugs = [value for value in CODE_SPAN_RE.findall(match.group(2)) if SLUG_RE.fullmatch(value)]
        citations.append(
            {
                "line": line_number,
                "label": match.group(1).lower(),
                "slugs": slugs,
                "text": line.strip(),
            }
        )
    return citations


def inspect_page_provenance(
    root: Path,
    metadata: dict[str, Any],
    body: str,
) -> dict[str, Any]:
    """Inspect page provenance without treating the legacy shape as invalid."""

    has_sources = "sources" in metadata
    has_legacy = any(key in metadata for key in LEGACY_FIELDS)
    issues: list[dict[str, Any]] = []
    migration_notes: list[str] = []
    citations = extract_citations(body)

    if has_sources and has_legacy:
        issues.append(
            _issue(
                "PROV-MIXED-FORMAT",
                "high",
                "Page mixes structured sources with legacy singular provenance",
            )
        )

    if has_sources:
        entries = metadata.get("sources")
        issues.extend(validate_source_entries(root, entries))
        declared = {
            entry.get("slug")
            for entry in entries
            if isinstance(entry, dict) and isinstance(entry.get("slug"), str)
        } if isinstance(entries, list) else set()
        for citation in citations:
            for slug in citation["slugs"]:
                if slug not in declared:
                    issues.append(
                        _issue(
                            "PROV-CITATION-UNDECLARED",
                            "medium",
                            "Citation references a source slug absent from page sources",
                            line=citation["line"],
                            slug=slug,
                        )
                    )
        if isinstance(entries, list) and len(entries) > 1 and body.strip() and not citations:
            issues.append(
                _issue(
                    "PROV-CITATION-MISSING",
                    "medium",
                    "Multi-source page has no explicit Evidence or Support citations",
                )
            )
        return {
            "format": "multi-source",
            "entries": entries if isinstance(entries, list) else [],
            "citations": citations,
            "issues": issues,
            "migration_notes": migration_notes,
        }

    if has_legacy:
        entry = legacy_entry(metadata)
        if entry is None:
            issues.append(
                _issue(
                    "PROV-LEGACY-SOURCE",
                    "high",
                    "Legacy source metadata is missing a valid raw/sources path",
                )
            )
        else:
            legacy_issues = _entry_issues(
                root,
                entry,
                0,
                require_hash=False,
            )[0]
            # A missing legacy hash is migrable and can be computed safely. A
            # missing path or a stale supplied hash is still a real finding.
            issues.extend(
                item
                for item in legacy_issues
                if item["code"] not in {"PROV-REQUIRED-FIELD", "PROV-HASH-FORMAT"}
            )
            if metadata.get("extracted") is None:
                migration_notes.append("extracted path will be inferred from the source slug")
            if metadata.get("sha256") is None:
                migration_notes.append("sha256 will be computed from the current source")
        return {
            "format": "legacy-singular",
            "entries": [entry] if entry else [],
            "citations": citations,
            "issues": issues,
            "migration_notes": migration_notes,
        }

    return {
        "format": "none",
        "entries": [],
        "citations": citations,
        "issues": issues,
        "migration_notes": migration_notes,
    }


def _document_parts(text: str) -> tuple[list[str], int]:
    lines = text.splitlines()
    if not lines or lines[0].strip() != "---":
        raise FrontmatterError("missing opening frontmatter delimiter")
    closing = next(
        (index for index, line in enumerate(lines[1:], start=1) if line.strip() == "---"),
        None,
    )
    if closing is None:
        raise FrontmatterError("missing closing frontmatter delimiter")
    return lines, closing


def _render_sources_entry(entry: dict[str, Any]) -> list[str]:
    return [
        "sources:",
        f"  - slug: {entry['slug']}",
        f"    source: {entry['source']}",
        f"    extracted: {entry['extracted']}",
        f"    sha256: {entry['sha256']}",
        f"    role: {entry['role']}",
    ]


def migrate_document_text(text: str, entry: dict[str, Any]) -> str:
    """Replace only the legacy top-level fields and preserve the page body."""

    lines, closing = _document_parts(text)
    key_lines = {
        index
        for index, line in enumerate(lines[1:closing], start=1)
        if re.match(r"^(source|extracted|sha256):(?:\s|$)", line)
    }
    if not key_lines:
        return text

    first = min(key_lines)
    rendered = []
    for index, line in enumerate(lines):
        if index == first:
            rendered.extend(_render_sources_entry(entry))
        if index in key_lines:
            continue
        rendered.append(line)
    newline = "\r\n" if "\r\n" in text else "\n"
    result = newline.join(rendered)
    if text.endswith(("\n", "\r")):
        result += newline
    return result


def _migration_documents(root: Path) -> list[Path]:
    paths: list[Path] = []
    for directory in (root / "wiki/pages", root / "wiki/syntheses"):
        if directory.is_dir():
            paths.extend(path for path in directory.rglob("*.md") if path.is_file())
    return sorted(paths)


def _migration_entry(root: Path, metadata: dict[str, Any]) -> tuple[dict[str, Any] | None, list[dict[str, Any]], list[str]]:
    errors: list[dict[str, Any]] = []
    warnings: list[str] = []
    source = metadata.get("source")
    slug = _legacy_slug(source)
    source_path = _path_value(source)
    if slug is None or source_path is None:
        errors.append(
            _issue(
                "MIGRATE-SOURCE-PATH",
                "high",
                "Cannot migrate legacy provenance without a valid source path",
                source=source,
            )
        )
        return None, errors, warnings

    source_file = root / source
    if not source_file.is_file():
        errors.append(
            _issue(
                "MIGRATE-SOURCE-MISSING",
                "high",
                "Cannot migrate because the current source file is missing",
                source=source,
            )
        )
        return None, errors, warnings

    extracted = metadata.get("extracted")
    if extracted is None:
        extracted = f"raw/sources/{slug}/extracted.md"
        warnings.append("extracted path inferred")
    digest = metadata.get("sha256")
    actual_digest = sha256_file(source_file)
    if digest is None:
        digest = actual_digest
        warnings.append("sha256 computed from current source")
    elif digest != actual_digest:
        errors.append(
            _issue(
                "MIGRATE-HASH-STALE",
                "high",
                "Cannot migrate a legacy page with a stale sha256",
                source=source,
                expected=actual_digest,
                actual=digest,
            )
        )

    entry = {
        "slug": slug,
        "source": source,
        "extracted": extracted,
        "sha256": digest,
        "role": "primary",
    }
    errors.extend(validate_source_entries(root, [entry]))
    return entry, errors, warnings


def plan_provenance_migration(root: Path) -> dict[str, Any]:
    """Build an all-or-nothing migration plan without changing the vault."""

    changes: list[dict[str, Any]] = []
    errors: list[dict[str, Any]] = []
    warnings: list[dict[str, Any]] = []
    for path in _migration_documents(root):
        relative_path = path.relative_to(root).as_posix()
        try:
            document = parse_file(path)
        except FrontmatterError as error:
            errors.append(
                _issue(
                    "MIGRATE-FRONTMATTER",
                    "high",
                    str(error),
                    path=relative_path,
                )
            )
            continue

        metadata = document.metadata
        has_legacy = any(key in metadata for key in LEGACY_FIELDS)
        if not has_legacy:
            continue
        if "sources" in metadata:
            errors.append(
                _issue(
                    "MIGRATE-MIXED-FORMAT",
                    "high",
                    "Cannot migrate a page that already mixes sources with legacy fields",
                    path=relative_path,
                )
            )
            continue

        entry, entry_errors, entry_warnings = _migration_entry(root, metadata)
        if entry_errors:
            for item in entry_errors:
                item["details"]["path"] = relative_path
            errors.extend(entry_errors)
            continue
        if entry is None:
            continue
        migrated = migrate_document_text(path.read_text(encoding="utf-8"), entry)
        original = path.read_text(encoding="utf-8")
        if migrated == original:
            continue
        changes.append(
            {
                "path": relative_path,
                "source": entry,
                "warnings": entry_warnings,
                "diff": "".join(
                    difflib.unified_diff(
                        original.splitlines(keepends=True),
                        migrated.splitlines(keepends=True),
                        fromfile=relative_path,
                        tofile=relative_path,
                    )
                ),
                "before": original,
                "after": migrated,
            }
        )
        if entry_warnings:
            warnings.append({"path": relative_path, "items": entry_warnings})

    status = "blocked" if errors else "needs-migration" if changes else "clean"
    return {
        "vault": str(root),
        "mode": "check",
        "status": status,
        "affected_pages": [change["path"] for change in changes],
        "changes": changes,
        "errors": errors,
        "warnings": warnings,
    }


def write_provenance_migration(root: Path, plan: dict[str, Any]) -> dict[str, Any]:
    """Apply a previously built plan, refusing if a page changed meanwhile."""

    if plan["status"] == "blocked" or not plan["changes"]:
        return plan

    for change in plan["changes"]:
        path = root / change["path"]
        current = path.read_text(encoding="utf-8")
        if current != change["before"]:
            raise RuntimeError(
                f"migration plan is stale; page changed before write: {change['path']}"
            )

    temporary_paths: list[tuple[Path, Path]] = []
    try:
        for change in plan["changes"]:
            path = root / change["path"]
            with tempfile.NamedTemporaryFile(
                "w",
                encoding="utf-8",
                dir=path.parent,
                prefix=f".{path.name}.",
                suffix=".provenance.tmp",
                delete=False,
            ) as handle:
                handle.write(change["after"])
                temporary_paths.append((Path(handle.name), path))
        for temporary, destination in temporary_paths:
            os.replace(temporary, destination)
    finally:
        for temporary, _ in temporary_paths:
            if temporary.exists():
                temporary.unlink()

    result = dict(plan)
    result["mode"] = "write"
    result["status"] = "written"
    result["written_pages"] = list(plan["affected_pages"])
    return result


def _public_migration_result(result: dict[str, Any]) -> dict[str, Any]:
    """Keep page snapshots internal so JSON output does not copy page bodies."""

    public = dict(result)
    public["changes"] = [
        {key: value for key, value in change.items() if key not in {"before", "after"}}
        for change in result["changes"]
    ]
    return public


def migrate_provenance(root: Path, *, write: bool = False) -> dict[str, Any]:
    plan = plan_provenance_migration(root)
    if write and plan["status"] != "blocked":
        return _public_migration_result(write_provenance_migration(root, plan))
    return _public_migration_result(plan)
