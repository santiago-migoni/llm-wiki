---
name: wiki-init
description: Turn a folder into a working LLM-maintained document vault. Use when the user says "init a wiki", "set up a new vault", "crea una wiki aquí", or wants to turn a folder into an LLM-wiki knowledge base. Distinct from wiki-ingest, wiki-query, and wiki-lint; this is the only skill that scaffolds a new vault — it refuses if one already exists there.
---

# Wiki init

Turn a folder into a working LLM-wiki vault: directory layout, a tailored schema file, the memory-layer files, and a git repo — ready for the first ingest.

Load the plugin's `references/vault-protocol.md` first; its vault-resolution and git rules govern this skill too. Resolve that reference from the plugin package, not from the current working directory.

## Steps

1. Resolve the target folder as an absolute path per the protocol. If it already contains a schema file (`CLAUDE.md`/`AGENTS.md`) or a `wiki/` directory, stop: say it is already a vault, change nothing, and offer to run `wiki-lint` instead. Existing unrelated files (an Obsidian `.obsidian/` directory, loose notes) are left untouched.
2. Ask two things, skipping anything the user already stated:
   - **Scope** — what will this wiki track (a project, a research topic, a book, personal or professional knowledge)? This tailors the schema's Scope section and the category directories.
   - **Git** — initialize a repository? Default yes.
3. Create the structure with file tools:

   ```text
   raw/inbox/                     # Pending documents supplied by the user
   raw/archive/                   # Immutable originals, grouped by document and revision
   raw/extracted/                 # Complete normalized text, grouped by document and revision
   raw/assets/                    # Embedded images and supporting assets
   raw/catalog.md                 # One routing row per logical document
   wiki/sources/                  # One summary page per archived revision
   wiki/knowledge/                # Stable living documents, updated over time
   wiki/syntheses/                # Filed answers and analyses
   wiki/<categories tailored to scope>/
   ```

   `sources/`, `knowledge/`, and `syntheses/` are always created. Categories otherwise follow the stated scope — a literature wiki might get `works/` instead of `organizations/`; a technical wiki might want `decisions/`.
4. Write `AGENTS.md` at the vault root from the plugin's `skills/wiki-init/assets/vault-template.md`, filling in the scope paragraph, the category directories, and the user's name. If the user explicitly requests a Claude-first vault, write `CLAUDE.md` instead. Never create both automatically.
5. Write the memory-layer files with file tools, each with frontmatter:
   - `raw/catalog.md` — an empty source catalog with columns for `document-id`, current `revision-id`, status, topics, and paths.
   - `wiki/index.md` — one section per category, `(none yet)` placeholders, plus `sources/` and `knowledge/` sections.
   - `wiki/log.md` — one initial entry: `## [YYYY-MM-DD] schema | Initial scaffold` (today's date).
   - `wiki/overview.md` — a stub noting the wiki is empty until sources accumulate.
   - Add `.gitkeep` files to empty archive directories so the layout survives a clone.
6. If git was accepted: probe with `git -C <vault-root> rev-parse --git-dir` (expected to fail here — there is no repo yet), then `git -C <vault-root> init`. Stage every file just created by name (there is nothing else to exclude at this point) and commit: `git -C <vault-root> commit -m "schema: initial wiki scaffold"`. If git is unavailable, say the vault was created but not committed, and offer `git init` for later.
7. Verify: confirm the schema file, `raw/catalog.md`, `wiki/index.md`, `wiki/log.md`, `wiki/overview.md`, `raw/inbox/`, `raw/archive/`, `raw/extracted/`, `raw/assets/`, `wiki/sources/`, `wiki/knowledge/`, and the category directories all exist under the resolved vault root.
8. Close by telling the user the vault is ready: drop a document in `raw/inbox/` and ask to ingest it, or open the folder in a file browser to browse the archive and wiki as they appear.
