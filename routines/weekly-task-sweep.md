---
name: weekly-task-sweep
version: 1
schedule: "0 10 * * 0"           # 6am EDT Sundays / 5am EST after Nov DST
slack_channel: planning-updates
repos:
  - andy-esch/taskflow           # write — the only repo; code + planning live together
working_dirs:
  write: ./taskflow
permissions:
  unrestricted_branch_pushes: taskflow only
allowed_tools: [Bash, Read, Write, Edit, Glob, Grep]
connectors:
  - slack
sweep_size: 5                    # tasks audited per run
cooldown_days: 7                 # skip tasks audited more recently, unless their code moved
last_modified: 2026-08-30
---

# Weekly Task Sweep

## Purpose

Every Sunday, pick up to 5 pending tasks (`next-up` or `ready-to-start`),
re-verify them against the current state of `internal/` and `cmd/`, refresh
frontmatter and body where the work has drifted or partially shipped, and open a
single PR with the changes plus any cross-cutting findings.

The goal isn't aggressive cleanup — it's keeping the backlog honest without the
user having to manually audit it. **Deference over deletion** is the operating
principle.

taskflow self-hosts its planning, so the code under audit and the task files
describing it are the same working tree. That makes verification cheap and
overreach easy — see SAFE-EDITING RULES.

## Prompt

```
You are running the weekly-task-sweep routine on the taskflow repo. Be careful
with prior content. This repo self-hosts its own planning: `planning/` describes
work on `internal/` and `cmd/` in the same checkout.

=== SCOPE — read this before anything else ===

**You write to `planning/` and nothing else.** That is the whole rule; the rest
of this block is just its consequences.

Allowed writes:
  - `planning/tasks/*.md` — additive body edits plus `tskflwctl` field and
    `task ac` commands
  - `planning/meta/no-op-log.md`
  - the PR itself

Forbidden — no exceptions, not even a one-line fix you are certain about:
  - `internal/`, `cmd/`, any `*.go` file, `go.mod`, `justfile`
  - `docs/`, `README.md`, `CLAUDE.md`
  - `planning/adrs/`, `planning/epics/`
  - any task's `status`, `tier`, `priority`, or scope
  - `depends_on` on any task (graph-owned; only guarded `task depend` may touch it)

taskflow self-hosts its planning, so code and planning share one working tree
and nothing structurally stops you from editing source. **The discipline is
yours to hold.** A defect you are certain about is a finding with a
`**Recommendation:**`, never a commit. Proposing the fix IS the deliverable.

Before opening the PR you MUST verify you stayed in bounds:

    git status --porcelain

Every line must be under `planning/`. If anything else is listed — including a
file you touched and meant to revert — **STOP**, restore it
(`git checkout -- <path>`), re-run the check, and say so in the PR body. If you
cannot get the check clean, post to Slack and exit without a PR.

=== START ===

0. Build the tool from this checkout — it IS this repo:

       just build          # → ./bin/tskflwctl

   If `just` is unavailable: `go build -o bin/tskflwctl ./cmd/tskflwctl`.
   If neither succeeds, **STOP**: post a Slack message saying the build failed,
   quote the error, and exit without a PR. Every later step uses
   `./bin/tskflwctl`. There is no fallback tool.

1. Read `CLAUDE.md` and `docs/ARCHITECTURE.md`. Skim
   `./bin/tskflwctl epic list` — epic status carries the committed project
   context (what's active vs retired) that informs tier/priority judgment in
   step 8. Skim `planning/adrs/` filenames so you recognise when a task's
   premise is governed by an accepted ADR.

2. List candidates:

       ./bin/tskflwctl task list --json -c slug,status,tier,priority,updated_at

   Filter to `status ∈ {next-up, ready-to-start}`. Then pull the full records
   for those (`task list --json`, no `-c`) so you have `audited`, `created`,
   `tags`, `depends_on`, and `audit_sources`.

=== DISCOVERY ===

3. Build a churn map of `internal/` and `cmd/` keyed by directory and the date
   of that directory's most recent commit, over the last 90 days (git log is
   newest-first, so the first time a dir appears is its latest touch):

       git log --since="90 days ago" --name-only --pretty=format:'%cs' \
         -- internal cmd docs \
         | awk '/^[0-9]{4}-[0-9]{2}-[0-9]{2}$/{d=$0; next} \
                NF{p=$0; sub(/\/[^/]*$/,"",p); \
                   if(!(p in seen)){print d"\t"p; seen[p]=1}}'

   The point is not "which dirs are hot" but "did anything a task cites move
   *after* that task was last audited" — that is the only churn that yields new
   information.

4. **Cooldown gate, then score.** Two stages — the gate decides who is eligible,
   the score orders who survives:

   (a) COOLDOWN GATE. Drop — do not score — any task whose `audited:` is within
       the last 7 days UNLESS one of these is true:
         - a directory it cites has a commit in the churn map dated *after* that
           task's `audited:`; or
         - its derived dependency state changed (step 4c).
       Never gate a task with `audited:` missing.
       Rationale: a task audited <7 days ago against code that hasn't moved is
       guaranteed to produce zero new findings on re-audit. The escape hatches
       keep genuinely-drifting tasks eligible.

   (b) SCORE survivors. **Staleness is the dominant axis** — the other bonuses
       only break ties among tasks of similar age; they can never let a fresh
       task leapfrog a much staler one:
         - staleness: +1 for every full 7 days since `audited:` (since `created:`
           if `audited:` is absent), capped at +9. Treat a missing `audited:`
           as the +9 cap.
         - +2 if a directory the task cites changed *after* its `audited:` date
           (real drift → high information gain). For never-audited tasks, +2 if
           it cites a dir touched in the last 30 days.
         - +2 if the task's dependency gate flipped since `audited:` (step 4c) —
           this is the taskflow-native signal and it is worth as much as code
           churn: a task that just became startable, or just became blocked, is
           exactly the one whose framing has moved.
         - +1 if tier ∈ {1, 2}
         - +1 if `priority: high`

       The non-staleness bonuses sum to at most +6, so any task ≥9 weeks stale
       (or never audited) still outranks a maximally-hot fresh one. The budget
       flows to the long tail, not to re-confirming the same few hot tasks.

   (c) Dependency-state check. For each surviving candidate:

           ./bin/tskflwctl task blockers <slug> --json

       Record `state.role`, `state.gate`, and `state.eligible`. Compare against
       what the task body claims about its prerequisites. A task whose body says
       "blocked on X" while the graph now reports `gate: clear` — or the reverse
       — has drifted and scores the +2 above. Also note whether the task appears
       in `./bin/tskflwctl task list --unblocked -q`.

5. Pick the top 5 by score; tie-break by oldest `audited:` (or `created:` if
   absent). If the cooldown gate leaves fewer than 5 eligible tasks, audit only
   those — do NOT pad the list with recently-verified tasks to hit 5. Note the
   short count in the PR body's TL;DR.

=== AUDIT (per task, medium depth) ===

6. Read the full task body and every file/line it cites in `internal/` or
   `cmd/`. For tasks about CLI surface, also check `docs/cli/` and the goldens
   under `internal/cli/testdata/golden/` — those are the shipped contract.

7. For each cited path/line: verify it still exists and the surrounding code
   matches what the task assumes. A task citing a symbol that was renamed, a
   test that was deleted, or a schema version that has since moved is drift
   worth recording.

8. Refresh frontmatter through the CLI — never hand-edit:

       ./bin/tskflwctl task set <slug> --description "..." --effort "..." ...
       ./bin/tskflwctl task set <slug> --set audited=YYYY-MM-DD

   - Add `description` if missing (required for `next-up`; lint enforces it).
   - Tighten `description` if >150 chars.
   - Adjust `effort` if scope has clearly shrunk or grown.
   - Adjust `tier` / `priority` only if the urgency calculus has materially
     changed; otherwise leave alone.
   - Append the sweep to `audit_sources:` only if this sweep produced a finding
     against that task.
   - Always bump `audited:` to today for every task you actually read.

9. Refresh the body conservatively per the SAFE-EDITING RULES below, using
   `task append <slug> --body "..."` for new sections and
   `task set --body-file` only when a whole-
   body rewrite is genuinely warranted (it rarely is).

=== SAFE-EDITING RULES — read carefully ===

- **Prefer additive edits.** Annotate sections as ✅ done, narrow scope, mark
  refs as "verified accurate (YYYY-MM-DD)" rather than deleting. Deletion is the
  last resort, not the default.
- **Never delete** these sections/fields:
  - `## Objective`, `## Context`, `## Origin`, `## Why` (or equivalent intro)
  - `audit_sources:` frontmatter
  - Prior `## Progress Log` / `## Implementation progress` entries (add new ones)
  - Any unresolved `- [ ]` decision checkbox (you may add a **Lean:** line below
    it if code state suggests an answer)
  - Any `## Open Questions` / `## Stress tests` / `## Sequencing` section
- **Acceptance criteria are owned by `task ac`.** Never hand-edit a `- [ ]` /
  `- [x]` line. If a criterion is now demonstrably met, use
  `./bin/tskflwctl task ac <slug> --check <n>`; if it has been decided rather
  than done, use `--defer|--wontfix|--tracked|--na <n> --reason "<why>"`. If you
  are <90% confident, leave the box alone and surface it as a finding.
- **Annotate, don't rewrite, embedded ⚠️ AUDIT blocks.** If a prior audit block
  is contradicted by current code, add a blockquote immediately below it:

  > **Update YYYY-MM-DD**: <what changed since the original audit>

  Leave the original intact so the reasoning history stays legible in git log.
- **Confidence threshold.** If you are <70% confident a section is obsolete,
  KEEP IT and surface the suspicion in the PR body's Findings section instead.
  Over-report; don't silently overwrite.
- **Lifecycle changes go through the CLI.** `task complete|deprecate|next|
  ready|defer`. Never `mv` a file or edit `status:` by hand — the store rejects
  it anyway, on both the `set` and `edit` paths.
  - **Do not `task complete` autonomously.** Completion is a human call; propose
    it in the PR body under "Possibly completable". `deprecate`/`defer` likewise.
- **Never touch `depends_on`.** It is graph-owned. Only the guarded
  `task depend add|remove` commands may change it, and this routine does not run
  them. A wrong edge is a finding.
- **Schema validity.** Run `./bin/tskflwctl lint` after editing tasks. Fix any
  error before moving on.

=== FINDINGS (the "what else did I notice") ===

While auditing, collect into a Findings list anything that is:

  (a) **Cross-cutting** — affects more than the task you're editing. Example:
      "`internal/core/dependency_graph.go` grew 980→1263 lines since 2026-08-12;
      three tasks reference its pre-Thread shape."
  (b) **Framing-shifting** — the underlying problem the task addresses has
      moved. Don't auto-rescope; surface for human review.
  (c) **Graph-shaped** — a task whose `depends_on` no longer matches its prose,
      a prerequisite that is `deprecated` (which makes the gate *broken*, not
      blocked), or a task that has silently become eligible. These are unique to
      this repo and worth calling out by name.
  (d) **Candidate new tasks** or **candidate deprecations** — bugs or
      inconsistencies you spotted that no task covers, or tasks whose problem
      appears resolved. Flag; don't create or close.

Be specific. "Lots of churn in core" is useless. "`task_lifecycle.go` has 6
commits in 30 days, all in the eligibility path; tasks A, B, C assume the
pre-1.55 candidate-only contract" is useful.

=== CLOSE ===

10. Add a one-line entry to each modified task's progress section:
    "YYYY-MM-DD: automated weekly sweep — <one-line summary>", via
    `./bin/tskflwctl task append <slug> --body "..."` (content goes in
    --body, not a positional).

11. Validate:

        git status --porcelain         # MUST list only planning/ paths
        just build && just test        # the sweep must not have broken anything
        ./bin/tskflwctl lint           # planning tree + dependency links

    `just test` runs the race detector. If it was already red before your edits,
    say so in the PR rather than trying to fix it — that's out of scope.

12. If NO task needed updating in any way (rare, but possible), do NOT open a
    PR. Append one line to `planning/meta/no-op-log.md` (date, `task-sweep`,
    "all audited tasks verified accurate"), post the SLACK-NOOP-TEMPLATE line,
    and exit.

13. Otherwise, open a PR against `main`:
    Title: "chore(planning): weekly task sweep — YYYY-MM-DD"
    Body: PR-BODY-TEMPLATE below.

14. Post to Slack #planning-updates using SLACK-MESSAGE-TEMPLATE below —
    verbatim. The PR URL MUST be wrapped in Slack's `<url|label>` angle-bracket
    syntax; a bare URL line absorbs the following text into the link and breaks
    it.

If anything fails (lint error you can't resolve, ambiguous code, unexpected repo
state), STOP, post a Slack message describing what happened, and exit without
opening a PR. Partial PRs are worse than no PR.

=== PR-BODY-TEMPLATE ===

  ## TL;DR

  <one or two sentences. If fewer than 5 tasks were eligible, say so and why.>

  ## Tasks changed

  - `<slug>` — <one line: what changed and why>
  - `<slug>` — ...

  ## Dependency-graph observations

  (Tasks whose derived state moved since last audit — became eligible, became
  blocked, or whose prose disagrees with `depends_on`. Omit if none.)

  - `<slug>` — body says "blocked on X"; graph reports `gate: clear` since <date>
  - ...

  ## Findings (cross-cutting, not actioned)

  - <specific observation with file:line evidence>
  - ...

  ## Suggested follow-ups (human review)

  - Possibly completable: `<slug>` — <evidence>
  - Possibly deprecatable: `<slug>` — <evidence>
  - Possibly worth promoting to next-up: `<slug>` — <evidence>

  ## Open questions noticed but not resolved

  - `<slug>` §<N>: "<question>" — <current code state suggests an answer but I
    deferred to keep the decision with the human>

=== SLACK-MESSAGE-TEMPLATE ===

  Post this exact structure to #planning-updates. Substitute the bracketed
  values; keep the literal `*…*` bold markers and the `<url|label>` angle
  brackets around the PR link:

  :robot_face: *taskflow weekly task sweep — YYYY-MM-DD*
  :link: <https://github.com/andy-esch/taskflow/pull/NN|PR #NN>
  *TL;DR:* <one sentence>

  *Counts:* <N> refreshed, <M> flagged, <K> graph-drift

  Notes:
  - The `<URL|PR #NN>` form fully delimits the link, so the TL;DR text that
    follows can never be absorbed into the URL. Do not fall back to a bare
    `:link: https://…` line.
  - Keep it to these four lines. If you want to surface one impactful finding
    inline, append it as a fifth line prefixed with `:mag:`.

=== SLACK-NOOP-TEMPLATE ===

  :robot_face: *taskflow weekly task sweep — YYYY-MM-DD* — no changes
  All audited tasks verified accurate against current code; no PR opened.
```

---

*Bootloader, configuration, tuning knobs, and change history for this routine
live in [`weekly-task-sweep.NOTES.md`](weekly-task-sweep.NOTES.md)
(human-maintainer reference, not read during a run).*
