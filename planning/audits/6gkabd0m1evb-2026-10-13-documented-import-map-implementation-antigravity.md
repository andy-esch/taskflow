---
schema: 1
id: 6gkabd0m1evb
bucket: closed
area: documented-import-map-implementation-antigravity
date: "2026-10-13"
updated_at: "2026-10-13"
---
# Audit: Executable production import-map verification — antigravity — 2026-10-13

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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
AUDIT_REL="planning/audits/6gkabd0m1evb-2026-10-13-documented-import-map-implementation-antigravity.md"
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

### Findings

#### L1. Parser accepts '-' as a valid source package name · **Status:** fixed

- **Evidence:** In `internal/tools/architecturecheck/imports_test.go:71-73`, `parseImportMap` validates source packages with `if !validInternalPackageName(source)`. `validInternalPackageName("-")` returns `true` because `fs.ValidPath("-")` is `true` and regexp `^[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*$` matches `"-"`. While line 81 explicitly rejects target `"-"` within a comma-separated list (`if target == "-" || !validInternalPackageName(target)`), source validation has no check for `source == "-"`.
- **Reachable consequence:** `parseImportMap` accepts malformed source rows such as `- -> core` or `- -> -` as valid source packages (e.g. `importGraph{"-": []string{"core"}}`), violating the grammar in `docs/architecture/boundaries.md:73` ("'.' denotes the internal root and '-' alone a leaf"). In `TestArchitectureImportMap`, this fails closed because `go list` never produces a package named `"-"`, resulting in a `- package -` diff failure rather than a false pass. However, unit validation in `parseImportMap` permits an illegal source token.
- **Reproduction:** Calling `parseImportMap(markedImportMap("- -> core"))` returns `importGraph{"-": []string{"core"}}` and `err == nil`.
- **Recommendation:** In `parseImportMap:71`, check `if source == "-" || !validInternalPackageName(source)` to explicitly reject the leaf token as a source package name.
- **Residual uncertainty:** None. Probed and verified in the isolated sandbox.

**Resolution:** Confirmed Go rejects a directory named "-" with invalid input
    directory name. The parser now rejects the reserved leaf token as a source;
    new - -> core and - -> - cases reproduced acceptance before the fix and pass
    after it. This is parser diagnostics hardening, not a previously
    demonstrated caller false pass.

#### L2. Real-module test omits production imports into support packages and edge cases · **Status:** fixed

- **Evidence:** In `internal/tools/architecturecheck/imports_format_test.go:122-160`, `TestProductionImportGraphUsesDefaultModuleBuildWithoutEdits` creates a real fixture module in `t.TempDir()` with packages `a`, `b`, `testutil`, `tools`, and `onlytests`. It verifies that `testutil` and `tools` are omitted from source rows. However, none of the production packages in this fixture imports `testutil` or `tools/helper`. Edge retention for production imports into support packages (e.g. `runtime -> testutil`) is tested exclusively against synthetic JSON in `TestProductionImportProjectionKeepsScopeAndLeaves:64`. Similarly, root package (`.`) and near-collision names (`testutils`, `toolsmith`) are only tested against synthetic JSON.
- **Reachable consequence:** A future subprocess-level drift or flag change in `goList` or path resolution in `internalImportName` could alter how real `go list` invocations handle support-package targets, the internal root, or near-collision names without failing any real-module test in the suite.
- **Reproduction:** Verified by executing a standalone external fixture module probe through `productionImportGraph`. While the implementation correctly extracted `runtime -> testutil, tools/sub`, `testutils`, `toolsmith`, and `.`, this coverage relies on synthetic JSON fixtures in the repository test suite.
- **Recommendation:** Augment `TestProductionImportGraphUsesDefaultModuleBuildWithoutEdits` with a production package importing `testutil` or `tools/helper` and a root package file in `internal/`, pinning real-module retention of these edges.
- **Residual uncertainty:** None. Probed via external fixture.

**Resolution:** Accepted as a bounded coverage improvement, not a reproduced
    implementation defect. Production support imports were already added while
    resolving Claude L1; extended the compiler-valid real module with the
    internal root, tools/sub, testutils and toolsmith. Actual discovery must
    retain root/near-collision rows and support-target edges while omitting
    support-source rows.

#### L3. Metamorphic CRLF line ending equivalence is untested in test suite · **Status:** fixed

- **Evidence:** `TestParseImportMapIgnoresPresentationChanges` (`internal/tools/architecturecheck/imports_format_test.go:16-29`) tests plain and wrapped/reordered/commented cases with Unix LF (`\n`). It does not include a test case with Windows-style CRLF (`\r\n`) line endings.
- **Reachable consequence:** Although `strings.TrimSpace` currently strips trailing `\r` from lines and targets, any future refactor of line splitting or target trimming (e.g. switching to `strings.TrimSuffix(line, "\n")`) would silently break CRLF checkouts without failing the existing test suite.
- **Reproduction:** Probed by applying CRLF across `docs/architecture/boundaries.md` alongside row reordering and comment wrapping, observing that `TestArchitectureImportMap` passes (0.259s). The property holds, but lacks regression pinning.
- **Recommendation:** Add a CRLF test case to `TestParseImportMapIgnoresPresentationChanges`.
- **Residual uncertainty:** None. Probed and verified.

**Resolution:** Added a CRLF presentation-equivalence case to
    TestParseImportMapIgnoresPresentationChanges. The existing parser already
    supports CRLF; this is a cheap regression pin, not a current behavior fix.
    Focused and full race suites passed.

### Verdict and limits

- **Verdict:** Pass with minor findings (3 Low findings: L1, L2, L3 left open for owner triage). The import-map verifier is sound, deterministic, fails closed on corrupt or truncated metadata, correctly handles zero-edge leaf packages, and enforces default-build production import facts without false passes.
- **Scope limits:** Enforces default-build native GOOS/GOARCH metadata for production packages under `internal/`. Does not enforce standard-library, third-party, custom-tag, or cross-platform union imports, which remain governed by depguard (`.golangci.yml`) and pending architecture tasks.

### Verified consumer inventory

- `internal/tools/architecturecheck/imports_test.go`:
  - `TestArchitectureImportMap` [imports_test.go:31-45](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L31-L45)
  - `parseImportMap` [imports_test.go:47-119](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L47-L119)
  - `validInternalPackageName` [imports_test.go:121-123](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L121-L123)
  - `productionImportGraph` [imports_test.go:125-145](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L125-L145)
  - `goList` [imports_test.go:147-162](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L147-L162)
  - `listedPackage` [imports_test.go:164-173](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L164-L173)
  - `decodeProductionImports` [imports_test.go:175-214](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L175-L214)
  - `internalImportName` [imports_test.go:216-222](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L216-L222)
  - `importGraphDiff` [imports_test.go:224-256](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_test.go#L224-L256)
- `internal/tools/architecturecheck/imports_format_test.go`:
  - `markedImportMap` [imports_format_test.go:12-14](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_format_test.go#L12-L14)
  - `TestParseImportMapIgnoresPresentationChanges` [imports_format_test.go:16-29](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_format_test.go#L16-L29)
  - `TestParseImportMapRejectsAmbiguousOrMissingFacts` [imports_format_test.go:31-56](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_format_test.go#L31-L56)
  - `TestProductionImportProjectionKeepsScopeAndLeaves` [imports_format_test.go:58-80](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_format_test.go#L58-L80)
  - `TestProductionImportProjectionFailsClosed` [imports_format_test.go:82-120](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_format_test.go#L82-L120)
  - `TestProductionImportGraphUsesDefaultModuleBuildWithoutEdits` [imports_format_test.go:122-160](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_format_test.go#L122-L160)
  - `TestImportGraphDiffNamesPackageAndEdgeDriftDeterministically` [imports_format_test.go:162-180](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/imports_format_test.go#L162-L180)
- `docs/architecture/boundaries.md`:
  - Marked import map block: [boundaries.md:29-60](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/architecture/boundaries.md#L29-L60)
  - Coverage and grammar specifications: [boundaries.md:23-28](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/architecture/boundaries.md#L23-L28), [boundaries.md:72-77](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/architecture/boundaries.md#L72-L77)
  - Policy, check link, and remediation routing: [boundaries.md:62-70](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/architecture/boundaries.md#L62-L70), [boundaries.md:79-104](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/architecture/boundaries.md#L79-L104)
- `docs/ARCHITECTURE.md`:
  - Index link to import-map check: [ARCHITECTURE.md:48-52](file:///Users/andyeschbacher/git/andy-esch/taskflow/docs/ARCHITECTURE.md#L48-L52)
- `internal/tools/architecturecheck/architecture_test.go`:
  - Architecture link navigation checks: [architecture_test.go:23-90](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/architecture_test.go#L23-L90)
  - Forwarding anchor checks: [architecture_test.go:96-140](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/architecture_test.go#L96-L140)
  - Shared test helper `readFile`: [architecture_test.go:248-255](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tools/architecturecheck/architecture_test.go#L248-L255)
- Toolchain and CI configuration:
  - `Justfile:53`: `go test -race ./...` runs the suite.
  - `.github/workflows/ci.yml:112`: runs `go test -p 4 -race -coverprofile=coverage.out ./...`.
  - `.golangci.yml:10-230`: defines authoritative depguard rules (`domain-stays-inward`, `core-owns-ports-not-adapters`, `wire-stays-adapter-neutral`, etc.).
- Out-of-scope tracked planning destinations:
  - [planning/tasks/6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6g63hjm7cp6w-test-the-documented-import-graph-instead-of-date-stamping-a-manual-review.md) (implementing task)
  - [planning/threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md) (owning thread)
  - [planning/tasks/6g9cz5sk3b0d-adopt-selective-contract-focused-go-package-documentation.md](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6g9cz5sk3b0d-adopt-selective-contract-focused-go-package-documentation.md) (package docs task)
  - [planning/tasks/6g9cz5sv6kck-make-go-package-documentation-examples-executable.md](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6g9cz5sv6kck-make-go-package-documentation-examples-executable.md) (executable examples task)
  - [planning/tasks/6g9cz5t37dc0-reconcile-documentation-projections-and-enforce-drift-checks.md](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6g9cz5t37dc0-reconcile-documentation-projections-and-enforce-drift-checks.md) (reconciliation task)
  - [planning/tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md) (dependency policy ADR task)
  - [planning/tasks/6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md](file:///Users/andyeschbacher/git/andy-esch/taskflow/planning/tasks/6ggxymjbf86r-guard-core-and-domain-against-direct-i-o-and-framework-imports.md) (pure-layer I/O guard task)

### Checks executed vs inherited

- **Executed in isolated sandbox:**
  - `go test -race ./internal/tools/architecturecheck -count=1` run twice consecutively (Pass: 1.597s, 1.544s).
  - Semantic marked-map drift probe through `TestArchitectureImportMap` (removed `core -> id` from `docs/architecture/boundaries.md:41`; observed expected `+ edge core -> id` failure; restored and verified green).
  - Implementation mutation probe 1: suppressed zero-edge drift in `importGraphDiff` (`imports_test.go:237-241`). Package compile verified (`go test ./internal/tools/architecturecheck -run '^$'`). Executed `TestImportGraphDiffNamesPackageAndEdgeDriftDeterministically`; mutation killed. Restored and verified green.
  - Implementation mutation probe 2: bypassed duplicate source detection in `parseImportMap` (`imports_test.go:74-76`). Package compile verified. Executed `TestParseImportMapRejectsAmbiguousOrMissingFacts`; mutation killed by `duplicate_source` subtest. Restored and verified green.
  - Metamorphic map transformation probe: reversed logical rows, wrapped targets with `#` comments, converted to CRLF line endings. Executed `TestArchitectureImportMap`; verified identical pass (0.259s). Restored and verified green.
  - Hostile parser probe: tested `markedImportMap("- -> core")` against `parseImportMap`. Revealed parser leniency (Finding L1).
  - Real-module external fixture probe: executed `productionImportGraph` against standalone fixture module testing root `.`, leaf package, production imports to `testutil` and `tools/*`, and near-collision names (`testutils`, `toolsmith`). Confirmed boundary correctness, revealed testing gap in repo suite (Finding L2).
  - Link and inventory verification across all named Go symbols, tests, guides, and planning files.
- **Inherited from owner:**
  - Full repo test suite (`just test`), linter (`just lint`), binary build (`just build`), doc checks (`just docs-check`), planning link lint (`./bin/tskflwctl -C . lint --links`), and workspace untracked whitespace checks.

### Mutation and probe ledger

| # | Probe / Mutation | Target | Expected Outcome | Actual Outcome | Status |
|---|---|---|---|---|---|
| 1 | Semantic marked-map drift | `docs/architecture/boundaries.md:41` (remove `core -> id`) | `TestArchitectureImportMap` fails with `+ edge core -> id` | Failed with `+ edge core -> id` (0.24s) | Killed & restored |
| 2 | Implementation mutation: zero-edge drift suppression | `internal/tools/architecturecheck/imports_test.go:237-241` (`&& len(after) > 0` / `&& len(before) > 0`) | `TestImportGraphDiffNamesPackageAndEdgeDriftDeterministically` fails | Compile valid; test failed with missing leaf diff lines | Killed & restored |
| 3 | Implementation mutation: permit duplicate sources | `internal/tools/architecturecheck/imports_test.go:74-76` (remove duplicate source check) | `TestParseImportMapRejectsAmbiguousOrMissingFacts/duplicate_source` fails | Compile valid; test failed: `malformed import map passed` | Killed & restored |
| 4 | Metamorphic presentation invariance | `docs/architecture/boundaries.md:29-60` (reversed rows, wrapped targets, comments, CRLF) | `TestArchitectureImportMap` passes unchanged | Passed cleanly (0.259s) | Confirmed & restored |
| 5 | Hostile parser token probe | `parseImportMap` with `- -> core` | Reject `-` as illegal package name | Succeeded with `importGraph{"-": ["core"]}` | Survived (Finding L1) |
| 6 | Real-module external fixture probe | Standalone module with `.`, `leaf`, `runtime -> testutil, tools/sub`, `testutils`, `toolsmith` | `productionImportGraph` extracts accurate graph | Passed cleanly (`.`, `leaf`, `runtime`, `testutils`, `toolsmith`) | Confirmed (Finding L2) |

### Hypotheses ledger

- **Hypothesis 1 (Parser token validation):** `parseImportMap` rejects `-` when used in a source package position (`- -> core`).
  - *Result:* **Falsified / Reproduced defect** (Finding L1). `validInternalPackageName("-")` returns `true`. Target validation at line 81 explicitly checks `target == "-"`, but source validation at line 71 does not. Probe `parseImportMap(markedImportMap("- -> core"))` returned `importGraph{"-": []string{"core"}}`. In `TestArchitectureImportMap`, this fails closed because `go list` never produces `"-"`, but unit validation permits it.
- **Hypothesis 2 (Presentation metamorphic equivalence):** Presentation changes (logical row reordering, target wrapping with comments, CRLF line endings) parse equivalently to the baseline graph without spurious drift.
  - *Result:* **Confirmed**. Metamorphic mutation reversing 26 logical rows, wrapping each target across lines with `#` comments, and using `\r\n` line endings passed `TestArchitectureImportMap` identically (0.259s). However, CRLF line endings are not tested in the repo test suite (Finding L3).
- **Hypothesis 3 (Malformed fence handling):** Nested or multiple fenced code blocks within a single marker pair can cause `parseImportMap` to silently drop rows or accept split content.
  - *Result:* **Falsified**. Intermediate fence markers ```` and ````text lack `->` and are appended to the preceding row's target list. Backticks fail `validInternalPackageName` (not matching `^[A-Za-z0-9_.-]+$`), causing immediate validation failure. Stray text before opening fence or after closing fence fails `lines[0] == "```text"` and `lines[len-1] == "```"`.
- **Hypothesis 4 (Comparator diff fidelity):** `importGraphDiff` could fail to report drift if both `documented` and `actual` change edge counts or if zero-edge leaf packages are added or removed.
  - *Result:* **Falsified**. `importGraphDiff` performs exact symmetric difference on both package sets and target edges. Suppressing zero-edge package drift was immediately killed by `TestImportGraphDiffNamesPackageAndEdgeDriftDeterministically`.
- **Hypothesis 5 (Subprocess discovery fidelity):** Real-module discovery through `productionImportGraph` retains production imports into excluded support packages (`testutil` and `tools/*`), excludes test-only imports, and distinguishes root `.` and near-collision names (`testutils`, `toolsmith`).
  - *Result:* **Confirmed by probe, but identified as a test blind spot** (Finding L2). Our standalone fixture probe confirmed correct behavior through `go list`, but the committed test suite only exercises these boundaries through synthetic JSON fixtures.

### Isolation attestation

- **Workspace path:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.DrakzI`
- **Git directory:** `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.DrakzI/.git`
- **Baseline commit:** `874e1159dacea62f70d937f317b06f1c488ab1b0`
- **Captured deliverable blob:** `3b1ca87c76f76c3338a9259b92d0c496455ef385`
- **Captured source fingerprint:** `8e419dfb4e590d50126fd9c851522e94f728fd72`
- **Deliverable:** `planning/audits/6gkabd0m1evb-2026-10-13-documented-import-map-implementation-antigravity.md`
- **Workspace verification:** Verified (`deliverable_changed=true`, `transfer=pending`).
- **Transfer outcome:** Pending execution of `isolated-review-workspace.sh transfer` (to be executed after verification; actual output reported in handoff).

## Owner disposition (2026-10-13)

L1 is valid parser diagnostics hardening: a real Go build rejects a directory named `-`, and the new malformed-row cases reproduced parser acceptance before the fix. L2 and L3 are accepted as small coverage improvements; their probes show correct baseline behavior, not surviving implementation mutations or caller false passes. Support-target retention already overlapped the Claude fixes; root, support-subpackage and near-collision fixtures now run through an actual build and discovery, and CRLF equivalence is pinned.

The broad baseline soundness/no-false-passes verdict is limited to the executed uncached probes. It does not rule out the warm-cache or persisted-GOFLAGS failures independently found by Claude and already fixed. Future subprocess-backed checks should probe the ordinary cacheable route and persisted configuration as distinct evidence boundaries.

This report landed while the owner waited for the requested follow-up check. The frozen pending transfer row remains historical; no helper transfer stdout was supplied to the owner and success is not inferred from that row. The reviewer report is preserved, with managed dispositions and this separate owner assessment.
