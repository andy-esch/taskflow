---
schema: 1
id: 6g9samx90yrd
bucket: closed
area: projected-json-and-cli-contract-fidelity-implementation-claude
date: "2026-09-13"
updated_at: "2026-09-13"
---
# Audit: Projected JSON and CLI contract fidelity implementation — claude — 2026-09-13

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

Reviewer: claude (Opus 5) · 2026-09-13 · both passes complete · every finding left `open` for
implementation-owner triage.

### Verdict

**Not a no-findings close: 2 Medium, 4 Low.** The seven source findings are fixed at the named
instances. Mutation probes confirm the new focused tests kill the H1/M1/M3/L3 regressions. The 35
changed goldens contain only the version bump plus the two intended descriptions. CLI docs and schema
comments regenerate byte-identical, and the fang gate behaves correctly on a real PTY for every
pflag bool spelling.

The compatibility story does not fully hold:

- **M1.** The new canonical name `updated_at` carries H1's fabricated created-date value in
  table/CSV. The drift moved from a JSON key to a table header; it was not removed.
- **M2.** For existing `--json -c updated|open` callers, 1.66 is a silent key rename. The changelog
  calls it a compatibility alias. I demonstrate a `jq` filter that silently returns a wrong answer.
- **L4.** The generic `Column` abstraction does not prevent this defect class. A plausible future
  column re-creates H1 with the whole suite green.

### Isolation attestation

Created with `scripts/isolated-review-workspace.sh create --source <handoff> --deliverable <audit>
--print-path` from the handoff checkout. After that, the only operations in the handoff checkout were
reading this brief and the final guarded transfer.

- Sandbox: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.k8sZrA`
- Resolved Git directory: `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.k8sZrA/.git`
  (independent clone; `origin` is a plain path remote, no worktree/alternates link)
- Baseline commit: `21ac191f8d1777dcaa86d592335d8a75560bca15` ("chore: capture isolated review
  baseline"), parent `e6a9c8078ada5798f3f618e64cacf614fbea38e2`
- Captured source deliverable blob: `f8a492333f225d484b70ef507ae17636e47dae44` · source fingerprint:
  `1ecca41ab0cdd85f401e891f1ae20a2820ca08de`
- Deliverable: `planning/audits/6g9samx90yrd-2026-09-13-projected-json-and-cli-contract-fidelity-implementation-claude.md`
- Verification: `scripts/isolated-review-workspace.sh verify --sandbox <sandbox>` exited 0 with
  `deliverable_changed=true`, `transfer=pending`. Before it ran, `.review-scratch/` was deleted and
  `git status --short` showed only ` M` on this audit.
- Transfer: performed immediately after this attestation with the helper's `transfer` subcommand. Its
  result (`transfer=succeeded` or a refusal) is reported in the reviewer's hand-off message, because
  this file is the payload being transferred. The sandbox is retained until the owner confirms receipt.

### Findings

#### M1. Canonical `updated_at` in table/CSV echoes the wire name over the created-date fallback · **Status:** fixed

**File:** `internal/cli/render/columns.go:99` | **Component:** cli/render
**Effort:** S · **Urgency:** soon

When a column is selected by its canonical name, `SelectColumns` renames the header to that name
(`columns.go:97-101`). It keeps `Extract`, which is still the display closure with the `created`
fallback (`columns.go:277-282` task, `337-342` research). `WriteTablePlain`/`WriteCSV` consume `c.Name`
and `c.Extract`. The result is that one name has two meanings:

- `-c updated_at` under `--json` means "raw frontmatter `updated_at`".
- `-c updated_at` under `-o table`/`-o csv` means "`updated_at`, else `created`".

That is the claim H1 objected to ("the fallback stops being presentation and becomes a claim"), now
made under the wire field's own name. The change also pushes users toward it:

- Completion, `--columns` help, and the generated docs now advertise only `updated_at` (`Specs`,
  `columns.go:64`).
- CLAUDE.md's lead agent triage path is `task list -o table -c …`.
- The parent binary rejected `-o table -c updated_at` (exit 11), so this path is new.

Reproduction (disposable fixture, candidate binary). `never-edited-delta` has
`created: "2026-09-13"` and no `updated_at`; `untouched-research` has `created: "2026-01-06"` and
no `updated_at`:

```
$ tskflwctl task list -o table -c slug,updated_at      → never-edited-delta	2026-09-13
$ tskflwctl task list --json -c slug,updated_at        → {"slug":"never-edited-delta","updated_at":""}
$ tskflwctl task list --json | jq -c '.tasks[]|select(.slug=="never-edited-delta")|{created,updated_at}'
                                                       → {"created":"2026-09-13","updated_at":null}
$ tskflwctl research list -o csv -c updated_at,slug    → 2026-01-06,untouched-research
$ tskflwctl research list --json -c updated_at,id      → {"updated_at":"","id":"6ds16km03taf"}
```

No test pins either behavior. MUT7e adds `if c.jsonExtract != nil { c.Extract = c.jsonExtract }` next
to the header echo, so a canonical table/CSV selection uses the raw value. With it,
`go test ./internal/cli/render/ ./internal/cli/` stays green. `open_findings` is unaffected because
its display and raw values are identical.

**Impact:** a table/CSV consumer computing staleness from the advertised canonical column gets a
last-updated date the authoritative envelope says does not exist, which is H1's failure. Default
output and the legacy `-c updated` selector are unaffected.

**Recommended repair:** bind value semantics to the selector. A canonical selection also switches to
the projection extractor, which is the MUT7e shape. A legacy `updated` selection keeps its header and
display fallback, and default output stays byte-identical. Add a test that, for every contract column
selected by its canonical name, table, CSV, and projected-JSON cells are identical for edited and
never-edited records. **In scope.**

**Resolution:** Explicit canonical updated_at selection now switches table/CSV
to the raw extractor; exact render tests and command-level task/research
fixtures pin raw canonical values while legacy/default views retain the created
fallback.

#### M2. Schema 1.66 records a silent key rename for legacy `--json -c updated|open` callers as compatible · **Status:** fixed

**File:** `internal/wire/wire.go:271` | **Component:** wire
**Effort:** XS · **Urgency:** soon

The written policy (`wire.go:21-23`) is "Adding a field bumps the minor; renaming/removing bumps the
major." For any caller already using `-c updated` or `-c open`, 1.66 changes three things:

- renames the projected output key (`updated`→`updated_at`, `open`→`open_findings`);
- changes `updated`'s value for never-edited records (the `created` date becomes `""`);
- rejects repeated columns (`-c slug,slug` is exit 11 now; parent exit 0 emitted
  `{"slug":…,"slug":…}`).

The 1.66 entry (`wire.go:271-274`) says the old names "remain accepted as compatibility aliases",
which reads as additive. The README hunk and the source audit's M1 Resolution say the same. None of
them states that the output keys and values changed.

**Precedent.** The project has shipped breaking changes as minor bumps before, so the minor number
matches practice:

- 1.26 retired fields (`wire.go:96`).
- 1.32 changed `created.id`'s value, with a "CONSUMERS: … must re-read" note (`wire.go:123-129`).
- 1.47 removed a vocabulary word, labelled "NOT additive" (`wire.go:191-194`).
- 1.59 replaced a shape (`wire.go:251`).

1.66 does not meet the disclosure bar those entries set, and the policy sentence those precedents
already contradict was left in place. The source audit's own M1 Follow-up flagged this exact choice:
"it changes the key a `-c updated` caller sees today. That is a contract call, not a cleanup."

Reproduction (parent built from `e6a9c80` vs candidate, same fixture; `2026-01-07-second-area` has
0 open findings):

```
$ tskflwctl audit list --all --json -c slug,open | jq -r '.audits[] | select(.open != "0") | .slug'
parent 1.65:     2026-01-02-fixture-area
candidate 1.66:  2026-01-02-fixture-area
                 2026-01-07-second-area        ← wrong row, exit 0
$ tskflwctl task list --json -c slug,updated | jq -c '[.tasks[]|{slug,updated}]'
parent 1.65:     …"updated":"2026-01-03"…"updated":"2026-09-13"
candidate 1.66:  "updated":null on every row
```

**Impact:** scripts keyed on the legacy projected names keep exiting 0 but read `null`, so filters
silently return wrong results. Nothing in the changelog tells them to re-key.

**Recommended repair (owner contract call):**

- *Option (a).* Keep 1.66 as a minor bump per precedent, but rewrite the entry to say **NOT additive
  for projected JSON**: selecting `updated`/`open` now emits `updated_at`/`open_findings`, `updated_at`
  is raw (empty when never edited), and repeated columns are rejected; consumers keyed on the old
  projected names must re-key. Mirror that in the README and the source audit's M1 Resolution, and
  amend `wire.go:23` to state the actual rule (defect corrections may ship as a minor with an explicit
  NOT-additive/CONSUMERS note).
- *Option (b).* Bump to 2.0.
  `TestSchemaVersionChangelogIsAscending` hardcodes major 1 (`wire_changelog_test.go:12,41`), so a
  correct `// 2.0:` entry would fail that guard (see L3).

**In scope.**

**Resolution:** Adopted zero-breakage compatibility: explicit updated/open JSON
selectors retain their legacy keys but use corrected raw values; canonical
selectors emit updated_at/open_findings. Schema 1.66 documents both paths.

#### L1. `-c` completion offers the canonical twin of an already-typed legacy alias, which then fails as a duplicate · **Status:** fixed

**File:** `internal/cli/listmode.go:230` | **Component:** cli
**Effort:** XS · **Urgency:** eventually

`columnCompleter` skips columns already typed by checking `used[s.Name]`. Specs now carry only
canonical names (`columns.go:64`), so a legacy token earlier in the value is never marked as used. The
completer then offers `updated_at`/`open_findings`, which `SelectColumns` rejects as a duplicate
(`columns.go:92-94`). The parent build hid the column after `updated,`, so this is a regression.

Help also teaches two spellings: `internal/cli/audit.go:111` and the regenerated
`docs/cli/tskflwctl_audit_list.md:29` still show `-c slug,open`, while `--columns` help lists only
`open_findings` and the README recipe was moved to `open_findings`.

```
$ tskflwctl __complete task list -c 'updated,'      → …updated,updated_at	last-updated date…
$ tskflwctl task list -c updated,updated_at         → error: validation failed: duplicate column "updated_at" (already selected as "updated")   [exit 11]
$ tskflwctl __complete audit list -c 'open,'        → …open,open_findings	open findings
$ tskflwctl __complete research list -c 'updated,'  → …updated,updated_at	last-updated date
```

**Recommended repair:** export alias resolution with the spec (for example `ColumnSpec.Aliases`, or a
`render` helper that canonicalizes a selector). Canonicalize typed tokens before the `used` check,
change the audit-list example to `open_findings`, and regenerate docs. Add a completer test with a
legacy prefix. **In scope.**

**Resolution:** ColumnSpec carries alias metadata and completion canonicalizes
already-used tokens, preventing task, research, and audit aliases from offering
an invalid canonical duplicate.

#### L2. The CLI regression test for never-edited `updated_at` cannot fail on the H1 mutation · **Status:** fixed

**File:** `internal/cli/pipeline_test.go:127` | **Component:** cli tests
**Effort:** XS · **Urgency:** soon

`TestColumns_JSONProjectionUsesCanonicalWireSelectorAndKey` uses `setupRepo`
(`internal/cli/task_test.go:22-23`), whose tasks carry neither `created` nor `updated_at`. The display
fallback therefore also yields `""`, so the "never-edited fixture invented updated_at" assertion
(`pipeline_test.go:143-145`) passes with the defect restored:

| Mutation | `internal/cli/render` | `internal/cli` |
| --- | --- | --- |
| MUT1a: task `jsonExtract` restored to the fallback closure | FAIL (`columns_test.go:158`) | **ok** |
| MUT1b: research-only fallback | FAIL (`columns_test.go:191`) | **ok** |
| MUT1c: `projectedValue` ignores `jsonExtract` | FAIL (`columns_test.go:158`) | **ok** |

Both the source audit's H1 Resolution and the task's validation note say "render and CLI regression
tests cover never-edited records". Only the render unit test does. Research and audit have no
CLI-level projection coverage, and `internal/cli/testdata/planning/` still has no never-edited task
(the source H1 "Tightening (adjacent)").

**Recommended repair:** give the CLI fixture a `created` date without `updated_at` (or add a
never-edited task to `testdata/planning/` plus a projected-JSON golden). Assert `updated_at == ""`
while the table cell shows the created date, add a research case, and re-run MUT1a to confirm
`internal/cli` fails. **In scope.**

**Resolution:** A command-level fixture now creates never-edited task and
research records with real created dates, proving canonical JSON/table/CSV
remain raw while legacy table/CSV retain their fallback.

#### L3. The schema changelog guard accepts deleted and skipped versions · **Status:** fixed

**File:** `internal/wire/wire_changelog_test.go:36` | **Component:** wire tests
**Effort:** XS · **Urgency:** eventually

The guard checks only that entries strictly ascend and that the last one equals `SchemaVersion`.
Mutation results against `internal/wire/wire.go`:

| Probe | Result |
| --- | --- |
| MUT4a: move the 1.65 entry below 1.66 | FAIL ✓ (`1.65 follows 1.66`) |
| MUT4b: append a second `// 1.65:` | FAIL ✓ |
| MUT4c: two `// 1.66:` entries | FAIL ✓ |
| MUT4f: `SchemaVersion = "1.66.1"` | FAIL ✓ |
| MUT4g: entry spelled `// 1.67 :`, version 1.67 | FAIL ✓ |
| MUT4h: unrelated `// 1.99:` comment after the const | FAIL (false positive, fails closed) |
| **MUT4d: delete the 1.40 entry** | **ok ✗** |
| **MUT4e: append `// 1.68:`, set 1.68, no 1.67** | **ok ✗** |

MUT4e is the concurrent-branch rebase the test's own doc comment cites. Two branches both claim 1.67;
while resolving the conflict, the second renumbers to 1.68 and drops the other's entry. Entries are
contiguous 1…66 today (verified with a regex over `wire.go`), so contiguity costs nothing to enforce.
The `1\.` hardcoding also blocks a policy-conforming major bump (M2).

**Recommended repair:** require `minor == previous+1` starting at 1, parse the major rather than
hardcoding `1.`, and optionally scope the scan to the comment block directly above
`const SchemaVersion`. **In scope.**

**Resolution:** The changelog guard now scopes itself to the SchemaVersion
block, parses arbitrary major.minor entries, requires contiguous releases, and
permits the policy-conforming N.x to N+1.0 transition.

#### L4. Column registries have no invariant guard, so the H1 defect class stays open for future columns · **Status:** tracked by 6g9txf8x9m70

**File:** `internal/cli/render/columns.go:22` | **Component:** cli/render
**Effort:** S · **Urgency:** eventually

The repair is instance-level. The abstraction itself enforces none of the invariants it depends on:

- `contractColumn` validates nothing.
- A struct literal inside `render` can set `jsonName` without `jsonExtract`, or the reverse.
- `SelectColumns` silently overwrites colliding selector/alias names (`columns.go:77-81`).
- No test compares every wire-named column's projected value to the full DTO. This is the general
  invariant the source audit's H1 Follow-up recommended.

The fields are unexported, so only package-local code can build a contradictory contract column.

**Hostile evidence:**

- *Scratch in-package probe (created, run, deleted).* Registry
  `{contractColumn("updated","updated_at",…), column("updated_at",…)}`:
  - `-c updated` resolves to the contract column.
  - `-c updated_at` silently resolves to the later plain column (`{"updated_at":"SHADOW"}`).
  - `Specs` lists `updated_at` twice, with no error.
  - A literal `Column{Name:"updated", Extract: fallback, jsonName:"updated_at"}` (no `jsonExtract`)
    projects `{"updated_at":"2026-01-01"}` for a never-edited task: H1 recreated.
  - `jsonExtract` without `jsonName` projects the raw value under the legacy key `updated`, and
    `updated_at` is rejected.
- *Coordinated MUT8.* Append `column("updated_at", …created fallback…)` to `EpicColumns` (copying the
  task pattern). The only failure is the byte header pin at `columns_test.go:284`. After that routine
  pin update, `go test -count=1 ./...` passes, while `epic list --json -c id,updated_at` emits
  `"updated_at":"2026-01-01"` and full `epic list --json` has no `updated_at` key.

**Recommended repair:** one table-driven registry test over all five registries:

- selector and alias names are unique;
- `jsonName != ""` if and only if `jsonExtract != nil`;
- for each column whose selector is a key of the entity's full DTO, the projected string equals the
  stringified full-JSON value for present, absent, and zero fixtures.

No reflection or registry redesign is needed. **Follow-up task.** The implementation task explicitly
scoped out registry consolidation, and this is the separate, non-reflective invariant.

**Resolution:** The systemic selector-collision and contract-column construction
guards are scoped in enforce-projected-column-registry-invariants and sequenced
after this repair in the Harden CLI machine contracts Thread.

### Consumer inventory (verified in the sandbox)

- **`render.Column` / `Specs` / `SelectColumns` / `WriteTablePlain` / `WriteCSV` /
  `ProjectedListJSON`.** The only production call sites are:
  - `internal/cli/listmode.go:167,171` (projected JSON), `:183,188,192` (table/CSV), and `:179`
    (`-o name`/`-q` via `cols[0].Extract`, id column, unaffected, pinned by
    `TestColumnRegistries_FirstColumnIsID`);
  - `bind` → `columnCompleter`/`specNames` (`listmode.go:50-60,219-250`).

  Registries bound through `renderList` + `lm.bind`: `task.go:225,232`, `research.go:263,270`,
  `audit.go:130,137` (audit list), `audit.go:178,185` (audit findings), `epic.go:288,295`. No other
  `Extract`/`Column.Name` consumer exists (grep over `internal`, `cmd`). The TUI does not use these.
- **Registry key vs full-DTO key audit.**
  - Task: every column except the fixed `updated_at` matches its `TaskJSON` key.
  - Epic: all 8 match `EpicMetaJSON`/`EpicJSON`.
  - Audit: all 6 match `AuditJSON` after the fix.
  - Research: all 6 match; `tags` is comma-joined per the documented string projection.
  - Finding: all match `FindingJSON` except the documented synthetic `ref`.
  - Sibling probe: a task without `tier` projects `"tier":"0"` while full JSON omits it. This happens
    only on a lint-invalid task (`internal/domain/lint.go:98`; `tskflwctl lint` exits non-zero on the
    fixture), outside the valid 1-5 range, under documented int stringification. **Not a defect.**
- **`wire.SchemaVersion`.** 37 goldens under `internal/cli/testdata/golden/`; 35 changed. Normalizing
  `"schema_version":"1.65"`→`"1.66"` in the parent version and diffing shows no other change except
  `schema_jsonschema.golden`, which changes exactly the `CriterionJSON.reason` description, the
  `FindingJSON.status` description, and the title `(schema_version 1.66)`. `task_list_csv.golden` and
  `task_list_name.golden` are unchanged.
- **Descriptions and vocabularies.** `schema --json` publishes
  `criterion_states=[deferred,n/a,tracked,wontfix]`, and all four require a reason
  (`CriterionState.NeedsReason`), so the new `reason` description is accurate. `finding_statuses` has
  all 7 statuses.
- **Description guard strength (observation, no finding).**
  - MUT5a (restore the parent literal) fails `TestDispositionDescriptionsReferencePublishedVocabularies`.
  - MUT5b (keep the `finding_statuses` token and append a stale list without `tracked`) and MUT5c
    (re-narrow `reason` while still naming `criterion_states`) both pass it.

  Only the byte golden `TestGolden_MachineContract` catches those, and `-update` would regenerate it.
- **Docs.** `go run ./internal/tools/docgen -out <tmp>` followed by `diff -r docs/cli <tmp>` exits 0.
  `go run ./internal/tools/schemacomments -out <tmp>` produces 264 comments, byte-identical to
  `internal/wire/schema_comments.json`. Stale legacy example: see L1.
- **Legacy-selector consumers in repo.**
  - `internal/cli/audit.go:111` example and generated `docs/cli/tskflwctl_audit_list.md:29` (L1).
  - `internal/domain/audit.go:73` comment mentions `-c open`.
  - `planning/tasks/6ffr4wc03za4-…:57` references `audit list -c open` back-compat.
  - No scripts, CI workflows, or VHS tapes use `-c …updated|open`.

### Selector matrix (candidate, disposable fixture)

The fixture is a copy of `internal/cli/testdata/planning` plus the following, all under
`.review-scratch/` (since deleted):

- a never-edited task;
- edited and never-edited research docs;
- a second audit with 0 open findings.

| Probe | Result |
| --- | --- |
| Default `task list -o table` / `-o csv` | header `…epic\tupdated\tdescription…`; never-edited row shows created `2026-09-13` (byte-compatible) |
| `-o table -c slug,updated` / `-c slug,updated_at` | headers `updated` / `updated_at`; both cells `2026-09-13` for never-edited (M1) |
| `--json -c slug,updated` ≡ `--json -c updated_at,slug` (reordered) | key `updated_at`, raw `""` for never-edited, requested order kept |
| `research list -o table` / `--json -c slug,updated,created` | table `updated` falls back; JSON `updated_at:""` for untouched, `2026-09-13` for edited |
| `audit list --all -o table -c open_findings,slug` / `-o csv -c slug,open` / `--json -c slug,open` | headers `open_findings` / `open`; JSON key `open_findings`, values `"1"`, `"0"` |
| `-c updated,updated_at`, `-c updated_at,updated`, `-c open,open_findings`, `-c slug,slug` | exit 11 `duplicate column …` (JSON envelope under `--json`) |
| `task list --json -c slug,open` | exit 11 unknown column (aliases are per-registry) |
| Empty result: `--status deferred --json -c slug,updated` / `-o csv -c updated_at` / `-o table` | `{"schema_version":"1.66","tasks":[]}` / header-only `updated_at` / full header only |
| `-q -c updated` | exit 11 `--columns applies to -o table, -o csv, or --json` |
| `-c 'slug, updated_at'` / `-c 'slug,'` | exit 11 unknown column `" updated_at"` / `""` (unchanged strict parsing) |
| `audit findings --json -c ref,status,code`, `epic list --json -c id,done,percent` | canonical keys, string values |

### Real-PTY fang evidence

Harness: Python `pty.openpty()` for stderr, stdout on a pipe, `TERM=xterm-256color`, `NO_COLOR`
unset, candidate `bin/tskflwctl`, missing task `zzz`.

| argv | exit | stdout | stderr channel |
| --- | --- | --- | --- |
| `task show zzz` | 10 | 0B | fang styled (ANSI ERROR badge) |
| `… --json`, `--json=true`, `=1`, `=t`, `=T`, `=TRUE`, `=True` | 10 | 0B | JSON envelope `not-found` |
| `--json=T task show zzz`, `--json=1 task show zzz`, `task --json=True show zzz` | 10 | 0B | JSON envelope |
| `… --json=false`, `=0`, `=f`, `=F`, `=FALSE`, `=False` | 10 | 0B | fang styled (human path kept) |
| `… --json=yes`, `--json=`, `--json=tRUE` (invalid) | 1 | 0B | fang styled pflag parse error |
| `task show -- --json`, `task show -- --json=1` | 10 | 0B | fang styled (`--` ends scan) |
| `task show zzz --json true` | 1 | 0B | JSON envelope (`accepts at most 1 arg`) |
| `task list --json -c nope` | 11 | 0B | JSON envelope `validation` |

Parent binary on the same PTY: `--json=1` and `--json=T` produced fang styling (exit 10), confirming
M3 was real and is fixed.

Residual behaviors, identical in the parent build, pre-existing, **not findings**:

- `-o json` errors are prose on a pipe and fang-styled on a TTY; the envelope keys only on root
  `--json` (`cmd/tskflwctl/main.go:50`, `internal/cli/exit.go:74`).
- The argv prescan cannot model pflag's last-wins or value consumption.
  - `task show zzz --json --json=false` gives plain prose on a TTY (gate closed, flag false).
  - `task list --epic -- --json -c nope` consumes `--` as the epic value, so a genuine `--json` run
    gets fang styling on a TTY while the pipe path emits the envelope.
  - Semantic exit codes are preserved in all of these.

### Mutation probes (all restored with `git checkout HEAD -- <file>`; `git status --short` clean after each)

| # | Mutation | Killing test(s) |
| --- | --- | --- |
| 1a/1b/1c | `updated_at` projection collapsed onto display fallback (task / research / shared `projectedValue`) | render `TestProjectedListJSON_UsesCanonicalWireKeysAndRawValues`; **cli test survives (L2)** |
| 2a | `ProjectedListJSON` keys on `c.Name` (legacy key) | render test + cli `TestColumns_JSONProjectionUsesCanonicalWireSelectorAndKey` |
| 2b | audit `jsonName` set to `open` | render `TestSelectColumns_CanonicalNamesAndLegacyAliases` + projection test |
| 3a | gate treats `T`/`TRUE` as false | `TestUseFang/--json=T_closes_it`, `/--json=TRUE_closes_it` |
| 3b | gate reverted to parent (`true` only) | 5 `TestUseFang` subtests (`1`, `t`, `T`, `TRUE`, `True`) |
| 3c | any `--json=<v>` closes gate | `TestUseFang/--json=false_stays_human`, `/invalid_json_value…` |
| 3d | `--` handling removed | `TestUseFang/literal_--json_after_--_does_not_count` |
| 4a–4h | changelog reorder/duplicate/delete/skip/malformed | see L3 (delete and skip survive) |
| 5a–5c | description drift | see inventory (token-preserving drift survives semantic test) |
| 6 | `created.id` regresses to the slug | `TestAuditNew_JSONEnvelope`, `TestDryRun_TaskNew`, `TestCreatedID_IsTheDurableHandle` |
| 7a | canonical header echo removed | render alias test + cli canonical-header assertion |
| 7b | duplicate detection disabled | render alias test only (no CLI-level duplicate test) |
| 7c | `Specs` advertises legacy name | render alias test |
| 7d | legacy alias registration removed | render alias + projection tests, cli test |
| 7e | canonical table/CSV selection uses raw value (M1 repair shape) | **nothing fails** |
| 8 | new epic `updated_at` fallback column + header-pin update | **full suite passes (L4)** |

### Planning reconciliation

- **Source audit.** `audit info 2026-09-08-contract-and-compatibility --json` reports bucket `closed`,
  findings `total 7 / done 7 / open 0`. `audit findings … --json -c code,status` shows H1, M1, M2, M3,
  L1, L2, L3 all `fixed`, each with a `**Resolution:**` paragraph in the tool's shape. `audit lint`
  and `lint` pass. Legitimately closable, apart from the Resolution wording corrections implied by M2
  and L2.
- **Task `6g9s401jtqb7`.** Completed, `ac 8/8`, `audit_sources` set. The AC "Legacy … table and CSV
  default headers and display fallback remain byte-compatible" holds for default output. The
  validation claims (race tests, lint, tidy, regenerated docs/comments, planning lint,
  `git diff --check`) all reproduced green. The "CLI regression tests cover never-edited records"
  claim does not (L2).
- **`created.id` task `6fq9zy15j5pz`.** Closed with executable coverage: MUT6 fails three create-path
  tests. Two claims in its Resolution go beyond that evidence:
  - "Existing machine-contract goldens … pin the distinction" is inaccurate: no created-envelope
    golden exists among the 37.
  - Research/Thread create coverage was not exercised by MUT6 (task/audit only).

  Minor wording, no finding.
- The unrelated `planning/tasks/6g9mz2shwmb0-…` change is present in the baseline and was not
  reviewed, edited, or restored.

### Commands run (sandbox only)

- **Race tests.** `GOCACHE=<scratch>/gocache go test -race ./...` → all packages ok.
  - A first run with `GOTMPDIR` inside the Claude scratch directory failed 9 tests in `cli`, `config`,
    `spacehealth`, and `workspacestore`.
  - Those tests discover a planning root by walking up from the temp dir, and they found the scratch
    tree's parent. That is environmental. Re-run with the default `TMPDIR`: all ok.
- **Static checks.**
  - `golangci-lint run ./...` → `0 issues.`
  - `go mod tidy -diff` → exit 0, no output.
  - `git diff --check HEAD~1 HEAD` and `git diff --check` → exit 0.
- **Planning lint.** `bin/tskflwctl lint` → `✔ all planning entities and dependency links pass lint`;
  `audit lint` → `✔ all audit findings pass lint`.
- **Regeneration diffs.** `docgen` and `schemacomments` into temp paths, then diffed (exit 0).
- **Builds.** Candidate `go build -o bin/tskflwctl ./cmd/tskflwctl` (gitignored). Parent from
  `git archive e6a9c80` into `.review-scratch/parent`, used for the behavioral comparisons above.
- **PTY harness.** `.review-scratch/ptyprobe.py`. Mutation loops as tabulated.
  `.review-scratch/` was deleted before `verify`.
