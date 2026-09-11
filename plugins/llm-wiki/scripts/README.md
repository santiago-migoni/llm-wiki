# Deterministic vault tools

This directory contains the dependency-free Python CLI used for deterministic LLM Wiki checks.

## Runtime

- Python 3.10 or later.
- Python standard library only.
- No PyYAML dependency.

Run the CLI from any working directory:

~~~bash
python3 <plugin-root>/scripts/llm-wiki inventory <vault-root>
python3 <plugin-root>/scripts/llm-wiki hashes <vault-root> --input <pending-file> --include-history
python3 <plugin-root>/scripts/llm-wiki links <vault-root>
python3 <plugin-root>/scripts/llm-wiki validate <vault-root>
python3 <plugin-root>/scripts/llm-wiki migrate-provenance <vault-root> --check
~~~

The inventory, links, hashes, and validate commands are read-only. Provenance
migration is read-only by default; pass `--write` explicitly to apply the
reviewed page-frontmatter changes. Migration never creates a Git commit and
prints a unified diff. Add `--format json` for a versioned machine-readable
result.

`hashes --input` computes the pending file's SHA-256 and reports matching current or historical source locations. `--include-history` can be slower because it reads source blobs from Git.
`migrate-provenance` converts legacy page-level `source`, `extracted`, and
`sha256` fields into one structured `sources` entry per page. It blocks mixed
formats and supplied stale hashes instead of repairing them silently.

## YAML subset

The CLI does not attempt to implement general YAML. It accepts the frontmatter subset used by LLM Wiki:

- a mapping at the document root;
- keys made from letters, numbers, `_`, and `-`;
- strings, integers, booleans, and null;
- inline scalar lists such as `tags: [research, policy]`;
- block lists of flat mappings, including the proposed `sources` records;
- indentation in multiples of two spaces.

It rejects aliases, anchors, tags, flow mappings, block scalars, tabs, duplicate keys, and deeper containers inside a list item. Unsupported syntax fails closed with an actionable error.

This is a compatibility boundary, not a replacement for a YAML parser. If future schemas require general YAML, add a pinned parser dependency rather than extending this module into an incomplete general-purpose implementation.
