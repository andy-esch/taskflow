---
schema: 1
id: 6gmdsxcevgkr
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Refuse task-rename source deletion if a raw writer changed or removed the freshly written destination, retaining source data and truthful recovery.
effort: half day
tier: 2
priority: high
autonomy_level: 3
tags: [correctness, rename, persistence, concurrency]
created: "2026-10-16"
---
# Verify the task rename destination before removing its source

## Objective and confirmed evidence

Keep `RenameTask` from removing its last intact source when a raw writer changes
or deletes the destination between creation and source removal. Existing source
CAS is insufficient: it proves only that the old copy has not changed.

Owner reproduction on 2026-10-16, from `e454d00b` plus the lint-relocation work:
`testHookBeforeTaskRenameSourceRemove` deletes the freshly created destination.
`RenameTask` returns nil with `SourceRemoved=true` and `Complete=true`, and both
source and destination are absent. The compiler-valid throwaway probe
`TestOwnerProbeRenameDeletesSourceAfterDestinationDisappears` reproduced this
under `go test -race -count=1 -v`; it was then removed, not retained as passing
safety coverage. No real planning records were used.

## Scope

- Verify the destination still contains this rename's materialized bytes immediately
  before deleting the old source. Refuse destination deletion/replacement without
  overwriting raw changes; retain the original source and accurate durable receipt.
- Preserve source CAS, source mode, exclusive destination creation, inbound-link
  prefix reporting, dry-run and committed failure recovery.
- Document that action flags record historical actions, not current file existence.
  Do not promise crash atomicity or elimination of the final check-to-unlink race.

## Acceptance criteria

- [ ] Red-first destination-deletion and destination-edit probes preserve the intact source and refuse completion.
- [ ] Returned errors retain original causes and truthful destination-written/source-not-removed progress; CLI advises inspection rather than blind retry.
- [ ] Existing rename/cascade, preview, authorization and normal quality gates pass.

## Sequencing

Found while implementing `6gmbfxy1d561`, which adds this check for lint spelling
relocation only. This task owns the separate task-rename path. Independently
actionable; do not add a prerequisite solely for delivery order.
