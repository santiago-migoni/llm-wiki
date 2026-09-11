# Fase 9 — Validación en Codex

## Test run

- Date: 2026-09-11 11:51:47 -03:00 (`America/Argentina/Mendoza`)
- Host: Codex CLI local
- Plugin commit: `30189840c06fbfb2d951b34f242dbf2bd07172a1` (Fase 8)
- Working-tree delta: `plugins/llm-wiki/tests/assess-vault-scale.sh` adds postponed annotation evaluation for Python 3.9 compatibility; uncommitted during this run
- Clean checkout: `/private/tmp/<redacted>/checkout`, status clean, `HEAD` matches the plugin commit
- Installed package: `llm-wiki@llm-wiki`, version `1.0.0`, installed and enabled through the local marketplace
- Scenarios: `C-001`, `C-002`, `F-001`, `H-001`

## Results

| ID | Status | Evidence |
| --- | --- | --- |
| C-001 | passed | `codex plugin marketplace add /private/tmp/<redacted>/checkout` and `codex plugin add llm-wiki@llm-wiki` completed successfully; `codex plugin list` reported `installed, enabled`. |
| C-002 | passed | The installed package exposed exactly `wiki-init`, `wiki-ingest`, `wiki-query`, and `wiki-lint`; each skill had valid `name` and `description` frontmatter. |
| F-001 | passed | `PYTHONDONTWRITEBYTECODE=1 python3 tests/test-functional-workflow.py` returned `functional workflow: ok`; the run covered `I-001`, `G-001..G-004`, `Q-001..Q-003`, `L-001..L-002`, `M-001`, `K-001`, and `S-001`. |
| H-001 | blocked | The semantic `init → ingest → query → update → historical query → lint` smoke test was not executed in this existing task. Changed skills require a new Codex task to verify host discovery and behavior, and a new task was not created implicitly. |

## Deterministic checks

The following checks passed from the repository working tree:

- `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s plugins/llm-wiki/scripts/tests -v` — 34 tests passed.
- `PYTHONDONTWRITEBYTECODE=1 python3 tests/test-fixture-lint.py` — passed.
- `bash tests/check-plugin-contract.sh` — passed.
- `git diff --check` — passed.

The initial clean-checkout run exposed a traceback in the scale assessor under the host's
Python `3.9.6`. The minimal compatibility fix is included in the working-tree delta above;
the same deterministic workflow then passed under Python `3.9.6`.

## Artefacts and hashes

The disposable vault used by `F-001` is removed at process exit. The source code and installed
skill files remain inspectable from the checkout/cache. SHA-256 values from the repository
working tree:

- `plugins/llm-wiki/plugin.json`: `52a39d004734f9b1ec079e7819be7f06a3ca58c344858d0b5abd2b62a53102d8`
- `plugins/llm-wiki/.codex-plugin/plugin.json`: `59acc6ed74a22a08457aee0bd9852a8383ebe605d3cb4d3e1fdde6a851d54e83`
- `plugins/llm-wiki/skills/wiki-init/SKILL.md`: `214d690e657fd15e2f539a7748a981e1a095d4f44c19bd19ea613497acfbfd5a`
- `plugins/llm-wiki/skills/wiki-ingest/SKILL.md`: `e16140c863269c03a0c78aef7283c9064a58429f848c7c9ebe07dde708879771`
- `plugins/llm-wiki/skills/wiki-query/SKILL.md`: `5c682203f509b90860b83bf63ceb2cc7f1954b394db6451e88bd5c4b242938d5`
- `plugins/llm-wiki/skills/wiki-lint/SKILL.md`: `dccb1ee4a3d02ed43a33914dd8f5af5db7fef40d2d22cca6b81a46edbfa94a04`

## Limitations

- The global Codex configuration contains `model_reasoning_effort = "max"`, which this local
  CLI rejects. Installation and inspection commands therefore used the non-persistent override
  `-c model_reasoning_effort=high`.
- No LLM prompt was issued in this task for the semantic smoke test, so this record does not
  claim that an agent selected a slug, interpreted a source, wrote a vault, created a commit,
  answered a citation query, or performed a historical query.
- The release gate remains open for `H-001` until the same procedure is run in a new Codex task.
