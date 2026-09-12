# Migration from v1.0

This guide separates two cases that are often called “v1.0”. The `1.0.0` package in this
repository already uses the stable-source Git-first layout, so installing a later package does
not rewrite an existing vault. An older nine-skill release or a revision-folder vault needs the
explicit migration below.

## Before migration

1. Make a backup or clone of the vault.
2. Create a migration branch and record the current commit.
3. Run read-only inventory and lint.
4. Do not ingest new files until the source identity and current tree are understood.

~~~bash
git switch -c migrate-llm-wiki-v1
python3 plugins/llm-wiki/scripts/llm-wiki inventory /absolute/path/to/vault --format json
python3 plugins/llm-wiki/scripts/llm-wiki validate /absolute/path/to/vault --format json
~~~

## Existing v1.0 Git-first vault

If the vault already has `raw/sources/<slug>/source.<ext>`, `extracted.md`, `wiki/pages/`, and a
single authoritative `AGENTS.md`, keep the content in place. Install the new plugin, review the
release notes, run lint, and migrate only metadata that the validator identifies. No source bytes
need to be copied merely because the plugin version changed.

## Legacy revision-folder vault

For layouts using `raw/archive/`, per-revision source pages, or revision IDs:

1. Map every logical document to one stable lowercase kebab-case slug, optionally composed of
   slash-separated namespace segments.
2. Select the current original and place it at `raw/sources/<slug>/source.<ext>`.
3. Generate `raw/sources/<slug>/extracted.md` with the format, method, status, coverage, and
   warnings required by [extraction coverage](extraction-coverage.md).
4. Move or create one canonical page at `wiki/pages/<slug>.md` only after reviewing duplicates,
   provenance, and contradictions.
5. Preserve the old revision tree until the new commit has been reviewed and verified. Git history
   is the archive; do not silently delete evidence as part of a bulk migration.
6. Update `wiki/index.md`, `wiki/overview.md`, and `wiki/log.md` with the migration scope.

The old filename, upload date, hash, or revision number must not become the stable identity. When
the same logical document is updated later, reuse its slug and let Git preserve the previous state.

## Singular provenance metadata

Pages with legacy `source`, `extracted`, and `sha256` fields remain readable. The safe migration
tool is check-only by default:

~~~bash
python3 plugins/llm-wiki/scripts/llm-wiki migrate-provenance /absolute/path/to/vault --check
~~~

Review the unified diff. Apply only the approved frontmatter changes with:

~~~bash
python3 plugins/llm-wiki/scripts/llm-wiki migrate-provenance /absolute/path/to/vault --write
~~~

The write mode changes reviewed page metadata, preserves the page body, creates no commit, and
must be followed by validation and an explicit migration commit.

## Verification and rollback

After migration, run the contract, fixture, and vault checks; inspect the exact staged paths; and
commit the migration separately from later ingests. If the result is wrong, use the [recovery
guide](recovery.md) and a corrective commit or revert. Never use migration as a reason to rewrite
Git history.
