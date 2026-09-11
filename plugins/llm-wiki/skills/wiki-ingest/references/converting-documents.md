# Converting non-markdown sources

Read references/source-record.md for source identity, duplicate handling, metadata, and update rules,
and [docs/extraction-coverage.md](../../../docs/extraction-coverage.md) for the coverage contract.
This reference covers only format-specific conversion.

The source is staged in raw/inbox/. During ingestion, keep the current original under
raw/sources/<slug>/source.<ext> and write the normalized extraction with its declared status and
coverage to raw/sources/<slug>/extracted.md.

The extraction is derived and regenerable. It is never a replacement for the current original. Git preserves prior committed source and extraction states.

Begin every extracted.md with a header containing:

- stable slug;
- current source path;
- original filename;
- SHA-256 when available;
- conversion method;
- extraction date;
- format matching source.<ext>;
- status: complete, representative, partial, or unsupported;
- coverage unit, expected count, and processed count;
- warnings or omitted regions.

All converters are optional. Probe before use. If a converter is unavailable, state the limitation and use the best available reading capability without blocking other source types.

## PDF

- Try pdftotext -layout on raw/sources/<slug>/source.pdf, then normalize the result into raw/sources/<slug>/extracted.md.
- Set coverage unit to pages. Record expected and processed page counts, preserve page markers,
  and distinguish text extraction from OCR in method and warnings.
- For scanned or image-only PDFs, inspect page images and use partial or complete only according to
  the pages and meaningful regions actually checked; record OCR-by-LLM plus any omitted regions.
- Preserve detectable headings, lists, and tables.
- Drop repeated running headers and footers when they do not carry meaning.
- For long documents, add page markers such as <!-- p.87 --> so claims can be traced to the original.

## DOCX

- Try pandoc raw/sources/<slug>/source.docx -t gfm, then write the result to raw/sources/<slug>/extracted.md.
- Set coverage unit to pages and declare how comments and tracked changes were treated.
- Preserve headings, lists, tables, footnotes, and links where possible.
- If comments or tracked changes are omitted, include an explicit warning and do not use complete
  when they contain meaningful content.
- Export meaningful embedded images to raw/sources/<slug>/assets/ and reference them from extracted.md.

## XLSX and CSV

- Inspect sheets directly or use an available spreadsheet-reading capability.
- For XLSX, set coverage unit to sheets. For CSV/TSV, set it to rows.
- For each sheet, record its purpose, column schema, row count, representative sample, and relevant aggregates.
- For large files, keep a complete structured derivative such as CSV per sheet or JSONL and use
  extracted.md as its map, schema, metrics, and samples. Use representative when only samples are
  present; cite sheet names, cells, and ranges for load-bearing values.
- A workbook with fewer than roughly 50 total rows may be represented as a complete Markdown table
  only when all relevant rows were processed.

## Images

- Inspect the image directly.
- Set coverage unit to regions and count the meaningful visual regions actually inspected.
- Describe meaningful visual content in extracted.md.
- Keep the current image under raw/sources/<slug>/source.<ext>; place additional meaningful assets in assets/.

## PPTX

- Set coverage unit to slides and record expected and processed slide counts.
- Extract text per slide, preserving slide numbers and meaningful speaker notes.
- Record important diagrams or visual information not represented as text in extracted.md and keep
  relevant images in assets/; warn when visual content was not inspected.

## HTML

- Set coverage unit to pages; for a non-paginated document, expected and processed value 1 means
  the document as a whole.
- Extract readable content to Markdown.
- Preserve title, headings, links, lists, tables, and meaningful metadata.
- Exclude navigation and repeated boilerplate when it is not part of the source.

## Audio and video

No local transcription is assumed. Set coverage unit to seconds, record total duration and the
transcribed range with timestamps, and use partial when segments are missing. Ask for a transcript
or use an available transcription capability; unsupported means no trustworthy segment was produced.

## Sanity check

After conversion:

1. compare extracted.md with the original;
2. check headings, lists, tables, images, signatures, page or slide markers, and footnotes;
3. record garbled text or missing regions;
4. reconcile expected and processed coverage;
5. never silently present an incomplete extraction as complete.
