import sys
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[2] / "scripts"
sys.path.insert(0, str(SCRIPTS))

from llm_wiki.frontmatter import FrontmatterError, parse_document  # noqa: E402


class FrontmatterTests(unittest.TestCase):
    def test_current_page_shape(self):
        parsed = parse_document(
            """---
type: knowledge
slug: policy
updated: 2026-09-11
tags: [security, governance]
status: current
---
# Policy
"""
        )
        self.assertEqual(parsed.metadata["type"], "knowledge")
        self.assertEqual(parsed.metadata["tags"], ["security", "governance"])
        self.assertIn("# Policy", parsed.body)

    def test_multisource_shape(self):
        parsed = parse_document(
            """---
type: knowledge
slug: policy
sources:
  - slug: source-one
    source: raw/sources/source-one/source.pdf
    role: primary
  - slug: source-two
    source: raw/sources/source-two/source.docx
    role: supporting
---
body
"""
        )
        self.assertEqual(len(parsed.metadata["sources"]), 2)
        self.assertEqual(parsed.metadata["sources"][1]["role"], "supporting")

    def test_rejects_duplicate_key(self):
        with self.assertRaisesRegex(FrontmatterError, "duplicate key"):
            parse_document("---\nslug: one\nslug: two\n---\n")

    def test_rejects_general_yaml_features(self):
        with self.assertRaisesRegex(FrontmatterError, "unsupported YAML feature"):
            parse_document("---\nvalue: &anchor thing\n---\n")

    def test_rejects_odd_indentation(self):
        with self.assertRaisesRegex(FrontmatterError, "multiples of two"):
            parse_document("---\nsources:\n - one\n---\n")


if __name__ == "__main__":
    unittest.main()
