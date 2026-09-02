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
