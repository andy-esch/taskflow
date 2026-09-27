---
schema: 1
id: 6ge7qn9ptaxv
status: ready-to-start
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
---

# Report post-commit guard-release failures from ordinary entity creation

## Objective

`createEntityFile` currently uses `writeLock()` and `defer unlock()`, which cannot report a
guard-release failure after the file has become durable. Give ordinary task, epic, audit, and
research creation the same explicit post-commit diagnosis already used by guarded graph and Thread
mutations, without suggesting an unsafe retry.

## Acceptance criteria

- [ ] A post-write guard-release failure reports that the file committed, with its exact local
      destination when available, while preserving the underlying error classification.
- [ ] Core service and CLI human/JSON recovery output retain the kind-specific create receipt; no
      generic mutation envelope or domain `Path` fallback is introduced.
- [x] Research ID-collision retry stops once a create has committed, including if cleanup wraps
      `ErrConflict` (guarded at the service boundary by the predecessor task).
- [ ] Tests inject release errors for every ordinary create kind and prove there is one durable file,
      no blind retry, and useful recovery output. Dry-run and pre-commit failures remain distinct.

## Out of scope

- Reworking unrelated edit, fix, or body-write paths that also use `writeLock()`.
- Changing the ordinary create serialization and no-clobber policy.

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Predecessor [Preserve local mutation outcome evidence outside domain records](6gdx7mcqq67d-preserve-local-mutation-outcome-evidence-outside-domain-records.md)
