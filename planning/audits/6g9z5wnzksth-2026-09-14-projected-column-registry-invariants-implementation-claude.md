---
schema: 1
id: 6g9z5wnzksth
bucket: closed
area: projected-column-registry-invariants-implementation-claude
date: "2026-09-14"
updated_at: "2026-09-14"
---
# Audit: Projected-column registry invariants implementation — claude — 2026-09-14

> Reviewer assignment: claude. This document is the review brief and the only file the reviewer should update.
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

Adversarially review the implementation of `enforce-projected-column-registry-invariants`
(`6g9txf8x9m70`). This is a small but shared CLI/render foundation introduced after Claude's L4
finding in the projected-output contract review. Try to prove that the new guard can still let a
plausible registry declaration silently shadow a selector, lie under a canonical header, or drift
from the full wire envelope.

Do not merely confirm the three named mutation probes. Challenge whether the construction and
validation boundaries are correctly placed, whether their panic/error policy is safe, and whether
the registry-wide test is genuinely independent of the implementation it claims to verify. A clean
verdict is welcome only after both evidence passes are complete.

## Review target

Review the live implementation snapshot on branch `fix/projected-column-registry-invariants`
relative to its parent `27e0efc`. The scoped implementation and planning changes are:

- `internal/cli/render/columns.go`;
- `internal/cli/render/columns_contract_test.go`; and
- `planning/tasks/6g9txf8x9m70-enforce-projected-column-registry-invariants.md`.

Inventory every production consumer of `Column`, `columnProjection`, `column`, `contractColumn`,
`columnRegistry`, `validateColumnRegistry`, `Specs`, `SelectColumns`, `WriteTablePlain`, `WriteCSV`,
and `ProjectedListJSON`. Include the task, epic, audit, research, and finding registries; list-mode
help and completion; `-q`; and any package-local direct struct construction. Confirm that no other
worktree changes have been silently folded into the target.

## Intended contract to challenge

1. A canonical selector override and its raw projector are one private `columnProjection` value;
   `contractColumn` refuses either half missing.
2. Every official registry validates at construction. `Specs` repeats fail-fast validation before
   help/completion derive a menu, while `SelectColumns` returns a diagnostic before constructing a
   selector map. Duplicate display names and every canonical/legacy cross-collision are rejected.
3. Static malformed registries are programmer errors and may panic during command construction;
   malformed slices passed directly to `SelectColumns` return an error. Ordinary unknown and
   duplicate user selections keep their existing validation behavior.
4. Canonically selected table, CSV, and projected-JSON cells agree with the corresponding scalar in
   the actual full JSON envelope for populated, absent, zero, and created-but-never-edited records.
   Research `tags` is an explicit comma-joined string-view exception; finding `ref` is an explicit
   synthetic `audit:code` exception.
5. Legacy `updated`/`open` selection behavior and keys, canonical selector names, default table/CSV
   output, requested order, CSV injection neutralization, full JSON, and completion remain unchanged.
6. This is an internal invariant hardening with no machine-output shape change, so schema 1.66 and
   generated docs/goldens do not move.

Non-goals: deriving registries through reflection, requiring DTO/list parity, exposing every DTO
field as a column, redesigning projected JSON as typed full JSON, or changing the schema-1.66
canonical/legacy compatibility decision.

## Mandatory evidence floor

- Record the isolated sandbox baseline and parent, inspect the complete scoped diff, and build the
  repository-wide consumer inventory before assigning severity.
- Run `go test -race ./...`, `golangci-lint run ./...`, `go mod tidy -diff`, planning lint, and
  `git diff --check` in the sandbox. Independently regenerate CLI docs and schema comments to
  temporary outputs and prove the committed artifacts remain current.
- Exercise every official registry through construction, `Specs`, canonical selection, legacy
  selection where present, table, CSV, projected JSON, and full JSON. Include empty lists and values
  beginning with spreadsheet formula characters.
- Run command-level task/research/audit probes for canonical and legacy selectors, including a
  created-but-never-edited record, duplicate canonical+alias selection, help, and shell completion.
- Execute and restore at least these mutations, requiring the focused test named by the
  implementation to fail for the intended reason: duplicate an official selector; remove a raw
  projector; remove the canonical table/CSV extractor switch; bypass validation in `Specs`; and
  bypass validation in `SelectColumns` while creating a canonical/alias collision.
- For every column in all five registries, compare the expected full-envelope field with the tested
  projection and verify every omission or non-scalar mapping is explicitly justified. Do not accept
  the helper's own string conversion as proof without inspecting the real encoded bytes.

## Required hostile angles

- Look for collisions the `seen` map misses: canonical-to-canonical, alias-to-alias,
  canonical-to-another-display-name, repeated names inside one selected list, empty or whitespace
  names, and a canonical name identical to its display name.
- Treat the panic boundary as suspect. Determine whether any malformed state can come from user or
  repository data, whether `Specs` can panic during otherwise valid startup, and whether returning an
  error from `SelectColumns` remains reachable for production registries.
- Try direct package-local `Column` literals and mutated slices. Determine whether the nested value
  actually closes the half-configured contract defect or merely makes it less obvious.
- Challenge test independence. Look for a shared mapper, extractor, formatter, omission rule, or
  exception table that lets coordinated drift keep both the full envelope and projection assertion
  green. Mutation-test the actual failure mode rather than only helper branches.
- Probe valid zero values and invalid domain zero values separately. Do not demand equality for a
  lint-invalid task solely because both formats stringify it differently, but require that any such
  boundary is named rather than accidentally skipped.
- Check selection reuse and copying: calling `SelectColumns` more than once, passing an already
  selected slice, canonical and legacy ordering, and ensuring `selectedName` cannot contaminate the
  registry returned to another caller.
- Verify all five official registry constructors use `columnRegistry` and no completion/help path
  can derive from an unvalidated registry. Search beyond the changed files before concluding this.
- Conduct a second pass for systemic issues: a future sixth registry, a new alias-bearing column, a
  non-string wire scalar, or a direct call site should fail loudly or have an explicit extension
  point. Prefer a demonstrated defect over speculative redesign advice.

## Validation and restoration

Perform all probes only inside the mandatory independent sandbox injected above. Restore every
mutation and scratch fixture to the sandbox baseline before writing the verdict. Do not stage,
commit, push, switch branches, edit the source checkout, or copy back anything except this assigned
audit through the isolation helper. Before transfer, `git status --short` may show only the assigned
audit and the helper's `verify` must pass.

## Deliverable

Replace the placeholder below with an evidence-backed review. Keep every finding `open` for owner
triage. For each finding include severity, exact file/line or symbol, reproduction, impact, minimum
repair, and whether it is in scope or belongs in a follow-up task. If there are no findings, record
the consumer inventory, mutation outcomes, command-level behavior, exception audit, validation
commands, and why both the guard and test independence held.

Include the mandatory isolation attestation: sandbox path, resolved independent Git directory,
baseline commit and parent, captured source fingerprint/blob, assigned deliverable, verification
result, and transfer result. Leave the sandbox intact until receipt is confirmed.

## Reviewer report

Reviewer: claude (Opus 5), 2026-09-14. Both passes are complete: the brief checklist first, then a
systemic second pass. All findings are left `open` for owner triage.

### Verdict

**Not clean. There is 1 medium and 3 low findings. None is a runtime regression.** The production
guard does what the task says for every official registry. All five registries validate when they
are built. Every real selector collision I tried is rejected. The five mandated mutations are all
killed. Across a 53-command matrix, the branch output is byte-identical to the parent `27e0efc`.

The weak spot is the claim that the new registry-wide test is independent. Its full-wire check
**fails open** in two ways:

- A column that no fixture populates skips the wire comparison.
- An exception column never looks at the wire at all.

Two plausible drifts survive the entire repository suite (M1). Two coordinated mutations of the
`SelectColumns` fail-closed boundary also survive all `./internal/cli/...` tests. One of them brings
back silent shadowing for known selectors (L1). The validator accepts a self-named contract column
that puts a display fallback under a canonical header; today only fixtures catch it (L2). The
lint-invalid zero-tier boundary is skipped without being named (L3). For each finding, a minimum
repair was applied in the sandbox, shown to kill its mutation, and then restored.

### Isolation attestation

| Item | Value |
|---|---|
| Sandbox path | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.27QgAf` |
| Resolved Git directory | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.27QgAf/.git` (in-tree, independent `--no-hardlinks` clone; no alternates, single worktree) |
| Baseline commit | `1b157d4dc63f311e1b58ce2451b4453a1334209d` ("chore: capture isolated review baseline") |
| Parent | `27e0efc457ae394ad76b9b7b3b2c5d095c5324d6` |
| Source root | `/Users/andyeschbacher/git/andy-esch/taskflow` |
| Source deliverable blob | `cf9480a371762d812425f864a75d18ff120a897a` |
| Source fingerprint | `fe62749a890a59f2a0d23613c83131335540d1b7` |
| Deliverable | `planning/audits/6g9z5wnzksth-2026-09-14-projected-column-registry-invariants-implementation-claude.md` |
| `verify` result | See the verbatim block below. It was run after this report was written. |
| `transfer` result | This cannot be recorded inside the file it copies. The helper allows exactly one transfer: a second attempt is refused because the source blob no longer matches (`scripts/isolated-review-workspace.sh:246-247`). The transfer attestation is therefore given verbatim in the reviewer's handoff message to the owner. |

Only one command ran in the shared checkout: the helper's `create`. Builds, the parent binary
(`git archive 27e0efc` from the sandbox's own object store), regenerated docs, and the scratch
planning repository all lived in a session-private scratch directory outside both checkouts. That
does not follow the letter of "scratch fixtures inside `$SANDBOX`", so it is disclosed here. Nothing
was written to `$SOURCE_ROOT`. Every mutation and probe test was restored with `git checkout --`
inside the sandbox, or deleted. `git status --short` was empty before this report was written.

Verify output (run after writing this report):

```
sandbox_path=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.27QgAf
git_dir=/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.27QgAf/.git
baseline_commit=1b157d4dc63f311e1b58ce2451b4453a1334209d
source_blob=cf9480a371762d812425f864a75d18ff120a897a
source_fingerprint=fe62749a890a59f2a0d23613c83131335540d1b7
deliverable=planning/audits/6g9z5wnzksth-2026-09-14-projected-column-registry-invariants-implementation-claude.md
deliverable_changed=true
transfer=pending
```

### Scope confirmation

`git show --stat 1b157d4` shows exactly four paths. `git diff --numstat 27e0efc HEAD` gives their
line counts:

- `internal/cli/render/columns.go`: +93/−21
- `internal/cli/render/columns_contract_test.go`: +324 (new)
- the task file: +17/−8
- this audit brief: +189 (new)

No other worktree change was folded in. The
helper's before/after fingerprints matched when the sandbox was created. Schema `1.66` is unchanged:
`schema --json` reports `"schema_version":"1.66"`, and nothing under `internal/wire` is in the diff.

### Consumer inventory (verified by `grep -rn` across `internal/` and `cmd/`)

- **Construction.** Only package-local helpers build columns:
  - `column` (`internal/cli/render/columns.go:40-42`)
  - `contractColumn` (`columns.go:44-52`), which panics on an empty selector or nil projector
  - `columnRegistry` (`columns.go:136-139`), which calls `mustValidateColumnRegistry`

  There are no production `Column[...]{}` literals outside those helpers. The only literals are in
  `columns_contract_test.go:39,52,57-59,65-67` and the pre-existing
  `columns_test.go:TestWriteCSV_NeutralizesFormulaInjection`. Nothing writes to `projection.name` or
  `projection.extract` after construction; the grep for assignments returned only the `==`
  comparisons at `columns.go:101,104`.
- **The five official registries all use `columnRegistry`:**
  - `TaskColumns` `columns.go:368-387`
  - `EpicColumns` `:391-405`
  - `FindingColumns` `:410-422`
  - `ResearchColumns` `:426-443`
  - `AuditColumns` `:446-457`

  Contract columns exist only in task (`updated`→`updated_at` `:375-380`), research (`:435-440`) and
  audit (`open`→`open_findings` `:453-455`).
- **Help and completion via `Specs`.** Each list command calls `lm.bind(cmd, render.Specs(render.XColumns()))`:
  - `internal/cli/task.go:232`
  - `epic.go:295`
  - `audit.go:137` (audits) and `audit.go:185` (findings)
  - `research.go:270`

  `listMode.bind` (`internal/cli/listmode.go:50-60`) derives the `-c` help text (`specNames`,
  `:256-262`) and `columnCompleter` (`:220-253`) from those specs. `NewRootCmd`
  (`internal/cli/root.go:243`) builds every command eagerly, so `Specs` runs on every invocation.
- **Selection and rendering via `renderList`** (`listmode.go:156-202`).
  - Called from `task.go:226`, `epic.go:289`, `audit.go:131`, `audit.go:179` and `research.go:264`,
    each with a fresh `XColumns()` slice.
  - `--json -c` → `SelectColumns` (`:168`) → `ProjectedListJSON` (`:172`).
  - `-o table|csv` → `SelectColumns` (`:184`) → `WriteCSV` (`:189`) or `WriteTablePlain` (`:193`).
  - `-q`/`-o name` reads `cols[0].Extract` directly (`:180`), with no `SelectColumns` call. That
    slice is still validated by `columnRegistry`.
  - Bare `--json` goes to `TasksJSON`/`EpicsJSON`/`AuditsJSON`/`ResearchJSON`/`FindingsJSON`
    (`internal/cli/render/render.go:60,569,687,712,853`), which wrap `wire.To*Envelope`.
- **No other consumers.** There is no TUI, config, or thread use of `Column`, `Specs` or
  `SelectColumns`. The only `byColumn` hit is an unrelated local in
  `internal/tui/thread_spatial.go:851`.

### Validation commands (sandbox)

| Command | Result |
|---|---|
| `go test -race -count=1 ./...` (go1.26.6 darwin/arm64) | 25 `ok` packages, no `FAIL`, exit 0 |
| `golangci-lint run ./...` | `0 issues.` exit 0 |
| `go mod tidy -diff` | empty, exit 0 |
| `git diff --check 27e0efc HEAD` / `git diff --check` | exit 0 / exit 0 |
| `go run ./internal/tools/docgen -out <tmp>` + `diff -r <tmp> docs/cli` | identical (95 files) |
| `go run ./internal/tools/schemacomments -out <tmp>` + `cmp` | identical (264 comments) |
| `tskflwctl lint` (sandbox tree, branch binary) | `✔ all planning entities and dependency links pass lint` |

### Mutation outcomes

Each mutation was applied, run against the focused tests below (or the named wider set), then
restored.

- `RC` = `TestColumnRegistriesRejectContradictoryDeclarations`
- `CC` = `TestContractColumnRequiresCanonicalSelectorAndRawProjector`
- `FW` = `TestColumnRegistriesMatchFullWireValues`

| # | Mutation | Outcome |
|---|---|---|
| 1 | Official cross-collision: task `revisit_at` renamed `updated_at` | **Killed.** `RC/task` panics with `selector "updated_at" for column "updated_at" at index 7 collides with column "updated" at index 5` |
| 2 | Research contract column raw projector → `nil` | **Killed.** `RC/research` panics with `render: a contract column requires both a canonical selector and raw projector` |
| 3 | Remove canonical extractor switch (`columns.go:194-196`) | **Killed.** `FW/task/created_but_never_edited` and `FW/research/created_but_never_edited` report `canonical table value for "updated_at" = "2026-09-01", full wire value = ""` (table and CSV) |
| 4 | Remove validation in `Specs` (`columns.go:143`) | **Killed.** All 5 invalid-table subtests report `panic = <nil>, want text …` |
| 5 | Remove validation in `SelectColumns` (`columns.go:158-160`), with the table's canonical/display collision | **Killed.** All 5 subtests report `SelectColumns() did not fail closed …: unknown column "anything" (available: canonical, canonical)` |
| 6 | Coordinated: validate only when a requested name is unknown | **Survived** all of `go test ./internal/cli/...`. Probe: registry `[contractColumn("legacy","canonical",A), column("canonical",B)]` with `-c canonical` gives `selected "canonical" extracting "B(other column)", err=<nil>`. Baseline gives `invalid column registry: selector "canonical" …`. → **L1** |
| 7 | Coordinated: validation moved below the `len(names)==0` return | **Survived** all of `go test ./internal/cli/...` → **L1** |
| 8 | Task contract column self-named: `contractColumn("updated_at","updated_at",…)` | Validator **accepts** it. Caught only by fixtures: `FW/task/created_but_never_edited` plus 13 other render/cli tests (goldens, completion, table/CSV). → **L2** |
| 9 | Add task columns `depends_on` (unsorted join), `tags` (join) and `started` (no wire key) | `FW` **passes**. Real drift: projected `"depends_on":"6gb000000002,6gb000000001"` vs full `["6gb000000001","6gb000000002"]`; `"started":"2026-09-10"` with no wire key. → **M1** |
| 10 | Wire stops emitting research tags (`internal/wire/dto.go:231`, drop `Tags: r.Tags`) | `FW` **passes**, and **`go test ./...` passes repository-wide.** → **M1** |
| 11 | `TaskColumns` returns a raw slice literal (bypassing `columnRegistry`) with a duplicate `status` | **Killed.** `Specs` panics at command construction (e.g. `TestAuditLint_DirtyExits11` panics); render tests fail |
| 12 | Coordinated: strip `columnRegistry` and `Specs` validation, then duplicate epic `done` | **Killed.** `RC/epic` reports `official registry is invalid: … selector "done"`; `FW/epic/*` fail through the `SelectColumns` error |
| 13 | Coordinated: `contractColumn` degrades to a plain column when the projector is nil, plus research projector nil | **Killed.** `CC` reports `panic = <nil>`; `FW/research/present` reports `canonical table value for "updated" = "2026-09-14", full wire value = ""` |
| 14 | `wireScalarString` bool branch returns `"yes"` | `FW` **passes**: no registry column is a bool, so the branch never runs (see M1) |
| 15 | `wireScalarString` treats an absent key as an error | 6 `FW` subtests fail on legitimate `omitempty` absences (`priority`, epic `status`, `area`, `created`, `description`). A strict per-fixture presence check is therefore wrong; the M1 repair must aggregate per column |

### Command-level behavior

Setup:

- A scratch repository was created with `init`, `epic new`, and `task new` (twice; one task was then
  edited with `task set --priority high`).
- Two research docs were created with `research new`; one was edited with `research set --tags "@evil,cli"`.
- `audit new` created an audit, and two findings were appended by hand, including `=HYPERLINK("x")`
  and component `-cli`.
- A hand-written, lint-invalid task has no `tier`.
- The task description `=SUM(A1) formula` exercises CSV neutralization.

The branch binary (built from the baseline) and the parent binary (built from
`git archive 27e0efc`) ran the same 53 invocations. Stdout, stderr and exit codes were
**byte-identical** (397 lines). The matrix covered:

- default table/CSV/JSON/`-q` for task/research/audit/findings/epic
- canonical and legacy `-c` in table, CSV and JSON
- duplicate canonical+alias in both orders, `slug,slug`, `,slug`, and an unknown column
- the `-c` help line for all five list commands
- `__complete` for all five, including alias-used-then-canonical

Highlights:

- **Never-edited task.** `task list -o table -c slug,updated_at` gives an empty cell; `-c slug,updated`
  gives `2026-09-14` (the created fallback). `--json -c updated_at,slug` gives `"updated_at":""`;
  `--json -c slug,updated` gives `"updated":""`. The full `task list --json` omits `updated_at` for
  that task.
- **Duplicates.** `-c slug,updated,updated_at` returns `{"error":{"code":"validation","message":"validation failed: duplicate column \"updated_at\" (already selected as \"updated\")"}}` with exit 11. Research and audit (`open_findings,open`) behave the same way.
- **Formula cells.** CSV cells are neutralized: `'=SUM(A1) formula`, `'@cmd edited`,
  `"'@evil,cli"`, `"'=HYPERLINK(""x"") probe"`, `'-cli`. JSON and table output carry the raw values.
- **Help.** `-c` help lists canonical names only (`…,epic,updated_at,description,revisit_at`).
- **Completion.** `__complete task list -c updated,` omits `updated_at`, and
  `__complete research list -c updated,up` returns nothing, because the alias is recognized as used.
- **Tier 0.** The no-tier task gives `"tier":"0"` under `--json -c tier`, while the full envelope omits
  `tier`; lint reports `tier: missing` (L3).
- **Empty lists** (render probe, all five registries). Table and CSV print the header only. Projected
  JSON gives `{"schema_version":"1.66","<key>":[]}`, identical to the full envelope.

### Exception audit (every column vs the full envelope)

Values were checked against the real encoded bytes, both from `decodeProjectedRow` on `wire.EncodeJSON`
output and from command-level `--json`, not from the helper's string conversion.

- **Task** (`TaskJSON`, `internal/wire/dto.go:24-42`):
  - `slug`, `status`, `priority`, `epic`, `description`, `revisit_at`: equal strings. Absent
    `omitempty` fields map to an empty string.
  - `tier` (int, `omitempty`): equal for 1-5. **Zero diverges** and is unnamed (L3).
  - `updated`→`updated_at`: the canonical selection is raw and equal. The legacy table/CSV fallback is
    the documented compatibility behavior.
  - Unexposed, by non-goal: `id`, `effort`, `autonomy_level`, `created`, `tags`, `depends_on`.
- **Epic** (`EpicJSON` + `EpicMetaJSON`, `dto.go:171-182,310-318`):
  - `id`, `status`, `priority`, `description`: equal.
  - `done`, `total`, `deprecated`, `percent` (ints, never omitted): equal. `percent` uses the same
    domain method, `EpicSummary.Percent()` (`internal/core/service_epic.go:105-110`, 0/0→0). That
    shared method is the domain definition, not a test helper.
  - Unexposed: `open`, `liveness`, `created`, `updated_at`, `tags`.
- **Audit** (`AuditJSON`, `dto.go:194-211`):
  - `slug`, `bucket`, `area`, `date`, `findings`: equal.
  - `open`→`open_findings`: its display and projector closures are identical (`columns.go:454-455`).
  - Unexposed: `id`, `updated_at`, the three band counts, and `ready_to_close` (the only nearby bool).
- **Research** (`ResearchJSON`, `dto.go:218-225`):
  - `slug`, `created`, `description`, `id`: equal.
  - `updated`→`updated_at`: canonical is raw.
  - **`tags` is an exception.** The comma-joined view is justified, but the exception is computed from
    the domain item, not the envelope (M1).
- **Finding** (`FindingJSON`, `dto.go:246-258`):
  - `code`, `audit`, `status`, `title`, `effort`, `urgency`, `component`, `file`: equal.
  - **`ref` is an exception.** It is a synthetic `audit:code` with no wire key. The justification
    holds, but it is also computed from the item (M1).
  - Unexposed: `bucket`, `status_decoration`, `note`.

### Hostile angles checked (no finding)

- **Collision coverage.** The validator rejects:
  - canonical↔canonical
  - alias↔alias
  - canonical-of-A = alias-of-B, and the reverse
  - a plain display name equal to a later alias
  - a duplicate display name
  - `Name == ""`, a nil `Extract`, a projection with an empty name, and a projection with a nil
    extract

  Repeated names in one selection get the existing `duplicate column` error (for example `slug,slug`,
  exit 11).

  It accepts `" "`, `"a,b"`, a projection name of `" "`, and a `"slug "` twin. These cannot be
  selected, because pflag's comma split does not trim, so they are not shadows. It also accepts a
  self-named contract column (L2).
- **Panic boundary.** Registry names are pure code, and neither user nor repository data can reach
  them. `NewRootCmd` evaluates `Specs` for all five list commands, so a malformed registry panics on
  every invocation. Mutation 11 shows an unrelated command's test panicking, which makes the failure
  loud in CI. A `--json` caller would see a Go panic rather than an envelope, but only in a build
  that cannot pass tests.

  `SelectColumns`' registry error cannot be reached for official registries, because
  `columnRegistry` panics first on the same fresh slice. That error is unwrapped, so it would exit 1
  (`internal/cli/exit.go:48`) rather than 11. This is appropriate for an internal defect.
- **Nested projection.** The pointer type does not make a half-configured pair unrepresentable:
  package-local literals at `columns_contract_test.go:57-59,65-67` build one. What closes the defect
  is the validator at the three entry points (`columnRegistry`, `Specs`, `SelectColumns`).
  `WriteTablePlain`, `WriteCSV` and `ProjectedListJSON` trust their input, and every production call
  reaches them through `SelectColumns`.
- **Selection reuse.** Two `SelectColumns` calls on one registry leave its `Name`s and `selectedName`s
  untouched. Copies share the `*columnProjection` pointer, which is never mutated. Re-selecting a
  canonically selected slice by its legacy alias fails with `unknown column "updated"`. Empty
  selection returns the registry's backing array. Both of these are pre-existing behavior and not
  production paths.
- **Sixth registry and direct call sites.** Structural validation extends automatically. Mutations 11
  and 12 show that a registry skipping `columnRegistry` still fails through `Specs` and
  `SelectColumns`. Wire fidelity does not extend automatically (see M1).

### Findings

#### M1. The registry-wide full-wire test fails open for unpopulated columns and for its own exceptions · **Status:** fixed

**File:** `internal/cli/render/columns_contract_test.go:305` | **Component:** cli/render tests
**Effort:** S · **Urgency:** soon

`assertRegistryMatchesFullWire` is meant to be the independent proof that canonical projections agree
with the full envelope. Two things in the helper let wire drift through without an explicit decision:

1. **A missing key looks like an empty value.** `wireScalarString` returns `"", nil` for a key that is
   missing from the full row (`:305-307`). That is correct for an `omitempty` absence in one fixture
   (mutation 15). But nothing requires a column's key to appear in *any* fixture. So a column whose
   source the hand-written fixtures never populate is never compared with the wire. It also never hits
   the "needs an explicit string-view exception" rule (`:234-236`). Mutation 9 added three plausible
   task columns and `FW` still passed:
   - `depends_on`: the projection gives unsorted `6gb000000002,6gb000000001`, while the wire sorts it
     to `["6gb000000001","6gb000000002"]` (`dto.go:41,51`).
   - `tags`: a non-scalar that needs the same exception as research.
   - `started`: a synthetic column with no wire key.

   With one populated fixture, the same harness does fail loudly: `column "depends_on" needs an
   explicit string-view exception`. The comparator works; the missing piece is a coverage
   requirement.
2. **Exceptions never read the envelope.** The `tags` and `ref` exceptions compute `want` from the
   domain item (`:167-172`, `:185-190`, applied at `:228-233`), so they re-implement the extractor
   instead of documenting a mapping from the wire. In mutation 10 `research list --json` stopped
   emitting `tags` entirely, and `FW` plus the whole `go test ./...` suite stayed green, while the
   `tags` exception still vouched for the column.

The `bool` branch (`:319-320`), the extension point for non-string scalars, is never executed
(mutation 14). The five registry calls are listed by hand (`:117-190`), so a sixth registry gets no
wire check unless someone adds one.

**Impact:** Nothing is wrong at runtime today. Every current column is populated by a "present"
fixture, and the command matrix matches the parent. But the change's stated guarantee (L4 / task
AC 3: "documented synthetic and string-view exceptions remain explicit") holds only by fixture
coincidence. The most likely future edit, adding a column without touching fixtures, can ship a
lying canonical cell.

**Recommendation:** In scope; test-only, about 15 lines.

- Record per column whether its canonical key was present and non-null in at least one fixture's full
  row. After the fixture loop, `t.Errorf` for any column that is neither present nor an exception.
- Change `projectionException.want` to take the decoded full row: decode `row["tags"]` as `[]string`
  and join it; build `ref` from `row["audit"]` + ":" + `row["code"]`.
- Optionally add a bool fixture column, or delete the unused branch.

This sketch was applied in the sandbox. It passed at baseline, failed mutation 9 with `task column
"depends_on" never reaches the full wire in any fixture` (and the same for `tags` and `started`), and
failed mutation 10 with `canonical table value for "tags" = "cli,contract", full wire value = ""`.
It was then restored.

**Resolution:** The registry-wide test now requires every canonical column to
reach a non-null full-wire value or declare a wire-derived exception. Research
tags decode the envelope array, finding ref derives from envelope audit/code,
and the unused bool branch was removed; both reviewer mutations now fail for the
intended reason.

#### L1. The `SelectColumns` fail-closed assertion uses only an unknown selector, so lazy or skipped validation survives · **Status:** fixed

**File:** `internal/cli/render/columns_contract_test.go:77` | **Component:** cli/render tests
**Effort:** XS · **Urgency:** eventually

Contract §2 says `SelectColumns` "returns a diagnostic before constructing a selector map". The test
only calls `SelectColumns(tc.cols, []string{"anything"})`. An unknown name already errors, so the
test separates the cases by message text alone. It cannot tell "validate first" apart from "validate
on the miss path", and it never exercises the empty selection.

- **Mutation 6** moves validation into the unknown-name branch. All of `go test ./internal/cli/...`
  passes, and the actual defect comes back: `-c canonical` on a canonical/display collision silently
  returns the later column (`selected "canonical" extracting "B(other column)", err=<nil>`).
- **Mutation 7** moves validation below `if len(names) == 0 { return all, nil }`. It also survives.
  The default table would then pass an unvalidated slice to `WriteTablePlain`.

**Impact:** Low. Official registries are still protected, because `columnRegistry` panics at
construction (`columns.go:136-139`). The boundary this test claims to pin is the one meant for
non-registry slices, and it only appears to fail closed.

**Recommendation:** In scope; test-only. For each invalid registry, assert the diagnostic for `nil`,
for an unknown name, and for a declared selector such as `tc.cols[len(tc.cols)-1].selectorName()`.
Applied in the sandbox: it passes at baseline. The declared-selector case alone kills mutation 6
(`SelectColumns(["canonical"]) did not fail closed on the invalid registry: <nil>`), and the `nil`
case kills mutation 7. Restored.

**Resolution:** Each invalid registry is now exercised through empty selection,
an unknown selector, and a declared selector. This pins validation before every
early return or selector-map lookup and kills both lazy-validation mutations.

#### L2. The validator accepts a contract column whose canonical selector equals its display name · **Status:** fixed

**File:** `internal/cli/render/columns.go:109` | **Component:** cli/render
**Effort:** XS · **Urgency:** eventually

`validateColumnRegistry` records only `selectorName()` when `Name == selectorName()` (`:109-112`). A
`projection` whose `name` equals `Name` is therefore valid. Because `SelectColumns` swaps in the raw
extractor only when `c.Name != canonical` (`:192-197`), this contradictory declaration puts the
display fallback under the canonical header:

- A probe with `contractColumn("updated_at","updated_at", display, raw)` gave the default table
  `"updated_at\ndisplay:\n"` and the canonical table `"updated_at\ndisplay:\n"`, but projected JSON
  `{"updated_at":""}`.
- `Specs` advertises no alias.

This is exactly the "lie under a canonical header" the brief targets. For official registries
(mutation 8) the guard does not catch it. Only the never-edited fixture and 13 golden/table tests
fail, so a registry or column without equivalent fixtures would ship it.

**Impact:** Low. It is defense-in-depth for a mistake that the fixture layer catches today.

**Recommendation:** In scope. In `validateColumnRegistry`, reject `c.projection != nil &&
c.projection.name == c.Name` ("contract column %q names itself as its canonical selector"), and add
a matching case to the invalid table. Applied in the sandbox: `go test ./internal/cli/...`
(review probes skipped) was all `ok`, and mutation 8 then panicked at construction with that message.
Restored.

**Resolution:** Resolved with a supported variant rather than prohibition:
explicit selection of a self-named contract column now switches to its raw
projector, while the default table retains the display fallback. A focused test
pins both views and projected JSON.

#### L3. The lint-invalid zero-tier boundary is silently avoided rather than named · **Status:** fixed

**File:** `internal/cli/render/columns_contract_test.go:124` | **Component:** cli/render tests
**Effort:** XS · **Urgency:** eventually

`TaskJSON.Tier` is `omitempty` (`internal/wire/dto.go:34`), and the task `tier` column is
`fmt.Sprintf("%d", t.Tier)` (`columns.go:372`). For a task file with no `tier` (reachable from
repository data; lint reports `tier: missing`, `internal/domain/lint.go:98-99`):

- `task list --json` omits `tier`
- `task list --json -c tier` emits `"tier":"0"`
- `-o table` shows `0`

The parent binary behaves the same way, so this is not a regression. The brief accepts the
divergence but requires the boundary to be named. The fixture named "absent optional strings" sets
`Tier: 3` (`:124-126`) with no comment, and the task outcome's "zero" claim covers only the epic and
audit registries. Nothing records that zero tier was deliberately excluded.

**Impact:** Low. The divergence is invisible to a future reader. If `tier` gains an exception or a
fixture drops `Tier`, the failure will look like a regression rather than a known lint-invalid
boundary.

**Recommendation:** In scope. Add a named lint-invalid zero-tier case that asserts the known
divergence explicitly, or at least a comment on the fixture naming the excluded boundary. Changing
the `tier` column's behavior is out of scope (non-goal, schema-1.66 compatibility); if wanted, that
belongs in a follow-up task.

**Resolution:** Added a named lint-invalid tier-zero compatibility test proving
lint reports the missing tier, full JSON omits it, and projected list JSON
retains the established string value 0. Runtime behavior is intentionally
unchanged.
