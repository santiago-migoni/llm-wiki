# Host compatibility

## Portability contract

The plugin is file-first. It provides skills and conventions; it does not provide document storage, a remote database, or a synchronization service.

The same vault contract applies in Codex and ChatGPT Work:

- AGENTS.md defines the vault-specific rules;
- raw/inbox/ receives pending documents;
- raw/sources/<slug>/ stores the current source record;
- wiki/ contains canonical knowledge and navigation;
- Git preserves prior states when the repository is available;
- the agent reports the context it actually accessed.

## Capability matrix

| Capability | Codex | ChatGPT Work |
|---|---|---|
| Load plugin skills | Yes, after installation and a new thread | Yes, when the plugin is available to the workspace |
| Read and write local vault files | Yes, when the vault is in the accessible workspace | Only when the vault is exposed through project/workspace files or a connected source |
| Read raw/inbox/ | Yes | Only when pending files are available to the conversation |
| Use Git history | Yes, when the repository and history are accessible | Only when the repository history is exposed; a current-tree export has no history |
| Full-corpus query | Yes, subject to file size and context limits | Yes only for the accessible current corpus; otherwise report incomplete coverage |
| Ingest binary documents | Based on available local readers/converters | Based on files and readers exposed by the host |
| Persist a filed synthesis | In the vault Git tree | In the exposed writable workspace/project, when write access exists |

Never infer a capability from the product name. Verify the files and operations available in the current conversation.

## Installation and refresh

Install the plugin through the host's supported local plugin or marketplace flow. The portable
package manifest is `plugin.json`; `.codex-plugin/plugin.json` remains a Codex compatibility
fallback. The repository catalog is `.agents/plugins/marketplace.json`.

During local development:

1. ensure the plugin source path is the one referenced by the configured local marketplace;
2. update the host's cachebuster using the plugin development workflow when required;
3. reinstall or refresh the plugin;
4. start a new thread before testing changed skills.

Do not hand-edit marketplace configuration during a routine refresh. If no marketplace entry exists, create one through the host's supported plugin workflow before using a reinstall command.

## Codex smoke test

1. Open a new thread with the updated plugin.
2. Create a disposable empty folder.
3. Ask to initialize a wiki there.
4. Verify the exact structure and initial files.
5. Add a small source to raw/inbox/ and ask to ingest it.
6. Query a fact from the source and inspect its citations.
7. Update the source, then ask what changed.
8. Run lint and verify a report-first result.

Record the plugin commit and the vault path used for the run.

## ChatGPT Work smoke test

1. Create a disposable project or workspace with the plugin available.
2. Expose the empty fixture or a disposable vault to the conversation.
3. Ask to initialize or inspect the vault.
4. Add a small source through the available file mechanism.
5. Ingest and query it.
6. Ask for full context and verify that the agent reports source coverage and pending files.
7. Ask a historical question only when Git history is exposed.
8. Run lint and verify the same report format.

If the host exposes read-only files, the agent may answer but must not claim that it persisted a synthesis or committed changes.

## Required limitation language

When the complete vault is unavailable, the agent should say which subset it could access and which files or history were unavailable.

Examples of accurate claims:

- “I answered from the three current source slugs exposed to this conversation.”
- “The pending inbox file was not included in the current-corpus pass.”
- “Git history is unavailable here, so I cannot verify the historical claim.”
- “The source was readable, but its extraction is partial because the host did not expose an image or converter.”

Avoid saying “I loaded the entire archive” unless the coverage checklist supports that claim.

## Release gate

A host is considered compatible only when:

- the plugin loads its skills;
- the host can resolve the vault root;
- the host follows the Git-first paths;
- targeted query citations work;
- full-corpus limitations are reported;
- historical behavior is honest about Git availability;
- no write is claimed when the host is read-only;
- lint produces the common report structure.

## Evidence status

The intended capability matrix above is not itself host evidence. The current checked results are
summarized in [compatibility-matrix.md](compatibility-matrix.md) and retain the distinction between
implemented, automatically tested, host-validated, and blocked. In particular, Codex installation
and skill discovery are recorded, while the semantic smoke test still requires a new task. No
ChatGPT Work capability is claimed until a workspace exposes a disposable vault and the smoke test
is run there.
