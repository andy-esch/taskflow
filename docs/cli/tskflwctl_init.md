## tskflwctl init

Scaffold a planning tree here, or point at an external planning repo

### Synopsis

Bootstrap a new planning topology: either scaffold a local planning tree or
point this repository at an existing external planning repo. A fresh config is
registered as a machine-local space best-effort; use --no-register to opt out.
Repeat init with the configured --taskflow-root to repair the scaffold, create
threads/, and remove only an unused legacy projects/ placeholder; user-owned
legacy content is preserved and reported.
Bare init against an existing configuration reports its topology without
changing it; use
`tskflwctl config migrate` for safe configuration upgrades.

```
tskflwctl init [flags]
```

### Examples

```
  tskflwctl init
  tskflwctl init --taskflow-root planning
  tskflwctl init --planning-repo ../desirelines-planning
  tskflwctl init --no-register
```

### Options

```
  -h, --help                   help for init
      --no-link-back           pointer mode: don't add this repo to the planning repo's tracked_repos
      --no-register            don't add a freshly initialized repo to this machine's space registry (also TSKFLW_NO_REGISTER)
      --path string            directory to initialize (default ".")
      --planning-repo string   point this repo at an external planning repo (relative to --path, or absolute): writes a pointer config, no tree
      --taskflow-root string   scaffold the planning tree in this subdirectory instead of the repo root (sets taskflow_root; e.g. planning)
      --track strings          record an impl repo this planning repo tracks (repeatable; scaffold mode only)
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path (conflicts with --space)
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager       do not pipe long human output through a pager
      --paginate       page long human output through $PAGER (on a TTY), even if disabled in config
      --space string   select a registered entry point by label (also TSKFLW_SPACE; conflicts with -C)
      --theme string   color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, Threads, epics, audits, research) over markdown

