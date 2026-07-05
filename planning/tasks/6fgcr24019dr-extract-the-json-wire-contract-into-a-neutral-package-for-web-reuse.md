---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: ""
effort: L
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, web]
created: "2026-06-27"
updated_at: "2026-06-27"
started_at: "2026-06-27"
completed_at: "2026-06-27"
id: 6fgcr24019dr
---
Audit 2026-06-27-consumer-data-flow-architecture H1 (+L2). Lift the --json envelopes/DTOs/SchemaVersion/JSONSchema out of internal/cli/render into a neutral internal/wire pkg (depending only on core/domain) so a future web adapter reuses them without importing a sibling primary adapter. Export the DTOs there to collapse the dual schema-description sources (L2). Leave the *Human/Style renderers in render.