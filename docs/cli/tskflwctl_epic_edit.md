## tskflwctl epic edit

Open an epic in your editor (whole file; re-validated on save)

### Synopsis

Open the epic's markdown file in $VISUAL/$EDITOR (falling back to vi). On
save the file is re-parsed: a frontmatter break (or a value the loader can't
read) reopens the editor with the error rather than landing on disk — deeper
field checks remain `lint`'s job. The human counterpart to `epic set`; agents
and scripts should drive `set` (deterministic) instead.

```
tskflwctl epic edit <epic> [flags]
```

### Examples

```
  tskflwctl epic edit 20-cli-ux
  tskflwctl epic edit   # pick from a list
```

### Options

```
  -h, --help   help for edit
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

* [tskflwctl epic](tskflwctl_epic.md)	 - Work with epics

