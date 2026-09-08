---
schema: 1
id: 6g810b6ptmrp
bucket: open
area: contract-and-compatibility
date: "2026-09-08"
updated_at: "2026-09-08"
---

# Code Quality Audit: contract-and-compatibility — 2026-09-08

> Edit findings through `tskflwctl audit finding` so status and resolution
> metadata stay queryable. Never hand-edit a `**Status:**` line.

Routine: `code-quality-audit` · lens `contract-and-compatibility` · ISO week `2026-W37`,
slot `Tue` (index 2).

## Punch list

One line per finding for fast triage. Order: Critical → High → Medium → Low.

- H1. `--json -c updated` invents a last-updated date the full envelope says is absent  (effort: S · urgency: soon)
- M1. `--json -c` column keys diverge from the wire keys for the same datum  (effort: S · urgency: soon)
- M2. `CriterionJSON.reason`'s published description omits `tracked`, a state that requires it  (effort: XS · urgency: soon)
- M3. The fang gate misses truthy `--json=<v>` spellings, dropping the JSON error envelope  (effort: XS · urgency: soon)
- L1. `FindingJSON.status`'s vocabulary literal has no drift pin, unlike its epic sibling  (effort: XS · urgency: eventually)
- L2. `AuditColumns` is undocumented — its godoc line was absorbed into `ResearchColumns`  (effort: XS · urgency: eventually)
- L3. The `SchemaVersion` changelog block is out of numeric order from 1.42 onward  (effort: XS · urgency: eventually)

## Files audited

- **Signal**: `internal/wire/envelopes.go` — 8 commits in 30 days, 1217 lines, 39 of the
  52 envelope types; the single largest published-contract surface in the tree.
- **Signal**: `internal/wire/wire.go` — 6 commits in 30 days; owns `SchemaVersion`
  (referenced by 9 files) and the changelog every consumer reads to learn what moved.
- **Adjacency** (from `envelopes.go`, across the seam to the *published description*
  contract): `internal/wire/dto.go` + `internal/wire/schema.go` — the DTOs whose
  `jsonschema:"description=…"` tags become the shipped schema. `schema_comments.json`,
  their generated companion, is the highest-churn file in the lens (11 commits in 30 days).
- **Adjacency** (from `wire.go`, to the registry it publishes as `task_fields`):
  `internal/domain/fields.go`.
- **Random**: `cmd/tskflwctl/main.go` — the entrypoint that decides whether a run takes
  fang's human path or the machine-contract path.
- **Chased** (the lens's "does a `-c` projection expose a field the full `--json`
  doesn't" question led here): `internal/cli/render/columns.go`, which owns the four
  column registries `--json -c` projects through.

## Commands run

```
go build -o bin/tskflwctl ./cmd/tskflwctl      # `just` unavailable in the runner; exit 0
go test -race ./...                             # exit 0, no pre-existing failures
./bin/tskflwctl audit list --json               # 1 open audit — below the 10 backpressure bar
./bin/tskflwctl task list --json                # full envelope keys, for the -c comparison
./bin/tskflwctl task list --json -c slug,updated
./bin/tskflwctl task list --json -c slug,updated_at   # rejected: "unknown column"
./bin/tskflwctl task ac 6g72ncs4xjdm --list --json    # a live `tracked` criterion with a reason
./bin/tskflwctl schema --json                   # published finding_statuses / criterion_states
./bin/tskflwctl task new … --dry-run --json     # created.id/slug probe (wrote nothing)
script -q -e -c "./bin/tskflwctl task show zzz --json=1" /dev/null   # pty-stderr fang probe
```

An AST sweep over `internal/wire/*.go` confirmed all 52 `*Envelope` structs declare
`SchemaVersion` — no envelope ships without one.

## Findings

### Critical

(none)

### High

#### H1. `--json -c updated` invents a last-updated date the full envelope says is absent  · **Status:** open

**File:** `internal/cli/render/columns.go:233` | **Component:** cli/render
**Effort:** S · **Urgency:** soon

The `updated` column falls back to `t.Created` when `t.Updated` is empty. For
`-o table`/`csv` that is a defensible *display* choice (an empty cell would sort last),
and the sibling `ResearchColumns` comment argues exactly that. But the same closure also
feeds `--json -c`, where the fallback stops being presentation and becomes a claim: the
projection asserts a last-updated date for a task whose authoritative envelope omits
`updated_at` entirely, because the file has never been edited.

35 of the tasks in this repo's own `planning/tasks/` carry no `updated_at`, so this is
not a corner case — it is roughly a fifth of the corpus.

CLAUDE.md steers agents to `--json -c` as "the cheap machine path" for triage. An agent
computing staleness from `-c updated` and one computing it from `--json`'s `updated_at`
get different answers for the same task, and the cheap path is the one that is wrong.

**Failing scenario:** `planning/tasks/6g7srp3py9fe-…` has `created: "2026-09-07"` and no
`updated_at:` key.

```
$ tskflwctl task list --status ready-to-start --json -c slug,updated
… {"slug":"make-adversarial-review-…","updated":"2026-09-07"}

$ tskflwctl task list --status ready-to-start --json
… {"slug":"make-adversarial-review-…","created":"2026-09-07"}   # no updated_at key at all
```

The projection reports an edit that never happened. `ResearchColumns` (`columns.go:294`)
carries the identical fallback and the identical divergence.

**Why tests didn't catch it:** the goldens under `internal/cli/testdata/golden/` cover
`task_list_json` and the table/csv projections separately, but nothing asserts that a
`--json -c` cell and the full `--json` value for the same field *agree*. The fixture in
`testdata/planning/` also has no never-edited task, so the fallback branch is never
exercised in a golden at all.

**Recommendation:** split the presentation fallback from the projected value. Keep the
`created`-fallback for the `table`/`csv` renderers, and have the `--json -c` path emit the
raw `t.Updated` (empty when unset) so the projection can never contradict the envelope.
The narrowest form is a per-column "display fallback" the JSON projector skips; a
`Column` field like `DisplayOnly func(...) string` alongside the existing accessor keeps
the change local to `columns.go` and its two callers.

**Tightening (adjacent):** add a never-edited task to `internal/cli/testdata/planning/`
so the fallback branch has golden coverage in both directions, and fix the two column
help strings — "last-updated date" is not what the closure returns.

**Follow-up:** a general invariant test — for every entity, every `-c` column whose name
matches a full-`--json` key must produce the same value for the same record — would close
this class rather than this instance. That is the shape of a task, not of this fix.

### Medium

#### M1. `--json -c` column keys diverge from the wire keys for the same datum  · **Status:** open

**File:** `internal/cli/render/columns.go:233` | **Component:** cli/render, wire
**Effort:** S · **Urgency:** soon

`wire.go:24` states the rule plainly: "JSON keys match the frontmatter keys exactly
(`created`, `updated_at`)." The `--json -c` projection uses the *column* name as the JSON
key, and three of the four column registries have a column whose name is not the wire key:

| entity | column name | full `--json` key |
| --- | --- | --- |
| task | `updated` | `updated_at` |
| research | `updated` | `updated_at` |
| audit | `open` | `open_findings` |

`epic` is clean — all eight of its columns match their wire keys.

The consequence is worse than a synonym, because the wire spelling is *rejected*:

```
$ tskflwctl task list --json -c slug,updated_at
{"schema_version":"1.61","error":{"code":"validation","message":"validation failed:
 unknown column \"updated_at\" (available: slug, status, tier, priority, epic, updated,
 description, revisit_at)"}}
```

So an agent that learned `updated_at` from `schema --json` or from a full envelope cannot
ask for it by that name, and gets a differently-named key back when it guesses right.
Note `revisit_at` in that same list *does* match its wire key, which is what makes
`updated` read as an oversight rather than a convention.

**Failing scenario:** an agent projects `-c slug,updated_at` to cheapen a staleness scan
and takes exit 11 instead of rows; it retries with `-c updated` and must now maintain a
column-name↔wire-key translation table that exists for exactly three fields.

**Why tests didn't catch it:** the column registries and the wire DTOs have no shared
source and no cross-check. Each side's goldens are internally consistent, so both pass.

**Context:** severity sits at Medium rather than High because CLAUDE.md already documents
`--json -c` as "a string-valued column **view** (like `-o table`/`csv`)" that does not
validate against `schema --json-schema` — the divergence is *disclosed*, if not intended.
H1 outranks it because a wrong value is not disclosed anywhere.

**Recommendation:** rename the three columns to their wire keys (`updated_at`,
`open_findings`), keeping the current spellings as accepted aliases so existing scripts
and completions keep working. Aliasing rather than replacing keeps this additive.

**Follow-up:** if the aliases are added, decide whether an alias should project under the
alias name or the canonical name — projecting under the canonical name is the only choice
that makes the two paths agree, but it changes the key a `-c updated` caller sees today.
That is a contract call, not a cleanup.

#### M2. `CriterionJSON.reason`'s published description omits `tracked`, a state that requires it  · **Status:** open

**File:** `internal/wire/dto.go:107` | **Component:** wire
**Effort:** XS · **Urgency:** soon

The shipped JSON Schema describes `reason` as:

> why the criterion is deferred/wontfix/n-a — required for those states

`domain.CriterionState.NeedsReason()` (`resolution.go:100`) returns true for **four**
states — `deferred`, `wontfix`, `n/a`, and `tracked` — and `ToCriteriaJSON` populates
`Reason` from exactly that predicate. The published description therefore under-reports
the field by one state, and it is not a hypothetical state: `schema --json` lists
`criterion_states: [deferred, n/a, tracked, wontfix]`, and this repo's own corpus has
tracked criteria carrying reasons:

```
$ tskflwctl task ac 6feeygw00jmx --list --json
… {"index":3,"checked":false,"text":"`audit sync <slug>` …","state":"tracked",
   "reason":"carried by 6g3ag8py12y9 — the linkage convention is a design call…"}
```

The same string also spells the state `n-a` where the vocabulary word is `n/a`, probably
because `/` is already doing separator duty inside the enumeration — which is the tell
that a re-typed literal was the wrong shape here.

**Failing scenario:** an agent generates a criterion validator from
`schema --json-schema`, treats `reason` as meaningless for `state: "tracked"`, and drops
the destination prose on the one state whose entire point is that the work went somewhere
followable.

**Why tests didn't catch it:** `TestJSONSchema_HasFieldDescriptions` asserts descriptions
are *non-empty*, never that they are *true*. The one description pinned to its vocabulary
is `EpicMetaJSON.status` (`TestEpicStatusDescriptionMatchesVocab`) — the guard exists,
it just was not extended to this field.

**Recommendation:** stop re-typing the vocabulary. The sibling `State` field on the same
struct already gets this right — "one of criterion_states in the schema contract" — so
make `Reason` point the same way: "why the criterion carries its state — required for
every criterion_states value". One literal removed is one literal that cannot drift.

**Tightening (adjacent):** while in the file, `dto.go:251`'s `FindingJSON.Status` has the
same re-typed-vocabulary shape (see L1).

#### M3. The fang gate misses truthy `--json=<v>` spellings, dropping the JSON error envelope  · **Status:** open

**File:** `cmd/tskflwctl/main.go:70` | **Component:** cmd
**Effort:** XS · **Urgency:** soon

`useFang` scans argv for `--json` and `--json=true` only. pflag parses a bool flag with
`strconv.ParseBool`, which also accepts `1`, `t`, `T`, `TRUE`, and `True`. For those
spellings cobra sets `app.JSON = true` while `useFang` returns true, so a TTY-stderr run
takes fang's styled human path — the exact path the file's own comment says "the whole
machine contract … lives on that fall-through path; fang never touches it".

**Failing scenario:** on a terminal (or any harness that allocates a pty, which many agent
runners do):

```
$ tskflwctl task show zzz-does-not-exist --json
{"schema_version":"1.61","error":{"code":"not-found","message":"task \"zzz-does-not-exist\": not found"}}

$ tskflwctl task show zzz-does-not-exist --json=1
  ␛[101m ␛[1;97;101mERROR␛[m …
  Task "zzz-does-not-exist": not found.
```

Verified for `1`, `t`, `T`, `TRUE`, and `True`. The exit code survives (10 in both cases)
and stdout on the success path stays JSON — it is the **error envelope** that is lost,
replaced by ANSI-styled prose with the message reworded (capitalised, full-stopped). A
consumer parsing stderr for `{"error":{"code":…}}` gets terminal escapes.

**Why tests didn't catch it:** `TestUseFang` (`main_test.go:170`) covers `--json` and
`--json=true` and stops there, while its own docstring claims to guard "any `--json`
run". The table asserts less than the comment above it promises. `TestSmoke` cannot
catch it either — it runs the binary in a subprocess where stderr is not a TTY, so
`useFang` is already closed by the TTY check.

**Recommendation:** parse the value rather than matching two literals — for an arg with
the `--json=` prefix, feed the suffix to `strconv.ParseBool` and close the gate on any
value that parses true (an unparseable value can stay open; cobra will reject the run
anyway). Add the four missing spellings to `TestUseFang`.

**Tightening (adjacent):** the sibling boolean `--dry-run` is not part of the gate and
does not need to be, but the same argv scan would misread `--json` if a shorthand is ever
added; a comment naming that assumption would age better than the current one.

### Low

#### L1. `FindingJSON.status`'s vocabulary literal has no drift pin, unlike its epic sibling  · **Status:** open

**File:** `internal/wire/dto.go:251` | **Component:** wire
**Effort:** XS · **Urgency:** eventually

`FindingJSON.Status`'s description re-types all seven finding statuses as a compile-time
literal. It is currently **correct** — it matches `domain.FindingStatuses()` exactly — so
this is a coverage gap, not a live defect. It earns a place next to M2 because this exact
literal has already drifted once: schema 1.47 dropped `landed` for `tracked`, and
`TestEpicStatusDescriptionMatchesVocab` exists precisely because "a compile-time literal
would silently leave it stale". `render.findingStatusOrder` got its registry pin
(`TestFindingStatusOrder_CoversRegistry`); this literal did not.

**Recommendation:** extend `TestEpicStatusDescriptionMatchesVocab`'s pattern to
`FindingJSON.status` against `domain.FindingStatuses()` — or better, resolve it the way
M2 recommends and point at `finding_statuses` in the schema contract instead of
enumerating.

#### L2. `AuditColumns` is undocumented — its godoc line was absorbed into `ResearchColumns`  · **Status:** open

**File:** `internal/cli/render/columns.go:282` | **Component:** cli/render
**Effort:** XS · **Urgency:** eventually

`ResearchColumns` was inserted between `AuditColumns`'s doc comment and its declaration,
so the comment block above `func ResearchColumns` opens with "AuditColumns is the
projectable column set for `audit list` (slug first)." and `func AuditColumns`
(`columns.go:304`) now has no doc comment at all. A godoc reader is told the wrong thing
about one exported function and nothing about the other. Adjacent to H1/M1, both of which
land in this file.

**Recommendation:** move the first line back above `func AuditColumns`.

#### L3. The `SchemaVersion` changelog block is out of numeric order from 1.42 onward  · **Status:** open

**File:** `internal/wire/wire.go:172` | **Component:** wire
**Effort:** XS · **Urgency:** eventually

The changelog ascends 1.1 → 1.42, jumps to 1.56, descends to 1.43, then ascends again
1.57 → 1.61. Every entry is present and each one is accurate; only the ordering is
scrambled, presumably from concurrent branches each inserting at their own cursor. It is
in-lens because this comment block *is* the compatibility record — it is where a consumer
looks to learn what changed between two versions, and a reader scanning for "what landed
after 1.50" currently has to read the whole block to be sure they found it all.

**Recommendation:** re-sort ascending and add a one-line "append new entries at the
bottom" note so the next insert has an obvious cursor.

## What audited clean

- `internal/wire/envelopes.go` — **clean for this lens.** An AST sweep confirmed all 52
  declared `*Envelope` structs carry `SchemaVersion`; no envelope can ship versionless.
  `TestJSONEnvelopes_RegistryIsComplete` closes the harder direction by *parsing the
  package source*, catching an envelope that is declared and emitted but never registered
  — the research-envelope bug that reflection structurally could not see. This is a
  stronger guard than most contract surfaces get.
- `internal/domain/fields.go` — **clean.** The known-field set and the int/list/date maps
  are all derived from one `taskFields` table rather than kept in lockstep by eye, the
  accessors are unexported-backed so no sibling package can mutate the type map, and
  `TestTaskFieldsMatchStruct` pins the table to the `domain.Task` yaml tags. The drift this
  lens hunts for has already been designed out here.
- `internal/wire/schema.go` — **clean.** The contract DTOs are thin and their comments
  explain *why* each published set exists (`criterion_states`, `research_fields`) rather
  than restating the type. `schema_comments.json`, despite being the highest-churn file in
  the lens, is guarded byte-for-byte by `TestSchemaComments_NotStale`, which additionally
  carries a floor guard so a broken generator fails loudly instead of false-passing on an
  empty map.
- `internal/wire/wire.go` — clean apart from L3. The `SchemaVersion` discipline itself
  holds: every entry names what changed and, where a change was not additive (1.32, 1.47),
  says so explicitly and names the consumers affected.
- `cmd/tskflwctl/main.go` — clean apart from M3. The fang/machine split is the right
  shape, `useFang` is pure and testable by deliberate design, and `repoColorScheme` takes
  its theme as a parameter for the same reason. The defect is in the gate's argv matching,
  not in the architecture around it.
- Epic's column registry — the one of four that matches its wire keys on every column.

## External research

Not done: four Medium-or-higher findings surfaced, above the step-11 threshold.

## Candidate tasks (human to triage)

No open task matched any finding's fingerprint. The nearest hits
(`honor-c-columns-and-compact-output-for-json`,
`populate-json-schema-field-descriptions-at-build-time`,
`audit-agent-discoverability-of-the-schema-contract-for-sibling-gaps`) are all
**completed** — they are the work that *built* these surfaces, not open owners of these
defects. Overlap classified NONE for all seven findings; no task annotations were made.

- `tskflwctl task new "Stop --json -c from inventing an updated date the envelope omits" --epic 20-cli-ux-and-ergonomics --tags cli,json,agent --tier 2 --priority high --description "The updated column falls back to created for display; that fallback leaks into --json -c, so the cheap machine path reports an edit the full envelope says never happened (H1)."`
- `tskflwctl task new "Align -c column names with the wire keys they project" --epic 20-cli-ux-and-ergonomics --tags cli,json,agent --tier 3 --priority medium --description "task/research updated and audit open project under names the full --json does not use, and the wire spelling updated_at is rejected outright (M1)."`
- `tskflwctl task new "Point published descriptions at the vocabulary instead of re-typing it" --epic 26-frontmatter-schema-declared-validation-contract --tags schema,agents --tier 3 --priority medium --description "CriterionJSON.reason omits tracked and misspells n/a; FindingJSON.status re-types seven statuses unpinned. Both should reference the schema contract's published sets (M2, L1)."`
- `tskflwctl task new "Close the fang gate on every truthy --json spelling" --epic 21-code-quality-architecture-hardening --tags cli,json --tier 2 --priority medium --description "--json=1|t|T|TRUE|True takes fang's human path on a TTY, replacing the machine error envelope with styled prose; TestUseFang covers only two spellings (M3)."`
- L2 and L3 are one-line repairs; fold them into whichever of the above touches the file
  rather than minting tasks for them.

## Related-task observations (propose-only)

- **Possibly already-shipped:** `planning/tasks/6fq9zy15j5pz-task-new-json-expose-the-minted-id-as-its-own-field.md`
  (status `ready-to-start`). It reports that `created.id` returns the slug and the minted
  id is only inside `created.path`. Schema 1.32 changed exactly that, and both of its
  acceptance criteria look satisfied today. Verified with a dry run (wrote nothing):

  ```
  $ tskflwctl task new "zzz probe delete me" --epic 21-… --tags cli --dry-run --json
  {"…","created":{"kind":"task","id":"6g81078bkh94","slug":"zzz-probe-delete-me",…}}
  ```

  `created.id` carries the stable id and `created.slug` is a distinct field, consistent
  across `task`/`epic`/`audit`/`research` (all four emit the one `CreatedEnvelope`). Worth
  a human read before completing — status left untouched, per this routine's scope.
