## tskflwctl status

At-a-glance project dashboard (counts, in-progress, epic progress)

### Synopsis

Show the current planning repo's dashboard. With --all, summarize every
logical planning space in the home registry and combine their in-progress work.
Multiple registered entry points sharing one planning id are read only once. The
command works from any directory; -C is used only when the registry is empty.

Broken registry entries remain inline and informational. Unreadable planning files
or a selected tree that fails to load still render every available result, then make
the command exit non-zero so automation can detect the partial result.

```
tskflwctl status [flags]
```

### Examples

```
  tskflwctl status
  tskflwctl status --json
  tskflwctl status --all
  tskflwctl status --all --json
```

### Options

```
      --all    summarize every logical planning space in the registry
  -h, --help   help for status
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

