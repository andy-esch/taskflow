## tskflwctl space forget

Drop an entry point from the registry (the repo itself is untouched)

### Synopsis

Remove an entry from the registry.

This never touches the repo on disk — forgetting is a registry edit, not a
deletion, so a space can always be re-added with `space add`.

```
tskflwctl space forget <id> [flags]
```

### Examples

```
  tskflwctl space forget old-thing
```

### Options

```
  -h, --help   help for forget
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

