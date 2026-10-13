---
schema: 1
id: 6g63hjm7cp6w
status: completed
epic: 21-code-quality-architecture-hardening
description: Verify the focused architecture guide’s exact production internal import map against default-host go list metadata.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, docs, go, dx]
created: "2026-09-02"
depends_on: [6g6x7e2ef37r]
updated_at: "2026-10-13"
audit_sources: [2026-09-28-arch-hexagonal-boundaries]
started_at: "2026-10-13"
completed_at: "2026-10-13"
---
# Test the documented import graph instead of date-stamping a manual review

## Objective

Verify the descriptive import map in
[boundaries](../../docs/architecture/boundaries.md#package-map) against real production
imports. Replace its presentation/process shorthand with exact internal package paths;
do not confuse an observed edge with permission to add that edge.

The focused guide owns the map, `.golangci.yml` owns executable import policy, and the
proposed dependency-policy ADR owns the remaining decisions. Those roles remain separate.
The checker runs in the ordinary test suite and reports concrete drift instead of
requiring another hand-review date.

## Coverage decision (2026-10-13)

- Source rows: packages selected by `go list ./internal/...` with non-test Go or cgo
  files, except `internal/testutil` and the `internal/tools` support namespace.
  Include packages with no internal imports, so a new leaf package cannot disappear.
  Wildcard-skipped directories and nested modules are not independently inventoried;
  fail if a covered source imports a non-support internal target without a source row.
- Targets: direct same-module `internal/` imports only. A production import into an
  excluded support package still appears as an edge; exclusions apply to source rows.
- Use ordinary native/default-build `go list` metadata, with workspace mode disabled,
  process/persisted GOFLAGS overridden and module edits forbidden. Do not claim a custom-tag,
  cross-platform union, or race-instrumented build inventory.
- Exclude TestImports, XTestImports, transitive Deps, standard library and third-party
  packages. Fail on malformed/incomplete metadata rather than comparing partial facts.
- Parse one explicitly marked `internal-import-map:v1` block. Rows use literal
  internal-relative paths, commas and an explicit no-import token; permit whitespace,
  ordering, wrapped lists and comments, but reject duplicates or unexplained shorthand.
- Compare both package inventory and edges, with deterministic added/removed diagnostics.
  No new dependency policy, generator framework or documentation platform is introduced.

## Acceptance criteria

- [x] A marked internal-import-map:v1 block in docs/architecture/boundaries.md
  exactly matches default-host production internal package rows and direct
  internal Imports from go list, including packages with no internal edges.
- [x] Drift reports deterministic added/removed packages and edges with
  owner/remediation guidance; malformed or incomplete metadata fails closed, and
  meaningful drift regressions are tested.
- [x] TestImports, XTestImports, transitive Deps, stdlib and third-party edges
  are excluded. Source exclusions are internal/testutil and the internal/tools
  namespace; production imports into them remain visible. No custom-tag or
  cross-platform union is claimed.
- [x] The focused map links to its executable checker instead of a review date
  or pending task, keeping descriptive facts separate from depguard policy and
  named exceptions.
- [x] The marked-block grammar is documented and tested for harmless whitespace,
  ordering, wrapping and comments; missing/malformed blocks, duplicate
  declarations and unsupported shorthand fail clearly.

## Out of scope

- Changing `.golangci.yml` dependency policy — this task verifies what the doc claims
- The adapter-edge classification table and named local exceptions; those are judgements, not derivable facts
- Any similar treatment of the package-role table above the graph block

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Thread [Make documentation layered, executable, and agent-navigable](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- Follows [Restructure the architecture documentation into focused guides](6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md)

The [composition-isolation task](6gcwcf8rxe72-isolate-cli-composition-wiring-and-enforce-controller-boundaries.md)
now owns audit M1's default-deny policy and L1's recursive coverage. This task independently verifies
the descriptive graph, including the new `cli/ports` and `appwiring` roles. It remains complementary
to depguard and to the [proposed policy ADR](6gg7e59mm68g-record-the-hexagonal-dependency-policy-and-composition-exceptions-in-an-adr.md).

## Architecture extraction handoff (2026-10-12)

The canonical descriptive map now lives in [boundaries](../../docs/architecture/boundaries.md#package-map).
The role table and named exceptions remain human judgments. The map retains explanatory shorthand
for presentation/process edges and is explicitly not a complete machine-generated inventory.
Choose and document its covered edge set/normalization before an exact go-list comparison; do not
silently bless omitted edges or compare shorthand labels literally with package imports.
The new architecture navigation check only checks routes and must not be mistaken for this task's
import-drift enforcement. Update the focused guide when the checker lands; keep the root index short.

## Implementation and validation (2026-10-13)

The marked boundaries map now covers 26 production packages and 78 direct internal edges in the observed host build, including four leaf packages. The ordinary architecture test compares both source rows and edges, prints deterministic +/- drift, and routes remediation to the focused guide and actual depguard/named-exception policy. No import policy, product behavior, Go dependency or machine-contract revision changed.

Parser regressions cover ordering, whitespace, comments and wrapped lists, plus missing/duplicate markers, malformed rows and duplicate declarations. Metadata fixtures pin exclusion of test/transitive/external imports, test-only/support sources, retention of production imports into support packages, cgo/leaf/root records and fail-closed incomplete/erroneous/truncated streams. A real throwaway module verifies that broken ambient GOWORK and GOFLAGS=-find cannot erase facts, and its module files remain unchanged.

Owner probes removed core -> id from the map and added a blank editor -> domain production import. The latter passed go build ./internal/editor; each probe failed TestArchitectureImportMap with the intended named edge. Both were restored and the focused race suite passed. Full just test (race), just lint (0 issues), just build, just docs-check, planning lint --links and whitespace checks passed. The two new test files also received explicit untracked-file whitespace checks.

At initial handoff the task remained in progress for focused adversarial review of
coverage/parsing; checked criteria reflected owner evidence, not an independent
verdict. No branch, commits or pushes had been made at that checkpoint.

## Independent review handoff (2026-10-13)

Prepared open, zero-finding audits for [Claude](../audits/6gkabd0b0bg6-2026-10-13-documented-import-map-implementation-claude.md) and [Antigravity](../audits/6gkabd0m1evb-2026-10-13-documented-import-map-implementation-antigravity.md). Claude leads discovery/coverage fidelity; Antigravity leads parser/comparator false passes, with bounded cross-checks. Both must capture this dirty/untracked implementation in separate clones, exercise the real native guard, kill a new compiler-valid implementation mutation, verify named evidence and transfer only their assigned audit. The previous review's false-pass and unverified-inventory lessons informed these requirements; no reviewer-specific finding quota or assumed verdict is imposed.

At handoff, review had not run. Each brief and the captured implementation are the
reviewer's independent baseline; do not edit an assigned audit while its reviewer
is working. The task remains in progress pending findings triage.

## Claude review triage (2026-10-13)

Claude's three findings are accepted, with owner-authored regressions reproducing each
failure before fixes. M1 now records the internal tree and module manifests in the test
process so its cached pass is invalidated by production-source drift. An end-to-end
throwaway-module regression first proves a warm cache hit, then requires an import edit,
new leaf and removed leaf to fail through the actual checker without `-count=1`.

L1 uses a non-empty GOFLAGS=-mod=readonly override instead of the empty value that falls
back to go env -w configuration. The real module includes competing build-tag imports and
a persisted GOENV file. Validation-before-filter tests also start with a valid row, so
an empty-graph refusal cannot mask ignored erroneous support/test-only metadata.

L2 refuses non-support internal targets with no covered source row and reports the
target/importer and wildcard-selection limitation. Compiler-valid `_hidden`, testdata
and nested-module fixtures pin this refusal without expanding the inventory into a
transitive or all-platform union. The guide and coverage decision now state that limit.

Removing cache-input recording, restoring empty GOFLAGS, moving metadata validation
after filtering, and disabling uncovered-target refusal each compiled and failed its
intended regression. Every probe was restored. Final full just test (race), just lint
(0 issues), just build, just docs-check, planning lint --links and whitespace checks
passed. All three findings are marked fixed and Claude's audit is closed.
The exact CI race/coverage invocation also passed with its coverage output outside
the checkout: `go test -p 4 -race -coverprofile=/private/tmp/taskflow-import-map-coverage.out ./...`.
At this checkpoint Antigravity's report was still pending; Claude's findings and
owner verification did not stand in for that independent review.

## Antigravity follow-up and closeout (2026-10-13)

Antigravity's report landed while the owner waited for the requested follow-up check.
Its Low parser-token finding is valid: Go rejects a directory named `-`, while the
parser previously accepted `- -> core` and `- -> -`. Both malformed-row regressions
failed before the source-token rejection was added. CRLF equivalence is now explicit.

The real-module coverage suggestion partly overlaps the Claude fixes rather than
demonstrating another implementation bug. Extended the compiler-valid fixture with
the internal root, support subpackages and testutils/toolsmith near-collisions; actual
discovery must retain those rows/target edges without admitting support-source rows.
The review's broad soundness claim is limited to its uncached probes; it does not
disprove Claude's separate warm-cache and persisted-configuration findings.

Final focused/full race suites, CI-equivalent coverage, lint (0 issues), build,
docs-check and planning link lint passed. Both audits are closed with all findings
fixed; the task is completed. The healthy Thread still exposes selective package
contracts and the dependency-policy ADR as its two eligible next tasks. No additional
follow-up ticket or policy/wire change was needed for these in-scope review refinements.
