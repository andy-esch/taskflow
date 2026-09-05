---
schema: 1
id: 6g73ymfrj67y
bucket: closed
area: thread-source-revision-cas-implementation-claude
date: "2026-09-05"
updated_at: "2026-09-05"
---
# Audit: Thread source revision CAS implementation — Claude — 2026-09-05

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer
> should update.
>
> Finding grammar is exact: use `#### H1. <title> · **Status:** open` (or M1/L1). Codes must match
> `[A-Z]+[0-9]+`; do not put status on a separate line or pre-resolve a finding.
>
> A second adversarial pass is mandatory. After the checklist, deliberately seek a systemic reason
> the locally symmetric patch is still unsafe: divergent read paths, a token-bearing contract that
> leaks through another adapter, a comparison that is correct only for filesystem order, or a
> guarded writer that silently retains the former `FileProblem` snapshot. A no-finding verdict must
> show the hostile evidence that falsified those hypotheses; green tests alone are insufficient.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may use the handoff checkout concurrently. Reading
this brief and performing the initial copy are the only operations allowed there until the final
guarded audit transfer. Capture the exact current contents, including staged, unstaged, untracked,
and deleted files:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6g73ymfrj67y-2026-09-05-thread-source-revision-cas-implementation-claude.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review-claude.XXXXXX")"

git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"
rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"
test -d "$SANDBOX/.git"
cd "$SANDBOX"

git add -A
git -c user.name='Taskflow Review Sandbox' \
  -c user.email='review-sandbox@invalid' \
  -c commit.gpgsign=false \
  -c core.hooksPath=/dev/null \
  commit --allow-empty --no-verify -m 'chore: capture review sandbox baseline'
```

Confirm `git rev-parse --absolute-git-dir` resolves inside `$SANDBOX` and no object alternate points
to the source. Perform all inspection, builds, formatting, fixtures, and mutation probes there.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If isolation cannot be verified, stop and report the blocker.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`; inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then transfer
only the audit, guarded against concurrent source edits:

```sh
test "$(git -C "$SOURCE_ROOT" hash-object "$SOURCE_AUDIT")" = "$SOURCE_AUDIT_BLOB" || {
  printf 'source audit changed; do not overwrite it; preserve sandbox at %s\n' "$SANDBOX" >&2
  exit 1
}
TRANSFER="$(mktemp "${SOURCE_AUDIT}.review-transfer.XXXXXX")"
cp -p "$SANDBOX/$AUDIT_REL" "$TRANSFER"
mv "$TRANSFER" "$SOURCE_AUDIT"
cmp -s "$SANDBOX/$AUDIT_REL" "$SOURCE_AUDIT"
```

Do not copy anything else back. Leave the sandbox in place and report its path until the
implementation owner confirms receipt.

## Review brief

Perform an independent adversarial implementation review of
[version unreadable Thread sources for repair-safe CAS](../tasks/6g72ncs4xjdm-version-unreadable-thread-sources-for-repair-safe-cas.md)
on branch `feat/unreadable-thread-source-revisions`, based on `main` at `aababec`. The implementation
under review is the complete uncommitted working tree captured by the sandbox checkpoint. Review its
diff against `aababec` and judge it against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md),
[the architecture guide](../../docs/ARCHITECTURE.md), the task acceptance criteria, the immediately
preceding task-source revision implementation, and every existing Thread mutation/OCC contract.

Assume the apparent task/Thread symmetry can conceal an asymmetry. Re-derive exact source bytes
through both filesystem Thread scanners, neutral `ThreadRead`, service projections,
`sameThreadSourceSnapshot`, `verifyThreadSourceSnapshot`, and every pre-write consumer. Do not edit
implementation or other planning files, create follow-up tasks, change finding statuses, close this
audit, push, or install anything globally.

## Intended contract to challenge

- `ThreadReadProblem.SourceVersion` is an opaque adapter-owned revision of the exact unreadable
  source. Core and guarded stores may copy, normalize, and compare it but may not parse it or expose
  it through Thread/domain values, validation errors, CLI, TUI, schema, JSON/YAML, or wire contracts.
- Production filesystem `ReadThreads` hashes the same bytes it hands `parseThread`, in one resilient
  scan. Ordinary `ListThreads` retains its established `[]domain.FileProblem` compatibility shape;
  tokenless compatibility reads are CAS-ineligible rather than falsely authoritative.
- Thread apply's body-aware scan carries the identical revision semantics. Creation, membership and
  lifecycle mutation, bulk apply, and task lifecycle all compare the complete revision-bearing
  `ThreadRead` snapshot before any write.
- Snapshot comparison is order-independent and does not mutate caller slices. It compares readable
  source identity/location plus a non-empty equal revision, and unreadable stable identity, optional
  location, message, plus a non-empty equal revision.
- Missing revision evidence fails closed. Pathless/remote adapters may provide stable opaque tokens
  without fabricating filesystem paths. Add/remove/rename/identity drift, readable/unreadable
  transitions, and changed bytes reproducing one diagnostic all invalidate equality; exact byte
  restoration restores the filesystem revision.
- Existing ordinary mutations still refuse malformed Thread input before planning. The stronger
  comparison protects transitions at their verify boundary and is reusable by graph repair, where
  malformed Threads will be non-blocking but impact evidence may be incomplete.
- The change adds no second scan, persisted token, universal entity revision scheme, repair policy,
  or graph-library dependency.

## Mandatory evidence floor

A `ready` verdict is not credible without all of the following:

1. Inventory `ReadThreads`, `ListThreads`, their test override, `threadReadFromFiles`, the
   revision-bearing conversion, `scanDirWithSourceVersions`, `listThreadApplyThreads`, every
   `ThreadStore` production composition, every service consumer, and all four guarded write families.
2. Exercise unchanged unreadable bytes; changed body bytes with the same diagnostic; exact restore;
   add/remove/rename; recovered identity drift; readable/unreadable transitions; pathless problems;
   reordered and duplicate problem identities; multiple readable records; and both one-sided and
   two-sided missing tokens.
3. Prove ordinary native listing, Thread service results, lint, human CLI, JSON CLI, TUI state,
   generated schema, and wire DTOs cannot reveal the token or change established diagnostics.
4. Verify the body-aware Thread apply scan did not lose body fidelity, change scan counts, or diverge
   from ordinary `ReadThreads` revision semantics.
5. Trace the whole-snapshot and immediate target CAS windows. State exactly what cooperating writers,
   raw editors, unreadable non-targets, and the future repair path can and cannot guarantee.
6. Apply and restore mutation probes that: omit unreadable revision comparison; accept two empty
   tokens; derive the token from path/message; return an empty filesystem token; route one writer
   back through `ListThreads`; make equality input-order-sensitive; expose the token from the service
   or wire; and drop unreadable problems. Name the shipped test killing each mutation and report any
   survivor as a finding.
7. Run repeated focused race tests, an uncached full race suite, static analysis, module tidiness,
   generated-doc checks, planning/audit lint, and `git diff --check`, with exact results.

## Required adversarial angles

1. **Source truth.** Challenge whether the hash covers the exact parser bytes, whether a read error
   stays fatal, and whether parse failures remain resilient without another filesystem pass.
2. **Portability and disclosure.** Challenge pathless adapters, tokenless fakes, direct struct
   serialization, service cloning, TUI retention, explicit wire mapping, schemas, and error strings.
3. **Comparator soundness.** Seek false equality, false conflicts, asymmetric non-empty checks,
   unstable duplicate pairing, aliasing, location-as-identity inference, or ordering-dependent CAS.
4. **Consumer completeness.** Inspect every guarded operation independently. Do not infer that fixing
   creation fixed mutation, task lifecycle, body-aware apply, or apply reprepare.
5. **Compatibility.** Compare healthy and malformed Thread list/lint/show/apply behavior to base;
   challenge the `threadListOverride` seam and planner re-entry protections.
6. **Systemic second pass.** Look beyond this patch for a shared resilient-read/CAS convention that
   leaves one entity or mutation path half-versioned. Demonstrate it, distinguish current reachability
   from forward risk, and avoid speculative architecture expansion.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- branch/base/checkpoint, runtime, exact validation, and isolation attestation;
- a compact source-to-CAS data flow and complete producer/consumer inventory;
- findings grouped by severity with exact grammar and `**Status:** open`;
- acceptance-criteria traceability plus hostile-evidence and restored-mutation ledgers;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but still provide every required evidence class. Do not
pre-resolve findings; the implementation owner will triage them with `tskflwctl audit finding`.

## Reviewer report

### Executive verdict

**ready with tracked follow-ups.**

The implementation is correct, fails closed, and discloses nothing. Every disclosure and comparator
hypothesis I raised was falsified with hostile evidence, not with green tests. The follow-ups are
evidence defects, not runtime defects: the unreadable-revision comparison that this task exists to
add is **unreachable through all four guarded writers today**, one acceptance criterion is ticked
without support, and a mutation that routes a guarded writer back onto the tokenless read path
survives the entire suite. No behavior regression, no leak, and no unsafe write was found.

### Attestation

| Field | Value |
| --- | --- |
| Branch | `feat/unreadable-thread-source-revisions` |
| Base | `aababec` (Merge pull request #192) |
| Sandbox checkpoint | `7448fe0` `chore: capture review sandbox baseline` |
| Sandbox | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review-claude.StWFAf` |
| Isolation | `git rev-parse --absolute-git-dir` → sandbox `.git`; **no** `objects/info/alternates` (clone `--no-hardlinks`, then `rsync -a --delete` of the live tree) |
| Source writes | none until the guarded transfer; source audit blob `fd3a2da4…` unchanged at transfer time |
| Runtime | go1.26.6 darwin/arm64 · golangci-lint 2.12.2 |

All inspection, builds, fixtures, mutation probes, and the base-vs-head binary comparison ran in the
sandbox. Every probe was restored from a byte-compared backup; `git status --short` in the sandbox is
empty against the checkpoint apart from this audit.

### Validation actually run

| Check | Result |
| --- | --- |
| `go build ./...` · `go vet ./...` | clean |
| `go test ./...` | all packages ok |
| `go test -count=1 -race ./internal/{store,core,tui,wire}/...` | ok (uncached) |
| `go test -count=5 -race ./internal/store -run Thread` | ok (repeated focused race) |
| `gofmt -l internal cmd` | empty |
| `go mod tidy` | no diff to `go.mod`/`go.sum` |
| `golangci-lint run ./...` | 0 issues |
| `docgen -out docs/cli` then `git diff --exit-code docs/cli` | clean, no drift |
| `tskflwctl lint` (sandbox planning) | all entities pass |
| `tskflwctl audit lint 6g73ymfrj67y` | all findings pass |
| `git diff --check aababec HEAD` | clean |

### Source-to-CAS data flow

```
threads/<id>-<slug>.md  ──┐
                          │  os.ReadFile (ONE pass, one os.ReadDir)
                          ▼
        scanDirSources(dir, parse, versionProblems=true)     resolve.go:32
           sourceVersion = hashContent(content)   ← EXACT bytes, hashed BEFORE parse
           parse(path, content) ──► ok ─► T                  (readable)
                                └─► err ─► sourceFileProblem{FileProblem, sourceVersion}
                          │
        ┌─────────────────┴──────────────────┐
        ▼                                    ▼
  FS.ReadThreads()                    FS.listThreadApplyThreads()
  threadstore.go:23                   threadapply.go:226  (+ body via splitFrontmatter)
        │                                    │
        └────────► threadReadFromSourceFiles(threads, problems) ─► core.ThreadRead
                          │                                          Threads[].SourceVersion (parseThread)
                          │                                          Problems[].SourceVersion (scan)
        ┌─────────────────┼──────────────────┬────────────────────────┐
        ▼                 ▼                  ▼                        ▼
  MutateThreadCreation  MutateThread  MutateThreadApply     MutateTaskLifecycle
        └─────────────── verifyThreadSourceSnapshot(expected, current) ──► domain.ErrConflict
                         cas.go:47 → sameThreadSourceSnapshot (order-independent, fail-closed)

  Read-only projections (token stripped / never mapped):
    Service.ListThreadViews  → problems[i].SourceVersion = ""   service_thread.go:224
    Service.Lint             → domain.FileProblem{Path,Message} service.go:456
    ComposeThreadApply       → validator reads name+message only
    wire.ToThreadsEnvelope   → explicit 4-field ThreadReadProblemJSON
    core.ThreadReadProblem.SourceVersion  `json:"-" yaml:"-"`
```

### Producer / consumer inventory

**Producers.** `scanDirSources` (`resolve.go:32`) is the single revision source; `scanDirWithSourceVersions`
(`:78`) and `scanDir` (`:87`) are its two projections. `parseThread` (`threadstore.go:131`) sets the
readable `domain.Thread.SourceVersion`. Two revision-bearing scans exist for Threads:
`FS.ReadThreads` (`threadstore.go:23`) and `FS.listThreadApplyThreads` (`threadapply.go:226`);
both terminate in `threadReadFromSourceFiles` (`threadstore.go:66`).

**Conversions.** `threadReadProblemFromFile` (`threadstore.go:84`) is the one place filename identity is
recovered; `threadReadFromFiles`/`threadReadProblemsFromFiles` (`:58`,`:77`) are its tokenless
counterparts, now reachable **only** through the `threadListOverride` test seam.

**Guarded write families (all four verified independently, not inferred from one another).**

| Writer | Prepare read | Pre-write re-read | CAS |
| --- | --- | --- | --- |
| `MutateThreadCreation` `threadcreation.go:50/83` | `ReadThreads` | `ReadThreads` | `:87` |
| `MutateThread` `threadmutation.go:50/88` | `ReadThreads` | `ReadThreads` | `:92` |
| `MutateThreadApply` `threadapply.go:53/119` | `listThreadApplyThreads` | `listThreadApplyThreads` | `:123` |
| `MutateTaskLifecycle` `lifecyclemutation.go:54/106` | `ReadThreads` | `ReadThreads` | `:110` |

`MutateThreadApply.reprepareThreadApply` (`:250`) re-reads and re-validates but performs no snapshot
CAS — correct, because its purpose is convergence after a durable prefix, and each remaining write is
still gated by per-file `verifyUnchanged`.

**Service consumers.** `Service.ListThreadViews` (`service_thread.go:214`), `Service.Lint`
(`service.go:452`), `Service.ComposeThreadApply` (`service_thread_apply.go:19`). **Wire:**
`ToThreadsEnvelope` (`wire/thread.go:145`). **TUI:** `thread_projection.go:26` via `ListThreadViews`
only — the TUI never touches `ThreadStore`.

### Findings

#### H1. Acceptance criterion 5 is ticked but no shipped test proves it, and the path it claims to cover is unreachable · **Status:** tracked by 6g4g8gatbnrs

AC 5 of `6g72ncs4xjdm` reads: *"Focused store tests prove graph-aware write guards return `ErrConflict`
before any write when an unreadable Thread source changes concurrently."* It is ticked `[x]`. No such
test exists, and today none can pass, because the unreadable-revision comparison is unreachable from
every guarded writer.

Proof (instrumentation probe M10, restored): I recorded a stack trace at every entry to
`sameThreadProblemSource` and ran `go test -count=1 ./internal/store ./internal/core ./internal/tui`.
Ten invocations, from exactly three call sites — all three the *new unit tests* in
`thread_source_snapshot_test.go`:

```
store.TestReadThreadsVersionsExactUnreadableSourceBytes
store.TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed
store.TestThreadSourceSnapshotRejectsRepresentationAndIdentityChanges
```

`grep 'store\.\(\*FS\)\.Mutate[A-Za-z]*'` over the same log returns **nothing**: no guarded writer ever
reaches the unreadable comparison.

The mechanism: all four writers call `ValidateThreadCreationSource`/`ValidateThreadMutationSource`
(`thread_creation.go:62`, `thread_mutation.go:196`), which hard-refuse with `ErrValidation` whenever
`len(unreadable) > 0` at prepare time. So a writer reaching the CAS always has **zero** problems on the
expected side. If a document becomes unreadable in the window, `sameThreadSourceSnapshot`'s
`len(left.Problems) != len(right.Problems)` check (`cas.go:56`) short-circuits, and `slices.EqualFunc`
never invokes the element comparator on two empty slices. The revision on an unreadable Thread is
therefore never compared by any shipped writer.

The nearest existing test, `TestGuardedMutationsRejectUnreadableThreadDocuments/Thread apply final
snapshot verification`, does return `ErrConflict` before any write — but it covers a Thread
*becoming* unreadable (count 0→1), not a source *changing*. `git diff aababec HEAD --stat --
internal/store/thread_unreadable_guard_test.go` is empty (the file is untouched by this patch) and it
**passes unmodified at base `aababec`**, so it cannot be evidence for a criterion about source
revisions.

This is an evidence defect, not a runtime defect: the shipped code is safe and fails closed. But the
criterion should be untucked and given a state (`deferred`/`tracked` with the repair task) rather than
left ticked, since `complete` treats a ticked box as a discharged obligation.

**Failure scenario.** A reader trusts AC 5, or a later change relaxes the validation gate for repair
believing writer-level regression tests exist. The first mutation that legitimately proceeds past a
malformed Thread has its unreadable-revision CAS exercised for the very first time in production, with
no test having ever executed that path.

**Resolution:** Reworded this task's fifth criterion to the verifier behavior
its focused tests actually prove, then added a tracked criterion assigning the
first reachable malformed-Thread writer proof to repair task 6g4g8gatbnrs; that
task's concurrency criterion now explicitly covers still-unreadable Thread byte
changes.

#### M1. No test pins any guarded writer to the revision-bearing read; routing one back to `ListThreads` is silently green · **Status:** fixed

Mutation probe M5 (applied and restored): I rewrote `MutateThreadCreation` (`threadcreation.go:50`
and `:83`) to read through the tokenless compatibility surface —

```go
legacyThreads, legacyProblems, err := s.ListThreads()
threadRead := threadReadFromFiles(legacyThreads, legacyProblems)
```

— on both the prepare read and the pre-write re-read. **The entire suite stayed green**
(`internal/store`, `internal/core`, `internal/wire`, `internal/tui` all `ok`). This was the only
survivor of nine probes.

It survives for the H1 reason — `threadReadFromFiles` still yields token-bearing *readable* threads
(`parseThread` sets `SourceVersion` regardless of scan flavor); only the *unreadable* tokens are lost,
and those are unreachable. So the regression is invisible today and would surface only once the repair
path lifts the validation gate — exactly when it matters most.

The intended contract states all four guarded write families compare the revision-bearing snapshot.
That is true as written, but nothing holds it there. A store-level test asserting that each guarded
writer's snapshot carries non-empty `Problems[].SourceVersion` when a malformed Thread is present (or a
seam that makes `ListThreads` unavailable to writers) would pin it.

**Failure scenario.** A future refactor consolidates Thread reads back onto `ListThreads` for
convenience. CI is green. When guarded repair later authorizes a write while a malformed Thread is
present, the snapshot's unreadable entries carry empty revisions; a raw editor's byte change behind an
unchanged diagnostic is invisible to the length-only check, and repair authorizes a write from stale
evidence — the precise failure this task was created to prevent.

**Resolution:** Removed the unused tokenless ListThreads route, its conversion
helpers, and its test override. General reads use revision-bearing ReadThreads,
while bulk apply retains its specialized revision-bearing body-aware read, so
guarded writers cannot select the reviewed alternative.

#### L1. `ListThreads` has no production callers left and survives only as a test seam · **Status:** fixed

`grep -rn '\.ListThreads()' internal cmd` returns five hits, **all in `_test.go` files**; the only
non-test occurrences are the method definition and its doc comment (`threadstore.go:43-46`). Its doc
comment says it is "retained for local maintenance and tests", but no maintenance path uses it. With
M1, an unused tokenless read that is API-compatible with the token-bearing one is an attractive
nuisance. Consider unexporting it, folding it into the test seam, or documenting that guarded writers
must never call it.

**Failure scenario.** A contributor picks the simpler-looking `ListThreads` for a new guarded Thread
mutation, gets tokenless problems and a green suite, and ships a half-versioned writer.

**Resolution:** Removed the production-dead ListThreads method and its tokenless
helper graph; internal/store has no parallel Thread listing surface remaining.

#### L2. `TestFSReadThreadsPerformsExactlyOneNativeThreadScan` no longer pins the production scan · **Status:** fixed

At base, `ReadThreads` delegated to `readThreads(s.ListThreads)`, so the override counter genuinely
proved `ReadThreads` performed exactly one native scan. `ReadThreads` now checks `threadListOverride`
directly (`threadstore.go:27`) and returns before touching `scanDirWithSourceVersions`, so the test
(`paths_test.go:100`) now asserts only that the override branch calls the override once. The production
single-scan property is true by inspection (one `os.ReadDir`, one `os.ReadFile` per entry in
`scanDirSources`) but is no longer covered.

**Failure scenario.** Someone adds a second scan or a stat pass to the production `ReadThreads` body —
for instance to recover mtimes — and the test named for exactly this property still passes.

**Resolution:** Removed the obsolete override and the misleading scan-count
test. ReadThreads now directly contains one revision-bearing scan with no
compatibility branch; exact-byte tests exercise the production path.

#### L3. Every readable Thread and task document is SHA-256 hashed twice per scan · **Status:** wontfix

`scanDirSources` hashes `content` for **every** entry when `versionProblems` is true (`resolve.go:58`),
but the result is used only on the parse-failure branch (`:63`). For readable documents `parseThread`
(`threadstore.go:154`) — and `parseTask` (`fsstore.go:340`) — hash the same bytes again. Guarded
mutations call `ReadThreads`/`listThreadApplyThreads` twice (prepare + pre-write), so a healthy
repository pays four SHA-256 passes per Thread document per mutation, plus the task-graph equivalent.

The pattern is inherited from the task-source slice rather than introduced here, and correctness is
unaffected; hoisting the hash into the failure branch (or reusing it in the parse closure) removes it.

**Failure scenario.** On a large planning repository the redundant hashing measurably slows every
guarded mutation and every `lint`/`thread list`, with no behavioral benefit.

**Resolution:** Retained eager pre-parse hashing as the defensive exact-byte
contract. The prerequisite audit measured the same shared scanner at roughly
five percent overhead for 2000 8 KB files, still one filesystem pass and linear
work; no production evidence justifies weakening that invariant.

### Acceptance-criteria traceability

| AC | Verdict | Evidence |
| --- | --- | --- |
| 1. Byte-identical unreadable sources equal; changed bytes with same diagnostic differ | **met** | `TestReadThreadsVersionsExactUnreadableSourceBytes` asserts the diagnostic is identical across both writes and the revision differs; probes M3/M4 (token from path / empty token) both killed by it |
| 2. Missing revisions fail closed; pathless adapters supply revisions; order-independent | **met** | `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed`; killed M2 (non-empty checks removed) and M6 (sorts removed). My probes add one-sided missing tokens for both readable and unreadable sides |
| 3. Readable/unreadable transitions + add/remove/rename/identity drift invalidate | **met** | `TestThreadSourceSnapshotRejectsRepresentationAndIdentityChanges` covers all six; killed by M1-probe and M8 |
| 4. Projections stay compatible and cannot expose the token | **met** | M7a and M7b both killed; base-vs-head CLI byte-comparison identical on 6 surfaces; 64-hex sweep over 10 surfaces clean |
| 5. Focused store tests prove write guards return `ErrConflict` when an unreadable Thread source changes concurrently | **not met** | **H1** — no writer-level test; path proven unreachable by probe M10 |

### Hostile-evidence ledger

Each row is a hypothesis I tried to confirm and could not; green tests were never accepted as the
answer.

| Hypothesis | Method | Outcome |
| --- | --- | --- |
| Hash covers something other than the exact parser bytes | Read `resolve.go:49-61`: `hashContent(content)` runs on the same slice later handed to `parse`, before any parser sees it | **falsified** |
| A read error was softened to a problem | `os.ReadFile` failure still returns fatally (`resolve.go:50`); only `parse` failures become problems | **falsified** |
| A second filesystem pass was added | `scanDirSources` performs one `os.ReadDir` and one `os.ReadFile` per entry; apply closure byte-identical to base (`git show aababec:…` diffed) | **falsified** |
| Apply lost body fidelity | Base and head closures are character-identical (`parseThread` + `splitFrontmatter`); only the scan function name changed | **falsified** |
| Token leaks via service | Probe M7a removed the strip → `TestServiceThreadListStripsOpaqueProblemSourceRevisions` failed | **falsified** |
| Token leaks via wire | Probe M7b added `source_version` to `ThreadReadProblemJSON` → `TestToThreadsEnvelopeRetainsPathlessIdentityWithoutParsingLocation` failed | **falsified** |
| Token leaks via validation error strings | `ValidateThreadCreationSource` formats `threadReadProblemName(problem)` + `problem.Message` only; no `%+v` of the struct | **falsified** |
| Token leaks via any real CLI surface | Ran head binary over a fixture with a healthy Thread, a malformed Thread, and a non-id-led file; grepped 10 surfaces (`thread list`, `--json`, `show`, `frontier`, `lint`, `lint --json`, `task list --json`, `epic list --json`, `schema --json-schema`, `schema thread`) for `\b[0-9a-f]{64}\b` and for `source_version`/`SourceVersion` | **falsified** — zero hits |
| Diagnostics or exit codes changed vs base | Built base binary from `git archive aababec`; base vs head byte-identical on stdout, stderr, and exit code for 6 command shapes over the malformed fixture (exit 1, 1, 11, 1, 1, 1) and on the healthy fixture (only the fixture path differs) | **falsified** |
| Comparison is correct only for filesystem order | Probe M6 removed all four `sort.SliceStable` calls → `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` failed | **falsified** |
| Duplicate problem identities pair unstably | My probe: `[P(id1,v1), P(id1,v1)]` vs itself equal; `[P(id1,v1), P(id1,v1)]` vs `[P(id1,v1), P(id1,v2)]` conflicts | **falsified** — the sort key includes `SourceVersion` and the element comparator checks exactly the same five fields, so the comparison is true multiset equality |
| Two readable Threads swapping revisions compare equal | My probe: `[A(vA), B(vB)]` vs `[A(vB), B(vA)]` → `ErrConflict` | **falsified** |
| One-sided missing token compares equal | My probe, both readable and unreadable sides → `ErrConflict` | **falsified** |
| Comparison mutates or reorders caller slices | My probe with three readable records verifies caller order after comparison; shipped test does the same for problems | **falsified** — both sides are copied before sorting |
| Location is inferred as identity | `threadReadProblemFromFile` recovers identity from the filename at the adapter boundary only; core compares `Location` as an opaque field and never parses it | **falsified** |
| A guarded writer silently retains the former `FileProblem` snapshot | All four writers converted; the old `sameThreadSourceSnapshot([]domain.Thread, []domain.FileProblem, …)` is deleted; `threadReadFromFiles` is reachable only via the test seam | **falsified** |
| Another graph-aware writer is left half-versioned | `MutateTaskGraph` (`graphmutation.go:87`) verifies the task graph but not Threads — inspected: it writes only task dependency frontmatter, reads no Thread document, and returns no Thread impacts, so Thread evidence is not part of its authority | **falsified — correct by design** |
| The shared convention leaves another entity half-versioned | Only two multi-record CAS helpers exist (`verifyTaskGraphSourceSnapshot`, `verifyThreadSourceSnapshot`); epics/audits/research use `scanDir` and are written one file at a time under `verifyUnchanged` per-file CAS | **falsified** |
| The unreadable comparison is reachable from a writer | Probe M10 stack-trace instrumentation over the full store/core/tui suite | **CONFIRMED unreachable → H1, M1** |

### Restored-mutation ledger

Every probe was applied in the sandbox, run, then restored from a byte-compared backup;
`git status --short` was empty after each.

| # | Mutation | Killing test |
| --- | --- | --- |
| M1 | Drop `SourceVersion` from `sameThreadProblemSource` | `TestReadThreadsVersionsExactUnreadableSourceBytes`, `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` |
| M2 | Accept two empty tokens (remove both `!= ""` guards) | `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` |
| M3 | Derive token from `path` instead of bytes | `TestReadThreadsVersionsExactUnreadableSourceBytes`, `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` |
| M4 | Filesystem returns an empty token | `TestReadThreadsVersionsExactUnreadableSourceBytes`, `TestUnreadableTaskSourceRevisionDefeatsStaleGraphPrewriteCAS` |
| M5 | Route `MutateThreadCreation` (both reads) back through `ListThreads` | **SURVIVOR → M1** |
| M6 | Make equality input-order-sensitive (remove all four sorts) | `TestThreadSourceSnapshotNormalizesOpaqueProblemsAndFailsClosed` |
| M7a | Service stops stripping the token | `TestServiceThreadListStripsOpaqueProblemSourceRevisions` |
| M7b | Wire publishes the token as `source_version` | `TestToThreadsEnvelopeRetainsPathlessIdentityWithoutParsingLocation` |
| M8 | Revision-bearing scan drops unreadable problems | `TestReadThreadsRecoversFilesystemProblemIdentityAtAdapterBoundary`, `TestReadThreadsVersionsExactUnreadableSourceBytes`, `TestGuardedMutationsRejectUnreadableThreadDocuments` (all 5 subtests) |
| M10 | Instrumentation only: record stack at every unreadable comparison | n/a — diagnostic; produced the H1 evidence |

### CAS windows — what is and is not guaranteed

**Whole-snapshot window** (prepare read → pre-write verify, under `checkedWriteLock`). Guarantees that
between planning and the first durable byte, no Thread document was added, removed, renamed, changed
identity, changed bytes, or crossed the readable/unreadable boundary, **and** that the task graph is
unchanged. Cooperating writers are excluded by the advisory lock; the snapshot exists to catch a raw
editor that ignored it.

**Immediate-target window** (`verifyUnchanged`, per file, `cas.go:100+`). Bounds each individual atomic
replacement after a durable prefix has landed, re-resolving by canonical slug and re-hashing that one
file.

- **Cooperating writers**: fully excluded — they take the same lock.
- **Raw editors**: caught for any change to a *readable* Thread (byte hash) and for any change to the
  *set* of unreadable Threads (length + identity). A byte change to an *already-unreadable* Thread is
  detectable by the comparator but, per H1, cannot arise in a current writer because validation
  refuses first.
- **Unreadable non-targets**: today they abort the mutation outright with `ErrValidation`. They never
  become a silent partial authority.
- **Future repair path**: will proceed past malformed Threads, so it — and only it — will exercise the
  unreadable revision comparison. The comparator is ready and sound (hostile ledger above); the
  *coverage* is not (H1, M1).
- **Not guaranteed anywhere**: multi-file atomicity. `MutateThreadApply` and `MutateTaskGraph` are
  deliberately resumable, not transactional; a durable prefix can survive a later failure, which the
  receipt names. That is ADR-0006's stated contract, not a defect.

### Separation of concerns

**Demonstrated defects** (reproduced in the sandbox): H1 (AC 5 unsupported; unreachability proven by
M10), M1 (M5 survivor), L1 (`ListThreads` caller inventory), L2 (seam test scope, diffed against base),
L3 (double hash, read from source).

**Source-supported risks** (real but not currently reachable): the unreadable-revision comparison
becomes live the moment guarded repair lands. Its correctness is well evidenced; its integration with a
writer is not.

**Unverified concerns** (stated, not claimed as defects): I did not exercise a genuine pathless or
remote `ThreadStore` in production — no such adapter exists yet — so "pathless adapters may supply
opaque tokens without fabricating paths" is verified only at the comparator level, where
`Location: ""` problems compare correctly by identity + revision. I also did not attempt an adversarial
SHA-256 collision; the `hashContent` doc comment already reasons about that choice.

### Settled concerns

- **`domain.Thread.SourceVersion` carries `yaml:"-"` but no `json:"-"`** (`domain/thread.go:54`, and the
  same on `domain/task.go:11`). I checked whether this patch newly exposes it: it does not. No
  production path marshals `domain.Thread` directly — `wire` maps `ThreadViewJSON` field by field — and
  the 64-hex sweep over every JSON surface came back empty. The *new* field,
  `core.ThreadReadProblem.SourceVersion`, is tagged `json:"-" yaml:"-"`, i.e. strictly stricter than the
  pre-existing domain fields.
- **`ThreadRead.Threads` shares backing arrays for `Tags`/`Tasks`** after
  `append([]domain.Thread(nil), …)`. The CAS compares only scalar fields, and planners receive
  `clonePlannerThreads`. Not a defect.
- **`threadReadProblemLess` ignores `SourceVersion`.** It orders diagnostics for display and for
  choosing which problem to name in a validation error; ties are already resolved identically. It is not
  the CAS ordering (that is `threadProblemSourceKey`, which does include the revision).
- **`\x00`-joined sort keys.** Only sort order depends on the join; equality is re-checked field by
  field by the element comparator, so a hypothetical separator collision cannot produce false equality.
- **Planner re-entry protections** unchanged: `ReadThreads`, `ListThreads`, and `listThreadApplyThreads`
  all call `rejectRepositoryPlannerCall` first, including on the override branch.
- **ADR-0006 and `docs/ARCHITECTURE.md`** were updated accurately. The ADR explicitly retains "Ordinary
  mutations still reject malformed Thread reads before planning; the complete comparison exists for
  their pre-write transitions and for repair" — that is a correct statement of what shipped, and it does
  not overclaim what H1 identifies.
- **Scope discipline**: no second scan, no persisted token, no universal entity revision scheme, no
  repair policy, and no graph-library dependency were introduced. `go mod tidy` is a no-op.
