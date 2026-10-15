---
schema: 1
id: 6gk9j6ff3ke9
bucket: closed
area: correctness-and-errors
date: "2026-10-13"
updated_at: "2026-10-14"
---

# Audit: correctness-and-errors — 2026-10-13

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

#### M1. Self-track guard compares a resolved slug against an unresolved ref, so two of four ref shapes bypass it · **Status:** fixed

**File:** internal/core/service_task.go:444 | **Component:** core/task-criteria
**Effort:** XS · **Urgency:** soon

`SetCriteriaTracked` refuses to track a criterion to its own task, but the guard
compares the **destination's resolved slug** against the **caller's unresolved
source reference**:

```go
destinationRecord, err := s.store.ReadTask(trackedTo)   // resolved
destination := destinationRecord.Value.Task
if destination.Slug == slug {                            // slug is the raw CLI arg
```

`slug` arrives straight from the `task ac <ref>` positional (`internal/cli/task.go:538`)
and is only resolved later, inside `setCriteriaState` → `transformTaskBody` →
`store.TransformTaskBody`. Task resolution accepts the stable id, the human slug, a
unique case-insensitive prefix, or a substring (`FS.resolve`,
`internal/store/fsstore.go:347`), so one task has several legal spellings. The guard
holds only for the single spelling where the source ref happens to be the exact slug.

The identity the write records is the destination's **id**
(`FormatTrackedReason(destinationRecord.Source.ID, reason)`), so the comparison is
already working in the wrong currency: it checks a slug to protect an id-keyed pointer.

Two of the four ref-shape combinations bypass it. Probed against a throwaway planning
space with the freshly built binary:

```
$ tskflwctl -C probe task ac self-track-probe --tracked 1 --to self-track-probe
error: validation failed: a criterion cannot be tracked to its own task      # caught

$ tskflwctl -C probe task ac 6gk9gxtwrf1v --tracked 1 --to self-track-probe
✔ set criterion 1 tracked self-track-probe                                   # bypassed

$ tskflwctl -C probe task ac self-track --tracked 2 --to self-track-probe
✔ set criterion 2 tracked self-track-probe                                   # bypassed (prefix)

$ tskflwctl -C probe task ac self-track-probe --list
[ ]  1. <observable outcome>  tracked: by 6gk9gxtwrf1v
```

Task `6gk9gxtwrf1v` now carries a criterion tracked to `6gk9gxtwrf1v`. That is precisely
the state the method's own doc comment says the design exists to prevent — "a destination
held only as prose is a string nothing can follow" — except worse, because it resolves
successfully to a task that has handed the work to itself. Nothing in the criterion
vocabulary can distinguish it from a real handoff, and `lint` does not flag it.

**Failing scenario:** a task `6gk9gxtwrf1v-self-track-probe` with one unmet criterion.
Run `tskflwctl task ac 6gk9gxtwrf1v --tracked 1 --to self-track-probe`. Expected: the
`ErrValidation` refusal (exit 11). Actual: exit 0 and the criterion is written
`· **tracked:** by 6gk9gxtwrf1v` — a self-referential handoff. The same happens with any
prefix or substring spelling of the source ref.

**Why tests didn't catch it:** there is no test for the guard at all. `grep -rn "Tracked"
internal/core/*_test.go` returns nothing, and the refusal string
`"cannot be tracked to its own task"` appears only at its definition site — no test
constructs the self-reference state in any ref shape, let alone a mismatched pair.

**Recommendation:** Resolve the source task first and compare stable ids: reject when destinationRecord.Source.ID equals the source record's Source.ID, not when destination.Slug equals the caller's ref.

**Resolution:** Resolve both endpoints through ReadTask, require canonical
    source evidence, compare RecordSource.ID and bind the body transform to the
    resolved source ID. Portable drift/same-slug tests and 50 real CLI
    alias/preview combinations pass. Fresh-binary ID/prefix self-tracking probes
    exit 11 without changing the SHA-256; valid destination IDs survive rename.

#### M2. A case-only entity-id defect is reported as non-Crockford and told to replace the identity · **Status:** fixed

**File:** internal/store/flatname.go:62 | **Component:** store/identity-diagnostics
**Effort:** S · **Urgency:** soon

`flatNameProblem` explains a misspelled entity id by asking `id.InvalidChar` for the
first character outside the Crockford alphabet, then appends repair advice from
`id.Canonicalize`. The alphabet it checks against is **lowercase**
(`internal/id/id.go:35`), and `id.Valid` is documented as deliberately strict: "ids this
package mints are always lowercase". So any uppercase character in a filename id is
reported as a Crockford exclusion, and because `id.Alias` canonically decodes only
`i`/`l`/`o` (and their uppercase forms), every other uppercase letter falls through to
the `u` branch:

```go
reason := fmt.Sprintf("id %q contains %q at position %d, which Crockford base32 excludes (no i, l, o, u)", ...)
if fixed, ok := id.Canonicalize(cand); ok {
    return reason + fmt.Sprintf(" — rename to %q and match the `id:` field", ...)
}
return reason + " — u has no canonical decode, so this id must be replaced deliberately"
```

Probed with the freshly built binary against a throwaway planning space, on a file whose
only defect is case:

```
$ tskflwctl -C probe task list
! unidentified task record
    validation failed: invalid entity id: "6GK9GXTWRF1W-upper-probe.md" id "6GK9GXTWRF1W"
    contains "G" at position 1, which Crockford base32 excludes (no i, l, o, u)
    — u has no canonical decode, so this id must be replaced deliberately
```

Three independent falsehoods in one sentence:

1. `G` is **in** Crockford base32. It is in this package's own alphabet. The id is not
   non-Crockford; it is miscased.
2. The parenthetical `(no i, l, o, u)` is irrelevant to `G` and sends the reader looking
   for a character that is not there.
3. The tail claims `u` has no canonical decode and concludes the id "must be replaced
   deliberately" — on an id containing no `u` at all. This is the damaging part: it tells
   the operator to mint a **new identity** when the correct repair is a case-only rename
   that preserves the id, its 60-bit value, its sort position, and every inbound body link.

Reaching the state is ordinary: a manual `mv`, a copy through a tool that upper-cases, or
a checkout on a case-insensitive filesystem (macOS/Windows), where the id-led file can end
up cased differently from the `id:` field with git reporting nothing.

**Failing scenario:** rename any task file to upper-case its id
(`mv 6gk9gxtwrf1v-x.md 6GK9GXTWRF1V-x.md`), leaving the `id:` frontmatter untouched. Run
`tskflwctl task list`. Expected: "id is upper-case; rename to `6gk9gxtwrf1v-x.md`".
Actual: a claim that the id is non-Crockford plus advice to replace the identity — which,
if followed, discards a stable id that was never wrong.

**Why tests didn't catch it:** `TestEntityNameProblemDistinguishesStrayFromMistypedID`
(`internal/store/flatname_test.go:54`) enumerates exactly the three cases the original fix
set out to separate — a stray, an `i`/`l`/`o` typo, and a `u` — and asserts the
"replaced deliberately" wording **only** for a real `u`
(`entityNameProblem("6fbj87000ut6-u-id.md")`). Case is covered one layer down as a
rejection but never as a message: `internal/store/flatname_test.go:37` asserts
`splitFlatName("6FJVR03MR9ZG-slug")` returns `ok=false` and stops there. The diagnostic
for the fourth case has no test because nobody enumerated a fourth case.

**Recommendation:** Branch on case before reaching the Crockford message: when strings.ToLower(cand) is id.Valid, report a case-only defect and suggest the lowercase rename; and gate the 'u has no canonical decode' clause on the id actually containing a u.

**Resolution:** Shared strict ID diagnostics distinguish casing from excluded
    characters across filename, create and update paths. Case-only and mixed
    alias advice preserves canonical spelling; u advice is conditional on actual
    u/U. Automatic repair remains unchanged and refuses casing with accurate
    manual advice. Real JSON/list/fix probes retain bytes and filename spelling;
    focused and full race tests pass.

#### M3. Map iteration makes task set's written key order and its rejection message nondeterministic · **Status:** fixed

**File:** internal/store/frontmatter.go:141 | **Component:** store/frontmatter,core/task-set
**Effort:** XS · **Urgency:** soon

Two `range` loops over the same `map[string]any` make `task set` non-reproducible.

**Symptom 1 — the bytes written depend on map iteration order.** `updateFrontmatter`
applies updates by iterating the map (`internal/store/frontmatter.go:141`) and
`setMapNode` **appends** any key not already in the mapping
(`internal/store/frontmatter.go:374`). Replacing an existing key is order-independent;
appending two or more new keys is not. Eight runs of one identical command against one
identical input file, with the freshly built binary:

```
$ for i in $(seq 8); do cp pristine.md $F
    tskflwctl -C probe task set self-track-probe --force \
      --set alpha=1 --set beta=2 --set gamma=3 --set delta=4 --set epsilon=5
    awk '/^---$/{n++;next} n==1 && /^(alpha|beta|gamma|delta|epsilon):/{printf "%s ",$1}' $F; echo
  done
alpha: beta: gamma: delta: epsilon:
delta: epsilon: alpha: beta: gamma:
gamma: delta: epsilon: alpha: beta:
alpha: beta: gamma: delta: epsilon:
delta: epsilon: alpha: beta: gamma:
beta: gamma: delta: epsilon: alpha:
alpha: beta: gamma: delta: epsilon:
delta: epsilon: alpha: beta: gamma:
```

Five distinct orderings from one command. `AGENTS.md` lists preserved key order among the
non-negotiables, and the surrounding code takes that seriously — `setMapNode` goes out of
its way to carry head, line and foot comments across a replacement. The *append* path
quietly gives up the same property.

**Symptom 2 — which error you get is also a coin flip.** `Service.SetFields` validates by
iterating the same map (`internal/core/service_task.go:654`) and returns on the first
rejection, so with more than one unknown field the message names a non-deterministically
chosen one. Six runs:

```
$ tskflwctl -C probe task set self-track-probe --set zzz=1 --set aaa=2
error: validation failed: unknown field "zzz" ...
error: validation failed: unknown field "aaa" ...
error: validation failed: unknown field "zzz" ...
error: validation failed: unknown field "zzz" ...
error: validation failed: unknown field "zzz" ...
error: validation failed: unknown field "zzz" ...
```

An agent that fixes the field it was told about re-runs and is told about a different one,
with no indication that two were wrong. The same loop also decides whether the
graph-owned, `status`, and unknown-field rejections fire in a stable order, so a request
with two different classes of bad field reports an arbitrary one of them.

**Context:** urgency is `soon` rather than matching severity because the cost lands on the
project's own reviewers and agents rather than on data: a diff that shuffles keys is
re-reviewed, a golden over a multi-new-field write cannot be written at all, and the
twice-weekly audit routine itself writes `audit_sources` through this path.

**Failing scenario:** `tskflwctl task set <ref> --force --set alpha=1 --set beta=2`.
Expected: the same two keys in the same positions on every run. Actual: the pair is
appended in whichever order the runtime's map hash seed produces, so the file content —
and therefore the git diff — differs run to run for identical input.

**Why tests didn't catch it:** `TestUpdateFrontmatter_PreservesCommentsAndOrder`
(`internal/store/frontmatter_update_test.go:46`) is the test that owns key order, and it
passes a **single, already-present** key (`{"status": "in-progress"}`). It asserts that
existing keys keep their relative order — which this code does correctly — and never
exercises the append branch, so no test in the suite writes two new keys at once. Go's
randomized map iteration means even a test that did would only fail intermittently unless
it asserted over repeated runs.

**Recommendation:** Iterate a sorted key slice in both loops (slices.Sorted(maps.Keys(updates))) so appended frontmatter keys land in a stable order and the first rejection is always the same field.

**Resolution:** Core validates update keys lexically; the shared YAML writer
    applies keys lexically while replacing old keys in place. Repeated tests pin
    deterministic errors/new-key order and LF/CRLF comments, block scalars and
    body preservation. Fresh binary reports aaa for 10/10 invalid multi-field
    calls without byte changes, and reversed custom updates append
    alpha/beta/zeta.

#### L1. NewTask does not enforce the Next/Start exclusivity it documents; only cobra does · **Status:** fixed

**File:** internal/core/service_task.go:776 | **Component:** core/task-create
**Effort:** XS · **Urgency:** eventually

`NewTaskParams` documents the constraint on the field itself —
`Start bool // create directly in in-progress (mutually exclusive with Next)`
(`internal/core/service_task.go:776`) — but `NewTask` never checks it. With both set,
`status` is computed as `StatusNextUp` and then the `if p.Start` branch
(`internal/core/service_task.go:866`) routes creation through the lifecycle mutation to
`StatusInProgress`, so `Next` is silently discarded.

The exclusivity is enforced only in the cobra adapter
(`cmd.MarkFlagsMutuallyExclusive("next", "start")`, `internal/cli/task.go:204`). That is
exactly the split `NewTask`'s own comment argues against four lines above the defaults
block: "Defaults for zero-valued fields live here so EVERY caller — not just the CLI flags
— produces a valid, lint-clean task; … A second adapter calling NewTask without
replicating them would otherwise get ErrValidation or scaffold a lint-failing file." The
sibling constraint in the same function already follows that rule: `templateBodyConflict`
rejects `Body` + `Template` in core, not in cobra.

Labelled Low, not Medium, because no live defect follows from it today: the CLI is the only
caller that sets either field, and cobra refuses the combination before `NewTask` is
reached. It is a latent one. Task `6g4ym5y2dkp8` (create-entities-from-tui, ready-to-start)
adds a second primary adapter over these same `New*` use cases; a TUI create flow that
offers both toggles would inherit silent precedence instead of a refusal.

**Recommendation note:** this is the kind of thing worth folding into whichever change
touches `NewTask` next rather than a commit of its own.

**Why tests didn't catch it:** nothing asserts the pair is refused, and nothing could
notice the precedence — a core-level test passing both would simply observe
`in-progress` and read as correct.

**Recommendation:** Reject Next && Start in NewTask beside the existing templateBodyConflict check, so a second adapter gets the refusal rather than silent Start precedence.

**Resolution:** NewTask rejects Next plus Start with ErrValidation before epic
    reads, ID minting or creation, for both preview and write callers. Spy tests
    demonstrate no effects and the explicit exclusivity refusal; existing
    individual Next/Start regressions remain green. This is application
    validation, not a Cobra-only guard.

#### L2. ID-only criterion handoffs are writable but fail parsed lint · **Status:** tracked by 6gkkzvr92waq

**File:** internal/core/service_task.go:456 | **Component:** core/domain/criteria
**Effort:** S · **Urgency:** eventually

Owner-discovered during the detour, not part of the external audit's initial four findings. The fresh binary in an unregistered throwaway space accepted task ac --tracked 1 --to valid-destination with no --reason (exit 0). Its JSON acceptance view contained a tracked_by ID and omitted reason. The immediate lint --links returned exit 11: criterion 1 is tracked with no reason. Repeating the handoff with explanatory prose made lint green.

SetCriteriaTracked uses FormatTrackedReason to insert a by-ID prefix. SetCriteriaState sees nonempty raw text, but domain.TrackedDestination separates that ID from prose and lint checks the parsed explanation. The writer and linter consequently disagree. Existing comments/help and historic decisions do not unambiguously settle whether a resolved destination alone counts as explanation; the follow-up needs an explicit policy decision, not an assumed parser or wire change.

**Recommendation:** Decide whether separate explanation is required, then make write validation and parsed lint agree.

**Resolution:** Owner reproduced accepted ID-only handoff followed by lint
    refusal in an unregistered throwaway space. Policy choice is explicitly
    scoped to follow-up 6gkkzvr92waq after this detour, with no reason-contract
    change here.

## Punch list

Order: Critical → High → Medium → Low. No Critical or High findings this run.

- M1. Self-track guard compares a resolved slug against an unresolved ref, so two of four ref shapes bypass it  (effort: XS · urgency: soon)
- M2. A case-only entity-id defect is reported as non-Crockford and told to replace the identity  (effort: S · urgency: soon)
- M3. Map iteration makes `task set`'s written key order and its rejection message nondeterministic  (effort: XS · urgency: soon)
- L1. NewTask does not enforce the Next/Start exclusivity it documents; only cobra does  (effort: XS · urgency: eventually)

## Routine

`code-quality-audit` · lens `correctness-and-errors` · ISO week `2026-W42`, slot `Tue`
(index 0 = (42 × 2 + 0) mod 6).

## Files audited

Signal = churn (30d) × blast radius, scored `log(churn+1) × log(blast+1)` within the
lens surface (`internal/core/`, `internal/domain/`, `internal/store/`).

- **Signal**: `internal/store/fsstore.go` — top score (12 commits in 30d; the widest
  referenced surface in the lens, implementing twelve core ports). The flat-layout
  parse/resolve/write path every entity read passes through.
- **Signal**: `internal/core/service_task.go` — 13 commits in 30d, comparable blast
  radius. `internal/core/service.go` scored nominally second by 0.01 but its lint
  region was audited under this same lens on 2026-09-22 (M1 there, since fixed), so
  the third-ranked file — within scoring noise — was taken instead.
- **Adjacency** (from `fsstore.go`): `internal/core/store.go` — the hexagonal-seam hop
  the routine prefers: the consumer-side port file `FS` asserts against at
  `internal/store/fsstore.go:84`.
- **Adjacency** (from `service_task.go`): `internal/domain/body.go` — the domain
  collaborator behind the criterion write path (`SetCriteriaState`,
  `TrackedDestination`, `FormatTrackedReason`), i.e. where M1's bad state is serialized.
- **Random**: `internal/domain/finding_header.go` — passed the quality filter
  (non-generated, non-fixture, non-test, ≥50 lines).

Read for context rather than audited in their own right: `internal/id/id.go`,
`internal/store/flatname.go`, `internal/store/resolve.go`,
`internal/store/frontmatter.go`, `internal/domain/status.go`,
`internal/domain/finding.go`, `internal/cli/task.go`, and the test files for the
packages above (`flatname_test.go`, `frontmatter_update_test.go`,
`frontmatter_setmapnode_test.go`, `id_test.go`). Two of the four findings landed in
that context set rather than in a seed file, which is the cross-file yield the lens
rotation is for.

## Commands run

Build and validation:

- `go build -o bin/tskflwctl ./cmd/tskflwctl` — green (`just` is not installed in this
  environment; the spec's documented fallback).
- `go test -race ./...` — see the Validation note below.
- `./bin/tskflwctl lint` and `./bin/tskflwctl audit lint 2026-10-13-correctness-and-errors`
  — green.

Discovery:

- `git log --since="30 days ago" --name-only --pretty=format: -- internal/core internal/domain internal/store | ... | sort -rn` — churn ranking.
- `grep -rln "<ExportedName>" internal cmd` per candidate — blast-radius estimate.

Mutation probes, all against a throwaway planning space created with
`tskflwctl init` under the session scratchpad — never this checkout:

- `task ac <ref> --tracked <n> --to <ref>` across all four id/slug/prefix ref-shape
  pairings (M1). Two bypassed the guard; transcript in the finding.
- `task set --force --set <k>=<v>` ×5 new keys, eight runs from one pristine input
  (M3). Five distinct key orderings; transcript in the finding.
- `task set --set zzz=1 --set aaa=2`, six runs (M3). Two different fields blamed.
- A task file renamed to upper-case its id, then `task list` (M2). Transcript in the
  finding.
- `task set --set audit_sources=...` — accepted without `--force`, confirming
  `audit_sources` is a known field and that this routine's own step-10 write goes
  through the M3 code path.

No probe wrote to `internal/`, and no source file was edited to verify a claim.

## What audited clean

Two of the five seed files produced nothing, and that is a result rather than a gap.

- `internal/core/store.go` — **clean.** Almost entirely port declarations and contract
  prose; the three functions with behaviour hold up. `ThreadRead.ValidateSources` sorts a
  copy before validating, so the error it reports on a duplicate or identity-disagreeing
  Thread set is deterministic — the opposite of M3's defect, in the same lens, which is
  what makes M3 a local lapse rather than a house style. `LoadedThreads` clones each
  Thread before handing it out; `SemanticThreads` deliberately does not, and the comment
  says why.
- `internal/domain/finding_header.go` — **clean.** The two things worth suspecting both
  check out. `strings.Fields(line[depth:])[0]` at line 76 cannot panic: it is reached only
  after `nearMissHeaderRe` matched, and that pattern requires `#{2,6}[ \t]+` followed by a
  non-space, so the slice is never empty. And `ClassifyFindingHeaders` passes the *raw*
  body to `ParseFindings` while scanning `blankFences(body)` itself — which looks like it
  would let a fenced `#### H1.` example (the audit scaffold emits one, as does this file)
  pollute the `defined` set and suppress repair of a real near-miss with that code. It
  does not: `ParseFindings` blanks fences internally (`internal/domain/finding.go:147`).
  The `findingsDepth` enter/exit bookkeeping at lines 54–61 reads the pre-reset value for
  `insideFindings`, which is the correct order.
- `internal/store/fsstore.go` — **no finding of its own.** The write path in `SetFields`
  sequences correctly for this lens: authorize → reject planner call → reject
  lifecycle/graph-owned fields → resolve → read → update → parse-before-commit → domain
  validation → dry-run return → flock → version-CAS → atomic replace. Validation runs
  before the mutation, not after, and the parse-before-commit branch distinguishes "your
  update is bad" from "this file was already broken" rather than blaming the file. The
  helpers it delegates to are deterministic: `os.ReadDir` returns sorted entries, so
  listings and `FileProblem` order are stable. Its collaborator `flatname.go` is where M2
  lives.
- `internal/core/service_task.go` — **M1 and L1.** Separately checked and sound: the
  `--status` filter compares raw strings (line 73) but `domain.ParseStatus` is exact-match
  with no normalization or aliasing, so there is no spelling that validates and then
  matches nothing. `canonicalTaskGraphLoadProblems` sorts load problems through an explicit
  `cmp.Or` chain before projecting them, so partial-read diagnostics are ordered.
  `validatedTaskGraphRead` nils out both compatibility projections after rejecting a
  record, so a record without a canonical source ID cannot be reintroduced through
  `Records` or `Tasks`.
- `internal/domain/body.go` — **no finding.** The criterion writer is careful about the
  cases that actually bite: every index is validated before any line is touched, indices
  are deduplicated, a reason containing a newline is refused by name ("can manufacture
  phantom checkboxes"), and the disposition suffix is written at `box.end` rather than
  `box.line` so a wrapped criterion is not split mid-sentence. `stripCriterionSuffixLine`
  re-attaches the line's own indentation, with a comment explaining that losing it would
  silently detach a continuation line from its criterion.

On the lens's own high-yield pattern — a `switch` over a closed domain vocabulary with a
`default:` arm returning a benign value — nothing new surfaced in the audited set. The
2026-09-22 run of this lens found two (L1, L2) and both are tracked by
`6gcqz5aqt2sg`, which remains ready-to-start; this run is not re-reporting them.

## Related-task observations (propose-only)

- **Partial overlap**: `planning/tasks/6gcqz5b0j0dg-give-lint-diagnostics-stable-kinds-and-repair-applicability.md`
  (ready-to-start, epic 26) vs finding `M2`. The task assigns every lint defect a stable
  kind plus a closed repair applicability (`fix` / `manual` / `refused`) and explicitly
  keeps human messages in scope only as prose. M2 is the other half of the same surface:
  the case-only id defect would currently be classified `refused` ("must be replaced
  deliberately") when it is in fact mechanically fixable by a lowercase rename. Doing the
  task before M2 would bake the wrong applicability into the machine contract, where it is
  harder to correct than a sentence. Sequencing call for a human — the finding stays open.
- **Completed task whose fix is partial**: `planning/tasks/6fq9zy126tas-invalid-non-crockford-entity-id-is-mishandled.md`
  (completed) is the task that introduced the stray-vs-mistyped-id distinction M2 reports
  against. Its description names the original symptom as "a non-Crockford id is misreported
  as 'no leading id'", and for `i`/`l`/`o`/`u` it is genuinely fixed. The uppercase case was
  never enumerated, so one shape of the original symptom survives in the shipped code. Not
  proposing a status change; flagging that a completed task's acceptance does not cover it.
- **Completed task whose guard was never tested**: `planning/tasks/6g31g9f8x4cv-let-an-acceptance-criterion-say-more-than-done-or-not-done.md`
  (completed) shipped the criterion vocabulary and the `tracked` destination requirement
  that finding `M1` reports against. The destination-required guard has a test; the
  self-reference guard beside it has none. Worth a human deciding whether M1 is a follow-up
  on that task or new scoped work.
- **Possible sequencing home**: `planning/tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md`
  (ready-to-start) is a design pass over "identity, schema, locking, atomic-write and
  diagnostic foundations". M2 is identity-diagnostic work and may belong in that sequence
  rather than as a standalone task. Not claiming it — the candidate row above proposes the
  standalone shape, and the design task may absorb it.
- **No open task owns M1, M3, or L1.** Every keyword candidate in `planning/tasks/`
  matched on a technical fingerprint but resolved to a **completed** task
  (`6g1dhhk6721x` block-scalar refolding, `6fpfcecdymca` structure-aware body mutation)
  or to unrelated scope. Nothing was marked `tracked`, and no task frontmatter was changed
  by this run.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ✔ M1 · fixed — Implemented in 6gkg89ay3b5b — canonical criterion identities
- ✔ M2 · fixed — Implemented in 6gkg89ay3b5b — identity-preserving diagnostics
- ✔ M3 · fixed — Implemented in 6gkg89ay3b5b — deterministic field operations
- ✔ L1 · fixed — Implemented in 6gkg89ay3b5b — application lifecycle option validation
- → L2 · tracked — Resolve in 6gkkzvr92waq after the canonical-identity detour
