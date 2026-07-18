# Demo tapes (vhs)

`.tape` scripts for the README demo GIFs, rendered with
[vhs](https://github.com/charmbracelet/vhs).

Regenerate every GIF into `assets/` with:

```
just gifs
```

Requires `vhs` (plus its `ttyd` + `ffmpeg` deps) on `PATH` — it isn't a build or
runtime dependency, only needed to (re)record the demos. `just gifs` builds the
binary first and runs each tape against `./bin/tskflwctl` (shown as `tskflwctl`
via a `PATH` prepend), so the GIFs always reflect the current code.

The featured tapes — **`tui`** (hero), **`status`**, **`audit-show`** — each `cd`
(hidden) into the curated [`assets/demo-planning/`](../demo-planning/) fixture, a
small planning tree authored to show the symbology off: epics mid-progress, tasks
across every status, and an open audit whose findings span fixed / landed /
in-progress / open / deferred / wontfix (so the segmented bar shows all its
bands). Regenerate the fixture itself by re-running the `tskflwctl epic/task/audit
new` commands, or edit the markdown in place.

`help.tape`, `epic-show.tape`, and `task-list.tape` are gallery demos linked from
[`assets/README.md`](../README.md), not the hero README — keep or prune as you
like; `just gifs` still renders them. `epic-show.tape` also shows **Tab
completion** of the epic id: it runs under `zsh` (not `bash`) and sources
`tskflwctl completion zsh` off-camera first, because macOS ships bash 3.2, which
is too old for cobra's bash completion. VHS gives every shell the same `> `
prompt, so it still matches the other demos.
