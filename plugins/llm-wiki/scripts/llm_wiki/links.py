from __future__ import annotations

import re
from pathlib import Path

from .models import relative


WIKILINK_RE = re.compile(r"\[\[([^\]]+)\]\]")


def wiki_markdown(root: Path) -> list[Path]:
    wiki = root / "wiki"
    if not wiki.is_dir():
        return []
    return sorted(path for path in wiki.rglob("*.md") if path.is_file())


def link_target(raw: str) -> str:
    return raw.split("|", 1)[0].split("#", 1)[0].strip()


def build_link_report(root: Path) -> dict:
    files = wiki_markdown(root)
    stems: dict[str, list[str]] = {}
    for path in files:
        stems.setdefault(path.stem.casefold(), []).append(relative(path, root))

    links = []
    missing = []
    ambiguous = []
    for path in files:
        try:
            text = path.read_text(encoding="utf-8")
        except (OSError, UnicodeDecodeError):
            continue
        for line_number, line in enumerate(text.splitlines(), start=1):
            for match in WIKILINK_RE.finditer(line):
                raw = match.group(1)
                target = link_target(raw)
                if not target:
                    continue
                candidates = stems.get(Path(target).name.casefold(), [])
                item = {
                    "path": relative(path, root),
                    "line": line_number,
                    "raw": raw,
                    "target": target,
                    "candidates": candidates,
                }
                links.append(item)
                if not candidates:
                    missing.append(item)
                elif len(candidates) > 1:
                    ambiguous.append(item)
    return {
        "vault": str(root),
        "links": links,
        "missing": missing,
        "ambiguous": ambiguous,
        "counts": {
            "links": len(links),
            "missing": len(missing),
            "ambiguous": len(ambiguous),
        },
    }
