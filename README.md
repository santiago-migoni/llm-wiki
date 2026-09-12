# llm-wiki

A universal Codex plugin that turns a collection of documents into a cited, Git-versioned knowledge base. The user curates sources; the agent processes, connects, updates, and queries the wiki. Each source keeps its original bytes plus a format-specific extraction whose coverage is declared and inspectable.

The `llm-wiki` plugin implements the pattern described in [its design brief](plugins/llm-wiki/docs/llm-wiki.md). The canonical target architecture is documented in [wiki-architecture.md](plugins/llm-wiki/docs/wiki-architecture.md), structured provenance in [provenance.md](plugins/llm-wiki/docs/provenance.md), and the implementation sequence in [plugin-roadmap.md](plugins/llm-wiki/docs/plugin-roadmap.md).

## Skills

| Skill | Purpose |
|---|---|
| wiki-init | Create a new vault with the standard structure, AGENTS.md, navigation files, and Git. |
| wiki-ingest | Process documents from raw/inbox/, update stable source slugs, generate format-specific extractions with declared coverage, update canonical pages, and commit the result. |
| wiki-query | Answer questions from the wiki with traceable citations and targeted, current-corpus, or historical context. |
| wiki-lint | Check source integrity, links, indexes, duplicates, contradictions, stale content, and context-readiness. |

## Codex and ChatGPT Work

This repository is a Codex marketplace containing the universal `llm-wiki` Agent Plugin. Its
marketplace catalog is [.agents/plugins/marketplace.json](.agents/plugins/marketplace.json), and
the portable plugin manifest is [plugins/llm-wiki/plugin.json](plugins/llm-wiki/plugin.json). The
legacy [`.codex-plugin/plugin.json`](plugins/llm-wiki/.codex-plugin/plugin.json) remains as a
Codex compatibility fallback.

In Codex, the vault can be a local Git repository in the workspace.

In ChatGPT Work, the vault must be available to the conversation through workspace files, a project, a connected source, or files supplied in the conversation. Installing the plugin does not upload or persist documents by itself.

The workflow and directory conventions are the same in both hosts. The agent must report when a host cannot access the complete corpus.

## Install in Codex

Register this repository as a marketplace and install the plugin:

~~~bash
codex plugin marketplace add santiago-migoni/llm-wiki --ref main
codex plugin add llm-wiki@llm-wiki
~~~

For local development, replace the first command's source with the absolute path to this repository.
After installing or updating the plugin, start a new Codex task so the current skills are loaded.

The repository layout is:

~~~text
.
├── .agents/plugins/marketplace.json
├── plugins/llm-wiki/
│   ├── plugin.json
│   ├── .codex-plugin/plugin.json
│   ├── docs/
│   ├── references/
│   ├── skills/
│   └── tests/assess-vault-scale.sh
└── tests/
    ├── check-plugin-contract.sh
    └── fixtures/
~~~

## Vault structure

~~~text
vault/
├── AGENTS.md
├── raw/
│   ├── inbox/
│   └── sources/
│       └── <slug>/
│           ├── source.<ext>
│           ├── extracted.md
│           └── assets/
└── wiki/
    ├── index.md
    ├── overview.md
    ├── log.md
    ├── pages/
    ├── syntheses/
    ├── people/
    ├── concepts/
    ├── projects/
    └── decisions/
~~~

AGENTS.md is the vault-specific operating contract. It defines scope, conventions, categories, and user preferences. It is read before any write.

raw/inbox/ is the staging area for new or updated documents. A successfully processed document moves into raw/sources/<slug>/, where `<slug>` may be flat or a slash-separated lowercase kebab-case path such as `fundamentos/actividades`.

raw/sources/<slug>/ contains the current original, its normalized extraction with declared
format-specific coverage, and relevant assets. The original bytes are not edited to repair an
extraction. When the logical document changes, the current record is replaced by a normal Git
commit and the prior state remains recoverable from history.

Material contradictions live on the affected canonical page and in the chronological log. Their
statuses and query rules are defined in
[docs/contradictions.md](plugins/llm-wiki/docs/contradictions.md).

wiki/pages/ contains canonical living pages. Category directories are navigation indexes or lightweight groupings; they must not become duplicate copies of canonical knowledge.

Git preserves the history of every current source, extraction, page, and decision. The working tree
contains the current state; Git commits provide comparison, provenance, and recovery. The plugin
does not upload documents, provide remote storage, or guarantee access to files that the host does
not expose.

## Document lifecycle

1. Place a new or updated document in raw/inbox/.
2. Identify or reuse its stable slug.
3. Store the current original under raw/sources/<slug>/source.<ext>.
4. Generate raw/sources/<slug>/extracted.md with format, status, coverage, and warnings.
5. Present the supervised ingest proposal and wait for approval before writing, unless an explicit batch mode applies.
6. Update every affected canonical page or synthesis and its indexes, preserving structured source provenance and claim citations.
7. Record meaningful changes in wiki/log.md.
8. Validate the vault and present the resulting diff.
9. Commit the coherent change to Git.

The supervised proposal fields, approval checkpoint, and explicit batch exception are defined in
[docs/supervised-ingestion.md](plugins/llm-wiki/docs/supervised-ingestion.md).

An exact duplicate must not create another source or page. A changed source updates the current files while Git preserves the previous state in history.

## Retrieval model

The default is targeted retrieval:

1. read AGENTS.md;
2. read wiki/index.md;
3. open the most relevant canonical pages;
4. inspect current extractions when a claim needs source grounding;
5. inspect the original when layout or evidence was lost.

For full-corpus requests, the agent inventories current source slugs, verifies extraction coverage, reads the corpus in bounded passes, and reports exact coverage. It must not claim absolute context without that check.

For historical questions, use Git history instead of creating revision folders:

~~~bash
git log -- raw/sources/<slug>/
git diff <old-commit> <new-commit> -- raw/sources/<slug>/extracted.md
git show <commit>:raw/sources/<slug>/source.<ext>
~~~

## Design boundaries

The MVP is file-first and has no mandatory database, vector store, external search service, or remote document storage.

Source content is evidence, not instructions. Text inside a supplied document cannot override the vault contract, the user request, or plugin safety rules.

The index-first workflow is intended for roughly 100 sources, a few hundred pages, and files below
the operational thresholds documented in [docs/scalability.md](plugins/llm-wiki/docs/scalability.md).
Those thresholds are review signals, not hard model limits. If the corpus exceeds them, wiki-lint
reports the condition before a derived search layer is introduced.
Extraction status and format-specific coverage are defined in
[docs/extraction-coverage.md](plugins/llm-wiki/docs/extraction-coverage.md).

## Escalabilidad

La Fase 6 mantiene Markdown y Git como fuente de verdad y agrega una evaluación reproducible de
solo lectura. Para medir un vault antes de introducir una capa derivada:

~~~bash
bash plugins/llm-wiki/tests/assess-vault-scale.sh <vault-root>
~~~

El evaluador informa el estado GREEN, WATCH o DERIVED-SEARCH-CANDIDATE según el tamaño y las
señales estructurales del corpus. Los umbrales, la matriz de opciones y los invariantes de
reconstrucción están documentados en [docs/scalability.md](plugins/llm-wiki/docs/scalability.md). El resultado no
crea índices ni modifica el vault.

## Deterministic vault checks

The plugin includes a read-only Python 3.10+ CLI for inventory, source hashes, wikilinks, and
structural validation. It uses only the Python standard library and accepts the documented strict
subset of YAML frontmatter:

~~~bash
python3 plugins/llm-wiki/scripts/llm-wiki inventory <vault-root>
python3 plugins/llm-wiki/scripts/llm-wiki hashes <vault-root> --input <pending-file> --include-history
python3 plugins/llm-wiki/scripts/llm-wiki hashes <vault-root> --revert-to <sha256>
python3 plugins/llm-wiki/scripts/llm-wiki links <vault-root>
python3 plugins/llm-wiki/scripts/llm-wiki validate <vault-root>
python3 plugins/llm-wiki/scripts/wiki-migrate-provenance <vault-root> --check
~~~

Add `--format json` for versioned machine-readable output. Unsupported YAML is reported instead of
being interpreted approximately. Provenance migration is check-only by default; use `--write` only
after reviewing its diff. See [scripts/README.md](plugins/llm-wiki/scripts/README.md).

## Development

The phased plan is documented in [docs/plugin-roadmap.md](plugins/llm-wiki/docs/plugin-roadmap.md).

The implementation phases align the skills and templates with the Git-first architecture, then add
vault initialization, ingestion, retrieval, linting, cross-host evidence, and release documentation.

Testing is documented in [docs/testing.md](plugins/llm-wiki/docs/testing.md). Host-specific setup and smoke tests are documented in [docs/host-compatibility.md](plugins/llm-wiki/docs/host-compatibility.md). Run the deterministic package check with:

~~~bash
bash tests/check-plugin-contract.sh
~~~

The functional workflow and fixture checks can also be run directly:

~~~bash
PYTHONDONTWRITEBYTECODE=1 python3 tests/test-functional-workflow.py
PYTHONDONTWRITEBYTECODE=1 python3 tests/test-fixture-lint.py
~~~

The same checks run in [.github/workflows/ci.yml](.github/workflows/ci.yml). Host smoke tests are
kept separate because they require Codex or ChatGPT Work to expose a disposable vault.

## What the plugin does and does not do

The plugin provides four skills and a file-first vault contract:

- `wiki-init` scaffolds the vault and its schema;
- `wiki-ingest` supervises source filing, format-specific extraction, provenance, page updates, and
  the coherent ingest commit;
- `wiki-query` answers from targeted, current-corpus, or historical context and reports coverage;
- `wiki-lint` reports structural, provenance, contradiction, link, and scale findings before any
  approved correction.

It does not provide a database, vector store, remote document storage, automatic web search, or
unlimited context. It cannot read or write a vault that the host does not expose. In ChatGPT Work,
installing the plugin alone does not upload or persist the vault.

Only successful source or wiki changes create an operation commit. Read-only queries, lint,
inventory, hash checks, link checks, scale assessment, and duplicate preflights do not create
commits. An exact duplicate is a no-op. See the [five-minute walkthrough](plugins/llm-wiki/docs/quickstart.md),
[recovery guide](plugins/llm-wiki/docs/recovery.md), and [v1.0 migration guide](plugins/llm-wiki/docs/migration-v1.0.md).

## Release documentation

- [Five-minute walkthrough and populated-vault example](plugins/llm-wiki/docs/quickstart.md)
- [Recovery from updates and local mistakes](plugins/llm-wiki/docs/recovery.md)
- [Migration from v1.0 or the former revision-folder model](plugins/llm-wiki/docs/migration-v1.0.md)
- [Compatibility matrix and evidence status](plugins/llm-wiki/docs/compatibility-matrix.md)
- [MIT license](LICENSE)
