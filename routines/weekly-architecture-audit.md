---
name: weekly-architecture-audit
version: 1
schedule: "30 11 * * 1"          # 7:30am EDT Mondays / 6:30am EST after Nov DST
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
max_open_audits: 10              # backpressure: at or above this, triage instead of authoring
lens_rotation:
  iso_week_mod_4:
    # Indexed so the first scheduled run (Mon 2026-08-31, ISO week 36,
    # 36 mod 4 = 0) lands on hexagonal-boundaries. Subsequent weeks advance
    # through the natural progression.
    0: hexagonal-boundaries      # ports/adapters, core purity, the store seam
    1: data-model-and-storage    # entity model, flat id-led layout, migration, OCC
    2: machine-contract          # --json envelopes, schema_version, agent ergonomics
    3: failure-and-recovery      # guard topology, atomicity, partial failure, repair
last_modified: 2026-08-30
---

# Weekly Architecture Audit (Single-Lens Rotation)

## Purpose

Weekly deep look at one **architectural lens** — not file-level bugs (that's
`code-quality-audit`'s job). The lens rotates on a 4-week cycle keyed to ISO
week, so full coverage takes about a month. Every run is research-mandatory:
assertions must be anchored to cited external sources. Output is a narrative +
findings hybrid; propose-only, with the same open-task cross-reference discipline
as the code audit.

taskflow has a distinguishing asset most codebases don't: **accepted ADRs under
`planning/adrs/`** that state the intended architecture. This routine's real job
is reconciliation. Every finding must land in one of three buckets:

- **Implementation drift** — the code diverges from an accepted ADR. A defect.
- **ADR gap** — the ADR is silent on something the code had to decide. Surface
  the decision so it can be recorded.
- **ADR pressure** — the code is right and the ADR is now wrong, or the external
  best practice contradicts an accepted decision. Do NOT rewrite the ADR;
  propose an amendment and say what evidence forces it.

Never edit an ADR. That is a human decision recorded through a superseding ADR
or an `## Amendments` entry, per ADR-0001.

## Prompt

```
You are running the weekly-architecture-audit routine on the taskflow repo.
Single lens per run, rotating across a 4-week cycle. Deep analysis of one
architectural concern across the whole system, grounded in cited external
best-practice research and reconciled against this repo's accepted ADRs. Be
specific, be honest, and prefer "this lens is in good shape, here's why" to
inventing mediocre findings.

taskflow self-hosts its own planning: `planning/` describes work on `internal/`
and `cmd/` in the same checkout.

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

1. Read, in this order:
   - `CLAUDE.md` — the non-negotiables in one screen.
   - `docs/ARCHITECTURE.md` — the orientation this audit measures against.
   - `ls planning/adrs/` and read every ADR whose subject touches today's lens
     (LENS-SCOPES names them). Read the **Decision** and **Amendments**
     sections in full; skim the rest. An accepted ADR is the authority — your
     job is reconciliation, not re-litigation.
   - `./bin/tskflwctl epic list` — committed project context for urgency
     scoring in step 9.

2a. BACKPRESSURE CHECK — before picking a lens:

        ./bin/tskflwctl audit list --json

    Count audits in the `open` bucket. If the count is 10 or more, do NOT
    author a new audit. Spend the run triaging the 3 oldest instead, exactly as
    `code-quality-audit` step 2a describes: re-verify each open finding, and for
    any the work has moved to a task, run
    `./bin/tskflwctl audit finding <audit> <code> --status "tracked by
    <task-id>"` (the destination goes INSIDE --status; the tool rejects a bare
    `tracked`). Close an audit
    only when every finding is terminal. Then post the Slack triage line, open
    the PR, and exit.

2b. Determine today's ISO week (America/New_York):

        TZ=America/New_York date "+%G-W%V %A %Y-%m-%d"

    Look up `week mod 4` in LENS-ROTATION below. If today is not Monday, note
    "off-schedule run" in the audit body and proceed — the lens is keyed to the
    week, not the day.

=== DISCOVERY (lens-specific, breadth-first) ===

3. Open LENS-SCOPES for the chosen lens. It's a guide for discovery, not a
   checklist. Goal: build a system-wide mental model of how this concern is
   implemented today, across the hexagonal layers.

4. Read 6–10 key files end-to-end, biased toward:
   - the **seams**: `internal/core/store.go` (the port definitions),
     `internal/core/service*.go`, `internal/cli/app.go`, `internal/wire/wire.go`;
   - the lens's own centre of gravity (LENS-SCOPES names it);
   - and 1–2 "edge" files that test whether the pattern actually holds
     everywhere — a rarely-touched adapter, a small package, a tool under
     `internal/tools/`.

   Quality over quantity. Reading `internal/core/dependency_graph.go` properly
   is worth more than skimming six CLI command files.

5. Sketch the mental model before analysis: the pattern in use today
   (layer by layer), where it is consistent or inconsistent, the implicit
   contract between layers, surprising gaps. This becomes the "State of the
   architecture" section. `file:line` evidence mandatory.

   Then, explicitly: **which ADR governs this?** Quote the clause. If no ADR
   governs it, that itself is worth recording (an ADR gap).

=== EXTERNAL RESEARCH (mandatory, lens-anchored) ===

6. Run 3–6 web searches for best practices in this lens, scoped to the actual
   stack: Go 1.2x, hexagonal / ports-and-adapters in Go, cobra command trees,
   bubbletea Elm architecture, `gopkg.in/yaml.v3` round-tripping, atomic file
   replacement on POSIX, optimistic concurrency over a filesystem, CLI JSON
   contract and schema-evolution design, agent-facing tool ergonomics.

   Prefer recent (last 12 months) authoritative sources — the Go blog and
   proposal docs, maintainer writing, practitioner case studies matching the
   stack. Established texts (Hexagonal Architecture / Cockburn, *A Philosophy of
   Software Design*, the Go proverbs) only when anchoring a specific principle.

   WebSearch to find, WebFetch to read. Cite every URL and quote the specific
   recommendation — don't paraphrase generic advice into something
   defensible-but-vague.

7. For each external recommendation, classify the codebase as **Follows** /
   **Partial** / **Diverges**, and if divergent, whether the divergence is
   justified by project constraints. taskflow is a **single-user, local-first,
   filesystem-backed planning CLI with an explicit agent-first machine
   contract**. That legitimises simplifications which would be antipatterns in a
   networked multi-tenant service — no database, no migrations framework, a
   whole-repo lock instead of row-level concurrency. Say so when it applies;
   don't score the project against a scale it deliberately isn't on. Note the
   reasoning either way.

=== ANALYSIS ===

8. Synthesize into the audit body — Executive summary, State of the
   architecture, ADR reconciliation, Best-practice comparison, Tensions and
   trade-offs (omit if none). Structure per AUDIT-FILE-TEMPLATE.

9. Generate findings only after the comparison is written. A finding is an
   observed divergence from a cited best practice OR from an accepted ADR, where
   the divergence is not justified by a project constraint.

   Every finding declares its **Class**: `implementation drift` (code vs ADR) ·
   `ADR gap` (undecided) · `ADR pressure` (ADR vs evidence) ·
   `unanchored` (no ADR governs; best-practice-only).

   Score each on **severity** (Critical/High/Medium/Low), **effort**
   (XS/S/M/L), and **urgency** (acute/soon/eventually, judged against
   `planning/epics/` and any in-flight ADR rollout). When severity and urgency
   diverge, add a one-line **Context** note.

   An architectural finding must name what it would cost to fix **later** rather
   than now — that's the whole reason to audit shape rather than lines. A
   finding with no compounding cost is a code finding; route it to the next
   `code-quality-audit` run instead of inflating it here.

10. List well-implemented patterns under "What audited clean" with a one-line
    "why" and the ADR clause or cited practice each satisfies. This is the
    counterweight to invented findings — don't skip it.

=== CROSS-REFERENCE OPEN TASKS ===

11. For each Medium+ finding, run the open-task cross-reference — same rules as
    `code-quality-audit` step 10:

    (a) Grep `planning/tasks/` for the finding's fingerprint (concept, paths,
        package names, ADR number, schema field).
    (b) Classify overlap as **FULL** / **PARTIAL** / **NONE**.
    (c) FULL → `./bin/tskflwctl audit finding <audit> <code> --status
        "tracked by <task-id>"` — the destination goes INSIDE --status; a bare
        `--status tracked` is rejected even if --note names the task; the
        candidate-tasks entry becomes `⏳ tracked in planning/tasks/…`; annotate
        the task additively — `audit_sources` is a LIST and `--set` REPLACES
        it, so read the current value and pass the full comma-separated set
        (`task set <slug> --set audit_sources=<existing>,<new>`); then add a
        body line via `task append <slug> --body "Reinforced by audit
        YYYY-MM-DD-arch-<lens>: <code>."`).
        PARTIAL → keep the finding open, cross-reference both sides.
        NONE → no change.

    Hard rule: never change task status/tier/priority/scope. Judgment calls go
    in the PR body's "Related-task observations" — propose-only.
    FULL-tracked findings still count toward the Medium+ tally in step 14.

=== AUDIT FILE ===

12. Create the audit file with the tool — never hand-build the path or
    frontmatter:

        ./bin/tskflwctl audit new arch-<lens>

    writes `planning/audits/<id>-YYYY-MM-DD-arch-<lens>.md` with valid
    `area`/`date` frontmatter in the `open` bucket. `<lens>` is the kebab-case
    slug from LENS-ROTATION. Use AUDIT-FILE-TEMPLATE below for the body.

    **Never hand-edit a `**Status:**` or `**Resolution:**` line** —
    `./bin/tskflwctl audit finding` owns both.

    Before generating candidate tasks, run `./bin/tskflwctl epic list` to use
    real epic IDs — template examples drift.
    `21-code-quality-architecture-hardening` is the usual home for
    architecture-derived work; `24-data-model-evolution-…` and
    `26-frontmatter-schema-declared-validation-contract` fit specific lenses
    better. Pick the closest.

13. Validate what you just wrote. **All four commands, not just the first:**

        git status --porcelain                        # MUST list only planning/ paths
        just build
        just test                                     # race-enabled
        ./bin/tskflwctl lint                          # entity tree + step-11 annotations
        ./bin/tskflwctl audit lint <date>-arch-<lens> # THIS audit's finding statuses

    `tskflwctl lint` does **not** check finding `**Status:**` values —
    `audit lint` is the one that does. Do not close out on a red `audit lint`.

    If `just test` was already red before your run, say so in the PR and do not
    attempt to fix it — out of scope for this routine.

=== CLOSE ===

14. If zero findings AND no real divergence in the comparison, do NOT open a PR.
    Append one line to `planning/meta/no-op-log.md` (date, `arch-<lens>`,
    "clean — see <cited source>"), post the Slack no-op line, and exit. Delete
    the audit file if you created one before concluding the lens was clean.

15. Otherwise, open a PR against `main`:
    Title: "audit(arch-<lens>): weekly architecture — YYYY-MM-DD"
    Body: AUDIT-PR-BODY below. The PR includes the new audit file plus any
    step-11 task annotations.

16. Post to Slack #planning-updates using SLACK-MESSAGE-TEMPLATE below —
    verbatim. The PR URL MUST be wrapped in Slack's `<url|label>` angle-bracket
    syntax; a bare URL line absorbs the following text into the link and breaks
    it.

If anything fails (build error, file write error, red lint, research
consistently turning up paywalls), STOP, post a Slack message describing what
happened, and exit without opening a PR. Don't push a half-finished audit.

=== LENS-ROTATION ===

  ISO week mod 4 == 0  →  hexagonal-boundaries      (first scheduled run, 2026-W36)
  ISO week mod 4 == 1  →  data-model-and-storage
  ISO week mod 4 == 2  →  machine-contract
  ISO week mod 4 == 3  →  failure-and-recovery

  Verify with: TZ=America/New_York date "+%G-W%V"
  (Week starts Monday; ISO week numbers reset annually per ISO 8601.)

=== LENS-SCOPES ===

(Guide for discovery — not prescriptive. Use judgment based on what's actually
in the repo today.)

  hexagonal-boundaries
    Governing ADRs: read `docs/ARCHITECTURE.md` first; then any ADR describing
      the cli/tui/core/store layering.
    Surface: internal/core/store.go (port definitions), internal/core/service*.go,
      internal/cli/app.go + the command tree, internal/tui/, internal/wire/,
      internal/domain/
    Look for: decisions duplicated across the seam (an adapter re-deriving what
      core owns); ports that leak persistence vocabulary into core; a core type
      that only exists to serve one adapter; capability interfaces vs one
      god-Store; whether a new entity kind can be added without touching every
      layer; whether the TUI could be deleted without disturbing core; DI
      discipline (one *cli.App, no package-level state); injected io.Writer
      honoured everywhere.
    Research themes: ports-and-adapters in Go, interface-segregation for storage
      ports, when a hexagon costs more than it earns in a single-binary CLI.

  data-model-and-storage
    Governing ADRs: ADR-0003 (stable-key id-addressed storage), ADR-0006
      (Threads as task DAGs), plus any frontmatter/schema ADR.
    Surface: internal/domain/ (entity.go, fields.go, task.go, thread.go,
      validate.go), internal/store/ (create.go, edit.go, atomic.go, fix.go,
      parsing/frontmatter surgery), internal/tools/*migrate*, planning/ layout
    Look for: whether frontmatter really is surgically edited (unknown fields,
      comments, key order preserved) on every write path; id minting and
      collision handling across entity kinds; how a new field is added and who
      must learn about it; whether the field registry is the single source of
      truth or one of several; migration story for a schema change with existing
      user repos in the wild; what `lint --fix` will and won't repair, and
      whether the gap is principled; entity kinds that have drifted apart
      structurally without a reason.
    Research themes: YAML round-tripping without data loss, schema evolution for
      file-backed records, content-addressed vs key-addressed local stores,
      idempotent repair tooling.

  machine-contract
    Governing ADRs: any ADR on the JSON contract / agent ergonomics; plus the
      `SchemaVersion` changelog in internal/wire/wire.go, which is the de-facto
      compatibility record.
    Surface: internal/wire/ (all of it), internal/cli/render/,
      internal/cli/testdata/golden/, docs/cli/, the `schema` command
      (internal/cli/schema*.go), cmd/tskflwctl exit codes
    Look for: whether schema_version bumps track *semantic* change or only shape
      change; whether reflected `jsonschema:"description=…"` strings still
      describe the real predicate; envelope consistency across commands
      (does every one carry workspace, schema_version, the same error shape);
      whether `--json -c` projections and full `--json` can disagree; exit-code
      taxonomy coherence (10/11/13/14, 12 retired) and whether every sentinel
      still maps; whether an agent can discover everything it needs from
      `schema` alone; goldens that pin the version string but not the behaviour.
    Research themes: CLI JSON output contracts, JSON Schema 2020-12 authoring,
      semantic versioning for machine payloads, designing tool surfaces for LLM
      agents, self-describing CLIs.

  failure-and-recovery
    Governing ADRs: ADR-0006 §guarded mutations; plus any ADR on the repository
      guard / concurrency model.
    Surface: internal/store/ (atomic.go, lock*, lifecyclemutation.go, the
      guarded dependency/thread mutation paths), internal/core/ guarded
      planners and typed errors, internal/cli/moves.go (batch partial failure),
      internal/tui/ error surfacing
    Look for: the guard topology — what the lock covers, what it doesn't, and
      whether a reader can observe a torn state; committed-vs-pre-commit failure
      handling (is every post-commit failure typed so a retry loop stops); CAS
      coverage (whole-snapshot and per-file) and what a non-cooperating external
      editor can still break; retry policy and whether a retried planner is
      genuinely idempotent; batch semantics (does a mid-batch failure leave a
      coherent tree, and does the receipt say so); what a user is told to do
      after each failure class, and whether a supported repair path exists for
      each defect `lint` can report.
    Research themes: crash-safe file replacement (fsync/rename semantics on
      macOS + Linux), advisory locking pitfalls, optimistic concurrency over
      files, designing recoverable partial-failure receipts, error taxonomy and
      exit codes for scriptable tools.

=== AUDIT-FILE-TEMPLATE ===

  (Frontmatter is written by `audit new` — do not hand-author it. Body:)

  # Weekly Architecture Audit: <lens> — YYYY-MM-DD

  > Edit findings through `tskflwctl audit finding` so status and resolution
  > metadata stay queryable. Never hand-edit a `**Status:**` line.
  > Architecture audits are propose-only — no code, docs, or ADR edits.

  Routine: `weekly-architecture-audit` · lens `<lens-slug>` · ISO week
  `<YYYY-WNN>` (mod 4 = <N>).
  ADRs consulted: <list> · Files read: <list> · Sources cited: <count>

  ## Executive summary

  (3–5 sentences. The lens, the headline finding or "this lens is healthy",
  and what to read next.)

  ## State of the architecture (this lens)

  (How the system implements this concern today. Descriptive only, no findings.
  Layer by layer where relevant. `file:line` evidence throughout.)

  ## ADR reconciliation

  For each governing ADR clause this lens touches:

  ### <ADR-NNNN §N> — <clause in one line>

  **Quoted:** "<the actual clause>"
  **Implementation:** <follows | drift | silent> — <evidence with file:line>
  **Class if divergent:** implementation drift | ADR gap | ADR pressure

  (If no ADR governs some part of the lens, say so explicitly here — an
  undocumented architectural decision is a finding in its own right.)

  ## Best-practice comparison

  ### <Practice 1 title>

  **Source:** <URL>
  **Guidance:** <one-line summary of the cited recommendation>
  **This codebase:** <follows | partial | diverges> — <evidence with file:line>
  **Justified?** <yes/no + the project constraint, if divergent>

  ### <Practice 2 title>
  ...

  ## Tensions and trade-offs

  (Where best practice conflicts with project realities — single-user scope,
  local-first, no server, agent-first machine contract, a deliberately small
  dependency set. Omit if nothing surfaced.)

  ## Findings

  Sections in order: Critical → High → Medium → Low. Critical/High get full
  entries; Medium uses M1, M2, …; Low stays tight — omit a Low that isn't
  adjacent to a higher-severity finding rather than creating churn.

  #### H1. <Title>  · **Status:** open

  **File:** `path:line` (or "system-wide") | **Component:** <component>
  **Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>
  **Class:** <implementation drift | ADR gap | ADR pressure | unanchored>
  **Anchored to:** <ADR-NNNN §N and/or URL from the comparison>

  <what's wrong architecturally, why it matters, evidence>

  **Compounding cost:** <what this costs to fix later vs now — which planned
  work it constrains, how many call sites grow per month it survives>

  **Recommendation:** <the minimum fix>

  **Proposed ADR amendment:** <only for ADR pressure / ADR gap findings — the
  clause you would propose, for a human to accept or reject. Never write it into
  the ADR yourself. Omit otherwise.>

  **Follow-up:** <out-of-scope work — design calls, cross-layer refactors.
  Omit if none.>

  ## What audited clean

  (Architectural patterns that are well-implemented, one line each with the ADR
  clause or cited practice they satisfy.)

  ## Candidate tasks (human to triage)

  Mix of `tskflwctl task new …` suggestions and `⏳ tracked in planning/tasks/…`
  cross-links. Do NOT run the `task new` lines yourself.

  - `tskflwctl task new "<title>" --epic <epic-id> --tags <tag> --tier <N> --priority <p> --description "..."`
  - ⏳ M2 tracked in `planning/tasks/<id>-<slug>.md`

  ## Related-task observations (propose-only)

  Judgment calls from the step-11 cross-reference. Omit if empty.

  - Possibly demotable: `planning/tasks/<id>-<slug>.md` — <evidence>
  - Possibly already-shipped: `planning/tasks/<id>-<slug>.md` — <evidence>
  - Scope conflict: `planning/tasks/<id>-<slug>.md` vs finding `<code>` — <line>

=== AUDIT-PR-BODY ===

  ## TL;DR

  <one sentence: lens + headline finding, or "lens audited clean">

  ## Lens

  <lens-slug> (ISO week <YYYY-WNN>, mod 4 = <N>)

  ## Counts

  - Critical: N
  - High: N
  - Medium: N
  - Low: N

  By class: <N> implementation drift · <N> ADR gap · <N> ADR pressure ·
  <N> unanchored

  ## Headline findings (top 3 by severity, or all if <3)

  - **H1**: <title> — `path:line` (anchored to <ADR-NNNN §N / short citation>)
  - **M1**: <title> — `path:line` (anchored to <…>)

  ## ADRs consulted

  - `planning/adrs/<file>` — <clause reconciled>

  ## Proposed ADR amendments (human decision — nothing was edited)

  - <ADR-NNNN>: <one-line proposed amendment> (finding <code>)
  - (omit if none)

  ## External sources cited

  - <URL 1>
  - <URL 2>

  ## Validation

  - `just build` · `just test` · `tskflwctl lint` · `tskflwctl audit lint <slug>`
  - <note any pre-existing red, explicitly marked as pre-existing>

  ## Candidate tasks

  (Same list as the audit file's section.)

  ## Open-task cross-references

  - <N> findings marked `tracked by <task-id>` (see audit body)
  - <N> annotations added to existing tasks
  - Related-task observations: see audit body (omit if empty)

=== SLACK-MESSAGE-TEMPLATE ===

  Post this exact structure to #planning-updates. The PR URL MUST be wrapped in
  Slack's <url|label> angle-bracket syntax; a bare URL line absorbs the
  following text into the link and breaks it.

  Normal run (PR opened) — three lines:

  🏛 *taskflow weekly arch — <lens>:* <N> Critical · <N> High · <M> Med · <L> Low
  • <code> <one-line finding title>
  <https://github.com/andy-esch/taskflow/pull/NN|PR #NN>

  If any finding is class `ADR pressure`, add a line before the PR link:
  ⚖️ proposes an amendment to <ADR-NNNN> (human decision)

  If any Critical: prepend "🚨 " to the first line.

  Backpressure run (triage instead of authoring):

  🧹 *taskflow weekly arch — backlog triage:* closed <N>, tracked <M> findings
  <https://github.com/andy-esch/taskflow/pull/NN|PR #NN>

  Clean run (no PR opened) — single line:

  🏛 *taskflow weekly arch — <lens>:* clean (see planning/meta/no-op-log.md)
```

---

*Bootloader, configuration, tuning knobs, open design questions, and change
history for this routine live in
[`weekly-architecture-audit.NOTES.md`](weekly-architecture-audit.NOTES.md)
(human-maintainer reference, not read during a run).*
