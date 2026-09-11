"""A deliberately small parser for the YAML subset used by LLM Wiki.

This is not a general YAML parser. It accepts mappings, lists, scalar values,
and lists of mappings with two-space indentation. Unsupported YAML features
fail closed with a readable error instead of being interpreted approximately.
"""

from __future__ import annotations

import csv
import io
import json
import re
from dataclasses import dataclass
from pathlib import Path
from typing import Any


KEY_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_-]*$")
INTEGER_RE = re.compile(r"^-?(?:0|[1-9][0-9]*)$")
UNSUPPORTED_VALUE_PREFIXES = ("&", "*", "!", "|", ">", "{")


class FrontmatterError(ValueError):
    """Raised when frontmatter is absent, malformed, or outside the subset."""


@dataclass(frozen=True)
class ParsedDocument:
    metadata: dict[str, Any]
    body: str


@dataclass(frozen=True)
class _Line:
    number: int
    indent: int
    text: str


def parse_document(text: str, *, source: str = "<document>") -> ParsedDocument:
    lines = text.splitlines()
    if not lines or lines[0].strip() != "---":
        raise FrontmatterError(f"{source}: missing opening frontmatter delimiter")

    closing = next(
        (index for index, line in enumerate(lines[1:], start=1) if line.strip() == "---"),
        None,
    )
    if closing is None:
        raise FrontmatterError(f"{source}: missing closing frontmatter delimiter")

    metadata = parse_yaml_subset(lines[1:closing], source=source)
    body = "\n".join(lines[closing + 1 :])
    if text.endswith("\n"):
        body += "\n"
    return ParsedDocument(metadata=metadata, body=body)


def parse_file(path: Path) -> ParsedDocument:
    try:
        text = path.read_text(encoding="utf-8")
    except UnicodeDecodeError as error:
        raise FrontmatterError(f"{path}: file is not UTF-8") from error
    except OSError as error:
        raise FrontmatterError(f"{path}: cannot read file: {error}") from error
    return parse_document(text, source=str(path))


def parse_yaml_subset(lines: list[str], *, source: str = "<frontmatter>") -> dict[str, Any]:
    tokens: list[_Line] = []
    for number, raw in enumerate(lines, start=2):
        if "\t" in raw:
            raise FrontmatterError(f"{source}:{number}: tabs are not supported")
        stripped = raw.strip()
        if not stripped or stripped.startswith("#"):
            continue
        indent = len(raw) - len(raw.lstrip(" "))
        if indent % 2:
            raise FrontmatterError(
                f"{source}:{number}: indentation must use multiples of two spaces"
            )
        tokens.append(_Line(number=number, indent=indent, text=raw[indent:]))

    if not tokens:
        return {}
    if tokens[0].indent != 0 or tokens[0].text.startswith("- "):
        raise FrontmatterError(f"{source}:{tokens[0].number}: root must be a mapping")

    value, position = _parse_mapping(tokens, 0, 0, source)
    if position != len(tokens):
        line = tokens[position]
        raise FrontmatterError(f"{source}:{line.number}: unexpected indentation")
    return value


def _parse_mapping(
    tokens: list[_Line], position: int, indent: int, source: str
) -> tuple[dict[str, Any], int]:
    result: dict[str, Any] = {}
    while position < len(tokens):
        line = tokens[position]
        if line.indent < indent:
            break
        if line.indent > indent:
            raise FrontmatterError(f"{source}:{line.number}: unexpected indentation")
        if line.text.startswith("- "):
            break

        key, raw_value = _split_key_value(line, source)
        if key in result:
            raise FrontmatterError(f"{source}:{line.number}: duplicate key: {key}")
        position += 1

        if raw_value:
            result[key] = _parse_scalar(raw_value, source, line.number)
            continue

        if position >= len(tokens) or tokens[position].indent <= indent:
            result[key] = None
            continue
        if tokens[position].indent != indent + 2:
            raise FrontmatterError(
                f"{source}:{tokens[position].number}: nested values must indent two spaces"
            )
        if tokens[position].text.startswith("- "):
            child, position = _parse_list(tokens, position, indent + 2, source)
        else:
            child, position = _parse_mapping(tokens, position, indent + 2, source)
        result[key] = child
    return result, position


def _parse_list(
    tokens: list[_Line], position: int, indent: int, source: str
) -> tuple[list[Any], int]:
    result: list[Any] = []
    while position < len(tokens):
        line = tokens[position]
        if line.indent < indent:
            break
        if line.indent != indent or not line.text.startswith("- "):
            break

        raw_item = line.text[2:].strip()
        if not raw_item:
            raise FrontmatterError(f"{source}:{line.number}: empty list items are unsupported")

        if ":" not in raw_item:
            result.append(_parse_scalar(raw_item, source, line.number))
            position += 1
            continue

        first_key, first_value = _split_inline_mapping(raw_item, source, line.number)
        item: dict[str, Any] = {
            first_key: _parse_scalar(first_value, source, line.number)
            if first_value
            else None
        }
        position += 1

        while position < len(tokens) and tokens[position].indent > indent:
            child = tokens[position]
            if child.indent != indent + 2 or child.text.startswith("- "):
                raise FrontmatterError(
                    f"{source}:{child.number}: list mappings use two-space indentation"
                )
            key, raw_value = _split_key_value(child, source)
            if key in item:
                raise FrontmatterError(f"{source}:{child.number}: duplicate key: {key}")
            if not raw_value:
                raise FrontmatterError(
                    f"{source}:{child.number}: nested containers inside list mappings are unsupported"
                )
            item[key] = _parse_scalar(raw_value, source, child.number)
            position += 1
        result.append(item)
    return result, position


def _split_key_value(line: _Line, source: str) -> tuple[str, str]:
    if ":" not in line.text:
        raise FrontmatterError(f"{source}:{line.number}: expected key: value")
    key, value = line.text.split(":", 1)
    key = key.strip()
    if not KEY_RE.fullmatch(key):
        raise FrontmatterError(f"{source}:{line.number}: unsupported key: {key!r}")
    return key, value.strip()


def _split_inline_mapping(raw: str, source: str, number: int) -> tuple[str, str]:
    key, value = raw.split(":", 1)
    key = key.strip()
    if not KEY_RE.fullmatch(key):
        raise FrontmatterError(f"{source}:{number}: unsupported key: {key!r}")
    return key, value.strip()


def _parse_scalar(raw: str, source: str, number: int) -> Any:
    if raw.startswith(UNSUPPORTED_VALUE_PREFIXES):
        raise FrontmatterError(f"{source}:{number}: unsupported YAML feature: {raw!r}")
    if raw == "[]":
        return []
    if raw.startswith("["):
        if not raw.endswith("]"):
            raise FrontmatterError(f"{source}:{number}: unterminated inline list")
        inner = raw[1:-1].strip()
        if not inner:
            return []
        try:
            values = next(csv.reader(io.StringIO(inner), skipinitialspace=True))
        except csv.Error as error:
            raise FrontmatterError(f"{source}:{number}: malformed inline list") from error
        return [_parse_scalar(value.strip(), source, number) for value in values]
    if raw.startswith('"'):
        try:
            value = json.loads(raw)
        except json.JSONDecodeError as error:
            raise FrontmatterError(f"{source}:{number}: invalid quoted string") from error
        if not isinstance(value, str):
            raise FrontmatterError(f"{source}:{number}: expected a quoted string")
        return value
    if raw.startswith("'"):
        if len(raw) < 2 or not raw.endswith("'"):
            raise FrontmatterError(f"{source}:{number}: invalid single-quoted string")
        return raw[1:-1].replace("''", "'")
    lowered = raw.lower()
    if lowered in {"true", "false"}:
        return lowered == "true"
    if lowered in {"null", "~"}:
        return None
    if INTEGER_RE.fullmatch(raw):
        return int(raw)
    if " #" in raw:
        raw = raw.split(" #", 1)[0].rstrip()
    return raw
