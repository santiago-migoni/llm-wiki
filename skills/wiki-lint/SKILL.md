---
name: wiki-lint
description: Health-check an LLM-maintained Obsidian wiki vault - find contradictions, stale claims, broken links, and gaps. Use when the user says "lint the wiki", "health check", "clean up the wiki", or after roughly every 10 ingests. Distinct from wiki-query: lint reports and, on approval, fixes; it never answers a question. Orphan pages are not computed here — Obsidian's own graph view already shows them.
---

# Wiki lint

Audit the wiki's structural and factual health; fix only what the user approves.

Load `${CLAUDE_PLUGIN_ROOT}/references/vault-protocol.md` first; its vault-resolution, schema-authority, path-lane, git, and write-procedure rules govern every step below.

## Steps

1. Resolve the vault per the protocol; read its schema file. The vault's schema overrides this skill where they conflict.
2. Run the deterministic checks, all read-only:
   - **Broken links** — collect every `[[wikilink]]` across `wiki/` with `grep -roh '\[\[[^]|#]*' <vault-root>/wiki/ | sed 's/\[\[//'`, then compare each target against the filename stems under `wiki/`. A target with no matching file is a page worth creating — a link referenced from three or more pages is a concept begging for one.
   - **Index drift** — compare the pages `wiki/index.md` lists against the actual `.md` files under `wiki/`, excluding `index.md`, `log.md`, and `overview.md` themselves. Report any file missing from the index and any index entry with no file.
   - **Scale ceiling** — `find <vault-root>/wiki -name '*.md' | wc -l`. Past roughly 100 sources or a few hundred pages, the index-first approach this plugin relies on stops holding: say so explicitly and point at proper search tooling, rather than reporting health as if the vault were still small.
   - **Pending sources** — `git -C <vault-root> status --short raw/` if the vault is a repository, else compare `raw/` against every page's `source:` field. Files staged here are un-ingested.
3. Read the wiki pages — or a representative sample if the vault is large, prioritizing well-linked pages and anything `wiki/log.md` shows as recently touched — checking for: contradictions between pages, claims a newer source has superseded, `updated:` frontmatter that looks stale relative to newer sibling pages, near-duplicate pages that should merge, and a page that has outgrown one coherent topic and should split.
4. Check `wiki/index.md`'s completeness against the file tree (from step 2), and `wiki/overview.md` against the wiki's current state.
5. Suggest gaps worth filling: questions the wiki raises but can't answer, and sources worth a web search.
6. Report every finding grouped by severity. Change nothing until the user responds.
7. Apply only the fixes the user approves. Follow the protocol's write procedure to close out: update the index if anything moved or merged, append one log entry summarizing findings and fixes (label `lint:`), commit, and report what was actually changed.

## Schema amendments

When a convention in the vault's schema genuinely isn't working — a category that never got used, a frontmatter field nobody fills in, a log format that's awkward to grep — propose amending the schema file itself rather than working around it silently. On approval, edit the schema file and commit it with a `schema:` label, separately from any `lint:` commit in the same session, so the two kinds of change stay auditable on their own.
