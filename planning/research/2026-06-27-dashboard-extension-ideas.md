# Dashboard extension ideas (research)

Date: 2026-06-27 · Epic: [[18-tui-bubble-tea-interactive-planning-browser]]

Pragmatic ways to extend the TUI landing **dashboard** beyond its v1 widgets
(in-progress · due-for-revisit · epic rollups · needs-attention). Grounded in what
`core.Service` and the data model already expose, so each idea is tagged by how
much new plumbing it needs. Companion to the deferred items in
[[dashboard-review-follow-ups-tier-3-polish-and-deferred-decisions]].

## Lay of the land — what's cheap vs what needs plumbing

- **Already a render away.** `core.Summary` gives status counts, in-progress,
  epic rollups (with `Done/Total/Deprecated/LastUpdated/Percent`), open audits
  (with finding tallies), misfiled, revisit-due, and unreadable files. Anything
  built from these is a pure new render.
- **Rich, barely-used dataset:** audit **findings** carry `status`, `effort`, AND
  `urgency` (acute/soon/eventually), and `Service.QueryFindings(FindingFilter)`
  already returns them cross-audit (`internal/core/finding.go`). Today they're only
  shown one-audit-at-a-time. Real audits here populate urgency/effort reliably (50
  findings carry urgency), so an aggregate is meaningful. **This is the biggest
  free opportunity** and the chosen first build (see below).
- **A data gap bounds "progress over time":** `completed_at` is a known, validated
  frontmatter field but is **NOT decoded into `domain.Task`** (only `updated_at` +
  the `completed/` directory survive load), and there is no move history. So
  throughput / velocity / burndown / the deferred **Pulse** widget all need a small
  schema add (decode + stamp `completed_at`) before they're honest.
- **Config is fully writable** (`config.AddTrackedRepo`, `InitPointer`, `LinkBack`,
  surgical TOML edits) — but only from the CLI; the TUI has no config seam. So a
  TUI config editor is an architecture/wiring question, not a "build the writer" one.
- **The TUI's mutation + overlay machinery is reusable:** routing via
  `dashTarget`/`jumpTo`/`applyView`; the registry-driven action menu; the generic
  inline-edit form (`editMenu` + a `fieldSetter` closure); a new overlay/screen is
  ~50–150 LOC (one `modal` impl + one registry entry).

## Tier 1 — cheap wins (pure render of data core already returns)

| Idea | What it shows | Source | Effort |
|---|---|---|---|
| **Findings inbox / audit breakdown** ⭐ | Open findings across all audits, aggregated by **urgency** + status; acute highlighted; rows jump to the audit | `QueryFindings` | Small |
| **Flow / WIP strip** | The pipeline funnel (next-up → ready → in-progress → done); in-progress as your WIP load | `Summary.Counts` | Trivial |
| **Next-up queue** | The on-deck tasks — "what to pick up next" | `ListTasks(status=next-up)` | Small |
| **Audit debt by area** | Open audits grouped by `Area` ("Auth: 2 · DB: 1") | group `Summary.OpenAudits` by `.Area` | Small |
| **Stale in-progress flag** | Flag/color in-progress tasks untouched > N days (extends the staleness date already on those rows) | `updated_at` vs `Service.Now()` | Small |

## Tier 2 — small plumbing, high value

| Idea | What it shows | Needs | Effort |
|---|---|---|---|
| **Pulse: done recently** (deferred v1 widget) | "Completed this week" — momentum | decode + stamp `completed_at` (already a known field) | Medium |
| **Stalled / almost-done epics** | `active` epics with tasks but zero in-progress (stalled), or ≥90% (push to finish) | cross-ref `EpicSummary` × `InProgress` | Medium |
| **Misaligned work** | In-progress tasks whose epic is `retired`/`deprecated`, or in-progress with no epic | cross-ref `task.Epic` × epic status | Medium |
| **Priority/tier triage** | "High-priority, not started: 3" — `tier`/`priority`/`effort` parsed but unused in lists | iterate `ListTasks` | Small–Med |

## Tier 3 — bigger bets / outside-the-box

- **"My day" focus view** ⭐ — instead of mirroring entities, *synthesize* one
  ranked "do this now" list: stale in-progress + revisit-due + acute findings + top
  next-up. Turns the dashboard from a status board into a daily driver. Pure
  composition of Tier-1 sources.
- **Config / "About this repo" panel** (the brought-up idea) — phase it:
  1. **Read-only** "where am I pointed" panel first: resolved planning root,
     `planning_repo` vs in-tree, `tracked_repos`, pager config, and **`doctor`
     link-health** (`config.CheckLinks` already exists). High value for the
     sibling-planning-repo setup; cheap.
  2. **Editable** later (re-point `planning_repo`, manage `tracked_repos`, toggle
     pager) via the existing inline-edit form — once a TUI↔config seam is decided
     (today the TUI only talks to `core.Service`, which never touches config).
- **Quick-actions from the dashboard** — reuse the action menu / inline edit to
  `start` the top next-up, or `defer`/`complete` without diving into a tab. NOTE: a
  deliberate shift from v1's read-only/navigational stance.
- **Effort-weighted epic progress** — weight % by `effort` (XS/S/M/L) instead of
  task count, for truer progress. Needs a light effort→weight convention (effort is
  free-form today).
- **Projects / "goals" widget** — the marquee cross-cutting-progress widget, but
  **blocked**: the Projects entity is designed (ADR-0002,
  `planning/research/2026-06-20-adrs-and-projects-format-design.md`) and unbuilt.
  Build the entity first; the widget is then a Tier-1 render.

## Chosen first build — Findings inbox (audit breakdown, aggregated)

Deepen the existing `needs attention → "N open audit(s)"` line into *what's inside*
those audits, triaged.

**Data:** `loadDashboard` additionally calls
`QueryFindings(FindingFilter{Status: ["open","in-progress"]})` — the actionable
findings (open findings only exist in open audits per the bucket invariant). The
result rides along on `dashLoadedMsg` next to the `Summary`; refresh (`r`/fsnotify)
already re-runs `loadDashboard`.

**Render (aggregated):**
- A header tally: `findings (N open · M in progress)`.
- A by-urgency line: `⚠ A acute · S soon · E eventually` (acute red, soon yellow,
  eventually dim; missing urgency → an "other" bucket). Urgency order is the
  template convention acute → soon → eventually.
- Optionally list the **acute** findings individually (small, sharp set): `code ·
  title · audit`, each navigable.
- Empty state: `✔ no open findings`.

**Navigation:** rows route via `dashTarget{kind: entityAudits, id: <audit slug>}` →
`jumpTo` lands on the parent audit's detail (which already renders a glyph-coded
finding index). Deep-linking to a *specific* finding is a follow-up — the audit
detail has no per-finding selection today, and there is no findings-list view.

**Aggregation location — the one decision:**
- **(A, recommended) Tally in the TUI** from `QueryFindings` results. No change to
  `core.Summary` or its JSON envelope/schema; lowest blast radius. Pick this unless
  the CLI `status` also wants the breakdown.
- **(B) Add a `FindingsRollup` to `core.Summary`** so CLI `status` shows it too —
  but that touches the Summary JSON envelope + schema + docs regen (the
  docs-check gate). Bigger, do only if CLI parity is wanted.

**Reusable bits:** finding-status glyphs (`theme.FindingStatus`), the
measure-then-pad column alignment (`rollupCounts` / the dashboard date-column
pattern), and the existing `nav`/`dashTarget` routing.

**Tests:** a repo with an open audit whose body has findings of mixed
urgency/status — assert the widget shows the open/in-progress tallies and the
by-urgency counts, that acute is surfaced, and that selecting a row jumps to the
parent audit. Deterministic (no wall-clock dependence in the finding fields).

**Validated against real data (`desirelines-planning`, 19 open + 45 closed
audits):**
- **Filter by FINDING status, not audit bucket.** A real audit in `audits/open/`
  had findings that were *all* `superseded` — "open bucket" ≠ "has actionable
  findings." Querying `Status: [open, in-progress]` is the accurate inbox.
- Real actionable mix is `in-progress`-dominated (30) over `open` (6); the bulk of
  findings are `superseded` (103 — the workflow converts a finding into a task,
  then marks it superseded-by-task). An **out-of-vocab status `tracked`** occurs in
  the wild; it's handled for free (doesn't match the open/in-progress filter), but
  any *full* status breakdown must tolerate unknown statuses.
- **`acute` urgency is rare** (1 of ~145) → highlighting it is high-signal; `soon`
  (51) / `eventually` (93) dominate.
- **`Component` is richly populated and hierarchical** (`stravapipe / write paths`,
  `topology / postgres + bigquery`, `dispatcher`, `web`, …) → a **by-component
  (subsystem) breakdown** is a strong second aggregation axis, grouping on the
  top-level component (before the first ` / `). Likely as useful as by-urgency.
- Audit frontmatter also carries `routine` / `lens` / `iso_week` / `files_audited`
  (unparsed by `domain.Audit` today) — a future "audit coverage / freshness by
  area+lens" angle, out of scope for the inbox.

## Data gaps to be honest about

- **`completed_at` not decoded** → blocks honest Pulse/throughput until added
  (decode the known field + stamp it on Move→completed). Small but a prerequisite.
- **No move history** → no real burndown/cycle-time without an event log or git-log
  parsing (out of scope; relates to epic
  [[24-data-model-evolution-stable-key-storage-read-model-content-occ]]).
- **List fields** (`dependencies`/`blocks`/`blocked_by`/`related_tasks`/`projects`/
  `adrs`) are validated in the schema but **not decoded into structs** → any
  "blockers"/dependency widget needs domain plumbing first.

## Already planned — don't duplicate, build complementarily

- **Pulse** (recently completed/updated) and the **goals/Projects** widget are in
  the original dashboard plan (deferred) — see
  `planning/tasks/completed/add-a-tui-landing-dashboard-the-default-view.md`.
- `onDash` bool → screen-enum refactor was **decided against**; widget-registry
  refactor deferred until 5–6 widgets — see
  [[dashboard-review-follow-ups-tier-3-polish-and-deferred-decisions]].
- External-planning (`planning_repo`/`tracked_repos`/`doctor`) is **shipped** for
  local/sibling repos; remote backends are epic-24 territory.
