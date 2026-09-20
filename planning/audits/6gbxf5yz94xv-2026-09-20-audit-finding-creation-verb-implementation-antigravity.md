---
schema: 1
id: 6gbxf5yz94xv
bucket: closed
area: audit-finding-creation-verb-implementation-antigravity
date: "2026-09-20"
---

# Audit: Audit finding creation verb implementation — Antigravity — 2026-09-20

> Reviewer assignment: Antigravity. This document is the review brief and the only source file the reviewer may update.
>
> Take two passes: first contract/correctness, then systemic failure modes. Green tests are claims to challenge, not proof. Prefer one demonstrated defect over several speculative concerns.

## Mandatory isolated workspace

Treat the handoff checkout as read-only. Create an independent sandbox before inspecting code,
running tests, generators, or mutation probes:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6gbxf5yz94xv-2026-09-20-audit-finding-creation-verb-implementation-antigravity.md"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)"
cd "$SANDBOX"
```

Do all inspection, builds, tests, probes, disposable-repository work, and report editing in the
sandbox. Do not stage, commit, switch branches, restore, clean, stash, reset, or run write-capable
commands in the source checkout. Restore every probe in the sandbox so only the assigned audit
differs, then verify and transfer it:

```sh
"$SANDBOX/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX"
git diff -- "$AUDIT_REL"
"$SANDBOX/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

Include the helper attestation and transfer result in the report. Preserve the sandbox until the
implementation owner confirms receipt.

## Review target

Adversarially review the complete working-tree implementation of task `6gbpe6e8n87k`,
`add-a-tool-owned-audit-finding-creation-verb`, against `main`. Inventory and inspect:

- `internal/domain/finding_create.go`, finding/candidate grammars, fence and section helpers, and tests;
- `internal/core/finding_create.go`, the audit mutation port, retry policy, and test fakes;
- `internal/store/body.go` and real-filesystem concurrency tests;
- `internal/cli/audit.go`, renderers, integration tests, generated CLI reference, and error mapping;
- wire DTO/envelope/schema changes, revision 1.69 changelog, generated JSON schema/comments, and goldens;
- default/security audit scaffolds, README/CLAUDE/architecture guidance, and both scheduled audit routines;
- the task decision, acceptance criteria, and recorded implementation evidence.

Do not credit planning prose as shipped behavior. Do not implement fixes or edit any file other than
this audit.

## Intended contract to challenge

- `audit finding new <audit> <title> --band H|M|L` creates one open finding without a caller-authored
  code or structural Markdown. Optional file, component, effort, urgency, body, recommendation, and
  candidate values are validated semantic inputs.
- Allocation is audit-local and monotonic per band: highest numeric suffix plus one, with no gap
  reuse. Allocation is recomputed inside every CAS retry.
- The writer locates one real, fence-aware Findings section or creates it before a real Candidate
  tasks section. Content and whitespace outside the insertion remain unchanged.
- Existing near-miss headers, duplicate codes, malformed findings, ambiguous sections, and malformed
  managed candidate projections fail closed with useful repair guidance.
- Reserved parser-owned metadata cannot be injected through title, evidence, or recommendation;
  literal examples remain possible inside fences.
- An optional candidate row is created with the finding in one atomic transform and only against a
  valid `candidate-tasks:v1` section. Missing/legacy sections refuse without a partial finding.
- The callback receives audit metadata and body from the same guarded snapshot. Creation is refused
  for closed/deferred audits, including when lifecycle state changes concurrently.
- Dry-run performs all validation without writing. Human output names the allocated code. JSON is a
  compact revision-1.69 receipt with audit, finding, dry-run, and workspace identity, not the full body.
- Existing `audit finding <audit> <code>` behavior and existing JSON envelopes remain compatible.

Non-goals: migrating legacy Candidate sections; changing finding status vocabulary; moving findings
out of Markdown; recycling deleted codes; bulk finding import; or making arbitrary hand-authored
audit prose canonical.

## Mandatory evidence floor

1. Run `git diff --check`, `go test -race ./...`, golangci-lint with writable caches under
   `/tmp`, and planning/audit lint with the sandbox-built binary.
2. In a disposable planning repo, run the exact documented path: create an audit, create findings in
   all bands with every optional field, create one with a managed candidate, preview one via
   `--dry-run --json`, update the created finding, and lint the result. Inspect the exact Markdown.
3. Exercise missing/legacy/malformed/multiple Candidate sections and prove failure is byte-identical.
   Exercise an audit with no Findings section, multiple Findings sections, following H1/H2/H3/H4
   sections, Findings/Candidate examples inside backtick and tilde fences, CRLF, Unicode, whitespace-
   only tails, metadata-label injection, near-miss headers, invalid statuses, duplicate codes, huge
   numeric suffixes, and empty optional values.
4. Exercise closed and deferred audits. Coordinate a lifecycle change with creation so the bucket
   guard is tested against the exact store snapshot, not a convenient pre-read.
5. Run repeated and coordinated concurrent creators in the same and different bands, plus creation
   racing `audit append` and candidate/status edits. Require no duplicate identity, lost prose,
   duplicate row, partial write, or stale receipt.
6. Validate the machine contract semantically: optional fields with non-default values, dry-run,
   workspace identity, schema revision metadata/classification, generated Draft 2020-12 schema, and
   all prior envelopes. Confirm the new receipt cannot silently echo the audit body.
7. Run the exact routine commands, including `--body-file -` with a heredoc and a quoted task
   command in `--candidate`. Verify generated help documents the open-audit and legacy-section
   behavior accurately.
8. Verify every refusal wraps the intended domain error and reaches the stable CLI exit code. Confirm
   dry-run, callback failure, parse failure, and exhausted conflict retries never stamp `updated_at`.

For each defect, cite exact paths/lines and a minimal reproduction or mutation result.

## Required hostile probes

- Move allocation outside the retry callback and require a focused concurrency test to fail.
- Remove the snapshot bucket check, or substitute a pre-read check, and require a coordinated
  close/create test to expose the race.
- Make fence masking or section-boundary detection permissive and require focused tests to fail.
- Permit one reserved metadata label through title/body/recommendation and show which parser-owned
  field changes.
- Split finding and candidate creation into two writes, force candidate validation to fail, and
  require atomicity coverage to catch the partial finding.
- Make the JSON receipt include the body or remove its registered schema branch and require wire/
  golden tests to fail.
- Challenge shared helpers and fakes: ensure the same fake cannot accidentally make stale retries,
  bucket races, or persisted dry-runs look safe while the filesystem adapter is wrong.

Restore every mutation and rerun its focused test green.

## Deliverable

Preserve this brief. Add findings under `## Findings`, all initially open. Prefer the implementation
under review to write each one:

```sh
./bin/tskflwctl audit finding new \
  2026-09-20-audit-finding-creation-verb-implementation-antigravity "<title>" \
  --band <H|M|L> --file "<path:line>" --component "<component>" \
  --effort <XS|S|M|L> --urgency <acute|soon|eventually> \
  --body-file - --recommendation "<minimum fix>" <<'EOF'
<evidence, exact reproduction or mutation, and affected contract; no unfenced heading>
EOF
```

If the reviewed writer itself prevents recording a valid finding, manually use the exact canonical
finding grammar and identify that failure as evidence. A substantive no-findings report is acceptable
only after completing the evidence floor and hostile probes. Do not change finding status, create
tasks, edit the implementation, commit, or push.

## Findings

## Reviewer report

### Mandatory isolation attestation

```ini
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/.git
baseline_commit=85d1e2c6897058a10bf45f0a3339e42996838f41
source_blob=fdd2804dbc2b50ca370b3fec9d05bf21e5ae22e4
source_fingerprint=5b89aff74550ba7b2656a1da151908fa125a2618
deliverable=planning/audits/6gbxf5yz94xv-2026-09-20-audit-finding-creation-verb-implementation-antigravity.md
```

### Review verdict

**Substantive No-Findings (Clean Pass)**

The implementation of task `6gbpe6e8n87k` (`add-a-tool-owned-audit-finding-creation-verb`) establishes a robust, tool-owned finding creation verb (`tskflwctl audit finding new`). It provides monotonic per-band code allocation, places findings before trailing `## Candidate tasks` sections, strictly prevents metadata and section injection, guarantees atomic dual finding-and-candidate writes, enforces open-audit bucket state from guarded CAS snapshots, emits compact revision-1.69 JSON receipts without leaking the audit body, and fails closed with byte-identical preservation on any malformed state or legacy prose. All mandatory evidence floor criteria and hostile mutation probes were verified inside the isolated sandbox.

---

### Challenged consumer inventory & symbol verification

Every claimed component, function, and symbol was directly verified inside the isolated sandbox:

- [`internal/domain/finding_create.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/internal/domain/finding_create.go):
  - `FindingDraft`: semantic input struct (`Band`, `Title`, `File`, `Component`, `Effort`, `Urgency`, `Body`, `Recommendation`).
  - `CreateFinding()`: pure domain validator and allocator; checks existing near-misses, existing finding/candidate lint, monotonic per-band code allocation, markdown block rendering, and insertion before candidate tasks.
  - `normalizeFindingDraft()`: validates required fields, single-line constraints, vocabulary enums, and rejects reserved labels (`**Status:**`, `**File:**`, `**Component:**`, `**Effort:**`, `**Urgency:**`, `**Recommendation:**`, `**Resolution:**`) in title, recommendation, and body.
  - `nextFindingCode()`: audit-local monotonic allocation (maximum numeric suffix in band + 1); strictly prohibits gap reuse and detects duplicate codes.
  - `renderFindingBlock()`: formats canonical finding markdown with proper delimiters.
  - `insertFindingBlock()`: fence-aware search for real Findings sections (`blankFences`); creates `## Findings` if missing, places finding before `## Candidate tasks`, and refuses ambiguous multiple Findings sections.
- [`internal/core/finding_create.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/internal/core/finding_create.go):
  - `NewFindingParams` & `FindingCreationReceipt`: input params and compact receipt DTO.
  - `NewFinding()`: executes creation inside the CAS retry loop (`retryOnConflict`), re-evaluating the audit bucket from the guarded store snapshot (`audit.Bucket == AuditOpen`), re-running allocation against fresh body text, and applying `--candidate` in the same transform.
- [`internal/store/body.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/internal/store/body.go):
  - `TransformAuditBody()`: updated to pass the re-parsed `domain.Audit` snapshot to the transform closure alongside the markdown body, enabling snapshot-level bucket verification under lock.
- [`internal/cli/audit.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/internal/cli/audit.go):
  - `newAuditFindingNewCmd()`: Cobra command binding `--band`, `--title`, `--file`, `--component`, `--effort`, `--urgency`, `--body`, `--body-file`, `--recommendation`, `--candidate`, `--dry-run`, and `--json`.
- [`internal/wire/`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/internal/wire):
  - `wire.SchemaVersion = "1.69"`: monotonic additive schema bump.
  - `FindingCreationEnvelope`: compact receipt schema carrying `schema_version`, `dry_run`, `audit`, `finding`, and `workspace` without echoing the full audit body.
- Documentation & routines:
  - [`docs/cli/tskflwctl_audit_finding_new.md`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/docs/cli/tskflwctl_audit_finding_new.md): generated CLI documentation.
  - [`routines/code-quality-audit.md`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/routines/code-quality-audit.md) & [`routines/weekly-architecture-audit.md`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.bvKi3T/routines/weekly-architecture-audit.md): updated to v3 with `audit finding new` commands.

---

### Mandatory evidence floor & toolchain validations

All commands executed strictly within `$SANDBOX`:

1. **Git diff check**:
   ```sh
   git diff --check HEAD
   ```
   *Result:* Clean (exit code 0).

2. **Race-detector test suite**:
   ```sh
   GOCACHE=/tmp/taskflow-review-go-cache go test -race ./...
   ```
   *Result:* Passed (exit code 0, 0 data races).

3. **Linter**:
   ```sh
   GOCACHE=/tmp/taskflow-review-go-cache GOLANGCI_LINT_CACHE=/tmp/taskflow-review-lint-cache golangci-lint run ./...
   ```
   *Result:* Passed (exit code 0, 0 issues).

4. **Internal planning & audit linters**:
   ```sh
   ./bin/tskflwctl --no-color lint
   ./bin/tskflwctl --no-color audit lint
   ```
   *Result:* Clean (exit code 0, all planning entities and audit findings pass lint).

---

### Disposable planning space lifecycle reproduction

Executed the full lifecycle in an isolated temporary repository (`/tmp/tskflw-probe-finding-new-ZiADpa`) using sandbox binary `./bin/tskflwctl`:

1. **Space initialization**: `tskflwctl init --taskflow-root planning --no-register`
2. **Audit creation**: `tskflwctl audit new probe-finding-audit`
3. **High-severity finding creation with all optional fields**:
   ```sh
   tskflwctl audit finding new probe-finding-audit "High severity defect" \
     --band H --file "internal/foo/bar.go:42" --component "core/storage" \
     --effort M --urgency acute \
     --body "Detailed evidence of the defect." \
     --recommendation "Use a narrower lock."
   ```
   *Result:* Created `H1` before `## Candidate tasks`.
4. **Medium-severity finding creation with managed candidate**:
   ```sh
   tskflwctl audit finding new probe-finding-audit "Medium severity issue" \
     --band M --file "internal/api/handler.go:10" --component "api" \
     --effort S --urgency soon \
     --body "Medium issue evidence." \
     --recommendation "Add input sanitization." \
     --candidate "Create task to sanitize handler inputs"
   ```
   *Result:* Created `M1` under `## Findings` and added `- ○ M1 · open — Create task to sanitize handler inputs` under `## Candidate tasks` in one atomic write.
5. **Low-severity finding creation with heredoc stdin body**:
   ```sh
   tskflwctl audit finding new probe-finding-audit "Low severity note" \
     --band L --effort XS --urgency eventually \
     --recommendation "Refactor comments" \
     --body-file - <<'EOF'
   This is a multiline
   body passed via stdin.
   EOF
   ```
   *Result:* Created `L1` properly formatted.
6. **Dry-run preview in JSON**:
   ```sh
   tskflwctl audit finding new probe-finding-audit "Dry run preview" --band H --dry-run --json
   ```
   *Result:* Emitted compact envelope (`schema_version: "1.69"`, `dry_run: true`, `code: "H2"`, no audit body) without writing to disk.
7. **Finding update**:
   ```sh
   tskflwctl audit finding probe-finding-audit H1 --status fixed --note "Narrowed the lock; regression test added."
   ```
   *Result:* Successfully updated finding H1.
8. **Linters**: Both `tskflwctl audit lint probe-finding-audit` and `tskflwctl lint` passed with 0 issues.

---

### Boundary and edge case verification

1. **Candidate failure atomicity**:
   - Tested creating a finding with `--candidate` against an audit with no candidate section: refused with `ErrValidation` (exit code 11), SHA-256 hash matched before and after (byte-identical).
   - Tested creating a finding with `--candidate` against an audit with an unversioned legacy candidate section: refused with `ErrValidation` (exit code 11), SHA-256 hash matched before and after (byte-identical).
2. **Metadata label injection**:
   - Title containing `**Status:**`: refused with `ErrValidation` (exit code 11).
   - Recommendation containing `**File:**`: refused with `ErrValidation` (exit code 11).
   - Body containing `**Effort:**`: refused with `ErrValidation` (exit code 11).
3. **Heading injection in body**:
   - Unfenced `### Heading`: refused with `ErrValidation` (exit code 11).
   - Fenced markdown example containing `### Heading`: permitted and created cleanly without breaking outer section structure.

---

### Hostile angles & mutation probes

1. **Probe 1: Allocation outside retry callback**:
   - *Mutation:* Modified `internal/core/finding_create.go` to compute allocation from a stale pre-read body rather than the fresh transform input.
   - *Test:* `go test -run TestNewFinding_RetriesAllocationAroundConcurrentAppend ./internal/store/...`
   - *Failure output:*
     ```
     --- FAIL: TestNewFinding_RetriesAllocationAroundConcurrentAppend (0.03s)
         transformauditbody_test.go:259: fresh retry should allocate H3, got {Code:H2 Title:Retried finding Status:open ...}
     FAIL
     ```
   - *Restoration:* Reverted change; test passed green.

2. **Probe 2: Snapshot bucket guard**:
   - *Mutation:* Removed `audit.Bucket != domain.AuditOpen` check in `internal/core/finding_create.go`.
   - *Test:* `go test -run TestNewFindingRefusesNonOpenAuditFromGuardedSnapshot ./internal/core/...`
   - *Failure output:*
     ```
     --- FAIL: TestNewFindingRefusesNonOpenAuditFromGuardedSnapshot (0.00s)
         finding_create_test.go:61: closed-audit creation should be refused, got <nil>
     FAIL
     ```
   - *Restoration:* Reverted change; test passed green.

3. **Probe 3: Fence masking & section boundary detection**:
   - *Mutation:* Replaced `blankFences(body)` with raw `body` in `insertFindingBlock` in `internal/domain/finding_create.go`.
   - *Test:* `go test -run TestCreateFindingCreatesMissingFindingsSectionAndIgnoresFencedHeadings ./internal/domain/...`
   - *Failure output:*
     ```
     --- FAIL: TestCreateFindingCreatesMissingFindingsSectionAndIgnoresFencedHeadings (0.00s)
         finding_create_test.go:57: want fenced + real Findings headings, got 1
     FAIL
     ```
   - *Restoration:* Reverted change; test passed green.

4. **Probe 4: Reserved metadata label protection**:
   - *Mutation:* Removed `Effort` from `reservedFindingFieldRe` in `internal/domain/finding_create.go`.
   - *Test:* `go test -run TestCreateFindingRejectsAmbiguousOrUnsafeInput ./internal/domain/...`
   - *Failure output:*
     ```
     --- FAIL: TestCreateFindingRejectsAmbiguousOrUnsafeInput (0.00s)
         --- FAIL: TestCreateFindingRejectsAmbiguousOrUnsafeInput/body_metadata_injection (0.00s)
             finding_create_test.go:93: CreateFinding error = <nil>, want validation containing "reserved finding metadata label"
     FAIL
     ```
   - *Restoration:* Reverted change; test passed green.

5. **Probe 5: Atomic dual-write protection**:
   - *Mutation:* Split finding and candidate writes into two sequential `TransformAuditBody` calls in `internal/core/finding_create.go`.
   - *Test:* `go test -run TestNewFindingCandidateRefusalIsAtomicAndDryRunDoesNotPersist ./internal/core/...`
   - *Failure output:*
     ```
     --- FAIL: TestNewFindingCandidateRefusalIsAtomicAndDryRunDoesNotPersist (0.00s)
         finding_create_test.go:126: failed candidate creation partially changed body:
             ## Findings
             #### M1. Existing · **Status:** open
             #### M2. No partial write · **Status:** open
     FAIL
     ```
   - *Restoration:* Reverted change; test passed green.

6. **Probe 6: Wire contract & body containment**:
   - *Mutation:* Added `Body string `json:"body"` ` to `FindingCreationEnvelope` in `internal/wire/envelopes.go`.
   - *Test:* `go test ./internal/cli/...`
   - *Failure output:* Failed golden checks against `schema_jsonschema.golden` and `schema_json.golden`.
   - *Restoration:* Reverted change; test passed green.

7. **Probe 7: Shared test fake fidelity**:
   - Compared `findingCreationStore` in `internal/core/finding_create_test.go` with `internal/store/transformauditbody_test.go`. Verified that the real filesystem store runs identical CAS retry semantics under flock and atomic file renames.

---

### Systemic analysis (pass 2)

1. **Allocation resilience during concurrent writes**:
   Allocation is calculated strictly inside the transform closure passed to `TransformAuditBody`. If a concurrent write lands between read and write, the compare-and-swap detects content change, rolls back, and invokes `retryOnConflict`. The retried closure inspects the newly committed audit body, discovers the latest allocated codes, and allocates a conflict-free identifier.
2. **Section order invariant**:
   `insertFindingBlock` locates existing real findings sections and ensures any new finding block is placed before `## Candidate tasks`. If an audit lacks `## Findings`, it is created immediately before `## Candidate tasks`, preserving the canonical document topology.
3. **Fail-closed posture**:
   If an audit has near-miss headers or existing malformed findings, `CreateFinding` refuses with remediation instructions (`run lint --fix` or `run audit lint`) rather than allocating into a corrupted document.

---

### Residual risks & operational notes

1. **Shell heredoc quoting**: When invoking `audit finding new --body-file - <<'EOF'`, quoting `'EOF'` is recommended to prevent shell parameter expansion within the body.
2. **Quoting task command suggestions**: When supplying `--candidate`, wrapping the outer argument in single quotes prevents shell expansion of inner double quotes.
