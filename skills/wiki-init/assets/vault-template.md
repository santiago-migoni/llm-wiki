# Vault schema template

Write this as `AGENTS.md` at the vault root for a Codex-first vault, or as `CLAUDE.md` when the user explicitly requests a Claude-first vault. Replace every `{{...}}` placeholder. Drop or rename category directories to match the scaffold actually created.

---

# {{Vault Name}} Wiki — Schema

This vault is an LLM-maintained knowledge base. The human ({{user name}}) curates sources and asks questions; the LLM writes and maintains everything in `wiki/`.

## Scope

{{One paragraph: what this wiki tracks, tailored to the user's stated purpose.}}

## Layout

```text
raw/
  inbox/              # User drop zone for documents awaiting ingest.
  archive/            # Immutable originals, grouped by logical document and revision.
  extracted/          # Complete, regenerable markdown conversions by document/revision.
  assets/             # Images and embedded assets by document/revision.
  catalog.md          # One routing row per logical document and current revision.
wiki/
  index.md            # Catalog of all wiki pages, by category. Updated on every write.
  log.md              # Append-only chronological record of all wiki activity.
  overview.md         # Evolving top-level synthesis of the whole wiki.
  sources/            # One summary page per archived revision.
  knowledge/          # Stable living documents, updated over time.
  syntheses/          # Filed answers: comparisons, analyses, decisions.
  {{category dirs, one line each, e.g.:}}
  people/             # Entity pages for individuals.
  concepts/           # Ideas, skills, methods, trends.
AGENTS.md              # This file. Co-evolves with the wiki via lint.
```

## Page conventions

- Filenames: kebab-case, e.g. `wiki/people/jane-doe.md`.
- Links: Obsidian wikilinks, e.g. `[[jane-doe]]`. Link liberally — a link to a page that does not exist yet marks it as worth creating.
- Every wiki page starts with YAML frontmatter:

```yaml
---
type: source | knowledge | synthesis | {{category singular, e.g. person}}
created: YYYY-MM-DD
updated: YYYY-MM-DD
tags: []
---
```

- Source pages additionally get `document-id:`, `revision-id:`, `source: raw/archive/<document-id>/<revision-id>/original.<ext>`, `sha256:`, `extracted: raw/extracted/<document-id>/<revision-id>.md`, and `source-date:` when known.
- Stable living pages in `wiki/knowledge/` get `type: knowledge`, `document-id:`, and `current-revision:`; they retain links to current and superseded source pages.
- `wiki/index.md` carries a one-line hook per page (`[[page]] — when/why it matters`). It is the retrieval layer: find the right page from the index before opening it.
- Keep pages focused: one entity or concept per page. Split a page when a section outgrows its host.
- When new information contradicts an existing claim, do not silently overwrite it. Note the contradiction on the page and flag it in the log entry.

## Document archive and context

- New documents enter through `raw/inbox/`.
- Ingest assigns a stable `document-id` and a monotonic `revision-id`, records a SHA-256 checksum, moves the original without changing its bytes into `raw/archive/<document-id>/<revision-id>/`, and writes a complete extraction to `raw/extracted/<document-id>/<revision-id>.md`.
- Each revision directory contains a small `manifest.md` recording the document ID, revision ID, original filename, checksum, ingest date, extraction path, and supersession links.
- `wiki/knowledge/<document-id>.md` is the stable living document. It may be updated as new revisions arrive, while revision summaries under `wiki/sources/` preserve the source-specific record.
- `raw/catalog.md` stays compact: one row per logical document with `document-id`, current `revision-id`, status, topics, stable page, current source page, and extraction path. Do not put document contents in the catalog.
- Targeted queries may use the index and relevant current extracts. When the user asks for full or exhaustive context, inventory the catalog, verify current extraction coverage, read every relevant extraction in bounded passes, and report any gap. Never claim complete context if a source is pending, missing, or unreadable.

## Workflows

### Ingest

A source lands in `raw/inbox/`. Read it fully, decide whether it is a new `document-id` or a revision, discuss key takeaways, archive the original, preserve a complete extraction, write a revision summary in `wiki/sources/`, update the stable living document in `wiki/knowledge/`, update every page it touches, update the catalog, index, log, and commit (`ingest:`). A source contradicting an existing claim gets the contradiction noted, not a silent overwrite.

### Query

Read `wiki/index.md` to find relevant pages, then read them. Answer with citations to the wiki page, `document-id`, `revision-id`, and extraction or page marker where applicable. A substantive answer can be filed to `wiki/syntheses/` — index, log, and commit it (`query:`).

### Lint

On request, check for contradictions, stale claims, orphan candidates, broken links, archive integrity, and gaps. Report findings; apply only what the user approves; log and commit (`lint:`). When a convention in this schema stops fitting the vault, propose amending this file and commit that too (`schema:`).

## Log format

Entries in `wiki/log.md`, newest last, prefix-parseable:

```text
## [YYYY-MM-DD] ingest | Source Title
## [YYYY-MM-DD] query | Question asked
## [YYYY-MM-DD] lint | Scope
## [YYYY-MM-DD] schema | What changed
```

## Git

The vault is a git repository. Commit after every write, with a labelled message (`ingest:`, `query:`, `lint:`, `schema:`). **Never `git add -A` or `git add .`** — stage only the exact paths the operation touched. A source in `raw/archive/` is committed for the first time by the ingest that files it; until then, `git status --short raw/inbox/` is the inbox.

## Style

Write for future retrieval: dense, factual, specific, no filler. Convert relative dates to absolute. This schema is negotiable — when a convention isn't working, propose a change through lint and update this file.
