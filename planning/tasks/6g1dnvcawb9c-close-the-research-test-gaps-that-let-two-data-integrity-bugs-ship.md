---
schema: 1
id: 6g1dnvcawb9c
status: completed
epic: 28-first-class-entities-new-planning-nouns
description: '13 of 30 mutations survived the suite: the CAS, flock, retryOnConflict, research edit, and the mintable-date seam are untested, plus five tests assert less than their names claim.'
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [testing, core]
created: "2026-08-18"
updated_at: "2026-08-19"
completed_at: "2026-08-18"
---

# Close the research test gaps that let two data-integrity bugs ship

## Objective

An independent mutation-testing review ran 30 mutations against the research code; **13
survived the full suite**. Two of the survivors are the direct reason the id-collision and
frontmatter-fabrication bugs shipped. Close the gaps, and fix the tests whose names claim
more than they check.

## Whole guarantees with ZERO coverage (mutation survived)

- **The version-CAS.** Deleting the `verifyUnchanged` call from `SetResearchFields` and
  stubbing both closures in `EditResearch`/`AppendResearchBody` to `return nil` leaves the
  suite green. Tasks have `store/occ_test.go`; audits have `store/audit_edit_test.go`;
  research has nothing. Research's write paths also lack the
  `testHookBeforeSetFieldsWrite`/`testHookBeforeBodyWrite` seams tasks use to reach the CAS
  window from a test — so add the seams, then the test.
- **The write flock.** Removing the `writeLock()` block and passing a no-op: suite green.
- **`retryOnConflict` in core.** Unwrapping both calls to direct store calls: suite green.
- **`research edit` — no test at all.** Nothing exercises the editor loop's
  parse-before-accept, the reopen-on-error path, the `changed == false` branch, or the
  success path. Mirror `store/audit_edit_test.go` + `cli/edit_test.go`.
- **The mintable-date seam.** Swapping `ValidateMintableDate` -> `ValidateDate` in
  `NewResearch`: suite green. `mintdate_test.go` tests the domain function in isolation and
  nothing asserts the CALLER uses it, so the guard shipped two PRs ago is unpinned. The
  existing "bad date" case (`2026-6-3`) can't distinguish the two validators — it needs a
  case like `research new --created 9026-06-15` exiting 11.

## Tests that claim more than they verify

- `TestFS_SetResearchFields_PreservesUnknownKeysAndOrder` — **never checks order**.
  Reversing every key in `updateFrontmatter` still passes (four independent
  `strings.Contains` calls). Compare the frontmatter block byte-for-byte instead.
- `TestFS_SetResearchFields_RefusesUnreloadableUpdate` — claims it "must blame the update
  rather than the file", but only asserts `errors.Is(err, ErrValidation)`, which both
  branches satisfy; collapsing them to always blame the file still passes. Also its
  `t.Skip` does not fire today but is a latent hazard: if the YAML round-trip ever makes
  that value parse, the guard goes unverified behind a green suite. Use a fixture that
  provably breaks the parse and assert on "would not reload".
- `TestResearchEdit_RejectsDryRunAndNonTTY` — **does not test the dry-run rejection**.
  Deleting the whole `if app.DryRun` block still passes, because the harness has no TTY so
  it errors anyway. Assert the message names dry-run, as
  `TestDryRun_TaskEditRejected` does.
- `TestValidateMintableDate_Boundaries` — claims it pins both edges "exactly", but
  `maxDay` is `Truncate(24h)`'d so it sits strictly below `id.MaxMillis`: changing
  `<= MaxMillis` to `< MaxMillis` passes all 8 subtests. (The min edge IS pinned.)
- The unknown-field gate on the **unset** path is untested — the unset subtests only use
  PROTECTED fields, which are rejected earlier. `research set --unset bogsu` would
  silently no-op.
- Three CLI input guards unexercised: empty append text, `--set` requiring `key=value`,
  and the same key given to both `--set` and `--unset`.

## Acceptance criteria

- [x] Research OCC/conflict test at the store layer and a `retryOnConflict` test in core.
- [x] A `research edit` test file covering parse-before-accept, reopen-on-error, no-change,
      and success.
- [x] A CLI case pinning that `research new` rejects an out-of-range `--created`.
- [x] The five weak assertions above fixed, each verified to FAIL against the mutation
      that previously survived.
- [x] Coverage for the unset-path force gate and the three CLI input guards.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md)
- The two bugs these gaps let through: [duplicate-stable-ids-brick-research-docs-silently-and-unrecoverably](6g1dnnfgyjap-duplicate-stable-ids-brick-research-docs-silently-and-unrecoverably.md),
  [research-write-paths-accept-frontmatter-less-files-and-fabricate-incomplete](6g1dnqxefhtz-research-write-paths-accept-frontmatter-less-files-and-fabricate-incomplete.md)
