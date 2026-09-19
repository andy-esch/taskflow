---
schema: 1
id: 6gbjq9149xmh
bucket: closed
area: machine-contract-revision-policy-implementation-antigravity
date: "2026-09-19"
updated_at: "2026-09-19"
---
# Audit: Machine contract revision policy implementation — antigravity — 2026-09-19

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

### Mandatory isolation attestation

```ini
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/.git
baseline_commit=edc9e9d689627c9f074771f947a8b5096cb4360c
source_blob=2d3a8eef61ee67cfe6d08960eaf7bcb272dccff8
source_fingerprint=db9640291e41ec9035cd795720d20f8e3f19b49b
deliverable=planning/audits/6gbjq9149xmh-2026-09-19-machine-contract-revision-policy-implementation-antigravity.md
deliverable_changed=true
transfer=succeeded
```

---

### Executive verdict & assessment

- **Verdict:** **Pass with findings**. The core implementation of task [`6ga968z46e9y`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/planning/tasks/6ga968z46e9y-define-and-enforce-the-cli-machine-contract-revision-policy.md) and accepted [`ADR-0008`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md) successfully decouples the machine-readable JSON contract from SemVer, introduces a strictly monotonic revision scheme, exposes the policy cleanly through `schema --json` and `schema --json-schema`, and test-enforces changelog classifications from boundary revision 1.68 onward.
- **Contract truthfulness:** Revision 1.68 is **truthfully additive**. The only modifications to existing machine envelopes are the version identifier change (`"1.67"` → `"1.68"`), while `schema --json` gains the new object property `revision_policy`. Draft 2020-12 schema validation remains strictly scoped to typed envelopes, while dynamic projections correctly participate in the monotonic revision scheme without false validation claims.
- **Findings summary:** 3 low-severity findings identified and kept `open`:
  - **L1**: Asymmetry in revision policy ownership between `SchemaHuman` and `SchemaJSON` when passed an unstamped `SchemaContract`.
  - **L2**: `internal/wire` package-level tests omit assertions for three policy constants (`SchemaRevisionCompatibilityDefault`, `SchemaRevisionReaderExpectation`, and `JSONSchemaValidationMode`).
  - **L3**: Changelog classification parser error diagnostic masks em-dash delimiter requirements for future contributors.

---

### 1. Repository-wide consumer inventory

An exhaustive audit of all production and test references to `SchemaVersion`, `schema_version`, JSON encoders, and revision policy constants identified the following consumers:

1. **Policy declaration & wire constants (`internal/wire/wire.go:280-311`):**
   - [`SchemaVersion`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L283) (`"1.68"`)
   - [`SchemaRevisionScheme`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L289) (`"monotonic-revision"`)
   - [`SchemaRevisionScope`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L291) (`"all-json-output"`)
   - [`JSONSchemaScope`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L294) (`"typed-envelopes"`)
   - [`SchemaRevisionClassificationSince`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L297) (`"1.68"`)
   - [`SchemaRevisionCompatibility`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L300) (`"additive"`)
   - [`SchemaRevisionCompatibilityDefault`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L304) (`"additive"`)
   - [`SchemaRevisionReaderExpectation`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L307) (`"ignore-unknown-object-fields"`)
   - [`JSONSchemaValidationMode`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire.go#L310) (`"exact-revision"`)
2. **Wire constructors & DTOs:**
   - [`CurrentSchemaRevisionPolicy`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/schema.go#L43-L55): Assembles `SchemaRevisionPolicy`.
   - [`ToSchemaEnvelope`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/envelopes.go#L1001-L1005): Enforces stamping `c.RevisionPolicy = CurrentSchemaRevisionPolicy()`.
   - [`JSONSchema`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/envelopes.go#L1208-L1228): Emits Draft 2020-12 document with version-qualified `$id` (`.../json-envelopes/1.68`) and root `x-taskflow-*` extras.
   - All 43 typed envelope constructors in `internal/wire/envelopes.go` stamp `SchemaVersion: SchemaVersion`.
3. **CLI rendering layer (`internal/cli/render`):**
   - [`render.SchemaVersion`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/cli/render/render.go#L27): Re-exports `wire.SchemaVersion`.
   - [`render.ProjectedListJSON`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/cli/render/columns.go#L321-L347): Writes dynamic projection envelope, stamping `schema_version` as the first ordered key.
   - [`render.SchemaHuman`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/cli/render/schema_render.go#L41-L50): Emits human-readable summary of `c.RevisionPolicy`.
   - [`render.SchemaJSON`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/cli/render/schema_render.go#L36-L38): Serializes `ToSchemaEnvelope(c)`.
4. **Error handling:**
   - [`cli.WriteError`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/cli/exit.go#L66-L127): Dispatches versioned `wire.ErrorEnvelope{SchemaVersion: wire.SchemaVersion}` with structured domain error payload.
5. **Changelog parser & enforcement tests:**
   - [`internal/wire/wire_changelog_test.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/wire/wire_changelog_test.go): Enforces monotonic contiguous versions, `ADDITIVE` or `NOT ADDITIVE` classification at or above 1.68, and agreement with `SchemaRevisionCompatibility`.
6. **Documentation & architectural guidelines:**
   - [`docs/ARCHITECTURE.md:507-514, 741-748`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/docs/ARCHITECTURE.md#L507-L514)
   - [`docs/THREADS_COMPATIBILITY.md:21, 44`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/docs/THREADS_COMPATIBILITY.md#L21)
   - [`README.md:378-384`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/README.md#L378-L384)
   - [`CLAUDE.md:28-31`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/CLAUDE.md#L28-L31)
   - [`routines/weekly-architecture-audit.md:351-361`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/routines/weekly-architecture-audit.md#L351-L361)
   - [`planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/planning/adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)

---

### 2. Validation suite & toolchain execution evidence

The full test and verification suite was executed inside the sandbox:

| Phase | Command | Result |
| :--- | :--- | :--- |
| **Race detector suite** | `go test -race ./...` | **PASS** (all 20+ packages clean, 0 races) |
| **Linter** | `golangci-lint run ./...` | **PASS** (0 issues) |
| **Module hygiene** | `go mod tidy -diff` | **PASS** (no go.mod / go.sum diff) |
| **CLI reference sync** | `go run ./internal/tools/docgen -out <tmp> && diff` | **PASS** (committed docs match generator) |
| **Schema comment sync** | `go run ./internal/tools/schemacomments -out <tmp> && diff` | **PASS** (265 comments up-to-date) |
| **Planning lint** | `go run ./cmd/tskflwctl --no-color lint` | **PASS** (all planning entities valid) |
| **Whitespace/diff check**| `git diff --check` | **PASS** (clean) |
| **Golden regeneration** | `go test ./internal/cli -update` | **PASS** (0 working-tree modifications generated) |

#### Golden semantic movement analysis
Independently rerunning golden generation confirmed byte-for-byte fidelity with the candidate commit:
- 38 golden files in `internal/cli/testdata/golden/` contain strictly one change: `"schema_version":"1.67"` → `"schema_version":"1.68"`.
- [`schema_json.golden`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/cli/testdata/golden/schema_json.golden): Version bumped to 1.68 and gained the `"revision_policy"` object with expected values.
- [`schema_jsonschema.golden`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.7o1u2U/internal/cli/testdata/golden/schema_jsonschema.golden): `$id` qualified with `/1.68`, added root `x-taskflow-*` extras, and added `$defs.SchemaRevisionPolicy`. No existing envelope definitions suffered deletions or key renames.

---

### 3. Live candidate probing (disposable planning space)

Probes were compiled and executed against a temporary disposable repository (`/var/folders/16/.../tskflw-probe.wpeO3Z`):

1. **Human self-description (`tskflwctl schema --no-color`):**
   ```text
   tskflwctl schema v1.68

   JSON revision: monotonic-revision · additive · classified since 1.68
   Scope: all-json-output · generated schema: typed-envelopes
   ```
2. **Machine self-description (`tskflwctl schema --json`):**
   ```json
   {
     "scheme": "monotonic-revision",
     "scope": "all-json-output",
     "default_compatibility": "additive",
     "current_compatibility": "additive",
     "classified_since": "1.68",
     "reader_expectation": "ignore-unknown-object-fields",
     "generated_json_schema_for": "typed-envelopes",
     "json_schema_validation": "exact-revision"
   }
   ```
3. **JSON Schema self-description (`tskflwctl schema --json-schema`):**
   ```json
   {
     "$schema": "https://json-schema.org/draft/2020-12/schema",
     "$id": "https://github.com/andy-esch/taskflow/internal/wire/json-envelopes/1.68",
     "x-taskflow-schema-version": "1.68",
     "x-taskflow-revision-scheme": "monotonic-revision",
     "x-taskflow-revision-compatibility": "additive"
   }
   ```
4. **Typed envelope (`tskflwctl task new "Title" --json` failure/success):**
   `{"schema_version":"1.68","error":{"code":"validation","message":"validation failed: --epic is required"}}`
5. **Dynamic projected list (`tskflwctl task list -c id,slug,status --json`):**
   `{"schema_version":"1.68","tasks":[]}`
6. **JSON error path (`tskflwctl task show missing-id --json`):**
   `{"schema_version":"1.68","error":{"code":"not-found","message":"task \"missing-id\": not found"}}`

**Surface analysis:**
- Typed envelopes, dynamic projections, and JSON error payloads all carry `schema_version: "1.68"`.
- By architectural design (ADR-0008 §2), `revision_policy` is published exclusively by `schema --json` (the global contract self-descriptor) and is intentionally omitted from operational envelopes (such as `task list` or `task show`) to prevent token and payload bloat.

---

### 4. Independent Draft 2020-12 schema validation probe

Using the in-tree `github.com/santhosh-tekuri/jsonschema/v6` validator, an independent validation harness evaluated the generated JSON schema document:

1. **Root & `$id` parsing:** The URI `https://github.com/andy-esch/taskflow/internal/wire/json-envelopes/1.68` parses as a valid absolute identifier. Compiling definitions with fragment anchors (e.g. `$id + "#/$defs/SchemaEnvelope"`) resolves correctly.
2. **Valid typed envelope:** Emitted `SchemaEnvelope` validates cleanly against `#/$defs/SchemaEnvelope`.
3. **Exact schema validation on unknown object fields:** Injecting `"extra_future_field": "test"` into `SchemaEnvelope` causes `sch.Validate()` to fail immediately with:
   `additional properties 'extra_future_field' not allowed`.
4. **Tolerant reader decoding:** Unmarshaling the exact same payload into `wire.SchemaEnvelope` succeeds without error, correctly extracting `schema_version: "1.68"` and the full `revision_policy`.

This confirms ADR-0008's distinction: the generated schema enforces **exact-revision** validation (`additionalProperties: false`), whereas semantic forward compatibility relies on **tolerant reader** decoding (`ignore-unknown-object-fields`).

---

### 5. Future change walkthroughs (author obligations)

1. **Future Additive Change (e.g., adding an optional field to `TaskJSON`):**
   - Author modifies `internal/wire/dto.go`.
   - Author bumps `wire.SchemaVersion` to `"1.69"`.
   - Author appends `// 1.69: ADDITIVE — added field ...` in `internal/wire/wire.go`.
   - Author keeps `wire.SchemaRevisionCompatibility = "additive"`.
   - Running `go test ./...` catches missing doc comments (`TestSchemaComments_NotStale`) and golden drift (`TestGoldens`).
   - Author runs `go test ./internal/cli -update` and regenerates comments with `schemacomments`.
   - Review confirms the classification is truthful.
2. **Future Not-Additive Change (e.g., renaming a field or altering a closed vocabulary):**
   - Author modifies wire struct.
   - Author bumps `wire.SchemaVersion` to `"1.69"`.
   - Author appends `// 1.69: NOT ADDITIVE — renamed field ...` in `internal/wire/wire.go`.
   - Author updates `wire.SchemaRevisionCompatibility = "not-additive"`.
   - If author leaves compatibility as `"additive"`, `TestSchemaVersionChangelogIsAscending` fails:
     `SchemaRevisionCompatibility = "additive", current changelog classification = "NOT ADDITIVE"`.
   - Author updates goldens: `schema_json.golden` and `schema_jsonschema.golden` visibly record `"not-additive"` in the pull request diff.

---

### 6. Mutation probes & hostile angle results

All 6 required mutation probes and an architectural coordinated mutation were executed and restored:

| Probe | Mutation | Target / Test | Observed Result |
| :--- | :--- | :--- | :--- |
| **1. Missing classification** | Removed `ADDITIVE —` from `// 1.68:` | `TestSchemaVersionChangelogIsAscending` | **FAIL**: `SchemaVersion changelog 1.68 has no compatibility classification` |
| **2. Unknown marker** | Changed `ADDITIVE` to `BREAKING` | `TestSchemaVersionChangelogIsAscending` | **FAIL**: `unknown compatibility classification "BREAKING"` |
| **3. Disagreeing constant** | Set `SchemaRevisionCompatibility = "not-additive"` | `TestSchemaVersionChangelogIsAscending` | **FAIL**: `SchemaRevisionCompatibility = "not-additive", current changelog classification = "ADDITIVE"` |
| **4. Weaken boundary** | Set `SchemaRevisionClassificationSince = "1.67"` | `TestSchemaRevisionPolicyBoundary` | **FAIL**: `ADR-0008 classification boundary changed to "1.67"` |
| **5. Drop policy stamping** | Commented `c.RevisionPolicy = ...` in `ToSchemaEnvelope` | `TestToSchemaEnvelopeStampsRevisionPolicy` | **FAIL**: `schema envelope policy = {}, want current wire policy` |
| **5b. Coordinated probe** | Commented `ToSchemaEnvelope` stamping, ran CLI tests | `go test ./internal/cli` | **PASS (Accidental Preservation)**: Demonstrates CLI `runSchemaContract` eagerly populates policy, masking wire regression unless caught by unit test. |
| **6a. De-version $id** | Removed `/" + SchemaVersion` from `s.ID` | `TestSchema_JSONSchema` | **FAIL**: `schema identity should carry revision "1.68"` |
| **6b. Remove root annotation** | Removed `x-taskflow-revision-compatibility` | `TestSchema_JSONSchema` | **FAIL**: `schema revision policy drifted: scheme="monotonic-revision" compatibility=""` |

#### Additional Hostile Attacks:
- **Punctuation variants:** Tested replacing the em dash `—` with an ASCII hyphen `-`. The classification parser fails closed (`has no compatibility classification`), preventing non-standard delimiters from slipping through (see Finding L3).
- **Duplicate entries:** Tested inserting two `// 1.68:` changelog entries. `validateSchemaChangelog` caught the jump (`1.68 follows 1.68`).
- **Major revision transitions:** Verified that major version bumps (e.g. `1.68` → `2.0`) are accepted contiguously if minor resets to 0 (`TestValidateSchemaChangelogAllowsMajorTransition`), while illegal skips (e.g. `2.1`) are rejected.
- **Agent ergonomics:** Confirmed all core entities publish durable IDs (`id` or `task_id`) alongside human slugs; error payloads include rich typed recovery objects (`Filesystem`, `TaskLifecycleRecoveryJSON`, etc.).

---

### Findings (kept open for implementation-owner triage)

#### L1. `SchemaHuman` renders empty revision policy if passed an unstamped `SchemaContract` · **Status:** fixed

**File:** `internal/cli/render/schema_render.go:42-47` | **Component:** cli / render
**Severity:** low · **Effort:** XS · **Urgency:** eventually

`wire.ToSchemaEnvelope(c)` protects the machine JSON output by explicitly stamping `c.RevisionPolicy = CurrentSchemaRevisionPolicy()`. However, `render.SchemaHuman(w, st, c)` does not stamp or validate `c.RevisionPolicy`, reading directly from the caller-supplied `c.RevisionPolicy` fields.

If an alternate primary adapter or command constructor constructs `c := render.SchemaContract{...}` without explicitly calling `wire.CurrentSchemaRevisionPolicy()`, `SchemaJSON` will succeed and output complete policy metadata, but `SchemaHuman` will emit blank strings for the policy:
```text
JSON revision: <empty> · <empty> · classified since <empty>
Scope: <empty> · generated schema: <empty>
```

Furthermore, because `internal/cli/schema.go:104` eagerly populates `c.RevisionPolicy`, all existing CLI integration and golden tests pass even if `ToSchemaEnvelope`'s stamping logic is disabled (as demonstrated in Coordinated Mutation Probe 5b).

**Recommendation:** In `render.SchemaHuman`, fall back to `wire.CurrentSchemaRevisionPolicy()` when `c.RevisionPolicy == (wire.SchemaRevisionPolicy{})`, or provide an exported helper / constructor in `render` that normalizes `SchemaContract` prior to rendering. Belongs in this task or an immediate polish follow-up.

---

**Resolution:** Resolved through the shared wire.NormalizeSchemaContract
boundary used by both human and machine renderers; CLI assembly no longer
supplies a duplicate policy value, and hostile empty/stale/contradictory inputs
are covered.

#### L2. `internal/wire` lacks package-level assertions for three `SchemaRevision` policy constants · **Status:** fixed

**File:** `internal/wire/wire_changelog_test.go:214-224` | **Component:** wire / tests
**Severity:** low · **Effort:** XS · **Urgency:** eventually

`TestSchemaRevisionPolicyBoundary` in `internal/wire/wire_changelog_test.go` checks `SchemaRevisionScheme`, `SchemaRevisionClassificationSince`, `SchemaRevisionScope`, and `JSONSchemaScope`. However, it does not check the remaining ADR-0008 constants defined in `internal/wire/wire.go`:
- `SchemaRevisionCompatibilityDefault` (`"additive"`)
- `SchemaRevisionReaderExpectation` (`"ignore-unknown-object-fields"`)
- `JSONSchemaValidationMode` (`"exact-revision"`)

If an author inadvertently mutates or clears any of these three constants in `wire.go`, `go test ./internal/wire` completely passes. The regression is only detected downstream by `internal/cli/schema_test.go`. As `internal/wire` is intended to be the self-contained wire-contract package, its boundary test should pin all wire-owned policy constants.

**Recommendation:** Add assertions for `SchemaRevisionCompatibilityDefault`, `SchemaRevisionReaderExpectation`, and `JSONSchemaValidationMode` into `TestSchemaRevisionPolicyBoundary` in `internal/wire/wire_changelog_test.go`. Belongs in this task or an immediate polish follow-up.

---

**Resolution:** TestSchemaRevisionPolicyBoundary now pins the default
compatibility, tolerant-reader expectation, and exact-revision schema-validation
mode in the wire package alongside the existing scheme, scope, and boundary
assertions.

#### L3. Changelog classification parser error diagnostic masks em-dash delimiter requirement · **Status:** fixed

**File:** `internal/wire/wire_changelog_test.go:17, 81-83` | **Component:** wire / changelog parser
**Severity:** low · **Effort:** XS · **Urgency:** eventually

The regex `schemaChangelogClassification` requires an em dash `—` (Unicode U+2014) following the classification token:
```go
schemaChangelogClassification = regexp.MustCompile(`(?m)^// (\d+)\.(\d+): ([A-Z][A-Z ]*) —`)
```
If an author writes a standard ASCII hyphen (`// 1.69: ADDITIVE - ...`), the classification regex fails to match. The resulting test failure reports:
`SchemaVersion changelog 1.69 has no compatibility classification`

While this fails closed safely, the message misdiagnoses the error by stating that no classification exists, leading authors to suspect the keyword (`ADDITIVE`) rather than the delimiter punctuation (`—`).

**Recommendation:** Either accept standard hyphen variants (e.g. `[—–-]`) in `schemaChangelogClassification`, or add a specific check in `TestSchemaVersionChangelogIsAscending` that detects lines matching `^// \d+\.\d+: [A-Z]+` without an em dash and warns that an em dash `—` is required. Belongs in a follow-up test ergonomics task.

**Resolution:** Classification parsing now detects a recognizable marker written
with the wrong delimiter and reports the required ADDITIVE — description / NOT
ADDITIVE — description grammar explicitly; a focused ASCII-hyphen regression
pins the diagnostic.
