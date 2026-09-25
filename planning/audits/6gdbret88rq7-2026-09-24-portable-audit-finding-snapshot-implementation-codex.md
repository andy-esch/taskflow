---
schema: 1
id: 6gdbret88rq7
bucket: closed
area: portable-audit-finding-snapshot-implementation-codex
date: "2026-09-24"
updated_at: "2026-09-25"
---
# Audit: Portable audit finding snapshot implementation — codex — 2026-09-24

> Reviewer assignment: codex. This document is the review brief and the only file the reviewer should update.
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

Perform an independent adversarial implementation review of task `6gcwcf7gjayh`,
`make-finding-queries-consume-a-portable-audit-snapshot`. Treat the task's checked acceptance
criteria as claims, not evidence. Reconstruct the control flow and data contracts from source,
then try to make the implementation return a plausible but wrong result. A passing suite is only a
starting point; identify what each regression test actually proves and what it leaves unguarded.

For this Codex review, prioritize precise execution traces, semantic before/after comparisons, and
small runnable counterexamples. Then take a separate architecture pass: assume a second adapter is
HTTP- or database-backed, has no paths, and may offer snapshot caching. Challenge every interface
or helper that makes that adapter inherit filesystem behavior or inconsistent selector rules.

## Review target

Review the complete uncommitted implementation on branch
`refactor/portable-audit-finding-snapshot` relative to base
`7676a433f35428108403ea9d1456f7aa9dfb82c5`. Include source, tests, generated machine-contract
goldens, architecture/ADR amendments, and task lifecycle bookkeeping. The primary planning target is
`planning/tasks/6gcwcf7gjayh-make-finding-queries-consume-a-portable-audit-snapshot.md`.

Inventory and trace at least these seams:

- `core.AuditSnapshot`, `AuditSnapshotSource`, `LintSource`, `Service.auditReads`, options, and
  constructor capability discovery;
- `Service.QueryFindings`, `Service.LintAudits`, and repository `Service.Lint`;
- `store.FS.ReadAuditSnapshot`, `resolveAudit`, `ListAuditsWithFindings`, scanner diagnostics, and
  removal of `GetAuditByPath`;
- list-mode rendering, portable partial-read errors, human streams, bare JSON, projected JSON;
- `FindingsEnvelope`, diagnostic DTO mapping, schema 1.74, generated schema/comments/goldens;
- Summary, audit list/show/info, finding repair/mutation, TUI consumers, and remaining local audit
  read contracts that might conflict with or accidentally bypass the new abstraction.

Do not implement fixes or edit any file other than this assigned audit.

## Intended contract to challenge

1. Empty-selector reads produce one coherent resilient snapshot: readable audit metadata and every
   body-derived projection come from the same source read, while unreadable records retain portable
   identity and optional location.
2. Selected reads preserve exact-ID authority, exact and case-insensitive slug behavior, unique
   prefix/substring resolution, validation, not-found, ambiguity, and the previous behavior when the
   selected audit itself is malformed. Unrelated malformed audits do not affect the selected read.
3. Finding queries and audit lint depend only on the narrow consumer-owned port. A pathless adapter
   can support them without broad Store or filesystem methods, and core never parses location.
4. The unfiltered path opens each audit source at most once and has no hidden list-then-reread,
   resolver rescan per record, fallback `GetAudit`, or second adapter call.
5. Filtering, deterministic audit/document order, attribution fields, lint issue sets, error
   categories, partial-read exits, and stdout/stderr discipline are unchanged.
6. Full and projected findings JSON deliberately map portable diagnostics. Local paths remain
   compatible, opaque locations remain locations, and no internal fields or Go casing leak.
7. Schema 1.74 is correctly classified additive, fully regenerated, and free of accidental drift
   concealed by the global golden revision update.
8. Removed path APIs have no remaining production, test-fake, composition-root, or documentation
   dependency that would force future adapters to recreate them.

## Mandatory evidence floor

- Produce a complete consumer inventory of audit reads and problem conversions. Mark each consumer
  as portable snapshot, deliberate local filesystem contract, mutation/CAS contract, or unresolved
  leakage.
- Trace exact call counts and bytes for one unfiltered query, one selected query, single-audit lint,
  repository lint, and Summary. Do not conflate one adapter method call with one source open.
- Create a minimal pathless adapter implementing only `AuditSnapshotSource`; prove query and lint
  operation with identity-bearing problems whose locations are empty, opaque, and misleading.
- Exercise filesystem reference resolution with exact ID versus colliding slug, case variants,
  unique/ambiguous prefix and substring, duplicate IDs, invalid separators/dot-dot, selected corrupt
  audit, and valid selected audit beside unrelated corruption. Compare error identity and message
  usefulness to the base implementation.
- Exercise finding filters and ordering with multiple audits/buckets, no findings, near-miss
  headings, candidate issues, duplicate slugs/IDs where representable, and a source that returns a
  one-record snapshot inconsistent with the requested selector. Determine whether core needs to
  validate that adapter postcondition or whether the trust boundary is explicit and safe.
- Test all service-wiring combinations: default FS discovery, narrow-only source, typed nil,
  `WithLintSource`, explicit audit override before and after the lint option, and a Store that lacks
  the new capability.
- Inspect regenerated goldens semantically and independently regenerate them. Validate a local
  unreadable audit and a pathless/opaque diagnostic through bare and projected JSON with both strict
  and tolerant decoders.
- Execute and restore at least eight mutations, including selector ignored, all-scan filtered in
  core, per-audit reread restored, problems discarded, identity reconstructed from location,
  `WithLintSource` not wiring audit reads, projected JSON serializing core structs directly,
  `LocationIsPath` ignored, and schema bump/changelog omitted. Record which focused test kills each.
- Run `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, planning lint, audit
  lint, complete golden/schema drift checks, and `git diff --check`.

## Required hostile angles

- Determine whether `ReadAuditSnapshot(selector string)` has one crisp contract or overloads two
  operations in a way that makes consistency, caching, pagination, or authorization error-prone.
  Compare against separate list/get ports, but require a demonstrated defect before calling the
  chosen shape wrong.
- Verify snapshot coherence at the byte level. Look for duplicated parsing that is harmless versus
  rereading that permits mixed revisions; inspect resolver enumeration followed by file reads and
  classify any race according to existing read-only semantics.
- Challenge adapter trust: can a source ignore the selector, return multiple selected records,
  return a mismatched audit, or attach non-audit problems? Decide whether core validation is required
  by current architecture, and back any finding with a plausible adapter or existing invariant.
- Attack typed-nil and option precedence. A narrow capability should not be shadowed by an aggregate
  Store or silently overwritten by an unrelated lint source unless option order intentionally says
  so and tests/documentation make that discoverable.
- Review the generic `renderListWithProblems` extraction as its own blast radius. Check every output
  mode, nil/empty problems, selected columns, wire ordering, and all pre-existing list callers.
- Confirm the new wire shape is really additive: `path` and `message` retention, `omitempty`
  behavior, JSON Schema required fields, strict old decoders, and projection/full parity.
- Search for dead/stale path-shaped APIs and documentation, but do not mislabel historical audit or
  completed-task narrative as active guidance. Focus on instructions a maintainer would still use.
- Review test independence. Identify fakes that reuse production helpers or broad embedded stores in
  a way that lets the implementation and test share the same bug.

## Validation and restoration

Run all probes only in the mandatory sandbox. Restore every mutation and generated artifact to the
captured baseline. Before transfer, ensure the assigned audit is the only sandbox diff, run the
isolation helper verification, inspect the audit diff, and transfer through the helper. Never write
to the shared source checkout except through that guarded one-file transfer.

## Deliverable

Preserve this brief. Add a severity-ranked reviewer report with exact path/symbol references,
consumer inventory, call traces, fixture matrix, mutation table, and command results. Add defects
under `## Findings` using exact repository grammar and leave statuses open for owner triage. If no
finding survives, document the hostile evidence that settled each major risk; “tests pass” is not a
sufficient verdict.

## Reviewer report

### Verdict and summary

The implementation's present behavior is sound, but the change is **not ready to close as
fully regression-proven**. Both review passes found the production control flow coherent: finding
queries and audit lint consume the narrow snapshot port; filesystem snapshots derive metadata,
findings, near misses, candidate issues, and problems from one source read; selected reads retain the
base resolver's authority and error behavior; portable diagnostics survive every output mode; and
mutations continue through their separate local CAS contracts.

Two evidence defects remain open. First, the task claims tests prove the absence of a per-record
reread, but a coordinated `GetAudit(record.Audit.Slug)` reread on aggregate stores passes the complete
core suite. Second, committed fixtures never combine a non-empty opaque location with
`LocationIsPath=false`; both filesystem-path fabrication in the wire mapper and conditional identity
recovery from location pass the relevant shipped suites. The review-only hostile tests killed both
mutations, proving the fixes are small and testable, but those tests are not part of the implementation
being reviewed.

Pass one reconstructed the data flow, contracts, resolver behavior, outputs, generated artifacts, and
consumer inventory. Pass two separately challenged the overloaded list/get port, adapter trust,
option precedence, cache/auth implications, scan coherence, generic rendering, mutation separation,
and test-helper independence. No additional runtime defect survived that second pass.

### Isolation attestation

- Workspace: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.kHzrXQ`
- Resolved Git directory:
  `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.kHzrXQ/.git`
- Sandbox baseline commit: `de47ac5facaf1f164d52e4015812d8e393264162`
- Captured source blob: `0ffe1f14f7059ec082a1f0567464790e31bcd00d`
- Captured source fingerprint: `391f75dd0c907b4d6aa4365c757a89ea4544517b`
- Deliverable:
  `planning/audits/6gdbret88rq7-2026-09-24-portable-audit-finding-snapshot-implementation-codex.md`
- Independence: `git rev-parse --absolute-git-dir` resolved inside the sandbox; the source checkout
  was treated as read-only after helper creation. No worktree, symlinked Git metadata, staging,
  branch switch, stash, reset, clean, or reviewer commit was used.
- Transfer result: successful through
  `scripts/isolated-review-workspace.sh transfer --sandbox <workspace>` after the helper's guarded
  verification; only this assigned audit was copied by the helper. The sandbox remains in place.

### Consumer inventory and execution traces

#### Audit-read and problem-conversion inventory

| Consumer or boundary | Classification | Verified behavior |
| --- | --- | --- |
| `core.Service.QueryFindings` (`internal/core/finding.go:154`) | Portable snapshot | Calls `ReadAuditSnapshot(f.Audit)` once, filters supplied findings without reading `Audit.Path`, and returns supplied `LintLoadProblem`s unchanged (`finding.go:158-175`). |
| `core.Service.LintAudits` (`internal/core/finding.go:318`) | Portable snapshot | Calls the same port once with the requested selector and consumes all body-derived fields from each record (`finding.go:326-337`). |
| Repository `core.Service.Lint` (`internal/core/service.go:483`) | Portable snapshot | Its audit phase calls `ReadAuditSnapshot("")` once, includes readable and unreadable identity in duplicate-ID analysis, and never parses location (`service.go:646-675`). |
| `store.FS.ReadAuditSnapshot` (`internal/store/lintsource.go:25`) | Portable adapter boundary | Selected mode enumerates filename candidates and opens only the chosen source; empty mode delegates to the one-pass body-aware scan. `lintLoadProblems` is the sole local `FileProblem` to neutral diagnostic conversion (`lintsource.go:57-72`). |
| `core.Service.Summary` (`internal/core/service.go:354`) | Unresolved leakage, explicitly outside this task | Still depends on `SummaryStore.ListAuditsWithFindings` and returns `domain.FileProblem`; it is coherent and one-pass, but is not a pathless summary port (`service.go:358-399`, `store.go:312-319`). |
| `core.Service.ListAudits`, CLI/TUI audit lists, fill helpers | Unresolved leakage, explicitly outside this task | Ordinary audit list remains the local `AuditStore.ListAudits`/`domain.FileProblem` contract (`service_audit.go:69-91`; `cli/audit.go:127`; `tui/commands.go:173`; `cli/fill.go:203,233`). |
| CLI completion (`internal/cli/completion.go:259`) | Deliberate local filesystem contract | Shell completion constructs `store.FS` directly and lists local audit names; it is navigation convenience, not a semantic finding/lint port. |
| Audit show/info and TUI detail (`internal/core/service_audit.go:94`) | Unresolved aggregate-store read | `ShowAudit` still uses broad `Store.GetAudit`; CLI show/info and TUI detail consume it (`cli/audit.go:477,524`; `tui/commands.go:193`). It does not participate in finding queries or audit lint. |
| `core.Service.AuditPath` (`internal/core/service_audit.go:99`) | Deliberate local filesystem contract | Explicit parse-free local navigation through `ResolveAuditPath`; remote adapters need not implement it merely to support the new snapshot port. |
| Finding create/edit, header repair, append, and bucket moves | Mutation/CAS contract | Writes retain `TransformAuditBody`, `AppendAuditBody`, and `MoveAudit`; `FixFindingHeaders` uses its initial local scan only to select candidates and recomputes under the guarded transform (`finding.go:232-247,277-309`; `finding_create.go:45`; `service_audit.go:105-135`). Read snapshots never authorize writes. |
| Human, bare JSON, projected JSON, and partial-read exit | Portable problem conversion | `renderListWithProblems` keeps problems on stderr for non-JSON modes and embeds them for JSON (`cli/listmode.go:169-218`); `ToLintLoadProblemsJSON` preserves identity/location and gates compatibility `path` on `LocationIsPath` (`wire/envelopes.go:952-968`); `portableProblemsError` uses identity before location (`cli/problems.go:74-103`). |

`GetAuditByPath` has no active production, test-fake, composition-root, or maintainer-guidance
dependency. The only remaining mentions are the accepted ADR/task narrative and historical task
`6fes83r01325`; those correctly describe its removal/history rather than instructing new code to use
it.

#### Exact calls, source opens, and bytes

A review-only read hook was placed immediately after the two production `os.ReadFile` sites, then
removed. The fixture contained a 102-byte readable alpha audit, a 68-byte readable beta audit, and a
39-byte malformed audit: 209 bytes total. A separate port spy distinguished adapter calls from body
opens.

| Operation | Adapter or store calls | Audit directory enumeration | Audit source opens and bytes |
| --- | --- | --- | --- |
| Unfiltered `QueryFindings` | one `ReadAuditSnapshot("")` | one | 3 opens, 209 bytes total |
| Selected `QueryFindings("alpha")` | one `ReadAuditSnapshot("alpha")` | one candidate enumeration | 1 open, 102 bytes |
| `LintAudits("alpha")` | one `ReadAuditSnapshot("alpha")` | one candidate enumeration | 1 open, 102 bytes |
| Repository `Lint` | one each of task, epic, research, and audit lint reads; audit selector `""` | one audit enumeration | 3 audit opens, 209 bytes |
| `Summary` | zero `AuditSnapshotSource` calls; one `ListAuditsWithFindings` call | one | 3 audit opens, 209 bytes |

The unfiltered byte path is `scanDirSources` (`internal/store/resolve.go:32-76`): one `ReadFile`
produces the byte slice passed to `parseAuditWithFindings`. Selected mode is
`internal/store/lintsource.go:29-44`: resolution opens no body, and one `ReadFile` feeds every
projection. Thus metadata and body projections are byte-coherent even though the repository-wide
scan remains the existing non-transactional read-only view; a concurrent replacement can affect
which revision is opened, but cannot mix metadata from one byte slice with findings from another.

### Hostile fixtures and mutation results

#### Fixture and architecture matrix

- A minimal adapter implementing only `AuditSnapshotSource` returned manually constructed findings
  with no filesystem methods. Queries and selected audit lint preserved audit ID, slug, bucket,
  document order, filtering, and three identity-bearing problems whose locations were empty,
  `db://audit/5`, and path-looking-but-nonpath. Core did not inspect the locations.
- Multiple buckets, an audit with no findings, near-miss headings, candidate issues, duplicate IDs,
  and duplicate slugs retained supplied audit/document order. Repository lint still attributed
  duplicate IDs to both records.
- Filesystem selection was compared directly with the base-equivalent `GetAudit` path for exact ID
  versus a colliding slug, case-insensitive exact slug, unique prefix, unique substring, ambiguous
  prefix, ambiguous substring, duplicate exact IDs, separators, dot-dot, not-found, selected corrupt
  audit, and a valid selected audit beside unrelated corruption. Successful identity and every error
  sentinel/message matched; unrelated corruption was never opened.
- All six list outputs were exercised: human, name, table, CSV, bare JSON, and projected JSON.
  Non-JSON problems stayed on stderr. Local diagnostics retained `path`; opaque and misleading
  locations retained `location` and emitted an empty `path`. Current strict decoding passed,
  tolerant old decoding retained `path`/`message`, and strict old decoding rejected the documented
  additive fields, consistent with the `ignore-unknown-object-fields` reader policy.
- The real JSON Schema semantic validator passed a `FindingsEnvelope` populated with local, opaque,
  and misleading non-default diagnostics. Independent full golden regeneration and generated schema
  comments produced no drift beyond the committed implementation.
- Both option orders were executed. Options are applied sequentially (`service.go:263-265`):
  `WithLintSource` legitimately supplies the embedded audit capability (`lint_source.go:36-41`), and
  a later `WithAuditSnapshotSource` wins. Default FS discovery, narrow-only injection, typed nils,
  and a broad `Store` without the capability behaved as documented. This is an explicit overlapping
  capability, not an unrelated source shadow.
- A hostile selected adapter returned two mismatched records and core accepted them. That is not a
  finding: `AuditSnapshotSource` explicitly promises one ordinarily resolved record
  (`audit_source.go:13-21`), and core cannot validate adapter-specific aliases, authorization, or
  resolver policy without recreating the abstraction it removed. Cache keys and authorization must
  include the selector inside the adapter. The current FS implementation satisfies the postcondition.
- The one repair command emitted by the hostile near-miss fixture, `lint --fix`, was executed against
  that recommending state; it canonicalized the heading and the subsequent selected `audit lint`
  was clean.

#### Mutation table

| Mutation | Result and killing evidence |
| --- | --- |
| Ignore `f.Audit` and call the empty selector | Killed by `TestQueryFindings_DelegatesSingleAuditResolutionToSnapshot`; the selected result became empty and the call spy observed `""`. |
| Read the all-snapshot and filter inside core | Killed by the same selected-port test and call spy, including the coordinated local filter that would otherwise preserve an exact-slug result. |
| Restore a per-record aggregate `GetAudit(record.Audit.Slug)` reread | Reviewer counting fake killed it with two extra calls. The shipped `go test ./internal/core` passed unchanged: finding M1. |
| Discard `snapshot.Problems` | Killed by `TestQueryFindings_PathlessSnapshotPreservesIdentityAndDiagnostics` and the hostile three-location fixture. |
| Reconstruct identity from `Location` | Overwriting explicit identity was killed by the hostile opaque/misleading fixture. A conditional reconstruction only when identity is absent passed shipped core and CLI suites: finding L1. |
| Stop `WithLintSource` from wiring audit reads | Killed by `TestLintPreservesPortableLoadProblemIdentityWithoutLocations` with a nil audit capability panic. |
| Serialize core problems directly in projected JSON | Killed by `TestAuditFindingsJSONReportsIdentityAwareUnreadableAudits`; fields became Go-cased and the projected identity/path contract disappeared. |
| Ignore `LocationIsPath` and assign `path = location` | Reviewer full/projected JSON fixture killed it. Shipped wire, render, and CLI suites all passed: finding L1. |
| Omit both schema 1.74 and its changelog entry | Killed by `TestMachineGoldenRevisionMatchesSchemaVersion`: committed machine goldens declared 1.74 while runtime declared 1.73. |

The generic renderer's blast radius was also exercised with nil/non-empty problem slices, every
output mode, selected columns, and both wire/human callbacks. No pre-existing list caller changed
shape or stream discipline. The mutation/CAS review found no snapshot-derived authorization: all
actions reread and validate through their existing write contracts, so the read abstraction does
not create a projection-to-action time-of-check/time-of-use hole.

### Validation

- `go test -race ./...` — passed all packages.
- `golangci-lint run ./...` — `0 issues.`
- `go mod tidy -diff` — clean, no output.
- `go test ./internal/cli -update` — passed; independent complete machine-golden regeneration left
  no diff.
- `go run ./internal/tools/schemacomments -out /tmp/taskflow-review-schema-comments.json` plus
  `diff -u internal/wire/schema_comments.json ...` — generated 269 comments; no diff.
- Review-only pathless/resolver/output probes — passed before mutation; every required mutation was
  then observed to fail its reviewer guard and was manually reversed.
- Review-only populated `TestJSONSchema_ValidatesRealOutput` — passed for the non-default findings
  unreadable branch, then was reversed.
- Review-only emitted-repair probe — `lint --fix` executed successfully and removed the recommending
  near-miss condition, then was removed.
- `go run ./cmd/tskflwctl -C . lint` — `✔ all planning entities and dependency links pass lint`.
- `go run ./cmd/tskflwctl -C . audit lint 2026-09-24-portable-audit-finding-snapshot-implementation-codex`
  — clean before this report deliberately added open findings.
- `git diff --check` — clean.
- `scripts/isolated-review-workspace.sh verify --sandbox <workspace>` — passed after all probes,
  hooks, generated artifacts, and mutations were removed; the assigned audit was the only diff.

## Findings

#### M1. The claimed no-reread regression guard cannot observe an aggregate-store reread · **Status:** fixed

The implementation is correct at `internal/core/finding.go:158-175`, but the acceptance claim in
`planning/tasks/6gcwcf7gjayh-make-finding-queries-consume-a-portable-audit-snapshot.md:45` says store
and core tests prove that no fallback remains. They do not. I inserted a coordinated fallback after
the snapshot read: when `s.store != nil`, each returned record called
`s.store.GetAudit(record.Audit.Slug)` and reparsed that body. `go test ./internal/core` still passed.

The current pathless test only counts calls on a narrow source (`internal/core/finding_test.go:82-109`),
while the ordinary cross-audit tests use the broad `fakeStore`, whose `GetAudit` succeeds and is not
counted (`service_epic_test.go:139-146,178-194`). The only committed one-scan counter covers
`Summary`, not `QueryFindings` (`usecases_test.go:163-202`). Consequently the default FS composition
can regress to N additional opens and mixed snapshot revisions while every claimed regression test
stays green. Add an aggregate fake that implements both `Store` and `AuditSnapshotSource`, make
`GetAudit` fail or count, execute unfiltered `QueryFindings`, and require zero fallback calls while
retaining the expected findings/problems.

**Resolution:** A counting aggregate Store plus AuditSnapshotSource fake now
proves QueryFindings performs one snapshot call and zero fallback GetAudit
calls; the coordinated reread mutation fails this test.

#### L1. Opaque non-path diagnostic semantics are absent from the committed machine-contract tests · **Status:** fixed

The mapper correctly gates compatibility `path` on `LocationIsPath`
(`internal/wire/envelopes.go:955-968`), and core currently returns snapshot problems untouched
(`internal/core/finding.go:175`). However, shipped fixtures exercise only a true local path
(`internal/cli/audit_test.go:38-73`) or a pathless diagnostic whose `Location` is empty
(`internal/cli/render/render_test.go:168-195`). The semantic schema case passes `nil` problems
(`internal/wire/envelopes_test.go:318-323`).

Changing the mapper to `path := problem.Location` passed the complete
`go test ./internal/wire ./internal/cli/render ./internal/cli` suites. A coordinated core mutation
that filled missing identity with `filepath.Base(problem.Location)` likewise passed
`go test ./internal/core ./internal/cli`. Either regression breaks the advertised remote/database
contract precisely when an adapter supplies useful opaque repair context: a URI is mislabeled as a
filesystem path or parsed into fabricated identity. Commit the hostile non-empty `db://...` and
path-looking-but-nonpath fixtures at the core, bare JSON, projected JSON, and semantic JSON Schema
layers; assert explicit identity precedence, unchanged opaque location, empty compatibility `path`,
and non-default optional fields.

**Resolution:** Committed hostile opaque and path-looking non-path fixtures
preserve explicit or absent identity, retain location, and keep compatibility
path empty across core, full/projected JSON, mapper, and schema tests.
