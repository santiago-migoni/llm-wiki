# Ingest source record

This reference defines the record created or updated for one logical source. It is used by wiki-ingest after loading references/vault-protocol.md.

## Stable source record

The current record for a logical document is:

~~~text
raw/sources/<slug>/
├── source.<ext>
├── extracted.md
└── assets/
~~~

There is exactly one current original matching source.*. The extension may change when the source format changes, but the directory keeps the same slug.

The slug may be flat (`policy`) or hierarchical (`fundamentos/actividades`).
Each segment must be lowercase kebab-case. Namespace directories such as
`raw/sources/fundamentos/` are organizational only; the leaf directory is the
record. The same sub-slug may appear under different namespaces because the
complete slash-separated slug is the identity.

Do not add revision numbers, upload dates, hashes, manifests, catalogs, or duplicate source pages to this record. Git commits provide the historical record.

## Classify the input

Use the source content, existing index routes, current source directories, page frontmatter, title, and semantic role together.

### New source

Create a new slug when no existing logical document represents the source.

### Update

Reuse the existing slug when the source is a newer or corrected version of the same logical document. The current original, extraction, and affected canonical pages are updated in place. Git must show the previous committed state.

### Exact duplicate

Compute the hash of the input and compare it with every current source.* file under raw/sources/**/, not only the likely matching slug. If no current source matches and the repository has history, inspect historical source.* blobs under every source slug and compare their bytes as well. If the hash already exists in the current state or in any slug's history, classify the input as a duplicate of an existing state. A filename or proposed slug change does not make identical bytes a new source. In either case:

- do not create a new source directory;
- do not regenerate pages or assets;
- do not append a log entry;
- do not create a commit;
- report the existing slug, current paths, and historical commit when applicable.

If the same bytes appear under a different filename, it is still a duplicate.

### Ambiguous identity

If the source could be a new document or an update, do not guess. Preserve the input in raw/inbox/ and ask the user which logical document it belongs to.

Filename similarity alone is not enough to identify an update.

## Protect local work

Before updating an existing slug, inspect Git status for:

- raw/sources/<slug>/;
- the expected canonical page;
- affected category indexes;
- wiki/index.md;
- wiki/overview.md;
- wiki/log.md.

If there are uncommitted changes that were not created by the current ingest, stop before replacement and ask how to proceed. Never overwrite a user's local source, extraction, page, or asset edits silently.

## Current source metadata

Include this header at the beginning of every extracted.md. The complete field and format-specific
coverage contract is defined in [docs/extraction-coverage.md](../../../docs/extraction-coverage.md).

~~~yaml
---
type: extraction
slug: <stable-slug>
source: raw/sources/<slug>/source.<ext>
original-filename: <filename supplied by the user>
sha256: <hash of current source>
extracted: YYYY-MM-DD
format: markdown | txt | html | pdf | docx | xlsx | csv | pptx | image | audio | video | other
method: <converter or reading method>
status: complete | representative | partial | unsupported
coverage:
  unit: pages | slides | sheets | rows | seconds | regions
  expected: <non-negative integer>
  processed: <non-negative integer not greater than expected>
warnings: []
---
~~~

Use a concise warning for every non-complete status. `complete` requires
`coverage.processed == coverage.expected`; `representative` identifies a deliberate sample;
`partial` identifies omitted or unreadable material; and `unsupported` means that no trustworthy
unit was processed. Do not claim complete extraction when meaningful sections, tables, pages,
slides, cells, images, or audio were not represented.

The hash is for duplicate detection and freshness checks. It is not an identifier and must not be used to create a new path.

## Canonical knowledge page

When the source represents durable knowledge, create or update one page under wiki/pages/. Prefer the existing page.

~~~yaml
---
type: knowledge
slug: <stable-page-slug>
created: YYYY-MM-DD
updated: YYYY-MM-DD
sources:
  - slug: <source-slug>
    source: raw/sources/<source-slug>/source.<ext>
    extracted: raw/sources/<source-slug>/extracted.md
    sha256: <64 lowercase hexadecimal characters>
    role: primary
tags: []
status: current
---
~~~

The page body should be selective and retrieval-oriented:

~~~markdown
# <Title>

## Current understanding

<Current claims, definitions, or procedures.>

## Evidence

- <Claim>.
  - Evidence: `<source-slug>`, <section or marker>.

## Open questions

- <Unresolved question, if any.>

## Related pages

- [[related-page]]
~~~

A page can cite several sources. Add one complete entry to `sources` for each
source and keep the entries in stable role/slug order. For a claim that combines
sources, use one `Evidence` or `Support` line per source:

~~~markdown
- <Claim supported by two records>.
  - Evidence: `<primary-slug>`, <section or marker>.
  - Support: `<supporting-slug>`, <section or marker>.
~~~

Citation slugs must already be declared in frontmatter. Do not copy the full
extraction into the page.

## Provenance updates

When a new source supports an existing page, append its structured entry in the
stable order, add citations to the affected claims, and list every affected
page before writing. When a source changes, keep its slug, update its source,
extraction, hash, and every affected page entry in one coherent operation. When
a source disappears or no longer supports a claim, preserve the old entry until
an explicit reviewed change removes it and updates the claims. Never drop a
source from a page merely because a lookup failed.

Existing pages with singular `source`, `extracted`, and `sha256` fields remain
readable. Treat them as migrable and use `wiki-migrate-provenance --check`
before an explicit `--write`; do not mix the singular and structured formats.

## Update rules

- Replace the current source only after computing its input hash and checking for uncommitted target changes.
- Preserve the supplied bytes in source.<ext>; do not edit the original to repair extraction.
- Regenerate extracted.md when the source bytes change.
- If the extension changes, keep one current source.<ext> and include the old path removal in the same coherent Git change. The old file remains recoverable from the prior commit.
- Add or update relevant assets. Do not delete possibly useful old assets automatically; let lint or an explicit cleanup request handle pruning.
- Update existing pages and indexes before creating new ones.
- A category directory may receive a short index entry or link, never a second copy of the canonical page.
- Record contradictions and extraction limitations in the canonical page or wiki/log.md.

## Commit boundary

A successful ingest commit may include:

- raw/sources/<slug>/source.<ext>;
- raw/sources/<slug>/extracted.md;
- raw/sources/<slug>/assets/ files;
- affected wiki/pages/ files;
- affected category indexes;
- wiki/index.md;
- wiki/overview.md when materially changed;
- wiki/log.md.

Stage only the exact paths touched. Use an ingest: commit message. A duplicate is a no-op and must not create a commit.

## Ingest report

Report at least:

- operation: new, update, duplicate, or blocked/ambiguous;
- stable slug;
- source and extraction paths;
- canonical pages created or updated;
- assets added or retained;
- source hash;
- extraction format, status, processed/expected coverage, and warnings;
- contradictions or open questions;
- Git commit, or why no commit was created.
