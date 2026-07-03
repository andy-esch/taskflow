---
status: completed
epic: 17-pm-go-cli
description: Coerce known typed fields under --set and validate epic existence in SetFields, so set can't write unreloadable frontmatter
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [go, cli, data-integrity]
created: "2026-06-11"
updated_at: "2026-06-11"
started_at: "2026-06-11"
completed_at: "2026-06-11"
id: 6fb7ym4008ma
---

# Harden task set against silent frontmatter corruption

> ⚠️ **Externally proposed — needs independent review before implementing.**
> This task was filed from an outside code-review pass (not by the implementing
> agent). The finding below was verified empirically (see Evidence), but the
> implementing agent should reach its own buy-in on the fix before coding it.

## Objective

`task set` can silently write frontmatter that the strict loader then refuses to
parse — corrupting the file (it becomes a `FileProblem` and drops out of
`status`/`list`/rollup sweeps). Two distinct gaps in `core.Service.SetFields`:

1. **Type bypass via the `--set key=value` escape hatch.** The typed flags
   (`--tier`, `--tags`) are *fine* — `cli/task.go` puts a real `int`/`[]string`
   into the updates map and `store.valueNode` serializes them as `!!int`/`!!seq`.
   But `--set tier=4` / `--set tags=ui,chart` route through the generic loop
   (`internal/cli/task.go:158`) as **strings**, so the store writes `tier: "4"`
   (`!!str`). Worse, validation *passes*: `ValidateField` stringifies the value
   and `strconv.Atoi("4")` succeeds, so the corruption is written with no error.
2. **No epic-existence check on update.** `NewTask` validates the epic exists
   (`service.go:132`), but `SetFields` only calls `domain.ValidateField`, which
   can't query the store — so `task set <t> --epic bogus` writes an orphan epic
   id, silently dropping the task from epic rollups until a `lint` sweep flags it.

### Evidence (verified 2026-06-11)

A yaml.v3 round-trip against the `Task` struct confirms the corruption:

```
tier: "4"          -> err: cannot unmarshal !!str `4` into int
tier: 4            -> ok (Tier=4)        # what --tier produces
tags: "ui,chart"   -> err: cannot unmarshal !!str `ui,chart` into []string
tags: [ui, chart]  -> ok                 # what --tags produces
```

## Approach (for the reviewing agent to weigh)

- Coerce **known typed fields** when they arrive as strings, before the store —
  the natural seam is the service (it already owns `ValidateField`). Map
  `tier`/`autonomy_level` → `int`, list fields (`tags`/`related_tasks`/…) →
  `[]string` (split on `,`). Leave genuinely-custom `--set` keys as strings.
- For the epic check: have `SetFields` list epics and reject an unknown `epic`
  value (reuse `epicExists`), mirroring `NewTask`. Decide whether to gate only
  when `epic` is among the updates (cheap; one extra scan on set).

## Acceptance criteria

- [x] `task set <t> --set tier=4` and `--set tags=a,b` write `!!int`/`!!seq`
      (or are rejected), and the file reloads cleanly — no `FileProblem`.
- [x] `task set <t> --epic <unknown>` fails with `ErrValidation` and writes
      nothing.
- [x] Tests: a set→reload round-trip for the corrupting cases; an unknown-epic
      rejection. Suite + lint green.

## Out of scope

- The typed flags themselves (already correct).
- Broader frontmatter-schema validation beyond the known typed/list fields.

## Progress Log

### 2026-06-11 — independent review + fix (suite + lint green)

**Review finding (the symptom was mis-described, the bug is real and worse).**
Reproduced empirically against the real store: `SetFields` does *not* fail
silently — `store.SetFields` writes the file **atomically first, then re-parses**
(`fsstore.go`), so `--set tier=4` **corrupts the file on disk and then returns a
post-hoc error**. Either way the file is left unreadable and drops out of sweeps.
So the proposed fix is right; I also hardened the store so the corruption can't be
written in the first place.

**Fix (three parts):**
1. **Core coercion** (`core.SetFields`) — string values from the `--set` escape
   hatch are coerced to the native type a known typed field needs before the store
   sees them: `tier`/`autonomy_level` → `int`, list fields (`tags`) → `[]string`
   (comma-split, trimmed). `intFields`/`listFields` tables + `coerceField`/
   `splitList`. Typed-flag values (already native) pass through untouched.
2. **Epic existence check** (`core.SetFields`) — when `epic` is among the updates,
   reject an unknown id with `ErrValidation`, mirroring `NewTask` (`epicExists`).
3. **Store parse-before-commit** (`store.SetFields`) — *beyond the proposed scope,
   on-objective:* parse the updated content and write only if it reloads. Any
   caller that would produce unreadable frontmatter is now rejected **without
   touching the file** — directly enforcing "set can't write unreloadable
   frontmatter" as a store invariant, not just a CLI-path fix.

**Verified:** real CLI `task set --set tier=4` → `tier: 4` (unquoted int);
`--set tags=x,y` → `tags: [x, y]`; both reload clean (`lint` green).
`--set epic=bogus` → `validation failed: unknown epic "bogus"` (exit 11), file
untouched. `--set tier=notnum` → rejected.

**Tests:** `core/setfields_coercion_test.go` (external `core_test`, real store):
typed-string round-trips reload clean, non-numeric rejected, unknown-epic
rejected. `store/setfields_test.go`: `TestFS_SetFields_RejectsUnreloadable` —
unreloadable update rejected, file left byte-identical.

**Also fixed in passing:** the stale `Service` doc comment ("reusable by a *future*
TUI" → reused by both shipped adapters).

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/core/service.go` (`SetFields`), `internal/cli/task.go`
  (`--set` loop), `internal/domain/validate.go`.
