---
schema: 1
id: 6g5rxq17px59
status: in-progress
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Separate durable row/navigation identity from display slugs so duplicate-slug tasks, audits, research, and Threads cannot collide across reloads or workspace sessions.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, architecture, identity, correctness]
created: "2026-09-01"
updated_at: "2026-09-02"
started_at: "2026-09-02"
---

# Make TUI entity navigation use stable identities

## Objective

Make the TUI's internal row, selection, reload, detail, back-stack, and workspace-session identity
match the stable IDs that persistence treats as canonical. Slugs remain the human-facing label and
filter vocabulary, but duplicate slugs must not make two distinct records select, restore, or open
as though they were one entity when Threads join the registry.

## Current constraint

`entityItem.id()` currently returns `Task.Slug`, `Audit.Slug`, and `Research.Slug`, even though those
documents have stable IDs and their stores deliberately permit duplicate slugs. The same value is
used for cursor restoration, selection checks, detail reads, navigation history, palette jumps, and
saved workspace sessions. Adding Thread rows with that convention would copy an existing ambiguous
identity bug into a new first-class entity.

## Scope

- Separate durable row/navigation identity from display/filter text in the entity-item contract;
  avoid overloading one `id()` method with both a canonical key and a human label.
- Use each entity's canonical stable reference for service calls, selection restore, stale-result
  guards, navigation/palette targets, the back stack, and per-workspace session state. Epics retain
  their existing canonical ID behavior.
- Keep slugs and descriptions in visible rows, detail titles, fuzzy filtering, and user-facing
  messages; include a short stable-ID disambiguator only where two rows in one loaded result would
  otherwise be indistinguishable.
- Define explicit behavior for malformed or drifting IDs that resilient list reads retain. Do not
  silently fall back to a colliding slug; preserve repair visibility and use the store's canonical
  resolvable identity when one is available.
- Make the entity registry contract safe for the later Thread item without adding the Thread tab in
  this task.

## Acceptance criteria

- [x] Two tasks with the same slug but distinct stable IDs remain independently selectable,
  detail-loadable, filterable, and cursor-restorable across manual and watcher reloads.
- [x] Equivalent duplicate-slug coverage exists for every stable-ID TUI entity whose store permits
  the state, and the future Thread item contract has one documented canonical-key path.
- [x] Navigation history, command-palette jumps, stale async result guards, and saved workspace
  sessions carry canonical identity rather than display slug.
- [x] Slug-based display and filtering remain readable; duplicate visible labels gain deterministic
  disambiguation without exposing IDs everywhere by default.
- [x] Missing/drifting-ID fixtures fail visibly or use a documented canonical filename identity;
  they never collapse onto another row or open an arbitrary duplicate.
- [x] Existing unique-slug navigation behavior and TUI golden output remain unchanged except where
  identity disambiguation is intentionally exercised.

## Stress tests

- Duplicate slugs in one list, duplicate slug hidden by a filter, reload with the second duplicate
  selected, detail loads completing out of order, back navigation across duplicates, workspace
  switch/restore, malformed frontmatter ID, and filename/frontmatter ID drift.

## Implementation progress (2026-09-02)

- The entity registry now exposes an `entityRef` with separate canonical key and human label.
  Cursor restoration, detail reads/stale guards, sorting/filter changes, palette/dashboard/follow
  navigation, the back stack, atlas landings, workspace sessions, lifecycle actions, and inline
  edits carry the canonical key; titles, breadcrumbs, yank, prompts, and mutation flashes keep the
  label.
- Task, audit, research, and future Thread records expose `CanonicalID()`: an adapter-provided
  filename identity wins when present so missing/drifting frontmatter cannot redirect a read, while
  portable adapters without filename semantics fall back to their semantic `ID`. Epics retain their
  existing ID contract. Acute audit-finding dashboard rows now receive that canonical audit identity
  from the core summary instead of re-resolving the visible audit slug.
- Duplicate labels gain a leading shortest-unique stable-ID prefix in entity rows, follow pickers,
  the epic task roster, dashboard work rows, and Atlas work rows; the prefix survives truncation and
  stays stable while filtering or paginating the loaded result. Unique labels keep their previous
  presentation. Stable IDs join the tab and palette filter corpus without replacing
  slug/description/tag search.
- Stress coverage now creates duplicate task, audit, and research slugs with distinct filename IDs,
  including missing and drifting frontmatter IDs. It exercises independent detail reads, out-of-order
  stale results, reload restoration, filtering, palette targets, back navigation, session snapshots,
  stable-key mutation, and collision-expanded short hints. Existing unique-slug and rendering tests
  remain unchanged in behavior.

## Adversarial review closeout (2026-09-02)

Three independent review passes are closed with every finding fixed: [Claude](../audits/6g67v0c4n6sv-2026-09-02-tui-stable-entity-navigation-implementation-claude.md),
[Antigravity](../audits/6g67v0cd1x0h-2026-09-02-tui-stable-entity-navigation-implementation-antigravity.md),
and Antigravity's [deeper systemic pass](../audits/6g6819a6hm1e-2026-09-02-tui-stable-entity-navigation-implementation-antigravity-38.md).
Review hardening made exact stable IDs outrank an unrelated sibling slug in repository, task-graph,
and Thread-snapshot resolution; rejects empty or repeated adapter keys at the generic entity
registry boundary; keeps friendly detail-error titles; pins every lifecycle/edit target to its
canonical key; copies an ID for ambiguous rows; and retries a raced cross-space Atlas landing in
`:all` without changing ordinary reload behavior. Misleading slug-valued `selectedID` test helpers
were removed in favor of explicit label/key assertions.

## Validation (2026-09-02)

The full `go test -race ./...` suite, `go vet ./...`, golangci-lint, module-tidy diff,
generated CLI-doc drift check, planning lint, shell syntax/prompt preflight, and diff hygiene pass.
Each of the three review audits passes `audit lint` after closeout. Repository-wide audit lint still
reports the pre-existing empty `**Resolution:**` placeholders in
`2026-09-01-cli-ergonomics-from-an-agent-session`; this branch does not alter that unrelated audit.

## Out of scope

- Persisting TUI sessions across process restarts, changing fuzzy resolver tiers beyond exact
  stable-ID precedence, renaming documents, adding Thread screens, or replacing user-facing slugs
  with IDs.

## Sequencing

Independent TUI foundation discovered while scoping Thread views. The Thread list/detail slice must
depend on it so the new entity never adopts slug-keyed state.

## Related

- Epic [18-tui-bubble-tea-interactive-planning-browser](../epics/18-tui-bubble-tea-interactive-planning-browser.md)
- Downstream [Thread list and detail views](6g5rwjqr7rt8-add-thread-list-and-detail-views-to-the-tui.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
