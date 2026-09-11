# Query context modes

This reference defines how wiki-query selects, reads, and reports context. It is loaded after references/vault-protocol.md when a query needs more than ordinary targeted retrieval.

## Mode selection

Choose the narrowest mode that satisfies the request.

- **Targeted** is the default for a topic, definition, procedure, or ordinary question.
- **Current-corpus** is required for full context, all current documents, exhaustive analysis, or claims about the complete current vault.
- **Historical** is required for changes over time, prior states, provenance, superseded claims, or comparisons between commits.

Do not infer current-corpus mode merely because several pages are relevant. Use it when the user asks for broad or exhaustive coverage.

## Targeted mode

1. Read the authoritative vault schema.
2. Read wiki/index.md.
3. Select the smallest relevant set of canonical pages.
4. Follow links to relevant syntheses, category indexes, and source slugs.
5. Read raw/sources/<slug>/extracted.md when a claim is load-bearing or the canonical page is only a summary.
6. Inspect raw/sources/<slug>/source.<ext> when the question depends on layout, images, signatures, tables, footnotes, or content not faithfully represented in Markdown.
7. Record the pages and every source slug actually used. When a page has a
   `sources` list, inspect the relevant entry's current extraction and hash
   before relying on its claim.

If the index misses the topic, search wiki/ and raw/sources/ read-only. State when the answer required fallback search.

## Current-corpus mode

Use the current source layer, not Git history.

### Inventory

Build a coverage list containing every current source slug under raw/sources/ and every canonical page under wiki/pages/. Also inspect raw/inbox/.

For every source slug, verify:

- exactly one current source.* file;
- extracted.md exists;
- source and extraction paths are readable;
- extraction status is complete or has declared warnings.

Files in raw/inbox/ are pending and are not part of the processed current corpus. If any exist, report that a complete corpus answer is not possible until they are ingested.

### Read

1. Read all canonical pages under wiki/pages/.
2. Read current extracted.md files in bounded passes.
3. Read original source files only when the extraction is incomplete or the question depends on source format.
4. Track each source slug as read, partially read, skipped, unreadable, unsupported, or pending. For every page, preserve the complete `sources` entry set and flag missing, stale, or contradictory provenance.
5. Do not silently truncate an extraction because of context limits. Continue in another bounded pass or report the omission.

### Report

For an exhaustive answer, include a compact coverage statement:

~~~text
Context mode: current-corpus
Current sources: <count>
Canonical pages: <count>
Sources fully read: <count>
Sources partially read: <count>
Sources skipped or unreadable: <count>
Pending inbox files: <count>
Exceptions: <paths and reasons>
~~~

Do not call this complete or exhaustive when any source is skipped, pending, unreadable, or materially partial.

## Historical mode

Use Git as the archive.

1. Identify the stable source slug or canonical page.
2. Inspect its path history with git log --follow or git log --all.
3. Identify the commits relevant to the question.
4. Use git show for prior source, extraction, or page content.
5. Use git diff for comparisons.
6. Cite commit hashes and exact paths in the answer.
7. Report shallow clones, missing objects, or unavailable history.

There are no revision IDs in the current model. A commit and path identify a historical state.

Example operations:

~~~bash
git log --follow -- raw/sources/<slug>/source.pdf
git log --all -- raw/sources/<slug>/
git diff <old-commit> <new-commit> -- raw/sources/<slug>/extracted.md
git show <commit>:raw/sources/<slug>/source.pdf
~~~

If a source changed extension, inspect the directory history and cite the exact path at each commit.

## Citation contract

Citations should let a future agent locate the evidence quickly.

For canonical knowledge, cite each load-bearing claim:

~~~text
[[page-slug]], section: Current understanding
source slug: <source-slug>
source: raw/sources/<source-slug>/extracted.md, section or marker
~~~

For exact source evidence, cite source.<ext> plus a page, slide, sheet, cell, timestamp, or other available marker.

For historical claims, add:

~~~text
commit: <short-hash>
path: raw/sources/<slug>/extracted.md
~~~

For a claim combining multiple sources, include one source slug and locator
per supporting record. The slugs must be declared in the page frontmatter's
`sources` list. Distinguish:

- direct source evidence;
- a current canonical interpretation;
- a cross-source synthesis;
- an external fact not found in the vault.

## Coverage ledger

For any current-corpus or historical query, maintain a private or user-visible ledger:

~~~text
- <slug or page> | status: fully read | evidence: <path and marker>
- <slug or page> | status: partial | reason: <reason>
- <slug or page> | status: skipped | reason: <reason>
~~~

A ledger is required for the agent's claim of complete coverage. It does not need to be persisted unless the user asks for an audit artifact or the answer is filed as a synthesis.

## Filing a synthesis

Only file an answer when it is durable and substantive, and only after offering it to the user.

Before creating a new file, search wiki/syntheses/ and wiki/index.md for an existing synthesis on the same question. Update an existing synthesis when appropriate.

A new synthesis should use the format in [docs/synthesis-format.md](../../../docs/synthesis-format.md):

~~~yaml
---
type: synthesis
slug: <stable-slug>
created: YYYY-MM-DD
updated: YYYY-MM-DD
status: current
context-mode: targeted | current-corpus | historical
sources:
  - slug: <source-slug>
    source: raw/sources/<source-slug>/source.<ext>
    extracted: raw/sources/<source-slug>/extracted.md
    sha256: <64 lowercase hexadecimal characters>
    role: primary | supporting | context | counterpoint
pages: []
commits: []
---
~~~

The body must link to the canonical pages and source paths used. For current-corpus or historical syntheses, include the coverage ledger or a compact equivalent.

After filing:

1. update wiki/index.md;
2. update wiki/overview.md only if the global orientation changed;
3. append wiki/log.md;
4. stage only touched paths;
5. commit with a query: label;
6. report the file, sources, coverage, and commit.
