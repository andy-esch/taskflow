---
schema: 1
id: 6gf9nsb0d8ca
bucket: closed
area: readable-source-location-implementation-claude
date: "2026-09-30"
updated_at: "2026-09-30"
---
# Audit: Readable source-location and local repair-path implementation — claude — 2026-09-30

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Adversarial implementation review of optional readable source locations and the graph-repair local-path split. The implementation is uncommitted on `feat/portable-readable-source-locations`; review the captured sandbox state, not only the branch commit. Take a second, systemic pass after the checklist. Do not treat green tests or this brief as proof.

## Review target

- Task: `planning/tasks/6ge1bacd3bd2-preserve-readable-source-locations-in-portable-entity-diagnostics.md`.
- Read path: `internal/core/entity_read.go`, `service.go`, `finding.go`, `dependency_graph.go`, `dependency_source.go`, `dependency_repair.go`.
- Adapters/contracts: `internal/store/graphrepair.go`, `internal/cli/task_dependency_repair.go`, list commands/renderers, `internal/wire/dto.go`, `dependency.go`, `dependency_repair.go`, `envelopes.go`, schema/goldens and generated CLI docs.

## Intended contract to challenge

`RecordSource.ID` is canonical entity identity. Optional `RecordSource.Location` attributes readable physical occurrences but is not a lookup selector or filesystem path. Graph repair selects by task ID/slug or independently supplied `TaskGraphSourceRef.LocalPath`; an opaque `Location` may be a stale-context check but cannot select or become a materialization target. Pathless records are diagnosed yet cannot be locally repaired. Local repair, source revisions, whole-graph CAS, and partial-prefix recovery must remain intact. Ordinary task/epic/audit/research list, projected JSON, human, lint, and graph views carry nonredundant location context without introducing URI-valued `path` fields.

## Mandatory evidence floor

- Build a consumer inventory for `RecordSource.Location`, `TaskGraphSourceRef.Location`, `LocalPath`, `GraphProblem.Location`, list/lint DTOs, source declarations, repair manifests/receipts, and the filesystem writer. Trace one value from adapter read to each output/write boundary.
- Prove actual mutation evidence: run the newly added regression tests against the exact smallest code mutation they claim to kill; report the failing assertion. Also test at least one coordinated mutation that might preserve a shallow test while breaking the contract.
- Exercise non-default values through full and projected JSON and the generated schema: two equal-value records with the same canonical ID/slug at distinct opaque locations, a pathless source, local path plus different opaque location, absent location, and contradictory declared ID/filename metadata.
- Run `go test ./...`, relevant race tests, and a local repair dry-run plus the emitted remedy command in the sandbox. Record exact commands and outcomes.

## Required hostile angles

1. Can an opaque URI in a selector, YAML legacy `location`, current `path`, or normalized repair source reach `PlannedFiles`, a file read, or a write? Challenge ID/slug-only selectors on pathless or duplicate-ID records, and check that `--auto` does not silently repair them.
2. Can two equal records with the same ID/slug but distinct locations collapse in graph sorting, duplicate attribution, lint, source declarations, repair diagnosis, or list projection? Look for representative-record and legacy-reference misattribution.
3. Can a changed location, local path, readable/unreadable revision, or duplicate-record ordering pass `SameSourceSnapshot` or `SameRepairSnapshot` incorrectly? Include a pre-write CAS or interrupted-prefix scenario rather than just a pure value test.
4. Check all repair source consumers for accidental use of diagnostic `Location` after the split, including human suggestions, JSON `source.location` versus `source.path`, CLI parser, YAML compatibility alias, simulated edits, and filesystem containment.
5. Challenge the architecture seam: does deriving `LocalPath` from transitional domain `Path` introduce an implicit capability or allow a URI to masquerade as local? Does any downstream consumer infer canonical ID from location or stale frontmatter?
6. Find at least one independent systemic issue not named above if the evidence permits. If none, explain which hostile probes rule out the likely classes; do not fill the audit with speculative findings.

## Validation and restoration

Use the mandatory independent sandbox procedure above before any implementation inspection or test. Make mutation probes only there. Restore every probe to the sandbox baseline so only this assigned audit differs; then run the helper's verify and guarded transfer. Do not run generators or repair writes in the shared source checkout.

## Deliverable

Write evidence-backed findings in this assigned audit only, using the repository's finding grammar and leaving statuses open. For each finding include trigger, actual versus expected behavior, severity, exact file/line, minimal reproduction, and a concrete fix direction. Distinguish a proven defect from a design tradeoff or future task. Include the sandbox transfer attestation.

## Reviewer report

To be completed by the reviewer. Preserve this brief, then add findings, validation evidence, and a concise verdict below it.

### Executive verdict — `ready after M1 and M2; L1–L3 may be tracked`

For the only shipped adapter (filesystem), the safety contract holds. No URI reaches a selector,
`PlannedFiles`, a file read, or a write. ID-only selectors fail closed on a duplicate pair,
pre-write CAS and interrupted-prefix recovery still work, and every emitted repair command applies
cleanly. The adapter always sets `RecordSource.Location == domain Path`
(`internal/store/entity_read.go:14`), so every opaque-location behavior is currently exercised only
by in-process fakes.

Two problems should block closing the task as written:

- **M1, a proven defect.** For pathless records, repair diagnostics cannot be attributed. The
  receipt even reports a declaration as both addressed and residual. This contradicts acceptance
  criterion 1.
- **M2, systemic.** The `Location`/`LocalPath` split is guarded by fixtures in which the two values
  are equal. Eight of 21 targeted mutations survive, including removing `Location` from both CAS
  snapshot comparisons. Acceptance criterion 3 claims those guarantees are retained but does not
  test them.

### Sandbox attestation (helper `verify`, captured before the report edit)

```text
sandbox_path=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/ff2ed3bb-c7d3-4181-ab77-1a2ed55f1043/scratchpad/isolated-review.KFSm7e
git_dir=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/ff2ed3bb-c7d3-4181-ab77-1a2ed55f1043/scratchpad/isolated-review.KFSm7e/.git
baseline_commit=e2e43a4a9a3aaa46c1596a0d79cb11ce4cdeddc5
source_blob=a0cb195979976d734a6a82fd61c02156b9572f10
source_fingerprint=238cd2d6833c822cf5c8017168ba8686507f61b8
deliverable=planning/audits/6gf9nsb0d8ca-2026-09-30-readable-source-location-implementation-claude.md
deliverable_changed=false
transfer=pending
```

- **Clone:** independent, with its own in-tree `.git`. The baseline commit is
  `e2e43a4 chore: capture isolated review baseline`, on top of `0abda9b`.
- **Transfer result:** the helper's `transfer` refuses a second copy once the source deliverable
  changes, so this file cannot record its own transfer. The transfer output is reported in the
  hand-off message instead.
- **Toolchain:** go1.27.1 darwin/arm64, macOS 26.6.1, golangci-lint 2.13.2.
- **Scope of probes:** all probes, fixtures, and mutations ran inside the sandbox or a scratch
  fixture repo outside both trees. Nothing ran in the shared checkout.

### Consumer inventory (one value traced from adapter read to every boundary)

The filesystem read is the origin: `store/entity_read.go:14` sets `Source{ID: FilenameID, Location:
task.Path}`.

`core.NewTaskGraphRead` (`dependency_graph.go:333`) then calls `taskGraphSourceRefForRecord`
(`dependency_source.go:309`). That function builds `TaskGraphSourceRef{TaskID: Source.ID,
TaskSlug, Location: Source.Location, LocalPath: Value.Path}` and stores it in `TaskGraph.sourceRefs`.
From there the value reaches these boundaries:

- **Graph sort and representative selection:** `dependency_graph.go:402` keys on
  `source.Location`. The representative is `record.source` (line 450).
- **Duplicate-ID message:** `idPaths` (line 416) uses the location, falling back to
  `displayPath(Path)`.
- **`GraphProblem.Location`:** filled through `addProblem` (line 630). The wire layer suppresses it
  when it equals `Path` (`wire/dependency.go:57`). No human graph renderer prints it (see L1).
- **Source records, declarations, and repair defects:** `taskSourcePairs` and
  `sourceDeclarationSortKey` include both fields.
- **Repair-unavailable problem:** `requireLocalRepairPath` (`dependency_repair.go:308`) carries the
  location.
- **Repair receipt JSON:** `wire/dependency_repair.go:28` emits `source.location` **and**
  `source.path` unsuppressed (see L3).
- **Snapshot and CAS:**
  - `sameReadableTaskSources` (`dependency_graph.go:984`) compares the full ref.
  - `SameRepairSnapshot` (`dependency_repair.go:1198`) compares the full ref.
  - `changedSet` is keyed by the full ref.
  - `store/graphrepair.go` `repairPlanForSource` and `repairSourceRecord` compare by `==` on the
    full ref.
- **Stale-context checks:**
  - `resolveSourceTask` (`dependency_source.go:296`) returns `ErrConflict`.
  - `normalizeRepairEditSource` (`dependency_repair.go:835`) returns `ErrConflict`.

`LocalPath` comes only from `Value.Path`. It is consumed in these places:

- **Selection:** `resolveSourceTask` and `normalizeRepairEditSource` (exact match), gated by
  `hasLocalRepairPath` (`:322`).
- **Write:** `materializeTaskGraphRepair` (`store/graphrepair.go:191`) uses it for the
  `tasksDir` containment check, then `os.ReadFile` and `writeFileAtomic`.
- **Display:** `displayRepairSource` (`:1262`) produces planned, applied, and remaining files;
  `render/dependency.go:177` produces selectors and human output.
- **CLI input:**
  - Selectors: `parseGraphRepairSelector` rejects `://` (`cli/task_dependency_repair.go:125`).
  - YAML: `path`, or the legacy `location` key, rejected on `://` (`:217`), and refused when both
    are set (`:211`).
  - Relative paths are joined under `Cfg.Root` by `normalizeGraphRepairPath`.

Ordinary lists flow through `LoadedRecord.Source.Location` into:

- `wire.ToLoaded{Task,Audit,Research}JSON` and `ToLoadedEpicMeta`, which suppress the value when it
  equals `Value.Path` (`dto.go:390`)
- the opt-in `location` column (`render/columns.go:530`, epic at `:434`)
- the human ` [location: …]` suffix (`render/render.go:70`)

Lint flows through `readableDiagnosticLocation` into `LintResult.Location`, then to
`LintTaskJSON.location` and the `location:` line in `LintHuman`. Duplicate load problems pass
`problem.Location` through, unless it equals `LocalPath` (`service.go:818`).

### Findings

#### M1. Pathless repair defects lose declaration and occurrence identity, and the receipt contradicts itself · **Status:** fixed

**Trigger.** A readable task with no local path (`Value.Path == ""`) and an opaque `Location`, plus
any declaration defect. `requireLocalRepairPath` (`internal/core/dependency_repair.go:308`)
converts the defect to `Repairable=false` and replaces `Problem` with a constant
`repair-source-unavailable` problem. That problem carries only `TaskID` and `Location`: no field,
no value, no related ID, and a fixed message.

Three downstream consumers assume a non-repairable defect has no meaningful target:

- **JSON** omits `target` unless the defect is repairable (`internal/wire/dependency_repair.go:145`).
- **Human output** prints only `reason/code: message` (`internal/cli/render/dependency.go:69`, `:143`).
- **`repairDefectIdentity`** (`dependency_repair.go:742`) keys non-repairable defects on
  reason, code, task, related ID, field, path, and message. It ignores `Location` and the target.

**Actual vs expected.** I built one task ID with two opaque occurrences (`db://tasks/a`,
`db://tasks/b`), each declaring `depends_on: [bad-one, bad-two]`.

Actual output:

- **Human:** four identical lines,
  `⚠ invalid-dependency-id/repair-source-unavailable: source has no explicit local repair path; edit it through its adapter`.
  They name no task, location, field, or value.
- **JSON:** the two `residual` entries per location are byte-identical, so `bad-one` and `bad-two`
  cannot be told apart. The `problems` array still holds the details, but nothing joins a defect to
  its problem.

Expected: each defect stays attributable to its occurrence and its declaration, as acceptance
criterion 1 claims ("individually attributable in graph/list/lint diagnostics").

**Demonstrated misattribution (core receipt).** I used a pathless task P (`Location: db://tasks/p`)
with `depends_on: [Q, R]`, where local Q and R each depend on P. This gives two cycles, and both of
P's declarations are non-repairable `cycle` defects with identical identity keys.

1. Plan `--drop <Q path>:depends_on=P#0`.
2. `repairDefectDifference(before, after)` reports **P→R** as addressed.
3. `After.Defects` still lists **P→R** as residual.

The declaration that was actually addressed, P→Q, appears in neither list. That is because Q sorts
before R and the count-based matcher consumes the wrong key.

Severity: M. This is a proven contract defect. It is latent for the filesystem adapter, which never
yields a pathless record, but it is exactly the scenario the task introduces.

**Fix direction:**

- Keep the declaration on the defect. Either emit `target` for every defect that has one, marked
  non-replayable, or copy `Field` and the raw value into the `repair-source-unavailable` problem and
  its message.
- Include `Location` and the target edit key in `repairDefectIdentity`.
- Have both human renderers print task and location for non-repairable defects.
- Add a regression test that gives one pathless task two distinct bad values and asserts distinct
  JSON entries and correct `addressed`/`residual` membership.

**Resolution:** Pathless defects now retain a non-replayable source declaration
in JSON and human output; receipt identity includes the exact raw occurrence,
with regression tests.

#### M2. Fixtures where `Location == Path` mask the whole `Location`/`LocalPath` defect class, including both CAS snapshot guards · **Status:** fixed

**Trigger.** Most graph fixtures arrive with `Location` set to the same string as the local path:

- the compatibility conversion (`internal/core/service_task.go:200`, `:219`: `Location: task.Path`)
- the wire `loadedTask` helper (`internal/wire/envelopes_test.go:22`)
- the filesystem adapter itself

So in nearly every test, `Location ≡ LocalPath`. Any code that uses one where the other belongs,
or drops `Location` from a key that also contains `LocalPath`, stays green. Only about three new
focused tests construct `Location != Path`.

**Evidence: restored-mutation ledger.** Each mutation is a single-site edit, restored with
`git checkout` afterwards. "Killed" lists the failing assertion.

| # | Mutation (site) | Result |
|---|---|---|
| a | `addProblem` stops filling `Location` (`dependency_graph.go:630`) | killed: `dependency_graph_test.go:777` |
| b | `taskGraphSourceRefForRecord` sets `Location: Value.Path` (`dependency_source.go:310`) | killed: `dependency_graph_test.go:772`, `dependency_repair_test.go:74` |
| c | `sameReadableTaskSources` ignores `Location` (`dependency_graph.go:984`) | **survived** (core) |
| d | drop `requireLocalRepairPath` (`dependency_repair.go:281`) | killed: `dependency_repair_test.go:74` |
| e | `normalizeRepairEditSource` skips `hasLocalRepairPath` (`:838`) | killed: `dependency_repair_test.go:86` |
| f | `hasLocalRepairPath` drops the `://` check (`:323`) | **survived** (`go test ./...`) |
| g | task `LintResult` drops `Location` (`service.go:693`) | killed: `lint_source_test.go:263` |
| h | `readableSourceLocation` never suppresses (`wire/dto.go:390`) | killed: `envelopes_test.go:158`, `:217`, golden |
| i | selector `://` guard removed (`cli/task_dependency_repair.go:125`) | killed: `task_dependency_repair_test.go:86` |
| j | YAML `://` guard removed (`:217`) | killed: `task_dependency_repair_test.go:107` |
| k | `SameRepairSnapshot` ignores `Location` (`dependency_repair.go:1198`) | **survived** (core, store) |
| l | `resolveSourceTask` stale-location conflict removed (`dependency_source.go:296`) | **survived** (core, store, cli) |
| m | `normalizeRepairEditSource` stale-location conflict removed (`dependency_repair.go:835`) | **survived** |
| n | `projectedSourceEdge` drops the `sourceRefs` equality (`dependency_source.go:448`) | **survived** |
| o | `legacyReferenceIndex` reverts to `Location: diagnostic.TaskPath` (`dependency_repair.go:1160`) | **survived** |
| p | graph sort uses `Path`, not `Location` (`dependency_graph.go:402`) | killed: `dependency_graph_test.go:785` |
| r | task human location suffix removed (`render/render.go:70`) | killed: `render_test.go:168` |
| s | opt-in columns join the default table (`render/columns.go:172`) | killed: `task_list_csv`/`audit_list_csv` goldens |
| t | repair-unavailable problem drops `Location` (`dependency_repair.go:316`) | killed: `dependency_repair_test.go:74` |
| u | duplicate-load-problem lint drops `Location` (`service.go:822`) | **survived** (core, cli) |

**Coordinated mutation (c + k together).** I removed `Location` from both snapshot comparisons, then
added a probe with a single healthy record whose opaque location changes from `db://tasks/a` to
`db://tasks/moved`.

- As shipped, `SameSourceSnapshot` and `SameRepairSnapshot` both return `false`, which is correct.
- Mutated, both return `true`.
- `go test ./internal/core ./internal/store ./internal/cli` is still all `ok`.

The shipped test `TestTaskGraphReadableDuplicateLocationsRemainDistinct` passes under mutation c
only because its fixture is a duplicate-ID pair. Its `duplicate-task-id` problems carry `Location`,
so `sameGraphProblem` fails the comparison for it. That is an accidental neighboring guard, not the
source-pair guard the test claims to cover.

**Probe o, by contrast, shows correct behavior as shipped.** A record with a local path and a
different opaque location still yields its `legacy-reference-missing` defect. That defect is
silently lost under mutation o, and nothing notices.

Severity: M (test coverage). The implementation is correct today. But AC3 ("guarded mutation/CAS
tests retain their existing guarantees") covers a new dimension, `Location`, that no test isolates.

**Fix direction:**

- Add a fixture builder that defaults to `Location != LocalPath`, for example
  `Location: "opaque:" + id`, and run the existing graph, repair, and lint suites through it.
- At minimum, add healthy-graph assertions for c and k, and a mixed-location assertion each for
  l/m (stale conflict), n/o (legacy edge and index), u (lint), and f (a domain `Path` containing
  `://` is non-repairable).

**Resolution:** Added healthy location-only snapshot/CAS checks and
mixed-location stale-context, legacy, URI-path, and unreadable-duplicate tests.

#### L1. Human graph diagnostics never render location and collapse distinct occurrences · **Status:** fixed

`graphDiagnosticsHuman` (`internal/cli/render/dependency.go:311`) deduplicates on
`Code + "\x00" + Message` (`:314`) and never prints `Location`. Its callers are task blockers and
unblocks, and Thread list and show.

Probe: one ID at `db://tasks/a` and `db://tasks/b`, both with `depends_on: [bad-one]`. That yields
four `GraphProblem`s, which render as two lines with no location. The human view cannot say which
occurrence carries the invalid declaration. The JSON does carry it. The dedupe predates this change,
but the task now claims occurrence attribution, and only the JSON surface delivers it.

The filesystem adapter is unaffected, because its location is suppressed and paths were never shown
here. Severity: L.

**Fix direction:** include `Location` in the dedupe key and print it as `(location: …)` when it is
non-empty.

**Resolution:** Human graph diagnostics now retain and display distinct
nonredundant opaque source locations.

#### L2. "Explicit local path" is inferred from the transitional domain `Path` and screened by a `://` substring heuristic · **Status:** tracked by 6gcwcf88z57p

`taskGraphSourceRefForRecord` (`internal/core/dependency_source.go:309`) sets
`LocalPath: record.Value.Path`. `hasLocalRepairPath` (`internal/core/dependency_repair.go:322`)
accepts any non-empty value that lacks `://`.

Probe: one self-dependent task per value, with `Value.Path` set to `urn:taskflow:task:abc` and then
`tasks/opaque-key`.

| `Value.Path` | Diagnosis | `--auto` plan |
|---|---|---|
| `urn:taskflow:task:abc` | `repairable=true automatic=true` | one op, `LocalPath` set to that string |
| `tasks/opaque-key` | `repairable=true automatic=true` | one op, `LocalPath` set to that string |
| `db://tasks/x` | `repairable=false` | no op |

In production, `materializeTaskGraphRepair`'s `tasksDir` containment check
(`store/graphrepair.go:195`) is what fails closed. Because that check runs before any write, a single
such record would abort the whole `--auto` run with "outside the task directory". Meanwhile core keeps
advertising the defect as an inferable repair. The `://` branch is also untested (mutation f).

This is a design tradeoff the task acknowledges, since domain `Path` retirement belongs to the
source/path separation task. It is not a shipped-path defect. Severity: L.

**Fix direction:** have the adapter supply `LocalPath` as an explicit capability on the
graph record, rather than deriving it, and drop the string heuristic. Record this as a requirement
on the source/path separation task.

**Resolution:** The sequenced source/path split task now explicitly requires an
adapter-owned local graph-repair capability, replacing domain-Path inference and
URI heuristics.

#### L3. The repair receipt's `source` is not replayable key-for-key, and repeats `location` for every filesystem source · **Status:** fixed

`wire/dependency_repair.go:28` always emits both `location` and `path`. For filesystem sources the
two values are identical. Running the fixture's `--drop … --dry-run --json` shows both keys with the
same absolute path. This is unlike lists, lint, and graph problems, which suppress a redundant
location.

The YAML plan refuses an operation carrying both keys, even when they are equal
(`cli/task_dependency_repair.go:211`). The fixture plan
`path: tasks/…alpha-task.md` + `location: tasks/…alpha-task.md` fails with
`graph repair plan operation 1 cannot set both path and legacy location`. A consumer that maps the
receipt's `source.{path,location}` onto the plan's `{path,location}` is therefore rejected.

Schema 1.79 also changes the meaning of `source.location`. In 1.78 it was the replayable owner, and
its values are unchanged for the filesystem adapter, which keeps the `ADDITIVE` claim defensible.
The changed meaning should be called out in the 1.79 note.

Severity: L.

**Fix direction:** suppress `location` when it equals `path`, for consistency with the other DTOs,
or accept equal `path`/`location` in YAML.

**Resolution:** YAML repair plans now accept paired path/location: path selects
the local source, location supplies a stale-context check; equal pairs remain
compatible with filesystem receipts.

### Hostile angles: settled with evidence

1. **URI into a selector, plan, `PlannedFiles`, a read, or a write.** Ruled out for every shipped
   entry point. Each case ran against the fixture repo:

   | Input | Result | Exit |
   |---|---|---|
   | `--drop 'file:///…/alpha-task.md:depends_on=not-an-id#0'` | "is an opaque location, not a local path or task selector" | 11 |
   | `file:/…` (single slash) | joined under root, `not found` | 10 |
   | YAML `location: db://tasks/x` | validation error | — |
   | ID-only selector on the duplicate-ID pair | `ambiguous match … use its exact local path` | 13 |

   - `--auto` over a pathless graph plans no operations (shipped test, plus probe C3 for `db://`).
   - Core resolves `LocalPath` only by exact equality with a record's `Value.Path`. The materializer
     re-checks `tasksDir` containment before any read.
   - The residual gap, non-`://` opaque keys, is L2.

2. **Equal records collapsing.** Only partly ruled out, and the open findings say where.

   Correctly distinct:
   - the graph sort key and `sourceRefCounts`, both keyed on the full ref
   - lint per-record `Location`
   - list rows
   - the opt-in column, verified against full wire values by the shipped contract test

   Schema-valid with non-default values: a probe validated `TasksEnvelope`, `EpicsEnvelope`,
   `AuditsEnvelope`, `ResearchListEnvelope`, `LintEnvelope`, and `TaskGraphRepairEnvelope` against
   their `$defs`, all containing `"location":"db://records/b"`. All validate, and none emits a
   `"path"` key.

   Collapse confirmed only in repair defect identity (M1) and the human graph dedupe (L1). Records
   with absent locations and identical refs remain indistinguishable, which is expected because
   there is nothing to attribute.

3. **Snapshot or CAS passing incorrectly.** Ruled out as shipped. Pure-value checks:
   - a healthy location change returns `false`
   - reordering returns `true`
   - a changed revision returns `false`
   - identical-ref duplicates with swapped versions return `true` (same set)
   - a changed revision among identical-ref duplicates returns `false`

   Store scenarios, driven through the shipped hooks on a duplicate-ID pair at two local files:
   - **Pre-write CAS:** `testHookBeforeGraphRepairWrite` moves the non-target duplicate. Result:
     `ErrConflict`, `committed=false`, no applied files, target bytes unchanged.
   - **Interrupted prefix:** `testHookAfterGraphRepairWrite` stops after the first write. Result:
     `applied=[…zz-dup.md] remaining=[…zz-dup-shadow.md]`. The retry applies the remainder, and a
     third run is a no-op (`changed=false ops=0 selected=2`).
   - The remaining weakness is test coverage only (M2).

4. **Repair consumers using diagnostic `Location`.** Ruled out:
   - `displayRepairSource`, `repairSourceDisplay`, and the materializer read `LocalPath`.
   - The YAML alias maps only to `LocalPath`.
   - `SimulateSourceEdits` carries `Location` into the prospective graph, which keeps
     `repairSourceRecord` equality intact.

   In the fixture, all four emitted `--drop` commands (invalid ID ×2 on the duplicate pair, dangling
   ID, legacy missing) applied for real from the recommending state. Each changed exactly the
   declared value and `updated_at`.

   `--auto --dry-run` left the tree unchanged. `--auto` applied five operations across three files.

5. **Architecture seam.** No consumer infers canonical ID from location or from stale frontmatter:
   - The graph ID is `Source.ID` (via `taskGraphTasks`).
   - Source-less wire records emit no ID.
   - In the fixture, frontmatter `id: 6gf9x9zzzzzz` on file `6gf9x9tw9kc3-…` produces JSON id
     `6gf9x9tw9kc3` in both full and projected `-c id,slug,location`, plus a lint drift issue.

   The derived `LocalPath` capability is L2.

6. **Independent systemic issue.** Found: M2, fixture-induced masking of the whole
   `Location`/`LocalPath` class.

### Validation performed (all inside the sandbox)

| Check | Command | Result |
|---|---|---|
| Unit tests | `go test ./...` | all packages `ok` |
| Race tests | `go test -race -count=1 ./internal/core ./internal/store ./internal/wire ./internal/cli/...` | all `ok` |
| Vet | `go vet ./...` | clean |
| Lint | `golangci-lint run ./...` | `0 issues.` |
| Planning lint | `bin/tskflwctl -C <sandbox> lint` | "all planning entities and dependency links pass lint" |

The fixture was a scratch repo created with `tskflwctl init --no-register`, using isolated `HOME`
and `XDG_CONFIG_HOME`. It holds three tasks plus a duplicate-ID shadow file and injected invalid,
duplicate, self, dangling, empty-legacy, and missing-legacy declarations. Against it I ran:

- `task list` and `--json -c slug,id,location`
- `-o table -c slug,location`
- `epic list --json`
- `lint`
- `task depend repair` (diagnose, every emitted `--drop` dry-run and real, `--auto`, selector and
  plan variants)

Restoration: all probe test files were deleted and every mutation was reverted with `git checkout`.
Before the report was written, `git status --short` was empty.
