---
schema: 1
id: 6g2k3qye4qma
bucket: open
area: multi-workspace-atlas
date: "2026-08-22"
---

# Audit: multi-workspace-atlas — 2026-08-22

Scope: Comprehensive, multi-agent adversarial review of the multi-workspace Atlas and workspace opening feature introduced in commit `6f3f0cc`. Covers `internal/core/workspace.go`, `internal/core/space_overview.go`, `internal/workspacestore/fs.go`, `internal/spacestore/fs.go`, `internal/userconfig/spaces.go`, `internal/spacehealth/diagnose.go`, `internal/cli/ui.go`, `internal/cli/root.go`, `internal/tui/model.go`, `internal/tui/session.go`, `internal/tui/atlas.go`, `internal/tui/watch.go`, `internal/tui/view.go`, `.golangci.yml`, and documentation.

> **Second pass — 2026-08-22.** Every first-pass finding was re-derived from the code and,
> where possible, reproduced. Method: `just build`, `go test -race ./...`,
> `golangci-lint run ./...`, `tskflwctl lint`, `just docs` (all clean, docs already in
> sync), real-binary probes from a directory outside any planning repo, and throwaway
> tests in `internal/tui` to reproduce H2, M5, and L2. Verdicts are recorded per finding
> as **Verified**; H1 is the one finding whose framing is partly rejected. M5–L4 are new.
>
> Overall: the slice meets its task's acceptance criteria. The boundary is clean
> (`go list -deps ./internal/tui` shows no `config`/`store`/`spacestore`, and the new
> adapter is depguarded), the watcher/session lifecycle is correct under review, and
> docs/tests are unusually thorough for a slice this size. What follows is refinement,
> not rework.
>
> **Resolution — 2026-08-22.** H1, H2, M1, M5, L1, L2, L3, and L4 are **fixed on the same
> branch**, each with a regression test. H1 was fixed with a revised behavior the owner
> specified, not the one the first pass recommended (see the finding). M2, M3, M4, and H3
> remain open as tracked tasks on epic 29, which was **un-retired**: the atlas's stated
> payload is still unbuilt, so the epic that owns it is not finished.

## Findings

#### H1. `ui` cannot reach the atlas from outside a planning repo  · **Status:** fixed

**File:** `internal/cli/root.go:209,252-267`, `internal/cli/ui.go:27-44` | **Component:** cli/tui
**Effort:** M · **Urgency:** acute

`ui` does not override the root `PersistentPreRunE`, so `repoPreRun` → `config.Discover`
runs before the TUI starts. **Reproduced 2026-08-22** from a clean directory with five
registered spaces: `tskflwctl ui -C /tmp/clean` exits 1 with
`not a taskflow planning repo (no .tskflwctl.toml or tasks/ found …) — run tskflwctl init`.
The atlas is a home-registry surface, yet it is unreachable from anywhere a space has not
already been chosen.

The minimum fix is not a null-workspace TUI (that would push a `m.svc == nil` mode into
every tab, the dashboard, the footer, and `:config`). Outside a repo, resolve the atlas's
own preferred entry point, open it through `core.WorkspaceService` as the initial
workspace, and land on the atlas.

**Recommendation:** Override `PreRunE` in `internal/cli/ui.go` to forgive
`config.ErrNoConfig` when neither `--space` nor `-C` was given, then seed the initial
workspace from the registry and start with `onAtlas = true`.

**Verified (2nd pass): confirmed for the non-repo half.** The first pass also called
landing on the atlas *inside* a repo with >1 space "inverted". That matched the shipped
research sketch ("**Land only when >1 space is registered**"), so as a *finding* it was
rejected — but the owner then re-decided the behavior, which is the one call an audit
cannot make for itself.

**Fixed 2026-08-22, with a revised rule.** Landing is now decided by **context, not
count**:

- Inside a planning repo (or one named by `--space`/`-C`), `ui` opens that repo's
  Overview. The atlas never interposes itself on someone who already said where they are.
- Outside one, `ui` forgives the ordinary discovery miss, opens the registry's preferred
  entry point as the runtime workspace, and lands on the atlas.

`config.ErrNoConfig` is a new sentinel on exactly the "no planning repo anywhere above
here" error, so `ui` can forgive that while a config that EXISTS but is broken stays
fatal — and so can `--space`/`-C`, which are assertions about *which* tree to open and
must never be answered with a different one. `SpaceRegistryService.PreferredEntry` supplies
the seed through the same direct-over-pointer routing the atlas uses; seeding avoids
teaching every tab, the dashboard, and `:config` a null-workspace mode. `WithAtlas` no
longer chooses the landing screen at all — the new `WithAtlasLanding` does, which is why
the space count no longer appears in the decision. Verified with the real binary: from a
non-repo cwd, `ui` now reaches the interactive-terminal gate instead of exiting 1 on
discovery. Tests: `TestUIStartup*`, `TestUIDoesNotSubstituteASpaceForABrokenExplicitSelection`,
`TestAtlasInRepoStartupOpensTheRepoAndKeepsTheAtlasOneKeyAway`,
`TestAtlasLandingHoldsEvenWithOneRegisteredSpace`,
`TestDiscover_NoPlanningRepoIsIdentifiableAndKeepsItsMessage`.

Follow-on from the same change: the seeded space keeps the browser's surfaces live but
was never chosen, so naming it in the top-rail chip would claim a context the user never
entered. The chip reads `[atlas]` — this project's word for the whole set rather than one
of it — until they leave the atlas by any route, after which it names the space for the
rest of the session. In-repo launches are unaffected: standing in a repo IS the choice.
Tests: `TestAtlasChipNamesTheAtlasUntilASpaceIsChosen`,
`TestAtlasChipKeepsTheRepoNameWhenLaunchedInsideOne`.

---

#### H2. Atlas status and error text lives inside the scrollable body, not the pinned footer  · **Status:** fixed

**File:** `internal/tui/atlas.go:378-391` | **Component:** tui
**Effort:** S · **Urgency:** acute

`atlas.view` appends `opening selected space…`, `✘ <openErr>`, and
`⚠ refresh failed: …` to the end of `lines`, then hands the whole slice — header included
— to `scrollTo(lines, focusRow, maxH)`. Everywhere else in this TUI, transient results go
to `m.flash`, which `footer()` renders on a pinned row (`view.go:340`); the atlas is the
only surface that puts them where they can scroll away.

**Reproduced 2026-08-22** against this machine's actual registry shape (4 logical spaces;
1, 1, 3, and 1 entry points): the atlas renders **25 rows**, while an 80×24 terminal
leaves roughly 20 body rows after the tab strip, footer, and pane frame. With the cursor
on a lower card the `atlas · 4 space(s) · order …` header scrolls off; with the cursor on
an upper card the trailing error row is clipped, so pressing `⏎` on a space that fails to
open produces no visible change at all. Both cases are live today, before anyone registers
a fifth space.

**Recommendation:** Pin the header and the status row outside the slice passed to
`scrollTo`, so only the cards scroll.

**Verified (2nd pass): confirmed,** with the diagnosis sharpened. The first pass framed
this as a viewport bug; the underlying cause is that the atlas bypasses the TUI's
established pinned-footer convention. Note the card body itself is fine — it mirrors
`dashboard.view`'s deliberate scroll behavior, and a broken entry's own `Detail`/`Remedy`
lines render inline next to the cursor, so the *entry-specific* explanation is visible.
The clipped text is the screen-level status.

**Fixed 2026-08-22.** `atlas.view` now composes a pinned header, a scrolled card body
sized to `maxH - header - status`, and a pinned status row. Routing the text through
`m.flash` — the first pass's recommendation — was tried and rejected: `flash` is a
one-shot cleared by the next keypress, which is right for "moved task → done" but wrong
for `opening selected space…` (a live status) and for an open error that should persist
until the cursor moves, which is exactly what `move`/`moveEntry` already implement.
Pinned by `TestAtlasPinsHeaderAndStatusAroundTheScrollingCards`, which asserts the header,
the status row, and the focused card all survive a 6-space registry in a 12-row body at
three cursor positions.

---

#### H3. Broken symlink `spaces.toml` silently masks registry data as empty rather than reporting error  · **Status:** open

**File:** `internal/userconfig/spaces.go:100-119` | **Component:** userconfig
**Effort:** XS · **Urgency:** acute

`internal/userconfig/userconfig.go:167-177` deliberately separates "no config file" from
"a symlink whose target is gone", re-probing with `os.Lstat` on `ErrNotExist` and
returning an actionable `broken symlink -> <target>` error — the documented reasoning is
that this audience symlinks its config out of a dotfiles repo that may not be mounted.
`spaces.go:readRegistry` uses a bare `os.ReadFile` and collapses every `ENOENT` to
`initialSpacesText, nil, nil`, i.e. "no spaces registered".

**Recommendation:** Port the `os.Lstat` probe from `userconfig.go:167-177` into
`readRegistry`.

**Verified (2nd pass): confirmed, with scope corrected.** This is pre-existing — the file
is untouched by `6f3f0cc` — so it is out of scope for this branch and now carries its own
task. The atlas is what raises its cost: the failure mode used to be a short `space list`;
it is now a TUI that says "No spaces registered. Run `tskflwctl space add <path>`" and
invites the user to re-register spaces they still have, after which `space add` would
write a fresh registry over the dangling link.

---

#### M1. Path-string identity drops the green `current` badge under symlinks or case differences  · **Status:** fixed

**File:** `internal/tui/atlas.go:405,441,472-474`, `internal/tui/session.go:69-74` | **Component:** tui/session
**Effort:** XS · **Urgency:** soon

`workspaceKey`/`filepathKey` compare `filepath.Clean` of two paths that are produced
differently: `workspace.Checkout` comes from `config.Discover`, which resolves symlinks
via `evalOr` (`config.go:99`), while `entry.Checkout` is `userconfig.ExpandTilde(space.Path)`
(`spacestore/fs.go:88`) — tilde-expanded but never symlink-evaluated. A registry row
pointing through a symlink, or spelled with different case on a case-insensitive APFS
volume, silently loses the `current` marker on the space the user is standing in.

**Recommendation:** Prefer durable identity: match `entry.ID == current.SpaceID` when
`SpaceID` is set, and fall back to comparing symlink-evaluated paths for the launch
workspace (whose `SpaceID` is empty unless `--space` was passed).

**Verified (2nd pass): confirmed as latent.** Checked this machine: none of the four
registered checkouts is a symlink (`pwd -P` matches the registry path), so the badge works
today. The defect is real but has no current reproduction, hence *soon*, not *acute*.

**Fixed 2026-08-22, on both sides.** Rendering now asks two different questions through
`isCurrentEntry` (is this exact checkout the open one — durable `SpaceID` first, path key
as the fallback) and `isCurrentSpace` (does the open workspace belong to this logical
space — planning id first), which is the right distinction for a card whose identity may
have three registered ways in. Underneath, `spacehealth.SpaceProblem` now records the
directory discovery actually resolved and `spacestore` uses it for `SpaceEntryPoint.Checkout`,
so both sides of any remaining path comparison come from `config.Discover` and are
symlink-evaluated. The registry spelling stays on `Path` for output; an entry too broken
to discover keeps its tilde-expanded path, which is all that is known about it. No
filesystem access was added to the render path. Tests:
`TestAtlasCurrentBadgeFollowsIdentityNotPathSpelling`,
`TestFSRegistryAdapter_CheckoutIsTheResolvedPathNotTheRegistrySpelling`,
`TestFSRegistryAdapter_UndiscoverableEntryKeepsItsRegisteredPath`.

---

#### M2. Atlas cards omit branch and worktree badges that `DescribeCheckout` already supplies  · **Status:** open

**File:** `internal/spacehealth/diagnose.go:119-127`, `internal/tui/atlas.go:439-444` | **Component:** spacehealth/tui
**Effort:** S · **Urgency:** soon

[`internal/config/worktree.go:134`](../../internal/config/worktree.go) exposes
`DescribeCheckout(dir)` (branch + `IsWorktree`), which `internal/cli/workspace.go:67`
already feeds into `wire.WorkspaceJSON`. `spacehealth.DiagnoseSpace` and `renderAtlasCard`
ignore it, so two registered worktrees of one repo render identically as
`● id  direct  ~/path` — the multi-entry-point case that motivated grouping in the first
place is the one the card cannot describe.

**Recommendation:** Carry `Branch`/`IsWorktree` through `spacehealth.SpaceProblem` into
`core.SpaceEntryPoint` and render them in the entry rows.

**Verified (2nd pass): confirmed.** Feature gap, not a defect; no worktrees are currently
registered on this machine.

---

#### M3. The cross-space in-progress rail — the sketch's stated payload — is discarded  · **Status:** open

**File:** `internal/tui/atlas.go:344-392`, `internal/core/space_overview.go:36-40` | **Component:** tui
**Effort:** M · **Urgency:** soon

`SpaceOverview.InProgress` is populated by `SpaceOverviewService.Overview` and rendered by
`status --all` ("In progress across spaces (3)"), but `atlas.view` uses only
`len(summary.InProgress)` as a count string and drops the rows. The research sketch names
this rail the atlas's actual payload — "Cards orient; the rail answers the question that
genuinely can't be asked today" — so the shipped atlas has the orienting half without the
answering half.

**Recommendation:** Render `overview.InProgress` as a bounded section beneath the cards,
sharing the body budget without making the card list unnavigable.

**Verified (2nd pass): confirmed, and it is the highest-value gap here.** See also L4:
commit `6f3f0cc` deleted the epic's out-of-scope line that had explicitly excluded the
rail, so it was left neither built nor excluded.

---

#### M4. No live filter (`/`) on the atlas  · **Status:** open

**File:** `internal/tui/atlas.go:123-154` | **Component:** tui
**Effort:** S · **Urgency:** soon

`/` opens live list filtering on every entity tab; `handleAtlasKey` leaves it unbound. The
atlas is the one screen whose row count scales with the user's whole machine rather than
with one repo, so it is the screen that most needs the key.

**Recommendation:** Wire `/` on the atlas to filter cards by space id, entry-point label,
and registry path, composing with `o`/`O` ordering.

**Verified (2nd pass): confirmed.** Also unbound on the atlas: `[`/`]` (which move between
the dashboard and tabs elsewhere) and `g`/`G`. Worth folding into the same pass.

---

#### M5. Visiting the atlas and returning resets focus and zoom for the space you never left  · **Status:** fixed

**File:** `internal/tui/atlas.go:156-169`, `internal/tui/atlas.go:140-141` | **Component:** tui
**Effort:** XS · **Urgency:** soon

`enterAtlas` sets `m.focus = focusList` and calls `m.unzoom()`, and the return path
(`keys.Atlas` / `keys.Back` → `m.onAtlas = false`) restores neither. `spaceSession`
records `focus` and `zoom` precisely because they are worth preserving across a switch, so
a *round trip that never switches* is the one path that loses them — while leaving the
stale detail body on screen.

**Reproduced 2026-08-22:** from `focus=focusDetail, zoom=true`, pressing `a` then `a`
returns with `focus=focusList, zoom=false`.

**Recommendation:** Snapshot focus/zoom on `enterAtlas` and restore them on return, or
route the return through the same session-restore path a switch uses.

**Verified (2nd pass): new, confirmed by test.**

**Fixed 2026-08-22.** `enterAtlas` snapshots focus/zoom into `atlasResume` and the new
`exitAtlas` restores them; `activateWorkspace` discards the snapshot, because a switch IS
navigation away and the incoming space's cached session owns its own state. Tests:
`TestAtlasRoundTripRestoresDetailFocusAndZoom` and
`TestAtlasSwitchDiscardsTheResumeSnapshotOfTheSpaceLeft`.

---

#### L1. Missing `Checkout` validation and nil-`Layout` guard at the workspace boundary  · **Status:** fixed

**File:** `internal/core/workspace.go:62-75`, `internal/tui/tui.go:42` | **Component:** core/tui
**Effort:** XS · **Urgency:** eventually

`WorkspaceService.Open` validates `Store`, `Layout`, and `PlanningRoot` but not
`Checkout`; an adapter returning an empty checkout makes `m.configStart` empty
(`session.go:98`) and `:config` falls back to cwd. `tui.Run` dereferences
`workspace.Layout.WatchPaths()` with no nil guard, so a partially-built `core.Workspace`
panics at startup rather than degrading to `watchOff` the way a failed watcher does.

**Recommendation:** Reject `source.Checkout == ""` in `Open` and guard
`workspace.Layout != nil` in `Run`.

**Verified (2nd pass): confirmed, with one sub-item dropped and one added.** The first
pass also asked for `strings.TrimSpace(start) != ""` in `workspacestore.FS`; that is
redundant — `WorkspaceService.Open` already rejects an empty start at `workspace.go:63`,
before the adapter is reached, and the adapter is not a public entry point. Added: in
`Open`, the `s == nil || s.store == nil` check sits *after* the request validation, so a
nil service with an empty start reports the wrong cause.

**Fixed 2026-08-22.** `Open` rejects an empty `Checkout` and checks its own capability
before the request; `tui.Run` treats a nil `Layout` as live-reload-unavailable — the
degradation the footer already reports honestly — instead of panicking. Tests:
`TestWorkspaceService_OpenRejectsAnEmptyCheckout`,
`TestWorkspaceService_OpenReportsAnUnavailableOpenerBeforeTheRequest`.

---

#### L2. `scopeSession`'s runtime-message guard misses pointer types and `tea.Sequence`  · **Status:** fixed

**File:** `internal/tui/session.go:36-49` | **Component:** tui
**Effort:** XS · **Urgency:** eventually

`scopeSession` keeps Bubble Tea's runtime-control messages visible by comparing
`reflect.TypeOf(msg).PkgPath()` against `"charm.land/bubbletea/v2"`. Two gaps:

1. **Pointer types report an empty `PkgPath`.** Verified:
   `reflect.TypeOf(tea.QuitMsg{}).PkgPath()` is `"charm.land/bubbletea/v2"` but
   `reflect.TypeOf(&tea.QuitMsg{}).PkgPath()` is `""`, and a pointer-typed runtime message
   is duly swallowed into `sessionMsg`.
2. **`tea.BatchMsg` is unwrapped recursively; `tea.Sequence`'s payload is not.**
   `sequenceMsg` is an unexported `[]Cmd`, so it passes the package check and its inner
   commands run entirely unscoped — silently defeating the workspace-generation guard.

Neither is reachable today: bubbletea v2.0.7 emits value-typed messages from commands, and
this package uses no `tea.Sequence` (only `tea.Batch`/`tea.Tick`). This is a trap for the
next person, not a live bug.

**Recommendation:** Unwrap `reflect.Ptr` before the package check, and either handle the
sequence payload or record the `tea.Sequence` prohibition in the comment with a guard test.

**Verified (2nd pass): new, item 1 confirmed by test, item 2 by inspection.**

**Fixed 2026-08-22.** `scopeSession` dereferences pointer types before the package check.
`tea.Sequence` cannot be unwrapped from outside bubbletea, so the prohibition is written
down and enforced instead: `TestSessionScopingHasNoSequenceCommandsToLeak` parses every
non-test file in the package and fails on a `tea.Sequence` call. Pointer handling is pinned
by `TestSessionScopingKeepsPointerShapedRuntimeMessagesVisible`, which also asserts a
pointer-typed *application* message is still stamped.

---

#### L3. `WithAtlasTheme` bakes the dark palette at option time  · **Status:** fixed

**File:** `internal/tui/model.go:131-140`, `internal/tui/tui.go:37-41` | **Component:** tui
**Effort:** XS · **Urgency:** eventually

`WithAtlasTheme` stores the theme *and* eagerly assigns `*m.atlasSt = newStyles(theme.Dark)`,
unconditionally choosing the dark variant. `Run` then resolves the terminal background and
overwrites it with `newStyles(atlasTheme.For(dark))`, so the eager assignment is dead work
in the real path — but wrong for any consumer that builds a Model without `Run` (which is
what the tests do). The option should carry the theme and let the single render-path
resolution own the variant, matching how `m.st` is handled.

**Recommendation:** Drop the eager assignment; keep `theme.For(dark)` in `Run` as the one
place the variant is chosen.

**Verified (2nd pass): new, by inspection.**

**Fixed 2026-08-22.** The option carries the theme only; `Run` remains the single place
the light/dark variant is chosen, exactly as it already was for `m.st`.

---

#### L4. Epic 29 was retired with the atlas's stated payload neither built nor excluded  · **Status:** fixed

**File:** `planning/epics/29-multi-space-planning-a-home-registry-and-the-atlas.md` | **Component:** planning
**Effort:** XS · **Urgency:** soon

Commit `6f3f0cc` flipped epic 29 to `retired` **and** deleted its out-of-scope line —
"The atlas's visual design — card layout, accent derivation, the cross-space rail.
Sketched, not specified." — in the same change. The rail (M3) and the per-space accent
(`core.SpaceEntryPoint.Accent` exists and is unused by the atlas) were therefore removed
from the exclusion list without being implemented and without landing anywhere else, which
is the one way this repo's planning can actually lose work.

**Recommendation:** Track the deferred atlas surface on an active epic rather than in a
retired one.

**Verified (2nd pass): new; fixed 2026-08-22.** Epic 29 is **active** again, with a
step 6 ("finish the atlas surface") and the four remaining findings as tasks under it.
Its Out of scope now names the per-space accent explicitly as a design commitment not
made, rather than leaving it deleted-and-unbuilt. The epic also records the revised
landing rule, and the sketch's superseded bullet is struck through in the research doc
so the old rule is not re-derived from it.

## Candidate tasks

**Fixed on the `feat/atlas-tui` branch** (H1, H2, M1, M5, L1, L2, L3, L4) — each has a
regression test; see the finding for what changed and why.

**Open, tracked on epic 29** (un-retired 2026-08-22, since the atlas surface is
unfinished):

- ⏳ `6g2nnkffgyeg` [Render the cross-space in-progress rail in the atlas](../tasks/6g2nnkffgyeg-render-the-cross-space-in-progress-rail-in-the-atlas.md) — M3, the highest-value gap.
- ⏳ `6g2nnkfk1em1` [Show branch and worktree badges on atlas entry points](../tasks/6g2nnkfk1em1-show-branch-and-worktree-badges-on-atlas-entry-points.md) — M2.
- ⏳ `6g2nnmmwp1gd` [Live-filter the atlas space list](../tasks/6g2nnmmwp1gd-live-filter-the-atlas-space-list.md) — M4.
- ⏳ `6g2nnp5tkaaj` [Detect a broken symlink in the spaces registry loader](../tasks/6g2nnp5tkaaj-detect-a-broken-symlink-in-the-spaces-registry-loader.md) — H3; pre-existing, in a file this branch does not touch.
