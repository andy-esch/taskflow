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

The featured tapes — **`tui`** (hero), **`status`**, **`audit-show`**, and
**`picker`** — are the ones shown in the READMEs. The first three each `cd`
(hidden) into the curated [`assets/demo-planning/`](../demo-planning/) fixture, a
bike-workshop planning tree authored to show the symbology off: epics
mid-progress, tasks across every status, and an open audit whose findings span
fixed / landed / in-progress / open / deferred / wontfix (so the segmented bar
shows all its bands). `picker` works on a **throwaway copy** of the fixture (in
`/tmp`) because it actually creates a task through the interactive prompts, and
mustn't dirty the committed tree. Regenerate the fixture itself by re-running the
`tskflwctl epic/task/audit new` commands, or edit the markdown in place.

`epic-show.tape` and `task-list.tape` are extra tapes `just gifs` renders but no
README currently links — keep, link, or prune as you like.

**Tab completion.** `epic-show`, `task-list`, and `audit-show` build their
commands from short prefixes + `Tab` (the epic/audit slug, the subcommands, and
`task list`'s `-o`/`-c` values) rather than typing them out. Those three run
under `zsh` and source `tskflwctl completion zsh` off-camera first, because macOS
ships bash 3.2 — too old for cobra's bash completion. VHS gives every shell the
same `> ` prompt, so they still match the `bash` demos.
