---
status: ready-to-start
epic: 19-web-companion-apps-over-a-shared-core
description: 'Write a research doc capturing web-companion ideas: tskflwctl serve, core/app module split, GitHub store adapter; capture only, no research yet'
effort: 1-2 days
tier: 2
priority: low
autonomy_level: 3
tags: [web, architecture, research]
created: "2026-06-13"
id: 6fbwhsw00a4b
updated_at: "2026-11-08"
audited: "2026-11-08"
audit_sources: [2026-11-08-weekly-task-sweep]
---

# Research doc: web companion directions (serve, shared core, GitHub-backed store)

## Objective

Produce a research doc (under `docs/` or `planning/`) that evaluates directions for a
web companion to the TUI and a sustainable package structure to support it. This task
captures the ideas from the 2026-06 design conversation so they aren't lost; the doc is
where the actual thinking/research happens later. Motivating concern: planning data
should stay raw markdown and git-friendly, but discoverability on github.com is weak —
directory buckets are a crude browsing UI and force file moves on status changes.

## Ideas to capture (seed material, not conclusions)

**Web companion, in tiers of ambition:**

1. **`tskflwctl serve`** (preferred starting point) — local HTTP server as a third
   primary adapter over `core.Service`, peer of `cli` and `tui`. Server-rendered
   templates (html/template or templ, maybe htmx); no auth, no GitHub API. Cheap way to
   validate the web adapter and shared-code story; everything carries forward.
2. **Hosted web UI over a GitHub-backed store** — implement `core.Store`
   (`internal/core/store.go`, ~12 methods) against the GitHub API: reads via git
   trees/contents, mutations land as commits (arguably more git-friendly than local
   writes). Auth: fine-grained PAT for personal use; GitHub App + OAuth for multi-user.
   Needs ETag caching for rate limits.
3. **WASM client-only app** — `core` is fs-free so it compiles to wasm; static hosting
   (GitHub Pages), browser talks to GitHub API directly. Caveats: 5–10MB payload, OAuth
   needs device flow or a tiny token-exchange function. Bubble Tea-in-browser via
   xterm.js exists if pixel-identical look/feel is wanted.

**Sustainable structure ("core + bundled applications"):** early days — the question is
how to separate a reusable `tskflwctl` core (domain, core service, frontmatter parsing,
lint) from the applications that consume it (cli, tui, serve, future web). Options to
weigh: keep one module with clean `internal/` seams (status quo, already enforced by the
ports in `core`), promote core packages out of `internal/` for external consumers, or a
multi-module split. Bias: don't split prematurely; the `core.Store` /
`core.Service` ports are already the seam that makes any later split mechanical.

**Adjacent, much cheaper alternative to note in the doc:** a generated `BOARD.md`
kanban-style index (regenerated on every mutation, lint-checked) solves GitHub
browsing for ~1% of the effort. The web UI is justified by what a board file can't do:
interactivity, mutations from anywhere, non-CLI collaborators.

**Aside:** Charm's `wish` can serve the existing TUI over SSH with zero new UI code —
different question than GitHub discoverability, but worth a line.

## Acceptance criteria

- [ ] Research doc exists, covering: the serve-first path, package/module structure
      options with a recommendation, GitHub-backed store sketch, auth options, and the
      BOARD.md alternative as a baseline comparison
- [ ] Doc recommends a first concrete slice (likely `tskflwctl serve`, read-only) and
      what it would prove
- [ ] Follow-up tasks proposed under epic [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md)

## Out of scope

- Any implementation or prototyping (including `serve` scaffolding)
- Deciding the module split now — the doc weighs options, it doesn't restructure
- Auth implementation details beyond option comparison

## Related

- Epic [19-web-companion-apps-over-a-shared-core](../epics/19-web-companion-apps-over-a-shared-core.md)
- `docs/ARCHITECTURE.md` — ports-and-adapters layout this builds on
- `internal/core/store.go` — the `Store` port a GitHub adapter would implement

## Sweep verification (2026-09-06)

Automated weekly sweep. The deliverable is still outstanding — no web-companion
research doc exists under `planning/research/` (nearest neighbours are
`6ffdv9g00d53-task-storage-model-files-logs-or-versioned-db.md`, which epic 19's
body already names as the read-side seed, and
`6g2qtp0022t7-the-atlas-as-a-dashboard-of-dashboards.md`). AC 1–3 all unmet;
nothing ticked.

Two seed facts have moved since this was filed on 2026-06-13. Both are recorded
here rather than rewritten into the "Ideas to capture" section, because the doc is
where the thinking happens and the seed material should stay as filed.

### 1. Half the motivating concern has been retired by ADR-0003

The Objective's motivating concern reads: *"directory buckets are a crude browsing
UI and force file moves on status changes."*

The second clause is **no longer true**. ADR-0003 §4 landed (epic
`24-data-model-evolution-...`, 95%): tasks/audits/research/threads are stored
**flat and id-led** (`tasks/<id>-<slug>.md`), `status` is authoritative in
frontmatter, and a lifecycle verb edits frontmatter in place with **no file move**.
There is no status directory left to browse or to churn.

What survives is the first clause — GitHub's flat directory listing is still a poor
browsing UI for 300+ id-led filenames, and arguably a *worse* one now that the slug
is prefixed by an opaque id. The `BOARD.md` baseline the task already names is, if
anything, a stronger comparator than when it was written. Worth the doc restating
the problem in post-ADR-0003 terms before weighing options.

### 2. The `core.Store` port is ~3× the surface this task assumes

Tier 2 ("Hosted web UI over a GitHub-backed store") sizes the work as
*"implement `core.Store` (`internal/core/store.go`, ~12 methods)"*. As of HEAD
`84b3798`, `core.Store` composes four interfaces totalling **34 methods**:

- `TaskStore` — 10 · `EpicStore` — 7 · `AuditStore` — 10 · `ResearchStore` — 7

and threads live in a **separate `ThreadStore`** (`store.go:137`, 2 methods) that a
complete adapter would also need, plus the narrow `Fixer` / `Linter` /
`SummaryStore` ports beside it. A GitHub-API adapter is therefore a much larger
commitment than the "~12 methods" figure implies — which strengthens the doc's
existing bias toward `tskflwctl serve` (option 1) as the first slice, and makes the
read-only framing of that slice more important, not less.

The `1-2 days` effort is unchanged: the deliverable is still a research doc, and
these corrections are inputs to it rather than extra scope.

> **Update 2026-11-08**: the count is now **31**, not 34 — `TaskStore` 9, `EpicStore` 7,
> `AuditStore` 8, `ResearchStore` 7 (`internal/core/store.go:435`). The October
> identity-binding refactor (`33d6284`, `a05faad`, `db7ad6b`, `84a58c9`) folded the
> separate "selected" write methods into target-carrying signatures such as
> `SetFields(target TaskMutationTarget, …)`, which is why the surface shrank without
> losing capability. `ThreadStore` is still 2 methods, but now at `store.go:251`, not
> `:137`. The conclusion is unchanged: 31 is still ~2.6× the "~12 methods" the task
> assumes, so the bias toward a read-only `serve` first slice holds.

## Sweep verification (2026-11-08)

Automated weekly sweep. The deliverable is **still outstanding**: no web-companion
research doc exists under `planning/research/` — nothing matching web, serve,
companion or github. ACs 1–3 remain unmet and nothing is ticked. Epic
`19-web-companion-apps-over-a-shared-core` is still `active` at 50% (1/2), so the
committed project context has not moved either.

Both substantive points from the 2026-09-06 block still stand, with the one number
corrected in the blockquote above:

- ADR-0003's flat, id-led layout is unchanged, so the "file moves on status changes"
  half of the motivating concern stays retired, and the first clause (GitHub's flat
  listing is a poor browsing UI for 300+ id-led filenames) is still live. The task
  tree has grown to 459 files under `planning/tasks/`, which if anything strengthens
  the `BOARD.md` baseline the task already names as the comparator.
- `core.Store` is still far larger than the "~12 methods" the task assumes, and the
  narrow `SummaryStore` / `Fixer` / `Linter` ports (1 method each, `store.go:444`,
  `:450`, `:467`) sit beside it as before.

One housekeeping defect in the prior block, recorded for re-checkability rather than
actioned: it anchors its verification to "HEAD `84b3798`", which is not reachable in
this repository (`git show 84b3798` → unknown revision). The same hash is cited by
the 2026-09-06 blocks on `6g50ry3dhgw7` and `6g63jj1dh0sb` and by `6fes83r03rhs`.
Sweep annotations should cite a date, or a hash that survives on `main`.

## Progress Log

- 2026-09-06: automated weekly sweep — doc still unwritten; recorded that ADR-0003 retired the "file moves on status changes" premise and that `core.Store` is now 34 methods (+ThreadStore), not ~12.
- 2026-11-08: automated weekly sweep — doc still unwritten; corrected the `core.Store`
  surface from 34 to 31 methods and `ThreadStore` to `store.go:251`, and flagged that
  the prior block's `84b3798` anchor is unreachable on `main`.
