---
schema: 1
id: 6gdbre2abzxz
bucket: closed
area: portable-audit-finding-snapshot-implementation-antigravity
date: "2026-09-24"
updated_at: "2026-09-25"
---
# Audit: Portable audit finding snapshot implementation — antigravity — 2026-09-24

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

Adversarially review the implementation of task `6gcwcf7gjayh`,
`make-finding-queries-consume-a-portable-audit-snapshot`. Do not summarize the diff and stop. Try to
disprove that the new port is actually adapter-neutral, that each source is read once, that selector
semantics are preserved, and that schema revision 1.74 is truthfully additive. Take a second,
systemic pass after the checklist: look for an abstraction that is locally green but likely to fail
when a remote adapter, another audit consumer, or concurrent repository change is introduced.

For Antigravity specifically: do not award confidence from the existing happy-path tests. Rebuild
the read topology independently, use coordinated mutations where one nearby fallback would mask a
defect, and create hostile fixtures that make path-based assumptions produce the wrong answer rather
than merely fail. If an apparent issue is out of scope, still establish whether it is a regression,
pre-existing debt, or a concrete follow-up; do not dilute real findings into general suggestions.

## Review target

Review the complete uncommitted implementation on branch
`refactor/portable-audit-finding-snapshot` relative to base
`7676a433f35428108403ea9d1456f7aa9dfb82c5`. Include source, tests, generated machine-contract
goldens, architecture/ADR amendments, and task lifecycle bookkeeping. The primary planning target is
`planning/tasks/6gcwcf7gjayh-make-finding-queries-consume-a-portable-audit-snapshot.md`.

Inventory and trace at least these seams:

- `core.AuditSnapshot`, `AuditSnapshotSource`, `LintSource`, `Service.auditReads`,
  `WithAuditSnapshotSource`, `WithLintSource`, and `NewService` capability discovery;
- `Service.QueryFindings`, `Service.LintAudits`, and the audit portion of `Service.Lint`;
- `store.FS.ReadAuditSnapshot`, `resolveAudit`, `ListAuditsWithFindings`, scanner problem identity,
  and the removed `GetAuditByPath` surface;
- CLI list-mode generalization, portable problem exit/rendering, full and projected findings JSON;
- `wire.FindingsEnvelope`, `LintLoadProblemJSON`, schema revision 1.74, schema comments, generated
  schema, and all updated machine goldens;
- remaining audit readers such as Summary, finding repair/mutation, show/info/list, and any adapter
  composition root that may still bypass or accidentally depend on the new capability.

Do not implement fixes or edit any file other than this assigned audit.

## Intended contract to challenge

1. An empty selector yields one resilient audit snapshot. Each readable source is opened at most
   once, and its metadata, tally, findings, near misses, and candidate issues derive from those same
   bytes. Unreadable records appear once with stable identity and optional location.
2. A non-empty selector preserves the established exact canonical-ID, exact slug,
   case-insensitive unique prefix, unique substring, validation, not-found, and ambiguity behavior.
   It reads only the chosen audit and is not poisoned by an unrelated malformed audit.
3. `QueryFindings`, repository lint, and single-audit lint consume the narrow snapshot capability.
   A pathless adapter can support them without implementing broad `Store`, `AuditStore`, or local
   path APIs. Typed-nil and option-order cases do not silently leave the wrong capability wired.
4. No core or primary-adapter query follows `Audit.Path` or diagnostic `Location` to load data or
   reconstruct identity. `GetAuditByPath` is absent from Go application ports and implementations,
   not merely unused on the default path.
5. Existing finding filtering, audit/document order, audit ID/slug/bucket attribution, lint issue
   sets, error identity, partial-result exit code, and human stream discipline remain unchanged.
6. Full and caller-projected `audit findings --json` both publish the same portable unreadable
   shape. Local files retain exact `path`; remote/opaque/pathless sources retain identity and
   location without inventing filesystem semantics or leaking Go field names.
7. Schema 1.74 is additive for tolerant readers: existing `path` and `message` stay compatible,
   new fields are optional where promised, schema comments/schema/goldens agree, and no unrelated
   generated contract drift was accepted under the bulk revision update.
8. The refactor does not change audit mutation authorization, content CAS, Summary semantics,
   filesystem repair evidence, finding grammar, or ordinary audit-list contracts.

## Mandatory evidence floor

- Build a producer/consumer inventory for all audit reads and all conversions among
  `domain.FileProblem`, `LintLoadProblem`, and `LintLoadProblemJSON`. State which paths deliberately
  remain local and which are portable.
- Trace the empty-selector and selected-selector flows from CLI/service construction through the
  filesystem adapter and back to human/full JSON/projected JSON output. Count directory scans,
  source opens, and finding parses rather than inferring “single scan” from a method name.
- Use a genuinely narrow pathless fake that does not embed `nopStore` or expose paths. Exercise
  readable findings plus unreadable records with: ID+slug and no location; opaque non-path location;
  misleading ID-shaped location; and no recoverable identity.
- Build filesystem fixtures for exact-ID precedence over a colliding slug, case-insensitive exact
  slug, ambiguous prefix, ambiguous substring, invalid `../` and separator queries, duplicate IDs,
  a malformed selected audit, and an unrelated malformed audit beside a valid selected audit.
- Compare before/after behavior for finding order and filtering, single-audit and repository lint,
  human stdout/stderr, exit code 11 on partial reads, bare JSON, and `--json -c` projection.
- Inspect every changed golden semantically. Independently regenerate schema comments and the full
  machine-contract goldens, then prove schema 1.74 contains exactly the intended findings diagnostic
  change plus the global revision movement.
- Execute and restore at least eight mutation probes: ignore the selector; implement selection by
  filtering an all-audit scan; reread each body after listing; derive findings from different bytes;
  drop unreadable identity; reinterpret Location as a path; bypass `auditReads` through `store`;
  fail to wire `auditReads` from `LintSource`; serialize raw core diagnostics in projected JSON;
  or omit the schema bump. Name the focused test that kills each mutation and report survivors.
- Run `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, planning lint,
  both assigned-audit lint, the complete golden drift check, and `git diff --check`.

## Required hostile angles

- Challenge the selector-bearing port itself. Does combining “all” and “one” behind an empty string
  create ambiguous semantics, adapter burden, or a future pagination/cache trap? Is selector
  resolution correctly owned, documented, and testable across non-filesystem adapters?
- Look for false snapshot coherence: metadata and findings may be returned together while having
  been parsed twice, resolved from one directory state and read from another, or mixed with problem
  evidence from a different scan. Separate harmless parse duplication from observable inconsistency.
- Attack duplicate and malformed identity. Determine whether an unreadable record can disappear
  from ambiguity, whether exact-ID precedence remains authoritative, and whether filename identity
  is still parsed only by the filesystem adapter.
- Attack constructor capability wiring with nil interfaces, `WithLintSource` plus
  `WithAuditSnapshotSource` in both orders, a broad Store lacking the new port, and a narrow source
  with no Store. Look for a panic or a misleading “unavailable” error in a valid composition.
- Check whether generalized list rendering subtly changed every existing list command, especially
  projected JSON ordering/omitempty behavior and stderr routing, even though only findings needed a
  new problem type.
- Check the claimed additive wire change using a strict old reader and a tolerant old reader. Verify
  `path` remains present where the prior contract required it and that opaque locations never fill
  it unless `LocationIsPath` is true.
- Search beyond named files for stale port documentation, test fakes that accidentally provide a
  fallback, direct filesystem calls in core/CLI/TUI, and old assumptions in generated or public docs.
- Ask what breaks when a future web adapter caches snapshots, when a source changes between resolve
  and read, and when audit count is large. Report only issues grounded in current contracts or a
  clearly demonstrated extensibility failure.

## Validation and restoration

Run all probes only in the mandatory sandbox. Mutations must be restored to its captured baseline.
Before transfer, ensure the assigned audit is the only sandbox diff, run the isolation helper's
verification, inspect the audit diff, and transfer through the helper. Do not format, regenerate,
stage, commit, or repair the shared source checkout.

## Deliverable

Preserve this brief. Record a severity-ranked verdict with exact file/symbol evidence, commands,
fixture outcomes, consumer inventory, and mutation table. Add each defect under `## Findings` using
the repository finding grammar and leave every finding open. A substantive no-findings verdict is
acceptable only after all evidence-floor work and the second systemic pass are documented.

## Reviewer report

### Verdict and summary

**Verdict: Accept with findings (3 Medium, 2 Low).**

The implementation of task `6gcwcf7gjayh` (`make-finding-queries-consume-a-portable-audit-snapshot`) successfully eliminates the persistence-shaped `ListAudits` → `GetAuditByPath(a.Path)` reread cycle from `Service.QueryFindings`, removes the `GetAuditByPath` method entirely from application ports and adapters, and introduces the adapter-neutral `AuditSnapshotSource` capability. Single-audit filtering, exact canonical-ID resolution, and resilient multi-audit scans behave correctly under normal operation. Schema revision 1.74 properly extends `FindingsEnvelope.unreadable` to `LintLoadProblemJSON` while retaining the backward-compatible `path` and `message` fields.

However, an adversarial audit across both passes identified five concrete failure modes and test-floor gaps:
1. **Option Clashing (`internal/core/service.go`)**: `WithLintSource` unconditionally assigns `s.auditReads = source`. If a caller explicitly injects a dedicated or remote `AuditSnapshotSource` before `WithLintSource`, that source is silently discarded and replaced by `LintSource`.
2. **Crash on Missing Capability (`internal/core/service.go:650`)**: Unlike `QueryFindings` and `LintAudits` which guard against nil `auditReads`, `Service.Lint()` directly invokes `s.auditReads.ReadAuditSnapshot("")` without checking `!isNilCapability(s.auditReads)`. If `auditReads` is nil, repository lint crashes with a SIGSEGV / nil pointer dereference.
3. **Misleading Error Contract (`internal/core/finding.go:323`)**: When `auditReads` is unavailable, `Service.LintAudits(slug)` reports `"repository lint reads are unavailable from this store"` even when called for a single audit (`slug != ""`), giving a confusing error for a targeted operation.
4. **Wire Branch Verification Gap (`internal/wire/envelopes.go`)**: Mutation probe 6 (unconditionally setting `path = problem.Location` in `ToLintLoadProblemsJSON`) survived the entire test suite because no semantic validator or fixture tests an unreadable record where `Location != ""` and `LocationIsPath == false`.
5. **Read Isolation & Single-Scan Test Gap (`internal/store/lintsource.go`)**: Mutation probe 2 (implementing single-audit selection by filtering an all-audit scan) and mutation probe 3 (rereading audit bodies after listing) both survived because `FS.ReadAuditSnapshot` has no test verifying read counts or asserting that selecting one audit avoids reading unrelated audits.

---

### Isolation attestation

The mandatory isolated-review protocol was executed using `scripts/isolated-review-workspace.sh`:

- **Sandbox Path**: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn`
- **Git Directory**: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/.git` (verified independent clone; no alternates, no worktrees, no shared Git metadata)
- **Baseline Commit**: `9ca48d13a6a590527157398b09d8f137b47af11e` (`chore: capture isolated review baseline`)
- **Captured Source Blob**: `057a63f97336836b7b7a44ff63b0362f07bb9be7`
- **Captured Source Fingerprint**: `391f75dd0c907b4d6aa4365c757a89ea4544517b`
- **Deliverable**: `planning/audits/6gdbre2abzxz-2026-09-24-portable-audit-finding-snapshot-implementation-antigravity.md`
- **Verification Result**: Verified clean sandbox baseline prior to report generation; exactly one unstaged file modification (`AUDIT_REL`); transfer executed via `scripts/isolated-review-workspace.sh transfer`.

---

### Consumer inventory and execution traces

#### Producer/Consumer Inventory

| Operation / Method | Producer / Adapter | Problem Type | Consumer(s) | Portability Status |
| :--- | :--- | :--- | :--- | :--- |
| `ReadAuditSnapshot(selector)` | `store.FS` (`internal/store/lintsource.go`) | `core.LintLoadProblem` | `Service.QueryFindings`, `Service.LintAudits`, `Service.Lint` | **Portable**: returns `core.AuditSnapshot` with metadata, parsed findings, and neutral `LintLoadProblem`. |
| `ListAuditsWithFindings()` | `store.FS` (`internal/store/auditstore.go`) | `domain.FileProblem` | `FS.ReadAuditSnapshot("")`, `Service.Summary`, `Service.FixFindingHeaders` | **Local**: filesystem-specific scan returning exact `domain.FileProblem` paths for local repair and summary tallies. |
| `ListAudits()` | `store.FS` (`internal/store/auditstore.go`) | `domain.FileProblem` | CLI `audit list`, CLI `audit info`, TUI | **Local**: filesystem listing. |
| `GetAudit(slug)` | `store.FS` (`internal/store/auditstore.go`) | `error` | CLI `audit show`, `audit move`, `audit append` | **Local**: slug resolution and body read. |
| `GetAuditByPath(path)` | *Removed* | N/A | *None* | **Removed**: completely deleted from `core.AuditStore` and `store.FS`. |
| `lintLoadProblems` | `internal/store/lintsource.go` | `domain.FileProblem` → `core.LintLoadProblem` | `FS.ReadAuditSnapshot`, `ReadLintTasks`, etc. | **Adapter Boundary**: maps local file problems to portable diagnostics. Sets `LocationIsPath=true`. |
| `ToLintLoadProblemsJSON` | `internal/wire/envelopes.go` | `core.LintLoadProblem` → `wire.LintLoadProblemJSON` | CLI `render.FindingsJSON`, `render.ProjectedListJSONWithProblems`, `ToLintEnvelope`, `ToFixEnvelope` | **Wire Boundary**: emits public schema 1.74 representation. Sets `path = Location` only if `LocationIsPath=true`; emits `path: ""` for opaque locations. |

#### Execution Traces

1. **Unfiltered Findings Query (`tskflwctl audit findings --json`)**:
   - **Call Path**: `cli.newAuditFindingsCmd` → `app.Service.QueryFindings(FindingFilter{})` → `s.auditReads.ReadAuditSnapshot("")` → `store.FS.ListAuditsWithFindings()` → `scanDir(s.auditsDir, ...)`.
   - **Operations**:
     - 1 directory scan (`os.ReadDir(s.auditsDir)`).
     - Exactly $N$ file opens (`os.ReadFile(path)`) for $N$ audit markdown documents.
     - Exactly $N$ finding parses via `parseAuditWithFindings(content, path)`. Findings, near-misses, and candidate issues derive from the exact same byte slice.
     - Unreadable files produce `domain.FileProblem`, translated to `core.LintLoadProblem` with recovered entity ID and slug.
     - Core receives `AuditSnapshot`, filters findings in memory, and returns them alongside `snapshot.Problems`.
     - CLI formats envelope via `wire.ToFindingsEnvelope` (using `ToLintLoadProblemsJSON`), writing versioned JSON. If unreadable records exist, command exits with code 11 (`domain.ErrValidation`).
   - **Counts**: 1 directory scan, $N$ file opens, $N$ body parses. Zero rereads.

2. **Single-Audit Finding Query (`tskflwctl audit findings --audit 2026-06-14-gateway`)**:
   - **Call Path**: `cli.newAuditFindingsCmd` → `app.Service.QueryFindings(FindingFilter{Audit: "2026-06-14-gateway"})` → `s.auditReads.ReadAuditSnapshot("2026-06-14-gateway")` → `store.FS.resolveAudit(...)` → `os.ReadFile(path)` → `parseAuditWithFindings(content, path)`.
   - **Operations**:
     - 1 directory scan of `s.auditsDir` (listing filenames only via `flatCandidates`).
     - Slug/ID resolution via `resolveID`: exact ID tier > exact slug tier > prefix tier > substring tier.
     - Exactly 1 file open (`os.ReadFile(path)` on the resolved path).
     - Exactly 1 finding parse via `parseAuditWithFindings`.
     - Returns `AuditSnapshot{Audits: [1], Problems: nil}`. Unrelated files (even if malformed) are never opened or parsed.
   - **Counts**: 1 directory listing (names only), 1 file open, 1 body parse. Zero unrelated file reads.

---

### Hostile fixtures and mutation results

#### Pathless Fake Scenarios

Exercised with a pathless fake implementing only `AuditSnapshotSource` without embedding `Store` or exposing filesystem paths:
- **Scenario 1 (ID + slug, empty location)**: Findings returned with exact attribution (`AuditID="6g0000000001"`, `Audit="audit-one"`). Wire output emitted `entity_kind: "audit"`, `entity_id: "6g0000000001"`, `entity_slug: "audit-one"`, `location: ""`, `path: ""`.
- **Scenario 2 (Opaque non-path location `s3://bucket/audits/audit-two.md`)**: Wire output emitted `location: "s3://bucket/audits/audit-two.md"` and `path: ""` (did not manufacture a local path).
- **Scenario 3 (Misleading ID-shaped location `Location: "6g0000000099"`, empty EntityID)**: Preserved `entity_id: ""` and `location: "6g0000000099"`. Human renderer output `! unidentified audit record` with `location: 6g0000000099`.
- **Scenario 4 (No recoverable identity, empty location)**: Emitted `entity_kind: "audit"`, empty ID/slug/location/path, and `! unidentified audit record`.

#### Filesystem Hostile Fixtures

All 8 requested filesystem scenarios were tested directly against `store.FS.ReadAuditSnapshot`:
1. **Exact-ID precedence over colliding slug**: Audit A (`ID: 6g0000000001`, `slug: colliding`) and Audit B (`ID: 6g0000000002`, `slug: 6g0000000001`). Query `6g0000000001` resolved uniquely to Audit A without ambiguity. (PASS)
2. **Case-insensitive exact slug**: Audit with slug `colliding` resolved when queried with uppercase `COLLIDING`. (PASS)
3. **Ambiguous prefix**: Audits `prefix-one` and `prefix-two` queried with `prefix-` returned `domain.ErrAmbiguous`. (PASS)
4. **Ambiguous substring**: Audits `prefix-one` and `prefix-two` queried with `prefix` returned `domain.ErrAmbiguous`. (PASS)
5. **Invalid `../` and path separator queries**: Queries `../audit`, `audits/one`, `foo\bar`, and `..` were rejected with `domain.ErrValidation`. (PASS)
6. **Duplicate IDs**: Two files declaring identical stable ID `6g0000000099` returned `domain.ErrAmbiguous` on single resolution and were both surfaced in repository lint duplicate identity checks. (PASS)
7. **Malformed selected audit**: Selected audit with corrupted YAML frontmatter returned an explicit parse error (`errBadFrontmatter`) rather than being silently dropped. (PASS)
8. **Unrelated malformed audit beside valid selected audit**: Malformed audit `6g0000000077-malformed.md` did not prevent successful resolution and parsing of sibling `6g0000000088-valid.md`. (PASS)

#### Mutation Probes Table

| # | Mutation Description | Target File & Seam | Focused Test Expected to Kill | Outcome & Observations |
| :- | :--- | :--- | :--- | :--- |
| 1 | Ignore selector in `ReadAuditSnapshot` (always return all audits) | `internal/store/lintsource.go:27` | `internal/store/lintsource_test.go:TestFSAuditSnapshotPreservesSingleAuditResolutionSemantics` | **Killed**: `single audit snapshot want 1, got 2`. |
| 2 | Implement selection by filtering an all-audit scan (`ListAuditsWithFindings`) | `internal/store/lintsource.go:27` | None (Test gap) | **SURVIVED**: All store and core tests passed. No test verifies that single-audit reads avoid scanning the whole directory. |
| 3 | Reread each body after listing in `ReadAuditSnapshot("")` | `internal/store/lintsource.go:46` | None (Test gap) | **SURVIVED**: All tests passed. Single-scan assertion exists only for `Service.Summary` (`countingAuditStore`), not `store.FS`. |
| 4 | Derive findings from different bytes (parse `""` instead of `bodyText`) | `internal/store/auditstore.go:220` | `internal/cli/integration_golden_test.go:TestGolden_MachineContract/audit_findings_json` | **Killed**: Drift against golden output; findings array was empty. |
| 5 | Drop unreadable identity (`entityID = ""; entitySlug = ""`) | `internal/store/lintsource.go:60` | `internal/cli/audit_test.go:TestAuditFindingsJSONReportsIdentityAwareUnreadableAudits` | **Killed**: Unreadable audit missing expected `entity_id` and `entity_slug`. |
| 6 | Reinterpret `Location` as a path unconditionally (`path = problem.Location`) | `internal/wire/envelopes.go:958` | None (Test gap) | **SURVIVED**: All tests passed because existing tests only provide empty locations or local filesystem paths with `LocationIsPath=true`. |
| 7 | Bypass `auditReads` through `store` in `Service.QueryFindings` | `internal/core/finding.go:158` | `internal/core/finding_test.go:TestQueryFindings_PathlessSnapshotPreservesIdentityAndDiagnostics` | **Killed**: SIGSEGV panic dereferencing nil `s.store`. |
| 8 | Omit wiring `s.auditReads = source` in `WithLintSource` | `internal/core/service.go:67` | `internal/core/lint_source_test.go:TestLintPreservesPortableLoadProblemIdentityWithoutLocations` | **Killed**: SIGSEGV panic on line 650 in `Service.Lint()` dereferencing nil `s.auditReads`. |
| 9 | Serialize raw core diagnostics in projected findings JSON | `internal/cli/audit.go:180` | `internal/cli/audit_test.go:TestAuditFindingsJSONReportsIdentityAwareUnreadableAudits` | **Killed**: Field name capitalization mismatch (`EntityKind` vs `entity_kind`). |
| 10 | Omit schema bump (leave `SchemaVersion = "1.73"`) | `internal/wire/wire.go:311` | `internal/cli/golden_test.go:TestMachineGoldenRevisionMatchesSchemaVersion` | **Killed**: Contract revision mismatch vs machine goldens. |

---

### Validation

All validation checks were executed strictly within the isolated review workspace:

- `go test -race ./...`: **PASS** across all packages (32 packages tested, zero data races).
- `golangci-lint run ./...`: **PASS** (0 issues detected).
- `go mod tidy -diff`: **PASS** (no module drift).
- `go run ./cmd/tskflwctl --no-color lint`: **PASS** (`✔ all planning entities and dependency links pass lint`).
- `go run ./cmd/tskflwctl --no-color audit lint 6gdbre2abzxz`: **PASS** (`✔ all audit findings pass lint`).
- `go run ./cmd/tskflwctl --no-color audit lint 6gdbret88rq7`: **PASS** (`✔ all audit findings pass lint`).
- `go test -v ./internal/cli -run "(Golden|Machine)"`: **PASS** (all 40+ machine contract golden snapshots match schema 1.74).
- `go run ./internal/tools/schemacomments`: **PASS** (verified zero drift against `internal/wire/schema_comments.json`).
- `git diff --check`: **PASS** (zero trailing whitespace or merge conflict markers).

---

## Findings

#### M1. WithLintSource unconditionally clobbers custom WithAuditSnapshotSource · **Status:** fixed

In [internal/core/service.go](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/internal/core/service.go#L63-L70), `WithLintSource` unconditionally overwrites `s.auditReads`:

```go
func WithLintSource(source LintSource) Option {
	return func(s *Service) {
		if !isNilCapability(source) {
			s.lintReads = source
			s.auditReads = source
		}
	}
}
```

If a caller provides a dedicated or remote `AuditSnapshotSource` via `WithAuditSnapshotSource` and also supplies `WithLintSource` for the other planning entities (tasks, epics, research), passing `WithAuditSnapshotSource` before `WithLintSource` results in `WithLintSource` silently overriding `s.auditReads` with `source`. The caller's explicitly configured audit capability is discarded without error or warning. `WithLintSource` should only set `s.auditReads` if `s.auditReads == nil` so that option order does not silently discard an explicit capability.

**Resolution:** Explicit audit snapshot injection now has
option-order-independent precedence over the broader lint source, matching
existing narrow-port configuration; both orders and the lint-source default are
regression-tested.

#### M2. Service.Lint panics on nil auditReads without capability validation · **Status:** fixed

In [internal/core/service.go](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/internal/core/service.go#L650), `Service.Lint()` invokes `ReadAuditSnapshot` directly on `s.auditReads`:

```go
	auditSnapshot, err := s.auditReads.ReadAuditSnapshot("")
	if err != nil {
		return nil, nil, err
	}
```

While `QueryFindings` ([finding.go:155](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/internal/core/finding.go#L155)) and `LintAudits` ([finding.go:323](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/internal/core/finding.go#L323)) check `if isNilCapability(s.auditReads)`, `Service.Lint()` contains no nil check before dereferencing `s.auditReads`. As demonstrated in mutation probe 8, if `s.auditReads` is nil, running `Service.Lint()` triggers a SIGSEGV / nil pointer dereference panic rather than returning a formatted capability error. `Service.Lint()` must check `if isNilCapability(s.auditReads)` before dereferencing.

**Resolution:** Repository lint validates the audit snapshot capability before
any entity read and returns a typed capability error instead of risking a nil
dereference; a regression test pins the fail-closed order.

#### M3. Service.LintAudits returns misleading repository lint error for single-audit queries · **Status:** fixed

In [internal/core/finding.go](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/internal/core/finding.go#L323-L325), `Service.LintAudits` handles missing capability with:

```go
	if isNilCapability(s.auditReads) {
		return nil, nil, fmt.Errorf("repository lint reads are unavailable from this store")
	}
```

When a user runs a targeted single-audit lint command (`tskflwctl audit lint <slug>`), a missing capability reports that `"repository lint reads are unavailable"`. Repository lint reads pertain to multi-entity scans across tasks, epics, and research (`LintSource`), whereas `audit lint <slug>` requested only an audit snapshot. The error message should accurately describe the missing capability as `"audit snapshot reads are unavailable from this store"`, matching `QueryFindings`.

**Resolution:** QueryFindings and LintAudits now report the missing audit
snapshot capability directly; focused tests pin the consumer-facing error.

#### L1. Lack of non-path location test coverage allows Location-to-Path wire leak to survive · **Status:** fixed

In [internal/wire/envelopes.go](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/internal/wire/envelopes.go#L955-L965), `ToLintLoadProblemsJSON` gates setting `path = problem.Location` on `problem.LocationIsPath`:

```go
		path := ""
		if problem.LocationIsPath {
			path = problem.Location
		}
```

Mutation probe 6 mutated this to `path = problem.Location` unconditionally, and the entire test suite (`go test ./...`) continued to pass. Existing tests (such as `TestFindingsJSON_PreservesPortableUnreadableIdentity`) provide mock diagnostics where `Location == ""`, while all filesystem tests set `LocationIsPath: true`. No test asserts that a diagnostic with a non-empty, non-path location (`Location != ""` and `LocationIsPath == false`, such as an opaque URI or remote identifier) emits `path: ""` on the wire. A test fixture exercising non-path locations should be added to prevent regression.

**Resolution:** Opaque non-path locations are covered through core passthrough,
the wire mapper, full findings JSON, projected findings JSON, and semantic
schema validation; mutation probes confirm URI-to-path leakage fails.

#### L2. Store FS.ReadAuditSnapshot lacks tests pinning single file open and read isolation · **Status:** fixed

In [internal/store/lintsource.go](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.xKvnEn/internal/store/lintsource.go#L25-L48), `FS.ReadAuditSnapshot` implements single-audit selection via `resolveAudit` and reads only the resolved file. However:
- Mutation probe 2 (implementing single-audit selection by filtering `ListAuditsWithFindings()`) survived all tests.
- Mutation probe 3 (rereading each audit body from disk after listing) survived all tests.

There are currently no store-level tests verifying that `FS.ReadAuditSnapshot(selector)` opens only the selected file and avoids reading unrelated audits, nor that `FS.ReadAuditSnapshot("")` reads each audit file at most once. (Existing read-counting tests only exist on `Service.Summary` with a test double `countingAuditStore` in `internal/core/usecases_test.go`). Tests asserting single-file opens on `store.FS` should be added to guard against regressions back to full directory sweeps or secondary rereads.

**Resolution:** The filesystem adapter has an internal audit-read seam and store
tests now count exact source opens: selected reads open only the resolved audit
and unfiltered reads open every source exactly once. Aggregate-store fallback
rereads are independently guarded in core.
