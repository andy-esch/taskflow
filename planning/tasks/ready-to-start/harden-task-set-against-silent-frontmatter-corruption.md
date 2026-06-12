---
status: ready-to-start
epic: 17-pm-go-cli
description: Coerce known typed fields under --set and validate epic existence in SetFields, so set can't write unreloadable frontmatter
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [go, cli, data-integrity]
created: "2026-06-11"
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

- [ ] `task set <t> --set tier=4` and `--set tags=a,b` write `!!int`/`!!seq`
      (or are rejected), and the file reloads cleanly — no `FileProblem`.
- [ ] `task set <t> --epic <unknown>` fails with `ErrValidation` and writes
      nothing.
- [ ] Tests: a set→reload round-trip for the corrupting cases; an unknown-epic
      rejection. Suite + lint green.

## Out of scope

- The typed flags themselves (already correct).
- Broader frontmatter-schema validation beyond the known typed/list fields.

## Related

- Epic [[17-pm-go-cli]]
- Touches `internal/core/service.go` (`SetFields`), `internal/cli/task.go`
  (`--set` loop), `internal/domain/validate.go`.
