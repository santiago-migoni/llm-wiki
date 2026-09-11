from __future__ import annotations

import subprocess
from pathlib import Path


REQUIRED_DIRECTORIES = (
    "raw/inbox",
    "raw/sources",
    "wiki",
    "wiki/pages",
)


def resolve_root(value: str) -> Path:
    root = Path(value).expanduser().resolve()
    if not root.exists():
        raise ValueError(f"vault root does not exist: {root}")
    if not root.is_dir():
        raise ValueError(f"vault root is not a directory: {root}")
    return root


def schema_paths(root: Path) -> list[Path]:
    return [path for path in (root / "AGENTS.md", root / "CLAUDE.md") if path.exists()]


def run_git(root: Path, *args: str, binary: bool = False):
    try:
        return subprocess.run(
            ["git", "-C", str(root), *args],
            check=False,
            capture_output=True,
            text=not binary,
        )
    except OSError:
        return None


def git_available(root: Path) -> bool:
    result = run_git(root, "rev-parse", "--is-inside-work-tree")
    return bool(result and result.returncode == 0 and result.stdout.strip() == "true")
