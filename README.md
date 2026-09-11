# llm-wiki

A universal Codex plugin that turns a collection of documents into a cited, Git-versioned knowledge base. The user curates sources; the agent processes, connects, updates, and queries the wiki.

The `llm-wiki` plugin implements the pattern described in [its design brief](plugins/llm-wiki/docs/llm-wiki.md). The canonical target architecture is documented in [wiki-architecture.md](plugins/llm-wiki/docs/wiki-architecture.md), structured provenance in [provenance.md](plugins/llm-wiki/docs/provenance.md), and the implementation sequence in [plugin-roadmap.md](plugins/llm-wiki/docs/plugin-roadmap.md).

## Skills

| Skill | Purpose |
|---|---|
| wiki-init | Create a new vault with the standard structure, AGENTS.md, navigation files, and Git. |
| wiki-ingest | Process documents from raw/inbox/, update stable source slugs, generate complete extractions, update canonical pages, and commit the result. |
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

raw/inbox/ is the staging area for new or updated documents. A successfully processed document moves into raw/sources/<slug>/.

raw/sources/<slug>/ contains the current original, its complete normalized extraction, and relevant assets. The slug remains stable when the logical document changes.

wiki/pages/ contains canonical living pages. Category directories are navigation indexes or lightweight groupings; they must not become duplicate copies of canonical knowledge.

Git preserves the history of every current source, extraction, page, and decision. The working tree contains the current state; Git commits provide comparison, provenance, and recovery.

## Document lifecycle

1. Place a new or updated document in raw/inbox/.
2. Identify or reuse its stable slug.
3. Store the current original under raw/sources/<slug>/source.<ext>.
4. Generate raw/sources/<slug>/extracted.md.
5. Update every affected canonical page or synthesis and its indexes, preserving structured source provenance and claim citations.
6. Record meaningful changes in wiki/log.md.
7. Validate the vault.
8. Commit the coherent change to Git.

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

The index-first workflow is intended for roughly 100 sources and a few hundred pages. If the corpus exceeds that scale, wiki-lint should report the limit before a derived search layer is introduced.

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

## Development

The phased plan is documented in [docs/plugin-roadmap.md](plugins/llm-wiki/docs/plugin-roadmap.md).

The first implementation phase aligns all skills and templates with the Git-first architecture. Later phases add vault initialization, ingestion, retrieval, linting, cross-host tests, and optional scale improvements.

Testing is documented in [docs/testing.md](plugins/llm-wiki/docs/testing.md). Host-specific setup and smoke tests are documented in [docs/host-compatibility.md](plugins/llm-wiki/docs/host-compatibility.md). Run the deterministic package check with:

~~~bash
bash tests/check-plugin-contract.sh
~~~
