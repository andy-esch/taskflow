---
schema: 1
id: 6gfrcytd9n9a
bucket: open
area: test-rigour
date: "2026-10-02"
updated_at: "2026-10-05"
---

# Audit: test-rigour — 2026-10-02

> Create findings with `audit finding new`; update them with `audit finding`.

## Findings

<!-- Example grammar; let `audit finding new` allocate and render real findings: -->

```
#### H1. <title>  · **Status:** open

**File:** <path:line> | **Component:** <component>
**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>

<what's wrong, why it matters, evidence>

**Recommendation:** <minimum fix>

**Resolution:** <how it was resolved — written by `audit finding --note`, not by hand>
```

#### M1. `thread compose`/`apply` refuse malformed and stale plans, and 24 of 37 refusals have no test · **Status:** fixed (locally; pending independent review and merge)

**File:** internal/core/thread_apply.go:214 | **Component:** core/thread-apply
**Effort:** S · **Urgency:** soon

The `thread compose` → `thread apply` pipeline is the widest mutation surface in the
repo — it writes repository-global `depends_on` edges and then a Thread document under
one guard. Its validation is real and it works: I drove eleven of its refusals through
the shipped binary against a throwaway planning repo and every one fired with the
right sentinel, exit 11, and an actionable message. The gap is that almost none of them
is pinned by a test.

Cross-package guard-coverage tracing over `internal/core/thread_apply.go` leaves
**24 of its 37 sentinel-returning refusal blocks with zero executions**.

The existing rejection tests are aimed at the invariants that were *hard to get right*,
and they cover those well — see *What audited clean*. What has no test is the
structural layer beneath them. On `compose`, the manifest shape (every message below
observed from the shipped binary):

| Predicate | line | Verified refusal |
| --- | --- | --- |
| manifest has no nodes | 214 | `Thread manifest requires at least one node` |
| node has no local key | 247 | `every manifest node requires a local key` |
| duplicate node key | 249 | `duplicate manifest node key "a"` |
| `task_id` is not an exact stable id | 251 | `node "a" task_id "alpha" is not an exact stable task id` |
| node references a missing task | 254 | `node "a" references missing task 6zzzzzzzzzzz` |
| no member nodes | 269 | `Thread manifest requires at least one member node` |
| dependency `from` is an unknown key | 279 | `dependency references unknown local key "zz"` |
| dependency `to` is an unknown key | 282 | `dependency references unknown local key "zz"` |
| self-dependency | 286 | `task 6gfr9parw1bq cannot depend on itself` |
| duplicate dependency | 289 | `duplicate manifest dependency a -> b` |

And on `apply`, the revalidation of a plan that is **stale** rather than tampered with.
`thread apply --help` promises that "an interrupted plan is safe to retry from the same
durable file", so a plan is expected to outlive the moment it was composed — and the
graph underneath it can move. Untested: `:347` (planned Thread is not `unstarted`),
`:373`–`:378` (planned dependency ids not exact, self-edge, duplicate), `:381`/`:385`
(planned prerequisite / dependent no longer exists), `:429` (authoritative body
unavailable).

**Failing scenario:** verified end-to-end on the shipped binary. Compose a plan over
tasks `alpha` + `beta` with `a -> b`; delete `planning/tasks/<beta>-beta.md`; then
`thread apply plan.json` → exit 11, `validation failed: planned dependent
6gfr9pb2jg5n does not exist`. Correct, and asserted nowhere: deleting `:381`–`:387` in
a throwaway copy of the tree left `go test ./...` fully green. Separately, editing a
composed plan's `status: unstarted` to `in-progress` → exit 11, `planned Thread must be
unstarted, got "in-progress"` — also unasserted.

**Context (why Medium, not High):** the blast radius is bounded by defence in depth.
With `:381`–`:387` removed, the dangling edge is still refused one layer down by
`internal/core/dependency_graph_mutation.go:77` (`planned dependency %s for task %s
does not exist`), which I confirmed by applying a stale plan with the unguarded binary.
A regression here costs the precise, early, actionable message and defers the refusal
to a later, vaguer one — it does not corrupt the task DAG. Urgency is `soon` rather
than `eventually` because epic 30 is at 95% and this file is still being extended.

**Why tests didn't catch it:** the rejection tables were written against the subtle
cases, so the structural predicates — each individually obvious — were left to the
happy-path tests, which never construct a malformed manifest.

**Fix sketch:** extend the two existing tables rather than adding new test
functions. `TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest` already takes
a `{manifest, want}` pair, so each of the ten structural rows is about two lines; add
`status: in-progress`, a deleted prerequisite, and a deleted dependent to
`TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity`.

**Tightening (adjacent):** that Compose table asserts only
`strings.Contains(err.Error(), tc.want)` with no sentinel check, so a row that began
returning the wrong error *class* — and therefore the wrong exit code — would still
pass as long as the wording survived. The Prepare tests already assert
`errors.Is(err, domain.ErrValidation)`; make the Compose table match.

**Follow-up:** `thread_apply.go` is 546 lines running two validation passes over
overlapping vocabulary (manifest-local keys vs. stable ids). Whether `Compose` and
`Prepare` should share one declared predicate table instead of two hand-maintained
sequences is an architecture question, not a test gap — route it to the architecture
audit if it recurs.

**Recommendation:** Extend the two existing rejection tables: add the ten structural manifest rows to TestComposeThreadApplyPlanRejectsMisleadingOrInvalidManifest, and a tampered status plus a deleted prerequisite and dependent to TestPrepareThreadApplyRejectsUnsafeOrEditedCreationIdentity — asserting errors.Is(err, domain.ErrValidation) alongside the substring.

**Resolution:** Reachable compose/prepare guards now have sentinel and
diagnostic assertions with valid baselines. Eleven real service/store
edited-plan or stale-corpus cases prove preview and refusal preserve repository
entries/bytes/modes and the full typed plan receipt. Compiler-valid compose,
endpoint, body-evidence, and lifecycle probes fail for the intended reason;
redundant endpoint/lifecycle checks are explicitly diagnostic protection.

#### M2. Thread creation's committed-conflict no-retry guard is untested while its mutation twin is covered · **Status:** fixed (locally; pending independent review and merge)

**File:** internal/core/service_thread.go:93 | **Component:** core/thread-mutations
**Effort:** XS · **Urgency:** soon

`runThreadCreationMutation` and `runThreadMutation` carry the same three-term retry
condition, and the third term is the data-safety one:

```go
for attempt := 1; attempt <= s.maxRetries && errors.Is(err, domain.ErrConflict) && !result.Committed; attempt++ {
```

`!result.Committed` is what stops the service retrying a write that already landed.
The mutation path proves it; the creation path does not.

Two sibling tests are named almost identically and look equivalent, but differ
decisively:

- `store.TestThreadMutationAttributesReleaseFailureAfterCommit` injects
  `fmt.Errorf("…: %w", domain.ErrConflict)`, drives it through
  `core.MustNewService(...)` with `core.WithRetry(3, …)`, and asserts `retries == 0`.
  The injected error *is* a conflict, so the second term is true and `!result.Committed`
  is the only thing left holding the loop shut. Deleting it turns that test red
  (`retries=3`). This is the test M2 wants copied.
- `core.TestServiceThreadCommittedFailureIsNotRetried` injects
  `errors.New("unlock failed")` — **not** a conflict. `errors.Is(err,
  domain.ErrConflict)` is therefore false and short-circuits the conjunction, so the
  loop never runs regardless of the third term. Its `fake.calls != 1` assertion passes
  for the wrong reason.
- `store.TestThreadCreationAttributesReleaseFailureAfterCommit` is the creation-side
  twin by name, but it calls `NewFS(root).MutateThreadCreation(...)` **directly on the
  store**, bypassing `core`'s retry loop entirely, and asserts nothing about retry
  count. It correctly proves the store attributes a post-commit release failure and
  that the committed Thread stays readable — a different, also-valuable claim.

The two CLI tests with `CommittedFailure` in their names
(`TestThreadCreationCommittedFailureHasStructuredRecovery` and its mutation twin)
hand-build the receipt and never invoke the service, so they cover the error *rendering*
only.

**Failing scenario:** verified by mutation probe. With `&& !result.Committed` deleted
from `service_thread.go:93` only, `go test ./...` is fully green. A regression therefore
ships silently, and the behaviour it unlocks is concrete: a release failure that wraps
`ErrConflict` after the Thread file is already durable gets retried, each retry
re-enters `MutateThreadCreation` with the same planner and the same `now`, the stem now
exists on disk, and `core/thread_creation.go:131` refuses it. The user is told
`thread id 6gfr82z0f37a is already used by …` for a Thread that *this same command just
created successfully* — and the `ThreadCreationMutationFailure` receipt that carried the
committed path is replaced by the last attempt's, so `receipt.Committed` flips to
`false` and the CLI's structured recovery loses the path. I reproduced exactly that:
`retries=3 committed=false err=thread id … is already used by …: conflict`.

**Why tests didn't catch it:** the creation-side test injects a generic error where its
mutation-side twin injects a conflict-wrapping one, and only the conflict reaches the
guard. Nothing about either test's name or shape reveals the difference.

**Fix sketch:** add the creation-side twin of the mutation test — drive
`svc.NewThread` through `core.MustNewService(NewFS(root), …WithRetry(3, …))` with
`testHookRepositoryUnlockError` returning a conflict-wrapping error, and assert
`errors.As(err, &*core.ThreadCreationMutationFailure)`, `receipt.Committed`, and
`retries == 0`. I wrote and validated exactly this test in the probe copy: it fails
with the guard removed and passes with it restored. It belongs in
`internal/store/threadcreation_test.go` beside the existing
`…AttributesReleaseFailureAfterCommit`.

**Tightening (adjacent):** give `core.TestServiceThreadCommittedFailureIsNotRetried` the
table treatment `core.TestLifecycleCommittedFailureIsNeverRetriedAndRetainsReceipt`
already uses for `service_task.go:539` — it loops over both
`errors.New("unlock failed")` and `fmt.Errorf("unlock failed: %w", domain.ErrConflict)`,
which is precisely the pair that makes the guard load-bearing. That is the third retry
loop of the four, and the only one whose guard is properly pinned at the core level.

**Recommendation:** Add the creation-side twin of store.TestThreadMutationAttributesReleaseFailureAfterCommit: drive svc.NewThread through core.MustNewService with WithRetry and a conflict-wrapping testHookRepositoryUnlockError, asserting ThreadCreationMutationFailure, receipt.Committed, and retries == 0.

**Resolution:** Generic and conflict-wrapping post-commit cleanup failures now
run through a portable fake and real service/store, preserving committed
identity/local receipt with one mint and no retry. Removing only the committed
guard kills both conflict rows; generic controls stay green. Restored tests
pass.

#### M3. `theme` can silently fall back to ambient presentation on a bad `--space`, untested · **Status:** fixed (locally; pending independent review and merge)

**File:** internal/cli/theme.go:41 | **Component:** cli/space-selection
**Effort:** XS · **Urgency:** soon

Three commands resolve a planning repo best-effort and fall back when discovery fails,
and all three carry the same refusal so an *explicit* selection cannot be silently
downgraded. `internal/cli/template.go:36` states the invariant plainly:

```go
// Repo discovery is optional; an explicit registry selection is not. A
// typo or drifted --space must never turn into a built-in-only success.
if app.wantsSpace() {
    return err
}
```

`internal/cli/theme.go:41` carries the identical guard with the identical rationale:

```go
// Ordinary cwd discovery is best-effort, but an explicit space selection
// is an address assertion and therefore cannot silently fall back.
if err := app.resolve(); err != nil && app.wantsSpace() {
    return err
}
```

`template`'s guard is tested. `theme`'s is not. `TestGlobalSpace_UnknownListsKnownLabels`
ends with exactly the right assertion — and only for `template`:

```go
_, _, err = runSelection(t, "--space", "missing", "template", "list")
if err == nil || ExitCode(err) != 10 {
    t.Fatalf("template swallowed explicit bad space: %v", err)
}
```

No test anywhere passes `--space` to `theme`; it is exercised only via `-C`
(`command_safety_test.go`) and bare (`userconfig_test.go`).

**Failing scenario:** verified by building the unguarded binary and comparing against
the shipped one, same `TSKFLW_CONFIG_HOME`, no registered spaces:

```
shipped:          tskflwctl --space typo theme list
  → exit 10, error: not found: unknown space "typo" — none are registered; run `space add`
guard removed:    tskflwctl --space typo theme list
  → exit 0, and a theme list (catppuccin / miami-vice / neon (default, active))
```

Removing the guard leaves `go test ./...` fully green. The silent-success mode is the
damaging one: the caller named a space, got exit 0 and a plausible answer, and is
reading the *ambient* theme configuration while believing it is that space's — which
for `theme list` includes which theme is marked `active`. Epic 25 (65%) and epic 29
(88%) are both live and intersect precisely here.

**Why tests didn't catch it:** the invariant is per-command but its test lives in one
shared table keyed by command, so a command added to the guard does not automatically
gain a row. `theme`'s guard was written correctly and simply never got its line.

**Fix sketch:** add the two-line `theme` case beside the existing `template` one in
`TestGlobalSpace_UnknownListsKnownLabels`. Same shape, same expected exit code.

**Tightening (adjacent):** `internal/cli/ui.go:37` is the third site and has a different
shape (`isCompletionCommand(cmd) || app.wantsSpace() || app.Chdir != ""`), so it does
not reduce to the same two-line case. Worth a glance while in the file, but it is
guarding entry into the TUI rather than a read that can quietly answer wrong.

**Follow-up:** three hand-maintained copies of one invariant is the shape that produced
this gap. Whether best-effort-with-explicit-override belongs in `App.resolve()` itself —
so no command can forget it and the test is written once — is a design call for the
architecture audit.

**Recommendation:** Add the two-line theme case beside the existing template case in TestGlobalSpace_UnknownListsKnownLabels, asserting exit 10 on --space missing.

**Resolution:** Explicit unknown-space tests cover theme list and noninteractive
preview alongside template list, requiring ErrNotFound, exit 10, and no ambient
output. Removing only the theme refusal guard returns ambient success and kills
both theme rows. Restored tests pass.

#### L1. The exit-taxonomy smoke test's filesystem probe is vacuous as root and aborts the rest of the test · **Status:** open

**File:** cmd/tskflwctl/main_test.go:179 | **Component:** cmd/smoke-tests
**Effort:** XS · **Urgency:** soon

`cmd/tskflwctl/main_test.go:179-188` pins the "unclassified filesystem failure maps to
exit 1" row of the published exit taxonomy by making the tasks directory read-only and
asserting the write fails:

```go
tasksDir := filepath.Join(root, "tasks")
if err := os.Chmod(tasksDir, 0o500); err != nil { t.Fatal(err) }
stdout, stderr, code := runStreams(t, root, "task", "new", "Blocked Write", …)
…
if code != 1 || stdout != "" { t.Fatalf("filesystem failure = exit %d, want 1; …") }
```

Mode `0o500` does not stop uid 0: root holds `CAP_DAC_OVERRIDE`, so the write succeeds,
the command exits 0, and the test fails. On this checkout, unmodified:

```
--- FAIL: TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior (2.44s)
    main_test.go:188: filesystem failure = exit 0, want 1; stdout={"schema_version":"1.80",…"created":{"kind":"task",…}}
```

CI is unaffected — `.github/workflows/ci.yml` runs `ubuntu-latest`, whose `runner` user
is unprivileged — so this is green where it is watched and red where it is not.

Two reasons it is worth more than its severity suggests. First, `t.Fatalf` aborts the
test, so every assertion *after* line 188 in the repo's single subprocess smoke test
never runs in a root container: the typed filesystem error envelope, and whatever
follows it, go unchecked rather than merely unverified. Second, agent and devcontainer
sessions routinely run as root — this routine itself hits the failure twice a week, and
has to re-derive that it is environmental before it can report a clean baseline.

**Context:** severity Low (no shipped behaviour is wrong, and CI is unaffected);
urgency `soon` rather than `eventually` because it costs every root-container
contributor a false failure and silently truncates the smoke test's remaining coverage.

**Why tests didn't catch it:** it *is* the test. Nothing asserts that the environment
the negative probe depends on actually holds.

**Fix sketch:** make the probe's precondition explicit rather than assumed. Either
skip with a reason when the chmod does not bite — after `os.Chmod`, attempt a throwaway
`os.WriteFile` inside `tasksDir` and `t.Skip("requires an unprivileged user")` if it
succeeds — or inject the failure in a way that does not depend on DAC (point the planning
root at a path whose parent is a regular file, so the write fails for every uid). The
skip is the smaller change; the injection keeps the coverage.

**Tightening (adjacent):** if the skip route is taken, assert the rest of the test body
still runs — the value lost today is the assertions after line 188, not the probe itself.

**Recommendation:** Make the probe's precondition explicit: after the chmod, attempt a throwaway write and t.Skip("requires an unprivileged user") if it succeeds — or inject the failure without depending on DAC so the coverage survives.

#### L2. `TestTemplateList_KindFilter` pins a catalog count that epic 22 exists to grow · **Status:** open

**File:** internal/cli/template_test.go:50 | **Component:** cli/templates
**Effort:** XS · **Urgency:** eventually

`internal/cli/template_test.go:39-57` asserts the size of the built-in template catalog:

```go
out := runRoot(t, "template", "list", "--kind", "audit", "--json")
…
if len(got.Templates) != 2 {
    t.Fatalf("--kind audit should list 2 templates, got %d", len(got.Templates))
}
```

The test's actual subject — that every returned row has `kind == "audit"` — is asserted
immediately after, and correctly. The count is the fragile part, and it is fragile in a
specific direction: epic 22 (`22-selectable-template-library`, "a growable store of
named task/epic/audit body templates — built-in + repo-local") exists to add templates.
Adding a third audit template is a success for that epic and a red test here, with a
message that reads like a regression rather than an expected catalog change.

Labelled Low and reported as a smell, not a defect: I can make it fail only by adding a
template, which is a deliberate act, so there is no repo state today that makes it
wrong. It is reported because it sits in the file the random pick landed on and shares a
root with M3 — a per-item invariant pinned by a whole-collection assertion.

Note the sibling `TestTemplateList_JSON` already does this well: `len(got.Templates) < 4`
is a floor, not an equality, so it survives growth.

**Fix sketch:** replace the equality with a floor plus the invariant that already
follows it — assert `len(got.Templates) >= 1` (or `>= 2` to keep the current signal) and
keep the per-row `tpl.Kind != "audit"` loop, which is what the test is actually for. If
the exact catalog is worth pinning, pin it where catalog contents are the subject —
`TestSchemaKind_AdvertisesTemplates` already names `default` and `security` explicitly
and is the right home for that.

**Recommendation:** Replace the len(got.Templates) != 2 equality with a floor and keep the per-row kind assertion that follows it; pin exact catalog contents in TestSchemaKind_AdvertisesTemplates instead.

#### L3. `thread new`'s four input refusals are untested, and all four are redundant with a deeper layer · **Status:** open

**File:** internal/core/service_thread.go:30 | **Component:** core/thread-creation
**Effort:** XS · **Urgency:** eventually

`Service.NewThread` opens with four input refusals and the coverage profile reports
every one unexecuted — 8 of 12 sentinel blocks in `service_thread.go` are uncovered, and
these are four of them (two more are retry-loop *bodies*, which correctly never run in a
green suite, and the remaining two are the guards M2 covers).

All four are reachable from the CLI. Verified against a throwaway planning repo with the
shipped binary:

| Guard | line | Input | Observed |
| --- | --- | --- | --- |
| description required | 30 | `thread new "T" --goal g` | exit 11 `Thread description is required` |
| goal required | 36 | `thread new "T" --description d` | exit 11 `Thread goal is required` |
| goal is single-line | 39 | `--goal $'a\nb'` | exit 11 `Thread goal must be a single line` |
| title slugifies to empty | 48 | `thread new "!!!"` | exit 11 `title produced an empty slug: "!!!"` |

**This is Low, and the reason is the interesting part.** My first read of the coverage
profile had this as a Medium on the grounds that the four guards were the only thing
between CLI input and a Thread document. That was wrong, and the mutation probe said so.
Deleting all four and rebuilding leaves every case still refused, one layer down:

```
thread new "T"                       → exit 11  thread description is required     (domain/thread.go:92)
thread new "T" --description d       → exit 11  thread goal is required            (domain/thread.go:98)
thread new "!!!" --description d --goal g
                                     → exit 11  Thread creation requires a non-empty slug
                                                                        (core/thread_creation.go:119)
```

No invalid Thread is written and `lint` stays clean. `domain.ValidateThreadDocument`
(`internal/domain/thread.go:84`) independently re-checks description-required,
description-length, goal-required and goal-single-line, and
`ValidateThreadCreationPlan` re-checks the slug. So the `service_thread.go` preamble is
a redundant earlier copy, and the missing tests guard almost nothing.

What survives as worth reporting is narrower: the two copies of these messages have
**diverged in case** — `core` says `Thread description is required`, `domain` says
`thread description is required`. Today a caller sees whichever layer fires first, which
is `core`; if the preamble were ever removed as redundant, every one of these
user-visible strings would silently change case. That is the kind of drift a test
asserting message text would pin and no current test does — and it is precisely the
"a word that means the same thing in two places is spelled once" rule from
`docs/ARCHITECTURE.md`, applied to a message rather than a vocabulary.

**Fix sketch:** do not add four negative tests for the preamble — they would pin a
redundant layer. Either (a) assert the four refusals once at the level that actually
enforces them, in `internal/domain`, where `ValidateThreadDocument`'s own table can take
the rows cheaply, or (b) if the preamble stays for its better messages, make that intent
explicit: one test that `NewThread` reports the `core` spelling, so the duplication is
load-bearing and documented rather than incidental.

**Tightening (adjacent):** the member-mutation preamble guards at `:121`
(`%s requires at least one task`) and `:136` (`duplicate member intent`) are also
uncovered. `:121` is unreachable from the CLI — Cobra's `RangeArgs` refuses
`thread add <thread>` with no task argument first — so it is defence-in-depth. `:136`
*is* reachable and verified: `thread add <thread> <task-id> <same-task-slug>` →
exit 11 `task references resolve to duplicate member intent 6gfr9parw1bq`. That one has
no deeper backstop in the path I drove, so of the six uncovered preamble guards in this
file it is the one most worth a row.

**Follow-up:** whether the `core` preamble should exist at all, given `domain` enforces
the same four rules, is a simplification question — it will come round on that lens in
three weeks. Flagging it here only so the duplication is on record.

**Recommendation:** Do not pin the redundant preamble; assert the four refusals once in internal/domain where ValidateThreadDocument enforces them, or add one test fixing the core spelling if the duplicated messages are deliberate.

## Review context

Routine: `code-quality-audit` · lens `test-rigour` · ISO week `2026-W40`,
slot `Fri` (index 3).

The lens question: does the suite prove **rejection** as rigorously as success?
Method was guard-coverage tracing — enumerate the load-bearing refusal predicates,
name the repository state that trips each, then look for a test that CONSTRUCTS that
state and ASSERTS the refusal.

This run added two mechanical steps to that method, because reading alone both
over- and under-reports. First, a **cross-package** coverage profile
(`go test ./... -coverpkg=./...`) to shortlist candidates: a per-package profile
credits only the package under test, so a `core` guard exercised by a `cli` test reads
as uncovered. That trap produced one hypothesis this audit discarded (see *What
audited clean*). Second, every surviving "no test covers this" claim was **confirmed by
mutation probe** — delete the guard in a throwaway copy of `HEAD` under the scratchpad,
re-run `go test ./...`, and see whether anything turns red. A claim that survived both
is reported below; one that did not is not reported at all.

No finding here is a live defect. Every guard named is present and correct today, and
I drove most of them through the shipped binary to confirm they fire with the right
sentinel and message. What is missing is the test that would notice if they stopped.

## Punch list

One line per finding for fast triage. Order: Critical → High → Medium → Low.

M1. `thread compose`/`apply` refuse malformed and stale plans, and 24 of 37 refusals have no test  (effort: S · urgency: soon)
M2. Thread creation's committed-conflict no-retry guard is untested while its mutation twin is covered  (effort: XS · urgency: soon)
M3. `theme` can silently fall back to ambient presentation on a bad `--space`, untested  (effort: XS · urgency: soon)
L1. The exit-taxonomy smoke test's filesystem probe is vacuous as root and aborts the rest of the test  (effort: XS · urgency: soon)
L2. `TestTemplateList_KindFilter` pins a catalog count that epic 22 exists to grow  (effort: XS · urgency: eventually)
L3. `thread new`'s four input refusals are untested, and all four are redundant with a deeper layer  (effort: XS · urgency: eventually)

## Files audited

- **Signal**: `internal/core/service_thread_test.go` — 9 commits in 30d × `internal/core`
  referenced by 148 files (log(10)×log(149) ≈ 11.5). Epic 30 is at 95% and still
  landing here.
- **Signal**: `internal/core/service_research_test.go` — 8 commits in 30d over the same
  core blast radius (≈ 11.0); research is the newest entity (no status at all) and was
  outside the 2026-09-11 test-rigour run's scope entirely.
- **Adjacency** (from `service_thread_test.go`): `internal/store/threadcreation.go` +
  `threadcreation_test.go` / `threadmutation_test.go` — the hexagonal-seam hop, from a
  `core` retry invariant to the `store` guard that triggers it. This hop produced M2.
- **Adjacency** (from `service_research_test.go`): `internal/core/service_research.go` —
  the test-rigour hop in reverse, from a test file to the production predicates it
  claims to cover.
- **Random**: `internal/cli/template_test.go` (+ `internal/cli/template.go`, and
  `internal/cli/theme.go` which the template file's own guard pointed at). Produced
  M3 and L2.

Pulled in by the high-yield rule ("a guard with no negative test"):
`internal/core/thread_apply.go` and its three test files, surfaced by the coverage
filter below.

A note on the raw signal: goldens dominate the 30-day churn ranking
(`template_show_security_json.golden` at 18 commits, twenty more at 17) because every
golden is rewritten on each `schema_version` bump. That is co-churn, not per-file
interest, so the ranking was recomputed over `*_test.go` files only. The highest-churn
golden did still earn its keep — it is what made the random pick's file interesting.

## Commands run

Build (`just` is not installed in this environment; the spec's documented fallback was
used):

```
go build -o bin/tskflwctl ./cmd/tskflwctl          # exit 0
```

Signal and blast radius:

```
git log --since="30 days ago" --name-only --pretty=format: \
  | grep -E '_test\.go$' | sort | uniq -c | sort -rn | head -25
grep -rln "internal/core\""  --include=*.go internal cmd | wc -l   # 148
grep -rln "internal/wire\""  --include=*.go internal cmd | wc -l   # 54
grep -rln "internal/store\"" --include=*.go internal cmd | wc -l   # 27
```

Mechanical guard filter (cross-package, so sibling-package coverage counts):

```
go test ./... -coverpkg=./... -coverprofile=cover_all.out -covermode=count
```

Zero-count blocks whose source wraps a domain sentinel, by package:

```
internal/core 123 · internal/store 72 · internal/cli 22
internal/domain 13 · internal/config 8 · internal/tui 1      # 239 total
```

Per audited file: `thread_apply.go` 24/37 · `service_thread.go` 8/12 ·
`service_research.go` 1/8.

Mutation probes, each in a throwaway copy of `HEAD` under the scratchpad
(`git archive HEAD | tar -x -C …` — never this tree), then `go test ./...`:

| Guard deleted | Suite result |
| --- | --- |
| `service_thread.go:93` `!result.Committed` (creation retry) | **green** → M2 |
| `service_thread.go:179` `!result.Committed` (mutation retry) | red (`store`: `TestThreadMutationAttributesReleaseFailureAfterCommit`, `retries=3`) → covered, not a finding |
| `thread_apply.go:381`–`:387` stale-plan prerequisite/dependent | **green** → M1 |
| `service_thread.go:30`,`:36`,`:39`,`:48` `thread new` preamble | **green**, but every case still refused one layer down → L3 (demoted from Medium) |
| `theme.go:41` `wantsSpace()` fall-back refusal | **green** → M3 |
| `service_research.go:182` unset-path unknown field | red (`cli`: `TestResearchSet_InputGuards`) → covered, hypothesis discarded |

Reachability against a throwaway planning repo (scratchpad), shipped binary:

```
thread compose --from <manifest>    # 10 malformed manifests → exit 11, precise messages
thread apply <stale plan>           # exit 11: planned dependent … does not exist
thread apply <tampered plan>        # exit 11: planned Thread must be unstarted, got "in-progress"
thread new                          # 4 bad inputs → exit 11 (incl. title "!!!" → empty slug)
--space typo theme list             # shipped: exit 10 · guard removed: exit 0 + a theme list
```

A regression test written for M2 was validated both ways: it fails
(`retries=3 committed=false`) with the guard removed and passes with it restored.

**Pre-existing red:** `cmd/tskflwctl`'s
`TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` fails on this checkout before
any change, because this environment runs as uid 0. Every other package is green.

## What audited clean

- `internal/core/service_research.go` — the `SetResearchFields` guard family is the
  best-covered surface I read: empty update, protected fields on **both** the set and
  unset paths (a 4-field table), unknown field needs `--force`, description length, tag
  coercion, and `updated_at` stamping all have negative tests. The one gap I
  hypothesised — the unset branch's unknown-field check at `:182`, whose production
  comment flags it as deliberate — is covered a layer out by
  `cli.TestResearchSet_InputGuards`, which the mutation probe confirmed. 1 of 8
  sentinel blocks uncovered, and that one (`:65`, empty slug from title) is shadowed.
- `internal/core/service_thread.go`'s **mutation** retry path (`runThreadMutation`,
  `:179`) — the committed-conflict guard that M2 reports missing on the creation path is
  properly covered here, by `store.TestThreadMutationAttributesReleaseFailureAfterCommit`,
  which injects a release failure that *wraps* `ErrConflict`, drives it through
  `core.MustNewService` with `WithRetry`, and asserts `retries == 0`. That test is the
  model M2 asks to be copied, not criticised.
- `internal/store/threadcreation.go:138` (`thread %q already exists` → `ErrConflict`) —
  uncovered, and deliberately not reported. It is shadowed by
  `core/thread_creation.go:131`, which is covered and fires first on every path I could
  construct; the store-level check is defence-in-depth behind a tested guard, not a gap.
- `internal/core/thread_apply.go`'s tampered-plan family — the cases that need judgment
  *are* covered: wrong planning identity, dependency cycle, a `../../` slug escape, a
  backdated `created` vs `composed_at`, cross-kind id collision, and a memberless plan
  (`TestPrepareThreadApplyRejectsWrongIdentityAndCycle`,
  `...RejectsUnsafeOrEditedCreationIdentity`). M1 is about the structural layer beneath
  them, not these.
- `internal/wire/envelopes_test.go` was re-checked rather than re-audited: the
  2026-09-11 run proved envelope exhaustiveness mechanically (53 declared = 53 tested)
  and that still holds, so it was dropped from the file set despite ranking second on
  churn.

## Owner triage (2026-10-04)

M1-M3 are tracked together by
[guarded planning mutation boundary regressions](../tasks/6ggdkzv2tnta-pin-guarded-planning-mutation-boundary-contracts.md).
The task is sequenced after workspace identity parity, explicit authorization policy, cross-kind
collision lint, and core-owned impact/recovery semantics for a final bounded closeout pass.
Behavior-level tests still belong in each implementation slice; this dependency does not defer
their verification until the end.

Workspace root/identity replacement tests stay in their existing task rather than being copied.
Shared-validator and resolver redesign, a generic conformance framework, and broader retry-policy
decisions are not approved by this regression task. Re-measure the current reachable refusals and
make each guard-removal probe fail for the intended condition, not unrelated invalid fixtures.

L1-L3 remain open outside this approved scope. Tracked findings are not fixed, and this audit
cannot close as fully triaged while those findings remain unresolved.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ✔ M1 · fixed — Malformed/stale Thread apply boundary contracts | implemented locally in 6ggdkzv2tnta | pending independent review and merge
- ✔ M2 · fixed — Committed Thread creation conflict recovery | implemented locally in 6ggdkzv2tnta | pending independent review and merge
- ✔ M3 · fixed — Explicit-space theme refusal | implemented locally in 6ggdkzv2tnta | pending independent review and merge
- ○ L1 · open — `tskflwctl task new "Make the exit-taxonomy filesystem probe honest under root" --epic 21-code-quality-architecture-hardening --tags tests,ci --tier 3 --priority low --description "chmod 0o500 does not bite uid 0, so the smoke test fails in root containers and aborts before its remaining assertions."`
- ○ L2 · open — `tskflwctl task new "Stop pinning built-in template catalog size in the kind-filter test" --epic 22-selectable-template-library --tags tests,cli --tier 4 --priority low --description "A third audit template would redden TestTemplateList_KindFilter; assert the per-row invariant and a floor instead of an exact count."`
- ○ L3 · open — `tskflwctl task new "Resolve the duplicated thread-validation messages between core and domain" --epic 21-code-quality-architecture-hardening --tags threads,tests,simplification --tier 4 --priority low --description "NewThread re-checks four rules domain.ValidateThreadDocument already enforces, and the two copies have diverged in case."`
