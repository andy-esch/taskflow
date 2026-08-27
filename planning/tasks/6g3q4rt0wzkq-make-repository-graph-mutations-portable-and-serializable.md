---
schema: 1
id: 6g3q4rt0wzkq
status: completed
epic: 30-threads-and-task-dependency-graphs
description: Provide a cross-platform store-owned guard for authoritative repository graph read, validation, and write operations.
effort: 2-4 days
tier: 1
priority: high
autonomy_level: 2
tags: [threads, graph, storage, concurrency]
created: "2026-08-25"
updated_at: "2026-08-27"
started_at: "2026-08-27"
completed_at: "2026-08-27"
---
# Make repository graph mutations portable and serializable

## Objective

Provide one store-owned repository mutation boundary that makes final graph read, validation, and write indivisible for cooperating writers on every supported platform.

## Scope

- Replace or explicitly close the non-Unix no-op locking gap.
- Expose a narrow control-inverted mutation port: the store locks, loads the strict snapshot, invokes a core-supplied pure planner, applies returned writes through package-internal lock-free helpers, and unlocks.
- Forbid Store calls from inside the planner callback and make invalid nested mutation fail explicitly rather than self-deadlock.
- Preserve atomic replacement and content-version conflict behavior for ordinary entity writes.

## Acceptance criteria

- [x] Every released platform has an explicit, tested cooperating-writer serialization contract; other platforms reject graph mutation rather than silently no-op.
- [x] A graph mutation performs its authoritative scan, pure core planning/validation, and writes within one repository guard without exposing filesystem locking or teaching the store graph semantics.
- [x] The callback contract accepts and returns taskflow-owned snapshot/planned-write values, permits no Store calls at canonical-root scope, and detects same- or second-Store nesting without hanging.
- [x] Lock acquisition/release errors are attributable and process termination does not leave unrecoverable stale state.
- [x] Existing optimistic concurrency and ordinary write behavior remain compatible outside the documented callback-exclusive contention window.
- [x] The store boundary consumes one canonical strict-snapshot loader and the unused `ReadTaskGraph` service seam is removed; lint and mutation share the same strict projection constructor and task parser.

## Stress tests

- Concurrent opposite edge intents cannot both commit a cycle.
- Concurrent same-file and different-file writers cannot lose updates silently.
- Error, panic-equivalent, and process-exit paths release or recover the guard as designed.
- A faithful nested-acquisition regression test cannot reproduce the current self-deadlock.

## Sequencing

The control-inversion contract is fixed by ADR-0006's 2026-08-26 amendment, so implementation can proceed alongside the strict read foundation. Guarded dependency writes require both tasks to land.

## Implementation notes (2026-08-27)

The production boundary uses `core.LoadTaskGraph` for guarded writes while lint builds the same strict projection from its body-bearing scan. `TaskGraphMutationStore` takes the repository guard, loads that immutable snapshot, calls a pure planner over taskflow-owned values, delegates source/plan/prefix/final validation to core, applies surgical dependency updates through internal lock-free atomic writes, and returns any durable applied prefix after failure. Planner-provided write order is preserved because it is recovery data; semantic dependency sets are still canonicalized. Whole-snapshot byte versions plus immediate per-target CAS catch raw edits without exposing persistence tokens to planners. Graph writes stamp `updated_at` from the injected clock only for a semantic change.

The repository guard combines a canonical-root in-process mutex with root-directory `flock` on the supported macOS/Linux release targets. Windows and other non-Unix builds reject mutation explicitly until they have a native-tested cross-user lock contract. Callback exclusion is root-global across independent `FS` values; Store access during the callback fails fast with an attributable conflict. Entity creation participates in the guard and retains its prior missing-root behavior.

## Validation (2026-08-27)

Adversarial coverage includes concurrent opposite edges, independent same-process stores, nested reads/writes/mutations, degraded legacy migration, broken snapshots, raw edits outside the planned write set, unsafe interruption prefixes, partial-write convergence, planner panic, injected apply/release failure, and child-process termination. `go test -race ./...`, GolangCI-Lint, formatting, module-tidiness, `git diff --check`, and planning lint pass. Store tests compile for Linux, Windows, and the explicit-unsupported WebAssembly target.

## Adversarial hardening (2026-08-27)

The Gemini and Claude audits are closed after dispositioning all 23 findings. The guard now scopes callback exclusion by canonical planning root, including second-`FS` access; planner order is preserved as semantic recovery data; pure source/plan/prefix validation lives in core; whole-snapshot CAS is followed by immediate per-target CAS; graph changes receive a caller-clock `updated_at`; exact problem/legacy identities participate in source comparison; missing-root creation compatibility is restored; and non-Unix builds fail closed rather than claiming an untested Windows cache-lock guarantee.

Production-path regressions cover the keyed process mutex, root aliases, same- and second-store re-entry, concurrent callback contention, process termination, real cross-process opposite-edge mutation, safe and unsafe edge-reversal order, raw edits after a durable prefix, exact unreadable-set comparison, no-op timestamp stability, and the previously unguarded lint helpers. ADR-0006 and ARCHITECTURE record the root-wide contention, dry-run, raw-editor, platform, and recovery-order contracts. Prefix-validation scale is tracked by `6g3q4rtv8d0a` as a release gate for bulk apply.

Validation after the amendments: `go test -race ./...` passes; GolangCI-Lint reports 0 issues; formatting, `go mod tidy -diff`, and `git diff --check` are clean; and store tests cross-compile for Linux, Windows (explicit unsupported runtime path), and js/wasm.
