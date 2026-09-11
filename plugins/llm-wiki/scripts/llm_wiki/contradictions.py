"""Deterministic checks for query-visible contradiction records."""

from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from .models import relative


CONTRADICTION_STATUSES = {"unresolved", "resolved", "superseded"}
FIELD_RE = re.compile(r"^\s*-\s*([^:]+):\s*(.*?)\s*$")
PAGE_SECTION_RE = re.compile(r"^##\s+Contradictions\s*$", re.IGNORECASE)
SUBHEADING_RE = re.compile(r"^###\s+(.+?)\s*$")
LOG_HEADING_RE = re.compile(
    r"^##\s+(?:\[[^\]]+\]\s+)?contradiction\s*\|\s*(.+?)\s*$",
    re.IGNORECASE,
)
PAGE_LINK_RE = re.compile(r"\[\[([a-z0-9]+(?:-[a-z0-9]+)*)")


@dataclass
class ContradictionEntry:
    entry_id: str
    path: Path
    line: int
    fields: dict[str, str]
    page: str | None
    kind: str


def _normalize_key(value: str) -> str:
    return " ".join(value.strip().lower().split())


def _normalize_id(value: str) -> str:
    value = value.strip().strip("`")
    return re.sub(r"[^a-zA-Z0-9]+", "-", value).strip("-").lower()


def _fields(lines: list[str], start: int, end: int) -> dict[str, str]:
    result: dict[str, str] = {}
    for line in lines[start:end]:
        match = FIELD_RE.match(line)
        if match:
            result[_normalize_key(match.group(1))] = match.group(2).strip()
    return result


def _read_lines(path: Path) -> list[str]:
    try:
        return path.read_text(encoding="utf-8").splitlines()
    except (OSError, UnicodeDecodeError):
        return []


def _page_from_field(fields: dict[str, str]) -> str | None:
    value = fields.get("page") or fields.get("canonical page")
    if not value:
        return None
    linked = PAGE_LINK_RE.search(value)
    if linked:
        return linked.group(1)
    return None


def _parse_page_entries(root: Path) -> list[ContradictionEntry]:
    entries: list[ContradictionEntry] = []
    pages_root = root / "wiki/pages"
    if not pages_root.is_dir():
        return entries

    for path in sorted(item for item in pages_root.rglob("*.md") if item.is_file()):
        lines = _read_lines(path)
        in_section = False
        entry_start: int | None = None
        entry_id: str | None = None

        def flush(end: int) -> None:
            if entry_start is None or entry_id is None:
                return
            entries.append(
                ContradictionEntry(
                    entry_id=entry_id,
                    path=path,
                    line=entry_start + 1,
                    fields=_fields(lines, entry_start, end),
                    page=path.stem,
                    kind="page",
                )
            )

        for index, line in enumerate(lines):
            if PAGE_SECTION_RE.match(line):
                in_section = True
                continue
            if not in_section:
                continue
            if re.match(r"^##\s+", line):
                flush(index)
                break
            heading = SUBHEADING_RE.match(line)
            if heading:
                flush(index)
                entry_id = _normalize_id(heading.group(1))
                entry_start = index
        else:
            if in_section:
                flush(len(lines))

    return entries


def _parse_log_entries(root: Path) -> list[ContradictionEntry]:
    path = root / "wiki/log.md"
    if not path.is_file():
        return []
    lines = _read_lines(path)
    entries: list[ContradictionEntry] = []
    starts: list[tuple[int, str]] = []
    for index, line in enumerate(lines):
        heading = LOG_HEADING_RE.match(line)
        if heading:
            starts.append((index, _normalize_id(heading.group(1))))

    for position, (start, entry_id) in enumerate(starts):
        end = starts[position + 1][0] if position + 1 < len(starts) else len(lines)
        next_heading = next(
            (
                index
                for index in range(start + 1, end)
                if re.match(r"^##\s+", lines[index])
            ),
            end,
        )
        fields = _fields(lines, start + 1, next_heading)
        entries.append(
            ContradictionEntry(
                entry_id=entry_id,
                path=path,
                line=start + 1,
                fields=fields,
                page=_page_from_field(fields),
                kind="log",
            )
        )
    return entries


def _issue(
    root: Path,
    code: str,
    message: str,
    entry: ContradictionEntry,
    **details: Any,
) -> dict[str, Any]:
    return {
        "code": code,
        "severity": "high",
        "message": message,
        "path": relative(entry.path, root),
        "line": entry.line,
        "details": details,
    }


def _validate_page_entry(root: Path, entry: ContradictionEntry) -> list[dict[str, Any]]:
    issues: list[dict[str, Any]] = []
    required = ("claim a", "evidence a", "claim b", "evidence b", "status")
    missing = [field for field in required if not entry.fields.get(field)]
    if missing:
        issues.append(
            _issue(
                root,
                "CONTRA-FIELD-MISSING",
                "Contradiction entry is missing required fields",
                entry,
                fields=missing,
            )
        )

    status = entry.fields.get("status", "").strip("`").lower()
    if status not in CONTRADICTION_STATUSES:
        issues.append(
            _issue(
                root,
                "CONTRA-STATUS",
                "Contradiction status is not supported",
                entry,
                allowed=sorted(CONTRADICTION_STATUSES),
                actual=status or None,
            )
        )
    if status in {"resolved", "superseded"}:
        missing_resolution = [
            field for field in ("resolution", "criterion") if not entry.fields.get(field)
        ]
        if missing_resolution:
            issues.append(
                _issue(
                    root,
                    "CONTRA-RESOLUTION-MISSING",
                    "Resolved or superseded contradiction must preserve both evidences and explain the resolution criterion",
                    entry,
                    fields=missing_resolution,
                )
            )
    return issues


def _validate_log_entry(root: Path, entry: ContradictionEntry) -> list[dict[str, Any]]:
    issues: list[dict[str, Any]] = []
    if entry.page is None:
        issues.append(
            _issue(
                root,
                "CONTRA-LOG-PAGE-MISSING",
                "Contradiction log entry must identify its canonical page with a wikilink",
                entry,
            )
        )
    status = entry.fields.get("status", "").strip("`").lower()
    if status not in CONTRADICTION_STATUSES:
        issues.append(
            _issue(
                root,
                "CONTRA-LOG-STATUS",
                "Contradiction log entry must declare a supported status",
                entry,
                allowed=sorted(CONTRADICTION_STATUSES),
                actual=status or None,
            )
        )
    return issues


def build_contradiction_report(root: Path) -> dict[str, Any]:
    page_entries = _parse_page_entries(root)
    log_entries = _parse_log_entries(root)
    issues: list[dict[str, Any]] = []

    for entry in page_entries:
        issues.extend(_validate_page_entry(root, entry))
    for entry in log_entries:
        issues.extend(_validate_log_entry(root, entry))

    page_by_key: dict[tuple[str, str], ContradictionEntry] = {}
    log_by_key: dict[tuple[str, str], ContradictionEntry] = {}
    for entry in page_entries:
        key = (entry.page or "", entry.entry_id)
        if key in page_by_key:
            issues.append(
                _issue(
                    root,
                    "CONTRA-DUPLICATE",
                    "Canonical page contains duplicate contradiction identifiers",
                    entry,
                    id=entry.entry_id,
                )
            )
        page_by_key[key] = entry
    for entry in log_entries:
        key = (entry.page or "", entry.entry_id)
        if key in log_by_key:
            issues.append(
                _issue(
                    root,
                    "CONTRA-LOG-DUPLICATE",
                    "Contradiction log contains duplicate page and identifier entries",
                    entry,
                    id=entry.entry_id,
                )
            )
        log_by_key[key] = entry

    for entry in page_entries:
        key = (entry.page or "", entry.entry_id)
        if key not in log_by_key:
            issues.append(
                _issue(
                    root,
                    "CONTRA-LOG-MISSING",
                    "Canonical contradiction is not recorded in wiki/log.md",
                    entry,
                    id=entry.entry_id,
                )
            )

    for entry in log_entries:
        if entry.page is None:
            continue
        key = (entry.page, entry.entry_id)
        page_entry = page_by_key.get(key)
        if page_entry is None:
            issues.append(
                _issue(
                    root,
                    "CONTRA-LOG-ONLY",
                    "Contradiction is mentioned in wiki/log.md but has no canonical page entry",
                    entry,
                    page=entry.page,
                    id=entry.entry_id,
                )
            )
            continue
        page_status = page_entry.fields.get("status", "").strip("`").lower()
        log_status = entry.fields.get("status", "").strip("`").lower()
        if page_status in CONTRADICTION_STATUSES and log_status in CONTRADICTION_STATUSES:
            if page_status != log_status:
                issues.append(
                    _issue(
                        root,
                        "CONTRA-STATUS-MISMATCH",
                        "Contradiction status differs between canonical page and wiki/log.md",
                        entry,
                        page_status=page_status,
                        log_status=log_status,
                    )
                )

    issues.sort(key=lambda item: (item["path"], item["line"], item["code"]))
    serializable_pages = [
        {
            "id": entry.entry_id,
            "page": entry.page,
            "path": relative(entry.path, root),
            "line": entry.line,
            "status": entry.fields.get("status"),
        }
        for entry in page_entries
    ]
    serializable_log = [
        {
            "id": entry.entry_id,
            "page": entry.page,
            "path": relative(entry.path, root),
            "line": entry.line,
            "status": entry.fields.get("status"),
        }
        for entry in log_entries
    ]
    return {
        "page_entries": serializable_pages,
        "log_entries": serializable_log,
        "issues": issues,
        "counts": {
            "page_entries": len(page_entries),
            "log_entries": len(log_entries),
            "unresolved": sum(
                entry.fields.get("status", "").strip("`").lower() == "unresolved"
                for entry in page_entries
            ),
            "resolved": sum(
                entry.fields.get("status", "").strip("`").lower() == "resolved"
                for entry in page_entries
            ),
            "superseded": sum(
                entry.fields.get("status", "").strip("`").lower() == "superseded"
                for entry in page_entries
            ),
        },
    }
