---
name: wiki-query
description: Answer a question from an LLM-maintained Obsidian wiki vault, with citations to wiki pages and raw sources. Use when the user asks a question against their wiki or knowledge base ("what does the wiki say about…", "based on my notes…", "compare X and Y from my sources"). Distinct from wiki-recall (out of scope in this plugin) and from wiki-ingest: query reads and answers, it never files a new source into raw/.
---

# Wiki query

Answer from the wiki layer, cite pages, and optionally file the answer back into the wiki.

Load `${CLAUDE_PLUGIN_ROOT}/references/vault-protocol.md` first; its vault-resolution, schema-authority, path-lane, git, and retrieval-first rules govern every step below.

## Steps

1. Resolve the vault per the protocol; read its schema file. The vault's schema overrides this skill where they conflict.
2. Read `wiki/index.md` to shortlist relevant pages. If the index hooks don't surface the topic, fall back to `grep -ril '<term>' <vault-root>/wiki/` to search page content directly.
3. Read the shortlisted pages, most relevant first. If more pages are relevant than fit comfortably in context, read the most relevant subset and say explicitly which pages were not read, rather than silently truncating.
4. Answer with citations to the wiki pages drawn on (`[[page-name]]`), and to the underlying `raw/` source where the claim is load-bearing. If two pages disagree on a point the question touches, surface the disagreement rather than picking a side silently.
5. If the vault has nothing on the topic, say so plainly — do not answer from outside the vault without saying that is what is happening.
6. If the answer is substantive — a comparison, an analysis, a decision rationale — offer to file it in `wiki/syntheses/`. If accepted, follow the protocol's write procedure: write the synthesis page with the vault's frontmatter conventions plus wikilinks, update `wiki/index.md`, append the log entry, and commit (label `query:`).
7. Answers need not be plain prose: when the question calls for it, the filed page can hold a comparison table or a Marp slide deck (`marp: true` frontmatter, `---` slide separators — renders in Obsidian's Marp plugin). A rich-output answer still gets filed the same way as any synthesis.

## Style

Keep answers grounded in what the wiki actually contains. Say plainly when it has nothing on the topic, and offer to ingest a new source instead of guessing.
