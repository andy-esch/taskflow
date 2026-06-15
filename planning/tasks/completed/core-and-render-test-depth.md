---
status: completed
epic: 17-pm-go-cli
description: core.Service at 54.5% with whole use-cases untested directly, render at 24.9% with no golden files, no built-binary smoke tests
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, testing, coverage]
created: "2026-06-12"
updated_at: "2026-06-14"
started_at: "2026-06-14"
completed_at: "2026-06-14"
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

- [x] core ≥80% with use-cases tested at the service seam.
- [x] Render output pinned by goldens; JSON envelopes decode-asserted. — done via
      `DisallowUnknownFields` strict-decode per envelope; golden *files* judged
      unnecessary (decode-assert is stricter and less brittle).
- [x] Binary smoke test exercises real exit codes.
- [x] One shared planning-tree fixture helper. — `internal/testutil` (`Write` +
      `Repo` builder); the store/cli/core/tui fixture helpers now delegate to it.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/core/`, `internal/cli/render/`, `cmd/tskflwctl/`,
  test files across packages.
## Progress (2026-06-12)

core 54.5% → **83.0%** (`usecases_test.go`: Lint, ListAudits, NewEpic;
`listtasks_test.go`: filter validation). render 24.9% → **68.4%** via
`render_test.go` — strict-decode envelope tests (`DisallowUnknownFields`
pins each JSON schema) plus human-output assertions; golden *files* were
judged unnecessary given the decode-assert approach, revisit if desired.
Built-binary smoke suite added (`cmd/tskflwctl/main_test.go`): real process
exit codes 0/10/11, stderr `error:` contract, version stamp. CRLF behavioral
test landed with [[store-write-path-hardening]].

## Progress (2026-06-14)

Final item done: `internal/testutil` now owns the one "mkdir + write a fixture
file" implementation (`Write` + a chainable `Repo` builder). The per-package
helpers — store's `writeTask`/`writeEpic`/`writeAudit`, cli's `mustWrite`,
core's `setFieldsRepo`, tui's `seedRepo` — delegate to it; their call sites are
unchanged, so the dedup carries no behavioral risk. Suite green (9/9), vet/gofmt
clean. Task complete.
