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
    hashes.add_argument(
        "--revert-to",
        metavar="SHA256",
        help="Explicitly report a requested restoration target without changing files",
    )
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
        lines = [
            f"Vault: {payload['vault']}",
            f"Schemas: {counts['schemas']}",
            f"Source records: {counts['source_records']}",
            f"Canonical pages: {counts['canonical_pages']}",
            f"Syntheses: {counts['syntheses']}",
            f"Pending inbox files: {counts['pending_inbox']}",
            f"Git: {payload['git']['state']}",
        ]
        lines = [
            *lines,
            *(
                f"WARNING {item['code']} {item.get('path', '-')}"
                f": {item['message']}"
                for item in payload["warnings"]
            ),
        ]
        return "\n".join(lines)
    if command == "links":
        counts = payload["counts"]
        lines = [
            f"Vault: {payload['vault']}",
            f"Wikilinks: {counts['links']}",
            f"Missing: {counts['missing']}",
            f"Ambiguous: {counts['ambiguous']}",
            f"Missing headings: {counts['missing_headings']}",
            f"Unindexed canonical pages: {counts['index_missing']}",
        ]
        lines.extend(
            f"MISSING {item['path']}:{item['line']} -> {item['target']}"
            for item in payload["missing"]
        )
        lines.extend(
            f"AMBIGUOUS {item['path']}:{item['line']} -> {item['target']}"
            for item in payload["ambiguous"]
        )
        lines.extend(
            f"MISSING HEADING {item['path']}:{item['line']} -> {item['target']}#{item['heading']}"
            for item in payload["missing_headings"]
        )
        lines.extend(
            f"INDEX MISSING {item['path']} (expected in {item['index']})"
            for item in payload["index_missing"]
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
            lines.append(f"Input classification: {payload['input']['match_type']}")
            lines.append(f"Input matches: {len(payload['matches'])}")
        reversion = payload["reversion"]
        if reversion["requested"]:
            lines.append(
                f"Reversion request: {reversion['status']}"
                + (f" ({reversion['kind']})" if reversion["kind"] else "")
            )
            lines.append(f"Reversion matches: {len(reversion['matches'])}")
        lines.extend(
            f"DUPLICATE {item['kind']} {item['sha256']} ({len(item['locations'])} locations)"
            for item in payload["duplicates"]
        )
        lines.extend(f"WARNING {warning}" for warning in payload["warnings"])
        return "\n".join(lines)
    if command == "validate":
        lines = [f"Vault: {payload['vault']}", f"Status: {payload['status'].upper()}"]
        lines.extend(
            "EXTRACTION "
            f"{item['slug']} status={item['status']} "
            f"coverage={item['coverage'].get('processed', '?')}/"
            f"{item['coverage'].get('expected', '?')} "
            f"{item['coverage'].get('unit', '?')}"
            for item in payload.get("extractions", [])
        )
        contradiction_counts = payload.get("contradictions", {}).get("counts", {})
        if contradiction_counts:
            lines.append(
                "CONTRADICTIONS "
                f"page={contradiction_counts.get('page_entries', 0)} "
                f"log={contradiction_counts.get('log_entries', 0)} "
                f"unresolved={contradiction_counts.get('unresolved', 0)}"
            )
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
        lines.extend(
            f"WARNING {item['code']} {item.get('path', '-')}"
            f": {item['message']}"
            for item in payload.get("inventory_warnings", [])
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
                include_history=args.include_history or bool(args.revert_to),
                input_path=Path(args.input) if args.input else None,
                revert_to=args.revert_to,
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
        if (
            args.command == "hashes"
            and data["reversion"]["requested"]
            and data["reversion"]["status"] == "not-found"
        ):
            return 1
        return 0
    except (OSError, ValueError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
