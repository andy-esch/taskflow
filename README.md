# taskflow

Home of **`tskflwctl`** — a local-first planning CLI over markdown+frontmatter
task/epic/audit files. It's the Go port of the Python `pm` prototype (now
retired — see below), and it dogfoods on its own planning under
[`planning/`](./planning/).

## Map

| Path | Purpose |
| :--- | :--- |
| **[`cmd/tskflwctl/`](./cmd/tskflwctl/)** | The CLI entrypoint (thin composition root). |
| **[`internal/`](./internal/)** | `domain` (pure) · `core` (use cases) · `store` (markdown adapter) · `cli` (cobra) · `config`. |
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
override). All commands take `--json` for scripting/agents.

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
not-found, `11` validation, `12` invalid-transition, `13` ambiguous, `14`
conflict (e.g. a name already taken).

Human output is colorized with status glyphs on a terminal and falls back to
plain text when piped. Control it with `--color=auto|always|never`, `--no-color`,
or the env vars [`NO_COLOR`](https://no-color.org) / `FORCE_COLOR` (the latter
forces color even off a TTY — handy for agents). `--json` is always plain.
`tskflwctl version` / `--version` report the build.

### `pm` is retired

The Python prototype (`bin/pm`) this tool was ported from is **no longer used** —
`tskflwctl` covers the full create → update → move → lint loop. `bin/pm` and
`tests/test_pm.py` are kept only as the historical executable spec.

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

> **Note:** `services/` and `internal/tui/` are leftovers from an earlier Go
> spike (a Python "brain" and a Bubble Tea TUI sketch) and are not part of the
> current CLI. The TUI will be rebuilt over the same `core` in a later phase.
