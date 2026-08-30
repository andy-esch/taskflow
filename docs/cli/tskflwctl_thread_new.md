## tskflwctl thread new

Create an unstarted Thread with optional initial task members

### Synopsis

Create one unstarted Thread. Repeat --task to add initial members; every reference is resolved and validated with current tasks and Threads under the repository guard.

```
tskflwctl thread new <title> [flags]
```

### Examples

```
  tskflwctl thread new "Thread delivery" --description "Ship production Threads" --goal "Dogfood the remaining implementation"
  tskflwctl thread new "Thread delivery" --description "Ship production Threads" --goal "Dogfood it" --task documents --task lifecycle
```

### Options

```
      --body string          replace the default body scaffold
      --body-file string     read body from a file, or - for stdin
      --description string   one-line description (<=200 chars; required)
      --goal string          one-line observable finish line (required)
  -h, --help                 help for new
      --tags strings         comma-separated tags
      --target-date string   optional human planning target (YYYY-MM-DD)
      --task stringArray     initial task reference (repeatable)
      --template string      body template name (default: default)
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

* [tskflwctl thread](tskflwctl_thread.md)	 - Work with initiative Threads over the task DAG

