---
schema: 1
id: 6gw9y271vckm
bucket: open
area: contract-and-compatibility
date: "2026-11-10"
updated_at: "2026-11-10"
---

# Audit: contract-and-compatibility — 2026-11-10

> Create findings with `audit finding new`; update them with `audit finding`.

## Findings

<!-- Example grammar; let `audit finding new` allocate and render real findings: -->

```
#### H1. <title>  · **Status:** open

**File:** <path:line> | **Component:** <component>
**Effort:** <XS|S|M|L> · **Urgency:** <acute|soon|eventually>

<what's wrong, why it matters, evidence>

**Recommendation:** <minimum fix>

**Resolution:** <how it was resolved — written by `audit finding --note`, not by hand>
```

#### H1. Deprecated task verbs print cobra's notice to stdout, breaking the --json single-document contract · **Status:** open

**File:** internal/cli/task.go:912 | **Component:** cli/output-contract
**Effort:** XS · **Urgency:** acute

`root.SetOut(out)` (internal/cli/root.go:307) points cobra's `outWriter` at the app's
stdout writer — correct for usage text, but cobra also prints the deprecation notice
through that writer: `command.go:911` calls `c.Printf`, which is `fmt.Fprint(c.OutOrStderr())`,
and `OutOrStderr` only falls back to stderr when `outWriter` is unset (cobra v1.10.2).

`deprecatedTransitionCmd` (internal/cli/task.go:912) sets `cmd.Deprecated` on `task promote`
and `task demote`, so both commands emit an English sentence on stdout ahead of their JSON
envelope. The published `commands` contract lists both with `deprecated: true, safety: mutating`,
and `SchemaCommand`'s own doc says hidden and deprecated leaves "remain present because agents
can still invoke them" — so the contract invites exactly the call that breaks.

This is the only place the output discipline leaks. Verified in the same run: error
envelopes go to stderr while the payload envelope stays on stdout (`audit close` on an
unsettled audit), and flag-parse failures leave stdout completely empty. Both keep stdout
a single parseable document; the deprecation notice does not.

`-o name` is corrupted the same way, so a shell pipeline consumes the sentence as a record.

**Recommendation:** Stop letting cobra own the notice on the injected stdout writer: either clear `cmd.Deprecated` in deprecatedTransitionCmd and emit the same sentence through the app's stderr writer in a PersistentPreRunE, or keep `Deprecated` and override the leaf's out writer to errOut so `c.Printf` lands on stderr. Add one `runRootStreams` case asserting `res.Out` parses as a single JSON document and the notice appears only in `res.Err`.

#### M1. Thread DTOs and the epic rollup trio ship in the published schema with no field descriptions · **Status:** open

**File:** internal/wire/thread.go:96 | **Component:** wire/json-schema
**Effort:** M · **Urgency:** soon

`schema.go`'s premise is that a machine consumer can drive the tool "without parsing
--help prose", and ADR-0008 publishes `$defs` descriptions as that contract. The Thread
surface — the newest and most semantically loaded one — publishes almost none.

Counted from this binary's own `schema --json-schema`: 687 of 945 `$defs` properties carry
no description. Most of those are envelope wrappers and workspace plumbing, where it is
defensible. The entity projections are not: `ThreadJSON` has 12 properties and exactly one
description (`status`); `ThreadRollupJSON` has four (`done`, `total`, `drained`,
`deprecated`) and none; `TaskGraphStateJSON` describes `eligible` but not `role`, `gate`,
`soundly_completed`, `drained`, or `inconsistent`. `internal/wire/thread.go` carries 28
`jsonschema` tags across 153 JSON fields; `dto.go` carries 90 across 99.

Which fields got a description is close to accidental. On `EpicJSON`, `open` and
`deprecated` are described only because someone wrote a Go doc comment that
`AddGoComments` happened to harvest; their siblings `total`, `done`, and `percent` have
neither tag nor comment and ship bare — and 1.11 gave `total`/`done` a non-obvious
derivation (deprecated tasks excluded) recorded only in the wire.go changelog.

**Failing scenario:** an agent reads `schema --json-schema`, sees `thread.rollup` =
`{done, total, drained, deprecated}` with no semantics, and computes progress as
`done/total`. `Drained` is strictly stronger — `dependency_graph.go:880` sets it to
`role == RoleNominallyComplete && sound` — so the agent reports a Thread 100% complete
while `drained < total`, which is precisely the state `thread complete` refuses and
`thread_projection.go:312` raises as an inconsistency. `ThreadShowHuman` prints both
numbers side by side ("3/5 done · 2/5 drained"); the published schema gives a machine
consumer no way to learn the difference exists.

**Why tests didn't catch it:** `TestJSONSchema_HasFieldDescriptions` checks a hand-written
`want` map of seven field names across TaskJSON/FindingJSON/AuditJSON. Its own comment
says it exists so the feature cannot "silently go inert" — it guards the mechanism, not
coverage, and was never extended when the Thread DTOs landed.

**Recommendation:** Add `jsonschema:"description=..."` tags to the Thread DTO family (ThreadJSON, ThreadRollupJSON, ThreadViewJSON, ThreadTaskJSON, ThreadGraphNodeJSON, TaskGraphStateJSON) and to EpicJSON's total/done/percent, saying for each what the derivation actually is — that `done` is nominal and `drained` is sound, and that epic total/done exclude deprecated tasks per 1.11. Then replace schema_descriptions_test.go's hand-kept `want` allowlist with a closure guard over the entity projection DTOs, the way registry_test.go closes the envelope-registration gap by parsing the package.

#### M2. Projected thread list reuses the typed envelope's frontier key for a count, not the member list · **Status:** open

**File:** internal/cli/render/columns.go:578 | **Component:** cli/projection-contract
**Effort:** S · **Urgency:** soon

Every other projected selector in the six list registries names the same datum as the
typed envelope key it shares a name with. `frontier` is the one exception.

Verified against this checkout:

    thread list --json            → threads[].frontier is an ARRAY of member objects
    thread list --json -c frontier → threads[].frontier is "0" — len() of that array

The other Thread columns avoid the collision by accident of nesting, not by design:
`done`/`total`/`drained`/`deprecated` live under `rollup` in the typed envelope and
`slug`/`status`/`description`/`id` under `thread`, so their flattened projected keys
collide with nothing. `graph_health`, `projection_health`, and `inconsistent` sit at the
same level in both and mean the same thing in both. Only `frontier` is published twice at
row level with two different meanings.

**Failing scenario:** an agent reads `threads[].frontier` from the typed envelope, then
switches to `-c` for token economy — the documented reason the projection exists. The key
is still present and still truthy, so nothing errors: it now holds `"3"`, and `len()` of
it is 1. A frontier of three members reads as one.

The projection contract does blunt this: `ProjectedListJSON`'s doc is explicit that rows
are a "string-valued VIEW, NOT the canonical typed envelope" and do not validate against
`schema --json-schema`. That puts a consumer on notice about *encoding*. It does not warn
that one key changes *which datum* it carries, and revision 1.72 names `frontier` among
the projected fields without saying it becomes a cardinality.

**Recommendation:** Rename the canonical selector to `frontier_count` via `contractColumn("frontier", "frontier_count", ...)` — the same mechanism already used for `updated`→`updated_at` and `open`→`open_findings` — so the legacy `frontier` spelling keeps working as a compatibility alias while the canonical projected key no longer collides with a typed key of different meaning. Declare it in the next revision entry and say what the projected value is.

#### L1. LocationIsPath is set from LocalPath != "", asserting an equality the sibling mapper checks · **Status:** open

**File:** internal/core/service_task.go:183 | **Component:** core/record-source
**Effort:** XS · **Urgency:** eventually

`RecordSource.LocationIsPath` is documented (internal/core/entity_read.go:26) as
"adapter-supplied presentation evidence that the location is already the separately
resolvable local path". Two mappers derive it from the same shaped input and disagree:

    core/dependency_source.go:228  LocalPath != "" && Location == LocalPath   ← checks it
    core/service_task.go:183       problem.LocalPath != ""                    ← asserts it

`taskGraphLoadProblemFromLoadProblem` is the compatibility bridge used when an adapter
exposes only `ReadTasks()`. Hand it a `LoadProblem{Location: "s3://bucket/tasks/x.md",
LocalPath: "/cache/x.md"}` — exactly the pathless/remote adapter shape that revisions
1.73–1.76 and 1.79 were added to support — and it returns `LocationIsPath: true`, claiming
an s3 URI *is* the local path `/cache/x.md`.

Labelled a smell rather than a defect, honestly: I could not make it produce wrong output.
Every consumer of the flag prefers `Path` when it is set (`taskGraphLocalPath`,
service_task.go:308; `taskGraphProblems`, store/entity_read.go:135), and the flag is only
wrongly true when `Path` is set, since the same mapper assigns `Path: problem.LocalPath`.
The shipped FS store sets `Location` and `LocalPath` to the identical value
(store/entity_read.go:53), so nothing in tree distinguishes them today.

What remains is a guard that mis-states its own contract, two lines from a sibling that
states it correctly, with no test pinning the two together — and the one observable
consequence, the `boolRank(LocationIsPath)` sort key in `canonicalTaskGraphLoadProblems`
(service_task.go:341), orders published `unreadable` records by it.

**Recommendation:** Make the predicate match its documented meaning: `LocationIsPath: problem.LocalPath != "" && problem.Location == problem.LocalPath`, matching dependency_source.go:228. Better still, give RecordSource a single constructor both sites call so the question is answered once.

Routine: `code-quality-audit` · lens `contract-and-compatibility` · ISO week `2026-W46`,
slot `Tue` (index 2 = (46 × 2 + 0) mod 6).

## Punch list

- H1. Deprecated task verbs print cobra's notice to stdout, breaking the --json single-document contract  (effort: XS · urgency: acute)
- M1. Thread DTOs and the epic rollup trio ship in the published schema with no field descriptions  (effort: M · urgency: soon)
- M2. Projected thread list reuses the typed envelope's frontier key for a count, not the member list  (effort: S · urgency: soon)
- L1. LocationIsPath is set from LocalPath != "", asserting an equality the sibling mapper checks  (effort: XS · urgency: eventually)

## Files audited

- **Signal**: `internal/wire/wire.go` — highest churn in the lens surface (6 commits/30d, 12/90d) and 71 non-wire files import the package; it owns `SchemaVersion`, the 90-entry changelog, and the ADR-0008 policy constants.
- **Signal**: `internal/domain/entity.go` — 2 commits/30d, and structurally the highest blast radius in the surface (350 files import `internal/domain`); it owns the entity registry whose order is declared part of the machine contract.
- **Adjacency** (from `wire.go`, across the hexagonal seam): `internal/cli/render/columns.go` — the `-c` projection layer that publishes the *other* half of the JSON contract, where the lens's "does a projection expose a field the full `--json` doesn't" question actually lands.
- **Adjacency** (from `entity.go`, across the seam): `internal/wire/schema.go` — the published self-description (`exit_codes`, `commands`, the vocabulary registries) that the domain registry feeds.
- **Random**: `internal/wire/dto.go` — the per-entity DTOs and their `jsonschema` description tags.

Read for context rather than audited: `wire_changelog_test.go`, `registry_test.go`,
`schema_descriptions_test.go`, `schema_comments_test.go`, `internal/cli/command_safety.go`,
`internal/cli/bareverb.go`, `internal/cli/harness_test.go`, `internal/domain/audit.go`,
`internal/domain/resolution.go`, `internal/core/entity_read.go`,
`internal/core/service_task.go`, `internal/core/thread_projection.go`,
`internal/core/dependency_source.go`, and cobra v1.10.2's `command.go`.

## Commands run

Build (no `just` on this runner, per the routine's documented fallback):

- `go build -o bin/tskflwctl ./cmd/tskflwctl` — OK, `v0.0.0-20261107172534-ba6e95dde14e`.

Contract probes against this checkout (all with the freshly built binary):

- `audit list --json` → 6 open audits, below the backpressure threshold of 10.
- `schema --json` → `exit_codes` (8 rows), `commands` (100 rows), 13 top-level registries.
- `schema --json-schema` → 155 `$defs`, 945 properties, 687 with no `description` (M1).
- Exit-code vocabulary checked against behavior: `task info <missing>` → 10 `not-found`;
  `task list --json -c nosuchcolumn` → 11 `validation`; `task info e` → 13 `ambiguous`;
  code 12 published `reserved` and never emitted. All match the published rows.
- Per-envelope `schema_version`: all 54 declared `*Envelope` types carry the field.
- Projection vs typed keys for all six list commands; every selector resolves to a real
  wire key except `ref` (documented as the `-o name` handle) and `frontier` (M2).
- Alias round-trip on a task with no `updated_at`: `--json -c updated` → `""`,
  `--json -c updated_at` → `""`, `-o table -c updated` → `2026-06-22` (display fallback),
  `-o table -c updated_at` → empty. Exactly what revision 1.66 declares.
- `--help` vs `docs/cli/*.md` column lists for all six list commands — identical.
- Stream purity: `audit close` on an unsettled audit → `moves` envelope on stdout (parses
  as one document), `error` envelope on stderr; `task list --json --nosuchflag` → stdout
  empty, error envelope on stderr; `task promote <task> --dry-run --json` → **stdout fails
  to parse** (H1), stderr empty.
- Golden honesty: 11 commits touching `internal/cli/testdata/golden/` in 45 days, each
  changing 4–176 lines beyond `schema_version`. Not version-string churn.

Validation, with `just` unavailable so the Justfile recipes were run directly:

- `go vet ./...` — clean.
- `./bin/tskflwctl -C . lint` — all planning entities and dependency links pass.
- `./bin/tskflwctl -C . audit lint 2026-11-10-contract-and-compatibility` — all findings pass.
- `go test -race ./...` — two packages fail, both for reasons belonging to this runner
  rather than to the repository or to this run (which adds one planning markdown file and
  nothing else):
  - `internal/tools/reviewworkspace` — every failure prints `error: rsync is required`;
    `which rsync` finds nothing on this container.
  - `cmd/tskflwctl` · `TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior` — it chmods
    `tasks/` to `0500` to force an unclassified filesystem failure and asserts exit 1. This
    runner is uid 0, and root bypasses the write bit, so the create succeeds and exits 0.
    Confirmed directly: writing into a fresh `0500` directory here succeeds.
  Every other package, `internal/wire` and `internal/cli` among them, passes under `-race`.

## What audited clean

- `internal/wire/wire.go` — the compatibility record is genuinely executable. All 90
  changelog entries are contiguous and ascending, every entry from the 1.68 boundary
  declares ADDITIVE or NOT ADDITIVE, `SchemaRevisionCompatibility` is pinned to the
  current entry's classification, and an ASCII hyphen in place of the em dash gets its own
  actionable diagnostic. Nothing here to report.
- `internal/wire/schema.go` — the published `safety` capability is not decoration. All five
  guarded mutation entry points on `store.FS` (`MutateTaskGraph`, `MutateTaskGraphRepair`,
  `MutateTaskLifecycle`, `MutateThread`, `MutateThreadApply`) reach `writeFileAtomic` only
  through `checkedWriteLock()`, which calls `authorizeMutation()` before taking the lock.
  The hidden bare-verb leaves classified `read-only` really do nothing but return
  `ErrValidation` (`bareverb.go`). The `exit_codes` vocabulary matches the binary exactly.
- `internal/cli/render/columns.go` — the trickiest contract in the surface and it holds.
  Canonical selectors and their legacy aliases both resolve; projected JSON uses the raw
  value under the requested key while table/CSV keeps the display fallback, verified on a
  real task with no `updated_at`; `validateColumnRegistry` fails closed on selector
  collisions and duplicate selections; and `ColumnSpec.Aliases` is not vestigial — the
  `-c` completer uses it to avoid re-offering a column already chosen under its other
  spelling. Only `frontier` breaks the typed-key correspondence (M2).
- Envelope registration — every declared `*Envelope` type carries `schema_version`, and
  `registry_test.go` closes the inverse gap by AST-parsing the package so a declared but
  unregistered envelope cannot escape `$defs`. This is the shape of guard M1 is asking for.
- `internal/wire/dto.go` (random pick) — where descriptions exist on the task/epic/audit/
  research/finding/criterion DTOs, they describe the actual predicate. Spot-checked
  `ready_to_close` against `Audit.Settled()` (including the deliberate divergence where
  `ValidateMove` permits closing an empty audit that `ReadyToClose()` does not advertise),
  `depends_on` against `sortedStrings`, `path` absoluteness against `task info`/`task path`,
  and `criterion_states` — the published four suffix words — against the `NeedsReason()`
  gate in `ToCriteriaJSON`, which admits exactly those four. No stale descriptions found.
- Goldens and generated reference — the goldens carry behavior, not just version strings,
  and the `-c` lists in `--help` and `docs/cli/*.md` agree for all six list commands.
- Stream discipline everywhere except H1 — error envelopes and flag-parse failures leave
  stdout a single parseable document. That consistency is what makes H1 an outlier rather
  than a systemic gap.

## Related-task observations (propose-only)

No OPEN task overlaps any finding. The nearest matches are all completed, which is itself
the observation: each shipped a contract this run found an uncovered edge of. Flagging for
a human to judge whether a follow-up belongs under the same epic — no statuses touched.

- Possibly incomplete as shipped: `planning/tasks/6fbj870023xj-pipeline-output-modes-q-plain-stderr-discipline.md`
  (completed, epic 20) established stderr discipline for diagnostics; cobra's own
  deprecation notice was never routed, which is H1.
- Possibly incomplete as shipped: `planning/tasks/6fe4my002nsc-populate-json-schema-field-descriptions-at-build-time.md`
  (completed, epic 20) built the description mechanism; 687 of 945 published properties
  still carry none, and the Thread DTOs carry almost none, which is M1.
- Related precedent, not an overlap: `planning/tasks/6g1c0wj3qah2-register-the-research-envelopes-in-the-json-schema-and-make-the-coverage-guard.md`
  (completed, epic 28) added the AST closure guard for envelope *registration* after the
  research envelopes shipped undescribed. M1 asks for the same shape of guard one level
  down, over field descriptions.
- Related, not an overlap: `planning/tasks/6g9s401jtqb7-restore-projected-json-and-cli-contract-fidelity.md`
  (completed, epic 20) is where the canonical-selector/alias machinery landed. M2 is the
  one column that was not converted to it.

## External research

None — three Medium-or-higher findings surfaced, so the routine's conditional research
step does not apply.

## Candidate tasks

<!-- candidate-tasks:v1 · ○ open · ● in-progress · ✔ fixed · → tracked · ◌ deferred · ◌ superseded · ✘ wontfix -->
<!-- Add or replace one row with `tskflwctl audit finding <audit> <code> --candidate "<one line>"`; an empty value removes it. -->

- ○ H1 · open — `tskflwctl task new "Keep stdout a single JSON document on deprecated task verbs" --epic 20-cli-ux-and-ergonomics --tags cli,json --tier 2 --priority high --description "Route cobra deprecation notices to stderr and assert per-stream purity for task promote/demote"`
- ○ M1 · open — `tskflwctl task new "Describe the Thread and epic rollup fields in the published JSON Schema" --epic 20-cli-ux-and-ergonomics --tags wire,schema --tier 2 --priority medium --description "Add jsonschema description tags to the Thread DTO family and EpicJSON total/done/percent, and replace the allowlist guard with a closure guard"`
- ○ M2 · open — `tskflwctl task new "Give the projected thread frontier column a non-colliding canonical selector" --epic 20-cli-ux-and-ergonomics --tags cli,json --tier 3 --priority medium --description "Publish frontier_count as the canonical -c selector with frontier retained as a compatibility alias, declared in the next revision"`
- ○ L1 · open — `tskflwctl task new "Derive RecordSource.LocationIsPath from one checked predicate" --epic 21-code-quality-architecture-hardening --tags core --tier 4 --priority low --description "Make service_task.go compare Location to LocalPath like dependency_source.go, behind a shared constructor"`
