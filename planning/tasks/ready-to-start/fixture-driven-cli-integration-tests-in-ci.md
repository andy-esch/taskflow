---
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Fixture-driven integration/e2e CLI tests (testscript, golden files, subprocess smoke); approach NOT yet chosen — options laid out for a decision
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [testing, ci, dx]
created: "2026-06-19"
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

## Acceptance criteria (decision-gated)

- [ ] Approach chosen (which of #2/#3/#4, or combination).
- [ ] A committed `testdata/planning/` fixture tree exists (or fixtures are inline
      per the chosen approach).
- [ ] The chosen harness runs in CI and fails on regressions
      (output drift / wrong exit code / wrong file state).
- [ ] At least the lifecycle + exit-code-contract paths are covered.

## Out of scope

- Replacing the existing in-process `runRoot` tests (they stay).
- Shell-based (`bats`) harnesses.

## Related

- Epic [[20-cli-ux-and-ergonomics]]
- Complements [[auto-generate-cli-reference-docs-with-a-ci-sync-check]] (both are
  CI-enforced contracts) and would lock the byte-stable output from
  [[consolidate-output-flags-into-output-and-columns]].
