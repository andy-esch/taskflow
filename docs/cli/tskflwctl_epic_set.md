## tskflwctl epic set

Set one or more epic frontmatter fields (validated, single atomic write)

```
tskflwctl epic set <epic> [flags]
```

### Examples

```
  tskflwctl epic set 20-cli-ux --priority high
  tskflwctl epic set --priority high   # pick the epic from a list
```

### Options

```
      --description string   one-line description (<=200 chars)
      --force                allow --set of a field tskflwctl doesn't know
  -h, --help                 help for set
      --priority string      high|medium|low
      --set stringArray      key=value (repeatable); known fields are typed+validated, unknown keys need --force
      --tags strings         comma-separated tags
      --unset stringArray    remove a frontmatter key (repeatable)
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

* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics

