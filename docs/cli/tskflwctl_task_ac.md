## tskflwctl task ac

List a task's acceptance criteria, or check/uncheck one by index

### Synopsis

List a task's acceptance criteria — the checkboxes under its `## Acceptance criteria` section — or flip one by 1-based index. Run with no flags (or --list) to number them, then --check <n> / --uncheck <n> to tick or clear one. Matching is index-based, not substring, for robustness. A flip rewrites only that one checkbox (the rest of the file is preserved), is atomic, and is idempotent — flipping to the current state writes nothing. Checkboxes in fenced code blocks are ignored, and a missing section or out-of-range index is a validation error (exit 11).

```
tskflwctl task ac <task> [flags]
```

### Examples

```
  tskflwctl task ac add-retry-backoff             # numbered list
  tskflwctl task ac add-retry-backoff --check 3   # tick criterion 3
  tskflwctl task ac add-retry-backoff --uncheck 3
```

### Options

```
      --check int     check the criterion at this 1-based index
  -h, --help          help for ac
      --list          list the acceptance criteria (the default)
      --uncheck int   uncheck the criterion at this 1-based index
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

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

