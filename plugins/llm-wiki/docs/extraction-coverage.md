# Extraction coverage contract

`raw/sources/<slug>/extracted.md` is a derived reading artifact, not a promise that every
format was represented perfectly. Every new extraction declares its format, processing method,
status, measurable coverage, and warnings so a query can distinguish complete, representative,
partial, and unsupported material.

## Required frontmatter

Use this header before the normalized extraction body:

~~~yaml
---
type: extraction
slug: <source-slug>
source: raw/sources/<source-slug>/source.<ext>
original-filename: <filename supplied by the user>
sha256: <64 lowercase hexadecimal characters>
extracted: YYYY-MM-DD
format: markdown | txt | html | pdf | docx | xlsx | csv | pptx | image | audio | video | other
method: <converter, reader, OCR, transcription, or supervised method>
status: complete | representative | partial | unsupported
coverage:
  unit: pages | slides | sheets | rows | seconds | regions
  expected: <non-negative integer>
  processed: <non-negative integer not greater than expected>
warnings: []
---
~~~

`sha256` is optional when the source cannot be hashed, but it must be valid lowercase SHA-256
when present. `warnings` is always a list of non-empty strings; a non-complete status must include
at least one warning explaining the limitation. The deterministic validator checks the header and
returns the coverage summary in JSON.

## Status semantics

- `complete`: every expected unit was processed (`processed == expected`). Warnings may describe
  non-material normalization, such as removed navigation, but must not hide omitted meaning.
- `representative`: the artifact intentionally contains a sample or map of a larger source
  (`processed < expected`). The complete structured or binary derivative must be retained when it
  is needed to support exact claims.
- `partial`: some expected units or meaningful regions were not represented, or a conversion
  limitation remains even when the unit count alone cannot express it. The warning names the
  omitted content and the locator or range where possible.
- `unsupported`: no trustworthy extraction was produced (`processed == 0`). Preserve the original
  and state what capability, input, or access is missing.

The status is about extraction coverage, not source authority. A `complete` extraction can still
contain a claim that conflicts with another source, and a `partial` extraction can contain useful
evidence. Query and lint must report the status before relying on the derived body.

## Units by format

| Source format | Coverage unit | Minimum format-specific evidence |
|---|---|---|
| Markdown, TXT, HTML | `pages` | All meaningful content; navigation or boilerplate omissions are warnings. For non-paginated text, one page represents the document as a whole. |
| PDF | `pages` | Expected and processed pages, page markers, text-versus-OCR method, and relevant tables or images. |
| DOCX | `pages` | Headings, lists, tables, links, footnotes, comments and tracked-change treatment, plus meaningful embedded images. |
| XLSX | `sheets` | Sheet names, purpose, schema, row counts, samples or complete structured derivatives, and relevant aggregates. |
| CSV/TSV | `rows` | Column schema, expected and processed rows, samples or a complete structured derivative, and range markers. |
| PPTX | `slides` | Slide count, slide text, speaker notes, and visual content not represented as text. |
| Image | `regions` | Meaningful visual regions, OCR method when used, and preserved supporting assets. |
| Audio, video | `seconds` | Total duration, transcribed range, timestamps, and omitted segments. |

The `format` and `unit` values must agree with the current `source.<ext>`. Unknown extensions use
`other` with `pages` and must explain the chosen reading method in `warnings` when it is not
complete.

## Locator and derivative rules

- PDF evidence uses page markers such as `<!-- p.87 -->`.
- Slide evidence uses slide numbers and preserves meaningful speaker notes.
- Workbook evidence cites sheet names and cell or range markers. Large workbooks keep a complete
  CSV/JSONL derivative when exact rows are load-bearing and use `extracted.md` as its map, schema,
  metrics, and samples.
- Audio and video evidence uses timestamps and records the transcribed interval.
- Images, diagrams, signatures, and layout-dependent evidence remain available in the original
  or `assets/`; a Markdown description does not replace them.

Never silently downgrade a missing section, page, slide, cell range, image region, or time segment.
If the source cannot be checked completely, use `partial`, `representative`, or `unsupported`
and explain the exception.
