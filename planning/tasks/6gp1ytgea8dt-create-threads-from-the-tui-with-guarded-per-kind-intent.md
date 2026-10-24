---
schema: 1
id: 6gp1ytgea8dt
status: completed
epic: 18-tui-bubble-tea-interactive-planning-browser
description: Create an empty, unstarted Thread through a staged TUI form with typed guarded receipts; membership and sequencing are separate workflows.
effort: 1-2 days
tier: 3
priority: medium
autonomy_level: 3
tags: [tui, authoring, ux]
created: "2026-10-21"
depends_on: [6gpb4y5yqmjp]
updated_at: "2026-10-24"
started_at: "2026-10-24"
completed_at: "2026-10-24"
---
# Create Threads from the TUI with guarded per-kind intent

## Objective

Create an empty, unstarted Thread from its own TUI page using shared staged-form mechanics and the actual guarded Thread creation contract.

## Scope decision

The owner confirmed that this slice is simple Thread creation only. Stage title,
description, goal and appropriate optional metadata, then submit once. Do not
include initial task membership, task creation, sequencing or a combined project
composition wizard. Use the existing guarded creation contract with no members;
this does not relax graph validation or receipt/recovery rules.

Membership curation and dependency editing need their own interaction-design
attention in the existing follow-ups below. Their design is not decided by this
creation form. They are distinct workflows: membership does not declare an edge,
and a dependency may involve a task outside the Thread. No new dependency between
these tasks is implied; existing Threads can be curated independently of creation.

## Acceptance criteria

- [x] Threads-page c and New Thread palette action stage title, description, goal and appropriate optional metadata before one explicit submission through the originating Service.
- [x] Typed Thread creation/graph receipts and failures retain committed/partial intent without UI retry loops, duplicate writes or invented task lifecycle semantics.
- [x] Creation supplies no task members and exposes no task picker, create-task
  action or dependency inputs.
- [x] Successful creation persists empty membership and unstarted lifecycle;
  existing task documents and repository-global dependency edges remain
  unchanged.
- [x] Portable and real guarded-adapter tests cover identity/capability drift, refusals, committed/uncertain results, origin switches, watcher landing and keyboard/paste/narrow modes with normal gates and docs.

## Out of scope

Task selection/membership, task creation, dependency editing/sequencing, combined
graph authoring, bulk compose/apply and graph layout.

## Implementation checkpoint

- [Thread creation](../../internal/tui/thread_create.go) is a concrete fifth
  consumer of the shared host and text-only controls. Page-local `c` and New Thread
  collect title, description, goal, optional target date and tags. The submitted
  `NewThreadParams` deliberately has no Tasks, Body, Template or lifecycle input.
  No new root-model result or displacement branch, core port, CLI/wire or persisted
  schema was needed.
- Core owns authoritative graph/Thread validation and bounded conflict retry.
  The typed receipt keeps Changed/planned evidence separate from Committed;
  known uncommitted refusals allow deliberate correction, while committed/unknown
  outcomes and marked creation failures retain non-replayable recovery. Missing
  optional creation currently yields an unclassified core error, so the form
  retains safe recovery rather than guessing retryability from message text.
- [Independent durable/portable tests](../../internal/tui/thread_create_test.go)
  verify metadata/scaffold, empty unstarted membership, exact task/dependency
  bytes and modes, real read-only/broken-graph/unreadable-Thread refusals, post-dispatch
  identity/authorization changes, capability source binding, proposals and actual
  committed failures, stale/duplicate results, same-ID incoming-space isolation,
  watcher/filter landing, narrow modes and live-program Keep/Discard/create.
  Preview/unconfirmed/contradictory receipts are explicit adapter-defense probes,
  not claims about actual filesystem producers.
- Shared host, retained-text, delayed/native input and visibility tests now cover
  all five kinds. The shell's creation hint now follows the host's actions instead
  of advertising only tasks/epics; README, key help and the
  [owning guide](../../docs/architecture/tui.md#thread-creation) describe the scope.
- Compiler-valid overlays adding an implicit task member, confusing Changed with
  Committed, or removing displaced recovery inspection protection each fail the
  intended durable, proposal/refusal or incoming-source regression. The source
  implementation was never mutated by these probes; normal focused race checks
  provide the restored-code baseline.

Validation passed: `just build`, full `just test` race suite, `just lint` (zero
issues), focused creation/native-input and restored mutation-case race checks,
planning/link/audit lint and whitespace checks. The build emitted the existing
nonfatal module stat-cache permission warning. Initial test expectations were
corrected for the actual Thread scaffold heading and incremental renderer output.

Membership and sequencing remain separate design work, not hidden follow-ups in
this form. The parent owns the combined creation walkthrough, not this slice.

## Independent review checkpoint

Both reviews completed on 2026-10-24 with no actionable product defects:

- [Codex](../audits/6gptxh01kx8x-2026-10-24-tui-empty-thread-creation-implementation-codex.md):
  guarded delivery confirmed; accepted and closed. Independent whole-scaffold and
  preexisting-tree checks, selected context, proposal/commit/core-retry, actual
  cleanup, native input and same-ID incoming-space probes passed. A displaced-result
  reload mutation exposed incoming reads after executed follow-ups; restoration passed.
- [Antigravity](../audits/6gptxh07vyj6-2026-10-24-tui-empty-thread-creation-implementation-antigravity.md):
  guarded delivery confirmed; bounded corroboration accepted and closed. Focused
  and running-program checks, empty-create/tree-preservation and origin/footer
  probes passed; omitting target date killed the intended durable regression and
  restored green. The owner note corrects source/retry/value-copy claims and
  distinguishes executed evidence from broader assertions and a nondetailed
  independent corrected-refusal flow.

Both used independent sandboxes and reported a second systemic pass. No code fix
or new follow-up was warranted. Full project gates remain attributed to the owner,
not represented as independently repeated by the reviewers. Historical report
transfer-pending attestations are separate from the confirmed delivery receipts.
No claim is made about every platform or asynchronous interleaving. The bounded
self-review changes below followed these independently reviewed snapshots.

## Final self-review

- Removed the footer's contextual-action blacklist: page creation metadata now
  has one owner shared by palette discovery and hints. Tests pin the exact action
  and exclude Atlas/dashboard creation without inheriting a selected epic.
- Bound the read-only fixture before opening its form instead of rewriting an
  origin token. Made fake-adapter outcomes exhaustive so a misspelled scenario
  cannot silently exercise unconfirmed success.
- Shortened the owning guide and removed stale read-only/next-extraction prose.
  Kept per-kind receipts and validation explicit rather than introducing a
  universal creation schema. No additional product defect or new follow-up was
  identified; membership/sequencing already have separately scoped owners.
- Final full-suite validation exposed a runtime-test ordering race: the success
  flash can precede asynchronous selection/detail landing, and the former wait
  also matched title text already rendered by the form. Wait for the loaded
  empty-Thread rollup before quitting. The adjacent audit/research runtime test
  had the same weak wait; it now requires the created scaffold heading, matching
  the existing task/epic tests. This fixes the test oracle, not creation behavior.

These metadata, fixture, test-wait and prose changes are owner self-review, not
attributed to the external reviewers. After fixing the wait, the Thread and
audit/research runtime flows passed ten consecutive race-enabled runs. Final
`just build`, full `just test` and `just lint` passed; the PR handoff also records
planning/link/audit and whitespace validation. This slice is complete; combined
integration acceptance remains with the parent.

## Related

- [Additional creation integration](6gmshzwphadw-extend-tui-creation-to-epics-audits-research-and-threads.md).
- [Membership/lifecycle curation and its design checkpoint](6gmshzxchdzc-manage-thread-membership-and-lifecycle-from-the-tui.md).
- [Dependency/sequencing interactions](6gmshzxksjcn-edit-task-dependencies-from-the-tui-with-guarded-graph-feedback.md).
- [Thread creation use case](../../internal/core/service_thread.go).
- [TUI architecture](../../docs/architecture/tui.md#staged-task-creation).
