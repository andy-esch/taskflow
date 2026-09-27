---
schema: 1
id: 6gdx7mcqm371
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Prevent semantic reads, local paths, and mutations from being composed across unrelated planning corpora.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, ports, composition]
created: "2026-09-26"
depends_on: [6gcwcf7rgxef]
updated_at: "2026-09-27"
started_at: "2026-09-27"
---

# Bind split planning capabilities to one source set

## Objective

Make independently supplied semantic-read, local-path, and mutation capabilities prove that they
refer to the same planning corpus. Option ordering and implicit-capability detachment prevent one
class of accidental pairing today, but cannot detect two explicit adapters that happen to expose the
same entity IDs from unrelated repositories, databases, services, or caches.

## Scope

- Introduce a small opaque, comparable source-set identity at the capability/composition boundary.
  Keep it distinct from record identity, source location, and optimistic-concurrency revisions.
- Require independently composable planning-data readers, local-path resolvers, and mutators to
  advertise their source set, or construct them through a bundle that owns it.
- Reject mismatched explicit capabilities during service/workspace construction before any read,
  editor launch, or mutation can occur. Preserve the current rule that replacing a reader detaches
  an implicitly inherited local capability.
- Fail closed when a guarded mutation pairing lacks usable source-set identity. Read-only and
  pathless adapters remain valid when unsupported capabilities are simply absent.
- Cover typed nils, option ordering, multi-workspace composition, fakes, and same-record-ID values
  returned by different corpora.

## Acceptance criteria

- [x] Explicitly pairing readers, path resolvers, or mutators from different source sets fails at
      construction with a typed, explanatory error.
- [x] A complete filesystem adapter, a pathless read-only adapter, and test fakes can each expose an
      honest source-set identity without putting that value on domain records or wire output.
- [x] Record IDs, locations, and source revisions cannot be mistaken for source-set identity.
- [x] Replacing semantic reads still detaches implicit local capabilities; explicitly restoring a
      compatible capability remains deterministic regardless of option order.
- [x] Thread, task, epic, audit, and research composition paths share the same rule and regression
      matrix rather than implementing entity-specific approximations.

## Out of scope

- Detecting staleness within one source set; guarded revisions and mutation-time comparison retain
  that responsibility.
- Distributed transactions across multiple writable planning corpora.
- Making unsupported local or mutation capabilities mandatory for read-only adapters.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make planning data access adapter neutral](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- [Portable entity-read design](6gcwcf7rgxef-design-portable-entity-reads-and-optional-local-source-capabilities.md)
- [Codex design finding H1](../audits/6gdx1bvdf16p-2026-09-26-portable-entity-read-local-source-capability-design-codex.md)
- [Antigravity design finding L1](../audits/6gdx1bvz683y-2026-09-26-portable-entity-read-local-source-capability-design-antigravity.md)

## Implementation progress (2026-09-27)

`core.NewService` now returns a typed construction error if any selected aggregate or split
planning-data capability omits, empties, changes, or disagrees on its opaque `SourceSetID`.
`WorkspaceService.Open` propagates this before returning a usable workspace; CLI resolution uses
the checked constructor. The filesystem adapter mints one process-local token per instance and
shares it across its ports, with an explicit option for a composition root that knows two
instances address the same corpus. The token is not a path, planning-space ID, record ID, revision,
or security credential.

Construction probes cover every currently injectable planning read/path/mutation port, option
ordering and typed nils. Filesystem and workspace fixtures cover same record ID in different
corpora, and same record ID/location/revision from separate unbound instances. The later
[source/path split](6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md)
must reuse this witness for the new entity-specific path ports.

## Adversarial review (2026-09-27)

The independent [Codex](../audits/6ge63hzm2g57-2026-09-27-source-set-capability-composition-implementation-codex.md)
and [Antigravity](../audits/6ge63hzwtwn2-2026-09-27-source-set-capability-composition-implementation-antigravity.md)
reviews found no reproducible implementation defects. Both exercised cross-corpus rejection,
compatible split ports, pathless reads, option ordering, and mutation-killing probes; the full
race suite passed in their isolated sandboxes. Both audits are closed with no findings.

The guarantee is deliberately a composition witness, not proof that an adapter describes the
physical corpus honestly or that the corpus remains fresh after construction. The existing
[source/path split](6gcwcf88z57p-split-local-path-capabilities-from-semantic-entity-reads.md)
must register each new path port in the same validator. Repo-local templates are not a shipped
planning-data adapter; they will need the same decision if introduced. `Layout` only supplies
watcher hints, while direct CLI fixer/linter channels are wired from the already validated FS
instance; neither is an independently selected service data port today. The Antigravity report's
`transfer=pending` text is the known review-protocol attestation issue tracked by
[make review transfer attestations self-finalizing](6g7srp3py9fe-make-adversarial-review-transfer-attestations-self-finalizing.md),
not a finding against this implementation.
