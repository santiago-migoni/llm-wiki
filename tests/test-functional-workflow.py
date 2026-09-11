#!/usr/bin/env python3
"""Exercise the observable deterministic vault workflow in a disposable Git vault."""

from __future__ import annotations

import hashlib
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
PLUGIN_SCRIPTS = REPO_ROOT / "plugins/llm-wiki/scripts"
CLI = PLUGIN_SCRIPTS / "llm-wiki"
ASSESSOR = REPO_ROOT / "plugins/llm-wiki/tests/assess-vault-scale.sh"
EMPTY_FIXTURE = REPO_ROOT / "tests/fixtures/empty-vault"
FUNCTIONAL_FIXTURE = REPO_ROOT / "tests/fixtures/functional-vault"


class FunctionalTestError(AssertionError):
    """Raised when an observable workflow invariant does not hold."""


def require(condition: bool, message: str) -> None:
    if not condition:
        raise FunctionalTestError(message)


def run_cli(*args: str) -> tuple[int, str, str]:
    environment = os.environ.copy()
    environment["PYTHONDONTWRITEBYTECODE"] = "1"
    result = subprocess.run(
        [sys.executable, str(CLI), *args],
        check=False,
        capture_output=True,
        text=True,
        env=environment,
    )
    return result.returncode, result.stdout, result.stderr


def cli_json(*args: str, expected_returncode: int = 0) -> dict:
    returncode, stdout, stderr = run_cli(*args, "--format", "json")
    require(
        returncode == expected_returncode,
        f"CLI {args} returned {returncode}, expected {expected_returncode}: {stderr}{stdout}",
    )
    try:
        return json.loads(stdout)
    except json.JSONDecodeError as error:
        raise FunctionalTestError(f"CLI {args} did not return JSON: {stdout}") from error


def git(root: Path, *args: str) -> str:
    result = subprocess.run(
        ["git", "-C", str(root), *args],
        check=True,
        capture_output=True,
        text=True,
    )
    return result.stdout.strip()


def snapshot(root: Path) -> dict[str, bytes]:
    return {
        path.relative_to(root).as_posix(): path.read_bytes()
        for path in root.rglob("*")
        if path.is_file() and ".git" not in path.relative_to(root).parts
    }


def expected_tree() -> list[str]:
    return [
        line.strip()
        for line in (FUNCTIONAL_FIXTURE / "expected-tree.txt").read_text(
            encoding="utf-8"
        ).splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    ]


def expected_history() -> list[str]:
    return [
        line.strip()
        for line in (FUNCTIONAL_FIXTURE / "expected-history.txt").read_text(
            encoding="utf-8"
        ).splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    ]


def source_bytes(name: str) -> bytes:
    return (FUNCTIONAL_FIXTURE / name).read_bytes()


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def extraction_text(slug: str, filename: str, digest: str, body: bytes) -> str:
    return (
        "---\n"
        "type: extraction\n"
        f"slug: {slug}\n"
        f"source: raw/sources/{slug}/source.txt\n"
        f"original-filename: {filename}\n"
        f"sha256: {digest}\n"
        "extracted: 2026-09-11\n"
        "format: txt\n"
        "method: functional fixture text reader\n"
        "status: complete\n"
        "coverage:\n"
        "  unit: pages\n"
        "  expected: 1\n"
        "  processed: 1\n"
        "warnings: []\n"
        "---\n\n"
        + body.decode("utf-8")
    )


def page_text(policy_digest: str, support_digest: str | None = None) -> str:
    if support_digest is None:
        sources = (
            "sources:\n"
            "  - slug: policy\n"
            "    source: raw/sources/policy/source.txt\n"
            "    extracted: raw/sources/policy/extracted.md\n"
            f"    sha256: {policy_digest}\n"
            "    role: primary\n"
        )
        body = (
            "# Policy\n\n"
            "- Review is described in the current policy.\n"
            "  - Evidence: `policy`, section Review cycle.\n"
        )
    else:
        sources = (
            "sources:\n"
            "  - slug: policy\n"
            "    source: raw/sources/policy/source.txt\n"
            "    extracted: raw/sources/policy/extracted.md\n"
            f"    sha256: {policy_digest}\n"
            "    role: primary\n"
            "  - slug: audit-report\n"
            "    source: raw/sources/audit-report/source.txt\n"
            "    extracted: raw/sources/audit-report/extracted.md\n"
            f"    sha256: {support_digest}\n"
            "    role: supporting\n"
        )
        body = (
            "# Policy\n\n"
            "- Review occurs quarterly in the current policy.\n"
            "  - Evidence: `policy`, section Review cycle.\n"
            "  - Support: `audit-report`, page 1.\n\n"
            "## Contradictions\n\n"
            "### review-frequency\n\n"
            "- Claim A: Review occurs annually.\n"
            "- Evidence A: `policy`, section Review cycle, historical wording.\n"
            "- Claim B: Review occurs quarterly.\n"
            "- Evidence B: `audit-report`, page 1.\n"
            "- Status: unresolved\n"
            "- Impact: Changes the compliance schedule.\n"
        )
    return (
        "---\n"
        "type: knowledge\n"
        "slug: policy\n"
        "created: 2026-09-11\n"
        "updated: 2026-09-11\n"
        + sources
        + "tags: []\n"
        "status: current\n"
        "---\n"
        + body
    )


def commit(root: Path, message: str, paths: list[str]) -> None:
    subprocess.run(["git", "-C", str(root), "add", "--", *paths], check=True)
    subprocess.run(["git", "-C", str(root), "commit", "-qm", message], check=True)


def commit_paths(root: Path) -> list[str]:
    return git(root, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD").splitlines()


def validate_read_only(root: Path) -> dict:
    before = snapshot(root)
    payload = cli_json("validate", str(root))
    require(payload["status"] == "valid", f"functional vault is invalid: {payload}")
    require(snapshot(root) == before, "validate mutated the vault")
    return payload


def main() -> int:
    policy_v1 = source_bytes("source-v1.txt")
    policy_v2 = source_bytes("source-v2.txt")
    support = source_bytes("source-support.txt")
    policy_v1_hash = sha256(policy_v1)
    policy_v2_hash = sha256(policy_v2)
    support_hash = sha256(support)

    with tempfile.TemporaryDirectory(prefix="llm-wiki-functional-") as directory:
        vault = Path(directory) / "vault"
        shutil.copytree(EMPTY_FIXTURE, vault)

        subprocess.run(["git", "init", "-q", str(vault)], check=True)
        git(vault, "config", "user.name", "LLM Wiki Functional Test")
        git(vault, "config", "user.email", "functional@example.invalid")
        commit(
            vault,
            "schema: initial wiki scaffold",
            ["AGENTS.md", "raw", "wiki"],
        )

        baseline = cli_json("inventory", str(vault))
        require(baseline["counts"]["source_records"] == 0, "empty scaffold has sources")
        require(baseline["counts"]["canonical_pages"] == 0, "empty scaffold has pages")
        validate_read_only(vault)
        links = cli_json("links", str(vault))
        require(links["counts"] == {
            "links": 0,
            "missing": 0,
            "ambiguous": 0,
            "missing_headings": 0,
            "index_routes": 0,
            "index_missing": 0,
        }, f"empty link report changed: {links}")

        pending = vault / "raw/inbox/policy.txt"
        pending.write_bytes(policy_v1)
        before_hash_check = snapshot(vault)
        new_report = cli_json("hashes", str(vault), "--input", str(pending))
        require(new_report["input"]["match_type"] == "new", "new input was not classified as new")
        require(snapshot(vault) == before_hash_check, "hash preflight mutated the vault")

        policy_record = vault / "raw/sources/policy"
        policy_record.mkdir(parents=True)
        shutil.move(str(pending), policy_record / "source.txt")
        (policy_record / "extracted.md").write_text(
            extraction_text("policy", "policy.txt", policy_v1_hash, policy_v1),
            encoding="utf-8",
        )
        (vault / "wiki/pages/policy.md").write_text(
            page_text(policy_v1_hash), encoding="utf-8"
        )
        (vault / "wiki/index.md").write_text("# Index\n\n[[policy]]\n", encoding="utf-8")
        (vault / "wiki/log.md").write_text(
            "# Log\n\n## [2026-09-11] ingest | policy\n\n- Operation: new\n",
            encoding="utf-8",
        )
        first_validation = validate_read_only(vault)
        require(first_validation["extractions"][0]["status"] == "complete", "v1 extraction is not complete")
        commit(
            vault,
            "ingest: policy v1",
            [
                "raw/sources/policy/source.txt",
                "raw/sources/policy/extracted.md",
                "wiki/pages/policy.md",
                "wiki/index.md",
                "wiki/log.md",
            ],
        )
        actual_v1_paths = commit_paths(vault)
        require(
            actual_v1_paths
            == [
                "raw/sources/policy/extracted.md",
                "raw/sources/policy/source.txt",
                "wiki/index.md",
                "wiki/log.md",
                "wiki/pages/policy.md",
            ],
            "v1 commit contains unexpected paths",
        )

        duplicate = vault / "raw/inbox/renamed-policy.txt"
        duplicate.write_bytes(policy_v1)
        before_duplicate = snapshot(vault)
        duplicate_report = cli_json("hashes", str(vault), "--input", str(duplicate))
        require(
            duplicate_report["input"]["match_type"] == "current-duplicate",
            "current duplicate was not classified as a no-op",
        )
        require(snapshot(vault) == before_duplicate, "duplicate preflight changed the vault")
        require(len(git(vault, "log", "--format=%H").splitlines()) == 2, "duplicate created a commit")
        duplicate.unlink()

        update = vault / "raw/inbox/policy-update.txt"
        update.write_bytes(policy_v2)
        update_report = cli_json("hashes", str(vault), "--input", str(update))
        require(update_report["input"]["match_type"] == "new", "changed input was not classified as new")
        shutil.move(str(update), policy_record / "source.txt")

        support_record = vault / "raw/sources/audit-report"
        support_record.mkdir(parents=True)
        (support_record / "source.txt").write_bytes(support)
        (policy_record / "extracted.md").write_text(
            extraction_text("policy", "policy-update.txt", policy_v2_hash, policy_v2),
            encoding="utf-8",
        )
        (support_record / "extracted.md").write_text(
            extraction_text("audit-report", "source-support.txt", support_hash, support),
            encoding="utf-8",
        )
        (vault / "wiki/pages/policy.md").write_text(
            page_text(policy_v2_hash, support_hash), encoding="utf-8"
        )
        (vault / "wiki/log.md").write_text(
            "# Log\n\n"
            "## [2026-09-11] ingest | policy update\n\n"
            "- Operation: update\n\n"
            "## [2026-09-11] contradiction | review-frequency\n"
            "- Page: [[policy]]\n"
            "- Status: unresolved\n",
            encoding="utf-8",
        )

        updated = validate_read_only(vault)
        require(updated["provenance"]["status"] == "current", "multisource provenance is not current")
        require(len(updated["extractions"]) == 2, "updated vault did not report both extractions")
        require(updated["contradictions"]["counts"]["unresolved"] == 1, "active contradiction is missing")
        require(updated["links"]["index_missing"] == 0, "canonical page is not indexed")
        inventory = cli_json("inventory", str(vault))
        require(inventory["counts"]["source_records"] == 2, "updated inventory has the wrong source count")
        scale = subprocess.run(
            ["bash", str(ASSESSOR), str(vault)],
            check=False,
            capture_output=True,
            text=True,
        )
        require(scale.returncode == 0, f"functional vault scale check failed: {scale.stdout}{scale.stderr}")
        require("Scale status: GREEN" in scale.stdout, "functional vault left the GREEN scale boundary")

        commit(
            vault,
            "ingest: policy v2",
            [
                "raw/sources/policy/source.txt",
                "raw/sources/policy/extracted.md",
                "raw/sources/audit-report/source.txt",
                "raw/sources/audit-report/extracted.md",
                "wiki/pages/policy.md",
                "wiki/index.md",
                "wiki/log.md",
            ],
        )
        actual_v2_paths = commit_paths(vault)
        require(
            actual_v2_paths
            == [
                "raw/sources/audit-report/extracted.md",
                "raw/sources/audit-report/source.txt",
                "raw/sources/policy/extracted.md",
                "raw/sources/policy/source.txt",
                "wiki/log.md",
                "wiki/pages/policy.md",
            ],
            f"v2 commit contains unexpected paths: {actual_v2_paths}",
        )

        historical = vault / "raw/inbox/historical-policy.txt"
        historical.write_bytes(policy_v1)
        before_history_check = snapshot(vault)
        historical_report = cli_json(
            "hashes",
            str(vault),
            "--include-history",
            "--input",
            str(historical),
        )
        require(
            historical_report["input"]["match_type"] == "historical-duplicate",
            "historical duplicate was not classified correctly",
        )
        reversion = cli_json(
            "hashes",
            str(vault),
            "--revert-to",
            policy_v1_hash,
        )
        require(reversion["reversion"]["status"] == "requested", "historical reversion was not reported")
        require(reversion["reversion"]["kind"] == "historical", "reversion was not historical")
        require(snapshot(vault) == before_history_check, "hash/history checks mutated the vault")
        historical.unlink()

        for command in (
            ("inventory", str(vault)),
            ("links", str(vault)),
            ("validate", str(vault)),
            ("hashes", str(vault)),
        ):
            before_read = snapshot(vault)
            cli_json(*command)
            require(snapshot(vault) == before_read, f"{command[0]} mutated the vault")

        require(sorted(snapshot(vault)) == expected_tree(), "final tree differs from expected snapshot")
        require(
            git(vault, "log", "--format=%s", "--reverse").splitlines() == expected_history(),
            "Git history differs from expected snapshot",
        )

    print("functional workflow: ok (I-001, G-001..G-004, Q-001..Q-003, L-001..L-002, M-001, K-001, S-001)")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (FunctionalTestError, OSError, subprocess.SubprocessError) as error:
        print(f"FAIL: {error}", file=sys.stderr)
        raise SystemExit(1)
