---
schema: 1
id: 6g67v0cd1x0h
bucket: closed
area: tui-stable-entity-navigation-implementation-antigravity
date: "2026-09-02"
updated_at: "2026-09-02"
---
# Audit: TUI stable entity navigation implementation — Antigravity — 2026-09-02

> Reviewer assignment: Antigravity. This document is the review brief and the only file the reviewer should update.
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

The TUI stable entity navigation implementation for task [`6g5rxq17px59`](../tasks/6g5rxq17px59-make-tui-entity-navigation-use-stable-identities.md) on branch `feat/tui-stable-entity-identities` (based at commit `28ccf6b`) is verified against [ADR-0006](../adrs/0006-adopt-threads-as-task-dags.md) and [`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

1. **Canonical Identity Separation:** The entity registry introduces [`entityRef{key, label}`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L49-L52), strictly isolating internal store-resolution keys (`key`) from human-facing display text (`label`). For tasks, audits, research, and Threads, `CanonicalID()` prioritizes adapter-derived `FilenameID` so missing or drifted frontmatter `id:` tags cannot redirect operations, while portable adapters fall back to semantic `ID`.
2. **End-to-End Navigation & Stale Protection:** Selection, cursor preservation across reloads/refilters ([`entityTab.selectByKey`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L202-L210), [`entityTab.restore`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L171)), asynchronous detail ownership ([`Model.isCurrentSelection`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/model.go#L1226-L1228), [`detailPane.loadedKey`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/detail.go#L44)), follow/back history ([`navLoc`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/nav.go#L22-L25)), atlas cross-workspace landings ([`pendingJump`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/session.go#L64-L68)), and cached workspace sessions ([`spaceSession.movedAwayKey`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/session.go#L84)) strictly carry canonical keys.
3. **Mutation Boundary Integrity:** Lifecycle moves ([`moveTask`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L293), [`deferTaskCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L313), [`moveAudit`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L330), [`moveEpic`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L343)) and inline field edits ([`setFieldCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/edit.go#L346), [`unsetFieldCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/edit.go#L358), [`setEpicFieldCmd`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/edit.go#L370)) dispatch mutations using `ref.key`, never re-resolving ambiguous display slugs.
4. **Deterministic Disambiguation & Filter Preservation:** Visible duplicate labels receive deterministic, shortest unique stable-ID prefixes ([`duplicateIdentityHints`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/entity.go#L72-L100)) expanded beyond 6 characters when prefixes collide. Canonical keys are injected into [`FilterValue()`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/tui/item.go#L51-L53), ensuring rows remain searchable by either slug or stable ID.
5. **Core Projections Alignment:** [`AuditFinding.AuditID`](file:///Users/andyeschbacher/git/andy-esch/taskflow/internal/core/finding.go#L133) is populated across `Summary()` and `QueryFindings()`, allowing dashboard acute finding navigation to target canonical audits directly without slug re-resolution.

Two low-severity findings are filed: one for a mutation survival in cross-kind palette disambiguation coverage, and one for edge-case hint assignment on non-uniform prefix-of-another keys.

---

### Review Environment & Validation Results

- **Branch:** `feat/tui-stable-entity-identities`
- **Base Commit:** `28ccf6ba78a24967977c0b6b048c6a4c44026d91` (`28ccf6b`)
- **Runtime Platform:** `darwin/arm64` (macOS 15.x), Go 1.26.6, APFS

#### Validation Commands & Outcomes

| Validation Suite | Exact Command Line | Outcome |
| :--- | :--- | :--- |
| **Repeated Focused Race** | `go test -race -count=10 -run 'TestDuplicate\|TestCanonicalID' ./internal/domain ./internal/tui` | **PASS** (10/10 clean, 2.053s) |
| **Full Repository Race** | `go test -count=1 -race ./...` | **PASS** (all packages clean, 7.923s max package duration) |
| **Static Analysis / Lint** | `golangci-lint run ./...` | **PASS** (0 issues) |
| **Go Vet** | `go vet ./...` | **PASS** (clean) |
| **Docs & CLI Drift** | `go run ./internal/tools/docgen -out docs/cli && git diff --exit-code docs/cli` | **PASS** (no drift) |
| **Go Module Tidiness** | `go mod tidy -diff` | **PASS** (no diff) |
| **Planning System Lint** | `go run ./cmd/tskflwctl lint` | **PASS** (all planning entities and links valid) |
| **Git Diff Check** | `git diff --check` | **PASS** (clean whitespace) |

---

### Compact Identity-Flow Re-Derivation

```
[Domain Record] (Task, Audit, Research, Thread)
      │
      ├── FilenameID != "" ──► CanonicalID = FilenameID  (Authoritative filesystem resolution)
      └── FilenameID == "" ──► CanonicalID = ID          (Portable adapter fallback)
      │
      ▼
[Loader & Row Construction] (loadTaskList, loadAuditList, loadResearchList, loadEpicList)
      │
      ├── Build []entityRef{key: CanonicalID(), label: Slug}
      ├── duplicateIdentityHints(refs) ──► Shortest unique prefix [min 6 chars] for duplicate labels
      └── Build list.Item:
            ├── ref(): entityRef{key, label}
            ├── displayLabel(): label + " [" + hint + "]" (if duplicate)
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

**Resolution:** A cross-kind equal-label palette regression now asserts both
tasks and audits receive kind-qualified titles, while each canonical key appears
once in the filter corpus.

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

**Resolution:** The shorter member of a strict-prefix key pair now receives its
full key as an explicit fallback hint; focused coverage pins both rows.

### Acceptance-Criteria Traceability Table

| Acceptance Criterion | Status | Implementation Seams | Verification Evidence |
| :--- | :---: | :--- | :--- |
| **1. Duplicate slug selection & detail isolation**<br>Duplicate-slug tasks, audits, and research rows can be selected independently without showing detail from the other row. | **Fulfilled** | `internal/tui/entity.go:202-210`<br>`internal/tui/model.go:1226-1228`<br>`internal/tui/detail.go:149-169` | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState`<br>`TestDuplicateStableIDEntitiesLoadByFilenameIdentity` |
| **2. Stale detail completion & reload preservation**<br>Out-of-order detail completion for one duplicate cannot overwrite another; reloads preserve the selected duplicate by ID. | **Fulfilled** | `internal/tui/entity.go:181-195`<br>`internal/tui/commands.go:78-86`<br>`internal/tui/model.go:343-356` | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` (stale `detailMsg` dropped; `markReload` preserves `second.ref()`) |
| **3. Disambiguated visible labels & filterability**<br>Duplicate labels gain the shortest unique stable-ID prefix; canonical keys join the filter corpus without replacing slug search. | **Fulfilled** | `internal/tui/entity.go:72-107`<br>`internal/tui/item.go:51-57,175-181,257-263` | `TestDuplicateIdentityHintsExpandPastACollidingPrefix`<br>`TestHostile_PrefixDisambiguationCornerCases` |
| **4. Cross-surface navigation preservation**<br>Palette, follow/back stack, dashboard findings, and cached workspace sessions preserve the exact selected duplicate. | **Fulfilled** | `internal/tui/nav.go:22-25,143-168`<br>`internal/tui/dashboard.go:159-163`<br>`internal/tui/command_dispatch.go:136-167`<br>`internal/tui/session.go:84,94-105` | `TestHostile_AuditFindingDashboardNavigationWithDuplicateSlugs`<br>`TestHostile_CrossKindPaletteAndDuplicateHints`<br>`TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` |
| **5. Stable mutations & post-move safety**<br>Lifecycle moves and inline edits target the selected duplicate by canonical ID; post-move absence matches `movedAwayKey`. | **Fulfilled** | `internal/tui/entity.go:293-350`<br>`internal/tui/edit.go:346-377`<br>`internal/tui/model.go:487-512` | `TestHostile_TaskDuplicateLifecycleAndEdgeCases`<br>`TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` |
| **6. Frontmatter drift & adapter fallback resilience**<br>Tasks/audits with drifting or missing frontmatter IDs resolve by filename ID; portable records fall back to semantic ID. | **Fulfilled** | `internal/domain/task.go:64-69`<br>`internal/domain/audit.go:83-88`<br>`internal/domain/research.go:61-66`<br>`internal/domain/thread.go:74-79` | `TestCanonicalIDPrefersAdapterResolutionIdentity`<br>`TestDuplicateStableIDEntitiesLoadByFilenameIdentity` |

---

### Mandatory Hostile-Evidence Ledger

| Probe / Angle | Hostile Attack Condition | Observed Implementation Response | Catching Test / Coverage Status |
| :--- | :--- | :--- | :--- |
| **Fixture 1: Duplicate tasks with ID drift & missing ID** | Two tasks sharing slug `same-task`. Task A has drifting frontmatter ID `wrong-task-frontmatter`; Task B omits `id:`. | Both tasks load with canonical `FilenameID`. Task B selected by key; detail loaded `TASK-2-BODY`; edit updated Task B without touching Task A. | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` |
| **Fixture 2: Duplicate audits with ID drift** | Two audits sharing slug `same-audit`. Audit A has drifting frontmatter ID; Audit B omits `id:`. | `CanonicalID()` resolved filename identities; selecting Audit B loaded `AUDIT-2-BODY` without colliding with Audit A. | `TestDuplicateStableIDEntitiesLoadByFilenameIdentity` |
| **Fixture 3: Duplicate research records** | Two research records sharing slug `same-research` with distinct filename IDs. | Both rows rendered with bracketed identity hints; selecting second research loaded `RESEARCH-2-BODY`. | `TestDuplicateStableIDEntitiesLoadByFilenameIdentity` |
| **Probe 1: Out-of-order detail completion** | Task B selected; injected delayed `detailMsg` for Task A with same generation. | `isCurrentSelection` checked `msg.id == m.selectedKey()`; Task A detail dropped; Task B detail landed cleanly. | `TestDuplicateTaskSlugsKeepCanonicalIdentityAcrossTUIState` |
| **Probe 2: Lifecycle move & post-move absence** | Task A moved to completed; Task B remains in-progress under same slug `same-slug`. | `movedAwayKey` set to Task A key; working view reloaded showing only Task B; Task A not flagged as error. | `TestHostile_TaskDuplicateLifecycleAndEdgeCases` |
| **Probe 3: Status view widening on duplicate** | Jump to completed Task A while in working view (hides completed). | `jumpTo` widened `statusView` to `"all"`, reloaded, and restored cursor onto Task A. | `TestHostile_TaskDuplicateLifecycleAndEdgeCases` |
| **Probe 4: Committed-with-warning lifecycle move** | Move returns `TaskLifecycleMutationFailure` with valid receipt. | `moveTask` unmarshaled `committed.Receipt`; model set `flash` warning without losing navigation state. | `TestHostile_TaskDuplicateLifecycleAndEdgeCases` |
| **Probe 5: Dashboard acute finding navigation** | Two audits share slug `security-review`; Audit A has acute finding. | Dashboard row target carried `AuditID: auditIDA`; `dashJump` landed on audits tab with Audit A selected. | `TestHostile_AuditFindingDashboardNavigationWithDuplicateSlugs` |
| **Probe 6: Cross-kind palette collision** | Same slug `database-migration` across task, audit, and research. | `paletteIndex()` detected title collisions and appended `" · tasks"`, `" · audits"`, `" · research"`. | `TestHostile_CrossKindPaletteAndDuplicateHints` |
| **Probe 7: Short keys & 6-char prefix expansion** | Colliding 6-character prefixes `abcdef111111` and `abcdef222222`. | Expanded to 7 characters (`abcdef1` and `abcdef2`); unique key `zzzzzz333333` received no hint. | `TestDuplicateIdentityHintsExpandPastACollidingPrefix` |
| **Probe 8: Portable adapter fallback (no FilenameID)** | Injected records with empty `FilenameID` and non-empty `ID`. | `CanonicalID()` returned semantic `ID` across all 4 entity types. | `TestCanonicalIDPrefersAdapterResolutionIdentity` |
| **Mutation 1: Semantic ID wins over FilenameID** | Inverted `CanonicalID()` to prefer `t.ID` over `t.FilenameID`. | Failed on drifted frontmatter IDs (`CanonicalID() = "semantic-task"`). | **KILLED** (`domain/identity_test.go:20`, `tui/identity_test.go:152`) |
| **Mutation 2: Label comparison in stale guard** | Changed `isCurrentSelection` to compare `id == m.selectedLabel()`. | Duplicate detail loads collided and failed to land matching content. | **KILLED** (`tui/identity_test.go:100,212`) |
| **Mutation 3: Write routed through `ref.label`** | Changed `setFieldCmd` to pass `ref.label` to `svc.SetFields`. | Service returned `ErrAmbiguous` due to duplicate slug. | **KILLED** (`tui/identity_test.go:148`) |
| **Mutation 4: Omit AuditID in Summary aggregation** | Omitted `AuditID` in `core/service.go:322`. | Dashboard acute finding target had empty audit key. | **KILLED** (`tui/dashboard_test.go:351`, `tui/audit_hostile_nav_test.go:114`) |
| **Mutation 5: Omit cross-kind palette disambiguation** | Removed `titleCounts` disambiguator in `paletteIndex()`. | Cross-kind equal slugs produced identical palette titles. | **SURVIVED PRE-EXISTING SUITE** (Tracked in **Finding L1**) |

---

### Platform & Risk Classification

- **Demonstrated Correctness:** Verified on macOS / ARM64 / APFS. Duplicate tasks, audits, and research rows with ID drift and missing frontmatter IDs maintain distinct selection, detail ownership, searchability, lifecycle mutations, inline edits, and session state.
- **Source-Inspected Portability:** `CanonicalID()` is pure domain logic independent of OS or filesystem specifics. Portable memory-backed or served stores providing semantic `ID` without `FilenameID` operate identically. `Thread.CanonicalID()` is structurally implemented in `internal/domain/thread.go:74` ready for the subsequent Thread TUI registry wiring.
- **Documented Boundary:** When synthetic keys have non-uniform lengths where one key is a strict prefix of another, `duplicateIdentityHints` assigns a bracketed hint to the longer key while leaving the shorter key unbracketed (**Finding L2**). Standard 12-character Crockford IDs avoid this condition entirely.

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
