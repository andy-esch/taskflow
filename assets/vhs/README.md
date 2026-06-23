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
via a `PATH` prepend), so the GIFs always reflect the current code. Tapes render
against this repo's own [`planning/`](../../planning/) tree (taskflow dogfoods
its planning), so output tracks real data.
