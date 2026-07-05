---
schema: 1
status: completed
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: init writes a pointer-only config (no tree) via --planning-repo; on a TTY asks here-vs-elsewhere. Bare non-TTY init still scaffolds.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [cli, init]
created: "2026-06-22"
updated_at: "2026-06-23"
started_at: "2026-06-23"
completed_at: "2026-06-23"
id: 6fes83r015px
---
# `init` pointer mode + interactive flow

Make `tskflwctl init` able to write a **pointer-only** config (no tree) and to
ask, on a TTY, whether planning lives here or elsewhere. Today `init` always
scaffolds the full tree — exactly what you don't want in an impl repo.

## Scope

1. Split `config.Init` into "scaffold tree + config" (today) vs. "write pointer
   config only" (writes `.tskflwctl.toml` with `planning_repo = "..."`, **no
   tree**), keyed by a mode/role.
2. Flags on the `init` command: `--planning-repo <path>` → pointer mode
   (validate target per the discovery task, write pointer config, no tree).
3. Interactive flow (TTY, no mode flag): prompt "Will planning live in *this*
   repo, or *another* repo?" via the existing `app.Gate` / `Prompter`
   (`SelectOne`), and if another, `Text`-prompt the path. Reuse the flag-twin
   pattern so flags stay the headless escape hatch.
4. Non-interactive + no flag → today's full scaffold (backward compat; headless
   "create a planning repo" still works).

## Acceptance criteria

- [ ] `init --planning-repo ../planning` writes a pointer config and **no** tree;
      a bad target errors (no half-written state).
- [ ] Bare `init` on a TTY asks here-vs-elsewhere; off a TTY it scaffolds as today.
- [ ] Idempotent; `--dry-run` previews; `--json` reflects the mode.
- [ ] Suite + lint green.

## Depends on

- Config schema + TOML parser; discovery validation (shared resolve/validate).

## Related

- [[23-point-an-impl-repo-at-an-external-planning-repo]].
- Interactive primitives mirror [[20-cli-ux-and-ergonomics]].

## Review hardening (2026-06-23)

Two adversarial reviewers (config/pointer + CLI/envelope). CLI/envelope: clean. Config found 3 MAJORs, all fixed:
- **Dir-creation parity**: InitPointer now MkdirAll(dir) (after validation) — was a sentinel-less exit 1 on a missing --path, vs scaffold which creates it.
- **Mode-collision silent no-op**: re-pointing to a different target / switching scaffold↔pointer now errors (ErrConflict, exit 14) instead of silently keeping the old config; same target stays an idempotent no-op.
- **Scaffold-over-pointer data fork**: Init now refuses (ErrConflict) to scaffold a local tree over an existing pointer config (would orphan the tree while discovery follows the pointer).
Plus doc nit: --planning-repo help now says 'relative to --path'. Tests added for all three.
