---
schema: 1
id: 6gch6xep2mh2
bucket: open
area: correctness-and-errors
date: "2026-09-22"
updated_at: "2026-09-22"
---

# Audit: correctness-and-errors — 2026-09-22

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

#### M1. Cross-kind task/Thread id-collision lint is blind to unreadable records · **Status:** open

**File:** internal/core/service.go:496-518 | **Component:** core/lint
**Effort:** S · **Urgency:** eventually

`Service.Lint` enforces the one cross-kind invariant ADR-0003/ARCHITECTURE state
as "task and Thread identities must be globally unique". It builds two identity
sets and checks each kind against the other.

Both sets are populated **only from readable records**. `taskIdentity`
(`service.go:475-481`) is filled from `ListTasksWithBodies`' decoded tasks;
`threadIdentity` (`service.go:506-515`) is filled inside `for _, thread := range
threads`. The unreadable records are skipped by both.

This is not a case of missing evidence. Three lines earlier in the *same loop*,
`service.go:500-504` reads `problem.ThreadID` off each `ThreadReadProblem` and
feeds it to `threadIDSources` for the duplicate-id check — the adapter
deliberately "recovers safe filename identity at that boundary"
(`docs/ARCHITECTURE.md`, `store/threadstore.go:48`). The recovered id is present
and already trusted for one check; the cross-kind check just never consumes it.

Net effect: the collision is reported while both documents parse, and **stops
being reported the moment either one becomes malformed** — that is, in exactly
the degraded state where a human is least able to see it by eye. It fails open,
not closed.

The write paths are not affected: `store/create.go`'s `ensureTaskIDNotThread`
and its Thread-creation counterpart scan *filenames*, so creation still refuses a
colliding id. Lint is therefore the only surface that reports a collision already
on disk, and it is the surface with the hole.

**Recommendation:** Seed threadIdentity from threadRead.Problems[].ThreadID and taskIdentity from the task FileProblems' EntityID, exactly as threadIDSources already does, so the collision check sees recovered identity.

#### L1. blockerReason's default arm mislabels a future status AND truncates the frontier · **Status:** open

**File:** internal/core/dependency_graph.go:1152-1158 | **Component:** core/graph
**Effort:** XS · **Urgency:** eventually

`dependency_graph.go` contains two switches over `domain.Status`. One is
protected against vocabulary growth; the sibling is not, and its default arm is
the more dangerous of the two.

`roleForStatus` (`:746`) returns `RoleUnknown` from its default, and
`TestTaskGraphLifecycleRoleCoversEveryPersistedStatus`
(`dependency_graph_test.go:180`) asserts every `domain.AllStatuses()` value maps
to a real role. `RoleUnknown` is also benign downstream: `isPendingWorkRole` is
false for it, so `deriveState` fails closed to not-eligible.

`blockerReason`'s trailing switch (`:1152`) has no such test and its default is
**not** benign — it returns `BlockerInvalidStatus` for a status that reached it
having already passed `task.Status.Valid()` at `:1143`. Two things then go wrong
at once: the blocker is reported to the user as `invalid-status` when the status
is in fact valid, and `isTerminalBlocker` (`:1118`) lists `BlockerInvalidStatus`
as terminal, so `projectBlockers` stops expanding there and
`BlockingFrontier` silently drops every constraint behind that task.

Labelled Low, not Medium, because it is latent: the four statuses that reach the
switch today are all handled, so no current repository state trips it. It is
reported because the cost of closing it is one test that already has a template
three hundred lines above it, and because the asymmetry looks like an oversight
rather than a decision.

**Recommendation:** Add a drift test mirroring TestTaskGraphLifecycleRoleCoversEveryPersistedStatus that asserts no domain.AllStatuses() value reaches blockerReason's default arm.

#### L2. Two vocabulary switches fall through to a zero value that reads as real data · **Status:** open

**File:** internal/domain/validate.go:179-186 | **Component:** domain,core/lint
**Effort:** XS · **Urgency:** eventually

Same class as L1, two more instances in the audited surface. Neither is
reachable today; both are reported because the fallthrough value is not
obviously wrong at the call site, which is what makes the class expensive later.

**`ParseRevisitDate`** (`domain/validate.go:179`) declares `var d time.Time`,
switches on the unit captured by `relativeDate`, and returns
`d.Format(time.DateOnly)` with no default arm. A unit that matches the regexp
but no case leaves `d` as the zero `time.Time` and returns the string
`"0001-01-01"` — a well-formed date that `IsRevisitDue` reports as immediately
due, so the snooze would fire instantly instead of erroring. The regexp
currently admits only `d|day|days|w|week|weeks`, so the gap opens the day
someone widens the pattern without touching the switch — and the code comment
right above it discusses adding months, which is exactly that edit.

**`dependencyLintIssues`** (`core/service.go:701`) switches over
`LegacyResolution` with no default. An unhandled resolution contributes nothing
to `parts`; if every reference on a field is of that kind, `message` is empty
and `:713` substitutes the literal text `"field is present but empty"` — lint
then makes a specific, checkable, false claim about a file that is not empty.

Note `ParseRevisitDate`'s other edge is already handled correctly: a huge offset
such as `999999999d` formats to a seven-digit year that `time.Parse` cannot read
back, and `task_lifecycle.go:218` re-validates with `ValidateDate` before any
write, so it fails closed (with a confusing message, but closed).

**Recommendation:** Give both switches an explicit default that returns a wrapped ErrValidation, and enable the exhaustive linter so the class is caught mechanically rather than by review.

## Punch list

- M1. Cross-kind task/Thread id-collision lint is blind to unreadable records  (effort: S · urgency: eventually)
- L1. blockerReason's default arm mislabels a future status AND truncates the frontier  (effort: XS · urgency: eventually)
- L2. Two vocabulary switches fall through to a zero value that reads as real data  (effort: XS · urgency: eventually)

Routine: `code-quality-audit` · lens `correctness-and-errors` · ISO week `2026-W39`,
slot `Tue` (index 0). Rotation: `(39 * 2 + 0) mod 6 = 0`.

## Files audited

- **Signal**: `internal/core/service.go` — churn 6/30d x blast radius 220 (score 10.5, highest in the lens surface); owns `Summary` and the whole-repository `Lint` sweep.
- **Signal**: `internal/core/dependency_graph.go` — churn 5/30d x blast radius 159 (score 9.1); the immutable graph projection every gate, frontier and Thread view reads through.
- **Adjacency** (from `service.go`): `internal/store/create.go` — hexagonal-seam hop from the core creation use case to the store guard that enforces its identity invariants.
- **Adjacency** (from `dependency_graph.go`): `internal/store/graphmutation.go` — hexagonal-seam hop from the graph invariant to the guarded writer that validates and CAS-protects it.
- **Random**: `internal/domain/validate.go` — drawn from the 76 non-test, >=50-line files in `core/`, `domain/`, `store/`.

## Commands run

- `go build -ldflags ... -o bin/tskflwctl ./cmd/tskflwctl` — OK (`just` is not installed in the routine container; the Justfile recipe was run directly).
- `go test -race ./...` — one FAIL, **pre-existing and environmental**, see below.
- `golangci-lint run ./...` — `0 issues.`
- `./bin/tskflwctl lint` — `all planning entities and dependency links pass lint`.
- **M1 reproduction**, against a scratch planning repo built with `tskflwctl init` / `epic new` / `task new` / `thread new`:
  - hand-write `threads/<TASK-ID>-collide.md` with valid frontmatter -> lint reports the collision from both sides, 2 items with issues;
  - replace that file's frontmatter with invalid YAML, keeping the filename -> `0 item(s) with issues, 1 unreadable file(s)`. The collision is gone.
  - control, proving the identity *is* recovered: an unreadable `threads/<THREAD-ID>-dup.md` beside a readable Thread of the same id **is** reported, by id and by recovered slug, through `DuplicateIDIssues`.
  - mirror, proving the gap is bidirectional: an unreadable `tasks/<THREAD-ID>-mirror.md` colliding with a readable Thread is likewise unreported.

### Pre-existing test failure (not caused by this run, not fixed here)

`TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` (`cmd/tskflwctl/main_test.go:188`)
fails with `filesystem failure = exit 0, want 1`. The test chmods `tasks/` to
`0o500` to force a write failure; this routine's container runs as **uid 0**, and
root bypasses the DAC write check, so the write succeeds. It is an artifact of the
execution environment, not a repository regression — the working tree was clean
(`git status --porcelain` empty) when it was observed. Out of scope per the routine spec.

## What audited clean

- `internal/store/create.go` — **clean.** `validEntityID` rejects an id that cannot legally appear in a flat filename before it can produce an unparseable file; slug legality is guaranteed upstream by `domain.Slugify`'s allowlist (every path separator and `..` is a word break) plus `NewTask`'s explicit empty-slug guard, and the store re-checks empty independently. `createEntityFile` keeps the identity scan and the `O_EXCL` write inside one repository critical section, which is the whole point of the helper. `CreateAudit` sets `a.Bucket` before serializing, not after.
- `internal/store/graphmutation.go` — **clean for this lens.** The named-return plus deferred `unlock` correctly `errors.Join`s a release failure onto an in-flight error rather than masking it. Materialization re-parses its own output and refuses to queue a write whose reloaded `depends_on` is not the planned set. A whole-snapshot CAS precedes the first write and a fresh per-target CAS precedes each one, so a durable prefix cannot be extended against changed bytes.
- `internal/core/dependency_graph.go` — **clean apart from L1**, including three things that looked wrong and are not:
  - `CandidateIDs[0]` (`service.go:703,705`) cannot panic: `LegacyResolved` is only ever assigned in the `case 1:` arm, and `LegacyUnsafe` is only ever a demotion of an already-resolved reference, so both carry exactly one candidate.
  - `RepresentativeCycles[componentIndex]` (`:501`) cannot panic: `stronglyConnectedCycles` builds `cycles` 1:1 from `components` (`:1345-1348`).
  - the `sync.Mutex` coverage is **not** partial. `unsoundPrerequisites` reads the `sound` cache instead of calling `computeSound`, so `g.sound`/`g.soundVisits` are written during construction only; the three post-construction caches are all taken under `g.mu`, and every accessor returns a deep copy. The documented "immutable, synchronized" contract holds.
- `internal/domain/validate.go` — **clean apart from L2.** Clearing a field routes through `domain.UnsetField` in `SetFields` (`service_task.go:467`) and never reaches `ValidateField`, so the enum validators cannot block a legitimate unset. `ValidateDescription` counts runes, not bytes.
- `internal/core/service.go` — `summarize`'s `append(append(p1, p2...), p3...)` (`:411`) is safe, though narrowly: `taskGraphFileProblems` returns `make([]FileProblem, len(problems))`, so there is no spare capacity for the nested append to write into. `Lint` defends the same hazard explicitly one layer up with `taskProblems := append([]domain.FileProblem(nil), problems...)` (`:493`). Every `fmt.Errorf` in all five audited files wraps a sentinel — zero bare error strings.

## External research

Run under step 11 (fewer than 3 Medium+ findings). All three strands support the
L1/L2 class rather than adding new defects.

- **`exhaustive` is the mechanical fix for the L1/L2 class, and it is already a golangci-lint linter** — no new tooling, one config block. Its enum definition is "any named type with an underlying float/string/integer type and at least one constant of its type in the same block", which matches `domain.Status`, `LifecycleRole`, `GateState`, `BlockerReason`, `LegacyResolution` and `AuditBucket` exactly. Crucially `default-signifies-exhaustive` defaults to **false**, so it flags a missing enum member *even when a default arm exists* — which is precisely the L1/L2 shape. Individual deliberate switches opt out with `//exhaustive:ignore <reason>`, which doubles as the comment explaining why the default is safe.
  - https://golangci-lint.run/docs/linters/
  - https://pkg.go.dev/github.com/nishanths/exhaustive
  - https://github.com/nishanths/exhaustive
- **`.golangci.yml` presently runs `default: standard` plus `depguard` only**, so nothing in CI checks switch exhaustiveness today. The file's own header already frames tightening as a tracked follow-up; `exhaustive` is a cheaper increment than the `gosec`/`wrapcheck`/`nolintlint` parity it names, because it needs no per-call annotations.
- **Slice-aliasing via `append(append(...))`** — the general hazard is that `append` reuses spare capacity and writes past `len` into a backing array a second name still points at; the standard defences are a full three-index slice expression `a[:len(a):len(a)]` or an explicit copy. This repo currently avoids it by allocating exact-length slices and, in `Lint`, by an explicit copy; worth knowing that the `summarize` case is safe by the callee's allocation rather than by anything visible at the call site.
  - https://blogtitle.github.io/go-slices-gotchas/
  - https://rednafi.com/go/slice-gotchas/

## Related-task observations (propose-only)

- Scope-adjacent, **not** a duplicate: `planning/tasks/6g5vm4efjcdv-make-repository-lint-load-diagnostics-adapter-neutral.md` vs finding `M1`. That task plumbs recovered identity *through* the lint diagnostic contract for pathless adapters; M1 is a lint **rule** that fails to consume identity the filesystem adapter already recovers. The task's "Out of scope" line says "Changing lint rules" in as many words, so M1 is not covered by it — but whoever picks it up will be holding exactly the right code, and its stress-test list already names "duplicate identities". Worth sequencing M1 immediately after, or folding in deliberately with a widened scope. Left `open` and cross-referenced on both sides rather than marked `tracked`.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ M1 · open — `tskflwctl task new "Lint cross-kind task/Thread id collisions on unreadable records" --epic 21-code-quality-architecture-hardening --tags lint,diagnostics,identity --tier 3 --priority medium --description "Seed the task and Thread identity sets in Service.Lint from recovered ids on unreadable records so a cross-kind collision stays reported when a document is malformed."`
- ○ L1 · open — `tskflwctl task new "Drift-test every domain.Status switch in the task graph" --epic 21-code-quality-architecture-hardening --tags testing,graph --tier 4 --priority low --description "Mirror TestTaskGraphLifecycleRoleCoversEveryPersistedStatus over blockerReason so a new status cannot become a terminal invalid-status blocker."`
- ○ L2 · open — `tskflwctl task new "Enable the exhaustive linter for closed-vocabulary switches" --epic 21-code-quality-architecture-hardening --tags lint,tooling --tier 4 --priority low --description "Turn on golangci-lint exhaustive with default-signifies-exhaustive false, annotate deliberate defaults with //exhaustive:ignore, and close the ParseRevisitDate and LegacyResolution fallthroughs."`
