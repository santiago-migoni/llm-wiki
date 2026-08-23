# llm-wiki

A Claude Code plugin that turns a folder into a knowledge base the LLM writes and maintains: you curate sources and ask questions, the LLM does the summarizing, cross-referencing, filing, and bookkeeping.

It implements the pattern described in [docs/llm-wiki.md](docs/llm-wiki.md) — the idea file this plugin is built from, credited here rather than claimed as original. Three layers: immutable raw sources, an LLM-owned wiki, and a per-vault schema file (`CLAUDE.md` or `AGENTS.md`) that governs everything else and that skills always defer to.

## Skills

| Skill | Purpose |
|---|---|
| `wiki-init` | Turn a folder into a working vault: directory layout, a tailored schema file, index/log/overview, git repo. |
| `wiki-ingest` | File a source from `raw/` into the wiki, updating every page it touches. Converts `.docx`/`.xlsx`/`.pdf` first when a converter is available; batch mode for several sources at once. |
| `wiki-query` | Answer questions from the wiki with citations; file substantive answers back as syntheses. |
| `wiki-lint` | Health check: broken links, contradictions, stale claims, index drift, pending sources — reports findings, fixes only what's approved. |

## Install

Add this repository as a plugin marketplace and install `llm-wiki`, in Claude Code or in Cowork. Both hosts are supported identically — see [Scale and hosts](#scale-and-hosts) below for what that depends on.

## Memory model

There is no search engine, no graph tooling, and no retrieval-hook frontmatter. `wiki/index.md` is the entire retrieval layer: one line per page, read before any page is opened. This holds at roughly 100 sources and a few hundred pages — `wiki-lint` says so explicitly when a vault crosses that ceiling, rather than silently degrading.

## Scale and hosts

Runs identically under Claude Code and Cowork. Every skill resolves the vault as an absolute path rather than assuming the working directory is the vault, writes pages with file tools rather than shell redirection, and probes for `git` before committing — degrading to "written, not committed" when it's unavailable instead of failing outright.

## Conventions assumed as defaults

- `raw/` holds immutable sources; conversions of non-markdown sources go to `raw/extracted/`.
- Wiki pages use YAML frontmatter, kebab-case filenames, and liberal wikilinks.
- `wiki/index.md` catalogs pages; `wiki/log.md` is an append-only activity log.
- The vault is a git repo; every write is committed atomically, with only the paths it touched.

All of these yield to whatever the vault's own schema file specifies — see `wiki-init`'s template at `skills/wiki-init/assets/vault-template.md`.

## Zero dependencies

No scripts, no runtime, no install step beyond the plugin itself. Everything the earlier version of this plugin delegated to a bundled Python tool is now either a documented `grep`/`find`/`git` one-liner, inside a skill, or isn't done at all — see `docs/llm-wiki.md`'s "Optional: CLI tools" section and this repository's `.specs/constitution.md` for what was deliberately left out and why.
