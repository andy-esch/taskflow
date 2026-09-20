## tskflwctl audit finding new

Create one canonical audit finding and allocate its code

### Synopsis

Create one open finding without hand-writing its code or Markdown grammar.

The audit itself must be open; reopen a closed or deferred audit before adding new
work so creation cannot make its bucket and finding state disagree.

--band is required and accepts H, M, or L. Allocation is monotonic within that
band: the highest existing suffix plus one, recomputed after every concurrent-write
retry so a stale caller cannot duplicate an identity. Optional file, component,
effort, urgency, body, and recommendation values are rendered into the canonical
block. --candidate adds the new finding's row to a managed candidate-tasks:v1
section in the same atomic write; legacy sections are refused without a partial
finding. Existing malformed findings or managed candidate projections also refuse
creation; inspect them with `audit lint` and use `audit edit` for explicit repair.
The compact JSON receipt returns the allocated code rather than the full body.

```
tskflwctl audit finding new <audit> <title> [flags]
```

### Examples

```
  tskflwctl audit finding new 2026-09-20-storage "Retry loop can lose evidence" --band H --component store --effort S --urgency soon --body "A conflicting write replaces the prior snapshot." --recommendation "Recompute inside the CAS callback."
  tskflwctl audit finding new my-audit "Bound the query" --band M --body-file finding.md --candidate "Create a bounded-query task"
```

### Options

```
      --band string             required finding identity band: H | M | L
      --body string             finding evidence as Markdown (unfenced headings are refused)
      --body-file string        read finding evidence from a file, or - for stdin
      --candidate string        one-line managed Candidate tasks entry (requires candidate-tasks:v1)
      --component string        optional component or subsystem
      --effort string           optional effort: XS | S | M | L
      --file string             optional source location, e.g. internal/store/body.go:194
  -h, --help                    help for new
      --recommendation string   one-line minimum recommendation
      --urgency string          optional urgency: acute | soon | eventually
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

* [tskflwctl audit finding](tskflwctl_audit_finding.md)	 - Set one finding's status, resolution, and candidate row (validated, atomic)

