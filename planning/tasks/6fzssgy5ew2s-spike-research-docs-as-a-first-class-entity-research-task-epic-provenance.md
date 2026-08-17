---
schema: 1
id: 6fzssgy5ew2s
status: in-progress
epic: 28-first-class-entities-new-planning-nouns
description: 'Explore research as a first-class noun: frontmatter contract, research new|list|show, and epic:/tasks: cross-refs for provenance rollups — without adding a lifecycle (ADR-0001 tension).'
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [core, design]
created: "2026-08-13"
updated_at: "2026-08-14"
started_at: "2026-08-14"
---

# Spike: research docs as a first-class entity + research↔task/epic provenance

## The idea

`planning/research/` is the oldest dir in the tree and the only one the tool has no
opinion about. 28 docs live there with no scaffold, no query surface, and no lint
coverage — so their conventions have drifted, and the provenance they carry ("this
design decision came from that spike") exists only as prose links. Make **research**
tenant #2 of this epic: give it a frontmatter contract, a read surface, and
*queryable* cross-references to the tasks/epics it backs.

## Crucial framing — this must not become a second ADR

[ADR-0001](../adrs/0001-adopt-adrs.md) rejected extending `research/` as the "why" home, on the grounds that
research docs are *"open-ended spikes with no lifecycle, numbering, or status
discipline."* That characterization is load-bearing for ADRs existing at all, so the
spike's job is **not** to overturn it. The split to hold:

- **ADR** — one decision, frozen on `accepted`, append-only amendments. Has a lifecycle.
- **research** — an exploration snapshot, true as of its date, freely superseded by a
  later doc. **Deliberately has no lifecycle.**

So the value being chased is **discoverability + provenance**, not status discipline.
If the design starts wanting lifecycle verbs, that's the signal it has drifted into
ADR's lane.

## Decisions (2026-08-14)

Settled up front, so the spike explores the remaining design surface rather than these:

1. **Ambition — scaffold + read surface only.** `research new|list|show` + a frontmatter
   contract + lint coverage. No status vocabulary, no lifecycle verbs.
2. **No cross-reference frontmatter — not even `epic:`.** Other entities link *to*
   research docs with ordinary body links (already cascade-safe and dangler-checked);
   research does not carry back-references. Provenance stays a body concern, so there is
   no rollup, no resolved reference, and no many-to-many modelling problem. This is the
   one thing to hold the line on if the design starts creeping.
3. **Naming — id-led, like tasks.** Drop the `YYYY-MM-DD-<slug>` convention for
   `<id>-<slug>.md` per ADR-0003, so research is flat and id-led like every other
   entity. No exception carved out.
4. **Ids are backdated from each doc's real date**, so lexical id order stays true
   chronological order across the legacy corpus (see below — no new code needed).
   **Relative order within a single date does not matter** — random intra-day order from
   `NewAt`'s low bits is accepted, including for the 9-doc 2026-01-03 batch.
5. **Migration renames and backfills all 28 docs.** Full frontmatter contract,
   best-effort with defaults where a value can't be recovered.
6. **Executed by a one-shot tool in `internal/tools/`** — throwaway, dry-run by
   default, git is the undo. Not a `tskflwctl` subcommand.
7. **No new lint warn severity.** Backfill best-effort; where a recovered value doesn't
   fit the frontmatter contract cleanly, keep it as prose near the top of the doc rather
   than forcing it into a field. `FileProblem` stays `{Path, Message}`.

## Current state (audited 2026-08-13)

28 docs in `planning/research/` — **18 with frontmatter, 10 bare; 19 date-led names**:

- **18 date-led with frontmatter** — `2026-06-28-color-palette-and-theming-overhaul.md`
  carries `status: reference`, `created`, `tags`.
- **10 bare** — no frontmatter at all. One is date-led
  (`2026-06-27-dashboard-extension-ideas.md`); the other 9 are bare-slug
  (`hybrid-search-architecture.md`, `ai-provider-abstraction.md`,
  `monorepo-structure-proposal.md`, …) and carry prose `**Status**: Proposal` /
  `**Created**: 2026-01-03` headers instead.

Names are **not** id-led, so ADR-0003's flat/id-led rule and the `errNotEntity`
carveout gate don't reach the dir. `lint` doesn't scan it at all.

## The migration is a near-copy of `flatmigrate`

`internal/tools/flatmigrate/` already did this exact job for tasks/audits and is the
template — copy its shape rather than inventing one:

- **`id.NewAt(unixMilli)`** (`internal/id/id.go:86`) is the backdating mint, and its doc
  comment names this use case verbatim: *"minting ids for known entities out of natural
  time order: the migration backfilling existing files from their created: date."* It is
  stateless and deliberately does **not** disturb the process monotonic counter.
  **No new code in `internal/id` — the API already exists.**
- **`mintUnique(millis, seen)`** — regenerates on the same-millisecond collision
  (`NewAt`'s low 17 bits are random). Heavily needed here: the 28 docs resolve to only
  **12 distinct dates**, and dates are day-precision, so most ids collide on the
  millisecond — 9 docs on 2026-01-03, 4 on 2026-06-23, 4 on 2026-06-06, 2 each on
  2026-06-12 and 2026-06-24.
- **`checkCollisions(renames)`** — fails the run on a duplicate target path or id.
- **`firstDate(content)`** — but it scans *frontmatter keys only*. The research corpus
  needs two more sources, in this precedence: **(a)** the `created:` field (18 docs),
  **(b)** the `YYYY-MM-DD` prefix in the filename (19 docs, incl. one bare), **(c)** a
  prose `**Created**: …` header (the 9 remaining bare docs, all → 2026-01-03).
  **Every doc resolves to a date — there are zero dateless docs, so no fallback is
  needed.** Verify this still holds at migration time; if it does, the tool can treat
  "no date" as a hard error exactly like `flatmigrate` does.
- Dry-run default, `-apply` to write, refuses a dirty git tree unless `-force`.

**Data-loss guard:** `2026-06-27-dashboard-extension-ideas.md` is bare *and* date-led —
its date exists **only** in the filename. `created:` must be backfilled from the filename
*before* the rename drops it. Any doc whose date lives only in its name has the same
hazard, so date extraction must run before renaming, not after.

**Link cascade is already handled** for free: `RenameTask`'s cascade
(`store/rename.go`) and `DanglingLinks` (`store/danglers.go`) both `WalkDir` the whole
planning root, and `lint --links` is currently clean — so it's the post-migration check
that proves no cross-reference broke.

**Non-issue:** `wikimigrate`'s `splitFlatName` would start treating research stems as
id-led (keying the index by human slug instead of whole stem), but that tool already ran
and is throwaway.

## What to explore

- **The frontmatter contract** — the field set `research new` scaffolds and lint checks.
  Is `status: reference` (which 17 docs already carry) a real vocabulary or vestigial and
  better dropped? A `supersedes:`/`superseded_by:` pair is how a research doc gets
  obsoleted without a lifecycle — is that enough?
- **Registry fan-out** — `domain/entity.go`'s `Descriptor` is built for this (one entry:
  kind, dir, `AuthoringFields`, `Conventions`, `Templates`, `Placeholders`), and
  `layout.go` has an unused `ProjectsDir` precedent. But the registry's doc-comment is
  honest that the **store scan and render/TUI delegates are still per-entity** (epic 21
  M9/M10). Scope it: `ResearchDir`, a `ListResearch` scan, `WatchPaths`, render delegates,
  the `research` cobra subtree, the `--json` envelope + `schema --json-schema`.
- **Templates** — one default scaffold, or named ones (`spike`, `comparison`,
  `design-doc`)? The existing corpus is roughly those three shapes.
## Open questions

- Does `research` overlap enough with the scaffolded `adr` entity group that they should
  land together as one "documents, not work-items" family (shared scan, shared render
  shape) rather than two independent tenants?
- Ordering against the other tenants — routines (tenant #1), projects, ADRs. Research has
  the largest existing corpus and the weakest contract, which argues for going early.

## Related

- Epic [28-first-class-entities-new-planning-nouns](../epics/28-first-class-entities-new-planning-nouns.md) (this epic's charter: the "which nouns exist" axis)
- Tenant #1: [spike-routines-as-a-first-class-entity-routine-audit-lineage](6fjw1d2jm6e9-spike-routines-as-a-first-class-entity-routine-audit-lineage.md)
- [ADR-0001](../adrs/0001-adopt-adrs.md) — the decision this must not reopen
- [ADR-0003](../adrs/0003-stable-key-id-addressed-storage.md) — flat id-led layout (§3 id-from-date policy, §6 cutover), which research now follows
- Epic [26-frontmatter-schema-declared-validation-contract](../epics/26-frontmatter-schema-declared-validation-contract.md) — a new noun's fields should ride the field registry once it lands
- Migration template: `internal/tools/flatmigrate/main.go`

## Progress — implementation landed (2026-08-14)

The design questions were settled in conversation (see Decisions above), so this went
straight to implementation rather than producing a research doc first. `go build`,
`go test ./...`, and `go vet ./...` are green; `golangci-lint` was NOT run (not
installed in this environment) — the one unverified gate.

### Landed

- **domain** — `Research` struct (`domain/research.go`), `ResearchDir`, and a registry
  `Descriptor` entry, so `schema research` / `template list` / the `kinds` contract all
  light up with no per-kind switch. New body template.
- **store** — `researchstore.go` (`ListResearch`/`GetResearch`/`parseResearch`/
  `resolveResearch`), `CreateResearch` + `researchFields`, `ResolveResearchPath`, and
  `researchDir` in `WatchPaths`. All the generic helpers (`scanDir`, `flatCandidates`,
  `resolveID`, `splitFlatName`) were reused as-is.
- **core** — `service_research.go` (`NewResearch`/`ListResearch`/`ShowResearch`/
  `ResearchPath`) + the `ResearchStore` port.
- **cli** — `research new|list|show|path`, tag filter, picker + completion.
- **wire** — `ResearchJSON`, `research_list`/`research_show` envelopes, `SchemaVersion`
  bumped to **1.30** with a changelog entry.
- **render** — `ResearchColumns`/`ResearchHuman`/`ResearchShowHuman` (+ JSON).
- **`init`** now scaffolds `research/` alongside the other entity dirs.
- **tests** — store, core, and cli-integration coverage, incl. the id-from-`created`
  property, the non-id-led carveout gate, and the "no status/bucket/epic key" contract.

### One design note worth keeping

`id.NewAt` already existed and its doc comment already named this exact use case, so
**no new id code was needed** — the `WithIDGen` test seam was widened to cover the
date-stamped mint so one injection still makes every create path deterministic.

A date-only string parses to UTC midnight, so a backdated id is UTC-anchored and
`id.Time` returns it a day earlier in a negative-offset zone. That is exactly what
`flatmigrate` did, so backdated research ids are consistent with the ids already in the
tree — deliberate, not a bug. Tests compare in UTC.

### Migration tool — written and verified, NOT yet applied

`internal/tools/researchmigrate/` (modelled on `flatmigrate`; dry-run default, refuses a
dirty git tree). Verified on a full copy of `planning/`:

- 28 docs → `<id>-<slug>.md`, ids backdated from each doc's own date
- 10 bare docs got a frontmatter block created, prose `**Status**:`/`**Created**:`
  headers left in the body (per decision 7)
- 18 existing blocks backfilled surgically — the vestigial `status: reference` survived
  as an unknown key, as intended
- 70 inbound links across 38 files repointed; `lint --links` clean afterward
- all 28 then parsed through `research list`

Deliberately not applied to the real tree: the working tree is dirty with these code
changes, so `-apply` correctly refuses, and git should be a clean undo point before a
28-file rename lands. Run it after committing the code:

    go run ./internal/tools/researchmigrate -root planning          # preview
    go run ./internal/tools/researchmigrate -root planning -apply

Until then `research list` reports all 28 as non-entity files (exit 11) — the expected
pre-migration state. Top-level `lint` is unaffected and stays clean.

### Deliberately not done

- **Research is not in the top-level `lint` roster.** `lint` covers tasks + epics only;
  audits live behind `audit lint`. Folding research in belongs with the existing
  [fold-audits-into-the-top-level-lint-command](6fm8p1cj11qf-fold-audits-into-the-top-level-lint-command.md) task, not ahead of it. Validation
  still happens — a malformed or non-id-led doc is a `FileProblem` on `research list`.
- **No TUI surface.** No research tab/detail view; the TUI is untouched.
- **No `research set`/`edit`/`append`.** `research path` + `$EDITOR` is the edit story;
  there are no cross-reference fields to maintain, which was the main reason `set` exists
  for tasks.
- **`supersedes:`/`superseded_by:`** left out of the first contract (still an open
  question above) — it is how research would obsolete itself without a lifecycle.
