# Changelog

## Unreleased

No unreleased changes.

## 1.2.0 — 2026-09-12

- Added hierarchical source slugs and nested canonical pages while preserving existing flat slugs.
- Made source inventory, provenance, current and historical hashes, duplicate detection, validation,
  and exact wikilink resolution work with complete hierarchical paths.
- Preserved namespace segments named `assets` instead of confusing them with record attachments.
- Moved the plugin's unit, fixture, scale, and functional tests into `plugins/llm-wiki/tests/`.

## 1.1.0 — 2026-09-11

- Clarified that originals are preserved byte-for-byte, updates replace only the current working
  tree through Git, and extraction quality is reported by format-specific coverage.
- Added the v1.1 quickstart, populated-vault example, recovery and migration guides, and an
  evidence-based host compatibility matrix.
- Added the portable Agent Plugins manifest at `plugins/llm-wiki/plugin.json`.
- Kept `.codex-plugin/plugin.json` as a compatible Codex fallback and aligned shared metadata.
- Added the repository MIT license and normalized the marketplace display name to `LLM Wiki`.
- Validated the plugin in a real Codex vault with initialization and a supervised batch ingest of
  Markdown and PDF sources, including format-specific coverage and Git commits.
- ChatGPT Work remains explicitly experimental until a workspace-backed smoke test is available.

## 1.0.0

First official stable release of the LLM Wiki Codex marketplace plugin.

**What changed:**

- Added the plugin's minimal, transparent logo at `plugins/llm-wiki/assets/logo.png`.
- Declared the logo in the plugin interface manifest and promoted the package version to `1.0.0`.
- Stabilized the Git-first vault workflow for document ingestion, canonical wiki pages, citations, and historical context.
- Published the repo-local marketplace package for installation in Codex and use with ChatGPT Work.

## 0.2.0

Codex and ChatGPT Work migration with a Git-first document model.

**What changed:**

- Replaced the Claude-specific manifests with a universal .codex-plugin/plugin.json manifest.
- Packaged the plugin in a repo-local Codex marketplace at .agents/plugins/marketplace.json.
- Documented use in Codex and ChatGPT Work, including the boundary that the plugin does not provide external document storage.
- Defined the current source lifecycle: raw/inbox/ → raw/sources/<slug>/source.<ext> + extracted.md → canonical wiki page.
- Established one stable slug per logical document and Git history as the versioning mechanism.
- Added SHA-256 duplicate detection, format-specific extraction coverage requirements, historical
  Git queries, and auditable context coverage reporting.
- Aligned the vault protocol, initialization template, ingestion, query, and lint skills with the simplified structure.
- Added the source-record contract for classifying new documents, updates, exact duplicates, ambiguous identities, local-edit protection, extraction metadata, and ingest reports.
- Added auditable query modes for targeted, current-corpus, and historical context, with coverage ledgers, source citations, Git provenance, and durable synthesis filing.
- Added report-first wiki linting for current source structure, extraction freshness, index drift, links, duplicates, contradictions, metadata, legacy layouts, and scale.
- Added a read-only vault scale assessor with explicit GREEN, WATCH, and DERIVED-SEARCH-CANDIDATE thresholds before any derived search layer is considered.
- Added a hybrid test strategy, an empty-vault fixture, a deterministic plugin contract check, and Codex/ChatGPT Work compatibility smoke-test guidance.

## 0.1.0

First release of this plugin in this repository.

It succeeds an earlier nine-skill version published separately as llm-wiki 1.x, which is not part of this repository's history. That version is superseded, not continued: this one was written from scratch against docs/llm-wiki.md after an audit of the earlier one found meaningful redundancy.

The version resets to 0.1.0 deliberately. Nothing here has been used in sustained real work yet, and 0.x says so; a 2.0.0 would have claimed a maturity this does not have.

**What it ships:**

- Four skills — wiki-init, wiki-ingest, wiki-query, wiki-lint — covering vault scaffolding, document ingestion, wiki queries, and quality checks.
- references/vault-protocol.md, the shared contract for vault resolution, schema authority, source handling, Git, safety, retrieval, and context coverage.
- Optional document conversion for PDF, DOCX, XLSX, PPTX, HTML, images, and other supported formats.
- Portable operation in Codex and ChatGPT Work, subject to the host's file access.
- A documented scale ceiling for the initial index-first approach.

**Deliberately not shipped:** a mandatory search engine, vector database, graph tooling, remote storage, or external service. The current state is plain files plus Git; derived search layers can be added when real scale requires them.
