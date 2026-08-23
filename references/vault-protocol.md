# Vault protocol

Every skill in this plugin loads this file first, before reading or writing anything. It is the contract all four skills share, so it is written once here instead of four times.

## Resolve the vault

A vault is a folder containing a schema file (`CLAUDE.md` or `AGENTS.md`) that describes a wiki, plus a `wiki/` directory.

- Resolve the vault's root as an absolute path before doing anything else. Never treat the current working directory as the vault — on some hosts it is scratch space, not the vault, even when a vault is open.
- If the user named a folder, resolve it to an absolute path and confirm it looks like a vault (or, for `wiki-init`, confirm it does not yet).
- If the user did not name a folder and more than one candidate vault is reachable, ask which one before reading or writing anything.
- Once resolved, derive every other path in the operation from this root — never a bare relative path, never a path assumed from context.

## The schema is the authority

Read the vault's schema file (`CLAUDE.md` or `AGENTS.md`) in full before writing anything. It is the authority over every skill in this plugin.

- Where the schema and a skill conflict — directory names, frontmatter fields, log format, commit labels, anything — follow the schema. A skill's instructions are defaults, not requirements.
- Where the schema is silent on something a skill needs, apply the skill's default and proceed.
- This holds for vaults this plugin did not create. A vault with its own conventions is followed on its own terms, not migrated toward this plugin's defaults.

## Two lanes: file tools and shell

The two hosts this plugin runs on — Claude Code and Cowork — can report different paths for the same file between their shell and their file tools. Cowork's shell is an isolated sandbox that reaches the user's folders through mounts; its working directory is temporary scratch space, not the vault.

- Read and write wiki pages, `wiki/index.md`, and `wiki/log.md` with file tools (Read/Write/Edit), not shell redirection.
- Reserve the shell for `git` and for read-only queries (`grep`, `find`, `git status`, `git log`).
- Never carry a path a shell command reported straight into a file-tool call. Re-resolve it against the vault root first — a shell path and a file-tool path for the same file are not guaranteed to match.
- Every shell command in this plugin's skills is invoked with `-C <vault-root>` or an absolute path, never by changing directory (`cd`) first.

## Git

Before any commit, probe with `git -C <vault-root> rev-parse --git-dir`. This answers two questions at once: whether `git` is available on this host, and whether the vault is a repository.

- **Probe succeeds:** proceed with the commit below.
- **Probe fails:** the write to the vault's files still happened — say so, then say the change was not committed and why (no `git`, or the folder is not a repository). For "not a repository," offer `git init`. Never block the rest of the operation on a missing probe.

When the probe succeeds:

- Stage only the exact paths this operation touched, by name. Never `git add -A` or `git add .` — an unrelated file sitting uncommitted in `raw/` must never be swept into an unrelated commit.
- Commit once, at the end of the operation, with a label matching the operation: `schema:`, `ingest:`, `query:`, `lint:`.
- A source file entering `raw/` is committed for the first time by the ingest that files it. Until then it stays uncommitted — `git -C <vault-root> status --short raw/` is how an ingest finds what is waiting.

## Read the index first

Read `wiki/index.md` before opening any page. It is the retrieval layer — a one-line hook per page is enough to shortlist what is relevant, so finding the right page costs one file read instead of a scan of the vault.

- Read only the pages a task actually needs. A vault is expected to grow larger than fits in context; reading everything defeats the point of the index.
- This holds at roughly 100 sources and a few hundred pages, per the vault's schema. Past that, `wiki-lint` says so rather than silently degrading — see `skills/wiki-lint/SKILL.md`.

## Every write ends the same way

Any operation that changes the vault — ingest, a filed synthesis, a lint fix, a schema amendment — finishes with the same four steps, in order:

1. Update `wiki/index.md` so every page created or materially changed has a current one-line entry.
2. Append one entry to `wiki/log.md`, dated and prefix-parseable: `## [YYYY-MM-DD] <op> | <subject>`.
3. Commit once, per the git procedure above.
4. Report what actually changed — which pages were created, which were updated, and any contradiction found — rather than a generic confirmation. If the commit step degraded (no git), say so here too.

Do not commit partway through a write. If an operation is interrupted before step 3, the vault is left with uncommitted changes that `git status` and `git diff` fully describe — never a half-written page committed as if it were finished.
