---
name: wiki-lint
description: Health-check an LLM-maintained wiki vault - find contradictions, stale claims, orphan candidates, broken links, and gaps. Use when the user says "lint the wiki", "health check", "clean up the wiki", or after roughly every 10 ingests. Distinct from wiki-query; lint reports and, on approval, fixes, but it never answers a question.
---

# Wiki lint

Audit the wiki's structural and factual health; fix only what the user approves.

Load the plugin's `references/vault-protocol.md` first; its vault-resolution, schema-authority, path-lane, git, and write-procedure rules govern every step below. Resolve that reference from the plugin package, not from the current working directory.

## Steps

1. Resolve the vault per the protocol; read its schema file. The vault's schema overrides this skill where they conflict.
2. Run the deterministic checks, all read-only:
   - **Broken links** — collect every `[[wikilink]]` across `wiki/` with `grep -roh '\[\[[^]|#]*' <vault-root>/wiki/ | sed 's/\[\[//'`, then compare each target against the filename stems under `wiki/`. A target with no matching file is a page worth creating — a link referenced from three or more pages is a concept begging for one.
   - **Index drift** — compare the pages `wiki/index.md` lists against the actual `.md` files under `wiki/`, excluding `index.md`, `log.md`, and `overview.md` themselves. Report any file missing from the index and any index entry with no file.
   - **Orphan candidates** — identify wiki pages that receive no inbound wikilink from another content page. Ignore `wiki/index.md` as an inbound link source, because the catalog links to everything by design. Report these as candidates for linking or review, not automatic deletion targets.
   - **Scale ceiling** — count archived originals under `raw/archive/` separately from Markdown pages under `wiki/`. Past roughly 100 sources or a few hundred pages, the index-first approach this plugin relies on stops holding: say so explicitly and point at proper search tooling, rather than reporting health as if the vault were still small.
   - **Pending sources** — `git -C <vault-root> status --short raw/inbox/` if the vault is a repository, otherwise list `raw/inbox/` and compare it against every page's `document-id:`/`revision-id:` or `source:` field. Files in `raw/archive/`, `raw/extracted/`, and `raw/assets/` are not pending sources.
   - **Archive integrity** — for every `raw/archive/<document-id>/<revision-id>/`, confirm there is exactly one original, a `manifest.md`, a corresponding `wiki/sources/<document-id>--<revision-id>.md`, and a complete `raw/extracted/<document-id>/<revision-id>.md`. Recompute the recorded SHA-256 and report mismatches, duplicate revision IDs, duplicate checksums, broken `supersedes`/`superseded-by` links, and archive records with no source page.
   - **Catalog integrity** — compare `raw/catalog.md` with the archive manifests and stable pages in `wiki/knowledge/`. Report logical documents missing from the catalog, catalog entries pointing to nonexistent paths, stale `current-revision` values, and stable documents with no current revision.
3. Read the wiki pages, stable knowledge records, and source records — or a representative sample if the vault is large, prioritizing well-linked pages and anything `wiki/log.md` shows as recently touched — checking for: contradictions between pages, claims a newer revision has superseded, `updated:` frontmatter that looks stale relative to newer sibling pages, near-duplicate pages that should merge, and a page that has outgrown one coherent topic and should split.
4. Check `wiki/index.md`'s completeness against the file tree (from step 2), and `wiki/overview.md` against the wiki's current state.
5. Suggest gaps worth filling: important concepts mentioned across multiple pages but lacking a page, questions the wiki raises but can't answer, missing cross-references, and sources worth a web search.
6. Report every finding grouped by severity. Change nothing until the user responds.
7. Apply only the fixes the user approves. Follow the protocol's write procedure to close out: update the index if anything moved or merged, append one log entry summarizing findings and fixes (label `lint:`), commit, and report what was actually changed.

## Schema amendments

When a convention in the vault's schema genuinely isn't working — a category that never got used, a frontmatter field nobody fills in, a log format that's awkward to grep — propose amending the schema file itself rather than working around it silently. On approval, edit the schema file and commit it with a `schema:` label, separately from any `lint:` commit in the same session, so the two kinds of change stay auditable on their own.
