---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: ""
effort: M
tier: 3
priority: high
autonomy_level: 3
tags: [architecture, core]
created: "2026-06-27"
updated_at: "2026-06-27"
started_at: "2026-06-27"
completed_at: "2026-06-27"
id: 6fgcr2402rbj
---
Audit 2026-06-27-consumer-data-flow-architecture H4. A core ErrorClass(err) over the domain sentinels so CLI->exit code, web->HTTP status, and TUI->{flash, reload-on-ErrConflict, inline field error} share one mapping. Drop the TUI strings.TrimPrefix sentinel hack (model.go:232); the TUI currently cannot distinguish ErrConflict (needs reload) from ErrValidation.