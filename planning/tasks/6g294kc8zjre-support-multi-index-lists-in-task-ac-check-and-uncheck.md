---
schema: 1
id: 6g294kc8zjre
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: Accept comma-separated and repeatable criteria indices in task ac --check/--uncheck (e.g. --check 1,2,4) to eliminate agent loops
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, task, ac, ergonomics, agents]
created: "2026-08-21"
---
# Support multi-index lists in task ac --check and --uncheck

## Objective

Enhance `tskflwctl task ac <slug>` so `--check` and `--uncheck` accept lists of 1-based criteria indices (e.g. `--check 1,2,4` or `--check 1 --check 2`) rather than only a single integer (`--check <N>`). This allows agents and human operators to flip multiple acceptance criteria in a single atomic command without scripting client-side shell loops or issuing redundant tool calls.

## Context & Motivation

In current taskflow implementations, `tskflwctl task ac <slug>` supports:
- `task ac <slug> [--list]` — reads and lists numbered acceptance criteria.
- `task ac <slug> --check <n>` — ticks a single checkbox at 1-based index `n`.
- `task ac <slug> --uncheck <n>` — clears a single checkbox at 1-based index `n`.

When an agent finishes multiple acceptance criteria during a work step (e.g., implementing criteria 1, 2, and 4), it currently faces significant friction:
1. **Tool-Call / Process Overhead:** The agent must issue separate shell commands (`tskflwctl task ac foo --check 1`, `tskflwctl task ac foo --check 2`, `tskflwctl task ac foo --check 4`), incurring multiple process launches, disk reads, atomic file rewrites, and `updated_at` timestamps.
2. **Fragile Shell Loops:** Agents frequently write ad-hoc shell loops (`for i in 1 2 4; do tskflwctl task ac foo --check $i; done`), introducing unnecessary bash complexity and potential failure modes if one iteration fails halfway through.
3. **Write Amplification & fsnotify:** Flipping $K$ items individually causes $K$ atomic file renames and $K$ filesystem change events, whereas a batch update should be a single atomic transaction.

Supporting list inputs directly on `--check` and `--uncheck` solves this cleanly while remaining 100% backward compatible with existing single-index calls like `--check 3`.

## Proposed CLI Interface

### Syntax
- Comma-separated list:
  ```bash
  tskflwctl task ac <slug> --check 1,2,4
  tskflwctl task ac <slug> --uncheck 1,3
  ```
- Repeatable flag:
  ```bash
  tskflwctl task ac <slug> --check 1 --check 2 --check 4
  ```
- Single index (backward compatible):
  ```bash
  tskflwctl task ac <slug> --check 1
  ```

### Flag Definition
- Replace `IntVar(&check, "check", 0, ...)` with `IntSliceVar(&check, "check", nil, "check criteria at 1-based indices (comma-separated or repeatable)")`.
- Replace `IntVar(&uncheck, "uncheck", 0, ...)` with `IntSliceVar(&uncheck, "uncheck", nil, "uncheck criteria at 1-based indices (comma-separated or repeatable)")`.
- Maintain mutual exclusivity:
  - `--check` and `--uncheck` remain mutually exclusive.
  - `--list` remains mutually exclusive with `--check` and `--uncheck`.

### Validation & Edge Cases
1. **Empty list:** If `--check` or `--uncheck` is provided with no elements or empty input, return `domain.ErrValidation` (exit code 11).
2. **Bounds checking:** All indices must satisfy $1 \le \text{index} \le \text{total criteria count}$. If any index is $\le 0$ or $> \text{len(boxes)}$, fail fast with `domain.ErrValidation` (exit code 11) naming the invalid index before modifying any state.
3. **Duplicates & Ordering:** Duplicate indices (e.g. `--check 1,1,2`) are deduplicated. Non-sequential or reverse-order indices (e.g. `--check 4,1`) are processed deterministically.
4. **Missing AC Section:** If the task has no `## Acceptance criteria` section, fail fast with `domain.ErrValidation` (exit code 11).

### Idempotency & Output Reporting
- **All Already in Target State:** If all requested criteria already match the requested state (e.g. `--check 2` when criterion 2 is already checked), return without modifying the file or bumping `updated_at`.
  - Human output: `• criterion 2 is already checked` (or `• 2 criteria are already checked`).
- **Partial or Full Mutation:** If at least one criterion changes state, perform a single atomic write via `store.EditBody`.
  - Human output:
    - If all changed: `✔ checked <slug>` / `✔ unchecked <slug>`
    - If partially changed: `✔ checked <slug> (2 flipped, 1 already checked)`
  - JSON output (`--json`): Emits `wire.TaskMutationEnvelope` with the updated task metadata and resulting full body (matching standard mutation envelopes).

## Architecture & Layering Strategy

1. **Domain Layer (`internal/domain/body.go`)**:
   - Introduce `SetAcceptanceCriteria(body string, indices []int, checked bool) (newBody string, flippedCount int, err error)`.
   - Iterate through identified checkboxes in a single pass, deduplicating target indices and validating bounds.
   - For each target checkbox whose state differs from `checked`, flip the marker `[ ]` $\leftrightarrow$ `[x]`.
   - If `flippedCount == 0`, return `(body, 0, nil)`.
   - Maintain `SetAcceptanceCriterion(body, n, checked)` as a thin wrapper over `SetAcceptanceCriteria(body, []int{n}, checked)` or retain for backward compatibility.
   - Extend `internal/domain/body_test.go` and fuzz tests.

2. **Core Layer (`internal/core/service_task.go`)**:
   - Add `SetAcceptanceCriteria(slug string, indices []int, checked, dryRun bool) (domain.Task, string, int, bool, error)`.
   - Reads task, invokes `domain.SetAcceptanceCriteria`.
   - If no checkboxes flipped, return `changed = false` without calling `store.EditBody`.
   - If checkboxes flipped, call `s.store.EditBody(slug, newBody, false, s.now(), dryRun)` and return `changed = true`.

3. **CLI Layer (`internal/cli/task.go`)**:
   - Update `newTaskAcCmd` flag bindings to use `IntSliceVar`.
   - Route parsed slice to `app.Svc.SetAcceptanceCriteria`.
   - Format human status messages for single and multi-criterion flips and idempotency notices.
   - Ensure `--json` continues emitting `render.TaskMutationJSON`.

4. **Documentation & Generated Artifacts**:
   - Update CLI help text, examples, and `README.md`.
   - Run `docgen` to synchronize `docs/cli/tskflwctl_task_ac.md`.

## Acceptance criteria

- [ ] `tskflwctl task ac <slug> --check 1,2,4` ticks criteria 1, 2, and 4 in a single atomic file write.
- [ ] `tskflwctl task ac <slug> --check 1 --check 2` works identically to comma-separated lists.
- [ ] `tskflwctl task ac <slug> --uncheck 1,3` clears criteria 1 and 3 in a single atomic file write.
- [ ] Single-index invocations like `task ac <slug> --check 3` continue to work without breaking changes.
- [ ] Passing duplicate indices (e.g. `--check 1,1,2`) deduplicates cleanly and flips each checkbox once.
- [ ] Out-of-bounds indices (e.g. `--check 0` or `--check 99` on a 3-item task) fail validation with exit code 11 before writing.
- [ ] Mutually exclusive combinations (`--check` + `--uncheck`, `--list` + `--check`, `--list` + `--uncheck`) fail fast with validation errors.
- [ ] Partial flips (some criteria already checked, some unchecked) flip only the unchecked criteria and report count accurately.
- [ ] Full no-op flips (all requested criteria already in target state) perform no disk writes and do not bump `updated_at`.
- [ ] `--dry-run` validates inputs and previews the mutation without writing.
- [ ] `--json` returns the standard `task_mutation` envelope matching schema version 1.40.
- [ ] Comprehensive unit, integration, and golden tests pass with race detector enabled (`go test -race ./...`).
- [ ] Generated CLI documentation (`docs/cli/tskflwctl_task_ac.md`) is updated with zero docgen drift.

## Test Strategy & Edge Cases

- **Unit Tests (`internal/domain/body_test.go`)**:
  - Empty criteria list, single index, multiple sorted indices, multiple unsorted indices (`--check 3,1`).
  - Duplicate indices (`1, 2, 1, 2`), partial flips, all already flipped.
  - Out of bounds (`0`, negative, index $> \text{total criteria}$).
  - Bodies with code fences containing sample checkboxes (ensure fences are ignored).
  - Bodies with missing `## Acceptance criteria` header or empty section.
  - CRLF line endings preserved/handled properly.
- **Integration Tests (`internal/cli/task_ac_test.go`)**:
  - Test CLI flag parsing with comma separation (`--check 1,2`).
  - Test CLI flag parsing with repeatable flags (`--check 1 --check 2`).
  - Test CLI `--uncheck 1,2` clearing checkboxes.
  - Test CLI partial idempotency messaging.
  - Test CLI `--json` payload verification.
  - Test CLI `--dry-run` non-mutating execution.

## Out of Scope

- String / regex pattern matching against criterion text (index-based matching remains the only supported approach).
- Re-ordering or deleting acceptance criteria via CLI.
- Adding new acceptance criteria checkboxes (handled via `task append` or body editing).
