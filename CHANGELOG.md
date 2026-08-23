# Changelog

## 0.1.0

First release of this plugin in this repository.

It succeeds an earlier nine-skill version published separately as `llm-wiki` 1.x, which is not part of this repository's history. That version is superseded, not continued: this one was written from scratch against [docs/llm-wiki.md](docs/llm-wiki.md) after an audit of the earlier one found meaningful redundancy — a bundled search command that was a substring `grep` reimplemented in Python, graph tooling (hub ranking, orphan export, DOT output) that duplicated Obsidian's own graph view, and a memory model (retrieval-hook frontmatter, budget-aware progressive loading) that added instruction weight without a demonstrated need.

The version resets to `0.1.0` deliberately. Nothing here has been used in sustained real work yet, and `0.x` says so; a `2.0.0` would have claimed a maturity this does not have.

**What it ships:**

- Four skills — `wiki-init`, `wiki-ingest`, `wiki-query`, `wiki-lint` — covering the three operations the source pattern names, plus vault scaffolding.
- `references/vault-protocol.md`, the contract all four skills load first: explicit vault resolution, schema authority, file-tool/shell path lanes, the git commit procedure, confirmation gates on destructive actions, index-first retrieval, and the write procedure every operation ends with. One written contract instead of the same rules restated — and drifting — across four files.
- Optional document conversion for `.pdf`, `.docx`, `.xlsx`, `.pptx`, `.html`, and images, each probing for its converter and degrading with a stated alternative rather than failing.
- Support for running identically in Claude Code and in Cowork, including never assuming the working directory is the vault, and degrading to write-without-committing when `git` is unavailable.
- A documented scale ceiling (roughly 100 sources, a few hundred pages) that `wiki-lint` reports explicitly when crossed, instead of the index quietly getting less useful.

**Deliberately not shipped**, each because nothing yet demonstrated a need: a search engine or any search tooling, graph tooling of any kind, a memory model, and the five other skills the earlier version carried (`wiki-recall`, `wiki-remember`, `wiki-research`, `wiki-digest`, `wiki-stats`). There is no executable code in this repository at all — everything the earlier graph script did that survives is a documented `grep`/`find`/`git` one-liner inside a skill.
