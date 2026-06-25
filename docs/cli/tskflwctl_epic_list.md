## tskflwctl epic list

List epics with task rollup

```
tskflwctl epic list [flags]
```

### Examples

```
  tskflwctl epic list
  tskflwctl epic list --status active
  tskflwctl epic list -o table -c id,status,percent,description
```

### Options

```
  -c, --columns strings   select columns for -o table/csv/json, comma-separated (implies -o table); available: id,status,priority,done,total,description,percent,deprecated
  -h, --help              help for list
  -o, --output string     output format: human|json|name|table|csv
  -q, --quiet             ids only, one per line (alias for -o name)
      --status string     filter by epic status (active|retired|deprecated)
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager       do not pipe long human output through a pager
      --paginate       page long human output through $PAGER (on a TTY), even if disabled in config
```

### SEE ALSO

* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics

