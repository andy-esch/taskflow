## tskflwctl audit findings

Query findings across audits (or one) by status/effort/urgency/component

### Synopsis

Search audit findings — the structured per-finding view, not the aggregate.
With no argument, searches every audit; with an audit slug, just that one.
status/effort/urgency match exactly (case-insensitive, comma = any-of);
--component is a case-insensitive substring. Each --json hit carries its
audit slug and bucket.

```
tskflwctl audit findings [audit] [flags]
```

### Examples

```
  tskflwctl audit findings --status open --effort XS,S --json
  tskflwctl audit findings 2026-06-14-simplify-apigateway --status in-progress
  tskflwctl audit findings --component stravapipe -o table
```

### Options

```
  -c, --columns strings        select columns for -o table/csv/json, comma-separated (-o table when no format is pinned); available: ref,code,audit,status,effort,urgency,component,file,title
      --component string       filter by component (case-insensitive substring)
      --cursor string          resume page.next_cursor with the same filters, sort, and limit; restart if results changed
      --effort strings         filter by effort XS,S,M,L (any-of)
  -h, --help                   help for findings
      --limit int              maximum returned rows (1-1000; omitted means unbounded)
  -o, --output string          output format: human|json|name|table|csv
  -q, --quiet                  concise command handles, one per line (alias for -o name)
      --search string          case-insensitive substring in list metadata, not full bodies (maximum 256 UTF-8 bytes)
      --sort string            sort by id|name|created|updated; prefix with - for descending (findings: id|name only)
      --status strings         filter by finding status (comma-separated, any-of)
      --updated-since string   unsupported for findings (no own update date); requesting it is a validation error
      --urgency strings        filter by urgency acute,soon,eventually (any-of)
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

