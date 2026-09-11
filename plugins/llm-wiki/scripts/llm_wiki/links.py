from __future__ import annotations

import re
from pathlib import Path

from .models import relative


WIKILINK_RE = re.compile(r"\[\[([^\]]+)\]\]")
HEADING_RE = re.compile(r"^#{1,6}[ \t]+(?P<title>.+?)[ \t]*$")


def wiki_markdown(root: Path) -> list[Path]:
    wiki = root / "wiki"
    if not wiki.is_dir():
        return []
    return sorted(path for path in wiki.rglob("*.md") if path.is_file())


def _markdown_files(path: Path) -> list[Path]:
    if not path.is_dir():
        return []
    return sorted(item for item in path.rglob("*.md") if item.is_file())


def link_target(raw: str) -> str:
    return split_link(raw)[0]


def split_link(raw: str) -> tuple[str, str | None, str | None]:
    target, has_alias, alias = raw.partition("|")
    path, has_heading, heading = target.strip().partition("#")
    return (
        path.strip(),
        heading.strip() if has_heading and heading.strip() else None,
        alias.strip() if has_alias and alias.strip() else None,
    )


def heading_key(value: str) -> str:
    value = re.sub(r"<[^>]*>", "", value).strip()
    value = re.sub(r"[ \t]+#+[ \t]*$", "", value)
    value = value.casefold()
    value = re.sub(r"[^\w\s-]", "", value, flags=re.UNICODE)
    return re.sub(r"[-\s]+", "-", value).strip("-")


def markdown_headings(text: str) -> list[dict]:
    headings = []
    for line_number, line in enumerate(text.splitlines(), start=1):
        match = HEADING_RE.fullmatch(line)
        if not match:
            continue
        title = match.group("title").strip()
        key = heading_key(title)
        if key:
            headings.append({"title": title, "line": line_number, "key": key})
    return headings


def _stem_target(target: str) -> str:
    name = Path(target).name
    return name[:-3] if name.casefold().endswith(".md") else name


def build_link_report(root: Path) -> dict:
    files = wiki_markdown(root)
    stems: dict[str, list[str]] = {}
    documents: dict[Path, str] = {}
    headings: dict[Path, list[dict]] = {}
    for path in files:
        try:
            text = path.read_text(encoding="utf-8")
        except (OSError, UnicodeDecodeError):
            continue
        documents[path] = text
        headings[path] = markdown_headings(text)
        stems.setdefault(path.stem.casefold(), []).append(relative(path, root))

    links = []
    missing = []
    ambiguous = []
    missing_headings = []
    invalid_heading_links: set[tuple[str, int, str]] = set()
    for path in files:
        text = documents.get(path)
        if text is None:
            continue
        for line_number, line in enumerate(text.splitlines(), start=1):
            for match in WIKILINK_RE.finditer(line):
                raw = match.group(1)
                target, heading, alias = split_link(raw)
                if not target and not heading:
                    continue
                if target:
                    candidates = stems.get(_stem_target(target).casefold(), [])
                else:
                    candidates = [relative(path, root)]
                item = {
                    "path": relative(path, root),
                    "line": line_number,
                    "raw": raw,
                    "target": target,
                    "heading": heading,
                    "alias": alias,
                    "candidates": candidates,
                }
                links.append(item)
                if not candidates:
                    missing.append(item)
                elif len(candidates) > 1:
                    ambiguous.append(item)
                elif heading:
                    candidate_path = root / candidates[0]
                    available = headings.get(candidate_path, [])
                    if heading_key(heading) not in {entry["key"] for entry in available}:
                        invalid_heading_links.add((item["path"], item["line"], item["raw"]))
                        missing_headings.append(
                            {
                                **item,
                                "available_headings": [entry["title"] for entry in available],
                            }
                        )

    canonical_pages = [relative(path, root) for path in _markdown_files(root / "wiki/pages")]
    indexed_pages = sorted(
        {
            candidate
            for item in links
            if item["path"] == "wiki/index.md" and len(item["candidates"]) == 1
            for candidate in item["candidates"]
            if candidate in canonical_pages and not any(
                (item["path"], item["line"], item["raw"]) == invalid
                for invalid in invalid_heading_links
            )
        }
    )
    index_missing = [
        {"path": page, "index": "wiki/index.md"}
        for page in canonical_pages
        if page not in indexed_pages
    ]
    return {
        "vault": str(root),
        "links": links,
        "missing": missing,
        "ambiguous": ambiguous,
        "missing_headings": missing_headings,
        "index_routes": indexed_pages,
        "index_missing": index_missing,
        "counts": {
            "links": len(links),
            "missing": len(missing),
            "ambiguous": len(ambiguous),
            "missing_headings": len(missing_headings),
            "index_routes": len(indexed_pages),
            "index_missing": len(index_missing),
        },
    }
