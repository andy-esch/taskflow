---
schema: 1
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: Once frontmatter is truth, drop the dir-vs-frontmatter cross-check; lint validates status/bucket (present + recognized), fail-open and flag (inherit the epic-status model). Per epic 24.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core]
created: "2026-07-01"
id: 6fhnydm034th
updated_at: "2026-07-03"
started_at: "2026-07-03"
completed_at: "2026-07-03"
---
# Replace Misfiled with frontmatter status/bucket validation

## Objective

Now that frontmatter `status` (tasks) / `bucket` (audits) is authoritative (ADR-0003
Phase A), a MISSING or UNRECOGNIZED frontmatter value silently falls back to the folder,
flagged nowhere — surfaced repeatedly in the Phase-A reviews, and a hard prerequisite for
the flat layout (where there's no folder to fall back to). Add lint validation that flags
it, fail-open (the fallback still lists the entity, it's just flagged), mirroring the
epic-status model.

## Done

- [x] `parseTask` / `parseAudit` record `StatusFellBack` / `BucketFellBack` when the
  frontmatter value is missing or unrecognized (they already fell back to the folder;
  the fallback is now observable to lint). A bool, so hand-built `Task`/`Audit` literals
  default to clean — no widespread test churn.
- [x] `FrontmatterStatusIssues(t)` / `FrontmatterBucketIssues(a)` flag it — fail-open, in
  ANY status (archived too), wired into `LintTask` + the archived branch + `LintAudits`.
  Removed the now-dead `if t.Status == ""` check.
- [x] Tests: domain unit + cli end-to-end (a missing task status and a missing audit
  bucket are each flagged); audit test fixtures gained a `bucket:`.

## Out of scope / deferred

- **Dropping the `Misfiled()` cross-check** (the "replace" in the title) — Phase B. In
  Phase A the directory mirror still exists, so dir-vs-frontmatter drift is a real,
  distinct check; `Misfiled` goes away only when the dirs do. Already captured in the
  flatten task's Phase B ("*removing* the Misfiled concept").
- **`lint --fix` backfilling a missing TASK status** from the folder (audits already
  backfill a missing `bucket:`). Tasks have always dual-stored `status:`, so a missing
  one is a rare hand-edit; the remedy is a lifecycle verb. A follow-on if we want the
  symmetry / for Phase B prep.
- A `bad_task_status` dashboard count (the epic model has `bad_epic_status`) — a
  Summary/dashboard parity nice-to-have.

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- [flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory](6fhnydm03edq-flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory.md) — the flat layout (Phase B) needs this (no folder fallback once flat), and owns the Misfiled removal.
