---
status: completed
epic: 17-pm-go-cli
description: "Wire cobra shell completion: install path for command/flag completion (Part A) + ValidArgsFunction slug completion for task/audit/epic (Part B)."
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: 2026-06-08
started_at: 2026-06-08
updated_at: 2026-06-08
completed_at: 2026-06-08
id: 6fa91vg02na1
---

# Shell autocomplete: command + dynamic task/audit/epic slug completion

## Objective

Make `tskflwctl` pleasant to drive by hand: TAB-completion for both the command
tree and the task/audit/epic **slug** arguments. Two separable concerns.

## Part A — command/flag completion (install path)

Cobra already generates this (`tskflwctl completion {bash,zsh,fish,…}`); the gap
is that the script isn't installed and it's undiscoverable. Low/zero code.

- [x] Verify the default `completion` command isn't clobbered; keep it enabled.
- [x] Add a `just completion-zsh` (and `just completion SHELL`) so the generated
      script lands on `$fpath`.
- [x] Document the one-time install in README.

## Part B — dynamic slug completion (`ValidArgsFunction`)

Wire slug completion on every slug-taking command:
- task: `show`, `set`, `move`, and the six verbs (`start/promote/demote/
  complete/defer/deprecate`)
- audit: `show`, `close`, `reopen`, `defer`
- epic: `show`

Design calls (deliberate — completion has different constraints):
- [x] **Complete from filenames, not parsed frontmatter.** Slug == filename
      stem, so glob `tasks/*/*.md` (audits `audits/*/*.md`, epics `epics/*.md`)
      without YAML parsing → fast (shell runs the binary per TAB) and immune to
      a malformed file (you can still complete the broken task you're fixing).
      `slugsFromGlob` in `cli/completion.go` — no `core`/store round-trip.
- [x] **Silent + repo-tolerant.** `__complete` *does* run `PersistentPreRunE`
      (confirmed — its "not a repo" error was aborting completion). Made the
      hook non-fatal under `isCompletionCommand`, and the completion funcs do
      their own forgiving `config.Discover` → outside a repo TAB is a quiet
      no-op (`NoFileComp`, exit 0), no stderr noise.
- [x] Three helpers (`completeTaskSlugs`/`completeAuditSlugs`/`completeEpicIDs`)
      so the pattern lives once (cf. `runMoves`).
- [x] **Dedup against already-typed args** — folded in early (nearly free via
      the `args` param); multi-slug `move`/`close` don't re-suggest typed slugs.

Fast-follow:
- [x] **Status-aware filtering:** `task start` offers only non-in-progress;
      audit `reopen` only non-open; etc. Resolved the "no parse" tension neatly
      — status/bucket *is* the directory, so the completer just globs the right
      dirs (excludes the target state's dir). Still zero YAML parsed.

Ride-along (cheap review-polish, done in the same pass):
- [x] `render.MoveResult` JSON key `"status"` → `"to"` (honest for both task
      status and audit bucket; pre-release, no shipped consumers).
- [x] `domain.ErrNotFound` is now generic ("not found"); the task/epic/audit
      stores prefix the noun (`task "x": not found`).

## Acceptance

- [x] `tskflwctl completion zsh` install documented + a `just` helper.
- [x] `task show <TAB>` / `audit close <TAB>` / `epic show <TAB>` complete real
      slugs; a malformed file doesn't break completion; outside a repo TAB is a
      quiet no-op.
- [x] Tests exercise the completion funcs (via `__complete` in-process).

## Progress Log

**2026-06-08 — Part A done (command/flag completion install path).** Confirmed
cobra's default `completion` command is live and not clobbered, so command +
flag completion already work via `__complete` (verified: top-level commands,
`task` subcommands, and `task list --` flags all complete *with descriptions*).
The only gap was install/discoverability — added `just completion-zsh` (writes
`~/.zsh/completions/_tskflwctl` + prints the one-time `fpath`/`compinit` setup)
and a generic `just completion SHELL` (bash|zsh|fish|powershell → stdout), plus
a README "Shell completion" section. No Go changes; build/test/lint untouched.
**Next: Part B** — `ValidArgsFunction` slug completion (filename-glob, silent +
repo-tolerant), starting with the task verbs.

**2026-06-08 — Part B done (dynamic slug completion).** New `cli/completion.go`:
`slugsFromGlob` (filename-stem glob, no YAML parse, prefix-filtered, drops
already-typed args) + `completeTaskSlugs`/`completeAuditSlugs`/`completeEpicIDs`,
wired as `ValidArgsFunction` on task `show`/`set`/`move` + the six transition
verbs, audit `show`/`close`/`reopen`/`defer`, and epic `show`. Discovered (and
fixed) that `__complete` runs `PersistentPreRunE`: its not-a-repo error was
killing completion, so the hook is now non-fatal under `isCompletionCommand`.
Verified end-to-end via `__complete`: real slugs with prefix-filter + dedup, a
**malformed** task still completes (no-parse design pays off), and outside a repo
TAB is a quiet no-op (exit 0, `NoFileComp`). Five in-process completion tests
added; full suite + vet + lint green. README updated.

Acceptance met. Remaining is the **status-aware filtering** fast-follow plus the
two ride-along polish items below (deferred — not started).

**2026-06-08 — status-aware filtering + ride-along polish (task done).**
Transition verbs now filter by destination: `task start` omits in-progress
tasks, `audit reopen` omits open audits, etc. — implemented by globbing all
state dirs *except* the target's (status/bucket == directory, so still no YAML
parse). Ride-alongs: `MoveResult` JSON key `status`→`to`; `ErrNotFound` made
generic with per-noun prefixes at the stores. Two status-aware completion tests
added; full suite + lint green. All acceptance criteria + fast-follow complete.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); follows the
  [port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec](6f9menr01nsd-port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec.md) work.
