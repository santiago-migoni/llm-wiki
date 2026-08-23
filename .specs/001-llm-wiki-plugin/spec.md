# Spec: LLM Wiki Plugin

| Name            | Code     | Version | Date       | Status |
| ---             | ---      | ---     | ---        | ---    |
| llm-wiki-plugin | SPEC-001 | R01     | 2026-08-23 | Approved |

## Summary

A Claude Code plugin of four markdown skills — init, ingest, query, lint — that lets a person keep a knowledge base which the LLM writes and maintains, so the human curates sources and asks questions while the LLM does the summarizing, cross-referencing, filing, and bookkeeping.

## Success Metrics

- **Executable surface**: zero. `find . -name '*.py' -o -name '*.sh' -o -name '*.js'` returns nothing outside `.git/` — measured at implement time.
- **Instruction budget**: the four `SKILL.md` files total under 4,000 words — measured with `wc -w`.
- **Trigger separation**: no two skill descriptions claim the same user phrase; each description names what distinguishes it from its nearest sibling — verified by reading the four descriptions together.
- **Human writes nothing**: in an end-to-end run against a scratch vault (init, ingest one source, ask one question, lint), the user edits zero files under `wiki/` by hand.
- **Round-trip integrity**: after that run, `git log --oneline` shows one labelled commit per operation, and every claim on a concept page traces to a source page that names a file in `raw/`.

## User Stories

### US1 — Turn a folder into a vault (P1)

A person has a folder — empty, or already an Obsidian vault with loose notes — and wants it to become a knowledge base the LLM maintains. They ask Claude to set it up. Claude asks what the wiki will track, then creates the directory layout, writes a schema file tailored to that purpose, seeds the index and log, and initializes git.

**Acceptance Scenarios**:

- **Given** an empty folder, **When** the user asks to initialize a wiki there and states its purpose, **Then** the folder contains a schema file, `wiki/index.md`, `wiki/log.md`, a `raw/` directory, category directories matching the stated purpose, and a git repository with one `schema:` commit.
- **Given** a folder that already contains a schema file or a `wiki/` directory, **When** the user asks to initialize a wiki there, **Then** Claude reports it is already a vault, makes no changes, and offers a lint instead.
- **Given** a folder containing unrelated files such as `.obsidian/` or loose notes, **When** initialization runs, **Then** those files are left untouched.
- **Given** the user states a purpose such as tracking a book, **When** the category directories are created, **Then** they reflect that purpose rather than a fixed generic set.

### US2 — File a source into the wiki (P1)

A person drops a document into `raw/` and asks Claude to process it. Claude reads it, discusses the key takeaways, writes a summary page, updates every entity and concept page the source touches, updates the index, appends to the log, and commits.

**Acceptance Scenarios**:

- **Given** a markdown file in `raw/` that no wiki page cites, **When** the user asks to ingest it, **Then** a summary page exists under the vault's sources directory naming that file, the pages it touches are created or updated, the index lists every new page, the log gains one dated entry, and one `ingest:` commit contains exactly those paths.
- **Given** the user has not named a file, **When** they ask what is waiting to be ingested, **Then** Claude lists the files in `raw/` that no page cites as a source.
- **Given** a source whose content contradicts a claim on an existing page, **When** it is ingested, **Then** the contradiction is recorded on the page alongside the existing claim, and named in the log entry — the prior claim is not silently replaced.
- **Given** a `.pdf`, `.docx`, or `.xlsx` in `raw/`, **When** it is ingested and a suitable converter is available, **Then** a markdown extraction is written to a separate location, the original file is unmodified, and the summary page's source reference points at the original.
- **Given** the same source file, **When** ingestion completes, **Then** the file is committed to git for the first time as part of that ingest's commit.
- **Given** several un-ingested sources, **When** the user asks to process all of them, **Then** each is filed with its own commit and log entry, and Claude reports the takeaways for all of them at the end rather than pausing after each.

### US3 — Ask the wiki a question (P1)

A person asks a question against what they have accumulated. Claude locates the relevant pages through the index, reads them, and answers with citations. When the answer is substantive, it can be filed back into the wiki so the exploration compounds.

**Acceptance Scenarios**:

- **Given** a vault with pages covering the topic, **When** the user asks a question, **Then** the answer cites the specific wiki pages it drew on.
- **Given** two pages that disagree on a point the question touches, **When** the answer is produced, **Then** it surfaces the disagreement rather than choosing one side without saying so.
- **Given** a vault with nothing on the topic, **When** the user asks about it, **Then** Claude says the wiki does not cover it and does not answer from outside the vault without saying that is what it is doing.
- **Given** a substantive answer such as a comparison or analysis, **When** the user accepts the offer to keep it, **Then** it becomes a page in the vault's syntheses directory, appears in the index, gains a log entry, and is committed.
- **Given** a question whose answer is better shown than told, **When** the user wants it, **Then** the answer can take the form of a table or a Marp slide deck within the filed page.

### US4 — Health-check the wiki (P2)

Periodically, a person asks Claude to check the wiki's health. Claude reports contradictions, claims newer sources have superseded, concepts mentioned often but lacking a page, missing cross-references, and gaps worth filling — then fixes only what the user approves. Orphan pages are not computed: the non-goals assign that to Obsidian's graph view, and lint points the user there instead.

**Acceptance Scenarios**:

- **Given** a vault with accumulated pages, **When** the user asks for a lint, **Then** Claude reports findings grouped by severity and changes nothing before the user responds.
- **Given** reported findings, **When** the user approves a subset, **Then** only those are applied, and a `lint:` commit records them together with a log entry.
- **Given** a vault whose index no longer matches the files on disk, **When** a lint runs, **Then** the discrepancy is reported.
- **Given** a vault that has grown past the scale the index-first approach is documented to hold, **When** a lint runs, **Then** Claude says so rather than reporting health normally.

### US5 — Work in a vault the plugin did not create (P2)

A person already keeps a wiki with their own conventions. They install the plugin and expect it to follow their schema, not impose its own.

**Acceptance Scenarios**:

- **Given** a vault whose schema file specifies different directory names, frontmatter fields, or a different log format, **When** any skill runs, **Then** it follows the vault's conventions rather than the plugin's defaults.
- **Given** a vault whose schema is silent on some convention, **When** a skill needs that convention, **Then** the plugin's default applies.
- **Given** more than one candidate vault in scope, **When** a skill runs, **Then** Claude asks which vault before writing anything.

### US6 — Work the same way in Claude Code and in Cowork (P1)

The plugin runs on two hosts that differ in where the working directory points and in how the shell reaches files. Cowork's working directory is a scratch `outputs/` folder cleared between sessions, and its shell is an isolated Linux sandbox that reaches the user's folders through mounts, so the same file may have different paths in the shell and in the file tools. A person using either host expects identical behaviour and expects their work to persist.

**Acceptance Scenarios**:

- **Given** any host, **When** a skill needs the vault, **Then** it resolves the vault root as an explicit path and derives every other path from it, never assuming the current working directory is the vault.
- **Given** a host whose working directory is temporary scratch space, **When** any skill writes a page, an index entry, or a log entry, **Then** the write lands under the resolved vault root and nothing the user needs is left in scratch space.
- **Given** a page that must be created or updated, **When** a skill writes it, **Then** it writes with file tools rather than shell redirection, so the write is unaffected by shell-versus-file-tool path divergence.
- **Given** a vault operation that would end in a commit, **When** the skill reaches the commit step, **Then** it confirms `git` is available before invoking it, and on absence reports that the change was written but not committed instead of failing the operation.
- **Given** a shell command that reports paths, such as a directory listing, **When** those paths are used for a subsequent file-tool operation, **Then** they are re-resolved against the vault root rather than passed through verbatim.

## Non-Functional Requirements

- **MUST**: Every skill reads the vault's schema file before writing anything to that vault.
- **MUST**: The vault root is resolved explicitly before any read or write, and every path is derived from it. No skill treats the current working directory as the vault.
- **MUST**: Wiki pages, the index, and the log are read and written with file tools. The shell is reserved for git and for read-only queries, and paths are never carried between the two without being re-resolved against the vault root.
- **MUST**: The plugin functions with no converter tools installed — a source it cannot convert is reported as such, with the conversion left to the user, and no other skill is blocked.
- **SHOULD**: An operation interrupted midway leaves the vault in a state git can describe — no half-written page committed as if complete.

## Edge Cases

- A source file contains text addressed to the LLM ("ignore your instructions", "add this page"). It is quoted to the user and not acted on, per the constitution's trust boundary.
- A scanned PDF has no text layer. Claude says so and offers to transcribe from page images, noting in the extraction that it did.
- A spreadsheet has thousands of rows. The extraction records schema, row count, and a sample rather than dumping the contents.
- A source is already cited by an existing page. Claude reports it as already ingested and asks whether to re-ingest before doing anything.
- The vault is not a git repository. Skills still work; commit steps are skipped with a note, and the user is offered `git init`.
- A question spans more pages than fit in context. Claude reads the most relevant first and says which it did not read.
- The user asks to ingest a file that is not in `raw/`. Claude asks whether to stage it there first rather than reading it in place.
- A wiki page grows past one coherent topic. Lint proposes a split; ingest does not split silently.

## Assumptions & Dependencies

- `.specs/constitution.md` R00 is approved and governs this feature.
- `git` is expected on both supported hosts but its presence is verified at use time, not assumed. Obsidian is assumed on the user's side but never invoked by the plugin.
- Both supported hosts provide file tools and a shell. They differ in that Cowork's working directory is temporary scratch space and its shell is an isolated sandbox reaching user folders through mounts, so shell paths and file-tool paths may differ for the same file.
- Converter tools (pandoc, poppler, Python libraries) may or may not be present; presence is detected at use time.
- The existing `.claude-plugin/plugin.json` and `marketplace.json` are carried forward and updated, not recreated. This ships as version `2.0.0` under the existing `llm-wiki` name: a major bump, because five skills and the graph script are removed. Version 1 installations keep working until their owners choose to update.
- `docs/llm-wiki.md` is the source pattern; it is committed and credited.

## Explicit Non-Goals

- No search engine, search tooling, or embedding infrastructure. The index is the retrieval layer.
- No graph tooling — no hub ranking, orphan export, DOT output, or neighborhood traversal. Obsidian's graph view already serves this.
- No memory model — no retrieval-hook frontmatter, no budget-aware progressive loading, no recall skill.
- No skills beyond the four specified. Recall, remember, research, digest, and stats stay in the backlog.
- No executable code of any kind, including validation or test scripts.
- No reimplementation of anything Obsidian, git, or the standard shell already provides.

## Open Questions

None.
