# Compatibility matrix and evidence status

The capability contract describes what can work when a host exposes the required files. This
matrix records what has actually been checked. `host-validated` requires an executed host flow and
inspectable artefacts; a manifest or deterministic test alone is not enough.

## Current evidence

| Host or layer | Version / commit | Evidence | Status | Limitation |
|---|---|---|---|---|
| Codex CLI local | Release `1.2.0`; host evidence from the `1.1.0` installed package | Local marketplace install and four discovered skills; [installation result](../../tests/results/codex/2026-09-11_codex_phase9.md) plus [real-vault follow-up](../../tests/results/codex/2026-09-11_codex_real-vault_followup.md) | `host-validated` for installation, initialization, and ingestion in `1.1.0`; hierarchical paths covered by automated tests | The referenced run did not produce a separate persisted transcript for every query mode; historical behavior is covered by the deterministic workflow. |
| Codex deterministic tools | Python `3.9.6` working tree plus CI Python `3.11` | 42 unit tests, fixture lint, functional workflow, scale tests, and plugin contract | `automated` | The documented CLI baseline remains Python 3.10+; the scale assessor also works on the observed 3.9 host. |
| ChatGPT Work | No host session exposed | [ChatGPT Work result](../../tests/results/chatgpt-work/2026-09-11_chatgpt-work_phase9.md) | `experimental` | A project or workspace must expose a disposable vault before read/write, partial-context, and Git behavior can be checked. |
| Portable package | Manifest version `1.2.0` | `plugin.json`, `.codex-plugin/plugin.json`, marketplace catalog, and contract check | `automated` | Availability still depends on the host's plugin installation flow. |

## Capability interpretation

| Capability | Codex when workspace is accessible | ChatGPT Work when workspace is accessible |
|---|---|---|
| Discover skills | Verified for `C-001`/`C-002` in the local Codex CLI | Not yet verified; do not infer from the portable manifest |
| Read current vault | Supported for accessible workspace files | Supported only for files exposed to the conversation |
| Write and commit | Supported when the workspace is writable and Git is available | Supported only when the exposed project/workspace permits it; otherwise report read-only |
| Full-corpus query | Requires an inventory, coverage ledger, and bounded reads | Same requirements, limited to the exposed current corpus |
| Historical query | Requires accessible Git history | Requires Git history to be exposed; a current-tree export is insufficient |
| Binary extraction | Depends on local readers/converters | Depends on readers/files exposed by the host; declare partial or unsupported coverage |

## Release decision

Release `v1.1.0` is prepared with Codex host evidence and ChatGPT Work explicitly marked
experimental. It must not be described as universally validated: ChatGPT Work still requires a
workspace-backed smoke test before it can move from `experimental` to `host-validated`.
