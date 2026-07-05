---
status: ready-to-start
epic: 20-cli-ux-and-ergonomics
description: 'DRAFT: a readiness axis (draft/finalized) orthogonal to status, tied to an open-questions checklist, possibly gating promotion to in-progress'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [planning-model, frontmatter, domain, draft]
created: "2026-06-12"
updated_at: "2026-06-21"
id: 6fbj87001m03
---
# Task readiness state (draft vs finalized) in frontmatter

> 🚧 **DRAFT — a sketch of an idea, not yet scoped.** Filed 2026-06-12 from
> the observation that four real draft tasks just had to improvise the
> mechanism (banner + `draft` tag + `DRAFT:` description prefix). This task
> is itself the kind of task it describes: it must not be picked up until
> its Open questions below are resolved. Self-demonstrating on purpose.

## Objective

Tasks have a lifecycle axis (`status` == directory) but no **definition
axis**: nothing distinguishes "fully specified, pick me up" from "sketch
with unresolved questions." Today that's improvised per-file. Make it a
first-class, tool-visible concept so:

- `task list` (and the TUI, and agents via `--json`) can tell ready work
  from drafts at a glance;
- a draft can't be started by accident — there is some gate between
  "questions open" and "in progress";
- the open-questions themselves are structured content the tool can count,
  not prose only a human notices.

## Design sketch (NOT decided — see Open questions)

- A `## Open questions` body section with checkboxes, machine-countable the
  way audit bodies already are (`auditstore.go` counts `#### H1.` headers
  and open `**Status:**` markers — content-derived state has precedent).
- Readiness is either a declared frontmatter field (`readiness: draft |
  finalized` or similar), derived purely from the unchecked-question count,
  or **declared + lint-checked against derived** — mirroring the existing
  status-vs-folder pattern (declared value, authoritative source, drift
  surfaced as ⚠ misfiled).
- A gate on promotion: `task start` (or next-up → in-progress generally)
  refuses — or warns — while questions are open / readiness is draft.

## Open questions

- [ ] **Field vs directory vs derived?** A new frontmatter field keeps
      status==directory sacred (a draft can still sit in next-up by
      priority). A `tasks/draft/` directory fits the invariant but conflates
      the two axes. Pure derivation needs no field but makes readiness
      invisible in raw frontmatter. Hybrid (declared + lint cross-check)
      matches the repo's existing declared-vs-folder pattern. Pick one.
- [ ] **Vocabulary:** `draft|finalized`? `draft|ready`? Is there a third
      state (e.g. `needs-review` between them)? Keep it minimal.
- [ ] **What exactly does the gate block?** `task start` only? Any move into
      in-progress? next-up → in-progress specifically (per the originating
      idea)? Hard error vs warning vs `--force` escape?
- [ ] **Exit code:** a readiness gate is the first REAL transition rule —
      D3 ([2026-06-12-pending-decisions](../research/2026-06-12-pending-decisions.md)) retired exit code 12
      (invalid-transition) but left it *reserved*. Does this feature
      un-retire 12, or is it a plain validation failure (11)?
- [ ] **Is the open-questions count authoritative for the gate, or the
      readiness field?** (i.e. can a task be `finalized` with unchecked
      questions? Can it be `draft` with none?) Define the invariant lint
      enforces.
- [ ] **Section format:** exact heading + checkbox convention the counter
      recognizes (and tolerates absent — most tasks have no questions).
      Fenced code blocks must not count (the audit counter already solves
      this with fenceRe).
- [ ] **Tooling surface:** `task new --draft`; how readiness is changed
      (`task finalize`? `task set --readiness`? auto when the last box is
      ticked?); `task list` rendering (chip/glyph) and a `--ready`/
      `--drafts` filter; TUI display; `--json` field (minor schema bump);
      should agents get a default that EXCLUDES drafts from pickup lists?
- [ ] **Lint rules:** is a draft in next-up a lint warning (priority says
      "soon", readiness says "not defined")? Is an unresolved question in a
      completed task an error?
- [ ] **Migration:** the four existing draft tasks (banner + `draft` tag +
      `DRAFT:` prefix) move to the new mechanism; does the `draft` tag stay
      as a search convenience or go?
- [ ] **Registry impact:** new field joins `domain/fields.go` (known fields,
      maybe an enum like epic status) and the future
      [schema-command-for-agent-self-discovery](6fbj87003e4g-schema-command-for-agent-self-discovery.md) output.

## ⚠️ Conflicts to resolve before starting

- **D3 interaction** (exit 12 reserved — see Open questions): touching it
  reopens a same-day decision; do that explicitly, not incidentally.
- The four draft tasks already in flight use the improvised convention —
  whatever is built must subsume it, and their "de-draft" acceptance
  criteria should then reference the real mechanism.
- `status == directory` is a CLAUDE.md non-negotiable: any option that adds
  a readiness *directory* needs an architecture-level sign-off, not just a
  task decision.

## Acceptance criteria (draft)

- [ ] All Open questions above resolved and recorded (here or in a
      decisions doc); task de-drafted.
- [ ] Then: re-scope this task into an implementable spec with its own
      acceptance criteria.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md) · [2026-06-12-pending-decisions](../research/2026-06-12-pending-decisions.md) (D3) ·
  the four 2026-06-12 draft tasks (motivating instances) ·
  `internal/store/auditstore.go` (content-derived counting precedent) ·
  `internal/domain/fields.go` (field registry).

## Reassigned off the port (2026-06-21)

Moved out of epic 17 (the pm→tskflwctl port): this is a NEW planning-model idea, not
pm parity, so it must not gate closing the port. Kept alive here as a draft. Like
Projects/ADRs, it is really a planning-*model* change — a natural future candidate to
be proposed as an ADR rather than slipped in as a task.
