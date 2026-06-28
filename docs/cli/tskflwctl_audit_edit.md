## tskflwctl audit edit

Open an audit in your editor (whole file; re-validated on save)

### Synopsis

Open the audit's markdown file in $VISUAL/$EDITOR (falling back to vi). On save
the file is re-parsed: a frontmatter break reopens the editor with the error rather
than landing on disk. The findings are then lint-checked and any issues (bad
**Status:**, bucket↔state drift) are surfaced as a warning. The human counterpart
to `audit append` (scriptable).

```
tskflwctl audit edit <audit> [flags]
```

### Examples

```
  tskflwctl audit edit 2026-06-20-api-gateway
  tskflwctl audit edit   # pick from a list
```

### Options

```
  -h, --help   help for edit
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

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

