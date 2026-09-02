---
schema: 1
id: 6g67v0c4n6sv
bucket: closed
area: tui-stable-entity-navigation-implementation-claude
date: "2026-09-02"
updated_at: "2026-09-02"
---
# Audit: TUI stable entity navigation implementation — Claude — 2026-09-02

> Reviewer assignment: Claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

## Review brief

Perform an independent adversarial implementation review of the uncommitted work for
[make TUI entity navigation use stable identities](../tasks/6g5rxq17px59-make-tui-entity-navigation-use-stable-identities.md)
on branch `feat/tui-stable-entity-identities`, based at commit `28ccf6b`. Review it against the
stable-identity contract in [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and
[docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md).

Assume the change can be subtly wrong despite green tests. Its 38-file diff crosses the
domain, core projections, nearly every TUI navigation surface, asynchronous reload state, cached
workspace sessions, lifecycle mutations, and display disambiguation. A missed slug-keyed path can
open or mutate the wrong duplicate while ordinary unique-slug fixtures keep passing. Re-derive the
identity flow; do not accept comments, types, or broad search-and-replace as proof. Equally, do not
manufacture findings: settle concerns when hostile evidence disproves them.

## Review target

Start with `git status --short`, `git diff --stat HEAD`, and the complete `git diff HEAD`. The
implementation is deliberately uncommitted in a dirty worktree. Primary review surfaces are:

- canonical identity methods and filesystem parse/population behavior for task, audit, research,
  and Thread records in `internal/domain` and the local stores;
- `entityRef`, `entityItem`, stable selection/restoration, duplicate-label hinting, list loading,
  filtering, and sorting in `internal/tui/entity.go`, `item.go`, and `commands.go`;
- detail requests/results/stale guards, dashboard and palette targets, follow/back navigation,
  atlas landings, watcher reloads, and workspace session save/restore throughout `internal/tui`;
- every lifecycle and inline-edit write path, including committed-with-warning task moves and the
  post-move `movedAwayKey` exception;
- `core.AuditFinding.AuditID` production in both query and summary paths and its dashboard consumer;
- `internal/domain/identity_test.go`, `internal/tui/identity_test.go`, and every modified pre-existing
  TUI test; and
- the architecture guide, ADR-0006 amendment, and implementation task.

The untracked `planning/tasks/6g63*.md` files are simultaneous unrelated planning work. Inspect
their names/status for scope classification, then leave them alone. Do not treat any other changed
file as unrelated without inspection.

## Intended contract to challenge

- TUI-internal identity is an entity's canonical store-resolution key, never its display slug.
  Task, audit, research, and future Thread rows prefer adapter-provided `FilenameID`; portable
  records without filename semantics fall back to semantic `ID`. Epics retain their canonical ID.
- A missing or drifting frontmatter `id:` remains visible to lint/repair, but cannot redirect a TUI
  read or write and cannot make two same-slug records collapse. An unusable empty canonical key must
  fail visibly rather than silently becoming a shared identity.
- Selection, cursor restoration, generation-stamped reloads/refilters, detail ownership and stale
  results, palette/dashboard/atlas jumps, follow history, the back stack, cached workspace sessions,
  lifecycle actions, and inline edits all preserve the canonical key end to end.
- Slugs remain human-facing labels and filter vocabulary. Unique labels retain existing rendering.
  Duplicate labels are deterministic and distinguishable through the shortest unique stable-ID
  prefix, expanded beyond the six-character minimum as needed; the canonical key itself is also
  searchable.
- Async completion order, filters, pagination, status views, external file changes, and workspace
  switches cannot apply a stale detail, restore the wrong duplicate, report a false not-found, or
  send a mutation to the visible label.
- Acute audit findings carry their parent audit's canonical key through core summary/query
  projections. The dashboard must not re-resolve a possibly duplicated audit slug.
- The contract belongs at domain/core/TUI boundaries without making the core depend on filesystem
  path rules or coupling future adapters to local filenames. It must be usable by the later Thread
  registry entry without pretending that a Thread screen exists in this task.
- No public JSON/CLI contract should change accidentally. Existing unique-slug navigation, output,
  lifecycle safety, and TUI rendering remain behaviorally compatible.

## Mandatory evidence floor

A `ready` verdict is not credible unless the report includes all of the following:

1. A repository-wide consumer inventory for `CanonicalID`, `FilenameID`, `entityRef`, `ref()`,
   `selectedRef`/`selectedKey`/`selectedLabel`, `selectByKey`, `restore`, `loadedKey`, `navLoc`,
   `pendingJump`, `movedAwayKey`, `AuditID`, every service read/write receiving a selected entity,
   and every message or session field carrying identity. Classify each use as canonical key, display
   label, or an actual defect; a grep count alone is not an inventory.
2. Hostile fixtures containing two records with the same slug and distinct stable IDs for tasks,
   audits, and research. At least one task and audit fixture must have filename/frontmatter ID drift,
   and one fixture must have a missing frontmatter ID. Show which exact record is selected and read,
   not merely that a row exists. Inspect the future Thread method separately rather than claiming
   unimplemented TUI behavior was exercised.
3. End-to-end task duplicate probes for selection, detail load, out-of-order detail completion,
   reload, async refilter, status-view widening, palette jump, follow/back, atlas cross-workspace
   landing, session switch/restore, inline edit, lifecycle move, committed-with-warning move, and
   post-move absence. Demonstrate the targeted file/key at each mutation boundary.
4. Equivalent risk-proportionate probes for audits and research: audit dashboard/findings navigation
   and bucket mutation must target the canonical audit; research navigation and restore must not
   collapse. Verify cross-kind equal labels in the global palette as well as within-kind duplicates.
5. Display probes for six-character collisions, prefix-of-another keys, very short keys, non-ASCII
   labels, narrow terminals, filter-hidden duplicates, pagination, and sort/view changes. Record
   whether the hint scope is the loaded set, filtered set, or rendered page and whether that matches
   the documented promise. Unique-label snapshots must remain unchanged.
6. Malformation probes for empty `FilenameID`, empty semantic `ID`, both empty, duplicate canonical
   keys, and multiple records whose keys cannot be distinguished by any prefix. Determine whether
   the TUI fails honestly, collapses rows, targets an arbitrary record, or merely exposes already
   invalid adapter output. Do not assume filesystem lint makes portable adapters safe.
7. At least these restored mutation probes, with the focused test that kills each mutation:
   - make semantic `ID` win over `FilenameID`;
   - change one selection/reload/detail comparison back to the label;
   - route one lifecycle or edit write through `ref.label`;
   - omit `AuditID` in one of the query/summary aggregation paths; and
   - reduce duplicate hints to a fixed six-character prefix or remove the cross-kind palette
     disambiguator.
   If a mutation survives, report a coverage finding even if the current production code is right.
8. Repeated focused tests under `-race`, the full `go test -race ./...`, static analysis/lint,
   docs drift checks, module tidiness, planning lint, and `git diff --check`, with exact commands and
   results. Distinguish cached from uncached runs and record runtime/Go version.

A report that restates the acceptance criteria, cites existing test names without hostile fixtures,
or says "all tests pass" without the consumer inventory and mutation ledger does not satisfy this
audit.

## Required hostile angles

1. **Canonical identity semantics.** Trace how local parsers derive `FilenameID`, how stores resolve
   ID versus slug, and how portable adapters populate records. Challenge whether a domain method
   defined as `FilenameID`-then-`ID` is truly adapter-neutral, whether it hides invalid data, and
   whether `CanonicalID` is the right name/owner. Require a concrete violated boundary before
   recommending a redesign.
2. **Malformed and adversarial rows.** Attack missing/drifting IDs, duplicate canonical IDs, empty
   keys, duplicate slugs across lifecycle buckets, a slug equal to another record's stable ID, case
   and Unicode lookalikes, and list reads that deliberately retain invalid records for repair. Check
   maps/equality operations for silent last-writer wins.
3. **Selection and reload ordering.** Interleave manual/watch reloads, filter application/removal,
   sort and status changes, stale generations, cursor movement, and detail results. Verify a restore
   intent belongs to the load generation that carries it and label drift cannot steal the cursor.
4. **Detail ownership.** Make duplicate A and duplicate B finish in both orders, fail refreshes, keep
   stale bodies, resize, toggle pretty/raw, and navigate away/back. Confirm titles stay friendly
   while every ownership/staleness test uses the canonical key.
5. **Mutation safety.** Open an action or edit menu for B, then reload, move the cursor, externally
   change the slug/status, or move B between lifecycle directories before the command completes.
   Check normal success, validation refusal, CAS conflict, committed-with-warning results, flashes,
   and post-move restoration. The write must never retarget A merely because the label is shared.
6. **Navigation graph.** Exercise dashboard findings, dashboard task groups, global palette,
   epic-to-task follow, task-to-epic follow, ctrl-o history, atlas pending jumps, inactive workspace
   sessions, missing targets, and view widening. Challenge both same-kind duplicates and cross-kind
   equal labels; inspect stale cached labels separately from stable keys.
7. **Duplicate presentation.** Prove the shortest-prefix algorithm terminates, is deterministic
   across input order, expands collisions correctly, and does not overwrite hints when keys repeat.
   Inspect truncation: a computed hint that is always clipped in a supported narrow view is not an
   effective disambiguator. Decide whether hints should react to filtering/pagination from the
   documented contract, not personal preference.
8. **Core projection completeness.** Inventory every `AuditFinding` constructor and consumer,
   including summary and direct query flows, empty/no-op test doubles, and any serialization. Verify
   adding `AuditID` neither leaks an accidental public wire field nor leaves dashboard targets empty.
9. **Portability and future Thread fit.** Use a minimal non-filesystem fake adapter to exercise the
   semantic-ID fallback. Inspect the future Thread record contract but do not demand a Thread TUI
   entry here. Flag filesystem knowledge leaking into core/TUI ports, or documentation that promises
   more adapter safety than the interfaces enforce.
10. **Regression and completeness search.** Search for remaining `.Slug`, `selectedID`, `id`, string
    message fields, map keys, and service calls in TUI code; classify them semantically. Pay special
    attention to old helper names/tests that may make a slug-keyed path look canonical. Verify epics
    were not broken by the generalized ref type.
11. **Performance and scale.** Measure or reason concretely about hint calculation, palette
    rebuilding, and list reloads for large duplicate groups and long common prefixes. File a finding
    only for a plausible planning-space size or a demonstrated pathological input, not theoretical
    big-O alone.
12. **Documentation and task truthfulness.** Compare source, tests, architecture text, ADR-0006,
    task checkboxes/progress, and actual Thread sequencing. Flag overclaims such as "every row",
    "never collapse", "visible duplicates", or adapter portability where runtime enforcement or
    evidence is weaker.

## Finding quality and proportionality

A finding must identify an observable wrong-entity read/write, lost or stale navigation state,
ambiguous rendering that defeats the accepted contract, a surviving mutation, an architectural
boundary violation, or a material documentation overclaim. Include stable severity code, exact
file/line evidence, reproduction or reasoning chain, impact, and minimum viable correction. Do not
report naming preferences, speculative future features, or broad rewrites as findings. If a concern
is real but belongs outside this task, leave it open; the implementation owner will decide whether
to fix it here or track it in a sequenced follow-up.

## Validation and restoration

Run proportionate validation, including hostile temporary-space or fixture probes and repeated race
tests. You may create scratch data outside the repository and temporary mutation probes inside it,
but restore every probe and generated artifact. Do not install dependencies, commit, push, edit the
implementation permanently, create follow-up tasks, change finding statuses, close this audit, or
edit the other reviewer's audit. At finish, the worktree must differ from its starting state only by
your edits to this assigned audit file. Verify that claim with `git status --short` and `git diff`.

## Deliverable

Preserve this brief and replace the reviewer-report placeholder below with:

- executive verdict: `ready`, `ready with tracked follow-ups`, or `not ready`;
- reviewed branch/base/worktree state, runtime, and exact validation results;
- a compact end-to-end identity-flow re-derivation rather than an acceptance-criteria paraphrase;
- findings grouped by severity, each with stable code and `**Status:** open` in the heading;
- an acceptance-criteria traceability table;
- a hostile-evidence ledger covering every required fixture, navigation/mutation path, and restored
  mutation probe, including observed result and the test that would catch regression;
- explicit separation of demonstrated defects, source-supported risks, and unverified concerns; and
- settled concerns with the evidence that settles them.

If there are no findings, say so plainly, but the evidence and mutation ledgers are still required.
Do not pre-resolve findings; the implementation owner will triage them with
`tskflwctl audit finding`.

## Reviewer report

### Verdict

**not ready** — one high finding and four medium ones, all localized; the identity core itself is
sound.

The hard part of this change is right. Canonical identity is genuinely separated from display text,
and I could not construct a wrong-entity read or write through the registry, the reload/restore
machinery, the detail stale guard, the back stack, the palette, the dashboard, or a cached workspace
session. The edit path is now demonstrably *more* correct than its predecessor: with the form open,
an external `mv` that changes only a record's slug no longer redirects the write to a same-named
sibling (probe P26 below).

What blocks a `ready` verdict is that the one user-visible half of the contract — telling two
duplicate rows apart — does not work at the terminal widths this tool is actually used at. The hint
is appended *after* the label and then truncated with it, so for this repo's median task slug the
two rows are byte-identical from 92 through 188 columns, the project's own `120×40` fixture
included (H1). Separately, the atlas work list kept its slug-keyed cursor three lines from a line
this diff edited (M1), and routing canonical keys through the store's fuzzy id-OR-slug resolver
introduces a narrow but real regression that can make a record unreachable (M2).

None of the four medium findings requires rethinking the design; H1 is a rendering-order fix, M1 is
a one-line key change, M2 is a resolver-selection choice, and M4 is test debt.

### Reviewed state

| | |
|---|---|
| Branch | `feat/tui-stable-entity-identities` |
| Base | `28ccf6b` (merge of #178) |
| Worktree | dirty, uncommitted, as the brief describes |
| Diff | 38 files, +521 −310 (`git diff --stat HEAD`) |
| Go | `go1.26.6 darwin/arm64` |
| golangci-lint | present; `0 issues` |

**Concurrency caveat.** Partway through this review an untracked
`internal/tui/audit_hostile_nav_test.go` (timestamped 17:05, the other reviewer's probe) appeared in
this shared worktree and did not compile, breaking `go build`/`go vet` on `internal/tui`. I did not
touch it. I moved all of my own probing to an out-of-repo copy of the tree
(`rsync -a --exclude .git` into the session scratchpad) and ran every fixture, probe, and mutation
there. The file was removed by its owner before I finished. Every result below is reproducible on a
clean checkout of this diff; the numbers were produced against a byte-identical copy.

**Restoration verified.** `git status --short` and `git diff HEAD` are byte-identical to the
snapshots I took before starting (`diff` reports no change). The only difference between the
worktree's starting state and its finishing state is this audit file. `bin/` is gitignored, so
`just build` left no trace; `go mod tidy` produced no change to `go.mod`/`go.sum`.

### Validation results

| Command | Result |
|---|---|
| `go test -race ./...` (uncached, `-count=1`) | all packages pass |
| `go test -race ./internal/tui/ -run 'TestDuplicate\|TestModel_Stale\|TestModel_ReloadDuringJump\|TestModel_Dashboard\|TestPalette\|TestModel_Follow'` ×5, `-count=1` | pass ×5, ~2.05 s each, no race reports |
| `gofmt -l internal/ cmd/` | empty |
| `go vet ./...` | clean |
| `golangci-lint run ./...` | `0 issues` |
| `git diff --check` | clean |
| `go mod tidy` then `git diff go.mod go.sum` | no change |
| `just docs-check` (regenerate `docs/cli` + `git diff --exit-code`) | clean, no drift |
| `just build` && `./bin/tskflwctl lint` | `✔ all planning entities and dependency links pass lint`, exit 0 |

First `go test ./...` of the session was partly cached (`internal/editor`, `id`, `listfilter`,
`tomledit`, `userconfig` reported `(cached)`); the `-race -count=1` run above is fully uncached and
is the one the verdict rests on.

### Identity flow, re-derived

Not a paraphrase of the criteria — this is the chain I walked and the boundary each step actually
crosses.

1. **Filesystem → domain.** `splitFlatName` (`internal/store/flatname.go:28`) accepts a stem only if
   it is exactly `id.Length` valid id characters, a `-`, and ≥1 slug character. So for any record the
   FS store returns, `FilenameID` is non-empty and exactly 12 chars. `parseTask`/`parseAudit`/
   `parseResearch` set `FilenameID = fnID` and `Slug = <rest>`; the frontmatter `id:` lands in `ID`
   independently and is never reconciled at parse time — which is what keeps drift lintable
   (`domain.IDDriftIssue`, `internal/domain/lint.go:142/203/233`).
2. **domain → canonical key.** `CanonicalID()` = `FilenameID` else `ID`. For local records this is
   always the filename id, so frontmatter drift and a missing `id:` are both inert for navigation.
   For an adapter with no filename semantics it degrades to the semantic `ID`. The four
   implementations are identical modulo receiver type.
3. **Canonical key → row.** `taskItem`/`auditItem`/`researchItem`'s `ref()` = `{key: CanonicalID(),
   label: Slug}`; `epicItem`'s = `{key: Epic.ID, label: Epic.ID}` (epics are directory-named, so id
   and label coincide and cannot duplicate).
4. **Row → everything stateful.** `selectByKey`, `markReload`/`reload`/`listLoadedMsg.restore`,
   `isCurrentSelection`, `detailPane.loadedKey`, `navLoc.ref`, `pendingJump.ref`, `dashTarget.ref`,
   `paletteItem.ref`, `spaceSession.movedAwayKey`, `actionMenu.ref`, `editMenu.ref` — all carry
   `ref.key`. I inventoried every one (table below); the split is clean.
5. **Canonical key → store.** Every write is `svc.X(ref.key, …)` and every detail read is
   `loadItem(svc, m.selectedKey())`. Inside the store these land on `FS.resolve` /
   `resolveAudit` / `resolveResearch`, i.e. **`resolveID`** — whose *exact* tier matches
   `c.id == query || c.slug == query`. This is the one place the chain is weaker than it looks; see
   M2.
6. **Canonical key → display.** `duplicateIdentityHints` groups the *loaded* refs by label and, for
   each group of ≥2, walks `n` from `min(6, len(key))` upward until no other key in the group shares
   the prefix. `labelWithIdentityHint` appends `" [" + hint + "]"`. Called at four sites: the three
   list loaders and the follow picker. The dashboard computes its own over the capped visible slice.
7. **Core projection.** `AuditFinding.AuditID` is populated in both producers — `QueryFindings`'
   `collect` closure and `summarize` — from `a.CanonicalID()`, and consumed by the dashboard's acute
   row. Verified live through both paths (P16).

The design is coherent. The defects below are all at the edges of this chain, not in it.

### Findings

#### High

#### H1. The duplicate-label identity hint is appended after the label and truncated away with it, so duplicate task rows render identically at every ordinary split-pane width · **Status:** fixed

**File:** `internal/tui/item.go:92-98` (also `internal/tui/entity.go:102-107`) | **Component:** TUI duplicate presentation
**Effort:** S · **Urgency:** acute

`taskDelegate.Render` computes one budget for the whole label and truncates the composed string:

```go
slugW := m.Width() - 2 - 2 - 3 - 10
if slugW < 8 { slugW = 8 }
slug := padRight(truncate(it.displayLabel(), slugW), slugW)
```

`displayLabel()` is `slug + " [" + hint + "]"`. Because the hint is the *tail* of that string and
`truncate` cuts from the right, the disambiguator is the first thing lost whenever the slug alone
approaches the column budget. The identical construction is used by `auditDelegate`
(`item.go:228-234`) and `researchDelegate` (`item.go:286-291`).

This is not a corner case for this repository. `planning/tasks/` holds 310 tasks with slug lengths
min 18, **median 53**, p75 62, max 80. The list pane is roughly `0.4 × terminal width` once the
detail pane splits.

**Measured** (probe P31: sweep terminal widths 80→320, render both duplicate rows through
`taskDelegate` at the same list index so cursor styling cannot mask the comparison; report the
widths at which the two rows are byte-identical):

| slug length | duplicate rows byte-identical at widths | first width that disambiguates |
|---|---|---|
| 18 | 92–100 | 104 |
| 30 | 92–132 | 136 |
| 40 | 92–156 | 160 |
| **53 (repo median)** | **92–188** | **192** |
| 62 (repo p75) | 80–84, 92–212 | 216 |
| 80 (repo max) | 80–256 | 260 |

Below ~88 columns the layout is single-pane and the list gets the full width, which is why the very
narrow end sometimes works; from 92 columns — where the detail pane appears — the hint is clipped
until the pane is wide enough to hold `slug + " [xxxxxx]"` outright.

The project's own standard test fixture is `120×40`. At 120 columns, with a median-length slug, the
two rows are indistinguishable:

```
term=120x40 listWidth=46 hints=("cjz0eb","zmp204")
   |│› ●   make-tui-entity-navigation-u…  yesterday│…
   |│  ●   make-tui-entity-navigation-u…  yesterday│…
   >>> duplicate rows visually identical: true
```

At 200 columns the same rows differ only in the sixth character of a clipped hint
(`[6g…` vs `[6g…`, cut mid-token).

**Impact.** Acceptance criterion 4 — "duplicate visible labels gain deterministic disambiguation" —
is not met for the primary entity at the widths the tool is used at, and the tests cannot see it
because they assert on `displayLabel()` (the pre-truncation string), never on rendered output.
`TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` and
`TestDuplicateStableIDEntitiesLoadByFilenameIdentity` both compare `displayLabel()` directly.

**Minimum viable correction.** Reserve the hint's width before truncating the label, i.e. truncate
the *slug* to `slugW - len(hint) - 3` and append the hint after — so the ellipsis eats slug
characters, which are redundant between duplicates, instead of the six characters that are the
entire point. Add a test that asserts on delegate output at a realistic width, not on
`displayLabel()`.

**Resolution:** Identity hints now lead shared labels, so supported row
truncation preserves the discriminator; delegate-level width regressions cover
task, audit, and research rows.

#### Medium

#### M1. The atlas work list still keys its cursor by `PlanningID + Task.Slug`, so duplicate in-progress slugs collapse and a re-sort can hand `openAtlasWork` the wrong row · **Status:** fixed

**File:** `internal/tui/atlas.go:499-501`, used at `atlas.go:481-497`, consumed at `atlas.go:366` | **Component:** Atlas cross-space navigation
**Effort:** XS · **Urgency:** acute

```go
func workKey(row core.SpaceInProgress) string {
	return row.PlanningID + "\x00" + row.Task.Slug
}
```

`applyWorkOrder` captures `workKey(selected)` before `sortWork()` and restores the cursor to the
first row whose key matches. Two in-progress tasks with the same slug in one planning space produce
one key, so the restore lands on whichever now sorts first. `openAtlasWork` — three lines above
`workKey`, and a line this diff *edited* to use `row.Task.CanonicalID()` — then builds the landing
`pendingJump` from that row. The canonical key is carried faithfully; it is simply the wrong row's.

**Reproduction** (probe P9, direct on the `atlas` value):

```
before re-sort: cursor=1 key="p1\x00same-slug" canonical="bbbbbbbbbbbb"
after  re-sort: cursor=0 key="p1\x00same-slug" canonical="aaaaaaaaaaaa"   <- wrong row
```

Pressing `o` (cycle work order) with the cursor on the second duplicate silently moves it to the
first; `⏎` then opens that workspace and lands on the wrong task. The atlas work rows also render
`row.Task.Slug` with no identity hint (`atlas.go:1057`), so nothing on screen reveals the swap.

**Impact.** A wrong-entity navigation landing — the exact defect class the task exists to remove —
surviving in a file the diff touched. `core.SpaceInProgress.Task` comes from an in-process store
scan (`core/space_overview.go:75` ← `Summary.InProgress`), so `FilenameID` *is* populated and the fix
is available.

**Minimum viable correction.** `return row.PlanningID + "\x00" + row.Task.CanonicalID()`. Optionally
extend `duplicateIdentityHints` to the work list's slug column for parity with the entity lists.

**Resolution:** Atlas work cursor identity now combines planning identity with
the task canonical ID, and duplicate work rows are visibly hinted; a re-sort
regression pins the selected duplicate.

#### M2. Canonical-key reads and writes route through the store's fuzzy id-OR-slug resolver, so a sibling whose slug equals this record's stable id makes the record unreadable and unwritable — a regression from the slug-keyed predecessor · **Status:** fixed

**File:** `internal/store/fsstore.go:249-259` (`FS.resolve` → `resolveID`, `internal/store/resolve.go:158-193`); TUI entry points `internal/tui/commands.go:78-86` and `internal/tui/edit.go:346-352` | **Component:** Store resolution / TUI read-write boundary
**Effort:** S · **Urgency:** soon

`resolveID`'s exact tier matches **either** key:

```go
if match(c.id) || (c.slug != "" && match(c.slug)) {
```

Before this change the TUI supplied slugs, and a slug is what that tier was designed around. The TUI
now always supplies an exact canonical id — but still through the same fuzzy entry point. If any
*other* record's slug happens to equal this record's stable id, the exact tier returns two hits and
the read fails as `ErrAmbiguous`.

The codebase already identified and solved this hazard for the CAS guard. `resolveExactID`'s own
doc comment (`resolve.go:203-207`) says it exists so "a same-named sibling — a task whose SLUG
happens to equal this file's id — can't turn the guard into a spurious ErrAmbiguous". The TUI's
read/write path did not get the same treatment.

**Reproduction** (probe P28): task A is `<idA>-alpha.md`; task B is `<idB>-<idA>.md`, i.e. B's human
slug is literally A's stable id. Both are well-formed and both lint clean.

```
row key=5bhrmwfbs1sf label=ch3r8r5xqwqp
row key=ch3r8r5xqwqp label=alpha
>>> A detail: loadedKey="" errMsg="\"ch3r8r5xqwqp\" matches 2 tasks:
      ch3r8r5xqwqp (5bhrmwfbs1sf), alpha (ch3r8r5xqwqp): ambiguous match"
>>> write to A's exact canonical key: tui.actionErrMsg
```

Task A becomes completely unreachable in the TUI — no detail, no edit, no lifecycle move — because
of a property of a *different* file. Under the previous slug-keyed code A resolved fine (its slug
`alpha` is unique), so this is a regression introduced by the change, not a pre-existing condition
merely re-exposed.

**Likelihood** is genuinely low: it needs a slug that is exactly 12 valid id characters. But it is
reachable with supported commands, it is silent until it bites, and the failure mode is total for
the affected record.

**Minimum viable correction.** Give the now-always-exact canonical read/write path an exact-id
resolve before falling back — `resolveExactID` is already written and already used by
`resolvePath`/`resolveAuditPath`/`resolveResearchPathExact`. Note the task lists "changing resolver
ambiguity policy" as out of scope, so reordering `resolveID`'s tiers globally is probably the wrong
fix here; adding an exact-first attempt on the canonical path is not a policy change. If the owner
would rather sequence this, it is a clean follow-up — the trigger is rare enough to defer
deliberately.

**Resolution:** Repository, task-graph, and Thread-snapshot resolvers now give
exact stable IDs their own precedence tier before slug and fuzzy matches, while
genuinely duplicated IDs remain ambiguous.

#### M3. A failed detail read titles the pane with the raw 12-character canonical key instead of the slug · **Status:** fixed

**File:** `internal/tui/model.go:361`, `internal/tui/detail.go:180-191` | **Component:** Detail pane presentation
**Effort:** XS · **Urgency:** soon

```go
m.detail.SetError(msg.id, msg.err.Error())      // model.go:361
func (d *detailPane) SetError(title, msg string) { … d.title = title … }
```

`msg.id` is now the canonical key, and `SetError`'s first parameter is the pane **title**. The
success path is unaffected (`SetContent` uses `c.Title()`, which is still `d.t.Slug`), so this only
shows on a failed or ambiguous read — precisely when the user most needs to know *which* record is
broken.

**Observed** (probe P7): with a task whose slug is
`make-tui-entity-navigation-use-stable-identities`, a failed read renders

```
>>> detail.title after failed read = "zmp204w8a52t"
>>> detailTitle()                  = "zmp204w8a52t"
```

The hostile brief asks that "titles stay friendly while every ownership/staleness test uses the
canonical key". Ownership is right; the title is not.

**Minimum viable correction.** Carry the label alongside the key into the error path (either widen
`detailErrMsg` to an `entityRef`, or title from `m.selectedLabel()` at the call site) and keep
`msg.id` for the `showing(id)` ownership check, which must stay canonical.

**Resolution:** Detail errors carry the selected human label beside the
canonical stale-guard key; a deleted duplicate fixture proves the friendly title
survives read failure.

#### M4. Ten of sixteen identity mutations survive the full suite, including every lifecycle write's ref plumbing and both remaining hint call sites · **Status:** fixed

**File:** `internal/tui/entity.go:295-347`, `internal/tui/edit.go:137-145,356-364`, `internal/tui/model.go:452`, `internal/tui/atlas.go:366`, `internal/tui/dashboard.go:95-96`, `internal/tui/nav.go:63`, `internal/tui/command_dispatch.go:150-160` | **Component:** Test coverage
**Effort:** M · **Urgency:** soon

I restored sixteen mutations one at a time against the full `go test ./internal/...`. Ten produced
no failure. The production code is correct in every case — this is coverage debt, which the brief
asks to be reported as a finding regardless.

The pattern is systematic: the *write functions* are covered, but the **ref construction and ref
plumbing around them** are not. `setFieldCmd(ref.key → ref.label)` is caught; `moveTask`,
`deferTaskCmd`, `moveAudit`, and `unsetFieldCmd` making the same substitution are not. Similarly
`dashTarget` for acute findings is covered but the in-progress task target is not; the follow-menu
jump is covered but the atlas landing is not.

Full ledger in the mutation section below. The four I would prioritise, in order:

1. `moveTask` / `deferTaskCmd` / `moveAudit` writing through `ref.label` — a status change landing on
   the wrong task is the worst outcome in this whole change, and nothing detects it.
2. `editMenu.open` keying the form by `t.Slug` — the write function is covered, its entry point is not.
3. `movedAwayKey = msg.ref.label` — silently reintroduces the false "not found" flash after a
   duplicate's lifecycle move.
4. `pendingJump` (atlas) and the dashboard in-progress `dashTarget` keyed by slug.

**Minimum viable correction.** One duplicate-slug lifecycle test that asserts *which file on disk*
changed (probe P10 below is a working template — it already distinguishes the two records), plus a
dashboard/atlas target-construction assertion.

**Resolution:** Hostile duplicate fixtures now pin edit capture, unset, task
move, defer, post-move suppression, audit move, dashboard targets, Atlas
landing, and both remaining hint surfaces to canonical keys.

#### Low

#### L1. The epic detail pane's task roster renders duplicate slugs unhinted while the follow picker over the same roster hints them · **Status:** fixed

**File:** `internal/tui/detail.go:521` vs `internal/tui/nav.go:57-64` | **Component:** TUI duplicate presentation
**Effort:** XS · **Urgency:** eventually

`followMenu.view` was updated to compute `duplicateIdentityHints` over `f.tasks`; `epicDetail`'s
body renders the same `ed.tasks` slice with a bare `t.Slug`. Side by side on the same fixture
(probe P8):

```
epic-detail roster |  ● make-tui-entity-navigation-use-stable-identities|
epic-detail roster |  ● make-tui-entity-navigation-use-stable-identities|
follow picker      |  › ● make-tui-entity-navigation-use-stable-identities [cjz0eb]|
follow picker      |    ● make-tui-entity-navigation-use-stable-identities [zmp204]|
```

The roster is read-only, so nothing mis-targets — but it is the surface a user reads *before*
pressing `f`, and it presents the epic as owning one task twice.

**Resolution:** Epic detail rosters now use the same loaded-set duplicate hints
as the follow picker, with focused rendering coverage.

#### L2. The command palette's match corpus silently widened from `id + entity` to the row's whole `FilterValue`, and concatenates the canonical key twice · **Status:** fixed

**File:** `internal/tui/command_dispatch.go:147` | **Component:** Command palette
**Effort:** XS · **Urgency:** eventually

```go
title: ei.displayLabel(), filter: ei.FilterValue() + " " + ref.key + " " + t.name,
```

Previously `filter: ei.id() + " " + t.name`. `FilterValue()` is slug + canonical id + description +
tags, so descriptions and tags are now palette-searchable, and `ref.key` is appended a second time
(it is already inside `FilterValue()` for task/audit/research, and inside it for epics too).
Observed (probe P21):

```
filter="alpha-task bqrw4h5k825h pineapple upside down cake zebra bqrw4h5k825h tasks"
  matches 'pineapple' (description): true
  matches 'zebra' (tag):             true
  canonical key appears 2 time(s)
```

The hunk's own comment says "Keep the ordinary one-word palette unchanged", which is true of titles
but not of matching. Widening the corpus may well be desirable — but it is an undocumented
behavioural change to a user-facing surface, made incidentally while adding the key. Either state it
in the task/ARCHITECTURE notes, or narrow it to `ei.ref().label + " " + ref.key + " " + t.name`.

**Resolution:** The duplicate canonical-key term was removed. The wider
FilterValue corpus is retained deliberately so palette search matches the source
tab vocabulary, and the contract is documented in code and task notes.

#### L3. `duplicateIdentityHints` degrades silently on three malformed-key shapes that only a portable adapter can produce — which is exactly the contract Threads are told to copy · **Status:** fixed

**File:** `internal/tui/entity.go:72-100` | **Component:** Duplicate presentation / adapter neutrality
**Effort:** S · **Urgency:** eventually

Filesystem keys are always exactly 12 valid id characters (`splitFlatName`), so none of this bites
today. It matters because ADR-0006 and ARCHITECTURE both present this as the adapter-neutral contract
a future Thread row and any non-filesystem adapter will implement. Measured (probes P2–P5):

| shape | result | note |
|---|---|---|
| key `"abc"`, key `"abcdef"`, same label | `hints = {abcdef: "abcdef"}` | the shorter key gets **no** hint; rows render `same` / `same [abcdef]` — still distinguishable, but by absence, and the hint is the full key rather than a prefix |
| two rows, same label, **same** key | both get the same hint | rows render identically; `selectByKey` silently takes the first |
| two rows, same label, both keys `""` | `hints = {"": ""}` | both render bare; `entityRef.empty()` is true, so selection/restore treat them as "nothing pending" rather than failing |

The inner loop's uniqueness test is `other != key && strings.HasPrefix(other, prefix)` — comparing
by *value*, so two rows carrying an identical key never see each other as a collision.

The brief asks whether the TUI "fails honestly, collapses rows, targets an arbitrary record, or
merely exposes already invalid adapter output". For the filesystem the answer is honest failure: I
built two files claiming one filename id (probe P14) and the store refused both the read and the
write with a precise message —

```
detail errMsg = "\"cs5bq1952dvp\" matches 2 tasks: alpha (cs5bq1952dvp), beta (cs5bq1952dvp): ambiguous match"
write result  = tui.actionErrMsg
```

— while only the row display and the cursor collapsed. For a portable adapter there is no such
backstop, and `CanonicalID()` returning `""` is currently indistinguishable from "no selection".

**Minimum viable correction.** Decide and document what an empty canonical key means at the registry
boundary (I would make it a loud row-level problem rather than a silent empty ref), and make the
prefix walk collision-safe for equal keys.

**Resolution:** Strict-prefix keys receive an explicit full-key fallback, and
the generic entity registry rejects empty or repeated canonical keys before
installing an ambiguous loaded result.

#### L4. Three documentation and comment claims are stronger than the implementation · **Status:** fixed

**File:** `internal/tui/entity.go:58-59`, `internal/tui/model.go:621-623`, `docs/ARCHITECTURE.md:509-522` | **Component:** Documentation truthfulness
**Effort:** XS · **Urgency:** eventually

1. `entity.go:58-59` — "displayLabel may append a short identity hint when duplicate labels are
   **simultaneously visible**." Hints are computed once per load over the whole *loaded* set and
   baked into the item. Measured (probe P13): with one duplicate `in-progress` and one `completed`,
   the default working view shows one row with **no** hint; `:all` shows both **with** hints; a
   filter that leaves one match keeps the hint on the survivor. Load-set scope is arguably the better
   behaviour (labels stay stable as you filter) — the comment is what is wrong.
2. `docs/ARCHITECTURE.md` repeats it: "only duplicate **visible** labels gain the shortest unique
   stable-ID prefix". Same correction.
3. `model.go:621-623` — `handleListLoaded`'s doc comment still says "Every tab restores its own cursor
   **by id**", stale vocabulary from before the rename. `entity.go:202`'s body comment and
   `messages.go:86-89` were updated; this one was missed.

Also worth a line in the task note: "Duplicate labels gain the shortest unique stable-ID prefix in
entity rows" is written as delivered, with no mention that H1 makes it invisible in practice.

**Resolution:** Architecture, ADR, and source comments now describe
loaded-result hint scope, leading truncation-safe hints, and canonical-key
restoration accurately.

#### L5. `y` yanks the bare ambiguous slug from a row the TUI itself renders as disambiguated · **Status:** fixed

**File:** `internal/tui/model.go:806` | **Component:** Clipboard yank
**Effort:** XS · **Urgency:** eventually

`m.yank(m.selectedLabel(), "slug")` is the right call in isolation — a slug is what you paste into a
CLI. But on a duplicate row the TUI now *shows* `slug [cjz0eb]` and yanks only `slug`, which no CLI
verb can resolve (probe P22):

```
row renders as "make-tui-entity-navigation-use-stable-identities [cjz0eb]"
>>> y yanks "make-tui-entity-navigation-use-stable-identities"
>>> resolving the yanked value: "…" matches 2 tasks: … : ambiguous match
```

Not a regression — the old code yanked the same string — but the change makes duplicates visible
without making the copied handle usable. Yanking `ref.key` when the row carries a hint would close
it.

**Resolution:** Yank now copies the canonical ID when a row carries a
duplicate-identity hint and retains the friendly slug for ordinary unique rows.

### Acceptance-criteria traceability

| # | Criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | Two same-slug tasks with distinct IDs stay independently selectable, detail-loadable, filterable, cursor-restorable across manual and watcher reloads | **met** | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState`; probes P6/P10/P25/P27. Reload via `markReload`→`reload` restores the second duplicate (P10, P30); filter by canonical key isolates one row |
| 2 | Equivalent duplicate coverage for every stable-ID entity; Thread contract has one documented canonical-key path | **met** | `TestDuplicateStableIDEntitiesLoadByFilenameIdentity` (tasks/audits/research); probes P29 (audit bucket move) and P30 (research restore + `ctrl+o`). `Thread.CanonicalID` exists, is tested in `identity_test.go`, and has no TUI entry point — `requestThreadDetail` has no external caller, so no Thread screen is implied |
| 3 | Nav history, palette jumps, stale guards, saved sessions carry canonical identity | **met** | Consumer inventory below; probes P17 (cross-kind palette), P20 (session switch/restore), P16 (dashboard acute finding). Out-of-order detail results rejected by key (`identity_test.go`) |
| 4 | Slug display/filtering stays readable; duplicate labels gain deterministic disambiguation without exposing IDs by default | **NOT met** | **H1** — deterministic and correct pre-truncation, invisible at 92–188 columns for a median slug. Unique labels are untouched (`labelWithIdentityHint(l, "") == l`), and hints are order-independent (P23) |
| 5 | Missing/drifting-ID fixtures fail visibly or use a documented canonical filename identity; never collapse or open an arbitrary duplicate | **met** for filesystem, **gap** for portable adapters | `duplicateEntityRepo` covers drift (record 0) and a missing `id:` (record 1) for all three kinds and reads the right body each time. Duplicate filename ids fail honestly (P14). **L3** covers the portable-adapter shapes |
| 6 | Existing unique-slug navigation and golden output unchanged except where disambiguation is intentionally exercised | **met** | Full `-race` suite green; `glamour_test`, `restore_test`, `sortcycle_test`, `yank_test`, `revisit_test` changes are mechanical `id()`→`ref()` renames. No `--json`/CLI surface touched — `AuditFinding` is a `core` type with no JSON tags reaching a public envelope, and `docs-check` reports no drift |

Every box in the task file is still `- [ ]`, correctly — the task is `in-progress`, and the progress
note does not claim otherwise. AC4 should not be ticked until H1 is resolved.

### Hostile-evidence ledger

Fixtures are as required: two records sharing a slug with distinct stable ids for **tasks, audits and
research**; record 0 of each kind carries a *drifting* frontmatter `id:`, record 1 *omits* it
entirely (`identity_test.go:duplicateEntityRepo`, which I re-used and extended).

| # | Probe | Observed | Regression guard |
|---|---|---|---|
| P1 | Hint truncation, synthetic widths | rows identical at slugW 23/33/43/53 | none — **H1** |
| P2 | Key is a strict prefix of another | shorter key gets no hint | none — L3 |
| P3 | Two rows, identical label **and** key | identical hint on both | none — L3 |
| P4 | Empty canonical keys | both bare; `empty()` true | none — L3 |
| P5 | 1-char keys, non-ASCII labels (`café`, `日本語`) | hints correct, widths sane | none |
| P6 | Rendered rows at 100/120/160/200 | identical at 100/120/160 | none — **H1** |
| P7 | Failed detail read | title = `zmp204w8a52t` | none — **M3** |
| P8 | Epic roster vs follow picker | roster unhinted, picker hinted | none — L1 |
| P9 | Atlas `applyWorkOrder` on duplicates | cursor jumps to other duplicate | none — **M1** |
| P10 | Lifecycle move on duplicate B | **correct file** moved; `movedAwayKey` = B's key; reload leaves 1 row, no false not-found | none — M4 |
| P11 | Rejected move | `actionErrMsg.ref` correct, flash uses label | none — M4 |
| P14 | Two files claiming one filename id | honest `ErrAmbiguous` on read **and** write; rows collapse | none — L3 |
| P16 | Acute findings, two audits sharing a slug | `AuditID` distinct via **both** `summarize` and `QueryFindings`; dash targets distinct | `TestModel_DashboardAcuteFindingJumpsToAudit` + `TestQueryFindings_CrossAudit_NoFilter` |
| P17 | Cross-kind equal labels, all tabs loaded | `shared-name · tasks` / `· audits` / `· research` | none — M4 |
| P18 | `jumpTo` a duplicate hidden by the status view | widens to `:all`, lands on the right one | `TestModel_JumpClearsAppliedFilter` (partial) |
| P19 | `jumpTo` a missing key | `"ghost-slug not found"`; bare key falls back to the key | — |
| P20 | Session save → `activateWorkspace` with `pendingJump` | restores the right duplicate | none — M4 |
| P21 | Palette match corpus | description + tags now match; key doubled | none — L2 |
| P22 | `y` on a duplicate row | yanks the ambiguous slug | none — L5 |
| P23 | Hint determinism across input order | identical maps | `TestDuplicateIdentityHintsExpandPastACollidingPrefix` (partial) |
| P24 | 50/200/1000-member duplicate group, long common prefix | ~0 ms, hint length stable | — no finding |
| P25 | Menu opened on B, cursor moved to A, then confirm | **B** moved, A untouched | none — M4 |
| P26 | Edit form open on B, external `mv` renames B's slug, then submit | write lands on **B**; A untouched; the old slug now resolves *uniquely to A* — the pre-change code would have written to the wrong task | none — M4 |
| P27 | External delete of the selected duplicate | survivor kept, label-shaped not-found | — |
| P28 | Sibling slug == this record's stable id | record unreachable for read **and** write | none — **M2** |
| P29 | Audit bucket move on duplicates | correct audit closed | none — M4 |
| P30 | Research reload / detail / `ctrl+o` on duplicates | right doc throughout | none — M4 |
| P31 | Width sweep 80→320 × 6 slug lengths | table in **H1** | none — **H1** |

P26 is worth singling out as evidence the change *works*: it is the wrong-entity write the old code
would have committed, and the new code does not.

**Consumer inventory** (every symbol the brief named, classified — counts are `internal/` non-test
occurrences):

| Symbol | Uses | Classification |
|---|---|---|
| `CanonicalID` | 31 | canonical key — 4 domain definitions, 2 core projections, 25 TUI ref/hint constructions. No display use |
| `FilenameID` | 45 | canonical key source — store parsers, lint drift checks, `CanonicalID`. No navigation use outside `CanonicalID` |
| `entityRef` / `ref()` | 72 | the contract itself |
| `selectedRef` | 5 | splitter only |
| `selectedKey` | 8 | **all canonical**: `handleTabMsg` prev, `updateList` prev, `afterSelectionChange`, `refreshDetail`, `isCurrentSelection`, `applySortToCurrent`, `toggleFilterMode` |
| `selectedLabel` | 3 | **all display**: window title (`view.go:100`), `y` yank (`model.go:806`) — see L5 |
| `selectByKey` | 6 | canonical — restore, jump, sort re-select |
| `loadedKey` | 7 | canonical — detail ownership, `showing()`, clears |
| `navLoc.ref` | 6 | key for `jumpTo`, label for the `ctrl+o` breadcrumb (`view.go:465`) — correct split |
| `pendingJump.ref` | 10 | canonical — atlas landing → `tabs[i].restore` |
| `movedAwayKey` | 8 | canonical, compared against `msg.restore.key` (like for like) |
| `AuditID` | 4 | canonical — 1 field, 2 producers, 1 dashboard consumer |
| `ref.key` consumers | 13 | every `svc.*` write, `selectByKey`, epic-detail identity check, `movedAwayKey`, palette filter |
| `ref.label` consumers | 10 | every flash, prompt heading, breadcrumb, hint grouping — **zero** reach a store call |

No defect found in the inventory itself: the label never reaches a service call and the key never
reaches a rendered title, with the single exception of M3's error path.

**Regression sweep.** All 34 remaining `.Slug` references in `internal/tui` are display, sort keys, or
ref *construction*; the only two that are identity are `atlas.go:500` (**M1**) and `detail.go:521`
(L1). No `selectedID`/`id()` remnants. Epics are unbroken by the generalised ref type — `epicItem`
returns `{ID, ID}`, epic follow/edit/move all key on `Epic.ID`, and epic tests pass unchanged.

### Mutation ledger

Sixteen mutations, each applied alone to a clean copy and run against the full
`go test ./internal/...`, then reverted. Baseline re-verified green afterwards.

| # | Mutation | Killed by |
|---|---|---|
| M-a | `Task.CanonicalID` prefers semantic `ID` over `FilenameID` | `TestCanonicalIDPrefersAdapterResolutionIdentity/task_filename`, `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` |
| M-b | `isCurrentSelection` compares `selectedLabel()` | 5 tests incl. `TestDuplicateStableIDEntitiesLoadByFilenameIdentity/tasks`, `TestModel_Glamour*` |
| M-b2 | `selectByKey` matches on the label | 9 tests incl. `TestModel_StaleReloadDoesNotStealRestore`, `TestModel_ReloadDuringJumpKeepsTarget`, `TestPalette_JumpSelectsEntity` |
| M-c | `setFieldCmd` writes through `ref.label` | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` |
| **M-c2** | **`moveTask` writes through `ref.label`** | **— SURVIVED** |
| M-d1 | `summarize()` omits `AuditID` | `TestModel_DashboardAcuteFindingJumpsToAudit` |
| M-d2 | `QueryFindings` omits `AuditID` | `TestQueryFindings_CrossAudit_NoFilter` |
| M-e1 | hint fixed at six chars, no collision expansion | `TestDuplicateIdentityHintsExpandPastACollidingPrefix` |
| **M-e2** | **cross-kind palette disambiguator removed** | **— SURVIVED** |
| **N1** | **`deferTaskCmd` writes through `ref.label`** | **— SURVIVED** |
| **N2** | **`moveAudit` writes through `ref.label`** | **— SURVIVED** |
| **N3** | **`unsetFieldCmd` writes through `ref.label`** | **— SURVIVED** |
| **N4** | **`editMenu.open` keys the form by `t.Slug`** | **— SURVIVED** |
| **N5** | **`movedAwayKey` records `ref.label`** | **— SURVIVED** |
| N6 | `markReload` returns a label-keyed ref | `TestEntityTab_MarkReloadCarriesPendingTarget`, `TestModel_RefreshFiresReloadMsg`, + 1 |
| **N7** | **atlas `pendingJump` keyed by the slug** | **— SURVIVED** |
| **N8** | **dashboard in-progress `dashTarget` keyed by the slug** | **— SURVIVED** |
| N9 | dashboard acute-finding target keyed by the audit slug | `TestModel_DashboardAcuteFindingJumpsToAudit` |
| N10 | follow-menu jump keyed by the slug | `TestModel_FollowEpicToTaskViaMenuMultiHop` |
| N11 | `taskItem.ref()` keyed by the slug | 9 tests |
| N12 | `researchItem.ref()` keyed by the slug | `TestDuplicateStableIDEntitiesLoadByFilenameIdentity/research` |
| N13 | `taskItem.displayLabel()` drops the hint | 3 tests |
| N14 | `taskItem.FilterValue()` drops the canonical key | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` |
| **N15** | **follow picker drops the identity hint** | **— SURVIVED** |
| **N16** | **dashboard rows drop the identity hint** | **— SURVIVED** |

All four mutations the brief named specifically (M-a, M-b, M-c, M-d, M-e1) are killed. The survivors
are M4.

### Settled concerns

Raised as hypotheses, then disproved by hostile evidence:

- **"A lifecycle move could target the wrong duplicate."** It does not (P10): B's file changes status,
  A's does not, `movedAwayKey` holds B's canonical key, and the post-move reload neither flashes a
  false not-found nor mistakes A for B. Untested (M4), but correct.
- **"An open action/edit menu could retarget when the cursor moves."** It does not (P25) — the ref is
  captured at open and is not re-read from the cursor.
- **"External renames could redirect a write."** The opposite: the canonical key makes the write
  *more* robust than before (P26).
- **"`AuditID` might be populated in only one of the two producer paths."** Both are populated, and
  the dashboard consumes distinct keys for two audits sharing a slug (P16).
- **"Cross-kind equal labels could collapse in the palette."** They do not (P17) — but only once every
  tab is loaded, which `openPalette` handles and `reindex` refreshes.
- **"Hints might be order-dependent."** They are not (P23) — grouping is by label and the prefix walk
  is a pure function of the group's key set.
- **"Hint computation could be pathological for large duplicate groups or long common prefixes."** It
  is not (P24): 1000 same-label rows sharing a 20-character prefix resolve in ~0 ms. Quadratic per
  group, but planning-space sizes make that irrelevant. **No finding.**
- **"The workspace-session cache could restore a stale label-keyed cursor."** It does not (P20) —
  `spaceSession.navStack`/`movedAwayKey` are canonical and `pendingJump` goes through the ordinary
  reload rather than selecting against cached rows.
- **"Threads might be half-wired."** They are not. `Thread.CanonicalID` exists and is tested;
  `requestThreadDetail` has no external caller and `threadProjectionState.detailRef` is a separate
  thread-only slot. No Thread screen is implied, matching the task's scope.
- **"The core might have gained a filesystem dependency."** It has not. `CanonicalID` is a domain
  method over two already-present fields; `core` never sees a path rule. `AuditFinding` is an internal
  `core` type with no JSON tags on the new field and no `--json` envelope exposure — `docs-check`
  confirms no CLI drift.
- **"Epics might be broken by the generalised ref."** They are not — `{ID, ID}`, all epic tests pass.

### Demonstrated defects vs risks vs unverified

**Demonstrated** (reproduced against real fixtures, evidence above): H1, M1, M2, M3, M4, L1, L2, L5.

**Source-supported risks** (correct reasoning from the code, not reachable through the filesystem
adapter today): L3 — every shape needs a portable adapter that does not yet exist, but the ADR and
ARCHITECTURE both promise this contract to one.

**Unverified / limits of this review:**

- I did **not** build a fake non-filesystem `core` store to exercise the semantic-`ID` fallback
  end-to-end through the TUI; the store interface is wide and a partial fake would have proved little.
  The fallback is covered at the domain level (`identity_test.go`) and by inspection. This is why L3
  is a risk rather than a demonstrated defect.
- Watcher-driven reloads were exercised through `reloadAll`/`markReload` (the code path a watcher
  event lands on), not through a live `fsnotify` event.
- H1's width table assumes the default split layout; I did not sweep zoom (`z`) or the
  single-pane fallback beyond noting that widths below ~88 columns behave differently.
- The concurrency caveat above: all probing ran against an out-of-repo copy of this exact diff.

### Commands run

```
git status --short ; git diff --stat HEAD ; git diff HEAD          # + snapshots for restoration
go version                                                          # go1.26.6 darwin/arm64
go test -race -count=1 ./...                                        # all pass, uncached
go test -race -count=1 ./internal/tui/ -run '<focused set>'  (×5)   # pass, ~2.05s each
gofmt -l internal/ cmd/ ; go vet ./... ; golangci-lint run ./...    # clean / clean / 0 issues
git diff --check ; go mod tidy && git diff go.mod go.sum            # clean / no change
just docs-check ; just build && ./bin/tskflwctl lint                # no drift / lint clean, exit 0
# out-of-repo copy (session scratchpad):
go test ./internal/tui/ -run 'TestProbe_' -v                        # probes P1..P31
zsh mutate.sh ; zsh mutate2.sh                                      # 16-mutation ledger, each reverted
diff baseline.status now.status ; diff baseline.diff now.diff       # worktree identical to start
```

All probes, fixtures, and mutations lived outside the repository or were reverted; the worktree
differs from its starting state only by this file.
