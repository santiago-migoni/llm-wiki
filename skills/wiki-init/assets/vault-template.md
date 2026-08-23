# Vault schema template

Write this as `CLAUDE.md` at the vault root, replacing every `{{...}}` placeholder. Drop or rename category directories to match the scaffold actually created.

---

# {{Vault Name}} Wiki — Schema

This vault is an LLM-maintained knowledge base. The human ({{user name}}) curates sources and asks questions; the LLM writes and maintains everything in `wiki/`.

## Scope

{{One paragraph: what this wiki tracks, tailored to the user's stated purpose.}}

## Layout

```text
raw/                  # Immutable source documents. Never modify. raw/assets/ holds images;
                       # raw/extracted/ holds regenerable markdown conversions of non-markdown originals.
wiki/
  index.md            # Catalog of all wiki pages, by category. Updated on every write.
  log.md              # Append-only chronological record of all wiki activity.
  overview.md         # Evolving top-level synthesis of the whole wiki.
  sources/            # One summary page per ingested raw source.
  syntheses/          # Filed answers: comparisons, analyses, decisions.
  {{category dirs, one line each, e.g.:}}
  people/             # Entity pages for individuals.
  concepts/           # Ideas, skills, methods, trends.
CLAUDE.md              # This file. Co-evolves with the wiki via lint.
```

## Page conventions

- Filenames: kebab-case, e.g. `wiki/people/jane-doe.md`.
- Links: Obsidian wikilinks, e.g. `[[jane-doe]]`. Link liberally — a link to a page that does not exist yet marks it as worth creating.
- Every wiki page starts with YAML frontmatter:

```yaml
---
type: source | synthesis | {{category singular, e.g. person}}
created: YYYY-MM-DD
updated: YYYY-MM-DD
tags: []
---
```

- Source pages additionally get `source: <filename in raw/>` and `source-date:` when known.
- `wiki/index.md` carries a one-line hook per page (`[[page]] — when/why it matters`). It is the retrieval layer: find the right page from the index before opening it.
- Keep pages focused: one entity or concept per page. Split a page when a section outgrows its host.
- When new information contradicts an existing claim, do not silently overwrite it. Note the contradiction on the page and flag it in the log entry.

## Workflows

### Ingest

A source lands in `raw/`. Read it, discuss key takeaways, write a summary page in `wiki/sources/`, update every page it touches, update the index, log, and commit (`ingest:`). A source contradicting an existing claim gets the contradiction noted, not a silent overwrite.

### Query

Read `wiki/index.md` to find relevant pages, then read them. Answer with citations. A substantive answer can be filed to `wiki/syntheses/` — index, log, and commit it (`query:`).

### Lint

On request, check for contradictions, stale claims, broken links, and gaps. Report findings; apply only what the user approves; log and commit (`lint:`). When a convention in this schema stops fitting the vault, propose amending this file and commit that too (`schema:`).

## Log format

Entries in `wiki/log.md`, newest last, prefix-parseable:

```text
## [YYYY-MM-DD] ingest | Source Title
## [YYYY-MM-DD] query | Question asked
## [YYYY-MM-DD] lint | Scope
## [YYYY-MM-DD] schema | What changed
```

## Git

The vault is a git repository. Commit after every write, with a labelled message (`ingest:`, `query:`, `lint:`, `schema:`). **Never `git add -A` or `git add .`** — stage only the exact paths the operation touched. A source in `raw/` is committed for the first time by the ingest that files it; until then, `git status --short raw/` is the inbox.

## Style

Write for future retrieval: dense, factual, specific, no filler. Convert relative dates to absolute. This schema is negotiable — when a convention isn't working, propose a change through lint and update this file.
