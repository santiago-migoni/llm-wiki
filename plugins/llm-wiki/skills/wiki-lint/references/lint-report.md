# Wiki lint checks and report

This reference defines the checks and report contract for wiki-lint. It is loaded after references/vault-protocol.md.

## Baseline

Run the audit against the resolved vault root and the current working tree.

An empty vault is valid when:

- the schema exists;
- the standard directories exist;
- wiki/index.md, wiki/overview.md, and wiki/log.md exist;
- raw/inbox/ is empty;
- raw/sources/ contains no source records;
- wiki/pages/ contains no canonical pages.

Do not report missing knowledge as a structural failure in an empty vault.

## Deterministic checks

### Pending inbox

List raw/inbox/ including untracked files. In a Git vault use:

~~~bash
git -C <vault-root> status --short --untracked-files=all -- raw/inbox/
~~~

Every file waiting there is pending. A pending file makes a full-corpus query incomplete, but it is not itself a lint failure when the user intentionally staged it.

### Current source structure

For every raw/sources/<slug>/:

- confirm the slug is lowercase kebab-case;
- confirm exactly one file matches source.*;
- confirm extracted.md exists;
- confirm assets/ exists when the extraction references local assets;
- report unexpected files separately;
- confirm no revision directories or duplicate source files exist.

A source record without an expected canonical page is an unfiled-source finding, not automatically a reason to create a page.

### Extraction freshness

For every source record:

- compute the current source hash when a hash utility is available;
- compare it with sha256 in extracted.md and any canonical page metadata;
- confirm extracted.md points to the current source path;
- inspect status and warnings;
- report missing, stale, partial, unreadable, or unsupported extractions.

If the source changed after extraction, classify the extraction as stale even if the file exists.

### Page and index coverage

Compare:

- canonical Markdown files under wiki/pages/;
- synthesis files under wiki/syntheses/;
- routes listed in wiki/index.md.

Report pages without an index route and index routes without a file. Category indexes may be navigational and should not be required to duplicate every canonical page.

### Broken wikilinks

Collect wikilinks from wiki/ and resolve them against the stems of Markdown files under wiki/. Preserve heading and alias suffixes while resolving the page target.

Report:

- missing target;
- ambiguous target with duplicate stems;
- link to a deprecated path;
- link that should point to a canonical page rather than a category copy.

A missing link target is a finding, not permission to create a page automatically.

### Duplicate candidates

Look for:

- more than one canonical page for the same stable slug;
- repeated canonical body content in category indexes;
- near-duplicate titles or pages with overlapping scope;
- source records whose hashes are identical;
- a new page created where an existing canonical page already covers the same unit.

Do not merge or delete pages automatically.

### Source/page coverage

Check that:

- canonical pages' source and extracted paths exist;
- source slugs referenced by pages exist;
- source records expected to become durable knowledge are discoverable from wiki/index.md;
- a page does not silently cite a superseded or missing path;
- synthesis pages list the source slugs or canonical pages they use.

### Metadata consistency

Check fields relevant to each page type:

- type;
- slug;
- created;
- updated;
- source;
- extracted;
- sha256;
- tags;
- status;
- context-mode, sources, pages, and commits for syntheses when applicable.

Report malformed dates, invalid slugs, missing required paths, stale hashes, and contradictory status values.

### Git state and history

Report:

- uncommitted changes in raw/sources/ and wiki/;
- source or page changes mixed with unrelated work;
- whether the current source history is available;
- shallow or missing Git history when historical traceability is requested.

Do not rewrite or repair Git history during lint.

### Legacy layout and scale

Report any occurrence of:

- raw/archive/;
- raw/catalog.md;
- per-revision source or extraction paths;
- revision identifiers used as current identity;
- wiki/knowledge/;
- wiki/sources/ as an operational revision layer.

Count current source records and canonical pages. Past roughly 100 sources or a few hundred pages, report that index-first retrieval may need a derived search layer.

For consistent scale reporting, use the thresholds in docs/scalability.md:

- GREEN: fewer than 80 source records, fewer than 200 canonical pages, no source-layer or wiki
  Markdown file at or above 5 MiB, and less than 50 MiB of wiki Markdown in total;
- WATCH: 80–100 source records, 200–300 canonical pages, a source-layer or wiki Markdown file at
  or above 5 MiB, or 50–249 MiB of wiki Markdown in total, unless the candidate threshold is
  reached;
- DERIVED-SEARCH-CANDIDATE: more than 100 source records, more than 300 canonical pages, or at
  least 250 MiB of wiki Markdown in total.

The numeric status is a decision signal, not a claim that the vault is context-ready. Report
pending files, missing extractions, incomplete access, Git limitations, and observed retrieval
degradation separately. A derived layer still requires evidence and must preserve the current
Markdown/Git tree as its source of truth.

## Semantic checks

Read relevant canonical pages, current extractions, and recent wiki/log.md entries. Check for:

- contradictions between current sources and canonical claims;
- stale claims that an update should revise;
- claims without evidence;
- missing cross-references;
- categories that duplicate canonical content;
- pages containing multiple unrelated topics;
- repeated concepts, people, projects, or decisions without a canonical page;
- unresolved open questions;
- extraction warnings that affect a load-bearing claim.

For large vaults, prioritize recently changed and highly linked material, then state the sample and unexamined scope.

## Severity

- **Blocker** — current context or provenance cannot be trusted.
- **High** — likely to produce materially incorrect retrieval or answer.
- **Medium** — structural or maintenance degradation.
- **Low** — navigational, clarity, or cosmetic improvement.

Severity describes impact, not effort.

## Report format

Use this compact structure:

~~~markdown
# Wiki lint report

- Date: YYYY-MM-DD
- Vault: <absolute path>
- Sources: <count>
- Canonical pages: <count>
- Pending inbox files: <count>
- Context readiness: ready | incomplete | blocked
- Git state: clean | uncommitted | unavailable

## Summary

<Short overall assessment.>

## Findings

### [HIGH] LINT-001 — <short title>

- Paths: <exact paths>
- Evidence: <what was observed>
- Impact: <why it matters>
- Suggested correction: <minimal correction>
- Approval required: yes | no

## Positive checks

- <Important invariant that passed.>

## Coverage and limitations

- Fully checked: <scope>
- Sampled: <scope>
- Not checked: <scope and reason>
~~~

Use stable finding IDs within one report. Do not present a proposed fix as already applied.

## Fix protocol

Lint is report-first.

Without explicit approval, make no changes. When the user approves selected findings:

1. apply only those corrections;
2. do not delete or merge pages unless the approval names the targets;
3. preserve source files, extractions, assets, and Git history;
4. update wiki/index.md for changed routes;
5. update wiki/overview.md only when the global picture changed;
6. append one lint entry to wiki/log.md;
7. validate the affected invariants;
8. commit with a lint: label.

A change to AGENTS.md or another vault convention is a schema amendment. Propose it separately and commit it with schema: rather than hiding it in a lint commit.
