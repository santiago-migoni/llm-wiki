---
name: wiki-init
description: Turn a folder into a working LLM-maintained document vault. Use when the user says "init a wiki", "set up a new vault", "crea una wiki aquí", or wants to turn a folder into an LLM-wiki knowledge base. Distinct from wiki-ingest, wiki-query, and wiki-lint; this is the only skill that scaffolds a new vault.
---

# Wiki init

Create a minimal, navigable LLM Wiki vault with a stable-source layout, a schema file, navigation files, and optional Git history.

Load references/vault-protocol.md first. Resolve that reference from the plugin package, not from the current working directory.

## Steps

1. Resolve the target folder as an absolute path under the protocol.
2. Run a read-only preflight. If the folder contains AGENTS.md, CLAUDE.md, raw/, raw/inbox/, raw/sources/, wiki/, or wiki/index.md, stop. Say it already looks like a vault or partial vault, change nothing, and offer wiki-lint or an explicit repair plan.
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
   - wiki/index.md with frontmatter, a canonical pages section, a syntheses section, and one section for each category;
   - wiki/overview.md with frontmatter and an empty-vault orientation;
   - wiki/log.md with frontmatter and one entry: ## [YYYY-MM-DD] schema | Initial scaffold.
   Use the initial-file shapes below. Replace the date and vault name; do not leave template placeholders.
7. Add .gitkeep files only to empty directories so the structure survives a clone. Do not create placeholder source or wiki pages.
8. If Git was accepted, probe git -C <vault-root> rev-parse --git-dir:
   - if it succeeds, use the existing repository and do not initialize a nested repository;
   - if it fails, run git -C <vault-root> init;
   - stage only the files created by this operation and commit with schema: initial wiki scaffold.
9. Verify the exact expected tree: AGENTS.md or CLAUDE.md, raw/inbox/, raw/sources/, wiki/index.md, wiki/overview.md, wiki/log.md, wiki/pages/, wiki/syntheses/, and all category directories. Confirm that no archive, catalog, or revision directories were created.
10. Report the absolute vault path, the created structure, whether an existing repository was reused or Git was initialized, and the next action: place a document in raw/inbox/ and run wiki-ingest.

## Initial file shapes

wiki/index.md:

~~~markdown
---
type: index
updated: YYYY-MM-DD
---
# Wiki index

## Canonical pages

- (none yet)

## Syntheses

- (none yet)

## People

- (none yet)

## Concepts

- (none yet)

## Projects

- (none yet)

## Decisions

- (none yet)
~~~

wiki/overview.md:

~~~markdown
---
type: overview
updated: YYYY-MM-DD
---
# Wiki overview

The vault is initialized and contains no ingested sources yet.
~~~

wiki/log.md:

~~~markdown
---
type: log
updated: YYYY-MM-DD
---
# Activity log

## [YYYY-MM-DD] schema | Initial scaffold

Created the Git-first vault structure.
~~~

## Safety and idempotency

- Do not initialize inside an existing vault.
- Do not mutate a partial vault during preflight.
- Do not overwrite unrelated files.
- Do not create archive or revision directories.
- Do not create duplicate category pages.
- If Git is unavailable, finish the file scaffold and report that it was not committed.
