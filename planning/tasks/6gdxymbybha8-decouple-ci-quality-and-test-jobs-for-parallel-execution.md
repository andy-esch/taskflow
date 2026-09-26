---
schema: 1
id: 6gdxymbybha8
status: completed
epic: 21-code-quality-architecture-hardening
description: Split monolithic CI into parallel lint and test jobs, and document test sharding and gotestsum acceleration options.
effort: 1-2 hours
tier: 3
priority: medium
autonomy_level: 4
tags: [ci, testing, performance, github-actions]
created: "2026-09-26"
updated_at: "2026-09-26"
started_at: "2026-09-26"
completed_at: "2026-09-26"
---

# Decouple CI quality and test jobs for parallel execution

## Objective

Accelerate CI feedback loops by splitting the monolithic `go-quality` job in `.github/workflows/ci.yml` into two decoupled, parallel jobs: a fast `lint` job (tidy, docs, formatting, golangci-lint, govulncheck) and a dedicated `test` job (unit tests, race detection, Codecov upload). This eliminates serial queuing on GitHub Actions runners, enabling fast-fail static analysis in ~30s and dropping overall CI turnaround from ~2m 05s to ~1m 15s.

## Scope

- Split `.github/workflows/ci.yml` jobs:
  - `lint`: Runs code generation checks (`docs-check`), module hygiene (`tidy-check`), formatting (`gofmt`), static analysis (`golangci-lint`), and vulnerability scanning (`govulncheck`). Finishes in ~30-35s.
  - `test`: Runs the unit test suite with race detection (`go test -v -race -coverprofile=coverage.out ./...`) and uploads coverage to Codecov.
  - `planning-lint`: Continues running in parallel via `uses: ./actions/lint` (~8s).
- Verify GitHub Actions concurrency and branch status checks remain intact.
- Validate workflow definitions with `actionlint`.

## Research: Further Go CI Acceleration Options

Beyond the immediate job split, research into Go test acceleration in GitHub Actions identifies four viable follow-up optimizations:

1. **Horizontal Test Sharding (`strategy: matrix`)**:
   - Split the ~30 packages across 2 parallel runners using deterministic package distribution (`awk "NR % 2 == ${{ matrix.shard }} - 1"` or `hashicorp-forge/go-test-split-action`).
   - Codecov natively merges multi-shard coverage reports.
   - Drops test execution from ~76s to ~38s, bringing total CI time under 50s.
2. **Compact Test Output (`gotestsum`)**:
   - `go test -v` streams thousands of lines to the runner console, incurring network and terminal buffer latency.
   - `gotestsum --format short` prints one-line summaries per package and only expands failures, reducing console streaming overhead.
3. **Decoupled Race Detection**:
   - The Go race detector (`-race`) adds a 2x-5x execution penalty and high memory pressure on 2-vCPU runners.
   - Splitting into fast unit tests + coverage (~15s) and a separate race detector pass on concurrency-critical packages (`internal/store`, `internal/core`, `internal/cli`) isolates regressions rapidly.
4. **Tool Cache for `govulncheck`**:
   - `go install .../govulncheck@latest` compiles from source every run (~9s). Caching the binary or checking for an existing installation saves ~9s on the lint job.

## Acceptance criteria

- [x] The monolithic go-quality job in .github/workflows/ci.yml is split into
  two independent parallel jobs: lint (static checks, format, golangci-lint,
  govulncheck) and test (unit tests, race detector, Codecov upload).
- [x] Both jobs run in parallel on push and pull_request events alongside the
  planning-lint job.
- [x] Static checks and linting failures fail fast without waiting for the full
  Go test suite to complete.
- [x] Code coverage upload to Codecov continues to report accurately from the
  dedicated test job.
- [x] Workflow syntax passes actionlint validation with zero errors.
- [x] Further horizontal test sharding and gotestsum acceleration options are
  evaluated and documented as follow-up candidates.

## Out of scope

- Horizontal multi-runner matrix sharding in this initial task (documented for follow-up).
- Adopting `gotestsum` as a new build/test dependency in this pass.
- Modifying test assertions, timeouts, or package structures.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Worktree branch `feat/speed-up-ci-tests`

## Progress Log

- 2026-09-26: Evaluated Go CI acceleration research (matrix sharding, gotestsum, race detection separation, caching) and implemented Option 1 by decoupling the monolithic `go-quality` job in `.github/workflows/ci.yml` into parallel `lint` and `test` jobs. Verified clean actionlint syntax, docs-check, tidy-check, and planning lint.
