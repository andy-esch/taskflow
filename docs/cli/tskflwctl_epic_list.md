## tskflwctl epic list

List epics with task rollup

```
tskflwctl epic list [flags]
```

### Examples

```
  tskflwctl epic list
  tskflwctl epic list -o table
  tskflwctl epic list -o json
```

### Options

```
  -c, --columns strings   select columns for -o table/csv, comma-separated (implies -o table); available: id,status,priority,done,total,description
  -h, --help              help for list
  -o, --output string     output format: human|json|name|table|csv
  -q, --quiet             ids only, one per line (alias for -o name)
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
```

### SEE ALSO

* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics

