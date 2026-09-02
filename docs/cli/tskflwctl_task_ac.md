## tskflwctl task ac

List a task's acceptance criteria, or check/uncheck one by index

### Synopsis

List a task's acceptance criteria — the checkboxes under its `## Acceptance criteria` section — or flip them by 1-based index. Run with no flags (or --list) to number them, then --check <n> / --uncheck <n> to tick or clear them. Both take a LIST: --check 1,2,4 (or --check 1 --check 2) flips all of them in ONE atomic write, so closing out several criteria costs one file rewrite and one updated_at bump rather than one per criterion. Indices are deduplicated and order-independent. Matching is index-based, not substring, for robustness. A flip rewrites only those checkboxes (the rest of the file is preserved) and is idempotent — flipping to the current state writes nothing. Checkboxes in fenced code blocks are ignored, and a missing section or an out-of-range index is a validation error (exit 11) that rejects the whole request before writing, so a bad index never leaves a half-applied body.

The criteria themselves can be edited too: --add <text> appends one, --remove <n> deletes one, and --replace <n> --text <new> rewords one. A reworded criterion KEEPS its checkbox and any state suffix — rewording is not a change of mind, and silently dropping a `wontfix` and its reason would lose a decision. Added and reworded text is wrapped to match the corpus. --add needs an existing `## Acceptance criteria` section: creating one would mean guessing where it belongs in a body the tool did not write.

```
tskflwctl task ac <task> [flags]
```

### Examples

```
  tskflwctl task ac add-retry-backoff               # numbered list
  tskflwctl task ac add-retry-backoff --check 3     # tick criterion 3
  tskflwctl task ac add-retry-backoff --check 1,2,4 # tick three, one atomic write
  tskflwctl task ac add-retry-backoff --uncheck 1,3
  tskflwctl task ac add-retry-backoff --defer 2 --reason "waiting on the schema ADR"
  tskflwctl task ac add-retry-backoff --add "Retries stop at the configured ceiling"
  tskflwctl task ac add-retry-backoff --replace 3 --text "Backoff is jittered"
  tskflwctl task ac add-retry-backoff --remove 4
```

### Options

```
      --add text          append a new unchecked criterion with this text
      --check indices     check the criteria at these 1-based indices (comma-separated or repeatable)
      --defer int         mark the criterion at this 1-based index deferred (needs --reason)
  -h, --help              help for ac
      --list              list the acceptance criteria (the default)
      --na int            mark the criterion at this 1-based index n/a — no longer applies (needs --reason)
      --reason string     why the criterion is deferred/wontfix/tracked/n-a — required for those, rejected otherwise
      --remove int        delete the criterion at this 1-based index
      --replace int       reword the criterion at this 1-based index (needs --text)
      --text string       the new wording for --replace
      --tracked int       mark the criterion at this 1-based index tracked — handed to another task (needs --reason naming it)
      --uncheck indices   uncheck the criteria at these 1-based indices (comma-separated or repeatable)
      --wontfix int       mark the criterion at this 1-based index wontfix (needs --reason)
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

* [tskflwctl task](tskflwctl_task.md)	 - Work with tasks

