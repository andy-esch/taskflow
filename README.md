# taskflow

Home of **`tskflwctl`** — a local-first planning CLI over markdown+frontmatter
task/epic/audit/research files. It dogfoods on its own planning under
[`planning/`](./planning/).

## Demos

The interactive TUI (`tskflwctl ui`) — navigate registered planning spaces from the
atlas, then tab across tasks, epics, audits, and research; status glyphs, epic rollup
bars, and an audit's **segmented finding bar** over its status-grouped **finding tree**:

![the tskflwctl TUI](./assets/tui.gif)

…and the same vocabulary on the CLI:

| | |
| :-- | :-- |
| `tskflwctl status` — counts, in-progress, epic bars, open audits | ![status](./assets/status.gif) |
| `tskflwctl audit show <id>` — segmented finding bar + finding tree | ![audit show](./assets/audit-show.gif) |
| `tskflwctl task new` with no `--epic` — on a TTY it prompts: an epic picker, then tags | ![epic picker](./assets/picker.gif) |

▸ **[All demos, how they're recorded, and the demo fixture →
`assets/README.md`](./assets/README.md)** — rendered with
[vhs](https://github.com/charmbracelet/vhs) against a curated planning tree;
regenerate with `just gifs`.

## Map

| Path | Purpose |
| :--- | :--- |
| **[`cmd/tskflwctl/`](./cmd/tskflwctl/)** | The CLI entrypoint (thin composition root). |
| **[`internal/`](./internal/)** | `domain` (pure) · `core` (use cases) · `store` (markdown adapter) · `cli` (cobra) · `tui` (Bubble Tea) · `config`/`userconfig` · `spacehealth`. |
| **[`planning/`](./planning/)** | This repo's own epics, tasks, and research (self-hosted). |
| **[`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md)** | One-screen orientation: the primary/secondary-adapter design. |

## Install

Distribution is **GitHub Releases only** — three paths, no Homebrew/external channels:

```bash
# 1) go install (needs Go toolchain)
go install github.com/andy-esch/taskflow/cmd/tskflwctl@latest   # or @vX.Y.Z

# 2) Prebuilt binary (no Go toolchain) — download the archive for your platform
#    (darwin/linux × amd64/arm64) from the GitHub Releases page, then:
tar xzf tskflwctl_*_linux_arm64.tar.gz && ./tskflwctl version

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

`tskflwctl` anchors to the nearest planning repo (walks up for `tasks/`; use `-C` for
a path or `--space <label>` for a registered entry point). All commands take `--json`
for scripting/agents, and mutating commands take `--dry-run` to preview the write
(full validation runs; nothing is written) — interactive editors have no preview and
reject `--dry-run`.

```bash
tskflwctl init                         # scaffold here and register this checkout as a space
tskflwctl init --taskflow-root planning  # ...or put it in a subdir, config at the repo root
tskflwctl init --planning-repo ../x-planning  # point at an external planning repo instead
tskflwctl init --no-register           # bootstrap only; leave this machine's registry alone
tskflwctl config show                  # raw repo/user scopes + effective values and sources
tskflwctl config migrate --dry-run     # preview safe config/id upgrades; omit flag to apply
tskflwctl config edit                  # typed theme/pager editor (human TTY only)
tskflwctl config doctor                # linkback + home space-registry health
tskflwctl status                       # at-a-glance board: counts, in-progress, epic progress
tskflwctl status --all                 # compact summaries + one in-progress rail across spaces

# create
tskflwctl task new "Add retry backoff" --epic <epic-id> --tags net
tskflwctl task new "Triage flake" --epic <epic-id> --tags ci --description "is CI red?" --start  # straight to in-progress (--next/--start need --description)
echo "$BODY" | tskflwctl task new "Long writeup" --epic <epic-id> --tags x --body-file -  # body from stdin/file
tskflwctl epic new "Billing overhaul" --description "Replace legacy pipeline"
tskflwctl audit new dispatcher          # → audits/<id>-YYYY-MM-DD-dispatcher.md (--date to override)
tskflwctl audit new auth --template security  # pick a body scaffold (default|security); --template is shell-completable
tskflwctl research new "Compare theming libs" --tags tui --description "Weighed three libs"
tskflwctl research new "Storage options" --created 2026-06-24  # backdate: the id is minted from --created

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
tskflwctl research list                # the whole corpus, newest first (--tag to filter)
tskflwctl research show <slug>         # metadata + body (--section / --frontmatter-only; research path <slug> for the file)
tskflwctl schema                       # the tool's contract for agents (statuses, fields, codes)
tskflwctl schema task --json           # how to author a task: sections, fields, conventions
tskflwctl template list                # body scaffolds `new --template` can use (--kind to filter)
tskflwctl template show audit security # inspect a template's rendered body (--json for the envelope)
tskflwctl ui                           # Bubble Tea browser; a / :atlas navigates spaces

# update + lifecycle
tskflwctl task set <slug> --priority high --tags a,b
tskflwctl task edit <slug>                          # open the whole file in $EDITOR (human; re-validated on save)
echo "## Findings" | tskflwctl task append <slug> --body-file -  # add a section (agent; atomic)
tskflwctl task set <slug> --body-file notes.md      # replace the body (agent; its own call)
tskflwctl task ac <slug>                            # numbered acceptance criteria; --check/--uncheck <n> to flip one
tskflwctl task start|next|ready|complete|defer|deprecate <slug>...   # defer takes --until <date>
tskflwctl task defer <slug> --until 2026-09-01      # snooze (revisit_at); on a TTY, prompts for the date
tskflwctl audit close|reopen|defer <slug>...
tskflwctl research set <slug> --description "…" --tags a,b   # settable fields only; `schema research` lists them
tskflwctl research edit|append <slug>                        # same human/agent pair as task edit|append

# hygiene
tskflwctl lint                         # validate active task, epic, and research frontmatter
tskflwctl lint --fix                   # auto-repair frontmatter (quote ':' values, normalize lists, backfill ids)
```

### Multiple planning repos

The home space registry is an advisory inventory of planning entry points on this machine.
It lives at `$XDG_CONFIG_HOME/tskflwctl/spaces.toml`, falling back to
`~/.config/tskflwctl/spaces.toml`, separate from both the repo-local `.tskflwctl.toml` and
the user-preference `config.toml`. Entry points that resolve to the same durable planning
id form one logical space: a direct planning checkout and its implementation-repo pointers
are grouped rather than presented as duplicate plans. The relationship is derived from the
repo configs; `spaces.toml` does not duplicate parent/child state. Registering an entry point
never changes ordinary cwd/`-C` discovery, and forgetting one never touches its repo.

```bash
tskflwctl space add                         # register the current checkout as an entry point
tskflwctl space add ../other --id other     # or name another checkout explicitly
tskflwctl space list                        # grouped spaces + entry-point health
tskflwctl space list --json                 # flat entries with role + planning_id
tskflwctl --space other status              # run against that exact registered entry point
tskflwctl status --all                      # all logical spaces + combined in-progress work
tskflwctl status --all --json               # one versioned envelope: per-space summaries + diagnoses
tskflwctl --space other task list --json    # works from any directory
TSKFLW_SPACE=other tskflwctl workspace      # environment twin; receipt names the selector
tskflwctl config doctor                     # linkback + space-registry health audit
tskflwctl space forget other --dry-run      # preview; the repo itself is never deleted
tskflwctl space forget other
```

The former top-level `tskflwctl doctor` remains as a hidden compatibility alias for
the current minor-release window; new usage should prefer `config doctor`.

Fresh scaffold and pointer `init` runs register their checkout automatically, using the
directory that contains `.tskflwctl.toml` rather than a nested planning root. Registration
is best-effort: a registry conflict or filesystem failure warns on stderr but never rolls
back or fails the new topology. Use `--no-register` or `TSKFLW_NO_REGISTER=1` to opt out;
use `space add` for an existing checkout, because bare `init` and `config migrate` never
silently mutate home state. `--space <label>` and its `TSKFLW_SPACE` twin make those local
labels exact checkout targets now, even when several labels share one planning identity.
`--space` and `-C` are mutually exclusive;
an unknown, broken, or durable-id-mismatched entry fails loudly rather than falling back
to cwd discovery. Machine mutation receipts carry the selected label as `workspace.space`.
`status --all` is the registry-wide exception to cwd anchoring: it runs from anywhere,
loads each logical planning identity once (healthy direct checkout first, then the first
healthy pointer), and keeps broken entry-point diagnoses inline without failing healthy
spaces. Registry health remains informational, but unreadable planning files or a selected
tree that cannot be loaded produce a non-zero partial-failure exit after every available
summary has rendered. With no registered spaces it emits the ordinary single-repo `status`
output.
The same grouped projection now powers the TUI atlas. With more than one logical space,
`tskflwctl ui` lands there; `j`/`k` select a space, `h`/`l` choose among its direct and
pointer entry points, and Enter opens it without restarting the TUI. Entry paths stay
visible in dim text. Cards default to name order; `o` cycles name/activity/registration
order and `O` reverses it. Press `a`, run `:atlas`, or use the command palette to return.
The atlas uses the home/global theme rather than whichever repo launched it. Each entry point keeps its own tab,
cursor, filter, dashboard, and navigation state across round trips; only the active
space is watched for filesystem changes. Broken entries remain visible but cannot be
entered, and a failed open leaves the current session untouched.
`TSKFLW_CONFIG_HOME` overrides the home-config directory, which is useful for an isolated
trial before populating the real registry.

### Configuration lifecycle

`init` is bootstrap-only: it creates a new local planning tree or a pointer to an
external planning repo. Running bare `init` where `.tskflwctl.toml` already exists reports
the current topology and exits without changing it. If an older file needs a durable
`id` or `planning_repo_id`, it points to the explicit, idempotent migration instead:

```bash
tskflwctl config migrate --dry-run    # exact planned keys and values; writes nothing
tskflwctl config migrate              # atomic, comment-preserving application
tskflwctl config migrate              # current files are a byte-for-byte no-op
```

For a legacy pointer whose planning target also lacks an `id`, migrate the target first,
then the implementation repo. Relative `planning_repo` spelling, comments, unknown keys,
and key order are preserved.

`config show` keeps the raw repository and user scopes separate, then prints the effective
theme and pager values with their provenance. Theme precedence is `--theme` →
`TSKFLW_THEME` → repository config → user config → built-in default. Pager-command
precedence is `TSKFLW_PAGER` → repository config → user config → `$PAGER` → `less -FRX`;
pager enablement is controlled by `--no-pager` / `--paginate`, then repository config,
then user config, then the default.

`config edit` is a typed TTY interface rather than a raw TOML editor. It starts in user
scope; press `s` to choose an explicit repository override. Theme changes repaint the
existing palette preview before save, pager enablement remains inherit/on/off, and `u`
removes the selected override to restore inheritance. Durable IDs, planning topology,
and `spaces.toml` are deliberately read-only there. The Config/About entry on the
`tskflwctl ui` Overview screen (also available as `:config` and in `?` help) embeds the
same component and uses the same application service and writer.

Tasks, audits, and research are stored flat and id-led (`tasks/<id>-<slug>.md`,
`audits/<id>-<slug>.md`, `research/<id>-<slug>.md`); `status:` / `bucket:` is
authoritative in frontmatter, with no mirror directory. Lifecycle verbs edit that
field in place and stamp the dates atomically — no file moves (`lint --fix`
re-normalizes a hand-edited drift). Errors carry semantic exit codes — `10`
not-found, `11` validation, `13` ambiguous, `14` conflict (e.g. a name already taken).

**Research** is the thinnest kind, and the omissions are the point: no status and
no lifecycle verbs (a later doc supersedes an earlier one — a decision that needs a
lifecycle is an [ADR](./planning/adrs/)), and no `epic:` field, so provenance is
ordinary body links rather than a rollup. Its id is minted from `created`, so
listing by id lists by authorship date; that also makes `created` un-settable.
See [`ADR-0001`](./planning/adrs/0001-adopt-adrs.md) for the research/ADR boundary
and `tskflwctl schema research` for the fields.

**Body templates.** Each kind ships named body scaffolds; `<entity> new
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
output stays byte-identical. See `config show` for the effective pager command,
enablement, and the source of each value.

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
tree, flags, registered `--space` labels, **and** task/audit/epic/research slugs —
e.g. `task show <TAB>`, `audit close <TAB>`, `epic show <TAB>` offer the real slugs
(and still work when a file's frontmatter is malformed).

## Development

`just` wraps the common tasks:

- `just build` — build `bin/tskflwctl`
- `just run *ARGS` — `go run ./cmd/tskflwctl …`
- `just test` — `go test ./...`
- `just lint` — `golangci-lint run ./...`
- `just fmt` — gofmt + lint formatting
- `just tidy` — `go mod tidy`

Design rationale lives in [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md).

### Interactive TUI (`tskflwctl ui`)

A Bubble Tea browser over the **same `core`** the CLI uses (a second primary
adapter — never the filesystem directly). When multiple logical spaces are registered,
the atlas is the landing screen: `j`/`k` select a space, `h`/`l` select a registered
entry point, `o`/`O` change card ordering, Enter navigates into it, and `a` / `:atlas`
returns. Inside a space, two
panes show an entity list (tasks / epics / audits / research) and a detail preview
rendered as **glamour markdown** (`R` toggles raw). Vim-first keys: `ctrl+p` command
palette (fuzzy-jump to any entity or run a command), `:` command-jump, `/` filter (slug/desc/tags),
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
