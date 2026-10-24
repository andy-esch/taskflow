---
schema: 1
id: 6gppv08araf8
status: unstarted
description: Finish making the CLI's machine contract discoverable from the binary and economical for agents, following the completed self-describing-contracts Thread.
goal: An agent learns every selector, envelope, vocabulary and diagnostic kind from schema output, and resumes or mutates work with compact receipts, without parsing prose.
created: "2026-10-23"
tags: [agents, schema, contracts]
tasks: [6fpnn6zk157b, 6gcqz5b0j0dg, 6gppr2k0exeb, 6gppr2k482rm, 6gppr2k7xk5s, 6gppr2kk1thc, 6gppr2kps7v9]
---
# Thread: Complete the agent-facing machine contract

**Goal.** An agent learns every selector, envelope, vocabulary and diagnostic kind from schema output, and resumes or mutates work with compact receipts, without parsing prose.

## Context

The [self-describing contracts Thread](6gbn4v2jpf2m-make-cli-contracts-self-describing-and-bounded.md)
published the command surface, bounded reads and revision policy. Audits since then found the
remaining places where an agent still has to provoke an error, re-parse a body, or read more than
it needs:
[2026-07-24 AI-agent ergonomics](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md),
[2026-10-06 adapter hygiene](../audits/6gh1frgyz2fh-2026-10-06-adapter-hygiene.md) and
[2026-10-12 machine contract](../audits/6gjzzvdy5jnw-2026-10-12-arch-machine-contract.md).
This Thread was formed during 2026-10-23 audit triage.

## Constraints

- Every observable output change is a classified ADR-0008 revision. Several members bump the
  revision; land them one at a time rather than batching unrelated changes into one revision.
- The vocabulary-annotation member starts with a human decision on the proposed ADR-0008
  amendment. Do not implement annotations before that decision is recorded.
- Making body-mutation receipts compact by default changes an established shape; classify it
  before it lands and keep an explicit opt-in for the full body.
- The replay/change-set member is decision-only. Implementation tasks it authorizes join this
  Thread only if the decision accepts them.

## Sequencing

Only one dependency is real: the task brief's latest-progress entry needs the canonical
progress-section shape from the task-log member. Everything else can proceed in any order.

## Out of scope

- JSON encoder modernization (the Go 1.27 Thread) and store integrity foundations
  ([the foundations design](../tasks/6g7wsr8yvt1a-define-the-next-shared-entity-integrity-foundations.md)).
