---
status: ready-to-start
epic: 17-pm-go-cli
description: core.Service at 54.5% with whole use-cases untested directly, render at 24.9% with no golden files, no built-binary smoke tests
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, testing, coverage]
created: "2026-06-12"
---
# Core and render test depth

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], finding M18 + testing
> lows). The seams already exist — this is cheap coverage of the
> architectural center, not new infrastructure.

## Objective

Coverage today: core **54.5%** (`ListTasks`, `Move`, `NewEpic`, `Lint`,
`LintFix`, and all audit use-cases at 0% direct — exercised only through CLI
tests, so core regressions surface as confusing CLI failures), render
**24.9%** (every formatter 0% direct, no golden files anywhere), cli 81.7%,
store 80.2%, tui 79.3%, domain 85.9%.

1. **fakeStore-based units for core.** The `fakeStore` praised in
   ARCHITECTURE.md exists but is only used for the epic rollup. Extend to
   Move / Lint / LintFix / audit paths.
2. **Golden-file tests for `cli/render`.** The formatters are pure functions
   over view-models — ideal golden targets; strict struct-decode assertions
   per JSON envelope (the `schema_version` contract is currently pinned only
   by loose substring checks).
3. **A built-binary smoke suite.** Exit codes 10–14 are tested only via
   `ExitCode(err)` mapping, never a real process exit; the ldflags version
   stamp is unverified. Build once in `TestMain`, run
   `init → epic new → task new → start → complete → lint`, assert codes.
4. **Behavioral edge tests:** CRLF round-trip with value assertions (fuzz
   seeds only assert no-panic today — pairs with
   [[store-write-path-hardening]]); a unicode slug/description case.
5. **Shared fixture builder:** four hand-rolled "build a planning tree"
   helpers across packages (`store/fsstore_test.go`, `cli/task_test.go`,
   `store/epicstore_test.go`, `core/setfields_coercion_test.go`) — extract
   an `internal/testutil` builder before they sprawl.

## Acceptance criteria

- [ ] core ≥80% with use-cases tested at the service seam.
- [ ] Render output pinned by goldens; JSON envelopes decode-asserted.
- [ ] Binary smoke test exercises real exit codes.
- [ ] One shared planning-tree fixture helper.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/core/`, `internal/cli/render/`, `cmd/tskflwctl/`,
  test files across packages.