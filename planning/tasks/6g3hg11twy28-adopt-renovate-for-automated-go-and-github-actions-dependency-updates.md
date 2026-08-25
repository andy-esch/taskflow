---
schema: 1
id: 6g3hg11twy28
status: completed
epic: 21-code-quality-architecture-hardening
description: Configure Renovate with gomodTidy, Charm ecosystem grouping, coordinated Go toolchain bumps, and Dependency Dashboard major version gates.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [ci, dependencies, infra, renovate]
created: "2026-08-25"
updated_at: "2026-08-25"
started_at: "2026-08-25"
completed_at: "2026-08-25"
---
# Adopt Renovate for automated Go and GitHub Actions dependency updates

## Objective

Automate dependency updates and security patches across Go modules (`go.mod`/`go.sum`) and GitHub Actions workflows (`.github/workflows/*.yml`) using Renovate, mirroring the battle-tested configuration patterns from `desirelines`.

`taskflow` has strict CI and architecture guards (e.g. `just tidy-check`, `just docs-check`, `golangci-lint` v2 schema, `govulncheck`), and depends on rapidly evolving Charm v2 beta libraries. Renovate configuration must ensure:
1. `postUpdateOptions: ["gomodTidy"]` is always run so `go mod tidy -diff` passes in CI.
2. Charm v2 ecosystem packages (`charm.land/*`, `github.com/charmbracelet/*`, `github.com/muesli/*`) bump together in a single coordinated PR to avoid peer-dependency mismatch breaks.
3. Go version updates are grouped across `go.mod`, `ci.yml`, and `release.yml`.
4. Major updates are held behind the Dependency Dashboard for human review.
5. Non-major updates are grouped by manager and batched with a 3-day minimum release age to protect against supply-chain churn.

## Acceptance criteria

- [x] `renovate.json` is created at the repository root extending `config:recommended`, `:dependencyDashboard`, `:semanticCommits`, and `group:monorepos`.
- [x] `postUpdateOptions: ["gomodTidy"]` is configured so `go.sum` and `go.mod` remain tidy on module updates.
- [x] Charm ecosystem packages (`charm.land/*`, `github.com/charmbracelet/*`, `github.com/muesli/*`) are grouped into a dedicated `charm` package rule.
- [x] Go version bumps across `go.mod`, `.github/workflows/ci.yml`, and `.github/workflows/release.yml` are grouped under a `go version` package rule.
- [x] GitHub Actions workflow updates are grouped into a `github actions` package rule.
- [x] Major version updates require manual approval via `dependencyDashboardApproval: true` with the `requires-review` label.
- [x] Non-major updates are grouped by manager with `minimumReleaseAge: "3 days"` and a weekly schedule (`before 6am on monday`).
- [x] `just tidy-check` and `just docs-check` pass cleanly with the new configuration in place.

## Out of scope

- Setting up automated merge (`automerge: true`) — all dependency PRs should go through normal CI and human review.
- Multi-language manager configs (npm, uv, terraform) — `taskflow` is a pure Go and GitHub Actions repo.
