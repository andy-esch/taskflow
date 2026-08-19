---
schema: 1
id: 6fsa47r4f7es
bucket: open
area: ai-agent-cli-ergonomics
date: "2026-07-24"
---
# Audit: AI-agent CLI ergonomics — 2026-07-24

This is a dogfooding audit based on sustained use of `tskflwctl` to inspect epics,
coordinate multi-step work, persist adversarial-review findings, update task bodies,
flip acceptance criteria, and transition tasks across an implementation repository
that routes to an external planning repository.

The baseline is strong. The CLI already has semantic exit codes, compact JSON errors,
stable JSON schemas, `--dry-run`, non-interactive behavior off-TTY, column projection,
section-scoped reads, token-cheap `task info` / `audit info`, atomic writes, and bounded
OCC retries. The findings below focus on the remaining places where an AI agent can
mutate the wrong workspace, lose its place across sessions, duplicate a write after an
unknown outcome, or spend substantially more tokens and round trips than the work
requires.

## Findings

### High

#### H1. Mutations neither assert nor return the resolved planning workspace  · **Status:** fixed

**File:** internal/cli/root.go; internal/wire/envelopes.go | **Component:** discovery / wire
**Effort:** M · **Urgency:** acute

External-planning routing is intentionally transparent: a command run in an
implementation repo can follow `.tskflwctl.toml` into a sibling planning repo. That is
ergonomic until an agent has a stale working directory, uses the wrong checkout, or
works with the same slug in two planning trees. Mutation envelopes return the entity
and sometimes a path relative to the planning root, but not the absolute resolved root,
the config/pointer that selected it, or a stable workspace identity. A successful
mutation therefore cannot prove which planning tree it changed without a separate
pre-read.

This is a wrong-repository write hazard, not merely a discoverability issue. The
existing `clearer-error-and-walk-up-discovery-when-anchored-outside-the-planning-tree`
task improves failure diagnostics but does not guard a successful resolution to the
wrong valid tree.

**Recommendation:** Add a cheap `workspace --json` / `root --json` read and include
`planning_root`, `config_path`, and a stable workspace identity in every mutation and
dry-run receipt. Add an optional `--expect-root` or `--expect-workspace` precondition
that fails with exit 14 before writing when routing resolves elsewhere. Keep human
output terse; this is primarily a machine-contract field.

**Resolution (2026-08-18, fixed).** Three parts, in the order the finding framed them:

1. **The cheap read** — `tskflwctl workspace` (human + `--json`) reports `planning_root`,
   `config_path`, and a `source` of `pointer|config|discovered`. A `pointer` resolution is
   called out explicitly in human output, since that is the case where the directory you
   stand in is not the tree you would change.
2. **The guard** — a global `--expect-root` compared on PHYSICAL paths, enforced in
   `App.resolve()` so it runs ahead of *every* command body. A mismatch is `ErrConflict`
   (exit 14, matching the CAS "world is not what you assumed" code) and nothing is written;
   `TestExpectRoot_MismatchRefusesBeforeWriting` asserts the target file is byte-identical
   afterwards, since reporting the wrong root after the fact would be too late.
3. **The receipts** — every mutation and dry-run envelope (`task set`/`append`,
   `epic set`, transitions, `*  new`, `audit` mutations, `lint --fix`) now carries a
   `workspace` object. Registered in `jsonEnvelopes`, so `schema --json-schema` describes
   it and the goldens pin it.

**Scoped deliberately:** "a stable workspace identity" is the **absolute resolved root**,
not a minted durable id. A minted id that survives a move is a real question, but it belongs
with epic 29's `space.id` rather than being invented twice — noted there.

#### H2. Create JSON calls a mutable slug `id` after the stable-ID migration  · **Status:** open

**File:** internal/cli/task.go:157; internal/cli/audit.go:77; internal/wire/envelopes.go:297-316 | **Component:** wire contract
**Effort:** S · **Urgency:** acute

`task new --dry-run --json` returned:

`{"created":{"kind":"task","id":"agent-create-envelope-probe",...,"path":"tasks/6fsa428vc2mm-agent-create-envelope-probe.md"}}`

The path proves that the task's immutable ID is `6fsa428vc2mm`, but
`created.id` contains the mutable slug. Audit creation has the same defect.
The call sites explicitly pass `t.Slug` / `a.Slug` into `CreatedJSON`; tests pin
the old behavior. Other task/audit JSON DTOs correctly distinguish stable `id`
from human `slug`, so an agent that saves `created.id` receives the wrong handle
only on the command where capturing the new immutable handle matters most.

**Recommendation:** Change `CreatedItem` to carry both `id` and `slug`, pass
`t.ID` / `a.ID` plus their slugs, bump the wire schema, and add contract tests
asserting the ID matches the filename prefix and a subsequent `task info <id>` /
`audit info <id>` lookup. Epic IDs can keep their existing NN-slug identity.

#### H3. Non-idempotent mutations have no replay identity after an unknown outcome  · **Status:** open

**File:** internal/core/service_task.go:134; internal/cli/task.go:588; internal/cli/audit.go:421 | **Component:** mutation reliability
**Effort:** M · **Urgency:** soon

OCC protects concurrent read-modify-write operations, and the local atomic rename makes
an in-process retry safe. It does not solve the agent-orchestrator version of a lost
acknowledgement: the process may commit an append, then the tool call may time out or
disconnect before the agent receives the result. Retrying `task append`, `audit append`,
or future `task log` / `audit finding new` duplicates durable content. Create and state
transition calls are naturally conflict-safe or idempotent; arbitrary appends are not.

The existing OCC design explicitly concluded that idempotency keys were unnecessary
for a local synchronous process. That conclusion is correct inside the process boundary
but does not cover a remote agent tool transport whose acknowledgement can be lost
after process success.

**Recommendation:** Add an optional agent-supplied operation key to non-idempotent
mutations and return it in the receipt. Persist only the minimum dedupe record needed
to return the prior result on replay, with bounded retention and a documented scope
per workspace. If persistent receipts are judged too heavy for a local-first CLI,
provide a narrower `--append-if-absent-hash` contract and document its limitations.

### Medium

#### M1. List and finding queries are projectable but still unbounded  · **Status:** open

**File:** internal/cli/listmode.go; internal/cli/task.go:196; internal/cli/audit.go:102 | **Component:** query / token economy
**Effort:** M · **Urgency:** soon

Column projection is a major improvement, but `--all` can still produce a single huge
JSON line. On this repository,

`task list --all -o json -c slug,status,description`

produced roughly 12,600 output tokens and was truncated by the calling environment.
There is no `--limit`, continuation cursor, deterministic sort selector,
`--updated-since`, or metadata query. An agent looking for a few historical tasks must
either ingest the full corpus, shell-filter it after paying the output cost, or bypass
the CLI with `rg`.

**Recommendation:** Add deterministic `--sort`, `--limit`, and `--after` pagination
to task/audit/epic lists and `audit findings`; return `total` and `next_after` in JSON.
Add `--query` over slug/title/description/tags and `--updated-since` for cross-session
resumption. Preserve the current unbounded behavior when no bound is requested.

#### M2. Body mutation JSON echoes the entire resulting document by default  · **Status:** open

**File:** internal/wire/envelopes.go:137-153,382-395 | **Component:** wire / token economy
**Effort:** S · **Urgency:** soon

`task append --json`, `task ac --check --json`, and `audit append --json` return the
complete resulting Markdown body. A dry-run that appends the eleven characters
`probe-only` to a normal task returned the entire multi-page task. The field-only
`task set` and transition receipts are compact, so the mutation contract is
inconsistent precisely where bodies are longest.

Agents normally need proof of target, operation, dry-run state, and post-write version;
they can call `show` explicitly when they need the full body.

**Recommendation:** Make body mutation receipts compact by default:
`workspace`, `id`, `slug`, `operation`, `dry_run`, `changed`, `updated_at`, and
post-write `version`. Add `--include-body` for the current echo behavior. A dry-run can
optionally expose a short change summary or unified diff without serializing unrelated
body content.

#### M3. `schema` describes entities but not the executable command surface  · **Status:** open

**File:** internal/cli/schema.go; internal/cli/root.go | **Component:** command discovery
**Effort:** M · **Urgency:** soon

`schema --json` successfully exposes statuses, fields, kinds, and semantic exit codes.
It deliberately does not inventory commands. An agent must still traverse root help,
noun help, and verb help prose to discover command paths, positional arguments, flag
types/defaults, mutual exclusions, mutation/destructive classification, and output
envelopes. Global flags are repeated at each help level, increasing context without
adding capability information.

**Recommendation:** Add `schema cli --json` or `capabilities --json`, derived from the
Cobra tree rather than hand-maintained. For every leaf command expose path, purpose,
required/variadic args, flags and types, conflicts, whether it mutates, whether it can
be destructive, dry-run support, input body modes, and the JSON envelope name. This is
a CLI manifest, not an MCP server and not a second execution path.

#### M4. Structure-aware body writes stop short of the edits agents make most  · **Status:** open

**File:** internal/domain/body.go; internal/cli/task.go:588; internal/cli/audit.go:421 | **Component:** structured authoring
**Effort:** M · **Urgency:** soon

The read side is good: section projection, AC enumeration/toggling, task/audit info,
and structured finding queries. The write side still forces raw Markdown surgery for:

- adding, removing, or replacing an acceptance criterion;
- appending a dated progress entry under an existing progress section;
- adding an audit finding and changing its status/resolution block.

`task append` and `audit append` are intentionally structure-blind, so agents can
create duplicate `## Progress Log` or `## Acceptance criteria` headings. The progress
and audit-finding pieces are already tracked by
`task-log-append-a-dated-progress-log-entry` and
`audit-finding-write-surface-status-write-and-candidate-list-sync`; AC evolution has no
equivalent write surface.

**Recommendation:** Finish those two existing tasks and add narrow
`task ac --add/--replace/--remove` operations using the existing fence-aware body
parser. Do not build a general Markdown editor. The useful abstraction is a small set
of domain operations over the conventions the tool already owns.

#### M5. `task complete` does not reconcile unfinished acceptance criteria  · **Status:** open

**File:** internal/cli/moves.go; internal/core/service_task.go | **Component:** workflow integrity
**Effort:** S · **Urgency:** soon

`task complete --dry-run --json` will happily move a task with every acceptance
criterion unchecked. Completed tasks in this repository demonstrate that unchecked
boxes can mean either genuinely unfinished work or deliberate descopes recorded later
in prose. A future agent cannot distinguish the two mechanically, and a task can be
reported complete while its structured contract says otherwise.

**Recommendation:** Before completion, count unchecked criteria. Block by default or
emit a distinct validation result; allow an explicit
`--allow-incomplete --reason <text>` escape hatch that appends a dated closure note.
Alternatively make this policy configurable per planning repo, but always surface the
unchecked count in the transition receipt and in lint. Never auto-check criteria merely
because the status changed.

#### M6. Multi-document agent workflows have no preflighted, restartable change set  · **Status:** open

**File:** internal/cli/moves.go | **Component:** orchestration
**Effort:** L · **Urgency:** eventually

A typical review closeout creates several tasks, appends findings to existing tasks,
flips criteria, and transitions completed work. Each command validates and writes
independently. Batch transitions explicitly attempt every item and can partially
succeed. If an agent is interrupted halfway through ten commands, the planning state is
valid file-by-file but semantically incomplete, and the next session must reconstruct
which operations landed.

True cross-file atomicity may be the wrong fit for a git-agnostic local-files tool.
Restartability and an explicit receipt are still feasible.

**Recommendation:** Design a versioned `apply --file <json|yaml>` change-set format
with workspace precondition, ordered operations, per-operation request IDs, full
preflight, `--dry-run`, and a machine result showing applied/no-op/failed operations.
Prefer resume-safe execution over pretending multiple renames are one transaction.
This also gives sandboxed agent environments one stable, least-privilege mutation
entry point instead of a growing set of approval prefixes.

#### M7. JSON errors classify domain failures but flatten OS recovery information  · **Status:** open

**File:** internal/cli/exit.go:61-87 | **Component:** errors / recovery
**Effort:** S · **Urgency:** eventually

Known domain errors produce useful stable codes. Filesystem failures fall back to
`{"code":"error","message":"..."}`. In dogfooding, a routed planning mutation failed
while creating its temp file with `operation not permitted`; the correct next action
was to request workspace permission and retry. An agent must parse prose to distinguish
permission denied, missing directory, read-only filesystem, disk full, or a generic
internal bug.

**Recommendation:** Keep the human message, but enrich JSON errors with optional
structured details for OS failures: `class` (`permission`, `not-found`,
`read-only`, `no-space`, `io`), `operation`, `path`, and `retryable`. Preserve exit 1
if no new exit-code policy is desired. Add `doctor --write-access` only if real use
shows preflighting is valuable; structured errors are the minimum useful fix.

### Low

#### L1. Generated help is source-synchronized but contains stale semantic copy  · **Status:** open

**File:** internal/cli/audit.go:20-24; internal/cli/listmode.go:49-55 | **Component:** help / docs
**Effort:** XS · **Urgency:** eventually

The generated docs accurately mirror Cobra, but that only proves synchronization with
the source string. `audit close`, `reopen`, and `defer` still say they move audits to
`closed/`, `open/`, and `deferred/` after the flat-layout migration made frontmatter
authoritative. The `-c` help says it “implies -o table” even though `-o json -c ...`
is supported and advertised as the token-cheap agent path. Both statements can steer
an agent away from the real contract.

**Recommendation:** Fix the immediate copy. Then derive lifecycle descriptions from
the frontmatter-state transition registry and test the `-c` wording against the
format-resolution matrix. Add a small stale-vocabulary test for retired layout terms;
doc generation alone cannot catch semantically obsolete prose.

#### L2. Cross-session resumption still requires several reads and body interpretation  · **Status:** open

**File:** internal/cli/task.go:283-315; internal/domain/body.go | **Component:** handoff / token economy
**Effort:** S · **Urgency:** eventually

`task info` is an excellent cheap primitive, but resuming work typically needs:
objective, unchecked acceptance criteria, blockers/dependencies, and the latest
progress entry. Today that means `task info`, `task ac`, and one or more
`task show --section ...` calls, followed by prose interpretation. Long progress
sections still grow without a tail option.

**Recommendation:** Add a composed `task brief <slug> --json` or
`task info --include=objective,unchecked-ac,latest-progress,blockers` view. Derive it
from the existing section/AC parser, include only the latest dated progress item, and
keep the underlying commands. This should be a compact read model, not a new stored
summary that can drift.

## Candidate tasks

- ⏳ `tskflwctl task new "Expose and guard the resolved planning workspace on every mutation" --epic 23-point-an-impl-repo-at-an-external-planning-repo --tags agents,config,safety` — H1.
- ⏳ `tskflwctl task new "Return stable IDs and slugs correctly from create JSON" --epic 24-data-model-evolution-stable-key-storage-read-model-content-occ --tags agents,json,schema` — H2.
- ⏳ `tskflwctl task new "Add replay identities to non-idempotent agent mutations" --epic 20-cli-ux-and-ergonomics --tags agents,reliability,cli` — H3.
- ⏳ `tskflwctl task new "Add bounded search pagination and changed-since filters to list queries" --epic 20-cli-ux-and-ergonomics --tags agents,query,cli` — M1.
- ⏳ `tskflwctl task new "Make mutation receipts compact with opt-in body echo" --epic 20-cli-ux-and-ergonomics --tags agents,json,cli` — M2.
- ⏳ `tskflwctl task new "Publish a machine-readable CLI capability manifest" --epic 20-cli-ux-and-ergonomics --tags agents,schema,cli` — M3.
- ⏳ Finish `task-log-append-a-dated-progress-log-entry`, finish `audit-finding-write-surface-status-write-and-candidate-list-sync`, and add AC-edit operations — M4.
- ⏳ `tskflwctl task new "Gate task completion on unresolved acceptance criteria" --epic 20-cli-ux-and-ergonomics --tags agents,workflow,safety` — M5.
- ⏳ `tskflwctl task new "Design a restartable multi-operation change-set apply command" --epic 20-cli-ux-and-ergonomics --tags agents,orchestration,reliability` — M6.
- ⏳ `tskflwctl task new "Add structured filesystem recovery details to JSON errors" --epic 20-cli-ux-and-ergonomics --tags agents,errors,cli` — M7.
- ⏳ `tskflwctl task new "Remove stale layout and output-mode semantics from CLI help" --epic 20-cli-ux-and-ergonomics --tags docs,cli,agents` — L1.
- ⏳ `tskflwctl task new "Add a compact cross-session task brief read model" --epic 20-cli-ux-and-ergonomics --tags agents,handoff,cli` — L2.

## What should remain unchanged

- Keep Markdown plus frontmatter as the source of truth.
- Keep the CLI as the primary agent interface; an MCP/server layer is not required for
  these improvements.
- Keep semantic exit codes, JSON error envelopes, schema versioning, output projection,
  off-TTY no-prompt behavior, `--dry-run`, section reads, and OCC retries.
- Keep structure-aware writes narrow and convention-driven rather than introducing a
  general-purpose Markdown AST mutation language.
