---
status: proposed
date: "2026-06-20"
deciders: [andy-esch]
tags: [adr, planning-model, project]
supersedes: []
superseded_by: null
---

# ADR-0002: Adopt Projects — cross-cutting groupings of tasks

> Follows the ADR format established in [[0001-adopt-adrs]]. Depends on it: this is a
> decision recorded *using* the convention 0001 introduces. Design rationale and the
> cross-tool product survey are in [[2026-06-20-adrs-and-projects-format-design]].

## Context and Problem Statement

Epics are **domain-specific homes**: durable clusters of related work — "ci/cd",
"visualization", "observability". A task belongs to **exactly one** epic, its
thematic home. That model is good for *taxonomy* but has a gap: **a cohesive push to
ship one larger thing routinely pulls tasks from several epics at once**, and some
tasks have no single "best" epic. There is no way to say "these 14 tasks, spread
across 4 epics, are the same initiative" — and no rollup that answers "how close is
that initiative to done?"

The singular `project:` tag from the `pm` prototype tried to fill this but was never
used (0 tasks carry it; `projects/` is empty in both repos), and its lone command was
a stub referencing a non-existent verb. So there is **no migration cost** — a clean
slate to model the concept properly, scoped here per [[0001-adopt-adrs]]'s opt-in
principle (a repo that doesn't need cross-cutting initiatives never creates
`planning/projects/`).

## Considered Options

- **A — Project as a first-class, opt-in document type, orthogonal to epics** (this
  decision). A task keeps its one `epic:` and gains zero-or-more `projects:`.
- **B — A validated `projects:` tag against a flat registry, no document type.**
  Cheaper, and a fine *interim*, but it can't carry a goal, a target date, a status,
  or a narrative — so it can't answer "what is this initiative and is it on track?",
  which is the whole point. Rejected as the end state.
- **C — Nest projects under epics (Jira/SAFe-style containment).** Rejected: it
  forces the one-parent tree the problem explicitly breaks (an initiative spans
  epics). Orthogonality is the requirement, not a nice-to-have.

## Decision

Adopt **Project** as an optional planning type: a cohesive, **completable**,
**cross-cutting** grouping of tasks that builds one larger feature and may pull tasks
from one or more epics.

### Orthogonal to epics (the core model)

- **Epic = a *place* work lives** (durable domain home; **one** per task; never
  "done"). **Project = a *finish line* work moves toward** (cross-cutting initiative;
  **zero-or-more** per task; completes). *If it can never be done it's an epic; if it
  has a target date and a definition of done it's a project.*
- Epic and project are **orthogonal axes over the same tasks** — not a hierarchy. A
  task with no obvious "best" epic still has exactly one (its least-bad home); the
  project stitches it together with its true initiative-siblings across epic
  boundaries.
- **Precedent:** GitHub Projects (an item can belong to multiple projects — the valid
  multi-membership model). Linear is cited only for its **status-vs-health
  separation**, *not* for membership: Linear restricts an issue to one project, so it
  is **not** a precedent for our many-valued model. (Terminology caution: SAFe's
  "Epic" means a funded time-bounded initiative — our *Project* — and its "Strategic
  Theme" is our *Epic*; the words are inverted there.)

### Membership lives on the task

A task carries `projects: [<slug>, …]` (already a registered field —
`internal/domain/fields.go` — today **unvalidated**; this ADR makes it validated
against `projects/`). The project's member list is **computed**, never
hand-maintained — same pattern as `epic show`. Single source of truth on the task; no
two-place drift.

### Lifecycle via verbs (not `set status`)

States: **`unstarted → in-progress → completed | abandoned`**. `completed` (not
`complete`) to match the epic/task vocabulary; `abandoned` is the project analog of a
task `deprecate` — cancelled, won't finish (Shape Up's "circuit breaker"; a
completable thing must be able to *not* complete). Transitions go through **verbs**
(`project start|complete|abandon`) that stamp dates — consistent with `task
start|complete` and with the repo's hard rule that `set` may not change status or
write stamps (`internal/domain/validate.go`). `project set` is for non-lifecycle
fields only.

### Rollup

`project show <slug>` computes `done / total` over `{task | slug ∈ task.projects}`,
member tasks **grouped by epic** to surface the cross-domain spread. Because the axes
are orthogonal, the project rollup and the epic rollup never share a denominator —
**no double-counting**, and project membership never affects epic rollups. Cost: like
`epic show`, this is O(all tasks) (no project→member index); acceptable and consistent
with existing behavior.

## Project frontmatter spec (decided)

```yaml
---
status: unstarted          # unstarted | in-progress | completed | abandoned
description: <one line, ≤150 chars>   # required for active; the triage primitive
goal: <the definition of done — what "shipped" means>   # one line
created: "2026-06-20"       # ISO, quoted (house style)
target_date: null           # the finish line; what makes it a Project not an Epic
started_at: null            # stamped by `project start`
ended_at: null              # stamped by `project complete|abandon`
tags: []
health: null                # optional: on-track | at-risk | off-track (≠ status)
spawned_by: null            # optional: the ADR that originated it, e.g. ADR-0007
---
```

- **Required:** `status`, `description`. **Recommended** (the traits that distinguish
  a Project from an Epic): `goal`, `target_date`. **Optional:** `health`,
  `spawned_by`. **`health` is a manually-set enum** — note it is *inspired by* but
  lighter than Linear's health (which is derived from periodic project updates); we
  don't model an update cadence.
- **No stored `epics:` list.** The set of epics an initiative touches is **always
  computed** from member tasks, never stored — honoring the earlier research's
  explicit "no `related_epics` (drift risk, low value)" decision.
- **No `owner`.** Single-user tool; omitted (add later if multi-user ever lands).
- Layout: `planning/projects/<slug>.md` — **flat, frontmatter status** (like epics, a
  grouping type), slug-only identity. Created only on opt-in (no reflexive directory).

## Cross-linking

| Link | Field / form | On | Validated |
| :-- | :-- | :-- | :-- |
| task → initiative(s) | `projects: [<slug>]` | task | yes (against `projects/`) |
| project → spawning decision | `spawned_by: ADR-NNNN` | project | yes (against `adrs/`) |
| project → epics touched | — | — | **computed, never stored** |
| project → member tasks | — | — | **computed** (no bidirectional write) |

No bidirectional-integrity hazard here: membership is computed, so there's only ever
one stored side (`task.projects`). `lint` validates each `projects:` slug resolves.

## CLI surface (sketch, for the implementation epic)

`project new "Title" --goal … [--target-date …]` · `project list|show` (cross-epic
rollup) · `project start|complete|abandon <slug>` (stamp dates) · `project set <slug>
<field> <val>` (non-lifecycle fields) · `project add|rm <slug> <task>…` (edits the
task's `projects:` list) · `task new … --project <slug>` / `task set --projects …`.
`schema` learns `project`; `lint` validates project frontmatter + `task.projects:`.

## Consequences

**Positive.** Cross-cutting initiatives get a cohesive, completable view and a real
rollup without distorting the stable epic taxonomy. Membership-on-task keeps one
source of truth. Verbs + computed rollups reuse the patterns tasks/epics already
established, so it fits the grain of the codebase.

**Negative / cost.** A second new type (command group, schema, lint, validated
`projects:`). `project show` is O(all tasks) (inherited from `epic show`). Risk of
under-use, mitigated by opt-in and by `project new` staying lightweight (only two
required fields).

## Amendments

_None yet (still `proposed`)._

## Related

- Format & convention this builds on: [[0001-adopt-adrs]].
- Design rationale & cross-tool survey: [[2026-06-20-adrs-and-projects-format-design]].
- Project model origin (orthogonal axes, membership-on-task, no `related_epics`):
  [[2026-06-06-project-concept-cross-cutting-initiatives]].
- Command spec (the `project` group): [[2026-06-06-tskflwctl-command-spec]].
