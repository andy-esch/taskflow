## tskflwctl space list

List planning spaces, entry points, and current health

### Synopsis

List every registered planning entry point and diagnose it without changing the registry.

Registered paths that resolve to the same durable planning identity are grouped as
entry points. The first direct checkout anchors a group when one is registered;
indentation means shared planning data, not filesystem ownership.

Healthy repos are `ok` or `empty`. Missing paths, non-repos, unreadable configs,
and durable-id mismatches stay listed with a remedy; none is auto-forgotten.

```
tskflwctl space list [flags]
```

### Examples

```
  tskflwctl space list
  tskflwctl space list --json
```

### Options

```
  -h, --help   help for list
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

* [tskflwctl space](tskflwctl_space.md)	 - Manage planning spaces and their registered entry points

