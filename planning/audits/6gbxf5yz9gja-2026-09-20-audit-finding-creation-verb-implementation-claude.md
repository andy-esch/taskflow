---
schema: 1
id: 6gbxf5yz9gja
bucket: closed
area: audit-finding-creation-verb-implementation-claude
date: "2026-09-20"
updated_at: "2026-09-20"
---

# Audit: Audit finding creation verb implementation — Claude — 2026-09-20

> Reviewer assignment: Claude. This document is the review brief and the only source file the reviewer may update.
>
> Take two passes: first contract/correctness, then systemic failure modes. Green tests are claims to challenge, not proof. Prefer one demonstrated defect over several speculative concerns.

## Mandatory isolated workspace

Treat the handoff checkout as read-only. Create an independent sandbox before inspecting code,
running tests, generators, or mutation probes:

```sh
SOURCE_ROOT="$(git rev-parse --show-toplevel)"
AUDIT_REL="planning/audits/6gbxf5yz9gja-2026-09-20-audit-finding-creation-verb-implementation-claude.md"
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
  2026-09-20-audit-finding-creation-verb-implementation-claude "<title>" \
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

#### M1. Guarded-snapshot bucket contract has no regression coverage · **Status:** fixed

**File:** internal/store/body.go:192 | **Component:** store

**Effort:** S · **Urgency:** soon

This task changed the audit store port so the transform callback receives the audit
parsed from the same snapshot the content CAS protects. That is the invariant the design
decision and the recorded implementation evidence both rest on, and AC 6 claims focused
tests for it. Nothing in the suite exercises it.

Mutation A, in the adapter: hand the callback a hardcoded open bucket instead of the
snapshot's own parsed metadata.

```
# internal/store/body.go:192
- newBody, err := transform(currentAudit, string(body))
+ newBody, err := transform(domain.Audit{Slug: currentAudit.Slug, Bucket: domain.AuditOpen}, string(body))
```

`go test -race ./...` stays fully green, and the binary built from it creates an open
finding inside a closed audit:

```
$ tskflwctl audit close 2026-09-20-closedarea
✔ moved 2026-09-20-closedarea -> closed
$ tskflwctl audit finding new 2026-09-20-closedarea "Should be impossible" --band H
✔ created H1 in 2026-09-20-closedarea       # exit 0, frontmatter still reads bucket: closed
```

The shipped binary refuses the identical call with exit 11, so the implementation is
correct; only the regression fence is missing.

Mutation B, in core: cache the bucket verdict from the first snapshot and replay it on
every conflict retry, which is exactly the "convenient pre-read" shape the brief names.
The whole suite is green again. The single bucket test,
TestNewFindingRefusesNonOpenAuditFromGuardedSnapshot, cannot observe either mutation,
because its fake synthesises the bucket in the same call it is asked to transform, so
snapshot-read and pre-read are indistinguishable to it.

A coordinated test does separate them. Driving a lifecycle move inside the write window
with testHookBeforeBodyWrite, then asserting the refusal and that no block landed, fails
under both mutations and passes on the shipped code for closed and for deferred. Two
such cases are enough to fence the port change this task exists to make.

**Recommendation:** Add a real-filesystem test that moves the audit to closed and to deferred inside the CAS window and asserts creation is refused and nothing is written.

**Resolution:** Added a real-filesystem close/defer race test; it fails when the
store fabricates an open snapshot and passes with guarded metadata.

#### M2. Creation is blocked by an unrelated candidate defect with no tool-owned repair · **Status:** fixed

**File:** internal/domain/finding_create.go:59 | **Component:** domain

**Effort:** M · **Urgency:** soon

CreateFinding runs LintCandidateTasks on every creation and refuses on any issue, even
when no candidate was requested. A creation without a candidate never touches the
Candidate tasks section, so the precondition is wider than the operation's blast radius,
and the state that trips it is reachable through the tool's own documented writes.

The default `audit new` scaffold ends with the managed Candidate tasks section, and
`audit append` appends to the end of the body. Appending ordinary prose therefore lands
inside the managed section, silently and with exit 0. Full reproduction in a disposable
repo, shipped binary throughout:

```
$ tskflwctl audit new trapped --date 2026-09-20
$ tskflwctl audit finding new 2026-09-20-trapped "First finding" --band H --candidate "A candidate row"
✔ created H1 in 2026-09-20-trapped                       # audit lint green

$ tskflwctl audit append 2026-09-20-trapped --body "A note about progress."
✔ appended to 2026-09-20-trapped                         # exit 0, no warning

$ tskflwctl audit finding new 2026-09-20-trapped "Unrelated finding" --band M
error: validation failed: existing Candidate tasks projection is malformed: line 32 is
not a canonical candidate row ...; run `audit lint` and repair it before creating
another finding                                          # exit 11  <-- new in this task

$ tskflwctl audit finding 2026-09-20-trapped H1 --candidate ""
error: validation failed: the managed Candidate tasks section is malformed ...   # exit 11
$ tskflwctl lint --fix
nothing to fix / could not auto-repair: candidate_tasks: line 32 ...             # exit 11
```

Every tool-owned route out is gated on the same check, and `lint --fix` declines it, so
the only recovery is hand-editing a section the CLI, README and routines all say must
never be hand-edited. The near-miss refusal names `lint --fix`, which really does repair
it; this refusal names `audit lint`, which only reports.

Attribution checked against main: the candidate-write refusal and the unrepairable lint
issue predate this task, but on main a poisoned section blocked only candidate writes.
Blocking all finding creation is new here, and it lands on the very verb the routines and
guidance were just rewritten to make the primary path. The generated reference documents
the open-audit and legacy-section preconditions but not this one.

**Recommendation:** Gate the candidate-projection check on a requested candidate, or ship a repair verb the refusal can name.

**Resolution:** Kept creation fail-closed, made audit append place narrative
before the trailing managed projection, and routed pre-existing corruption
through audit lint plus audit edit.

#### L1. CRLF audits gain an extra blank line at the insertion point · **Status:** fixed

**File:** internal/domain/finding_create.go:246 | **Component:** domain

**Effort:** XS · **Urgency:** eventually

insertMarkdownBlock decides its separator with strings.HasSuffix on the raw body against
LF-only literals. On a CRLF body the text before the insertion point ends with a CR LF
pair, so the two-newline case never matches, the one-newline branch wins, and a stray bare
LF is prepended to an already blank-line-terminated paragraph. The store folds the result
back to the file's own ending, so the artefact persists as a doubled blank line rather
than a mixed ending.

Domain-level A/B on two bodies that differ only in line ending:

```
lf   => "...#### H1. x · **Status:** open\n\nEvidence.\n\n#### H2. New ...\n\n## Candidate tasks..."
crlf => "...#### H1. x · **Status:** open\r\n\r\nEvidence.\r\n\r\n\n#### H2. New ...\n\n## Candidate tasks\r\n..."
```

Persisted through the CLI in a disposable repo, the same two bodies produce a one-line gap
under LF and a two-line gap under CRLF:

```
LF   file: 16 Evidence.        CRLF file: 16 Evidence.
     17 (blank)                     17 (blank)
     18 #### H2. Probe ...           18 (blank)
                                     19 #### H2. Probe ...
```

The resulting file is uniformly CRLF, 23 CRLF endings and no bare LF, so the EOL contract
itself holds and lint stays green. Only the whitespace differs, which is why no test sees
it: TestCreateFindingPreservesWhitespaceOutsideInsertion uses an LF fixture, and the CRLF
round-trip test at the CLI boundary covers candidate edits rather than creation.

**Recommendation:** Compare separators against the normalised body, or fold to LF before deciding the prefix.

**Resolution:** Made insertion separator detection newline-normalized and added
LF versus CRLF vertical-spacing coverage.

#### L2. Rendered finding block no longer matches the scaffold and routine templates · **Status:** fixed

**File:** internal/domain/finding_create.go:199 | **Component:** domain

**Effort:** XS · **Urgency:** eventually

renderFindingBlock joins every rendered part with a blank line, so the location line and
the planning line land in separate paragraphs. Both audit scaffolds in internal/domain
entity.go, and the AUDIT-FILE-TEMPLATE in both scheduled routines, show them on adjacent
lines. This task rewrote the surrounding guidance to make the writer authoritative while
leaving the templates describing the older shape, so the documented canonical grammar and
the only tool that now writes it disagree.

Scaffold and routine templates:

```
#### H1. <title>  · **Status:** open

**File:** <path:line> | **Component:** <component>
**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>
```

What `audit finding new` actually writes:

```
#### H1. Retry loop can lose evidence · **Status:** open

**File:** internal/store/body.go:194 | **Component:** store

**Effort:** S · **Urgency:** soon
```

Parsing is unaffected, since each label is matched independently anywhere in the section,
and lint stays green either way. The cost is that the templates the routines still tell
agents to follow no longer describe the tool's output, which is the drift the repository
elsewhere goes out of its way to make impossible.

**Recommendation:** Join the location and planning lines with a single newline, or update both audit scaffolds and both routine templates to the shape the renderer emits.

**Resolution:** Grouped location and planning metadata as adjacent lines,
matching the scaffold and routine templates.

## Reviewer report

### Isolation attestation

Every command below ran in an independent clone created by the mandatory helper; the source
checkout was never staged, committed, switched, restored, cleaned, or written to.

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.h7Vus4
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.h7Vus4/.git
baseline_commit=fe80e68fb87f43ddc66891cae1e45aa1fa52f5f4
source_blob=1c413a541dfd6f3f5858c2af1892fda519d19c7c
source_fingerprint=5b89aff74550ba7b2656a1da151908fa125a2618
deliverable=planning/audits/6gbxf5yz9gja-2026-09-20-audit-finding-creation-verb-implementation-claude.md
deliverable_changed=true
transfer=pending            # verify; transfer result appended after the final copy
```

After the last probe, `git status --porcelain -uall` in the sandbox reported exactly one
entry, this audit. Every mutation described below was restored byte-for-byte, which that
clean status proves independently of the reruns. All four findings were written by the
implementation under review, through `audit finding new --body-file -`.

### Evidence floor

1. `git diff --check` clean. `go test -race ./...` green across all 33 packages, before and
   after every probe. `golangci-lint run ./...` with `GOLANGCI_LINT_CACHE`, `GOCACHE` and
   `HOME` under `/tmp`: 0 issues. `tskflwctl lint` and `audit lint` with the sandbox-built
   binary: both green on `planning/`.
2. Disposable repo, documented path end to end: `init`, `audit new`, findings in H, M and L
   with every optional field, one with `--candidate`, `--dry-run --json` preview, then
   `audit finding <code> --status fixed --pr 42 --note ... --candidate ...`, then lint.
   Inspected the exact Markdown at each step. Allocation was H1, M1, L1, H2; the candidate
   row rendered as `- ○ H2 · open — ...`; lint stayed green.
3. Structural corpus, each against the shipped binary with a byte comparison of the file:
   no Findings section (creates one), no Findings section with a managed candidate section
   (creates one immediately before it), two real Findings sections (refused, byte-identical),
   `## Findings` examples inside backtick and tilde fences (correctly ignored), H3/H4
   sections following Findings, CRLF, Unicode and emoji titles with a decorated status,
   whitespace-only tails, near-miss `#### H-1.`, unknown status, missing status, duplicate
   codes, `H99999999999999999999`, `H2147483647`, allocation over the gap set {H1, H3, H07}
   yielding H8, and all-empty optional values. Candidate variants: missing, legacy,
   malformed row, two managed sections, drifted marker legend. Every refusal left the file
   byte-identical with no partial finding. Metadata-label injection through title, body and
   recommendation is refused, while a literal fenced `#### H9. ... **Status:** open`
   survives inside a fence, including in this report's own findings.
4. Closed and deferred audits are refused with exit 11. The coordinated case is finding M1:
   I wrote the missing test, drove a lifecycle move inside the write window, and confirmed
   the shipped code refuses while two plausible mutations of it do not.
5. Five concurrency rounds, 12 simultaneous creators across mixed bands plus three
   `audit append` writers plus concurrent `--status` and `--candidate` edits on the same
   audit. Zero duplicate receipt codes and zero duplicate codes in any file across all
   rounds; every successful creator's evidence prose and header were present; receipt
   tallies formed a strict serialised sequence (findings=2..11 for ten winners), so no
   receipt was stale. Losers failed cleanly with exit 14 and wrote nothing. Rounds 4 and 5
   are what surfaced finding M2.
6. Machine contract: the live receipt validates against the generated Draft 2020-12 schema
   with no extra and no missing required properties, recursively through `AuditJSON`,
   `FindingJSON` and `WorkspaceJSON`. `additionalProperties: false` plus the required list
   make a body echo a schema violation by construction. Optional non-default fields appear
   in the receipt; dry-run and workspace identity are present. `schema --json` reports
   revision 1.69 with `current_compatibility: additive`, and the golden diff against `main`
   is exactly the revision bump plus the new `FindingCreationEnvelope` branch, so no prior
   envelope changed.
7. Ran both routines' exact documented command, `--body-file -` with a heredoc and a quoted
   `task new` command in `--candidate`; the row rendered with its inner double quotes
   intact and lint stayed green. Regenerated `docs/cli` with the docgen tool and diffed:
   in sync. The generated help documents the open-audit and legacy-section behaviour
   accurately, but not the precondition in finding M2.
8. Refusal matrix, 16 cases: not-found maps to 10, every validation refusal to 11, cobra
   usage errors to 1. `updated_at` was pinned to a sentinel beforehand and was unchanged
   after all 16, as after every dry-run, callback failure, parse failure and exhausted
   retry observed in the concurrency rounds.

### Hostile probes

Each mutation was applied to the sandbox, built, run against the full suite, then restored
and re-run green.

- **Allocation outside the retry callback.** Precomputed the body and identity once and
  replayed it on retry. Caught twice: `TestNewFindingRetriesAllocationAgainstFreshBody`
  (core fake) and `TestNewFinding_RetriesAllocationAroundConcurrentAppend` (real
  filesystem), which reported H2 where H3 was required. Well fenced.
- **Snapshot bucket check replaced by a pre-read.** Two variants. A literal
  `GetAudit` pre-read fails the core tests only because the shared `nopStore` returns
  not-found, which is the fake failing rather than the race being caught. The faithful
  variant, caching the first snapshot's verdict across retries, passes the entire suite.
  See finding M1.
- **Permissive fence masking and section boundaries.** Making `insertFindingBlock`
  fence-blind fails `TestCreateFindingCreatesMissingFindingsSectionAndIgnoresFencedHeadings`;
  widening the boundary to `#{1,6}` fails
  `TestCreateFindingPreservesWhitespaceOutsideInsertion`, which caught the new block
  prepending ahead of H1. Both well fenced.
- **One reserved label permitted.** Dropping `Status` from the reserved pattern fails the
  `title metadata injection` case. The parser-owned field that changes is the status: with
  the guard removed, `--band H` with title `Title · **Status:** wontfix` renders a header
  carrying two status markers, and `statusRe` takes the first, so `audit findings` reports
  the brand-new finding as `wontfix`. A finding born settled would count toward the audit's
  done band and toward the headline settled percent that gates `audit close`.
- **Finding and candidate split into two writes.** Caught at both layers, by the core fake
  (`TestNewFindingCandidateRefusalIsAtomicAndDryRunDoesNotPersist`) and by the real-store
  CLI test (`TestAuditFindingNewCandidateRefusalLeavesLegacyAuditUntouched`), each showing
  the partial finding left behind. Well fenced.
- **Receipt echoes the body / loses its schema branch.** Adding a `body` field fails an
  explicit CLI assertion plus `TestSchemaComments_NotStale` and the jsonschema golden.
  Removing the registry entry fails `TestJSONEnvelopes_RegistryIsComplete`,
  `TestJSONSchema_ValidatesRealOutput` and the golden. Well fenced, and the new envelope is
  in the exact-validation fixture list with non-default optional fields.
- **Shared helpers and fakes.** This is where the gap is. `findingCreationStore`
  synthesises the bucket in the same call it transforms, so snapshot-read and pre-read are
  indistinguishable to it, and no real-filesystem test constrains the adapter half:
  hardcoding `Bucket: domain.AuditOpen` in `TransformAuditBody` leaves the whole suite
  green while the binary happily creates open findings inside closed audits. Stale retries
  and persisted dry-runs are not exposed this way, both having real-filesystem coverage.

### Verdict

The implementation is correct on every contract point I could reach. Allocation genuinely
recomputes inside the CAS callback, the bucket guard genuinely reads the snapshot the write
is protected against, creation and the candidate row are genuinely one transform, fence
masking and boundary detection hold against backtick and tilde fences and CRLF, injection
through title, evidence and recommendation is closed while fenced literals still work, the
receipt is compact and schema-exact, and the 1.69 bump is additive with no prior envelope
disturbed. Under 14-way contention it never duplicated an identity, lost prose, wrote a
duplicate row, produced a partial write, or returned a stale receipt. No finding here
disputes shipped behaviour.

Two things are worth the owner's time. M1 is a coverage defect on the one invariant this
task changed the store port to establish: the guarded-snapshot bucket read is unverified at
both layers, so AC 6's "focused tests" and the recorded evidence overstate what is fenced,
and two mutations of correct code ship green. M2 is a real interaction defect: the new verb
refuses on a candidate-projection defect it does not touch, that state is reachable through
the tool's own `audit append` against the tool's own scaffold, and no tool-owned path
repairs it. L1 and L2 are cosmetic.

### Residual risks

- **Contention budget.** With 14 concurrent writers on one audit, 2 to 6 of 12 creators
  exhausted their retries and returned exit 14. Correct under OCC, and no work was lost,
  but a routine that authors a dozen findings in parallel will need to serialise or retry
  at the caller.
- **`H99999999999999999999` leaks a Go internal.** The refusal reads `cannot participate in
  allocation: strconv.Atoi: parsing ...: value out of range` and offers no repair. It fails
  closed and only blocks the matching band, so it is a message-quality nit rather than a
  finding.
- **Codes above int32.** An existing `H2147483647` allocates `H2147483648`. Fine for this
  codebase, which uses `Atoi` into `int`, but any 32-bit consumer of the code string would
  overflow.
- **Setext and indented headings.** The body guard rejects ATX headings only. A setext
  underline, or a heading indented up to three spaces, renders as a heading in CommonMark
  while this repository's scanners ignore it. Self-consistent for the tool, so nothing
  breaks, but the "one creation cannot open another section" promise is an ATX-only promise.
- **H3 and H4 placement.** A `###` subsection between `## Findings` and the next `##`
  section is treated as part of the Findings section, so a new block lands after it rather
  than beside its siblings. Markdown-correct and the finding still parses; noted only
  because the evidence floor names those depths.

The sandbox is retained at the path in the attestation for owner-confirmed cleanup.
