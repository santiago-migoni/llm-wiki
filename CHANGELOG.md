# Changelog

## 2.0.0

Rewrite from scratch, per `.specs/001-llm-wiki-plugin`. An audit of 1.x found meaningful redundancy: a bundled search command that was a substring `grep` reimplemented in Python, graph tooling (hub ranking, orphan export, DOT output) that duplicated Obsidian's own graph view, and a memory model (retrieval-hook frontmatter, budget-aware progressive loading) that added instruction weight without a demonstrated need. This release removes all of it.

**Removed:**

- Five skills: `wiki-recall`, `wiki-remember`, `wiki-research`, `wiki-digest`, `wiki-stats`. Each stays available to reconsider through the backlog if real use demands it — none is gone because it was a bad idea, only because none had shipped evidence it was needed.
- `scripts/wikigraph.py` and its test suite — 400 lines of Python, and everything it did that survives is now a documented `grep`/`find`/`git` one-liner inside a skill.
- The `summary:` retrieval-hook frontmatter field and the progressive-loading protocol built around it.

**Added:**

- `references/vault-protocol.md` — the contract all four remaining skills load first: explicit vault resolution, schema authority, file-tool/shell path lanes, the git commit procedure, and the write procedure every operation ends with. Four skills sharing one written contract, instead of the same rules restated (and drifting) four times.
- Explicit support for running identically in Claude Code and in Cowork, including the write-without-committing degradation when `git` is unavailable, and never assuming the working directory is the vault.
- A documented scale ceiling (roughly 100 sources, a few hundred pages) that `wiki-lint` reports explicitly when crossed, rather than the index silently getting less useful.

**Unchanged:** the three-layer architecture (raw sources, wiki, schema), schema authority over every skill, and the core ingest/query/lint operations from `docs/llm-wiki.md`.

## 1.0.1

Marketplace metadata release.

## 1.0.0

Initial release: nine skills (`wiki-init`, `wiki-ingest`, `wiki-query`, `wiki-lint`, `wiki-recall`, `wiki-stats`, `wiki-remember`, `wiki-research`, `wiki-digest`) plus `scripts/wikigraph.py`, a stdlib-only Python tool for wikilink-graph analysis.
