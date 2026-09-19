---
schema: 1
id: 6ga968z46e9y
status: completed
epic: 20-cli-ux-and-ergonomics
description: Record the JSON machine contract as a monotonic revision and make compatibility declarations visible and test-enforced.
effort: S
tier: 2
priority: high
autonomy_level: 3
tags: [cli, json, contract, adr]
created: "2026-09-15"
updated_at: "2026-09-19"
started_at: "2026-09-15"
depends_on: [6ga3z57hqa70]
completed_at: "2026-09-19"
---

# Define and enforce the CLI machine-contract revision policy

## Objective

Record the policy already practiced by the `--json` contract: one global,
monotonic revision identifies all machine-readable JSON behavior, additive
evolution is preferred, and an incompatible revision is called out explicitly
rather than disguised as SemVer. Put the decision where humans can understand
its rationale and make its operational parts visible to authors, tests, agents,
and future primary adapters.

## Acceptance criteria

- [x] ADR-0008 defines what `schema_version` does and does not version, what a
  consumer may infer from it, and why it is a monotonic revision rather than a
  SemVer compatibility range.
- [x] The ADR incorporates the architecture audit's primary sources and records
  the durable-id/readable-slug, bounded-projection, actionable-error, and
  generated-schema obligations that make the contract useful to agents.
- [x] `SchemaVersion`'s nearby documentation and the architecture guide point to
  ADR-0008 and no longer describe the value as SemVer.
- [x] Every changelog revision introduced after ADR-0008 declares `ADDITIVE` or
  `NOT ADDITIVE`, and a focused test rejects missing or unknown declarations.
- [x] `schema --json` publishes the revision policy without requiring source or
  prose parsing, and its human rendering explains the same policy concisely.
- [x] The generated Draft 2020-12 schema has a versioned `$id` and an explicit
  root revision annotation, with focused and golden coverage.
- [x] The additive wire revision and all generated artifacts are recorded, and
  focused tests, the full suite, lint, planning lint, and diff checks pass.

## Out of scope

- Do not implement the separate exit-code taxonomy or Thread-list projection
  findings; this task only establishes the policy they must follow.
- Do not introduce per-envelope versions, SemVer range negotiation, migrations,
  or a remote schema registry.
- Do not claim the generated JSON Schema describes projected `--json -c` rows:
  their keys and order are caller-selected even though their machine behavior is
  covered by the global revision. The on-disk frontmatter `schema:` key remains
  a separate contract.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Audit [2026-09-14 machine-contract architecture review](../audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md), especially H1 and L1
- ADR [ADR-0003: Stable-key, ID-addressed storage](../adrs/0003-stable-key-id-addressed-storage.md)
- Thread [Harden CLI machine contracts](../threads/6g9txm0s3yyh-harden-cli-machine-contracts.md)

## Implementation outcome (2026-09-19)

[ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
records one monotonic revision for every JSON machine surface, including dynamic
projections, while limiting the generated Draft 2020-12 schema to typed envelopes.
It defines tolerant-reader compatibility separately from exact-revision schema
validation and retains the on-disk `schema:` key as an independent contract.

Revision 1.68 is classified `ADDITIVE`. `schema --json` now publishes the scheme,
scope, current/default compatibility, classification boundary, reader expectation,
and schema-validation scope. `schema --json-schema` uses a revision-qualified `$id`
and root revision/scheme/classification annotations. Wire constructors stamp the
policy so another primary adapter cannot omit or override it.

The source-adjacent changelog test requires every revision from 1.68 onward to
declare `ADDITIVE` or `NOT ADDITIVE`, rejects missing/unknown labels, and verifies
the current marker against the executable classification. Architecture, contributor,
Threads-compatibility, audit-routine, and README guidance now point to the decision.
Audit findings H1 and L1 are fixed with resolution notes.

Validation: focused wire/CLI/render tests, `go test -race ./...`, golangci-lint,
generated CLI-doc drift, module-tidy drift, generated schema-comment drift, live
schema probes, planning lint, and `git diff --check` all pass.

## Adversarial review remediation (2026-09-19)

Independent Codex and Antigravity reviews confirmed revision 1.68 is additive
and found enforcement gaps rather than a policy-design failure. The remediation
normalizes wire-owned policy before both human and machine rendering, constrains
every registered envelope's JSON Schema `schema_version` to `const: "1.68"`, and
adds a validator regression for a wrong-revision payload.

The ordinary golden updater now refuses to replace changed or new JSON fixtures
until `SchemaVersion` advances beyond their committed revision marker. A separate
projection-contract snapshot covers ordered selectors and aliases that cannot be
represented in the typed-envelope schema, and the reserved major line is guarded
pending a superseding ADR. Remaining review polish pins every policy constant and
gives malformed changelog delimiters an actionable diagnostic.

Reviews:

- [Codex implementation audit](../audits/6gbjq90ts0cr-2026-09-19-machine-contract-revision-policy-implementation-codex.md)
- [Antigravity implementation audit](../audits/6gbjq9149xmh-2026-09-19-machine-contract-revision-policy-implementation-antigravity.md)

Post-remediation validation: all six findings are fixed and both audits are
closed; `go test -race ./...`, focused wire/CLI/render tests, the guarded full
golden update, golangci-lint, module-tidy drift, generated CLI-doc drift,
planning lint, JSON Schema envelope/const counts, and `git diff --check` pass.
