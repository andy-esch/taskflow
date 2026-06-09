---
status: completed
epic: 17-pm-go-cli
description: 'Finishing polish: version command and flag, help examples, styled command output and next-step hints, no-color flag and FORCE_COLOR support'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: "2026-06-09"
updated_at: "2026-06-09"
completed_at: "2026-06-09"
---

# CLI polish bundle (version, help examples, color consistency, hints, color env)

## Objective

Finishing polish to make the CLI feel cohesive and complete (chose this over
glamour-rendering `show`, which loses raw-text fidelity + adds a heavy dep).

## Done

1. **`version`** — `tskflwctl version`, `--version` (cobra), and `--json version`
   (`schema_version` + `version`). Stamped from git via `-ldflags` in the
   Justfile (`build`/`install`); falls back to module build info. Works with no
   planning repo (own PreRun).
2. **Help examples** — cobra `Example:` on the main commands (`task
   new`/`list`/transitions, `epic new`/`list`, `audit list`/move, `lint`,
   `init`).
3. **Color consistency** — styled the previously-plain outputs (`init` ✔ + dim
   scaffold, `task set` ✔, `lint` "all pass" ✔) through `Style`.
4. **Next-step hints** — dim `→ next: …` after `init` (epic new), `epic new`
   (task new under it), `task new` (task start).
5. **`--no-color` + `FORCE_COLOR`/`CLICOLOR_FORCE`** — full color control with a
   clear precedence (flag → FORCE_COLOR → NO_COLOR → TTY), so **agents get
   deterministic color** independent of TTY detection.

## Acceptance

- [x] `version` (+`--version`/`--json`); examples in `--help`; styled output and
      hints; `--no-color`→plain, `FORCE_COLOR` forces color even when piped.
- [x] Tests: version subcommand/flag/json, color precedence (incl. NO_COLOR vs
      FORCE_COLOR vs flags). Full suite + lint green; git-stamped build demoed.

## Out of scope

- `glamour`-rendered `task show` body (filed separately; debatable value).
- The dashboard (its own sprint).

## Related

- Epic [[17-pm-go-cli]]; caps the color/formatting line that began with
  [[cli-color-glyphs-table-headers-render-styling]] and
  [[relative-dates-and-width-aware-truncation-in-task-list]].
