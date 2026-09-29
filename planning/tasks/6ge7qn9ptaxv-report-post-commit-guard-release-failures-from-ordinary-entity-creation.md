---
schema: 1
id: 6ge7qn9ptaxv
status: completed
epic: 21-code-quality-architecture-hardening
description: Do not silently discard a repository guard release error after an ordinary entity file was created.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 3
tags: [architecture, mutations, filesystem]
created: "2026-09-27"
depends_on: [6gdx7mcqq67d]
updated_at: "2026-09-27"
started_at: "2026-09-27"
completed_at: "2026-09-27"
---

# Report post-commit guard-release failures from ordinary entity creation

## Objective

`createEntityFile` currently uses `writeLock()` and `defer unlock()`, which cannot report a
guard-release failure after the file has become durable. Give ordinary task, epic, audit, and
research creation the same explicit post-commit diagnosis already used by guarded graph and Thread
mutations, without suggesting an unsafe retry.

## Acceptance criteria

- [x] A post-write guard-release failure reports that the file committed, with its exact local
      destination when available, while preserving the underlying error classification.
- [x] Core service and CLI human/JSON recovery output retain the kind-specific create receipt; no
      generic mutation envelope or domain `Path` fallback is introduced.
- [x] Research ID-collision retry stops once a create has committed, including if cleanup wraps
      `ErrConflict` (guarded at the service boundary by the predecessor task). A pre-commit collision
      joined with a guard-release failure also stops rather than discarding the cleanup error.
- [x] Tests inject release errors for every ordinary create kind and prove there is one durable file,
      no blind retry, and useful recovery output. Dry-run and pre-commit failures remain distinct.
- [x] A committed-create JSON error retains transient OS details without advertising that the
      whole create command is safe to retry.

## Out of scope

- Reworking unrelated edit, fix, or body-write paths that also use `writeLock()`.
- Changing the ordinary create serialization and no-clobber policy.

## Implementation notes

- Ordinary creates use the checked repository guard and return committed receipts alongside
  post-write release errors. Pre-commit failures and dry runs never claim a committed file.
- The CLI reuses the created-item identity and workspace vocabulary in a committed-error recovery
  field; it does not infer a destination from domain `Path` or introduce a generic mutation port.
- JSON revision 1.78 adds this recovery field without changing successful create envelopes.
- Adversarial review found two retry hazards: a Research collision joined with a failed guard
  release, and a transient filesystem hint on a committed-create error. An adapter-neutral
  finalization error now prevents the former retry; committed-create JSON suppresses the latter
  retry hint while retaining the OS detail. Both are pinned by regressions.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Predecessor [Preserve local mutation outcome evidence outside domain records](6gdx7mcqq67d-preserve-local-mutation-outcome-evidence-outside-domain-records.md)
- Adversarial implementation reviews: [Codex](../audits/6ge9gcfjnwte-2026-09-27-ordinary-create-guard-release-implementation-codex.md) and [Antigravity](../audits/6ge9gcfv95ej-2026-09-27-ordinary-create-guard-release-implementation-antigravity.md)
- Delivery: [PR #268](https://github.com/andy-esch/taskflow/pull/268)
