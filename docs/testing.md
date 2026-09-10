# Testing strategy

## Why this is a hybrid test suite

LLM Wiki has no runtime or external service. Its behavior is expressed by Codex skills, so testing has two layers:

1. deterministic contract checks for the plugin package;
2. behavioral smoke tests performed by an agent in a disposable vault.

The first layer catches broken packaging and drift. The second verifies that a host follows the workflow and produces the expected files and Git history.

## Deterministic contract check

Run from the repository root:

~~~bash
bash tests/check-plugin-contract.sh
~~~

The check validates:

- the Codex plugin manifest;
- the four skill entrypoints;
- required supporting references;
- the expected Git-first paths in documentation;
- the read-only vault scale assessor against the empty fixture;
- the absence of obsolete revision-folder instructions in operational skills.

It does not prove that an LLM will make a good semantic decision.

## Fixture

tests/fixtures/empty-vault/ is a small valid vault used as a baseline for lint and query smoke tests. It has no sources and no canonical pages.

For initialization tests, use a fresh temporary directory rather than mutating the repository or the fixture:

~~~bash
test_root=$(mktemp -d)
~~~

Remove the temporary directory after the test has been reviewed. Do not run initialization tests against the plugin repository itself.

## Test matrix

| ID | Scenario | Expected evidence |
|---|---|---|
| C-001 | Manifest contract | JSON parses; name, version, skills, and interface are present. |
| C-002 | Skill contract | All four skills have valid frontmatter and discoverable references. |
| I-001 | Initialize empty folder | Exact Git-first tree, schema, index, overview, log, and category directories are created. |
| I-002 | Initialize existing vault | No files are changed; the skill reports an existing or partial vault. |
| I-003 | Initialize inside existing Git repo | The parent repository is reused; no nested .git directory is created. |
| G-001 | Ingest new Markdown source | Stable slug, source.<ext>, complete extracted.md, page, index, log, and ingest commit exist. |
| G-002 | Update existing source | Same slug and current paths are updated; prior state is available through Git. |
| G-003 | Exact duplicate | No new source, page, log entry, or commit is created. |
| G-004 | Historical duplicate | A hash already present in the slug history is recognized without creating a new state. |
| G-005 | Ambiguous identity | Input remains pending and the agent asks before filing. |
| G-006 | Contradictory source | The conflict is recorded and not silently overwritten. |
| G-007 | Conversion | PDF, DOCX, spreadsheet, image, or other supported input has a complete or explicitly partial extraction. |
| Q-001 | Targeted query | Answer cites canonical pages and the source/extraction markers actually read. |
| Q-002 | Current-corpus query | All current slugs are inventoried, coverage is reported, and pending files prevent a false exhaustive claim. |
| Q-003 | Historical query | Git commits and exact paths support the change or provenance claim. |
| Q-004 | Filed synthesis | Existing syntheses are checked first; a new durable page records mode, sources, pages, and coverage. |
| L-001 | Lint empty vault | Empty scaffold is healthy and missing knowledge is not reported as a failure. |
| L-002 | Lint populated vault | Source, extraction, link, index, duplicate, contradiction, metadata, and scale checks report evidence. |
| S-001 | Assess empty-vault scale | The read-only assessor reports GREEN, zero sources, zero pages, and zero pending files without mutating the fixture. |
| S-002 | Assess populated-vault scale | The assessor reports current source/page counts, byte totals, structural gaps, Git state, and the highest applicable threshold. |
| H-001 | Codex smoke test | Plugin loads, skills are discoverable, and a new thread can access the current package. |
| H-002 | ChatGPT Work smoke test | Workflow works when the vault is exposed to the conversation; inaccessible files are reported. |

## Behavioral smoke test

Use a disposable Git vault and a small text source.

1. Run wiki-init and verify the exact structure.
2. Put a new source in raw/inbox/ and run wiki-ingest.
3. Query a fact that requires the extracted text.
4. Copy the source with one meaningful change into raw/inbox/ and ingest it using the same slug.
5. Use Git history to compare the two states.
6. Place the unchanged source in raw/inbox/ and verify the duplicate no-op.
7. Add a contradiction and verify that it is surfaced.
8. Run wiki-query in targeted, current-corpus, and historical modes.
9. Run wiki-lint and verify the report format.
10. Run the read-only scale assessor and record its status with the vault commit.
11. Repeat the smoke test in each supported host.

The smoke test must inspect actual files and Git commits. A response that merely claims completion is not evidence.

## Result record

For each run, record:

~~~markdown
## Test run

- Date: YYYY-MM-DD
- Host: Codex | ChatGPT Work
- Plugin commit: <commit>
- Vault location: <temporary or project path>
- Scenarios: <IDs>
- Passed: <IDs>
- Failed: <IDs>
- Blocked: <IDs and reason>
- Notes: <context-access or format limitations>
~~~

A host test is blocked, not passed, when the host cannot expose the vault or Git history required by the scenario.
