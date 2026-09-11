# Testing strategy

## Why this is a hybrid test suite

LLM Wiki has no runtime or external service. Its behavior is expressed by Codex skills, so testing has three layers:

1. deterministic contract checks for the plugin package;
2. deterministic functional checks over a disposable Git vault;
3. behavioral smoke tests performed by an agent in a disposable vault.

The first layer catches broken packaging and drift. The second verifies the observable file and Git
invariants without pretending to be an LLM host. The third verifies that a host follows the workflow
and produces the expected files and Git history.

## Deterministic contract check

Run from the repository root:

~~~bash
bash tests/check-plugin-contract.sh
~~~

The check validates:

- the portable and Codex compatibility plugin manifests;
- the four skill entrypoints;
- required supporting references;
- the expected Git-first paths in documentation;
- the read-only vault scale assessor against valid, malformed, and boundary fixtures;
- the absence of obsolete revision-folder instructions in operational skills.
- the dependency-free Python CLI and its strict frontmatter parser;
- deterministic inventory warnings, hash classification, wikilink resolution, structural validation,
  format-specific extraction coverage, and contradiction behavior.
- checked-in fixture outcomes and a disposable Git functional workflow with tree/history snapshots.

It does not prove that an LLM will make a good semantic decision.

Run the deterministic Python suite directly with:

~~~bash
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s plugins/llm-wiki/scripts/tests -v
~~~

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
| G-001 | Ingest new Markdown source | Stable slug, source.<ext>, format-specific extracted.md with declared coverage, page, index, log, and ingest commit exist. |
| G-002 | Update existing source | Same slug and current paths are updated; prior state is available through Git. |
| G-003 | Exact duplicate | Identical bytes are recognized across filenames and source slugs; no new source, page, log entry, or commit is created. |
| G-004 | Historical duplicate | A hash already present in any source slug's history is recognized without creating a new state. |
| G-005 | Ambiguous identity | Input remains pending and the agent asks before filing. |
| G-006 | Contradictory source | The conflict is recorded and not silently overwritten. |
| G-007 | Conversion | PDF, DOCX, spreadsheet, image, or other supported input has a complete or explicitly partial extraction. |
| G-008 | Supervised ingest checkpoint | Before any vault write, the proposal reports identity, hash, extraction coverage, takeaways, authority, affected paths, contradictions, and commit plan; approval can correct slug or scope. |
| Q-001 | Targeted query | Answer cites canonical pages and the source/extraction markers actually read. |
| Q-002 | Current-corpus query | All current slugs are inventoried, coverage is reported, and pending files prevent a false exhaustive claim. |
| Q-003 | Historical query | Git commits and exact paths support the change or provenance claim. |
| Q-004 | Filed synthesis | Existing syntheses are checked first; a new durable page records mode, sources, pages, and coverage. |
| L-001 | Lint empty vault | Empty scaffold is healthy and missing knowledge is not reported as a failure. |
| L-002 | Lint populated vault | Source, extraction, link, index, duplicate, contradiction, metadata, and scale checks report evidence. |
| S-001 | Assess empty-vault scale | The read-only assessor reports GREEN, zero sources, zero pages, and zero pending files without mutating the fixture. |
| S-002 | Assess populated-vault scale | The assessor reports current source/page counts, source and wiki Markdown byte totals, largest files, structural gaps, Git state, and the highest applicable threshold. |
| S-003 | Assess malformed vault | Wrong path types, missing extractions, duplicate current originals, invalid slugs, unexpected entries, and inaccessible directories produce `Scale status: INVALID` with stable diagnostics and no traceback. |
| S-004 | Assess exact scale boundaries | Source counts 79/80/100/101 and canonical-page counts 199/200/300/301 produce GREEN/WATCH/DERIVED-SEARCH-CANDIDATE according to the documented inclusive boundaries. |
| M-001 | Multi-source provenance | A page stores every source slug, current source/extraction path, role, and matching SHA-256; the validator checks all entries. |
| M-002 | Legacy provenance compatibility | A page using singular source/extracted/sha256 remains readable and is reported as migrable rather than invalid solely for its shape. |
| M-003 | Migration check | Default `migrate-provenance` mode makes no changes and prints exactly the affected pages plus a unified diff. |
| M-004 | Migration write | Explicit `--write` converts only reviewed page frontmatter, preserves the body, and creates no commit. |
| M-005 | Claim citations | Multi-source claims cite declared source slugs and locators; undeclared citation slugs are reported. |
| D-001 | Inventory warnings | Missing or malformed structural entries produce stable warning codes in JSON and human output. |
| D-002 | Hash classification | Current duplicates, historical duplicates, and explicit reversion targets are distinguished without writes. |
| D-003 | Link resolution | Aliases resolve, headings are checked, and canonical pages missing from `wiki/index.md` are reported. |
| E-001 | Extraction coverage contract | Extraction frontmatter declares a format-compatible unit, expected/processed counts, a valid status, and warnings when coverage is not complete. |
| K-001 | Query-visible contradiction | A structured contradiction is present on the affected page and has a matching dated `wiki/log.md` event with the same page, id, and status. |
| K-002 | Log-only contradiction | A contradiction event without a canonical page entry is reported as a stable high-severity finding. |
| K-003 | Reviewed resolution | `resolved` and `superseded` entries retain both claims/evidences and require `Resolution` plus `Criterion`; missing fields are invalid. |
| F-001 | Deterministic functional workflow | A disposable Git vault exercises scaffold, new ingest, update, current and historical duplicate no-ops, multifuente provenance, contradiction visibility, links, validation, scale, exact commit paths, and expected tree/history snapshots. |
| F-002 | Fixture lint | The checked-in empty, populated, legacy, malformed, and duplicate fixtures produce their declared validation outcomes. |
| H-001 | Codex smoke test | Plugin loads, skills are discoverable, and a new thread can access the current package. |
| H-002 | ChatGPT Work smoke test | Workflow works when the vault is exposed to the conversation; inaccessible files are reported. |

## Matriz de promesas públicas y evidencia

Esta matriz conecta las promesas visibles del plugin con escenarios y artefactos
que pueden verificarlas. El estado es el de la línea base de la remediación;
`pending` no significa que la promesa haya sido validada.

| ID | Promesa pública | Evidencia verificable | Estado de línea base |
| --- | --- | --- | --- |
| P-001 | El paquete y sus skills son descubribles por el host | `C-001`, `C-002`, `H-001`, `H-002` | `static-validated`; host pendiente |
| P-002 | La operación es Git-first y conserva fuente, extracción e historial | `I-001`–`I-003`, `G-001`, `G-002`, `Q-003` | `pending-semantic` |
| P-003 | Duplicados exactos e históricos no generan registros actuales falsos | `G-003`, `G-004` y commits inspeccionados | `pending-semantic` |
| P-004 | Las consultas de corpus completo informan su cobertura real y sus límites | `Q-002`, `H-001`, `H-002` y artefactos de consulta | `pending-host` |
| P-005 | Las respuestas conservan citas trazables a fuentes verificables | `Q-001`, `Q-003`, `Q-004` y páginas generadas | `pending-semantic` |
| P-006 | Las contradicciones materiales permanecen visibles y no se sobrescriben silenciosamente | `G-006`, `L-002`, `K-001`–`K-003`, registro de auditoría y página afectada | `static-validated; semantic/host pending` |
| P-007 | La extracción es completa o declara explícitamente sus limitaciones | `G-007`, artefactos de extracción y reporte | `pending-semantic` |
| P-008 | Lint y pruebas de escala son de solo lectura y reportan evidencia | `L-001`, `L-002`, `S-001`, `S-002` | `pending-semantic` |
| P-009 | Cada host comunica honestamente sus capacidades de lectura y escritura | `H-001`, `H-002`, registros por host | `pending-host` |

Los registros de ejecución se almacenan en
[`tests/results/`](../../tests/results/). Cada fila debe enlazar, en el
registro correspondiente, el commit del plugin, el host, los escenarios y los
artefactos que sustentan el estado declarado.

## Automated functional workflow

The deterministic functional runner creates a disposable Git vault from the empty fixture and
materializes a small, reviewed workflow with the expected files. It checks the observable
invariants that do not require an LLM host:

- new source classification and current duplicate no-op;
- update under the same slug with Git history preserved;
- historical duplicate and explicit reversion reporting;
- declared format-specific extraction coverage and multifuente provenance;
- indexed canonical page and query-visible contradiction with a matching log event;
- read-only inventory, links, hashes, validation, and scale checks;
- exact paths in each ingest commit;
- final tree and Git history against
  [`expected-tree.txt`](../../tests/fixtures/functional-vault/expected-tree.txt) and
  [`expected-history.txt`](../../tests/fixtures/functional-vault/expected-history.txt).

Run it from the repository root:

~~~bash
PYTHONDONTWRITEBYTECODE=1 python3 tests/test-functional-workflow.py
PYTHONDONTWRITEBYTECODE=1 python3 tests/test-fixture-lint.py
~~~

The runner does not claim that a Codex or ChatGPT Work agent selected the slug, interpreted the
source, or requested approval. Those behavioral and host-specific claims require the manual smoke
tests below and a result record with the host and plugin commit.

## Behavioral smoke test

Use a disposable Git vault and a small text source.

1. Run wiki-init and verify the exact structure.
2. Put a new source in raw/inbox/ and run wiki-ingest.
3. Query a fact that requires the extracted text.
4. Copy the source with one meaningful change into raw/inbox/ and ingest it using the same slug.
5. Use Git history to compare the two states.
6. Place the unchanged source in raw/inbox/ under a different filename and verify the cross-slug duplicate no-op.
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
