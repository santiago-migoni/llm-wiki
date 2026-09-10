# Archive model

The archive model is defined in [wiki-architecture.md](wiki-architecture.md).

The current design uses:

- one stable slug per logical document;
- the current original and extraction under raw/sources/<slug>/;
- one canonical living page in wiki/pages/<slug>.md when the source has a corresponding page;
- Git history for versioning, comparison, provenance, and recovery.

The former revision-folder model is obsolete. Do not create raw/archive/, raw/catalog.md, per-revision source pages, or revision identifiers as part of normal ingestion.
