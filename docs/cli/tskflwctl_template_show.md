## tskflwctl template show

Show a template's rendered body (name defaults to "default")

```
tskflwctl template show <kind> [name] [flags]
```

### Examples

```
  tskflwctl template show audit security
  tskflwctl template show task --json
```

### Options

```
  -h, --help   help for show
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

* [tskflwctl template](tskflwctl_template.md)	 - List and inspect the body scaffolds `new --template` can use

