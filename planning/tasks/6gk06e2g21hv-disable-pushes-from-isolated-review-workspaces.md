---
schema: 1
id: 6gk06e2g21hv
status: completed
epic: 21-code-quality-architecture-hardening
description: Prevent ordinary sandbox pushes from mutating source refs and verify the accident-prevention guard without claiming hostile-agent isolation.
effort: 2-4 hours
tier: 2
priority: medium
autonomy_level: 4
tags: [scripts, reviews, git, safety]
created: "2026-10-12"
updated_at: "2026-10-12"
started_at: "2026-10-12"
completed_at: "2026-10-12"
---
# Disable pushes from isolated review workspaces

## Objective

Prevent accidental Git pushes from the general review sandbox back into its source repository.
The helper makes independent Git storage, but `git clone --no-hardlinks` leaves `origin` pointing
at the live source checkout. Independence alone therefore does not prevent a reviewer from
creating or changing source refs with an accidental push.

## Scope

- Keep `scripts/isolated-review-workspace.sh` a standalone shell tool, not a taskflow Go feature.
- Disable ordinary pushes through the helper-created remotes/defaults without changing source
  Git configuration, refs, files, or user-wide Git settings.
- Have verification detect a missing or weakened push guard before accepting a report.
- Preserve dirty-state capture, independent objects/administration, sandbox-only baseline creation,
  and guarded one-deliverable transfer.
- State the limit: a configurable Git guard is accident prevention, not a security boundary
  against an agent deliberately overriding configuration or pushing to an explicit source path.

## Acceptance criteria

- [x] Newly created review sandboxes refuse ordinary default and explicit-origin pushes without
      modifying source refs or files, including when the source permits updates to other branches.
- [x] Verify/transfer reject removal or weakening of the configured push guard with actionable
      diagnostics; source configuration and user-global settings remain untouched.
- [x] Disposable-repository tests compare source refs and bytes before/after attempted pushes,
      exercise a weakened guard, and still prove the normal report/verify/transfer workflow.
- [x] Scripts guidance explains the guard, remediation, and explicit-URL/config-override limits
      without claiming that local Git configuration is a hostile-agent security sandbox.

## Sequencing

This is independent of the documentation ownership delivery and the existing transfer-attestation
task. Group them in the review-workspace hardening Thread without inventing a hard prerequisite.
The docs Thread may continue or close without waiting for this tooling improvement.

## Out of scope

- OS/network sandboxing, protection against a deliberately malicious reviewer, remote execution,
  paid-agent orchestration, or sandbox deletion.
- Changing the source repository's hooks, protections, refs, or push configuration.
- Redesigning the task-attached review entity model or moving review tooling into the Go library.

## Evidence and related work

- [Documentation ownership review](../audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md):
  surviving-hypothesis ledger observed the source path as the sandbox origin; no accidental push occurred.
- [General helper](../../scripts/isolated-review-workspace.sh) and [protocol](../../scripts/README.md)
- [Transfer-attestation task](6g7srp3py9fe-make-adversarial-review-transfer-attestations-self-finalizing.md)
- [Review-workspace hardening Thread](../threads/6gk06gktr1q5-harden-isolated-review-workspace-safety-and-delivery-evidence.md)

## Delivery and verification (2026-10-12)

- `create` sets every sandbox push URL to `isolated-review-push-disabled://source-is-read-only`,
  an unknown transport, immediately after cloning. Fetch still reads the source.
- `verify` and `transfer` refuse a removed `origin`, any remote whose effective push URL after
  `insteadOf` expansion is not that guard, and a workspace `remote.pushDefault` or
  `branch.<name>.pushRemote`/`remote` naming a path or URL. Each refusal prints the restoring
  `git -C <sandbox> …` command. Local upstreams (`remote = .`) remain allowed.
- [Disposable-repository tests](../../internal/tools/reviewworkspace/review_workspace_test.go)
  attempt default, origin, new-branch, existing-branch, `--all`, and tag pushes, comparing the
  source's refs, config, HEAD, index, objects, and working files before and after. A control push
  to the explicit source path then succeeds. That proves the source accepts other-branch updates
  and documents the explicit-URL limit. Both `verify` and `transfer` reject eight weakenings,
  and the redirected user-global configuration stays unchanged. The normal
  capture/verify/transfer workflow still moves only the deliverable.
- Mutations: removing guard installation fails the push test, and removing the `verify`/`transfer`
  check fails all eight weakening cases. ShellCheck, vet, lint, and the package tests pass on
  macOS `/bin/bash` 3.2 with openrsync; CI exercises Linux.
- [Scripts guidance](../../scripts/README.md) and the generated review brief state the guard,
  remediation, and limits: explicit-URL pushes, `git -c`/environment overrides, and user-wide or
  system configuration are outside a workspace-local accident guard.
