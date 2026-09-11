import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[1]
CLI = SCRIPTS / "llm-wiki"
FIXTURE = Path(__file__).resolve().parents[4] / "tests/fixtures/empty-vault"


def run_cli(*args):
    return subprocess.run(
        [sys.executable, str(CLI), *args],
        check=False,
        capture_output=True,
        text=True,
    )


class CliTests(unittest.TestCase):
    def test_empty_fixture_inventory(self):
        result = run_cli("inventory", str(FIXTURE), "--format", "json")
        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["schema_version"], 1)
        self.assertEqual(payload["counts"]["source_records"], 0)
        self.assertEqual(payload["counts"]["canonical_pages"], 0)

    def test_empty_fixture_is_valid(self):
        result = run_cli("validate", str(FIXTURE), "--format", "json")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual(json.loads(result.stdout)["status"], "valid")

    def test_multiple_source_files_are_invalid(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "raw/inbox").mkdir(parents=True)
            record = root / "raw/sources/policy"
            record.mkdir(parents=True)
            (root / "wiki/pages").mkdir(parents=True)
            (root / "AGENTS.md").write_text("# Schema\n", encoding="utf-8")
            (record / "source.pdf").write_bytes(b"one")
            (record / "source.docx").write_bytes(b"two")
            (record / "extracted.md").write_text(
                "---\ntype: extraction\nslug: policy\nsource: raw/sources/policy/source.pdf\n"
                "sha256: value\nstatus: complete\n---\ntext\n",
                encoding="utf-8",
            )
            result = run_cli("validate", str(root), "--format", "json")
            self.assertEqual(result.returncode, 1)
            messages = [item["message"] for item in json.loads(result.stdout)["findings"]]
            self.assertIn("Source record has multiple current source.* files", messages)

    def test_hashes_can_compare_an_input(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "raw/inbox").mkdir(parents=True)
            record = root / "raw/sources/policy"
            record.mkdir(parents=True)
            (root / "wiki/pages").mkdir(parents=True)
            (root / "AGENTS.md").write_text("# Schema\n", encoding="utf-8")
            source = record / "source.txt"
            source.write_bytes(b"same bytes")
            pending = root / "raw/inbox/copy.txt"
            pending.write_bytes(b"same bytes")
            result = run_cli(
                "hashes", str(root), "--input", str(pending), "--format", "json"
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            payload = json.loads(result.stdout)
            self.assertEqual(len(payload["matches"]), 1)
            self.assertEqual(payload["matches"][0]["slug"], "policy")

    def test_historical_hashes_work_inside_parent_repository(self):
        with tempfile.TemporaryDirectory() as directory:
            repository = Path(directory)
            vault = repository / "vault"
            (vault / "raw/inbox").mkdir(parents=True)
            record = vault / "raw/sources/policy"
            record.mkdir(parents=True)
            (vault / "wiki/pages").mkdir(parents=True)
            (vault / "AGENTS.md").write_text("# Schema\n", encoding="utf-8")
            source = record / "source.txt"
            source.write_bytes(b"version one")
            subprocess.run(["git", "init", "-q", str(repository)], check=True)
            subprocess.run(
                ["git", "-C", str(repository), "config", "user.name", "LLM Wiki Test"],
                check=True,
            )
            subprocess.run(
                ["git", "-C", str(repository), "config", "user.email", "test@example.invalid"],
                check=True,
            )
            subprocess.run(["git", "-C", str(repository), "add", "vault"], check=True)
            subprocess.run(
                ["git", "-C", str(repository), "commit", "-qm", "first"], check=True
            )
            source.write_bytes(b"version two")
            subprocess.run(["git", "-C", str(repository), "add", "vault"], check=True)
            subprocess.run(
                ["git", "-C", str(repository), "commit", "-qm", "second"], check=True
            )

            result = run_cli(
                "hashes", str(vault), "--include-history", "--format", "json"
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            historical = json.loads(result.stdout)["historical"]
            self.assertEqual(len(historical), 2)
            self.assertTrue(
                all(item["path"] == "raw/sources/policy/source.txt" for item in historical)
            )

    def test_required_path_wrong_type_is_reported_without_traceback(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "raw").mkdir()
            (root / "raw/inbox").mkdir()
            (root / "raw/sources").write_text("not a directory", encoding="utf-8")
            (root / "wiki/pages").mkdir(parents=True)
            (root / "AGENTS.md").write_text("# Schema\n", encoding="utf-8")
            result = run_cli("validate", str(root), "--format", "json")
            self.assertEqual(result.returncode, 1)
            self.assertNotIn("Traceback", result.stderr)
            messages = [item["message"] for item in json.loads(result.stdout)["findings"]]
            self.assertIn("Required path is not a directory", messages)


if __name__ == "__main__":
    unittest.main()
