---
name: llm-wiki
description: Use when the user explicitly invokes llm-wiki or names one of its vault workflows.
---

# llm-wiki

Invoke this router only for an explicit llm-wiki workflow. Select the single skill matching the requested action; do not trigger background work from a general mention of notes or projects. The package has no hooks or scheduled tasks configured.

Before vault operations, read the vault manual and establish its explicit project/vault selection. Load relevant context and the folder map; do not infer a vault path from shell variables. Read [capability-boundaries.md](../../references/capability-boundaries.md) when a workflow needs external access.

Use [workflows.json](../../references/workflows.json) for the 47 F-ID-to-skill mappings, acceptance cases and original write policies. The detailed skill is the procedure: keep its mode distinctions, output shape, and persistence policy. Do not chain an optional follow-up unless requested.

Use available `wiki_*` tools and inspect returned results. A catalog entry is not proof that a tool ran. If a capability is absent, continue the portions supported by available/readable input and name the stage that could not run.
