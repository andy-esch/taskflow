---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: handleKey's modal guard chain grows per overlay; restore/detail use shared single slots. Add an overlay stack and stamp intent with loadGen.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, architecture]
created: "2026-06-22"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r01yqn
---
# Harden TUI dispatch modal stack and generation-stamped reload

## Objective

Two related TUI-core issues. (1) handleKey is a ~160-line guard chain that grows one
block + one Model bool + one bodyView case per overlay, with ForceQuit duplicated 5x:
introduce a modal/overlay interface (active/handleKey/view) with an ordered stack.
(2) tab.restore is one slot shared by reload + jump, and detailGen isn't tied to a
tab's loadGen: carry the target id + originating loadGen in the load result rather
than a mutable field. (H5 was a targeted patch over the same single-slot smell.)

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M14** (modal stack) and **M6** (generation-stamped restore/detail). Relates to epic 18.

## Acceptance criteria

- [x] New overlays are a registry entry, not edits across handleKey + bodyView + Model.
- [x] A reload overlapping a jump can't clobber the navigation target; detail tied to loadGen.
- [x] just test + just lint green.

## Implementation plan

**Approach.** Two independent refactors that share the "single mutable slot" smell.
(1) **Modal stack:** introduce an `overlay` interface and an ordered slice the
reducer loops, replacing the three `if m.X.active` guard blocks (help/action/follow)
at the top of `handleKey` and the parallel `switch` in `bodyView`. (2)
**Generation-stamped restore:** replace `entityTab.restore string` (written by both
`markReload` and `jumpTo`, consumed asynchronously) with an intent carried *in the
load result* and stamped with the originating `loadGen`, so a reload firing between a
jump's request and its landing can't silently steal the cursor target. Do them as two
PRs; the restore fix (M6) is the higher-value/lower-churn half and H5 was already a
point-patch over the same slot, so consider landing M6 first.

**Steps — modal stack (M14).**
1. Define `type overlay interface { active() bool; handleKey(m *Model, msg tea.KeyMsg)
   (handled bool, cmd tea.Cmd); view(w, h int) string }` in a new
   `internal/tui/overlay.go`. Make `actionMenu`, `followMenu`, and a thin `helpOverlay`
   wrapper satisfy it (their key handlers already exist as `handleActionKey`/
   `handleFollowKey` and the inline help-scroll block in handleKey; move that block
   into a `helpOverlay.handleKey`).
2. Hold them as `m.overlays []overlay` in order of precedence (help, action, follow —
   matching today's 0/0b/0c ordering). In `handleKey`, after the flash-clear, loop:
   `for _, o := range m.overlays { if o.active() { if handled, cmd := o.handleKey(&m,
   msg); handled { return m, cmd } } }`. Each overlay's `handleKey` keeps the
   `keys.ForceQuit → tea.Quit` escape (this removes the 5× ForceQuit duplication noted
   in M14 — it lives once in the loop preamble OR once per overlay; prefer the loop
   preamble: check ForceQuit before the overlay loop and the global switch).
3. In `bodyView`, replace the `switch {case m.showHelp … case m.action.active …}` with
   a loop that asks each active overlay for its `view(m.width-2, m.paneOuterH-2)` and
   composites the topmost active one via `overlay()`; fall through to `base` when none
   active. `helpMaxScroll` moves onto `helpOverlay`.
4. The command bar (`m.cmd`) and the list filter / detail-find captures stay as
   special early returns — they are *input captures*, not floating boxes, so folding
   them into the same interface is optional; note that boundary in the body. A future
   modal (peek/confirm/tag-picker) is then one struct + one append to `m.overlays`.

**Steps — generation-stamped restore (M6).**
5. Add a `restore` field to `listLoadedMsg` (messages.go) — the id to re-select after
   this specific load. `entityTab.reload`/the `loadList` Cmds already snapshot
   `loadGen`; thread the intended restore id into the loader so it rides back on the
   matching-gen message. Drop (or keep transitional) `entityTab.restore`: the consumer
   in `handleListLoaded` already gates on `msg.gen == tab.loadGen`, so reading
   `msg.restore` instead of the mutable `tab.restore` closes the race where a
   `markReload`-driven fs-debounce reload overwrites a `jumpTo`-set target.
6. `markReload` (entity.go) and `jumpTo`/`applyView` (nav.go/model.go) currently set
   `tab.restore` then call `tab.reload`; change them to pass the id *into* `reload`
   (e.g. `reload(svc, restoreID string)`), which stamps it onto the load it fires.
   `handleTabMsg`'s pending-restore-after-refilter path (the filtered-list case where
   `VisibleItems` is empty at SetItems time) still needs a per-tab pending slot — but
   now keyed to the gen that set it, so a newer load supersedes it.
7. Tie detail to the tab's load: when a reload changes the selection, `refreshDetail`
   already bumps `detailGen`; verify the `detailMsg`/`detailErrMsg` stale guard
   (`msg.gen != m.detailGen`) still holds with the new restore flow — the audit notes
   detailGen isn't tied to a tab's loadGen, but the kind+id+gen triple already defeats
   the cross-tab case (see the audit's "Refuted" section), so the concrete fix is just
   ensuring a restore-driven selection change triggers a fresh `loadDetail`.

**Tests.** All message-injection (`internal/tui/model_test.go`/`nav_test.go`):
(a) M6 regression — set a `/filter`, fire a `jumpTo` whose load is in flight, inject a
`reloadMsg`/debounce that calls `markReload`, then land both `listLoadedMsg`s out of
order; assert the cursor lands on the jump target, not the reload's captured id, and
no spurious "not found" flash. (b) Modal stack — assert help/action/follow each
capture keys while active and that `ForceQuit` quits from each; assert `bodyView`
composites the active overlay (substring assertion on `View()`); assert a key falls
through to base routing when no overlay is active. Keep `action_test.go`,
`help_test.go`, `nav_test.go`, `polish_test.go` green.

**Risks / gotchas.** (a) `Model` is value-typed and `Update` returns a new model —
overlay `handleKey` taking `*Model` must be reconciled with that (either operate on a
local copy returned, or have the loop pass `&m` where `m` is the by-value receiver's
copy — the existing handlers already mutate the value copy and return it). Keep the
"value model, mutate-the-copy" idiom; don't introduce shared pointers. (b)
`TestModel_ViewFitsTerminal` (the layout invariant) must still pass — the overlay
compositing math (`m.width-2`, `m.paneOuterH-2`, `lipgloss.Place` + `overlay()`) must
be reproduced exactly when generalized. (c) Update is serial, so these are
message-ordering bugs, not data races — tests must inject messages in the hazardous
order, not rely on timing. (d) Don't regress H5's `movedAway` guard — it reads
`tab.restore`/the moved-id absence; if `restore` moves into the message, re-express the
guard against the new field.

**Done when.** Adding a new overlay is a struct + one `m.overlays` entry (no new
handleKey guard block, no new `bodyView` case, no new `Model` bool); a reload
overlapping a jump preserves the navigation target with no false "not found"; and
`go build ./...`, `go test ./...`, `golangci-lint run ./...` are green.
