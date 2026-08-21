## tskflwctl research new

Create a new research doc

### Synopsis

Create a research doc: an exploration snapshot, true as of its date.

The id is minted from --created, so backdating a doc to when the work actually
happened sorts it into place chronologically. Research has no status and no
lifecycle verbs — a later doc supersedes an earlier one. A decision that needs a
lifecycle is an ADR, not research.

```
tskflwctl research new <title> [flags]
```

### Examples

```
  tskflwctl research new "Compare theming libraries" --tags tui,color
  tskflwctl research new "Storage model options" --created 2026-06-24
```

### Options

```
      --body string          override the default scaffold
      --body-file string     read the body from a file, or - for stdin (replaces --body)
      --created string       date the research was done, YYYY-MM-DD (default today); the id is minted from it
      --description string   one-line description (<=200 chars)
  -h, --help                 help for new
      --tags strings         comma-separated tags
      --template string      body scaffold to use (default "default"); completes the available names
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

