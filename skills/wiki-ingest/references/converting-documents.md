# Converting non-markdown sources

The original file is staged in `raw/inbox/` and becomes immutable when moved to `raw/archive/<document-id>/<revision-id>/original.<ext>` — never edit or replace it. Write the converted Markdown extraction to `raw/extracted/<document-id>/<revision-id>.md` (create the directories if needed). Extractions are derived artifacts: regenerable, and the source page's `source:` field still points at the **archived original**.

Begin every extraction with a short header noting the `document-id`, `revision-id`, archived original path, original filename, SHA-256, conversion method, and date, so staleness is detectable if the original changes.

Every converter below is optional. Probe for it before use; if it is absent, say what is missing and how to convert manually, and leave the rest of `wiki-ingest` usable for other sources.

## .pdf

- Try `pdftotext -layout raw/archive/<document-id>/<revision-id>/original.pdf raw/extracted/<document-id>/<revision-id>.txt` (poppler) first, then normalize that text into the required Markdown extraction; if unavailable, use whatever PDF-reading capability is already at hand to extract the text.
- For scanned or image PDFs with no text layer, view the page images directly and transcribe what matters; note "OCR-by-LLM" in the extraction header and record any omitted regions.
- Preserve heading structure where detectable; drop running headers and footers.
- For a large work (a book, a long report), insert page markers as HTML comments (`<!-- p.87 -->`) at each page break instead of dropping page numbers — citations in wiki pages reference them (`ch. 3, p. 87`), keeping every claim traceable to the raw file.

## .docx

- Try `pandoc raw/archive/<document-id>/<revision-id>/original.docx -t gfm -o raw/extracted/<document-id>/<revision-id>.md` first; if unavailable, use whatever document-reading capability is already at hand.
- Keep tracked changes and comments only if the user says they matter.
- Export embedded images to `raw/assets/<document-id>/<revision-id>/` and reference them from the extraction if they carry meaning.

## .xlsx / .csv

- Inspect sheets directly, or with a spreadsheet-reading tool if one is at hand.
- Do not dump a large sheet wholesale. For each sheet, record: purpose, column schema, row count, and a representative sample (roughly 10 rows) as a markdown table, plus any aggregate that matters (totals, ranges, distributions).
- A workbook small enough (under roughly 50 rows total) can be dumped as a full markdown table instead.

## Images (.png, .jpg)

- View directly; describe the content in the extraction. Store the image itself in `raw/assets/<document-id>/<revision-id>/`.

## Other formats

- .pptx: extract per-slide text, one section per slide.
- .html: convert to markdown, or extract the readable text directly.
- Audio/video: no local transcription is assumed — ask the user for a transcript.

## Sanity check

After conversion, skim the extraction against the original for garbled text, missing sections, or mangled tables before filing. Note any known gap in the extraction header.
