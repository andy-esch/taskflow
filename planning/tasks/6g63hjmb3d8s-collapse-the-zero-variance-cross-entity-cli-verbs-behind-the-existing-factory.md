---
schema: 1
id: 6g63hjmb3d8s
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: path is five identical copies modulo a noun; extend the newTransitionCmd factory across entities where variance is zero
effort: 2-4 hours
tier: 3
priority: low
autonomy_level: 3
tags: [cli, refactor, architecture]
created: "2026-09-02"
audited: "2026-10-11"
updated_at: "2026-10-11"
---
# Collapse the zero-variance cross-entity CLI verbs behind the existing factory

## Objective

`task path`, `epic path`, `audit path`, `research path` and `thread path` are
identical implementations modulo four substitutions: a noun for the error text, a
completion function, an options function, and the `core.Service` method. The tail
was already factored (`emitPath`); the command shape was not. Several other verbs
sit on the same spectrum.

The codebase already owns the fix and uses it in six places — `newTransitionCmd`,
`deprecatedTransitionCmd`, `newAuditMoveCmd`, `newTaskDependencyEdgeCmd`,
`newThreadMembershipCmd`, `newThreadLifecycleCmd` — but always *within* one
entity, never *across* entities.

This is deliberately NOT the data-driven collapse that
`docs/ARCHITECTURE.md` considered and declined ("for five heterogeneous entities
it trades clarity for machinery"). That reasoning is right for `list`/`show`/`set`,
where the entities genuinely differ — tasks carry status/tier/priority, audits
carry findings/buckets, research has no lifecycle at all. It does not hold for
verbs whose per-entity variance is *zero*. The job here is to identify that subset
honestly and leave everything else alone.

## Acceptance criteria

- [ ] The per-entity `path` commands are built by one parameterized factory in the style of `newTransitionCmd`
- [ ] Each candidate verb is assessed for genuine per-entity variance before being folded in; the ones with real variance are listed and left alone with a one-line reason
- [ ] Per-entity help text, examples, and completions stay exactly as specific as they are today — the factory takes them as parameters rather than generating generic prose
- [ ] No change to any command's observable output, `--json` envelope, or exit code; existing goldens pass unchanged
- [ ] `docs/ARCHITECTURE.md`'s fan-out paragraph is updated to distinguish zero-variance verbs from the shaped ones it correctly declines to collapse

## Out of scope

- `list`, `show`, and `set` — their per-entity shape is real and the declined collapse still applies
- A CLI command registry or descriptor table; this is the existing factory pattern applied one axis over
- Collapsing the `core.Service` per-entity method fan-out (`TaskPath`/`ResearchPath`/…) behind a generic seam
- Touching the completion functions beyond passing them in

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)

## Sweep refresh (2026-10-11)

2026-10-11: automated weekly sweep — no drift; every premise and all six cited factories
re-verified against current code, so the task stands exactly as written.

**Verified accurate (2026-10-11):**
- All five `path` commands are still separate, hand-written constructors:
  `newTaskPathCmd` (`internal/cli/task.go:360`), `newEpicPathCmd` (`epic.go:28`),
  and the audit / research / thread counterparts (`audit.go:571`, `research.go:331`,
  `thread.go:318`). Read side by side, the variance is still exactly the four
  substitutions the Objective names — noun, completion function, options function, and
  the `core.Service` method.
- The tail is still factored and the shape is not: `emitPath` (`internal/cli/task.go:384`)
  is shared, `resolveOne` + `absPath` + the `Annotations: {"safety": "read-only"}` line are
  copied per entity.
- All six cited in-entity factories still exist: `newTransitionCmd`,
  `deprecatedTransitionCmd`, `newAuditMoveCmd`, `newTaskDependencyEdgeCmd`,
  `newThreadMembershipCmd`, `newThreadLifecycleCmd`.
- All five `core.Service` `*Path` methods still exist (`TaskPath`, `EpicPath`, `AuditPath`,
  `ResearchPath`, `ThreadPath`), so the "Out of scope" note about not collapsing that
  fan-out still describes a live seam.
- `docs/ARCHITECTURE.md:240` still carries the declined data-driven collapse verbatim
  ("for five heterogeneous entities it trades clarity for machinery"), so the last
  acceptance criterion still has the paragraph it expects to amend.

**One small thing for whoever picks this up** (noted, not actioned — the sweep does not
edit `internal/`): `emitPath`'s doc comment says "Shared by task/epic/audit path", but it
is shared by all five, including research and thread. It is the kind of undercount the
factory in this task would make structurally impossible to repeat.
