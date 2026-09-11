# Contradiction contract

Material conflicts are queryable knowledge, not transient notes. A contradiction is recorded on
the affected canonical page and in `wiki/log.md`; Git preserves the historical versions. The
deterministic validator checks the record shape and the page/log cross-reference. It does not decide
which source has authority.

## Canonical page entry

Add a `## Contradictions` section to the affected page. Each stable identifier is a lowercase
kebab-case name and has two competing claims with their evidence:

~~~markdown
## Contradictions

### review-frequency

- Claim A: Review occurs annually.
- Evidence A: `corporate-policy`, section 4.
- Claim B: Review occurs quarterly.
- Evidence B: `audit-report`, page 14.
- Status: unresolved
- Impact: Changes the compliance schedule.
~~~

The required fields are `Claim A`, `Evidence A`, `Claim B`, `Evidence B`, and `Status`. `Impact`
is strongly recommended when the conflict can change an action, date, risk, or interpretation.
Evidence must name the declared source slug and a locator such as a section, page, slide, sheet,
cell range, or timestamp.

## Statuses and resolution

- `unresolved`: both claims remain active and no reviewed decision has selected an interpretation.
  Query responses must present the conflict and must not state either claim as certain.
- `resolved`: a reviewed decision explains which interpretation is currently used and why. Keep
  both claims and both evidence fields, and add `Resolution` and `Criterion`:

  ~~~markdown
  - Status: resolved
  - Resolution: The signed policy has precedence.
  - Criterion: The effective-date clause is the governing authority.
  ~~~

- `superseded`: one claim has been replaced by a later or more authoritative state. Preserve the
  superseded claim and its evidence, identify the current interpretation in `Resolution`, and state
  the review basis in `Criterion`.

No status is inferred from source order, filename, recency, or a presumed authority hierarchy.
Resolving or superseding a conflict is a reviewed semantic change, not an automatic lint repair.

## Chronological log event

Every page entry has one matching event in `wiki/log.md`:

~~~markdown
## [YYYY-MM-DD] contradiction | review-frequency

- Page: [[policy]]
- Status: unresolved
~~~

The identity is the pair `(Page wikilink, contradiction identifier)`. The validator reports a
canonical contradiction missing from the log, a log-only contradiction, duplicate events, and
status mismatches. A log entry without a canonical page is not sufficient evidence for a query.

## Query and update rules

Before answering a targeted query, inspect `## Contradictions` on each relevant canonical page and
the corresponding log state. For an unresolved conflict, cite both claims and state that the
answer is contested. For a resolved or superseded conflict, preserve the competing evidence and
explain the recorded criterion before using the current interpretation. If a page, log entry, or
source is unavailable, report the limitation instead of assuming the conflict is absent.

When a source update introduces or changes a material conflict:

1. identify every affected canonical page before writing;
2. preserve both source citations and the existing claim history;
3. add or update the structured contradiction entry;
4. append the matching dated event to `wiki/log.md`;
5. validate the page, log, provenance, links, and status before committing.

The canonical page is the place for the claims and evidence; the log is the chronological pointer.
Neither should silently erase the other.
