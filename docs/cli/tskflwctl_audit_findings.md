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
  -c, --columns strings    select columns for -o table/csv/json, comma-separated (implies -o table); available: ref,code,audit,status,effort,urgency,component,file,title
      --component string   filter by component (case-insensitive substring)
      --effort strings     filter by effort XS,S,M,L (any-of)
  -h, --help               help for findings
  -o, --output string      output format: human|json|name|table|csv
  -q, --quiet              ids only, one per line (alias for -o name)
      --status strings     filter by finding status (comma-separated, any-of)
      --urgency strings    filter by urgency acute,soon,eventually (any-of)
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

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

