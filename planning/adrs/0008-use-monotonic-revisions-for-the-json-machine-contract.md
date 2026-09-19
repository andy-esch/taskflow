---
status: accepted
date: "2026-09-15"
deciders: [andy-esch]
tags: [adr, cli, json, machine-contract, compatibility]
supersedes: []
superseded_by: null
---

# ADR-0008: Use Monotonic Revisions for the JSON Machine Contract

> ✔ **Accepted 2026-09-19 — finalized.** The decision sections below are frozen; add
> new information only under `## Amendments`. Reverse via a superseding ADR.

> Follows the ADR format established in
> [0001-adopt-adrs](0001-adopt-adrs.md). This decision closes H1 and L1 from the
> [2026-09-14 machine-contract architecture audit](../audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md)
> and makes explicit the policy behind the executable changelog that already
> accompanies `internal/wire.SchemaVersion`.

## Context and Problem Statement

Every machine-readable JSON response carries one global `schema_version`. Its
source comment called the value SemVer and promised that additions incremented
the minor component while renames and removals incremented the major component.
Practice was different: revisions 1.25, 1.26, 1.32, 1.47, and 1.59 each included
an incompatible change without leaving major version 1. The changelog described
those changes candidly, but the number advertised a compatibility promise the
project did not keep.

That divergence matters increasingly as the same neutral `internal/wire` values
become usable outside Cobra. Agents, a future served adapter, and other callers
need to know what the number means without inferring policy from a Go comment or
treating its SemVer-like spelling as a version range.

The contract also has two different forms:

- statically typed envelopes, reflected into `schema --json-schema`; and
- projected list JSON, whose selected row keys and order are chosen at runtime
  with `-c`.

Both are machine JSON and both carry `schema_version`, but only the first can be
fully represented by one generated JSON Schema. Separately, the `schema:` field
in Markdown frontmatter versions persisted file shapes and is not a wire
contract at all.

The machine contract is designed for agents as well as conventional programs.
That creates additional obligations beyond naming a version: return the durable
identifier a later operation needs while retaining readable names, allow bounded
projections, and report actionable structured errors. The stable-ID list
projection work established this shape before this decision was recorded.

## Considered Options

### A — Enforce Semantic Versioning strictly

Any incompatible change would advance the major version. This gives the number
familiar range semantics, but it would retroactively make five shipped revisions
policy violations and would encourage compatibility inference from the number
instead of inspection of the binary's self-described contract. For this young,
single-binary tool, repeated major releases would communicate maturity and
migration guarantees the project does not yet offer.

### B — Use one monotonic contract revision with explicit compatibility labels

Keep the established `major.minor` spelling and one global counter, but define it
as identity and ordering rather than SemVer. Prefer additive evolution and label
every new revision `ADDITIVE` or `NOT ADDITIVE`. This matches practice while
making incompatibility visible to authors, reviewers, and consumers.

### C — Version each envelope independently

This localizes changes but makes a caller reason about dozens of coupled versions
from one shipped binary. It also complicates errors, shared nested DTOs, and
cross-envelope workflows without solving dynamic projections.

### D — Publish only the generated JSON Schema

Content-addressed schemas can describe static shapes, but they do not capture
command semantics, exit codes, closed vocabularies, or caller-selected projected
rows. Schema identity complements rather than replaces the contract revision.

## Decision

Choose **B: one monotonic contract revision with explicit compatibility labels**.

### 1. `schema_version` is a revision, not SemVer

`schema_version` is a monotonically increasing identifier for the complete
machine-readable JSON contract shipped by one `tskflwctl` binary. Its
`major.minor` syntax is retained for continuity; it does not grant SemVer range
semantics. A consumer may compare revisions for equality and ordering, but must
not infer that every `1.x` response is compatible with every other `1.x`
response.

The major component is reserved. Changing its meaning or starting a new major
line requires another ADR; ordinary contract changes continue the minor counter.
There are no per-envelope versions.

### 2. The revision covers all JSON output; the generated schema covers typed envelopes

The revision covers:

- every typed `--json` success and error envelope;
- `schema --json` and the semantics it publishes;
- projected list JSON, including selector names, value semantics, and requested
  column order; and
- the generated schema document associated with that revision.

The Draft 2020-12 document describes only statically typed envelopes. A projected
row is intentionally caller-selected and therefore is not claimed to validate
against those definitions. It still carries the global revision because its
selectors and values are part of the same machine contract.

The revision does **not** cover human rendering, ANSI styling, Markdown body
templates, or the on-disk frontmatter `schema:` key. Those surfaces may have
their own explicit contracts.

### 3. Every new revision declares compatibility

Beginning with the first revision implementing this ADR, each changelog entry
must declare exactly one of:

- **`ADDITIVE`** — existing valid inputs and documented value meanings remain
  valid; existing keys, types, and closed-vocabulary members retain their
  meanings. Consumers must ignore unknown object keys.
- **`NOT ADDITIVE`** — a key, type, referent, requiredness, selector, error code,
  or documented meaning changes incompatibly; or a published closed vocabulary
  adds, removes, or renames a member that an exhaustive consumer may switch on.

Additive evolution is preferred, not guaranteed. An incompatible revision is
permitted when justified, but the marker and changelog must explain the consumer
action. The current revision's classification and the policy's starting revision
are published by `schema --json`; the source-adjacent test verifies that the
declaration and executable constants agree.

`ADDITIVE` describes a tolerant semantic JSON reader. The generated schema uses
`additionalProperties: false` to catch drift within one exact revision, so an old
schema is not a forward validator for a newer payload. A validator must use the
schema whose `$id` carries the payload's exact revision; a decoder consuming an
additive revision must ignore unknown object fields.

### 4. The generated schema has revision identity

`schema --json-schema` uses a `$id` containing the exact revision and publishes
the same revision and compatibility classification as root annotations. A cache
or served adapter can therefore distinguish schema resources without parsing the
human title. The schema continues to use Draft 2020-12 and one `$defs` registry
for all typed envelopes.

### 5. Consumer-facing output remains agent-usable

Machine surfaces follow these obligations:

- expose a readable slug or title for comprehension and the durable ID needed
  for stored references or later calls;
- offer bounded filtering or projection where an unbounded response would waste
  context;
- return structured, specific recovery information instead of requiring prose
  parsing; and
- derive published vocabularies and registries from their owners rather than
  maintaining a second transcription.

These are design constraints, not a demand that every envelope contain every
field. A curated projection may remain small; it must expose a durable handle
when callers are expected to retain one.

### 6. Consumers inspect the invoked binary's contract

Consumers should read `schema --json` and, when validating typed envelopes,
`schema --json-schema` from the binary they invoke. They should compare exact
revisions rather than accept a SemVer range. Unknown additive object fields must
be ignored; closed vocabulary changes and revisions marked `NOT ADDITIVE`
require deliberate handling.

## Consequences

- The existing global counter and executable changelog remain useful; no
  historical payload or envelope gains a second version.
- A SemVer-looking value no longer makes a false compatibility promise.
- Contract authors must classify each revision, and review must still decide
  whether the classification is truthful. A test can require a declaration but
  cannot infer semantic compatibility.
- `schema --json` becomes the machine-readable source for revision semantics,
  while the ADR remains the canonical rationale.
- Versioned JSON Schema identities prepare the contract for caching and a future
  served adapter without committing to such an adapter now.
- Dynamic projections remain outside the generated schema, but no longer sit
  outside the stated revision policy.
- The first classified revision is additive because it publishes policy and
  schema metadata without removing existing output.

## Amendments

### 2026-09-19 — Make exact revision and author enforcement executable

Implementation review exposed two places where the accepted policy was only
descriptive. The generated schema's revision-qualified `$id` identified the
resource but did not reject an envelope whose own `schema_version` named another
revision. Every registered envelope definition now constrains that property to
the exact revision, so pairing a cached schema and payload incorrectly fails
closed.

Machine-output goldens now carry a separate committed revision marker. The
ordinary `go test ./internal/cli -update` workflow refuses to rewrite a changed or
new JSON golden while that marker equals the running `SchemaVersion`; authors must
first advance and classify the revision. A dedicated golden records every dynamic
projection's ordered canonical selectors and compatibility aliases, covering the
contract surface the static schema deliberately cannot describe. A successful
full update advances the marker only after all CLI tests pass. Major revision
movement remains test-blocked pending the superseding ADR required by this
decision.

## Related

- [2026-09-14 machine-contract architecture audit](../audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md)
- [ADR-0003: Stable-key, ID-addressed storage](0003-stable-key-id-addressed-storage.md)
- [ADR-0007: Planning state vocabularies](0007-planning-state-vocabularies.md)
- [Semantic Versioning 2.0.0](https://semver.org/) — the compatibility semantics
  this revision deliberately does not claim
- [JSON Schema: Structuring a complex schema](https://json-schema.org/understanding-json-schema/structuring) — `$id` as schema-resource identity
- [Anthropic: Writing effective tools for agents](https://www.anthropic.com/engineering/writing-tools-for-agents) — meaningful context, downstream identifiers, bounded responses, and actionable errors
- [Command Line Interface Guidelines](https://clig.dev/) — machine-readable output and meaningful exit behavior
- Task [Define and enforce the CLI machine-contract revision policy](../tasks/6ga968z46e9y-define-and-enforce-the-cli-machine-contract-revision-policy.md)
