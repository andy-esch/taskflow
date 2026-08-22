# Demos

Recorded GIFs of `tskflwctl`, linked from the [root README](../README.md#demos).
Each is rendered with [vhs](https://github.com/charmbracelet/vhs) against a curated
fixture, so the output always tracks the current code. Most record against the single
[`demo-planning/`](./demo-planning/) tree; the atlas needs several, and stages its own.

## The TUI (`tskflwctl ui`)

Tab across tasks, epics, and audits — status glyphs, epic rollup bars, and an
audit's **segmented finding bar** over its status-grouped **finding tree**.

![the tskflwctl TUI](./tui.gif)

## The atlas (`tskflwctl ui`, outside a planning repo)

Standing in a planning repo, `ui` opens that repo. Run it anywhere else and there is no
statement of which space you meant, so it opens the **atlas**: one card per durable
planning identity, every registered entry point beneath it, and reversible navigation
into a space and back without restarting the program.

`bike-workshop` appears once but is reachable two ways — the planning checkout itself and
an impl repo pointing at it — because a card is a planning *identity*, not a directory.
`h`/`l` selects between them, `o`/`O` reorders the cards, and `a` returns from a space to
the atlas with that space still open behind it.

![the tskflwctl atlas](./atlas.gif)

## `tskflwctl status`

The at-a-glance board: status counts, the in-progress set, epic rollup bars, and
the **Open-audits** section with its segmented finding bar.

![tskflwctl status](./status.gif)

## `tskflwctl audit show <id>`

The **segmented finding bar** (`█` done · `▓` in-progress · `▒` dropped · `░`
open) above the status-grouped **finding tree**.

![tskflwctl audit show](./audit-show.gif)

## Interactive pickers (`task new` with no `--epic`)

On a TTY, a missing required input **prompts** instead of erroring: `task new`
without `--epic` opens a type-to-filter epic picker, then a free-text tags
prompt. Off a TTY (a pipe, `--json`, `--no-input`) the same command fails with
the flag to pass — so nothing interactive can block a script. Prompts render to
stderr, so stdout stays a clean data stream.

![task new epic picker](./picker.gif)

---

## How they're made

- **[`vhs/`](./vhs/)** — the `.tape` scripts and the `just gifs` recipe that
  renders them. See [`vhs/README.md`](./vhs/README.md). vhs is a dev-only tool,
  not a build or runtime dependency.
- **[`demo-planning/`](./demo-planning/)** — the curated planning tree the tapes
  record against, shaped to exercise the symbology. See
  [`demo-planning/README.md`](./demo-planning/README.md).
- **[`demo-kitchen/`](./demo-kitchen/)** and **[`demo-bike-shop/`](./demo-bike-shop/)** —
  the second planning identity and the pointer repo that give the atlas more than one
  card and more than one entry point. Only `atlas.tape` uses them, via
  [`vhs/atlas-setup.sh`](./vhs/atlas-setup.sh), which stages copies into a throwaway
  registry rather than ever touching your real one.

Regenerate every GIF with `just gifs` (builds `bin/tskflwctl` first, so the GIFs
reflect the current code).
