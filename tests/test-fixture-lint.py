#!/usr/bin/env python3
"""Validate the checked-in fixtures and their intentionally malformed cases."""

from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CLI = ROOT / "plugins/llm-wiki/scripts/llm-wiki"
FIXTURES = ROOT / "tests/fixtures"


def run(fixture: str) -> tuple[int, dict]:
    environment = os.environ.copy()
    environment["PYTHONDONTWRITEBYTECODE"] = "1"
    result = subprocess.run(
        [sys.executable, str(CLI), "validate", str(FIXTURES / fixture), "--format", "json"],
        check=False,
        capture_output=True,
        text=True,
        env=environment,
    )
    try:
        payload = json.loads(result.stdout)
    except json.JSONDecodeError as error:
        raise AssertionError(f"{fixture} did not return JSON: {result.stderr}{result.stdout}") from error
    return result.returncode, payload


def main() -> int:
    return_codes = {}

    code, payload = run("empty-vault")
    assert code == 0 and payload["status"] == "valid", "empty fixture must be valid"
    return_codes["empty-vault"] = code

    code, payload = run("populated-vault")
    assert code == 0 and payload["status"] == "valid", "populated fixture must be valid"
    return_codes["populated-vault"] = code

    code, payload = run("legacy-vault")
    assert code == 0 and payload["status"] == "valid", "legacy fixture must remain readable"
    assert payload["provenance"]["status"] == "migrable", "legacy fixture must be migration-visible"
    return_codes["legacy-vault"] = code

    code, payload = run("malformed-vault")
    messages = {item["message"] for item in payload["findings"]}
    assert code == 1 and payload["status"] == "invalid", "malformed fixture must fail validation"
    assert "Source slug is not lowercase kebab-case" in messages
    assert "Source record is missing extracted.md" in messages
    return_codes["malformed-vault"] = code

    code, payload = run("duplicate-source-files")
    messages = {item["message"] for item in payload["findings"]}
    assert code == 1 and payload["status"] == "invalid", "duplicate fixture must fail validation"
    assert "Source record has multiple current source.* files" in messages
    return_codes["duplicate-source-files"] = code

    print(f"fixture lint: ok ({', '.join(f'{name}={code}' for name, code in return_codes.items())})")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (AssertionError, OSError, subprocess.SubprocessError) as error:
        print(f"FAIL: {error}", file=sys.stderr)
        raise SystemExit(1)
