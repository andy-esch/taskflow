---
schema: 1
id: 6gdwqmzxp7ka
status: completed
epic: 21-code-quality-architecture-hardening
description: Publish a reusable GitHub Action (action.yml) in taskflow that sets up tskflwctl and validates planning trees with tskflwctl lint.
effort: 1-2 hours
tier: 3
priority: medium
autonomy_level: 4
tags: [ci, github-actions, lint, distribution]
created: "2026-09-26"
updated_at: "2026-09-26"
started_at: "2026-09-26"
completed_at: "2026-09-26"
---

# Publish a reusable GitHub Action for planning tree linting

## Objective

Provide a first-class, reusable GitHub Action (`action.yml`) in this repository so that any project using `taskflow` can validate its planning tree with `tskflwctl lint` in CI with minimal workflow boilerplate. The action should handle binary acquisition (leveraging pre-built release binaries for fast execution) and execute validation, failing CI runs when frontmatter, acceptance criteria, or task DAG constraints are violated.

## Scope

- Add a root composite action (`action.yml`) to `taskflow`:
  - Fetch and install `tskflwctl` using the pre-compiled release archive for the runner architecture (`linux_amd64` / `linux_arm64`), with a fallback to `go install github.com/andy-esch/taskflow/cmd/tskflwctl@<version>` when running outside GitHub release environments.
  - Expose clean inputs:
    - `version`: Version tag or ref (default: `latest`).
    - `links`: Boolean to run with `--links` to check markdown cross-links (default: `false`).
    - `working-directory`: Path to directory containing `.tskflwctl.toml` or `planning/` (default: `.`).
    - `args`: Passthrough string for additional `tskflwctl lint` flags.
  - Execute `tskflwctl lint` in the designated directory and bubble up the process exit code (`11` for validation failure, `1` for general failure).
- Dogfood the action in `taskflow`'s own CI (`.github/workflows/ci.yml` or a dedicated workflow) against its `planning/` tree.
- Document action usage and minimal caller workflow snippets in `README.md` and CLI documentation.

## Acceptance criteria

- [x] A root action.yml composite GitHub Action is added to taskflow that
  installs tskflwctl and executes tskflwctl lint.
- [x] The action supports configurable inputs: version (default latest), links
  (default false), working-directory (default .), and extra args.
- [x] The installation step fetches the pre-built static binary release asset
  for linux-amd64 / linux-arm64, falling back to go install if unavailable.
- [x] Validation failures in tskflwctl lint exit with a non-zero exit code that
  fails the GitHub Actions step.
- [x] A test workflow or step in taskflow dogfoods the action against taskflow's
  own planning tree in CI.
- [x] Usage instructions and example workflow snippets are documented in the
  repository docs or README.

## Out of scope

- Auto-remediation (`--fix`) within the action execution.
- GitHub pull request review commenting or bot annotations (can be added as an optional enhancement later).
- Windows runner support in the initial composite action (focus on Linux runners).

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Release automation in [6fbj87000zs7](6fbj87000zs7-binary-releases-via-goreleaser.md), which publishes the release archives consumed by the action
- Decoupled planning epic [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md)

## Progress Log

- 2026-09-26: Implemented the composite GitHub Action at `action.yml`.
  - Added configurable inputs for `version` (default: `latest`), `links` (default: `false`), `working-directory` (default: `.`), `args`, and `github-token`.
  - Configured multi-platform asset detection (`linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64`) with fast direct downloads (~1s) from GitHub Releases via `gh release download`, plus support for `version: local` (building from checkout) and a fallback to `go install` when pre-built binaries are not available for a requested git ref.
  - Hardened action steps against script injection and shellcheck warnings by passing inputs via `env:` rather than inline template expansion.
  - Added a dogfooding `planning-lint` job to `.github/workflows/ci.yml` that executes `uses: ./` against taskflow's own planning tree.
  - Documented the action, invocation syntax, and inputs in `README.md`.
  - Verified `just test`, `just lint`, `just docs-check`, `just tidy-check`, `actionlint`, and `shellcheck` all pass cleanly with zero errors.

- 2026-09-26: Relocated composite action to `actions/lint/action.yml` (option 2) to maintain a clean separation between repo-internal CI and shareable public actions, added `actions/README.md` documentation, updated dogfooding in `ci.yml` to `uses: ./actions/lint`, and updated `README.md`.
