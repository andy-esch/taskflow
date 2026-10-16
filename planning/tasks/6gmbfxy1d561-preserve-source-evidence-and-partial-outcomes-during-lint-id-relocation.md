---
schema: 1
id: 6gmbfxy1d561
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Make lint ID spelling relocation no-clobber and mode-preserving, recheck source bytes, and report destination-written/source-retained failures.
effort: 1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, repair, persistence, concurrency]
created: "2026-10-16"
---
# Preserve source evidence and partial outcomes during lint ID relocation

## Objective

Harden the misspelled-ID relocation in `internal/store/fix.go` without changing
which spelling repairs are authorized. The October 16 concurrency audit's M1 and
M2 share this path; the earlier completed lint-fix locking task addressed cooperating
writers, not raw-edit CAS, permission preservation or per-file partial outcomes.

## Scope

- Preserve the source mode using exclusive destination creation; refuse an occupied
  destination, including one appearing after the availability check.
- Recheck original source evidence before destination creation and again before
  source removal. Never remove a concurrent raw edit.
- Report the destination-written/source-retained stage alongside errors and earlier
  batch progress. Choose an adapter-neutral maintenance receipt consistent with the
  existing maintenance port; assess wire compatibility before changing public output.
- Keep both copies for explicit inspection on a post-create conflict. Do not promise
  crash atomicity, add signal handling or automatically delete ambiguous records.
- Correct stale relocation/rename comments in the owning fixer and persistence guide.

## Acceptance criteria

- [ ] A 0600 source retains its mode after repair; destination collision tests prove no clobber.
- [ ] Deterministic pre-create and pre-remove source drift tests preserve concurrent bytes.
- [ ] Failed unlink or post-create conflict reports both paths and the durable stage without claiming completion or hiding prior repairs.
- [ ] Preview makes no changes; permission, authorization, preservation and normal quality gates pass.

## Evidence and sequencing

[Concurrency audit M1/M2](../audits/6gm8g0f66k80-2026-10-16-concurrency-and-atomicity.md).
The source confirms both missing safeguards. A process kill can still leave two
copies; reporting and non-destructive recovery, not eliminating that window, are
the contract. The audit's copied-valid-ID probe is not an actual spelling-repair
interruption: the original misspelled filename has no canonical source ID, so
that probe does not establish a new duplicate-canonical-ID failure. The original
was already unhealthy. The missing CAS and partial receipt remain valid defects.
Do not reopen the historical status-folder move tasks or expand
Finish selected identity safety: this is an independent maintenance-write task.
The broader entity-integrity foundations task may reuse the resulting evidence,
but is not a prerequisite.
