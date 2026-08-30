## tskflwctl research

Work with research docs

### Options

```
  -h, --help   help for research
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

* [tskflwctl](tskflwctl.md)	 - Local-first planning CLI (tasks, Threads, epics, audits, research) over markdown
* [tskflwctl research append](tskflwctl_research_append.md)	 - Append a section to a research doc's body (atomic; agent-facing)
* [tskflwctl research edit](tskflwctl_research_edit.md)	 - Open a research doc in your editor (whole file; re-validated on save)
* [tskflwctl research list](tskflwctl_research_list.md)	 - List research docs (newest first)
* [tskflwctl research new](tskflwctl_research_new.md)	 - Create a new research doc
* [tskflwctl research path](tskflwctl_research_path.md)	 - Print the absolute path to a research doc's file
* [tskflwctl research set](tskflwctl_research_set.md)	 - Set one or more frontmatter fields (validated, single atomic write)
* [tskflwctl research show](tskflwctl_research_show.md)	 - Show a research doc's metadata and body

