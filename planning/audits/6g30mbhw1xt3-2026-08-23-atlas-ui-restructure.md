---
schema: 1
id: 6g30mbhw1xt3
bucket: closed
area: atlas-ui-restructure
date: "2026-08-23"
---

# Audit: atlas-ui-restructure — 2026-08-23

Scope: Adversarial multi-lens review of the atlas UI restructure in branch `feat/atlas-ui-restructure` (`v0.17.0..HEAD`, commits `8969c05` and `150570b`). Evaluated across five distinct passes: Correctness/Concurrency, Architecture, UX/Rendering, Test Quality, and Planning Integrity. Covers `internal/tui/atlas.go`, `internal/tui/atlas_test.go`, `internal/tui/help.go`, `internal/tui/keys.go`, `internal/tui/model.go`, `internal/tui/session.go`, `internal/tui/view.go`, `.golangci.yml`, and `planning/`.

> **Verdict:** The restructuring delivers a major qualitative improvement: the dual-view atlas (`spaces` ⇄ `work` via `v`), the pinned entry-point band beneath the table, and the promotion of the atlas into the `[`/`]` page ring are coherent and well-conceived. Architectural boundaries are strictly respected (`go list -deps ./internal/tui` remains clean of secondary adapters and discovery). However, two high-severity concurrency/caching bugs affect `pendingJump` and workspace reactivation, and the spaces table has a column alignment shear caused by unpadded inline breakdowns.

## Findings

#### H1. `pendingJump` intent leaks across workspace switches on spaces selection or atlas exit  · **Status:** fixed

**File:** `internal/tui/atlas.go:255-278,296-302`, `internal/tui/session.go:148-156` | **Component:** tui/session
**Effort:** XS · **Urgency:** acute

`Model.pendingJump` carries the "land on this specific task" intent when entering a space from the work view (`openAtlasWork`). However, `pendingJump` is only cleared when `activateWorkspace` consumes it or when `workspaceOpenedMsg.err != nil`. If a user selects a task in the work view (`openAtlasWork` sets `pendingJump` and fires async `openWorkspace`), but before that open completes (or immediately after switching to spaces view) triggers `openAtlasSelection()`, `openAtlasSelection` does not clear `pendingJump`.

When the new workspace opens successfully, `activateWorkspace()` observes `m.pendingJump.set == true` and fires `loads = m.jumpTo(jump.kind, jump.id)` targeting the task slug from the *previous* space inside the *new* space. This results in spurious view widening to `:all` or false `<slug> not found` flashes. `pendingJump` is likewise not cleared if the user leaves the atlas via `esc`, `a`, or `[`/`]`.

**Recommendation:** Explicitly clear `m.pendingJump = pendingJump{}` at the start of `openAtlasSelection()`, `enterAtlas()`, and `exitAtlas()`.

**Fixed 2026-08-23. Verified worse than reported.** A throwaway test reproduced it, and
the failure is quieter than the finding assumed: in a fixture where two spaces both contain
a task slugged `alpha`, the stale intent did not produce a "not found" flash — it silently
selected a *different task that happens to share the slug*, because slugs are unique only
within a tree. `openAtlasSelection` now clears the intent as its first act, so arming any
open supersedes a previous one and `openAtlasWork` re-arms afterwards. Pinned by
`TestAtlasStaleLandingIntentCannotFollowIntoAnotherSpace`.

---

#### H2. Work view jump into a cached workspace skips tab reload, leaving data stale or dropping newly-started tasks  · **Status:** fixed

**File:** `internal/tui/session.go:148-161`, `internal/tui/nav.go:178-196` | **Component:** tui/session
**Effort:** S · **Urgency:** acute

When `openAtlasWork()` enters a workspace that was already visited in the active session, `activateWorkspace()` restores the cached `spaceSession` (where `tab.loaded == true`). In `activateWorkspace`, the `m.pendingJump.set` branch executes `loads = m.jumpTo(jump.kind, jump.id)` instead of `m.reloadAll()`.

However, `jumpTo()` only calls `tab.reload()` if `!tab.loaded`. When `tab.loaded` is already `true`, `jumpTo()` simply calls `tab.selectByID(id)` on the in-memory item list without reading from `m.svc`. If a task was started, renamed, or modified on disk while the user was browsing another space, `selectByID()` fails (causing an unintended escalation to `:all` or a `<slug> not found` flash). Even if `selectByID()` succeeds, neither the task tab nor any other tab (`epics`, `audits`, `research`, or `dashboard`) is refreshed from disk, violating `activateWorkspace`'s contract ("Every loaded surface is refreshed so preserved cursors/filters do not imply preserved filesystem data"). In contrast, entering a cached space from the `spaces` view properly executes `m.reloadAll()`.

**Recommendation:** When `m.pendingJump.set` is true during `activateWorkspace` on a restored session, mark the restore ID on the target tab and execute `m.reloadAll()` so that all loaded surfaces refresh from the filesystem while landing on the selected task.

**Fixed 2026-08-23.** Reproduced: entering a space from the work view, returning, writing a
new in-progress task to that tree, then re-entering rendered the same two cached rows with
the new task absent. `jumpTo` is no longer used on this path — the target tab's `restore`
id is stamped and the ordinary `reloadAll` runs, giving the landing and fresh data in one
pass instead of selecting against rows read on the previous visit. Pinned by
`TestAtlasWorkReEntryRefreshesARestoredSession`.

---

#### M1. Unpadded inline counts breakdown shears column alignment for `⚠` attention markers and dates  · **Status:** fixed

**File:** `internal/tui/atlas.go:611-624` | **Component:** tui/atlas
**Effort:** S · **Urgency:** soon

In `spaceRows()`, `atlasCountsLine()` returns an unpadded, variable-length breakdown (e.g., `"4 epics · 1 audit"`, `"1 epic"`, or `""`). This string is concatenated inline directly between the `▸active` count and the attention marker cell (`strings.Repeat(" ", attentionW)` / `padRight("⚠...", attentionW)`) and the recency date cell (`dates[i]`).

Because `counts` has no fixed width or padding across spaces, the column positions of `⚠` and the recency date shift on every line depending on the length of that space's breakdown text. For spaces with 0 epics and audits, `counts` is empty and the `⚠` marker immediately follows `▸0`, while for spaces with multiple epics and audits, `⚠` appears 20+ columns further to the right. This defeats the core acceptance criterion of task `6g2zqyra2s6h` ("One aligned row per space, columns measured across the visible set so the bar, percentage, counts, and date columns line up").

**Recommendation:** Compute `countsW := max(ansi.StringWidth(atlasCountsLine(...)))` across all spaces during the column measurement pass and pad `counts` to `countsW`, or position the variable-length breakdown at the end of the row after the fixed columnar fields (`attention`, `date`).

**Fixed 2026-08-23.** Reproduced at 18 columns of shear between a space rendering
`12 epics · 7 findings` and one rendering `1 epic`; the screenshots that passed review
happened to have equal-width breakdowns. The breakdown is now measured across the set and
padded like every other column. Pinned by
`TestAtlasSpaceRowsKeepColumnsAlignedAcrossUnequalBreakdowns`, which measures **display
columns** rather than byte offsets — a first attempt using `strings.Index` reported phantom
shear, because rows carry different numbers of multi-byte `·` separators.

---

#### M2. Empty work view prompt directs users to press `[/]` for spaces, which exits the atlas  · **Status:** fixed

**File:** `internal/tui/atlas.go:784-786` | **Component:** tui/atlas
**Effort:** XS · **Urgency:** soon

When the work view has no in-progress tasks (`len(a.work) == 0`), `workView()` renders the prompt:
`"Start something with tskflwctl task start <slug>, or press [/] for spaces."`

Pressing `[` or `]` from the atlas activates `PrevTab`/`NextTab`, which navigates out of the atlas to the workspace dashboard or last entity tab. The key to cycle between the atlas's own views (`work` ⇄ `spaces`) is `v` (`keys.View`). The prompt instructs the user to leave the atlas rather than switch views.

**Recommendation:** Update the empty state prompt to reference `keys.View.Help().Key` (`"or press v for spaces"`).

**Fixed 2026-08-23.** A stale string: it predated the key moving from `[`/`]` to `v` and was
missed because no test asserted the empty state named a working key. Now derived from
`keys.View` and pinned by `TestAtlasEmptyWorkViewNamesTheRealViewKey`, so a future rebind
carries it.

---

#### L1. Narrow terminal degradation relies on line-level string truncation instead of structural column dropping  · **Status:** fixed

**File:** `internal/tui/atlas.go:561,840-845`, `planning/tasks/6g2zqyra2s6h-compose-the-atlas-spaces-view-as-a-comparable-table.md:61-62` | **Component:** tui/atlas
**Effort:** S · **Urgency:** eventually

Task `6g2zqyra2s6h` acceptance criterion 8 requires: "Degrades on narrow terminals by dropping the least load-bearing columns rather than overflowing or wrapping." In implementation, `atlas.view()` passes full rows to `truncateAll(lines, maxW)`, which truncates with a trailing `…`. On viewports <= 80 columns, dates, attention badges, and count breakdowns are chopped mid-word or mid-number instead of dropping lower-priority columns cleanly.

**Recommendation:** Implement progressive column inclusion in `spaceRows()` based on `maxW`, or explicitly record in `6g2zqyra2s6h` implementation notes that column pruning was deferred in favor of hard truncation.

**Reopened and fixed the same day — the deferral was wrong.** What follows was the original
reasoning; it was judged against the demo fixture, whose spaces are called `kitchen` and
`bike-workshop`. Against the real registry (`desirelines-planning`, 76-character slugs) the
work view needed **138 columns** on a 92-column terminal, so everything after the epic
column — priority, the staleness age that is the entire point of the view, and the
description — never rendered at all. The spaces table lands at exactly 92 and shears at 80.

Both now FIT rather than truncate: columns are dropped whole in a declared sacrifice order,
with slug and age always surviving in the work view and the bar and rollup always surviving
in the spaces table. The slug is capped at a share of the terminal first so one long name
cannot starve the columns beside it, then grows back into whatever is spare. Pinned by
`TestAtlasWorkRowsFitColumnsRatherThanShearingThem` and
`TestAtlasSpaceRowsShedContextColumnsBeforeShearing`, both asserting at several widths.

*Original note:* Confirmed: rows are truncated at the right edge rather
than shedding whole columns, so a narrow terminal can cut mid-cell (`1mo a…`). Two things
make this a real but low finding. Truncation from the right does drop the rightmost —
least load-bearing — columns first, so the *ordering* is already correct; what is missing
is cutting at column boundaries. And the fix wants a small column-spec abstraction, which
is more machinery than the current two-view atlas justifies. Left open with the task
below, and criterion 8 of `6g2zqyra2s6h` left unchecked so the gap is visible in the
planning data rather than only here.

---

#### L2. Spaces table tests assert substring presence instead of column alignment  · **Status:** fixed

**File:** `internal/tui/atlas_test.go:866-884` | **Component:** tui/test
**Effort:** XS · **Urgency:** eventually

`TestAtlasSpacesTableRendersComparableColumns` asserts only substring presence (`strings.Contains(view, "▸")`, `strings.Contains(view, "%")`). It does not verify that columns (such as `⚠` or date cells) line up at identical character offsets across multiple rows with unequal breakdown lengths. Consequently, the column shearing bug in `spaceRows()` was not caught by the test suite.

**Recommendation:** Add assertions in `atlas_test.go` checking that rune column indices of `⚠` and date tokens match across all rendered table lines.

**Fixed 2026-08-23** as part of M1. This finding correctly identified the coverage gap that
let M1 ship: the table tests asserted substring presence and row count, never column
offsets. The new alignment test asserts them directly.

---

#### L3. Acceptance criteria checkboxes in completed tasks `6g2nnkffgyeg` and `6g2zqyra2s6h` remain unchecked  · **Status:** fixed

**File:** `planning/tasks/6g2nnkffgyeg-render-the-cross-space-in-progress-rail-in-the-atlas.md:37-57`, `planning/tasks/6g2zqyra2s6h-compose-the-atlas-spaces-view-as-a-comparable-table.md:41-65` | **Component:** planning
**Effort:** XS · **Urgency:** eventually

Both task `6g2nnkffgyeg` (work view) and task `6g2zqyra2s6h` (spaces table) have `status: completed` in frontmatter, but all acceptance criteria checkboxes under `## Acceptance criteria` in both markdown bodies remain empty (`- [ ]`). In self-hosted planning repositories, checking off shipped criteria confirms that each item was reviewed and delivered.

**Recommendation:** Update the completed acceptance criteria in `6g2nnkffgyeg` and `6g2zqyra2s6h` to `- [x]` and note any deferred criteria in the implementation notes.

**Fixed 2026-08-23** using `tskflwctl task ac --check`, which exists precisely for this and
had not been used. Ticking per-criterion rather than wholesale is what surfaced L1:
criterion 8 of `6g2zqyra2s6h` ("degrades by dropping the least load-bearing columns") would
not tick honestly, because it was genuinely unmet. It was left unchecked until the column
fitting landed, then checked — so the planning data never claimed a criterion the code did
not meet.

> **Resolution, 2026-08-23.** Every finding was independently reproduced before being acted
> on; none was taken on the reviewer's word. **All seven are fixed**, each with a
> regression test. L1 was initially deferred as "eventually" and that call was wrong: it had
> been judged against the demo fixture's short space names, and real-registry widths showed
> it was acute. It was reopened and fixed the same day. Every proposed candidate task was
> small enough to fix directly rather than track.
>
> Also surfaced while running the gate, unrelated to this branch: `FORCE_COLOR=3` in the
> developer's environment makes `TestColor_DefaultIsPlainForNonTTY` and
> `TestAuditClose_ChangesBucketInPlace` fail. Confirmed pre-existing by reproducing at tag
> `v0.17.0`; `env -u FORCE_COLOR go test ./...` passes. The tool honouring FORCE_COLOR is
> arguably correct, so the gap is that the suite is not hermetic against it. Worth a task
> if it bites anyone again.

## Candidate tasks

<!-- Mirror each finding: ✅ done · ⚠️ partial · ⏳ open · ⛔ won't do -->

- ✅ ~~`tskflwctl task new "Fix pendingJump lifecycle leak and stale workspace cache on atlas work jump" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags tui,atlas,concurrency` — Clear pendingJump across all atlas exits/switches and reload cached workspace surfaces on work-view jump.~~ — fixed directly (H1, H2); regression tests added rather than a task.
- ✅ ~~`tskflwctl task new "Pad atlas counts breakdown column to align attention and date cells" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags tui,atlas,ux` — Compute maximum width for atlasCountsLine across spaces so ⚠ attention and date columns align across rows.~~ — fixed directly (M1).
- ✅ ~~`tskflwctl task new "Fix empty work view keybinding prompt and verify column alignment in tests" --epic 29-multi-space-planning-a-home-registry-and-the-atlas --tags tui,atlas,ux,test` — Fix empty state prompt from [/] to v and add column alignment assertions in atlas_test.go.~~ — fixed directly (M2, L2).

- ✅ ~~`tskflwctl task new "Drop atlas table columns structurally on narrow terminals"`~~ — L1 fixed directly once real-registry widths showed it was acute, not eventual.
