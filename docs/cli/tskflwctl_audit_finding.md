## tskflwctl audit finding

Set one finding's status and resolution note in place (validated, atomic)

### Synopsis

Stamp a finding's **Status:** and **Resolution:** without touching the rest of the audit.

The status is validated against the finding vocabulary, and only the leading token
is normalised — decoration the line formats carry (`fixed 2026-08-24 (PR #12)`,
`deferred (see ADR-0003)`, `superseded by <link>`) is written verbatim, because it
holds dates, links, and document names whose spelling is not the tool's to flatten.
`tracked` additionally REQUIRES a destination (`tracked by <task-id>`), so a finding
handed to a task always says where it went.

--note writes the `**Resolution:**` paragraph as the finding's last block: one
paragraph, no newlines, placed inside the right finding by construction rather than
by careful typing. Passing an empty --note removes it. Both flags REPLACE what was
there, and given together they land in a single atomic write.

--pr N is sugar for the canonical `(PR #N)` decoration, so the reference is spelled
one way across the corpus and stays greppable.

```
tskflwctl audit finding <audit> <code> [flags]
```

### Examples

```
  tskflwctl audit finding 2026-06-14-gateway H1 --status fixed
  tskflwctl audit finding 2026-06-14-gateway M2 --status "deferred (see ADR-0003)"
  tskflwctl audit finding 2026-06-14-gateway H1 --status "tracked by 6g392b0rps7w"
  tskflwctl audit finding 2026-06-14-gateway H1 --status fixed --note "Widened the regex; regression test added."
```

### Options

```
  -h, --help                   help for finding
      --note **Resolution:**   the finding's **Resolution:** paragraph — how it was resolved; empty removes it
      --pr (PR #N)             append (PR #N) to the status — the one canonical spelling, so the reference stays greppable
      --status string          the finding's new status — one of: deferred | fixed | in-progress | open | superseded | tracked | wontfix (decoration after the token is kept verbatim)
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

* [tskflwctl audit](tskflwctl_audit.md)	 - Work with code audits

