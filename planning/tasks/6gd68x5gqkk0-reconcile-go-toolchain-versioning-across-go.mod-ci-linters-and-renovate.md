---
schema: 1
id: 6gd68x5gqkk0
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Document Go-version roles, enforce CI/container/linter alignment, and coordinate Renovate line upgrades without delaying toolchain patch updates.
effort: 2-4 hours
tier: 3
priority: medium
autonomy_level: 3
tags: [ci, tooling, go]
created: "2026-09-24"
updated_at: "2026-10-09"
started_at: "2026-10-09"
---
# Reconcile Go toolchain versioning across go.mod, CI linters, and Renovate

## Objective

Make Go toolchain coordination explicit and executable without conflating the minimum supported
Go version, the CI/release line, the reproducible container patch pin, and a developer's local Go.
Keep line upgrades deliberate while allowing prompt patch updates, and make stale or incompatible
configuration fail early rather than discovering it during release qualification.

## Scope refresh (2026-10-09)

The original prerequisites shipped: `go.mod` declares Go 1.26.0, CI uses golangci-lint v2.13,
and PR #223 is merged. v0.23.0 qualification exposed the remaining gap: local Go 1.27.1 and
the pinned container's Go 1.26.6 were vulnerable; updating to 1.27.2 and 1.26.9 cleared the gates.
The container update is merged in [PR #291](https://github.com/andy-esch/taskflow/pull/291).
Do not repeat those upgrades or raise the module's minimum just to refresh the release compiler.

## Scope

- Document the current policy in the release guide: `go.mod` owns the minimum and CI test line;
  the release workflow owns the compiler line shared with CI lint and the container. Compatible
  newer release/local Go is allowed without raising the minimum. Linter build compatibility
  remains a runtime preflight against the minimum and active compiler line, not a version guess.
- Add an offline `just toolchain-check` backed by real-config regression tests. Wire it into CI
  and the shared release gate, checking setup-go selectors/latest-patch behavior, container Go
  precision/minimum/line, and the CI/container linter release line. Keep vulnerability scans as
  the authority for patch security: a static drift check cannot know when a new CVE appears.
- Align the container linter with the already selected CI line. Track its ARG pin through
  Renovate's regex manager so grouping is actionable, not just a label on untracked values.
- Give Go-specific rules precedence over broad manager grouping. Explicitly propose minimum Go
  updates with `rangeStrategy: "bump"`, grouping line changes behind dashboard approval.
  Separate toolchain patches from line upgrades; let patches bypass the weekly schedule and
  release-age delay without enabling automerge or automatically raising the module minimum.
- Validate the Renovate configuration with Renovate itself and exercise the actual extraction
  and rule application on representative Go, container, linter, and unrelated-action updates.
- Key the cached Go-built vulnerability scanner by setup-go's resolved patch version, so patch
  refreshes do not retain a binary from the old compiler. Keep developer vulnerability scanning
  consistent with CI/release package-level scanning.

## Acceptance criteria

- [x] The release guide distinguishes minimum Go, CI/release line, exact
  container pin, and local toolchain compatibility without adding a duplicate
  version manifest.
- [x] An offline just toolchain-check rejects workflow/container/linter drift
  and runs in CI and the shared release validation gate.
- [x] Renovate discovers Go directives, setup-go selectors, the Containerfile
  base image, and the container linter ARG; Go-specific grouping wins over broad
  manager rules.
- [x] Go line and minimum-version upgrades require dashboard approval, while
  toolchain patch updates have a separate unscheduled, zero-release-age,
  non-automerge path.
- [x] Regression tests exercise missing and incompatible configuration,
  latest-patch selection, and unchanged module-minimum behavior; Renovate
  extraction and rule application are verified with the actual engine.
- [x] Focused release-validator tests, full race tests, lint, build, module
  tidiness, planning lint, and strict Renovate config validation pass.

## Out of scope

- Upgrading the supported/release Go line to 1.27 or forcing developers onto an exact patch.
- Automatically merging dependency updates, guaranteeing Renovate runs instantly, or treating
  Docker/stdlib vulnerabilities as GitHub module alerts that Renovate necessarily receives.
- Adding another version manifest or a runtime dependency to the published CLI.
- Introducing new linter suites into `.golangci.yml`.

## Implementation and validation (2026-10-09)

Implemented on `chore/coordinate-go-toolchains`, based on v0.23.0
(`ee54b930a9bb673c53882622546bd851a2355eae`). No runtime CLI, module minimum, or
machine-contract change. The container linter moves from v2.12.2 to v2.13.0.

- `just toolchain-check`, focused release-validator tests, the full race suite,
  `just lint`, `just build`, `just tidy-check`, and planning `lint --links` pass.
- Host: Go 1.27.2; golangci-lint 2.14.0 built with Go 1.27.1. The preflight
  accepts a compatible build on the same Go line without demanding patch identity;
  regressions reject older module/active-line compilers and unknown build metadata.
- The updated container builds. Its Go and linter build Go are both 1.26.9;
  golangci-lint is 2.13.0. Focused tests and real lint pass with the source mounted
  read-only. This is container qualification, not a claim that the complete clean
  release-candidate gate has run on these uncommitted changes.
- Renovate 44.77.0 passes strict config validation and
  `just renovate-toolchain-check <installed-package-directory>`: real extraction
  covers the module directive, workflow selectors, container base, and linter ARG;
  rule application covers reviewed line changes, prompt patches, and unrelated actions.
  Hosted scheduling and inherited-preset resolution are not simulated by that check.
- The package vulnerability scan passes; findings limited to unused modules remain
  informational. Static coordination checks do not certify remote patch freshness.

Independent review is the remaining checkpoint before closing this task:
[Codex](../audits/6gj70e4p18b3-2026-10-09-go-toolchain-coordination-implementation-codex.md)
owns compiler/preflight probes;
[Antigravity](../audits/6gj70e4z9zrg-2026-10-09-go-toolchain-coordination-implementation-antigravity.md)
owns actual Renovate extraction and precedence probes. Both briefs require an
independent sandbox, a valid mutation probe, and a bounded cross-check of the other lens.

Antigravity review is complete and closed with no findings. Its report independently
extracted all four dependency surfaces and killed both the Renovate ordering and
active-linter-line mutations. Owner reruns of the focused tests and actual-engine
check pass. Acceptance is bounded to stable compiler cases, extraction, and supplied
update-type rule application; remote classification/scheduling are not certified.
Codex remains pending; the task stays in progress until that checkpoint is handled.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- PR [#223](https://github.com/andy-esch/taskflow/pull/223) (Renovate `x/term` bump blocked by linter Go version skew)
- PR [#152](https://github.com/andy-esch/taskflow/pull/152) (Renovate GitHub Actions bump including `golangci-lint v2.13`)
- Audit [2026-09-14 Go 1.27 modernization](../audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md)
- Release [v0.23.0 qualification](6gj5zz9efdtt-cut-v0.23.0-as-an-adapter-boundary-and-planning-safety-release.md)
- Guide [Releasing taskflow](../../docs/RELEASING.md)
- Primary references: [Renovate Go directives](https://docs.renovatebot.com/modules/manager/gomod/),
  [Dockerfile/Containerfile extraction](https://docs.renovatebot.com/modules/manager/dockerfile/),
  [rule precedence and update policy](https://docs.renovatebot.com/configuration-options/#packagerules).
