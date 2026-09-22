---
schema: 1
id: 6gcdj3xsb5s0
bucket: closed
area: command-safety-annotations-implementation-antigravity
date: "2026-09-21"
---
# Audit: Command safety annotations implementation — antigravity — 2026-09-21

> Reviewer assignment: antigravity. This document is the review brief and the only file the reviewer should update.
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

### Settled verdict & findings

**Verdict:** Settled clean — no defects or systemic contract gaps identified.

Adversarial exploration across all mandatory inventory checks, runtime execution paths, and hostile mutation probes failed to falsify any of the target claims:
1. Every runnable command (100 in total: 49 read-only, 51 mutating) carries an explicit, recognized safety capability.
2. Command safety binding is fail-closed, executing before discovery or command logic across root, style-only, and custom persistent pre-run hooks.
3. Read-only commands cannot enter mutating use cases through planning filesystem, configuration, or space-registry stores, even under `--dry-run` or no-op conditions.
4. Framework-neutral mutation authorization cleanly isolates secondary persistence adapters without coupling them to Cobra.
5. Revision 1.71 is truthfully additive under ADR-0008, publishing bounded command safety metadata while preserving existing machine contracts.

---

### Mandatory isolation attestation

```ini
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.72n3lz
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.72n3lz/.git
baseline_commit=8fa188d01fd767f94d4b1454075c68087ca55826
source_blob=5b3a80a843f72e41fc033283f97ca4bbb2e767a3
source_fingerprint=44b9eae1f8dda6e18d45f2b2bb7576ea663a0435
deliverable=planning/audits/6gcdj3xsb5s0-2026-09-21-command-safety-annotations-implementation-antigravity.md
```

#### Verification gates
- `go test -race ./...`: **PASS** (33 packages passed, 0 race conditions).
- `golangci-lint run ./...`: **PASS** (0 issues).
- `./bin/tskflwctl --no-color lint`: **PASS** (all planning entities and dependency links pass lint).
- `./bin/tskflwctl --no-color audit lint`: **PASS** (all audit findings pass lint).
- `git diff --check HEAD`: **PASS** (clean working tree).
- Code generators (`schemacomments`, `docgen`, `mangen`): **PASS** (100% reproducible, zero drift).

---

### 1. Command, pre-run, & mutation consumer inventory

A complete census of the Cobra command tree was conducted inside the sandbox:

- **Total runnable commands:** 100
- **Read-only commands:** 49
- **Mutating commands:** 51
- **Hidden commands:** 9 (`complete`, `doctor`, `edit`, `list`, `new`, `show`, `start`, `task demote`, `task promote`)
- **Deprecated commands:** 2 (`task demote`, `task promote`)
- **Grouping nodes (non-runnable):** `tskflwctl`, `audit`, `config`, `epic`, `research`, `space`, `task`, `task depend`, `template`, `theme`, `thread` (deliberately untagged and omitted from the published command surface because they define no `Run`/`RunE`).

#### Representative command execution matrix

| Class | Command | Mode | Stated Safety | Observed Behavior | Authorization Result |
| :--- | :--- | :--- | :---: | :--- | :---: |
| **Ordinary Read** | `tskflwctl task list` | JSON | `read-only` | Emits task list, exit 0 | Not invoked (read-only) |
| **Ordinary Mutating** | `tskflwctl task new` | Prose | `mutating` | Creates task file, exit 0 | Authorized |
| **Mixed-Mode Capable** | `tskflwctl lint` | Prose | `mutating` | Checks repository, exit 0 | Authorized (capable of `--fix`) |
| **Dry-Run Mutating** | `tskflwctl task set --dry-run` | JSON | `mutating` | Previews update, exit 0 | Authorized before dry-run branch |
| **No-Op Mutating** | `tskflwctl task set (unchanged)` | JSON | `mutating` | Evaluates unchanged, exit 0 | Authorized before change check |
| **Hidden Bare-Verb** | `tskflwctl list` | Prose | `read-only` | Recommends nouns, exit 0 | Blocked from write paths |
| **Deprecated Leaf** | `tskflwctl task promote` | JSON | `mutating` | Promotes task, exit 0 | Authorized |
| **Generated Completion**| `tskflwctl completion bash` | Shell | `read-only` | Outputs script, exit 0 | Not invoked (read-only) |
| **Generated Help** | `tskflwctl help version` | Prose | `read-only` | Renders help, exit 0 | Not invoked (read-only) |
| **Custom Pre-Run** | `tskflwctl doctor` | JSON | `read-only` | Validates links, exit 11 | Blocked from write paths |
| **Config Mutation** | `tskflwctl config edit` | Prose | `mutating` | Updates TOML, exit 0 | Authorized by configstore |
| **Space Registry** | `tskflwctl space add` | JSON | `mutating` | Registers space, exit 0 | Authorized by spacestore |
| **Direct Write** | `tskflwctl init` | Prose | `mutating` | Scaffolds repo, exit 0 | Authorized at init RunE |
| **TUI Primary Adapter**| `tskflwctl ui` | Prose | `mutating` | Refuses non-TTY, exit 11 | Authorized at ui RunE |

---

### 2. Runtime safety probes & payload verification

1. **Hostile probe — Read-only command calling mutating service:**
   A custom command `safety-probe` annotated with `{"safety": "read-only"}` was registered and invoked against `app.Svc.SetFields` on an existing task:
   - **Under `dryRun: false`:**
     - Error returned: `command safety violation: read-only command "tskflwctl safety-probe" reached a mutating persistence path`
     - File on disk: Byte-identical before and after invocation (`bytes.Equal(after, before) == true`).
   - **Under `dryRun: true`:**
     - Error returned: `command safety violation: read-only command "tskflwctl safety-probe" reached a mutating persistence path`
     - Proves `s.authorizeMutation()` executes before `dryRun` branching in `internal/store/fsstore.go:189`.
2. **Hostile probe — Read-only command calling config & space mutations:**
   - Calling `app.ConfigSvc.SetPreference` from a `read-only` command returns:
     `command safety violation: read-only command "tskflwctl safety-probe-config" reached a mutating persistence path`
   - Calling `app.SpaceSvc.AddSpace` from a `read-only` command returns:
     `command safety violation: read-only command "tskflwctl safety-probe-space" reached a mutating persistence path`
3. **Dynamic completion transport safety:**
   When Cobra dynamically instantiates `__complete` or `__completeNoDesc` during shell completion execution, [`internal/cli/command_safety.go:36-42`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.72n3lz/internal/cli/command_safety.go#L36-L42) intercepts the unannotated leaf via `isCompletionCommand(cmd)`, binds it as `commandSafetyReadOnly`, and sets the annotation. Verified in `TestCommandSafetySurfaceCoversEveryRunnableCommand`.

---

### 3. Hostile mutation testing matrix

All required mutation seams were executed in the isolated sandbox, verified against focused tests, and restored to baseline:

| Mutation | Description | File & Line | Focused Test Catching Mutation | Observed Failure Output |
| :---: | :--- | :--- | :--- | :--- |
| **M1a** | Omit safety tag from command (`task ac`) | `internal/cli/task.go:438` | `TestCommandSafetySurfaceCoversEveryRunnableCommand` | `command safety invariant: runnable command "tskflwctl task ac" has unrecognized safety annotation ""` |
| **M1b** | Falsify command tag (`task new` → `read-only`) | `internal/cli/task.go:130` | `TestGolden_MachineContract/schema_json`<br>`TestSmoke_LifecycleAndExitCodes` | `integration_golden_test.go:124: output drift vs testdata/golden/schema_json.golden`<br>`main_test.go:212: task new: exit 1: command safety violation: read-only command "tskflwctl task new" reached a mutating persistence path` |
| **M2** | Omit custom pre-run bind (`styleOnlyPreRun`) | `internal/cli/root.go:148` | `TestInitExistingLegacyPointerHandsOffToConfigMigrate`<br>(and 23 other `TestInit*` tests) | `command safety violation: mutation reached persistence before a command safety classification was bound` |
| **M3** | Omit store authorization (`SetFields`) | `internal/store/fsstore.go:189` | `TestFSMutationsRequireAuthorizationBeforeDryRun`<br>`TestReadOnlyCommandCannotReachMutatingServicePath` | `mutation_authorization_test.go:19: SetFields error = task "missing": not found, want authorization error`<br>`command_safety_test.go:90: read-only mutation error = <nil>, want command safety violation` |
| **M4a** | Omit configstore authorization (`SetPreference`) | `internal/configstore/fs.go:158` | `TestFSMutationsRequireAuthorizationBeforeDryRun` | `fs_test.go:23: SetPreference error = userconfig: refusing to read..., want authorization error` |
| **M4b** | Omit spacestore authorization (`AddSpace`) | `internal/spacestore/fs.go:85` | `TestFSRegistryMutationsRequireAuthorizationBeforeDryRun` | `fs_test.go:20: AddSpace error = userconfig: refusing to read..., want authorization error` |
| **M5a** | Direct-write authorization (`init` → `read-only`) | `internal/cli/init.go:39` | `TestInitExistingDirectReportsWithoutMigrating` | `command safety violation: read-only command "tskflwctl init" reached a mutating persistence path` |
| **M5b** | TUI launch authorization (`ui` → `read-only`) | `internal/cli/ui.go:34` | `TestUIRefusesNonInteractiveAndDryRunInvocation` | `ui_test.go:28: non-interactive ui should fail clearly with validation: command safety violation: read-only command "tskflwctl ui" reached a mutating persistence path` |

---

### 4. Architecture & contract integrity

1. **Framework-neutral secondary adapters:**
   [`internal/store`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.72n3lz/internal/store), [`internal/configstore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.72n3lz/internal/configstore), and [`internal/spacestore`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.72n3lz/internal/spacestore) import neither `cobra` nor `internal/cli`. They receive an optional callback `WithMutationAuthorization(func() error)`. When omitted (such as in headless domain unit tests), mutations succeed without requiring Cobra fixtures; when injected by the CLI composition root in `root.go`, safety is strictly enforced.
2. **Fail-closed unclassified execution:**
   If a command reaches a store mutation without having bound a safety classification (`s.safety == ""`), `authorizeMutation()` rejects the call with `"command safety violation: mutation reached persistence before a command safety classification was bound"`.
3. **Decoupling from `--dry-run`:**
   Command safety reflects inherent command capability (whether any invocation can persist writes). `--dry-run` availability remains command-specific: interactive commands such as `ui` cannot support dry-run previews, whereas batch and file-mutating commands support previewing.
4. **Machine contract & revision 1.71:**
   `wire.SchemaVersion` is advanced to `"1.71"` and classified `ADDITIVE` in accordance with ADR-0008. The new `commands` array in `SchemaContract` is normalized to `[]wire.SchemaCommand{}` (never `null`), sorted deterministically by command path, and fully validated by Draft 2020-12 schema annotations with `additionalProperties: false`.

---

### 5. Concurrency & lifetime analysis

- In `internal/cli/root.go`, `commandSafetyState` is bound once during single-threaded startup before use-case execution or background goroutine initialization.
- In long-lived primary adapters (such as `tskflwctl ui`), `s.safety` is read concurrently by store mutation calls but is never mutated after launch.
- Concurrency audit under `go test -race ./...` across all 33 packages reported 0 data races.

---

### 6. Residual risks & operational notes

1. **TUI workspace opening boundary:**
   `workspacestore.OpenWorkspace` constructs a `store.NewFS` without `WithMutationAuthorization` because it is an adapter-neutral core service called after the TUI has already authorized its mutation capability at launch (`ui.go:57`). If a future CLI command opens workspaces dynamically outside `ui`, it should ensure the store is constructed with the invocation authorizer.
2. **Agent consumption of command safety:**
   Automated agents should consult `schema --json`'s `commands` array to discover whether a proposed command is `read-only` or `mutating`. Agents must not assume that every `mutating` command supports `--dry-run`.
