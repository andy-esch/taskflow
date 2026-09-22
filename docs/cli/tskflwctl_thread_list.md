## tskflwctl thread list

List Threads with nominal and sound progress

```
tskflwctl thread list [flags]
```

### Examples

```
  tskflwctl thread list
  tskflwctl thread list -o table -c slug,status,done,total,frontier
  tskflwctl thread list --json -c id,slug,graph_health,projection_health
```

### Options

```
  -c, --columns strings   select columns for -o table/csv/json, comma-separated (-o table when no format is pinned); available: slug,status,done,total,drained,deprecated,frontier,graph_health,projection_health,inconsistent,description,id
  -h, --help              help for list
  -o, --output string     output format: human|json|name|table|csv
  -q, --quiet             concise command handles, one per line (alias for -o name)
      --status string     filter by Thread status
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

