## tskflwctl research set

Set one or more frontmatter fields (validated, single atomic write)

### Synopsis

Update a research doc's frontmatter in one atomic, validated write. Unknown keys,
comments, and key order are preserved.

`created` cannot be set: the stable id is minted from it, so changing one would
desync the pair and break the id-order-is-date-order property. Re-dating a doc
means creating a new one.

```
tskflwctl research set <research> [flags]
```

### Examples

```
  tskflwctl research set theming-libs --description "Weighed three TUI theming libs"
  tskflwctl research set theming-libs --tags tui,color
```

### Options

```
      --description string   one-line description (<=200 chars)
      --force                allow --set of a field tskflwctl doesn't know
  -h, --help                 help for set
      --set stringArray      key=value (repeatable); known fields are typed+validated, unknown keys need --force
      --tags strings         comma-separated tags
      --unset stringArray    remove a frontmatter key (repeatable)
```

### Options inherited from parent commands

```
  -C, --chdir string   anchor to the planning repo at this path (conflicts with --space)
      --color string   colorize output: auto|always|never (default "auto")
      --dry-run        preview the mutation without writing (validation still runs)
      --json           machine-readable JSON output
      --no-color       disable colored output (alias for --color=never)
      --no-input       never prompt; missing required input is an error (for scripts/agents; also TSKFLW_NO_INPUT)
      --no-pager       do not pipe long human output through a pager
      --paginate       page long human output through $PAGER (on a TTY), even if disabled in config
      --space string   select a registered entry point by label (also TSKFLW_SPACE; conflicts with -C)
      --theme string   color theme name (overrides TSKFLW_THEME and [theme].name in config)
```

### SEE ALSO

* [tskflwctl research](tskflwctl_research.md)	 - Work with research docs

