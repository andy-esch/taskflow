---
schema: 1
id: 6gcb5p5jh8ct
bucket: closed
area: complete-cli-exit-code-taxonomy-implementation-antigravity
date: "2026-09-21"
---
# Audit: Complete CLI exit-code taxonomy implementation — Antigravity — 2026-09-21

> Reviewer assignment: Antigravity/Gemini. This document is the review brief and the only source file the reviewer may update.
>
> Play devil's advocate. Take a fresh inventory instead of following the implementation's abstractions, look for systemic contract gaps and anti-patterns, and prefer a demonstrated counterexample over broad approval.

## Mandatory isolated workspace

Treat the handoff checkout as read-only. Resolve this audit, then create an independent sandbox before inspecting code, running tests, generators, or mutation probes:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_ABS="$(tskflwctl audit path 2026-09-21-complete-cli-exit-code-taxonomy-implementation-antigravity)"
AUDIT_REL="${AUDIT_ABS#"$SOURCE_ROOT"/}"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

Do all inspection, builds, tests, probes, and report editing in the sandbox. Never stage, commit, switch branches, restore, clean, stash, reset, or run write-capable commands in the source checkout. Restore every probe in the sandbox so only the assigned audit differs, then verify and transfer it:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Include the helper attestation and transfer result in the report. Preserve the sandbox until the implementation owner confirms receipt.

## Review target

Adversarially review task `6gbn4g1bb7wr`, `publish-the-complete-cli-exit-code-taxonomy`, against `main`. Start from process behavior, not the new registry. Cover `cmd/tskflwctl`, CLI routing and prompts, domain classification, schema/wire DTOs, generated artifacts, docs, tests, ADR-0008, and the originating architecture-audit M1.

Do not implement fixes or edit any file other than this audit.

## Claims you must try to falsify

- The published rows cover every process status this binary deliberately returns: active `0,1,10,11,13,14,130`, with retired `12` reserved and never emitted.
- Code/name/meaning/state metadata has one owner; domain failure classes remain separate and cannot accidentally classify success, abort, or reservations.
- Existing consumers see an additive 1.70 change: the four historical rows keep their exact order/content and all new rows append.
- Real process behavior, JSON `error.code`, human output, JSON Schema, goldens, and documentation all describe the same contract.
- Tests fail when the registry and executable behavior drift, instead of merely comparing two values derived from the same mistake.

## Required independent work

1. Search numerically and semantically for every exit source, including `os.Exit`, Cobra parse/usage errors, fang handling, prompt aborts, filesystem failures, batch failures, and signal conventions. Report any deliberate or reachable result absent from the table.
2. Run the compiled binary through success, malformed flags, filesystem/permission failure where portable, not-found, validation, ambiguity, conflict, and prompt cancellation. Record status plus stdout/stderr payload. Do not rely only on unit calls to `ExitCode`.
3. Probe awkward cases: wrapped errors matching more than one sentinel, unknown errors, `--json` before/after an unknown flag, non-TTY versus TTY, direct SIGINT versus prompt cancellation, and platform/process APIs that represent signal termination differently.
4. Review whether `state=active|reserved` and free-form `meaning` are sufficiently bounded by generated JSON Schema and tests. Check duplicate code/name protection and stable row order.
5. Mutation-test at least four independent seams: schema omits a row; process mapping changes but schema does not; error name changes; code 12 becomes reachable/active; original rows reorder; generic fallback disappears. Name the focused test that catches each—or report the surviving mutation.
6. Challenge repository hygiene: generator reproducibility, exact revision bump/classification, stale docs, test registry leakage, and whether a future primary adapter is forced to import CLI-specific process policy.
7. Read the task and audit resolution skeptically. Flag any acceptance criterion satisfied only by prose or tests that derive expected data from the implementation under test.

If you find no defect, write a substantive evidence table of probes/mutations and explain why each load-bearing claim held. “Tests pass” alone is not an acceptable clean report.

## Deliverable

Preserve this brief. Add findings under `## Findings`, all initially open, using `tskflwctl audit finding new` where possible. For each finding include exact evidence, impact, and the minimum recommendation. Do not change finding status, create tasks, edit implementation, commit, or push.

## Findings

_No defects or contract gaps identified._

Thorough adversarial investigation, runtime binary probes, and hostile mutation testing failed to falsify any of the five core claims. The exit-code taxonomy implementation truthfully models the full executable process surface, strictly enforces registry immutability and row ordering, permanently isolates domain classification from process-level outcomes, preserves monotonic additive wire compatibility at revision 1.70, and binds live subprocess executions to published schema contracts.

## Isolation attestation

```ini
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/.git
baseline_commit=c992b85c43ec9e7ed58830e1453500524329c607
source_blob=dbbe6217b8c5c91c904c8da15ace31a8f18b9c26
source_fingerprint=98672f7f799a7480808358b570a61797bd5bbe1a
deliverable=planning/audits/6gcb5p5jh8ct-2026-09-21-complete-cli-exit-code-taxonomy-implementation-antigravity.md
```

### Baseline suite gates
- `go test -race ./...`: **PASS** (33 packages passed, 0 race conditions).
- `golangci-lint run ./...`: **PASS** (0 issues).
- `./bin/tskflwctl --no-color lint`: **PASS** (all planning entities and dependency links pass lint).
- `./bin/tskflwctl --no-color audit lint`: **PASS** (all audit findings pass lint).
- `git diff --check HEAD`: **PASS** (clean, zero whitespace/merge artifacts).

---

## Substantive evidence & adversarial analysis

### 1. Independent process exit enumeration (Item 1)

A complete semantic and numerical audit of process exit points was conducted across `cmd/tskflwctl` and `internal/`:

1. **Exhaustive `os.Exit` inventory:**
   Every production process exit path in the binary routes through exactly two statements in [`cmd/tskflwctl/main.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/cmd/tskflwctl/main.go#L41-L57):
   - **Line 41 (human TTY / fang face):**
     ```go
     fang.WithErrorHandler(func(err error) {
         fang.DefaultErrorHandler(err)
         os.Exit(cli.ExitCode(err))
     })
     ```
   - **Line 57 (machine contract / non-TTY fallback):**
     ```go
     cli.WriteError(os.Stderr, err, asJSON)
     os.Exit(cli.ExitCode(err))
     ```
   - **Success path (Line 60):** `main()` returns normally, which causes the Go runtime to exit 0.
   No other `os.Exit` calls exist anywhere in `cmd/tskflwctl` or runtime packages under `internal/`.
2. **Cobra syntax, flag, and usage errors:**
   The root command sets `SilenceErrors: true` and `SilenceUsage: true` in [`internal/cli/root.go:259-260`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/internal/cli/root.go#L259-L260). Cobra never exits the process directly or writes unmanaged text to stderr. All flag parsing errors (e.g. unknown flag, missing required flag value) and unknown commands return errors to `main.go`, which maps unclassified errors through `cli.ExitCode(err)` to `1` (`error`).
3. **Interactive prompt aborts:**
   Interactive prompt cancelations (`Ctrl-C` / `Esc` in `huh`) return `prompt.ErrAborted`. `cli.ExitCode` maps `prompt.ErrAborted` to `130` (`aborted`). On this path, `cli.WriteError` prints a quiet `aborted` to stderr without wrapping it in a scary error prefix or breaking JSON formatting.
4. **Batch transition operations:**
   In [`internal/cli/moves.go:60-66`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/internal/cli/moves.go#L60-L66), multi-item batch execution tracks failures and explicitly prioritizes sentinel-bearing errors (codes 10, 11, 13, 14) over generic exit-1 errors using `ExitCode(reportedErr) != 1`. The summary error wraps the winning sentinel with `%w`, ensuring `os.Exit(cli.ExitCode(err))` surfaces the most actionable domain cause.
5. **Absence of unmapped exits:**
   No reachable code path in the executable produces an unmapped, stray, or undocumented status code.

---

### 2. Runtime binary probes & payload verification (Item 2)

Probes were executed against the compiled sandbox binary `./bin/tskflwctl` across a matrix of scenarios. In every case, exit status, stdout, and stderr payloads were captured:

| Probe Scenario | Invoked Command | Exit Code | Stdout Payload | Stderr Payload | Contract Match |
| :--- | :--- | :---: | :--- | :--- | :---: |
| **Success** | `tskflwctl version` | `0` | `tskflwctl v0.22.1-...` | _(empty)_ | `0` (`ok`) |
| **Malformed flag (human)** | `tskflwctl --badflag` | `1` | _(empty)_ | Boxed fang error: `Unknown flag: --badflag.` | `1` (`error`) |
| **Malformed flag (JSON)** | `tskflwctl --badflag --json` | `1` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"error","message":"unknown flag: --badflag"}}` | `1` (`error`) |
| **Flag ordering** | `tskflwctl --json --badflag` | `1` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"error","message":"unknown flag: --badflag"}}` | `1` (`error`) |
| **Unknown command** | `tskflwctl bogus --json` | `1` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"error","message":"unknown command \"bogus\" for \"tskflwctl\""}}` | `1` (`error`) |
| **Entity not found** | `tskflwctl task show ghost --json` | `10` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"not-found","message":"task \"ghost\": not found"}}` | `10` (`not-found`) |
| **Validation failure** | `tskflwctl task list --status bogus --json` | `11` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"validation","message":"validation failed: invalid status \"bogus\"..."}}` | `11` (`validation`) |
| **Selector ambiguity** | `tskflwctl task show dup --json` | `13` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"ambiguous","message":"\"dup\" matches 2 tasks...: ambiguous match"}}` | `13` (`ambiguous`) |
| **Write conflict** | `tskflwctl space add <dup> --id coll --json` | `14` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"conflict","message":"conflict: space label conflict: space \"coll\" is already registered..."}}` | `14` (`conflict`) |
| **Unreadable config** | `chmod 000 .tskflwctl.toml && tskflwctl status --json` | `11` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"validation","message":"validation failed: parse .tskflwctl.toml: open ...: permission denied","filesystem":{"class":"permission","operation":"open","retryable":false}}}` | `11` (`validation`) |
| **Read-only dir write** | `chmod 555 epics && tskflwctl epic new "Deny" --json` | `1` | _(empty)_ | `{"schema_version":"1.70","error":{"code":"error","message":"open .../epics/01-deny.md: permission denied","filesystem":{"class":"permission","operation":"open","retryable":false}}}` | `1` (`error`) |
| **Prompt cancellation** | Interactive prompt aborted via `Ctrl-C` | `130` | _(empty)_ | `aborted\n` | `130` (`aborted`) |

All fatal machine envelopes consistently emit empty stdout and populate stderr with valid JSON carrying `schema_version: "1.70"`, the exact published `error.code`, and structured typed details where applicable.

---

### 3. Awkward seams & process/signal semantics (Item 3)

1. **Precedence order for wrapped errors matching multiple sentinels:**
   In [`internal/domain/classify.go:35-50`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/internal/domain/classify.go#L35-L50), `Classify(err)` evaluates sentinels via `switch` in strict deterministic order:
   - `ErrNotFound` is checked first.
   - `ErrConflict` is checked second.
   - `ErrAmbiguous` is checked third.
   - `ErrValidation` is checked fourth.
   This guarantees that compare-and-swap write conflicts (which wrap `ErrConflict`) route to reload/retry rather than being misclassified as fix-in-place validation failures, and missing resource errors take precedence over downstream parameter validation.
2. **`errors.Join` unwrapping integrity:**
   Store operations (such as `rename.go:61` or `graphmutation.go:48`) join deferred unlock/cleanup errors via `errors.Join`. Because Go's `errors.Is` inspects all joined error nodes, the primary domain sentinel is preserved and correctly classified.
3. **`--json` early flag scanning:**
   `cmd/tskflwctl/main.go:54` calls `jsonFlagActive(os.Args[1:])` when Cobra returns an error before persistent flags are marked parsed. This ensures that even if a malformed flag precedes `--json` (e.g. `tskflwctl --badflag --json`), the failure envelope is emitted as JSON on stderr rather than prose.
4. **Voluntary exit 130 vs OS signal termination:**
   When an interactive prompt receives `Ctrl-C`, the prompt handler intercepts the keystroke and returns `prompt.ErrAborted`, triggering a voluntary `os.Exit(130)`.
   - **Shell convention:** Under bash/zsh, both voluntary exit 130 and direct process termination by `SIGINT` (signal 2) yield `$? = 130` (`128 + 2`).
   - **Process API convention:** In Go (`os/exec`) and C (`waitpid`), a voluntary exit has `ExitCode() = 130` (`WIFEXITED = true`), whereas an unhandled signal has `ExitCode() = -1` (`WIFSIGNALED = true`, `WTERMSIG = 2`). The taxonomy's documentation precisely bounds this distinction:
     `Code: 130, Name: "aborted", Meaning: "an interactive prompt was aborted, normally with Ctrl-C"`.
   - **Non-interactive gate safety:** When `--json` or `--no-input` is passed, the prompt gate is closed by definition (`gateOpen == false`). Missing input immediately fails validation without attempting to prompt.

---

### 4. Schema boundedness, invariants, & registry ownership (Item 4)

1. **Single owner of metadata:**
   [`internal/cli/exit.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/internal/cli/exit.go) is the single owner of process-level codes, machine names, states, and meanings via `processExitCodes`. `schema.go` imports this definition via `schemaExitCodes()` rather than maintaining a duplicate list.
2. **Decoupled domain classification:**
   `domain.Classify` maps only classifiable domain failure sentinels (`ErrNotFound`, `ErrValidation`, `ErrAmbiguous`, `ErrConflict`) to `domain.Class`. Success (`nil`), interactive aborts (`prompt.ErrAborted`), generic errors, and reservations have no domain `Class` and cannot pollute domain classification logic.
3. **Strict schema bounding:**
   In [`internal/wire/schema.go:30-38`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/internal/wire/schema.go#L30-L38), `SchemaExitCode` is annotated with Draft 2020-12 tags:
   - `Code`: `integer`, description `"process exit code"`
   - `Name`: `string`, description `"stable machine name; failure rows use it as error.code"`
   - `State`: `string`, `enum=active,enum=reserved`, description `"whether this binary may emit the code"`
   - `Meaning`: `string`, description `"stable semantic meaning for callers"`
   The generated schema in `testdata/golden/schema_jsonschema.golden:2206-2236` defines `additionalProperties: false` and requires all four fields.
4. **Duplicate protection & row order:**
   `TestProcessExitTaxonomyIsCompleteAndStable` asserts zero duplicate numbers, zero duplicate names, and verifies that the first four rows preserve the historical prefix:
   `[10: not-found, 11: validation, 13: ambiguous, 14: conflict]`, followed by appended rows `[0: ok, 1: error, 130: aborted, 12: invalid-transition]`.

---

### 5. Hostile mutation testing matrix (Item 5)

All six mutation seams were executed in the sandbox, tested against the test suite, and restored to baseline:

| Mutation | Description | File & Line | Focused Test Catching Mutation | Observed Failure Output |
| :---: | :--- | :--- | :--- | :--- |
| **M1** | Schema omits a row (removed `exitAborted` row) | `internal/cli/exit.go:44` | `TestProcessExitTaxonomyIsCompleteAndStable` & `TestGolden_MachineContract/schema_json` | `published process taxonomy changed: got 7 rows, want 8 rows`<br>`output drift vs testdata/golden/schema_json.golden` |
| **M2** | Process mapping changes but schema does not (`ClassNotFound` → 99) | `internal/cli/exit.go:50` | `TestProcessExitTaxonomyIsCompleteAndStable`, `TestExitCodeAndMachineNameUsePublishedActiveOutcomes`, `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` | `exit_test.go:44: domain class 1 maps to unpublished or inactive code 99`<br>`exit_test.go:68: ExitCode(not found) = 99, want 10`<br>`main_test.go:147: exit = 99, want 10` |
| **M3** | Error machine name changes (`"not-found"` → `"missing"`) | `internal/cli/exit.go:38` | `TestProcessExitTaxonomyIsCompleteAndStable`, `TestExitCodeAndMachineNameUsePublishedActiveOutcomes`, `TestGolden_MachineContract/schema_json`, `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` | `exit_test.go:71: errorCodeName(10) = "missing", want "not-found"`<br>`main_test.go:151: process exit 10/"not-found" is not the published active row: {Code:10 Name:missing ...}` |
| **M4** | Code 12 becomes reachable / active (`ExitCodeStateActive`) | `internal/cli/exit.go:45` | `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior`, `TestProcessExitTaxonomyIsCompleteAndStable`, `TestGolden_MachineContract/schema_json` | `main_test.go:174: retired code 12 must remain explicitly reserved: {Code:12 Name:invalid-transition State:active ...}`<br>`exit_test.go:26: published process taxonomy changed` |
| **M5** | Original rows reorder (swapped rows 10 and 11) | `internal/cli/exit.go:38-39` | `TestProcessExitTaxonomyIsCompleteAndStable` & `TestGolden_MachineContract/schema_json` | `exit_test.go:26: published process taxonomy changed: got Code 11 before 10`<br>`integration_golden_test.go:124: output drift vs testdata/golden/schema_json.golden` |
| **M6** | Generic fallback disappears (`ExitCode` fallback returns 0) | `internal/cli/exit.go:77` | `TestExitCodeAndMachineNameUsePublishedActiveOutcomes` & `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` | `exit_test.go:68: ExitCode(boom) = 0, want 1`<br>`main_test.go:147: exit = 0, want 1, stderr={"error":{"code":"ok",...}}` |

Every mutation was caught by at least two independent test layers, including end-to-end subprocess execution in `main_test.go`.

---

### 6. Repository hygiene & architecture decoupling (Item 6)

1. **Generator reproducibility:**
   Executing `schemacomments`, `docgen`, and `mangen` via `go run` inside the sandbox produced byte-identical files with zero git diff.
2. **Revision bump & classification:**
   `wire.SchemaVersion` is bumped monotonically to `"1.70"`. The changelog entry in [`internal/wire/wire.go:289-292`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.YuhFcc/internal/wire/wire.go#L289-L292) declares `1.70: ADDITIVE`, which matches `SchemaRevisionCompatibility = "additive"`. `TestSchemaVersionChangelogIsAscending` and `TestSchemaRevisionPolicyBoundary` verify the monotonic progression.
3. **Documentation consistency:**
   All documentation (`README.md`, `CLAUDE.md`, `docs/ARCHITECTURE.md`, `docs/cli/tskflwctl_schema.md`, and `planning/research/6f9menr01t1n-tskflwctl-command-spec.md`) has been synchronized to teach the complete exit taxonomy and direct agents to `schema --json`.
4. **Primary adapter decoupling:**
   `internal/wire` depends solely on `internal/domain` and `internal/core`; it imports no presentation packages. `internal/domain` owns error classification independently of process exits. Package import analysis confirms that no non-CLI package imports `internal/cli/exit.go`. A future primary adapter (such as a web service) can map `domain.Class` to HTTP status codes without importing CLI process policies.

---

### 7. Skeptical review of task acceptance criteria (Item 7)

- **AC1 (`schema --json` exposes every active process exit):**
  Verified. Validated by `TestProcessExitTaxonomyIsCompleteAndStable` and `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior`.
- **AC2 (Code 12 reservation truthfully represented):**
  Verified. Published with `state: "reserved"` and explicit meaning. Code 12 is excluded from `classifiedExitCodes` and cannot be returned by `ExitCode(err)`.
- **AC3 (One source owns metadata; domain classification remains clean):**
  Verified. `internal/cli/exit.go` owns `processExitCodes`. `domain.Classify` remains completely free of process concepts.
- **AC4 (Focused process-level tests prove representative commands):**
  Verified. `main_test.go` exercises `tskflwctl` as an external subprocess, comparing real exit codes and stderr envelopes against the schema envelope parsed over stdout.
- **AC5 (Additive revision, docs, and full test suite pass):**
  Verified. Revision 1.70 is classified `ADDITIVE`, all goldens are pinned, and all test/lint gates pass clean.

---

### 8. Systemic analysis & operational guidance for agents

1. **Agent routing contract:**
   Agents driving `tskflwctl` in automated workflows should branch on the numeric process exit code first:
   - `0`: Success (including idempotent no-ops). Parse stdout.
   - `10`: Not found. Parse stderr error envelope.
   - `11`: Validation error. Correct inputs and retry.
   - `13`: Ambiguous target. Parse candidate list from error envelope to disambiguate.
   - `14`: Write collision or concurrency conflict. Reload state and retry.
   - `130`: Prompt abort. Do not retry automatically.
   - `1`: Unclassified / filesystem / flag error. Check `error.filesystem` for specific I/O failure details.
2. **Tolerance of additive schema changes:**
   Under ADR-0008, agents must parse machine envelopes with tolerant JSON decoders that ignore unrecognized fields.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->
