from __future__ import annotations

from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Any


@dataclass
class Finding:
    id: str
    severity: str
    message: str
    path: str | None = None
    details: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class SourceRecord:
    slug: str
    path: str
    sources: list[str]
    extraction: str | None

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


def relative(path: Path, root: Path) -> str:
    return path.relative_to(root).as_posix()
