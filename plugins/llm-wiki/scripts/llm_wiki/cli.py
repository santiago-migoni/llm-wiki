from __future__ import annotations

import argparse
import json
import sys
from typing import Any

from . import SCHEMA_VERSION
from .hashes import build_hash_report
from .inventory import build_inventory
from .links import build_link_report
from .provenance import migrate_provenance
from .validate import validate
from .vault import resolve_root


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="llm-wiki",
        description="Deterministic tools for an LLM Wiki vault.",
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    for name in ("inventory", "links", "validate"):
        command = subparsers.add_parser(name)
        command.add_argument("vault")
        command.add_argument("--format", choices=("json", "text"), default="text")

    hashes = subparsers.add_parser("hashes")
    hashes.add_argument("vault")
    hashes.add_argument("--include-history", action="store_true")
    hashes.add_argument("--input", help="Optional source file to compare with the vault")
    hashes.add_argument("--format", choices=("json", "text"), default="text")

    migration = subparsers.add_parser(
        "migrate-provenance",
        help="Plan or explicitly apply legacy singular page provenance migration.",
    )
    migration.add_argument("vault")
    migration_mode = migration.add_mutually_exclusive_group()
    migration_mode.add_argument(
        "--check",
        dest="write",
        action="store_false",
        help="Only report the migration plan (the default).",
    )
    migration_mode.add_argument(
        "--write",
        dest="write",
        action="store_true",
        help="Apply the reviewed migration without creating a Git commit.",
    )
    migration.set_defaults(write=False)
    migration.add_argument("--format", choices=("json", "text"), default="text")
    return parser


def _envelope(command: str, data: dict[str, Any]) -> dict[str, Any]:
    return {"schema_version": SCHEMA_VERSION, "command": command, **data}


def _text(command: str, payload: dict[str, Any]) -> str:
    if command == "inventory":
        counts = payload["counts"]
        return "\n".join(
            (
                f"Vault: {payload['vault']}",
                f"Schemas: {counts['schemas']}",
                f"Source records: {counts['source_records']}",
                f"Canonical pages: {counts['canonical_pages']}",
                f"Syntheses: {counts['syntheses']}",
                f"Pending inbox files: {counts['pending_inbox']}",
                f"Git: {payload['git']['state']}",
            )
        )
    if command == "links":
        counts = payload["counts"]
        lines = [
            f"Vault: {payload['vault']}",
            f"Wikilinks: {counts['links']}",
            f"Missing: {counts['missing']}",
            f"Ambiguous: {counts['ambiguous']}",
        ]
        lines.extend(
            f"MISSING {item['path']}:{item['line']} -> {item['target']}"
            for item in payload["missing"]
        )
        lines.extend(
            f"AMBIGUOUS {item['path']}:{item['line']} -> {item['target']}"
            for item in payload["ambiguous"]
        )
        return "\n".join(lines)
    if command == "hashes":
        lines = [
            f"Vault: {payload['vault']}",
            f"Current source hashes: {len(payload['current'])}",
            f"Historical source hashes: {len(payload['historical'])}",
            f"Duplicate hashes: {len(payload['duplicates'])}",
        ]
        if payload["input"]:
            lines.append(f"Input SHA-256: {payload['input']['sha256']}")
            lines.append(f"Input matches: {len(payload['matches'])}")
        lines.extend(f"WARNING {warning}" for warning in payload["warnings"])
        return "\n".join(lines)
    if command == "validate":
        lines = [f"Vault: {payload['vault']}", f"Status: {payload['status'].upper()}"]
        provenance = payload.get("provenance", {})
        if provenance.get("legacy_pages"):
            lines.append(
                "Migrable legacy provenance pages: "
                + ", ".join(provenance["legacy_pages"])
            )
        lines.extend(
            f"[{item['severity'].upper()}] {item['id']} {item['path'] or '-'}: {item['message']}"
            for item in payload["findings"]
        )
        return "\n".join(lines)
    if command == "migrate-provenance":
        lines = [
            f"Vault: {payload['vault']}",
            f"Mode: {payload['mode']}",
            f"Status: {payload['status'].upper()}",
            f"Affected pages: {len(payload['affected_pages'])}",
        ]
        lines.extend(
            f"ERROR {item['details'].get('path', '-')} {item['code']}: {item['message']}"
            for item in payload["errors"]
        )
        for change in payload["changes"]:
            lines.append(f"AFFECTED {change['path']}")
            if change["warnings"]:
                lines.extend(f"WARNING {warning}" for warning in change["warnings"])
            if change["diff"]:
                lines.extend(change["diff"].rstrip("\n").splitlines())
        return "\n".join(lines)
    raise ValueError(f"unknown command: {command}")


def main(argv: list[str] | None = None) -> int:
    args = _parser().parse_args(argv)
    try:
        root = resolve_root(args.vault)
        if args.command == "inventory":
            data = build_inventory(root)
        elif args.command == "links":
            data = build_link_report(root)
        elif args.command == "hashes":
            from pathlib import Path

            data = build_hash_report(
                root,
                include_history=args.include_history,
                input_path=Path(args.input) if args.input else None,
            )
        elif args.command == "migrate-provenance":
            data = migrate_provenance(root, write=args.write)
        else:
            data = validate(root)
        payload = _envelope(args.command, data)
        if args.format == "json":
            print(json.dumps(payload, ensure_ascii=False, indent=2, sort_keys=True))
        else:
            print(_text(args.command, payload))
        if args.command == "validate" and data["findings"]:
            return 1
        if args.command == "migrate-provenance" and data["status"] == "blocked":
            return 1
        return 0
    except (OSError, ValueError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
