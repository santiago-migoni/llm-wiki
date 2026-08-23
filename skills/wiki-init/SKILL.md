---
name: wiki-init
description: Turn a folder into a working LLM-maintained Obsidian wiki vault. Use when the user says "init a wiki", "set up a new vault", "crea una wiki aquí", or wants to turn a folder into an LLM-wiki knowledge base. Distinct from wiki-ingest, wiki-query, and wiki-lint: this is the only skill that scaffolds a new vault — it refuses if one already exists there.
---

# Wiki init

Turn a folder into a working LLM-wiki vault: directory layout, a tailored schema file, the memory-layer files, and a git repo — ready for the first ingest.

Load `${CLAUDE_PLUGIN_ROOT}/references/vault-protocol.md` first; its vault-resolution and git rules govern this skill too.

## Steps

1. Resolve the target folder as an absolute path per the protocol. If it already contains a schema file (`CLAUDE.md`/`AGENTS.md`) or a `wiki/` directory, stop: say it is already a vault, change nothing, and offer to run `wiki-lint` instead. Existing unrelated files (an Obsidian `.obsidian/` directory, loose notes) are left untouched.
2. Ask two things, skipping anything the user already stated:
   - **Scope** — what will this wiki track (a project, a research topic, a book, personal or professional knowledge)? This tailors the schema's Scope section and the category directories.
   - **Git** — initialize a repository? Default yes.
3. Create the structure with file tools:

   ```text
   raw/assets/          wiki/sources/
   raw/extracted/        wiki/syntheses/
                          wiki/<categories tailored to scope>/
   ```

   `sources/` and `syntheses/` are always created. Categories otherwise follow the stated scope — a literature wiki might get `works/` instead of `organizations/`; a technical wiki might want `decisions/`.
4. Write `CLAUDE.md` at the vault root from `${CLAUDE_PLUGIN_ROOT}/skills/wiki-init/assets/vault-template.md`, filling in the scope paragraph, the category directories, and the user's name.
5. Write the memory-layer files with file tools, each with frontmatter:
   - `wiki/index.md` — one section per category, `(none yet)` placeholders.
   - `wiki/log.md` — one initial entry: `## [YYYY-MM-DD] schema | Initial scaffold` (today's date).
   - `wiki/overview.md` — a stub noting the wiki is empty until sources accumulate.
6. If git was accepted: probe with `git -C <vault-root> rev-parse --git-dir` (expected to fail here — there is no repo yet), then `git -C <vault-root> init`. Stage every file just created by name (there is nothing else to exclude at this point) and commit: `git -C <vault-root> commit -m "schema: initial wiki scaffold"`. If git is unavailable, say the vault was created but not committed, and offer `git init` for later.
7. Verify: confirm the schema file, `wiki/index.md`, `wiki/log.md`, `wiki/overview.md`, `raw/`, and the category directories all exist under the resolved vault root.
8. Close by telling the user the vault is ready: drop a source in `raw/` and ask to ingest it, or open the folder in Obsidian to browse as pages appear.
