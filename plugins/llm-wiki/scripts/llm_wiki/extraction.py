"""Validation for format-specific extraction metadata."""

from __future__ import annotations

import re
from pathlib import Path
from typing import Any

from .models import relative


STATUS_VALUES = {"complete", "representative", "partial", "unsupported"}
UNIT_VALUES = {"pages", "slides", "sheets", "rows", "seconds", "regions"}
DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}$")
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")

FORMAT_BY_SUFFIX = {
    ".md": "markdown",
    ".markdown": "markdown",
    ".txt": "txt",
    ".html": "html",
    ".htm": "html",
    ".pdf": "pdf",
    ".docx": "docx",
    ".xlsx": "xlsx",
    ".xls": "xlsx",
    ".csv": "csv",
    ".tsv": "csv",
    ".pptx": "pptx",
    ".ppt": "pptx",
    ".png": "image",
    ".jpg": "image",
    ".jpeg": "image",
    ".gif": "image",
    ".webp": "image",
    ".bmp": "image",
    ".svg": "image",
    ".mp3": "audio",
    ".wav": "audio",
    ".m4a": "audio",
    ".flac": "audio",
    ".mp4": "video",
    ".mov": "video",
    ".mkv": "video",
    ".webm": "video",
}

UNIT_BY_FORMAT = {
    "markdown": "pages",
    "txt": "pages",
    "html": "pages",
    "pdf": "pages",
    "docx": "pages",
    "xlsx": "sheets",
    "csv": "rows",
    "pptx": "slides",
    "image": "regions",
    "audio": "seconds",
    "video": "seconds",
    "other": "pages",
}


def _issue(code: str, message: str, **details: Any) -> dict[str, Any]:
    return {"code": code, "severity": "high", "message": message, "details": details}


def expected_format(source_path: Path) -> str:
    return FORMAT_BY_SUFFIX.get(source_path.suffix.lower(), "other")


def inspect_extraction(
    root: Path,
    extraction_path: Path,
    source_path: Path,
    slug: str,
    metadata: dict[str, Any],
) -> dict[str, Any]:
    """Return stable issues and the machine-readable coverage summary."""

    issues: list[dict[str, Any]] = []
    source_format = expected_format(source_path)
    expected_unit = UNIT_BY_FORMAT[source_format]

    if metadata.get("type") != "extraction":
        issues.append(
            _issue(
                "EXT-TYPE",
                "Extraction frontmatter type must be extraction",
                expected="extraction",
                actual=metadata.get("type"),
            )
        )

    if metadata.get("slug") != slug:
        issues.append(
            _issue(
                "EXT-SLUG",
                "Extraction slug does not match source record",
                expected=slug,
                actual=metadata.get("slug"),
            )
        )

    for field in ("original-filename", "method"):
        value = metadata.get(field)
        if not isinstance(value, str) or not value.strip():
            issues.append(
                _issue(
                    "EXT-METADATA-MISSING",
                    "Extraction metadata field is missing or invalid",
                    field=field,
                )
            )

    extracted_date = metadata.get("extracted")
    if not isinstance(extracted_date, str) or not DATE_RE.fullmatch(extracted_date):
        issues.append(
            _issue(
                "EXT-DATE",
                "Extraction date must use YYYY-MM-DD",
                field="extracted",
                actual=extracted_date,
            )
        )

    recorded_hash = metadata.get("sha256")
    if recorded_hash is not None and (
        not isinstance(recorded_hash, str) or not SHA256_RE.fullmatch(recorded_hash)
    ):
        issues.append(
            _issue(
                "EXT-SHA256",
                "Extraction SHA-256 must be 64 lowercase hexadecimal characters",
                field="sha256",
            )
        )

    status = metadata.get("status")
    if not isinstance(status, str) or status not in STATUS_VALUES:
        issues.append(
            _issue(
                "EXT-STATUS",
                "Extraction status is not supported",
                allowed=sorted(STATUS_VALUES),
                actual=status,
            )
        )

    format_name = metadata.get("format")
    if format_name != source_format:
        issues.append(
            _issue(
                "EXT-FORMAT",
                "Extraction format does not match current source extension",
                expected=source_format,
                actual=format_name,
            )
        )

    coverage = metadata.get("coverage")
    coverage_summary: dict[str, Any] = {}
    valid_counts = False
    if not isinstance(coverage, dict):
        issues.append(
            _issue(
                "EXT-COVERAGE-MISSING",
                "Extraction coverage must declare unit, expected, and processed",
            )
        )
    else:
        unit = coverage.get("unit")
        expected = coverage.get("expected")
        processed = coverage.get("processed")
        coverage_summary = {
            "unit": unit,
            "expected": expected,
            "processed": processed,
        }
        if not isinstance(unit, str) or unit not in UNIT_VALUES:
            issues.append(
                _issue(
                    "EXT-COVERAGE-UNIT",
                    "Extraction coverage unit is not supported",
                    allowed=sorted(UNIT_VALUES),
                    actual=unit,
                )
            )
        elif unit != expected_unit:
            issues.append(
                _issue(
                    "EXT-COVERAGE-UNIT",
                    "Extraction coverage unit does not match source format",
                    expected=expected_unit,
                    actual=unit,
                )
            )

        if (
            isinstance(expected, bool)
            or not isinstance(expected, int)
            or expected < 0
            or isinstance(processed, bool)
            or not isinstance(processed, int)
            or processed < 0
        ):
            issues.append(
                _issue(
                    "EXT-COVERAGE-COUNT",
                    "Extraction coverage counts must be non-negative integers",
                )
            )
        else:
            valid_counts = True
            if processed > expected:
                issues.append(
                    _issue(
                        "EXT-COVERAGE-RANGE",
                        "Extraction processed coverage cannot exceed expected coverage",
                        expected=expected,
                        processed=processed,
                    )
                )

    warnings = metadata.get("warnings")
    valid_warnings = isinstance(warnings, list) and all(
        isinstance(warning, str) and warning.strip() for warning in warnings
    )
    if not valid_warnings:
        issues.append(
            _issue(
                "EXT-WARNINGS",
                "Extraction warnings must be a list of non-empty strings",
            )
        )
        warnings = []

    if isinstance(status, str) and status in {"representative", "partial", "unsupported"} and not warnings:
        issues.append(
            _issue(
                "EXT-WARNING-REQUIRED",
                "Non-complete extraction status requires at least one warning",
                status=status,
            )
        )

    if valid_counts:
        expected = coverage["expected"]
        processed = coverage["processed"]
        if status == "complete" and processed != expected:
            issues.append(
                _issue(
                    "EXT-STATUS-COVERAGE",
                    "Complete extraction must process all expected coverage",
                    expected=expected,
                    processed=processed,
                )
            )
        if status == "representative" and processed >= expected:
            issues.append(
                _issue(
                    "EXT-STATUS-COVERAGE",
                    "Representative extraction must identify a subset of expected coverage",
                    expected=expected,
                    processed=processed,
                )
            )
        if status == "unsupported" and processed != 0:
            issues.append(
                _issue(
                    "EXT-STATUS-COVERAGE",
                    "Unsupported extraction must process zero coverage units",
                    processed=processed,
                )
            )

    return {
        "slug": slug,
        "path": relative(extraction_path, root),
        "format": source_format,
        "status": status,
        "coverage": coverage_summary,
        "warnings": warnings,
        "issues": issues,
    }
