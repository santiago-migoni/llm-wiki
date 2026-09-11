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
python3 <plugin-root>/scripts/llm-wiki hashes <vault-root> --revert-to <sha256>
python3 <plugin-root>/scripts/llm-wiki links <vault-root>
python3 <plugin-root>/scripts/llm-wiki validate <vault-root>
python3 <plugin-root>/scripts/llm-wiki migrate-provenance <vault-root> --check
~~~

The inventory, links, hashes, and validate commands are read-only. `validate`
also checks every extraction header and returns each source's format, status, and
processed/expected coverage. Provenance
migration is read-only by default; pass `--write` explicitly to apply the
reviewed page-frontmatter changes. Migration never creates a Git commit and
prints a unified diff. Add `--format json` for a versioned machine-readable
result. Inventory JSON includes stable structural warning codes. Link reports
also identify missing headings and canonical pages absent from `wiki/index.md`.
Validation also checks query-visible contradiction records, including page/log
cross-references and reviewed resolution criteria.

`hashes --input` computes the pending file's SHA-256 and classifies it as a new
file, current duplicate, historical duplicate, or a current-and-historical
duplicate. `--include-history` can be
slower because it reads source blobs from Git. `--revert-to` explicitly records
a requested restoration target and never changes files; its target is classified
as current or historical when found.
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
