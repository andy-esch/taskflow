---
status: in-progress
epic: 17-pm-go-cli
description: CI installs golangci-lint v1.64.5 against the v2-schema .golangci.yml so the lint gate cannot run; just test lacks -race; no govulncheck
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [ci, tooling, go]
created: "2026-06-12"
updated_at: "2026-06-12"
started_at: "2026-06-12"
---
# Repair the CI lint gate and local test parity

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], findings H3/M20 + the
> -race drift). H3 was hand-verified: `.github/workflows/ci.yml:47` installs
> `golangci-lint@v1.64.5` while `.golangci.yml` declares `version: "2"` —
> v1 cannot parse v2 configs, so the repo's own "all three green" definition
> of done is currently unverifiable for lint.

## Objective

1. **H3 — Fix the lint gate.** Pin golangci-lint v2
   (`github.com/golangci/golangci-lint/v2/cmd/golangci-lint`) or switch to
   `golangci/golangci-lint-action` (which also caches). Confirm the same v2
   binary is what `just lint` documents/expects. Then actually run it and
   burn down whatever it reports — it has not been gating.
2. **`just test` lacks `-race` while CI has it.** The fsnotify watcher and
   debounce ticks are exactly where races live; developers should see them
   before CI does. Add `-race` to `just test` (or a `just test-race` the
   docs point at).
3. **M20 — No vulnerability scanning.** Add `govulncheck ./...` as a CI step
   and a Justfile alias. (`gosec` etc. stay the tracked `.golangci.yml`
   follow-up — out of scope here.)

## Acceptance criteria

- [ ] CI lint step runs the v2 binary against the v2 config and passes (or
      failures are triaged into follow-ups).
- [x] `just lint` works on a fresh machine following the README.
- [x] `just test` (or a documented variant) runs with `-race`.
- [x] `govulncheck` runs in CI.

## Related

- Epic [[17-pm-go-cli]]
- Touches `.github/workflows/ci.yml`, `Justfile`, possibly `.golangci.yml`.
## Progress (2026-06-12)

CI workflow now uses `golangci/golangci-lint-action@v8` with v2.1 (the v1
install couldn't parse the v2 config); `just test` runs `-race`; `just
vulncheck` + a CI govulncheck step added. Verified locally with golangci-lint
v2.1.6: **0 issues** — no burn-down needed. Remaining: confirm the first CI
run goes green after push (the only unchecked box). Note: `-race` needs cgo +
a C compiler — fine on dev Macs and CI's ubuntu image, unavailable in the
agent container.
