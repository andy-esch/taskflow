## tskflwctl research path

Print the absolute path to a research doc's file

```
tskflwctl research path <research> [flags]
```

### Examples

```
  tskflwctl research path tui-design-decisions
  $EDITOR "$(tskflwctl research path tui-design-decisions)"
```

### Options

```
  -h, --help   help for path
```

### Options inherited from parent commands

```
  -C, --chdir string         anchor to the planning repo at this path
      --color string         colorize output: auto|always|never (default "auto")
      --dry-run              preview the mutation without writing (validation still runs)
      --expect-root string   fail (exit 14) unless this directory resolves to this planning root — a wrong-repo write guard for agents
      --json                 machine-readable JSON output
      --no-color             disable colored output (alias for --color=never)
      --no-input             never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager             do not pipe long human output through a pager
      --paginate             page long human output through $PAGER (on a TTY), even if disabled in config
      --theme string         color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl research](tskflwctl_research.md)	 - Work with research docs

