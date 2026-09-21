---
schema: 1
id: 6gcb5n1q8w97
bucket: closed
area: complete-cli-exit-code-taxonomy-implementation-claude
date: "2026-09-21"
updated_at: "2026-09-21"
---
# Audit: Complete CLI exit-code taxonomy implementation — Claude — 2026-09-21

> Reviewer assignment: Claude. This document is the review brief and the only source file the reviewer may update.
>
> Take two passes: first public-contract compatibility and source ownership, then systemic failure modes. Treat green tests as claims to challenge, not proof.

## Mandatory isolated workspace

Treat the handoff checkout as read-only. Resolve this audit, then create an independent sandbox before inspecting code, running tests, generators, or mutation probes:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_ABS="$(tskflwctl audit path 2026-09-21-complete-cli-exit-code-taxonomy-implementation-claude)"
AUDIT_REL="${AUDIT_ABS#"$SOURCE_ROOT"/}"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

Do all inspection, builds, tests, probes, and report editing in the sandbox. Do not stage, commit, switch branches, restore, clean, stash, reset, or run write-capable commands in the source checkout. Restore every probe in the sandbox so only the assigned audit differs, then verify and transfer it:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Include the helper attestation and transfer result in the report. Preserve the sandbox until the implementation owner confirms receipt.

## Review target

Adversarially review task `6gbn4g1bb7wr`, `publish-the-complete-cli-exit-code-taxonomy`, against `main`. Inventory and inspect:

- `internal/cli/exit.go`, `schema.go`, their focused tests, and every caller of `ExitCode` / `WriteError`;
- `cmd/tskflwctl/main.go` plus the real-binary process tests;
- `internal/wire/schema.go`, schema revision 1.70, generated comments/schema, and machine goldens;
- human schema rendering and README, CLAUDE, architecture, generated CLI, and command-spec guidance;
- ADR-0008 and the originating architecture-audit M1 finding.

Do not credit planning prose as shipped behavior. Do not implement fixes or edit any file other than this audit.

## Intended contract to challenge

- One CLI-owned registry publishes stable process code, machine name, meaning, and `active|reserved` state.
- The previous published rows `10/11/13/14` remain an unchanged prefix. Rows `0/1/130/12` append, and `state`/`meaning` are additive object fields under ADR-0008.
- Code 12 is retired and reserved; no current execution path emits it.
- Domain classification remains adapter-neutral and covers only not-found, validation, ambiguity, and conflict. Success, generic errors, prompt abort, and reservations do not become domain classes.
- Active failures use the published name as `error.code`; success has no error envelope; prompt abort remains a quiet human-only exit 130.
- Schema revision 1.70, generated JSON Schema, human schema, docs, and goldens agree.

## Evidence floor and hostile probes

1. Independently enumerate every deliberate process-exit path from `main`, Cobra/fang routing, prompts, and `ExitCode`; compare it to `schema --json` without assuming the new table is complete.
2. Build the real binary and reproduce representative 0, 1, 10, 11, 13, and 14 exits. Verify stdout/stderr separation and exact JSON error names. Exercise abort 130 through a PTY if feasible; otherwise prove the prompt-error path and state the limitation.
3. Verify code 12 cannot be returned, is not represented as active, and cannot silently acquire a new meaning.
4. Check uniqueness and ordering of codes/names, fallback behavior for unknown errors, wrapped sentinel precedence, and the relationship between domain classes and process metadata.
5. Challenge the additive claim: original array prefix, added object fields, tolerant-reader policy, exact-revision JSON Schema, changelog classification, and every versioned golden must agree.
6. Mutate at least three load-bearing points: remove an active row, change one code/name, make code 12 active or reachable, reorder the original prefix, or disconnect schema assembly from the registry. Require focused tests to fail for each mutation.
7. Check the subprocess test is hermetic: no leaked home registry state, dependence on test order, accidental prose parsing, or false confidence from comparing output to the same faulty source.
8. Inspect human output and all guidance for stale claims such as “10–14 only,” active code 12, or an obsolete JSON error shape.

For every defect, cite exact paths/lines and a minimal reproduction or mutation result. Distinguish a demonstrated bug from a hardening suggestion.

## Deliverable

Preserve this brief. Add findings under `## Findings`, all initially open, using `tskflwctl audit finding new` where possible. A substantive no-findings report is acceptable only after completing the evidence floor and documenting the probes. Do not change finding status, create tasks, edit implementation, commit, or push.

## Review evidence

Reviewer: Claude (Opus 5), 2026-09-21. All inspection, builds, tests, probes, and mutations ran in the isolated sandbox. The only file changed there is this audit.

**Workspace attestation** (`isolated-review-workspace.sh verify`, run before this section was written; `transfer=pending` at that point):

```
sandbox_path=/private/tmp/claude-501/-Users-andyeschbacher-git-andy-esch-taskflow/87c8a001-534f-4832-992d-a10443fc3e76/scratchpad/isolated-review.ODnfmJ
baseline_commit=9b3a31d624d84928b20de56e85871bf7a6e18e7b
source_blob=2fb56462eec5b5df0d1ba4462f861ad0aa92fa1f
source_fingerprint=98672f7f799a7480808358b570a61797bd5bbe1a
deliverable=planning/audits/6gcb5n1q8w97-2026-09-21-complete-cli-exit-code-taxonomy-implementation-claude.md
deliverable_changed=true
transfer=pending
```

The sandbox was created with `--sandbox-parent` set to the session scratchpad instead of the default `$TMPDIR`; everything else followed the brief. The transfer can only happen once, after this file is final, so the transfer result is reported in the reviewer's handoff message, not here.

**Target.** `origin/main` = `b4ddf7b`. The diff under review is the handoff working tree captured in baseline `9b3a31d`. Inventory: `internal/cli/exit.go` (registry, `ExitCode`, `errorCodeName`, `schemaExitCodes`, `WriteError`), `internal/cli/exit_test.go`, `schema.go` + `schema_test.go`, `render/schema_render.go`, `cmd/tskflwctl/main.go` + `main_test.go` + `testmain_test.go`, `internal/wire/{schema.go,wire.go,schema_comments.json,envelopes_test.go}`, the 38 changed JSON machine goldens, the revision marker, README/CLAUDE/ARCHITECTURE, `docs/cli/tskflwctl_schema.md`, the command-spec research, ADR-0008, and architecture-audit M1. Production callers of `ExitCode`: `main.go:41` (fang path), `main.go:57` (machine path), `main.go:66` (the abort check in `fangErrorHandler`), and `exit.go:128` (the envelope name inside `WriteError`). Production callers of `WriteError`: `main.go:56` and `main.go:67`.

**Gates (sandbox).** `just build` ok. `go test -race ./...`: 25 packages ok, exit 0. `golangci-lint run ./...`: 0 issues. `tskflwctl lint` and `audit lint` are clean. `git diff --check` is clean. Regenerating `docs/cli` (docgen) and `schema_comments.json` produced no drift. No machine golden still names 1.69, and `machine_contract_revision.txt` is 1.70.

**1. Exit-path enumeration.** The binary exits deliberately in three places. Success returns 0 (`main.go:43` and the fall-through). The fang human path calls `os.Exit(ExitCode(err))` (`main.go:41`), and the machine path does the same (`main.go:57`). `ExitCode` yields only 0, 130 (`prompt.ErrAborted`), 10/11/13/14 (via `classifiedExitCodes`), and 1 (the fallback). fang v2.0.1 never calls `os.Exit`, and signal notification is not enabled. `huh.ErrUserAborted` maps to `ErrAborted` at both prompt sites (`prompt/tty.go:108,127`). Other prompt and TUI errors fall back to 1, and a TUI quit is 0. The only other `os.Exit`/`log.Fatal` calls live in `internal/tools/*`, which is not shipped. Every deliberate outcome therefore matches `schema --json`. Two non-deliberate outcomes are unpublished and are not findings: an unrecovered Go panic exits 2, and fatal signals (for example SIGPIPE) are reported by the shell as 128+n.

**2. Real-binary probes** (hermetic `TSKFLW_CONFIG_HOME`; stdout and stderr captured separately):

```
version                               exit=0   stdout=29B  stderr=∅
--badflag --json                      exit=1   stdout=∅    {"code":"error","message":"unknown flag: --badflag"}
no-such-command --json                exit=1   stdout=∅    {"code":"error",…}
task show ghost --json                exit=10  stdout=∅    {"code":"not-found",…}
task list --status bogus --json       exit=11  stdout=∅    {"code":"validation",…}
task show same --json                 exit=13  stdout=∅    {"code":"ambiguous",…}
space add <other> --id coll --json    exit=14  stdout=∅    {"code":"conflict",…}
read-only tasks/: task new … --json   exit=1   stdout=∅    {"code":"error","filesystem":{"class":"permission",…}}
```

All seven error envelopes above validate against the same binary's exact-revision `schema --json-schema` (`#/$defs/ErrorEnvelope`), and so does `schema --json` (`#/$defs/SchemaEnvelope`). This was checked with an out-of-tree santhosh-tekuri validator. As ADR-0008 §3 intends, the 1.69 schema rejects the 1.70 payload (`schema_version` const plus `additionalProperties`). On the fang path the codes are preserved: a PTY run gives 10, 13, 11, and 1 with styled prose. **Abort 130 was exercised through a real PTY** (Python `pty.fork`): `task new "Prompted" --description d --tags t` opens the epic picker, and both Ctrl-C and Esc exit **130** with a quiet `aborted` line and no file written. With `--json` on a PTY the gate stays closed, so the same command exits 11 with an envelope (`--epic is required`). Pre-existing and out of scope: batch transitions and lists or lint with unreadable files deliberately print a payload on stdout **and** an envelope on stderr (for example `task complete ghost --json` → 10), so `WriteError`'s comment "stdout stays empty on failure" is not universally true.

**3. Code 12.** It is not in `classifiedExitCodes`, it is published as `reserved`, and no current path returns it. Mutations M3 and M5 (below) are killed. Direct emission is **not** guarded; see M2.

**4. Uniqueness, order, fallback, precedence.** Codes and names are unique, and this is tested. The order is the 1.69 prefix plus appended rows, and the 1.69 code/name pairs are byte-identical. A scratch probe test, since removed, showed: `errors.Join(ErrAborted, ErrNotFound)` → 130; `Join(ErrValidation, ErrNotFound)` → 10; `Join(ErrAmbiguous, ErrConflict)` → 14; `%w: %w`(Validation, Ambiguous) → 13. So abort wins, followed by `domain.Classify`'s unchanged NotFound > Conflict > Ambiguous > Validation order. `errorCodeName` gives 0 → "ok", 999 → "error", and 12 → "invalid-transition" (see M2). `internal/domain` is untouched. Its five classes (Unknown plus four) map only to active rows, and Unknown falls through to 1.

**5. Additive claim.** Object fields are additive and the prefix is stable. The closed-vocabulary and documented-meaning analysis is in M1.

**6. Mutations.** Each mutation was run with `go test -count=1 ./internal/cli ./cmd/tskflwctl ./internal/wire` and restored with `git checkout` in the sandbox:

| # | Mutation | Result |
|---|---|---|
| M1 | remove active row 130 | **killed**: taxonomy DeepEqual, name case `aborted`, `schema_json` golden |
| M2 | rename `conflict` → `collision` | **killed**: taxonomy, name case, golden, subprocess `conflict`, recovery-envelope tests |
| M3 | make 12 `active` | **killed**: taxonomy, golden, subprocess reserved assertion |
| M4 | swap 10/11 in the original prefix | **killed**: taxonomy, golden |
| M5 | map ClassConflict → 12 in the table | **killed**: taxonomy mapping guard, name case, subprocess `conflict` |
| M6 | direct `return exitInvalidTransition` for filesystem errors in `ExitCode` | **survived**; the full `go test ./...` is also green (M2) |
| M7 | replace `schemaExitCodes()` in `schema.go` with a stale 4-row literal | **killed**: `TestSchemaContract_JSON`, golden, every subprocess case |
| M8 | drop the human `[reserved]` marker | **survived** (L2) |
| M9 | hide reserved rows from human output | **survived** (L2) |
| M10 | drop the `ErrAborted` → 130 mapping | **killed**: name case `aborted`, `TestExitCode_Abort` |

**7. Subprocess hermeticity.** `TestMain` pins `TSKFLW_CONFIG_HOME` to a temp dir, so `space add`/`forget` never touch the real home registry, and a check of the real registry found no test entry. The test passes with an ambient `TSKFLW_SPACE=bogus-space` because `-C` wins, and it also passes with `-count=3 -shuffle=on`. It decodes JSON rather than parsing prose. Each case asserts literal `wantCode`/`wantName` values as well as the self-published row, so it does not only compare the binary against itself. Pre-existing and out of scope: `binary()` never removes its `tskflwctl-smoke-*` build dir; this machine's `$TMPDIR` holds 416 such dirs (≈930 MB).

**8. Guidance sweep.** README, CLAUDE.md, ARCHITECTURE, generated `docs/cli/tskflwctl_schema.md`, `schema --help`, and `routines/code-quality-audit.md` contain no stale "10–14 only", no active-12 claim, and no obsolete envelope shape. "10–14" still appears in completed tasks and historical research, which are records rather than guidance. The command-spec residue is L3. A minor observation, not a finding: `fangErrorHandler` hard-codes `130` (`main.go:66`) because the registry constant is unexported. A registry change would be caught by the literal taxonomy test, but not the handler's copy.

**Planning.** Architecture-audit M1 is `tracked by 6gbn4g1bb7wr` and that audit is now `closed`. That matches the settled-share policy, and `audit close` does not stamp `updated_at`, so the unchanged date is expected. The task's ticked criteria are credited only where the probes above confirm them. Criterion 5's claim that the additive revision is correct is contested by M1.

## Findings

#### M1. Revision 1.70 is labeled ADDITIVE although ADR-0008 defines its exit-table change as NOT ADDITIVE · **Status:** fixed

**File:** internal/wire/wire.go:289 | **Component:** wire / machine contract
**Effort:** XS · **Urgency:** soon

The 1.70 changelog entry (`internal/wire/wire.go:289-292`) and `SchemaRevisionCompatibility = "additive"` (`internal/wire/wire.go:308`) publish `revision_policy.current_compatibility: "additive"`. ADR-0008 §3 (`planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md:124-129`) defines **NOT ADDITIVE** as including "a published closed vocabulary adds, removes, or renames a member that an exhaustive consumer may switch on" and any change to a "documented meaning". `docs/THREADS_COMPATIBILITY.md:21` explicitly lists "error codes" as part of the revisioned contract and says "closed-vocabulary changes are non-additive because exhaustive consumers may switch on them". 1.70 does both of these things:

- **Members added to a published vocabulary.** `exit_codes[].name` goes from `{not-found, validation, ambiguous, conflict}` to also include `ok`, `error`, `aborted`, and `invalid-transition`.
- **Documented row meaning changed.** In 1.69 the JSON Schema described every row as "one exit code and its stable machine name (also the `code` in the --json error envelope)". In 1.70, three rows never produce an envelope: `0/ok`, `130/aborted`, and reserved `12`.

Reproduction (1.69 binary built from `origin/main`, 1.70 binary built from the sandbox):

```
$ tskflwctl schema --json | jq -c '{v:.schema_version, compat:.revision_policy.current_compatibility,
    exhaustive_names_unchanged: ([.exit_codes[].name]|sort == ["ambiguous","conflict","not-found","validation"]),
    zero_listed_as_error_row: ([.exit_codes[].code]|index(0) != null)}'
1.69 → {"v":"1.69","compat":"additive","exhaustive_names_unchanged":true,"zero_listed_as_error_row":false}
1.70 → {"v":"1.70","compat":"additive","exhaustive_names_unchanged":false,"zero_listed_as_error_row":true}
```

The documented 1.69 meaning makes a simple check reasonable: if a status appears in `exit_codes[].code`, it is a documented failure. On 1.70 that check classifies success (`0`) as a failure. The published reader expectation, `ignore-unknown-object-fields`, covers the new `state`/`meaning` keys but not new array members. The changelog justifies the label only by prefix stability.

**Mitigating:** the codes that actually appear in error envelopes did not change. `error` was already emitted before 1.70, and `ok`, `aborted`, and `invalid-transition` never appear in an envelope. The code/name prefix is byte-identical. Only consumers of the table itself break. **Demonstrated:** policy-classification defect. It is not a runtime bug, but it is exactly the truthfulness judgment ADR-0008 assigns to review.

**Recommendation:** Before merge, either reclassify 1.70 as NOT ADDITIVE with a consumer-action note, or amend ADR-0008 to declare exit_codes an append-only registry and publish that rule in revision_policy and THREADS_COMPATIBILITY.

**Resolution:** Reclassified schema revision 1.70 as NOT ADDITIVE under ADR-0008
and regenerated the executable policy, schema, docs, and goldens.

#### M2. Reserved code 12 is guarded only at the classified-table boundary; a direct emission path passes the whole suite · **Status:** fixed

**File:** internal/cli/exit.go:82 | **Component:** cli / exit taxonomy
**Effort:** S · **Urgency:** soon

`errorCodeName` (`internal/cli/exit.go:82-88`) looks up **every** `processExitCodes` row, reserved rows included. Nothing between `ExitCode` and `os.Exit`/`WriteError` checks that an emitted code is an active row. A review probe confirmed that `errorCodeName(12) == "invalid-transition"`. The only reachability guard is `internal/cli/exit_test.go:75`, and it checks just one input: `ExitCode(domain.ErrValidation)`. The structural guard in `TestProcessExitTaxonomyIsCompleteAndStable` covers the `classifiedExitCodes` table, not the body of `ExitCode`.

**Mutation M6 (survived):** insert `if filesystemDetails(err) != nil { return exitInvalidTransition }` before `domain.Classify` in `ExitCode`. The focused packages (`./internal/cli ./cmd/tskflwctl ./internal/wire`) **and** the full `go test ./...` still pass. The mutant binary against a read-only `tasks/` directory:

```
unmutated: task new "Blocked" --epic 01-exit ... --json → exit=1  {"code":"error","fs":"permission"}
mutant:    task new "Blocked" --epic 01-exit ... --json → exit=12 {"code":"invalid-transition","fs":"permission"}
mutant:    schema --json → {"code":12,"name":"invalid-transition","state":"reserved","meaning":"retired and reserved; this binary never emits it"}
```

So code 12 can quietly take on a new meaning with every test green. This is the brief's hostile probe 3 failing. There is a related gap: the published meaning of `1` explicitly names "filesystem" failures, which M1 of the architecture audit calls the common exit-1 path. Yet no test pins a filesystem failure to `1`/`error`. `internal/cli/fserror_test.go` decodes the envelope but never asserts `Error.Code`, and the subprocess `generic` case (`cmd/tskflwctl/main_test.go:137`) is a flag-parse error only.

For comparison, mutations M1–M5, M7, and M10 were all killed (see Review evidence). **Classification:** test/hardening gap. The shipped binary behaves correctly.

**Recommendation:** Resolve every emitted code through one lookup that accepts only active rows, and pin a filesystem failure to 1/error in both unit and subprocess tests.

**Resolution:** All candidate outcomes now pass through an active-only registry
guard; reserved and unknown codes fall back to 1/error, with unit and
real-process filesystem coverage.

#### L1. Published exit-row semantics overstate error.code coverage and what code 10 means · **Status:** fixed

**File:** internal/wire/schema.go:35 | **Component:** wire / machine contract
**Effort:** XS · **Urgency:** eventually

- **`aborted` never reaches `error.code`.** The field description reads "stable machine name; failure rows use it as error.code" (`internal/wire/schema.go:35`), and the type comment says "Failure names are also emitted as the `code` in the --json error envelope". The active failure row `130/aborted` can never appear there: the prompt gate is closed under `--json` (`internal/cli/color.go:82-84`), and `WriteError` special-cases the abort to prose (`internal/cli/exit.go:119-122`). The PTY probe confirmed a quiet `aborted` line with exit 130. No machine-readable field tells a consumer which active failure rows can appear in an envelope.
- **A missing planning repo is not exit 10.** Code 10's meaning is "the requested entity or resource does not exist". But `-C /does/not/exist task list --json` exits **1** with `{"code":"error","message":"not a taskflow planning repo: …"}`, while `--space nosuch task list --json` exits **10**. The task puts changing exit behavior out of scope, so the meaning text is what overreaches.

**Classification:** contract-precision defect. It is low-impact because a tolerant consumer sees a superset.

**Recommendation:** Scope the error.code description to envelope-producing rows (or add a per-row flag) and narrow code 10's meaning to named entities and registered spaces.

**Resolution:** Narrowed code 10 to named entities and registered spaces and
documented exactly which active failures can appear as error.code.

#### L2. Human schema exit table is untested and inherits the wire's append order · **Status:** fixed

**File:** internal/cli/render/schema_render.go:72 | **Component:** cli / human render
**Effort:** XS · **Urgency:** eventually

The task changed `SchemaHuman` (`internal/cli/render/schema_render.go:72-79`) to add the meaning column and the `[reserved]` marker, but no test exercises this rendering. There is no human-schema golden. Mutations **M8** (drop the `[reserved]` marker) and **M9** (skip reserved rows, hiding code 12 from humans entirely) both pass `./internal/cli ./cmd/tskflwctl ./internal/wire`.

The human table also inherits the wire's append order:

```
  10  not-found … 11 validation … 13 ambiguous … 14 conflict … 0 ok … 1 error … 130 aborted
  12  invalid-transition  retired and reserved; this binary never emits it [reserved]
```

Append order exists to keep the machine prefix stable. ADR-0008 §2 excludes human rendering from the revision, though, so humans pay for a machine-only constraint. The reserved row also says "reserved" twice. **Classification:** hardening/UX.

**Recommendation:** Add a focused human-schema assertion or golden for the reserved row, and consider sorting the human table by code.

**Resolution:** Human schema rows now sort numerically and a focused renderer
test pins the reserved marker and row visibility.

#### L3. Command-spec exit guidance still uses ERR_* labels and a structured candidate list that are not on the wire · **Status:** fixed

**File:** planning/research/6f9menr01t1n-tskflwctl-command-spec.md:230 | **Component:** docs / agent guidance
**Effort:** XS · **Urgency:** eventually

The task rewrote the exit-code bullet (`planning/research/6f9menr01t1n-tskflwctl-command-spec.md:229-237`) to point agents at `schema --json`, but kept labels that never appear on the wire:

- It uses `ERR_NOT_FOUND`, `ERR_VALIDATION`, `ERR_AMBIGUOUS_TARGET`, `ERR_LOCK_CONFLICT`, and `ERR_INVALID_TRANSITION`. The published names are `not-found`, `validation`, `ambiguous`, `conflict`, and `invalid-transition`.
- `ERR_LOCK_CONFLICT` narrows 14 relative to its published meaning, "a write collided with existing or concurrently changed state". The smoke test's own conflict case, a space-label collision, is not a lock.
- Line 239 still promises "exit `13` + candidate list". The real ambiguous envelope carries its candidates only as prose inside `message`: `{"code":"ambiguous","message":"\"same\" matches 2 tasks: same (6gcb770xxgvj), same (6gcb7713q312): ambiguous match"}`.

This is planning prose, not shipped behavior, but the brief lists it as agent guidance. **Classification:** low-severity documentation drift.

**Recommendation:** Replace the ERR_* labels with the published names and mark the structured ambiguity candidates as not implemented.

**Resolution:** Updated command guidance to the published wire names and
documented that ambiguity candidates remain prose pending a future typed
contract.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ✔ M1 · fixed — Reconcile the 1.70 compatibility label with ADR-0008's closed-vocabulary rule
- ✔ M2 · fixed — Enforce active-only exit emission and test the filesystem exit-1 path
