---
name: code-quality-audit
version: 1
schedule: "0 10 * * 2,5"         # 6am EDT Tuesdays + Fridays; lens picked by rotation index
slack_channel: planning-updates
repos:
  - andy-esch/taskflow           # write — the only repo; code + planning live together
working_dirs:
  write: ./taskflow
permissions:
  unrestricted_branch_pushes: taskflow only
allowed_tools: [Bash, Read, Write, Edit, Glob, Grep, WebFetch, WebSearch]
connectors:
  - slack
lens_rotation:                   # index = (ISO week * 2 + slot) mod 6; slot 0 = Tue, 1 = Fri
  0: correctness-and-errors
  1: concurrency-and-atomicity
  2: contract-and-compatibility
  3: test-rigour
  4: adapter-hygiene
  5: simplification
max_open_audits: 10              # backpressure: at or above this, triage instead of authoring
last_modified: 2026-08-30
---

# Code Quality Audit (Twice-Weekly Lens Rotation)

## Purpose

Targeted, propose-only code audit of this Go codebase through one **lens** per
run, rotating through six bins so the full set is covered every three weeks. The
lens — not a directory — is the unit of rotation: the same file can be worth
reading twice through different questions, and the interesting defects in this
codebase are cross-file (a guard in `store` protecting an invariant defined in
`core` and rendered in `cli`).

The agent picks ~4–5 files via signal (churn × blast radius), adjacency, and one
random pick, all constrained to the lens's surface; surfaces
severity/effort/urgency-ranked findings; cross-references open tasks; and opens a
PR. The human triages what enters the backlog.

This is deliberately distinct from `weekly-architecture-audit`, which asks
"is the system shaped right?" This one asks "is this code right?"

## Prompt

```
You are running the code-quality-audit routine on the taskflow repo. Be
specific, be honest, and prefer "I didn't find anything noteworthy" to
inventing mediocre findings.

taskflow self-hosts its own planning: `planning/` describes work on `internal/`
and `cmd/` in the same checkout. You are auditing code you can also write.

=== SCOPE — read this before anything else ===

**You write to `planning/` and nothing else.** That is the whole rule; the rest
of this block is just its consequences.

Allowed writes:
  - ONE new audit file under `planning/audits/`, created with
    `tskflwctl audit new` (never hand-built)
  - additive annotations on existing `planning/tasks/*.md`
    (`audit_sources:` + one `task append` line)
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
   quote the error, and exit without a PR.

1. Read `CLAUDE.md` and `docs/ARCHITECTURE.md` fully before anything else —
   they define the non-negotiables this audit measures against (DI via one
   `*cli.App`, output through injected `io.Writer`, `--json` everywhere with a
   `schema_version`, core never touches fs/cobra, status authoritative in
   frontmatter, atomic writes via `store/atomic.go`, errors wrapping the domain
   sentinels). A finding that contradicts one of those is a real finding; a
   finding that merely disagrees with them is an ADR question, not a code bug —
   route it to the architecture audit instead.

   Skim `./bin/tskflwctl epic list` for committed project context; it informs
   urgency scoring in step 8.

2a. BACKPRESSURE CHECK — do this before picking a lens:

       ./bin/tskflwctl audit list --json

    Count audits in the `open` bucket. **If the count is 10 or more, do NOT
    author a new audit.** An audit nobody reads is worse than no audit: it costs
    a run, adds a file, and re-finds what an earlier audit already found.
    Instead, spend this run closing the backlog:

    (a) Take the 3 oldest open audits (`audit list --json`, sort by date).
    (b) For each open finding, check whether an existing task already owns the
        work — `./bin/tskflwctl task list --json`, match by file path and
        symbol, not by title.
    (c) If a task owns it:
            ./bin/tskflwctl audit finding <audit> <code> \
              --status "tracked by <task-id>" --note "<why it moved>"
        `tracked` REQUIRES the destination, and it goes INSIDE the --status
        value — `--status "tracked by <id>"`. A bare `--status tracked` is
        rejected even if --note names the task. Never hand-edit `**Status:**`.
    (d) Re-verify anything you flip. A finding whose code has moved is `fixed`
        (`--status fixed --note "<what shipped>"`), not tracked.
    (e) When every finding in an audit is terminal, `./bin/tskflwctl audit close
        <slug>`. It refuses while findings are open — that refusal is correct;
        don't force it.

    Then post a Slack summary of what you closed and skip to step 15 (PR). Do
    not author an audit on a backpressure run.

2b. Determine today's lens from the rotation index:

        TZ=America/New_York date "+%A %Y-%m-%d %V"

    `%V` is the ISO week number. Compute `index = (week * 2 + slot) mod 6`,
    where `slot` is 0 on Tuesday and 1 on Friday, then look up LENS-ROTATION
    below. If today is neither Tuesday nor Friday, stop and post a Slack message
    saying so.

    The rotation exists so the twice-weekly cadence still covers all six lenses
    — a full cycle every three weeks. **Do not key the lens to the day name**:
    with only two slots a week, that would starve four of the six lenses.

=== DISCOVERY (signal-driven + adjacency + random, within the lens) ===

3. Open LENS-SCOPES for today's lens. It names a surface area and the questions
   to ask. Compute the "churn × blast radius" signal *within that surface*:

   (a) List churn in the lens's paths (last 30 days):

           git log --since="30 days ago" --name-only --pretty=format: \
             -- <today's lens paths> \
             | grep -v '^$' | sort | uniq -c | sort -rn | head -20

   (b) For each of the top ~10, estimate blast radius — how many other files
       reference it. For Go, grep for the package's exported identifiers:
       `grep -rln "<ExportedName>" internal cmd`. A file in `internal/core` or
       `internal/domain` has structurally higher blast radius than one in
       `internal/tui`; weight accordingly.

   (c) Rank: score = log(churn + 1) × log(blast_radius + 1). Pick the top 2.

4. For each of the 2 signal picks, choose ONE adjacency hop:
   - the file it imports most heavily, OR
   - a file that imports it heavily, OR
   - its **test file** if the lens is `test-rigour`, OR
   - a sibling in the same package untouched for 6+ months (stale neighbor —
     high serendipity).

   In this codebase the highest-yield adjacency is usually **across the
   hexagonal seam**: from a `core` invariant to the `store` guard that enforces
   it, or to the `wire` type that publishes it. Prefer that hop when the lens
   allows.

5. Pick 1 fully random file from the lens's surface. Quality filter:
   - SKIP generated code (`docs/cli/*.md`, `internal/cli/testdata/golden/*`,
     `internal/tools/*gen*`)
   - SKIP `internal/testutil/` fixtures
   - SKIP `*_test.go` UNLESS the lens is `test-rigour`
   - SKIP files < 50 lines

6. You now have ~4–5 files. Read each fully, plus enough surrounding context
   (callers, the domain types it uses, its tests, the goldens it feeds) to
   evaluate correctness. For anything in `core` or `store`, read the test file
   too — this codebase encodes most of its contract in tests, and a finding that
   ignores existing coverage is usually wrong.

=== AUDIT ===

7. Work the lens's questions from LENS-SCOPES. Stay in the lens: a
   concurrency-lens run that reports a naming nit is a wasted run, and the
   naming nit will come around on the simplification lens in three weeks.

   Beyond the lens questions, always check the repo's own non-negotiables when
   a file you're reading touches them:
   - errors wrap `ErrNotFound` / `ErrValidation` / `ErrAmbiguous` / `ErrConflict`
     so the CLI maps exit codes 10/11/13/14 (12 is retired but reserved);
   - new file writes go through `store/atomic.go`
     (`writeFileAtomic` / `createFileAtomic`);
   - frontmatter edits are surgical — unknown fields, comments, and key order
     preserved;
   - `core` imports no fs and no cobra; `tui` reads through `core.Service` as
     `tea.Cmd`s with no I/O in `Update`/`View`;
   - every `--json` envelope carries `schema_version`.

8. Score each finding on three axes — **severity** (Critical/High/Medium/Low),
   **effort** (XS/S/M/L), and **urgency** (acute/soon/eventually, judged against
   committed project state in `planning/epics/` and any accepted ADR the code
   implements).

   Severity and urgency *can* diverge — a Medium real-bug can outrank a High
   low-probability one. When they do, leave a one-line **Context** note on the
   finding explaining why.

9. Be honest about what you DIDN'T find. If two of the five files audit clean,
   say so explicitly in the audit body — that's a real result, not a failure.
   Don't invent mediocre findings to fill a quota.

   For any finding you believe is a live defect, state a **concrete failing
   scenario**: inputs or repo state → wrong output. A finding you cannot make
   fail is a code-smell observation and should be labelled Low, not Medium.

=== CROSS-REFERENCE OPEN TASKS ===

10. For each Medium+ finding, check whether an open task already tracks it:

    (a) Grep `planning/tasks/` for the finding's technical fingerprint (file
        paths, symbols, flag names, schema fields, error strings). Restrict to
        active statuses:
            ./bin/tskflwctl task list --json -c slug,status,description
    (b) Read each candidate task body and classify overlap: **FULL** (task
        captures this work), **PARTIAL** (covers some, or proposes a different
        fix for the same symptom), **NONE** (keyword hit only).
    (c) Act per overlap:
        - **FULL** — mark the finding tracked, with the destination:
              ./bin/tskflwctl audit finding <audit> <code> \
                --status "tracked by <task-id>" --note "<why it moved>"
          The candidate-tasks list gets `⏳ tracked in planning/tasks/<file>`
          instead of a fresh `task new` suggestion. Annotate the task
          additively: append this audit's path to `audit_sources:` (create if
          missing). `audit_sources` is a LIST and `--set` REPLACES it —
          read the current value first (`task list --json`, no `-c`) and pass
          the full comma-separated set:
              tskflwctl task set <slug> --set audit_sources=<existing>,<new>
          Passing only the new path silently drops the earlier ones.
          Then add one
          body line:
              ./bin/tskflwctl task append <slug> \
                --body "Reinforced by audit YYYY-MM-DD-<lens>: <code>."
        - **PARTIAL** — leave the finding `open`; cross-reference both sides
          (audit body + a `task append <slug> --body "..."` line).
        - **NONE** — no change.

    Hard rule: **never** change task status/tier/priority/scope. Judgment calls
    (suspected obsolescence, scope conflict, promotion candidates) go in the PR
    body's "Related-task observations" — propose-only, same as findings.

    Findings resolved by a FULL cross-ref still count toward step 11's Medium+
    threshold: the work is real, the cross-ref just routes it.

=== EXTERNAL RESEARCH (conditional) ===

11. If you found < 3 Medium-or-higher findings, do light web research on the
    audited lens:

    - Search for known Go antipatterns in this lens that today's code does NOT
      violate — proactive recommendations, not bugs.
    - Search for recent (last 12 months) writing on the pattern in question,
      scoped to the actual stack: Go 1.2x, cobra, bubbletea, `gopkg.in/yaml.v3`,
      atomic file replacement on POSIX, CLI/agent JSON contract design.
    - WebSearch to find, WebFetch to read. 3–5 queries max. Cite every URL.

    If you found ≥ 3 Medium+ findings, skip this step — there's enough concrete
    work surfaced already.

=== AUDIT FILE ===

12. Create the audit file with the tool — never hand-build the path or
    frontmatter:

        ./bin/tskflwctl audit new <lens-slug>

    writes `planning/audits/<id>-YYYY-MM-DD-<lens-slug>.md` with valid
    `area`/`date` frontmatter in the `open` bucket. `<lens-slug>` is the
    kebab-case name from LENS-ROTATION. Add `--date` only to backdate.

    Write the body with `audit append` or `audit edit`, following
    AUDIT-FILE-TEMPLATE below. Each finding carries `**Status:** open` from the
    scaffold. **Never hand-edit a `**Status:**` or `**Resolution:**` line
    afterwards** — `./bin/tskflwctl audit finding <audit> <code> --status <v>
    [--note <text>]` owns both, in one validated atomic edit.

    Before generating the candidate-tasks list, run `./bin/tskflwctl epic list`
    to use real epic IDs in `task new` suggestions — template examples drift.
    `21-code-quality-architecture-hardening` is the usual home for
    audit-derived work; pick a better-fitting epic when one exists.

13. Validate what you just wrote. **All four commands, not just the first:**

        git status --porcelain                      # MUST list only planning/ paths
        just build
        just test                                   # race-enabled
        ./bin/tskflwctl lint                        # entity tree + step-10 annotations
        ./bin/tskflwctl audit lint <date>-<lens>    # THIS audit's finding statuses

    `tskflwctl lint` does **not** check finding `**Status:**` values —
    `audit lint` is the one that does, and passing it the slug scopes it to the
    file this run produced. Do not close out on a red `audit lint`; a bad status
    is a one-word fix now and an archaeology problem later.

    If `just test` was already red before your run, say so in the PR and do not
    attempt to fix it — that is out of scope for this routine.

=== CLOSE ===

14. If the audit found zero findings AND external research yielded nothing
    noteworthy, do NOT open a PR. Append one line to
    `planning/meta/no-op-log.md` — date, lens, "clean" — and post the
    SLACK clean-run line. Exit.

    Do not leave an empty audit file behind: delete it if you created one before
    concluding the lens was clean.

15. Otherwise, open a PR against `main`:
    Title: "audit(<lens-slug>): code quality — YYYY-MM-DD"
    Body: AUDIT-PR-BODY below. The PR includes the new audit file plus any
    step-10 task annotations.

16. Post to Slack #planning-updates using SLACK-MESSAGE-TEMPLATE below —
    verbatim. The PR URL MUST be wrapped in Slack's `<url|label>` angle-bracket
    syntax; a bare URL line absorbs the following text into the link and breaks
    it.

If anything fails (build error, file write error, unexpected repo state), STOP,
post a Slack message describing what happened, and exit without opening a PR.
Don't push a half-finished audit.

=== LENS-ROTATION ===

  index = (ISO week number * 2 + slot) mod 6,  slot: Tuesday = 0, Friday = 1

    0 → correctness-and-errors
    1 → concurrency-and-atomicity
    2 → contract-and-compatibility
    3 → test-rigour
    4 → adapter-hygiene
    5 → simplification

  Compute with: TZ=America/New_York date "+%V %A"
  Full cycle: 6 lenses / 2 runs per week = every 3 weeks.

=== LENS-SCOPES ===

(A guide for discovery, not a checklist. Use judgment based on what's actually
in the tree today.)

  correctness-and-errors
    Surface: internal/core/, internal/domain/, internal/store/
    Ask: nil and zero-value handling on exported entry points; off-by-one and
      boundary cases in slicing/indexing; error paths that swallow or reclassify
      a sentinel (does the CLI still map the right exit code?); errors created
      with fmt.Errorf that forget %w; validation that runs after a mutation
      instead of before; map iteration where output order must be deterministic;
      `default:` switch arms that silently absorb a new enum value.
    High-yield: any `switch` over domain.Status / a Role / a Gate / an
      AuditBucket with a `default:` that returns a benign value.

  concurrency-and-atomicity
    Surface: internal/store/ (atomic.go, *lock*, *mutation*), internal/core/
      guarded planners, internal/tui/ (tea.Cmd goroutines, fsnotify/debounce)
    Ask: is every write atomic (writeFileAtomic / createFileAtomic) or is there
      a raw os.WriteFile hiding somewhere; TOCTOU between a read, a plan, and a
      write; is the CAS token derived from content the writer actually compared;
      retry loops that can retry a *committed* write; lock acquisition ordering;
      what happens if the process dies between two writes that must land
      together; shared state captured by a tea.Cmd closure; `go test -race` gaps
      (a race only reproducible under contention that no test creates).
    High-yield: any mutation path that reads a snapshot, validates against it,
      and writes without re-checking the snapshot.

  contract-and-compatibility
    Surface: internal/wire/, internal/cli/testdata/golden/, docs/cli/,
      internal/domain/entity.go + fields.go, cmd/tskflwctl/
    Ask: did a `--json` field change meaning without a `SchemaVersion` bump or
      changelog line; is every envelope carrying schema_version; do the
      `jsonschema:"description=..."` strings still describe the actual predicate
      (a stale description is a shipped lie); is a golden asserting the behaviour
      or only the version string; does `--help` text match what the flag does;
      does a `-c` column projection expose a field the full `--json` doesn't;
      is a renamed flag still aliased.
    High-yield: a field whose derivation changed while its published description
      or name did not.

  test-rigour
    Surface: internal/**/*_test.go, internal/cli/testdata/
    The question: does the suite prove **rejection** as rigorously as success?
    Technique — **guard-coverage tracing**, done entirely by reading:
      (1) Enumerate the load-bearing predicates in the package under review —
          every `if !ok`, `if err != nil { return ... ErrValidation }`, and
          every early `return` that refuses a mutation. `grep -n 'ErrValidation\|
          ErrConflict\|ErrNotFound\|ErrAmbiguous' <pkg>/*.go` finds most of them.
      (2) For each, ask: what repository state trips this? Then grep the package's
          tests for a case that CONSTRUCTS that state and ASSERTS the refusal —
          not merely a test that happens to pass through the function.
      (3) A predicate with no such test is the finding. Name the predicate, the
          state that would trip it, and the test that should exist.
    Do NOT verify this by editing source. Reason about coverage statically; if a
    claim genuinely cannot be settled by reading, label the finding's confidence
    "unverified — needs a probe" and say what probe a human should run. You have
    no write access to `internal/`; a finding you cannot confirm read-only is
    still worth reporting, honestly labelled.
    Also look for: table tests whose `default:` arm asserts something trivially
      true (a new enum value would satisfy it silently); goldens that changed
      only by version string when the behaviour changed; assertions on error
      *text* that break on rewording but pass on the wrong error type; missing
      coverage for the combination matrix a recent feature introduced.
    High-yield: a guard added in the last 90 days (`git log -S'<predicate>'`)
      with no negative test.

  adapter-hygiene
    Surface: internal/cli/, internal/tui/, internal/wire/, internal/core/service*.go
    Ask: does an adapter reconstruct a decision that `core` already owns (the
      canonical smell — a `cli` or `wire` file re-deriving eligibility,
      readiness, or a status predicate from primitive fields); does anything in
      `core` import `os`, `filepath`-for-IO, or cobra; is output going through
      the injected io.Writer or straight to os.Stdout; is there a package-level
      var doing the work DI should; does `tui` do I/O in Update/View; does a
      renderer format a domain decision instead of displaying one.
    High-yield: `grep -rn "core\.Role\|core\.Gate\|core\.Status" internal/cli
      internal/tui internal/wire` — every hit is a candidate leak.

  simplification
    Surface: the two largest packages by line count (currently internal/cli and
      internal/tui), plus whatever the churn signal surfaces
    The single question: **can this be simpler without losing functionality and
      without resorting to tricks?** Best-practice code, not maximally terse
      code. A change that shaves lines but makes the code harder to read,
      removes a name that aids comprehension, or leans on a language quirk is
      NOT a simplification — reject it.
    Ask: duplicated logic that wants one helper; a two-level abstraction with
      one implementation; parameters always passed the same value; a struct
      field only ever set once; error handling repeated verbatim across sibling
      commands; a switch that a table would express better (or a table that a
      switch would); dead code and unreferenced exports.
    Score each candidate on **Reduction** (lines/branches removed) ×
      **Risk** (behaviour-change likelihood) × **Effort**. Anything Risk-high
      goes in Follow-up, not Recommendation.

=== AUDIT-FILE-TEMPLATE ===

  (Frontmatter is written by `audit new` — do not hand-author it. Body:)

  # Code Quality Audit: <lens> — YYYY-MM-DD

  > Edit findings through `tskflwctl audit finding` so status and resolution
  > metadata stay queryable. Never hand-edit a `**Status:**` line.

  Routine: `code-quality-audit` · lens `<lens-slug>` · ISO week `<YYYY-WNN>`,
  slot `<Tue|Fri>` (index <N>).

  ## Punch list

  One line per finding for fast triage. Order: Critical → High → Medium → Low.
  Format: `<code>. <title>  (effort: <X> · urgency: <Y>)`

  ## Files audited

  - **Signal**: `<path>` — <one line: why it scored high>
  - **Signal**: `<path>` — ...
  - **Adjacency** (from `<signal>`): `<path>` — <which hop and why>
  - **Adjacency** (from `<signal>`): `<path>`
  - **Random**: `<path>`

  ## Commands run

  (What you executed and what it returned — build, test, lint, any mutation
  probe. This is how a reader knows a claim was verified rather than inferred.)

  ## Findings

  ### Critical

  (none — or full findings)

  ### High

  #### H1. <Title>  · **Status:** open

  **File:** `path:line` | **Component:** <component>
  **Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>

  <what's wrong, why it matters, evidence>

  **Failing scenario:** <concrete inputs or repo state → wrong output. Omit only
  for a finding explicitly labelled as a smell rather than a defect.>

  **Why tests didn't catch it:** <one line>

  **Recommendation:** <the minimum fix>

  **Tightening (adjacent):** <small wins to roll in with the fix — a missing
  test invariant, a stale comment, a name. Omit if none.>

  **Follow-up:** <out-of-scope work that should become its own task — design
  calls, cross-file refactors, API surface changes. Omit if none.>

  ### Medium

  (M1, M2, … — same per-finding structure)

  ### Low

  (L1, L2, … — keep tight. If a Low isn't adjacent to a higher-severity finding,
  omit it rather than creating churn.)

  ## What audited clean

  (One line per file you read and found nothing notable in, with the *why*.
  This is a real result, not a gap. Do not skip this section.)

  ## External research (only if <3 Medium+ findings)

  <patterns/antipatterns researched, with cited URLs>

  ## Candidate tasks (human to triage)

  Mix of `tskflwctl task new …` suggestions (findings with no matching open
  task) and `⏳ tracked in planning/tasks/…` cross-links (already covered — see
  step 10). Do NOT run the `task new` lines yourself.

  - `tskflwctl task new "<title>" --epic <epic-id> --tags <tag> --tier <N> --priority <p> --description "..."`
  - ⏳ M2 tracked in `planning/tasks/<id>-<slug>.md`

  ## Related-task observations (propose-only)

  Judgment calls from the step-10 cross-reference. Do NOT change task
  status/tier/priority autonomously. Omit the section if empty.

  - Possibly demotable: `planning/tasks/<id>-<slug>.md` — <evidence>
  - Possibly already-shipped: `planning/tasks/<id>-<slug>.md` — <evidence>
  - Scope conflict: `planning/tasks/<id>-<slug>.md` vs finding `<code>` — <line>

=== AUDIT-PR-BODY ===

  ## TL;DR

  <one sentence: lens + headline finding, or "lens audited clean">

  ## Lens

  <lens-slug> (ISO week <YYYY-WNN>, slot <Tue|Fri>, index <N>)

  ## Counts

  - Critical: N
  - High: N
  - Medium: N
  - Low: N

  ## Headline findings (top 3 by severity, or all if <3)

  - **H1**: <title> — `path:line`
  - **M1**: <title> — `path:line`

  ## Validation

  - `just build` · `just test` · `tskflwctl lint` · `tskflwctl audit lint <slug>`
  - <note any pre-existing red, explicitly marked as pre-existing>

  ## External research

  (2–3 lines if done, else "none — sufficient findings surfaced".)

  ## Candidate tasks

  (Same list as the audit file's section — easy to copy/paste from the PR.)

  ## Open-task cross-references

  - <N> findings marked `tracked by <task-id>` (see audit body)
  - <N> annotations added to existing tasks (audit_sources + reinforcement note)
  - Related-task observations: see audit body (omit if empty)

=== SLACK-MESSAGE-TEMPLATE ===

  Post this structure to #planning-updates. Goal: helpful but not verbose — the
  counts say HOW MUCH, the bullets say WHAT, so a reader can decide whether to
  open the PR without opening it. Bare counts fail that test; always name the
  findings. The PR URL MUST be wrapped in Slack's <url|label> angle brackets — a
  bare URL on its own line gets auto-linked and Slack pulls the following text
  into the link target, producing a broken link.

  Normal run (PR opened):

  📋 *taskflow code quality — <lens>:* <N> Critical · <N> High · <M> Med · <L> Low
  • <code> <one-line finding title>
  • <code> <one-line finding title>
  <https://github.com/andy-esch/taskflow/pull/NN|PR #NN>

  Bullets: top findings by severity (highest first), reusing the punch-list
  titles verbatim — no bodies, no file paths, each under ~80 chars. Up to 3. If
  the run surfaced no findings but opened a PR for external research, replace
  the bullets with one "• research: <one-line gist>".

  If any Critical: prepend "🚨 " to the header line — the critical is already the
  first bullet, so don't repeat it on a separate line.

  Backpressure run (triage instead of authoring):

  🧹 *taskflow code quality — backlog triage:* closed <N>, tracked <M> findings
  <https://github.com/andy-esch/taskflow/pull/NN|PR #NN>

  Clean run (no PR opened) — single line:

  📋 *taskflow code quality — <lens>:* clean (see planning/meta/no-op-log.md)
```

---

*Bootloader, configuration, tuning knobs, and change history for this routine
live in [`code-quality-audit.NOTES.md`](code-quality-audit.NOTES.md)
(human-maintainer reference, not read during a run).*
