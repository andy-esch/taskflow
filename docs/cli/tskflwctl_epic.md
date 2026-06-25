## tskflwctl epic

Work with epics

### Options

```
  -h, --help   help for epic
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
* [tskflwctl epic edit](tskflwctl_epic_edit.md)	 - Open an epic in your editor (whole file; re-validated on save)
* [tskflwctl epic list](tskflwctl_epic_list.md)	 - List epics with task rollup
* [tskflwctl epic move](tskflwctl_epic_move.md)	 - Transition epic(s) to <status> (active|retired|deprecated)
* [tskflwctl epic new](tskflwctl_epic_new.md)	 - Create a new epic (auto-numbered NN-slug)
* [tskflwctl epic set](tskflwctl_epic_set.md)	 - Set one or more epic frontmatter fields (validated, single atomic write)
* [tskflwctl epic show](tskflwctl_epic_show.md)	 - Show an epic and the tasks under it

