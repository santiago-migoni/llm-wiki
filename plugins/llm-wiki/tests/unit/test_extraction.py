import sys
import tempfile
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[2] / "scripts"
sys.path.insert(0, str(SCRIPTS))

from llm_wiki.extraction import inspect_extraction  # noqa: E402


class ExtractionContractTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.root = Path(self.directory.name)
        self.source = self.root / "raw/sources/policy/source.pdf"
        self.extraction = self.root / "raw/sources/policy/extracted.md"
        self.source.parent.mkdir(parents=True)
        self.extraction.write_text("---\n---\n", encoding="utf-8")
        self.source.write_bytes(b"source")

    def tearDown(self):
        self.directory.cleanup()

    def metadata(self, **overrides):
        metadata = {
            "type": "extraction",
            "slug": "policy",
            "source": "raw/sources/policy/source.pdf",
            "original-filename": "policy.pdf",
            "sha256": "0" * 64,
            "extracted": "2026-09-11",
            "format": "pdf",
            "method": "pdftotext -layout",
            "status": "complete",
            "coverage": {"unit": "pages", "expected": 12, "processed": 12},
            "warnings": [],
        }
        metadata.update(overrides)
        return metadata

    def issues(self, metadata):
        return {
            issue["code"]
            for issue in inspect_extraction(
                self.root,
                self.extraction,
                self.source,
                "policy",
                metadata,
            )["issues"]
        }

    def test_complete_pdf_coverage_is_valid(self):
        report = inspect_extraction(
            self.root,
            self.extraction,
            self.source,
            "policy",
            self.metadata(),
        )
        self.assertEqual(report["issues"], [])
        self.assertEqual(report["coverage"]["processed"], 12)
        self.assertEqual(report["coverage"]["unit"], "pages")

    def test_format_units_match_supported_source_types(self):
        cases = (
            ("source.txt", "txt", "pages"),
            ("source.html", "html", "pages"),
            ("source.docx", "docx", "pages"),
            ("source.xlsx", "xlsx", "sheets"),
            ("source.csv", "csv", "rows"),
            ("source.pptx", "pptx", "slides"),
            ("source.png", "image", "regions"),
            ("source.mp3", "audio", "seconds"),
            ("source.mp4", "video", "seconds"),
        )
        for filename, format_name, unit in cases:
            source = self.root / "raw/sources/policy" / filename
            source.write_bytes(b"source")
            metadata = self.metadata(
                format=format_name,
                coverage={"unit": unit, "expected": 1, "processed": 1},
            )
            self.assertEqual(
                inspect_extraction(
                    self.root,
                    self.extraction,
                    source,
                    "policy",
                    metadata,
                )["issues"],
                [],
                filename,
            )

    def test_format_requires_its_declared_unit(self):
        metadata = self.metadata(
            format="xlsx",
            coverage={"unit": "rows", "expected": 2, "processed": 2},
        )
        codes = self.issues(metadata)
        self.assertIn("EXT-FORMAT", codes)
        self.assertIn("EXT-COVERAGE-UNIT", codes)

    def test_complete_requires_all_expected_units(self):
        metadata = self.metadata(
            coverage={"unit": "pages", "expected": 12, "processed": 11}
        )
        self.assertIn("EXT-STATUS-COVERAGE", self.issues(metadata))

    def test_partial_and_representative_need_explanation(self):
        partial = self.metadata(
            status="partial",
            coverage={"unit": "pages", "expected": 12, "processed": 12},
        )
        self.assertIn("EXT-WARNING-REQUIRED", self.issues(partial))

        representative = self.metadata(
            status="representative",
            coverage={"unit": "pages", "expected": 12, "processed": 12},
            warnings=["Only the executive summary was sampled."],
        )
        self.assertIn("EXT-STATUS-COVERAGE", self.issues(representative))

    def test_unsupported_must_have_zero_processed_units(self):
        metadata = self.metadata(
            status="unsupported",
            coverage={"unit": "pages", "expected": 1, "processed": 1},
            warnings=["No compatible reader was available."],
        )
        self.assertIn("EXT-STATUS-COVERAGE", self.issues(metadata))

    def test_unexpected_scalar_shapes_fail_closed(self):
        metadata = self.metadata(
            status=["complete"],
            coverage={"unit": ["pages"], "expected": 1, "processed": 1},
        )
        codes = self.issues(metadata)
        self.assertIn("EXT-STATUS", codes)
        self.assertIn("EXT-COVERAGE-UNIT", codes)


if __name__ == "__main__":
    unittest.main()
