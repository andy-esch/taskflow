---
schema: 1
id: 6gbjq90ts0cr
bucket: closed
area: machine-contract-revision-policy-implementation-codex
date: "2026-09-19"
updated_at: "2026-09-19"
---
# Audit: Machine contract revision policy implementation — codex — 2026-09-19

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

Adversarially review the implementation of task
`define-and-enforce-the-cli-machine-contract-revision-policy` (`6ga968z46e9y`) and accepted
ADR-0008. The change replaces the previous SemVer-like description of `schema_version` with an
explicit monotonic revision policy, advances the global JSON revision from 1.67 to 1.68, exposes
the policy through `schema --json`, and gives the generated JSON Schema a revision-qualified
identity and compatibility annotations.

Do not merely check that the new constants and goldens agree. Try to falsify the claimed contract:
find machine-readable JSON surfaces that escape the revision, readers for which the additive
classification is dishonest, author workflows that can evade the changelog guard, and documentation
that still teaches a contradictory rule. Take a second pass for systemic weaknesses that would let a
future feature ship a machine-contract change without a truthful revision or classification.

## Review target

Review the isolated sandbox baseline containing the uncommitted implementation on branch
`feat/machine-contract-revision-policy`, relative to base
`fb09b335136c5622e36ca9662ee2285c085b7c8f`. Inventory every production and test consumer of:

- `wire.SchemaVersion`, `SchemaRevisionPolicy`, `CurrentSchemaRevisionPolicy`, and all
  `SchemaRevision*` / `JSONSchema*` policy constants;
- typed JSON envelopes, caller-selected/projected JSON, error envelopes, and any JSON-producing
  command or adapter that claims the global revision;
- `ToSchemaEnvelope`, `schema --json`, `schema --json-schema`, and both human and machine schema
  renderers;
- `JSONSchema`, its `$id`, extension annotations, `$defs`, and any tests, docs, fixtures, caches, or
  downstream commands that consume or compare the generated schema;
- the source changelog parser and classification guard in `wire_changelog_test.go`;
- ADR-0008, architecture and compatibility docs, README/CLAUDE guidance, the weekly architecture
  routine, task/thread planning, updated audit dispositions, and regenerated goldens.

The untracked audit
`planning/audits/6ga5dr4g73zq-2026-09-14-go-1-27-modernization.md` is unrelated concurrent work.
Do not review, edit, restore, or transfer it. The generated golden movement from 1.67 to 1.68 is in
scope, but verify it contains no accidental contract changes.

## Intended contract to challenge

1. `schema_version` is a monotonically increasing repository-wide revision for all CLI machine JSON,
   not SemVer and not the persisted frontmatter `schema:` field. It covers typed envelopes, dynamic
   projections, and JSON error behavior even though only typed envelopes can be represented by the
   generated static JSON Schema.
2. A revision is classified `additive` only when a tolerant semantic JSON reader that ignores
   unknown object fields can continue operating. Any other change is `not-additive`; this is a
   disclosure, not a promise of runtime negotiation or automatic migration. Major revision movement
   remains reserved rather than acquiring SemVer meaning.
3. Revision 1.68 is truthfully additive: `schema --json` gains `revision_policy`, while existing
   machine envelopes otherwise change only from `schema_version: 1.67` to `1.68`. The policy declares
   `classified_since: 1.68`, so pre-1.68 changelog entries need no retroactive classification.
4. Every changelog entry at or after 1.68 must state exactly `ADDITIVE` or `NOT ADDITIVE`; the current
   entry must agree with `SchemaRevisionCompatibility`; revisions remain contiguous and ascending.
   A normal future author should receive a focused test failure when any part of this obligation is
   missed or contradictory.
5. `schema --json` is the canonical self-description. It publishes the scheme, scope, default and
   current compatibility, classification boundary, reader expectation, generated-schema scope, and
   exact-revision validation mode from wire-owned constants. `ToSchemaEnvelope` stamps the current
   policy so another primary adapter cannot emit an empty or stale policy supplied by a caller.
6. `schema --json-schema` describes typed envelopes only. Its `$id` is stable for revision 1.68 and
   revision-qualified, while root `x-taskflow-*` annotations identify the exact revision, scheme, and
   current classification. Exact JSON Schema validation is intentionally stricter than tolerant
   forward-compatible decoding.
7. Documentation and inline comments consistently instruct callers to persist durable IDs as entity
   handles, treat slugs as readable/renamable selectors, request bounded projections, tolerate
   additive object fields, and surface actionable errors. These claims must match shipped behavior
   and must not imply that all projected shapes are statically schema-validated.
8. The policy is owned at the wire boundary and reusable by future CLI, TUI, web, or other primary
   adapters rather than being coupled to one Cobra rendering path.

Non-goals: redesigning persisted Markdown/frontmatter storage, adding runtime schema negotiation,
retrofitting classifications before 1.68, adopting a schema registry, versioning every envelope
independently, or resolving the unrelated Go 1.27 audit.

## Mandatory evidence floor

- Record the sandbox baseline commit and captured source fingerprint. Inspect the complete scoped
  diff and build a repository-wide consumer inventory before assigning severity. Search for every
  JSON encoder and every reference to `SchemaVersion`, `schema_version`, JSON Schema `$id`, revision
  policy, compatibility, SemVer language, and persisted `schema:`.
- Run `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, documentation checks,
  planning lint, and `git diff --check` inside the sandbox. Use sandbox-local or `/tmp` Go/lint caches
  when necessary.
- Independently regenerate the machine goldens and schema comments in the sandbox, then compare them
  with the captured baseline. Inspect every changed golden semantically; do not accept a bulk update
  as evidence that only the intended revision changed.
- Build or `go run` the candidate against a disposable planning space. Probe `schema`,
  `schema --json`, and `schema --json-schema`; at least one typed envelope, one dynamic projected
  list, and one JSON error path. Confirm which surfaces carry revision 1.68 and which cannot carry the
  full revision policy by design.
- Parse the generated JSON Schema with an independent Draft 2020-12-capable validator if one is
  already available without modifying dependencies. Resolve representative root properties and
  local `$defs`; test a valid typed envelope and an envelope with an unknown object field. Clearly
  separate exact-revision schema behavior from tolerant-reader behavior.
- Trace at least one future additive change and one future not-additive change through the author's
  expected workflow. Verify the documentation, source changelog, constants, generated schema,
  comments, goldens, and tests give a coherent and discoverable obligation rather than relying on
  tribal knowledge.
- Perform and restore at least six mutation probes: remove the 1.68 classification; change it to an
  unknown marker; make the current compatibility constant disagree; move or weaken the 1.68 boundary;
  stop `ToSchemaEnvelope` from stamping the policy; and remove or de-version the JSON Schema `$id` or
  one root annotation. The relevant focused test must fail for the intended semantic reason. Add a
  coordinated mutation where a nearby constant or test fixture would otherwise preserve the invariant
  accidentally.

## Required hostile angles

- Challenge the scope inventory. Search for raw `json.NewEncoder`, `json.Marshal`, hand-written JSON,
  projected rows, completion/helper responses, fatal error paths, and alternate adapters. Determine
  whether “all-json-output” is true, aspirational, or misleading when some outputs may not carry a
  visible revision field.
- Challenge the compatibility model with concrete consumer shapes. Test tolerant decoders, strict
  struct decoders, exact JSON Schema validation, key-order/byte comparisons, and projected JSON whose
  caller-selected keys are dynamic. Do not call strict consumers compatible merely because the policy
  tells them to be tolerant; decide whether the documentation makes the precondition explicit enough.
- Treat `ADDITIVE` for 1.68 as a claim to prove. Compare representative 1.67 and 1.68 outputs beyond
  updated version strings. Look for a newly required field, renamed key, changed omission/null/empty
  behavior, changed enum, exit code, channel, ordering guarantee, or schema constraint hidden by
  shared golden-update helpers.
- Attack the changelog regex and classification validator with malformed punctuation/case, duplicate
  entries, missing entries, a future major transition, comments elsewhere in the bounded source
  region, inconsistent boundary/current constants, reordered declarations, and a skipped revision.
  Determine whether a plausible author can evade the guard while all tests remain green.
- Attack policy ownership. Construct a `SchemaContract` with empty, stale, and contradictory policy
  values and follow every renderer/adapter. Verify machine and human paths derive from the same source
  without allowing a future adapter to bypass stamping or a test helper to conceal such a bypass.
- Attack JSON Schema identity and caching semantics. Determine whether the new version-qualified `$id`
  is a valid absolute identifier, whether local references still resolve, whether any repo consumer
  assumes the former absent/stable identity, and whether the annotations are discoverable without
  being mistaken for validation rules.
- Reconcile prose across ADR-0008, `docs/ARCHITECTURE.md`, `docs/THREADS_COMPATIBILITY.md`, README,
  CLAUDE guidance, routine docs, source comments, and `schema --json`. Search for surviving “minor is
  additive / major is breaking” or envelope-only claims. Distinguish durable architectural guidance
  from duplicated prose likely to drift.
- Review agent ergonomics claims skeptically. Verify stable IDs are actually published wherever the
  docs tell callers to store them, slugs remain usable selectors where claimed, bounded projections
  behave consistently, and errors contain enough machine-readable context for correction. If a gap is
  pre-existing and genuinely outside this task, demonstrate it and recommend a specifically scoped
  follow-up rather than expanding this implementation casually.
- Conduct a final systemic pass for shared fixtures, constructors, or renderers that can make broad
  classes of revision-policy defects pass together. Prefer a reproducible invariant failure over
  speculative architecture advice, and look for the smallest core seam that future entities and
  interfaces can share.

## Validation and restoration

Run every probe only in the mandatory independent sandbox. Capture commands and decisive output in
the report. Restore each source mutation to the sandbox baseline before the final verdict. Do not
stage, commit, push, switch branches, regenerate in the source checkout, touch the unrelated Go 1.27
audit, or copy back any file except the assigned audit through the isolation helper. Before transfer,
`git status --short` may show only the assigned audit and the helper's `verify` must pass.

## Deliverable

Replace the placeholder below with an evidence-backed review. Keep every finding `open` for the
implementation owner. For each finding include severity, exact file/line or symbol, reproduction,
impact, recommended repair, and whether it is in scope or belongs in a follow-up task. If no findings
survive, explicitly report the consumer inventory, compatibility experiments, changelog attacks,
schema-validation results, mutation outcomes, commands run, and why the claimed revision policy held.

Include the mandatory isolation attestation: sandbox path, resolved independent Git directory,
baseline commit, captured source audit fingerprint/blob, assigned deliverable, verification result,
and transfer result. Leave the sandbox intact until receipt is confirmed.

## Reviewer report

Reviewer: codex

### Verdict

The shipped 1.68 behavior is internally coherent and truthfully additive, but the review found two
systemic enforcement gaps and one validation-boundary mismatch. The current binary publishes the
right policy, revision, schema identity, typed envelopes, projected rows, and JSON errors. The open
findings concern what the implementation and its tests permit next: routine golden regeneration can
bless a contract change without a revision, the human and machine schema renderers do not own policy
normalization at the same seam, and the schema advertised as exact-revision does not constrain an
instance's `schema_version` to its own revision.

#### M1. Golden regeneration can bless a machine-contract change without a new revision · **Status:** fixed

**Severity:** Medium | **Scope:** in scope | **Files:**
`internal/wire/wire_changelog_test.go:42-64,67-92`,
`internal/cli/golden_test.go:10-34`, `internal/wire/wire.go:283-310`

The changelog guard proves that the source comment is internally contiguous, classified from 1.68,
and equal to the current constants. The golden helper proves that checked-in output equals current
output. Neither guard binds a changed contract to a *different* revision. The normal repair printed
by a golden failure—`go test ./internal/cli -update`—rewrites the evidence in place, after which the
classification test sees the unchanged, already-valid 1.68 entry.

**Reproduction:** in the isolated sandbox I added a required `contract_probe` field to
`wire.SchemaContract`, populated it from `runSchemaContract`, and ran the documented golden update.
That changed both `schema_json.golden` and `schema_jsonschema.golden`. With
`SchemaVersion` still exactly `1.68` and no new changelog entry,
`go test ./internal/wire ./internal/cli -count=1` passed. The source and regenerated goldens were then
restored and matched the baseline byte-for-byte. A second probe appended
`// 2.0: ADDITIVE ...`, changed only `SchemaVersion` to `2.0`, and showed both
`TestSchemaVersionChangelogIsAscending` and `TestSchemaRevisionPolicyBoundary` pass even though
ADR-0008 reserves major movement for another ADR; this is also explicitly enabled by
`wire_changelog_test.go:54-55,158-163`.

**Impact:** a normal author who changes a typed envelope or self-description and follows the test's
own regeneration instruction can ship a new machine contract under an old identity. Consumers that
cache or dispatch on exact revision then receive changed shape/semantics without the declaration this
task exists to require. Dynamic projections are even easier to miss because they have no static
schema definition.

**Recommended repair:** add a non-overwritable revision fixture or manifest keyed by revision (for
example, `testdata/machine-contract/1.68/`) that includes the generated typed schema plus projection
registries and other non-schema machine metadata. A normal update command should refuse to replace
an existing revision fixture when its canonical content changes and should direct the author to add
the next changelog entry/classification first. Pin the allowed major line to 1 until a superseding ADR
changes that boundary. Add the exact unchanged-revision mutation above as a regression test.

**Resolution:** The golden updater now reads a committed machine-contract
revision marker and refuses changed or new JSON snapshots at the same revision.
A successful full CLI update advances the marker only after every machine golden
runs; partial updates cannot advance it. A projection-contract golden covers
ordered selectors and aliases outside the static schema, the reserved major line
is pinned, and focused tests exercise same-revision refusal.

#### M2. Human schema policy can diverge from the wire-stamped machine policy while tests stay green · **Status:** fixed

**Severity:** Medium | **Scope:** in scope | **Files:**
`internal/cli/schema.go:104-130`, `internal/cli/render/schema_render.go:35-47`,
`internal/wire/envelopes.go:994-1006`, `internal/cli/output_coverage_test.go:84-90`

`SchemaJSON` reaches `ToSchemaEnvelope`, which overwrites empty, stale, or contradictory policy input
with `CurrentSchemaRevisionPolicy`. `SchemaHuman` instead prints the caller-supplied
`c.RevisionPolicy` directly. The production Cobra assembly currently initializes that duplicate
value correctly, but the two renderer paths do not derive at the same ownership boundary.

**Reproduction:** constructing `SchemaContract` with empty, stale, and contradictory non-default
policies showed `ToSchemaEnvelope` stamped the current policy in all three cases. Passing a stale
contract through the two renderers produced current policy from `SchemaJSON` and the literal stale
policy from `SchemaHuman`. More concretely, I removed only the initializer at
`internal/cli/schema.go:105`; `go test ./internal/wire ./internal/cli -count=1` remained green, machine
`schema --json` still returned the complete current policy, but human `schema` rendered:

```text
JSON revision:  ·  · classified since
Scope:  · generated schema:
```

The existing human test checks only the older section headings and therefore masks this divergence.
The initializer was restored after the probe.

**Impact:** a cleanup that removes what appears to be redundant policy initialization—or a future
renderer/adapter that supplies a partially built `SchemaContract`—can make the human and machine
self-descriptions contradict each other without a failing test. This weakens the claimed wire
ownership and leaves the policy coupled to one Cobra assembly path.

**Recommended repair:** normalize policy at one shared boundary used by both renderers. The smallest
repair is for `SchemaHuman` to read `wire.CurrentSchemaRevisionPolicy()` rather than caller data;
alternatively expose a wire-owned normalized contract constructor and make both paths consume it.
Add hostile empty/stale/contradictory cases plus a human-output assertion for all published policy
values.

**Resolution:** Schema contracts now pass through wire.NormalizeSchemaContract
for both human and machine rendering, and the redundant CLI initializer was
removed. Empty, stale, and contradictory caller policies are replaced by the
running wire policy in focused renderer and envelope tests.

#### L1. The 1.68 schema accepts payloads that declare a different revision · **Status:** fixed

**Severity:** Low | **Scope:** in scope | **Files:** `internal/wire/envelopes.go:100-109,1204-1227`,
`internal/cli/testdata/golden/schema_jsonschema.golden:3627-3643`,
`planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md:137-148`

The schema's `$id` and root annotations identify 1.68, but every envelope's `schema_version` property
is reflected only as `{ "type": "string" }`. `$id` establishes the schema resource's identity; it
does not constrain instance data.

**Reproduction:** using the repository's independent Draft 2020-12 validator, I validated the real
`task show --json` payload against
`https://github.com/andy-esch/taskflow/internal/wire/json-envelopes/1.68#/$defs/TaskShowEnvelope`, then
changed only its `schema_version` to `0.0`. Both instances validated. The same validator rejected an
unknown object property, so the result is specifically a revision constraint gap rather than a
failure to honor `additionalProperties: false`.

**Impact:** `json_schema_validation: "exact-revision"` is only an out-of-band instruction to select
the right resource. A caller that accidentally pairs a cached schema with the wrong payload revision
gets no fail-closed signal from validation, even though the payload carries the comparison value.

**Recommended repair:** make `schema_version` a `const` of `SchemaVersion` in each reflected envelope
definition (post-processing the generated schema is sufficient) and add a validator regression that
changes only this field. If out-of-band comparison is intentional, rename the published validation
mode and state explicitly that the schema validates shape only after the caller separately compares
the payload revision to the schema annotation.

**Resolution:** JSONSchema now adds const: SchemaVersion to schema_version on
every registered envelope definition and fails generation if any registered
envelope cannot be constrained. A Draft 2020-12 validator regression accepts a
current TaskShowEnvelope and rejects the same payload after only its revision is
changed.

### First-pass inventory and 1.68 compatibility evidence

- The captured change was reviewed relative to `fb09b335136c5622e36ca9662ee2285c085b7c8f`.
  Excluding the two reviewer briefs and the explicitly unrelated Go 1.27 audit, the scoped diff is
  55 files / 652 insertions / 74 deletions. Production changes are confined to the wire policy,
  schema/envelope generation, CLI assembly/rendering, documentation, planning dispositions, and
  generated artifacts.
- `SchemaVersion` and all `SchemaRevision*` / `JSONSchema*` policy constants are owned together at
  `internal/wire/wire.go:286-310`. `CurrentSchemaRevisionPolicy` maps them into the wire DTO at
  `internal/wire/schema.go:29-55`. The only production policy consumers are the schema contract
  assembly, machine/human schema renderers, `ToSchemaEnvelope`, and `JSONSchema`; repository-wide
  search found no second policy definition.
- The registry guard reported **53 declared envelope types, all registered**. Success JSON emitters
  call `wire.EncodeJSON` through `internal/cli/render`; fatal JSON uses the named `ErrorEnvelope` at
  `internal/cli/exit.go:62-126`. The two deliberate exceptions are caller-selected list projections,
  whose ordered-object writer inserts `SchemaVersion` at
  `internal/cli/render/columns.go:306-346`, and the generated schema document itself, whose visible
  revision is the `$id`/root annotation rather than an envelope field. No production `json.Marshal`
  or `json.NewEncoder` path was found that emits another CLI JSON surface.
- Live disposable-space probes confirmed `schema`, `schema --json`, and
  `schema --json-schema`; a typed `task show`; a projected `task list --json -c id,slug`; and a
  not-found JSON error. The typed and projected outputs and stderr error all carried `1.68`; the
  error exited 10 with zero stdout bytes. The schema document carried revision `1.68`, scheme
  `monotonic-revision`, compatibility `additive`, the revision-qualified absolute `$id`, and 148
  `$defs` entries.
- All 37 changed JSON goldens were compared to the captured base semantically. After normalizing
  `schema_version`, the 35 non-schema goldens were identical. `schema_json.golden` adds only
  `revision_policy`; `schema_jsonschema.golden` adds the qualified `$id`, policy definition/required
  property, title revision, and three root annotations. No hidden key/type/null/empty/ordering,
  vocabulary, channel, or exit-code change was found. Independent `go test ./internal/cli -update`
  regeneration left the baseline clean.
- The durable-handle claims held in the exercised surfaces: full task output and the bounded
  `id,slug` projection both published the stable ID beside the readable slug. Dynamic projections
  retained requested key order and are explicitly excluded from static schema validation. A
  duplicate-dependency disposable fixture emitted exactly
  `tskflwctl task depend repair --auto`; running that command from the recommending state changed
  graph health from broken to healthy, and a repeat diagnosis reported no defects.
- Author-workflow probes for a truthful future `1.69: ADDITIVE` and
  `1.69: NOT ADDITIVE` both propagated the revision/classification through constants, `schema
  --json`, the generated schema identity/annotations, and all 37 goldens; after complete regeneration,
  wire and CLI tests passed. This proves the intended workflows work, while M1 records the missing
  guard when the author omits the revision step.
- Primary guidance is consistent across ADR-0008, `docs/ARCHITECTURE.md`,
  `docs/THREADS_COMPATIBILITY.md`, README, CLAUDE.md, the weekly architecture routine, source
  comments, and `schema --json`: the revision is not SemVer, dynamic projections are revisioned but
  not statically described, callers compare exact revisions, and additive readers ignore unknown
  object fields. Historical planning prose that quotes the superseded SemVer rule remains evidence
  inside the resolved architecture audit, not current guidance. The on-disk `schema:` field remains
  clearly separate.

### Compatibility and schema experiments

- A tolerant Go struct decoder accepted a future unknown object field and retained the known task
  ID/slug. `DisallowUnknownFields` rejected the same payload, byte comparison differed, and the exact
  1.68 schema rejected the extra field. This supports the `ADDITIVE` classification only for the
  documented tolerant-reader precondition; strict structs, exact schema validation, and byte/key
  comparisons are intentionally not forward compatible.
- The independent validator parsed the 1.68 Draft 2020-12 document, resolved its local `$defs`, and
  accepted a real `TaskShowEnvelope`. `TestJSONSchema_ValidatesRealOutput` separately validates one
  real constructed value for every registry-derived envelope definition. L1 records the distinct
  wrong-revision case.
- Changelog attacks behaved as intended for removal, lowercase/malformed casing, unknown marker,
  duplicate release, skipped revision, current-constant mismatch, and a weakened 1.68 boundary.
  Each focused test failed with the targeted semantic diagnostic. A valid 1.68 baseline and both
  additive/non-additive 1.69 workflows passed. The accepted reserved-major probe is included in M1.

### Required mutation outcomes

| Mutation | Focused result |
| --- | --- |
| Remove the 1.68 classification | Failed: `1.68 has no compatibility classification` |
| Replace it with `BREAKING` | Failed: unknown compatibility classification |
| Set current compatibility to `not-additive` while the marker stays additive | Failed: current constant/changelog mismatch |
| Move `classified_since` to 1.69 | Failed: ADR-0008 boundary changed |
| Remove `ToSchemaEnvelope` policy stamping | Failed: empty policy instead of current wire policy |
| De-version the JSON Schema `$id` | Failed: ID did not carry `1.68` |
| Remove the root compatibility annotation | Failed: annotation decoded empty |
| Coordinated removal of the wire stamp and CLI initializer | Failed: CLI contract policy empty |
| Add a contract field, update goldens, keep revision 1.68 | **Passed** wire/CLI suites; M1 |
| Remove only the CLI policy initializer | **Passed** wire/CLI suites while human output went blank; M2 |

Every source/golden mutation was restored immediately. Before report editing, `git status` in the
sandbox was clean.

### Commands and results

- `go test -race ./...` — passed all packages.
- `golangci-lint run ./...` — passed, 0 issues.
- `go mod tidy -diff` — passed after downloading pinned transitive test modules into a sandbox-local
  module cache.
- `just docs-check` — passed; regenerated CLI reference had no drift.
- `go run ./internal/tools/schemacomments -out /tmp/taskflow-review-schema-comments.json` plus
  `diff -u` — passed; 265 generated comments matched.
- `go run ./cmd/tskflwctl --no-color lint` — passed all planning entities and links.
- `git diff --check <base>..HEAD -- . ':(exclude)<unrelated-go-1.27-audit>'` — passed. The unscoped
  command reports only a trailing blank line in the expressly excluded concurrent audit; this review
  did not edit or transfer it.
- `go test ./internal/cli -update -count=1` followed by a baseline diff — passed and reproduced every
  committed golden exactly.

### Isolation attestation

- Workspace: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.smWQKV`
- Resolved Git directory:
  `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.smWQKV/.git`
- Baseline commit: `0ccd8b8c94980956eb31bdd7b88859860dbaa318`
- Captured source deliverable blob: `18d06b026720dda7d8bd9df1694f829dfdffc42a`
- Captured source fingerprint: `db9640291e41ec9035cd795720d20f8e3f19b49b`
- Assigned deliverable:
  `planning/audits/6gbjq90ts0cr-2026-09-19-machine-contract-revision-policy-implementation-codex.md`
- Verification result: succeeded; only the assigned audit differed from the sandbox baseline.
- Transfer result: succeeded; the helper copied only the assigned audit through its guarded atomic
  transfer and retained the sandbox for owner confirmation.
