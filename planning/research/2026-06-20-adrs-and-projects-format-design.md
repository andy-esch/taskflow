---
date: "2026-06-20"
topic: Generic ADR and Project document formats for tskflwctl
purpose: >-
  Synthesize ADR best practices, cross-tool Project/initiative product research, and
  the two repos' house style into generic ADR + Project templates tskflwctl will
  scaffold, plus a cross-linking scheme. The decision record behind ADR-0001 (ADRs)
  and ADR-0002 (Projects).
status: proposed
related_adrs:
  - ADR-0001
  - ADR-0002
related_research:
  - 2026-06-06-project-concept-cross-cutting-initiatives.md
  - 2026-06-06-tskflwctl-command-spec.md
---

# ADRs & Projects: format design and decision record

Rationale behind [0001-adopt-adrs](../adrs/0001-adopt-adrs.md) and [0002-adopt-projects](../adrs/0002-adopt-projects.md). It synthesizes
three research streams — (1) online ADR best practices, (2) cross-tool product
research on "Project / initiative" types, (3) a house-style mining of
`taskflow/planning/` and `desirelines-planning/` — into two **generic** document
formats `tskflwctl` will scaffold for any repo, plus a cross-linking scheme. The two
repos are *fodder* (they set the house style the new types must match), not subjects
to restructure.

> **Revised after adversarial review (2026-06-20).** An earlier draft of this doc and
> a combined ADR-0001 contained errors caught by review and corrected here: the Linear
> "multiple projects per issue" claim was **false** (Linear is single-membership;
> GitHub Projects is the multi-membership precedent); "immutability is unanimous" was
> **false** (Joel Parker Henderson favors mutability in practice) — replaced by an
> append-only *amendments* model; the project status enum now uses `completed` (not
> `complete`) to match the repo; and the concept was **split into two ADRs**. See
> §Corrections.

## TL;DR decisions

| Question | Decision | Why |
| :-- | :-- | :-- |
| Split | ADR-0001 = adopt ADRs · ADR-0002 = adopt Projects | One decision per ADR; 0001 is acceptable on its own; 0002 builds on its format. |
| Support model | **Opt-in**, never reflexive (init offers; `new` lazily creates) | "Support, not require." Directly answers the desirelines deprecation. |
| ADR core sections | Title, Context, Decision, Consequences (+ Considered Options) | The irreducible core every lineage agrees on, plus MADR's one addition. |
| ADR mutability | Frozen on accept; append-only `## Amendments`; reversal → supersede | Honest "append-only" without rigid immutability; fits a solo/agent flow. |
| ADR status location | Frontmatter field, no `## Status` section | Matches MADR + the repo's frontmatter norm; avoids two-places drift. |
| ADR identity | `ADR-NNNN`, one canonical spelling everywhere; `NNNN-slug.md`, flat | Nygard numbering; behaves like epics (durable, flat, frontmatter status). |
| ADR numbering race | scan-max → exclusive-create (`O_EXCL`) → retry | No git, no `flock`; reuses `createFileAtomic`. |
| ADR home | `planning/adrs/`, hardcoded (not configurable for v1) | Keeps layout dirs the single-source domain constants; supersedes command-spec. |
| Project layout | `planning/projects/<slug>.md`, flat, frontmatter status | Orthogonal grouping like epics; matches prior research; not board-like. |
| Project lifecycle | `unstarted → in-progress → completed \| abandoned`, via **verbs** | `completed` matches repo; verbs stamp (set-can't-stamp rule). |
| Project ↔ Epic | Orthogonal (1 epic, 0..N projects per task) | GitHub Projects is the precedent; Jira/Asana/SAFe nest (rejected). |
| Membership | On the task (`projects: []`, validated); members computed | Single source of truth; rollup is a filtered done/total, no double-count. |
| Project `epics` list | **Computed, never stored** | Honors prior "no `related_epics`" decision (drift risk). |
| Cross-link integrity | One stored side where possible; `lint` is the backstop | Non-transactional, git-agnostic tool can't guarantee multi-file writes. |

## 1. ADR best practices (online research)

Sources: Nygard (2011, the canonical five sections), MADR (the Markdown tooling
standard; frontmatter since 3.0), Tyree & Akerman (IEEE 2005, the 14-field
heavyweight), Y-statements (Zimmermann's one-liner), adr-tools (Nygard's CLI), Joel
Parker Henderson's collection, plus AWS / Microsoft / Fowler / UK GDS guidance.

**Irreducible core (no source disagrees):** Title (numbered) · Status · Context ·
Decision · Consequences. The one strongly-recommended addition is **Considered
Options** (MADR core; Tyree's "Positions") — we scaffold it by default.

**On the default template name.** MADR's *minimal* template carries **no status at
all**. So our default is honestly "MADR-minimal **+ a frontmatter status we add for
lint**," not a named off-the-shelf format. Keeping status in frontmatter (not a
`## Status` section) is consistent with MADR 3.0+; Nygard/AWS/Microsoft instead keep
a prose Status section — we deliberately depart from them there to match the repo.

**Lifecycle:** `proposed → accepted → superseded | deprecated`, with `rejected`
reachable only from `proposed`. Status-set size varies across sources (Nygard 5, AWS
4, MS 3, MADR open-ended); we take the 5-state superset. **deprecated** = retired
without replacement; **superseded** = replaced by a *specific* new ADR.

**Mutability — NOT unanimous.** AWS/Microsoft say "immutable, append-only." But Joel
Parker Henderson — a cited lineage source — explicitly prefers **mutability in
practice**. For a solo, agent-driven workflow, strict "never edit an accepted ADR" is
mismatched (the agent will want to extend a record). Reconciliation (ADR-0001): the
decision body **freezes on accept**, and post-acceptance changes are **append-only
dated `## Amendments`** entries; only a genuine reversal triggers a superseding ADR.
This is literally append-only without the dogma.

**Numbering:** `NNNN-kebab-title.md`, 4-digit zero-padded (lexical = numeric),
"monotonic… never reused" (Nygard). The race this creates (ADRs are the only
integer-identity type) is solved without locks: scan max, exclusive-create, retry.
**Canonical id `ADR-NNNN`** everywhere (prose, wikilink, `supersedes`/`superseded_by`)
— the earlier draft's bare-number/slug/`ADR-NNNN` mix was a real interop hazard
(cf. adr-tools' "Superceded" typo, which broke its own link parser).

**Storage path** is the one real cross-source disagreement (`doc/adr` vs `docs/adr`
vs MADR's `docs/decisions`). We hardcode `planning/adrs/` for v1; a configurable dir
(e.g. an impl-repo `docs/adr`) is deferred because layout dir names are single-source
domain constants threaded through the watcher/completion/init.

## 2. Project / initiative product research

Surveyed Jira, Linear, GitHub Projects/Milestones, Asana, Shape Up, OKRs, SAFe.

**Two camps on epic↔project:** *nested* (Jira, Asana, SAFe — strict containment tree)
vs *orthogonal* (work lives in one home but groups into projects independently). Ours
is orthogonal.

**Multi-membership precedent is GitHub Projects, not Linear.** GitHub lets an item sit
in multiple projects — matching our many-valued `projects: []`. **Linear restricts an
issue to exactly one project** (the earlier draft's "Linear is the validated twin" was
wrong). Cite Linear only for two things it genuinely does: (a) **status and health are
separate fields** (lifecycle vs on-track/at-risk/off-track), and (b) manual
completion. Note Linear's *health* is derived from periodic project **updates**, not a
poked enum — so our manually-set `health` is *lighter than* Linear's, not "almost
exactly" it.

**Terminology clash (name it in docs):** in SAFe an *Epic* is a funded, time-bounded
initiative (≈ our Project) and the durable direction is a *Strategic Theme* (≈ our
Epic). Our usage is the colloquial/Linear usage; SAFe's is inverted.

**Lifecycle & rollup:** the industry baseline is a status field + binary `done/total`
rollup (GitHub/Asana milestones; Linear adds optional partial credit). We adopt
`unstarted → in-progress → completed | abandoned` (Linear's set collapsed; `abandoned`
= Shape Up circuit-breaker), spelled to match the repo's `completed`. Membership is
stored on the work item in every orthogonal tool; the project's member list and the
set of epics it touches are **computed** — we store neither (honoring the prior
"no `related_epics`" decision).

**Sharpest distinction:** *Epic = a place work lives (durable home, one per task,
never done); Project = a finish line work moves toward (cross-cutting, time-bounded,
completes; zero-or-more per task).*

## 3. House-style fingerprint (the two repos as fodder)

Mined ~14 tasks, 6 epics, 5 audits, research/incident/issue docs across both repos.
Conventions any new type must follow to feel native:

1. **Tool-first, schema-driven** — created/moved/edited via `tskflwctl <noun> <verb>`;
   frontmatter is whatever `schema` emits and `lint --fix` normalizes. No hand-`mv`,
   no ad-hoc fields.
2. **`status == directory`** is the invariant for *flowing* completable types (tasks,
   audits). *Durable/grouping* types (epics — and now ADRs, projects) go **flat +
   frontmatter status**. (This adds a status mechanism; accepted as a conscious cost,
   not silent drift.)
3. **ISO dates `YYYY-MM-DD`, quoted in taskflow**; stamps in frontmatter, not filenames.
4. **One-line `description` ≤150 chars**, required for active items (the `--json`
   triage primitive).
5. **Bare enums, bare `NN-slug` epic refs, inline `[tag, tag]` lists** (quote only if
   the value contains `:`).
6. **Filenames stable across lifecycle**; slug-only for slug-identity types,
   `NN-`/`NNNN-` numeric prefix for numbered types.
7. **Canonical body skeleton:** H1 prose title → Objective/Goal → Context →
   done/acceptance section *with `- [ ]` checkboxes* → Out of scope → Related.
8. **Append-only dated history in the body** (`## Progress Log`, dated `## Closure`),
   never a parallel tracking doc. (The ADR `## Amendments` mechanism is this pattern.)
9. **`epic:` is the only validated structured link today**; everything else
   (wikilinks, `related_tasks`, relative paths, supersedes) is unvalidated prose, with
   up to three coexisting syntaxes for one purpose. The new types validate their *own*
   Tier-1 fields; the broader cross-link cleanup is deliberately out of scope.

## 4. The two scaffolds

### `tskflwctl adr new "Title"` → `planning/adrs/NNNN-slug.md`

```markdown
---
status: proposed
date: "YYYY-MM-DD"
deciders: [you]
tags: [adr]
supersedes: []
superseded_by: null
---

# ADR-NNNN: <title>

## Context and Problem Statement
## Considered Options
- **Option A** — …
## Decision
## Consequences
**Positive.** …
**Negative.** …
## Amendments
<!-- append-only, dated, added AFTER acceptance -->
## Related
```

### `tskflwctl project new "Title"` → `planning/projects/<slug>.md`

```markdown
---
status: unstarted
description: <one line, ≤150 chars>
goal: <definition of done>
created: "YYYY-MM-DD"
target_date: null
started_at: null
ended_at: null
tags: []
health: null            # on-track | at-risk | off-track (≠ status)
spawned_by: null        # ADR-NNNN that originated it
---

# Project: <title>

## Goal
## Scope / Out of scope
## Milestones
- [ ] …
## Related
```

Tasks gain `projects: []` (already registered, now **validated**) and a **new**
`adrs: []` field (registry + struct + schema-doc + sync-test change).

## 5. Cross-linking scheme

```
   ADR ── the WHY (decision; frozen + amendments)
    │ spawns ▲ cites back (task.adrs / project.spawned_by)
    ▼        │
 Epic + Project ── the WHAT/HOW
    │ contain ▲ computed rollup
    ▼         │
   Task ──────┘ ── the WORK (epic: one home, projects: 0..N, adrs: cites)
    ▲
    │ spawns (audit finding → task)
  Audit
```

**Tier 1 — validated frontmatter:** task `epic:` (existing), task `projects: []`,
task `adrs: []`; project `spawned_by:`; ADR `supersedes:`/`superseded_by:`. The only
genuinely *bidirectional* link is ADR↔ADR supersession — and because the tool is
non-transactional and git-agnostic, the *write* attempts both sides but **`lint` is
the integrity backstop** that detects/`--fix`es a one-sided pair. Everything else has
a single stored side (membership and touched-epics are computed).

**Tier 2 — `[[ADR-NNNN]]` / `[[slug]]` prose refs** with one canonical form per type.
A general resolver + TUI `f`-nav + migration of legacy links is **out of scope** for
these two ADRs (a separate effort if ever justified).

## 6. The desirelines lesson (engaged, not ignored)

ADRs were tried in `desirelines` and the convention was **declined** (2026-05-26: "no
precedent in the repo, and the versioning policy lives directly in openapi.yaml").
The takeaway shaped two decisions: **opt-in** (the tool supports ADRs but never forces
a repo to adopt them), and **ADRs don't replace inlining** a narrow, locally-enforced
decision (Considered Option C in ADR-0001). Supporting the type ≠ requiring its use.

## Corrections (post-review changelog)

- Linear-as-multi-membership-precedent → **GitHub Projects**; Linear demoted to
  status/health-separation only.
- "Immutability is unanimous" → **append-only `## Amendments`** model (Henderson's
  mutability stance + the solo/agent reality).
- "MADR-minimal-plus-Status, the smallest no source would call wrong" → honest
  "MADR-minimal + an added frontmatter status."
- Project terminal state `complete` → **`completed`** (repo vocabulary).
- `project set status …` → **lifecycle verbs** (`start|complete|abandon`) that stamp.
- Stored `project.epics` → **computed only**.
- ADR identity unified to **`ADR-NNNN`**; numbering race spec'd (exclusive-create).
- ADR home **hardcoded** `planning/adrs/` (was a phantom `adr_dir` config); the
  command-spec's impl-repo line is **superseded** by ADR-0001.
- One combined ADR → **split** into ADR-0001 (ADRs) + ADR-0002 (Projects).
- `adrs:` flagged as a **new** task field (real schema change), not a footnote.

## References

- ADRs: [0001-adopt-adrs](../adrs/0001-adopt-adrs.md) · [0002-adopt-projects](../adrs/0002-adopt-projects.md).
- Project model origin: [2026-06-06-project-concept-cross-cutting-initiatives](2026-06-06-project-concept-cross-cutting-initiatives.md).
- Command spec (`adr` / `project` groups): [2026-06-06-tskflwctl-command-spec](2026-06-06-tskflwctl-command-spec.md).
- ADR lineage: Nygard (cognitect.com/blog/2011/11/15/documenting-architecture-decisions),
  MADR (adr.github.io/madr), Tyree & Akerman (IEEE Software 2005), adr-tools
  (github.com/npryce/adr-tools), Joel Parker Henderson
  (github.com/joelparkerhenderson/architecture-decision-record), AWS Prescriptive
  Guidance, Microsoft Azure WAF, Martin Fowler, UK GDS Way.
- Project lineage: GitHub Projects (docs.github.com), Linear (linear.app/docs), Jira
  Advanced Roadmaps, Asana, Shape Up (basecamp.com/shapeup), SAFe.
