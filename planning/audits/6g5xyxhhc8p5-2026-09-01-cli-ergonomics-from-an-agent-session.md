---
schema: 1
id: 6g5xyxhhc8p5
bucket: open
area: cli-ergonomics-from-an-agent-session
date: "2026-09-01"
updated_at: "2026-09-02"
---
# Audit: cli-ergonomics-from-an-agent-session — 2026-09-01

> Findings from using `tskflwctl` v1.58 as an agent, unprompted, to reconcile a drifted board in a
> separate impl repo (`ethyca/pie`): completing a finished task, splitting out follow-on work, wiring
> a dependency, and reframing two tasks whose premise had changed. Roughly 20 invocations. Reported
> because the friction was real, not because it was asked for.

## What worked — recorded because an audit that only lists faults misleads

**The acceptance-criteria gate on `task complete` changed the outcome of the work.** Completing a task
with 9 unmet criteria was refused, with the states offered inline (`--defer|--wontfix|--tracked|--na`)
and `--force` available. Left to my own judgement I would have force-closed it. Instead I read the
criteria and found five were real, unfinished, and belonged to a *different* task — so they moved
there with reasons instead of disappearing. **The guard did not slow the work down; it corrected it.**

**`task depend migrate --dry-run` showed the exact edges** it would add and clear before writing. That
is the right shape for a mutation an agent is about to run on someone else's repo sight-unseen.

**`task blockers` output is excellent** — `graph: healthy`, `state: queued/blocked eligible=false`,
and the offending edge printed as `A -> B`. It answers the question completely. See H1 for why almost
nobody will run it.

## Findings

#### H1. `board` and `task list` do not surface blocked/eligibility state · **Status:** open

**File:** board and task list output paths | **Component:** cli-read-surfaces
**Effort:** S · **Urgency:** soon

After adding a hard prerequisite, `board` showed the blocked task **first** under `next-up (4)`, with
nothing to distinguish it from three actionable ones:

```
next-up (4)
  design-the-prompt-artifact-evaluation-matrix       high  Replace one-off prompting experiments...
  review-the-conventions-guide-for-answer-leakage    high  The biggest measured lever may be...
```

The first is blocked *by* the second. `task list --status next-up` is likewise silent. The state is
known — `task blockers` reports `queued/blocked eligible=false` — so this is a display gap, not a
model gap.

Why it matters more for agents than humans: an agent reads `board`, picks the top high-priority item,
and starts work that `task start` will then refuse. A human eventually remembers the dependency; an
agent re-derives the board from scratch every session and will re-make the same wrong choice every
time. The board is the tool's answer to "what should I do next", and it is currently answering with
work that cannot be started.

**Recommendation:** mark ineligible tasks in `board` and `task list` — a `⛔`/dim treatment, or sort
them last within their status. Ideally `board` gains the frontier view for free: eligible work first,
blocked work visibly parked.

**Resolution:**

#### H2. `--tracked` takes an unvalidated free-text reason, so the pointer rots · **Status:** open

**File:** task ac --tracked | **Component:** acceptance-criteria
**Effort:** S · **Urgency:** soon

`schema audit` says a `tracked` **finding** "needs the destination (`tracked by <id>`)". For a
**task acceptance criterion**, `--tracked N --reason "..."` accepts any prose. I wrote five of these:

```
--tracked 11 --reason "handed to make-the-train-test-split-more-rigorous: the split artifact is
                       that task's deliverable, not the corpus's"
```

Nothing resolves that slug, nothing links it, and nothing will notice if the destination task is
renamed (`task rename` cascades *body* links, not AC reasons) or completed without absorbing the
criterion. The criteria I moved are the substantive residue of a completed task — exactly the content
that must not rot — and they are held together by a string.

**Recommendation:** `--tracked N --to <task-ref>` that resolves the reference at write time and
records the id, keeping `--reason` for the prose. Then `task unblocks`/`blockers` can see it, and lint
can flag a criterion tracked to a task that is completed or deprecated. Same treatment the finding
statuses already get.

**Resolution:**

#### H3. Repo-global graph degradation surfaces at an unrelated mutation · **Status:** open

**File:** graph guard on task transitions | **Component:** dependency-graph
**Effort:** S · **Urgency:** eventually

`task complete <finished-task>` failed with:

```
validation failed: repository task graph is degraded: 3 legacy dependency field occurrence(s)
remain; run `tskflwctl task depend migrate`
```

The message is good — names the cause, names the remedy — and the latch is defensible: mutating a
graph you cannot trust is worse than stopping. But the degradation had nothing to do with the task
being completed, and it surfaced *mid-operation* in an unfamiliar repo. The three legacy fields had
presumably been there for weeks.

**Recommendation:** surface it where it can be fixed calmly — a line in `status` and `board`
("⚠ graph degraded: 3 legacy dependency fields · run `task depend migrate`"), and in `lint` output.
`lint` currently reports "all planning entities and dependency links pass lint" on a repo the graph
guard considers degraded, which reads as a contradiction.

**Resolution:**

#### H4. A nested code fence silently drops every finding after it · **Status:** fixed

**File:** audit body parser; `task append` / `audit new --body-file` | **Component:** body-mutation
**Effort:** S · **Urgency:** acute

**This audit demonstrates the bug.** As first written it contained six findings. `audit show` renders
all six and `lint` passes, but the tool's finding index sees **four**:

    $ tskflwctl audit info <this-audit>
    findings: 4 total · 4 open · 0 in-progress · 0 done · 0 dropped

    $ tskflwctl audit finding <this-audit> H5 --status open
    error: not found: no finding "H5" in this audit

The cause was a four-backtick fence in this finding, wrapping a three-backtick example. That is valid
CommonMark, but the parser is line-based and treated everything after it as fenced, swallowing H5 and
H6. The failure is silent at every surface an author would check: creation succeeded, `lint` passed,
and the rendered body looked right.

The consequence is worse than a display glitch. A swallowed finding **can never be stamped** —
`audit finding H5 --status fixed` fails permanently — and the progress bar reports `4/4 settled` on an
audit with two unaddressed findings. An audit that under-reports its own scope is the one artifact
here that must not do that.

Related but distinct: a truncated `task append` earlier in the same session wrote a body ending inside
an **unterminated** fence, and that was accepted too. Same root cause from the other side — nothing
validates fence structure on write, and nothing validates it on read.

**Recommendation:** parse fences by their opening run-length (CommonMark: a fence closes only on a run
of the same character at least as long), so nested fences work. Independently, reject or warn on a
body with unbalanced fences at `append`/`new --body-file` time — the write is where the author can
still fix it cheaply. A cross-check that `audit info`'s count matches the `####` headings in the body
would have surfaced this immediately.

**Secondary, docs-only:** every multi-line example in `task append --help` uses `printf`. Percent signs
are ubiquitous in planning bodies — metrics, coverage, gate thresholds; mine had eight — and an
unescaped `%` truncated one of my writes. Lead with a heredoc into `--body-file -`, which has no
format-string surface.

**Resolution:** blankFences now scans lines through the shared fenceScanner
(CommonMark run-length) instead of pairing ``` runs with a regex, so nested,
tilde and unterminated fences all mask correctly. store.writeBody refuses a body
ending inside an open fence, beside its existing parse-before-commit guard. The
task/audit append examples lead with a heredoc into --body-file - rather than
printf, which truncated a real write at an unescaped percent sign.

#### H5. `task depend add A B` fails with a bare arity error · **Status:** fixed

**File:** task depend add | **Component:** cli-arg-parsing
**Effort:** XS · **Urgency:** eventually

`task depend add <dependent> <prerequisite>` — the shape the verb suggests — returns:

```
error: accepts 1 arg(s), received 2
```

The real form is `add <task> --on <prereq>`, which `--help` shows plainly. But the error is Cobra's
default and teaches nothing; two task-shaped arguments is a strong signal of exactly this mistake.

**Recommendation:** when `depend add` receives extra positional args, suggest the flag form —
`did you mean: task depend add A --on B?`. Same for `depend remove`.

**Resolution:** task depend add/remove now reject extra positional args with the
flag form spelled out — 'did you mean: tskflwctl task depend add alpha --on
beta?' — instead of cobra's bare arity count. Several trailing prerequisites
each get their own --on in the suggestion.

#### H6. Bare-verb suggestion points at `lint` for `list` · **Status:** fixed

**File:** root command suggestions | **Component:** cli-arg-parsing
**Effort:** XS · **Urgency:** eventually

`tskflwctl list` (forgetting the noun) suggests:

```
Did you mean this?
	lint
```

Edit distance is doing the work, and `lint` is a validator — a plausible thing for an agent to run on
a repo it does not own, with a very different effect from listing tasks. `board`, `status` and
`task list` are all closer in intent.

**Recommendation:** hand-map the common bare verbs (`list`, `show`, `new`, `start`, `complete`) to
their noun-qualified forms, or append "run `tskflwctl board` for active work" to the root error.

**Resolution:** Hidden redirect commands intercept the bare verbs (list, show,
new, start, complete, edit) and name their noun-qualified forms, so 'list' no
longer answers with 'lint'. They use the styleOnlyPreRun opt-out so a usage
error does not depend on repo discovery; genuine typos still reach cobra's
distance matching.

## Relationship to `2026-07-24-ai-agent-cli-ergonomics`

That audit covered the same surface eight weeks earlier, from design review rather than field use.
Read together:

- **Its M5 — "`task complete` does not reconcile unfinished acceptance criteria" — is marked
  `fixed 2026-08-24`, and this session is field confirmation that the fix works.** The gate caught a
  real force-close and changed what I did. Recorded under "What worked" above; no new task needed.
- **Its M4 — "Structure-aware body writes stop short of the edits agents make most"
  (`tracked by 6fpnn6zk157b`) — is adjacent to H4 here** but not the same. M4 is about *which* edits
  are expressible; H4 is that an expressible edit produces a body the tool then mis-parses. Worth
  checking whether 6fpnn6zk157b already covers fence handling before opening a new task.
- **Its L2 — "Cross-session resumption still requires several reads and body interpretation" — is the
  general case of H1.** An agent resuming cold reads `board`, and `board` currently omits the one
  fact that determines whether the top item is actionable.

Nothing here duplicates an open finding in that audit; H1–H6 are all things that bit during use.

## Method note

Six findings from roughly 20 invocations, reconciling a drifted board in an impl repo the tool points
at from outside (`-C`). Every finding is something that cost time or produced a wrong artifact in the
session — none is speculative. Two of the six (H4, H6) were found by making the mistake, not by
inspecting the code, which is the sample this audit is drawn from and its main limitation: it says
nothing about surfaces I did not touch (`ui`, `thread`, `routine`, `epic`, `research`, `space`).

## Candidate tasks

- ⏳ `tskflwctl task new "Surface blocked and ineligible tasks in board and task list" --epic 30-threads-and-task-dependency-graphs --tags cli,board,graph` — H1; the eligibility data already exists in `task blockers`, this is a display change
- ⏳ `tskflwctl task new "Resolve and validate the destination of a tracked acceptance criterion" --epic 20-cli-ux-and-ergonomics --tags cli,acceptance-criteria` — H2; `--to <task-ref>` resolved at write time so the pointer cannot rot
- ⏳ `tskflwctl task new "Report graph degradation in status, board and lint" --epic 30-threads-and-task-dependency-graphs --tags cli,graph,lint` — H3; and reconcile lint passing on a graph the guard rejects
- ⏳ `tskflwctl task new "Fix nested-fence parsing in audit bodies and validate fences on write" --epic 21-code-quality-architecture-hardening --tags audit,parser,body-mutation` — H4, **acute**: a nested fence silently drops findings from the index and they can never be stamped; check 6fpnn6zk157b first for overlap
- ⏳ `tskflwctl task new "Improve arg-shape and bare-verb error suggestions" --epic 20-cli-ux-and-ergonomics --tags cli,errors` — H5 and H6 together; both are one-line suggestion improvements

#### H7. `audit finding --note` duplicates the Resolution block instead of filling an empty one · **Status:** open

**File:** audit finding --note; internal/domain/finding.go note writer | **Component:** body-mutation
**Effort:** XS · **Urgency:** soon

Found while stamping H4 of this audit. Five of the six findings here carry a bare
`**Resolution:**` placeholder with no paragraph — a shape `audit lint` already flags
("empty `**Resolution:**` label"). Stamping one of them appended a SECOND block rather
than filling the empty one:

    **Resolution:**

    **Resolution:** blankFences now scans lines through …

The command reported success. `audit lint` then swapped one complaint for a worse one —
"more than one `**Resolution:**` block — only the first is read" — which means the note
just written is the one being ignored, while the empty placeholder above it wins. A
resolution that is silently unread is the same class of defect as H4: the write succeeds,
the file renders fine, and the tool's own index disagrees with what the author sees.

`NoteSpan` is documented as covering "the note including its label, so re-noting replaces
the block rather than nesting a second label inside the first". That holds when a note
already has a paragraph; an empty label leaves the span empty, so the writer falls through
to the append path.

**Recommendation:** treat an empty `**Resolution:**` label as the target span so the note
fills it, or refuse the write naming the malformed label. Either is better than producing
a state whose own lint rule says the new content is unread. The repository is not clean of
these today, so the fill path is worth preferring over the refusal.
