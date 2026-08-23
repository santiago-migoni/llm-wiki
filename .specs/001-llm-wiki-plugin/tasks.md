# Tasks: LLM Wiki Plugin

| Name            | Code      | Version | Date       | Status |
| ---             | ---       | ---     | ---        | ---    |
| llm-wiki-plugin | TASKS-001 | R01     | 2026-08-23 | Approved |

## Phase 1: Setup

- [x] T001 [setup] Commit the working-tree deletion of every v1 file (`skills/`, `scripts/`, `README.md`, `CHANGELOG.md`, `.gitignore`) as one `chore:` commit, so the branch starts from a recorded clean slate
- [x] T002 [P][setup] Commit `docs/llm-wiki.md`, currently untracked, as the source pattern this plugin implements
- [x] T003 [P][setup] Write `.gitignore` with `.DS_Store` only — no build artefacts exist to ignore
- [x] T004 [setup] Create the directory skeleton from plan.md: `references/`, `skills/wiki-init/assets/`, `skills/wiki-ingest/references/`, `skills/wiki-query/`, `skills/wiki-lint/`
- [x] T005 [P][setup] Set `.claude-plugin/plugin.json` to version `2.0.0` with a description naming the four skills
- [x] T006 [P][setup] Set `.claude-plugin/marketplace.json` to version `2.0.0` with a matching description

## Phase 2: Shared Protocol (US5, US6) (P1)

- [x] T007 [US6] Write the vault-resolution section of `references/vault-protocol.md`: resolve an absolute vault root once, ask when more than one candidate exists, never treat the working directory as the vault
- [x] T008 [US5] Add the schema-authority section to `references/vault-protocol.md`: read the vault's schema file first, follow it wherever it conflicts with a skill, apply plugin defaults only where it is silent
- [x] T009 [US6] Add the path-lanes section to `references/vault-protocol.md`: file tools for pages, index, and log; shell for git and read-only queries; re-resolve shell-reported paths against the vault root before any file-tool call
- [x] T010 [US6] Add the git procedure to `references/vault-protocol.md`: probe with `git -C <vault> rev-parse --git-dir`, stage named paths only, commit labels, and the write-without-commit degradation when the probe fails
- [x] T011 [US5] Add the retrieval rule to `references/vault-protocol.md`: read `wiki/index.md` before opening any page, and read only the pages the task needs — the constitution's index-first Performance MUST
- [x] T012 [US5] Add the write procedure to `references/vault-protocol.md`: update `wiki/index.md`, append one `## [YYYY-MM-DD] <op> | <subject>` entry to `wiki/log.md`, commit once at the end of the operation, then report what actually changed — pages created, pages updated, contradictions found — rather than announcing success generically
- [x] T013 [TEST][US6] Verify `references/vault-protocol.md` contains no `cd` and no bare relative vault path, and that every documented command uses `-C <vault>` or an absolute path
- [x] T014 [TEST][US5] Run a skill against a scratch vault whose schema file specifies non-default directory names and a different log format, and verify the vault's conventions win over the plugin's defaults

## Phase 3: Create a Vault (US1) (P1)

- [x] T015 [US1] Write `skills/wiki-init/assets/vault-template.md`: the schema file written into the user's vault, with placeholders for scope, categories, and user name, carrying conventions, workflows, log format, and commit labels
- [x] T016 [US1] Write `skills/wiki-init/SKILL.md`: refuse if already a vault, ask scope and git, scaffold directories tailored to scope, seed `index.md`/`log.md`/`overview.md`, initialize git
- [x] T017 [TEST][US1] Run `wiki-init` against a scratch folder and verify the schema file, `wiki/index.md`, `wiki/log.md`, `raw/`, scope-tailored categories, and one `schema:` commit all exist
- [x] T018 [TEST][US1] Run `wiki-init` against a folder that already contains a `wiki/` directory and verify it refuses, changes nothing, and offers a lint

## Phase 4: File a Source (US2) (P1)

- [x] T019 [P][US2] Write `skills/wiki-ingest/references/converting-documents.md`: per-format conversion for pdf, docx, xlsx, pptx, html, and images, each probing for its converter and degrading with a stated alternative
- [x] T020 [US2] Write the core flow of `skills/wiki-ingest/SKILL.md`: read the source, discuss takeaways, write the source page with `source:`, update affected pages, index, log, and commit
- [x] T021 [US2] Add source triage to `skills/wiki-ingest/SKILL.md`: `git -C <vault> status --short raw/` with the `source:`-comparison fallback for non-git vaults; an already-cited source prompts before re-ingesting; a file outside `raw/` prompts to stage it there first
- [x] T022 [US2] Add the contradiction rule and batch mode to `skills/wiki-ingest/SKILL.md`; if the file exceeds 1,000 words, move batch mode to `skills/wiki-ingest/references/batch-ingest.md` per plan.md's stated fallback
- [x] T023 [TEST][US2] Ingest a markdown source into the scratch vault and verify the source page names the raw file, the index lists new pages, the log gains one entry, the commit contains only the touched paths plus the raw file, and the original under `raw/` is byte-identical afterwards
- [x] T024 [TEST][US2] Ingest a source whose body contains instructions addressed to the LLM and verify they are quoted to the user and not acted on
- [x] T025 [TEST][US2] Ingest a source contradicting an existing page and verify both claims survive and the log entry names the contradiction
- [x] T026 [TEST][US2] Attempt to ingest a PDF with no converter available and verify the skill reports what is missing, suggests how to convert, and leaves the other skills usable

## Phase 5: Ask the Wiki (US3) (P1)

- [x] T027 [US3] Write `skills/wiki-query/SKILL.md`: index-first shortlisting, `grep -ril` fallback, citations to pages, surfacing disagreement, and offering to file the answer under `wiki/syntheses/`
- [x] T028 [US3] Add output forms and context budgeting to `skills/wiki-query/SKILL.md`: a filed answer may be a comparison table or a Marp deck; when a question spans more pages than fit in context, read the most relevant first and state which pages were not read
- [x] T029 [TEST][US3] Ask the scratch vault a covered question and verify the answer cites the pages it used
- [x] T030 [TEST][US3] Ask the scratch vault an uncovered question and verify Claude says the wiki does not cover it rather than answering from outside it

## Phase 6: Health-Check (US4) (P2)

- [x] T031 [US4] Write `skills/wiki-lint/SKILL.md`: broken links, index-versus-disk drift, contradictions, stale claims, pages that outgrew one topic, the scale-ceiling count, and findings grouped by severity with nothing applied before approval; orphan pages are referred to Obsidian's graph view rather than computed
- [x] T032 [US4] Add the schema-amendment path to `skills/wiki-lint/SKILL.md`: when a convention is not working, propose amending the vault's schema file and commit it with a `schema:` label
- [x] T033 [TEST][US4] Run `wiki-lint` on the scratch vault and verify it reports findings and changes nothing until the user approves a subset

## Phase 7: Packaging

- [x] T034 [P] Write `README.md`: what the plugin is, the four skills, install, the documented scale ceiling, and credit to `docs/llm-wiki.md` as the source pattern
- [x] T035 [P] Write `CHANGELOG.md` with a `2.0.0` entry naming what was removed — five skills and the graph script — and why
- [x] T036 Verify trigger separation by reading the four skill descriptions together: no shared claim phrase, each naming its boundary against its nearest neighbour

## Verification

- [x] VERIFY All acceptance scenarios in spec.md pass
- [x] VERIFY All Non-Functional Requirements in spec.md are met
- [x] VERIFY No constitution MUST principle relevant to this feature is violated
- [x] VERIFY No files were created that are not listed in plan.md's file structure
- [x] VERIFY No new dependencies were added beyond those listed in plan.md
- [x] VERIFY `find . -name '*.py' -o -name '*.sh' -o -name '*.js'` returns nothing outside `.git/` — the spec's zero-executable-surface metric
- [x] VERIFY The four `SKILL.md` files total under 4,000 words by `wc -w`, each under 1,000
- [x] VERIFY An end-to-end scratch-vault run (init, ingest, query, lint) required no hand-editing of any file under `wiki/`
- [x] VERIFY Round-trip integrity: `git log --oneline` on the scratch vault shows one labelled commit per operation, and every claim on a page traces to a source page naming a file in `raw/`
- [x] VERIFY Every writing operation reported what actually changed — pages created, pages updated, contradictions found — not a generic success message
- [x] VERIFY An operation interrupted before its commit leaves changes that `git status` and `git diff` fully describe
