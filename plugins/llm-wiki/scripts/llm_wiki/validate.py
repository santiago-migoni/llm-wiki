from __future__ import annotations

from pathlib import Path

from .frontmatter import FrontmatterError, parse_file
from .extraction import inspect_extraction
from .hashes import sha256_file
from .inventory import build_inventory
from .links import build_link_report
from .models import Finding, relative
from .contradictions import build_contradiction_report
from .paths import (
    canonical_page_slug,
    is_valid_slug,
    source_namespace_files,
    source_record_dirs,
    source_slug,
)
from .provenance import inspect_page_provenance
from .vault import REQUIRED_DIRECTORIES, schema_paths


class _Findings:
    def __init__(self):
        self.items: list[Finding] = []
        self.counter = 0

    def add(self, severity: str, message: str, path: str | None = None, **details):
        self.counter += 1
        self.items.append(
            Finding(
                id=f"VAL-{self.counter:03d}",
                severity=severity,
                message=message,
                path=path,
                details=details,
            )
        )


def _validate_document(root: Path, path: Path, findings: _Findings):
    try:
        return parse_file(path)
    except FrontmatterError as error:
        findings.add("high", str(error), relative(path, root))
        return None


def validate(root: Path) -> dict:
    findings = _Findings()
    extraction_reports: list[dict] = []
    schemas = schema_paths(root)
    if not schemas:
        findings.add("blocker", "Missing AGENTS.md or CLAUDE.md")
    elif len(schemas) > 1:
        findings.add("blocker", "Multiple schema files", schemas=[path.name for path in schemas])
    elif not schemas[0].is_file():
        findings.add("blocker", "Schema path is not a regular file", schemas[0].name)

    for required in REQUIRED_DIRECTORIES:
        path = root / required
        if not path.exists():
            findings.add("blocker", "Missing required directory", required)
        elif not path.is_dir():
            findings.add("blocker", "Required path is not a directory", required)

    inventory = build_inventory(root)
    source_root = root / "raw/sources"
    if source_root.is_dir():
        for namespace, files in source_namespace_files(root):
            findings.add(
                "high",
                "Source namespace contains unexpected files",
                relative(namespace, root),
                entries=[item.name for item in files],
            )
        for record in source_record_dirs(root):
            path = relative(record, root)
            slug = source_slug(root, record)
            if not is_valid_slug(slug):
                findings.add("medium", "Source slug is not a safe lowercase kebab-case path", path)
            sources = sorted(
                item for item in record.iterdir() if item.is_file() and item.name.startswith("source.")
            )
            if not sources:
                findings.add("high", "Source record has no current source.*", path)
            elif len(sources) > 1:
                findings.add(
                    "high",
                    "Source record has multiple current source.* files",
                    path,
                    sources=[item.name for item in sources],
                )
            extraction = record / "extracted.md"
            if not extraction.is_file():
                findings.add("high", "Source record is missing extracted.md", path)
                continue
            document = _validate_document(root, extraction, findings)
            if document and len(sources) == 1:
                extraction_report = inspect_extraction(
                    root,
                    extraction,
                    sources[0],
                    slug,
                    document.metadata,
                )
                extraction_reports.append(
                    {
                        key: value
                        for key, value in extraction_report.items()
                        if key != "issues"
                    }
                )
                for issue in extraction_report["issues"]:
                    findings.add(
                        issue["severity"],
                        issue["message"],
                        relative(extraction, root),
                        code=issue["code"],
                        **issue["details"],
                    )
                metadata = document.metadata
                expected_path = relative(sources[0], root)
                if metadata.get("source") != expected_path:
                    findings.add(
                        "high",
                        "Extraction source path does not match current source",
                        relative(extraction, root),
                        expected=expected_path,
                        actual=metadata.get("source"),
                    )
                recorded_hash = metadata.get("sha256")
                if recorded_hash and recorded_hash != sha256_file(sources[0]):
                    findings.add(
                        "high",
                        "Extraction SHA-256 is stale",
                        relative(extraction, root),
                    )

    legacy_pages: list[str] = []
    for pages_root, expected_type in (
        (root / "wiki/pages", "knowledge"),
        (root / "wiki/syntheses", "synthesis"),
    ):
        if not pages_root.is_dir():
            continue
        for page in sorted(path for path in pages_root.rglob("*.md") if path.is_file()):
            page_slug = canonical_page_slug(root, page)
            if not is_valid_slug(page_slug):
                findings.add("medium", "Page filename is not a safe lowercase kebab-case path", relative(page, root))
            document = _validate_document(root, page, findings)
            if document is None:
                continue
            metadata = document.metadata
            if metadata.get("type") != expected_type:
                findings.add(
                    "medium",
                    f"Page type must be {expected_type}",
                    relative(page, root),
                )
            if metadata.get("slug") != page_slug:
                findings.add(
                    "medium",
                    "Canonical page slug does not match filename",
                    relative(page, root),
                    expected=page_slug,
                    actual=metadata.get("slug"),
                )
            provenance = inspect_page_provenance(root, metadata, document.body)
            if provenance["format"] == "legacy-singular":
                legacy_pages.append(relative(page, root))
            for issue in provenance["issues"]:
                findings.add(
                    issue["severity"],
                    issue["message"],
                    relative(page, root),
                    code=issue["code"],
                    **issue["details"],
                )

    links = build_link_report(root)
    for item in links["missing"]:
        findings.add(
            "medium",
            "Broken wikilink",
            item["path"],
            line=item["line"],
            target=item["target"],
        )
    for item in links["ambiguous"]:
        findings.add(
            "medium",
            "Ambiguous wikilink",
            item["path"],
            line=item["line"],
            target=item["target"],
            candidates=item["candidates"],
        )
    for item in links["missing_headings"]:
        findings.add(
            "medium",
            "Missing wikilink heading",
            item["path"],
            line=item["line"],
            target=item["target"],
            heading=item["heading"],
            available_headings=item["available_headings"],
        )
    for item in links["index_missing"]:
        findings.add(
            "medium",
            "Canonical page is not linked from wiki/index.md",
            item["path"],
            index=item["index"],
        )

    contradictions = build_contradiction_report(root)
    for issue in contradictions["issues"]:
        findings.add(
            issue["severity"],
            issue["message"],
            issue["path"],
            code=issue["code"],
            line=issue["line"],
            **issue["details"],
        )

    severity_order = {"blocker": 0, "high": 1, "medium": 2, "low": 3}
    ordered = sorted(findings.items, key=lambda item: (severity_order[item.severity], item.id))
    return {
        "vault": str(root),
        "status": "invalid" if ordered else "valid",
        "findings": [item.to_dict() for item in ordered],
        "counts": {
            severity: sum(item.severity == severity for item in ordered)
            for severity in severity_order
        },
        "inventory": inventory["counts"],
        "inventory_warnings": inventory["warnings"],
        "extractions": extraction_reports,
        "contradictions": {
            key: value
            for key, value in contradictions.items()
            if key != "issues"
        },
        "links": links["counts"],
        "provenance": {
            "status": "migrable" if legacy_pages else "current",
            "legacy_pages": legacy_pages,
        },
    }
