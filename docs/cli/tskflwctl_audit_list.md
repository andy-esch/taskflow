## tskflwctl audit list

List audits (open by default)

### Synopsis

List audits with a segmented progress bar per row.

The headline number is the SETTLED share — findings that have reached a terminal
disposition, however they got there — so 100% is exactly the point an open audit
becomes `✔ ready to close`. The bar says how it settled, grouping the seven statuses
into four bands so the shape reads at a glance:

  █ green   settled here         fixed · tracked
  ▓ yellow  still being worked   in-progress
  ▒ gray    settled by dropping  deferred · superseded · wontfix
  ░ dim     still open           open

The glyphs differ as well as the colors, so the bands survive --color=never.

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
  -c, --columns strings   select columns for -o table/csv/json, comma-separated (implies -o table); available: slug,bucket,area,date,findings,open
      --deferred          deferred audits only
  -h, --help              help for list
  -o, --output string     output format: human|json|name|table|csv
  -q, --quiet             ids only, one per line (alias for -o name)
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

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

