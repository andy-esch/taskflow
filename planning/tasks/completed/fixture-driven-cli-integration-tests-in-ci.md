---
status: completed
epic: 20-cli-ux-and-ergonomics
description: Fixture-driven integration/e2e CLI tests (testscript, golden files, subprocess smoke); approach NOT yet chosen — options laid out for a decision
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [testing, ci, dx]
created: "2026-06-19"
updated_at: "2026-06-20"
started_at: "2026-06-20"
completed_at: "2026-06-20"
---
## Objective

Add fixture-driven integration / end-to-end testing for the CLI: run the
command surface against a known set of fixture planning files and assert
stdout/stderr/exit-codes/file-state. This complements the existing in-process
tests with real-CLI fidelity, reusable fixtures, and readable, doc-like cases.

## ⚠️ Decision needed — NOT yet opted into any approach

The owner has **not** chosen an approach. This task captures the options for a
decision; the first step is to pick one (or a combination), then implement.

## Where we are today

The suite already does a form of integration testing: `runRoot(t, args...)`
builds the cobra tree and runs it **in-process** against a `t.TempDir()` planning
repo (`setupRepo`/`freshRepo`). That covers most of the *logic*. What's missing is
real-binary fidelity, byte-stable output snapshots, and shared committed fixtures.

## Options (the ladder — cheapest/fastest first)

1. **Keep leaning on in-process `runRoot` tests** — fastest, already the backbone;
   gives coverage but not real-binary fidelity or output snapshots.
2. **Golden-file (snapshot) tests** — capture output to `testdata/*.golden`,
   compare, regenerate with `-update`. Ideal for the byte-stable contracts
   (`-o table`/`csv`, the `--json` envelopes). `charmbracelet/x/exp/golden` is
   **already in the module graph** (near-zero friction).
3. **`testscript` (txtar)** — `rogpeppe/go-internal/testscript`, the tool the Go
   team uses for `cmd/go`; `.txtar` scripts with inline fixtures + stdout/stderr/
   exit/file assertions, run either in-process (fast) or against a built binary.
   Reads like living docs. `go-internal` is **already a transitive dep** (would
   become a direct test import). The closest match to "fixture tasks + the CLI."
4. **Subprocess smoke** — build the real binary (once, in `TestMain`) and `os/exec`
   it against a committed `testdata/planning/` fixture. Only thing that catches
   `main.go` wiring / ldflags `--version` / signals. Worth a handful, not the whole
   suite. (Skip `bats`/shell harnesses — leaves Go's tooling for no gain.)

## Recommendation (pending the owner's call)

testscript (#3) as the primary add — promote today's `setupRepo` fixtures into a
committed `testdata/planning/` tree and write a handful of `.txtar` scripts
covering the lifecycle (create → list/project → transition → lint) and the
exit-code contract; add golden files (#2) for the rich-output contracts; one or
two subprocess smokes (#4) in CI for true end-to-end confidence. Pairs with the
docs drift-check.

## Decision & shipped (2026-06-20)

**Chosen: golden snapshots (#2) + subprocess smoke (#4). testscript (#3)
deferred** — the in-process `runRoot` suite already covers lifecycle *logic*, so
txtar would largely duplicate it; the real gaps were byte-stable output snapshots
and real-binary fidelity, which #2 and #4 fill directly. (testscript remains a
fine future add for living-doc lifecycle scripts — noted, not built.) Rolled a
~20-line dep-free golden helper instead of importing the charm lib (repo's
"no library for a 15-line job" ethos; full control of the diff).

- **Fixture:** committed `internal/cli/testdata/planning/` — 3 tasks (one per
  status), an epic, fixed 2026-01 dates so every snapshot is date-stable. Lints
  clean.
- **Golden (`golden_test.go` + `integration_golden_test.go`):** `assertGolden`
  with `-update`; 11 snapshots of the date-stable machine contract — `task/epic
  list|show --json`, `status --json`, `lint --json`, `-o csv`/`name`, and the
  self-description trio incl. **`schema --json-schema` (894 lines — the entire
  Draft 2020-12 contract pinned byte-for-byte)**. Drift fails with a regenerate
  hint; proven by mutating a golden → FAIL.
- **Subprocess smoke (`smoke_test.go`):** binary built once in `TestMain`
  (`-short`-skippable), then `--version` shape, not-found → **exit 10**, the
  `--json` **error envelope** (which only `main.go`'s `WriteError` emits — invisible
  in-process), a clean `list --json`, and a full **create→start→complete lifecycle**
  through the binary asserting file movement.
- **CI:** no workflow change — both run under the existing `go test -race ./...`.
  The in-process `runRoot` tests stay (unchanged), as scoped.

## Acceptance criteria (decision-gated)

- [x] Approach chosen: golden (#2) + subprocess smoke (#4); testscript (#3)
      deferred with rationale.
- [x] A committed `testdata/planning/` fixture tree exists
      (`internal/cli/testdata/planning/`).
- [x] The chosen harness runs in CI (`go test -race ./...`) and fails on
      regressions — output drift (golden), wrong exit code / wrong file state (smoke).
- [x] Lifecycle (smoke create→start→complete) + the exit-code contract (exit 10 +
      JSON error envelope) are covered.

## Out of scope

- Replacing the existing in-process `runRoot` tests (they stay).
- Shell-based (`bats`) harnesses.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Complements [[auto-generate-cli-reference-docs-with-a-ci-sync-check]] (both are
  CI-enforced contracts) and would lock the byte-stable output from
  [[consolidate-output-flags-into-output-and-columns]].
