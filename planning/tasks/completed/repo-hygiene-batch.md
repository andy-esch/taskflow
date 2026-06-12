---
status: completed
epic: 17-pm-go-cli
description: Stray empty top-level planning dirs, tests/test_pm.py untracked while docs call it the spec, no LICENSE or CHANGELOG, dead .gitignore sections
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [hygiene, docs, git]
created: "2026-06-12"
updated_at: "2026-06-12"
started_at: "2026-06-12"
completed_at: "2026-06-12"
---
# Repo hygiene batch

> ⚠️ **Externally proposed — filed from the 2026-06-12 review**
> ([[2026-06-12-critical-code-review-multi-lens]], finding M19 + hygiene
> lows). Mostly deletions and doc truth-keeping; one git decision for the
> human.

## Objective

1. **Stray empty top-level dirs** `tasks/`, `epics/`, `audits/`, `projects/`
   at the repo root (untracked, empty). Without the root `.tskflwctl.toml`,
   walk-up discovery anchors to this empty skeleton instead of `planning/` —
   the misconfigured-root trap materialized. `rmdir` them. (Discovery-side
   guard is tracked in [[discovery-and-slug-edge-case-robustness]].)
2. **M19 — Docs point at a spec that is not in git.** README/CLAUDE.md say
   `tests/test_pm.py` is "kept as the historical executable spec", but only
   `bin/pm` is tracked — `tests/` is untracked, so a fresh clone loses it.
   Either commit it (and gitignore `.pytest_cache/`) or delete both relics
   and update README/CLAUDE.md together. **Needs the user's call.**
3. **No LICENSE, no CHANGELOG** — the module path is public-shaped
   (`github.com/andy-esch/taskflow`); add a LICENSE at minimum.
4. **`.gitignore` drift:** a large dead-Python section for the retired
   prototype, a stale "Docker & DB" section (`.taskflow/data/`) from the
   abandoned pre-rewrite design, and `bin/` is ignored while `bin/pm` is
   tracked inside it. Prune to match reality.
5. **Stale in-progress tasks:**
   [[port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec]] and
   [[tui-sprint-3-fsnotify-live-reload]] both look shipped (the latter in
   `f76254a`). Verify remaining checklist items, then `task complete` them.

## Acceptance criteria

- [x] Walk-up discovery from repo root cannot anchor to an empty skeleton.
- [x] Docs only reference files that exist in a fresh clone.
- [x] `.gitignore` matches the Go reality; LICENSE present.
- [x] in-progress contains only genuinely active work.

## Related

- Epic [[17-pm-go-cli]]
- Touches `.gitignore`, `README.md`, `CLAUDE.md`, `planning/tasks/`.
## Closure (2026-06-12)

Per decisions D10/D11/D12: `bin/pm` + `tests/test_pm.py` deleted (git history
`39f1b83` keeps them; README/CLAUDE.md updated to say so); stray empty
top-level `tasks/ epics/ audits/ projects/` skeleton removed; `.gitignore`
rewritten to Go reality (dead Python/Docker sections gone). LICENSE
intentionally **not** added (D11: private for now) — revisit before any public
release. Stale tasks `port-pm-…` and `tui-sprint-3-…` verified and completed
(closure note in the port task explains the descoped adr/schema items).
