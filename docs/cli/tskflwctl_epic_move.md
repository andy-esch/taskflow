## tskflwctl epic move

Transition epic(s) to <status> (active|retired|deprecated)

```
tskflwctl epic move <epic>... <status> [flags]
```

### Examples

```
  tskflwctl epic move 18-tui retired
  tskflwctl epic move 18-tui 20-cli deprecated --dry-run
```

### Options

```
  -h, --help   help for move
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

