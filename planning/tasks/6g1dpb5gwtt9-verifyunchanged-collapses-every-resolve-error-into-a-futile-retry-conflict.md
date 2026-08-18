---
schema: 1
id: 6g1dpb5gwtt9
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: An ambiguous id or an I/O error becomes 'changed on disk; retry', burning four backoffs on advice that cannot work; also no create path validates id.Valid before writing.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [store, robustness]
created: "2026-08-18"
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

## Also: no create path validates the id it is about to write

`CreateResearch` guards `ID == ""` but not `id.Valid(ID)`. Confirmed:
`CreateResearch(Research{Slug:"bogus", ID:"abcde"})` returns nil and writes
`research/abcde-bogus.md`, which `ListResearch` then reports as `not an entity file` and
`resolveResearch` cannot find. Out-of-alphabet ids (`uuuuuuuuuuuu` — Crockford drops
i/l/o/u) behave the same.

Identical gap in `CreateTask` and `CreateAudit`, reachable via `WithIDGen` or any direct
store caller. Every MUTATE path parses before committing; no CREATE path does. Bundled here
because both findings are "the store trusts an input it should verify", shared across all
entities.

## Acceptance criteria

- [ ] `verifyUnchanged` distinguishes a real conflict (path moved / content changed) from a
      resolve FAILURE, and propagates the underlying error — an ambiguous id must surface as
      `ErrAmbiguous` (exit 13) naming the duplicate, not as a retryable conflict.
- [ ] `retryOnConflict` does not retry a non-conflict.
- [ ] The create paths reject an id that is not `id.Valid`, for task, audit, and research.
- [ ] Tests: an ambiguous id gives a useful error and does not sleep through four retries;
      an invalid id is refused at create.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- The failure chain this enables: see the epic-28 duplicate-stable-ids task
- Found by an independent adversarial correctness review, 2026-08-18
