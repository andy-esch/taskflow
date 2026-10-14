---
schema: 1
id: 6gk06gktr1q5
status: completed
description: Harden the general external-review sandbox against accidental source-ref writes and make guarded transfer evidence truthful and achievable.
goal: Ordinary sandbox pushes cannot change source refs, and compliant guarded transfers produce truthful, verifiable delivery evidence.
created: "2026-10-12"
tags: [scripts, reviews, safety]
tasks: [6g7srp3py9fe, 6gk06e2g21hv]
updated_at: "2026-10-12"
started_at: "2026-10-12"
ended_at: "2026-10-12"
---
# Thread: Harden isolated review workspace safety and delivery evidence

## Goal

External reviews preserve source work and refs, and produce truthful evidence of guarded delivery,
using the general standalone shell protocol.

## Scope and sequencing

- [Push accident prevention](../tasks/6gk06e2g21hv-disable-pushes-from-isolated-review-workspaces.md)
  addresses the source checkout remaining a writable Git remote in a nominally independent clone.
- [Self-finalizing transfer attestations](../tasks/6g7srp3py9fe-make-adversarial-review-transfer-attestations-self-finalizing.md)
  addresses the report needing to freeze before it can record successful transfer.

These are independent hardening slices, not prerequisites for each other or for the documentation
Thread. Both preserve dirty snapshot capture, independent Git administration, restoration to a
sandbox-only baseline, and one-file guarded transfer. Completion requires their scoped criteria;
it does not claim that a configurable Git guard contains a deliberately hostile agent.

Keep this tool generic and shell-based; task-attached review entities and agent orchestration
remain separate work. No source/global Git configuration changes or automated workspace deletion.

## Evidence

- [Documentation ownership review](../audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md)
- [Scripts protocol](../../scripts/README.md)
