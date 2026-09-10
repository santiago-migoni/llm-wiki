# Converting non-markdown sources

The source is staged in raw/inbox/. During ingestion, keep the current original under raw/sources/<slug>/source.<ext> and write the complete normalized extraction to raw/sources/<slug>/extracted.md.

The extraction is derived and regenerable. It is never a replacement for the current original. Git preserves prior committed source and extraction states.

Begin every extracted.md with a short header containing:

- stable slug;
- current source path;
- original filename;
- SHA-256 when available;
- conversion method;
- extraction date;
- warnings or omitted regions.

All converters are optional. Probe before use. If a converter is unavailable, state the limitation and use the best available reading capability without blocking other source types.

## PDF

- Try pdftotext -layout on raw/sources/<slug>/source.pdf, then normalize the result into raw/sources/<slug>/extracted.md.
- For scanned or image-only PDFs, inspect page images and record OCR-by-LLM plus any omitted regions.
- Preserve detectable headings, lists, and tables.
- Drop repeated running headers and footers when they do not carry meaning.
- For long documents, add page markers such as <!-- p.87 --> so claims can be traced to the original.

## DOCX

- Try pandoc raw/sources/<slug>/source.docx -t gfm, then write the result to raw/sources/<slug>/extracted.md.
- Preserve headings, lists, tables, footnotes, and links where possible.
- Keep tracked changes and comments only when the user says they matter.
- Export meaningful embedded images to raw/sources/<slug>/assets/ and reference them from extracted.md.

## XLSX and CSV

- Inspect sheets directly or use an available spreadsheet-reading capability.
- For each sheet, record its purpose, column schema, row count, representative sample, and relevant aggregates.
- Do not dump a large workbook wholesale.
- A workbook with fewer than roughly 50 total rows may be represented as a complete Markdown table.
- Preserve sheet names and cell or range markers for load-bearing values.

## Images

- Inspect the image directly.
- Describe meaningful visual content in extracted.md.
- Keep the current image under raw/sources/<slug>/source.<ext>; place additional meaningful assets in assets/.

## PPTX

- Extract text per slide, preserving slide numbers and meaningful speaker notes.
- Record important diagrams or visual information in extracted.md and keep relevant images in assets/.

## HTML

- Extract readable content to Markdown.
- Preserve title, headings, links, lists, tables, and meaningful metadata.
- Exclude navigation and repeated boilerplate when it is not part of the source.

## Audio and video

No local transcription is assumed. Ask for a transcript or use an available transcription capability. If only a partial transcript is available, record the coverage and omissions.

## Sanity check

After conversion:

1. compare extracted.md with the original;
2. check headings, lists, tables, images, signatures, page or slide markers, and footnotes;
3. record garbled text or missing regions;
4. never silently present an incomplete extraction as complete.
