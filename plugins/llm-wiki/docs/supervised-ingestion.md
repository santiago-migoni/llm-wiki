# Supervised ingestion checkpoint

Single-source ingestion is supervised by default. The agent performs read-only discovery and
prepares a proposal before changing `raw/sources/`, `wiki/`, or `wiki/log.md`. The checkpoint lets
the user correct the proposed slug, source authority, affected scope, or extraction coverage before
any vault file needs to be reverted.

## Proposal phase

The proposal must be based on the source and current vault, but it must not write to the vault. A
temporary conversion artifact is allowed outside the vault and must be discarded after review. The
report contains:

~~~markdown
## Ingest proposal

- Operation: new | update | duplicate | ambiguous
- Input: raw/inbox/<filename>
- Proposed slug: <stable-slug>
- Identity basis: <why this is new or the existing logical document>
- Proposed source authority: primary | supporting | context | counterpoint
- SHA-256: <hash or unavailable>
- Extraction: <format>, <method>, <status>, <processed>/<expected> <unit>
- Extraction warnings: <warnings or none>
- Key takeaways: <concise source-grounded points>
- Affected pages and syntheses: <exact paths, including inspected-but-unchanged paths>
- Contradictions: <new, changed, unresolved, resolved, or none>
- Planned files: <exact paths>
- Planned commit: ingest: <new|update> <slug>
~~~

The agent must also report duplicate or ambiguous identity findings and stop before applying them.
A duplicate is a no-op; an ambiguous identity remains in `raw/inbox/` until the user chooses the
logical document.

## Approval and apply phases

For a supervised single-source ingest, the proposal ends with an explicit request for approval. Do
not move the input, create the source record, update a page, append the log, or create a commit
before approval. The user may revise the slug, authority, extraction scope, affected pages, or
contradiction treatment; the agent regenerates the proposal from that correction.

After approval:

1. apply only the approved source, extraction, assets, pages, indexes, and log paths;
2. validate provenance, extraction coverage, contradictions, links, and exact touched paths;
3. present the resulting Git diff and validation result;
4. create one coherent `ingest:` commit when Git is available;
5. report the final paths, hash, extraction coverage, contradictions, and commit.

If validation or the diff differs materially from the approved proposal, stop before committing and
show the difference. Never treat approval of a proposal as approval to delete, merge, or alter an
unlisted path.

## Explicit batch mode

An explicitly requested batch may omit the individual pause, but it must still produce a summary
before writes containing one proposal row per input: operation, slug, hash, extraction coverage,
affected paths, contradictions, and planned commit boundaries. Ambiguous identities and duplicates
remain blocked or no-op respectively. A batch does not authorize deletion, merging, or silent
authority changes.
