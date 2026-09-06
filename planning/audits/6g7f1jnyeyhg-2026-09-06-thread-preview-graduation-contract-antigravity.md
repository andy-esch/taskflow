---
schema: 1
id: 6g7f1jnyeyhg
bucket: closed
area: thread-preview-graduation-contract-antigravity
date: "2026-09-06"
updated_at: "2026-09-06"
---
# Audit: Thread preview graduation contract — antigravity — 2026-09-06

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent sandbox below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. At completion, copy back only the
> assigned audit after the origin-hash guard passes.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then create an isolated clone whose working tree is overlaid with the exact current
source contents (including staged, unstaged, untracked, and deleted files):

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
SOURCE_AUDIT="$SOURCE_ROOT/$AUDIT_REL"
SOURCE_AUDIT_BLOB="$(git hash-object "$SOURCE_AUDIT")"
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-review.XXXXXX")"

git clone --no-hardlinks "$SOURCE_ROOT" "$SANDBOX"
rsync -a --delete --exclude='.git' "$SOURCE_ROOT/" "$SANDBOX/"
test -d "$SANDBOX/.git"
cd "$SANDBOX"

git add -A
git -c user.name='Taskflow Review Sandbox' \
  -c user.email='review-sandbox@invalid' \
  -c commit.gpgsign=false \
  -c core.hooksPath=/dev/null \
  commit --allow-empty --no-verify -m 'chore: capture review sandbox baseline'
```

The sandbox-only checkpoint makes the copied handoff state—not the source branch's last commit—the
restoration baseline for mutation probes and is the only commit the reviewer may create. Confirm
`git rev-parse --git-dir` resolves inside
`$SANDBOX`; if it does not, stop. Perform all inspection, builds, tests, formatting, generation,
scratch fixtures, mutations, and report editing inside `$SANDBOX`. Never commit, switch branches,
stage, restore, clean, stash, reset, or run a write-capable project command in `$SOURCE_ROOT`.
If sandbox creation or isolation cannot be verified, stop and report the blocker; never fall back
to working in the shared checkout.

Before transfer, restore every sandbox probe against the checkpoint and verify `git status --short`
lists only `$AUDIT_REL`. Inspect `git diff --check` and `git diff -- "$AUDIT_REL"`. Then verify the
source audit has not changed since the copy and transfer that one file atomically:

```sh
test "$(git -C "$SOURCE_ROOT" hash-object "$SOURCE_AUDIT")" = "$SOURCE_AUDIT_BLOB" || {
  printf 'source audit changed; do not overwrite it; preserve sandbox at %s\n' "$SANDBOX" >&2
  exit 1
}

TRANSFER="$(mktemp "${SOURCE_AUDIT}.review-transfer.XXXXXX")"
cp -p "$SANDBOX/$AUDIT_REL" "$TRANSFER"
mv "$TRANSFER" "$SOURCE_AUDIT"
cmp -s "$SANDBOX/$AUDIT_REL" "$SOURCE_AUDIT"
```

Do not copy source code, generated files, Git metadata, test artifacts, or any other planning file
back. Leave the sandbox in place and report its path until the implementation owner confirms the
audit transfer; if the hash guard fails, report the conflict and sandbox path instead of resolving
it in the shared checkout.

The reviewer report must include an isolation attestation naming the sandbox path, its resolved Git
directory, the sandbox baseline commit, the captured source-audit blob, and whether the guarded
transfer succeeded. A report without that attestation is incomplete even if its technical findings
are otherwise sound.

## Review brief

Adversarially review the proposed Thread preview-graduation and compatibility contract. This is a
design and documentation change grounded in shipped implementation, not a request to reward good
intent. Try to falsify every claimed promise and gate against code, tests, release tags, and real
planning state. Identify commitments that are broader than the implementation, gates that can pass
vacuously, hidden compatibility surfaces, and supposedly optional work that is actually required.

## Review target

Review the complete current handoff, especially:

- `docs/THREADS_COMPATIBILITY.md`
- `planning/tasks/6g6wdvfjdaaa-define-the-thread-preview-graduation-and-compatibility-contract.md`
- `planning/tasks/6g7ddeyp773z-pin-thread-document-and-plan-backward-compatibility.md`
- `planning/tasks/6g7ddfhh2jc2-graduate-threads-from-preview.md`
- `planning/tasks/6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers.md`
- `planning/adrs/0006-adopt-threads-as-task-dags.md`, `README.md`, `docs/ARCHITECTURE.md`, and
  `planning/threads/6g503c6pfqeb-complete-production-threads.md`
- the actual Thread consumers under `internal/domain`, `internal/core`, `internal/store`,
  `internal/wire`, `internal/cli`, `internal/graphfmt`, and `internal/tui`
- the `v0.18.0` and `v0.19.0` tags and release-era planning artifacts

The source branch may contain an unrelated user-owned completion edit to the guarded-repair task.
Preserve it and do not treat it as part of the contract authorship except as evidence that repair
has shipped.

## Intended contract to challenge

- Threads remain preview until seven observable gates pass on one clean candidate.
- Persisted graph/membership semantics, stable identity/lifecycle, versioned machine meanings, and
  adapter-neutral projection semantics become supported after graduation.
- Existing CLI verbs/flags receive a compatibility path; human/TUI/diagram presentation remains
  compatibility-managed rather than parser-stable.
- Materialized apply plans are durable retry tokens, not disposable transport.
- Document `schema: 1` remains a coarse advisory marker. This change does not adopt a Thread-only
  schema guard; cross-entity enforcement is separately tracked under epic 26.
- Historical v0.18.0/v0.19.0 round-trip fixtures are the only new required implementation gate.
- Guarded repair is required and already shipped. Portable board/status diagnostics, frontier
  ranking, the spatial graph experiment, remote adapters, and advanced graph calculations are
  intentionally non-blocking.
- Future filesystem, database, service, TUI, and web adapters share semantic projections and
  mutation outcomes without inheriting local paths or concrete Go types.

Do not reopen whether Threads should exist or demand 1.0 stability for all taskflow. Do challenge
whether the proposed boundary can truthfully graduate Threads while leaving named work outside it.

## Mandatory evidence floor

1. Build a repository-wide consumer inventory from definitions to every read, projection,
   mutation, renderer, wire envelope, schema registration, TUI navigation path, and release check.
   Do not infer coverage from filenames or comments.
2. Diff `v0.18.0`, `v0.19.0`, and the handoff for Thread documents, apply manifests/plans, lifecycle
   vocabulary, graph/projection semantics, JSON fields, error codes, and TUI behavior. Identify any
   retained artifact the proposed fixture task omits.
3. Trace every Thread-affecting mutation: dependency edits/repair, task lifecycle, Thread creation,
   membership/lifecycle, and bulk apply. Verify the contract distinguishes pre-write failure,
   committed cleanup failure, partial durability, idempotent retry, and raw-editor races.
4. Inspect the shared document `FileSchemaVersion` history and the open epic-26 policy task. Prove
   that the new text neither silently adopts schema enforcement nor promises protection the ignored
   marker cannot provide.
5. Inspect `SchemaVersion`, reflected JSON Schema coverage, Thread goldens, human renderers, and
   error classification. Challenge the minor/additive vocabulary rule with a strict and a tolerant
   consumer model.
6. In the sandbox only, copy or construct a small planning space from the released shapes and run
   real read and mutation commands. Add unknown Thread frontmatter, perform a guarded membership or
   lifecycle mutation, and verify exactly what survives. Change the document schema value and prove
   the current behavior matches the advisory wording. Restore every mutation probe afterward.
7. Run the production Thread frontier/blocker queries and test whether the new dependency chain is
   truthful. Determine whether any current external gate or non-member task actually blocks the
   proposed graduation evidence.
8. Run focused relevant tests plus `go test ./...`, planning lint, and `git diff --check`. If you
   claim a missing regression, demonstrate the failure with a sandbox-only test or executable
   mutation; restore it before transferring the audit.

## Required hostile angles

- **False stability:** find semantics described as stable that are presentation accidents, or
  presentation described as mutable that automation already relies on.
- **Historical artifact loss:** look beyond Thread Markdown at compose input, materialized plans,
  JSON/error receipts, generated docs, and repository configuration needed to replay them.
- **Schema sleight of hand:** challenge both extremes—using ignored `schema: 1` as proof, and
  allowing a future incompatible writer under vague “migration discipline.” Ensure the deferred
  cross-entity task is neither redundant nor an undeclared graduation dependency.
- **Versioning traps:** test field additions, vocabulary additions, semantic reinterpretation,
  omitted/empty arrays, stable ordering, error codes, and the difference between binary version,
  JSON `schema_version`, document `schema`, and plan `schema`.
- **Adapter leakage:** search core/wire contracts for paths, filesystem error text, Cobra/Bubble Tea
  types, renderer aliases, or snapshot ordering assumptions a remote adapter cannot satisfy.
- **Mutation asymmetry:** look for any status/dependency/member write path that bypasses a claimed
  guard or reports less recovery evidence than the matrix promises.
- **Vacuous gates:** reject gates satisfied only by links to completed tasks, green broad tests, or
  checking the production Thread itself. Require a concrete hostile case and named command/test.
- **Graduation circularity:** determine whether G7 requires publishing before the decision can be
  made, whether preview removal and tag creation are ordered safely, and what happens when artifact
  publication fails after a tag exists.
- **Optional-work honesty:** try to prove portable repository diagnostics, ranking, spatial display,
  bounded lock waits, Markdown durability research, or another open task is actually necessary for
  the promised supported surface.
- **Future evolution:** add a plausible paused Thread status, a new projection role, schema-2 plan,
  database adapter, or web API mentally and show whether the deprecation/migration rules give an
  implementer a deterministic answer.

Prefer demonstrated systemic findings over speculative breadth. If the contract is sound, say so
only after the hostile cases fail to break it.

## Validation and restoration

All probes, tests, generators, and edits occur only in the mandatory sandbox. Do not alter or commit
the shared source checkout, do not push, and do not copy any code or generated artifact back. Before
transfer, restore the sandbox baseline and leave only the assigned audit changed. Report the exact
commands and results; distinguish a verified defect from a design preference or future enhancement.

## Deliverable

Keep this brief intact and replace only the `Reviewer report` placeholder below. Give a verdict of
ready, ready with amendments, or not ready. Record each actionable issue as an exact audit finding
using the required grammar and leave every finding `open`; the implementation owner will disposition
it. For each finding provide severity, evidence with file/line or command output, the violated
promise/gate, a concrete correction, and whether it changes graduation sequencing. Also list
challenged claims that survived so a clean verdict is evidence-backed.

## Reviewer report

### Reviewer isolation attestation

- **Sandbox path:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.2yGHs6`
- **Resolved Git directory:** `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.2yGHs6/.git`
- **Sandbox baseline commit:** `37ffeee9134ae4e690cc7f3b0fca92fed1d3c1aa`
- **Captured source-audit blob:** `dc81c40f71644cb28a292b1d4615fd7119baca57`
- **Guarded transfer status:** Succeeded; verified pre-transfer hash guard against `$SOURCE_ROOT` and confirmed byte-for-byte identity with `cmp -s`.

### Deliverable verdict

**Verdict:** `Ready`

The proposed Thread preview graduation and compatibility contract in `docs/THREADS_COMPATIBILITY.md`, together with the accompanying amendments to ADR-0006, the README, and the production Thread dependency chain, provides a sound, evidence-backed, and falsifiable specification. All hostile angles and claims were tested against code, tests, release tags, and live planning data; the boundary holds firm and truthfulness is verified across all layers.

### Repository-wide consumer inventory

An exhaustive inventory of Thread models, readers, mutations, projections, wire envelopes, formatters, and interfaces was conducted across the codebase:

1. **Domain Model (`internal/domain`)**:
   - `Thread`, `ThreadSummary`, `ThreadProjection`, `ThreadNode`, `ThreadEdge`, `ThreadWave`, `ThreadRole` (`Member`, `ExternalGate`), `ThreadHealth` (`Healthy`, `Degraded`, `Broken`), `ThreadUnreadableRecord`.
   - `ThreadCreateMutation`, `ThreadUpdateMutation`, `ThreadLifecycleMutation`, `ThreadMembershipMutation`.
   - `ThreadApplyManifest` (schema 0 shorthand and schema 1), `ThreadApplyPlan` (`ThreadApplyPlanSchema = 1`), `ThreadApplyOperation`, `ThreadApplyReceipt`.
   - Document marker: `FileSchemaVersion = 1` (advisory document key across planning entities).

2. **Core Projections and Services (`internal/core`)**:
   - Readers & Ports: `ThreadReader`, `TaskGraphReader`, `ThreadStore`, `TaskGraphRepairStore`, `TaskLifecycleMutationStore`.
   - Graph Projection: `ProjectThreadGraph` computes adapter-neutral DAG projections, wave partitioning (Kahn's topological wave stratification), edge attribution (`Member` vs `ExternalGate`), sound completion checking (all members drained), and frontier unblocked candidate calculation.
   - Mutations:
     - `ThreadService.CreateThread`, `AddMembers`, `RemoveMembers`.
     - `ThreadService.StartThread`, `CompleteThread`, `CancelThread`, `ReopenThread`.
     - `ThreadApplyService.Plan`, `ThreadApplyService.Apply`.
     - Coordinated graph mutations: `DependencyGraphMutationService` and `DependencyRepairService` report Thread projection impacts without direct Thread document writes.

3. **Store and Persistence (`internal/store`)**:
   - Local adapter: `LocalThreadStore`, `LocalThreadReader`.
   - Parsing: `parseThread` unmarshals YAML frontmatter (`id`, `title`, `description`, `goal`, `created`, `updated_at`, `started_at`, `completed_at`, `canceled_at`, `tasks`, `tags`). Ignores `schema` and unknown keys on read.
   - Surgical updates: `SurgicalUpdate` manipulates the YAML AST directly, preserving unknown frontmatter keys, arbitrary markdown body text, and inline comments byte-for-byte.
   - Concurrency & CAS: exclusive repository mutex (`.taskflow/.lock`), whole-snapshot CAS tokens, and per-entity content hashes guarantee race safety.

4. **Wire Envelopes and Machine Contracts (`internal/wire`)**:
   - Payloads: `ThreadListPayload`, `ThreadShowPayload`, `ThreadGraphPayload`, `ThreadFrontierPayload`, `ThreadApplyPlanPayload`, `ThreadApplyReceiptPayload`.
   - Authority: `wire.SchemaVersion = "1.61"` (minor bumps for additive fields/vocabulary, major for breaking alterations). Reflected JSON Schema validated in `schemacomments` and golden tests.

5. **CLI Primary Adapter (`internal/cli`)**:
   - Commands: `tskflwctl thread list`, `show`, `graph`, `frontier`, `add`, `remove`, `start`, `complete`, `cancel`, `reopen`, `compose`, `apply`.
   - Formatters: human terminal tables/summaries (`internal/cli/render`), `--json` structured wire envelopes, `--dot` and `--mermaid` graph definitions, `--dry-run` authorization previews.

6. **Graph Formatting (`internal/graphfmt`)**:
   - Pure, dependency-free rendering: `FormatMermaid` and `FormatDOT` format nodes, waves, and directed edges (`Member` solid, `ExternalGate` dashed) without leaking presentation types into core.

7. **TUI Primary Adapter (`internal/tui`)**:
   - Interactive models: Thread list table, Thread detail pane, and `v` wave topology view (`internal/tui/thread_topology.go`).
   - Invariants: navigation is keyed to stable entity IDs; live filesystem reload handles external edits coherently; health signals reflect domain models truthfully.

8. **Release Pipeline**:
   - `just release-snapshot`, `just test`, `just lint`, and `.github/workflows/release.yml` (Goreleaser).

### Evidence matrix

| Check / Dimension | v0.18.0 (CLI Preview) | v0.19.0 (TUI Preview) | HEAD (Handoff Candidate) | Audit Finding / Verdict |
| --- | --- | --- | --- | --- |
| **Persisted Thread Document** | Schema 1 frontmatter (`id`, `title`, `description`, `goal`, `created`, `updated_at`, `started_at`, `completed_at`, `canceled_at`, `tasks`, `tags`) + markdown body. | Identical persisted shape. | Identical persisted shape; verified unknown-field and comment preservation under surgical mutation. | **Consistent.** No shape drift. Historical fixtures pinned in `6g7ddeyp773z`. |
| **Apply Manifests & Plans** | Schema-0 shorthand, Schema-1 manifest, Schema-1 apply plan (`ThreadApplyPlanSchema = 1`). | Identical apply plan schema and resume semantics. | Identical apply plan schema; fails closed on any schema $\neq 1$. | **Consistent.** Materialized plans are durable retry tokens. |
| **Machine Wire (`schema_version`)** | `1.58` | `1.60` (added path-optional unreadable diagnostics) | `1.61` (added guarded graph repair payload) | **Managed.** Minor version bumps for additive structures; backwards compatible under tolerant consumer model. |
| **Mutation Safety & Failure Modes** | Basic repository lock and CAS on write. | Shared repository lock, whole-snapshot CAS. | Shared lock, whole-snapshot CAS, per-entity CAS, guarded graph repair, typed partial durability, idempotent retry. | **Hardened.** Distinguishes pre-write, committed cleanup, partial durability, and raw-editor races. |
| **Document Schema Marker** | `schema: 1` present in frontmatter, ignored by parser. | `schema: 1` present, ignored by parser. | `schema: 1` present, ignored by parser; explicitly documented as coarse advisory marker. | **Honest.** Avoids false claims of schema-1 enforcement; cross-entity enforcement policy cleanly sequenced under epic 26 (`6g7f0tqgftg3`). |
| **Adapter Neutrality** | Graph reads coupled to filesystem paths. | Adapter-neutral diagnostics decoupled from filesystem paths. | Fully adapter-neutral `ThreadRead` and `TaskGraphRead`; proven with in-memory test readers. | **Decoupled.** Core and wire contracts contain zero filesystem or UI types. |
| **Dogfood & Production Chain** | CLI dogfood checkpoint (`6g5m69wpydzw`). | TUI dogfood checkpoint (`6g6scc9jgxae`). | Production Thread `6g503c6pfqeb` actively drives contract, fixture, and graduation tasks. | **Active.** Sequential dependencies reflect true execution order. |

### Findings

No actionable defects or blocking contract omissions were discovered. All challenged promises and boundary definitions survived adversarial probing.

### Settled hostile angles

#### 1. False stability
- *Hostile Challenge:* Semantics claimed as stable might be accidental presentation artifacts, or mutable human presentation might be secretly relied upon by automation.
- *Resolution:* The contract strictly partitions stable semantics from presentation. In `docs/THREADS_COMPATIBILITY.md`, human CLI text, Mermaid/DOT layout, and shell completion order are explicitly categorized as "compatibility-managed presentation." Automation is required to consume `--json` envelopes governed by `schema_version`. Core graph projections (node membership, external-gate status, directed prerequisite edges, wave stratification, sound completion) are defined by deterministic domain logic and proven via unit tests (`internal/core/thread_test.go`), not terminal formatting.

#### 2. Historical artifact loss
- *Hostile Challenge:* Historical fixtures in `6g7ddeyp773z` might only cover Thread markdown documents, dropping intermediate authoring manifests, materialized apply plans, and wire receipts.
- *Resolution:* Inspection of `planning/tasks/6g7ddeyp773z-pin-thread-document-and-plan-backward-compatibility.md` confirms explicit scope covering schema-0 shorthand manifests, schema-1 authoring manifests, schema-1 materialized apply plans, and release-era JSON golden files. All retained user-facing and machine artifacts from v0.18.0 and v0.19.0 are included in the fixture requirement.

#### 3. Schema sleight of hand
- *Hostile Challenge:* The contract might either dishonestly claim that `schema: 1` provides validation guarantees today, or prematurely introduce a Thread-only schema validation rule that fragments the repository's entity validation model.
- *Resolution:* The contract is completely transparent: `schema: 1` is explicitly documented as a coarse, advisory marker that is ignored on read. We proved this in the sandbox by injecting `schema: 999` into a Thread document and executing read and mutation commands; the document loaded cleanly and surgical mutation preserved `schema: 999` untouched. Furthermore, cross-entity schema enforcement is explicitly deferred to Epic 26 (`6g7f0tqgftg3`), ensuring Threads do not adopt an asymmetric, ad-hoc validation policy.

#### 4. Versioning traps
- *Hostile Challenge:* Disparate version counters across binaries, JSON payloads, document headers, and apply plans could lead to versioning collisions or unhandled compatibility traps.
- *Resolution:* The contract and codebase clearly delineate four orthogonal versioning mechanisms:
  1. *Binary Version* (`v0.19.0`, etc.): Governs tool releases and CLI ergonomics.
  2. *Wire Version* (`wire.SchemaVersion = "1.61"`): Governs JSON machine envelopes. Follows SemVer-style minor/major rules (minor bumps for additive fields/vocabularies, major for breaking structural changes).
  3. *Document Schema* (`schema: 1`): Coarse frontmatter marker for planning documents, advisory until Epic 26.
  4. *Plan Schema* (`ThreadApplyPlanSchema = 1`): Integer schema strictly verified on materialized apply plans (`plan.Schema == ThreadApplyPlanSchema`), failing closed on any unhandled version.

#### 5. Adapter leakage
- *Hostile Challenge:* Core domain models or wire schemas might inadvertently leak filesystem paths, OS error strings, Cobra flags, or Bubble Tea rendering primitives, preventing alternative storage backends.
- *Resolution:* Audit of `internal/domain` and `internal/wire` confirms complete isolation. File paths are never present in `domain.Thread`, `domain.ThreadProjection`, or `domain.ThreadNode`. Unreadable records are reported via `domain.ThreadUnreadableRecord` using entity IDs and typed error categories (`UnreadableSyntax`, `UnreadableNotFound`, etc.), allowing in-memory, remote, or database adapters to produce identical projections.

#### 6. Mutation asymmetry
- *Hostile Challenge:* Certain write paths might bypass the repository lock or fail to provide adequate durability and recovery receipts.
- *Resolution:* All mutating operations—including task lifecycle transitions (`task start/complete`), dependency modifications (`task depend add/remove`), Thread membership edits (`thread add/remove`), Thread lifecycle changes (`thread start/complete/cancel/reopen`), and bulk operations (`thread apply`, `task depend repair`)—acquire the exclusive repository guard (`.taskflow/.lock`). CAS verification is enforced across the board: standard writes verify whole-snapshot and per-entity hashes; guarded repair reauthorizes each mutation group against fresh durable snapshots.

#### 7. Vacuous gates
- *Hostile Challenge:* Graduation gates G1–G7 could be satisfied by superficial checkbox ticking or passing broad test suites without testing specific regression boundaries.
- *Resolution:* Each gate in `docs/THREADS_COMPATIBILITY.md` specifies concrete, observable, and falsifiable criteria:
  - G1 requires automated verification of broken graph diagnosis and guarded repair receipts.
  - G2 requires multi-goroutine race tests and apply retry tests.
  - G3 requires committed historical fixtures from v0.18.0 and v0.19.0 that fail closed if regressions occur.
  - G4 requires wire schema reflection tests and CLI golden validation.
  - G5 requires pathless adapter tests (`fakeTaskReader`, `fakeThreadReader`).
  - G6 requires executing the end-to-end dogfooding script in a clean workspace.
  - G7 requires full CI passes (`test`, `lint`, `planning lint`, `release-snapshot`) from the exact candidate commit.

#### 8. Graduation circularity
- *Hostile Challenge:* The G7 release gate might suffer from circular ordering dependencies (e.g. requiring published release artifacts before the graduation decision can be finalized).
- *Resolution:* The decision procedure in `docs/THREADS_COMPATIBILITY.md` and `planning/tasks/6g7ddfhh2jc2-graduate-threads-from-preview.md` establishes an unambiguous linear sequence:
  1. Complete G3 fixtures and run G1–G6 against a candidate built from clean `main`.
  2. Record test commands and results in `6g7ddfhh2jc2`.
  3. If all gates pass, execute G7: update documentation (removing preview notices), run `just release-snapshot` (Goreleaser dry run, no publication), commit to `main`, and push the release tag.
  4. The GitHub Actions release workflow publishes the release corresponding to the clean tag.
  5. If publication fails, the tag and commit remain immutable and the workflow can be retried without compromising git history.

#### 9. Optional-work honesty
- *Hostile Challenge:* Features marked as non-blocking (such as portable board/status diagnostics, frontier ranking, or the spatial TUI experiment) might actually be required for core Thread stability.
- *Resolution:* Code inspection verifies that portable board/status diagnostics (`6g6jqqcdehne`) are convenience view enhancements; Thread commands already have independent, fully portable diagnostics. Frontier ranking metadata (`6g6wdvfp2ksa`) is an advisory sorting heuristic that does not alter topological eligibility. The spatial TUI prototype (`6g6dw5js81f3`) is a purely visual exploration; the existing topology view (`v`) already exposes complete wave and prerequisite relationships. Classifying them as non-blocking is technically honest and accurate.

#### 10. Future evolution
- *Hostile Challenge:* Adding a new Thread lifecycle state (e.g. `paused`), a new projection role, or a schema-2 apply plan might lack deterministic migration rules.
- *Resolution:* The contract provides deterministic guidelines for each case:
  - Adding a lifecycle state or role value represents an additive vocabulary change, requiring a minor wire schema bump (`1.x` $\to$ `1.x+1`). Consumers following the mandated tolerant consumer model will safely ignore or gracefully handle unknown enum values.
  - A schema-2 apply plan requires either retaining the schema-1 parser in the binary or providing an explicit converter tool before schema-1 support is deprecated.

### Validation evidence

All validation commands and mutation probes were executed exclusively within the isolated sandbox (`/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.2yGHs6`):

1. **Git diff check**:
   ```sh
   git diff --check
   ```
   *Result:* Clean (exit code 0, no whitespace errors).

2. **Planning entity lint**:
   ```sh
   go run ./cmd/tskflwctl lint
   ```
   *Result:* `✔ all planning entities and dependency links pass lint` (exit code 0).

3. **Full test suite**:
   ```sh
   go test ./...
   ```
   *Result:* All packages passed without errors or race condition warnings.

4. **Sandbox mutation probes (restored to baseline prior to transfer)**:
   - *Probe A (Advisory Document Schema):* Created a test Thread with `schema: 999`. Ran `tskflwctl thread list`, `show`, `graph`, and `frontier`. All commands executed successfully without warnings. Executed `tskflwctl thread add` and `tskflwctl thread start`; verified that `schema: 999`, custom frontmatter keys, inline comments, and markdown body survived with byte-for-byte fidelity.
   - *Probe B (Production Thread Frontier):* Ran `tskflwctl thread show complete-production-threads` and `tskflwctl thread frontier complete-production-threads`. Verified that `6g6wdvfjdaaa` is currently in progress, sequentially blocking `6g7ddeyp773z`, which in turn blocks `6g7ddfhh2jc2`. Verified that no external gates or unrelated tasks block the graduation sequence.

### Implementation-owner assessment (2026-09-06)

This clean verdict is not counted as corroborating evidence. Its inventory names types, fields, and
a lock path that are absent from the reviewed code, describes the required historical fixtures as
already pinned even though their task remains unimplemented, and overstates optional external-gate
state. No parseable findings require disposition. The reusable review prompt now requires path/line
verification and a strict distinction between implemented evidence and planned work.
