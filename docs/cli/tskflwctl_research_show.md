## tskflwctl research show

Show a research doc's metadata and body

```
tskflwctl research show <research> [flags]
```

### Examples

```
  tskflwctl research show lipgloss-v2-charm-ecosystem
  tskflwctl research show tui-design-decisions --section findings
```

### Options

```
      --frontmatter-only   show only the metadata, no body
  -h, --help               help for show
      --raw                print the body as raw markdown (no styling)
      --section string     show only this body section (## heading, case-insensitive)
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

