import hashlib
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[2] / "scripts"
CLI = SCRIPTS / "llm-wiki"


def run_cli(*args):
    return subprocess.run(
        [sys.executable, str(CLI), *args],
        check=False,
        capture_output=True,
        text=True,
    )


def create_vault(root: Path) -> None:
    for directory in (
        "raw/inbox",
        "raw/sources",
        "wiki/pages",
        "wiki/syntheses",
    ):
        (root / directory).mkdir(parents=True, exist_ok=True)
    (root / "AGENTS.md").write_text("# Schema\n", encoding="utf-8")
    for name in ("index.md", "overview.md", "log.md"):
        contents = "# Empty\n\n[[policy]]\n" if name == "index.md" else "# Empty\n"
        (root / "wiki" / name).write_text(contents, encoding="utf-8")


def add_source(root: Path, slug: str, content: bytes) -> str:
    record = root / "raw/sources" / slug
    record.mkdir(parents=True)
    source_path = record / "source.txt"
    source_path.write_bytes(content)
    digest = hashlib.sha256(content).hexdigest()
    (record / "extracted.md").write_text(
        "---\n"
        "type: extraction\n"
        f"slug: {slug}\n"
        f"source: raw/sources/{slug}/source.txt\n"
        "original-filename: source.txt\n"
        f"sha256: {digest}\n"
        "extracted: 2026-09-11\n"
        "format: txt\n"
        "method: test fixture\n"
        "status: complete\n"
        "coverage:\n"
        "  unit: pages\n"
        "  expected: 1\n"
        "  processed: 1\n"
        "warnings: []\n"
        "---\n"
        "Extracted text.\n",
        encoding="utf-8",
    )
    return digest


def page_frontmatter(slug: str, sources: str) -> str:
    return (
        "---\n"
        "type: knowledge\n"
        f"slug: {slug}\n"
        "created: 2026-09-11\n"
        "updated: 2026-09-11\n"
        f"{sources}"
        "tags: []\n"
        "status: current\n"
        "---\n"
    )


class ProvenanceTests(unittest.TestCase):
    def test_validate_checks_every_structured_source_hash(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            first_hash = add_source(root, "alpha", b"alpha")
            second_hash = add_source(root, "beta", b"beta")
            sources = (
                "sources:\n"
                "  - slug: alpha\n"
                "    source: raw/sources/alpha/source.txt\n"
                "    extracted: raw/sources/alpha/extracted.md\n"
                f"    sha256: {first_hash}\n"
                "    role: primary\n"
                "  - slug: beta\n"
                "    source: raw/sources/beta/source.txt\n"
                "    extracted: raw/sources/beta/extracted.md\n"
                f"    sha256: {second_hash}\n"
                "    role: supporting\n"
            )
            (root / "wiki/pages/policy.md").write_text(
                page_frontmatter("policy", sources)
                + "# Policy\n\n"
                "- The policy is reviewed annually.\n"
                "  - Evidence: `alpha`, section Review cycle.\n"
                "  - Support: `beta`, page 2.\n",
                encoding="utf-8",
            )

            result = run_cli("validate", str(root), "--format", "json")
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertEqual(json.loads(result.stdout)["provenance"]["status"], "current")

            (root / "raw/sources/beta/source.txt").write_bytes(b"beta changed")
            result = run_cli("validate", str(root), "--format", "json")
            self.assertEqual(result.returncode, 1)
            findings = json.loads(result.stdout)["findings"]
            self.assertTrue(
                any(
                    item["message"] == "Provenance SHA-256 does not match the current source"
                    and item["details"]["index"] == 1
                    for item in findings
                )
            )

    def test_validate_checks_order_and_declared_citation_slugs(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            first_hash = add_source(root, "alpha", b"alpha")
            second_hash = add_source(root, "beta", b"beta")
            sources = (
                "sources:\n"
                "  - slug: beta\n"
                "    source: raw/sources/beta/source.txt\n"
                "    extracted: raw/sources/beta/extracted.md\n"
                f"    sha256: {second_hash}\n"
                "    role: supporting\n"
                "  - slug: alpha\n"
                "    source: raw/sources/alpha/source.txt\n"
                "    extracted: raw/sources/alpha/extracted.md\n"
                f"    sha256: {first_hash}\n"
                "    role: primary\n"
            )
            (root / "wiki/pages/policy.md").write_text(
                page_frontmatter("policy", sources)
                + "# Policy\n\n"
                "- Claim.\n"
                "  - Evidence: `unknown`, section One.\n",
                encoding="utf-8",
            )
            result = run_cli("validate", str(root), "--format", "json")
            self.assertEqual(result.returncode, 1)
            codes = {item["details"].get("code") for item in json.loads(result.stdout)["findings"]}
            self.assertIn("PROV-ORDER", codes)
            self.assertTrue(
                any(
                    item["message"] == "Citation references a source slug absent from page sources"
                    for item in json.loads(result.stdout)["findings"]
                )
            )

    def test_nested_provenance_and_canonical_page_are_valid(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            digest = add_source(root, "fundamentos/actividades", b"activities")
            page = root / "wiki/pages/fundamentos/actividades.md"
            page.parent.mkdir(parents=True)
            page.write_text(
                page_frontmatter(
                    "fundamentos/actividades",
                    "sources:\n"
                    "  - slug: fundamentos/actividades\n"
                    "    source: raw/sources/fundamentos/actividades/source.txt\n"
                    "    extracted: raw/sources/fundamentos/actividades/extracted.md\n"
                    f"    sha256: {digest}\n"
                    "    role: primary\n",
                )
                + "# Activities\n",
                encoding="utf-8",
            )
            (root / "wiki/index.md").write_text(
                "# Index\n\n[[fundamentos/actividades]]\n", encoding="utf-8"
            )

            result = run_cli("validate", str(root), "--format", "json")
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            payload = json.loads(result.stdout)
            self.assertEqual(payload["status"], "valid")
            self.assertEqual(payload["provenance"]["status"], "current")

    def test_provenance_rejects_unsafe_hierarchical_slug(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            digest = add_source(root, "fundamentos/actividades", b"activities")
            for slug in (
                "/absolute",
                "fundamentos//actividades",
                "fundamentos/../actividades",
                "fundamentos/./actividades",
                "Fundamentos/actividades",
                "fundamentos/actividades_",
            ):
                (root / "wiki/pages/policy.md").write_text(
                    page_frontmatter(
                        "policy",
                        "sources:\n"
                        f"  - slug: {slug}\n"
                        "    source: raw/sources/fundamentos/actividades/source.txt\n"
                        "    extracted: raw/sources/fundamentos/actividades/extracted.md\n"
                        f"    sha256: {digest}\n"
                        "    role: primary\n",
                    )
                    + "# Policy\n",
                    encoding="utf-8",
                )

                result = run_cli("validate", str(root), "--format", "json")
                self.assertEqual(result.returncode, 1, slug)
                codes = {
                    item["details"].get("code")
                    for item in json.loads(result.stdout)["findings"]
                }
                self.assertIn("PROV-SLUG", codes, slug)

    def test_provenance_rejects_unsafe_paths(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            digest = add_source(root, "fundamentos/actividades", b"activities")
            unsafe_paths = (
                "/absolute/source.txt",
                "raw/sources/fundamentos/../actividades/source.txt",
                "raw/sources/fundamentos//actividades/source.txt",
                r"raw\sources\fundamentos\actividades\source.txt",
            )
            for source_path in unsafe_paths:
                (root / "wiki/pages/policy.md").write_text(
                    page_frontmatter(
                        "policy",
                        "sources:\n"
                        "  - slug: fundamentos/actividades\n"
                        f"    source: {source_path}\n"
                        "    extracted: raw/sources/fundamentos/actividades/extracted.md\n"
                        f"    sha256: {digest}\n"
                        "    role: primary\n",
                    )
                    + "# Policy\n",
                    encoding="utf-8",
                )
                result = run_cli("validate", str(root), "--format", "json")
                self.assertEqual(result.returncode, 1, source_path)
                codes = {
                    item["details"].get("code")
                    for item in json.loads(result.stdout)["findings"]
                }
                self.assertIn("PROV-SOURCE-PATH", codes, source_path)

    def test_legacy_page_is_valid_but_reported_as_migrable(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            digest = add_source(root, "policy", b"policy")
            (root / "wiki/pages/policy.md").write_text(
                page_frontmatter("policy", "")
                .replace(
                    "tags: []\n",
                    "source: raw/sources/policy/source.txt\n"
                    "extracted: raw/sources/policy/extracted.md\n"
                    f"sha256: {digest}\n"
                    "tags: []\n",
                )
                + "# Policy\n",
                encoding="utf-8",
            )
            result = run_cli("validate", str(root), "--format", "json")
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            payload = json.loads(result.stdout)
            self.assertEqual(payload["provenance"]["status"], "migrable")
            self.assertEqual(payload["provenance"]["legacy_pages"], ["wiki/pages/policy.md"])

    def test_migration_defaults_to_check_and_writes_only_with_write(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            digest = add_source(root, "policy", b"policy")
            page = root / "wiki/pages/policy.md"
            page.write_text(
                page_frontmatter("policy", "")
                .replace(
                    "tags: []\n",
                    "source: raw/sources/policy/source.txt\n"
                    "extracted: raw/sources/policy/extracted.md\n"
                    f"sha256: {digest}\n"
                    "tags: []\n",
                )
                + "# Policy\n",
                encoding="utf-8",
            )
            before = page.read_text(encoding="utf-8")

            result = run_cli("migrate-provenance", str(root), "--format", "json")
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            payload = json.loads(result.stdout)
            self.assertEqual(payload["status"], "needs-migration")
            self.assertEqual(payload["affected_pages"], ["wiki/pages/policy.md"])
            self.assertEqual(page.read_text(encoding="utf-8"), before)
            self.assertIn("sources:", payload["changes"][0]["diff"])
            self.assertNotIn("before", payload["changes"][0])
            self.assertNotIn("after", payload["changes"][0])

            result = run_cli(
                "migrate-provenance", str(root), "--write", "--format", "json"
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertEqual(json.loads(result.stdout)["status"], "written")
            migrated = page.read_text(encoding="utf-8")
            self.assertIn("sources:\n", migrated)
            self.assertNotIn("\nsource: raw/sources/policy/source.txt\n", migrated)
            self.assertIn(f"sha256: {digest}", migrated)

    def test_mixed_provenance_is_blocked_without_writing(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            create_vault(root)
            digest = add_source(root, "policy", b"policy")
            page = root / "wiki/pages/policy.md"
            page.write_text(
                page_frontmatter(
                    "policy",
                    "sources: []\n"
                    "source: raw/sources/policy/source.txt\n"
                    "extracted: raw/sources/policy/extracted.md\n"
                    f"sha256: {digest}\n",
                )
                + "# Policy\n",
                encoding="utf-8",
            )
            before = page.read_text(encoding="utf-8")
            result = run_cli("migrate-provenance", str(root), "--write", "--format", "json")
            self.assertEqual(result.returncode, 1)
            self.assertEqual(json.loads(result.stdout)["status"], "blocked")
            self.assertEqual(page.read_text(encoding="utf-8"), before)


if __name__ == "__main__":
    unittest.main()
