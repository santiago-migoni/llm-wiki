# Structured provenance

LLM Wiki stores the current original and its normalized extraction in one
stable source record. A canonical page or durable synthesis can refer to one
or more source records through an ordered `sources` list. Markdown remains the
source of truth; Git preserves previous source, extraction, and page states.

## Page contract

Pages written with the current contract use this frontmatter:

~~~yaml
---
type: knowledge
slug: security-policy
created: 2026-09-11
updated: 2026-09-11
sources:
  - slug: corporate-policy
    source: raw/sources/corporate-policy/source.pdf
    extracted: raw/sources/corporate-policy/extracted.md
    sha256: <64 lowercase hexadecimal characters>
    role: primary
  - slug: audit-report
    source: raw/sources/audit-report/source.docx
    extracted: raw/sources/audit-report/extracted.md
    sha256: <64 lowercase hexadecimal characters>
    role: supporting
tags: []
status: current
---
~~~

Each entry requires:

- `slug`: the existing lowercase kebab-case source-record slug;
- `source`: the current `raw/sources/<slug>/source.<ext>` path;
- `extracted`: the current `raw/sources/<slug>/extracted.md` path;
- `sha256`: the hash of the current original bytes;
- `role`: one of `primary`, `supporting`, `context`, or `counterpoint`.

The source and extraction paths must belong to the same slug, exist in the
vault, and remain vault-relative. A page cannot repeat a slug or source path.
The hash is checked against every entry, not only the first one.

## Stable order and roles

The order is part of the page's readable provenance. New writes use this
deterministic order: `primary`, `supporting`, `context`, then `counterpoint`;
entries with the same role are ordered by slug. There is no requirement that a
page have a primary source: this allows a synthesis of equally authoritative
or only contextual sources. The role communicates how a claim uses the source,
not an absolute trust ranking.

When an existing source changes, retain its slug, replace the current source
and extraction together, recompute its hash, and update every affected page
entry. The old bytes and page state remain available through Git. When a
source disappears, do not remove its provenance silently: the page becomes
invalid until the source is restored or an explicit update removes the entry
and adjusts the affected claims in the same reviewable change.

## Claim-level citations

Page metadata says which sources are available; the body says which source
supports each claim. A multi-source claim must identify the source slugs in
backticks and include a marker that lets a reader find the evidence:

~~~markdown
- The policy requires annual review.
  - Evidence: `corporate-policy`, section “Review cycle”.
  - Support: `audit-report`, page 14.
~~~

`Evidence` is the direct basis for the claim. `Support` is corroborating,
contextual, or countervailing material. A citation slug must be present in the
page's `sources` list. If sources disagree, keep both citations and state the
contradiction instead of choosing silently.

The deterministic validator checks the metadata, hashes, paths, roles, order,
duplicate entries, and explicit citation slugs. Whether a citation actually
supports the prose remains a semantic review responsibility.

## Compatibility and migration

Readers continue to accept the legacy singular fields:

~~~yaml
source: raw/sources/security-policy/source.pdf
extracted: raw/sources/security-policy/extracted.md
sha256: <hash>
~~~

The singular shape is reported as `migrable`, not as invalid solely because it
is old. New writes must use `sources`. To inspect a migration without changing
the vault:

~~~bash
python3 <plugin-root>/scripts/llm-wiki migrate-provenance <vault> --check
~~~

The command is also available as
`<plugin-root>/scripts/wiki-migrate-provenance`. `--check` is the default and
prints the exact affected pages plus a unified diff. `--write` is required to
apply the plan; it writes pages only, creates no commit, and leaves the diff
available for review. A stale supplied hash or a page that mixes singular and
structured fields blocks the whole migration.

Migration computes a missing legacy hash from the current source, but never
silently repairs a supplied stale hash. It infers a missing extraction path
only when the canonical `extracted.md` exists at the source-record path.
