---
status: ready-to-start
epic: 17-pm-go-cli
description: Guard error not sentinel-wrapped, unknown --set keys written silently with no unset, epic unclearable, updated_at clobbered, list-field drift
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [go, cli, store, validation]
created: "2026-06-12"
---
# `task set` follow-ups (sentinels, unknown keys, canonical field table)

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], findings M4/M5 + B1
> residuals). Direct follow-ups to the just-completed
> [[harden-task-set-against-silent-frontmatter-corruption]]. Item 2 was
> demonstrated live while filing these tasks: `task set <t> --set
> bogus_field=1` exits 0, writes the junk key, `lint` never flags it, and
> there is no way to remove it through the tool.

## Objective

1. **M4 — The new parse-before-commit guard returns a non-sentinel,
   misleading error.** `internal/store/fsstore.go:166-169` wraps
   `errBadFrontmatter` ("malformed frontmatter…") instead of
   `domain.ErrValidation` — exit 1 instead of 11, and the message blames the
   *file* when nothing was written. Wrap `ErrValidation` with an
   "update would not reload; nothing written" message.
2. **Unknown `--set` keys are silently written, and there is no `unset`.**
   "Arbitrary key=value" is documented behavior, but for the agent audience
   a typo (`prioirty=high`) silently pollutes frontmatter, `lint` doesn't
   flag unknown keys, and recovery requires hand-editing the file. Decide:
   warn on keys outside the known set (with `--force` to override), and/or
   add `--unset key` so mistakes are recoverable through the tool.
3. **M5 — "List fields" knowledge is duplicated and divergent.**
   `core/service.go:114` coerces only `tags`; `store/diagnose.go:14-17` /
   `fix.go:171` treat `related_tasks`, `dependencies`, `blocks`… as lists —
   so `--set related_tasks=a,b` writes a string the project's own
   `lint --fix` then rewrites. One canonical table (domain), ideally derived
   from `domain.Task`'s yaml tags, used by core coercion, store fixer, and
   the unknown-key warning above.
4. **B1 residuals:** clearing an epic is now impossible (`--epic ""` →
   `unknown epic ""`, `service.go:94-102`) — special-case detach or return
   a clear "epic cannot be cleared" message; `--set updated_at=…` validates,
   then is silently clobbered by the stamp (`service.go:103`) — reject the
   key explicitly like `status`, or honor it.

## Acceptance criteria

- [ ] Rejected updates exit 11 with a message that says nothing was written.
- [ ] A typo'd key cannot silently persist (warn/error), and `--unset`
      (if accepted) round-trips.
- [ ] `--set related_tasks=a,b` and `lint --fix` agree on the written form.
- [ ] Epic detach has a defined, tested behavior; so does `--set updated_at`.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/core/service.go`, `internal/store/fsstore.go`,
  `internal/store/diagnose.go`, `internal/domain/`.