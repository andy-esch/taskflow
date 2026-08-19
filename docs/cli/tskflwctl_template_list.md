## tskflwctl template list

List available body templates (kind, name, description)

```
tskflwctl template list [flags]
```

### Examples

```
  tskflwctl template list
  tskflwctl template list --kind audit --json
```

### Options

```
  -h, --help          help for list
      --kind string   restrict to one kind (task|epic|audit)
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

* [tskflwctl template](tskflwctl_template.md)	 - List and inspect the body scaffolds `new --template` can use

