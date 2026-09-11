# Fase 9 — Validación de vault real en Codex

## Test run

- Date: 2026-09-11
- Host: Codex local
- Plugin under test: package `1.0.0` installed from the local marketplace before the `1.1.0` metadata release
- Vault location: `/Users/<redacted>/Documents/<redacted>`
- Scenarios: `I-001`, `G-001`, `G-007`, `M-001`

## Results

| ID | Status | Evidence |
| --- | --- | --- |
| I-001 | passed | A real health-personal vault was initialized with `AGENTS.md`, the Git-first tree, and the initial `schema: initial wiki scaffold` commit `fddc922`. |
| G-001 | passed | A supervised explicit batch ingested a Markdown source, created its stable slug, extraction, canonical page, indexes, log entry, and commit. |
| G-007 | passed | Two PDF sources were inspected with text extraction and visual review; one declared `partial` coverage because ECG graphics remained in the original, and one declared `complete` coverage with a non-material converter warning. |
| M-001 | passed | Three source records and three canonical pages were validated with matching provenance, resolved links, no contradictions, and an empty inbox. Three independent `ingest:` commits were created: `be7b1ff`, `5b90994`, and `a543fb5`. |

## Limitations

- This record covers the real-vault initialization and ingestion flow. The deterministic functional
  workflow covers historical duplicate handling, query-visible provenance, contradictions, links,
  and scale checks.
- ChatGPT Work was not available in this run and remains explicitly experimental for release
  `1.1.0`.
- The original source documents are not copied into this result record.
