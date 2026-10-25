---
schema: 1
id: 6gpxc9zegzzd
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Eliminate synchronous 530ms TUI cursor-blink test sleeps, quadratic domain inline-code scanning, and sequential subprocess bottlenecks across CI tests.
effort: 2-4 hours
tier: 2
priority: high
autonomy_level: 4
tags: [ci, testing, performance, tui, domain]
created: "2026-10-24"
updated_at: "2026-10-24"
started_at: "2026-10-24"
---

# Eliminate CI test suite bottlenecks and parallelize slow package tests

## Objective

Reduce the CI `Tests` job runtime (`go test -p 4 -race -coverprofile=coverage.out ./...`, which grew to 7m 02s in GitHub Actions run `40091666145` / job `120865430802` on PR #329) back under 1 minute by eliminating synchronous 530ms cursor-blink timer sleeps in `internal/tui` staged-creation tests, fixing quadratic inline-code backtick scanning and redundant full-corpus parses in `internal/domain`, and parallelizing isolated subprocess and unit tests across the slowest packages.

## Findings & Actionable Work Slices

### CI Baseline (Job `120865430802` — 6m 46s `go test` wall-clock)

| Package | CI Time | Local (`-race`) | Primary Bottleneck |
| --- | ---: | ---: | --- |
| `internal/tui` | 312.99s | 208.63s | Synchronous 530ms `cursor.Blink()` timer waits in `settleCreationCommands` across ~350 creation subtests; `0` `t.Parallel()` calls |
| `internal/domain` | 51.47s | 16.01s | Quadratic `findingParagraphBreakRe` scanning in `blankFindingInlineCode`, 3 sequential full-corpus walks in `nearmiss_corpus_test.go`, and 980-iteration wrapping loop in `finding_note_test.go` |
| `internal/tools/reviewworkspace` | 25.15s | ~18.00s | 22 sequential tests (including 21 subtests in `TestVerifyAndTransferRejectAWeakenedPushGuard`) spawning `git`, `rsync`, and `bash` subprocesses |
| `internal/store` | 11.74s | 9.01s | Hardcoded 50-100ms `time.After` negative concurrency sleeps (tracked in `6gppr2jh0j3x`) + 3 sequential subprocess tests |
| `internal/cli` | 11.29s | 6.80s | Sequential CLI integration tests (`0` `t.Parallel()`) |
| `internal/tools/architecturecheck` | 9.25s | 5.71s | `TestImportMapCacheInvalidatesProductionDrift` runs 9 child `go test -race` subprocesses sequentially |
| `cmd/tskflwctl` | 8.64s | 6.13s | Sequential CLI subprocess/golden tests |

### Slice 1: `internal/tui` Cursor Blink Settling (~313s → ~15-20s)

- `creationState.focus` in `internal/tui/staged_creation.go` calls `inputs[row].Focus()`, which invokes `charm.land/bubbles/v2/cursor.Model.Focus()` -> `cursor.Model.Blink()`.
- `cursor.Model.Blink()` blocks on `context.WithTimeout(..., m.BlinkSpeed)` where `BlinkSpeed` defaults to **530ms** before returning `cursor.BlinkMsg`. `textinput.Model.Update` on `tea.PasteMsg` also schedules a new `Blink()` command.
- `settleCreationCommands` in `internal/tui/staged_creation_helpers_test.go` executes `msg := cmd()` synchronously (sleeping 530ms wall-clock per focus/paste) before `creationWidgetMaintenance` discards the resulting `cursor.BlinkMsg`.
- Fix by avoiding the 530ms real-time wait during test command settling (e.g., configuring non-blocking or sub-millisecond cursor blink behavior for creation test inputs while keeping `TestCreationHiddenWidgetKeepsBlinkContinuation` and `TestCreationHarnessStopsMaintenanceAndBoundsWork` verifying blink continuation at a fast interval).

### Slice 2: `internal/domain` Inline-Code Scanner & Corpus Tests (~51.5s → ~3-5s)

- In `internal/domain/finding_status.go`, `blankFindingInlineCode` scans every backtick run and, when looking for a matching closing backtick run, evaluates `findingParagraphBreakRe.MatchString(source[pos:begin])` on a growing slice for every subsequent unmatched backtick in the document. Bounding the search by locating the next paragraph break once per open backtick eliminates the repeated regex work.
- In `internal/domain/nearmiss_corpus_test.go`, `TestNearMissFindingHeaders_LiveCorpusIsClean`, `TestFindingHeaderClassifierAcrossAllPlanningBodies`, and `TestNearMissFindingHeaders_DoesNotShadowCanonicalFindings` walk and re-parse `planning/audits/*.md` three separate times sequentially.
- In `internal/domain/finding_note_test.go`, `TestSetFindingNote_WrappingPreservesDocumentStructure` runs 980 iterations (`7 fragments x 2 pads x 70 widths`) with 5 full document parses per iteration.

### Slice 3: `internal/tools/reviewworkspace` & `internal/tools/architecturecheck` Subprocess Parallelism (~34.4s → ~6-8s)

- Mark isolated tests and the 21 `t.TempDir()` subtests in `TestVerifyAndTransferRejectAWeakenedPushGuard` (`internal/tools/reviewworkspace/review_workspace_test.go`) with `t.Parallel()`.
- In `internal/tools/architecturecheck/imports_cache_test.go`, drop `-race` from the child `go test` invocation inside `runArchitectureCheck` (which tests `GOCACHE` invalidation on a synthetic temp module) and run the 3 subtests of `TestImportMapCacheInvalidatesProductionDrift` in parallel.

### Slice 4: Targeted `t.Parallel()` Adoption

- Add `t.Parallel()` to stateless unit test functions in `internal/tui` and `internal/domain` so `go test` utilizes all available runner cores instead of running the tail packages single-threaded.

## Acceptance criteria

- [x] internal/tui staged-creation tests and settleCreationCommands no longer
  block on 530ms cursor.Blink() timers during focus and paste settling, while
  TestCreationHiddenWidgetKeepsBlinkContinuation and
  TestCreationHarnessStopsMaintenanceAndBoundsWork continue to verify blink
  maintenance behavior.
- [ ] blankFindingInlineCode in internal/domain avoids repeated paragraph-break
  regex scans across unmatched backtick runs, and internal/domain corpus and
  finding-note wrapping tests avoid redundant re-parses.
- [ ] Subprocess-heavy tests in internal/tools/reviewworkspace (including
  TestVerifyAndTransferRejectAWeakenedPushGuard subtests) and
  internal/tools/architecturecheck
  (TestImportMapCacheInvalidatesProductionDrift) run with t.Parallel(), and
  child cache-probe go test invocations in architecturecheck omit redundant
  -race instrumentation.
- [ ] Stateless unit tests in internal/tui and internal/domain use t.Parallel()
  where safe under the race detector.
- [ ] just build, just test, just lint, and ./bin/tskflwctl -C . lint pass with
  zero regressions and substantially reduced wall-clock test runtime.

## Out of scope

- Migrating `internal/store` negative concurrency `time.After` waits to `testing/synctest` (already tracked in task [6gppr2jh0j3x](6gppr2jh0j3x-adopt-go-1.24-1.26-idioms-in-error-mapping-tests-and-dependencies.md)).
- Multi-runner CI matrix sharding or changing `.github/workflows/ci.yml` job topology.
- Altering production TUI cursor blink visuals in interactive terminal sessions.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Prior CI split task [6gdxymbybha8](6gdxymbybha8-decouple-ci-quality-and-test-jobs-for-parallel-execution.md)
- Store `testing/synctest` modernization task [6gppr2jh0j3x](6gppr2jh0j3x-adopt-go-1.24-1.26-idioms-in-error-mapping-tests-and-dependencies.md)

## Progress Log

- 2026-10-24: Completed Slice 1 (`internal/tui` cursor-blink settling). Added `newCreationInput()` with configurable `creationBlinkSpeed` (1ms in package tests; 0/default ~530ms in production) and stopped unwrapped maintenance ticks in `settleCreationCommands`, dropping `go test -race -count=1 ./internal/tui` from `208.63s` to `37.58s` locally while keeping `TestCreationHiddenWidgetKeepsBlinkContinuation` and `TestCreationHarnessStopsMaintenanceAndBoundsWork` verifying real `cursor.BlinkMsg` continuation.
