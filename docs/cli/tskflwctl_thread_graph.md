## tskflwctl thread graph

Export a deterministic Mermaid or DOT Thread graph

### Synopsis

Render Thread members, immediate external gates, and every dependency edge between those bounded nodes from the shared runtime projection. Pass --around TASK to select the one-hop neighborhood around a member or external gate; --depth 2 widens it to two hops. The bounded output discloses omitted nodes and exact crossing edges. Compact labels prioritize the human task title, lifecycle state and role, then stable ID; adapters without a body-derived title fall back to a humanized slug. --details adds a bounded description. Mermaid is the default. Generated output is never persisted; --json emits the full or selected neutral projection instead of renderer text and cannot be combined with renderer flags.

```
tskflwctl thread graph <thread> [flags]
```

### Options

```
      --around string   show a bounded neighborhood around a member or external-gate task
      --depth int       neighborhood hop depth: 1|2 (requires --around) (default 1)
      --details         include a bounded task description in each graph node
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

