## tskflwctl task list

List tasks (active by default)

```
tskflwctl task list [flags]
```

### Examples

```
  tskflwctl task list
  tskflwctl task list -q --tag tui | xargs tskflwctl task start
  tskflwctl task list -o table -c slug,status,epic
```

### Options

```
      --all               include completed/deprecated/deferred
  -c, --columns strings   select columns for -o table/csv/json, comma-separated (implies -o table); available: slug,status,tier,priority,epic,updated,description
      --epic string       filter by epic
  -h, --help              help for list
  -o, --output string     output format: human|json|name|table|csv
  -q, --quiet             ids only, one per line (alias for -o name)
      --status string     filter by status
      --tag string        filter by tag
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
```

### SEE ALSO

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

