import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPTS))

from llm_wiki.contradictions import build_contradiction_report  # noqa: E402


CLI = SCRIPTS / "llm-wiki"


class ContradictionContractTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.root = Path(self.directory.name)
        for directory in ("raw/inbox", "raw/sources", "wiki/pages"):
            (self.root / directory).mkdir(parents=True)
        (self.root / "AGENTS.md").write_text("# Schema\n", encoding="utf-8")
        (self.root / "wiki/index.md").write_text("# Index\n\n[[policy]]\n", encoding="utf-8")
        (self.root / "wiki/overview.md").write_text("# Overview\n", encoding="utf-8")

    def tearDown(self):
        self.directory.cleanup()

    def page(self, *, status="unresolved", resolution=True):
        resolution_fields = (
            "- Resolution: The signed policy has precedence.\n"
            "- Criterion: The effective-date clause is the governing authority.\n"
            if resolution
            else ""
        )
        return (
            "---\n"
            "type: knowledge\n"
            "slug: policy\n"
            "created: 2026-09-11\n"
            "updated: 2026-09-11\n"
            "sources: []\n"
            "tags: []\n"
            "status: current\n"
            "---\n"
            "# Policy\n\n"
            "## Contradictions\n\n"
            "### review-frequency\n\n"
            "- Claim A: Review occurs annually.\n"
            "- Evidence A: `corporate-policy`, section 4.\n"
            "- Claim B: Review occurs quarterly.\n"
            "- Evidence B: `audit-report`, page 14.\n"
            f"- Status: {status}\n"
            "- Impact: Changes the compliance schedule.\n"
            + resolution_fields
        )

    def log(self, *, page=True, status="unresolved"):
        page_field = "- Page: [[policy]]\n" if page else ""
        return (
            "# Log\n\n"
            "## [2026-09-11] contradiction | review-frequency\n"
            + page_field
            + f"- Status: {status}\n"
        )

    def report(self):
        return build_contradiction_report(self.root)

    def test_valid_unresolved_contradiction_is_linked_to_log(self):
        (self.root / "wiki/pages/policy.md").write_text(self.page(), encoding="utf-8")
        (self.root / "wiki/log.md").write_text(self.log(), encoding="utf-8")

        report = self.report()
        self.assertEqual(report["issues"], [])
        self.assertEqual(report["counts"]["unresolved"], 1)
        self.assertEqual(report["counts"]["page_entries"], 1)
        self.assertEqual(report["counts"]["log_entries"], 1)

    def test_log_only_contradiction_is_reported(self):
        (self.root / "wiki/pages/policy.md").write_text(
            self.page().split("## Contradictions")[0], encoding="utf-8"
        )
        (self.root / "wiki/log.md").write_text(self.log(), encoding="utf-8")

        codes = {issue["code"] for issue in self.report()["issues"]}
        self.assertIn("CONTRA-LOG-ONLY", codes)

    def test_page_contradiction_requires_a_log_event(self):
        (self.root / "wiki/pages/policy.md").write_text(self.page(), encoding="utf-8")
        (self.root / "wiki/log.md").write_text("# Log\n", encoding="utf-8")

        codes = {issue["code"] for issue in self.report()["issues"]}
        self.assertIn("CONTRA-LOG-MISSING", codes)

    def test_resolution_requires_both_evidence_and_criterion(self):
        (self.root / "wiki/pages/policy.md").write_text(
            self.page(status="resolved", resolution=False), encoding="utf-8"
        )
        (self.root / "wiki/log.md").write_text(
            self.log(status="resolved"), encoding="utf-8"
        )
        codes = {issue["code"] for issue in self.report()["issues"]}
        self.assertIn("CONTRA-RESOLUTION-MISSING", codes)

        (self.root / "wiki/pages/policy.md").write_text(
            self.page(status="resolved", resolution=True), encoding="utf-8"
        )
        self.assertEqual(self.report()["issues"], [])

    def test_page_and_log_status_mismatch_is_reported(self):
        (self.root / "wiki/pages/policy.md").write_text(
            self.page(status="unresolved"), encoding="utf-8"
        )
        (self.root / "wiki/log.md").write_text(
            self.log(status="resolved"), encoding="utf-8"
        )
        codes = {issue["code"] for issue in self.report()["issues"]}
        self.assertIn("CONTRA-STATUS-MISMATCH", codes)

    def test_log_event_without_page_is_reported(self):
        (self.root / "wiki/log.md").write_text(
            self.log(page=False), encoding="utf-8"
        )
        codes = {issue["code"] for issue in self.report()["issues"]}
        self.assertIn("CONTRA-LOG-PAGE-MISSING", codes)

    def test_validate_exposes_contradiction_report(self):
        (self.root / "wiki/pages/policy.md").write_text(
            self.page().split("## Contradictions")[0], encoding="utf-8"
        )
        (self.root / "wiki/log.md").write_text(self.log(), encoding="utf-8")
        result = subprocess.run(
            [sys.executable, str(CLI), "validate", str(self.root), "--format", "json"],
            check=False,
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["contradictions"]["counts"]["log_entries"], 1)
        self.assertTrue(
            any(
                item["details"].get("code") == "CONTRA-LOG-ONLY"
                for item in payload["findings"]
            )
        )


if __name__ == "__main__":
    unittest.main()
