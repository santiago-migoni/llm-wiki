# llm-wiki

A universal Codex plugin that turns an archive of documents into a knowledge base the LLM writes and maintains: you curate sources and ask questions, the LLM does the summarizing, cross-referencing, filing, and bookkeeping.

It implements the pattern described in [docs/llm-wiki.md](docs/llm-wiki.md) — the idea file this plugin is built from, credited here rather than claimed as original. A vault has four layers: an inbox for new documents, an immutable archive of revisions, complete extracted text for reading, and an LLM-owned wiki. A per-vault schema file (`AGENTS.md`, or `CLAUDE.md` for a Claude-first vault) governs everything else and the skills defer to it.

## Skills

| Skill | Purpose |
|---|---|
| `wiki-init` | Turn a folder into a working vault: archive layout, a tailored schema file, index/log/overview, git repo. |
| `wiki-ingest` | Move a document from `raw/inbox/` into the immutable archive as a new revision, preserve a complete extraction, update its stable living page, and update every wiki page it touches. Converts `.docx`/`.xlsx`/`.pdf` first when a converter is available; batch mode for several sources at once. |
| `wiki-query` | Answer questions from the wiki with citations; file substantive answers back as syntheses. |
| `wiki-lint` | Health check: archive integrity, orphan candidates, broken links, contradictions, stale claims, index drift, pending sources — reports findings, fixes only what's approved. |

## Codex and ChatGPT Work

This repository is a universal Agent Plugin with its manifest at `.codex-plugin/plugin.json`. Install it through the Codex plugin directory or a local marketplace. The same installed plugin can be invoked in Codex and ChatGPT Work; start a new conversation after installation so the host loads its skills.

The plugin itself is file-first and has no external service. In Codex, the vault can be a local repository. In ChatGPT Work, the archive must be available to that conversation through its workspace files or a connected source; installing the plugin alone does not upload or persist documents.

## Memory model

There is no search engine, no graph tooling, and no retrieval-hook frontmatter. `wiki/index.md` is the navigation layer: one line per wiki page, read before any page is opened. The archive remains the source of truth, while the wiki is the fast, structured memory layer. This holds at roughly 100 sources and a few hundred pages — `wiki-lint` says so explicitly when a vault crosses that ceiling, rather than silently degrading.

## Document lifecycle

New documents go into `raw/inbox/`. Ingest assigns or reuses a logical `document-id`, creates the next `revision-id`, records a SHA-256 checksum, moves the original to `raw/archive/<document-id>/<revision-id>/`, and writes a complete normalized extraction to `raw/extracted/<document-id>/<revision-id>.md`. The original revision is never edited. The stable page in `wiki/knowledge/` is updated over time; revision summaries in `wiki/sources/` remain individually traceable.

```text
raw/inbox/                         # Drop zone for documents awaiting ingest
raw/archive/<document-id>/<revision-id>/      # Immutable original + manifest
raw/extracted/<document-id>/<revision-id>.md  # Complete text derived from revision
raw/assets/<document-id>/<revision-id>/        # Images and embedded assets
raw/catalog.md                               # Current revision and routing map
wiki/sources/<document-id>--<revision-id>.md # Revision summary and citations
wiki/knowledge/<document-id>.md              # Stable living document
wiki/                                         # Cross-source pages and filed syntheses
```

For ordinary questions, the agent uses the index and reads the relevant current extracts. When asked for *full context*, *exhaustive analysis*, or an answer about *all documents*, it inventories the catalog, checks for pending or missing current revisions, reads every current extraction in bounded passes, and reports its coverage. Historical revisions are loaded when the question concerns changes or provenance. “Absolute context” is therefore an auditable workflow over the complete archive, not a promise that an arbitrarily large corpus fits in one model window. See [docs/archive-model.md](docs/archive-model.md).

## Scale and hosts

Runs with the same workflow in Codex and ChatGPT Work. Every skill resolves the vault as an absolute path rather than assuming the working directory is the vault, writes pages with file tools rather than shell redirection, and probes for `git` before committing — degrading to "written, not committed" when it's unavailable instead of failing outright.

## Conventions assumed as defaults

- `raw/inbox/` is the staging area; `raw/archive/` holds immutable revisions; conversions go to versioned paths under `raw/extracted/`.
- Wiki pages use YAML frontmatter, kebab-case filenames, and liberal wikilinks.
- `raw/catalog.md` catalogs logical documents and current revisions; `wiki/index.md` catalogs knowledge pages.
- `wiki/log.md` is an append-only activity log.
- The vault is a git repo; every write is committed atomically, with only the paths it touched.

All of these yield to whatever the vault's own schema file specifies — see `wiki-init`'s template at `skills/wiki-init/assets/vault-template.md`.

## Zero dependencies

No scripts, no runtime, and no install step beyond the plugin itself. Everything the earlier version of this plugin delegated to a bundled Python tool is now either a documented `grep`/`find`/`git` one-liner, inside a skill, or isn't done at all — see `docs/llm-wiki.md`'s "Optional: CLI tools" section for the deliberate boundary.
