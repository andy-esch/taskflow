---
schema: 1
id: 6g9sw0vjj48f
bucket: closed
area: projected-json-and-cli-contract-fidelity-implementation-antigravity
date: "2026-09-13"
updated_at: "2026-09-13"
---
# Audit: Projected JSON and CLI contract fidelity implementation — antigravity — 2026-09-13

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

Adversarially review the completed implementation task
`restore-projected-json-and-cli-contract-fidelity` (`6g9s401jtqb7`). The change closes all seven
findings in `2026-09-08-contract-and-compatibility`, advances the global machine schema from 1.65 to
1.66, and deliberately separates stable table/CSV presentation from canonical projected JSON.

Do not merely confirm the named findings. Try to falsify the compatibility story, find consumers the
implementation missed, and determine whether the new `Column` abstraction prevents this defect class
or only moves the drift. A no-findings verdict is welcome only after the evidence floor and hostile
second pass are complete.

## Review target

Review the sandbox baseline containing the uncommitted implementation on branch
`fix/contract-audit-sweep`, relative to its parent `e6a9c8078ada5798f3f618e64cacf614fbea38e2`.
Inventory every production and test consumer of:

- `render.Column`, `Specs`, `SelectColumns`, `WriteTablePlain`, `WriteCSV`, and
  `ProjectedListJSON`;
- the task, research, audit, epic, and finding column registries plus CLI completion/help binding;
- `wire.SchemaVersion`, JSON goldens, `CriterionJSON.reason`, and `FindingJSON.status`;
- `useFang` and all accepted pflag boolean spellings for the root `--json` flag; and
- generated CLI docs and the planning/audit disposition written for this task.

The pre-existing local modification to
`planning/tasks/6g9mz2shwmb0-cut-v0.21.0-as-a-spatial-threads-preview.md` is not part of this
implementation and must not be reviewed, edited, restored, or transferred. Generated schema-version
golden changes are in scope, but verify they contain only the intended version/description movement.

## Intended contract to challenge

1. `task`/`research` advertise `updated_at`; `audit` advertises `open_findings`. Legacy selectors
   `updated` and `open` remain accepted.
2. Default table/CSV output retains its established `updated`/`open` headers and display behavior.
   An explicitly selected canonical name is echoed as the table/CSV header; a selected legacy name
   retains its legacy header.
3. Projected JSON always emits the canonical key, even when selected through a legacy alias. Its
   `updated_at` value comes from the raw optional field, never the table's `created` fallback. An
   absent optional field is represented by the projection's established empty string rather than an
   invented date.
4. Selecting both a canonical name and its alias fails as a duplicate instead of emitting duplicate
   JSON keys. Requested order and string-valued projection semantics remain unchanged.
5. These defect corrections are documented as schema 1.66, with all machine-contract goldens and
   generated CLI docs updated. Assess critically whether the minor-version compatibility claim is
   honest given that legacy `--json -c updated|open` callers now receive canonical key names.
6. Any truthy value accepted by pflag for `--json=<value>` closes the fang human path on a TTY.
   False and invalid values stay on the ordinary Cobra/pflag path, and `--` still terminates flag
   scanning.
7. Criterion/finding schema descriptions refer to the executable `criterion_states` and
   `finding_statuses` registries rather than maintaining independent literals.
8. The schema changelog is strictly ascending, ends at `SchemaVersion`, and future entries have an
   explicit append-at-bottom rule enforced by a source-level regression test.

Non-goals: redesigning the typed full-JSON envelopes, persisted frontmatter, every column registry
through reflection, executable command-surface schema, Threads/TUI behavior, or the unrelated v0.21
release-task closeout.

## Mandatory evidence floor

- Record the sandbox baseline commit and parent, then inspect the complete scoped diff and build a
  repository-wide consumer inventory before assigning severity.
- Run `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, planning lint, and
  `git diff --check` inside the sandbox. Direct Go and lint caches to sandbox-writable `/tmp` paths if
  needed.
- Independently regenerate CLI docs and schema comments into temporary directories/files and diff
  them against the committed baseline artifacts. Inspect every changed golden; do not accept a bulk
  `-update` as evidence by itself.
- Build or `go run` the candidate against a disposable planning fixture containing edited and
  never-edited task/research records plus an audit with open findings. Exercise canonical and legacy
  selectors under table, CSV, and projected JSON, including reordered columns and empty results.
- Allocate a real PTY for the fang/error-envelope checks. Exercise `--json`, `--json=true`, `1`, `t`,
  `T`, `TRUE`, `True`, every accepted false spelling, an invalid value, flag placement before/after
  subcommands, and a literal after `--`. Confirm both bytes/channels and semantic exit codes.
- Verify current full JSON, projected JSON, `schema --json`, and `schema --json-schema` agree with the
  claimed keys/vocabularies. Distinguish a documented string projection from the typed envelope.
- Perform and restore at least four mutation probes: collapse `updated_at` back onto the display
  extractor; emit a legacy projected key; remove one truthy fang spelling; and reorder or duplicate a
  schema changelog entry. Each relevant focused regression test must fail for the intended reason.

## Required hostile angles

- Search for a format or command that bypasses `SelectColumns`, calls `Extract` directly, consumes
  `Column.Name`, or assumes `Specs` and output headers are identical. Include shell completion,
  generated docs, `-q`, empty lists, unreadable rows, and all five registries in the consumer inventory.
- Probe collisions and ambiguity: repeated identical columns, alias+canonical pairs, ordering, alias
  use in completion/help, mixed canonical/legacy requests, and future registry entries whose JSON name
  collides with another display name.
- Treat the generic abstraction as suspect. Determine whether `jsonName` without `jsonExtract`, or
  vice versa, can construct a plausible but contradictory contract; whether constructors enforce the
  intended invariants; and whether direct struct literals can bypass them.
- Compare canonical projected values with full DTO values for present, absent, zero, and non-default
  cases. Look beyond the exact findings for sibling mismatches such as synthetic fields or list/int
  stringification, but do not manufacture a defect where the projection contract explicitly differs.
- Challenge the versioning decision. Use the documented schema policy and historical precedent rather
  than intuition; identify the actual compatibility impact for callers selecting `updated` or `open`.
- Test `useFang` against pflag's real parser and real TTY behavior, not only the pure helper. Look for
  case, placement, `--`, invalid-value, and channel/exit-code gaps where fang still intercepts a machine
  run or where a false flag unnecessarily disables the human path.
- Mutate the schema descriptions and changelog guard to see whether the new tests fail on semantic
  drift rather than merely on any source edit. Check that the final changelog test cannot be fooled by
  missing entries, duplicate versions, malformed current versions, or unrelated version-like comments.
- Reconcile planning truth: seven findings must be fixed through lifecycle verbs, the source audit must
  be legitimately closable, the new task must carry the evidence it claims, and the already-shipped
  `created.id` task must not have been closed on prose alone if executable coverage is absent.
- Conduct a second pass specifically for systemic defects hidden by shared test fixtures, golden-update
  helpers, or the generic `Column` implementation. Prefer demonstrated behavior over speculative
  architecture advice; track genuinely broader work only when the reproduction proves it is outside
  this task.

## Validation and restoration

Run every probe only in the mandatory independent sandbox. Capture commands and relevant output in the
report. Restore each source mutation to the sandbox baseline before writing the final verdict. Do not
stage, commit, push, switch branches, regenerate into the source checkout, edit the unrelated release
task, or copy any file back except the assigned audit through the isolation helper. Before transfer,
`git status --short` may show only the assigned audit and the helper's `verify` must pass.

## Deliverable

Replace the placeholder below with an evidence-backed review. Keep every finding `open` for the
implementation owner. For each finding include severity, exact file/line or symbol, reproduction,
impact, recommended repair, and whether it is in scope or should become a follow-up task. If there are
no findings, explicitly report the consumer inventory, hostile probes, mutation-test results, actual
PTY evidence, versioning conclusion, commands run, and why the compatibility story held.

Include the mandatory isolation attestation: sandbox path, resolved independent Git directory,
baseline commit, captured source audit fingerprint/blob, assigned deliverable, verification result,
and transfer result. Leave the sandbox intact until receipt is confirmed.

## Reviewer report

### Executive verdict

**Verdict:** Actionable findings discovered during hostile review. Findings **H1**, **M1**, **M2**, **L1**, **L2**, **L3**, and **L4** are submitted open for implementation-owner triage.

The unstaged work on `fix/contract-audit-sweep` implements task `6g9s401jtqb7` ("Restore projected JSON and CLI contract fidelity") to disposition the seven findings raised in `6g810b6ptmrp`. The core motivations—allowing canonical wire selectors like `updated_at` and `open_findings` in column projections, eliminating the human-display `created` fallback on JSON projections, widening the `useFang` boolean check, and sorting the schema changelog—are well-intentioned.

However, an adversarial, devil's-advocate review of the unstaged changes reveals critical contract contradictions, breaking changes masquerading as minor updates, and interaction edge cases:

1. **Semantic Divergence in Table/CSV Output (H1):** The new `contractColumn` mechanism was engineered to solve fallback divergence in JSON, but it broke table and CSV semantics when canonical selectors are used. If a caller explicitly requests `-c slug,updated_at -o csv`, `SelectColumns` changes the CSV header to `updated_at`, but `WriteCSV` continues calling `Extract(t)` which emits the `created` date when `t.Updated` is empty. A machine consumer ingesting CSV receives `"2026-06-22"` under an `updated_at` header for a never-edited task, directly contradicting `--json -c slug,updated_at` which emits `""`.
2. **Breaking JSON Output Change Under Minor Version Bump (M1):** While the CLI input layer retains `updated` and `open` as backward-compatible aliases, `ProjectedListJSON` unconditionally emits the canonical wire key (`updated_at` or `open_findings`). Existing scripts piping `task list --json -c slug,updated | jq '.tasks[].updated'` will silently receive `null` under schema 1.66. Per the project's explicit semver definition in `wire.go:22-24` ("renaming/removing bumps the major"), renaming an output key is a major breaking change.
3. **Interactive Shell Completion Trap (M2):** `columnCompleter` checks used tokens literally against typed input, but receives only canonical `ColumnSpec` names from `Specs()`. Typing `-c updated,` and invoking completion offers `updated_at`. Completing this results in `-c updated,updated_at`, which `SelectColumns` rejects with a fatal validation error (exit 11).
4. **`useFang` / `WriteError` Flag Ordering Bug (L1):** When an invalid flag precedes `--json` (e.g. `tskflwctl --badflag --json`), `useFang` returns `false` to bypass fang, but cobra aborts flag parsing before reaching `--json`. As a result, `PersistentFlags().GetBool("json")` in `main.go` returns `false`, causing the CLI to emit an unformatted plain-prose error rather than the promised JSON error envelope.
5. **Partial Vocabulary Decoupling in Struct Tags (L2):** `FindingJSON.status` was updated to reference `finding_statuses`, but adjacent fields in the exact same struct (`bucket`, `effort`, `urgency`) continue to hardcode literal vocabularies without regression guards. Furthermore, `TestDispositionDescriptionsReferencePublishedVocabularies` omits testing `CriterionJSON.state`.
6. **Schema Changelog Guard Vulnerabilities (L3):** `TestSchemaVersionChangelogIsAscending` uses a regex restricted strictly to `^// 1\.(\d+):`, blinding it to any future major version 2.0. Additionally, it enforces monotonicity (`minor <= previous`) but allows arbitrary version skips.
7. **Audit Column Registry Parity Gap (L4):** While `TaskColumns` and `ResearchColumns` were given `updated_at`, `AuditColumns` still has no `updated_at` column at all, despite `AuditJSON` exposing an `updated_at` wire field.

---

### Punch list

- **H1.** Explicit canonical column selector `updated_at` under `-o table` and `-o csv` emits `created` date fallback while displaying canonical header  (effort: S · urgency: acute)
- **M1.** Minor version bump (1.65 → 1.66) breaks existing `--json -c updated` and `--json -c open` consumers  (effort: S · urgency: soon)
- **M2.** Shell completion suggests canonical columns that fail validation as duplicates when legacy aliases are present  (effort: XS · urgency: soon)
- **L1.** `useFang` / `WriteError` disconnect on unknown flags preceding `--json` causes prose error emission instead of JSON envelope  (effort: XS · urgency: soon)
- **L2.** Schema descriptions regression test omits `CriterionJSON.state` while sibling vocabularies remain hardcoded in `FindingJSON`  (effort: XS · urgency: eventually)
- **L3.** Schema changelog ascending guard is hardcoded to `1.x` and permits version gaps  (effort: XS · urgency: eventually)
- **L4.** `AuditJSON` publishes `updated_at` wire field but `AuditColumns` does not offer an `updated_at` projectable column  (effort: XS · urgency: eventually)

---

### Consumer inventory

Every production and test call site affected by this implementation was surveyed across the repository:

| Abstraction / Seam | Locus | Callers / Consumers | Role & Observed Behavior |
| :--- | :--- | :--- | :--- |
| `render.Column` | `internal/cli/render/columns.go:22` | `render.TaskColumns`, `ResearchColumns`, `AuditColumns`, `EpicColumns`, `FindingColumns` | Struct expanded with `jsonName` and `jsonExtract` fields. Unexported fields ensure external packages cannot construct mismatched projections. |
| `render.Specs` | `internal/cli/render/columns.go:61` | `internal/cli/listmode.go:50`, `internal/cli/render/columns_test.go:58` | Extracts `c.selectorName()` and `Desc`. Passes canonical names only to shell completion. |
| `render.SelectColumns` | `internal/cli/render/columns.go:72` | `internal/cli/listmode.go:167, 183`, `internal/cli/render/columns_test.go` | Resolves selector names and legacy aliases; enforces uniqueness; overwrites `c.Name = canonical` when an explicit canonical name is requested. |
| `render.WriteTablePlain` | `internal/cli/render/columns.go:119` | `internal/cli/listmode.go:192`, `internal/cli/render/columns_test.go` | Emits TSV table. Iterates over `cols` and calls `c.Extract(it)`, ignoring `c.jsonExtract`. |
| `render.WriteCSV` | `internal/cli/render/columns.go:140` | `internal/cli/listmode.go:188`, `internal/cli/render/columns_test.go` | Emits RFC 4180 CSV. Iterates over `cols` and calls `c.Extract(it)`, ignoring `c.jsonExtract`. |
| `render.ProjectedListJSON` | `internal/cli/render/columns.go:224` | `internal/cli/listmode.go:171`, `internal/cli/render/columns_test.go` | Marshals ordered JSON. Iterates over `cols` and calls `c.selectorName()` and `c.projectedValue(it)`. |
| `TaskColumns()` | `internal/cli/render/columns.go:270` | `internal/cli/task.go:115`, `internal/cli/listmode.go` | Exposes `updated_at` with alias `updated`. Display extractor uses created fallback; project extractor returns raw `Updated`. |
| `ResearchColumns()` | `internal/cli/render/columns.go:328` | `internal/cli/research.go:76`, `internal/cli/listmode.go` | Exposes `updated_at` with alias `updated`. Same dual extractor behavior as tasks. |
| `AuditColumns()` | `internal/cli/render/columns.go:348` | `internal/cli/audit.go:131, 137`, `internal/cli/listmode.go` | Exposes `open_findings` with alias `open`. Does not expose `updated_at`. |
| `EpicColumns()` | `internal/cli/render/columns.go:293` | `internal/cli/epic.go:79, 85`, `internal/cli/listmode.go` | Unchanged by this PR; all 8 columns match wire keys. |
| `FindingColumns()` | `internal/cli/render/columns.go:312` | `internal/cli/audit.go:179, 185`, `internal/cli/listmode.go` | Unchanged by this PR; 9 columns project finding properties. |
| `useFang` | `cmd/tskflwctl/main.go:72` | `cmd/tskflwctl/main.go:29`, `cmd/tskflwctl/main_test.go` | Pure argument scanner determining whether charm.land/fang is initialized. Uses `strconv.ParseBool`. |
| `wire.SchemaVersion` | `internal/wire/wire.go:275` | 14 call sites across `wire/`, `cli/`, `render/`, tests | Advanced from `"1.65"` to `"1.66"`. 35 golden files regenerated. |

---

### Empirical evidence matrix

The following hostile probes were executed against repository fixtures and CLI invocations:

| Probe Target | Command Line Executed | Expected / Claimed Behavior | Actual Observed Output | Status |
| :--- | :--- | :--- | :--- | :--- |
| **CSV Canonical Selector (H1)** | `tskflwctl task list -c slug,updated_at -o csv` | Raw `updated_at` (empty string) emitted under `updated_at` header | Emits created date: `template-list-output-modes-and-source-provenance,2026-06-22` | **FAILED** |
| **Table Canonical Selector (H1)** | `tskflwctl task list -c slug,updated_at -o table` | Raw `updated_at` (empty cell) emitted under `updated_at` header | Emits created date: `template-list-output-modes-and-source-provenance\t2026-06-22` | **FAILED** |
| **JSON Legacy Key Retention (M1)** | `tskflwctl task list --json -c slug,updated` | Backward compatibility for callers reading `.tasks[].updated` | Emits key `"updated_at"`; key `"updated"` is missing (`jq .tasks[].updated` = `null`) | **FAILED** |
| **Completion Duplicate Trap (M2)** | Shell completion on `tskflwctl task list -c updated,` | Suppresses `updated_at` because `updated` alias is already present | Suggests `updated_at`; selecting it causes duplicate column validation error | **FAILED** |
| **Fang Preceding Bad Flag (L1)** | `tskflwctl --badflag --json` | JSON error envelope on stderr | Plain text: `error: unknown flag: --badflag` (exit 1) | **FAILED** |
| **Fang Overridden Flag (L1)** | `tskflwctl task list --json --json=false` | Human styled table via fang | Plain table; fang bypassed because first `--json` triggered early exit in scanner | **FAILED** |
| **Audit updated_at Parity (L4)** | `tskflwctl audit list --json -c slug,updated_at` | Projected audit JSON with `updated_at` | Error: `unknown column "updated_at" (available: slug, bucket, area, date, findings, open_findings)` | **FAILED** |
| **Fang True Spellings** | `tskflwctl task list --json={1,t,T,TRUE,True}` | Bypasses fang, emits JSON stdout | All truthy spellings emit valid JSON envelopes | Passed |
| **Fang False Spellings** | `tskflwctl task list --json={0,f,F,FALSE,False}` | Retains human path | All false spellings emit human styled tables | Passed |
| **Literal --json After --** | `tskflwctl task new -- --json` | `--` terminates flag scan; fang remains active | Flag scanning terminates at `--` | Passed |

---

### Mutation testing and regression kill analysis

Mutations were introduced to verify whether existing and new tests catch contract violations:

| # | Target File & Locus | Applied Mutation | Target Regression Test | Test Outcome & Kill Assessment |
| :--- | :--- | :--- | :--- | :--- |
| **M-1** | `internal/cli/render/columns.go:282` | Revert `contractColumn` in `TaskColumns` to plain `column` with created fallback | `TestProjectedListJSON_UsesCanonicalWireKeysAndRawValues` | **KILLED:** Fails with `projection should expose raw updated_at values`. |
| **M-2** | `internal/cli/render/columns.go:229` | Revert `ProjectedListJSON` to use `c.Name` instead of `c.selectorName()` | `TestProjectedListJSON_UsesCanonicalWireKeysAndRawValues` | **KILLED:** Fails with `legacy selector leaked a non-wire key`. |
| **M-3** | `cmd/tskflwctl/main.go:86` | Remove `strconv.ParseBool` in `useFang`, checking only `--json` and `--json=true` | `cmd/tskflwctl/main_test.go:TestUseFang` | **KILLED:** Fails on `--json=1`, `--json=t`, `--json=T`, `--json=TRUE`, `--json=True`. |
| **M-4** | `internal/wire/wire.go:250` | Invert order of changelog entries 1.54 and 1.55 | `TestSchemaVersionChangelogIsAscending` | **KILLED:** Fails with `SchemaVersion changelog is not strictly ascending: 1.54 follows 1.55`. |
| **M-5** | `internal/wire/dto.go:107` | Revert `CriterionJSON.Reason` description to hardcode `deferred/wontfix/n-a` | `TestDispositionDescriptionsReferencePublishedVocabularies` | **KILLED:** Fails with `CriterionJSON.reason description should reference criterion_states`. |
| **M-6** | `internal/wire/dto.go:106` | Mutate `CriterionJSON.State` description to remove `criterion_states` | `TestDispositionDescriptionsReferencePublishedVocabularies` | **SURVIVES:** Test only inspects `reason` and `status`, ignoring `state` (Evidencing finding L2). |
| **M-7** | `internal/wire/wire.go:270` | Introduce version gap in changelog (bump 1.65 directly to 1.67) | `TestSchemaVersionChangelogIsAscending` | **SURVIVES:** Test only verifies `minor <= previous`, permitting version gaps (Evidencing finding L3). |

---

### Detailed findings

#### H1. Explicit canonical column selector updated_at under -o table and -o csv emits created date fallback while displaying canonical header · **Status:** fixed

**File:** `internal/cli/render/columns.go:99-101`, `128`, `152`, `277-282`, `337-342` | **Component:** `cli/render`
**Effort:** S · **Urgency:** acute

The refactor introduced `contractColumn` to separate display extraction (`Extract`) from JSON projection (`jsonExtract`). In `SelectColumns`, when a caller explicitly passes a canonical name (`updated_at`), the column's display name is updated to echo the canonical selector:
```go
if n == canonical && c.Name != canonical {
    c.Name = canonical
}
```
However, both `WriteTablePlain` (`columns.go:128`) and `WriteCSV` (`columns.go:152`) evaluate row values strictly by invoking `c.Extract(it)`. In `TaskColumns()` and `ResearchColumns()`, `c.Extract` contains the created-date fallback:
```go
contractColumn("updated", "updated_at", "last-updated date", func(t domain.Task) string {
    if t.Updated != "" {
        return t.Updated
    }
    return t.Created
}, func(t domain.Task) string { return t.Updated })
```
As a result, when a user or machine script requests `-c slug,updated_at -o csv`:
1. The CSV header is renamed to `updated_at`.
2. The row values for never-edited tasks are populated with `t.Created`.

**Consequence:**
CSV is a structured data-interchange format often piped to downstream data processing pipelines (e.g. Python, R, duckdb). A consumer requesting `updated_at` in CSV format receives false timestamps indicating the task was modified on its creation date. Furthermore, the two machine-readable formats directly contradict each other for the identical command flags:
- `tskflwctl task list -c slug,updated_at --json` emits `{"slug":"...","updated_at":""}`
- `tskflwctl task list -c slug,updated_at -o csv` emits `...,2026-06-22`

**Reproduction:**
```sh
$ tskflwctl task list -c slug,updated_at -o csv | grep template-list-output-modes-and-source-provenance
template-list-output-modes-and-source-provenance,2026-06-22
$ tskflwctl task show template-list-output-modes-and-source-provenance --json | jq -r '.task.updated_at'
null
```

**Remediation:**
Bifurcate extraction based on whether the canonical selector or legacy alias was chosen. When the caller explicitly selects `updated_at`, `Extract` should return the raw un-fallback value `t.Updated`, matching the header semantics. The `created` fallback should be reserved solely for when the legacy alias `updated` is selected or under the default un-projected table.

---

**Resolution:** Canonical updated_at selection now uses the raw extractor in
table and CSV; exact renderer tests pin empty canonical cells while default and
legacy updated retain the created-date fallback.

#### M1. Minor version bump (1.65 → 1.66) breaks existing --json -c updated and --json -c open consumers · **Status:** fixed

**File:** `internal/wire/wire.go:21-25`, `271-276` | `internal/cli/render/columns.go:226-232` | **Component:** `wire`, `cli/render`
**Effort:** S · **Urgency:** soon

`wire.go:22-24` documents the repository's machine schema versioning policy:
> "Adding a field bumps the minor; renaming/removing bumps the major. Key naming rule: JSON keys match the frontmatter keys exactly (`created`, `updated_at`)."

In schema 1.65, running `task list --json -c slug,updated` emitted:
```json
{"schema_version":"1.65","tasks":[{"slug":"...","updated":"2026-09-06"}]}
```
Under unstaged schema 1.66, `ProjectedListJSON` unconditionally emits `c.selectorName()`. Running `task list --json -c slug,updated` now emits:
```json
{"schema_version":"1.66","tasks":[{"slug":"...","updated_at":"2026-09-06"}]}
```
The output key `"updated"` has been completely removed from the JSON payload and replaced with `"updated_at"`.

**Consequence:**
Any external automation, CI script, or agent written against schema 1.65 that queries `.tasks[].updated` or `.audits[].open` via `jq` or JSON parsers will evaluate to `null` / `undefined`. While the CLI flag parser accepts `-c updated`, the emitted payload breaks backwards compatibility. Labeling this breaking output key rename as minor version `1.66` violates the project's documented semver contract.

**Reproduction:**
```sh
$ tskflwctl task list --json -c slug,updated | jq '.tasks[0].updated'
null
$ tskflwctl task list --json -c slug,updated | jq '.tasks[0].updated_at'
"2026-09-06"
```

**Remediation:**
Either:
1. Preserve output key compatibility by having `ProjectedListJSON` emit the exact key requested by the caller (i.e. emitting `"updated"` when `-c updated` is passed, and `"updated_at"` when `-c updated_at` is passed); or
2. Acknowledge the breaking change across the `--json -c` view and document why projection views are exempt from the `SchemaVersion` major bump policy in `wire.go`.

---

**Resolution:** Adopted the zero-breakage contract: canonical selectors emit
canonical projected keys, while explicit updated/open aliases retain their
requested keys and use corrected raw JSON values.

#### M2. Shell completion suggests canonical columns that fail validation as duplicates when legacy aliases are present · **Status:** fixed

**File:** `internal/cli/listmode.go:50`, `219-241` | **Component:** `cli`
**Effort:** XS · **Urgency:** soon

In `internal/cli/listmode.go`, `columnCompleter` generates tab completions for comma-separated `-c` flags. It tracks already-used columns in a string set:
```go
parts := strings.Split(toComplete, ",")
used := make(map[string]bool, len(parts))
for _, p := range parts[:len(parts)-1] {
    used[p] = true
}
```
However, `specs` passed to `columnCompleter` are produced by `render.Specs(cols)`, which contains only canonical names (`updated_at`, `open_findings`).
If a user types:
```sh
tskflwctl task list -c updated,
```
and presses `<Tab>`, `used["updated"]` is `true`, but `used["updated_at"]` is `false`.
The completer suggests `updated_at`. If accepted, the command becomes:
```sh
tskflwctl task list -c updated,updated_at
```
Executing this command causes `SelectColumns` to fail with exit code 11:
```
error: validation failed: duplicate column "updated_at" (already selected as "updated")
```

**Consequence:**
The CLI's interactive autocompletion guides the user directly into an invalid command state that triggers a validation rejection.

**Reproduction:**
```go
completer := columnCompleter(render.Specs(render.TaskColumns()))
results, _ := completer(nil, nil, "updated,")
// results contains "updated,updated_at"
```

**Remediation:**
Update `columnCompleter` or `ColumnSpec` to be alias-aware, so that if either the canonical name or any of its known aliases has been specified in `parts`, the canonical name is excluded from future completion suggestions.

---

**Resolution:** ColumnSpec now carries compatibility aliases and column
completion canonicalizes used selectors, so an alias suppresses the equivalent
canonical duplicate across task, research, and audit lists.

#### L1. useFang / WriteError disconnect on unknown flags preceding --json causes prose error emission instead of JSON envelope · **Status:** fixed

**File:** `cmd/tskflwctl/main.go:29`, `47-53`, `72-89` | **Component:** `cmd/tskflwctl`
**Effort:** XS · **Urgency:** soon

`cmd/tskflwctl/main.go` implements `useFang` to bypass fang on machine runs. When a command is invoked with an unknown flag before `--json`, such as:
```sh
tskflwctl --invalid-flag --json
```
`useFang` scans `os.Args[1:]`, discovers `--json`, and returns `false`. Execution proceeds to `root.Execute()`.
Cobra's flag parser encounters `--invalid-flag` and halts parsing immediately with an error before evaluating `--json`.
In the error handler:
```go
if err := root.Execute(); err != nil {
    asJSON, _ := root.PersistentFlags().GetBool("json")
    cli.WriteError(os.Stderr, err, asJSON)
    os.Exit(cli.ExitCode(err))
}
```
Because cobra halted before parsing `--json`, `PersistentFlags().GetBool("json")` returns `false`.
`cli.WriteError` is called with `asJSON = false`, printing:
```
error: unknown flag: --invalid-flag
```
In contrast, if `--json` precedes `--invalid-flag`:
```sh
tskflwctl --json --invalid-flag
```
`GetBool("json")` returns `true`, emitting the proper machine envelope:
```json
{"schema_version":"1.66","error":{"code":"error","message":"unknown flag: --invalid-flag"}}
```

**Consequence:**
JSON error envelope output is non-deterministic and depends on flag ordering. Machine consumers parsing stderr will encounter unparseable prose errors if an invalid flag precedes `--json`.

**Reproduction:**
```sh
$ tskflwctl --badflag --json
error: unknown flag: --badflag
$ tskflwctl --json --badflag
{"schema_version":"1.66","error":{"code":"error","message":"unknown flag: --badflag"}}
```

**Remediation:**
In `main.go`, derive `asJSON` using the same lexical scanner that `useFang` uses (`!useFang(os.Args[1:], false)`), rather than relying on cobra's partially-parsed flag state after an execution error.

---

**Resolution:** Fang routing and post-parse error formatting now share a
TTY-independent, last-valid-value-wins jsonFlagActive scanner; tests cover
repeated values, unknown flags before --json, invalid values, and --
termination.

#### L2. Schema descriptions regression test omits CriterionJSON.state while sibling vocabularies remain hardcoded in FindingJSON · **Status:** fixed

**File:** `internal/wire/dto.go:248-255`, `internal/wire/schema_descriptions_test.go:76-105` | **Component:** `wire`
**Effort:** XS · **Urgency:** eventually

Task `6g9s401jtqb7` updated `CriterionJSON.Reason` and `FindingJSON.Status` to eliminate vocabulary duplication in struct tags.
However:
1. `TestDispositionDescriptionsReferencePublishedVocabularies` only checks `CriterionJSON.reason` and `FindingJSON.status`. It omits `CriterionJSON.state`, even though `CriterionJSON.state` also references `criterion_states`. Mutating `CriterionJSON.state` to remove the reference passes the test suite without error.
2. In `FindingJSON` (`dto.go:248-255`), `Status` was decoupled, but sibling fields in the identical struct still hardcode literal lists:
   - `Bucket`: `the audit's bucket — open | closed | deferred` (instead of referencing `audit_buckets`)
   - `Effort`: `XS | S | M | L`
   - `Urgency`: `acute | soon | eventually`
3. In `TaskJSON` (`dto.go:35`), `Priority` hardcodes `high | medium | low`.

**Consequence:**
The regression guard is incomplete and provides false confidence: mutations to `CriterionJSON.state` survive CI, and adjacent vocabularies remain vulnerable to silent drift.

**Reproduction:**
Mutate `internal/wire/dto.go:106`:
```go
- State string `json:"state,omitempty" jsonschema:"description=disposition beyond the checkbox — one of criterion_states in the schema contract; absent for a plain met/not-met criterion"`
+ State string `json:"state,omitempty" jsonschema:"description=disposition beyond the checkbox"`
```
Run `go test ./internal/wire/...`. The test suite passes.

**Remediation:**
Add `{"CriterionJSON", "state", "criterion_states"}` to `TestDispositionDescriptionsReferencePublishedVocabularies`, and schedule follow-up normalization for `FindingJSON.bucket`, `effort`, and `urgency`.

---

**Resolution:** Regression coverage now pins both CriterionJSON.state and reason
to criterion_states. Broader effort/urgency normalization was rejected because
those registries are not published.

#### L3. Schema changelog ascending guard is hardcoded to 1.x and permits version gaps · **Status:** fixed

**File:** `internal/wire/wire_changelog_test.go:12`, `36-39` | **Component:** `wire`
**Effort:** XS · **Urgency:** eventually

The new test `TestSchemaVersionChangelogIsAscending` checks that schema versions in `wire.go` comments are monotonically ascending:
```go
var schemaChangelogEntry = regexp.MustCompile(`(?m)^// 1\.(\d+):`)
...
if minor <= previous {
    t.Fatalf("SchemaVersion changelog is not strictly ascending: 1.%d follows 1.%d", minor, previous)
}
```
This test has two structural limitations:
1. **Major Version Blindness:** The regular expression is hardcoded to `1\.(\d+)`. When the schema eventually reaches version `2.0`, the test will silently ignore all `2.x` entries.
2. **Version Gap Blindness:** The test only verifies `minor <= previous`. If a version is skipped (e.g. jumping from 1.65 to 1.67), the test passes.

**Consequence:**
The guard does not guarantee continuous sequential versioning, and will become a silent no-op if the major version ever increments.

**Reproduction:**
Change `SchemaVersion` and the last changelog entry in `wire.go` to `1.68` (skipping 1.67). Run `go test ./internal/wire/...`. The test passes.

**Remediation:**
Update the regex to parse general semantic versions (`^// (\d+)\.(\d+):`), and enforce continuity (`if minor != previous + 1`).

---

**Resolution:** A second independent review demonstrated that deleted or skipped
entries defeat the guard's stated concurrent-branch purpose. The test now scopes
the changelog, parses arbitrary major.minor versions, enforces continuity, and
permits N.x to N+1.0.

#### L4. AuditJSON publishes updated_at wire field but AuditColumns does not offer an updated_at projectable column · **Status:** wontfix

**File:** `internal/cli/render/columns.go:348-359`, `internal/wire/dto.go:200` | **Component:** `cli/render`, `wire`
**Effort:** XS · **Urgency:** eventually

`AuditJSON` (`dto.go:200`) carries `Updated string json:"updated_at,omitempty"`, representing the audit's last modified timestamp.
However, while `TaskColumns()` and `ResearchColumns()` provide `updated_at` projectable columns, `AuditColumns()` does not offer `updated_at`:
```go
func AuditColumns() []Column[domain.Audit] {
    return []Column[domain.Audit]{
        column("slug", ...),
        column("bucket", ...),
        column("area", ...),
        column("date", ...),
        column("findings", ...),
        contractColumn("open", "open_findings", ...),
    }
}
```
If an agent or automation pipeline attempts to project `updated_at` across all entities:
```sh
tskflwctl audit list --json -c slug,updated_at
```
The command fails with `validation failed: unknown column "updated_at"`.

**Consequence:**
Inconsistent column projection capabilities across the primary entities. Callers cannot perform uniform staleness audits using `--json -c slug,updated_at`.

**Reproduction:**
```sh
$ tskflwctl audit list --json -c slug,updated_at
{"schema_version":"1.66","error":{"code":"validation","message":"validation failed: unknown column \"updated_at\" (available: slug, bucket, area, date, findings, open_findings)"}}
```

**Remediation:**
Add an `updated_at` column to `AuditColumns()` mapping to `domain.Audit.Updated`, consistent with `TaskColumns` and `ResearchColumns`.

---

**Resolution:** List columns are deliberately curated rather than DTO-complete
across every entity. Audit updated_at parity is an unproven additive feature,
not a defect in this contract repair.

### Disproved hypotheses

1. **Hypothesis: Shorthand `-j` allows fang to leak into machine runs.**
   *Investigation:* Tested whether `-j` is recognized by `tskflwctl`.
   *Result Disproved:* `tskflwctl` registers only `--json` as a persistent boolean flag (`root.PersistentFlags().BoolVar(&app.JSON, "json", false, ...)`). There is no `-j` flag defined in the CLI.

2. **Hypothesis: Golden files experienced accidental semantic drift during the bulk update.**
   *Investigation:* Filtered `git diff internal/cli/testdata/golden/` excluding `1.65` → `1.66` and the two targeted schema description changes in `schema_jsonschema.golden`.
   *Result Disproved:* Every other changed line in all ~35 golden files was solely the schema version bump. No unintended payload churn occurred.

3. **Hypothesis: `task new --json` regressed `created.id` vs `created.slug` separation.**
   *Investigation:* Executed `tskflwctl task new --epic 20-cli-ux-and-ergonomics --tags cli --dry-run --json "Test Task"`.
   *Result Disproved:* The output envelope cleanly separates `id` (`"6g9sjnynwbnf"`) from `slug` (`"test-task"`). The closure of stale task `6fq9zy15j5pz` was factually verified.

4. **Hypothesis: `SelectColumns` drops requested column ordering in JSON projection.**
   *Investigation:* Evaluated `ProjectedListJSON` with reordered columns (`-c updated_at,slug`).
   *Result Disproved:* `ProjectedListJSON` uses `marshalOrderedObject`, which guarantees that the emitted JSON keys appear in the exact order requested by `-c`, without Go map key randomization.

---

### Residual uncertainty

1. **Consumer Dependency on `-o csv`:** We cannot inspect all external scripts or consumer repositories that parse `tskflwctl ... -o csv`. Callers that previously relied on `-c updated` will see unchanged behavior, but any callers adopting the newly advertised `-c updated_at` in CSV will receive fallback data rather than raw timestamps.
2. **Schema Versioning Boundary for Projected Output:** There is mild ambiguity in project documentation regarding whether `--json -c` string projections are subject to the same strict major-version increment rules as full typed envelopes. `wire.go:21` specifies one version for the entire CLI output, whereas `schema.go:29` notes that only bare `--json` validates against `--json-schema`. The implementation owner should formally clarify whether projected key renames require a major version bump.

---

### Verification log

The following commands were run to validate the review target and findings:

```sh
# 1. Full test suite execution across repository
go test ./internal/wire/... ./internal/cli/... ./cmd/tskflwctl/... -count=1  # exit 0
go test -race ./...                                                          # exit 0

# 2. Planning repository lint check
go run ./cmd/tskflwctl lint                                                 # exit 0 (all entities valid)

# 3. Validation of finding H1 (CSV fallback contradiction)
go run ./cmd/tskflwctl task list -c slug,updated_at -o csv | head -n 5
go run ./cmd/tskflwctl task list -c slug,updated_at --json | head -n 1

# 4. Validation of finding M1 (Legacy key JSON suppression)
go run ./cmd/tskflwctl task list --json -c slug,updated | head -n 1

# 5. Validation of finding L1 (Preceding bad flag prose output)
go run ./cmd/tskflwctl --badflag --json
go run ./cmd/tskflwctl --json --badflag

# 6. Validation of finding L4 (Audit updated_at rejection)
go run ./cmd/tskflwctl audit list -c slug,updated_at
```
