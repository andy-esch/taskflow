# Demos

Recorded GIFs of `tskflwctl`, linked from the [root README](../README.md#demos).
Each is rendered with [vhs](https://github.com/charmbracelet/vhs) against the
curated [`demo-planning/`](./demo-planning/) fixture, so the output always tracks
the current code.

## The TUI (`tskflwctl ui`)

Tab across tasks, epics, and audits — status glyphs, epic rollup bars, and an
audit's **segmented finding bar** over its status-grouped **finding tree**.

![the tskflwctl TUI](./tui.gif)

## `tskflwctl status`

The at-a-glance board: status counts, the in-progress set, epic rollup bars, and
the **Open-audits** section with its segmented finding bar.

![tskflwctl status](./status.gif)

## `tskflwctl audit show <id>`

The **segmented finding bar** (`█` done · `▓` in-progress · `▒` dropped · `░`
open) above the status-grouped **finding tree**.

![tskflwctl audit show](./audit-show.gif)

---

## How they're made

- **[`vhs/`](./vhs/)** — the `.tape` scripts and the `just gifs` recipe that
  renders them. See [`vhs/README.md`](./vhs/README.md). vhs is a dev-only tool,
  not a build or runtime dependency.
- **[`demo-planning/`](./demo-planning/)** — the curated planning tree the tapes
  record against, shaped to exercise the symbology. See
  [`demo-planning/README.md`](./demo-planning/README.md).

Regenerate every GIF with `just gifs` (builds `bin/tskflwctl` first, so the GIFs
reflect the current code).
