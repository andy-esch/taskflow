---
schema: 1
id: 6gh1frgyz2fh
bucket: open
area: adapter-hygiene
date: "2026-10-06"
updated_at: "2026-10-06"
---

# Audit: adapter-hygiene — 2026-10-06

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

#### H1. init silently ignores -C/--space and scaffolds the caller's cwd · **Status:** tracked by 6ggjmtmdd54w

**File:** internal/cli/init.go:50 | **Component:** cli
**Effort:** S · **Urgency:** acute

`-C/--chdir` is documented as "anchor to the planning repo at this path", and
`App.startDir()` (`internal/cli/root.go:383`) is described in its own doc comment as
"the single source of the discovery start directory ... so the 'where do we start
discovery' contract can't drift between consumers". `repoPreRun` states the invariant
in so many words for the completion path: "Defer repo composition to CompletionService
so **-C/--space cannot silently target cwd**" (`root.go:363`).

`init` is the one mutating command that breaks it. It overrides the root hook with
`PersistentPreRunE: app.styleOnlyPreRun` (`internal/cli/init.go:50`), which skips
`resolve()` and therefore never reaches `startDir()`. It then takes its target from its
own `--path` flag, defaulting to `"."`, resolved with `filepath.Abs(path)` against the
process cwd (`init.go:57`). `app.Chdir` and `app.Space` are parsed, accepted, and
dropped on the floor.

Three consequences, all reproduced against this build:

First, an explicit `-C` is ignored and the scaffold lands in the caller's cwd. Second,
the `-C`/`--space` mutual-exclusion guard — which lives inside `startDir()` — is
bypassed, so a combination every other command rejects with `ErrValidation` writes
files and exits 0. Third, an unknown `--space` is swallowed; `registeredSpaceStart`'s
stated rule that "explicit selection is also the wrong-repo guard, so a missing,
unreadable, or identity-mismatched target must fail loudly and can never fall back to
cwd discovery" does not apply, because that code is never reached.

The general mechanism is a known Cobra property rather than a local slip: Cobra does
not chain `PersistentPreRunE`. A child's hook *replaces* the parent's rather than
running after it, so every command that overrides the root hook silently opts out of
whatever that hook enforced — see spf13/cobra#2039 and the user guide. `init`,
`doctor`, `version` and the completion tree all override it here; `init` is the only
one that both mutates the filesystem and has a target to get wrong. (`config`, which
`styleOnlyPreRun`'s comment groups with `init`, keeps its own hook but still honors
`-C` — verified.)

Note the dry-run preview does not reveal the mistake either: `init --dry-run` prints
relative entries (`+ audits`, `+ .tskflwctl.toml`) and names the planning tree as `.`,
so the operator sees a plausible plan for the wrong directory.

**Failing scenario:** verified against `bin/tskflwctl` built from this checkout, in
two throwaway directories outside the repo.

```
$ cd /tmp/probe/cwdhere            # empty, not a planning repo
$ tskflwctl -C /tmp/probe/target init --no-register
  + tasks … + .tskflwctl.toml
    planning tree  .
$ ls /tmp/probe/target             # the -C target
                                   # (empty)
$ ls -a /tmp/probe/cwdhere         # the cwd
.tskflwctl.toml audits epics research tasks threads
```

The full tree plus `.tskflwctl.toml` was created in the cwd; the directory named by
`-C` was never touched, and the command exited 0 with no warning. Run from a checkout
that *is* already initialized, the same invocation instead reports that checkout's
topology and offers it a scaffold repair — a write prompt aimed at a repository the
operator explicitly pointed away from.

The two selector variants task 6ggjmtmdd54w records as "not yet reproduced" both
reproduce:

```
$ tskflwctl -C /tmp/probe/t2 --space nope init --no-register   # → scaffolds cwd, exit 0
$ tskflwctl -C /tmp/probe/t2 --space nope status               # → error: validation failed:
                                                               #   --space and -C are two answers
                                                               #   to one question; pass one
$ tskflwctl --space does-not-exist init --no-register          # → scaffolds cwd, exit 0
$ tskflwctl --space does-not-exist task list                   # → error: not found: unknown
                                                               #   space "does-not-exist"
```

**Why tests didn't catch it:** `internal/cli` has no test that passes a target flag to
`init` at all — `grep -n 'Chdir\|"--space"\|"-C"' internal/cli/*_test.go` returns only
`task set` cases. The existing init tests drive `--path` directly, which works
correctly, so they exercise the one selector that is honored and never the two that are
not. There is also no two-directory fixture anywhere in the suite that would notice a
write landing in the caller's cwd rather than the named target.

**Context:** urgency is acute rather than soon even though the task that owns it is
`tier: 2 / priority: medium`. This writes files into a directory the operator named
*away from*, exits 0, and the dry-run preview does not disclose the target — the three
properties that make a wrong-target mutation hard to notice before it has happened.
The repo's own rule for this class ("an adapter cannot claim that a missing checkout is
healthy") is the same shape.

**Tightening (adjacent):** have `init`'s success and dry-run receipts print the
absolute directory they acted on, not `planning tree .`. The relative form is what
makes the wrong target invisible in the transcript.

**Follow-up:** the general defect is structural, not local — any command overriding
`PersistentPreRunE` silently opts out of the root hook's guards, and nothing fails when
it does. A fitness test that asserts every command either uses `repoPreRun` or appears
on an explicit allowlist with a recorded reason would catch the next one; that is its
own task, and it belongs with the controller-boundary work in 6gcwcf8rxe72 rather than
with this fix.

**Recommendation:** Route init's bootstrap target through startDir() — honor -C (and resolve --path relative to it, as git does), or reject the selectors it cannot honor before any write.

**Resolution:** Already owned by 6ggjmtmdd54w
(prevent-silently-ignored-target-selectors-during-init), which records the -C
case from 2026-10-04 dogfood. This run adds deterministic reproductions of the
two variants that task lists as not yet reproduced: -C plus --space bypasses the
mutual-exclusion guard, and an unknown --space is swallowed, both scaffolding
cwd at exit 0. git -C's applied-first semantics are offered as the precedent for
its open policy question.

#### M1. ShowAudit discards the findings and near-misses its own read parsed, so both primary adapters re-parse the body · **Status:** open

**File:** internal/core/entity_read.go:105 | **Component:** core + cli + tui
**Effort:** S · **Urgency:** soon

The secondary adapter parses an audit's findings on *every* audit read.
`parseAuditWithFindings` (`internal/store/auditstore.go:262`) runs
`domain.ParseFindings`, `domain.TallyFindings`, `domain.NearMissFindingHeaders` and
`domain.LintCandidateTasks` over one body and returns all four, and the file's own
comment at `:43` notes the snapshot path exists so callers need not re-read a body.

The single-audit read throws three of the four away. `core.AuditWithBody`
(`internal/core/entity_read.go:105`) is `{Audit, Body}` — no findings, no near-misses,
no candidate issues — and `Service.ShowAudit` returns exactly that. So both primary
adapters re-parse the body they were just handed:

```go
// internal/cli/audit.go:490
findings := domain.ParseFindings(record.Value.Body)

// internal/tui/detail.go:1720
if findings := domain.ParseFindings(body); len(findings) > 0 {
```

Core already models the richer shape — `core.AuditWithFindings`
(`internal/core/store.go:278`) carries `Findings`, `NearMisses` and `CandidateIssues`,
and `Service.Lint`/`Summary` consume all three. The `show` path simply does not use it,
so each adapter re-derives the narrow half independently.

Two consequences follow, and the second is the one a user feels.

First, the duplication is the ordinary drift risk: one grammar, parsed in three places,
with the two adapter copies reachable only through a `domain` call that the store's
copy does not share a test with.

Second, and because the adapters re-derive only the *narrow* half, the near-miss
evidence the same read already computed is unreachable from `audit show` on either
surface. `NearMissFindingHeaders` exists to catch a heading that "reads as a finding
but would silently drop", and `lint` describes the consequence in its own message:
"it parses to nothing and its work is invisible". `audit show` is the cheapest triage
path and the one an agent reaches for first, and it says nothing at all.

The machine surface is weaker still. `AuditShowHuman` takes
`findings []domain.Finding` (`internal/cli/render/render.go:825`) and renders a
status-glyphed finding index; `AuditShowEnvelope` (`internal/wire/envelopes.go:791`)
is `{schema_version, audit, body}`, so `audit show --json` publishes no structured
findings whatsoever. An agent gets raw markdown and the frontmatter-derived counters,
and must either run a second command (`audit findings --json`, which does publish
them) or re-implement the finding grammar itself. Neither path reports a near-miss.

**Failing scenario:** verified in a throwaway planning repo outside this checkout.
Create an audit, add one canonical finding with the tool, then append a heading that
the near-miss recognizer claims and `ParseFindings` drops (`#### M-2.` — a separator
between letters and digits is enough):

```
$ tskflwctl audit show 2026-10-06-probe-area --frontmatter-only
findings: ░░░░░░░░░░ 0% settled  0/1  (1 open)
└── open
    └── H1  A real canonical finding            # M-2 absent, and unmentioned

$ tskflwctl audit show 2026-10-06-probe-area --json | jq 'keys'
["audit","body","schema_version"]               # no findings at all

$ tskflwctl audit lint 2026-10-06-probe-area
  findings: line 36: "#### M-2. A heading that reads as a finding but is not
  canonical" is not a finding header, so it parses to nothing and its work is
  invisible — write "#### M2. …" (`lint --fix` does this)
exit 11
```

Same body, same read, same already-computed `NearMisses`: `audit lint` calls the work
invisible, and `audit show` — human and `--json` — presents the audit as if it were
whole.

**Why tests didn't catch it:** the near-miss contract is tested where it is
*enforced*, not where it is *reported*. The write guards refuse to introduce one
(`domain.NearMissWriteError`) and `audit lint`/`lint --fix` cover drift already in a
file, so the recognizer has real coverage. Nothing asserts what a READ surface says
about a body that already contains one, because `AuditWithBody` gives the show-path
tests nothing to assert on. The duplicate parse is invisible to tests for the same
reason: both adapter copies agree with the store's copy today, so no assertion can
tell one spelling from three.


**Tightening (adjacent):** `audit show --json` should publish the structured findings
once core carries them — that closes the human/machine asymmetry in the same edit and
is additive on the wire (a minor `SchemaVersion` bump under the rule at the top of
`wire.go`).

**Follow-up:** `audit list`'s counters come from the same narrow tally, so the
"confidently false on the cheapest triage path" symptom task 6g77rn6em6n8 describes
also applies there. Whether `list` grows a near-miss column or just a repository-level
nudge is a UX call that belongs with that task, not with this fix.

**Recommendation:** Carry parseAuditWithFindings' already-computed findings and near-misses through ShowAudit, reduce both adapters to consuming them, and let audit show report a near-miss count.

#### L1. The epics tab's lifecycle and view vocabularies are hand-typed where the task axis is pinned to the domain · **Status:** open

**File:** internal/tui/action.go:63 | **Component:** tui
**Effort:** XS · **Urgency:** eventually

The task status axis in the TUI is pinned to the domain and the other two axes are
not, which is a coverage asymmetry rather than a live defect.

`statusViews` (`internal/tui/statusview.go:16`) says it is "the single source of truth
for task status views" and `TestStatusViewsCoverAllStatuses`
(`internal/tui/model_test.go:1202`) walks `domain.AllStatuses()` and asserts every one
is reachable as a `:` view. That is the right guard and it works.

The epic and audit axes have the same shape and no such guard, even though the domain
publishes both vocabularies as closed sets with accessors — `domain.AllEpicStatuses()`
(`internal/domain/epic.go:69`) and `domain.AllAuditBuckets()`
(`internal/domain/audit.go:20`):

```go
// internal/tui/statusview.go:56 — epicViews
{"live", ""}, {"retired", "retired"}, {"deprecated", "deprecated"}, {"all", "all"}

// internal/tui/action.go:63 — epicTransitions
{verb: "activate", to: "active"},
{verb: "retire", to: "retired"},
{verb: "deprecate", to: "deprecated"},
```

`epicTransitions` is the clearer case because the file states a reason that no longer
holds. Its comment says epics stay inline because they "have no CLI verb vocabulary to
share", but `epic move <epic>... <status>` exists (`internal/cli/epic.go:182`) and
validates through `domain.ValidateEpicStatus`. The destination strings are also raw
literals where named constants exist (`domain.EpicStatusActive` and siblings), so the
sibling tables sourced from `domain.TaskTransitions()`/`domain.AuditTransitions()` get
compile-time coupling to the vocabulary and this one gets none.

There is in-repo precedent for closing exactly this gap on the epic vocabulary:
`internal/wire/schema_descriptions_test.go:49` pins the `jsonschema` description text
to `domain.AllEpicStatuses()` with the comment "leave it stale. Pin it to
domain.AllEpicStatuses() so the drift fails loudly here."

No failing scenario — this is a smell, not a defect. All three literals match the
domain today, so nothing misbehaves; what is missing is the mechanism that would make
the next vocabulary change fail loudly. The failure mode when it comes is silent in
both directions: a renamed status leaves a menu row that `svc.MoveEpic` rejects at
apply time as `ErrValidation`, and an added status is simply absent from the menu and
the `s`/`S` cycle with nothing to notice it.


**Tightening (adjacent):** `dropArchivedRecords` (`internal/tui/commands.go:322`)
defines the archived set by negation (`!= StatusCompleted && != StatusDeprecated`),
so a newly added terminal status would silently enter the TUI's default working view
while `task list`, which asks `Status.IsActive()`, would correctly hide it. Its
neighbour `rankOf` already handles the unknown-value case deliberately; this one does
not. A `domain.IsTaskArchived` predicate beside the existing `IsEpicArchived` would
let the adapter stop enumerating. Same edit, same file.

**Follow-up:** `sortEpicsForView` (`internal/tui/commands.go:314`) re-spells core's
live-first comparator character-for-character from the unexported `dashboardEpics`
(`internal/core/service_epic.go:229`), and `filterEpicsByView` re-spells its live-set
predicate. Exporting the partition from core so both dashboards and the epics tab
share one spelling is the same question the previous adapter-hygiene audit's L1 asked
from the CLI side; it is worth its own task rather than riding this one.

**Recommendation:** Add domain-coverage tests for epicViews/auditViews mirroring TestStatusViewsCoverAllStatuses, and use the domain.EpicStatus* constants in epicTransitions.

## Run context

Routine: `code-quality-audit` · lens `adapter-hygiene` · ISO week `2026-W41`,
slot `Tue` (index 4 — `(41 * 2 + 0) mod 6`).

## Punch list

- `H1. init silently ignores -C/--space and scaffolds the caller's cwd  (effort: S · urgency: acute)`
- `M1. ShowAudit discards the findings and near-misses its own read parsed, so both primary adapters re-parse the body  (effort: S · urgency: soon)`
- `L1. The epics tab's lifecycle and view vocabularies are hand-typed where the task axis is pinned to the domain  (effort: XS · urgency: eventually)`

## Files audited

Signal ranked by `log(churn + 1) × log(blast_radius + 1)` within the lens surface
(`internal/cli/`, `internal/tui/`, `internal/wire/`, `internal/core/service*.go`),
churn over the last 30 days, blast radius as the count of other files referencing the
file's exported identifiers. Generated goldens, `testdata/`, and `_test.go` were
excluded per the lens quality filter.

- **Signal**: `internal/core/service.go` — churn 14 × blast 156, score ≈ 13.7, the
  highest in the surface; it is the application seam every adapter reads through.
- **Signal**: `internal/core/service_task.go` — churn 12 × blast 117, score ≈ 12.2.
- **Adjacency** (from `internal/core/service.go`): `internal/wire/envelopes.go` — the
  hop across the hexagonal seam to the neutral contract that publishes core's results.
- **Adjacency** (from `internal/core/service_task.go`): `internal/tui/item.go` — chosen
  because the lens's high-yield probe (`grep -rn "core\.Role\|core\.Gate\|core\.Status"`)
  put a `core.Role` switch here, making it a candidate leak.
- **Random**: `internal/tui/commands.go` (seeded shuffle over the 91 eligible files).

`internal/wire/wire.go` scored ≈ 11.3 and was deliberately passed over: it was the top
signal pick of the previous adapter-hygiene run (2026-09-15) and two of its findings
are still open.

Read for context while evaluating the above: `internal/cli/init.go`,
`internal/cli/root.go`, `internal/cli/epic.go`, `internal/cli/audit.go`,
`internal/cli/render/thread.go`, `internal/cli/render/status.go`,
`internal/cli/render/render.go`, `internal/tui/action.go`,
`internal/tui/statusview.go`, `internal/tui/detail.go`,
`internal/core/thread_projection.go`, `internal/core/dependency_graph.go`,
`internal/core/service_epic.go`, `internal/core/store.go`,
`internal/store/auditstore.go`, `internal/domain/status.go`,
`internal/domain/epic.go`, `internal/domain/finding.go`, `.golangci.yml`,
and the 2026-09-15 adapter-hygiene audit (to avoid re-finding its open items).

## Commands run

Build and validation (`just` is not installed in the routine's container; the spec's
fallback was used):

- `go build -o bin/tskflwctl ./cmd/tskflwctl` → exit 0.
- `go test -race ./...` → **1 pre-existing failure**, see the note below. 26 packages ok.
- `./bin/tskflwctl lint` → see the closing validation note.
- `./bin/tskflwctl audit lint 2026-10-06-adapter-hygiene` → see the closing note.

Discovery and lens probes:

- `git log --since="30 days ago" --name-only --pretty=format: -- internal/cli/ internal/tui/ internal/wire/ internal/core/service*.go`
  → churn ranking (filtered as described above).
- `grep -rn "core\.Role\|core\.Gate\|core\.Status" internal/cli internal/tui internal/wire`
  → 18 production hits. All are counting or formatting supplied values, which
  `threadItem`'s own contract comment explicitly sanctions; none traverse dependencies
  or recreate eligibility. No finding.
- `grep -rn '"os"\|"path/filepath"\|spf13/cobra\|charmbracelet' internal/core`
  → two hits, both pure string work (`filepath.ToSlash` at `dependency_graph.go:619`,
  `filepath.Base/Clean` at `space_registry.go:225`). Not filepath-for-IO. No finding.
- Envelope contract sweep: extracted all 54 `*Envelope` struct definitions in
  `internal/wire` and checked each for a `SchemaVersion` field → all 54 carry it.
- `grep -rnoE 'domain\.(Lint|Validate|Is[A-Z]|Parse[A-Z])[A-Za-z]*' internal/cli internal/tui internal/wire`
  → the probe that surfaced M1 (`domain.ParseFindings` called from both primary adapters).

Mutation probes for H1 and M1 were run against throwaway planning repos created under
the routine's scratchpad, never against this checkout. H1: two sibling temp
directories, `init` invoked with `-C` naming one while cwd was the other, then both
inspected — plus the `--space`/`-C` conflict and unknown-`--space` variants, and a
`config show -C` control that confirmed `config` honors the flag `init` drops. M1: a
scaffolded temp repo, one canonical finding created with `audit finding new`, then a
near-miss heading appended by hand, then `audit show` (human, `--frontmatter-only`,
and `--json`) compared against `audit lint` and `lint` on the same body.

**Pre-existing red:** `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior`
(`cmd/tskflwctl/main_test.go:188`) fails in this container with
`filesystem failure = exit 0, want 1`. The routine runs as uid 0, so the test's
permission-denied probe does not deny root and the asserted failure never occurs. This
is already recorded as the open `L1` of audit `2026-10-02-test-rigour` ("the
exit-taxonomy smoke test's filesystem probe is vacuous as root"). It is environmental,
predates this run, and is out of scope for this routine. Every other package passed
under `-race`.

## What audited clean

Three of the five audited files produced nothing, and that is the result rather than a
gap in the reading.

- `internal/core/service.go` — **clean** on this lens, and it is the file most likely
  to leak. Its imports are `domain` + `id` + stdlib only. More to the point, the
  decisions adapters would otherwise re-derive are deliberately hoisted into it and
  consumed, not copied: `Summary.SplitCounts()` is called by both dashboards
  (`cli/render/status.go:16`, `tui/dashboard.go:79`) rather than re-partitioned, and
  `Summary.ReadyToClose` is read as a field at all four of its consumers rather than
  re-walked off `OpenAudits`. The code comments name the audits (M2, M9, H2) that
  closed those gaps and the aggregates have held since.
- `internal/wire/envelopes.go` — **clean**. The contract stays adapter-neutral: no
  presentation or filesystem import, and every one of the 54 envelope types carries
  `schema_version` (checked mechanically, not by eye). Its `To*Envelope` constructors
  copy core values rather than recomputing them.
- `internal/tui/item.go` — **clean**, and it is the interesting clean result because
  the lens's own high-yield probe pointed here. `activityForThread` does count member
  roles in the adapter, which looks like a leak until you check what is counted: the
  dispatchable figure is `len(view.Frontier)` straight from core, and the `pending`
  tally only partitions roles core already published. The arithmetic is also sound, not
  merely conventional — `Eligible` requires `isPendingWorkRole` (`dependency_graph.go:826`,
  `:127`), so `Frontier` is a subset of the queued/candidate members and
  `pending - dispatchable` cannot go negative; the `max(0, …)` is belt-and-braces, not
  a clamp hiding a mismatch. Epic liveness is likewise consumed through
  `es.Liveness()`/`es.Live()` rather than re-derived from the rollup.
- `internal/core/service_task.go` — **clean** on this lens. Imports are stdlib +
  `domain`. `ListTasks` keeps filter validation in core and fails loudly on an unknown
  status or epic rather than returning an empty list, which is the behaviour an agent
  routing on exit codes needs.
- `internal/tui/commands.go` — the random pick, and the only audited file with a
  finding attached to it (L1's tightening note). The `tea.Cmd` discipline itself is
  intact: every loader closes over a snapshotted generation counter and returns a
  message, no loader touches the store, and no I/O appears in `Update`/`View`.
  `rankOf` deliberately sorts an unrecognized status last rather than letting a bare
  map index float it to rank 0 — the author was alert to exactly the
  new-enum-value problem that `dropArchivedRecords` eleven lines below still has.

One more negative result worth recording, since it would otherwise look like an
unexamined gap: `internal/graphfmt` has no `depguard` rule of its own, so its doc
comment's claim to have "no CLI, TUI, HTTP, filesystem, styling, or graph-library
dependencies" is convention rather than enforcement. It is **not** filed as a finding
because it is not special: `theme`, `design`, `progressbar` and `themepreview` are in
the same position, so "presentation utilities are unguarded" is a question about the
shape of the fitness-rule set, which belongs to `weekly-architecture-audit`, not here.

## External research

Two Medium-or-higher findings surfaced (H1, M1), so the conditional research step
applied. Three queries, scoped to the actual stack.

**Cobra does not chain `PersistentPreRunE`** — this is the named general mechanism
behind H1, not a local slip. A child command's hook *replaces* the parent's rather
than running after it, so any command that defines its own hook silently opts out of
whatever the root hook enforced, with no compile or test signal. The documented
workaround is to call the parent's function explicitly from the child's. Relevant
beyond the fix: `init`, `doctor`, `version` and the completion tree all override
`repoPreRun` in this repo, which is the population a fitness test should cover.
- https://github.com/spf13/cobra/issues/2039
- https://umarcor.github.io/cobra/

**`git -C` is the design precedent for H1's open policy question.** Task 6ggjmtmdd54w
leaves "support `-C` or reject it in favour of `--path`" undecided; git already
answers the analogous question. `-C` runs "as if git was started in" that path, is
applied *first*, and later relative path options resolve against it — `git
--git-dir=a.git --work-tree=b -C c status` is defined as equivalent to `git
--git-dir=c/a.git --work-tree=c/b status`. The direct translation is that `init
--path` should resolve relative to `-C` rather than the process cwd, which honors both
flags without a conflict rule and matches what a user of `git -C` already expects.
- https://www.kernel.org/pub/software/scm/git/docs/git.html
- https://man.archlinux.org/man/git.1.en

**Import-boundary enforcement for Go, as a proactive note rather than a bug.** This
repo's `.golangci.yml` depguard set is ahead of the common practice — default-deny
with exact package allowances, per-file exceptions, and tests excluded deliberately —
and today's code violates none of it. Two tools in this space cover ground depguard
does not and are worth knowing about if the rule set is ever revisited: `go-arch-lint`
adds tiers and deep-scan of method calls and DI rather than imports alone, and
`go-arch-guard` expresses layer-direction and blast-radius rules as ordinary `go test`
assertions, which would suit this repo's existing preference for fitness tests over
more lint configuration. Neither is a recommendation to adopt; both are the shape of
the answer to the unguarded-presentation-packages question noted above.
- https://github.com/vsfedorenko/go-arch-lint
- https://github.com/NamhaeSusan/go-arch-guard

## Related-task observations (propose-only)

- Possibly under-prioritized: `planning/tasks/6ggjmtmdd54w-prevent-silently-ignored-target-selectors-during-init.md`
  is `tier: 2 / priority: medium`, and its Evidence section records the `-C` case as
  observed once during dogfood while noting the `--space`/`TSKFLW_SPACE` variants
  "have not yet been reproduced". Both variants now reproduce deterministically (see
  H1), and the defect is a mutating command writing to a directory the operator named
  away from, at exit 0, with a dry-run preview that does not disclose the target. That
  reads like a higher priority than it currently carries. Not changed — flagging only.
- Empty scaffold on an otherwise well-specified task:
  `planning/tasks/6g77rn6em6n8-report-unparsed-findings-on-the-audit-read-surfaces.md`
  has a precise `description` and `priority: high`, but its Objective, acceptance
  criteria, and out-of-scope sections are still the unfilled `<...>` placeholders. M1
  supplies the mechanism and a verified reproduction; whoever picks it up has material
  to fill them in with.
- Audit ready to close, not touched by this run: `audit list --json` reports
  `2026-10-05-arch-data-model-and-storage` with `ready_to_close: true` (1 finding, 0
  open). This was not a backpressure run — 7 open audits against a threshold of 10 —
  so the spec kept authoring in scope and triage out of it. Noting it so the next
  backpressure run is not the first to see it.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- → H1 · tracked — Tracked in planning/tasks/6ggjmtmdd54w-prevent-silently-ignored-target-selectors-during-init.md
- ○ M1 · open — `tskflwctl task new "Carry the audit finding parse through ShowAudit instead of re-parsing it in each adapter" --epic 21-code-quality-architecture-hardening --tags audit,cli,tui,contracts --tier 2 --priority medium --description "ShowAudit returns AuditWithBody, so cli/audit.go and tui/detail.go each re-run domain.ParseFindings and neither can see the near-misses the same store read already computed."`
- ○ L1 · open — `tskflwctl task new "Pin the TUI epic and audit view vocabularies to the domain accessors" --epic 21-code-quality-architecture-hardening --tags tui,tests,contracts --tier 3 --priority low --description "epicViews/auditViews/epicTransitions are hand-typed literals with no domain-coverage guard, while the task axis is pinned to domain.AllStatuses by TestStatusViewsCoverAllStatuses."`
