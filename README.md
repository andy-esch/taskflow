# taskflow

Home of **`tskflwctl`** — a local-first planning CLI over markdown+frontmatter
task/epic/audit files. It's the Go port of the Python `pm` prototype (now
retired — see below), and it dogfoods on its own planning under
[`planning/`](./planning/).

## Demos

The interactive TUI (`tskflwctl ui`) — tab across tasks, epics, and audits;
status glyphs, epic rollup bars, and an audit's **segmented finding bar** over its
status-grouped **finding tree**:

![the tskflwctl TUI](./assets/tui.gif)

…and the same vocabulary on the CLI:

| | |
| :-- | :-- |
| `tskflwctl status` — counts, in-progress, epic bars, open audits | ![status](./assets/status.gif) |
| `tskflwctl audit show <id>` — segmented finding bar + finding tree | ![audit show](./assets/audit-show.gif) |

▸ **[All demos, how they're recorded, and the demo fixture →
`assets/README.md`](./assets/README.md)** — rendered with
[vhs](https://github.com/charmbracelet/vhs) against a curated planning tree;
regenerate with `just gifs`.

## Map

| Path | Purpose |
| :--- | :--- |
| **[`cmd/tskflwctl/`](./cmd/tskflwctl/)** | The CLI entrypoint (thin composition root). |
| **[`internal/`](./internal/)** | `domain` (pure) · `core` (use cases) · `store` (markdown adapter) · `cli` (cobra) · `tui` (Bubble Tea) · `theme` (shared glyphs/colors) · `config`. |
| **[`planning/`](./planning/)** | This repo's own epics, tasks, and research (self-hosted). |
| **[`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)** | One-screen orientation: the primary/secondary-adapter design. |

## Install

Distribution is **GitHub Releases only** — three paths, no Homebrew/external channels:

```bash
# 1) Prebuilt binary (no Go toolchain). The repo is private, so use gh (it
#    handles auth); pick your platform: darwin/linux × amd64/arm64.
gh release download -R andy-esch/taskflow -p "*linux_arm64*"
tar xzf tskflwctl_*_linux_arm64.tar.gz && ./tskflwctl version

# 2) go install from source (needs Go + git auth to the private repo)
GOPRIVATE=github.com/andy-esch/* \
  go install github.com/andy-esch/taskflow/cmd/tskflwctl@latest   # or @vX.Y.Z

# 3) From a checkout
just install            # → go install onto $PATH (version-stamped)
```

## Build / dev

```bash
just build              # → bin/tskflwctl
just run task list      # run without installing
just install            # put tskflwctl on $PATH
just release-snapshot   # dry-run a full release into ./dist (publishes nothing)
```

Releases are cut by pushing a tag (`vX.Y.Z`), which runs `.github/workflows/release.yml`
(goreleaser). A manual `workflow_dispatch` run builds a `--snapshot` and uploads
the binaries as workflow artifacts without minting a Release.

## Daily workflow

`tskflwctl` anchors to the nearest planning repo (walks up for `tasks/`; `-C` to
override). All commands take `--json` for scripting/agents, and mutating
commands take `--dry-run` to preview the write (full validation runs; nothing
is written) — except the interactive `task edit`, which has no preview.

```bash
tskflwctl init                         # scaffold a planning tree here
tskflwctl status                       # at-a-glance board: counts, in-progress, epic progress

# create
tskflwctl task new "Add retry backoff" --epic 17-pm-go-cli --tags net
tskflwctl task new "Triage flake" --epic 17-pm-go-cli --tags ci --description "is CI red?" --start  # straight to in-progress (--next/--start need --description)
echo "$BODY" | tskflwctl task new "Long writeup" --epic 17-pm-go-cli --tags x --body-file -  # body from stdin/file
tskflwctl epic new "Billing overhaul" --description "Replace legacy pipeline"
tskflwctl audit new dispatcher          # → audits/open/YYYY-MM-DD-dispatcher.md (--date to override)
tskflwctl audit new auth --template security  # pick a body scaffold (default|security); --template is shell-completable

# read
tskflwctl task list                    # active tasks (--all / --status / --epic / --tag)
tskflwctl task list --revisit-due      # deferred tasks whose snooze date has arrived
tskflwctl task show <slug>             # metadata + body (--section <name> / --frontmatter-only to narrow)
tskflwctl task info <slug> --json      # token-cheap metadata: path, status, epic, ac:{checked,total} (no body)
tskflwctl task path <slug>             # just the absolute file path — $EDITOR "$(tskflwctl task path <slug>)"
tskflwctl epic list                    # rollup: done/total per epic
tskflwctl epic show <id> --section goal # epic body section (or --frontmatter-only); epic path <id> for the file
tskflwctl audit list                   # open audits (--all / --closed / --deferred)
tskflwctl audit show <slug> --section findings  # audit body section (or --frontmatter-only)
tskflwctl audit info <slug> --json     # token-cheap: path, bucket, findings:{total,open,in_progress,done,dropped}
tskflwctl audit path <slug>            # just the absolute file path (like task path)
tskflwctl audit findings --status open --effort XS,S --json  # query findings across audits
tskflwctl audit lint                   # validate finding status vocab + missing status + bucket↔state
tskflwctl schema                       # the tool's contract for agents (statuses, fields, codes)
tskflwctl schema task --json           # how to author a task: sections, fields, conventions
tskflwctl template list                # body scaffolds `new --template` can use (--kind to filter)
tskflwctl template show audit security # inspect a template's rendered body (--json for the envelope)
tskflwctl ui                           # interactive Bubble Tea browser (tasks/epics/audits)

# update + lifecycle
tskflwctl task set <slug> --priority high --tags a,b
tskflwctl task edit <slug>                          # open the whole file in $EDITOR (human; re-validated on save)
echo "## Findings" | tskflwctl task append <slug> --body-file -  # add a section (agent; atomic)
tskflwctl task set <slug> --body-file notes.md      # replace the body (agent; its own call)
tskflwctl task ac <slug>                            # numbered acceptance criteria; --check/--uncheck <n> to flip one
tskflwctl task start|next|ready|complete|defer|deprecate <slug>...   # defer takes --until <date>
tskflwctl task defer <slug> --until 2026-09-01      # snooze (revisit_at); on a TTY, prompts for the date
tskflwctl audit close|reopen|defer <slug>...

# hygiene
tskflwctl lint                         # validate active task frontmatter
tskflwctl lint --fix                   # auto-repair (normalize, relocate misfiled, backfill ids)
```

A task's `status:` is authoritative in frontmatter; `tasks/<status>/` is a
lock-step mirror of it. Lifecycle verbs change the status and relocate the file,
stamping dates atomically (`lint --fix` re-syncs a hand-edited drift). Errors
carry semantic exit codes — `10`
not-found, `11` validation, `13` ambiguous, `14`
conflict (e.g. a name already taken).

**Body templates.** Each kind ships named body scaffolds; `task/epic/audit new
--template <name>` picks one (omit it for `default`). Names are shell-completable
and an unknown one fails with exit `11` listing what's available. `audit` ships a
`security` template (threat model + checklist) alongside `default`. `--template` is
mutually exclusive with `--body`/`--body-file` (pick a scaffold *or* supply your
own). Discover what's available with `template list` (`--kind` to filter, `--json`
for agents) and `template show <kind> [name]`. Repo-local and custom templates are
the next step (see `planning/epics/22-selectable-template-library.md`).

Human output is colorized with status glyphs on a terminal and falls back to
plain text when piped. Control it with `--color=auto|always|never`, `--no-color`,
or the env vars [`NO_COLOR`](https://no-color.org) / `FORCE_COLOR` (the latter
forces color even off a TTY — handy for agents). `--json` is always plain.
`tskflwctl version` / `--version` report the build.

### Pipelines

The `list` commands (`task`/`epic`/`audit list`) share one output-format flag,
`-o/--output`, plus an orthogonal column selector:

| `-o` | Output | For |
| :--- | :--- | :--- |
| `human` *(default)* | colorized table | reading on a terminal |
| `name` | ids only, one per line | `… \| xargs` |
| `table` | tab-separated, header row, absolute dates, no color/truncation | `cut`/`awk`; stable across versions |
| `csv` | RFC 4180 comma-separated, header row | spreadsheets; cells with commas are quoted |
| `json` | full records + `schema_version` | `jq` |

`-q`/`--quiet` is shorthand for `-o name`; `--json` (on every command) equals
`-o json`. `-c/--columns slug,status,…` projects the columnar formats (`table`,
`csv`) to the columns you name, in the order you name them (and implies
`-o table`) — both the formats and the column names are shell-completable. `-o table` is a documented contract under
the one `schema_version` (a column add/reorder is a schema bump), and always
emits the header row — even with zero results — so a consumer gets a stable
schema and detects "no rows" by line count. Recipes:

```bash
# start every ready-to-start task tagged `tui`
tskflwctl task list -q --tag tui | xargs tskflwctl task start

# audits with open findings, projected to slug + open count
tskflwctl audit list --all -o table -c slug,open | awk -F'\t' 'NR>1 && $2>0 {print $1}'

# in-progress slugs via jq
tskflwctl task list --status in-progress -o json | jq -r '.tasks[].slug'
```

**stdout is data, stderr is diagnostics** — per-item transition failures,
file-read problems, and prompts go to stderr, so a partial `… | xargs` never
interleaves errors into the data stream.

### Interactivity: one capability, two faces

When a required input is missing, `tskflwctl` picks the *face* based on the
terminal, never the capability ([clig.dev](https://clig.dev/#interactivity)):

- **On a TTY** (a human): it prompts — `task new` without `--epic` opens an epic
  picker; a bare `task start` opens a picker over ready-to-start tasks. Prompts
  render to **stderr**, so stdout stays byte-identical to the flag-driven run.
- **Off a TTY** (a pipe, an agent, `--json`, or `--no-input` / `TSKFLW_NO_INPUT=1`):
  it never prompts — it fails with today's exit code (11) naming the flag to pass.

So every prompt has a flag twin and nothing interactive can ever block a script.
Ctrl-C out of a prompt exits **130** (the SIGINT convention) with a quiet
`aborted`, not an error.

Long human output pages the same way: `show` / `schema` pipe through a pager (like
git) **only on a TTY** — never under a pipe, `--json`, or `--no-input`, so machine
output stays byte-identical. Program precedence: `TSKFLW_PAGER` → `[pager].command`
(in `.tskflwctl.toml`) → `$PAGER` → `less -FRX`. On/off: `--no-pager` → `--paginate`
→ `[pager].enabled` → default on.

### `pm` is retired

The Python prototype (`bin/pm`) this tool was ported from is **gone** —
`tskflwctl` covers the full create → update → move → lint loop, and the Go
test suite is the executable spec now. The prototype and its tests live only
in git history (last at commit `39f1b83`) if archaeology is ever needed.

## Shell completion

`tskflwctl` ships cobra-generated completion for bash/zsh/fish. For zsh:

```bash
just install         # put tskflwctl on $PATH (completion shells out to it)
just completion-zsh  # writes ~/.zsh/completions/_tskflwctl + prints the one-time setup
```

If `~/.zsh/completions` isn't already on your `$fpath`, add once to `~/.zshrc`:

```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

Other shells: `just completion bash` / `just completion fish` print the script
to stdout (see `tskflwctl completion --help`). Completion covers the command
tree, flags, **and** task/audit/epic slugs — e.g. `task show <TAB>`,
`audit close <TAB>`, `epic show <TAB>` offer the real slugs (and still work when
a file's frontmatter is malformed).

## Development

`just` wraps the common tasks:

- `just build` — build `bin/tskflwctl`
- `just run *ARGS` — `go run ./cmd/tskflwctl …`
- `just test` — `go test ./...`
- `just lint` — `golangci-lint run ./...`
- `just fmt` — gofmt + lint formatting
- `just tidy` — `go mod tidy`

Design rationale lives in [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) and
[`planning/epics/17-pm-go-cli.md`](./planning/epics/17-pm-go-cli.md).

### Interactive TUI (`tskflwctl ui`)

A Bubble Tea browser over the **same `core`** the CLI uses (a second primary
adapter — never the filesystem directly). Two panes: an entity list (tasks /
epics / audits) and a detail preview rendered as **glamour markdown** (`R`
toggles raw). Vim-first keys: `ctrl+p` command palette (fuzzy-jump to any
task/epic/audit or run a command), `:` command-jump, `/` filter (slug/desc/tags),
`F` toggles the filter between fuzzy and substring,
`o`/`O` sort, `s`/`S` status views, `[`/`]` tabs, `m` move (lifecycle:
start/complete/defer/…), `e` to edit a task's fields inline, `E` to open the
whole file in `$EDITOR` (any entity; re-read on save via live-reload), `f` to
follow a reference (task ⇄ epic) with `ctrl+o` to
jump back, `y`/`Y` to copy the selection's slug / file path to the system
clipboard (a native tool — pbcopy/wl-copy/xclip — when available, else OSC 52 so
it still works over SSH), `/`+`n`/`N` find-in-body when the
detail is focused, `?` for the
full keymap, `r` to refresh. The detail pane's title is a **click-to-open link**
(OSC 8) to the entity's file, and the terminal window/tab title tracks the current
selection. It **live-reloads** via `fsnotify` — edits from
your editor or a CLI `task move` in another terminal show up within ~200ms,
cursor preserved. See
[`planning/epics/18-tui-bubble-tea-interactive-planning-browser.md`](./planning/epics/18-tui-bubble-tea-interactive-planning-browser.md).

> **Clickable titles under tmux + Ghostty.** The detail-title link is correct OSC 8
> and works in a bare terminal as-is; tmux needs two things.
> **(1) Let tmux pass the hyperlink through** — tmux ≥ 3.4 with
> `set -as terminal-features ',xterm-ghostty:hyperlinks'` (match your real `$TERM`),
> then a full **`tmux kill-server`** — a `source-file` reload often does *not* apply
> `terminal-features`. Confirm with `tmux info | grep -i hyperlink`.
> **(2) Click through tmux's mouse capture** — with `mouse on`, open links via
> **shift+cmd+click** (Shift bypasses tmux's grab; make sure Ghostty's
> `mouse-shift-capture` isn't `true`), or `set -g mouse off` for plain cmd+click
> (you lose tmux mouse scroll/select). Quick isolate, in a tmux pane:
> `printf '\e]8;;https://example.com\e\\Click Me\e]8;;\e\\\n'` — if **Click Me** is
> underlined, rendering already works and it's only the click modifier (step 2);
> if it's plain text, tmux isn't passing OSC 8 yet (step 1).
