---
schema: 1
id: 6gjtz31j4kpg
bucket: closed
area: documentation-ownership-and-agent-routing-implementation-external
date: "2026-10-11"
updated_at: "2026-10-12"
---
# Audit: Documentation ownership and shared agent routing implementation — external — 2026-10-11

> Reviewer assignment: external. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief's bounded evidence floor rather than exhaustively mutating every new test.

> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.

> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Substitute the repository-relative assigned audit path printed in the
handoff prompt, then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="/Users/andyeschbacher/git/andy-esch/taskflow"
AUDIT_REL="planning/audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. If creation fails, report the blocker; never fall back to the
shared checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then use
the helper for fail-closed verification and transfer:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent `.git`, source-deliverable
drift, and empty reports; it copies back only the assigned audit through a same-directory atomic
rename. Do not copy anything else manually. Leave the workspace in place and report its path until
the implementation owner confirms receipt. On refusal, preserve it and report the conflict rather
than resolving it in the shared checkout.

Include the helper's attestation—workspace path, resolved Git directory, baseline commit, captured
source blob/fingerprint, deliverable, and transfer result—in the report. A report without it is
incomplete even if its technical findings are otherwise sound.

## Review brief

Independently challenge the documentation-ownership and agent-routing slice. This is one focused
external review, not a repository-wide architecture redesign. The primary lens is instruction
integrity: did moving durable rules from CLAUDE.md into AGENTS.md weaken safety, invent policy,
misroute contributors, or substitute aspirational documentation for implemented contracts?
Cross-check the small CLI/schema correction and its revision/fixture consequences.

Do not assume checked ACs, a shorter guide, or green tests establish maintainability or reliable
agent loading. Prefer demonstrated consequences over stylistic preferences; no finding quota.

## Review target

Source checkout: `/Users/andyeschbacher/git/andy-esch/taskflow`, branch `main`, base commit
`09974fbad069a2a77ac72cd8bf61b634f4be3f9f`. The review target is the captured **dirty working tree**,
including staged/unstaged edits, deleted CLAUDE.md, untracked AGENTS.md and research/planning files,
and this audit. HEAD alone omits the implementation. Use the helper's independent sandbox baseline
as the restoration target, not the source base commit.

Implementing records and governing decisions:

- [Ownership task](../tasks/6g9cz5saenme-establish-documentation-ownership-and-agent-routing.md)
- [Research prerequisite](../tasks/6gjthp9frrhq-research-documentation-ownership-and-portable-agent-routing.md)
- [Dated primary-source research](../research/6gjggjw01raw-documentation-ownership-and-portable-agent-routing.md)
- [Documentation Thread](../threads/6g9czp7g9pt3-make-documentation-layered-executable-and-agent-navigable.md)
- [ADR-0001](../adrs/0001-adopt-adrs.md) and [ADR-0008](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)

Verify this consumer inventory in the sandbox before relying on it:

- Root `AGENTS.md` replaces the base commit's `CLAUDE.md`; neither repository CLAUDE.md nor
  GEMINI.md remains. README's contributor section routes to the shared guide; its TUI section
  becomes a peer rather than a contributor subsection.
- `routines/code-quality-audit.md`, `routines/weekly-architecture-audit.md`, and
  `routines/weekly-task-sweep.md` now name AGENTS.md, use scoped reading routes, require a clean
  starting checkout, and stop on unexpected changes without reverting unknown work.
- `internal/store/fuzz_test.go` changes only the guide reference in a comment.
- `internal/domain/entity.go` owns the corrected task `depends_on` convention. Adjacent registry,
  schema, and layout comments are clarified; `internal/wire/schema_comments.json` is regenerated.
- `internal/cli/schema_test.go` adds `TestSchemaTask_DependencyGuidanceNamesShippedCommands`.
  `internal/cli/task_dependency.go` wires existing guarded dependency commands; their real-path
  regressions include `TestTaskDependAddRemoveJSONDryRunAndNoop` and
  `TestTaskDependAddRejectsCycleWithValidationExit` in `internal/cli/task_dependency_test.go`.
- `internal/wire/wire.go` declares additive revision 1.88. Updated CLI goldens should contain only
  the revision substitution and corrected convention, not unrelated payload changes.

Already queued, not delivered by this slice:

- [Architecture split](../tasks/6g6x7e2ef37r-restructure-the-architecture-documentation-into-focused-guides.md)
- [Selective package contracts](../tasks/6g9cz5sk3b0d-adopt-selective-contract-focused-go-package-documentation.md)
- [Final reconciliation and drift checks](../tasks/6g9cz5t37dc0-reconcile-documentation-projections-and-enforce-drift-checks.md)

The final reconciliation explicitly owns remaining routine-policy conflicts, including the task
sweep's existing contradictory tier/priority-write instructions. The current slice does not execute
those routines, add a general prose/link CI gate, rewrite the architecture manual, or change
product architecture. Challenge a deferral if you can show a new in-scope consequence.

## Intended contract to challenge

1. AGENTS.md is the single shared root owner for durable working rules and routing, not an
   exhaustive mutable CLI inventory. Root-tool removal is the maintainer's explicit choice.
2. Guarded persistence, adapter-neutral core boundaries, explicit capability/policy binding,
   adapter-established identity, opaque locations, coherent snapshots, injected I/O, managed
   planning writes, and limits on Git authority remain discoverable and unweakened.
3. Code/tests describe implemented behavior; accepted ADRs and compatibility documents describe
   decisions/promises; planning describes work/history. A mismatch is investigated, not silently
   resolved by saying code overrides a promise.
4. The source-ownership table names actual owners. Static link validity does not prove model-client
   loading or adherence. Research distinguishes native Claude/Antigravity support and its limits
   from Gemini CLI's separate default/import behavior; no user settings changed.
5. Routine prompts preserve planning-only scope. Existing dirty state is not permission to restore
   or discard someone else's changes, and an unexpected path stops the run without cleanup.
6. Authoring metadata recommends already-shipped guarded `task depend add/remove`, retains generic
   write restrictions, and changes no command behavior, payload shape, vocabulary, or persisted format.
   Revision 1.88's classification and generated artifacts must agree with ADR-0008.

## Mandatory evidence floor

Produce bounded evidence for each of these probes; record reproduced / falsified / unresolved.

1. **Guardrail comparison:** read `git show 09974fbad069a2a77ac72cd8bf61b634f4be3f9f:CLAUDE.md`
   and map its durable safeguards to the shared guide or an explicit linked owner. For any apparent
   omission, identify a realistic contributor action the new route could wrongly authorize.
   Distinguish intentional removal of inventories/prototype history from lost safety policy.
2. **Route walkthrough:** take six concrete changes—planning persistence, JSON contract, CLI help,
   TUI I/O, release/toolchain, and documentation—and follow the guide to actual code/checks.
   Validate local paths and anchors; identify where the route promises evidence that does not exist.
   Scan live non-planning consumers for references to removed entry points. Historical planning
   mentions are not defects merely because they name an old file.
3. **Schema production path:** run the fresh binary's `schema task --json` from a bare temporary
   directory and inspect the corrected convention; inspect runnable `task depend add/remove`
   help. Execute the focused CLI dependency tests named above as bounded corroboration. Do not
   mutate the real planning corpus or execute the scheduled routines.
4. **One compiler-valid mutation:** in the sandbox only, replace the corrected convention with
   its base-commit historical clause. Run exactly
   `go test ./internal/cli -run '^TestSchemaTask_DependencyGuidanceNamesShippedCommands$' -count=1`.
   The intended test must fail for the stale guidance, not for a compilation problem. Restore
   the captured sandbox implementation and show that exact test green. No quota of extra mutations.
5. **Wire/artifact comparison:** compare every changed golden to the source base. Verify changes
   are restricted to 1.87 → 1.88 and the task-convention correction; check schema identity and
   compatibility declarations, plus the generated comment-map delta. Execute focused schema/wire
   tests without update flags. A planned check or inherited green result is not independent evidence.
6. **Routine safety tabletop:** trace clean start, pre-existing planning edits, pre-existing source
   edits, and an unexpected source edit during a run through each updated prompt. Verify no branch
   instructs a destructive revert or submission of pre-existing work. Use textual/control-flow
   evidence, not an execution of the write-capable routines. Do not treat this as proof that a
   future model obeys prose instructions.

If verifying loader facts, use the dated research's official sources and record version/settings
qualifications. Actual client loading may remain unverified: do not change settings, authenticate,
or invoke paid/model sessions merely for this review. A version check is not a context-loading check.

## Required hostile angles

- Could the neutral guide read as granting path authority from identity/location, permission for
  generic graph/status writes, or implicit Git mutations? Follow the linked owner when deciding.
- Is safety routing so indirect that a normal scoped edit plausibly misses a relevant invariant?
  Conversely, does it claim enforceable behavior based only on manually maintained prose?
- Could the conflict policy bless implementation drift against an accepted compatibility promise?
  Could research or a checked task AC be mistaken for an accepted architectural decision?
- Are native-loader claims qualified correctly rather than treating all Gemini-powered tools or
  nested instruction loaders as interchangeable? Removal itself is an approved choice, not a bug.
- Are the newly changed routine preflight/stop instructions internally consistent with their
  planning-only scope and AGENTS.md's preservation rule? Separate introduced regressions from
  already tracked, pre-existing policy contradictions.
- Does the new schema test check useful semantic guidance, or merely fragile wording/command
  registration? Can malformed advice still satisfy it? Demonstrate consequence before assigning severity.
- Does the mechanical-looking revision churn conceal an unrelated schema/payload change or a stale
  generated artifact? Is the compatibility classification consistent with the actual delta?

Second pass: inspect the shared documentation pattern rather than repeating the checklist. Look
for an owner/projection loop, a route that sends readers back without giving a contract, enforcement
claims copied across layers, or a helper/test whose green result masks a systemic gap. Settle each
important hypothesis with evidence; do not turn future documentation aspirations into current bugs.

## Validation and restoration

Use the minter's Mandatory reviewer sandbox protocol. Source is read-only except for the helper's
final guarded one-file transfer. Its sandbox baseline checkpoint is authorized; all other commits,
staging, source restores, branch changes, and pushes are forbidden.

Build a fresh sandbox binary with `just build`. Place temporary planning fixtures, mutation backup
material, and caches outside the sandbox's tracked/untracked tree, using separate `mktemp -d`
directories. Override GOCACHE and GOLANGCI_LINT_CACHE there if your environment needs writable caches.

Required checks inside the sandbox:

```sh
go test ./internal/cli ./internal/wire ./internal/store \
  -run 'TestSchema|TestValidateSchema|TestParseSchema|TestTaskDependAddRemoveJSONDryRunAndNoop|TestTaskDependAddRejectsCycleWithValidationExit|FuzzSplitFrontmatter|FuzzUpdateFrontmatter' -count=1
just docs-check
./bin/tskflwctl -C . lint --links
git diff --check
```

The owner previously passed full race-enabled `just test`, `just lint`, `just build`, and
`just docs-check`; after the routine/comment fixes, the fresh build and focused schema/wire/fuzz
seed tests passed again. You may inherit the full-suite baseline if labeled as inherited, but must
independently perform the mandatory probes and focused checks. There are no changed optional wire
fields, repair behavior, or concurrency algorithms requiring exhaustive stress tests here.

Restore mutation probes and any generated deltas to the captured baseline, preserving the assigned
report. Only that audit may differ at verification/transfer. Keep the sandbox on failure and report
the blocker; never make the source checkout pass by reverting unknown work.

## Deliverable

Preserve this brief and fill only the Reviewer report slot. No implementation edits or follow-up
task creation. Use the sandbox's fresh `audit finding new` to create real findings and leave them
**open** for owner triage; do not pre-set fixed/tracked/settled states or close the audit.

Start with ready / needs changes / incomplete, exact snapshot, and limits; then include:

- isolation attestation: sandbox path, independent Git directory, captured baseline, source-audit
  fingerprint, verify result, and guarded-transfer result;
- verified consumer inventory with file/line or command evidence;
- findings consolidated by root cause, with reachable consequence, reproduction, evidence,
  recommendation, scope, and residual uncertainty;
- a compact probe/mutation ledger: exact edit, intended test, expected failure, observed
  killed/survived/invalid result, and restored green result;
- a bounded surviving-hypothesis ledger distinguishing tested negatives from unresolved concerns.

Separate introduced defects, coverage gaps, already tracked followups, defense-in-depth, and taste.
A no-findings verdict still needs the mandatory evidence; a blocked review is incomplete, not clean.

## Reviewer report

**Verdict: needs changes (minor).** No durable safeguard from the base `CLAUDE.md` was lost
or weakened. Every AGENTS.md route resolves to real code, checks, or tests. The wire delta is
exactly 1.87 → 1.88 plus the corrected convention, and the routine prompts no longer tell
anyone to revert. Two low-severity findings remain open for owner triage. **L1**: the new
schema regression can be satisfied by malformed advice. **L2**: AGENTS.md's authority wording
grants hand-authored published prose and declaration-only revision tests more proof value
than their owners claim. Neither finding blocks a safe merge.

Reviewed 2026-10-11 to 2026-10-12 by the external reviewer (Claude Code), entirely inside
the sandbox below.

**Snapshot.** Source `/Users/andyeschbacher/git/andy-esch/taskflow`, branch `main`, base
`09974fbad069a2a77ac72cd8bf61b634f4be3f9f`. The dirty working tree was captured as sandbox
baseline `b291d2a8c9aaeda07fee343e4838b6ef091d0581`. 62 paths differ from the base
(`git diff --stat 09974fba b291d2a8`). The fresh binary `bin/tskflwctl` (`tskflwctl version
v0.23.0-39-gb291d2a8`) was built in the sandbox with `just build`. Caches, fixtures, and
mutation backups lived in separate `mktemp -d` directories outside the sandbox tree.

**Limits.**

- I didn't verify model-client loading or adherence. One incidental observation: this review
  session was launched from the source checkout, and its harness injected that checkout's
  `AGENTS.md` as project instructions. That covers one installation only; I didn't check the
  client version or settings, and it says nothing about adherence.
- Loader and version facts in the dated research (Claude ≥ v2.1.277, Antigravity, Gemini CLI)
  were not re-fetched. They are inherited and labeled as such.
- I didn't execute the routines; the routine probe is a textual tabletop.
- `task edit` needs a TTY. Its refusal is corroborated by store tests, not an interactive run.
- `lint --links` checks planning-entity body links only. I checked the AGENTS.md, README, and
  routine links by hand.

### Isolation attestation

Created with `scripts/isolated-review-workspace.sh create --source <source> --deliverable
<this audit> --print-path`. Nothing else in the source checkout was read or written: the
harness's default shell directory is the source path, so every later command addressed the
sandbox or a temp directory by absolute path.

```text
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.TTHn5P
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.TTHn5P/.git
baseline_commit=b291d2a8c9aaeda07fee343e4838b6ef091d0581
source_blob=936306db8fbe5718bbf58823d1b6fda857596c65
source_fingerprint=08801f9ff961cef34dd03ab5b258e40b95aa9dec
deliverable=planning/audits/6gjtz31j4kpg-2026-10-11-documentation-ownership-and-agent-routing-implementation-external.md
```

- **Independence:** the in-tree `.git` is a real directory with no alternates, and
  `git worktree list` shows exactly this checkout (the helper's `assert_independent_clone`
  passed on create and every verify). `HEAD` is the helper's baseline commit; I made no other
  commits, staging, branches, or pushes.
- **Pre-review verify:** `deliverable_changed=false`, `transfer=pending`, exit 0.
- **Pre-transfer verify:** `deliverable_changed=true`, `transfer=pending`, exit 0; all
  probes were restored, and only this audit differed.
- **Guarded transfer:** runs after this text is frozen, so this file cannot record its own
  result. The transfer receipt (`transfer=succeeded` or the refusal) is reported in the
  reviewer's handoff message. The owner can confirm it with
  `cmp <sandbox>/<deliverable> <source>/<deliverable>`. The sandbox is retained until the owner
  confirms receipt.

### Verified consumer inventory

| Brief claim | Result | Evidence |
| --- | --- | --- |
| Root `AGENTS.md` replaces `CLAUDE.md`; no repository CLAUDE.md or GEMINI.md | Verified | `git ls-files` matches only `AGENTS.md`. The one other `CLAUDE.md` on disk is `.claude/worktrees/…/CLAUDE.md`, ignored by `.gitignore:24` (`git check-ignore -v`). |
| No other tracked agent-instruction file | Verified | No tracked copilot/cursor/windsurf/aider/`.gemini`/`.codex` files; `.claude/` is ignored. |
| README contributor section routes to the shared guide; TUI becomes a peer | Verified | `README.md:533` `## Contributing` links `./AGENTS.md` and `./justfile`; `README.md:540` is now `## Interactive TUI`, an H2 (base: `###` at 548). |
| README race-detector correction | Verified | `README.md:537` says race-enabled; `justfile:52-53` runs `go test -race ./...`. The unrelated `## Build / dev` block (`README.md:99`) remains and is assigned to reconciliation. |
| The three routines name AGENTS.md, use scoped routes, require a clean start, and stop without reverting | Verified | `code-quality-audit.md:70,80-88,100-103`; `weekly-architecture-audit.md:81,91-99,112-114`; `weekly-task-sweep.md:57,67-75,88-89`. |
| No live non-planning reference to removed entry points | Verified | `git grep -n -I -E 'CLAUDE(\.md)?\|GEMINI\.md' -- ':!planning/'` → no matches. Remaining planning mentions are historical or contextual (epic 24 decided items, `6fbj87001m03`'s retired premise, `6g2pjjsc8jv1`'s quoted old roster). |
| `internal/store/fuzz_test.go` changes only a guide reference | Verified | One-line comment diff at `fuzz_test.go:13`. |
| `entity.go` owns the corrected convention; adjacent comments clarified; comment map regenerated | Verified | `entity.go:68`. The base comment named `TaskStatusDirs`/`AuditBucketDirs`, which exist in neither the base nor the baseline (`git grep`), so the correction is accurate. `layout.go` now names `wire.SchemaVersion`, which exists. `TestSchemaComments_NotStale` passes, and the changed `Descriptor` map entries reach no golden or wire output. |
| New schema test exists | Verified | `internal/cli/schema_test.go:164`. |
| `task_dependency.go` wires existing commands; named regressions exist | Verified, unchanged by this slice | Both files are identical at base and baseline. The commands date from `53732d7b` (2026-08-28). Tests are at `task_dependency_test.go:45,163`. |
| `wire.go` declares additive 1.88 | Verified | `wire.go:373-376`; `SchemaRevisionCompatibility = "additive"`. The changelog-classification tests pass. |
| Goldens contain only the revision substitution and corrected convention | Verified | All 42 changed goldens diffed against base (see probe 5). |

### Mandatory evidence floor

1. **Guardrail comparison: reproduced (no lost safety policy).** I mapped each durable
   safeguard in base `CLAUDE.md` to the shared guide or a linked owner:
   - **Build and test gates and the entry point** → `AGENTS.md:38-46`.
   - **Adapter direction, appwiring, cli/ports, local exceptions** → `AGENTS.md:14-17`,
     `ARCHITECTURE.md:22-150`, with the lint rules at `ARCHITECTURE.md:112-142`.
   - **Planning-data checklist** → `AGENTS.md:24-25,76`.
   - **One `cli.App`, lazy opening, injected writers, TUI `tea.Cmd`** → `AGENTS.md:18-20`,
     `ARCHITECTURE.md:730-737`.
   - **`--json` and `schema_version` under ADR-0008** → `AGENTS.md:28-30`. `--json` is a
     global flag (`ARCHITECTURE.md:737`).
   - **Status/bucket authoritative in frontmatter and the flat id-led layout** → not restated
     in AGENTS. They live in the `schema task` convention (`entity.go:67`),
     `ARCHITECTURE.md:211-218`, and ADR-0003. AGENTS still forbids `set`/`edit` for status
     (`AGENTS.md:96-97`).
   - **AC and finding ownership, tracked destination** → `AGENTS.md:31-33,96-98` plus the
     schema conventions.
   - **Projection caveat** → `AGENTS.md:92-93`.
   - **Git authority** → `AGENTS.md:11-13`.
   - **Sentinel wrapping** → `AGENTS.md:27-28`. `schema --json` publishes the exit taxonomy.
   - **Atomic helpers** → `AGENTS.md:26`; `internal/store/atomic.go:48,84`.
   - **Surgical frontmatter** → `AGENTS.md:26-27`.

   Intentional removals: the command and verb inventories, the `pm` prototype history, the
   exit-code table, and "match the surrounding code" (taste). For each apparent omission I
   looked for an action the new route could wrongly authorize and found none. A status-directory
   move is still refused by lifecycle verbs and lint. Direct body-prose edits are discussed
   under the hypotheses below.
2. **Route walkthrough: reproduced (all routes resolve; no promised evidence missing).** Every
   local link target in AGENTS.md exists. The three anchors map to `ARCHITECTURE.md:22,64,861`,
   and "named exceptions" maps to `:144`. Six routes:
   - **Planning persistence** → checklist `:64-108`, which links real tests.
   - **JSON contract** → wire, ADR-0008 and its golden-update amendment, `golden_test.go`.
   - **CLI help** → `just docs`/`docs-check` (`justfile:80-92`).
   - **TUI I/O** → `ARCHITECTURE.md:861`, `internal/tui/nav_test.go`, `watch_test.go`.
   - **Release/toolchain** → `docs/RELEASING.md:23-40`, `just toolchain-check`.
   - **Documentation** → the ownership table plus manual link checks (not claimed as automated).

   Every command AGENTS.md quotes is registered with the expected safety class in
   `schema --json` and runs against this checkout.
3. **Schema production path: reproduced.** Run from an empty `mktemp -d` directory, the fresh
   binary's `schema task --json` exits 0 and writes nothing. It reports `schema_version` 1.88
   and the corrected convention, byte-equal to `schema_task_json.golden` apart from newlines.
   `task depend --help`, `add --help`, and `remove --help` are runnable and classified
   `mutating`. On a throwaway fixture repo (temporary `HOME`, `--no-register`):
   - `--dry-run task depend add` wrote nothing.
   - The real `add` wrote `depends_on`, and `remove` cleared it.
   - `task set --set depends_on=[]` exited 11 with or without `--force`, naming the guarded
     commands.
   - `lint --fix` reported nothing to fix and left the edge alone.

   The named focused tests pass, as do the store create/edit rejection tests
   (`dependency_persistence_test.go:40,66,96,125`). I didn't touch the real planning corpus.
4. **Compiler-valid mutation: killed.** Ledger below.
5. **Wire/artifact comparison: reproduced.** I normalized every base golden
   (`1.87`→`1.88`, old→new convention) and compared it with the baseline: 42 changed, 0
   residual. Each substituted site is a revision field: 54 envelope `const`s, `$id`,
   `x-taskflow-schema-version`, and the title in `schema_jsonschema.golden`; 39 compact
   `schema_version`s; the pretty `config_show`; and the marker. No unchanged golden mentions
   `1.87` or `schema_version`.
   - **Classification:** ADDITIVE matches ADR-0008 §3, since keys, types, vocabularies, and
     inputs are unchanged and only guidance text changed.
   - **Tests run without update flags:** `TestGolden_MachineContract`,
     `TestGolden_ProjectionContract`, `TestMachineGolden*`, `TestValidateMachineGoldenUpdate`,
     and the changelog/classification tests all pass.
6. **Routine safety tabletop: reproduced (no destructive branch).** For all three prompts:
   - **Clean start** → proceeds. `just build` writes only the ignored `bin/`.
   - **Pre-existing planning or source edits** → the pre-start `git status --porcelain` stops
     with "do not restore, stash, clean, or commit someone else's work".
   - **Unexpected source edit during a run** → the pre-PR repeat and the later validation
     checks (code-quality step 13, architecture step 13, sweep step 11) fall through to
     "STOP … exit without opening a PR". Preserving the checkout is the explicit instruction.

   No `git checkout --`, restore, reset, stash, or clean instruction remains (grep across the
   three specs). The bootloaders in `*.NOTES.md` only read the spec from a fresh `main` clone
   (`routines/README.md:39`). This is textual evidence; it doesn't show that a model obeys it.

### Probe and mutation ledger

| Probe | Exact edit | Intended test | Expected | Observed | Restored |
| --- | --- | --- | --- | --- | --- |
| Mandatory | `entity.go:68` convention → the base-commit clause "It is preservation-only until guarded `task depend add/remove` commands land: …" (text identical to base `entity.go:74`; `go vet ./internal/domain` ok) | `go test ./internal/cli -run '^TestSchemaTask_DependencyGuidanceNamesShippedCommands$' -count=1` | Fail on stale guidance | **Killed**: `schema_test.go:179` "dependency authoring must route to shipped guarded commands" | Baseline copy restored (sha1 `ce1bceaf…`, porcelain empty); same command → `ok` |
| Hostile (L1) | Convention → "Use guarded `task depend add/remove` commands once they land; until then `task set`, `task edit`, and `lint --fix` may add, remove, or reinterpret it." (compiles) | Same test, then `TestGolden_MachineContract` | Semantic test should fail | **Survived** the semantic test; only the golden `schema_task_json` drift check failed | Restored; both tests `ok` |

My own invalid probes, not product defects: the first restore failed because `cp` is aliased
to interactive mode, so I re-restored with `/bin/cp -f` and re-checked. An early bare-directory
run hung on shell aliases (`ls`, then a missing `timeout`) before the binary ever ran.

### Required checks (sandbox, independent)

- **Focused `go test`** (the brief's exact package and `-run` set) → `ok` for `internal/cli`,
  `internal/wire`, and `internal/store`. 23 named tests ran, including all those the brief
  names.
- **Checks exiting 0:** `just docs-check`, `./bin/tskflwctl -C . lint --links`, plain `lint`,
  `git diff --check` (baseline range and worktree), and `audit lint` for this audit after the
  findings were created.
- **Also run independently** rather than inherited: the full race-enabled `just test`
  (exit 0, 27 packages ok, no failures) and `just lint` (`0 issues.`).

### Classification summary

- **Introduced defects:** L2, a wording precision issue in the new guide. No safety
  regression.
- **Coverage gaps:** L1, the new regression's substring assertions.
- **Already tracked follow-ups, not counted as findings here:**
  - `weekly-task-sweep.md:59` forbids tier/priority writes, while `:186` allows them. This
    pre-dates the slice and is owned by `6g9cz5t37dc0`.
  - The routines' inline non-negotiable lists (e.g. `code-quality-audit.md:102-107` still
    names "status authoritative in frontmatter, atomic writes via `store/atomic.go`") now
    restate rules that AGENTS.md delegates. Reconciliation's "repeated architecture summaries"
    covers this.
  - README `## Build / dev` and the remaining product reference → reconciliation.
  - The architecture split → `6g6x7e2ef37r`.
- **Defense-in-depth:** see the direct-prose and helper-remote hypotheses below.
- **Taste:** the fuzz comment cites "AGENTS.md's surgical-frontmatter rule", but AGENTS
  phrases it as preserving unknown frontmatter, comments, and key order. "Match the
  surrounding code" was dropped.

### Surviving-hypothesis ledger

| Hypothesis | Status | Evidence |
| --- | --- | --- |
| The neutral guide grants path authority from identity/`Location` | Tested negative | `AGENTS.md:21-22` denies it explicitly, matching `ARCHITECTURE.md:73-76`. |
| The guide permits generic graph/status writes or implicit Git mutation | Tested negative | `AGENTS.md:11-13,31-33,96-97`; live `task set` refusal (exit 11); routine PRs are the user's standing request. |
| Safety routing is too indirect for a scoped edit | Tested negative, with a residual | The high-risk rules are inline in `AGENTS.md:9-33`. Status authority is reachable only via `schema task` or the architecture map; I found no reachable bad action. |
| AGENTS invents identity, snapshot, receipt, or capability policy | Tested negative | Each clause traces to the planning-data checklist `ARCHITECTURE.md:66-90`. |
| "Ordinary body prose may be edited directly" weakens base guidance that pointed agents to atomic `task append`/`--body` | Tested negative (defense-in-depth) | Accepted ADR-0006:978-983 and ADR-0003:290-292 already treat raw editors as detected "advisory outsiders". A stale whole-file write can still clobber a concurrent guarded write in the verify→rename window. Optional hardening: prefer `--body-file` when another writer may be active. |
| The conflict policy blesses drift against a promise, or research/ACs read as decisions | Tested negative | `AGENTS.md:62,65-70`; ADR-0001 amendment rule cited. |
| Loader claims conflate Gemini CLI, Antigravity, and nested loaders | Tested negative (textual) | Research lines 60-81 and 112-118 separate them and qualify version, settings, and ancestors. External facts were not re-fetched. |
| Other `schema <kind>` conventions are stale like `depends_on` was | Falsified for a bounded sample | Live fixture checks: tags required, epic description required, Thread creation has no status flag, `research set` refuses managed keys. Not exhaustive. |
| Revision churn hides a payload change or a stale artifact | Falsified | Probe 5; `TestSchemaComments_NotStale`; `docs-check`. |
| Routine preflight conflicts with planning-only scope or the preservation rule | Tested negative | Probe 6. The new rule stops rather than reverting even the run's own stray edit, which is more conservative. |
| Mislabeled future revisions pass the suite | Unresolved by design | ADR-0008:181-183 concedes it; it feeds L2's wording, and I didn't simulate it. |
| The review helper keeps `origin` pointed at the source, so an accidental push could create source refs | Unresolved, out of scope | Sandbox `.git/config` `remote.origin.url` = source path. The helper predates this slice and I pushed nothing; noted for the review-tooling owner. |

## Findings

#### L1. Dependency-guidance regression accepts malformed advice · **Status:** fixed

**File:** internal/cli/schema_test.go:164 | **Component:** cli/schema-tests
**Effort:** S · **Urgency:** eventually

**Classification:** coverage gap introduced with this slice's new regression. The delivered guidance itself is correct.

**Claim under test.** `TestSchemaTask_DependencyGuidanceNamesShippedCommands` (`internal/cli/schema_test.go:162-194`) says the dependency guidance "must name a live mutation route rather than a historical rollout limitation" and reports "lost generic-write restriction" on failure. The task record says the regression pins "the live guarded route and generic-write restriction". The assertions check only that the substring ``Use guarded `task depend add/remove` `` appears, that each of `` `task set` ``, `` `task edit` ``, and `` `lint --fix` `` appears somewhere in the sentence, and that `task depend add`/`remove` are runnable.

**Reproduction (sandbox, compiler-valid, restored).** Replace the convention at `internal/domain/entity.go:68` with:

```text
… DAG. Use guarded `task depend add/remove` commands once they land; until then `task set`, `task edit`, and `lint --fix` may add, remove, or reinterpret it.
```

`go vet ./internal/domain` passes. `go test ./internal/cli -run '^TestSchemaTask_DependencyGuidanceNamesShippedCommands$' -count=1` → `ok` (mutation **survived**). Only `TestGolden_MachineContract/schema_task_json` fails, with "output drift … regenerate with -update if intended". Restoring the baseline file (sha1 `ce1bceaf…`) returns both tests to green.

**Reachable consequence.** The exact-text golden is the only guard against this regression class, and it is designed to be regenerated: `internal/cli/golden_test.go` refuses `-update` only until the revision marker advances (ADR-0008 2026-09-19 amendment). At the next legitimate revision, a convention that again calls the commands future work, or one that authorizes generic graph writes, can be accepted with the dedicated "semantic" regression green. That is the same failure that left the base convention stale from 2026-08-28 (`53732d7b` shipped the commands) through 1.87. Runtime protection is unaffected: `task set --set depends_on=…` (with or without `--force`) exits 11 with the guarded remedy, and `internal/store/dependency_persistence_test.go:40,66` cover create/edit. The impact is misleading published agent guidance, not a write bypass. By reading only (not executed): the exact-phrase check also fails a benign rewording such as "with guarded `task depend add` or `task depend remove`", so the test is wording-coupled in both directions.

**Scope and uncertainty.** Test-only change. Severity depends on how carefully regenerated `schema_task_json` diffs are reviewed, which I can't observe.

**Recommendation:** Assert the regression class rather than phrases: reject rollout/future markers, require the generic commands inside the negated clause, or share the remedy with internal/core/service_task.go:719.

**Resolution:** Pinned the present-tense guarded route and complete
    generic-write prohibition, rather than command-name presence. The review's
    stale-route mutant and a separate permission-only mutant each compile and
    fail the intended test; restored guidance passes. This remains a bounded
    prose regression, not a general semantic validator.

#### L2. Authority wording treats hand-authored contracts as executable proof · **Status:** fixed

**File:** AGENTS.md:65 | **Component:** agent-guide
**Effort:** XS · **Urgency:** soon

**Classification:** introduced documentation-authority imprecision in the slice's primary deliverable. Instruction integrity, not a lost safeguard.

**Mismatch.** `AGENTS.md:65` says "Code, executable checks, and generated contracts describe implemented behavior". `AGENTS.md:60` routes entity authoring to `schema <kind>`, and `AGENTS.md:61` says "wire revision declarations and tests enforce the machine contract". Two owners this slice relies on state narrower guarantees:

- `internal/domain/schema.go:6-9` (rewritten in this slice): descriptions, examples, and authoring conventions are "hand-authored metadata … Sync tests check field membership/types; they do not prove every prose claim."
- ADR-0008 (accepted) Consequences, lines 181-183: "review must still decide whether the classification is truthful. A test can require a declaration but cannot infer semantic compatibility." The validator agrees: `validateSchemaClassifications` (`internal/wire/wire_changelog_test.go:68-94`) checks only label presence, membership in {ADDITIVE, NOT ADDITIVE}, and agreement with `SchemaRevisionCompatibility`.

**Demonstrated consequence.** At base `09974fba`, the published `schema task` convention, a generated contract under AGENTS.md's definition, said guarded `task depend add/remove` "commands land" in the future. Those commands had shipped on 2026-08-28 (`53732d7b`, `internal/cli/task_dependency.go:81`), and the committed golden pinned the stale sentence green. A reader applying `AGENTS.md:65` literally would take that text as implemented behavior and avoid the shipped guarded route. Likewise, `AGENTS.md:61` invites treating a green revision suite as evidence that an ADDITIVE label is true, which ADR-0008 explicitly disclaims. The conflict rule at `AGENTS.md:67-68` mitigates this only if the reader cross-checks a second source; the routed authoring owner gives them no reason to.

**Scope and uncertainty.** Wording-only change in `AGENTS.md`. I didn't simulate a mislabeled future revision, because ADR-0008 already concedes the limit. The consequence is interpretive: it depends on readers applying the authority sentence literally.

**Recommendation:** Qualify AGENTS.md:61 and :65: published schema prose is maintained metadata, and revision tests require a classification while review judges its truth (ADR-0008).

**Resolution:** Qualified AGENTS.md's ownership matrix and conflict policy:
    generated shapes/registries do not prove maintained prose, and revision
    tests check declarations/covered behavior while reviewers judge
    compatibility truth.

## Implementation-owner closeout (2026-10-12)

The transferred report matched the retained sandbox report byte-for-byte before owner edits.
Both findings are accepted and fixed; the reviewer's original snapshot, verdict, and evidence
remain intact above.

- L1 now pins both the present-tense guarded route and the full generic-write prohibition. The
  reported stale/future-route mutant fails the intended test at its route assertion; a separate
  permission-only mutant fails at the prohibition assertion. Both are compiler-valid. Restoring
  the real convention passes. This is a bounded phrase-based regression, not general proof of
  natural-language semantics; no product behavior or further wire revision changes were needed.
- L2 now explicitly limits generated/prose and declaration-test evidence in AGENTS.md. Compatibility
  truth still requires review under ADR-0008.
- The out-of-scope source-origin concern is tracked by
  [push-guard hardening](../tasks/6gk06e2g21hv-disable-pushes-from-isolated-review-workspaces.md), grouped
  with the existing receipt/attestation work in
  [review-workspace safety and delivery evidence](../threads/6gk06gktr1q5-harden-isolated-review-workspace-safety-and-delivery-evidence.md).
  The task explicitly targets accident prevention, not malicious-agent containment.

After the fixes, full race-enabled `just test` and `just lint` passed. Existing architecture,
package-documentation, routine-policy, and drift-check followups remain sequenced rather than
being silently declared delivered by this review.
