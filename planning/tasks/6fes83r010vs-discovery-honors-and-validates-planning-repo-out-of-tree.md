---
schema: 1
status: completed
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Discover follows planning_repo out of tree (the sanctioned escape) and errors if the target is not a real planning root.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [config, discovery]
created: "2026-06-22"
started_at: "2026-06-22"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r010vs
---
# Discovery honors + validates `planning_repo`

The load-bearing behavior change. `config.Discover`/`configuredRoot` currently
**reject** any root that escapes the config file's own tree (the `..` containment
check at `config.go:74`) — the "don't fork the data" guardrail. `planning_repo`
is the *sanctioned* escape: when set, discovery follows it out of the tree.

## Scope

1. In `configuredRoot` (or a sibling), when `planning_repo` is set: resolve it
   relative to the config dir (absolute allowed; `..` **permitted here**),
   then **validate** it's a real planning root (has `tasks/`) and **error**
   loudly otherwise — matches the "require + error" decision.
2. Precedence: when both `planning_repo` and a non-default `taskflow_root` are
   set, `planning_repo` wins (ignore `taskflow_root` with a `⚠`, since it
   defaults to "." and "both set" is hard to distinguish from default).
3. Leave the existing in-tree path (no `planning_repo`) untouched, including the
   `..`-escape rejection for `taskflow_root`.

## Acceptance criteria

- [ ] `planning_repo = "../sibling"` resolves and is followed out of tree.
- [ ] A `planning_repo` that doesn't exist / has no `tasks/` is a loud error
      with guidance (run `init` there first).
- [ ] `taskflow_root` still cannot escape the tree.
- [ ] Discovery from inside an impl repo lands on the external planning root.
- [ ] Suite + lint green.

## Depends on

- Config schema + TOML parser (the `PlanningRepo` field).

## Related

- [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md).

## Resolution (2026-06-22)

`Discover` now follows `planning_repo` out of tree via `resolveRoot`/`resolvePlanningRepo`: relative-to-config (absolute allowed), validated to contain tasks/ or it's a loud error. The `..` containment check still guards `taskflow_root` — planning_repo is the *sanctioned* escape. Raw planning_repo + tracked_repos still ride along for linkback checks.

**Deviation from scope — conflict is an error, not a ⚠.** The task floated 'planning_repo wins, ignore taskflow_root with a ⚠'. The config package has no warning channel (warnings go through app.ErrOut — that's the T5 CheckLinks/doctor work). So: planning_repo wins over a default/empty taskflow_root silently, but a *non-default* taskflow_root alongside planning_repo (two genuinely different roots) is a loud error ('keep one'). Safer than a silent override. Ambient ⚠ warnings land in T5.

e2e verified: impl repo with planning_repo="../planning" resolves task ops to the sibling.
