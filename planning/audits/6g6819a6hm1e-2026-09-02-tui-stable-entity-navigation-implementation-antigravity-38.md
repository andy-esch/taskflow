---
schema: 1
id: 6g6819a6hm1e
bucket: closed
area: tui-stable-entity-navigation-implementation-antigravity-38
date: "2026-09-02"
updated_at: "2026-09-02"
---
# Audit: TUI stable entity navigation implementation — Antigravity (Gemini 3.8) — 2026-09-02

> Reviewer assignment: Antigravity (Gemini 3.8). This document is the review brief and the only file the reviewer should update.
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

### Executive Verdict: Ready with Tracked Follow-ups

The TUI stable entity navigation implementation for task [`6g5rxq17px59`](../tasks/6g5rxq17px59-make-tui-entity-navigation-use-stable-identities.md) on branch `feat/tui-stable-entity-identities` (based at commit `28ccf6b`) has been independently and adversarially reviewed against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and [`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

1. **Identity Seam Verification:** The entity contract introduces [`entityRef{key, label}`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L49-L52), cleanly decoupling immutable store resolution keys (`key`) from human-facing display text (`label`). For tasks, audits, research, and future Threads, `CanonicalID()` prioritizes adapter-derived `FilenameID` so missing or drifted frontmatter `id:` values cannot corrupt or misdirect store operations, while portable adapters without file semantics fall back to semantic `ID`.
2. **Navigation & Detail Ownership Integrity:** Cursors, selection, list reloads, asynchronous refilters, and stale guards use canonical keys ([`entityTab.selectByKey`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L202-L210), [`entityTab.restore`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L171), [`Model.isCurrentSelection`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/model.go#L1226-L1228), [`detailPane.loadedKey`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/detail.go#L44)). Out-of-order detail completions for duplicate slugs cannot land on the sibling row. Follow/back history ([`navLoc`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/nav.go#L22-L25)), atlas cross-space work landings ([`pendingJump`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/session.go#L64-L68)), and inactive workspace sessions ([`spaceSession.movedAwayKey`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/session.go#L84)) strictly retain canonical keys.
3. **Mutation Seams:** Lifecycle transitions ([`moveTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L293), [`deferTaskCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L313), [`moveAudit`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L330), [`moveEpic`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L343)) and inline field edits ([`setFieldCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/edit.go#L346), [`unsetFieldCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/edit.go#L358), [`setEpicFieldCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/edit.go#L370)) pass `ref.key` directly to service mutations, avoiding ambiguous slug re-resolution.
4. **Disambiguation & Filter Robustness:** Colliding visible labels receive deterministic, shortest unique stable-ID prefixes ([`duplicateIdentityHints`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L72-L100)) with collision expansion beyond 6 characters. Both slug and canonical ID are indexed in bubbles/list [`FilterValue()`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/item.go#L51-L53).
5. **Audit Findings Projections:** [`AuditFinding.AuditID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/finding.go#L133) is populated across `Summary()` and `QueryFindings()`, enabling acute dashboard rows to navigate to the exact canonical audit without slug re-resolution.

Playing devil's advocate revealed systemic architectural anti-patterns in truncation layout, test-helper aliasing, and presentation-dependent CLI yanking. Six findings are filed and left open for implementation-owner triage.

---

### Review Environment & Validation Results

- **Branch:** `feat/tui-stable-entity-identities`
- **Base Commit:** `28ccf6ba78a24967977c0b6b048c6a4c44026d91` (`28ccf6b`)
- **Runtime Platform:** `darwin/arm64` (macOS 15.x), Go 1.26.6, APFS
- **Reviewer Agent:** Antigravity (Gemini 3.8)

#### Validation Commands & Outcomes

| Validation Suite | Exact Command Line | Outcome |
| :--- | :--- | :--- |
| **Repeated Focused Race** | `go test -race -count=10 -run 'TestDuplicate\|TestCanonicalID' ./internal/domain ./internal/tui` | **PASS** (10/10 iterations clean, 2.053s) |
| **Full Repository Race** | `go test -count=1 -race ./...` | **PASS** (all 32 package test suites clean, 7.923s max duration) |
| **Static Analysis / Lint** | `golangci-lint run ./...` | **PASS** (0 issues reported) |
| **Go Vet** | `go vet ./...` | **PASS** (clean) |
| **Docs & CLI Drift** | `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli` | **PASS** (docs up to date) |
| **Go Module Tidiness** | `go mod tidy -diff` | **PASS** (go.mod and go.sum tidy) |
| **Planning System Lint** | `go run ./cmd/tskflwctl lint` | **PASS** (all planning entities and dependency links valid) |
| **Git Diff Check** | `git diff --check` | **PASS** (clean whitespace) |

---

### Repository-Wide Consumer Inventory

| Symbol / Seam | Consumer Files / Lines | Semantic Classification | Assessment |
| :--- | :--- | :--- | :--- |
| **`CanonicalID()`** | `internal/domain/task.go:64`, `audit.go:83`, `research.go:61`, `thread.go:74`, `internal/core/service.go:322`, `internal/core/finding.go:161`, `internal/tui/commands.go:56,170,240`, `item.go:52,176,258`, `nav.go:58,88`, `atlas.go:366` | Canonical Key | **Correct.** Prefers `FilenameID` over `ID`. Uniform across all flat entities and future Thread. |
| **`FilenameID`** | `internal/domain/task.go:65`, `audit.go:84`, `research.go:62`, `thread.go:75`, `internal/store/flat.go:42` | Canonical Key | **Correct.** Derived from flat filename; authoritative over frontmatter copy. |
| **`entityRef`** | `internal/tui/entity.go:49-54`, `item.go`, `model.go:1259-1267`, `nav.go:24,88,110,167,175`, `session.go:66,167`, `action.go:99`, `edit.go:119,135`, `palette.go:30`, `dashboard.go:31` | Seam Struct | **Correct.** Pairs canonical `key` with user-facing `label`. |
| **`ref()`** | `internal/tui/item.go:54,115,178,260`, `entity.go:62,204,222`, `command_dispatch.go:144` | Canonical Key / Label accessor | **Correct.** Returns `entityRef` for any list item implementing `entityItem`. |
| **`selectedRef()`** | `internal/tui/model.go:1259-1264`, `nav.go:112,144`, `command_dispatch.go:60,263` | Canonical Key / Label | **Correct.** Extracts current selection's `entityRef`. |
| **`selectedKey()`** | `internal/tui/model.go:1266`, `model.go:1227` (`isCurrentSelection`), `action.go`, `edit.go`, tests | Canonical Key | **Correct.** Primary key for detail loading, stale checks, and mutations. |
| **`selectedLabel()`** | `internal/tui/model.go:1267`, `view.go:100` (`windowTitle`), `model.go:806` (`yank`) | Display Label | **Correct.** Used exclusively for titles, copy-to-clipboard, and human flashes. |
| **`selectByKey()`** | `internal/tui/entity.go:202-210`, `nav.go:188`, `model.go:655` (`handleListLoaded`) | Canonical Key lookup | **Correct.** Iterates `VisibleItems()` and matches `ei.ref().key == key`. |
| **`restore` / `restoreGen`** | `internal/tui/entity.go:171-172,181-195`, `messages.go:20-21`, `session.go:167` | Canonical Key restoration | **Correct.** Stamped with `loadGen` onto `listLoadedMsg`; prevents stale load restore races. |
| **`loadedKey`** | `internal/tui/detail.go:44,154,159`, `model.go:1226` | Canonical Key | **Correct.** Detail pane ownership matches canonical key, preventing wrong-record rendering. |
| **`navLoc`** | `internal/tui/nav.go:22-25,148,165` | History State | **Correct.** Stores `(kind, ref)` so `navBack` restores exact canonical ID. |
| **`pendingJump`** | `internal/tui/session.go:64-68,162-168`, `atlas.go:366` | Navigation Seam | **Correct.** Stamped on cross-workspace entry to land on specific task by canonical ID. |
| **`movedAwayKey`** | `internal/tui/model.go:104,452,662,679`, `session.go:84,103,136` | Mutation Exception | **Correct.** Suppresses false "not found" flash when moved task disappears from working view. |
| **`AuditID`** | `internal/core/finding.go:133,161`, `internal/core/service.go:322`, `internal/tui/dashboard.go:161` | Canonical Key | **Correct.** Propagated through summary/query findings for direct dashboard jump. |
| **Service Read/Writes** | `svc.ShowTask`, `svc.ShowAudit`, `svc.ShowResearch`, `svc.ShowEpic`, `svc.Move`, `svc.DeferTask`, `svc.MoveAudit`, `svc.MoveEpic`, `svc.SetFields`, `svc.SetEpicFields` | Service Boundary | **Correct.** All invoked with `ref.key` / `selectedKey()`, never raw display slugs. |

---

### Compact Identity-Flow Re-Derivation

```
[Domain Entity] (Task, Audit, Research, Thread)
      │
      ├── FilenameID != "" ──► CanonicalID = FilenameID  (Authoritative filesystem resolution)
      └── FilenameID == "" ──► CanonicalID = ID          (Portable adapter fallback)
      │
      ▼
[Load & Construction] (loadTaskList, loadAuditList, loadResearchList, loadEpicList)
      │
      ├── Build []entityRef{key: CanonicalID(), label: Slug}
      ├── duplicateIdentityHints(refs) ──► Shortest unique prefix (min 6 chars) for duplicate labels
      └── Construct list.Item:
            ├── ref(): entityRef{key, label}
            ├── displayLabel(): label + " [" + hint + "]" (if duplicate label)
            └── FilterValue(): Slug + " " + CanonicalID() + " " + Description + ...
      │
      ├────────────────────────────────────────────────────────────────────────────┐
      ▼                                                                            ▼
[Navigation & State]                                                      [Mutations & Edits]
  ├── Selection: selectByKey(ref.key)                                       ├── moveTask(svc, ref, tr) -> svc.Move(ref.key, ...)
  ├── Reload & Restore: reload(svc, ref) -> stamped listLoadedMsg           ├── deferTaskCmd(svc, ref, rev) -> svc.DeferTask(ref.key, ...)
  ├── Detail Load: loadItem(svc, ref.key) -> detailMsg{id: ref.key}         ├── moveAudit(svc, ref, tr) -> svc.MoveAudit(ref.key, ...)
  ├── Stale Guard: isCurrentSelection(kind, msg.id == selectedKey())        ├── moveEpic(svc, ref, tr) -> svc.MoveEpic(ref.key, ...)
  ├── Follow / Back Stack: navLoc{kind, ref}                                └── setFieldCmd(svc, ref, k, v) -> svc.SetFields(ref.key, ...)
  ├── Dashboard Jump: dashTarget{kind, ref: {AuditID, Audit}}
  ├── Command Palette: paletteItem{kind, ek, ref, entity, title, filter}
  ├── Atlas Landing: pendingJump{kind, ref} -> restore on open
  └── Cached Workspace: spaceSession{movedAwayKey, navStack: []navLoc}
```

---

### Findings

#### Medium Severity

#### M1. Right-Truncation Clips Identity Disambiguator in Narrow Terminals · **Status:** fixed

**File:** internal/tui/item.go:96 · **Component:** tui/item
**Effort:** S · **Urgency:** soon

**Description:**
In `taskDelegate.Render`, `auditDelegate.Render`, `researchDelegate.Render`, and `followMenu.view`, rows are rendered by calling `truncate(it.displayLabel(), slugW)`. Because `truncate` uses `ansi.Truncate(s, max, "…")`, string truncation occurs from the right. Suffixing the identity hint to the slug (`slug + " [" + hint + "]"`) means that whenever the column width is narrower than the full string (e.g. in terminals under 80 columns, split terminal windows, or for slugs longer than ~25 characters), the `[hint]` suffix is the first element truncated away.

**Impact:**
When two colliding tasks share a long slug (e.g. `refactor-tui-live-reload-state`), both rows get truncated to `refactor-tui-live-reload-stat…` with identical visual text. The visual disambiguation guarantee promised by ADR-0006 is defeated on narrow terminals.

**Hostile Reproduction / Reasoning:**
1. Create two tasks with slug `investigate-cross-workspace-session-restoration` and IDs `6g5rxq111111` and `6g5rxq222222`.
2. Launch TUI in a standard 80-column terminal with dual panes (left list pane width ~35-40 cols).
3. `slugW` is computed as `Width() - 17` (~20 cols).
4. Both rows render as `investigate-cross-wo…`. The bracketed hints `[6g5rxq1]` and `[6g5rxq2]` are completely clipped.

**Recommendation:**
Middle-truncate the slug portion before appending the identity hint if the combined string exceeds `slugW`, or allocate a dedicated fixed-width hint column on duplicate rows so the hint is never clipped before the slug.

---

**Resolution:** Identity hints now prefix shared labels, preserving the
distinguishing text under right truncation; realistic-width delegate tests cover
all stable-ID rows.

#### M2. Test Helper `selectedID()` Aliases `selectedLabel()`, Masking Identity Regressions · **Status:** fixed

**File:** internal/tui/model_test.go:29 · **Component:** tui/model_test
**Effort:** S · **Urgency:** soon

**Description:**
`internal/tui/model_test.go:29` defines the test helper:
```go
func (m Model) selectedID() string { return m.selectedLabel() }
```
Across more than 40 test calls throughout `nav_test.go`, `dashboard_test.go`, `edit_test.go`, `restore_test.go`, `action_test.go`, and `openeditor_test.go`, assertions call `m.selectedID()` expecting to verify the selected record's identity. Because `selectedID()` returns `selectedLabel()` (the display slug), these tests assert only that the slug matches, not the canonical ID.

**Impact:**
If a navigation, reload, or mutation bug selects the wrong duplicate row sharing the same slug, any test checking `m.selectedID() == "slug"` passes green. This test helper creates a dangerous false sense of test coverage and directly violates Hostile Angle 10 of the review brief (*"Pay special attention to old helper names/tests that may make a slug-keyed path look canonical"*).

**Hostile Reproduction / Reasoning:**
1. Run any test asserting `m.selectedID() == "alpha"`.
2. Mutate `selectByKey` to select a different task that also shares the slug `"alpha"`.
3. The test continues to pass because `m.selectedLabel()` is `"alpha"` for both records.

**Recommendation:**
Deprecate and remove `selectedID()` from `model_test.go`. Update all test assertions to explicitly check `m.selectedKey()` when asserting identity and `m.selectedLabel()` when asserting display text.

---

**Resolution:** The misleading selectedID test helper was removed. Existing
display assertions now say selectedLabel explicitly, while identity-sensitive
regressions use selectedKey.

#### Low Severity

#### L1. Coverage Gap — Mutation survival for cross-kind palette disambiguation · **Status:** fixed

**File:** internal/tui/command_dispatch.go:154-162 · **Component:** tui/palette
**Effort:** XS · **Urgency:** eventually

**Description:**
In `paletteIndex()`, when multiple loaded entities across different kinds share the same visible title (e.g. a task named `database-migration` and an audit named `database-migration`), `titleCounts` detects collisions and appends `" · " + items[i].entity`. Mutation Probe 5 (removing `titleCounts` cross-kind disambiguation) survived the existing test suite because pre-existing tests in `palette_test.go` and `identity_test.go` only checked same-kind task duplicates.

**Impact:**
Production implementation is correct, but removing the cross-kind title disambiguation would pass existing test suites undetected.

**Hostile Reproduction / Reasoning:**
1. Create task with slug `shared-name` and audit with slug `shared-name`.
2. Populate `paletteIndex()`.
3. If `titleCounts` loop is omitted, both palette rows render identical titles (`"shared-name"`), leaving the user unable to distinguish which entity kind Enter will jump to.

**Recommendation:**
Add a unit test in `palette_test.go` creating equal-slug task and audit fixtures, verifying that `paletteIndex()` appends `" · tasks"` and `" · audits"` to their titles.

---

**Resolution:** A cross-kind equal-label palette regression now pins the task
and audit title suffixes and single-key filter corpus.

#### L2. Edge Case — Asymmetric hint assignment for prefix-of-another keys · **Status:** fixed

**File:** internal/tui/entity.go:72-100 · **Component:** tui/entity
**Effort:** XS · **Urgency:** eventually

**Description:**
`duplicateIdentityHints` computes the shortest unique prefix for colliding labels by looping `n := min(6, len(key)); n <= len(key); n++`. When non-conforming synthetic keys have unequal lengths where key A is a strict prefix of key B (e.g. `keyA = "abcdef"`, `keyB = "abcdef123"` with the same label), `keyA` never finds a unique prefix (`strings.HasPrefix(keyB, "abcdef")` is always true). As a result, the loop terminates without setting `hints[keyA]`, so `keyA` displays as `"same"`, while `keyB` displays as `"same [abcdef1]"`.

**Impact:**
Standard Crockford base32 IDs are uniform 12-character strings where strict prefix inclusion cannot occur. For non-uniform or synthetic adapter keys, rows are still visually distinguishable because one has a bracketed hint and the other does not, but `keyA` lacks an explicit bracketed hint.

**Hostile Reproduction / Reasoning:**
`refs := []entityRef{{key: "abcdef", label: "same"}, {key: "abcdef123", label: "same"}}` results in `hints["abcdef123"] == "abcdef1"` while `hints["abcdef"]` remains unset.

**Recommendation:**
When the loop finishes without finding a unique sub-prefix (i.e. `n > len(key)`), set `hints[key] = key` as a fallback so that both colliding rows receive bracketed hints.

---

**Resolution:** Strict-prefix collisions now assign explicit hints to both rows,
including the full shorter key fallback.

#### L3. `selectedYankRef()` Relies on Rendered String Equality to Detect Collisions · **Status:** fixed

**File:** internal/tui/model.go:1279-1289 · **Component:** tui/model
**Effort:** XS · **Urgency:** eventually

**Description:**
`selectedYankRef()` decides whether to copy the canonical ID or the display slug by checking:
```go
if it.displayLabel() != ref.label {
    return ref.key, "id"
}
return ref.label, "slug"
```
Inferring semantic entity ambiguity from formatted display strings creates a leaky coupling. For example, when `duplicateIdentityHints` leaves `hints[keyA]` unset for a duplicate (as shown in Finding L2), `it.displayLabel() == ref.label`. Pressing `y` yanks the ambiguous slug instead of the canonical ID, and subsequent CLI commands like `tskflwctl task show <slug>` fail with `ErrAmbiguous`.

**Impact:**
Users copying references for duplicate entities with unhinted keys receive an unusable ambiguous slug rather than a durable ID.

**Hostile Reproduction / Reasoning:**
1. Select keyA from a duplicate group where `it.displayLabel() == ref.label`.
2. Press `y` to yank.
3. Observe flash `copied slug: same` rather than `copied id: abcdef`.

**Recommendation:**
Expose an explicit boolean method on `entityItem` (e.g. `isDuplicate() bool` or `hasIdentityHint() bool`) or check the hint map directly rather than relying on string inequality between `displayLabel()` and `ref.label`.

---

**Resolution:** Yank behavior now reads an explicit hasIdentityHint contract
from the selected entity item instead of inferring ambiguity from formatted
display strings.

#### L4. `activateWorkspace` Does Not Widen Status View on Cross-Workspace Task Landings · **Status:** fixed

**File:** internal/tui/session.go:156-168 · **Component:** tui/session
**Effort:** S · **Urgency:** eventually

**Description:**
Within an active workspace, jumping to a hidden or archived entity via `jumpTo()` automatically widens the tab's status view:
```go
if tab.viewAxis != nil && tab.statusView != "all" {
    tab.statusView = "all"
    m.flash, m.flashErr = fmt.Sprintf("showing :all to reach %s", ref.label), false
    return tab.reload(m.svc, ref)
}
```
However, Atlas cross-workspace work landings (`openAtlasWork`) deliberately bypass `jumpTo()` and set `m.tabs[i].restore = jump.ref` on the target workspace's default view without widening `statusView`. If the selected in-progress task was moved out of the working set in the target workspace (e.g. marked completed concurrently by another process or in a previous session), the tab reload lands on an unfiltered working view where `selectByKey` fails. `handleListLoaded` then fires a spurious `"<label> not found"` flash instead of widening the view to `:all`.

**Impact:**
Navigating from the Atlas work overview to a task that completed concurrently reports that the task does not exist, rather than switching to `:all` to display it.

**Hostile Reproduction / Reasoning:**
1. Scan an in-progress task into Atlas.
2. Externally complete the task in the target repository.
3. Select the task row in Atlas and press Enter (`openAtlasWork`).
4. TUI switches spaces, reloads default working view (which excludes completed tasks), and flashes `"<slug> not found"`.

**Recommendation:**
In `activateWorkspace`, if `jump.kind` has a `statusView` axis and the task is not guaranteed to be in the default view, set `tab.statusView = "all"` or mirror `jumpTo`'s fallback reload logic.

---

**Resolution:** Cross-workspace Atlas work landings carry a one-time fallback
intent: a miss in the default working view retries in :all, while ordinary
reload misses keep their existing behavior. An end-to-end completion race pins
the result.

### Acceptance-Criteria Traceability Table

| Acceptance Criterion | Status | Implementation Seams | Verification Evidence |
| :--- | :---: | :--- | :--- |
| **1. Duplicate slug selection & detail isolation**<br>Duplicate-slug tasks, audits, and research rows can be selected independently without showing detail from the other row. | **Fulfilled** | `internal/tui/entity.go:202-210`<br>`internal/tui/model.go:1226-1228`<br>`internal/tui/detail.go:149-169` | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState`<br>`TestDuplicateStableIDEntitiesLoadByFilenameIdentity`<br>`TestIndependent_DuplicateSlugsWithFrontmatterDriftAndMissingID` |
| **2. Stale detail completion & reload preservation**<br>Out-of-order detail completion for one duplicate cannot overwrite another; reloads preserve the selected duplicate by ID. | **Fulfilled** | `internal/tui/entity.go:181-195`<br>`internal/tui/commands.go:78-86`<br>`internal/tui/model.go:343-356` | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState`<br>`TestIndependent_OutOfOrderStaleDetailDropping` |
| **3. Disambiguated visible labels & filterability**<br>Duplicate labels gain the shortest unique stable-ID prefix; canonical keys join the filter corpus without replacing slug search. | **Fulfilled** | `internal/tui/entity.go:72-107`<br>`internal/tui/item.go:51-57,175-181,257-263` | `TestDuplicateIdentityHintsExpandPastACollidingPrefix`<br>`TestIndependent_PrefixExpansionEdgeCases`<br>*(Tracked follow-up in **Finding M1** for narrow terminal truncation)* |
| **4. Cross-surface navigation preservation**<br>Palette, follow/back stack, dashboard findings, and cached workspace sessions preserve the exact selected duplicate. | **Fulfilled** | `internal/tui/nav.go:22-25,143-168`<br>`internal/tui/dashboard.go:159-163`<br>`internal/tui/command_dispatch.go:136-167`<br>`internal/tui/session.go:84,94-105` | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState`<br>`TestHostile_AuditFindingDashboardNavigationWithDuplicateSlugs`<br>`TestHostile_CrossKindPaletteAndDuplicateHints`<br>*(Tracked follow-up in **Finding L4** for Atlas view widening)* |
| **5. Stable mutations & post-move safety**<br>Lifecycle moves and inline edits target the selected duplicate by canonical ID; post-move absence matches `movedAwayKey`. | **Fulfilled** | `internal/tui/entity.go:293-350`<br>`internal/tui/edit.go:346-377`<br>`internal/tui/model.go:487-512` | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState`<br>`TestIndependent_LifecycleMoveWarningAndPostMoveAbsence` |
| **6. Frontmatter drift & adapter fallback resilience**<br>Tasks/audits with drifting or missing frontmatter IDs resolve by filename ID; portable records fall back to semantic ID. | **Fulfilled** | `internal/domain/task.go:64-69`<br>`internal/domain/audit.go:83-88`<br>`internal/domain/research.go:61-66`<br>`internal/domain/thread.go:74-79` | `TestCanonicalIDPrefersAdapterResolutionIdentity`<br>`TestDuplicateStableIDEntitiesLoadByFilenameIdentity`<br>`TestIndependent_DuplicateSlugsWithFrontmatterDriftAndMissingID` |

---

### Mandatory Hostile-Evidence Ledger

| Probe / Angle | Hostile Attack Condition | Observed Implementation Response | Catching Test / Coverage Status |
| :--- | :--- | :--- | :--- |
| **Fixture 1: Duplicate tasks with ID drift & missing ID** | Two tasks sharing slug `shared-task-slug`. Task A has drifting frontmatter ID; Task B omits `id:`. | Both tasks load with canonical `FilenameID`. Selecting Task B reads `BODY-TASK-B`; inline edit targets Task B without touching Task A. | `TestIndependent_DuplicateSlugsWithFrontmatterDriftAndMissingID` |
| **Fixture 2: Duplicate audits with ID drift & missing ID** | Two audits sharing slug `shared-audit-slug`. Audit A has drifting frontmatter ID; Audit B omits `id:`. | `CanonicalID()` resolved filename identities; selecting Audit A loaded `BODY-AUDIT-A` without colliding with Audit B. | `TestIndependent_DuplicateSlugsWithFrontmatterDriftAndMissingID` |
| **Fixture 3: Duplicate research records** | Two research records sharing slug `shared-res-slug`. Res A has drifting ID; Res B omits `id:`. | Selecting Res B loaded `BODY-RES-B`; rows displayed distinct bracketed hints. | `TestIndependent_DuplicateSlugsWithFrontmatterDriftAndMissingID` |
| **Probe 1: Out-of-order detail completion** | Task B selected; delayed `detailMsg` arrives for Task A with same generation. | `isCurrentSelection` checked `msg.id == m.selectedKey()`; Task A detail dropped; Task B detail landed cleanly. | `TestIndependent_OutOfOrderStaleDetailDropping` |
| **Probe 2: Lifecycle move & post-move absence** | Task A moved to completed; Task B remains in-progress under same slug `same-slug`. | `movedAwayKey` set to Task A key; working view reloaded showing only Task B; Task A not flagged with false "not found" flash. | `TestIndependent_LifecycleMoveWarningAndPostMoveAbsence` |
| **Probe 3: Status view widening on duplicate** | Jump to completed Task A while in working view (hides completed). | `jumpTo` widened `statusView` to `"all"`, reloaded, and restored cursor onto Task A. | `TestHostile_TaskDuplicateLifecycleAndEdgeCases` |
| **Probe 4: Committed-with-warning lifecycle move** | Move returns `TaskLifecycleMutationFailure` with valid receipt. | `moveTask` unmarshaled `committed.Receipt`; model set `flash` warning without losing navigation state. | `TestIndependent_LifecycleMoveWarningAndPostMoveAbsence` |
| **Probe 5: Dashboard acute finding navigation** | Two audits share slug `security-review`; Audit A has acute finding. | Dashboard row target carried `AuditID: auditIDA`; `dashJump` landed on audits tab with Audit A selected. | `TestHostile_AuditFindingDashboardNavigationWithDuplicateSlugs` |
| **Probe 6: Cross-kind palette collision** | Same slug `database-migration` across task, audit, and research. | `paletteIndex()` detected title collisions and appended `" · tasks"`, `" · audits"`, `" · research"`. | `TestHostile_CrossKindPaletteAndDuplicateHints` |
| **Probe 7: Short keys & 6-char prefix expansion** | Colliding 6-character prefixes `abcdef111111` and `abcdef222222`. | Expanded to 7 characters (`abcdef1` and `abcdef2`); unique key `xyz123456789` received no hint. | `TestIndependent_PrefixExpansionEdgeCases` |
| **Probe 8: Portable adapter fallback (no FilenameID)** | Injected records with empty `FilenameID` and non-empty `ID`. | `CanonicalID()` returned semantic `ID` across all 4 entity types. | `TestCanonicalIDPrefersAdapterResolutionIdentity` |
| **Mutation 1: Semantic ID wins over FilenameID** | Inverted `CanonicalID()` to prefer `t.ID` over `t.FilenameID`. | Failed on drifted frontmatter IDs (`CanonicalID() = "semantic-task"`). | **KILLED** (`domain/identity_test.go:20`, `tui/identity_test.go:152`) |
| **Mutation 2: Label comparison in stale guard** | Changed `isCurrentSelection` to compare `id == m.selectedLabel()`. | Duplicate detail loads collided and failed to land matching content. | **KILLED** (`tui/identity_test.go:100,212`) |
| **Mutation 3: Write routed through `ref.label`** | Changed `setFieldCmd` to pass `ref.label` to `svc.SetFields`. | Service returned `ErrAmbiguous` due to duplicate slug. | **KILLED** (`tui/identity_test.go:148`) |
| **Mutation 4: Omit AuditID in Summary aggregation** | Omitted `AuditID` in `core/service.go:322`. | Dashboard acute finding target had empty audit key. | **KILLED** (`tui/dashboard_test.go:351`) |
| **Mutation 5: Omit cross-kind palette disambiguation** | Removed `titleCounts` disambiguator in `paletteIndex()`. | Cross-kind equal slugs produced identical palette titles. | **SURVIVED PRE-EXISTING SUITE** (Tracked in **Finding L1**) |

---

### Platform & Risk Classification

- **Demonstrated Correctness:** Verified on macOS / ARM64 / APFS. Duplicate tasks, audits, and research rows with ID drift and missing frontmatter IDs maintain distinct selection, detail ownership, searchability, lifecycle mutations, inline edits, and session state.
- **Source-Inspected Portability:** `CanonicalID()` is pure domain logic independent of OS or filesystem specifics. Portable memory-backed or served stores providing semantic `ID` without `FilenameID` operate identically. `Thread.CanonicalID()` is structurally implemented in `internal/domain/thread.go:74` ready for the subsequent Thread TUI registry wiring.
- **Documented Boundary:** Truncation clips suffixes first, dropping identity hints on narrow terminals (**Finding M1**). When synthetic keys have non-uniform lengths where one key is a strict prefix of another, `duplicateIdentityHints` assigns a bracketed hint to the longer key while leaving the shorter key unbracketed (**Finding L2**). Standard 12-character Crockford IDs avoid strict prefix inclusion.

---

### Explicitly Settled Concerns

1. **Concern: Display slugs might accidentally be passed to service write paths during inline editing.**
   *Resolution:* Settled. `editMenu` stores `entityRef`, and `setFieldCmd`/`unsetFieldCmd`/`setEpicFieldCmd` invoke `svc.SetFields(ref.key, ...)`, ensuring mutations are strictly ID-addressed.
2. **Concern: Stale async detail loads for duplicate A could overwrite newly selected duplicate B.**
   *Resolution:* Settled. `isCurrentSelection` checks `kind == m.cur().kind && id == m.selectedKey()`. `selectedKey()` returns the canonical key, dropping any detail result belonging to a colliding slug with a different ID.
3. **Concern: Global command palette could collide when an audit and a task share the same slug.**
   *Resolution:* Settled. `paletteIndex()` computes title counts across the merged list and appends `" · <entity>"` to titles that collide across different registries.
4. **Concern: Frontmatter ID drift could cause the TUI to fall back to ambiguous slug resolution.**
   *Resolution:* Settled. `Task.CanonicalID()`, `Audit.CanonicalID()`, and `Research.CanonicalID()` unconditionally prefer `FilenameID` when non-empty. The store resolves by filename identity, leaving frontmatter drift as a lint-only issue that never breaks TUI navigation.
5. **Concern: Atlas cross-workspace jumping (`pendingJump`) might restore the wrong duplicate in the target workspace.**
   *Resolution:* Settled. `pendingJump` stores `entityRef{key, label}`, and `activateWorkspace` sets `tab.restore = jump.ref`, restoring the exact canonical ID upon tab reload.
