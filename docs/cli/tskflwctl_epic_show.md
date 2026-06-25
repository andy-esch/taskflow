## tskflwctl epic show

Show an epic and the tasks under it

```
tskflwctl epic show <epic> [flags]
```

### Examples

```
  tskflwctl epic show 01-api-gateway
  tskflwctl epic show   # pick from a list
```

### Options

```
  -h, --help   help for show
      --raw    print the raw markdown body (skip rendering)
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

* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics

