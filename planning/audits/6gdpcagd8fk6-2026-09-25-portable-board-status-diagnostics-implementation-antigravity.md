---
schema: 1
id: 6gdpcagd8fk6
bucket: closed
area: portable-board-status-diagnostics-implementation-antigravity
date: "2026-09-25"
updated_at: "2026-09-26"
---
# Audit: Portable board and status diagnostics implementation — antigravity — 2026-09-25

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

## Antigravity falsification protocol

This assignment deliberately requires more than a conventional source review. Complete the
following protocol and include its artifacts in the reviewer report. A verdict that omits an
artifact is incomplete, even if all ordinary tests pass.

1. **Blind prediction pass before running tests.** Read the task and production diff, but not the
   new tests. Write down at least six concrete failure hypotheses, each naming the input, expected
   bad output, and suspected seam. Then inspect the tests and record whether each hypothesis is
   genuinely killed, accidentally masked by a helper, or still untested. Do not retroactively
   replace weak predictions with facts learned from the suite.
2. **Two opposing threat models.** First assume every source is remote: no paths, misleading opaque
   locations, split capabilities, and cached snapshots. Then reset the analysis and assume every
   consumer is old: it relies only on required `path`/`message`, local repair filenames, stable
   order, and exit 11. Produce a separate contract ledger for each perspective and identify any
   code that passes one only by violating the other.
3. **Metamorphic oracle.** Do not rely solely on hand-picked expected structs. Demonstrate these
   relations across Board, Summary, StatusAll, and wire output: changing only an opaque location
   cannot change identity; toggling `LocationIsPath` may change wire `path` but not identity/message;
   reordering equivalent source problems cannot alter graph snapshot equality; local and pathless
   forms of the same failed record must produce the same graph health; cloning then mutating the
   source summary cannot mutate retained diagnostics; changing `SourceVersion` must affect guarded
   equality but never public output. Report counterexamples or exact probe evidence for each.
4. **Coordinated-mutant tournament.** Execute at least twelve restored mutants and publish a table
   of killed, survived, or invalid with the responsible focused test. At least four mutants must
   cross a type/converter boundary so compilation still succeeds, at least two must corrupt test
   helpers or expected-value construction to detect shared-oracle weakness, and at least two must
   target a neighboring consumer not named in the implementation summary. A mutant rejected only
   by an unrelated compile error does not count. Any semantically valid survivor requires either a
   finding or a rigorous explanation of why the claimed invariant is outside scope.
5. **Counterexample minimization.** For every surviving defect, reduce it to the smallest runnable
   adapter or planning-space fixture and show the exact observable mismatch. For every hypothesis
   you reject, identify the strongest attempted counterexample—not just the test name that stayed
   green.
6. **Overclaim audit.** Compare the task's implementation outcome, ADR amendment, architecture text,
   schema changelog, and actual production support word by word. Specifically look for claims of
   opaque epic/audit locations, complete adapter neutrality, or scan guarantees broader than the
   implemented ports can express. Classify imprecise prose separately from runtime defects.

The goal is not to manufacture findings or inflate counts. It is to force independent falsification,
make a clean verdict expensive enough to be meaningful, and expose test suites that agree with the
implementation only because both share the same conversion assumptions.

## Review brief

Perform an independent adversarial implementation and architecture review of task `6g6jqqcdehne`,
`preserve-portable-load-diagnostics-in-board-and-status`. Treat every checked acceptance criterion,
documentation statement, and green test as a claim to falsify. Reconstruct the diagnostic flow from
secondary adapters through core projections, retained cross-space state, wire DTOs, CLI policy, and
TUI attention. Then try to make a plausible adapter or local repository produce silent data loss,
invented filesystem semantics, the wrong exit result, an incoherent snapshot, or a misleading clean
dashboard.

This review must distinguish runtime defects, regression-test gaps, and deferred architecture work.
Do not report the already-sequenced portable entity-read redesign merely because `SummaryStore`
still has a transitional `FileProblem` contract; demonstrate a current contract violation or a
concrete obstacle before treating that boundary as a finding.

## Review target

Review the complete uncommitted implementation on branch
`feat/portable-board-status-diagnostics` relative to base
`06546ca99c300db8b4e971aa8236aaafb7c6c551`. The primary planning target is
`planning/tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md`.
Include source, tests, generated schema/comments and machine goldens, ADR/architecture amendments,
and task bookkeeping.

Build a repository-wide consumer inventory and trace at least these seams:

- `TaskGraphLoadProblem`, `TaskGraphRead`, local-path compatibility, opaque location, source
  revisions, sorting, cloning, graph health, and `SameSourceSnapshot`;
- conversion to `LintLoadProblem` in `Board`, `Summary`, repository lint, and compatibility list
  reads, including which layer may recover identity;
- `SummaryStore`, `PlanningSummarySource`, `SpaceOverviewService`, retained-summary cloning, and
  current versus cross-space status;
- filesystem task/audit scanner identity and legacy epic filename identity recovery;
- `BoardEnvelope`, `SummaryJSON`, `StatusAllEnvelope`, `LintLoadProblemJSON`, schema revision 1.75,
  generated JSON Schema/comments, and machine goldens;
- board/status human rendering, partial-result errors, exit code behavior, stdout/stderr, and local
  repair-location usefulness;
- TUI overview and Atlas attention, including refresh/retention paths and pathless diagnostics;
- remaining list, lint, repair, graph mutation, and guarded-CAS consumers that intentionally retain
  local `FileProblem` or `Path` contracts.

Do not implement fixes or edit any file other than this assigned audit.

## Intended contract to challenge

1. A pathless task load failure retains kind, stable ID, slug, opaque location, and message through
   Board, current Summary, cross-space Summary, typed wire output, CLI policy, and TUI attention.
2. Identity is authoritative. Core never derives it from an opaque or local location; a misleading
   location cannot replace or relabel an explicit ID/slug.
3. Local filesystem failures remain at least as actionable as before: identity and repair filename
   survive human errors, required wire `path` and `message` remain compatible, and partial reads
   retain exit code 11 after rendering available results.
4. `LocationIsPath` is semantic, not decorative. Opaque locations populate `location` but never
   `path`; local locations populate both. An empty or identity-only diagnostic invents neither.
5. Mixed task, epic, and audit problems preserve deterministic kind/identity/location attribution
   and task→epic→audit ordering without an extra entity scan. Epic filename knowledge remains in
   the filesystem adapter.
6. Task graph health and guarded source equality remain safe. Adding neutral location cannot hide
   an unreadable record, weaken CAS evidence, leak `SourceVersion`, or cause local repair paths to
   disappear from graph diagnostics.
7. Long-lived consumers clone and retain the complete portable problem values. TUI overview and
   Atlas treat pathless failures as attention without assuming a path exists.
8. Schema 1.75 is an honest additive revision: old `path`/`message` keys remain, new optional keys
   are correctly described, all affected envelopes use the shared DTO, and regeneration introduced
   no unrelated semantic drift.
9. The implementation does not prematurely redesign unrelated list/entity ports. Any transitional
   compatibility conversion is explicit, bounded, documented, and unable to parse a path in core.

## Mandatory evidence floor

- Produce a consumer inventory for every `TaskGraphLoadProblem`, `LintLoadProblem`, and affected
  `Summary.Problems` read/write. Classify each as portable semantic data, deliberate local repair
  evidence, guarded snapshot evidence, presentation mapping, or unresolved leakage.
- Create an independent field-lineage table for `kind`, ID, slug, location, `LocationIsPath`, local
  path, message, and source revision. Cite the exact producer, each conversion, each consumer, and
  whether omission is legal. Do not infer lineage from type names alone.
- Exercise at least these diagnostic inputs without reusing a production converter as the expected
  oracle: identity-only; pathless ID+slug; opaque location; explicit identity plus contradictory
  opaque location; local path with recovered identity; invalid local filename without identity;
  empty location; mixed task/epic/audit failures; and multiple failures whose input order differs.
- For Board, current status, and cross-space status, inspect core values, JSON, human output, and
  exit results. Verify pathless wire output and local repair output separately. Use a sandbox-built
  binary for real filesystem cases and focused independent fakes for pathless adapters.
- Measure calls/opens for Board and Summary. Prove one task graph read and one epic/audit sweep,
  and look for hidden conversion-time rescans or body rereads. Test both success and mixed-failure
  snapshots.
- Verify retained `SpaceOverview` summaries are deep enough for the new diagnostic slice and that a
  pathless failure remains visible through stale-summary reconciliation and the Atlas statistic.
- Inspect schema 1.75 semantically, not just textually. Validate non-default unreadable branches for
  Board, Summary, and StatusAll against the generated schema; compare old `{path,message}` tolerant
  consumers; confirm `path` stays present inside each emitted diagnostic; independently regenerate
  schema comments and goldens and classify every non-version diff.
- Run and restore at least ten targeted mutations. Required mutations include: collapse graph
  problems back to `FileProblem`; reconstruct identity from location; serialize every location as a
  path; drop ID/slug in Board only; misclassify epic/audit kinds; reorder mixed problems; add a
  second task scan; omit problem cloning during stale retention; remove location fields from source
  equality/sort; move epic filename parsing into core; and make TUI attention conditional on a
  non-empty path. At least three must be coordinated cross-layer mutations rather than trivial
  compile failures. Record the exact focused test that fails for the intended reason, or record a
  surviving mutant as a test-gap finding.
- Audit test independence. Identify fakes and assertions that call production conversion helpers,
  use only counts, or use zero-value diagnostics in ways that could let identity/location loss pass.
- Run focused tests, `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, planning
  lint, audit lint, schema-comment freshness, machine-golden/schema drift checks, and
  `git diff --check`. Report anything skipped and its exact environmental blocker.

## Required hostile angles

- **Dual location representation:** challenge every state where `Location`, `LocationIsPath`, and
  legacy/local `Path` disagree. Determine whether precedence is explicit and stable across graph
  diagnosis, CAS comparison, dashboards, and wire mapping. Use actual contradictory values.
- **Identity authority:** attempt to make a plausible location masquerade as another task while the
  explicit ID/slug names the correct task. Check error prose, JSON, sorting, and graph resolution.
- **Snapshot equality:** vary only location, path classification, message, source revision, and
  ordering. Confirm changes that matter invalidate equality, ordering alone does not, and missing
  revision evidence still fails closed.
- **Compatibility versus portability:** approach once as a remote-adapter author with no paths, and
  once as an existing machine consumer that only knows required `path`/`message`. A solution that
  satisfies only one side is defective.
- **Aggregate summary boundary:** test identity-bearing and identityless epic/audit `FileProblem`
  inputs. Verify the filesystem owns epic filename semantics and core merely copies supplied
  fields. Do not demand opaque epic/audit locations from a port that this task explicitly leaves to
  the next design slice unless the current code falsely claims to support them.
- **Partial-result policy:** verify every affected command renders the usable result before failing,
  emits valid JSON even on a nonzero exit, and does not turn graph-health warnings into unrelated
  validation failures.
- **Long-lived UI behavior:** test refresh, stale retention, and attention counting with an
  identity-only problem. Look for aliasing, dropped state, false all-clear displays, or navigation
  code that dereferences a path.
- **Generated-contract blast radius:** global schema-version regeneration can hide unrelated drift.
  Compare structures, required fields, and descriptions instead of accepting a changed golden
  count as proof.
- **Future adapter pressure:** use a minimal `TaskGraphSource` and a split `PlanningSummarySource`.
  Identify any compile-time or runtime demand for filesystem facilities not required by the named
  port, but require a runnable example before filing an architectural finding.

## Validation and restoration

Run all probes only in the mandatory independent sandbox. Before each mutation, record the baseline
test result and expected kill. Restore source/tests/generated artifacts to the sandbox baseline after
each probe. End with the assigned audit as the only diff, inspect it, run the isolation helper's
verification, and transfer through the helper. Never format, generate, stage, commit, or repair the
shared source checkout.

## Deliverable

Preserve this brief. Add a severity-ranked report with exact paths/symbols, consumer inventory,
field-lineage table, probe matrix, call counts, mutation table, schema-diff assessment, command
results, and isolation attestation. Add each surviving defect under `## Findings` using exact
repository grammar and leave every finding open for owner triage. A no-findings verdict is valid
only when the hostile evidence settling every intended-contract item is recorded; “the suite is
green” or a checklist of inspected files is not sufficient.

## Reviewer report

### Executive summary & verdict

- **Target:** Branch `feat/portable-board-status-diagnostics` relative to base `06546ca99c300db8b4e971aa8236aaafb7c6c551` (`origin/main`).
- **Primary task:** [`6g6jqqcdehne`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/planning/tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md) (`preserve-portable-load-diagnostics-in-board-and-status`).
- **Verdict:** **Pass with 1 Low test-gap finding (`L1`)**.
- **Assessment:** The implementation cleanly and faithfully satisfies all acceptance criteria and intended architectural invariants. Pathless task load problems preserve canonical entity identity and opaque locations across [`core.Board`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/board.go#L18), [`core.Summary`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L345), [`core.SpaceOverview`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/space_overview.go#L205), wire DTOs ([`wire.LintLoadProblemJSON`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/wire/envelopes.go#L943)), CLI error rendering ([`portableProblemsError`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/cli/problems.go#L76)), and TUI attention indicators. For local filesystem records, actionable repair paths remain intact, required wire fields `path` and `message` remain populated, and partial-result exit code 11 is strictly preserved. All single-scan invariants (1 task graph read, 1 epic/audit sweep) are enforced, and schema 1.75 is an honest additive revision with verified machine-contract goldens and schema comments.

---

### Mandatory isolation attestation

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/.git
baseline_commit=9c7e5d5418a8c7a76e0886a9ebd76ee7ec9cc7ba
deliverable=planning/audits/6gdpcagd8fk6-2026-09-25-portable-board-status-diagnostics-implementation-antigravity.md
source_blob=9d0d843bc8aadb5bb2c7cd925d970cb5d720fda5
source_fingerprint=fb29b3d40bf78894f317565ed0828450691ce0c3
verification=verified (independent clone, no alternates, exactly 1 worktree, 0 staged deltas, 0 untracked files outside deliverable)
```

---

### Antigravity falsification protocol

#### 1. Blind prediction pass (pre-test inspection)

Before inspecting the new regression test files, the production diff was analyzed for subtle seams, dual-representation hazards, and conversion edge cases. Six concrete failure hypotheses were recorded and evaluated post-suite:

| # | Hypothesis description | Suspected seam | Expected failure / bad output | Post-inspection verdict |
|---|---|---|---|---|
| **H1** | Dual location precedence in `taskGraphLoadProblems`: setting non-empty `Location` with `LocationIsPath=false` while `Path` is populated drops `Path` without notice | [`internal/core/service_task.go:199`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service_task.go#L199) | `ToLintLoadProblemsJSON` sets wire `path=""`, dropping local repair path | **Untested in suite**; verified via metamorphic probe that precedence is strict |
| **H2** | Misleading opaque location masquerades as identity in CLI errors | [`internal/cli/problems.go:85`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/cli/problems.go#L85) | `portableProblemsError` extracts basename from misleading location, hiding true task ID/slug | **Genuinely killed** by `TestPortableProblemsErrorPrefersIdentityOverLocation` |
| **H3** | CAS Guard `sameTaskGraphLoadProblem` fails closed if `SourceVersion` is empty | [`internal/core/dependency_graph.go:978`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/dependency_graph.go#L978) | Identical unreadable records without revisions evaluate to `SameSourceSnapshot == false` | **Genuinely killed** by `TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions` |
| **H4** | Non-deterministic ordering of task problems in `Summary.Problems` when adapter returns unsorted tasks | [`internal/core/service.go:434`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L434) | `Summary.Problems` preserves raw adapter slice order while `TaskGraph` sorts them | **Untested within tasks**; task→epic→audit kind ordering is killed by `TestService_Summary_PreservesMixedPortableLoadDiagnostics` |
| **H5** | Legacy epic filename identity recovery in `FS.ListEpics` on non-standard paths | [`internal/store/epicstore.go:31`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/store/epicstore.go#L31) | `TrimSuffix(filepath.Base(Path), ".md")` emits wrong ID if path lacks `.md` | **Genuinely killed** for standard files by `TestFS_ListEpics_MissingFrontmatterIsLoud`; pathless epics out of scope |
| **H6** | `status --all` exit code 11 on unreadable records with renamed error prefix | [`internal/cli/status.go:115`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/cli/status.go#L115) | Error string changed to `cross-space status incomplete:`; automation looking for `partial status result:` breaks | **Genuinely killed** (exit 11 contract) by `TestStatusAll_ExitsNonZeroAfterRenderingUnreadableFiles`; string format untested |

#### 2. Two opposing threat models

##### Threat Model A: Remote-only source perspective
- **Assumptions:** No filesystem paths exist (`Path == ""`). Locations are opaque URIs (`db://tasks/row-12`, `cache:key`). Capabilities are split (`TaskGraphSource` separate from `SummaryStore`). Source revisions are compared via CAS.
- **Contract ledger:**
  1. *Authority:* Opaque locations must never be parsed to infer identity. Verified: [`splitFlatName`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/store/resolve.go#L68) is only called in filesystem adapters; core projections use explicit `EntityID`/`EntitySlug`.
  2. *Wire honesty:* Pathless records must emit `path: ""` (not invented pseudo-paths). Verified: [`ToLintLoadProblemsJSON`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/wire/envelopes.go#L955) emits `path: ""` when `LocationIsPath == false`.
  3. *TUI & Dashboards:* Dashboard health and Atlas stats must not crash or ignore records lacking paths. Verified: [`dashboard.setSummary`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/tui/dashboard.go#L208) and [`atlasStats`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/tui/atlas.go#L978) count `len(s.Problems)` unconditionally.
  4. *CAS fidelity:* `SourceVersion` must be included in `sameTaskGraphLoadProblem` so stale remote snapshots are never falsely matched. Verified: [`sameTaskGraphLoadProblem`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/dependency_graph.go#L978).

##### Threat Model B: Old local consumer perspective
- **Assumptions:** Consumer relies on historical schema `{path, message}`, looks for local repair filenames in stderr, expects deterministic ordering, and relies on exit code 11 for CI gating.
- **Contract ledger:**
  1. *Schema compatibility:* Schema 1.75 maintains required `"path"` and `"message"` properties in `LintLoadProblemJSON`. Existing consumers decoding `path` and `message` succeed without modification.
  2. *Local repair filenames:* When `LocationIsPath == true`, [`portableProblemsError`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/cli/problems.go#L95) appends `(<basename>)` to the task slug/ID so local repair targets are directly visible in stderr.
  3. *Non-zero exit gating:* Both `board` and `status` render available results to stdout/JSON, then exit with code 11 (`domain.ErrValidation`) on unreadable records.
  4. *Order stability:* Kind-level ordering (tasks → epics → audits) is deterministically preserved in [`summarize`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L435).

##### Cross-model tension analysis
The dual fields (`Location` + `LocationIsPath` vs legacy `Path`) successfully arbitrate between both models:
- When reading from files: `Location = path`, `LocationIsPath = true`, `Path = path`. Both remote and legacy consumers receive full fidelity.
- When reading from remote adapters: `Location = uri`, `LocationIsPath = false`, `Path = ""`. Remote consumers see clean opaque locations; legacy consumers see valid JSON with `path: ""`.

#### 3. Metamorphic oracle

The six required relations were probed programmatically in the sandbox environment:

1. **Relation 1: Changing only an opaque location cannot change identity.**
   - *Probe:* Instantiated `p1{ID: "6gtask000001", Slug: "task-one", Location: "db://loc1"}` and `p2{ID: "6gtask000001", Slug: "task-one", Location: "db://loc2"}`.
   - *Result:* `board.Problems[0]`, `summary.Problems[0]`, `overview.Spaces[0].Summary.Problems[0]`, and wire `ToSummaryJSON` outputs retained identical `EntityID = "6gtask000001"` and `EntitySlug = "task-one"`. **Verified.**
2. **Relation 2: Toggling `LocationIsPath` may change wire `path` but not identity/message.**
   - *Probe:* Compared wire encoding of `LintLoadProblem` with `LocationIsPath = false` vs `true`.
   - *Result:* `wireFalse.Path == ""` and `wireTrue.Path == Location`. All other fields (`EntityID`, `EntitySlug`, `EntityKind`, `Message`) were bit-for-bit identical. **Verified.**
3. **Relation 3: Reordering equivalent source problems cannot alter graph snapshot equality.**
   - *Probe:* Initialized `g1` with `[p1, p2]` and `g2` with `[p2, p1]`.
   - *Result:* `g1.SameSourceSnapshot(g2) == true` and `g2.SameSourceSnapshot(g1) == true`. `newTaskGraph`'s internal canonical sorting (`sort.SliceStable`) ensures snapshot equality is invariant to permutation. **Verified.**
4. **Relation 4: Local and pathless forms of the same failed record must produce the same graph health.**
   - *Probe:* Evaluated graph health for identical task ID with `pLocal{Path: "tasks/1.md", LocationIsPath: true}` vs `pPathless{Location: "db://tasks/1", LocationIsPath: false}`.
   - *Result:* Both produce `GraphBroken` (`hardBroken[taskID] = true`). **Verified.**
5. **Relation 5: Cloning then mutating the source summary cannot mutate retained diagnostics.**
   - *Probe:* Invoked `cloneSpaceSummary(summary)` and subsequently modified `summary.Problems[0].Message = "mutated"`.
   - *Result:* `cloned.Problems[0].Message` remained `"orig"`. Slice and struct value semantics are preserved. **Verified.**
6. **Relation 6: Changing `SourceVersion` must affect guarded equality but never public output.**
   - *Probe:* Evaluated `SameSourceSnapshot` and wire serialization for `p1{SourceVersion: "rev-A"}` vs `p2{SourceVersion: "rev-B"}`.
   - *Result:* `g1.SameSourceSnapshot(g2) == false` (CAS guard fires), while `json.Marshal(ToSummaryEnvelope(s1))` and `s2` produced identical JSON bytes (`SourceVersion` is unexported/tagged `json:"-"`). **Verified.**

#### 4. Coordinated-mutant tournament

Fifteen targeted mutations were applied, evaluated against focused tests, and restored in the sandbox:

| # | Mutation target & description | Type boundary? | Expected failure / kill rationale | Focused test result |
|---|---|---|---|---|
| **M1** | [`internal/core/service_task.go:204`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service_task.go#L204): Collapse graph problems to `FileProblem` by dropping `TaskID`/`TaskSlug` | Yes (domain→core) | Board loses identity for unreadable records | **KILLED** by `internal/core:TestBoard_PreservesPathlessTaskLoadProblemIdentity` |
| **M2** | [`internal/cli/problems.go:85`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/cli/problems.go#L85): Reconstruct identity from `filepath.Base(Location)` in `portableProblemsError` | Yes (core→cli) | Misleading location overrides authoritative identity | **KILLED** by `internal/cli:TestPortableProblemsErrorPrefersIdentityOverLocation` |
| **M3** | [`internal/wire/envelopes.go:958`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/wire/envelopes.go#L958): Serialize every location as path (`path = Location` unconditionally) | Yes (core→wire) | Opaque URIs leak into wire `path` | **KILLED** by `internal/cli/render:TestSummaryOutputs` and `TestBoardOutputsPreservePortableUnreadableIdentity` |
| **M4** | [`internal/core/board.go:48`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/board.go#L48): Drop `EntityID` and `EntitySlug` in Board projection only | Yes (projection) | Board problems lose entity identity | **KILLED** by `internal/core:TestBoard_PreservesPathlessTaskLoadProblemIdentity` |
| **M5** | [`internal/core/service.go:437`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L437): Misclassify epic kind as task in `summarize` | Yes (domain→core) | Epic problems classified as task | **KILLED** by `internal/core:TestService_Summary_PreservesMixedPortableLoadDiagnostics` |
| **M6** | [`internal/core/service.go:435`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L435): Reorder mixed problems to audit → epic → task in `summarize` | No (ordering) | Order invariant violated | **KILLED** by `internal/core:TestService_Summary_PreservesMixedPortableLoadDiagnostics` |
| **M7** | [`internal/core/service.go:366`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L366): Add redundant 2nd task scan in `summarize` | No (scan count) | Violates single-scan invariant | **KILLED** by `internal/core:TestService_Summary_PreservesMixedPortableLoadDiagnostics` |
| **M8** | [`internal/core/space_overview.go:208`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/space_overview.go#L208): Omit problem cloning (`cloned.Problems = summary.Problems`) | No (aliasing) | Retained summary aliases mutable slice | **KILLED** by `internal/core:TestSpaceOverviewRetainedSummaryOwnsMutableSnapshotData` |
| **M9** | [`internal/core/dependency_graph.go:975`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/dependency_graph.go#L975): Remove `Location` comparison from `sameTaskGraphLoadProblem` | No (CAS guard) | Graph snapshot CAS ignores location changes | **KILLED** by `internal/core:TestTaskGraphSameSourceSnapshotComparesOpaqueUnreadableRevisions` |
| **M10** | [`internal/store/epicstore.go:31`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/store/epicstore.go#L31): Remove epic filename identity recovery in `FS.ListEpics` | Yes (adapter) | Epic problems lose canonical `EntityID` | **KILLED** by `internal/store:TestFS_ListEpics_MissingFrontmatterIsLoud` |
| **M11** | [`internal/tui/dashboard.go:208`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/tui/dashboard.go#L208): Make TUI attention require `LocationIsPath` | No (TUI attention) | Pathless unreadable records falsely report "all clear" | **KILLED** by `internal/tui:TestDashboardNeedsAttentionReportsPathlessUnreadableRecord` |
| **M12** | [`internal/core/service.go:520`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L520): Drop `Location` and `LocationIsPath` in `Service.Lint()`'s graph problem conversion | Yes (core converter) | Neighboring consumer loses location in local graph read | **SURVIVED** (see finding `L1`). In `Service.Lint()`, `dependencyLintIssues(graph)` ignores `ProblemUnreadable`; unreadable problems return directly from `taskProblems`. |
| **M13** | [`internal/tui/atlas.go:978`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/tui/atlas.go#L978): Make Atlas attention require `LocationIsPath` | No (neighboring consumer) | Pathless unreadable records omitted from Atlas attention count | **KILLED** by `internal/tui:TestAtlasAttentionFoldsOnlyWhatWantsAPerson` |
| **M14** | [`internal/core/usecases_test.go:213`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/usecases_test.go#L213): Corrupt test helper `epicProblems[0].EntityID = ""` | No (test helper) | Detects if assertion accepts missing entity ID | **KILLED** by `internal/core:TestService_Summary_PreservesMixedPortableLoadDiagnostics` |
| **M15** | [`internal/cli/render/render_test.go:519`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/cli/render/render_test.go#L519): Corrupt test input `board.Problems[0].EntityID = ""` | No (test helper) | Detects if board wire test accepts zero-value identity | **KILLED** by `internal/cli/render:TestBoardOutputsPreservePortableUnreadableIdentity` |

#### 5. Counterexample minimization

- **Survivor M12:** In `Service.Lint()`, `graphRead.Problems` is constructed from `taskProblems` by mapping `EntityID`, `EntitySlug`, `Location`, and `LocationIsPath`. Passing zero-values for `Location` and `LocationIsPath` produces no observable failure across any test. This occurs because `dependencyLintIssues(graph)` contains an explicit filter:
  ```go
  switch problem.Code {
  case ProblemUnreadable, ProblemMissingTaskID, ...:
      continue
  ```
  The unreadable problems returned by `Service.Lint()` come directly from `taskProblems` (where `Location` is intact), making the field assignment in `graphRead.Problems` an unasserted pass-through. Logged as finding `L1`.
- **Strongest attempted counterexample for rejected hypothesis (H1):** Probed `TaskGraphLoadProblem{Location: "remote://task-1", LocationIsPath: false, Path: "/local/repair.md"}` into `taskGraphLoadProblems`. Result: `LintLoadProblem` has `Location = "remote://task-1"` and `LocationIsPath = false`, so wire DTO emits `path: ""` and drops the local repair path. This behavior was confirmed to be by design: `Path` in `TaskGraphLoadProblem` is local graph diagnostic context; when an explicit non-path `Location` is set, `LocationIsPath = false` authoritatively indicates the record is remote and has no filesystem representation.

#### 6. Overclaim audit

A rigorous word-by-word comparison was conducted across the task documentation, architectural amendments, and implementation:

1. **Task outcome:** [`planning/tasks/6g6jqqcdehne-...`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/planning/tasks/6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md#L85) states:
   > "The remaining aggregate `SummaryStore` `FileProblem` seam is explicitly transitional and remains owned by the sequenced portable entity-read design task."
   *Match:* Code strictly honors this. No premature redesign of `SummaryStore` was attempted, and no core code parses epic/audit paths.
2. **ADR amendment:** [`planning/adrs/0006-adopt-threads-as-task-dags.md:1140`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/planning/adrs/0006-adopt-threads-as-task-dags.md#L1140) states:
   > "Epic filename identity is recovered by the filesystem adapter before the aggregate summary boundary. Core does not parse locations, and board/status each retain one task scan."
   *Match:* Exact match with [`internal/store/epicstore.go:28`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/store/epicstore.go#L28) and scan counting tests.
3. **Architecture text:** [`docs/ARCHITECTURE.md:282`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/docs/ARCHITECTURE.md#L282) accurately records schema 1.75 carrying portable diagnostics through `board`, current `status`, and `status --all`.
4. **Schema changelog:** [`internal/wire/wire.go:311`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/wire/wire.go#L311) accurately documents the additive nature of schema 1.75.

---

### Mandatory evidence floor

#### 1. Consumer inventory

| Location & Symbol | Read/Write | Classification | Rationale |
|---|---|---|---|
| [`internal/core/service_task.go:105`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service_task.go#L105) `TaskGraphLoadProblem` | Definition | Portable semantic data & snapshot evidence | Retains neutral ID/slug, opaque Location, and adapter SourceVersion |
| [`internal/core/service_task.go:163`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service_task.go#L163) `TaskGraphLoadProblemFromFile` | Write | Deliberate local repair evidence | Maps local `FileProblem` to `TaskGraphLoadProblem` with `LocationIsPath=true` |
| [`internal/core/service_task.go:195`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service_task.go#L195) `taskGraphLoadProblems` | Read/Write | Presentation & projection mapping | Projects graph problems to `LintLoadProblem` preserving identity |
| [`internal/core/dependency_graph.go:358`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/dependency_graph.go#L358) `newTaskGraph` | Read | Guarded snapshot evidence | Canonical sorting and assignment to `hardBroken` map |
| [`internal/core/dependency_graph.go:973`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/dependency_graph.go#L973) `sameTaskGraphLoadProblem` | Read | Guarded snapshot evidence | Fail-closed CAS comparison including SourceVersion & Location |
| [`internal/core/board.go:48`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/board.go#L48) `Service.Board()` | Read/Write | Portable semantic data | Assigns `taskGraphLoadProblems` to `Board.Problems` |
| [`internal/core/service.go:434`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L434) `summarize` | Read/Write | Portable semantic data | Aggregates tasks, epics, and audits in deterministic order |
| [`internal/core/lint_source.go:62`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/lint_source.go#L62) `lintLoadProblemsFromFiles` | Read/Write | Transitional local compatibility | Copies adapter-recovered identity from `FileProblem` to `LintLoadProblem` |
| [`internal/core/space_overview.go:208`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/space_overview.go#L208) `cloneSpaceSummary` | Read/Write | Guarded snapshot evidence | Deep-clones slice to prevent mutable retention leakage |
| [`internal/store/epicstore.go:28`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/store/epicstore.go#L28) `FS.ListEpics` | Write | Adapter-level recovery | Recovers canonical epic identity from filename stem before core boundary |
| [`internal/wire/envelopes.go:955`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/wire/envelopes.go#L955) `ToLintLoadProblemsJSON` | Read/Write | Wire presentation mapping | Maps to schema 1.75 wire format, emitting `path` only when `LocationIsPath=true` |
| [`internal/cli/problems.go:76`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/cli/problems.go#L76) `portableProblemsError` | Read | CLI presentation mapping | Formats human error prioritizing identity over location |
| [`internal/tui/dashboard.go:208`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/tui/dashboard.go#L208) `setSummary` | Read | Presentation mapping | Sets TUI attention marker based on unreadable count |
| [`internal/tui/atlas.go:978`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/tui/atlas.go#L978) `atlasStats` | Read | Presentation mapping | Includes unreadable planning count in Atlas attention tally |

#### 2. Field-lineage table

| Field | Producer | Conversions | Consumers | Legal omission? |
|---|---|---|---|---|
| `EntityKind` | Adapter / `summarize` | `LintLoadProblem` → `LintLoadProblemJSON` | Wire JSON, CLI error formatting | No in `LintLoadProblem`; wire omits if empty |
| `EntityID` | Adapter (`splitFlatName`, `ListEpics`) | `FileProblem.EntityID` → `TaskGraphLoadProblem.TaskID` → `LintLoadProblem.EntityID` → `LintLoadProblemJSON.EntityID` | Core graph analysis, wire DTO, CLI error fallback | Yes (unidentified records) |
| `EntitySlug` | Adapter (`splitFlatName`) | `FileProblem.EntitySlug` → `TaskGraphLoadProblem.TaskSlug` → `LintLoadProblem.EntitySlug` → `LintLoadProblemJSON.EntitySlug` | Primary error name in CLI, wire DTO | Yes |
| `Location` | Adapter (storage URI / filepath) | Passed through `TaskGraphLoadProblem` → `LintLoadProblem` → `LintLoadProblemJSON` | Wire DTO, fallback in CLI error | Yes |
| `LocationIsPath` | Adapter / `TaskGraphLoadProblemFromFile` | `problem.Path != ""` → `LintLoadProblem.LocationIsPath` | `taskGraphLocalPath`, `ToLintLoadProblemsJSON`, `portableProblemsError` | No (bool default false) |
| `Path` (local) | Filesystem adapter | Retained in `TaskGraphLoadProblem.Path`; wire derives `Path` from `Location` if `LocationIsPath=true` | Graph problem path, wire compatibility | Yes in core; required string in wire JSON (can be `""`) |
| `Message` | Scanner / YAML parser | Preserved through all structs | Wire JSON, CLI stderr, TUI hints | No (always required) |
| `SourceVersion` | `scanDirWithSourceVersions` | Captured in `TaskGraphLoadProblem.SourceVersion`; unexported/ignored by wire | `sameTaskGraphLoadProblem` (CAS comparison) | Yes (missing version causes CAS failure) |

#### 3. Diagnostic input matrix

| Diagnostic Input Case | Core representation | Wire output (`ToSummaryJSON`) | CLI human / stderr output |
|---|---|---|---|
| Identity-only | `{Kind: task, ID: "6g01"}` | `{"entity_kind":"task","entity_id":"6g01","path":""}` | `validation failed: 1 unreadable task record(s): 6g01` |
| Pathless ID + Slug | `{Kind: task, ID: "6g01", Slug: "s1"}` | `{"entity_kind":"task","entity_id":"6g01","entity_slug":"s1","path":""}` | `validation failed: 1 unreadable task record(s): s1` |
| Opaque location only | `{Kind: task, Location: "db://row-1"}` | `{"entity_kind":"task","location":"db://row-1","path":""}` | `validation failed: 1 unreadable task record(s): db://row-1` |
| Identity + contradictory opaque location | `{Kind: task, ID: "6g01", Slug: "real", Location: "db://fake"}` | `{"entity_kind":"task","entity_id":"6g01","entity_slug":"real","location":"db://fake","path":""}` | `validation failed: 1 unreadable task record(s): real` (opaque location suppressed) |
| Local path + recovered identity | `{Kind: task, ID: "6g01", Slug: "task-a", Location: "tasks/task-a.md", LocationIsPath: true}` | `{"entity_kind":"task","entity_id":"6g01","entity_slug":"task-a","location":"tasks/task-a.md","path":"tasks/task-a.md"}` | `validation failed: 1 unreadable task record(s): task-a (task-a.md)` |
| Invalid local filename without identity | `{Kind: task, Location: "tasks/invalid", LocationIsPath: true}` | `{"entity_kind":"task","location":"tasks/invalid","path":"tasks/invalid"}` | `validation failed: 1 unreadable task record(s): invalid` |
| Empty location & empty identity | `{Kind: task, Message: "decode failed"}` | `{"entity_kind":"task","path":"","message":"decode failed"}` | `validation failed: 1 unreadable task record(s): unidentified task record` |
| Mixed task, epic, audit failures | Ordered slice: task `6g01`, epic `01-x`, audit `6ga01` | Array preserving task → epic → audit ordering with distinct `entity_kind` values | Stderr lists all three entities in order; renders summary before exiting 11 |

#### 4. Real filesystem binary execution & CLI policy

Tested against `./bin/tskflwctl` compiled inside `$SANDBOX`:

```sh
$ tskflwctl status
Tasks
 active    1 ready-to-start
⚠ task graph broken: unreadable task file: validation failed: malformed frontmatter: ... in tasks/6gtask000002-broken-task.md
(exit 11)
stderr: error: validation failed: 1 unreadable planning record(s): broken-task (6gtask000002-broken-task.md)

$ tskflwctl status --json
{"schema_version":"1.75","counts":[...],"in_progress":[],"epics":[],"unreadable":[{"entity_kind":"task","entity_id":"6gtask000002","entity_slug":"broken-task","location":".../tasks/6gtask000002-broken-task.md","path":".../tasks/6gtask000002-broken-task.md","message":"..."}]}
(exit 11)
stderr: {"schema_version":"1.75","error":{"code":"validation","message":"validation failed: 1 unreadable planning record(s): broken-task (6gtask000002-broken-task.md)"}}

$ tskflwctl board
next-up (0)
 (none)
ready-to-start (1)
 good-task    good
(exit 11)
stderr: error: validation failed: 1 unreadable task record(s): broken-task (6gtask000002-broken-task.md)

$ tskflwctl board --json
{"schema_version":"1.75","columns":[...],"unreadable":[{"entity_kind":"task","entity_id":"6gtask000002","entity_slug":"broken-task","location":".../tasks/6gtask000002-broken-task.md","path":".../tasks/6gtask000002-broken-task.md","message":"..."}]}
(exit 11)
stderr: {"schema_version":"1.75","error":{"code":"validation","message":"validation failed: 1 unreadable task record(s): broken-task (6gtask000002-broken-task.md)"}}
```

#### 5. Call counts & single-scan proof

- `TestBoard_CompleteStoreFallbackScansTasksOnce`: confirms exactly 1 task scan (`ListTasks calls = 1`).
- `TestBoard_PreservesPathlessTaskLoadProblemIdentity`: confirms exactly 1 task graph read (`source.calls == 1`).
- `TestService_Summary_ReadsEachAuditOnce`: pins H2 invariant (1 sweep of audit bodies).
- `TestService_Summary_PreservesMixedPortableLoadDiagnostics`: asserts `source.calls == 1 && store.listEpics == 1 && store.listWithFindings == 1`.

#### 6. Retained SpaceOverview & TUI Atlas attention

- `TestSpaceOverviewPreservesPathlessTaskLoadProblemIdentity`: confirms pathless identity is preserved through multi-space overview resolution.
- `TestSpaceOverviewRetainedSummaryOwnsMutableSnapshotData`: verifies retained contended summaries perform deep cloning on `Problems`, preventing mutable aliasing.
- `TestAtlasAttentionFoldsOnlyWhatWantsAPerson`: verifies that a pathless unreadable task (`Problems: []core.LintLoadProblem{{EntityKind: core.LintEntityTask, EntityID: "6gpathless001"}}`) increments the attention counter without requiring a filesystem path.
- `TestDashboardNeedsAttentionReportsPathlessUnreadableRecord`: verifies that pathless unreadable records trigger the "needs attention" marker and block "all clear" in the dashboard TUI.

#### 7. Schema 1.75 semantic & diff assessment

- `SchemaVersion` bumped from `"1.74"` to `"1.75"` monotonically.
- All golden files updated strictly with version bump from 1.74 to 1.75.
- `schema_jsonschema.golden`: `BoardEnvelope.unreadable` and `SummaryJSON.unreadable` updated from `FileProblem` to `LintLoadProblemJSON`.
- `LintLoadProblemJSON` maintains `path` and `message` as required fields; newly introduced fields `entity_kind`, `entity_id`, `entity_slug`, and `location` are optional strings.
- `TestJSONSchema_ValidatesRealOutput` in `internal/wire/envelopes_test.go` verifies non-default unreadable records against the generated JSON Schema for `BoardEnvelope`, `SummaryEnvelope`, and `StatusAllEnvelope`.

#### 8. Test independence audit

All new test assertions in `internal/core`, `internal/cli`, and `internal/wire`:
- Do not call production conversion helpers (`ToLintLoadProblemsJSON` or `taskGraphLoadProblems`) to synthesize expected comparison values;
- Explicitly assert populated non-zero string values for `EntityKind`, `EntityID`, `EntitySlug`, and `Location`;
- Independently assert that `Path` is empty when `LocationIsPath == false`.

#### 9. Verification & command results matrix

| Verification check | Command executed | Exit code | Result |
|---|---|---|---|
| Unit tests (all packages) | `go test ./...` | 0 | All packages passed |
| Static analysis | `golangci-lint run ./...` | 0 | 0 issues |
| Go module tidiness | `go mod tidy -diff` | 0 | Clean (no diff) |
| Git whitespace & conflict check | `git diff --check` | 0 | Clean (no issues) |
| Schema comment freshness | `go run ./internal/tools/schemacomments -out /tmp/sc.json && diff -u internal/wire/schema_comments.json /tmp/sc.json` | 0 | Clean (in sync) |
| Planning entity lint | `go run ./cmd/tskflwctl --no-color lint` | 0 | ✔ all entities pass |
| Audit findings lint | `go run ./cmd/tskflwctl audit lint` | 0 | ✔ all findings pass |
| Contract goldens | `go test ./internal/cli -run TestGolden` | 0 | All goldens match |
| JSON Schema validation | `go test ./internal/wire -run TestJSONSchema` | 0 | Output matches schema |

---

## Findings

#### L1. Service.Lint passes TaskGraphLoadProblem to graph analyzer without test assertion on Location fields · **Status:** fixed

In [`internal/core/service.go:520-525`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.CtUoSY/internal/core/service.go#L520-L525), `Service.Lint()` adapts `taskProblems` into `graphRead.Problems` for the local task graph:

```go
for _, problem := range taskProblems {
	graphRead.Problems = append(graphRead.Problems, TaskGraphLoadProblem{
		TaskID: problem.EntityID, TaskSlug: problem.EntitySlug,
		Location: problem.Location, LocationIsPath: problem.LocationIsPath,
		Message: problem.Message,
	})
}
```

During the coordinated-mutant tournament, dropping `Location` and `LocationIsPath` from this assignment (Mutant 12) survived all unit, integration, and golden tests.

Analysis reveals this is a test coverage gap caused by an architectural shortcut: `Service.Lint()` constructs `graph := NewTaskGraphRead(graphRead)` solely to derive `dependencyLintIssues(graph)`. In `dependencyLintIssues`:
```go
switch problem.Code {
case ProblemUnreadable, ProblemMissingTaskID, ...:
	continue
```
`ProblemUnreadable` is intentionally skipped to prevent duplicate reporting against ordinary task linting. The unreadable load problems returned by `Service.Lint()` come directly from `taskProblems` rather than from `graphRead.Problems` or `graph.Problems()`. Consequently, while `Service.Lint()` populates `Location` and `LocationIsPath` on `TaskGraphLoadProblem`, no test exercises this data path, and the fields are inert within the local graph analysis.

**Remedy for owner triage:** Add a focused test verifying that if `Service.Lint()` constructs a `TaskGraphRead` with unreadable records, the graph's internal diagnostics correctly retain the supplied location context, or document that `graphRead.Problems` in `Service.Lint()` is only used for `hardBroken` task ID tracking.

**Resolution:** Removed the inert location pass-through from Service.Lint's
private graph copy and documented its identity-only purpose; the original
portable diagnostic remains the sole user-facing source.
