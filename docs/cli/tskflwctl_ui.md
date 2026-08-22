## tskflwctl ui

Launch the interactive TUI (Bubble Tea)

### Synopsis

Navigate registered spaces from the atlas, then browse and update planning entities
without restarting the full-screen TUI. Press `a` or run `:atlas` to return; `o` changes
atlas ordering and `O` reverses it. Open the shared
Config/About editor from Overview, with `:config`, or from the command palette; writes
use the same typed application service as `tskflwctl config edit`.

```
tskflwctl ui [flags]
```

### Examples

```
  tskflwctl ui
```

### Options

```
  -h, --help   help for ui
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits, research) over markdown

