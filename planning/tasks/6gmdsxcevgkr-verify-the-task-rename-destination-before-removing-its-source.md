---
schema: 1
id: 6gmdsxcevgkr
status: completed
epic: 21-code-quality-architecture-hardening
description: Refuse task-rename source deletion if a raw writer changed or removed the freshly written destination, retaining source data and truthful recovery.
effort: half day
tier: 2
priority: high
autonomy_level: 3
tags: [correctness, rename, persistence, concurrency]
created: "2026-10-16"
started_at: "2026-10-17"
updated_at: "2026-10-17"
completed_at: "2026-10-17"
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

- [x] Red-first destination-deletion and destination-edit probes preserve the intact source and refuse completion.
- [x] Returned errors retain original causes and truthful destination-written/source-not-removed progress; CLI advises inspection rather than blind retry.
- [x] Existing rename/cascade, preview, authorization and normal quality gates pass.
- [x] Independent review challenges destination/source preservation and
  recovery; meaningful findings are fixed or explicitly tracked.

## Sequencing

Found while implementing `6gmbfxy1d561`, which adds this check for lint spelling
relocation only. This task owns the separate task-rename path. Independently
actionable; do not add a prerequisite solely for delivery order.

## Implementation checkpoint (2026-10-17)

Started from merged main `36348b5b` (PR #312), on
`fix/verify-task-rename-destination`. Retained compiler-valid red-first tests
reproduced destination deletion/edit/replacement reporting success and removing
the old source, with and without a committed inbound-link prefix.

The adapter now verifies the destination's regular-file type and exact materialized
bytes before the existing source CAS and unlink. A byte-identical symlink is not
accepted as a planning record: scans exclude it. Missing, changed and non-regular
destinations refuse source removal, preserve concurrent edits and keep the already
committed cascade/destination actions in the existing partial receipt. Original
non-ENOENT filesystem causes remain wrapped. No rollback, crash transaction or
inode-continuity promise was added; the final check-to-effect windows remain.

Core's recovery message describes inspection and reconciliation of surviving paths,
not guaranteed current existence. The low-level CAS helper no longer independently
advises retry. CLI recovery suppresses the filesystem retry hint after destination
commit, including synthetic transient errors; the pre-destination convergent prefix
keeps its existing retry guidance. Portable/core and CLI-port fixtures are distinct
from the real-store drift probes; `EBUSY` is not claimed as a real OS reproduction.

This restores the existing completion/source-removal and recovery promises. No
wire key, type, requiredness, vocabulary, selector or successful-input meaning
changes; the wire revision remains 1.89. Existing schemas and goldens are unchanged.

### Evidence

- `TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt`: deletion, edit,
  directory replacement and byte-identical symlink, each with/without cascade;
  checks intact 0600 source bytes, untouched raw replacement, exact historical
  counts/paths, incomplete receipt and no replay.
- `TestVerifyTaskRenameDestinationPreservesFilesystemCause`: real `ENOTDIR` from
  path inspection retains its `PathError` and failing path.
- `TestRenameTaskPreservesCauseAndSelectsRecoveryFromDurableActions`: pathless
  portable outcomes retain causes/counts and select recovery by durable stage.
- CLI command and JSON-error tests preserve structured receipts, human inspection
  advice, cause classification, and distinguish pre-/post-destination retry hints.

Owner mutation checks all compiled and failed the intended tests: removing the
production destination check lost the source after deletion; bypassing the type
check accepted the byte-identical symlink; removing CLI retry suppression advertised
replay after commit. All were restored, then focused race tests passed (`-count=3`).

`just build`, full `just test` (race), `just lint` (0 issues), `just docs-check`,
schema-comment generation and focused core/store/CLI/wire races pass. Normal
rename/cascade, no-op/same-slug, preview and authorization regressions remain green.
All test fixtures use automatic temporary-directory cleanup; no planning spaces
were registered. Independent review remains before completion/PR handoff.

### Independent review handoff

- [Claude: persistence/source-deletion ordering](../audits/6gmke6f26gn2-2026-10-17-task-rename-destination-safety-implementation-claude.md).
- [Antigravity: core/CLI receipt and retry contract](../audits/6gmke6f7nnw0-2026-10-17-task-rename-destination-safety-implementation-antigravity.md).

At handoff, both audits contained prompts only, no findings or verdict. The briefs require
independent clones, exact source declarations rather than reconstructed API
inventories, compiler-valid mutation evidence and truthful synthetic/real-probe
labels. The narrower inventory/trace requirement responds to material baseline
mismatches in the previous review round, not an assumption about model capability.
No reviewers were launched by the owner. Reviewers work from the captured snapshot;
task and Thread remain in progress with the independent-review AC unchecked.

### Antigravity owner triage (2026-10-17)

Verified the helper's successful guarded transfer receipt for sandbox
`/private/tmp/isolated-review.rKZHgW`, baseline `aff2bda6ca13fdf8d94bd4a0e6c4356f155893a3`,
before changing the returned audit. M1 is valid: the retained red-first test
`TestRenameTask_PreexistingDanglingDestinationRejectsBeforeCascade` reproduced a
committed inbound-link prefix on live rename and a false-positive preview when a
dangling symlink already occupied the target. `Lstat` preflight now refuses both
before effects, preserving source, inbound bytes and the foreign symlink. The
exclusive-create symlink regression now injects after planning, independently of
preflight, rather than relying on its old false negative.

The review also demonstrated that substituting ordinary creation for exact-mode
relocation survived 0600-only rename fixtures. Added a subprocess-isolated Unix
rename test with 0664 under umask 022: that compiler-valid mutation now fails
(0644 instead of 0664), then passes with the production exact-mode call restored.
Broader `Stat` preflights and lint relocation's symlink-following verifier are
explicit investigation items in `6g7wsr8yvt1a`; no shared guarantee is claimed fixed.

Antigravity's report is preserved as evidence about its captured baseline. At this
checkpoint the second audit was awaiting review and remained untouched; it did not
yet cover these new owner edits. The independent-review AC remained unchecked.

### Final review and completion (2026-10-17)

Codex completed the reassigned Claude persistence brief without consulting its
sibling report. Verified successful guarded delivery from sandbox
`/private/tmp/isolated-review.TnXuCp`, baseline `2a01d3675f4d37e8bb646d37e62f576d737d83c8`.
No additional actionable findings: independent raw-file recovery probes and
compiler-valid mutations challenged destination verification, regular-file checks,
source CAS and post-destination retry advice. Retained the audit's original filename
for continuity and recorded the reviewer reassignment in its owner disposition.

Both reviews are settled. Later owner fixes have separate red-first and mutation
evidence above; neither historical review is represented as covering them. Fresh
`just build`, full race `just test`, `just lint` and `just docs-check` pass. Planning
links and audit findings are linted before PR handoff. The maintenance Thread stays
open for its independently actionable in-place lint repair member `6gmdsxccqqnm`.
