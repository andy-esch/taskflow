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
tskflwctl task show <slug>
tskflwctl epic list                    # rollup: done/total per epic
tskflwctl audit list                   # open audits (--all / --closed / --deferred)
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
`F` toggles the filter between fuzzy and substring,
`o`/`O` sort, `s`/`S` status views, `[`/`]` tabs, `a` task actions
(start/complete/…), `f` to follow a reference (task ⇄ epic) with `ctrl+o` to
jump back, `/`+`n`/`N` find-in-body when the detail is focused, `?` for the
full keymap, `r` to refresh. It **live-reloads** via `fsnotify` — edits from
your editor or a CLI `task move` in another terminal show up within ~200ms,
cursor preserved. See
[`planning/epics/18-tui-bubble-tea-interactive-planning-browser.md`](./planning/epics/18-tui-bubble-tea-interactive-planning-browser.md).
