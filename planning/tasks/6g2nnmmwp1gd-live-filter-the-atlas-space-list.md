---
schema: 1
id: 6g2nnmmwp1gd
status: ready-to-start
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: Wire the `/` filter key on the atlas screen so a large registry is navigable by label, id, or path instead of scrolling.
effort: S
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, atlas, ux]
created: "2026-08-22"
updated_at: "2026-10-04"
audit_sources: [2026-10-04-weekly-task-sweep]
audited: "2026-10-04"
---
# Live-filter the atlas space list

## Objective

`/` opens live list filtering on every entity tab. `handleAtlasKey` leaves it unbound, so
a registry with many repos and worktrees is navigable only by `j`/`k` scrolling or by
cycling `o` order. The atlas is the one screen whose row count grows with the user's whole
machine rather than with one repo.

## Acceptance criteria

- [ ] `/` on the atlas opens a live filter over the cards, matching space id, entry-point
  label, and registry path.
- [ ] `esc` clears the filter and restores the previously selected logical space; `enter`
  on a filtered list opens the selection through the ordinary open path.
- [ ] The filter composes with `o`/`O` ordering without losing the selected space.
- [ ] The `?` help and the atlas footer name the key, consistent with the entity tabs.
- [ ] Tests cover filter → select → open and filter → clear → cursor restore.

## Out of scope

- Fuzzy/substring mode toggling (`F`) unless it falls out of reusing `internal/listfilter`.
- Persisting the filter across sessions.

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Audit finding M4: [2026-08-22-multi-workspace-atlas](../audits/6g2k3qye4qma-2026-08-22-multi-workspace-atlas.md)

## Scope note (2026-08-23)

The atlas is becoming two cycled views. This task is the **spaces** view's filter — match
against space id, entry-point label, and registry path, as written above.

Whether the work view needs its own filter is a separate question and deliberately not
answered here: its rows are tasks, so a filter there would match task slugs and
descriptions and is closer to the entity tabs' existing `/` than to this one. Decide it
after living with the work view.

Order of work: this can land before or after the tile grid — it is cursor and matching
logic, not layout — but if it lands after, the filter's chrome has to fit whatever the
tile header becomes.

## Sweep verification (2026-10-04)

Premise verified still accurate. `handleAtlasKey` (`internal/tui/atlas.go:237`)
binds `NextTab`, `PrevTab`, `View`, `j`/`k`, `h`/`l`, `Sort`/`SortRev`, `enter`,
`Atlas`/`Back`, `Command`, `Palette`, `Help`, `Refresh`, `Quit` — and still **no
filter key**. The Objective's reading of why holds: the atlas renders its own cards
rather than delegating to a bubbles list model, so it does not inherit the list's
built-in `/`. `internal/tui/help.go:39` makes that explicit for the entity tabs —
`{"/", "filter the list (label, id, metadata)"}` carries the comment *"the list's
own filter (no keyMap binding)"* — and `atlasHelpSection` (`help.go:74+`) correctly
omits `/` today, which is the gap the fourth criterion describes.

`internal/listfilter` exists and is unchanged since 2026-09-11, so the Out-of-scope
note about `F` falling out of reuse still stands (`keys.FilterMode` is bound to `F`
on the entity tabs, `keys.go:30`).

**The 2026-08-23 scope note's premise has since shipped.** It was written while "the
atlas is becoming two cycled views" was prospective; `keys.View` (`v`) is now bound
in `handleAtlasKey`, and the handler branches on a `work` boolean (the `h`/`l` cases
are guarded `!work`). The work view is real, and `help.go:187-189` documents both
views. So the note's deferred question — whether the work view needs its own filter,
"decide it after living with the work view" — is now answerable rather than
speculative. Left for the human: deciding it is a scope call, not a verification.

2026-10-04: automated weekly sweep — `/` confirmed still unbound in `handleAtlasKey`;
noted that the two-view atlas the scope note anticipated has shipped, making its
deferred work-view filter question now answerable.
