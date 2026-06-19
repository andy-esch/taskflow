## tskflwctl audit list

List audits (open by default)

```
tskflwctl audit list [flags]
```

### Examples

```
  tskflwctl audit list
  tskflwctl audit list --all -o table -c slug,open
  tskflwctl audit list --closed -o json
```

### Options

```
      --all               all buckets
      --closed            closed audits only
  -c, --columns strings   select columns for -o table/csv, comma-separated (implies -o table); available: slug,bucket,area,date,findings,open
      --deferred          deferred audits only
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

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

