---
schema: 1
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'Optional ergonomic sugar: named verbs epic retire/deprecate/activate over the existing epic move <status> path (and TUI m menu), mirroring task start/complete. Redundant with move; low priority.'
effort: S
tier: 3
priority: low
autonomy_level: 3
tags: [cli]
created: "2026-06-25"
---
## Objective

Optional: named epic lifecycle verbs (`epic retire`/`deprecate`/`activate`) as sugar over `epic move <status>`, mirroring the task verbs (start/complete/...). 

## Why low priority
`epic move <id> <status>` + the TUI `m` action menu already cover status movement fully. Named verbs are pure ergonomics over an existing path. Deferred from [[epic-mutation-add-epic-set-and-epic-edit]]. Pick up only if the shortcut is wanted.