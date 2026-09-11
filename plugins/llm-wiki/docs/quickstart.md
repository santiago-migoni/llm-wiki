# Five-minute walkthrough

This walkthrough uses a disposable Git vault. It shows the smallest useful loop: initialize,
ingest with approval, query with citations, update while preserving history, and lint. The plugin
does not upload the vault or infer access to files that the host has not exposed.

## 1. Install and initialize

In Codex, register the repository and install the plugin:

~~~bash
codex plugin marketplace add santiago-migoni/llm-wiki --ref main
codex plugin add llm-wiki@llm-wiki
~~~

Start a new Codex task after installation or an update. Ask the task to initialize a named empty
folder, for example: `Initialize a document wiki in /absolute/path/to/demo-vault`.

The initialization is idempotent and stops when the target already looks like a vault. It creates
the schema and navigation files, optionally initializes Git, and makes a `schema: initial wiki
scaffold` commit. It does not create placeholder sources or knowledge pages.

## 2. Ingest a source

Place `security-policy.md` in `raw/inbox/` and ask the agent to ingest it. Before writing, the
agent proposes the stable slug, SHA-256, extraction format and coverage, affected pages,
contradictions, exact paths, and planned commit. Approve that proposal when it is correct.

The resulting current record is stored at `raw/sources/security-policy/`. The original bytes remain
in `source.md`; `extracted.md` is a derived reading artifact with declared coverage. A complete
status means all expected units were processed; it does not mean that a layout-dependent detail
can never require inspection of the original.

## 3. Query and update

Ask a targeted question such as `What review interval does the security policy require?`. The
answer should cite the canonical page, source slug, and a section or other locator. Ask for full
context only when needed; the agent must inventory current sources, pending inbox files, extraction
coverage, and read scope before claiming exhaustive coverage.

Replace the source with a meaningful update in `raw/inbox/` and approve the update proposal. The
same slug is reused, current files are replaced in one coherent change, and Git preserves the
prior state. Ask `What changed historically?` to use `git log`, `git diff`, and `git show`.

## 4. Lint and inspect

Ask the agent to lint the vault. Lint is report-first and read-only until specific corrections
are approved. The deterministic checks can also be run directly:

~~~bash
python3 plugins/llm-wiki/scripts/llm-wiki validate /absolute/path/to/demo-vault --format json
bash plugins/llm-wiki/tests/assess-vault-scale.sh /absolute/path/to/demo-vault
git -C /absolute/path/to/demo-vault log --oneline --all
~~~

Only a successful source or wiki change creates an operation commit. Queries, lint, inventory,
hash preflights, link checks, scale assessment, and exact duplicate detection are read-only.

## Populated-vault example

This is the expected shape after one source and one canonical page have been ingested. `.git/` is
omitted from the display; it is the archive when Git is available.

~~~text
demo-vault/
├── AGENTS.md
├── raw/
│   ├── inbox/                         # empty after a successful ingest
│   └── sources/
│       └── security-policy/
│           ├── source.md              # current original bytes
│           ├── extracted.md            # normalized text plus coverage metadata
│           └── assets/                 # only when supporting assets exist
└── wiki/
    ├── index.md
    ├── overview.md
    ├── log.md
    ├── pages/
    │   └── security-policy.md
    ├── syntheses/
    ├── people/
    ├── concepts/
    ├── projects/
    └── decisions/
~~~

The extraction header declares what was actually processed:

~~~yaml
---
type: extraction
slug: security-policy
source: raw/sources/security-policy/source.md
original-filename: security-policy.md
sha256: <64 lowercase hexadecimal characters>
extracted: 2026-09-11
format: markdown
method: supervised text reader
status: complete
coverage:
  unit: pages
  expected: 1
  processed: 1
warnings: []
---
~~~

The canonical page preserves the source relationship and cites load-bearing claims:

~~~yaml
---
type: knowledge
slug: security-policy
created: 2026-09-11
updated: 2026-09-11
sources:
  - slug: security-policy
    source: raw/sources/security-policy/source.md
    extracted: raw/sources/security-policy/extracted.md
    sha256: <same current-source hash>
    role: primary
tags: [security]
status: current
---
~~~

~~~markdown
# Security policy

## Current understanding

- Reviews occur quarterly.
  - Evidence: `security-policy`, section “Review cycle”.
~~~

For a larger executable example, see the deterministic [functional fixture](../../tests/fixtures/functional-vault/README.md)
and its [workflow runner](../../tests/test-functional-workflow.py). Those tests prove file and Git
invariants; they do not pretend to be an LLM host interaction.

## Next references

- [Recovery](recovery.md) for accidental edits, wrong ingests, and deleted current files.
- [Migration from v1.0](migration-v1.0.md) for existing vaults and legacy layouts.
- [Extraction coverage](extraction-coverage.md) for format-specific status and locators.
- [Host compatibility](host-compatibility.md) for Codex and ChatGPT Work access boundaries.
