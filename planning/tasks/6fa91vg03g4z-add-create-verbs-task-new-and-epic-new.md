---
status: completed
epic: 17-pm-go-cli
description: Add tskflwctl task new and epic new create verbs with validated frontmatter and a handoff-ready body scaffold
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [pm-tooling, go, cli]
created: 2026-06-08
started_at: 2026-06-08
updated_at: 2026-06-08
completed_at: 2026-06-08
id: 6fa91vg03g4z
---

# Add create verbs (task new and epic new)

## Objective

The keystone gap for a bare-bones release: `tskflwctl` can read/update/move/lint
tasks but can't **create** them — you still fall back to Python `pm new`. Add
`task new` (and `epic new`) so the Go CLI is self-sufficient for the daily loop.

## Decisions (2026-06-08)

- **Body template: handoff-ready** — Objective / Acceptance criteria (checkbox)
  / Out of scope / Related `[[epic]]`. (Not the minimal pm "TBD" body.)
- **Scope:** both create verbs (`task new` + `epic new`). Epics created rarely,
  but shipping both makes the tool fully pm-independent for the core loop.

## Implementation Plan

- [x] `domain.Slugify(title)` (matches pm's rules — verified against real pm
      slugs) + unit tests.
- [x] Store: `CreateTask` / `CreateEpic` — `buildFile` serializes frontmatter
      in **canonical order** via a fresh `yaml.Node` (reuses `valueNode`, so a
      colon-in-description is quoted at the source), atomic write, **refuses to
      clobber**; `CreateEpic` auto-numbers `NN-slug`. Extracted shared
      `assembleFile`. Added to the `core.Store` interface + `fakeStore`.
- [x] Core: `Service.NewTask`/`NewEpic` — validate epic exists, tier/autonomy/
      priority/description via `domain.ValidateField`, stamp `created`, build the
      handoff body, derive slug, delegate. Default `ready-to-start` (`--next` →
      `next-up`); epic auto-numbered, description required.
- [x] CLI: `task new <title>` (`--epic`[req, completes] `--description`
      `--effort` `--tier` `--priority` `--autonomy` `--tags` `--next` `--body`)
      and `epic new <title>` (`--description` `--status` `--priority` `--tags`).
      Human (`created <path>`) + `--json`; exit 11 on validation/clobber.
- [x] Tests: slugify units; store create (order, quoting, clobber, auto-number);
      core validation (unknown epic → nothing written, `--next`); CLI happy path
      round-trips `show`+`lint`, `--next`, unknown-epic exit 11, clobber, epic
      new + missing-description.

## Acceptance

- [x] `tskflwctl task new "Title" --epic X` writes a lint-clean, handoff-ready
      task in `ready-to-start/` (`--next` → `next-up/`); re-running errors
      instead of clobbering.
- [x] `tskflwctl epic new <title>` writes a lint-clean, auto-numbered epic.
- [x] Created files pass `tskflwctl lint` (demoed end-to-end). Same
      markdown+frontmatter pm reads — interop preserved.
- [x] Full suite + lint green.

## Out of scope

- `--edit` ($EDITOR) interactive flow — keep `new` non-interactive (cf. `init`).
- `adr`/`project` create (separate deferred features).
- Templating/config-driven bodies — a single built-in scaffold for now.

## Progress Log

**2026-06-08 — done.** `task new` + `epic new` landed end-to-end. New files:
`domain/slug.go` (`Slugify`), `store/create.go` (`buildFile`/`CreateTask`/
`CreateEpic`/auto-number), shared `assembleFile`; `core.Service.NewTask`/
`NewEpic` (+ `NewTaskParams`/`NewEpicParams` + handoff body templates);
`epicMetaJSON`-style `render.Created{Human,JSON}`; CLI `task new`/`epic new`
(+`App.rel`, `--epic` completion). `domain.Epic` gained `Created`. Defaults
mirror pm (effort Unknown, tier/autonomy 3, priority medium); epic status
defaults `planning`. Demoed: created an epic + a colon-titled task, both
lint-clean, frontmatter canonical + correctly quoted. Full suite + lint green
(11 new tests across domain/store/core/cli). **This closes the keystone gap —
the Go CLI can now run the full create→update→move→lint loop without pm.**

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); unblocks dropping Python `pm` for the daily loop
  ([port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec](6f9menr01nsd-port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec.md)).
