# taskflow

Home of **`tskflwctl`** — a local-first planning CLI over markdown+frontmatter
task/epic/audit files. It's the Go port of the Python `pm` prototype (now
retired — see below), and it dogfoods on its own planning under
[`planning/`](./planning/).

## Map

| Path | Purpose |
| :--- | :--- |
| **[`cmd/tskflwctl/`](./cmd/tskflwctl/)** | The CLI entrypoint (thin composition root). |
| **[`internal/`](./internal/)** | `domain` (pure) · `core` (use cases) · `store` (markdown adapter) · `cli` (cobra) · `tui` (Bubble Tea) · `theme` (shared glyphs/colors) · `config`. |
| **[`planning/`](./planning/)** | This repo's own epics, tasks, and research (self-hosted). |
| **[`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)** | One-screen orientation: the primary/secondary-adapter design. |

## Quick start

```bash
just build              # → bin/tskflwctl
just run task list      # run without installing
just install            # put tskflwctl on $PATH
```

## Daily workflow

`tskflwctl` anchors to the nearest planning repo (walks up for `tasks/`; `-C` to
override). All commands take `--json` for scripting/agents, and every mutating
command takes `--dry-run` to preview the write (full validation runs; nothing
is written).

```bash
tskflwctl init                         # scaffold a planning tree here
tskflwctl status                       # at-a-glance board: counts, in-progress, epic progress

# create
tskflwctl task new "Add retry backoff" --epic 17-pm-go-cli --tags net
tskflwctl epic new "Billing overhaul" --description "Replace legacy pipeline"

# read
tskflwctl task list                    # active tasks (--all / --status / --epic / --tag)
tskflwctl task show <slug>
tskflwctl epic list                    # rollup: done/total per epic
tskflwctl audit list                   # open audits (--all / --closed / --deferred)
tskflwctl ui                           # interactive Bubble Tea browser (tasks/epics/audits)

# update + lifecycle
tskflwctl task set <slug> --priority high --tags a,b
tskflwctl task start|promote|demote|complete|defer|deprecate <slug>...
tskflwctl audit close|reopen|defer <slug>...

# hygiene
tskflwctl lint                         # validate active task frontmatter
tskflwctl lint --fix                   # auto-repair (quote colons, normalize lists)
```

A task's `status:` **is** its directory (`tasks/<status>/`); lifecycle verbs move
the file and stamp dates atomically. Errors carry semantic exit codes — `10`
not-found, `11` validation, `13` ambiguous, `14`
conflict (e.g. a name already taken).

Human output is colorized with status glyphs on a terminal and falls back to
plain text when piped. Control it with `--color=auto|always|never`, `--no-color`,
or the env vars [`NO_COLOR`](https://no-color.org) / `FORCE_COLOR` (the latter
forces color even off a TTY — handy for agents). `--json` is always plain.
`tskflwctl version` / `--version` report the build.

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
toggles raw). Vim-first keys: `:` command-jump, `/` filter (slug/desc/tags),
`o`/`O` sort, `s`/`S` status views, `[`/`]` tabs, `a` task actions
(start/complete/…), `f` to follow a reference (task ⇄ epic) with `ctrl+o` to
jump back, `/`+`n`/`N` find-in-body when the detail is focused, `?` for the
full keymap, `r` to refresh. It **live-reloads** via `fsnotify` — edits from
your editor or a CLI `task move` in another terminal show up within ~200ms,
cursor preserved. See
[`planning/epics/18-tui-bubble-tea-interactive-planning-browser.md`](./planning/epics/18-tui-bubble-tea-interactive-planning-browser.md).
