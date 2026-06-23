## tskflwctl template

List and inspect the body scaffolds `new --template` can use

### Options

```
  -h, --help   help for template
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, epics, audits) over markdown
* [tskflwctl template list](tskflwctl_template_list.md)	 - List available body templates (kind, name, description)
* [tskflwctl template show](tskflwctl_template_show.md)	 - Show a template's body (name defaults to "default"; --raw for the unrendered source)

