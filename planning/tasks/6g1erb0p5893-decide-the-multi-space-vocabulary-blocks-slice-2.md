---
schema: 1
id: 6g1erb0p5893
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Pick and commit the user-facing nouns for epic 29 (item + collective). The only thing blocking the five slice-2 tasks; free to decide today, expensive once space add ships.
effort: Unknown
tier: 3
priority: high
autonomy_level: 3
tags: [design, cli, multi-repo]
created: "2026-08-18"
completed_at: "2026-08-18"
updated_at: "2026-08-18"
---
# Decide the multi-space vocabulary (blocks slice 2)

## Objective

Pick the user-facing nouns for epic 29 and commit them. This is the **only** thing
blocking the five remaining slice-2 tasks — it is the TOML key, the global flag, the
command noun, and the package name, all of which become user-facing contract the
moment the first one ships.

Slice 1 shipped deliberately free of this vocabulary (`internal/userconfig` names
nothing), so the decision is still cost-free today and expensive the day after
`space add` exists.

## What is already constrained (not up for debate)

- **"Project" is unavailable.** [ADR-0002](../adrs/0002-adopt-projects.md) defines a
  Project as a cross-cutting initiative *inside* one planning repo. A second, outer
  meaning would be the worst available ambiguity.
- **The internal type stays `Workspace`** regardless of the skin — the resolved
  config + store + service bundle for one root. Already the vocabulary in
  [reusable-workspace-discovery-seam](6fgcr2403sjn-reusable-workspace-discovery-seam-lift-init-doctor-fix-off-the-cli.md),
  and the audit-H1 work reuses it. The noun chosen here is the *user-facing skin* over
  that type, not a rename of it.
- **Two nouns are needed, and they need not share a metaphor** — the item (turns up
  everywhere, wants to be short and plain) and the collective (turns up once, on a
  screen, can carry the flavor).

## Where the item noun actually surfaces

Worth reading these aloud before committing — this is the real test, not the pitch:

```
tskflwctl space list
tskflwctl space add ~/git/andy-esch/desirelines-planning
tskflwctl space forget old-thing
tskflwctl task list --space desirelines
TSKFLW_SPACE=desirelines tskflwctl status
```
```toml
# ~/.config/tskflwctl/spaces.toml
[[space]]
id = "taskflow"
```
```
error: unknown space "desirelnies" — known: taskflow, desirelines, dotfiles
error: --space and -C are two answers to one question; pass one
```

## The candidates

| item | collective | the case for it | the case against |
| --- | --- | --- | --- |
| **space** | **atlas** | Short, plain, reads well in every line above. An atlas is literally a bound collection of separate maps — evocative without corn. | `space` is overloaded in dev tooling (namespace, workspace, HF Spaces, Confluence spaces), and sits awkwardly next to the internal `Workspace` type. |
| **orbit** | **constellation** | Most sci-fi without tipping into Trek; "constellation" is real ops vocabulary (satellites). | An orbit is a *path*, not a place — semantically off for "a repo". `tskflwctl constellation` is 11 characters. |
| **station** | **bridge** | Strongest verb metaphor: on a bridge you monitor everything and take the helm of one station. | "Station" is heavy for "a directory"; `bridge` collides with the networking/git senses. |
| **space** | **universe** | Takes the original instinct at face value; unambiguous. | Grandiose for three repos. |

## Decision aids

- **The item noun is the one that matters.** The collective appears on one screen and
  in one command; the item appears in config, a flag, an env var, and every error
  message for the life of the tool.
- **Test `space` against `workspace` specifically.** If `--space` and `core.Workspace`
  in the same codebase would make you hesitate when writing a doc comment, that is the
  signal to pick a different item noun.
- **The collective can be deferred** — the board (and therefore the name for "all of
  them") is not built until slice 3, and slice 2 needs only the item noun. Splitting
  the decision is legitimate if only one half is clear.

## Decision (2026-08-18)

**`space` (item) · `atlas` (collective).** Chosen by andy-esch.

- **`space`** is one planning repo/identity — the TOML key (`[[space]]`), the global
  flag (`--space`), the env var (`TSKFLW_SPACE`), and the command noun
  (`space list|add|forget`). The 2026-08-20 refinement below distinguishes the logical
  space from the registry rows that enter it.
- **`atlas`** is the whole set, and the name of the TUI screen over it (slice 3+).

Rationale: it reads best in the lines that matter — `--space desirelines` and
`unknown space "x" — known: …` are the phrasings you live with, and the alternatives
were all worse there. An atlas is literally a bound collection of separate maps:
evocative without corn, and it never appears in a config key.

**Rejected, and why:** `orbit`/`constellation` — an orbit is a path, not a place, and
`tskflwctl constellation` is 11 characters. `station`/`bridge` — "station" is heavy for
"a directory" and `bridge` collides with the networking and git senses.
`space`/`universe` — grandiose for three repos.

**The accepted cost:** `space` is overloaded in dev tooling and sits next to the
internal `core.Workspace` type. That is tolerable because the two never appear in the
same register — `Workspace` is the internal resolved bundle, `space` is the user-facing
registered repo — and the alternative was a worse item noun to protect a type name no
user ever sees. If it grates in a year, the item noun is a rename of one TOML key, one
flag, and one command group; the durable data (`id`, `path`) is unaffected.

### 2026-08-20 refinement — space identity vs registry entry

The noun is unchanged; its referent is now more precise. A **space is one durable
planning identity**, while each `[[space]]` row is a locally addressable **entry point**
(a direct planning checkout or a repo whose config points there). Several rows may
therefore belong to one space. This follows from the shipped config model: the planning
repo's `id` and its pointers' `planning_repo_id` are the same natural grouping key. The
distinction prevents `status --all` and the atlas from counting one task tree several
times without changing any accepted CLI spelling.

## Acceptance criteria

- [x] The item noun and the collective noun are chosen and recorded here
- [x] Epic 29, its five slice-2 tasks, and the research doc use the chosen vocabulary
- [x] The choice is stated with its rejected alternatives, so it is not relitigated

## Out of scope

- Any implementation — this task is the decision and its propagation through the
  planning docs only.
- Renaming `core.Workspace` / the `Resolve() → Workspace` seam.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch, with the full naming discussion and the open forks:
  [multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)
- The word that is unavailable: [ADR-0002](../adrs/0002-adopt-projects.md)
