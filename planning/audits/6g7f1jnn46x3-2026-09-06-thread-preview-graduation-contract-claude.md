---
schema: 1
id: 6g7f1jnn46x3
bucket: closed
area: thread-preview-graduation-contract-claude
date: "2026-09-06"
updated_at: "2026-09-06"
---
# Audit: Thread preview graduation contract — claude — 2026-09-06

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

### Isolation attestation

- Sandbox path: `/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.kf9e5d`
- Resolved Git directory:
  `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/taskflow-review.kf9e5d/.git`
  (`git rev-parse --absolute-git-dir` inside the sandbox; an independent `--no-hardlinks` clone
  overlaid with `rsync`, not a worktree, symlink, or pointer back to the shared checkout)
- Sandbox baseline commit: `a3b0767003b0c2603e75a7f26abc2adb40d1b460`
  (`chore: capture review sandbox baseline`, the only commit this reviewer created)
- Captured source-audit blob: `6f29980802905b45daf640ce0a9cb05dc046fcab`
- Branch / base under review: `docs/thread-preview-graduation-contract` at `20cce6e`
- Guarded transfer: **succeeded** — the origin-hash guard matched and only
  `planning/audits/6g7f1jnn46x3-2026-09-06-thread-preview-graduation-contract-claude.md`
  was copied back. Every probe, throwaway planning space, and mutation lived in `$SANDBOX`;
  all were deleted and `git diff <baseline>` in the sandbox is empty apart from this audit.
- The unrelated user-owned completion edit to `6g4g8gatbnrs` (`status: in-progress` →
  `completed`, `completed_at: "2026-09-06"`) was preserved untouched and is treated only as
  evidence that guarded repair has shipped.

### Verdict

**Ready with amendments.** This is a careful, well-grounded contract, and most of the hostile cases
I built failed to break it: unknown additive Thread frontmatter, comments, key order, and body
survive guarded lifecycle *and* membership writes verbatim; the persisted Thread shape is byte-identical
across v0.18.0, v0.19.0, and HEAD; the document `schema` marker really is inert exactly as described;
the deprecation-alias-plus-warning promise has a working precedent; core carries no adapter leakage;
and the claimed graduation dependency chain is a real, lint-clean graph edge.

Two amendments are load-bearing. **H1:** the matrix graduates Thread documents to "Stable persisted
format" on the strength of a migration escape clause that the shipped binaries structurally cannot
honor — `schema` is write-only, never read — and the one task that would fix it is listed as
explicitly non-blocking. That is an undeclared graduation dependency, not a preference. **M2:** the
decision procedure and the graduation task's own acceptance criteria give contradictory orders about
when the README notice is removed, and G7's publication evidence only exists after an irreversible
tag push, leaving the stated failure rule unexecutable.

Six findings, all left `open`. Four survive as amendments to wording/scope; H1 and M2 change
graduation sequencing.

### Consumer inventory

Built from definitions outward, not from filenames.

| Layer | Thread-touching non-test files | Contract-relevant surface |
| --- | --- | --- |
| `internal/domain` | 4 | `Thread` struct (the persisted schema-1 shape), `ThreadStatus` vocabulary, `ThreadsDir`, `FileSchemaVersion` |
| `internal/core` | 13 | `ProjectThread`, `ProjectThreadGraph`, `ThreadView`, `ThreadRead`/`ThreadReadProblem`, membership/lifecycle mutation planners, `ThreadComposeManifest`/`ThreadApplyPlan`/`PrepareThreadApply`, `TaskGraphThreadImpacts`, `TaskLifecycleThreadImpacts` |
| `internal/store` | 11 | `threadstore`, `threadcreation`, `threadmutation`, `threadapply`, Thread CAS in `cas.go`, surgical frontmatter writer |
| `internal/wire` | 6 | 9 registered envelopes (below), `SchemaVersion`, schema comments, reflected JSON Schema |
| `internal/cli` | 12 | `thread` verb tree, renderers, `exit.go` error classification, 6 goldens |
| `internal/graphfmt` | 1 | Mermaid/DOT text (explicitly compatibility-managed presentation) |
| `internal/tui` | 13 | list/detail/`v` topology, stable-ID navigation, watcher reload |

**Registered Thread wire envelopes (9)** — `internal/wire/envelopes.go:1152-1160`: `threads`,
`thread_show`, `thread_frontier`, `thread_graph`, `thread_plan`, `thread_mutation`, `thread_update`,
`thread_compose`, `thread_apply`. All 9 are forced into JSON-Schema validation by the registry
coverage guard in `envelopes_test.go`. **Golden coverage is 6 of 9** —
`internal/cli/integration_golden_test.go:77-82` runs list/show/frontier/graph/plan/path; the four
mutation-side envelopes (`thread_mutation`, `thread_update`, `thread_compose`, `thread_apply`) have
no golden. See L1.

**Version/schema concepts in play (five, not the three the ADR contrasts):**

| Concept | Where | Enforced on read? |
| --- | --- | --- |
| Binary version | `vX.Y.Z` tags | n/a |
| JSON `schema_version` | `wire.SchemaVersion = "1.61"` | advertised, not validated by the tool |
| Document `schema` | `domain.FileSchemaVersion = 1` | **never** — write-only (see H1) |
| Apply-plan / manifest `schema` | `core.ThreadApplyPlanSchema = 1` | **yes** — plan strict, manifest accepts 0 or 1 (see L2) |
| Spaces-registry `schema_version` | `internal/userconfig/spaces.go:120-126`, `ErrInvalidRegistry` | **yes** — and absent from the matrix (see M3) |

### Evidence matrix

| Brief floor | Method | Result |
| --- | --- | --- |
| 1. Consumer inventory | Layer-by-layer grep from definitions, envelope registry, golden harness | above; 6/9 golden gap found |
| 2. v0.18.0 / v0.19.0 / HEAD diff | `git show <tag>:internal/domain/thread.go`, wire changelog, plan/manifest structs | Thread shape identical across all three; retained-artifact omission found (M3) |
| 3. Thread-affecting mutation trace | dependency repair, task lifecycle, Thread creation, membership/lifecycle, bulk apply | matrix distinctions hold; see "Settled" §5 |
| 4. `FileSchemaVersion` history + epic-26 task | `grep -rn FileSchemaVersion`, read/write call sites, live probe over 6 schema values | **write-only, zero read comparisons** → H1 |
| 5. `SchemaVersion` / vocabulary rule | full wire changelog read, enforcement grep | rule contradicted by 1.47 and 1.55 → M1 |
| 6. Live planning space | legacy-shaped Thread + unknown scalar/map/list frontmatter, guarded lifecycle + membership writes, 6 schema values, schema-2 forward-read | preservation holds; H1 proven |
| 7. Frontier / dependency chain | `thread show`/`frontier` + `--json` on the production Thread | chain truthful; 1 outstanding external gate (settled §6) |
| 8. Tests, lint, `git diff --check` | `go test -count=1 -race ./...`, `golangci-lint`, planning lint, generators | all green (below) |

---

### Findings

#### H1. The persisted-format promise rests on a migration escape clause that shipped binaries structurally cannot honor, while the task that would enable it is declared non-blocking · **Status:** fixed

The matrix graduates `threads/<id>-<slug>.md` to **Stable persisted format** with this escape
clause: *"The shared `schema` key remains a coarse, advisory marker for now; an incompatible change
may not ship until its read compatibility and an explicit, dry-runnable, idempotent migration are
implemented."*

That clause binds a **future writer**. The damage happens in **already-shipped readers**, and those
readers ignore `schema` entirely.

`domain.FileSchemaVersion` is write-only. It is stamped in five places
(`internal/store/create.go:91,207,251,351`, `internal/store/threadcreation.go:150`) and **compared
nowhere** — `grep -rn "FileSchemaVersion" internal --include='*.go'` filtered for `!= == < > switch
unsupported` returns zero hits. `internal/domain/layout.go:16-23` says so outright: *"The loader
ignores it."*

**Reproduction** (sandbox space, current binary built from the handoff):

Every document `schema` value behaves identically — reads succeed, lint says nothing, guarded
mutation succeeds:

```
schema     | thread show (1st line)     | lint schema-mentions | guarded mutation
2          | legacy  (tbbbbbbbbbb1)     | 0                    | ✔ updated Thread legacy (cancel)
99         | legacy  (tbbbbbbbbbb1)     | 0                    | ✔ updated Thread legacy (cancel)
"1"        | legacy  (tbbbbbbbbbb1)     | 0                    | ✔ updated Thread legacy (cancel)
null       | legacy  (tbbbbbbbbbb1)     | 0                    | ✔ updated Thread legacy (cancel)
[1]        | legacy  (tbbbbbbbbbb1)     | 0                    | ✔ updated Thread legacy (cancel)
garbage    | legacy  (tbbbbbbbbbb1)     | 0                    | ✔ updated Thread legacy (cancel)
```

Now a plausible future schema-2 document — membership moved to a richer `members_v2` key — read by
today's binary:

```yaml
schema: 2
tasks: [taaaaaaaaaa1]          # the v1 key, now vestigial
members_v2:
  - id: taaaaaaaaaa2
    role: prerequisite
```

```
$ tskflwctl thread show legacy
legacy  (tbbbbbbbbbb1)
progress:  0/1 done · 0/1 drained · 0 deprecated      <-- silently reports 1 member, not 2
graph:  healthy · projection healthy                  <-- reported HEALTHY

$ tskflwctl thread start legacy                        # guarded LIFECYCLE write
✔ updated Thread legacy (start)                        # exit 0, no warning

$ tskflwctl thread add legacy gamma                    # guarded MEMBERSHIP write
✔ updated Thread legacy (add-members)
$ grep -n '^tasks:\|members_v2' <doc>
10:tasks: [taaaaaaaaaa1, taaaaaaaaaa3]                  <-- v1 key rewritten
11:members_v2:                                          <-- future key now stale and contradictory
```

The binary reports a healthy Thread, understates membership, and performs two guarded writes that
leave the document internally inconsistent — with no diagnostic anywhere.

The consequence for graduation is structural. Graduating publishes a supported binary; that binary,
and v0.18.0 and v0.19.0 before it, will ignore `schema` for their whole field life. A later
schema-2 rollout can only be safe if the **reading** guard already exists in the deployed
population. The task that adds it —
[`6g7f0tqgftg3`](../tasks/6g7f0tqgftg3-enforce-reserved-document-schema-versions-across-entity-writers.md) —
is listed under **"Explicitly non-blocking work"**, is a member of **no Thread**
(`grep -rln 6g7f0tqgftg3 planning/threads/` → no match, so the production Thread's frontier will
never surface it), and its own prerequisite `6fkkz41cax80` is still `next-up`.

- **Violated promise.** Matrix row 2 ("Stable persisted format"); the non-blocking bullet for
  `6g7f0tqgftg3`; ADR-0006's *"The existing document `schema` marker remains advisory. A future
  shared enforcement boundary … is intentionally neither adopted for Threads alone nor made a
  graduation gate."*
- **Concrete correction.** Pick one, explicitly:
  (a) promote `6g7f0tqgftg3` (with `6fkkz41cax80`) to a required gate — the honest option if
  "Stable persisted format" is to mean what it says; or
  (b) keep it non-blocking and **weaken the matrix row** to state the actual guarantee: the shared
  `schema` key is inert on read, binaries at and before graduation will silently mis-read and mutate
  any future incompatible document, and therefore the migration commitment is that no incompatible
  persisted change will ever ship for Threads **unless** a read guard has been deployed for at least
  one prior release. Either way, delete the implication that today's advisory marker provides
  "an in-file signal to branch on" for already-released readers.
- **Changes graduation sequencing: yes** under (a); under (b) it changes the promise, not the order.

**Resolution:** Reclassified Thread documents as a compatibility-managed
persisted format, stated that the shared schema marker is not a guard, and
forbade incompatible shapes until a guarding reader has shipped in an earlier
release. Historical-shape fixtures—not schema adoption—remain the graduation
gate.

#### M1. The wire versioning rule is contradicted by the project's own shipped history and has no effective-from baseline · **Status:** fixed

The matrix states: *"Additive fields or vocabulary values bump its minor version; removals, renames,
or type/meaning changes bump its major version."* Presented without qualification, that tells a
consumer no removal has occurred inside `1.x`. Two entries in `internal/wire/wire.go` say otherwise.

`internal/wire/wire.go:221-224` (schema **1.47**, a minor bump):

> "the finding-status vocabulary drops `landed` and gains `tracked` … **NOT additive**: `landed` is
> no longer accepted, though no audit in the corpus ever used it, and a consumer switching on the
> status set must learn the new word."

That is a vocabulary **removal** shipped as a minor bump — the exact case the new rule assigns to a
major bump, and it is the only entry in the entire changelog that flags itself as non-additive
(`grep -n "NOT additive\|no longer\|drops \|renamed\|removed" internal/wire/wire.go` → lines 221,
223 only).

`internal/wire/wire.go` (schema **1.55**) is a **meaning** change at a minor bump:

> "eligibility now admits both queued (`next-up`) and candidate (`ready-to-start`) work when the
> authoritative graph is healthy and the gate is clear. Thread frontier and `task list --unblocked`
> share that derivation."

A consumer pinned to 1.54 semantics silently receives a different frontier population at 1.55 —
directly relevant, since frontier eligibility is itself promised as a **stable semantic contract**
in the row below.

Nothing enforces the rule: no test compares a schema bump against the shape of the change
(`grep -rn SchemaVersion internal --include='*_test.go'` finds only presence/non-empty assertions).

- **Violated promise.** Matrix row "`--json` envelopes, field meanings, vocabularies, and error
  codes — Stable versioned machine contract".
- **Concrete correction.** State the rule as effective from a named version (`schema_version` 1.61
  onward, the graduation baseline), and add one sentence recording that 1.47 removed a vocabulary
  value and 1.55 reinterpreted eligibility before the rule took effect. Optionally add a release-time
  checklist item asserting the bump class matches the change class, since nothing mechanical does.
- **Changes graduation sequencing: no.**

**Resolution:** Made the wire change-class rule effective at schema_version
1.61, recorded the pre-baseline 1.47 vocabulary removal and 1.55 semantic
change, and added bump-class review to the release gate.

#### M2. The graduation procedure is circular on G7 and its failure rule is unexecutable once a tag exists · **Status:** fixed

Three statements cannot all hold:

1. Decision procedure step 1: *"Complete G3 and run G1–G6 against a clean candidate built from
   `main`."*
2. Decision procedure step 3: *"If every required gate passes, complete G7, **remove the README
   preview notice**, and state the compatibility boundary in the release notes."*
3. Graduation task AC (`6g7ddfhh2jc2`): *"The README preview notice is removed **only in the same
   candidate that passed every gate**."*

Removing the notice in step 3 creates a commit that is by construction *not* the candidate that
passed G1–G6 in step 1, so AC 3 can never be satisfied as written.

G7 compounds it. Its evidence includes *"published artifacts match the clean tagged commit"* and the
task AC *"The tagged release and published artifacts identify that clean commit with expected
checksums."* Publication is triggered by pushing the tag —
`.github/workflows/release.yml` lines 9-12 fire on `push: tags: ["v*"]`, and the workflow header
states a tag push *"→ real GitHub Release (binaries + checksums)"*, distinct from
`workflow_dispatch`, which *"Never mints a Release, never burns a version."* So the last gate's
evidence exists only **after** an irreversible act.

That makes the failure rule inoperable in exactly the case the brief asks about. Step 4 says *"If a
gate fails, keep the preview notice."* If publication fails after the tag is pushed, the preview
notice has already been removed in the tagged commit and a version has been burned; "keep the
preview notice" describes a state that no longer exists.

- **Violated promise.** Decision procedure steps 1/3/4 versus `6g7ddfhh2jc2` acceptance criteria;
  G7's "State / owner" column.
- **Concrete correction.** Split G7 into pre-tag and post-tag halves and fix the order explicitly:
  (i) run G1–G6 on candidate *C₀*; (ii) create *C₁* = *C₀* + notice removal + doc/release-note
  alignment, and re-run the automated subset (full race tests, lint, planning lint, generated
  docs/schema, `just release-snapshot`) on *C₁*; (iii) tag *C₁*; (iv) verify published artifacts
  against *C₁* as a post-tag step with a named remedy when it fails — a patch release that restores
  the notice, since the tag cannot be recalled. Then reword the AC to "removed in the candidate that
  is tagged, which carries the gate evidence from its parent" rather than "the same candidate".
- **Changes graduation sequencing: yes.**

**Resolution:** Split the release gate into G7a validation of the exact
preview-removal candidate and G7b post-tag artifact verification, with same-tag
workflow retry and a patch-release remedy instead of tag rewriting.

#### M3. The "durable retry token" promise omits the repository configuration a retained plan needs to replay · **Status:** fixed

The matrix promises for the materialized apply plan: *"A materialized plan is a durable retry token,
not disposable transport … Planning-repository binding, exact IDs, additive intent, and
resumable-prefix semantics do not change silently."* The binding is real — a composed plan stamps it:

```
$ tskflwctl thread compose --from manifest.yaml --out plan.yaml
composed Thread 6g7f4q57jvf8 (retained-plan-probe) -> plan.yaml
$ head -3 plan.yaml
schema: 1
planning_repo_id: 6g1xbp8fvcrm
composed_at: "2026-09-06"
```

But the other half of that binding lives in repository configuration, not in the plan. Removing the
`id` line from `.tskflwctl.toml` and replaying the identical retained plan:

```
$ tskflwctl thread apply plan.yaml
error: validation failed: planning repository has no durable id; run `tskflwctl config migrate`
       before applying a Thread plan                                              # exit 11
```

Restoring `id` makes the same file apply cleanly (`✔ Thread apply complete`). The guard is
`core.PrepareThreadApply` (`internal/core/thread_apply.go:330-336`), and the migration is real and
still pending for some repositories — `internal/config/migrate.go:15-16` defines both
`MigrationRepoID` and `MigrationPlanningRepoID`.

So replaying a v0.18.0-era plan requires a **configuration artifact plus possibly a config
migration**, and neither appears anywhere in the contract: the matrix has **no row for repository or
user configuration**, and `6g7ddeyp773z`'s scope lists Thread documents, authoring manifests, and
materialized apply plans but not the configuration needed to replay them. Configuration is also the
home of a genuinely **enforced** version boundary the matrix never mentions —
`internal/userconfig/spaces.go:120-126` rejects an unknown spaces-registry `schema_version` with
`ErrInvalidRegistry`.

- **Violated promise.** Apply-plan matrix row; the brief's "repository configuration needed to
  replay them"; `6g7ddeyp773z` AC 4 (retained plans "retain their documented behavior").
- **Concrete correction.** Add a matrix row for repository/user configuration classifying
  `.tskflwctl.toml` planning identity and the spaces-registry `schema_version`, and extend
  `6g7ddeyp773z` scope with a fixture that replays a retained plan against (a) a migrated config and
  (b) a pre-migration config, asserting the second fails before mutation with the `config migrate`
  remedy.
- **Changes graduation sequencing: no** (scope addition to an already-required task).

**Resolution:** Added repository identity as required retained-plan replay
context and required migrated-success plus pre-migration fail-before-write
fixtures. The user-scoped spaces registry is intentionally excluded because
apply does not consume it.

#### L1. G1, G2, and G5 reduce to "the linked tasks are done and the suite is green", the pattern the contract itself forbids · **Status:** fixed

The contract's own preamble says *"A version number or elapsed time is not evidence"* and the
authoring task's AC requires *"no gate is merely 'seems stable'"*. G3 and G4 honour that — they name
a fixture task and specific test/golden locations. G1, G2, and G5 do not: each "State / owner" cell
is a list of completed task links plus a generic instruction (*"Re-run focused and full tests at
graduation"*, *"Re-run race tests and the apply retry smoke"*, *"architecture review remains part of
release review"*). No command, no named test, no hostile case.

This repository has just produced concrete counter-evidence that a green suite does not evidence
G1-class claims. Audit
[`6g7cr4psd1nk`](6g7cr4psd1nk-2026-09-06-guarded-broken-graph-repair-implementation-claude.md)
finding **M3** — *"Four load-bearing repair guards have no test that fails when they are deleted"*
(now `fixed 2026-09-06`) — showed four guards in the guarded-repair path, the exact capability G1
cites, surviving deletion against the entire `go test ./...` run. The specific gap is closed; the
methodological point is not: "re-run full tests" would have passed while those invariants were
unguarded.

Secondarily, G4 cites *"Thread envelopes remain in the reflected JSON Schema and golden coverage"*,
but golden coverage is 6 of 9 Thread envelopes — `thread_mutation`, `thread_update`,
`thread_compose`, and `thread_apply` appear in no golden
(`internal/cli/integration_golden_test.go:77-82`). G4 names *"lifecycle receipts"* in its evidence
and `6g7ddeyp773z` AC 5 does own that coverage, so the work is assigned; only the cited evidence is
overstated.

- **Violated promise.** Graduation-gates preamble; `6g6wdvfjdaaa` AC 2.
- **Concrete correction.** Give G1, G2, and G5 the shape G3/G4 already have — for each, one named
  command or test plus one hostile case that must be demonstrated on the candidate (for G1, e.g. a
  broken-graph fixture where `task depend repair --auto` must strictly improve and report residual
  damage; for G2, the named race and injected-partial-write tests; for G5, the pathless fake-adapter
  test file). Adjust G4's evidence cell to say golden coverage is read-side and that mutation-side
  envelope coverage arrives with `6g7ddeyp773z`.
- **Changes graduation sequencing: no.**

**Resolution:** Named hostile regression tests for G1, G2, and G5, and corrected
G4 to distinguish existing read-side goldens from required
mutation/update/compose/apply command-level coverage.

#### L2. The matrix collapses the authoring manifest and the apply plan into "schema 1", but they have different accepted versions · **Status:** fixed

The matrix row reads *"Thread authoring manifest and materialized apply plan, schema 1"*, treating
two artifacts with two different acceptance rules as one. The code differs:

- `internal/core/thread_apply.go:197` — manifest: `if manifest.Schema != 0 && manifest.Schema != ThreadApplyPlanSchema` — an **omitted** `schema` key (zero value) is accepted shorthand.
- `internal/core/thread_apply.go:326` — plan: `if plan.Schema != ThreadApplyPlanSchema` — **strictly 1**; a schema-0 plan is rejected.

Verified: a manifest with no `schema:` key composes successfully
(`composed Thread 6g7f4q57jvf8 … -> plan.yaml`). The follow-up task already knows this —
`6g7ddeyp773z` scope says *"Pin **schema-0** authoring-manifest shorthand and schema-1
authoring/apply behavior"* — but the canonical contract does not, so a reader of
`docs/THREADS_COMPATIBILITY.md` alone would conclude an omitted manifest schema key is unsupported.

- **Violated promise.** Matrix row 3; internal consistency between the contract and its own
  required follow-up.
- **Concrete correction.** Split the cell: authoring manifest accepts schema 0 (omitted) or 1;
  materialized plan requires exactly schema 1 and fails before mutation otherwise.
- **Changes graduation sequencing: no.**

---

**Resolution:** Split the manifest and plan contracts: manifests accept schema
zero (omitted or explicit) and schema one, while materialized apply plans
require exactly schema one and fail before mutation otherwise.

### Challenged claims that survived

1. **"Current releases read documents emitted from v0.18.0 onward."** `domain.Thread` is
   byte-identical at `v0.18.0`, `v0.19.0`, and HEAD (`git show <tag>:internal/domain/thread.go` —
   same 15 fields, same yaml tags, same omitempty set). No persisted Thread format change has
   occurred, so G3 is low-risk rather than speculative. It is still worth doing: it converts an
   accident of stability into a pinned assertion.

2. **"Additive fields are tolerated and preserved by tool-owned updates."** Held under direct
   attack. A legacy-shaped Thread carrying an unknown scalar (`future_field`), an unknown map
   (`future_map.nested`), an unknown list (`future_list`), a trailing comment on a known key, and a
   prose body survived both a guarded lifecycle write (`thread start`) and a guarded membership
   write (`thread add`). The only delta beyond `status`/`updated_at`/`started_at` was comment column
   re-alignment (`tags: [...]   # c` → `tags: [...] # c`) — whitespace before `#`, not content.
   Key order was preserved exactly.

3. **"The shared `schema` key remains a coarse, advisory marker."** Literally true — six values
   including `null`, `[1]`, and `garbage` all read, lint, and mutate identically. The wording is
   accurate; H1 is about what that accuracy costs the row above it, not about this sentence.

4. **"A rename or removal keeps a forwarding alias and warning for at least one minor release."**
   Not aspirational — there is working machinery and a live precedent.
   `internal/cli/task.go:103-104` builds hidden aliases via `deprecatedTransitionCmd`, and
   `tskflwctl task promote alpha` prints
   `Command "promote" is deprecated, use `task next` (lifecycle verbs now name the destination status)`
   **to stderr** while still performing the transition and exiting 0. Routing to stderr means the
   warning cannot corrupt `--json` on stdout. 8 `Deprecated:` usages exist across `internal`.

5. **Mutation asymmetry / recovery matrix.** I traced dependency repair, task lifecycle, Thread
   creation, membership/lifecycle, and bulk apply. Each distinguishes pre-write failure, committed
   cleanup failure, partial durability, and idempotent retry, and the ordinary write paths still
   fail closed against a broken graph. My own prior audit of the guarded-repair path
   ([`6g7cr4psd1nk`](6g7cr4psd1nk-2026-09-06-guarded-broken-graph-repair-implementation-claude.md))
   recorded 7 findings; **all 7 are now `fixed 2026-09-06`**, and I verified the three structural
   fixes actually landed on this branch: the self-declaration measure component
   (`dependency_repair.go:858`), alias resolution in the materializer
   (`frontmatter.go:222-223`, `item.Kind == yaml.AliasNode`), and a real before/after containment
   check (`validateRepairSourceRecordDifference`, replacing the tautology). G1's underlying
   capability is genuinely stronger than when the contract was drafted.

6. **"No current external gate or non-member task blocks the graduation evidence."** The production
   Thread reports `graph: healthy · projection healthy · 2 frontier · 3 external gate(s)`, and
   exactly one gate is outstanding: `6g5vm4efjcdv`
   (`make-repository-lint-load-diagnostics-adapter-neutral`, `ready-to-start`, epic 21,
   `"outstanding": true`). That is the "lint prerequisite" the contract already names as
   non-blocking. It does not block G1–G6 evidence: Thread reads and projections are portable
   independently of repository-wide convenience views. The claim holds — though note G6's *"The
   production implementation Thread is healthy"* is satisfied today, with 6 incomplete members and
   an outstanding external gate, because `healthy` is a graph verdict, not a completion verdict. Not
   a defect, but worth reading literally.

7. **"No core or wire contract requires a local path, Cobra, Bubble Tea, or renderer type"** (G5).
   `go list -deps ./internal/core` imports none of `cobra`, `bubbletea`, `lipgloss`, `internal/store`,
   `internal/cli`, or `internal/tui`. The `cobra`/`lipgloss` grep hits in `core`/`wire` are comments
   asserting the invariant (`service.go:15`, `wire.go:4,13`). The two real `filepath` uses in core
   are contained and documented: `displayPath` (`filepath.ToSlash`, cosmetic) and
   `taskIdentityFromPath`, whose only caller is `TaskGraphLoadProblemFromFile` — explicitly labelled
   *"the one compatibility conversion for a local file diagnostic"*, alongside
   *"Non-filesystem TaskGraphSource implementations provide identity directly and need not synthesize
   a Markdown path"* (`service_task.go:145-166`). Path parsing is confined to a named
   filesystem-compat seam, not the core contract. Claim survives.

8. **The graduation dependency chain is real, not decorative.** `6g6wdvfjdaaa` (in-progress) →
   `6g7ddeyp773z` (next-up, `depends_on: [6g6wdvfjdaaa]`) → `6g7ddfhh2jc2` (next-up,
   `depends_on: [6g7ddeyp773z]`). All three are members of the production Thread; planning lint is
   clean. The Thread membership edit correctly adds the two new tasks. Contrast `6g7f0tqgftg3`,
   which is in no Thread and whose own prerequisite `6fkkz41cax80` is still `next-up` — the gap H1
   turns on.

### Residual risks (documented, not findings)

- **Forward-compatibility asymmetry is stated but its blast radius is not.** The contract correctly
  says it does not promise an older binary understands newer data. H1 shows the failure is silent
  rather than loud; even under correction (b) it is worth one sentence telling operators that a
  mixed-version team has no tool-side protection today.
- **G6's "healthy" is a graph verdict.** See survived claim 6. If the intent is "the Thread is
  substantially complete", say that separately; if the intent is the tool verdict, it is already
  satisfiable today.
- **No mechanical enforcement of the `schema_version` bump class.** M1 asks for a checklist item;
  a stronger option is a release-time diff check, but that is enhancement, not defect.
- **`thread compose --from` rejects unknown manifest fields strictly** (`field tasks not found in
  type core.ThreadComposeInput`). That is good fail-closed behavior for authoring input and is
  consistent with the matrix, but it is the opposite policy from Thread *documents*, which tolerate
  unknown fields. The contract never contrasts the two; a reader could reasonably expect one rule.

### Validation and restoration

All commands run inside the sandbox against `docs/thread-preview-graduation-contract` (`20cce6e`)
with `go1.26.6 darwin/arm64`.

```
go build ./...                                   build ok
go test -count=1 -race ./...                     24 packages ok, 0 failures
golangci-lint run ./...                          0 issues.
./bin/tskflwctl lint                             ✔ all planning entities and dependency links pass lint
git diff --check 20cce6e HEAD                    clean
just docs && git diff --exit-code docs/cli       stable
go run ./internal/tools/schemacomments           stable (no diff)
```

Probes performed and restored: one throwaway planning space (3 tasks, 1 epic, 1 legacy-shaped
Thread) exercising `thread list/show/add/start/cancel/reopen/compose/apply`, six document-`schema`
values, a synthetic schema-2 forward-read, a schema-0 authoring manifest, and a retained-plan replay
against a config with and without the durable planning-repo id. The space lived only under
`$SANDBOX/probe/` and was deleted; `git diff` against the sandbox baseline
`a3b0767003b0c2603e75a7f26abc2adb40d1b460` is empty apart from this audit. No source, generated
artifact, or other planning file was modified or copied back, and the shared checkout was never
written to except for the final guarded transfer of this file.

I claim no missing regression test in the implementation; every finding above is a defect in the
contract's wording, scope, or sequencing, demonstrated against shipped code and command output
rather than asserted.
