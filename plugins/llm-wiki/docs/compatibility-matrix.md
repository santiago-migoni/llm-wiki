# Compatibility matrix and evidence status

The capability contract describes what can work when a host exposes the required files. This
matrix records what has actually been checked. `host-validated` requires an executed host flow and
inspectable artefacts; a manifest or deterministic test alone is not enough.

## Current evidence

| Host or layer | Version / commit | Evidence | Status | Limitation |
|---|---|---|---|---|
| Codex CLI local | Plugin `1.0.0`, commit `30189840c06fbfb2d951b34f242dbf2bd07172a1` | Local marketplace install, enabled package, and four discovered skills; see [Codex result](../../tests/results/codex/2026-09-11_codex_phase9.md) | `host-validated` for `C-001`/`C-002` | The semantic `H-001` cycle requires a new task and remains blocked. |
| Codex deterministic tools | Python `3.9.6` working tree plus CI Python `3.11` | 34 unit tests, fixture lint, functional workflow, scale tests, and plugin contract | `automated` | The documented CLI baseline remains Python 3.10+; the scale assessor also works on the observed 3.9 host. |
| ChatGPT Work | No host session exposed | [ChatGPT Work result](../../tests/results/chatgpt-work/2026-09-11_chatgpt-work_phase9.md) | `blocked` / experimental | A project or workspace must expose a disposable vault before read/write, partial-context, and Git behavior can be checked. |
| Portable package | Manifest version `1.0.0` | `plugin.json`, `.codex-plugin/plugin.json`, marketplace catalog, and contract check | `automated` | Availability still depends on the host's plugin installation flow. |

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

Documentation and manifests are prepared for v1.1, but v1.1 must not be described as released while
`H-001` is blocked and `H-002` has no host evidence. The release gate can be reconsidered after a
new Codex task runs the semantic cycle and ChatGPT Work either passes its smoke test or is clearly
declared experimental in the release notes.
