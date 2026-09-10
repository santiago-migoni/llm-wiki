---
name: wiki-query
description: Answer questions from an LLM-maintained wiki vault with traceable citations. Use when the user asks what the wiki says, asks for an answer based on their documents, requests a comparison, or asks for full or historical context. Reads the current source layer and Git history; it never ingests a new source.
---

# Wiki query

Answer from the current wiki and source layer. Use Git history for historical questions. Do not ingest, move, or rewrite sources during a query.

Load references/vault-protocol.md first. Load references/context-modes.md when selecting an exhaustive or historical mode, filing a synthesis, or when a coverage report is required. Resolve references from the plugin package, not from the current working directory.

## Steps

1. Resolve the vault and read AGENTS.md or the authoritative schema in full.
2. Choose the narrowest context mode that satisfies the request:
   - **Targeted** by default;
   - **Current-corpus** for full context, all current documents, or exhaustive analysis;
   - **Historical** for changes, prior states, provenance, or superseded claims.
3. Follow the selected procedure in references/context-modes.md.
4. In targeted mode, read wiki/index.md before the relevant canonical pages. Use a read-only search over wiki/ and raw/sources/ only when the index does not surface the topic.
5. In current-corpus mode, inventory raw/sources/, inspect raw/inbox/, verify source/extraction coverage, read all wiki/pages/ pages and current extractions in bounded passes, and track coverage by slug.
6. In historical mode, use Git log, diff, and show for the relevant stable slug or page. Cite the commit and exact path for historical claims.
7. Answer only from material actually read. If outside knowledge is useful, label it explicitly as outside the vault.
8. Cite the evidence path:
   - canonical page and section;
   - stable source slug;
   - current source or extracted.md path;
   - page, section, slide, sheet, cell, timestamp, or other marker;
   - Git commit and path for historical claims.
9. If sources disagree, surface the disagreement and identify the competing evidence. Do not silently select one source.
10. If the vault has no relevant information, say so and offer to ingest a source.
11. State context limitations. Never claim complete or absolute context when a source is pending, skipped, unreadable, unsupported, or materially partial.
12. If the answer is durable and substantive, offer to file it in wiki/syntheses/. If accepted, follow the synthesis procedure in references/context-modes.md and create a query: commit.

## Context claims

“Absolute context” means an auditable coverage procedure, not that an unlimited corpus fits in one model window.

A broad answer must distinguish:

- sources fully read;
- sources partially read;
- sources skipped or unreadable;
- pending inbox files;
- inaccessible or unsupported formats;
- contradictions that affect the answer.

A targeted answer does not need a full-corpus ledger, but it must still identify the canonical pages and source paths used.

## Synthesis boundary

Do not file transient answers or duplicate an existing canonical page. Before creating a synthesis, search wiki/syntheses/ and wiki/index.md for an existing analysis on the same question. Prefer updating the existing synthesis when it is the durable home.

A filed synthesis must preserve its context mode, source slugs, page links, and coverage limitations.

## Style

Keep answers dense, factual, and grounded in the vault. Prefer canonical pages for conceptual answers, extracted.md for source-grounded claims, and source.<ext> for layout-dependent evidence.
