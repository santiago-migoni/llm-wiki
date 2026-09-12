# Recovery guide

Git is the archive for current source records, extractions, pages, and decisions. The working tree
holds only the current state. Recovery should preserve history and make the correction explicit.
The `<slug>` path in the commands below may be flat or hierarchical; use the complete
slash-separated slug for nested records and pages.

## Inspect before changing anything

From the vault root:

~~~bash
git status --short
git log --oneline --decorate --all -20
git diff -- raw/sources/<slug>/ wiki/pages/<slug>.md wiki/log.md
~~~

Do not overwrite local edits created by another operation. Ask for a decision when the working tree
contains uncommitted changes in the affected source or page paths.

## Uncommitted accidental edit

If a file was edited accidentally and the intended state exists in `HEAD`, review the target first
and then restore only that exact path:

~~~bash
git diff -- raw/sources/<slug>/source.<ext>
git restore -- raw/sources/<slug>/source.<ext>
~~~

`git restore` discards the selected uncommitted bytes. Copy or save anything worth keeping before
running it. Never use a broad restore for a vault with unrelated local work.

## Wrong or incomplete committed ingest

Find the operation commit and inspect its paths:

~~~bash
git log --oneline --all -- raw/sources/<slug>/ wiki/pages/<slug>.md
git show --stat <ingest-commit>
git show <ingest-commit> -- raw/sources/<slug>/ wiki/pages/<slug>.md
~~~

If the commit should be undone, use a new corrective commit:

~~~bash
git revert <ingest-commit>
~~~

Resolve any reviewable conflict, validate the vault, and use a clear commit message. Do not amend
or rewrite shared history as part of normal maintenance.

## Recover a previous source state

Git can restore an earlier committed original, extraction, or page into the working tree:

~~~bash
git show <known-good-commit>:raw/sources/<slug>/source.<ext> > /tmp/source-backup.<ext>
git restore --source=<known-good-commit> -- raw/sources/<slug>/source.<ext> raw/sources/<slug>/extracted.md
~~~

Review the diff, regenerate or verify the extraction, update affected provenance and log entries,
then create a new coherent commit. The old commit remains available for audit.

If the file never existed in a commit, Git cannot recover its uncommitted bytes. Check the host's
trash, backup, or file history separately; do not claim recovery from the vault.

## Host limitations

A current-tree export without `.git` cannot answer historical questions. A read-only workspace may
be queried, but the agent must not claim that it persisted a synthesis or committed changes. A host
that exposes only part of the current corpus must report the accessible subset and pending or
unreadable paths.
