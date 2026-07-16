---
schema: 1
id: 6fpfcecfygt7
status: ready-to-start
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: ""
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, config, discovery, agents]
created: "2026-07-15"
---
> ⚠️ **Externally proposed — filed 2026-07-15** from an agent dogfooding
> session. The config side of decoupled planning is done (this epic shipped
> `planning_repo`; `desirelines-planning` moved its entities under `planning/`
> in [isolate-desirelines-planning-entities-under-a-dedicated-planning-directory](6fjvdf9t848k-isolate-desirelines-planning-entities-under-a-dedicated-planning-directory.md)).
> What remains is the *ergonomic* when the tool is run from the wrong place.

## Objective

Running `tskflwctl` from an impl repo (e.g. `../desirelines`) failed with
`planning_repo … has no tasks/`, because the target's entities live under
`planning/tasks/`, not `tasks/` at the root. Two friction points fall out:

1. **The error is opaque.** `has no tasks/` doesn't say what layout was
   expected, that `taskflow_root = "planning"` (or a `planning/` subdir) is the
   fix, or that `-C <path>` re-anchors. An agent can't self-correct from it.
   Make the error name the resolved path it checked, the expected layout, and
   the `-C` / `taskflow_root` remedies.
2. **No walk-up discovery (stretch).** The tool anchors to the physical
   cwd/root, so from a subdir of either repo it can't find the planning tree.
   Walking up to find the planning repo (or the `planning_repo` pointer) from
   anywhere in either tree would remove a whole class of wrong-cwd errors — and
   would defuse the persistent-shell `cd`-leak footgun the agent hit.
   **Design tension:** the tool deliberately anchors to physical paths
   (discovery is path-based everywhere — see epic 24 / the remote-backends
   research), and walk-up must never silently pick the *wrong* tree. Scope it
   carefully, or land the clearer error first and treat walk-up as a follow-on.

## Acceptance criteria

- [ ] The "no tasks/" failure names the path it checked, the expected layout,
      and the `-C` / `taskflow_root` fixes — legible to an agent.
- [ ] Decide walk-up discovery: either implement bounded upward search that
      refuses ambiguous matches, or explicitly record it as out of scope with
      the path-anchoring rationale.
- [ ] Suite + lint green; docs / error copy regenerated as needed.

## Related

- Epic [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md).
- Discovery/config cousins:
  [discovery-honors-and-validates-planning-repo-out-of-tree](6fes83r010vs-discovery-honors-and-validates-planning-repo-out-of-tree.md),
  [config-robustness-symlink-safe-discovery-and-toml-escapes](6fes83r00ztg-config-robustness-symlink-safe-discovery-and-toml-escapes.md).
- Touches discovery/config in `internal/` + the CLI error copy.
