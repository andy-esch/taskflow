---
date: 2026-06-20
topic: Generic ADR and Project document formats for tskflwctl
purpose: >-
  Synthesize ADR best practices, cross-tool Project/initiative product research, and
  the two repos' house style into generic ADR + Project templates tskflwctl will
  scaffold, plus a cross-linking scheme. The decision record behind ADR-0001.
status: proposed
related_adrs:
  - 0001-adopt-adrs-and-projects.md
related_research:
  - 2026-06-06-project-concept-cross-cutting-initiatives.md
  - 2026-06-06-tskflwctl-command-spec.md
---

# ADRs & Projects: format design and decision record

This is the rationale behind [[0001-adopt-adrs-and-projects]]. It synthesizes three
research streams — (1) online ADR best practices, (2) cross-tool product research on
"Project / initiative" types, (3) a house-style mining of `taskflow/planning/` and
`desirelines-planning/` — into two **generic** document formats that `tskflwctl`
will scaffold for any repo, and a cross-linking scheme. The two repos are *fodder*
(they set the house style the new types must match), not subjects to restructure.

## TL;DR decisions

| Question | Decision | Why |
| :-- | :-- | :-- |
| ADR core sections | Title, Status, Context, Decision, Consequences | The irreducible core *every* lineage agrees on (Nygard → MADR → AWS → MS → Fowler). |
| ADR default extra | Considered Options (+ Related) | The one section everyone wishes Nygard had; MADR core. |
| ADR optional (flags) | Decision Drivers, Confirmation, per-option Pros/Cons | MADR-full / Tyree heavyweight extras; keep the default ~1–2 pages. |
| ADR status location | Frontmatter field (canonical), no `## Status` section | Avoids two-places drift; matches the repo's frontmatter-everywhere norm. |
| ADR identity / layout | `planning/adrs/NNNN-slug.md`, flat, monotonic 4-digit | Nygard numbering rule; behaves like epics (durable, flat, frontmatter status). |
| ADR home | `planning/adrs/` (configurable to `docs/adr`) | Co-located knowledge type; the `pm` impl-repo precedent was never actually used. |
| ADR immutability | Append-only; supersede via a new ADR + bidirectional links | Unanimous across sources; the `supersede` command keeps both sides in sync. |
| Project layout | `planning/projects/<slug>.md`, flat, frontmatter status | Orthogonal grouping like epics; matches prior project research; not board-like. |
| Project lifecycle | `unstarted → in-progress → complete \| abandoned` | Linear's set collapsed to essentials; `abandoned` = Shape Up circuit-breaker. |
| Project ↔ Epic | Orthogonal axes (1 epic, 0..N projects per task) | Linear is the validated twin; Jira/Asana/SAFe nest, which we reject. |
| Membership | On the task (`projects: []`); member list computed | Every orthogonal tool does this; task stays single source of truth. |
| Progress rollup | Filtered `done/total` over members, independent of epics | Orthogonal axes never share a denominator → no double-counting. |
| Cross-linking | Tier 1 structured frontmatter (validated) + Tier 2 `[[wikilink]]` | One form per purpose; fixes today's three-syntaxes-for-one-thing drift. |

## 1. ADR best practices (online research)

Sources: Nygard (2011, the canonical five sections), MADR (the Markdown tooling
standard, introduced frontmatter in 3.0), Tyree & Akerman (IEEE 2005, the 14-field
heavyweight), Y-statements (Zimmermann's one-liner), adr-tools (Nygard's CLI),
Joel Parker Henderson's template collection, plus AWS / Microsoft / Fowler / UK GDS
adoption guidance.

**Irreducible core (no source disagrees):** Title (numbered) · Status · Context ·
Decision · Consequences. AWS reduces the very definition of an ADR to "the
decision, its **context**, and its **consequences**."

**The one strongly-recommended addition:** *Considered Options* — core in MADR and
Tyree ("Positions"), absent from bare Nygard. Captures the rejected alternatives so
future readers see the road not taken. We scaffold it by default.

**Lifecycle (canonical state machine):**

```
proposed ──accept──▶ accepted ──superseded-by-new-ADR──▶ superseded (links to replacement)
   │                    │
 reject               retire, no replacement
   ▼                    ▼
rejected            deprecated
```

Status-set size varies (Nygard 5, AWS 4, MS 3, MADR open-ended); we adopt the
5-state superset. Key semantic distinction everyone honors: **deprecated** = retired
*without* replacement; **superseded** = replaced by a *specific* new ADR.

**Immutability is unanimous.** "When the team accepts an ADR, it becomes immutable…
the team proposes a new ADR [that] supersedes the previous" (AWS). "Append-only log…
don't edit accepted records" (Microsoft). The only permitted edits are the status
line and trivial fixes. **Bidirectional links** ("Superseded by ADR-N" / "Supersedes
ADR-M") are the consensus mechanism — and the universal failure mode is updating
only one side, which is exactly why `adr supersede` should write both.

**Numbering:** `NNNN-kebab-title.md`, 4-digit zero-padded (so lexical sort = numeric
order), "monotonic… numbers will not be reused" (Nygard). Storage path is the one
real disagreement (`doc/adr` vs `docs/adr` vs MADR's `docs/decisions`) — hence we
make it configurable.

**Frontmatter caveat:** the *original* lineage (Nygard/AWS/MS/Fowler) carries status
in a prose `## Status` section, **not** frontmatter — YAML frontmatter is a MADR/
tooling convention. We adopt frontmatter because this repo already does everywhere,
but it's a deliberate choice, not a universal standard.

**What a generic scaffolder should emit by default vs. behind flags:** default to
*MADR-minimal-plus-Status* (Title, Status=proposed, Context, Considered Options,
Decision, Consequences) — the smallest format no source would call wrong — with
sequential `NNNN-kebab.md`, an auto-maintained index, and a bidirectional-supersede
command. Leave Decision Drivers, Confirmation, per-option Pros/Cons, and the
heavyweight (Tyree) / one-liner (Y-statement) templates as opt-in.

## 2. Project / initiative product research

Surveyed Jira, Linear, GitHub Projects/Milestones, Asana, Shape Up, OKRs, SAFe.

**Two camps on epic↔project:**
- **Nested** (Jira, Asana, SAFe): strict containment tree — Initiative ⊃ Epic ⊃
  Story. One parent per child.
- **Orthogonal** (Linear, GitHub): work lives in one home (team/repo) but groups into
  projects independently; a work item can be in multiple projects. **This is our
  model — Linear is the closest validated precedent.**

**Terminology clash (name it in the docs):** "Epic" is overloaded. In SAFe an *Epic*
is a large *funded, time-bounded initiative* (≈ our Project) and the durable
direction is a *Strategic Theme* (≈ our Epic). Our usage is the Linear usage; SAFe's
is inverted.

**Linear is the de-facto standard and maps onto our proposal almost exactly:**
- Status and **health** are *separate* fields (lifecycle vs. on-track/at-risk/
  off-track). We keep them separate; health is optional.
- Lifecycle: Backlog/Planned → In Progress → Completed/Canceled. Collapsed to
  essentials: `unstarted → in-progress → complete | abandoned`. `abandoned` is
  essential — a completable thing must be able to *not* complete (Shape Up's
  "circuit breaker").
- Membership stored **on the work item** (the many-valued, orthogonal axis), while
  the single-valued home (team/epic) is a field on the item. Progress = `done/total`
  over members (Linear gives in-progress 25% partial credit; the defensible baseline,
  matching GitHub/Asana milestones, is binary done/total).

**Irreducible Project fields:** `slug`, `title`/`description`, `status` required;
`goal` + `target_date` recommended (the time-bound + definition-of-done traits that
distinguish a Project from an Epic); `health`, `epics` touched, `spawned_by`,
milestones, `owner` optional. (The earlier project research dropped `owner` as a
personal-project simplification; we keep it optional for the generic tool.)

**Sharpest distinction derived:** *Epic = a place work lives (durable home, one per
task, never done); Project = a finish line work moves toward (cross-cutting,
time-bounded, completes; zero-or-more per task).*

## 3. House-style fingerprint (the two repos as fodder)

Mined ~14 tasks, 6 epics, 5 audits, research/incident/issue docs across both repos.
The conventions any new type must follow to feel native:

1. **Tool-first, schema-driven** — created/moved/edited via `tskflwctl <noun> <verb>`;
   frontmatter is whatever `schema` emits and `lint --fix` normalizes. No hand-`mv`,
   no ad-hoc fields.
2. **`status == directory`** is the invariant for *flowing* completable types (tasks,
   audits). *Durable/grouping* types (epics — and now ADRs, projects) go **flat +
   frontmatter status**. This is the structural tell that decided ADR and Project
   layout.
3. **ISO dates `YYYY-MM-DD`**, quoted in taskflow; stamps in frontmatter, not
   filenames.
4. **One-line `description` ≤150 chars**, required for active items (the triage
   primitive surfaced in `--json`).
5. **Bare enums, bare `NN-slug` epic refs, inline `[tag, tag]` lists** (quote only
   when the value contains `:`).
6. **Filenames stable across lifecycle**; slug-only for slug-identity types,
   `NN-`/`NNNN-` numeric prefix for numbered types.
7. **Canonical body skeleton:** H1 prose title → Objective/Goal → Context →
   done/acceptance section *with `- [ ]` checkboxes* → Out of scope → Related.
8. **Append-only dated history in the body** (`## Progress Log`, dated `## Closure`),
   never a parallel tracking doc.
9. **`[[wikilink]]` for prose refs; `epic:` is the only validated structured link
   today** — every other cross-link (wikilinks, `related_tasks`, `dependencies`,
   relative paths, supersedes) is unvalidated prose.

**The cross-linking problem (most important finding):** the same purpose has up to
**three coexisting syntaxes** — a related task appears as `related_tasks: [slug]`
(frontmatter), `[[slug]]` (body), and `[label](../bucket/slug.md)` (markdown link),
sometimes all in one file; the `.md` suffix comes and goes; `audit_sources` mixes
resolvable paths with prose labels. This is the drift the new scheme must end.

## 4. The cross-linking scheme

Information flow we want legible:

```
   ADR ── the WHY (immutable decision + rationale)
    │ spawns ▲ cites back (task.adrs / project.spawned_by)
    ▼        │
 Epic + Project ── the WHAT/HOW
    │ contain ▲ rollup
    ▼         │
   Task ──────┘ ── the WORK (epic: one home, projects: 0..N, adrs: cites)
    ▲
    │ spawns (audit finding → task)
  Audit
```

**Tier 1 — structured frontmatter (validated, tool-maintained bidirectionality):**

| Link | Field | On | Validated against |
| :-- | :-- | :-- | :-- |
| task → home | `epic:` (existing) | task | `epics/` |
| task → initiative(s) | `projects: []` | task | `projects/` |
| task → decision(s) | `adrs: []` | task | `adrs/` |
| project → decision | `spawned_by:` | project | `adrs/` |
| project → epics touched | `epics: []` | project | `epics/` (or derived from members) |
| ADR ↔ ADR | `supersedes:` / `superseded_by:` | ADR | `adrs/` |

**Tier 2 — `[[wikilink]]` in prose, one canonical form:** no `.md`; slug for tasks/
projects/research, `NN-slug` for epics, `ADR-NNNN` for ADRs. `lint` resolves them;
the TUI `f`-nav follows them (closing the gap the cross-link-navigation task
explicitly deferred).

**Relative paths only for implementation code** (`internal/...`, `../desirelines/...`),
never for sibling planning docs. Drop the inline `[slug.md]` and mixed
`related_tasks` styles.

## 5. The scaffolds (what the tool emits)

### `tskflwctl adr new "Title"`

```markdown
---
status: proposed
date: "YYYY-MM-DD"
deciders: [you]
tags: [adr]
supersedes: []
superseded_by: []
---

# ADR-NNNN: <title>

## Context and Problem Statement
<!-- The forces at play (technical, project, social), value-neutral. -->

## Considered Options
- **Option A** — …
- **Option B** — …

## Decision
<!-- "We will …" — active voice. The chosen option, and *because*. -->

## Consequences
**Positive.** …
**Negative.** …

## Related
<!-- Spawns: [[NN-epic]] · Supersedes: ADR-MMMM · Cited by: task-slug -->
```

### `tskflwctl project new "Title"`

```markdown
---
status: unstarted
description: <one line, ≤150 chars>
goal: <the definition of done — what "shipped" means>
created: "YYYY-MM-DD"
target_date: null
started_at: null
ended_at: null
tags: []
health: null            # on-track | at-risk | off-track (≠ status)
epics: []               # thematic areas this cuts across
spawned_by: null        # the ADR that originated it, e.g. ADR-0007
---

# Project: <title>

## Goal
## Scope / Out of scope
## Milestones
- [ ] …
## Related
```

Tasks gain optional `projects: []` and `adrs: []` in frontmatter (the `projects`
list replaces the old singular `project` tag).

## 6. Implementation path (for the spawned epic)

(a) ADR support — `planning/adrs/`, `adr` command group, schema, lint, race-free
numbering; (b) Project support — `planning/projects/`, `projects:`/`adrs:` task
fields, `project` command, cross-epic rollup; (c) cross-link Tier-1 validation +
`[[wikilink]]` resolver + TUI `f`-nav; (d) docs + `schema adr|project` + generated
CLI reference. Scoped to start only once [[0001-adopt-adrs-and-projects]] is accepted.

## References

- ADR: [[0001-adopt-adrs-and-projects]].
- Project model origin (orthogonal axes, membership-on-task): [[2026-06-06-project-concept-cross-cutting-initiatives]].
- Command spec (the `adr` / `project` groups): [[2026-06-06-tskflwctl-command-spec]].
- ADR lineage: Nygard (cognitect.com/blog/2011/11/15/documenting-architecture-decisions),
  MADR (adr.github.io/madr), Tyree & Akerman (IEEE Software 2005), adr-tools
  (github.com/npryce/adr-tools), Joel Parker Henderson
  (github.com/joelparkerhenderson/architecture-decision-record), AWS Prescriptive
  Guidance, Microsoft Azure WAF, Martin Fowler, UK GDS Way.
- Project lineage: Linear (linear.app/docs), Jira Advanced Roadmaps, GitHub Projects,
  Asana, Shape Up (basecamp.com/shapeup), SAFe (framework.scaledagile.com).
