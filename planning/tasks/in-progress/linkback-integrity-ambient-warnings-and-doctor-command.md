---
schema: 1
status: in-progress
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Ambient warning on link mismatch via config.CheckLinks in resolve(); plus tskflwctl doctor for an on-demand bidirectional audit (nonzero exit for CI).
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, config, doctor]
created: "2026-06-22"
updated_at: "2026-06-23"
started_at: "2026-06-23"
---
# Linkback integrity — ambient warnings + `doctor`

Keep the two sides honest: passive `⚠` warnings on normal runs, plus an explicit
on-demand audit.

## Scope

1. **`config.CheckLinks(cfg, startDir)`** called from `app.resolve()` right after
   `Discover` succeeds — ambient on every command that resolves a planning repo
   (`init`/`schema`/`version` run their own PreRunE and skip it). Non-fatal: one
   `⚠` line per finding to `app.ErrOut` (stderr, so `--json` stdout stays clean),
   using the established `app.Style.Warn("⚠")` idiom (see `edit.go:71`).
   - Resolved *via* an impl `planning_repo` → planning repo's `tracked_repos`
     doesn't list this repo back → "one-sided link."
   - Resolved *inside* a planning repo → a `tracked_repos` entry that doesn't
     exist, or whose `planning_repo` doesn't point back here.
   - Compare on **physical** paths to avoid false positives (the trap).
   - Suppress via `TSKFLWCTL_NO_LINK_WARN=1`; silent when there are no links.
2. **`tskflwctl doctor`** (new command): full bidirectional audit on demand,
   prints every inconsistency and **exits nonzero** when links are broken — a CI
   gate. Maps to a domain sentinel for the exit code.

## Acceptance criteria

- [ ] One-sided / dangling / mismatched links each produce a `⚠` on normal runs.
- [ ] Relative vs. absolute vs. symlinked paths that resolve equal produce **no**
      warning (false-positive guard test).
- [ ] `TSKFLWCTL_NO_LINK_WARN=1` silences ambient warnings.
- [ ] `tskflwctl doctor` audits both directions and exits nonzero on breakage;
      `--json` output.
- [ ] Suite + lint green.

## Depends on

- `tracked_repos` seeding + link-back (needs both sides populated).

## Related

- [[23-point-an-impl-repo-at-an-external-planning-repo]].
