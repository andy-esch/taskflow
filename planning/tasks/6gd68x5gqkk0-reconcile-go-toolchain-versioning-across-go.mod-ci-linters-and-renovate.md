---
schema: 1
id: 6gd68x5gqkk0
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Upgrade CI golangci-lint to support Go 1.26+, bump go.mod to 1.26, and configure Renovate to group and gate Go version bumps.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [ci, tooling, go]
created: "2026-09-24"
updated_at: "2026-09-24"
---
# Reconcile Go toolchain versioning across go.mod, CI linters, and Renovate

## Objective

Reconcile the Go toolchain versioning skew across `go.mod`, GitHub Actions workflows,
release validation, and Renovate. CI tests and release builds run on Go 1.26+, but
`go.mod` remains on `go 1.25.12` and CI pins `golangci-lint` to `v2.4` (compiled with
Go 1.25), causing automated dependency bumps (such as PR #223 for `golang.org/x/term`) to
break CI when Go's MVS toolchain updates the `go` directive to 1.26. Configure Renovate to
actively manage, group, and gate Go language version bumps alongside workflow toolchains.

## Scope

- Unblock Go 1.26 in CI by bumping `golangci-lint` in `.github/workflows/ci.yml` from `v2.4`
  to `v2.13` (or `v2.12`+), matching official Go 1.26 support (which began in v2.9.0).
- Update `go.mod` to target `go 1.26.0` (or `go 1.26`) and run `go mod tidy`.
- Configure `renovate.json` with `matchDepTypes: ["golang"]` and `rangeStrategy: "bump"` so
  Renovate manages the `go.mod` `go` directive, grouping `go.mod`, GitHub Actions `setup-go`
  (`actions/go-versions`), and `build/release-validation/Containerfile` into a single
  reviewable PR gated by dashboard approval (`dependencyDashboardApproval: true`).
- Verify that `just lint`, `just test`, and `scripts/release-validate.sh` pass cleanly under
  the upgraded configuration.

## Acceptance criteria

- [ ] `.github/workflows/ci.yml` pins a `golangci-lint` release built with Go 1.26+ (v2.12 or v2.13), resolving the `can't load config: Go language version is lower than targeted` CI failure.
- [ ] `go.mod` declares `go 1.26.0` (or `go 1.26`) and `go mod tidy` is clean.
- [ ] PR #223 (`golang.org/x/term` v0.46.0) is rebased or updated and passes all CI checks.
- [ ] `renovate.json` has an explicit package rule for `gomod`'s `go` directive with `rangeStrategy: "bump"`.
- [ ] `renovate.json` groups `go.mod`, GitHub Actions Go versions, and `Containerfile` into the `"go version"` group with `dependencyDashboardApproval: true` to prevent uncoordinated toolchain drift.
- [ ] `scripts/release-validate.sh` and `internal/tools/releasevalidate` tests pass without regressions.

## Out of scope

- Upgrading to Go 1.27 before the codebase modernizes its Go 1.26 features.
- Introducing new linter suites into `.golangci.yml`.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- PR [#223](https://github.com/andy-esch/taskflow/pull/223) (Renovate `x/term` bump blocked by linter Go version skew)
- PR [#152](https://github.com/andy-esch/taskflow/pull/152) (Renovate GitHub Actions bump including `golangci-lint v2.13`)
- Audit [2026-09-14 Go 1.27 modernization](../audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md)
