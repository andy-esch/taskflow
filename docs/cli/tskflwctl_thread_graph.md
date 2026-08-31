## tskflwctl thread graph

Export a deterministic Mermaid or DOT Thread graph

### Synopsis

Render Thread members, immediate external gates, and every dependency edge between those bounded nodes from the shared runtime projection. Mermaid is the default. Generated output is never persisted; --json emits the neutral projection instead of renderer text and cannot be combined with an explicit --format.

```
tskflwctl thread graph <thread> [flags]
```

### Options

```
      --format string   graph output format: mermaid|dot (default "mermaid")
  -h, --help            help for graph
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

* [tskflwctl thread](tskflwctl_thread.md)	 - Work with initiative Threads over the task DAG

