# routines/

Versioned specs for scheduled Claude Code routines that act on this repo.

Modelled on the same pattern used in
[`desirelines-planning/routines/`](https://github.com/andy-esch/desirelines-planning/tree/main/routines),
adapted for taskflow's single-repo, self-hosting layout.

## Why this exists

Claude Code routines are configured at [claude.ai/code/routines](https://claude.ai/code/routines),
but configuring everything there means:

- The prompt isn't reviewable in PRs.
- You can't tell what a routine does without logging into the web UI.
- Tweaks have no history.

This directory holds the **source-of-truth spec** for each routine. The routine
in claude.ai runs a small bootloader prompt that reads the spec from this repo
and executes it. Changes to behavior go through a normal PR; the routine itself
only needs to be touched if you rename a file or change the bootloader contract.

## Layout

Each routine gets a spec file `routines/<name>.md` — frontmatter (schedule,
rotation, channel, version) plus the full prompt the bootloader will execute.
This is the only file the agent reads at run time, so it stays lean.

Human-maintainer content the agent never executes — the bootloader prompt,
operational notes, tuning knobs, open design questions, and the changelog —
lives in a sidecar `routines/<name>.NOTES.md`. Keeping it out of the spec means
it isn't loaded into agent context on every run.

## Updating a routine

1. Open a PR that edits `routines/<name>.md`.
2. Bump `version` and add a Changelog entry in `routines/<name>.NOTES.md`.
3. Once merged to `main`, the next scheduled run picks up the new spec
   automatically (the bootloader clones `main` fresh each run).

Changing the bootloader itself — schedule, repo list, permissions, connectors —
is a manual update at [claude.ai/code/routines](https://claude.ai/code/routines).
Note the change here too.

## How this differs from desirelines-planning

The desirelines routines coordinate **four repos**: planning is a separate repo
from the code under audit, so read-only access is enforced structurally and a
`scripts/install-tskflwctl.sh` Step 0 is needed to get the tool into the sandbox.

taskflow is **one repo that self-hosts its own planning**. That collapses two
things and creates one hazard:

- **No install script.** `tskflwctl` *is* this repo. Step 0 is `just build`
  (or `go build -o bin/tskflwctl ./cmd/tskflwctl`), and every later step calls
  `./bin/tskflwctl`. The routine should run the binary this repo's `main`
  produces, which is why it builds rather than installs. taskflow is public, so
  `go install github.com/andy-esch/taskflow/cmd/tskflwctl@latest` also works
  with no token — that's the right call for *other* repos' routines, not this
  one's.
- **No separate planning repo.** The audit file, the code it audits, and the
  tasks it cross-references are all in the same working tree, so there is one
  PR per run instead of a planning PR referencing another repo's SHAs.
- **⚠ Propose-only is a behavioural rule here, not a structural one.** In
  desirelines the agent physically cannot write to the code it audits. Here it
  can. Every spec therefore opens with a SCOPE block stating one rule —
  **write to `planning/` and nothing else** — enumerating what that forbids
  (`internal/`, `cmd/`, any `*.go`, `docs/`, `justfile`, `planning/adrs/`,
  `planning/epics/`, and any task's status/tier/priority/scope), and requiring a
  `git status --porcelain` bounds check before the PR: every line must be under
  `planning/`, or the run stops. Proposing the fix IS the deliverable — a
  routine never commits one, however certain it is.

## `tskflwctl` conventions these routines rely on

The tool owns its own planning metadata; routines must go through it rather
than hand-editing markdown. The rules that matter most at run time:

- **Audit files are tool-managed.** `audit new` mints the id and frontmatter;
  `audit show|list|findings|append|close` read and move them. Never hand-build
  a path or a `bucket:` value.
- **Never hand-edit a finding's `**Status:**` or `**Resolution:**`.**
  `tskflwctl audit finding <audit> <code> --status <v> [--note <text>]` writes
  both in one validated, atomic edit. Statuses: `open | in-progress | fixed |
  tracked | deferred | superseded | wontfix`. **`tracked` requires the
  destination** (`tracked by <task-id>`) — this is where taskflow diverges from
  the older desirelines convention of writing `superseded by <link>` by hand.
- **Never hand-edit task `status:`.** Lifecycle verbs (`task start|next|ready|
  complete|defer|deprecate`) own it and edit frontmatter in place.
- **Two validators, both required.** `tskflwctl lint` checks the entity tree and
  dependency links; `tskflwctl audit lint <slug>` checks *finding* status
  vocabulary. The first does not cover the second.
- Machine reads are cheapest as `--json -c <cols>`; reach for full `--json`
  only when you need every frontmatter field.

Full authoring guidance is self-describing: `tskflwctl schema`,
`tskflwctl schema task|audit|thread|epic|research`.

## Routines

| File | Schedule | Purpose |
|------|----------|---------|
| [weekly-task-sweep.md](weekly-task-sweep.md) | 6am EDT Sundays | Re-verify 5 pending tasks against current code and the dependency graph; refresh frontmatter/body additively; open a PR with cross-cutting findings. |
| [code-quality-audit.md](code-quality-audit.md) | 6am EDT Tuesdays + Fridays | Targeted code audit through one of six rotating **lenses** (correctness · concurrency · contract · test-rigour · adapter-hygiene · simplification). Signal + adjacency + random file picks. Output: `planning/audits/<id>-YYYY-MM-DD-<lens>.md` + PR. |
| [weekly-architecture-audit.md](weekly-architecture-audit.md) | 7:30am EDT Mondays | Single-lens architectural review per week, rotating on ISO week mod 4 (hexagonal-boundaries / data-model-and-storage / machine-contract / failure-and-recovery). ADR-reconciling, research-mandatory with citations. Output: `planning/audits/<id>-YYYY-MM-DD-arch-<lens>.md` + PR. |

Schedules are deliberately spread across three different days so no two routines
share a usage window.

## Shared conventions

- **Backpressure.** The two audit routines check `tskflwctl audit list --json`
  first. At **10 or more** open audits they stop authoring and spend the run
  triaging the oldest three instead. An audit nobody reads costs a run, adds a
  file, and re-finds what an earlier audit already found.
- **No-op log.** A run that finds nothing does *not* open a PR. It appends one
  line to [`planning/meta/no-op-log.md`](../planning/meta/no-op-log.md) and posts
  the clean-run Slack line. "I found nothing" is a real result.
- **Fail loudly.** Any failure the agent can't resolve — build error, red lint,
  unexpected repo state — means STOP, post to Slack, and exit **without** a PR.
  A half-finished audit is worse than none.
