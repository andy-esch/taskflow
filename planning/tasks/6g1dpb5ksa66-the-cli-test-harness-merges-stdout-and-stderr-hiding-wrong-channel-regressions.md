---
schema: 1
id: 6g1dpb5ksa66
status: completed
epic: 21-code-quality-architecture-hardening
description: runRoot/runRootRC use one buffer for both streams, so moving diagnostics from stderr to stdout leaves the whole internal/cli suite green despite breaking -o name and --json consumers.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [testing, cli]
created: "2026-08-18"
updated_at: "2026-09-02"
started_at: "2026-09-02"
completed_at: "2026-09-02"
---

# The CLI test harness merges stdout and stderr, so no test can catch a wrong-channel regression

## The defect

`runRoot` and `runRootRC` both do `NewRootCmd(…, &out, &out)` plus `SetOut(&out)` /
`SetErr(&out)` — **one buffer for both streams**. Every assertion in `internal/cli` is
therefore blind to which stream output landed on.

Demonstrated: changing `render.ProblemsHuman(app.ErrOut, …)` to `app.Out` in
`internal/cli/listmode.go` leaves the **entire `internal/cli` suite green**. That is a real
regression — `FileProblem` lines on stdout corrupt `research list -o name` and any `--json`
consumer, and `internal/cli/problems.go` documents stderr as the contract precisely because
of it.

The stream discipline is a stated contract ("diagnostics go to stderr, matching the list
commands — scripts that capture stderr for problems must see them on one consistent
stream"), and nothing enforces it.

## Why it matters beyond research

This weakens every test in the package, not just the research ones. It came up because a
research test (`TestResearchList_NonIDLedFileIsReported`) asserts on the merged buffer and
so its claim is channel-agnostic; another (`TestResearchShow_AmbiguousWordingIsNotResearchs`)
is explicitly honest about checking both — but honesty about a blind spot doesn't remove it.

## Acceptance criteria

- [x] The harness exposes stdout and stderr separately, with the merged view still available
      for tests that legitimately don't care.
- [x] At least one test pins the contract directly: diagnostics on stderr, payload on
      stdout, verified by the mutation above FAILING.
- [x] Existing assertions migrated to the right stream where the distinction matters —
      notably anything checking a `FileProblem` or an error message.

## Out of scope

- Rewriting every existing assertion; the point is to make the distinction POSSIBLE and
  pin the contract, not to churn the whole package.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Found by an independent adversarial test-quality review (mutation testing), 2026-08-18
