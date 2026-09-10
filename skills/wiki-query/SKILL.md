---
name: wiki-query
description: Answer a question from an LLM-maintained wiki vault, with citations to wiki pages, archived originals, and complete extractions. Use when the user asks a question against their wiki or knowledge base ("what does the wiki say about…", "based on my notes…", "compare X and Y from my sources"). Distinct from wiki-recall (out of scope in this plugin) and from wiki-ingest; query reads and answers, it never files a new source into raw/inbox/ or raw/archive/.
---

# Wiki query

Answer from the wiki layer, cite pages, and optionally file the answer back into the wiki.

Load the plugin's `references/vault-protocol.md` first; its vault-resolution, schema-authority, path-lane, git, and retrieval-first rules govern every step below. Resolve that reference from the plugin package, not from the current working directory.

## Steps

1. Resolve the vault per the protocol; read its schema file. The vault's schema overrides this skill's configurable conventions where they conflict.
2. Decide the context mode from the request:
   - **Targeted** (default): read `wiki/index.md` to shortlist relevant pages. If the index hooks don't surface the topic, fall back to `grep -ril '<term>' <vault-root>/wiki/` to search page content directly.
   - **Current-corpus**: use when the user asks for "full context" or an exhaustive answer about the current state. Inventory `raw/catalog.md`, `raw/archive/`, `raw/extracted/`, and `raw/inbox/`; stop the full-context claim if there are pending sources, missing current extractions, or unreadable files.
   - **Historical**: use when the user asks what changed, requests all versions, or asks about superseded claims. Include the relevant prior `revision-id`s in the coverage checklist.
3. In targeted mode, read the shortlisted pages, most relevant first, then open the underlying current complete extraction whenever a claim is load-bearing. Read historical extractions only when the question needs them. If more pages are relevant than fit comfortably in context, read the most relevant subset and say explicitly which pages were not read, rather than silently truncating.
4. In current-corpus mode, read every current `wiki/knowledge/<document-id>.md` page and its current source summary, then every current complete extraction in bounded passes when exhaustive source grounding is required. In historical mode, add every relevant superseded revision. Maintain a coverage checklist by `document-id` and `revision-id`; if the corpus exceeds one context window, continue across passes and report exactly what was covered and what was not. Do not claim absolute or complete context without that coverage check.
5. Answer with citations to the wiki pages drawn on (`[[page-name]]`), the stable `document-id`, exact `revision-id`, archived original, and extraction or page marker where applicable. If two pages or sources disagree on a point the question touches, surface the disagreement rather than picking a side silently.
6. If the vault has nothing on the topic, say so plainly — do not answer from outside the vault without saying that is what is happening.
7. If the answer is substantive — a comparison, an analysis, a decision rationale — offer to file it in `wiki/syntheses/`. If accepted, follow the protocol's write procedure: write the synthesis page with the vault's frontmatter conventions plus wikilinks, record context mode and `document-id`/`revision-id` coverage, update `wiki/index.md`, append the log entry, and commit (label `query:`).
8. Answers need not be plain prose: when the question calls for it, the filed page can hold a comparison table or a Marp slide deck (`marp: true` frontmatter, `---` slide separators — renders in Obsidian's Marp plugin). A rich-output answer still gets filed the same way as any synthesis.

## Style

Keep answers grounded in what the wiki actually contains. Say plainly when it has nothing on the topic, and offer to ingest a new source instead of guessing.
