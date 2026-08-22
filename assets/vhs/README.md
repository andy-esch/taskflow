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

The featured tapes — **`tui`** (hero), **`atlas`**, **`status`**, **`audit-show`**,
and **`picker`** — are the ones shown in the READMEs. The first three each `cd`
(hidden) into the curated [`assets/demo-planning/`](../demo-planning/) fixture, a
bike-workshop planning tree authored to show the symbology off: epics
mid-progress, tasks across every status, and an open audit whose findings span
fixed / landed / in-progress / open / deferred / wontfix (so the segmented bar
shows all its bands). `picker` works on a **throwaway copy** of the fixture (in
`/tmp`) because it actually creates a task through the interactive prompts, and
mustn't dirty the committed tree. Regenerate the fixture itself by re-running the
`tskflwctl epic/task/audit new` commands, or edit the markdown in place.

**The atlas needs more than one tree.** `atlas.tape` is the exception to
"one fixture, `cd` into it": a cross-space navigator has nothing to show with a single
space, so it stages a throwaway **registry** via
[`atlas-setup.sh`](./atlas-setup.sh) instead of a throwaway copy. That script is the
only piece of this directory that could touch something outside the repo, so it is
deliberate about two things: `TSKFLW_CONFIG_HOME` redirects the whole home config into
the staging dir, so `space add` can never reach the recorder's real `spaces.toml`; and
every tree is a copy, so the committed fixtures stay clean. It stages three:

| staged as | from | role |
| --- | --- | --- |
| `bike-workshop` | [`demo-planning/`](../demo-planning/) | the direct planning checkout |
| `bike-shop` | [`demo-bike-shop/`](../demo-bike-shop/) | an impl repo *pointing at* that same tree |
| `kitchen` | [`demo-kitchen/`](../demo-kitchen/) | a second, unrelated planning identity |

`bike-workshop` and `bike-shop` share one durable planning id, so the atlas groups them
into **one card with two entry points** — which is the only reason the `h`/`l` entry-point
selection has anything to demonstrate. `demo-bike-shop/` exists solely for that: it is a
config file and nothing else. The directories are renamed on copy so the paths the atlas
prints read as the space you are looking at rather than as fixture plumbing, and the
pointer's relative `planning_repo` is rewritten to match.

The tape then records from the staging dir itself, which has no planning repo above it —
the case `ui` answers by landing on the atlas, so the demo opens there without a keystroke
spent getting to it.

`epic-show.tape` and `task-list.tape` are extra tapes `just gifs` renders but no
README currently links — keep, link, or prune as you like.

**Tab completion.** `epic-show`, `task-list`, and `audit-show` build their
commands from short prefixes + `Tab` (the epic/audit slug, the subcommands, and
`task list`'s `-o`/`-c` values) rather than typing them out. Those three run
under `zsh` and source `tskflwctl completion zsh` off-camera first, because macOS
ships bash 3.2 — too old for cobra's bash completion. VHS gives every shell the
same `> ` prompt, so they still match the `bash` demos.
