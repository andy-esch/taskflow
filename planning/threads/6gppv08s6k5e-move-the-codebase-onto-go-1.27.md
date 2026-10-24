---
schema: 1
id: 6gppv08s6k5e
status: unstarted
description: Adopt the Go 1.24-1.27 idioms from the September modernization audit and raise the supported minimum to 1.27 with no unclassified machine-output change.
goal: go.mod requires Go 1.27, the CLI encodes JSON with json/v2, and machine goldens change only through classified revisions.
created: "2026-10-23"
tags: [go, modernization, toolchain]
tasks: [6gppr2jh0j3x, 6gppr2js1bqq, 6gppr2jwqnpm]
---
# Thread: Move the codebase onto Go 1.27

**Goal.** go.mod requires Go 1.27, the CLI encodes JSON with json/v2, and machine goldens change only through classified revisions.

## Context

The [2026-09-14 modernization audit](../audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md)
found idioms from Go 1.24 through 1.27 that remove boilerplate, test latency and hand-rolled JSON
plumbing. During 2026-10-23 triage the owner chose to raise the supported minimum from 1.26 to
1.27 rather than defer the 1.27-only findings. Follow the [release guide](../../docs/RELEASING.md)
for the bump itself.

## Constraints

- Refactors change no behavior or machine output. Any intended byte change in JSON (for example
  json/v2 no longer HTML-escaping `<`, `>` and `&`) is a classified ADR-0008 revision.
- `omitempty` means different things in the two encoders. Verified on go1.27.2: the same struct
  encodes as `{"p":{}}` under v1 and `{"b":false,"n":0}` under v2. Field omission must be made
  explicit with `omitzero` before the encoder swap.

## Sequencing

- The 1.24-1.26 idioms task is independent and can land first.
- The `omitzero` task precedes the 1.27 bump task, a real prerequisite recorded as a dependency.
- [Store containment with `os.Root`](../tasks/6gppr2jnaahh-contain-planning-store-filesystem-access-with-os.root.md)
  came from the same audit but is deliberately not a member: it is sequenced by
  [the foundations design](../tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md).
