---
schema: 1
id: 6gcdj3xgt95z
bucket: closed
area: command-safety-annotations-implementation-claude
date: "2026-09-21"
updated_at: "2026-09-22"
---
# Audit: Command safety annotations implementation — claude — 2026-09-21

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
>
> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.

> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.

> Review-effectiveness floor: execute the exact mutation each new regression test claims to kill and require that test to fail; exercise newly added optional wire branches with non-default values in semantic validators; actually run every emitted repair command against the state that recommends it; and use coordinated mutations when a nearby call site would otherwise preserve an architectural invariant accidentally.

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
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/<your-assigned-audit-file>.md"
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

Adversarially review the implementation of `make-the-command-safety-annotations-load-bearing`.
Do not merely confirm that the new tests pass. Try to demonstrate a runnable command, flag mode,
custom pre-run path, direct filesystem write, secondary adapter, or long-lived TUI path that can
mutate while classified read-only or can bypass command-safety binding. Conversely, challenge
whether legitimate non-Cobra adapters have been accidentally coupled to or blocked by CLI policy.

## Review target

Review the complete working-tree diff on branch `feat/load-bearing-command-safety` against `main`,
including planning and generated artifacts. Begin with a repository-wide consumer inventory of:

- every runnable Cobra command, including hidden, deprecated, generated help/completion leaves,
  the runtime-only `__complete` transport, aliases, and commands whose safety changes by flags;
- every `PersistentPreRunE`/`PreRunE` route and whether it binds the final selected command before
  discovery or execution;
- every mutation route through `core.Service`, `core.ConfigurationService`,
  `core.SpaceRegistryService`, direct CLI filesystem/config calls, and TUI/workspace opening;
- every constructor/consumer of `store.FS`, `configstore.FS`, and `spacestore.FS`, identifying
  which composition roots inject authorization and why an omitted callback is or is not sound; and
- every machine-contract consumer and generated artifact affected by `SchemaCommand` and revision
  1.71.

Key implementation files include `internal/cli/command_safety.go`, `internal/cli/root.go`, custom
pre-run command files, `internal/store`, `internal/configstore`, `internal/spacestore`,
`internal/wire/schema.go`, `internal/wire/wire.go`, schema render/goldens, and the architecture and
agent guidance. Follow the inventory wherever the code leads; this list is not an allowed scope
boundary.

## Intended contract to challenge

- Every stable runnable command has exactly one recognized capability: `read-only` or `mutating`.
  Group-only nodes need no capability. Cobra's runtime-only completion transport is bound and
  annotated read-only when it materializes.
- Binding is fail-closed and happens before repository discovery or use-case execution, including
  every custom persistent pre-run.
- A read-only command cannot enter a mutation use case through the planning filesystem,
  configuration, or space-registry adapters. The rejection also applies to dry-run mutation paths.
  Direct init, durable Thread-plan output, and the mutating TUI launch are explicitly authorized.
- Secondary adapters remain framework-neutral: they receive an optional `func() error`, know
  nothing about Cobra, and remain reusable by TUI, tests, and future web adapters.
- Safety is a command capability, not a per-invocation write claim. Thus `lint` is mutating because
  `--fix` can write; read-only invocations of a mutating-capable command remain legal.
- Safety does not derive `--dry-run` support. Preview behavior remains command-specific because
  interactive mutating commands cannot meaningfully preview.
- `schema --json` revision 1.71 additively publishes a deterministic, exhaustive runnable-command
  surface with path, safety, hidden, and deprecated fields. Exact-revision JSON Schema, human
  guidance, generated CLI docs, and machine goldens agree.
- Non-goals: a generalized command descriptor registry, authentication/security authorization,
  flag-level command metadata, and displaying safety in ordinary `--help` prose.

## Mandatory evidence floor

1. Build an explicit command/pre-run/mutation consumer inventory from the sandbox source. Do not
   infer completeness from grep counts alone. Exercise representative ordinary, hidden,
   deprecated, generated completion/help, custom-pre-run, read-only, mutating, dry-run, no-op,
   config, space-registry, direct-write, and TUI-start paths.
2. Run the targeted safety tests and the full race-enabled suite. Run golangci-lint, planning lint,
   `git diff --check`, schema-comment generation verification, CLI-doc generation verification,
   and JSON Schema validation. Record exact commands and results.
3. Inspect `schema --json` semantically: verify deterministic ordering, unique paths, recognized
   enums, meaningful hidden/deprecated non-default rows, generated help/completion coverage, and
   agreement with the Cobra tree. Do not accept only default-valued DTO fixtures as evidence.
4. Perform sandbox-only mutation testing. At minimum, remove or falsify one ordinary command tag,
   one custom-pre-run bind, one store authorization call, one config/space authorization call, and
   the direct-write/TUI authorization where feasible. Demonstrate which focused test fails for
   each. If a mutation survives, report the systemic gap rather than restoring confidence from an
   unrelated failing golden.
5. Probe a deliberately read-only command that calls both a dry-run and real mutating service path
   and prove no durable bytes change. Probe a legitimately mutating command to ensure authorization
   does not reject it. Include home registry/config paths, not only task frontmatter.
6. Examine concurrency and lifetime: run with the race detector and reason about the mutable
   invocation safety state captured by callbacks during a long-lived TUI and any plausible command
   reuse. Separate a demonstrated race from a hypothetical unsupported embedding.
7. Verify every named symbol, test, generated artifact, and claimed capability against exact source
   or command evidence. Separate shipped behavior from planning intent.

## Required hostile angles

- Find a runnable leaf omitted by the tree walk, dynamically registered after schema assembly, or
  represented only by an alias. Decide explicitly whether `__completeNoDesc` and aliases require
  separate machine rows or are correctly represented by one canonical command.
- Challenge Cobra hook semantics: child persistent hooks replacing the root hook, commands reached
  through parents, help/version/completion behavior, parse/validation errors before pre-run, and
  a future command with `PreRunE` rather than `PersistentPreRunE`.
- Search for writes that bypass `checkedWriteLock` and the new adapter authorizers: mkdir-before-lock,
  temp/editor files, init/linkback helpers, Thread plan output, config preferences/migrations,
  space registration, fixes, no-op mutations, and stores opened indirectly by workspaces/TUI.
- Test whether dry-run early returns, unchanged/no-op returns, or validation-before-authorization can
  let a read-only command execute semantically mutating callbacks or observable side effects.
- Challenge the optional-callback default. Determine whether fail-open is correct for adapter reuse
  or creates an accidental unguarded CLI construction path that should be structurally impossible.
- Challenge safety taxonomy accuracy, especially mixed-mode commands (`lint`, `task ac`, config
  editors, Thread compose) and commands that write outside the planning tree.
- Stress machine compatibility: revision classification, nil-vs-empty arrays, exact schema enum,
  command-path stability, sorting/duplicates, hidden/deprecated booleans, and whether initializing
  generated completion commands changes help/docs behavior unexpectedly.
- Look for overreach: core/Cobra coupling, a security boundary implied by naming, duplicated guard
  calls with surprising effects, or policy that prevents future TUI/web adapters from composing
  stores cleanly.

## Validation and restoration

All probes, generation, and mutations must stay inside the mandatory isolated sandbox. Restore each
mutation to the sandbox baseline before the final report. The final sandbox diff must contain only
the assigned audit file. Do not commit to, stage, clean, reset, or push from the source checkout.
Use `/tmp`-scoped Go and golangci caches if host cache permissions prevent validation.

## Deliverable

Update only the assigned audit. Findings must be evidence-backed and use the repository's exact
finding grammar with open status. Rank systemic bypasses, incomplete mutation coverage, incorrect
wire compatibility, or adapter-boundary regressions above cosmetic issues. If no findings survive,
record a concise settled verdict plus the inventories, hostile mutations, and command evidence that
earned it; a bare “looks good” is not sufficient.

## Reviewer report

**Verdict: not ready as claimed; no shipped command is misclassified or able to mutate.**

The production enforcement works for every shipped command. Every runnable leaf carries exactly
one recognized tag. Every persistent pre-run binds the selected leaf before discovery. All 49
read-only commands, with and without `--json`, ran against a fixture copy with zero safety errors
and no bytes changed in the planning tree or the home config/registry. A read-only probe that calls
the lifecycle, preference, migration, and space add/forget use cases is rejected in both dry-run
and write modes, again with no durable bytes changed. Legitimate mutating commands still write. The
adapters remain Cobra-free, `schema --json` is deterministic and validates against its own
exact-revision JSON Schema, and the race suite is green.

It is not ready *as claimed* for two Medium reasons, both about the claim that the tag is now
load-bearing:

- **M1:** the CLI composition root still builds a writable planning store with no authorizer,
  `App.WorkspaceSvc`. A read-only command wrote through it in a sandbox probe. Only the untested
  `ui` launch check keeps that invariant true today.
- **M2:** the evidence offered for acceptance criterion 3 exercises only the `SetFields` path,
  which is guarded twice. The `checkedWriteLock` check is the only guard on six guarded mutation
  families (task lifecycle, dependency add/remove/repair, and every Thread write). It can be
  deleted, as can 14 of the 17 explicit store entry checks and the config/space wiring, and the
  whole suite stays green.

Three Low findings follow (L1–L3).

### Isolation attestation

Helper output from `scripts/isolated-review-workspace.sh verify --sandbox <sandbox>`, taken after
this report was finalized and every probe was restored:

| item | value |
| --- | --- |
| Sandbox path | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8OjFPH` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.8OjFPH/.git` (real in-tree directory; no alternates; single worktree; no `core.worktree`) |
| Baseline commit | `3dccd58e4a6e03efeb8338fb739ec0ffc72c9d3e` `chore: capture isolated review baseline`, on `3d4905c` (= `origin/main`) |
| Captured source blob (deliverable) | `9fa9a6b6d5a163cf06c82f00b374245107ef9126` |
| Captured source fingerprint | `44b9eae1f8dda6e18d45f2b2bb7576ea663a0435` |
| Deliverable | `planning/audits/6gcdj3xgt95z-2026-09-21-command-safety-annotations-implementation-claude.md` |
| Transfer | one-shot guarded `transfer`, run after this text was sealed. The helper refuses a second transfer once the source blob changes, so the transfer result cannot be recorded here without breaking its own origin-hash guard. The transfer attestation is in the reviewer's handoff reply. |

The workspace was created with `scripts/isolated-review-workspace.sh create --source <handoff>
--deliverable <audit> --print-path` before any implementation file was read. Every build, test,
generator, fixture, mutation, and edit ran in the sandbox. The shared checkout was used only to
read this brief and for the helper's initial copy. No commit, branch switch, stage, restore, clean,
stash, reset, or write-capable project command was run there. The helper's baseline commit is the
only commit. Temporary probe tests (`internal/cli/zz_review_probe_test.go`,
`internal/store/zz_review_probe_test.go`, and a JSON Schema probe) were removed. Every mutation was
restored from the baseline and checked with `git status --short`, which was empty apart from this
audit. Go and golangci caches used a session scratch directory.

### Reviewed state and validation results

| item | value |
| --- | --- |
| Branch | `feat/load-bearing-command-safety`, uncommitted handoff diff captured in the baseline |
| Base | `3d4905c` (merge of #251) |
| Delta | 89 files, +1360/−154 (includes both review audits, 39 golden files, and 5 new generated docs) |
| Go / lint | `go1.26.6 darwin/arm64` · `golangci-lint 2.12.2` built with go1.26.2 |

| check | command | result |
| --- | --- | --- |
| full race suite, uncached | `go test -race -count=1 ./...` | **PASS**: exit 0, 25 `ok` packages, 0 `FAIL` |
| targeted safety tests | `go test -count=1 ./internal/cli ./internal/store ./internal/configstore ./internal/spacestore -run 'CommandSafety\|ReadOnlyCommand\|Authorization'` | **PASS** at baseline. This is also the "focused" column in the mutation ledger. |
| static analysis | `golangci-lint run ./...` | **0 issues** |
| formatting / tidy | `gofmt -l cmd internal` · `go mod tidy -diff` | clean · clean |
| generated CLI docs | `go run ./internal/tools/docgen -out <tmp>` then `diff -ruN docs/cli <tmp>` | **no drift** |
| generated schema comments | `go run ./internal/tools/schemacomments -out <tmp>` then `diff -u internal/wire/schema_comments.json <tmp>` | **no drift** (268 comments) |
| planning lint | `tskflwctl -C <sandbox> lint` | `✔ all planning entities and dependency links pass lint` |
| whitespace | `git diff --check HEAD~1 HEAD` | 5 × `new blank line at EOF`, all in the new `docs/cli/tskflwctl_completion*.md`; clean outside `docs/cli`. This is generator convention: `git diff --check` against the empty tree flags all 96 of `main`'s `docs/cli` pages the same way, and CI runs `just docs-check`, not `diff --check`. The closeout's "`git diff --check` is clean" holds only if the new docs were untracked when it ran; that was not inspected in the shared checkout, per protocol. |
| JSON Schema, real output | sandbox test: compile `schema --json-schema` `$defs/SchemaEnvelope` and validate real `schema --json` | **PASS** on 100 rows (9 `hidden: true`, 2 `deprecated: true`). Negative controls fail as they should: `"safety":"readonly"` is rejected by the enum, and removing one `hidden` is rejected by `required`. |

### Consumer inventory (built from sandbox source)

**Runnable surface.** `schema --json` publishes 100 rows: 49 `read-only` and 51 `mutating`, which
matches the closeout. Paths are byte-sorted and unique, and three runs across two cwds and two
config homes were byte-identical. Hidden rows are `complete`, `doctor`, `edit`, `list`, `new`,
`show`, and `start` (bare-verb redirects plus the doctor forwarder), and `task promote|demote`,
which are also `deprecated`. The surface was cross-checked against the Cobra tree through the
generated docs:

- every visible runnable path has a `docs/cli` page except `help`, which Cobra's docgen excludes
  by design;
- every doc page missing from the schema is a group node (`tskflwctl`, `task`, `audit`, `epic`,
  `research`, `thread`, `space`, `template`, `theme`, `task depend`, `completion`).

There are no `Aliases:` in `internal/cli`, and all 19 non-root `AddCommand` sites run at
construction, so no leaf is registered dynamically. Cobra's `__complete` is created by
`initCompleteCmd` only while it serves a completion request; `__completeNoDesc` is an alias of
that same command. Leaving that runtime row unpublished is correct, because the transport is not a
stable contract path. `bind` classifies it read-only at runtime (`command_safety.go:36-43`), and
`TestCommandSafetySurfaceCoversEveryRunnableCommand` proves it.

Eagerly initializing help and completion changed no user-visible behavior. These outputs are
byte-identical between a `main` binary (built from `git archive HEAD~1`) and the branch binary:
`--help`, `completion --help`, `help completion`, `completion zsh`, `__complete task s`,
`__complete ''`, and `__completeNoDesc -C <fixture> task show ''`. The only effect is five new
`docs/cli` completion pages and one index line.

**Pre-run routes.** No command uses `PreRun`, `PreRunE`, non-E `PersistentPreRun`, or
`EnableTraverseRunHooks`, and Cobra runs only the nearest persistent hook. Every route binds first:

- `repoPreRun` (`root.go:335`): every command without its own hook, including `help`, the
  `completion` subtree, and `__complete`;
- `styleOnlyPreRun` (`root.go:147`): `init`, `version`, `schema`, the `space` group, and the six
  bare verbs;
- the custom hooks in `doctor.go:44`, `template.go:25`, and `theme.go:35`;
- `status.go:34`: the `--all` branch binds at `:38`, and every other branch delegates to
  `repoPreRun`;
- `ui.go:36`: the completion, space, and `-C` branches delegate to `repoPreRun`, and the ambient
  branch binds at `:40`.

`--help`, `--version`, parse errors, argument validation, and non-runnable groups all return before
any hook or `RunE`, so they need no binding. A future leaf with its own `PreRunE` still runs the
binding ancestor hook first.

**Mutation routes.**

- `core.Service` → `store.FS`:
  - 17 explicit entry checks: `MoveAudit`, `AppendAuditBody`, `EditBody`, `TransformAuditBody`,
    `TransformTaskBody`, `createEntityFile` (covers `CreateTask`, `CreateAudit`, `CreateResearch`,
    and `CreateEpic`), `MoveEpic`, `SetEpicFields`, `EditEpic`, `EditTask`, `EditAudit`,
    `FixFrontmatter`, `SetFields`, `RenameTask`, `SetResearchFields`, `EditResearch`, and
    `AppendResearchBody`.
  - The check in `checkedWriteLock` (`lock.go:115-118`), which is the **only** guard for
    `MutateTaskGraph` (`graphmutation.go:38`), `MutateTaskGraphRepair` (`graphrepair.go:33`),
    `MutateTaskLifecycle` (`lifecyclemutation.go:32`), `MutateThreadApply` (`threadapply.go:30`),
    `MutateThread` (`threadmutation.go:31`), and `MutateThreadCreation` (`threadcreation.go:31`).
    All six take the lock before dry-run planning.
- `ConfigurationService` → `configstore`: `MigrateConfiguration` and `SetPreference`.
- `SpaceRegistryService` → `spacestore`: `AddSpace` and `ForgetSpace`; `RegisterInitialized` also
  routes through `add`.
- Direct CLI writes:
  - `init` (`config.Init`, `InitPointer`, `AddTrackedRepo`, `LinkBack`) is authorized first
    (`init.go:54`).
  - `thread compose --out` is authorized on its real path only (`thread_apply.go:69`).
  - `$EDITOR` temp files are created inside store edit callbacks, after the entry check.
- TUI: launched by `ui` (`ui.go:58`).

Every other filesystem write in `internal/` either belongs to these routes or is a dev tool under
`internal/tools/*`.

**Adapter constructors** (production only):

| constructor | site | authorizer | soundness |
| --- | --- | --- | --- |
| `store.NewFS` | `root.go:445-451` (discovered repo) | yes | enforced |
| `store.NewFS` | `workspacestore/fs.go:23`, via `App.WorkspaceSvc` (`root.go:268`) | **no** | returns a full writable `core.Service` (`core/workspace.go:108`); sound only while `ui` is its sole consumer (**M1**) |
| `store.NewFS` | `spacestore/fs.go:112` `OpenPlanningStore` | no | sound: exposed only as the read-only `core.PlanningSummarySource` port |
| `store.NewFS` | `completion.go:222,259` | no | sound: `ListTasks`/`ListAudits` inside the read-only completion transport |
| `configstore.New` | `root.go:264` | yes | enforced |
| `spacestore.New` | `root.go:261` | yes | enforced |

`internal/tui` constructs no adapter. `store`, `configstore`, `spacestore`, and `core` import no
Cobra; the only `cobra` mentions under `internal/core` are comments.

**Machine-contract consumers.**

- The contract side is `wire.SchemaCommand` and `SchemaContract.Commands`, with the JSON Schema
  generated by reflection: all four fields are `required`, the `safety` enum is exact, and
  `additionalProperties` is `false`.
- `schema_comments.json` was regenerated with no drift. `NormalizeSchemaContract` turns a nil slice
  into `[]`.
- 1.71 is declared `ADDITIVE` and `SchemaRevisionCompatibility` is `additive`, which the changelog
  test enforces.
- Human schema output: `Command safety: 100 runnable (49 read-only, 51 mutating)`.
- The revision-gated goldens: `schema_json` pins every row, and the other machine goldens carry the
  1.71 bump.
- Prose: `docs/cli/tskflwctl_schema.md`, README, CLAUDE.md, ARCHITECTURE, and the research spec
  (see L3).

### Hostile-evidence ledger (baseline code, sandbox probes)

| probe | result |
| --- | --- |
| Read-only probe command → `app.Svc.Move(alpha → in-progress)`, dry-run and real | rejected with `command safety violation`; repo digest unchanged |
| Read-only probe → `ConfigSvc.SetPreference(user theme)`, `ConfigSvc.Migrate`, `SpaceSvc.Add`, `SpaceSvc.Forget`, each dry-run and real, with an isolated `TSKFLW_CONFIG_HOME` | all 8 rejected; digest of the repo **and the home config/registry** unchanged |
| Read-only probe → `app.WorkspaceSvc.Open(repo).Planning.SetFields(alpha, priority=high)` | **`err=nil`, task file rewritten** with `priority: high` and `updated_at` (M1) |
| Mutating commands: `task set --priority high`, `task start`, `space add --id probe`, `space forget probe`, `lint --fix`, `lint --fix --dry-run` | all accepted; task and registry bytes written; `lint` exits on unfixable validation findings, never on safety |
| All 49 read-only commands, `--json` and human, against a copy of `internal/cli/testdata/planning` with a scratch home | 0 safety errors; fixture and home digests unchanged |
| Outside any repo: `help task`, `completion bash`, `__complete task s`, `__completeNoDesc task s`, `ui`, `ui --dry-run`, `doctor`, `status --all --json`, `space list --json`, `theme list` | all behave as on `main`; `ui` exits 11 with `needs an interactive terminal` / `no --dry-run preview`; scratch home untouched |
| `store.NewFS(<missing root>, deny).MutateThreadCreation`, dry-run and real | returns the denial, but **`<missing root>` now exists** (L2) |

**Concurrency and lifetime.** `commandSafetyState` is written once, in the persistent pre-run on
the Cobra goroutine, before `tui.Run` starts. The only other write is `bind` stamping the annotation
on the per-process `__complete` command. Every later access is a read through the `app` pointer
that the adapter callbacks capture, including from TUI `tea.Cmd` goroutines. So there is no
write-after-start and no data race, and the full race suite agrees. Two cases are unsupported, not
demonstrated:

- concurrent `Execute` on one shared root;
- re-executing an `App` whose new hook skips `bind`, which would reuse the previous binding. No
  shipped hook skips it, and production builds one root per process.

### Restored-mutation ledger

Each mutation was applied alone (or as the named coordinated pair), run against three test sets,
then restored from the baseline:

- **focused:** `go test -count=1 ./internal/{cli,store,configstore,spacestore} -run 'CommandSafety|ReadOnlyCommand|Authorization'`;
- **full suite:** `go test ./... -skip ReviewProbe`;
- **reviewer probe:** the sandbox probes above.

"killed" names the first failing test.

| # | mutation | focused | full suite | reviewer probe |
| --- | --- | --- | --- | --- |
| M01 | delete the check in `checkedWriteLock` | survives | **survives (25 ok)** | killed: lifecycle probe |
| M02 | delete `SetFields` entry check | killed: `TestReadOnlyCommandCannotReachMutatingServicePath` and the store `TestFSMutationsRequireAuthorizationBeforeDryRun` | killed | — |
| M03 | delete `SetEpicFields` entry check | survives | **survives** | survives |
| M04 | delete `RenameTask` entry check | survives | **survives** | survives |
| M05 | delete `TransformTaskBody` entry check (`task ac`) | survives | **survives** | survives |
| M06 | delete `createEntityFile` entry check | killed: store test | killed | — |
| M07 | unwire configstore authorizer in `root.go` | survives | **survives** | killed: home probe |
| M08 | unwire spacestore authorizer in `root.go` | survives | **survives** | killed: home probe |
| M09 | unwire store authorizer in `resolveFrom` | killed | killed | killed |
| M10 | delete configstore `SetPreference` check | killed: configstore test | killed | killed |
| M11 | delete spacestore `AddSpace` check | killed: spacestore test | killed | killed |
| M12 | remove `task set` tag | killed: surface test | killed (plus 8 behaviour tests) | killed |
| M13 | falsify `task list` → `mutating` | survives | killed only by the revision-gated `TestGolden_MachineContract` | — |
| M14–M17 | drop the bind in the doctor, status `--all`, template, and theme hooks | survive | **survive** | survive; inert, because an unbound state denies mutation |
| M18 | drop the bind in the ambient `ui` hook | survives | **survives** | survives. A binary built with this mutation fails `tskflwctl ui` outside a repo with `command safety violation: mutation reached persistence before a command safety classification was bound`; baseline gives `needs an interactive terminal` (L1) |
| M19 | drop the bind in `styleOnlyPreRun` | survives | killed (init, space, audit, and task tests) | killed |
| M20 | drop the bind in `repoPreRun` | killed: surface test | killed | killed |
| M21 | drop the `init` authorize | survives | survives (inert, because `init` is mutating) | — |
| M22 | drop the `init` authorize **and** falsify the init tag to read-only | survives | killed only by registration-coupled init tests and the golden. The scaffold write itself is unguarded, so `--no-register` would pass (L1) | — |
| M23 | falsify the init tag, keep the authorize | survives | killed | — |
| M24 | drop the compose authorize **and** falsify the compose tag | survives | killed only by the golden (L1) | — |
| M25 | drop the `ui` `RunE` authorize | survives | **survives** | survives (L1, M1) |
| M26 | drop the `__complete` special case in `bind` | killed: surface test | killed (plus completion tests) | — |
| M27 | make unbound `authorizeMutation` return nil | survives | **survives** | survives (L1) |

A tag falsified from read-only to mutating (M13) is caught only by a golden. That is acceptable
here, because updating that golden requires a machine-contract revision advance
(`TestMachineGoldenUpdateRequiresRevisionAdvance`), so it cannot be regenerated silently.

### Acceptance-criteria and planning truthfulness

- AC1 (schema emits the surface, golden pins it): **met**. It is pinned by `schema_json.golden`,
  the real output validates, and the revision is additive.
- AC2 (every registered command, including hidden and deprecated, carries a recognized value):
  **met** for the stable tree and the runtime `__complete`.
- AC3 (a read-only command reaching a mutating path fails a test): **met only for `SetFields`,
  `createEntityFile`, `FixFrontmatter`, and the two config and two registry adapter methods**. The
  lock choke point, 14 explicit store checks, and the composition-root wiring for config and space
  are not pinned (M2).
- AC4 (`--dry-run` derivation decided and recorded): **met** in ARCHITECTURE and the closeout.
- AC5 (agent guidance): **met** in CLAUDE.md, README, the `schema` long help, and the human schema
  output.
- ARCHITECTURE's statement that the filesystem, config, and space adapters "receive a
  framework-neutral authorization callback" is true for the discovered-repo store but not for
  `workspacestore`, which the same composition root builds (M1).
- The research spec still describes an invocation-level taxonomy that contradicts the shipped
  capability rule (L3).

### Settled concerns (challenged, then closed with evidence)

- **Hook replacement.** Every custom persistent hook binds first. Cobra runs one persistent hook,
  and `PreRunE` does not displace it. Unbound state denies mutation, although that deny branch is
  untested (L1).
- **Validation before authorization and dry-run early returns.** Every explicit entry check sits
  before any planning, `prepare()`, or dry-run branch. The six guarded planners lock before
  planning in dry-run too. Only argument checks such as a nil planner or zero time come earlier,
  and those have no side effects. The one exception is L2.
- **No-op mutations.** Idempotent `MoveAudit`, unchanged `EditTask`, and unchanged Thread writes
  (`!materialized.changed`) are all checked first, so a read-only no-op is still rejected.
- **Mixed-mode taxonomy.** `lint` (`--fix`), `task ac` (`--check`), `config edit`,
  `thread compose`, and `init` are `mutating`. These read-only commands have no write flag or
  route: `config` (an alias of `show`), `config doctor`, `doctor`, `audit lint`, `thread plan`,
  `workspace`, `status [--all]`, `template show`, and `theme preview`. None of the 49 read-only
  commands errors on safety.
- **Adapter neutrality and overreach.** The callback is a plain `func() error`, and core and the
  adapters import no Cobra. The TUI reaches mutations only through injected core services. A future
  web adapter can pass its own policy or omit the option. The duplicate calls (explicit check plus
  the lock check) have no side effects.
- **Wire compatibility.** The field is additive and non-nil (`[]`). `hidden` and `deprecated` are
  always emitted and required. The enum is exact, and ordering is byte-stable. The JSON Schema is
  exact-revision, and the real output, including non-default rows, validates.

### Residual risks

- The fail-open adapter defaults are intentional for reuse (ARCHITECTURE). Beyond M1, protection
  rests on port typing (`PlanningSummarySource`) and on call-site discipline (`completion.go`),
  with no structural test.
- `thread compose --dry-run` is not authorized, while every adapter dry-run is. This is harmless
  today because the command is tagged mutating (L1).
- Reusing an `App` across `Execute` calls is unsupported. A future hook that skipped `bind` would
  inherit the previous invocation's capability there. It is not reachable in production.

## Findings

#### M1. The CLI composition root still builds a writable, unauthorized planning store behind App.WorkspaceSvc · **Status:** fixed

**File:** internal/workspacestore/fs.go:23 | **Component:** cli/workspacestore
**Effort:** S · **Urgency:** soon

The contract says a read-only command cannot enter a mutation use case through the planning
filesystem, and `docs/ARCHITECTURE.md:411-415` says the filesystem secondary adapter receives
the authorization callback. Yet the same composition root that wires `store`, `configstore`,
and `spacestore` also builds `App.WorkspaceSvc = core.NewWorkspaceService(workspacestore.New())`
(`internal/cli/root.go:268`). `workspacestore.OpenWorkspace` calls `store.NewFS(cfg.Root)` with
no `WithMutationAuthorization` (`internal/workspacestore/fs.go:23`). `core.WorkspaceService.Open`
then wraps that store in a full, writable `core.Service` (`internal/core/workspace.go:108`).
Because the option defaults to allow when omitted, every write through a workspace-opened service
skips command safety entirely.

Hostile evidence, using the shape of the implementation's own
`TestReadOnlyCommandCannotReachMutatingServicePath`: a hidden probe command annotated `read-only`,
added through `newRootCmd`, run with `-C <repo>`:

```go
ws, _ := app.WorkspaceSvc.Open(core.WorkspaceRequest{Start: repo})
_, err = ws.Planning.SetFields("alpha", map[string]any{"priority": "high"}, false, false)
// observed: err=<nil>; alpha's file now contains `priority: high` and a new updated_at
```

The same probe routed through `app.Svc` is rejected. No shipped read-only command uses
`WorkspaceSvc` today. Its only consumers are `ui` (`ui.go:73`, `ui.go:113`), where the TUI opens
additional spaces after launch. The invariant therefore holds only by convention, backed by one
call to `app.authorizeMutation()` in `ui`'s `RunE` (`ui.go:58`). Deleting that call (mutation M25)
leaves the full suite green. `WorkspaceSvc` is a general `App` field ("opens arbitrary local
planning entry points"), so a future read-only command that reuses it gets a silently writable
store. That is the "accidental unguarded CLI construction path" the brief asks to make
structurally impossible. The TUI's secondary workspaces also depend entirely on the launch check:
their store has no guard of its own.

**Recommendation:** Give workspacestore a WithMutationAuthorization option threaded into store.NewFS and wire app.authorizeMutation at root.go:268 (ui stays mutating, so the TUI keeps working); add a read-only WorkspaceSvc probe to the safety test

**Resolution:** Propagated the invocation authorizer through workspacestore into
every opened planning store; a CLI-level read-only workspace probe now proves
both preview and write are rejected.

#### M2. Enforcement is proven through one doubly guarded path; the lock choke point, 14 of 17 store entry checks, and the config/space wiring can be deleted with the suite green · **Status:** fixed

**File:** internal/store/lock.go:115 | **Component:** store/cli-tests
**Effort:** M · **Urgency:** soon

Acceptance criterion 3 ("a read-only command that reaches a mutating service path fails a test")
rests on one CLI-level test, `TestReadOnlyCommandCannotReachMutatingServicePath`
(`internal/cli/command_safety_test.go:67`). It only calls `app.Svc.SetFields`, which has both an
explicit entry check and the lock check. A store test, `TestFSMutationsRequireAuthorizationBeforeDryRun`,
covers only `CreateTask`, `SetFields`, and `FixFrontmatter`. Only four tests in the repository use
a denying authorizer, found with
`grep -rn 'WithMutationAuthorization\|newRootCmd(' --include='*_test.go' internal`.

Restored sandbox mutations, each run against the full suite (`go test ./...`):

- **M01:** deleting the `s.authorizeMutation()` call from `checkedWriteLock` (`internal/store/lock.go:115-118`)
  leaves all 25 packages green. That call is the *only* guard for `MutateTaskGraph`,
  `MutateTaskGraphRepair`, `MutateTaskLifecycle`, `MutateThreadApply`, `MutateThread`, and
  `MutateThreadCreation`, in both dry-run and real modes. Those carry every task lifecycle verb,
  `task depend add|remove|repair|migrate`, `thread new|add|remove|start|complete|cancel|reopen`,
  and `thread apply`. A reviewer probe (read-only command → `app.Svc.Move`) kills it.
- **M03, M04, M05:** deleting the entry check from `SetEpicFields`, `RenameTask`, or
  `TransformTaskBody` (`task ac`) leaves the suite green. Their dry-run branches return before any
  lock (`epicstore.go:175`, `rename.go:48`, `writeBody`), so the explicit check is the only dry-run
  guard. The other 11 unpinned entry checks share this shape: `MoveAudit`, `AppendAuditBody`,
  `EditBody`, `TransformAuditBody`, `MoveEpic`, `EditEpic`, `EditTask`, `EditAudit`,
  `SetResearchFields`, `EditResearch`, and `AppendResearchBody`.
- **M07, M08:** replacing
  `configstore.New(configstore.WithMutationAuthorization(app.authorizeMutation))` or the
  `spacestore` equivalent with a bare `New()` (`root.go:261,264`) leaves the suite green. The
  adapter-level tests pass because they build their own denying store; nothing checks that the CLI
  wires it. A reviewer probe (read-only command → `ConfigSvc.SetPreference` / `Migrate`,
  `SpaceSvc.Add` / `Forget`, dry-run and real, isolated `TSKFLW_CONFIG_HOME`) kills both.

So the single probe masks broad classes: the one choke point behind six mutation families,
fourteen entry points, and two composition-root wirings can each regress with CI green. The
code is correct today. This finding is about the claim that the tag "gates or verifies
something", which is what the task set out to make true ("a test failure, not a code-review
catch").

**Recommendation:** Add a table-driven deny-authorizer test over every store mutation entry and all six guarded planners (dry-run and real, asserting no bytes change) plus a CLI-level read-only probe through ConfigSvc, SpaceSvc and a lifecycle use case

**Resolution:** Added exhaustive deny-authorizer coverage for all 26 filesystem
mutation entries in preview and write modes, plus CLI probes for planning,
lifecycle, configuration, registry, and opened-workspace boundaries.

#### L1. Custom pre-run binds, the unbound-deny branch, and the init/compose/ui direct authorizations have no killing tests; the ambient ui launch can break silently · **Status:** fixed

**File:** internal/cli/ui.go:40 | **Component:** cli
**Effort:** S · **Urgency:** eventually

Several enforcement sites that the brief names explicitly have no test that kills their removal.
All were restored after probing.

- **M18:** removing `app.bindCommandSafety(cmd)` from `ui`'s ambient pre-run branch
  (`internal/cli/ui.go:40`) leaves the suite green, yet it breaks the default launch. With a
  binary built from that mutation, `tskflwctl ui` outside a repo exits 1 with
  `command safety violation: mutation reached persistence before a command safety classification was bound`.
  Baseline gives `ui needs an interactive terminal`, exit 11, so in a real terminal the TUI would
  refuse to start. Every `ui` test in `internal/cli/ui_test.go` passes `-C`, which routes through
  `repoPreRun`. The `uiStartup` tests call the method directly and skip Cobra.
- **M25:** removing `ui`'s `RunE` authorization (`ui.go:58`) survives. This is the only guard for
  TUI-opened workspaces (M1).
- **M27:** making unbound `commandSafetyState.authorizeMutation` return nil
  (`internal/cli/command_safety.go:59`) survives. The claim that "a hook that forgets to bind fails
  closed" is therefore untested. That is the property that makes M14–M17 (dropping binds in the
  doctor, status `--all`, template, and theme hooks) harmless today.
- **M22, M24 (coordinated):** dropping `init`'s authorization (`init.go:54`) and also falsifying
  its tag to read-only is caught only incidentally, by the space-registration tests and the golden.
  The scaffold write itself runs unguarded, and `--no-register` would bypass even those. The same
  pairing for `thread compose` (`thread_apply.go:69`) is caught only by the golden. Separately,
  `compose --dry-run` skips authorization while every adapter dry-run requires it. That is
  harmless today but inconsistent with "the rejection also applies to dry-run mutation paths".

None of these is a shipped defect: the tags are right and every bind is present. But the
fail-closed guarantee around custom hooks and direct writes cannot currently fail a test, and the
most visible consequence is a silently unlaunchable `ui`.

**Recommendation:** Test ambient ui through Cobra (expect the interactive-terminal error, not a safety error), unit-test unbound authorizeMutation denial, and add coordinated read-only probes for init --no-register and compose --out; decide whether compose --dry-run should authorize

**Resolution:** Pinned unbound denial, custom pre-run binding, ambient UI
binding, and coordinated read-only probes for init, compose previews and writes,
and UI launch.

#### L2. Thread creation creates the planning root before authorization, even for a denied dry-run · **Status:** fixed

**File:** internal/store/threadcreation.go:28 | **Component:** store
**Effort:** XS · **Urgency:** eventually

`MutateThreadCreation` runs `os.MkdirAll(s.root, 0o755)` (`internal/store/threadcreation.go:28`)
*before* `checkedWriteLock` (`:31`), which is where this change placed the authorization check.
Every other store mutation authorizes before touching disk, and `createEntityFile` authorizes
before its own `MkdirAll`. Sandbox probe:

```go
store := NewFS(filepath.Join(t.TempDir(), "not-yet"), WithMutationAuthorization(deny))
_, err := store.MutateThreadCreation(time.Now(), dryRun, planner)
// dryRun=true  → err=mutation denied, planner not called, root CREATED
// dryRun=false → err=mutation denied, planner not called, root CREATED
```

This cannot be reached through the CLI today. `config.Discover` requires `<root>/tasks/` to exist
(`configuredRoot`), so the `MkdirAll` is a no-op for a discovered root. It is a durable side effect
of a *denied* dry-run at the reusable adapter boundary, which the brief lists explicitly
(mkdir-before-lock). The unconditional `MkdirAll` on dry-run was pre-existing. What this change
added is placing the check after it.

**Recommendation:** Authorize before the MkdirAll (an explicit entry check like the other mutations) and skip the MkdirAll on dry-run

**Resolution:** Thread creation now authorizes before any filesystem effect and
creates the planning root only for a real authorized write.

#### L3. The research spec still classifies lint per invocation, contradicting the shipped capability rule · **Status:** fixed

**File:** planning/research/6f9menr01t1n-tskflwctl-command-spec.md:228 | **Component:** planning/docs
**Effort:** XS · **Urgency:** eventually

This change edited the command-safety bullet of
`planning/research/6f9menr01t1n-tskflwctl-command-spec.md` to say that adapters "reject mutation
from a read-only command". The sub-bullets directly beneath it still classify per invocation:
"*Read-only:* … `lint` *without* `--fix`" (line 228) and "*Mutating:* … `lint --fix`" (line 230).
The shipped rule, stated in ARCHITECTURE, the task closeout, and `lint.go`'s own comment, is that
safety is a *command capability*, so `lint` is `mutating` in `schema --json` whatever flags an
invocation uses. An agent reading the spec will expect a `read-only` row for `lint`.

**Recommendation:** List lint (and task ac) under Mutating as command capabilities and note that read-only invocations of a mutating command remain legal

**Resolution:** The command spec now classifies lint and task ac by
whole-command capability and explains that read-only invocations of mutating
commands remain legal.
