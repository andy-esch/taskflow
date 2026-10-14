---
schema: 1
id: 6gkabd0b0bg6
bucket: closed
area: documented-import-map-implementation-claude
date: "2026-10-13"
updated_at: "2026-10-13"
---
# Audit: Executable production import-map verification — claude — 2026-10-13

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief's bounded evidence floor rather than exhaustively mutating every new test.

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
SOURCE_ROOT="/Users/andyeschbacher/git/andy-esch/taskflow"
AUDIT_REL="planning/audits/6gkabd0b0bg6-2026-10-13-documented-import-map-implementation-claude.md"
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

Does the new import-map guard genuinely detect drift in its declared production scope,
or can parsing, discovery, projection or comparison quietly make wrong facts look equal?
This is a test/documentation change, not a runtime graph engine or a new import policy.

Claude leads Go discovery and coverage fidelity. Antigravity leads parser/comparator
false passes and test adequacy. Both inspect the whole shared contract and perform one
bounded cross-check of the other lens. Do not read the sibling report until your own
independent report is final. Do not assign a finding quota.

## Review target

Source checkout: /Users/andyeschbacher/git/andy-esch/taskflow
Source branch: main
Base: 86e255be59f7389dde526d139bfc9ef42f0f5582

The implementation is uncommitted, including two untracked Go test files. HEAD alone
contains none of this checker. Capture staged, unstaged and untracked state through
the mandatory independent-clone helper before inspection or execution.

Read the implementing task and Thread:

- planning/tasks/6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md
- planning/threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md

Verify a consumer inventory rather than assuming it from this list:

- internal/tools/architecturecheck/imports_test.go: TestArchitectureImportMap,
  parseImportMap, productionImportGraph, goList, listedPackage,
  decodeProductionImports, internalImportName and importGraphDiff.
- internal/tools/architecturecheck/imports_format_test.go: parser, projection, error,
  real-module and deterministic-diff regressions. Verify each cited test name below.
- docs/architecture/boundaries.md: marked map, coverage/grammar, policy/remediation.
- docs/ARCHITECTURE.md: index claim and link to the new guard.
- internal/tools/architecturecheck/architecture_test.go: existing navigation checks;
  those checks are not proof that the import map is semantically correct.
- Justfile and .github/workflows/ci.yml: ordinary race/coverage test entry points.
- .golangci.yml: independently enforced import rules; unchanged by this work.

Out-of-scope work is already tracked: selective package documentation
(6g9cz5sk3b0d), executable examples (6g9cz5sv6kck), broader reconciliation
(6g9cz5t37dc0), the proposed dependency-policy ADR (6gg7e59mm68g), and the
pure-layer I/O/framework guard (6ggxymjbf86r). Verify destinations before citing them.
The ADR is not accepted or delivered here. A gap in those tasks is not an in-scope
finding without a concrete consequence for this guard's declared claim.

## Intended contract to challenge

- One explicitly marked internal-import-map:v1 fenced text block describes literal
  internal-relative package paths; '.' denotes the internal root and '-' alone a leaf.
  Ordering, whitespace, comments and wrapped comma-separated lists are immaterial.
  Ambiguous/duplicate/malformed declarations must not be accepted by guessing.
- Covered source rows are go-list-selected production Go/cgo packages under the
  current module's internal namespace, excluding testutil and the tools namespace.
  Include zero-edge packages. Source exclusions must not erase a production import
  into support code; do not conflate testutil with testutils, or tools with toolsmith.
- Only direct Imports in the internal namespace count. TestImports, XTestImports,
  transitive Deps, stdlib and third-party targets do not. Check prefix boundaries and
  the root-package case, not just today's convenient package names.
- Discovery is default-build metadata for the selected GOOS/GOARCH, not an all-platform
  or custom-tag/race union. The child commands disable GOWORK and ambient GOFLAGS,
  request -mod=readonly, capture errors, and have a bounded lifetime. Real failures
  must not become an empty/partial successful comparison.
- Both package inventory and edges are compared deterministically, including new or
  removed leaf packages. Drift gives actionable +/- facts without recommending that
  an illegal edge be ratified merely to clear the diff.
- Observed imports are descriptive facts. Package roles, named exceptions, accepted
  decisions and depguard policy retain their own authority. No product behavior,
  wire revision, dependency or general Markdown framework changes are intended.

These are hypotheses to test. Checked ACs and owner measurements of 26 packages /
78 edges are not independent evidence or fixed quotas for future packages.

## Mandatory evidence floor

1. Independently run the focused race suite twice from the captured sandbox:
   go test -race ./internal/tools/architecturecheck -count=1.
   Verify the named tests/declarations and actual native go-list facts. Use local
   go help list/build or primary Go documentation for metadata/flag claims.
2. Through TestArchitectureImportMap, make one marked-map semantic drift (for example
   remove core -> id), observe the intended named edge failure, restore it, and get
   green. Merely showing a parser unit failure is insufficient for this caller check.
3. Each reviewer must also make one new compiler-valid implementation mutation:
   Claude: make decodeProductionImports omit leaves OR erase runtime imports into
   support packages. Require TestProductionImportProjectionKeepsScopeAndLeaves to
   kill it, not just an unrelated current-map mismatch. Antigravity: suppress zero-edge
   package drift in importGraphDiff OR make parseImportMap accept conflicting duplicate
   declarations. Require the appropriate TestImportGraphDiffNamesPackageAndEdgeDriftDeterministically
   or TestParseImportMapRejectsAmbiguousOrMissingFacts test to kill it.
   Build the test package (go test ./internal/tools/architecturecheck -run '^$')
   before counting the probe. A compile
   error is invalid, not a killed mutation. Record exact edits, expected/actual failing
   tests, killed/survived/invalid outcome and restored green result.
4. Claude: use an independently authored throwaway Go module through the actual
   productionImportGraph/goList path. Show that test-only imports stay out, leaves
   stay in, and a production import into support code stays visible. Cross-check one
   documented formatting transformation and one invalid block. Antigravity: show a
   valid metamorphic map edit remains equal (reorder/wrap/comment/CRLF) and a hostile
   malformed block is refused. Cross-check one real-module coverage boundary, not
   only hand-authored JSON. Temporary tests are allowed only in the sandbox and must
   be removed before transfer; external fixture modules/caches are preferred.
5. Settle at least three bounded hypotheses in your primary lens with actual evidence.
   For a no-findings verdict, meet items 1-4 and label remaining concerns unresolved
   where untested. Do not generalize one successful helper probe to all callers.

## Required hostile angles

Go/discovery angles: source-vs-target exclusions; test-only/cgo/root packages; namespace
near-collisions; empty or truncated streams after valid records; duplicate records;
Incomplete/Error/DepsErrors/ForTest; ambient flags/workspaces; failed Go commands;
module-file effects and error diagnostics. Can failures be filtered before validation?
Are claimed protections reachable through the real subprocess rather than only JSON
fixtures? Respect the explicit native/default-build scope when assessing platform tags.

Parser/comparator angles: marker ambiguity or wrong order; fences/examples/comments
masking a missing fact; continuation joining; repeated declarations; '-' mixed with
targets; literal-path validation; changes to only package inventory; comparisons that
drop information in both representations; stable output independent of map ordering.
For a suspected missing fixture, demonstrate a surviving valid mutation or reachable
false result rather than asking for more tests in the abstract.

Make a distinct systemic second pass over the shared extraction/comparison pipeline
and immediate consumers. Could a coordinated helper/caller change preserve a misleading
green test? Does the guide say more than the actual scope? Do instructions silently
change policy or treat a future ADR as accepted? Is this narrow verifier proportionate
and maintainable, or does a concrete grammar/test coupling cause recurrent false alarms?
Distinguish current defects, demonstrated coverage gaps, defense-in-depth and taste.

Review-loop provenance: the preceding architecture-guide review reproduced a fenced
navigation false pass despite green helper tests, and corrected unverified destination
references in a clean report. This brief therefore requires real caller probes and
verified inventories. It does not presume a reviewer/model will find a particular bug.

## Validation and restoration

Owner ran full just test (race), just lint, just build, just docs-check, planning lint
--links and whitespace checks. Owner also killed a removed map edge and a compilable
editor -> domain import. You may inherit those broader gates, but not count them as
your own execution or reuse those two probes as your required new implementation
mutation. No wire/schema/repair-command mutation is applicable to this change.

Run repository commands only in your independent sandbox. Keep GOCACHE, lint cache,
built binaries, coverage and standalone fixtures in a reviewer-owned external temp
directory; never use the source tree for scratch. Review the listed paths, not ignored
nested workspaces such as .claude/worktrees or their Git control files. Restore every
probe to the sandbox capture, not origin/main or the older source HEAD. No staging,
branch changes, additional commits or pushes; the helper's one baseline checkpoint
is allowed. Do not run broad clean/reset/revert operations in the source checkout.

Before verification/transfer, only your assigned audit may differ. Run a final focused
green suite, audit finding lint using a sandbox-built binary if available, and diff
whitespace checks. Preserve the brief and the sibling audit. On isolation, execution
or transfer failure, retain the workspace and report incomplete; do not improvise a
shared-worktree fallback. Frozen report attestations must distinguish pre-transfer
pending from a real succeeded outcome; report the helper's actual final transfer
output in your handoff rather than inventing success before running it.

## Deliverable

Fill only the assigned report section. Findings first, consolidated by root cause,
with exact source/test/guide evidence, reachable consequence, reproduction, smallest
sound recommendation and residual uncertainty. Leave findings open for owner triage.
Coverage/maintenance gaps are normally Low unless a concrete consequence warrants
more; no findings is acceptable, evidence-free certainty is not.

Include verdict/limits, verified consumer inventory, checks executed vs inherited,
concise mutation/probe ledger, a bounded reproduced/falsified/unresolved hypothesis
ledger, and isolation attestation (sandbox/Git directory, baseline, captured audit
blob/fingerprint, verify and actual transfer outcomes). Verify every named file,
symbol, test and planning destination. Do not implement fixes, commit or push.

## Reviewer report

Reviewer: Claude (Opus 5.5), discovery and coverage-fidelity lead. All inspection, builds, tests, fixtures,
mutations and edits ran in the independent sandbox listed in the isolation attestation below.
GOCACHE, the built binary and fixture modules live in a reviewer-owned scratch directory outside both trees.

### Verdict and limits

**Not ready as claimed: one Medium and two Low findings, all left open.**

On the default `go list` path the parser, projection and diff are correct for the declared scope.
Every required probe behaved as intended:

- Test-only imports are excluded and leaves are kept.
- Production imports into `testutil` and `tools/*` stay visible.
- The root package and the `testutils`/`toolsmith` near-misses are handled correctly.
- Failed subprocesses refuse to compare.
- Map drift fails with the named edge.
- The required decoder mutation is killed by the intended test.

The defects sit between the test and the child `go list` process:

- **M1:** Go's test-result cache cannot see the sources `go list` reads. The documented `just test` route
  therefore replays a cached pass after real import drift.
- **L1:** `GOFLAGS=` does not neutralize a persisted `go env -w GOFLAGS`. This contradicts the guide's
  isolation claim and makes `-e` reachable, where one fail-closed ordering is unpinned.
- **L2:** Wildcard discovery is not closure. Importable production packages in `_`-prefixed or `testdata`
  directories get no source row, and nothing fails.

None of these changes the current map's facts. 26 rows, 78 edges and four leaves match native `go list` on
darwin/arm64, linux/amd64, windows/amd64, linux/386 and js/wasm.

Limits:

- I did not run anything in GitHub Actions or on Windows.
- The CI cache argument in M1 is reasoned from Go source, not observed.
- Nested modules under `internal/` and the hang/timeout path are unresolved (see the ledger).

## Findings

#### M1. Warm Go test cache replays a pass after production import drift · **Status:** fixed

**File:** internal/tools/architecturecheck/imports_test.go:38 | **Component:** tools/architecturecheck
**Effort:** XS · **Urgency:** soon

`TestArchitectureImportMap` gets its code-side facts from a child process (`imports_test.go:38`, which calls
`goList` at `:140` and `:147-161`). Go caches a test result under a key built from three things:

- the test binary;
- the cacheable flags;
- the environment variables and files that the **test process itself** opened.

Regular files are hashed by size and mtime. Directories are hashed by their entries' stat
(`hashOpen`, `$GOROOT/src/cmd/go/internal/test/test.go:2094`).

Three facts combine here:

- The architecturecheck test binary depends only on the standard library. `go list -test -deps -f '{{if not .Standard}}…'`
  lists only the package itself.
- The test opens the guides but no `internal/**/*.go` file.
- `-race` and `-coverprofile` are both cacheable (`test.go:1855-1873`).

So an import edit that does not touch `boundaries.md` leaves the cache key unchanged.

Reproduction in the sandbox (external GOCACHE, Go 1.27.2 darwin/arm64):

1. Warm the cache: run `go test -race ./...` twice. The second run reports `ok …/architecturecheck (cached)`.
2. Edit in place: add `_ "github.com/andy-esch/taskflow/internal/domain"` to `internal/editor/editor.go`.
   `go build ./internal/editor` succeeds.
3. Wait 3 s, then rerun. The result is **`ok …/architecturecheck (cached)`**.

This held in 4/4 controlled trials: two with the literal `just test` command (`go test -race ./...`) and two
with the focused `go test -race ./internal/tools/architecturecheck`. On the same tree, `-count=1` fails with
`+ edge editor -> domain`.

One earlier uncontrolled `just test` run did re-execute and fail. It ran right after a `git checkout` restore
had invalidated the entry. The controlled trials show the warm-cache behavior, which is the normal state of a
repeated `just test`.

Consequences:

- `boundaries.md:67-68` offers `just test` as an equivalent way to run the check. AGENTS.md gates code
  changes on `just test` (`Justfile:52-53`, no `-count`).
- A contributor or agent who adds or removes an internal import without touching the guide gets a green
  local run. That is exactly the drift this guard exists to catch.
- CI is probably protected, but only incidentally. `actions/checkout` writes fresh mtimes, so the opened
  guide's hash changes every run even when setup-go restores GOCACHE (`ci.yml:112`). I reasoned this from
  `hashOpen`; I did not observe it in Actions.
- A workspace that keeps mtimes across runs (for example, a persistent runner) would not have that protection.
- Map-only edits do invalidate the cache, because the guide is an opened file. That is why the owner's and my
  map-drift probes fail correctly. It also explains why this went unnoticed: those probes use `-count=1` or
  edit the guide.

**Recommendation:** Before calling `productionImportGraph`, make the inputs that `go list` reads visible to the
test cache. Walk `filepath.Join(root, "internal")` with `filepath.WalkDir`, which opens every directory and so
records each file's size and mtime, and stat `go.mod`. Add a comment explaining why.

I checked this in the sandbox with a three-line `WalkDir` before `:38`, then restored it. In 2/2 warm-cache
trials, the drift then failed with `+ edge editor -> domain`, and the run went green again once the drift was
undone. A weaker alternative is to drop "(or `just test`)" from the guide and require `-count=1`.

**Resolution:** Recorded internal directories/Go-file metadata and module
    manifests in the test process before child discovery.
    TestImportMapCacheInvalidatesProductionDrift proves a cache hit, then
    rejects changed imports and added/removed leaves without -count=1.
    Suppressing input recording compiled and made all three cases replay a
    cached pass; restored full race suite passed.

#### L1. Persisted go env GOFLAGS still configures the "default build" discovery · **Status:** fixed

**File:** internal/tools/architecturecheck/imports_test.go:154 | **Component:** tools/architecturecheck
**Effort:** XS · **Urgency:** later

`goList` appends `GOFLAGS=` (`imports_test.go:154`). Its comment says this excludes "-test/-find/custom-tag flags"
(`:152-153`), and the guide says the check "disables workspace mode and ambient GOFLAGS" (`boundaries.md:64`).
In fact, the go command treats an **empty** environment value as unset. It then falls back to the per-user
go env file that `go env -w` writes (`cmd/go/internal/cfg/cfg.go:437-442`; `base.InitGOFLAGS` reads
`cfg.Getenv("GOFLAGS")`). `GOWORK=off` is non-empty, so it does win.

The probe uses an external fixture with `internal/tagged/tagged.go`, plus `tagged_tag.go`
(`//go:build reviewtag`, imports `c`) and `tagged_default.go` (`//go:build !reviewtag`, imports `d`).
The GOENV file contains `GOFLAGS=-tags=reviewtag`.

| GOFLAGS environment value | `go list` imports of `tagged` |
| --- | --- |
| unset | `[c]` |
| `GOFLAGS=` (what `goList` sets) | `[c]` |
| `GOFLAGS=-mod=readonly` | `[d]` |

`GOENV=<file> GOFLAGS= go env GOFLAGS` prints `-tags=reviewtag`.

Through the real `productionImportGraph` (a temporary sandbox test, since deleted):

- With that GOENV, discovery yields `tagged -> c`. A map claiming `tagged -> c` diffs as `""`, a pass,
  although the default build is `tagged -> d`.
- With `GOENV=off`, the same map yields `- edge tagged -> c` / `+ edge tagged -> d`.

The same path also makes `-e` reachable, and here one ordering is unpinned:

- With GOENV holding `GOFLAGS=-e`, a module whose `internal/tools/t` has a missing import still fails closed
  in the unmodified decoder (`…incomplete, erroneous or test-variant metadata for ".../internal/tools/t"`).
- Mutant X3 moves the error check (`:185-187`) after the support/test-only skip (`:192-194`). It compiles and
  **survives every shipped test**.
- Under persisted `-e`, X3 returned `map[core:[]]` with a nil error.
- Every erroneous record in `TestProductionImportProjectionFailsClosed` is `core` (`imports_format_test.go:84-91`).
- The owner's isolation regression only uses the environment-variable form (`imports_format_test.go:144`).

Consequence today:

- No current-map effect. No `internal/` file uses a custom or `race` build tag; the only constrained files are
  the `!unix` `lock_other.go` files, which add no new internal edge.
- The impact is local and latent. A developer with persisted `-tags`/`-race` GOFLAGS checks a non-default build.
  That produces false drift, or a false local pass that CI later contradicts.
- The comment and guide overstate the isolation.

**Recommendation:**

- Use a non-empty override such as `GOFLAGS=-mod=readonly` instead of `GOFLAGS=`. I validated this at the
  command level above. `GOENV=off` also works, but it would drop user GOPROXY/GOPRIVATE settings.
- In the real-module test, add `t.Setenv("GOENV", <file with GOFLAGS=-tags=x>)` and a tag-constrained fixture file.
- Add one FailsClosed case with an erroneous `tools/...` or test-only record, to pin validate-before-skip
  (it kills X3).

**Resolution:** GOFLAGS=-mod=readonly overrides both process and persisted
    configuration while preserving other Go settings. The real-module regression
    proves default-tag imports with persisted GOENV flags. Erroneous
    support/test-only records are tested after a valid row; moving validation
    after filtering compiles and is now killed by the intended cases. Full race
    suite and lint passed.

#### L2. Wildcard discovery omits importable production packages without failing · **Status:** fixed

**File:** internal/tools/architecturecheck/imports_test.go:140 | **Component:** tools/architecturecheck
**Effort:** XS · **Urgency:** later

Source rows come from `go list ./internal/...` (`imports_test.go:140`). Go wildcards skip directories that
begin with `_` or `.`, and `testdata` directories. Packages in those directories are still importable and get
compiled into production binaries. Targets come from `Imports` (`:199-206`), and nothing requires a
non-support target to have its own row.

Fixture A: `internal/e` imports `internal/_hidden`, which imports `internal/secret`, and `cmd/app` imports `e`.

- `go list -deps ./cmd/app` includes `_hidden` and `secret`.
- `productionImportGraph` returned `e -> _hidden` and a `secret` row, but **no `_hidden` row**.
  So `_hidden -> secret` is invisible.

Fixture B: `internal/p` imports `internal/p/testdata/td`.

- It builds.
- `go list ./internal/...` lists only `leaf` and `p`, while `-deps` shows `p/testdata/td`.

Consequences:

- The guide claims the map "records direct module-internal imports for production packages, including leaves"
  (`boundaries.md:24-25`).
- The task's coverage decision says "every package under this module's `internal/` with selected non-test Go
  or cgo files".
- Both claims are broader than what is enforced. A production package in a wildcard-skipped directory can
  import any internal package unnoticed; only its parent's edge appears.
- These directory names are unusual, hence Low.
- In the current tree, every documented target also has a row (computed from the parsed map).

**Recommendation:**

- After decoding, fail when a non-support internal target (not `testutil`, `tools` or `tools/*`) has no covered
  source row. Name the target and mention wildcard-skipped directories. Add a JSON regression in which `core`
  imports `_hidden` and there is no `_hidden` record.
- The same check would flag nested modules under `internal/`, whose imports are counted as internal targets
  today (not executed).
- At minimum, state the wildcard-selection limit in the guide.

**Resolution:** Reject non-support internal targets lacking a source row, with
    deterministic target/importer and wildcard-scope diagnostics. Compiler-valid
    hidden, testdata and nested-module fixtures fail through actual discovery;
    disabling the refusal compiles and is killed by all three. Guide/task scope
    now explicitly excludes unreferenced wildcard-skipped packages rather than
    claiming universal closure.

## Reviewer evidence

### Verified consumer inventory

Every name below was verified in the sandbox at baseline `4b8c7d0c`.

`internal/tools/architecturecheck/imports_test.go`:

| Symbol | Line |
| --- | --- |
| `importMapStart`/`importMapEnd` | :21-22 |
| `importGraph` | :29 |
| `TestArchitectureImportMap` | :31 |
| `parseImportMap` | :47 |
| `validInternalPackageName` | :121 |
| `productionImportGraph` | :125 |
| `goList` | :147 (env :154, 30 s timeout :148) |
| `listedPackage` | :164 |
| `decodeProductionImports` | :175 |
| `internalImportName` | :216 |
| `importGraphDiff` | :224 |

`internal/tools/architecturecheck/imports_format_test.go`:

| Symbol | Line |
| --- | --- |
| `markedImportMap` | :12 |
| `TestParseImportMapIgnoresPresentationChanges` | :16 |
| `TestParseImportMapRejectsAmbiguousOrMissingFacts` | :31 |
| `TestProductionImportProjectionKeepsScopeAndLeaves` | :58 |
| `TestProductionImportProjectionFailsClosed` | :82 |
| `TestProductionImportGraphUsesDefaultModuleBuildWithoutEdits` | :122 |
| `TestImportGraphDiffNamesPackageAndEdgeDriftDeterministically` | :162 |

Other consumers and unchanged files:

- **Callers:** none outside these two files. A repository grep finds only the guide links
  (`boundaries.md:62`, `:67`) and the index link (`ARCHITECTURE.md:48`).
- **`docs/architecture/boundaries.md`:**
  - markers `:29` and `:60`;
  - scope prose `:24-27`;
  - check, remediation and grammar prose `:62-77`.
- **`docs/ARCHITECTURE.md`:** check claim `:48-51`; map ownership `:55-56`.
- **Unchanged against base `86e255be`** (`git diff --stat` is empty):
  - `internal/tools/architecturecheck/architecture_test.go`;
  - `.golangci.yml`;
  - `Justfile` (`test` at `:52-53`);
  - `.github/workflows/ci.yml` (`go test -p 4 -race -coverprofile=coverage.out ./...` at `:112`, ubuntu-latest only);
  - `go.mod` and `go.sum`.
- **Planning destinations** (sandbox-built `tskflwctl`):
  - `6g9cz5sk3b0d`, `6g9cz5sv6kck`, `6g9cz5t37dc0`, `6gg7e59mm68g` and `6ggxymjbf86r` exist, all `ready-to-start`;
  - `6g63hjm7cp6w` is `in-progress`;
  - Thread `6g9czp7g9pt3` shows `6g63hjm7cp6w` in flight.
- **ADR:** `planning/adrs/` contains no dependency-policy ADR. That ADR exists only as task `6gg7e59mm68g`.

### Checks executed and inherited

Executed by me in the sandbox:

- `go test -race ./internal/tools/architecturecheck -count=1` twice, both green with 14 tests.
- A full `just test` (`go test -race ./...`), green.
- Native `go list` facts:
  - the selected-package inventory (35 directories, 35 listed);
  - ignored files (`lock_other.go` in config, store and userconfig);
  - the projected graph for five GOOS/GOARCH targets, identical across all five.
- `go build` of the CLI to an external path.
- The final focused suite, `git diff --check` and `audit lint` (results in the attestation).

Inherited from the owner and not counted as mine:

- `just lint`, `just build`, `just docs-check`, and planning lint with `--links`.
- The owner's `core -> id` and `editor -> domain` probes.

### Mutation and probe ledger

| Probe | Exact change | Expected | Actual | Outcome |
| --- | --- | --- | --- | --- |
| Map drift (floor 2) | `boundaries.md:41` `core -> domain, id` → `core -> domain` | `TestArchitectureImportMap` names the edge | Failed: `+ edge core -> id` | Killed; restored and green |
| X1 support erasure (floor 3) | `:200` adds `&& target != "testutil" && target != "tools" && !strings.HasPrefix(target, "tools/")` | `TestProductionImportProjectionKeepsScopeAndLeaves` | Only that test failed (`runtime` lost `testutil`, `tools/helper`). `TestArchitectureImportMap` and the owner's real-module test passed. My real-module probe also failed. | Killed (compile gate `-run '^$'` ok) |
| X2 omit leaves | `:208` wrapped in `if len(targets) > 0` | Leaf tests | Failed: `KeepsScopeAndLeaves`, `UsesDefaultModuleBuildWithoutEdits`, `TestArchitectureImportMap` | Killed (compile ok) |
| X3 filter before validate | Error check `:185-187` moved after skip `:192-194` | A fail-closed test | All shipped tests passed. Under persisted `-e`, a broken `tools/t` gave `map[core:[]]` with nil error. | **Survived** (see L1) |
| Cache false pass | In-place `editor -> domain` blank import after a warm cache | `TestArchitectureImportMap` fails | `ok (cached)`, 4/4 | **Reproduced** (M1) |
| Cache fix check | `filepath.WalkDir(root/internal)` before `:38` | Drift fails despite warm cache | 2/2 failed, green again once undone | Fix viable; restored |

Floor 4: an independently written module drove the real `productionImportGraph`/`goList` path.

- **Fixture contents:**
  - `a.go` imports `b`, `testutil`, `tools/x` and `fmt`;
  - `a_test.go` imports `c`;
  - `a_x_test.go` (package `a_test`) imports `d`;
  - `testutil -> b`, `tools/x -> d`, `toolsmith -> b`;
  - leaves `testutils`, `b`, `c`, `d`;
  - test-only `onlytests`;
  - root `internal/root.go -> b`.
- **Result:** `{".":[b], a:[b testutil tools/x], b, c, d, e:[_hidden], secret, tagged:[d], testutils, toolsmith:[b]}`.
  - Test-only imports are absent.
  - Leaves are present.
  - Production imports into support code are visible.
  - `testutil`, `tools/x` and `onlytests` have no rows.
- **Parser cross-check:**
  - Reversing all logical rows of the real guide map, rewrapping `tui` with a trailing comment, and adding
    comment lines parsed equal to the original.
  - A continuation that lost its comma (`core -> domain` / `id`) was refused:
    `invalid import-map target "domain id"`.
- **Real-subprocess failures all returned errors with diagnostics:**
  - a broken import: exit 1, `import lookup disabled by -mod=readonly`;
  - no `internal/` directory: `lstat ./internal/`;
  - a missing root: `chdir`.
- **Module files:** every fixture `go.mod` was unchanged, and no `go.sum` or `go.work` was created.
- **Probe files:** the temporary probe tests were deleted from the sandbox before verification.

### Hypothesis ledger

| # | Hypothesis | Result | Evidence |
| --- | --- | --- | --- |
| 1 | TestImports/XTestImports leak into rows | Falsified | Real module: `a` lacks `c` and `d` |
| 2 | Leaves dropped | Falsified | Real module rows; X2 killed by three tests |
| 3 | Source exclusions erase production imports into support | Falsified | `a -> testutil, tools/x`; X1 killed |
| 4 | Root package or prefix near-collisions mis-scoped | Falsified | `.` row; `testutils` and `toolsmith` kept; prefix uses `root+"/"` (`:221`) |
| 5 | Failed `go list` becomes an empty or partial comparison | Falsified on the default path | Three subprocess failures error out |
| 6 | Incomplete/Error/DepsErrors/ForTest checks are only reachable from JSON fixtures | Partly reproduced | Unreachable by default (`go list` exits 1 first); reachable via persisted `-e`; ordering unpinned (X3) |
| 7 | Ambient GOFLAGS cleared | **Reproduced false** | L1 |
| 8 | Wildcard selection equals production closure | **Reproduced false** | L2 |
| 9 | Test result reflects current sources | **Reproduced false** under a warm cache | M1 |
| 10 | Platform-dependent map causes recurrent false alarms | Falsified for this tree | Identical projection on five targets; CI is ubuntu-only |
| 11 | Child uses a different toolchain | Falsified | `go test` prepends `$GOROOT/bin` to PATH (`test.go:92`, `:1676`) |
| 12 | Hang beyond the 30 s timeout | Unresolved | `CommandContext` without `WaitDelay`; not exercised |
| 13 | Nested module under `internal/` | Unresolved | Not executed; same class as L2 by code reading |
| 14 | `testutil` versus `tools` exclusion asymmetry | Taste, not a defect | `testutil/x` would need a row because `:192` is exact-match, so it fails noisily; no `testutil` subpackage exists today |

### Second pass (systemic)

- **Shared pipeline.** The weak point is the boundary between the test and the child process, not parsing or
  diffing. M1 and L1 are both cases of the child's inputs or configuration differing from what the test
  harness assumes. They have separate root causes and separate fixes.
- **Test that could mask a defect.** Support-target visibility is pinned only by the JSON fixture. A real-module
  production edge into `testutil` would make that real, but the fixture does kill X1, so this is not a gap.
- **Guide versus scope:**
  - `boundaries.md:64` overstates GOFLAGS isolation (L1).
  - `:24-25` overstates coverage (L2).
  - `:68` "(or `just test`)" is the route that serves stale results (M1).
- **Policy.** The check introduces no policy change:
  - The failure message (`:43`) and the guide say observed imports are not an allowlist, and they route review
    to depguard and the named exceptions.
  - The comment at `imports_test.go:28` says "the policy ADR remain[s] separate". That ADR is only planned
    (`6gg7e59mm68g`), but the wording does not treat it as accepted. Not a finding.
- **Proportionality.** About 256 lines of checker and 180 lines of tests are proportionate. I found no
  grammar/test coupling that would cause recurring false alarms; new packages needing rows is intended.

### Isolation attestation

| Item | Value |
| --- | --- |
| Source checkout (read-only) | `/Users/andyeschbacher/git/andy-esch/taskflow`, branch `main`, base `86e255be59f7389dde526d139bfc9ef42f0f5582` |
| Sandbox | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.ZiebVb` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.ZiebVb/.git` (independent clone; not a worktree or symlink) |
| Baseline commit | `4b8c7d0cc8c05e7d3682621e713fd7c06fbd5f26` ("chore: capture isolated review baseline"; the only sandbox commit) |
| Captured source blob | `636cac46ce6d80cdabc0d5b0233858790ae4f433` |
| Source fingerprint | `8e419dfb4e590d50126fd9c851522e94f728fd72` |
| Deliverable | `planning/audits/6gkabd0b0bg6-2026-10-13-documented-import-map-implementation-claude.md` |
| Scratch (GOCACHE, binary, fixtures) | `/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/1a882109-cddd-4beb-a131-ae6a7ab341d6/scratchpad/review` |
| Restoration | Every probe was restored to the baseline; before verification only this audit differs |
| Final checks | Probe files absent from the package directory; `go test -race ./internal/tools/architecturecheck -count=1` ok; `git diff --check` clean; sandbox-built `audit lint 6gkabd0b0bg6` passed (3 open findings); planning `lint` passed |
| Verify | Helper `verify` exited 0 with `baseline_commit=4b8c7d0c…`, `source_blob=636cac46…`, `source_fingerprint=8e419dfb…`, `deliverable_changed=true` and `transfer=pending`. It was rerun after this row was written, immediately before transfer; that result is in the handoff. |
| Transfer | Pending when this report was frozen. The helper's actual transfer output is reported in the handoff message, not asserted here. |

## Owner disposition (2026-10-13)

Accepted all three findings and reproduced them with owner-authored regressions before fixes. Four compiler-valid reversions (cache recording, GOFLAGS override, validation-before-filter and uncovered-target refusal) failed their intended tests, then were restored. Final just test (race), just lint (0 issues), just build, just docs-check, planning lint --links and whitespace checks passed.

The reviewer verdict and frozen attestation above describe the reviewed baseline, not the subsequently corrected implementation. This report has landed in its assigned source audit; no final transfer stdout was supplied to the owner, so the frozen pending row is not rewritten as proof of transfer. Antigravity remains independent and pending. The implementing task remains in progress until that report is triaged.
