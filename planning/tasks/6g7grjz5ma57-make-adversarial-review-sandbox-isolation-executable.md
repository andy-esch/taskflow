---
schema: 1
id: 6g7grjz5ma57
status: completed
epic: 21-code-quality-architecture-hardening
description: Replace repeated reviewer clone and transfer recipes with one tested fail-closed sandbox helper.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 4
tags: [reviews, agents, scripts, safety]
created: "2026-09-06"
started_at: "2026-09-06"
updated_at: "2026-09-06"
completed_at: "2026-09-06"
---

# Make adversarial review sandbox isolation executable

## Objective

Replace the long, repeatedly reinterpreted clone/checkpoint/transfer recipe in external-review
briefs with one standalone shell tool. Keep it general enough for any repository-relative review
deliverable and deliberately separate from taskflow's Go library and CLI until broader use proves a
different home is warranted.

## Scope

- Provide `create`, `verify`, and `transfer` operations over an arbitrary Git repository and one
  repository-relative deliverable.
- Capture staged, unstaged, untracked, and deleted source state in an independent local clone with a
  sandbox-only restoration baseline.
- Record private source identity, deliverable hash, snapshot fingerprint, and baseline evidence in
  the sandbox's own `.git` directory.
- Fail closed on shared Git metadata, source changes during capture, reviewer commits or staging,
  unrelated sandbox changes, source-deliverable drift, an empty report, or an unsafe target path.
- Update the adversarial-audit generator and script guidance to consume the general helper rather
  than embedding its implementation in every prompt.

## Acceptance criteria

- [x] `scripts/isolated-review-workspace.sh create` produces a non-worktree, no-hardlink clone whose
      baseline contains the source's current staged, unstaged, untracked, and deleted state.
- [x] The helper accepts any existing regular repository-relative deliverable; absolute paths,
      traversal, symlinks, and sandboxes inside the source repository are refused.
- [x] `verify` permits only one unstaged deliverable delta and reports reusable isolation
      attestation fields; commits, staging, unrelated changes, and source-deliverable drift fail.
- [x] `transfer` rechecks the guard, requires a changed deliverable, atomically copies back only that
      file, compares the result, and retains the sandbox.
- [x] The adversarial-review generator emits the short helper workflow while preserving the handoff
      freeze, sandbox-only mutation rules, exact finding grammar, and report requirements.
- [x] Disposable-repository smoke tests exercise capture, isolation, contamination refusal,
      source-hash refusal, successful one-file transfer, and unsafe target rejection on the local
      supported platform; shell syntax, ShellCheck, planning lint, and diff checks pass.

## Out of scope

- Adding this workflow to the Go library or `tskflwctl` command tree.
- General multi-file synchronization, automatic sandbox deletion, remote cloning, containers, or a
  security boundary against a malicious reviewer with direct source-checkout access.
- Retrofitting the two already-running compatibility audits; their original briefs remain frozen.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Audit preparation helper [`scripts/prepare-adversarial-review-audits.sh`](../../scripts/prepare-adversarial-review-audits.sh)

## Implementation result

The standalone Bash helper now owns isolated capture, verification, attestation, and guarded
single-file transfer. Future generated adversarial-review briefs call that helper; the two reviews
already in flight retain their frozen instructions. No Go package or `tskflwctl` command was added.

Validation covered syntax and ShellCheck, a disposable Git-repository smoke test, generator
preflight, `git diff --check`, and planning lint. The smoke test exercised staged, unstaged,
untracked, and deleted capture plus staged/empty/unrelated-change refusal, source drift, path and
symlink rejection, and a successful atomic transfer.
