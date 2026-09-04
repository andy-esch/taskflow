---
schema: 1
id: 6g697mp8s4tx
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Surface a degraded task graph on status and lint instead of only at the mutation it refuses, reconciling lint's advisory/exit-zero treatment of legacy dependency fields.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, graph, lint]
created: "2026-09-02"
updated_at: "2026-09-04"
started_at: "2026-09-03"
completed_at: "2026-09-03"
---
# Report graph degradation in status and lint

## Objective

A degraded repository task graph is latched at mutation time: `task complete` refuses
with "repository task graph is degraded: 3 legacy dependency field occurrence(s) remain;
run `tskflwctl task depend migrate`". The message is good and the latch is defensible —
mutating a graph you cannot trust is worse than stopping — but the degradation surfaces
*mid-operation*, on an unrelated task, in a repository the caller may not own. The
occurrences had been there for weeks.

Make the same fact visible on the surfaces where it can be fixed calmly.

## Already landed

`board` prints a graph-health warning naming the cause and the remedy, and `board --json`
carries a `graph` object. That covers the surface an agent reads first. This task is the
rest: `status`, and the lint contradiction.

## The lint concern was narrower than it looked

`lint` reporting "all planning entities and dependency links pass lint" on a repository
the graph guard rejects is the sharp end of this finding, but a repository-level verdict
added naively is wrong, and the tests say so:

- A **resolvable** legacy dependency field is already reported, deliberately as an
  `advisory` with **exit zero** — pinned by
  `TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero` in `internal/cli/lint_test.go`.
  Appending a blocking repository-level issue flips that exit code and breaks the
  contract.
- Initial inspection suspected silence because `dependencyLintIssues`
  (`internal/core/service.go`) skips `ProblemLegacyMissing` and
  `ProblemLegacyAmbiguous`. Implementation proved that suspicion false: the grouped
  `LegacyDependencyDiagnostic` pass immediately below owns those renderings. The problem
  was an inaccurate ownership comment plus missing regression coverage, not missing
  runtime behavior.

So the work reconciles two existing designs rather than bolting on a repository-level lint
warning: preserve lint's severity model and pin the skipped classes to their actual grouped
owner without printing any defect twice.

## Scope

- `status`: a graph-health line, matching the one `board` already prints.
- `lint`: report degradation the guard would refuse on, honouring the existing
  advisory/exit-zero treatment of resolvable legacy fields.
- Cover the `ProblemLegacyMissing` / `ProblemLegacyAmbiguous` classes specifically —
  confirm first whether their "established renderings" actually fire, since that
  assumption is what left the gap.

## Acceptance criteria

- [x] `status` names a degraded or broken task graph, its cause, and the remedy, matching
      the wording `board` already uses.
- [x] `lint` cannot report a clean repository while the graph guard would refuse a
      mutation on it; a test asserts the two agree.
- [x] The resolvable-legacy-field case keeps its advisory severity and exit zero
      (`TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero` still passes unchanged).
- [x] A legacy field that is missing or ambiguous is reported by `lint`, once — not twice,
      and not silently.
- [x] `--json` carries the graph verdict on every surface that renders it.

## Related

- Audit [2026-09-01-cli-ergonomics-from-an-agent-session](../audits/6g5xyxhhc8p5-2026-09-01-cli-ergonomics-from-an-agent-session.md), finding H3

## Implementation outcome (2026-09-03)

`core.Summary` now derives task counts, in-flight work, and graph health from one explicit `TaskGraphSource` snapshot. Cross-space summary adapters must provide that graph capability rather than relying on optional filesystem-shaped discovery. Human `status`, `status --all`, the TUI overview/atlas, and their shared JSON summary report the same non-healthy verdict and remedy as `board`; healthy JSON omits the optional `graph` object. Dashboard graph warnings remain informational, while unreadable files retain their existing partial-result failure semantics.

The suspected lint gap was not present: grouped `LegacyDependencyDiagnostic` rendering already owns missing and ambiguous references while the raw graph-problem loop skips them to avoid duplication. The misleading comment was corrected and regression coverage now proves missing/ambiguous references appear exactly once, unsafe cases remain blocking, and exactly resolved legacy debt remains advisory/exit-zero without ever claiming the repository passes lint. The same fixture proves an ordinary graph mutation plan refuses that degraded snapshot.

Machine schema 1.60 adds the optional summary graph verdict. CLI reference, JSON schema comments, and golden contracts were regenerated. Full race tests, focused adapter/core/CLI/TUI/wire tests, golangci-lint, module tidy check, diff check, current-repository `status --json`, and current-repository `lint --json` are clean.

A second-pass portability review found one intentionally separate seam: board/status still collapse
neutral graph load problems into filesystem-shaped `FileProblem` results, losing task identity for a
pathless adapter. [Preserve portable load diagnostics in board and status](6g6jqqcdehne-preserve-portable-load-diagnostics-in-board-and-status.md)
tracks that work after the shared repository-lint diagnostic vocabulary lands.

## Adversarial review closeout (2026-09-04)

Two independent reviews converged on a real weakness in the legacy-diagnostic de-duplication test.
The test now asserts the core Lint result contains exactly one blocked_by issue for the dependent
task; a restored competing raw-problem owner produces three issues and fails the targeted test. All
additional Claude findings were accepted: lint and graph-query diagnostics recommend migration only
for resolved or present-empty legacy fields, while missing, ambiguous, and unsafe fields prescribe
direct repair plus lint; non-healthy summary graph objects now exercise the real JSON Schema
validator; and compile-time assertions independently pin both `PlanningSummarySource` capabilities.

The review loop also improved. Generated briefs now require a frozen handoff, an isolation
attestation, the exact mutation each regression claims to kill, non-default optional wire values in
semantic validation, execution of advertised repair commands, and coordinated mutations that
challenge architectural invariants rather than only their nearest call site. Full `go test -race`,
repeated focused race tests, golangci-lint, tidy/doc checks, planning/audit lint, and diff checks
pass.
