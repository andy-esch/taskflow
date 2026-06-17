---
status: ready-to-start
epic: 17-pm-go-cli
description: 'Resurrect the deferred audit finding-level surface: per-finding parser plus findings query, finding-status write, audit lint, candidate-list sync'
effort: Unknown
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, core, audit]
created: "2026-06-17"
---

# Audit finding-level operations query write lint sync

> 📥 **Externally proposed — filed 2026-06-17** from feedback by an agent
> dogfooding `tskflwctl` to pick + work a tranche on `desirelines-planning`
> (audit `2026-06-14-simplify-apigateway`). The friction it hit is the
> mirror image of a known gap: the **audit finding-level surface was specced
> in [[2026-06-06-tskflwctl-command-spec]] and explicitly *deferred* during
> the Go port** ("Deferred (audit finding-level): `status`/`fixed`/`landed`/
> `followup`/`sync` … `findings`/`stats`") and never re-filed. This task is
> that deferred surface, scoped to the four items the feedback grounded in
> real use.

## Objective

The audit is a **second-class citizen** in the tool: readable in aggregate,
not queryable or writable per-finding. Today the store only *regex-counts*
findings (`auditstore.go` — `findingHeaderRe` / `openFindingRe` over
fence-stripped prose); there is **no structured per-finding parse**, so the
fields a tranche-picker needs (status / effort / urgency / component) are
locked in prose and the only way to flip a finding's status is to hand-edit
the `**Status:**` line to a format the cheat sheet dictates. Tasks never work
this way — you'd never hand-edit a task's `status:`; you run `task start`.
This closes the same gap for findings.

The grammar is fixed by `desirelines-planning/audits/HOWTO-execute.md` (the
per-finding header + cheat sheet) and the scaffold this tool already writes
in `core/service.go` (`auditBodyTemplate` — the `#### CODE.` header,
`**Status:** open`, `**File:**`/`**Component:**`/`**Effort:**`/`**Urgency:**`
lines, and the `✅⚠️⏳⛔` candidate-tasks list).

### 1. Per-finding parser (the dependency for everything below)

Extend the finding model from *count* to *parse*. A `domain.Finding` struct —
code (`H1`/`M2`/`S1`), title, status, effort, urgency, component, file:line,
and the body offsets needed for surgical status edits. This is the natural
extension of the `domain.CountFindings(body) (total, open int)` move already
planned in [[scaffold-schema-version-key-and-domain-level-audit-finding-counter]]
(its item #2) — fold the two: build `domain.ParseFindings(body) []Finding`
and derive the counts from it, so the counting invariant and the parse share
one fence-aware, table-tested code path.

### 2. `audit findings` — finding-level query (feedback #1, the headline)

```
tskflwctl audit findings --status open --effort XS,S --urgency soon --json
tskflwctl audit findings --component stravapipe --json
```

Cross-audit (and single-audit via `<slug>`) finding search with filters on
the parsed fields. Turns "read 17 files and build a mental map" into one
query — the audit-execution hot path. `--json` carries the full per-finding
record + its audit slug. Specced as `findings [--status --severity]`; the
feedback's effort/urgency/component filters are a superset — include them.

### 3. `audit finding <slug> <code> --status` — write (feedback #2)

```
tskflwctl audit finding 2026-06-14-simplify-apigateway S1 --status in-progress
tskflwctl audit finding 2026-06-14-simplify-apigateway S1 --status fixed --pr 724
```

Surgically rewrite one finding's `**Status:**` line, stamping the cheat-sheet
format the human currently has to remember: `in-progress (since YYYY-MM-DD)`,
`fixed YYYY-MM-DD (PR #N)` when `--pr` is given, etc. On `--status fixed`,
prompt for / append the 1–3 line resolution block HOWTO mandates but nothing
enforces (the interactive *face* of that prompt belongs to
[[interactive-prompt-layer-gh-style-pickers]]; the non-interactive append +
format stamping is core and lives here). Edits go through the atomic /
surgical-write discipline (`store/atomic.go`) — the rest of the audit body is
byte-preserved, like `task set` preserves frontmatter. Specced as
`status <slug> <code> <value> [--pr N] [--note]`.

### 4. `audit lint` — findings are linted (feedback #3)

`tskflwctl lint` validates tasks + epics, **not** audits (confirmed in
`CLAUDE.md` + HOWTO; the only post-edit check today is "does `audit show`
load"). Add audit linting (a new `audit lint [<slug>]`, or fold audits into
top-level `lint`) that checks:

- finding-header grammar (`#### CODE.` + the required `**Status:**`/field
  lines parse);
- the **status vocabulary** is legal (`open · in-progress · fixed · landed ·
  deferred · superseded · wontfix` — reject typos a free-text edit allows);
- the **bucket == finding-state invariant** is sane (e.g. flag a `closed`
  audit with `open` findings not named in a Closeout block — HOWTO's close
  rule).

### 5. Candidate-list ↔ finding-status sync (feedback #7)

HOWTO calls the bottom-of-audit `✅⚠️⏳⛔` list and the per-finding `Status`
lines "the same state from two angles," kept in sync by hand. Two parts,
both falling out of the parser:

- **Detect drift** as an `audit lint` check (item 4): a finding `fixed` whose
  candidate mark is still `⏳` is a lint warning.
- **Re-derive** via `audit sync <slug>` — rewrite the candidate-list symbols
  from the finding `Status` lines. Specced verbatim as `sync <slug>`.

## Acceptance criteria

- [x] `domain.ParseFindings(body) []Finding` exists with table tests (fenced-
      code exclusion, the `open-ish`/`openness` guards, missing optional
      fields); `auditstore.go` counts derive from it; the regex-only counter
      no longer lives in the store. (Subsumes the item-#2 half of
      [[scaffold-schema-version-key-and-domain-level-audit-finding-counter]].)
- [ ] `audit findings` filters on status/effort/urgency/component, single- and
      cross-audit, with a stable `--json` per-finding schema (`schema_version`).
- [ ] `audit finding <slug> <code> --status <v> [--pr N]` surgically rewrites
      one Status line in the cheat-sheet format; body otherwise byte-identical;
      `--status fixed` appends/prompts the resolution block; `--dry-run` works.
- [ ] Audits are linted (status vocabulary, header grammar, bucket==state) —
      `audit lint` or folded into `lint`; `--json` envelope like the others.
- [ ] `audit sync` re-derives candidate-list symbols from Status lines; lint
      flags drift between them.
- [ ] Errors wrap the domain sentinels (exit 10/11/13/14); suite + lint green;
      README "audit" / agent-use sections updated.

## Progress Log

- **2026-06-17**: Item 1 (the parser foundation) **done**. `domain/finding.go`:
  `Finding{Code,Title,Status,File,Component,Effort,Urgency}` + `ParseFindings` +
  `CountOpenFindings`, fence-aware and table-tested. `store/auditstore.go` now
  derives both counts from it and the finding regexes are gone from the store —
  one grammar in the domain. This subsumes the item-#2 half of
  [[scaffold-schema-version-key-and-domain-level-audit-finding-counter]] (close
  that half there). **Caveat:** the struct is *read*-shaped — it carries no body
  offsets, so the item-3 surgical `--status` write still needs a small extension
  (offset/section spans) before it can rewrite a line in place. Items 2–5 remain;
  item 2 (`audit findings` query) is being designed first — its value is the
  matching interface (filter dimensions + exact-vs-fuzzy per field + cross-audit
  aggregation), which is an explicit design pass, not a fill-in.

## Out of scope

- `audit followup <slug> <code>` (create+link a follow-up task) and
  `audit stats` — also specced/deferred, but they don't appear in this
  feedback; file separately if wanted. `findings`/`status`/`lint`/`sync` are
  the four the feedback grounded in real use.
- `audit new --routine`/`noop` (routine generation) — desirelines-side.
- Changing the finding grammar or the cheat-sheet vocabulary — this parses
  and validates the *existing* contract, it doesn't redesign it.
- The interactive picker/prompt UX for the resolution block — that's
  [[interactive-prompt-layer-gh-style-pickers]]; here it's a non-TTY append.

## Note — the audit `status:` frontmatter (feedback #4, NOT in scope here)

The feedback also flagged `status: completed` on audit frontmatter as
overloaded/misleading. **For this tool that's already resolved:** `audit new`
writes only `area` + `date` (no status field) and directory==bucket is the
audit-level state, exactly as [[2026-06-06-tskflwctl-command-spec]]
recommends. The `status: completed` the feedback hit is a **desirelines**
artifact (legacy / routine-generated audits) — the fix lives there (routine
audit-generator + back-fill + HOWTO), not in tskflwctl. The only tool-side
hook is optional: `audit lint` (item 4) *could* flag an unexpected `status:`
key on an audit as drift.

## Related

- Epic [[17-pm-go-cli]] — this is the port's deferred finding-level tail.
- Source spec: [[2026-06-06-tskflwctl-command-spec]] (`audit` table:
  `status` / `findings` / `lint` / `sync`; "Open gap — audit frontmatter").
- Parser dependency / fold:
  [[scaffold-schema-version-key-and-domain-level-audit-finding-counter]] (#2).
- Resolution-block prompt (interactive face):
  [[interactive-prompt-layer-gh-style-pickers]].
- Grammar contract: `desirelines-planning/audits/HOWTO-execute.md`.
- Touches `internal/domain/` (new `Finding` + `ParseFindings`),
  `internal/store/auditstore.go`, `internal/core/service.go`,
  `internal/cli/audit.go`, `internal/cli/lint.go`, `README.md`.
