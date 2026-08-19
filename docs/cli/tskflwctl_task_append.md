## tskflwctl task append

Append a section to a task's body (atomic; agent-facing)

### Synopsis

Append markdown to the end of a task's body in one atomic, validated write —
the scriptable counterpart to `task edit`. Content comes from --body, --body-file,
or stdin (--body-file -); a blank line separates it from the existing body.

```
tskflwctl task append <task> [flags]
```

### Examples

```
  tskflwctl task append my-task --body 'a one-line note'
  printf '## Review\n- looks good\n' | tskflwctl task append my-task --body-file -
```

### Options

```
      --body string        markdown to append
      --body-file string   read the markdown to append from a file (or - for stdin)
  -h, --help               help for append
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

