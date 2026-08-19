## tskflwctl audit lint

Validate audit findings (status vocabulary, missing status, bucket↔state)

### Synopsis

Lint audit findings — the audit analog of `lint` (which covers tasks/epics).
Checks every finding has a legal **Status:** (catching typos a free-text edit
allows) and that a non-open audit has no still-open findings. With no argument
it lints every audit; with a slug, just that one. Exit 11 when issues are found.

```
tskflwctl audit lint [audit] [flags]
```

### Examples

```
  tskflwctl audit lint
  tskflwctl audit lint 2026-06-14-gateway --json
```

### Options

```
  -h, --help   help for lint
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
      --theme string   color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

