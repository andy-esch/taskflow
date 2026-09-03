---
schema: 1
id: 6g697mp8s4tx
status: ready-to-start
epic: 30-threads-and-task-dependency-graphs
description: Surface a degraded task graph on status and lint instead of only at the mutation it refuses, reconciling lint's advisory/exit-zero treatment of legacy dependency fields.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, graph, lint]
created: "2026-09-02"
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

## The lint problem is narrower than it looks

`lint` reporting "all planning entities and dependency links pass lint" on a repository
the graph guard rejects is the sharp end of this finding, but a repository-level verdict
added naively is wrong, and the tests say so:

- A **resolvable** legacy dependency field is already reported, deliberately as an
  `advisory` with **exit zero** — pinned by
  `TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero` in `internal/cli/lint_test.go`.
  Appending a blocking repository-level issue flips that exit code and breaks the
  contract.
- The actual silence comes from `dependencyLintIssues` (`internal/core/service.go`), which
  skips `ProblemLegacyMissing` and `ProblemLegacyAmbiguous` on the grounds that they
  "already have established ordinary-lint/FileProblem renderings". Where those renderings
  do not fire, the defect is reported nowhere while the graph is still degraded.

So the work is to reconcile two existing designs, not to bolt on a new warning: decide
what a repository-level graph verdict means for lint's severity model and exit code, and
close the gap for the skipped classes without printing any defect twice.

## Scope

- `status`: a graph-health line, matching the one `board` already prints.
- `lint`: report degradation the guard would refuse on, honouring the existing
  advisory/exit-zero treatment of resolvable legacy fields.
- Cover the `ProblemLegacyMissing` / `ProblemLegacyAmbiguous` classes specifically —
  confirm first whether their "established renderings" actually fire, since that
  assumption is what left the gap.

## Acceptance criteria

- [ ] `status` names a degraded or broken task graph, its cause, and the remedy, matching
      the wording `board` already uses.
- [ ] `lint` cannot report a clean repository while the graph guard would refuse a
      mutation on it; a test asserts the two agree.
- [ ] The resolvable-legacy-field case keeps its advisory severity and exit zero
      (`TestLintResolvedLegacyDependencyIsAdvisoryWithExitZero` still passes unchanged).
- [ ] A legacy field that is missing or ambiguous is reported by `lint`, once — not twice,
      and not silently.
- [ ] `--json` carries the graph verdict on every surface that renders it.

## Related

- Audit [2026-09-01-cli-ergonomics-from-an-agent-session](../audits/6g5xyxhhc8p5-2026-09-01-cli-ergonomics-from-an-agent-session.md), finding H3
