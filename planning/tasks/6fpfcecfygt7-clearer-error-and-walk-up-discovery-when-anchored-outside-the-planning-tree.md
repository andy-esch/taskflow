---
schema: 1
id: 6fpfcecfygt7
status: completed
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Make the 'no tasks/' failure name the path checked, the expected layout, and the -C / taskflow_root fixes, and decide walk-up discovery either way.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, config, discovery, agents]
created: "2026-07-15"
updated_at: "2026-08-23"
started_at: "2026-08-23"
completed_at: "2026-08-23"
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

- [x] The "no tasks/" failure names the path it checked, the expected layout,
      and the `-C` / `taskflow_root` fixes — legible to an agent.
- [x] Decide walk-up discovery: either implement bounded upward search that
      refuses ambiguous matches, or explicitly record it as out of scope with
      the path-anchoring rationale.
- [x] Suite + lint green; docs / error copy regenerated as needed.

## Related

- Epic [23-point-an-impl-repo-at-an-external-planning-repo](../epics/23-point-an-impl-repo-at-an-external-planning-repo.md).
- Discovery/config cousins:
  [discovery-honors-and-validates-planning-repo-out-of-tree](6fes83r010vs-discovery-honors-and-validates-planning-repo-out-of-tree.md),
  [config-robustness-symlink-safe-discovery-and-toml-escapes](6fes83r00ztg-config-robustness-symlink-safe-discovery-and-toml-escapes.md).
- Touches discovery/config in `internal/` + the CLI error copy.

## Implementation notes (2026-08-23)

Reproduced every scenario in the report before changing anything, and two of the three
complaints had already been fixed by later epic-23 work. What remained was one real bug the
report had not identified.

**Walk-up discovery already exists — criterion 2 is "implemented", not "out of scope".**
`Discover` climbs from the start directory trying `.tskflwctl.toml`, then `tasks/`, then
`planning/tasks/` at each level, terminating at a `.git` boundary or the filesystem root.
Verified both halves: it resolves from `outer/inner/src/deep` up to `outer`, and once
`inner` becomes its own git repo it correctly refuses rather than borrowing the outer
tree.

The AC asked for a search that "refuses ambiguous matches". No refusal is needed, because
ambiguity cannot arise: the climb takes the NEAREST match and stops, so exactly one tree
can win. Crossing a repository boundary is deliberately not walk-up's job — that is what
`planning_repo` and `--space` are for, and the `.git` stop is what keeps an impl repo from
silently adopting a parent's planning tree. The path-anchoring rationale the AC offered as
the alternative is therefore the reason the boundary exists, not a reason to skip the
feature.

**The real bug: a pointer and `-C` disagreed about the same directory.** For a config-less
planning repo whose entities live under `planning/`, `Discover` accepted it (its ladder
tries `planning/tasks/`) but `resolvePlanningRepo` rejected it, because that branch checked
only `tasks/`. So `-C ../plan2` worked while `planning_repo = "../plan2"` failed on the
identical directory — violating the invariant `resolvePlanningRepo`'s own doc comment
states: "a pointer at the repo root resolves the same tree `-C <target>` would". The
pointer now climbs the same ladder, and the subdir name is a single named constant so the
two cannot drift apart again. Pinned by
`TestPointerResolvesTheSameTreeAsAnchoringAtTheTarget`.

**Messages.** The three discovery failures were already better than the July report
described — they named the resolved path and a remedy — but none mentioned `-C`. All three
now name what was searched for, where, and every way out (`init`, `-C`, `--space`, or
fixing `taskflow_root`). Pinned by `TestDiscoveryFailuresNameThePathTheLayoutAndTheRemedies`.

Validation: `go test -race ./...`, `golangci-lint`, planning lint, generated docs,
`git diff --check` — all clean. Run the suite with `env -u FORCE_COLOR`.
