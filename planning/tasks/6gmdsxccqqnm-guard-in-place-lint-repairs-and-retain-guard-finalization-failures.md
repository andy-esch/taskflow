---
schema: 1
id: 6gmdsxccqqnm
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Add raw-edit CAS and checked guard handling to ordinary lint fixes while retaining completed repairs and refusing whole-batch replay.
effort: 1 day
tier: 2
priority: high
autonomy_level: 2
tags: [correctness, repair, concurrency, persistence]
created: "2026-10-16"
---
# Guard in-place lint repairs and retain guard finalization failures

## Objective

Close the ordinary `FixFrontmatter` write path's remaining raw-edit and guard
finalization gaps without reworking spelling relocation or inventing a generic
mutation framework. Cooperating writers are serialized today, but the shared
in-place `writeFileAtomic(path, fixed, ...)` still has no immediate source CAS.

## Scope

- Capture each original file version and refuse raw byte drift immediately before
  in-place replacement. Preserve malformed-document repair, comments, unknown keys,
  permissions, graph-field refusals, preview and no-op behavior for every entity kind.
- Use the checked repository guard so authorization is current after waiting; join
  guard-release errors with the original failure and retain completed batch evidence.
- Keep partial relocation distinct from completed in-place repairs. Never replay
  the whole batch after durable progress or advertise such a retry to automation.
- Add deterministic fault seams/probes before claiming the raw-editor scenario is
  reproduced. This follow-up is currently supported by source inspection, not a
  timed concurrent-editor reproduction.

## Acceptance criteria

- [ ] A red-first raw-edit probe preserves newer bytes and refuses stale in-place fixes.
- [ ] Post-wait authorization changes refuse effects; release failure retains completed repairs and the original error.
- [ ] Dry-run/no-op, graph-owned field refusal, mode and source-preservation tests stay green.
- [ ] Core/CLI recovery and normal quality gates pass without whole-batch retries.

## Sequencing

Discovered beside `6gmbfxy1d561` (lint ID relocation). It is independently
actionable; no artificial prerequisite. The broader entity-integrity design may
reuse this evidence, but must not block a bounded safety fix. Do not reopen the
completed selected-identity Thread.
