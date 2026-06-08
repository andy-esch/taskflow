---
date: 2026-06-06
topic: Projects as first-class cross-cutting initiatives
purpose: Design `project` as an optional, cross-domain grouping of tasks (orthogonal to epics), and surface the open gaps before implementing.
status: in-progress
related_tasks:
  - tighten-pm-cli-ergonomics.md
  - phase-0.5-formal-tskflwctl-command-hierarchy-purpose-spec.md
---

# Projects: optional, cross-cutting initiatives

## The concept (from the user, 2026-06-06)

- An **epic** is **domain-specific** and long-lived (e.g. `14-observability`,
  `09-postgresql-database-backend`). A task has exactly one epic — its home.
- A **project** is a **cross-domain, bounded deliverable** — a "feature"
  that cuts across backend + database + infra + frontend + devops, and thus
  across *several* epics. (e.g. "Multi-user onboarding.")
- Projects are **optional**, and a task can belong to **zero or more**
  projects.

So **epic and project are two orthogonal axes** over the same tasks:
*epic = which domain it lives in* (1 per task); *project = which initiative(s)
it serves* (0..N per task).

## This overturns the current AI_README hierarchy

AI_README today says:

> **Epic** (long-term strategy) → **Project** (sprint/tactic, via `project:`
> tag) → **Task**

That linear nesting is **wrong** under the new model — a project is not
"inside" an epic; it spans them. Revise to:

> **Epic** (domain home, 1 per task) and **Project** (cross-cutting
> initiative, 0..N per task) are orthogonal groupings of **Tasks**.

Good news: **0 tasks use the `project:` field today** (verified) and the
only command is a stub `project list` (its "no projects" message even
references a non-existent `project start`). So there's **no migration** —
clean slate to formalize.

## Proposed model

- **Directory:** `projects/<slug>.md`, first-class like `epics/`. Slug =
  filename minus `.md` (formal, referenceable).
- **Membership stored on the task** (single source of truth):
  `projects: [<slug>, ...]` frontmatter list (optional). Validated against
  `projects/` slugs, exactly like `epic:` is validated against `epics/`.
  The project's task list is **computed** (rollup), never hand-maintained —
  same pattern as `epic show`.
- **Project frontmatter:** `description`, `tags`, `status` ∈
  {`unstarted`, `in-progress`, `complete`, `abandoned`}, `created`
  (nullable), `ended` (nullable — set when it reaches complete/abandoned).
  **No `owner`** (personal project). **No `related_epics`** (drift risk,
  low value — derive nothing, store nothing).
- **Lifecycle:** flat dir + `status` field, **not** status-subdirectories.
  **No auto-close** — the tool surfaces *remaining (incomplete) tasks*
  (`project show`, `task list --project <slug> --status …`); the agent/human
  closes explicitly by setting `status: complete` when none remain ("what's
  left in project X?" → "nothing" → close it). `ended` stamped on close.
- **Rollup:** `project show <slug>` lists member tasks **grouped by epic**
  (surfacing the cross-domain spread) and highlights how many are still
  open; `project list` shows progress. A task appears under both its epic
  rollup and each project rollup.

## Command surface

`project list | show <slug> | new "<title>" | set <slug> <field> <val> |
lint | add <slug> <task>... | rm <slug> <task>...`. `add`/`rm` edit the
task's `projects:` list; `set` changes project frontmatter (e.g.
`status complete` → also stamps `ended`). `task list --project <slug>`
filters tasks by membership (+ combine with `--status` for "what's left").

## Resolved decisions (2026-06-06)

1. ✅ **Slug-only** (no number prefix).
2. ✅ **No `related_epics`** — don't store, don't derive (drift risk, low value).
3. ✅ **Frontmatter:** `status` ∈ {unstarted, in-progress, complete,
   abandoned}, nullable `created` + `ended`, `description`, `tags`. **No owner.**
   `pm schema` gains a `project` type.
4. ✅ **`projects` (validated list)** replaces the old singular `project`
   tag; update `OPTIONAL_FIELDS`/schema + `pm lint` (resolve entries against
   `projects/`).
5. ✅ **Enforce slug uniqueness** across projects/epics in `lint`.
6. ✅ **No auto-close.** Surface remaining open tasks; the agent closes
   explicitly (`set status complete`, stamps `ended`). Interaction is
   conversational ("what's left in X?" → "nothing" → close).
7. ✅ **Single repo** — all tasks/epics/projects live in ONE repo (a
   dedicated planning repo *or* an impl repo with a taskflow subdir). No
   cross-repo task tracking; impl-code references in task bodies stay prose.
   → motivates a **`tskflwctl init`** to scaffold the tree + config (see
   the command spec + its task).

### Still to confirm (small)
- The 4th status name: **`abandoned`** proposed (stopped, won't complete —
  the project analog of an audit's `deferred`). OK, or prefer `paused`/`shelved`?
- **Where built first:** prototype in Python `pm` (cheap) then port, vs
  straight into `tskflwctl`. *Lean: prototype in pm.*

## References

- `research/2026-06-06-tskflwctl-command-spec.md` (the `project` group).
- `AI_README.md` "Hierarchy" section (needs the orthogonal-axes rewrite).
- `bin/pm` `cmd_project_list` (the stub to replace) + `OPTIONAL_FIELDS`
  (`project` → `projects`).
