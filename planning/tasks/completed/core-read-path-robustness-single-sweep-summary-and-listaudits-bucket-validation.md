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
---
Audit 2026-06-27-consumer-data-flow-architecture H2 (+M7). Summary() calls QueryFindings after ListAudits already parsed every audit body, so each audit file is read from disk twice on the hottest path; reuse the already-parsed bodies (or one combined pass). Also validate the ListAudits bucket like ListTasks validates status, so an unknown bucket returns ErrValidation not a silently-empty list (web ?bucket= trap).