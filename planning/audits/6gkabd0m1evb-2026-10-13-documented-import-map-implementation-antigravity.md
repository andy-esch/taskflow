---
schema: 1
id: 6gkabd0m1evb
bucket: open
area: documented-import-map-implementation-antigravity
date: "2026-10-13"
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

Awaiting independent review. No findings or verdict have been authored by the owner.
