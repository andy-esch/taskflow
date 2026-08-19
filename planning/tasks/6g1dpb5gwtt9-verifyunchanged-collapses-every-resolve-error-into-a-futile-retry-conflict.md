---
schema: 1
id: 6g1dpb5gwtt9
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: An ambiguous id or an I/O error becomes 'changed on disk; retry', burning four backoffs on advice that cannot work.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [store, robustness]
created: "2026-08-18"
updated_at: "2026-08-19"
---

# verifyUnchanged collapses every resolve error into a futile "retry" conflict

## The defect

`internal/store/cas.go`:

    if err != nil || curPath != path { return conflict() }

Any error from the re-resolve — `ErrAmbiguous`, or a real `ReadDir` I/O failure — becomes
`"<entity> \"x\" changed on disk during <op>; retry: conflict"`. The same function is
careful to distinguish `os.IsNotExist` from a genuine read error seven lines later, so the
inconsistency is the defect rather than a deliberate simplification.

Cost: `retryOnConflict` burns its four backoff sleeps and then hands the user advice that
can never work, for a condition that is not a conflict. This is the mechanism that makes a
duplicate stable id **unrecoverable** rather than merely annoying — the sibling epic-28 task
documents that chain end to end.

## Moved out: create-path id validation

This task originally also carried "no create path validates `id.Valid` before writing"
(`CreateResearch(Research{ID:"abcde"})` writes `research/abcde-bogus.md`, which then reads
back as `not an entity file`; same in `CreateTask`/`CreateAudit`).

That belongs to [invalid-non-crockford-entity-id-is-mishandled](6fq9zy126tas-invalid-non-crockford-entity-id-is-mishandled.md), whose
final acceptance criterion is already *"Ids are validated at write time so a bad id can't
persist"* — the same fix. That task is older, higher priority, and owns the whole
invalid-id story (diagnosis wording, not aborting the command, `lint --fix` repair), so
write-time validation is a natural part of it rather than a rider here. Removed from this
task's criteria to avoid two epics fixing one thing.

## Acceptance criteria

- [ ] `verifyUnchanged` distinguishes a real conflict (path moved / content changed) from a
      resolve FAILURE, and propagates the underlying error — an ambiguous id must surface as
      `ErrAmbiguous` (exit 13) naming the duplicate, not as a retryable conflict.
- [ ] `retryOnConflict` does not retry a non-conflict.
- [ ] Tests: an ambiguous id gives a useful error and does not sleep through four retries.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Create-path id validation now lives in [invalid-non-crockford-entity-id-is-mishandled](6fq9zy126tas-invalid-non-crockford-entity-id-is-mishandled.md)
- The failure chain this enables: see the epic-28 duplicate-stable-ids task
- Found by an independent adversarial correctness review, 2026-08-18
