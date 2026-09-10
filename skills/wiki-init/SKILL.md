---
name: wiki-init
description: Turn a folder into a working LLM-maintained document vault. Use when the user says "init a wiki", "set up a new vault", "crea una wiki aquí", or wants to turn a folder into an LLM-wiki knowledge base. Distinct from wiki-ingest, wiki-query, and wiki-lint; this is the only skill that scaffolds a new vault.
---

# Wiki init

Create a minimal, navigable LLM Wiki vault with a stable-source layout, a schema file, navigation files, and optional Git history.

Load references/vault-protocol.md first. Resolve that reference from the plugin package, not from the current working directory.

## Steps

1. Resolve the target folder as an absolute path under the protocol.
2. If the folder already contains AGENTS.md, CLAUDE.md, raw/sources/, or wiki/, stop. Say it already looks like a vault, change nothing, and offer wiki-lint.
3. Ask only what is missing:
   - scope: what the wiki tracks;
   - Git: initialize a repository? Default yes.
4. Create this structure with file tools:

~~~text
raw/
  inbox/
  sources/
wiki/
  pages/
  syntheses/
  people/
  concepts/
  projects/
  decisions/
~~~

5. Write AGENTS.md from assets/vault-template.md, replacing every placeholder. Create CLAUDE.md only when the user explicitly requests a Claude-first vault. Never create both automatically.
6. Write the initial memory files:
   - wiki/index.md with sections for pages, syntheses, people, concepts, projects, and decisions;
   - wiki/overview.md with an empty-vault orientation;
   - wiki/log.md with one entry: ## [YYYY-MM-DD] schema | Initial scaffold.
7. Add .gitkeep files to empty directories so the structure survives a clone.
8. If Git was accepted, probe git -C <vault-root> rev-parse --git-dir. It should fail for a new vault. Run git -C <vault-root> init, stage only the files created by this operation, and commit with schema: initial wiki scaffold.
9. Verify AGENTS.md or CLAUDE.md, raw/inbox/, raw/sources/, wiki/index.md, wiki/overview.md, wiki/log.md, wiki/pages/, wiki/syntheses/, and all category directories.
10. Report the absolute vault path, the created structure, whether Git was initialized, and the next action: place a document in raw/inbox/ and run wiki-ingest.

## Safety and idempotency

- Do not initialize inside an existing vault.
- Do not overwrite unrelated files.
- Do not create archive or revision directories.
- Do not create duplicate category pages.
- If Git is unavailable, finish the file scaffold and report that it was not committed.
