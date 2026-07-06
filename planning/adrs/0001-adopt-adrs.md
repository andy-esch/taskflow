---
status: proposed
date: "2026-06-20"
deciders: [andy-esch]
tags: [adr, planning-model]
supersedes: []
superseded_by: null
---

# ADR-0001: Adopt Architecture Decision Records (ADRs)

> **Bootstrapping note.** This is the first ADR, and it proposes the very concept of
> ADRs — so it is written by hand (there is no `tskflwctl adr` command yet) and it
> deliberately *follows the format it proposes*. Same move Nygard's ADR-0001 ("record
> architecture decisions") makes. This file is the format's first instance; the
> eventual `adr new` template must match it, pinned by a golden-file test so the two
> can't drift. Design rationale and the best-practice survey behind it live in
> [2026-06-20-adrs-and-projects-format-design](../research/2026-06-20-adrs-and-projects-format-design.md). **Projects** — the cross-cutting
> grouping concept — are a separate decision in [0002-adopt-projects](0002-adopt-projects.md), built on the
> format this ADR establishes.

## Context and Problem Statement

tskflwctl is a generic, local-first work-management tool whose planning types are
knowledge-organization primitives it scaffolds for **any** repo. Today it has three:
**tasks** (the unit of work; `status == directory`), **epics** (a durable thematic
*home*; flat, frontmatter status), and **audits** (point-in-time reviews;
`status == directory`).

There is no durable home for **why**. Significant decisions — "split list output into
`-o`/`-c`", "pickers use bubbles/list not huh.Select", "prompts are TTY-gated" —
live scattered across task bodies, "Decided (date)" notes, and session history.
Nothing records *a decision and its rationale* in a discoverable, citable form, and
there is no first-class way to mark one decision as superseding another. For an
agent-driven workflow this is a real gap: an agent can read *what* to do but not
*why* the ground rules are what they are.

**A negative precedent worth naming.** ADRs were tried once in the sibling
`desirelines` repo and the convention was **declined**
(`desirelines-planning/tasks/deprecated/decide-on-adr-adoption-as-a-documentation-convention.md`,
2026-05-26: *"no precedent in the repo, and the versioning policy lives directly in
openapi.yaml"*). The lesson is not "ADRs are bad" — it's that **a repo should opt
into the convention deliberately, and a decision that lives better inlined should
stay inlined.** This ADR therefore makes tool support for ADRs *opt-in*, never
reflexive (see Decision → *Opt-in*). Supporting the type ≠ requiring its use.

## Considered Options

- **A — Add ADR as a first-class, opt-in type** (this decision). The tool can
  scaffold/lint/cross-link ADRs for repos that want them; no repo is forced to.
- **B — Extend `research/` (already branded "TDRs") with a `supersedes:` field.**
  Cheaper, reuses a working home for "why." Rejected as the *primary* model because
  research docs are open-ended spikes with no lifecycle, numbering, or status
  discipline — but its spirit survives in *opt-in* (a repo can decline ADRs and keep
  using `research/`).
- **C — Keep inlining decisions where they're enforced** (code comments, the
  relevant config). This is the right call for *narrow, locally-enforced* decisions
  (the desirelines lesson) and ADRs do not replace it — ADRs are for *cross-cutting*
  decisions with no single enforcement site. Not mutually exclusive with A.

## Decision

Adopt **ADR** as an optional planning type with the format, lifecycle, and layout
below. (Projects are decided separately in [0002-adopt-projects](0002-adopt-projects.md).)

### Format

A durable record of one significant decision: its **context**, the **decision**, and
its **consequences** — the *why* layer. Lineage: Michael Nygard's ADRs + MADR. The
default body, smallest that no source would call wrong, is **Nygard's core plus
MADR's Considered Options**:

- **Required:** Title · Context and Problem Statement · Decision · Consequences.
- **Scaffolded by default:** Considered Options · Amendments (empty) · Related.
- **Optional, behind flags** (`adr new --with-drivers`, `--with-confirmation`):
  Decision Drivers, Confirmation/Validation, per-option Pros & Cons (the MADR-full /
  Tyree heavyweight extras). Kept out of the default so the common case stays ~1–2 pages.

**Status lives in frontmatter only** — canonical and lint-validated — *not* also in a
`## Status` section. (Nygard/AWS/Microsoft keep a prose Status section; we follow
MADR's frontmatter placement to avoid two-places drift, consistent with how every
other type here carries status.)

### Append-only after acceptance (amendments, not edits)

- While **`proposed`**, an ADR is a draft — freely editable.
- On **`accepted`**, the decision content (Context, Considered Options, Decision,
  Consequences) **freezes**. New information is added **append-only** as dated entries
  under **`## Amendments`** (a clarification, a consequence discovered later, a link
  to a related decision). Status stays `accepted`.
- **`adr accept` stamps a visible "finalized" banner** at the top of the file (a
  blockquote directly under the H1), e.g.:
  `> ✅ **Accepted 2026-07-01 — finalized.** The decision sections below are frozen;
  add new information only under \`## Amendments\`. Reverse via a superseding ADR.`
  The banner makes the convention **in-band and unmissable** — a reader (human or
  agent) opening the file is told, where they're reading, that the frozen sections are
  off-limits. `lint` enforces the banner's *presence* on every `accepted` ADR (and its
  absence on `proposed` ones); it can't police the content-freeze itself (see
  Consequences).
- A genuine **reversal** is not an amendment — you write a *new* ADR and mark this one
  `superseded` (see Lifecycle). The only frontmatter edits permitted post-acceptance
  are the `status` line and `superseded_by`.
- This makes "append-only history" literally true while still letting a solo/agent
  workflow extend a record it has learned more about.

### Lifecycle

```
proposed ──accept──▶ accepted ──new ADR replaces it──▶ superseded  (terminal; links to replacement)
   │                    │
 reject               retire, no replacement
   ▼                    ▼
 rejected            deprecated  (terminal)
 (terminal)
```

`rejected` is reachable only from `proposed` (you don't reject an already-accepted
decision — you `deprecate` or `supersede` it). `superseded` = replaced by a specific
new ADR (bidirectional link); `deprecated` = retired with no replacement.

### Opt-in (never reflexive)

The tool does **not** create `planning/adrs/` unless asked. `tskflwctl init`
*offers* ADRs (and other optional types) as an interactive opt-in; otherwise the
directory is created lazily by the first `adr new`. A repo that wants its decisions
inlined or in `research/` simply never opts in. This is the direct answer to the
desirelines precedent.

### Spawns work, isn't work

An accepted ADR can spawn an epic, a project, or tasks; those work items cite it back
via a task `adrs:` field (see Cross-linking). ADRs are never on the task board.

## ADR frontmatter spec (decided)

```yaml
---
status: proposed          # proposed | accepted | superseded | deprecated | rejected
date: "2026-06-20"         # ISO, quoted (house style); the proposed/created date
deciders: [andy-esch]     # list; who owns the decision
tags: [adr]               # inline list; `adr` plus any topical tags
supersedes: []            # list of ADR ids this replaces, e.g. [ADR-0003]
superseded_by: null       # an ADR id when superseded, else null
---
```

- **Identity is the number**, derived from the `NNNN-` filename prefix (not stored as
  a field — same as epics derive their `NN` from the filename). Quoted ISO dates,
  bare-enum `status`, inline tag lists — all house style.
- **One canonical id form everywhere: `ADR-NNNN`.** Used in prose, in `[[ADR-NNNN]]`
  wikilinks, and in `supersedes:`/`superseded_by:`. The resolver strips `ADR-` and
  matches `NNNN-*.md`. No bare-number / slug / `ADR-NNNN` spelling drift.

## Data model & layout

| | ADR |
| :-- | :-- |
| Location | `planning/adrs/NNNN-slug.md` (**flat**) — created only on opt-in |
| Identity | the number (`ADR-NNNN`), monotonic, **never reused** |
| Status | frontmatter field (durable type, like epics — not `status == directory`) |
| Mutability | frozen on accept; append-only `## Amendments`; reversal via supersede |

- **Home:** `planning/adrs/`, alongside the other planning types. tskflwctl manages
  one local planning tree as a group — it does **not** track files across repos — so
  ADRs live *with* the planning docs, full stop (no configurable dir, no impl-repo
  split). This supersedes the earlier command-spec idea of writing ADRs to "the impl
  repo's ADR dir" ([2026-06-06-tskflwctl-command-spec](../research/2026-06-06-tskflwctl-command-spec.md)).
- **Numbering allocation** (race without git or locks): scan `planning/adrs/NNNN-*.md`
  for the max, attempt an **exclusive create** of `max+1` (the existing
  `createFileAtomic` / `O_EXCL`); on collision, re-scan and retry. No `flock` needed.

## Cross-linking

ADR-related links only (Projects' links are in [0002-adopt-projects](0002-adopt-projects.md)):

| Link | Field / form | On | Validated |
| :-- | :-- | :-- | :-- |
| ADR ↔ ADR supersession | `supersedes:` / `superseded_by:` (`ADR-NNNN`) | ADR | yes (against `adrs/`) |
| work → decision | `adrs: [ADR-NNNN]` | task | yes (against `adrs/`) |
| prose reference | `[[ADR-NNNN]]` | any body | resolver (future) |

- **`adrs:` is a new task field** — it does not exist in the registry today
  (`internal/domain/fields.go` has `projects` but not `adrs`), so adding it is a real
  schema change: `domain.Task` struct + field registry + `schema task` doc + sync
  test. Enumerated, not hand-waved.
- **`adr supersede <old> --by <new>` writes both sides** and flips the old status —
  but because the tool is non-transactional and doesn't touch git, a crash between the
  two file writes can leave a one-sided link. **`lint` is the integrity backstop**: it
  detects and `--fix`es a dangling `supersedes`/`superseded_by` pair. The *write*
  attempts both; *lint* guarantees consistency. (This is the only genuinely
  bidirectional ADR link.)
- **Out of scope here:** a general `[[wikilink]]` resolver across all planning docs,
  the TUI `f`-nav extension, and migrating existing inconsistent links. Those are a
  separate concern from adopting ADRs and get their own ADR if ever justified.

## CLI surface (sketch, for the implementation epic)

`adr new "Title"` (auto-numbered, `--with-drivers`/`--with-confirmation`) ·
`adr list|show` · `adr accept` (sets `status`, stamps the **finalized banner**) ·
`adr amend <n>` (append a dated `## Amendments` entry) · `adr supersede <old> --by
<new>` (both links + status flip) · `adr deprecate|reject`. `schema` learns `adr`
(positional, like `schema task`); `lint` validates ADR frontmatter, numbering, the
supersession pair, the accepted-banner presence, and `task.adrs:`.

## Consequences

**Positive.** Decisions get a durable, discoverable, citable home with real
supersession. The append-only-amendments model keeps history honest without rigid
immutability. Opt-in respects the desirelines lesson and the "support, not require"
principle. The format is standard (Nygard+MADR), short by default, and lints cleanly.

**Negative / cost.** A new type: command group, schema entries, lint rules, a new
`task.adrs:` field, and the `Amendments`/supersede machinery. Monotonic numbering
introduces the tool's only identity-allocation race (mitigated above). Append-only
discipline is a convention: the stamped banner makes it visible and `lint` enforces
the banner's presence, but neither can tell an "amendment" from a silent edit to a
frozen section — git review remains the human backstop.

## Amendments

<!-- Append-only, dated entries added AFTER this ADR is accepted. Format:
     ### 2026-07-01 — <what changed and why> -->

_None yet (still `proposed`)._

## Related

- Design rationale & best-practice survey: [2026-06-20-adrs-and-projects-format-design](../research/2026-06-20-adrs-and-projects-format-design.md).
- Follow-up — Projects: [0002-adopt-projects](0002-adopt-projects.md).
- Command spec (the `adr` group; its impl-repo location line is superseded here):
  [2026-06-06-tskflwctl-command-spec](../research/2026-06-06-tskflwctl-command-spec.md).
- Format lineage: Michael Nygard's ADRs + MADR (Markdown Any Decision Records).
