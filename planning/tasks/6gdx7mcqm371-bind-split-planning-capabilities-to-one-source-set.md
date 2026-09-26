---
schema: 1
id: 6gdx7mcqm371
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Prevent semantic reads, local paths, and mutations from being composed across unrelated planning corpora.
effort: 2-3 days
tier: 2
priority: medium
autonomy_level: 2
tags: [architecture, ports, composition]
created: "2026-09-26"
depends_on: [6gcwcf7rgxef]
updated_at: "2026-09-26"
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

- [ ] Explicitly pairing readers, path resolvers, or mutators from different source sets fails at
      construction with a typed, explanatory error.
- [ ] A complete filesystem adapter, a pathless read-only adapter, and test fakes can each expose an
      honest source-set identity without putting that value on domain records or wire output.
- [ ] Record IDs, locations, and source revisions cannot be mistaken for source-set identity.
- [ ] Replacing semantic reads still detaches implicit local capabilities; explicitly restoring a
      compatible capability remains deterministic regardless of option order.
- [ ] Thread, task, epic, audit, and research composition paths share the same rule and regression
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
