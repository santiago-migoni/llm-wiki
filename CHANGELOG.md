# Changelog

## 0.2.0

Codex and ChatGPT Work migration with a Git-first document model.

**What changed:**

- Replaced the Claude-specific manifests with a universal .codex-plugin/plugin.json manifest.
- Documented use in Codex and ChatGPT Work, including the boundary that the plugin does not provide external document storage.
- Defined the current source lifecycle: raw/inbox/ → raw/sources/<slug>/source.<ext> + extracted.md → canonical wiki page.
- Established one stable slug per logical document and Git history as the versioning mechanism.
- Added SHA-256 duplicate detection, complete extraction requirements, historical Git queries, and exhaustive context coverage reporting.
- Aligned the vault protocol, initialization template, ingestion, query, and lint skills with the simplified structure.
- Added the source-record contract for classifying new documents, updates, exact duplicates, ambiguous identities, local-edit protection, extraction metadata, and ingest reports.
- Added auditable query modes for targeted, current-corpus, and historical context, with coverage ledgers, source citations, Git provenance, and durable synthesis filing.
- Added report-first wiki linting for current source structure, extraction freshness, index drift, links, duplicates, contradictions, metadata, legacy layouts, and scale.

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
