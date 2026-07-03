## tskflwctl lint

Validate active task and epic frontmatter (--fix repairs tasks/audits and assigns missing ids)

```
tskflwctl lint [flags]
```

### Examples

```
  tskflwctl lint
  tskflwctl lint --fix --dry-run
  tskflwctl lint --json
```

### Options

```
      --fix    auto-repair frontmatter: quote ':' values, normalize lists, realign task status, backfill missing task/audit ids; epics are text-only
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown

