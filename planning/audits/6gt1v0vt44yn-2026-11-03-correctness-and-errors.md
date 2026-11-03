---
schema: 1
id: 6gt1v0vt44yn
bucket: open
area: correctness-and-errors
date: "2026-11-03"
updated_at: "2026-11-03"
---
# Code Quality Audit: correctness-and-errors — 2026-11-03

> Create findings through `tskflwctl audit finding new`; update them through
> `tskflwctl audit finding` so identity, status, resolution, and managed candidate
> metadata stay synchronized. Never hand-edit those fields.

Routine: `code-quality-audit` · lens `correctness-and-errors` · ISO week `2026-W45`,
slot `Tue` (index 0).

## Punch list

One line per finding for fast triage. Order: Critical → High → Medium → Low.

- `M1.` Epic and research field updates validate in random map order, so one request reports different errors  (effort: XS · urgency: soon)
- `M2.` A nil aggregate store panics instead of refusing on most task entry points  (effort: S · urgency: soon)
- `L1.` `EpicRefKey` compares the NN key lexically, so `--epic 0` misses epic `00-…`  (effort: XS · urgency: eventually)
- `L2.` `taskGraphWithStatus` drops the `exists` bool and would fabricate an empty-ID node  (effort: XS · urgency: eventually)
- `L3.` Capability-unavailable errors wrap no sentinel in six places and `ErrValidation` in one  (effort: S · urgency: eventually)

## Files audited

- **Signal**: `internal/core/store.go` — highest churn in the lens surface (13 commits/30d) against the broadest blast radius; it is the consumer-side port file every adapter implements.
- **Signal**: `internal/core/service_task.go` — 8 commits/30d, 1047 lines, and the single largest concentration of exported task use cases (score 11.5).
- **Adjacency** (from `internal/core/store.go`): `internal/store/auditstore.go` — the hexagonal-seam hop from the `AuditStore.MoveAudit` port contract to the adapter that enforces it, plus `internal/domain/audit.go` for the `ValidateMove` policy those counts feed.
- **Adjacency** (from `internal/core/service_task.go`): `internal/core/task_lifecycle.go` — the guarded lifecycle authorization boundary every status write in the service routes through.
- **Random**: `internal/domain/epic.go` — survived the quality filter (158 lines, not generated, not a fixture).

Supporting reads pulled in to settle specific claims rather than audited in full:
`internal/domain/finding.go` (status vocabulary and tally), `internal/core/finding.go`
(rollup), `internal/domain/fields.go` (task field registry), `internal/domain/validate.go`
(`ValidateField`), `internal/core/service_epic.go` and `internal/core/service_research.go`
(the sibling field-update paths), `internal/core/retry.go`, `internal/core/service.go`
(`NewService`), `internal/domain/body.go` (criteria indexing), `internal/domain/lint.go`.

## Commands run

Build (`just` is not installed in this environment, so the documented fallback was used):

```
$ go build -o bin/tskflwctl ./cmd/tskflwctl     # ok, go1.26.0
$ ./bin/tskflwctl --version                     # v0.0.0-20261101181409-474c8e2d5177
```

Discovery signal:

```
$ git log --since="30 days ago" --name-only --pretty=format: \
    -- internal/core internal/domain internal/store | ... | sort -rn
13 internal/core/store.go   8 internal/core/service_task.go   5 internal/store/auditstore.go ...
```

M1 reproduction — identical request, repeated, `--dry-run` so nothing is written:

```
$ for i in $(seq 60); do ./bin/tskflwctl -C . epic set 21-code-quality-architecture-hardening \
    --set priority=bogus --set created=not-a-date --dry-run --json; done
52  validation failed: priority must be high|medium|low, got "bogus"
 8  validation failed: must be YYYY-MM-DD, got "not-a-date"

$ # same shape against the task path, which sorts its keys:
$ for i in $(seq 20); do ./bin/tskflwctl -C . task set 6fbj87001m03 \
    --set priority=bogus --set created=not-a-date --dry-run --json; done
20  validation failed: must be YYYY-MM-DD, got "not-a-date"

$ for i in $(seq 40); do ./bin/tskflwctl -C . research set tui-authoring-interaction-checkpoint \
    --set description=<250 chars> --set tags= --set bogusfield=1 --dry-run --json; done
32  validation failed: description too long (250 > 200 chars)
 8  validation failed: unknown research field "bogusfield" ...
```

L1 reproduction (read-only):

```
$ ./bin/tskflwctl -C . task list --epic 00 --all   # 13 tasks
$ ./bin/tskflwctl -C . task list --epic 0  --all   # error: validation failed: unknown epic "0"
$ ./bin/tskflwctl -C . task list --epic 00-taskflow-v1-core --all   # 13 tasks
```

Hypothesis that did NOT hold (recorded because the probe is the evidence):

```
$ ./bin/tskflwctl -C . task set 6fbj87001m03 --set updated_at=2020-01-01 --dry-run --json
{"error":{"code":"validation","message":"validation failed: updated_at is stamped automatically and cannot be set"}}
```

Validation of this run is recorded in the PR body.

## Findings

#### M1. Epic and research field updates validate in random map order, so one request reports different errors · **Status:** open

**File:** internal/core/service_epic.go:341 | **Component:** core/entity-field-updates
**Effort:** XS · **Urgency:** soon

`Service.prepareEpicFieldUpdates` (`internal/core/service_epic.go:341`) and
`Service.SetResearchFields` (`internal/core/service_research.go:179`) both validate with
`for field, val := range updates` and return on the first rejection. Go specifies map
iteration order as unspecified, so when a request carries more than one invalid field,
*which* rejection the caller receives changes between runs of a byte-identical command.

The task twin does not have this defect, and says why in a comment directly above its loop
(`internal/core/service_task.go:794`):

> Maps do not preserve CLI argument order. Choose stable lexical validation order so the
> same request reports the same first error before any write.

That comment is the artifact of a fix for this same defect, in this same lens, three cycles
ago — see the Related-task observations section. The epic and research paths were left on the
old pattern even though their own comments claim to mirror the task contract ("mirroring the
set path (and the task/epic contract)", `internal/core/service_research.go:184`).

**Failing scenario:** `tskflwctl epic set 21-code-quality-architecture-hardening --set priority=bogus --set created=not-a-date`.
Both fields are invalid, so the command must fail — the defect is *which* failure it names.
Over 60 runs of that exact command (`--dry-run`, nothing written):

```
52  validation failed: priority must be high|medium|low, got "bogus"
 8  validation failed: must be YYYY-MM-DD, got "not-a-date"
```

The identically-shaped task command returned the same message 20/20 times. `research set`
reproduces it too: 32/40 vs 8/40 across `description` too long and an unknown field.

The practical cost is the one the originating audit described: an agent or script fixes the
field it was told about, re-runs, and is told about a different one, with no indication that
two were wrong — and a golden or doc example over a multi-invalid-field request cannot be
written at all.

**Why tests didn't catch it:** the regression that pinned this contract was written against
the task path only. Its evidence line reads "10/10 invalid multi-field requests named aaa" —
a repeated-run assertion, which is the right technique, but it was never parameterized over
the epic and research entry points. Go's randomized iteration also means a single-run test
over two bad fields passes most of the time (here, 87% and 80%), so an unrepeated test would
not have failed reliably even if one existed.

**Tightening (adjacent):** `ValidateDate`'s message ("must be YYYY-MM-DD, got ...") names no
field, so the epic `created` rejection above does not say which key it is about; the task
path's int rejection by contrast formats `%s must be an integer`. Worth naming the field
while touching this code.

**Follow-up:** the determinism contract is currently prose in one comment on one of three
sibling paths. A shared prepare-field-updates helper, or one table-driven regression
parameterized over task/epic/research, would stop the next entity from re-introducing it.
That is a design call, not part of the minimum fix.

**Recommendation:** Iterate slices.Sorted(maps.Keys(updates)) in prepareEpicFieldUpdates and SetResearchFields, exactly as prepareTaskFieldUpdates already does.

#### M2. A nil aggregate store panics instead of refusing on most task entry points · **Status:** open

**File:** internal/core/service_task.go:101 | **Component:** core/service-capabilities
**Effort:** S · **Urgency:** soon

`NewService` deliberately permits an absent aggregate store: `isNilCapability(store)`
normalizes a typed-nil to untyped nil, the field is set to that nil, every optional
capability is then detected by type assertion, and `validateSourceSets` passes a nil
store without complaint (`internal/core/service.go:322-394`). There is no error path for
a nil store — the constructor returns a usable `*Service`.

The Service guards that state in exactly three places: `planning_maintenance.go:64`,
`workspace.go:81`, and `ListTasks` (`service_task.go:39`), which returns a clean
`"epic reads are unavailable from this store"`. Roughly eighteen other call sites in
`service_task.go` alone dereference `s.store` with no guard, including these exported
entry points:

- `ShowTask` → `s.store.ReadTask` (`service_task.go:101`)
- `AcceptanceCriteria` → `s.store.ReadTask` (`:388`)
- `SetCriteriaTracked` → `s.store.ReadTask` (`:442`, `:449`, `:459`)
- `selectTaskMutationTarget` → `s.store.ResolveTaskSource` (`:501`), reached by
  `SetFields`, `ReplaceBody`, `AppendBody` and every body transform
- `EditTask` (`:554`), `NewTask` → `s.store.ReadEpics` (`:930`), `RenameTask` (`:1040`)

Calling a method on a nil interface panics, so each of these is a nil-pointer dereference
rather than the typed refusal the sibling capabilities all return. The asymmetry is the
finding: `TaskPath` guards `s.taskPaths`, `QueryFindings` guards `s.auditReads`,
`runTaskLifecycleMutation` guards `s.lifecycleMutations`, and `loadTaskGraphRecords` guards
its source — the convention is established and applied everywhere *except* the aggregate
store that the constructor explicitly allows to be nil.

**Failing scenario:** `core.MustNewService(nil, core.WithThreadStore(threads)).ShowTask("x")`
panics with a nil-pointer dereference. That composition is not hypothetical — this repo
builds it 20+ times across three packages (`internal/cli/ui_test.go:90`,
`internal/cli/composition_test.go:172`, `internal/cli/thread_test.go:350`,
`internal/tui/thread_projection_test.go:307`, `internal/core/service_thread_test.go` passim).
It is the standard way to assemble a Service from narrow capabilities only. Today those
tests happen to call only guarded or narrow-capability methods, so nothing panics in CI.

Not reachable through the shipped CLI, which always composes a complete `store.FS` — which
is why this is Medium rather than High.

**Why tests didn't catch it:** the nil-store compositions in the suite are built to exercise
*other* capabilities, so none of them calls an aggregate-store method. There is no negative
test asserting that an absent aggregate store produces `ErrValidation`-style refusals, the
way the suite does assert the refusals for task paths, audit snapshots and lifecycle
mutations.

**Context:** urgency is `soon` rather than matching severity because epic
`19-web-companion-apps-over-a-shared-core` (active) is precisely a second adapter over this
core, and epic `28-first-class-entities-new-planning-nouns` keeps adding entry points. A new
consumer composing narrow capabilities is the exact caller that meets a panic first.

**Tightening (adjacent):** six capability-unavailable errors in this file and
`internal/core/finding.go` wrap no domain sentinel — see L3; the guards added here should use
the same wrapped-sentinel form so the new refusals map to a real exit code.

**Follow-up:** whether the aggregate store should be *required* at construction instead
(making `NewService(nil, ...)` an error and narrow-capability composition explicit through
options only) is a composition-design question worth its own task. That would remove the
whole class rather than guarding each entry point, but it changes a constructor contract
20+ call sites depend on, so it is not the minimum fix.

**Recommendation:** Guard s.store with the existing isNilCapability helper on the exported entry points that dereference it, returning the same 'unavailable from this store' error ListTasks already returns.

#### L1. EpicRefKey compares the NN key lexically, so a zero-padded epic misses its unpadded reference · **Status:** open

**File:** internal/domain/epic.go:17 | **Component:** domain/epic-refs
**Effort:** XS · **Urgency:** eventually

`EpicRefKey` extracts the leading digit run and compares it as a *string*
(`internal/domain/epic.go:17`), so `"0"` and `"00"` are different keys even though its own
doc comment describes the value as "the leading NN number of the id/stem". Every epic join
in the tool runs through this key: `epicExists`, `canonicalEpic`, the `--epic` filter, and
the rollup that builds epic progress.

This repository has an epic whose stem is `00-taskflow-v1-core`, so the mismatch is live:

```
$ ./bin/tskflwctl -C . task list --epic 00 --all   # 13 tasks
$ ./bin/tskflwctl -C . task list --epic 0  --all   # error: validation failed: unknown epic "0"
```

Same epic, same intent, one spelling rejected. Any epic numbered with a leading zero is
reachable only by its padded spelling or its full stem.

**Failing scenario:** the command pair above. The wider latent case is a task whose
frontmatter carries `epic: 7` while the epic is `07-...`: it would join to no epic and sit
outside that epic's rollup entirely.

Severity stays Low because both symptoms are loud rather than silent. The filter returns
`ErrValidation` (exit 11) with "unknown epic", and `LintTask` independently flags a task
whose epic ref does not resolve as `unknown epic %q` (`internal/domain/lint.go:95`), so a
misjoined task is reported rather than quietly dropped from the rollup. No task in the
current corpus uses an unpadded or otherwise non-resolving epic ref — all 460 live refs are
full `<NN>-<slug>` stems (verified by grep over `planning/tasks/`).

**Why tests didn't catch it:** the corpus and the fixtures both use full stems or padded NN
spellings, so no test compares a padded stem against an unpadded reference. The behaviour is
a property of two spellings of the same number, which only shows up when a caller types the
shorter one.

**Recommendation:** Normalize both sides of the NN comparison (strip leading zeros, or compare the parsed integer) in EpicRefKey, keeping the whole-string fallback for legacy non-NN stems.

#### L2. taskGraphWithStatus discards the exists bool and would fabricate an empty-ID graph node · **Status:** open

**File:** internal/core/task_lifecycle.go:397 | **Component:** core/lifecycle-projection
**Effort:** XS · **Urgency:** eventually

`taskGraphWithStatus` reads `task, _ := graph.Task(taskID)` and discards the existence bool
(`internal/core/task_lifecycle.go:397`). For an unknown ID that yields a zero-value
`domain.Task` whose `ID` is `""`. It then sets the status and hands it to
`taskGraphWithTask`, where the replacement never matches any existing ID, so `found` stays
false and the zero task is *appended* as a new node keyed `""`
(`internal/core/task_lifecycle.go:403-418`). The result is an after-graph containing a
phantom empty-ID task, which is then used to derive the lifecycle receipt's `After` state
and downstream impacts.

This is latent, not live. Both call sites are guarded today:

- `validateExistingLifecyclePlan` checks `task, exists := graph.Task(plan.TaskID)` and
  returns `ErrNotFound` before reaching it (`:308-311`).
- `validateCreateAndStartPlan` validates `id.Valid(task.ID)` and inserts the seed into
  `beforeGraph` first (`:276-297`).

So there is no reachable input that trips it now, and I could not construct one read-only.
It is reported as a smell rather than a defect: the helper's safety is entirely a property
of its two callers, with nothing in the helper itself recording that contract. A third
caller — the lifecycle surface is actively growing — would get silent graph corruption
rather than an error, and the empty-ID node would flow into receipts and impact lists.

**Why tests didn't catch it:** correctly so — the guarded callers mean the branch is
unreachable, so no test can exercise it without calling the unexported helper directly.
The point is the missing guard, not a missing test.

**Tightening (adjacent):** `taskGraphAfterDependencyPlan` handles the same situation
explicitly (`if task, ok := tasks[write.TaskID]; !ok { continue }`, `:430-433`) — that is the
shape this helper is missing, two functions away in the same file.

**Recommendation:** Have taskGraphWithStatus check the exists bool from graph.Task and return the graph unchanged (or an error) rather than projecting a zero-value task.

#### L3. Capability-unavailable errors wrap no sentinel in six places and ErrValidation in one · **Status:** open

**File:** internal/core/service_task.go:40 | **Component:** core/error-policy
**Effort:** S · **Urgency:** eventually

The "this capability is not wired" condition is reported two different ways in `core`.

Wrapping no sentinel, so the CLI falls through to the generic exit code:

- `internal/core/service_task.go:40` — "epic reads are unavailable from this store"
- `internal/core/service_task.go:353` — "task graph reads are unavailable from this store"
- `internal/core/service_task.go:652` — "task lifecycle mutations are unavailable from this store"
- `internal/core/finding.go:156`, `:301`, `:362` — "audit snapshot reads are unavailable from this service"

Wrapping `domain.ErrValidation`, so the same class maps to exit 11:

- `internal/core/service_task.go:380` (`TaskPath`) — "task path resolution is unavailable
  from this service"

`AGENTS.md` lists wrapping the domain sentinels among the non-negotiables precisely so the
CLI's 10/11/13/14 mapping stays truthful, and the external pattern is the same one: exit
codes are an external contract, classified by walking the wrapped chain rather than by
message text. One condition that can surface as either 1 or 11 depending on which method
you called is a contract the caller cannot route on.

Low because it is not reachable through the shipped CLI — the real `store.FS` implements
every capability, so these branches only fire for a narrow or third-party adapter. It
becomes reachable at the same moment M2 does, and the two should be fixed together: the
guards M2 asks for will add more errors of exactly this class.

I did not pick the correct sentinel here on purpose. `ErrValidation` is defensible (the
caller asked for something this service cannot do) and so is leaving it unclassified (a
composition bug is not user input), but that is an error-policy decision for the owner, not
an audit finding to settle unilaterally.

**Why tests didn't catch it:** no test asserts an exit code or sentinel for the
capability-unavailable condition on these methods; the ones that cover absent capabilities
assert the message, which is identical either way.

**Recommendation:** Pick one sentinel for the capability-unavailable class and wrap it everywhere, or document why these are deliberately unclassified; do not leave the same condition mapping to two different exit codes.

## What audited clean

- `internal/core/store.go` — **clean.** Its one piece of real logic, `ThreadRead.ValidateSources`, is order-stabilized before iterating (`sort.Slice` on a copy), requires an adapter source ID per record, and rejects both duplicate canonical IDs and declaration/source disagreement with wrapped `ErrValidation`. The rest is port declarations and contract comments, which this lens cannot falsify.
- `internal/store/auditstore.go` + `internal/domain/audit.go` — **clean, and notably so.** `MoveAudit` runs `ValidateMove` against counts parsed from the exact content its CAS protects, and does so *before* the same-bucket no-op and the dry-run return, which is the ordering its port comment promises. The settlement arithmetic fails closed by construction: `UnsettledFindings = Findings - (Done + Dropped)`, so `tallyFindingStatus`'s `default:` arm — the exact shape this lens warns about — counts an unrecognized status toward *nothing* and therefore blocks a close rather than permitting one. That is the right polarity for a `default:`.
- `internal/core/task_lifecycle.go` — **clean.** Every refusal precedes effects; `ValidateTaskLifecycleSource` demands a fully healthy graph before authorizing; the override vocabulary is validated through an explicit `default: return false`; create-and-start rejects six distinct malformed-seed cases. The `switch plan.Override` without a `default:` is safe because `validTaskLifecycleOverride` already gated it.
- `internal/domain/body.go` criteria indexing — **clean.** `SetCriteriaState` bounds-checks `n < 1 || n > len(boxes)` against a 1-based index used as `boxes[n-1]`, validates *every* index and de-duplicates before mutating any line (the comment names partial application as the hazard), and rejects a newline in a reason that could manufacture phantom checkboxes.
- `internal/core/retry.go` — **clean.** The conflict loop is bounded, jittered, guarded against shift overflow, and never retries a committed write (`!result.Committed` in the lifecycle loop).
- The `updated_at` write path — **clean, hypothesis refuted.** I expected `task set --set updated_at=<date>` to validate and then be silently clobbered by the unconditional stamp at the end of `prepareTaskFieldUpdates`. `domain.ValidateField` rejects it first, and its comment states precisely that reasoning ("an explicit value would validate and then be silently clobbered — reject it instead"). The unset path has its own separate rejection. Both the task and epic paths refuse on set and on unset.
- No `fmt.Errorf` in the lens surface passes an error without `%w`, and none downgrades one with `%v`/`%s`. Checked by grep across `internal/core`, `internal/domain`, `internal/store`.

## External research

Three searches; `WebFetch` is blocked by this session's network policy (`go.dev` and
`reports.zellic.io` both failed to resolve), so these are search-result summaries that a
reader should re-verify at the source rather than fetched-and-quoted primary text.

- **Map-order nondeterminism is treated as a finding even when currently harmless.** A Zellic audit of Nibiru's oracle keeper raised `range` over a map inside an `EndBlocker` as a *High* finding while noting the operations then in place were order-independent and the immediate risk low; the reasoning was that a later change could make order matter. The recommendation was the same one M1 makes: iterate a sorted key slice. M1 is the stronger case — the ordering is already observable in output, not merely latent. <https://reports.zellic.io/publications/nibiru/findings/high-xoraclekeeper-iterating-over-maps-is-nondeterministic>, <https://oneuptime.com/blog/post/2026-01-23-go-map-iteration-order/view>, <https://leapcell.io/blog/iterating-over-maps-in-go-methods-order-and-best-practices>
- **The typed-nil interface trap is the real hazard behind M2, and this repo already owns the right tool for it.** A plain `x == nil` check on an interface misses a nil pointer wrapped in it, which is why a correct guard needs more than `== nil`. `core.isNilCapability` already does this, and is applied consistently to every *optional* capability; M2 is purely about extending the existing helper to the aggregate store rather than inventing a pattern. <https://oneuptime.com/blog/post/2026-01-23-go-nil-pointer-panics/view>, <https://www.claudepluginhub.com/agents/lerianstudio-ring-default-default/agents%2Fnil-safety-reviewer>
- **Exit codes are an external contract, classified by chain not by message.** The pattern these sources converge on — one sentinel or typed error per code, `%w` at every layer, a single classifier at the top of `main`, and message wording free to change without moving a code — is what taskflow already does via `ErrNotFound`/`ErrValidation`/`ErrAmbiguous`/`ErrConflict` → 10/11/13/14. L3 is the handful of places that sit outside it. <https://pkg.go.dev/github.com/glapsfun/gskill/internal/errs>, <https://pkg.go.dev/github.com/ANIAN0/filebrowser-cli/internal/errcode>, <https://dev.to/bala_paranj_059d338e44e7e/error-handling-in-a-go-cli-that-security-researchers-trust-2jj3>

No antipattern from the research is present in today's code beyond what M1–L3 already name.

## Related-task observations (propose-only)

- **Precedent, not overlap:** `planning/tasks/6gkg89ay3b5b-fix-task-identity-guards-and-deterministic-field-updates.md` (completed 2026-10-14) fixed exactly M1's defect class — its M3 came from `planning/audits/6gk9j6ff3ke9-2026-10-13-correctness-and-errors.md` (same lens, three cycles back) and its met criterion reads "invalid multi-field requests report a deterministic first error and write nothing", evidenced by "10/10 invalid multi-field requests named aaa". That work was scoped to the task path and to `store/frontmatter.go`; the epic and research twins were never extended. Both are completed records, so nothing was annotated — but a reviewer should know M1 is a *gap in a verified fix*, not a fresh discovery.
- No open task owns M1 or M2. Fingerprint greps over `planning/tasks/` for `prepareEpicFieldUpdates`, `SetResearchFields`, `slices.Sorted`, validation order, `isNilCapability`, "unavailable from this store", and nil-store composition returned only completed records (`6fbj87000m0a`, `6gcwcf8gzn50`, `6g5fy1m967ka`, `6gg7e59cyxxh`).

## Validation

| Check | Result |
| --- | --- |
| `git status --porcelain` | clean — one new file, `planning/audits/6gt1v0vt44yn-2026-11-03-correctness-and-errors.md` |
| `just build` | `just` is not installed here; documented fallback `go build -o bin/tskflwctl ./cmd/tskflwctl` passed (go1.26.0) |
| `just test` (`go test -race ./...`) | **2 packages red, both pre-existing and environmental** — see below. Every other package `ok`, including `internal/core` and `internal/store`. |
| `just lint` (`golangci-lint run ./...`) | **NOT RUN** — the installed binary was built with go1.25 and `go.mod` targets 1.26.0, so it refuses to load the config. This is exactly the toolchain skew the `Justfile` comment warns about; fixing it is a toolchain change, out of scope for this routine. |
| `tskflwctl lint` | ✔ all planning entities and dependency links pass lint |
| `tskflwctl audit lint 2026-11-03-correctness-and-errors` | ✔ all audit findings pass lint |

The two red packages are environment defects, not code defects, and neither touches this
run's change (both operate in their own temp directories, not the repository's `planning/`):

- `internal/tools/reviewworkspace` — every failure is `error: rsync is required`.
  `rsync` is not installed in this container.
- `cmd/tskflwctl` — only
  `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior`'s final filesystem assertion
  (`main_test.go:188`). It does `os.Chmod(tasksDir, 0o500)` and expects the write to be
  refused with exit 1; this session runs as uid 0, and root bypasses the permission bits,
  so the write succeeds and the assertion fails. All six published exit-code subtests pass.

Neither was repaired: per the routine's own instruction, a suite that was already red is
reported, not fixed.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — `tskflwctl task new "Make epic and research field-update validation order deterministic" --epic 21-code-quality-architecture-hardening --tags correctness,determinism,core --tier 2 --priority medium --description "Extend the task path deterministic-first-error fix to the epic and research field-update paths."`
- ○ M2 · open — `tskflwctl task new "Guard the aggregate store on core service entry points" --epic 21-code-quality-architecture-hardening --tags correctness,core,capabilities --tier 2 --priority medium --description "Return a typed refusal instead of panicking when the aggregate store capability is absent."`
- ○ L1 · open — `tskflwctl task new "Normalize the epic NN reference key numerically" --epic 21-code-quality-architecture-hardening --tags correctness,domain,refs --tier 3 --priority low --description "Make EpicRefKey treat 0 and 00 as the same epic number while keeping the non-NN exact-match fallback."`
- ○ L3 · open — `tskflwctl task new "Give capability-unavailable errors one sentinel policy" --epic 21-code-quality-architecture-hardening --tags correctness,errors,exit-codes --tier 3 --priority low --description "Classify the core capability-unavailable errors consistently so one condition maps to one exit code."`
