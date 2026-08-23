# Converting non-markdown sources

The original file in `raw/` is immutable — never edit or replace it. Write the converted markdown extraction to `raw/extracted/<original-stem>.md` (create the directory if needed). Extractions are derived artifacts: regenerable, and the source page's `source:` field still points at the **original** file.

Begin every extraction with a short header noting the original filename, conversion method, and date, so staleness is detectable if the original changes.

Every converter below is optional. Probe for it before use; if it is absent, say what is missing and how to convert manually, and leave the rest of `wiki-ingest` usable for other sources.

## .pdf

- Try `pdftotext -layout file.pdf out.txt` (poppler) first; if unavailable, use whatever PDF-reading capability is already at hand to extract the text.
- For scanned or image PDFs with no text layer, view the page images directly and transcribe what matters; note "OCR-by-LLM" in the extraction header.
- Preserve heading structure where detectable; drop running headers and footers.
- For a large work (a book, a long report), insert page markers as HTML comments (`<!-- p.87 -->`) at each page break instead of dropping page numbers — citations in wiki pages reference them (`ch. 3, p. 87`), keeping every claim traceable to the raw file.

## .docx

- Try `pandoc file.docx -t gfm -o out.md` first; if unavailable, use whatever document-reading capability is already at hand.
- Keep tracked changes and comments only if the user says they matter.
- Export embedded images to `raw/assets/` and reference them from the extraction if they carry meaning.

## .xlsx / .csv

- Inspect sheets directly, or with a spreadsheet-reading tool if one is at hand.
- Do not dump a large sheet wholesale. For each sheet, record: purpose, column schema, row count, and a representative sample (roughly 10 rows) as a markdown table, plus any aggregate that matters (totals, ranges, distributions).
- A workbook small enough (under roughly 50 rows total) can be dumped as a full markdown table instead.

## Images (.png, .jpg)

- View directly; describe the content in the extraction. Store the image itself in `raw/assets/`.

## Other formats

- .pptx: extract per-slide text, one section per slide.
- .html: convert to markdown, or extract the readable text directly.
- Audio/video: no local transcription is assumed — ask the user for a transcript.

## Sanity check

After conversion, skim the extraction against the original for garbled text, missing sections, or mangled tables before filing. Note any known gap in the extraction header.
