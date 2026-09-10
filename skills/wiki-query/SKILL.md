---
name: wiki-query
description: Answer questions from an LLM-maintained wiki vault with traceable citations. Use when the user asks what the wiki says, asks for an answer based on their documents, requests a comparison, or asks for full or historical context. Reads the current source layer and Git history; it never ingests a new source.
---

# Wiki query

Answer from the wiki and current source layer. Use Git history for historical questions. Do not create archive or revision folders.

Load references/vault-protocol.md first. Resolve that reference from the plugin package, not from the current working directory.

## Steps

1. Resolve the vault and read AGENTS.md or the authoritative schema in full.
2. Determine the context mode:
   - **Targeted** by default;
   - **Current-corpus** for full context, all documents, or exhaustive current analysis;
   - **Historical** for changes, prior states, provenance, or superseded claims.
3. In targeted mode:
   - read wiki/index.md;
   - shortlist relevant pages in wiki/pages/;
   - read relevant syntheses and category indexes;
   - read raw/sources/<slug>/extracted.md for load-bearing claims;
   - inspect source.<ext> when layout, images, signatures, tables, or other evidence matters;
   - if the index misses the topic, use a read-only search over wiki/ and raw/sources/.
4. In current-corpus mode:
   - inventory every current source slug under raw/sources/;
   - inspect raw/inbox/ and report pending files;
   - verify every source has exactly one current source.<ext> and extracted.md;
   - read all canonical pages in wiki/pages/;
   - read current extractions in bounded passes;
   - track coverage by slug;
   - report inaccessible, unsupported, oversized, or contradictory material.
5. In historical mode:
   - use Git log, diff, and show on the relevant source or page paths;
   - read only the history needed for the question;
   - identify the commits and paths supporting historical claims.
6. Answer only from the material actually read. If outside knowledge is useful, label it explicitly as outside the vault.
7. Cite:
   - canonical page paths;
   - stable source slugs;
   - current source or extracted.md paths;
   - page, section, slide, cell, or extraction markers;
   - Git commits or diffs for historical claims.
8. If sources disagree, surface the disagreement and identify the competing sources. Do not silently choose one.
9. If the vault has no relevant information, say so and offer to ingest a source.
10. If the answer is durable and substantive, offer to file it in wiki/syntheses/. If accepted:
    - write a synthesis with links to the canonical pages and sources;
    - record context mode and coverage;
    - update wiki/index.md;
    - update wiki/overview.md only if the global orientation changed;
    - append wiki/log.md;
    - commit with a query: label.

## Context claims

“Absolute context” means an auditable coverage procedure, not that an unlimited corpus fits in one model window.

Never say that all documents were considered unless the current-corpus checklist was completed. If a source was skipped, pending, unreadable, unsupported, or outside the available host context, state that limitation.

## Style

Keep answers dense, factual, and grounded in the vault. Prefer canonical pages over raw extraction when the question is conceptual, and prefer the extraction or original when the question depends on exact source wording or layout.
