---
schema: 1
id: 6gppr2jh0j3x
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Use errors.AsType and new(expr), move timing-based store concurrency tests to testing/synctest, adopt b.Loop and wg.Go, and prune redundant module dependencies.
effort: 1 day
tier: 3
priority: medium
autonomy_level: 3
tags: [modernization, go, testing]
created: "2026-10-23"
audit_sources: [planning/audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md]
updated_at: "2026-10-30"
---
# Adopt Go 1.24-1.26 idioms in error mapping, tests, and dependencies

## Objective

The 2026-09-14 modernization audit found idioms available on the Go 1.26 minimum that remove
boilerplate and real test latency. Adopt them where they simplify, settle the claimed TUI shutdown
watcher leak with evidence, and prune dependencies that modern tooling makes redundant. These are
refactors: behavior and machine output must not change.

## Acceptance criteria

- [ ] CLI error mapping, `fserror`, and the other listed `errors.As` sites use `errors.AsType`; test pointer helpers and `cloneBool` use `new(expr)`.
- [ ] Negative store concurrency assertions that wait on `time.After` run under `testing/synctest`, removing the wall-clock waits without weakening what they prove.
- [ ] Benchmarks use `b.Loop()`; plain completion wait groups use `wg.Go`, while readiness/rendezvous groups stay as they are.
- [ ] The claimed orphaned `fsnotify` watcher when quitting during an in-flight workspace open is reproduced and fixed, or falsified with recorded evidence.
- [ ] `mango-cobra`/`roff` are replaced by `cobra/doc` only if the generated man pages stay acceptable, and the dual JSON Schema library decision is recorded.
- [ ] `just test`, `just lint`, and `just docs-check` pass with no golden or wire changes.

## Out of scope

- Go 1.27-only features, `os.Root` containment, and `omitzero` envelope changes (separate tasks).

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit [2026-09-14 Go 1.27 modernization](../audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md), findings H1, M5, L1, L2
- Coordinate the watcher evidence with [watcher session binding](6gmbfxy96tpw-bind-tui-watcher-health-and-reload-messages-to-their-originating-session.md)

Partially reinforced by audit 2026-10-30-simplification: M2. Two of this task's named idioms — `new(expr)` and `wg.Go` — are automated by the Go 1.26 `go fix` modernizers `newexpr` and `waitgroup`, which `go.mod`'s current 1.26.0 minimum already supports. `go fix -diff ./internal/core/... ./internal/store/...` additionally flags `stringsseq` and `slicescontains` sites in 10 production files that this task's hand-enumerated list does not include. Finding M2 stays open and proposes coordinating with this task rather than duplicating it; consider whether the site inventory should come from `go fix -diff` instead of a typed list.
