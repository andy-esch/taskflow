---
schema: 1
id: 6gmdsxccqqnm
status: completed
epic: 21-code-quality-architecture-hardening
description: Add raw-edit CAS and checked guard handling to ordinary lint fixes while retaining completed repairs and refusing whole-batch replay.
effort: 1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, repair, concurrency, persistence]
created: "2026-10-16"
started_at: "2026-10-17"
updated_at: "2026-10-17"
completed_at: "2026-10-17"
---
# Guard in-place lint repairs and retain guard finalization failures

## Objective

Close the ordinary `FixFrontmatter` write path's remaining raw-edit and guard
finalization gaps without reworking spelling relocation or inventing a generic
mutation framework. Cooperating writers are serialized today, but the shared
in-place `writeFileAtomic(path, fixed, ...)` still has no immediate source CAS.

## Scope

- Capture each original file version and refuse raw byte drift immediately before
  in-place replacement. Preserve malformed-document repair, comments, unknown keys,
  permissions, graph-field refusals, preview and no-op behavior for every entity kind.
- Use the checked repository guard so authorization is current after waiting; join
  guard-release errors with the original failure and retain completed batch evidence.
- Keep partial relocation distinct from completed in-place repairs. Never replay
  the whole batch after durable progress or advertise such a retry to automation.
- Add deterministic fault seams/probes before claiming the raw-editor scenario is
  reproduced. This follow-up is currently supported by source inspection, not a
  timed concurrent-editor reproduction.

## Acceptance criteria

- [x] A red-first raw-edit probe preserves newer bytes and refuses stale in-place fixes.
- [x] Post-wait authorization changes refuse effects; release failure retains completed repairs and the original error.
- [x] Dry-run/no-op, graph-owned field refusal, mode and source-preservation tests stay green.
- [x] Core/CLI recovery and normal quality gates pass without whole-batch retries.

## Sequencing

Discovered beside `6gmbfxy1d561` (lint ID relocation). It is independently
actionable; no artificial prerequisite. The broader entity-integrity design may
reuse this evidence, but must not block a bounded safety fix. Do not reopen the
completed selected-identity Thread.

## Implementation checkpoint (2026-10-17)

Started from merged main `43af9bb1` on `fix/guard-in-place-lint-repairs`.
Retained red-first production-hook tests reproduced newer bytes being overwritten,
deleted sources resurrected, byte-identical symlinks replaced, and revoked policy
being ignored after waiting for the process guard. A directory replacement already
refused the write, but without the intended conflict classification. No real
planning records were used for hostile probes.

`FixFrontmatter` now reauthorizes under the checked guard and joins release errors
with any primary error. In-place repairs check regular-file type and captured
source bytes immediately before the existing atomic helper. No typed parse gate
was added: malformed YAML remains repairable. Tests cover all five entity kinds,
restrictive mode, comments/unknown fields/body preservation, preview, idempotence,
graph-field refusal, earlier completed rows and separate partial-relocation evidence.
Raw writers still retain the final check-to-replace window; shared create cleanup,
relocation symlink semantics and unrelated legacy writers remain out of scope.

Core classifies completed repairs/cleanup and written relocation destinations as
historical committed progress; skipped rows and preview proposals never count.
Failed ordinary prefixes carry inspection/plain-lint guidance, while partial moves
retain their specific reconciliation advice. CLI stdout/error recovery retains the
existing receipt after repair, post-lint and returned JSON-writer failures, and
suppresses transient filesystem retry hints after commit. Tests use explicit
synthetic `EBUSY`/`EAGAIN`; no actual kernel busy-file failure is claimed.

No JSON keys, types, vocabularies or registered exit codes are added. The internal
commit flag is not serialized; the populated ordinary-prefix recovery validates against the
existing schema. Revision 1.89 is retained because these fixes restore completed-
prefix/no-whole-batch-retry promises using the existing optional recovery field,
not a new receipt format. Independent review must challenge that classification.

### Validation and limits

- `just build`, full race `just test` and `just lint` pass; focused
  store/core/CLI/wire repair tests also run repeatedly under `-race -count=3`.
- Compiler-valid mutations bypassing byte CAS, discarding the joined release error,
  and bypassing committed-prefix retry suppression each fail the intended retained
  test. All restored before the final focused run. The first release-error mutation
  left an unused import and was invalid; only the corrected compiling probe counts.
- `just docs` regenerated the lint page. `just docs-check` reports exactly that
  intentional unstaged help diff against the index, not generator instability;
  separate generation into a temporary directory matches `docs/cli` byte-for-byte.
  At PR handoff, `just docs-check` passes with the generated page staged normally.
  Schema-comment generation produces no tracked change. Planning/link lint passes.
- All hostile entity fixtures use `t.TempDir`; no planning spaces were registered.
  Guard-release failure is hook-injected; `ENOTDIR` source inspection is a real OS
  error. No crash transaction, inode continuity or universal mutation policy is claimed.

Implementation is checkpointed for independent review; task and maintenance Thread
remain in progress. AC4 stays unchecked until review and final validation settle.

### Independent review handoff

- [Codex: persistence and authorization](../audits/6gmnhv65r2f0-2026-10-17-in-place-lint-repair-safety-implementation-codex.md).
- [Antigravity: core/CLI recovery](../audits/6gmnhv6b5da7-2026-10-17-in-place-lint-repair-safety-implementation-antigravity.md).

At handoff, both contained prompts only, no findings or verdict. Briefs require independent
captured clones, complementary hostile probes, compiler-valid mutations, populated
schema/semantic checks, an executed plain-lint recovery, and guarded transfer receipts.
No reviewers were launched. Freeze implementation and assigned audits during review;
later owner fixes must be distinguished from each report's captured baseline.

### Codex owner triage (2026-10-17)

Confirmed guarded delivery from `/private/tmp/isolated-review.cAXHXI`, baseline
`4802130c5079cdaa84aabdd902c7f963e3b0b119`, before changing the returned audit.
Accepted its bounded no-demonstrated-findings assessment after checking the actual
producer/consumer code and independent probe evidence. Its four compiling mutations
challenged byte CAS, post-wait authorization, release-error joining and symlink type
checks. Additional independent probes exercised combined denial/release errors,
real store-to-CLI `ENOTDIR`, returned synthetic writer errors, populated wire
semantics and the plain-lint recovery against preserved concurrent bytes.

No source changes were needed. Reviewer-only probes are not represented as retained
regressions. Codex also independently passed `just docs-check` against the captured
baseline; the intentional generated page is reproducible. This audit is settled;
Antigravity and AC4 remain pending, so the task and maintenance Thread stay open.

### Antigravity triage and final validation (2026-10-17)

Confirmed guarded delivery from `/private/tmp/isolated-review.QX35Vz`, baseline
`f582588c0e98259e0896a34540c266a493269658`. Accepted and fixed M1/L1:

- Filesystem diagnostics had combined the matched error's operation/path with
  another joined cause's class/retryability. They now classify that same matched
  cause. Retained tests cover PathError/LinkError pairs and lint command failures
  across skipped, preview and committed prefixes.
- A JSON-writer failure after successful repair had hidden residual validation
  (exit 1 rather than 11) and ordinary committed-prefix guidance. Validation is
  now joined with output failure, and a shared core result recovery method retains
  causes and advice without duplicating it or replacing relocation guidance.
  Clean, unreadable and remaining-issue cases have separate assertions.

The affected writer-failure path now reports the existing validation exit 11
instead of generic exit 1 when residual validation is present. This is a restored
cause/classification, not a new exit-code mapping or vocabulary.

Both defects reproduced before the fixes. Full race `just test`, `just build`,
`just lint` and focused core/CLI/wire races (`-count=3`) pass afterward. Both audits
are settled; later owner fixes are not attributed to their historical baselines.
The generated-help/index caveat above is unchanged. This completes the final
member of the bounded maintenance Thread; broader investigation stays in its
existing scoped tasks, not a new implicit prerequisite for feature work.
