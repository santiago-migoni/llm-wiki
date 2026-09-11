# Functional workflow fixture

This fixture contains inputs and expected snapshots for the deterministic
functional workflow. The runner copies `tests/fixtures/empty-vault/` into a
temporary Git vault, materializes approved ingest artifacts, and verifies the
current tree, source hashes, provenance, links, contradictions, validation,
scale report, duplicate no-op, and Git history.

It does not stand in for a Codex or ChatGPT Work host execution. Those smoke
tests require the host to expose a disposable vault and are recorded separately
under `tests/results/`.
