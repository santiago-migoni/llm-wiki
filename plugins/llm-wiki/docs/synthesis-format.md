# Synthesis format

Durable analyses under `wiki/syntheses/` use the same structured provenance
contract as canonical knowledge pages. They may combine source records and
canonical pages, but must not pretend that one source is authoritative when
the analysis depends on several.

## Frontmatter

~~~yaml
---
type: synthesis
slug: policy-comparison
created: 2026-09-11
updated: 2026-09-11
status: current
context-mode: current-corpus
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
    role: counterpoint
pages: [security-policy, audit-findings]
commits: []
---
~~~

`context-mode` is `targeted`, `current-corpus`, or `historical`. `sources` is
an ordered list of complete provenance entries. It may be empty when the
synthesis is based only on canonical pages, but every source-grounded claim
must have the corresponding entry. `pages` lists canonical pages used by the
analysis. `commits` lists Git commits used for historical claims.

## Body

Use claim-level citations and distinguish the nature of each statement:

~~~markdown
## Finding

- The review interval is annual.
  - Evidence: `corporate-policy`, section “Review cycle”.
  - Support: `audit-report`, page 14.

## Interpretation

- The second document narrows the operational scope; this is an interpretation
  of the two cited records, not a direct source statement.

## Limitations

- Sources fully read: 2
- Sources skipped or unreadable: 0
- Pending inbox files: 0
~~~

For a current-corpus or historical synthesis, preserve the coverage ledger or
an equivalent compact report. Do not call the synthesis exhaustive when any
source is skipped, pending, unreadable, unsupported, or materially partial.
