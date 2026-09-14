---
schema: 1
id: 6g9sxk32b2er
bucket: closed
area: projected-json-and-cli-contract-remediation-design-antigravity
date: "2026-09-13"
---
# Audit: Projected JSON and CLI contract remediation design — antigravity — 2026-09-13

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

Perform a focused second-round adversarial design review of the uncommitted implementation for
`restore-projected-json-and-cli-contract-fidelity` (`6g9s401jtqb7`) on branch
`fix/contract-audit-sweep`. Start from the findings in
`planning/audits/6g9sw0vjj48f-2026-09-13-projected-json-and-cli-contract-fidelity-implementation-antigravity.md`,
but do not merely restate them. The implementation owner needs an evidence-backed remediation design
that is internally consistent and does not exchange one compatibility defect for another.

The parent audit's strongest signals are H1, M1, M2, and L1. Challenge their proposed repairs and
settle the exact contract. Reassess L2, L3, and L4 skeptically: reject scope inflation, parity for its
own sake, or a proposed invariant that the repository does not actually promise.

## Review target

Review the captured sandbox baseline for `fix/contract-audit-sweep` relative to
`e6a9c8078ada5798f3f618e64cacf614fbea38e2`, including the completed task, the closed source audit,
and the first Antigravity implementation audit. The pre-existing local modification to
`planning/tasks/6g9mz2shwmb0-cut-v0.21.0-as-a-spatial-threads-preview.md` remains unrelated and must
not be edited or treated as implementation evidence.

Build a precise model of these independently observable dimensions of a selected column:

- accepted selector spellings and aliases;
- advertised completion/help name;
- table/CSV header;
- table/CSV value semantics;
- projected-JSON key;
- projected-JSON value semantics; and
- canonical full-envelope wire field, if one exists.

Determine whether the current `Column` representation and `SelectColumns` mutation can express that
model without contradictory states. Prefer a small enforceable abstraction over scattered format
conditionals, but demonstrate why every additional field or constructor invariant is needed.

## Intended contract to challenge

1. Under schema 1.66, what exact output should each of these emit for a never-edited task, and why?
   Test table, CSV, and projected JSON for both `-c updated` and `-c updated_at`.
2. Does accepting legacy `updated` / `open` while changing their projected JSON output keys violate
   the repository's documented global `SchemaVersion` rule? Decide from repository evidence whether
   the sound repair is legacy-key preservation, an explicit projection-contract exception, or a
   major schema bump. Account separately for output key compatibility and the intentionally corrected
   raw-value semantics.
3. Can canonical selectors use raw canonical values in table/CSV while default and explicitly legacy
   selectors preserve their historical display headers and fallbacks? Show the minimal representation
   and regression matrix needed to prevent header/value lies.
4. How should completion canonicalize already-selected aliases so it never suggests a duplicate?
   Probe canonical-first, alias-first, repeated values, partial prefixes, mixed ordering, and every
   affected registry.
5. Specify an argv scanner for root `--json` that matches pflag's effective bool behavior closely
   enough for both fang routing and post-parse error envelopes. Exercise repeated flags with last-value
   precedence, all accepted true/false spellings, invalid values, unknown flags before and after the
   JSON flag, subcommand placement, and `--`. Do not recommend deriving error formatting from
   `useFang(..., false)`, because non-TTY alone forces that function false and loses the JSON signal.
6. For L2, identify only executable vocabularies that actually have a published registry today.
   Separate the missing `CriterionJSON.state` regression assertion from broader schema-vocabulary
   normalization, and say whether the latter has a real current owner.
7. For L3, establish whether contiguous minor versions are a documented invariant or an invented one.
   A future-major-aware parser is sensible only if the resulting test correctly handles `1.N -> 2.0`
   and still agrees with the actual version policy.
8. For L4, determine whether list-column parity with every full DTO field is promised anywhere. Add
   `AuditColumns.updated_at` only if user workflows or a stated contract justify it; otherwise reject
   the finding or recommend a separately scoped additive task.

## Mandatory evidence floor

- Build and report a repository-wide consumer inventory for `Column`, `ColumnSpec`, `Specs`,
  `SelectColumns`, every list renderer, the completion callback, `useFang`, and `SchemaVersion` before
  proposing a repair.
- Use disposable fixtures with absent and populated `updated_at` values and nonzero audit findings.
- Record exact bytes, keys, and exit codes for the selector/format matrix; do not infer behavior from
  helpers alone.
- Drive completion through the actual completion callback or shell protocol and demonstrate the
  duplicate trap and proposed suppression rule.
- Exercise argv behavior in a real PTY as well as focused pure tests. Include `--json --json=false`
  and the inverse ordering; unknown and invalid flags before/after valid JSON flags; and a literal
  after `--`.
- Mutation-test the recommended abstraction: make a canonical table header use the legacy fallback,
  leak a canonical key through a legacy compatibility path (or vice versa, according to the chosen
  contract), remove alias-aware completion, and change repeated-flag precedence. Each relevant test
  should fail for the intended semantic reason.
- Inspect existing docs, schema policy, changelog precedent, and released behavior at v0.21.0. Label
  requirements that exist only in the new task separately from previously shipped promises.

## Required hostile angles

- Assume the current alias design can preserve input compatibility while still breaking output
  compatibility. Trace the requested spelling all the way from argv to each emitted key and header.
- Assume table and CSV are machine-consumed contracts, not merely human presentation. Reject any
  canonical header whose cell semantics belong to a legacy display alias.
- Treat repeated bool flags and early parser failures as state-machine problems. A few enumerated
  spellings are not adequate evidence unless precedence and termination behavior are also pinned.
- Challenge completion using aliases that are intentionally hidden from `Specs`; hidden compatibility
  inputs must still participate in duplicate suppression.
- Require repository evidence before broadening schema vocabularies or demanding column parity. A
  nearby field is not automatically an invariant.

## Required result

Produce one recommended contract matrix and one bounded implementation plan. For each first-round
finding, state `confirmed`, `confirmed but remediation revised`, `rejected`, or `follow-up`, with exact
evidence. If the contract cannot be settled from repository evidence, name the smallest user design
decision needed rather than hiding it in a code recommendation.

Any new findings must use the repository's exact finding grammar and remain `open` for owner triage.
Do not edit implementation or planning files other than this assigned audit.

## Validation and restoration

Run all probes only in the mandatory independent sandbox. Restore every mutation to the sandbox
baseline before final verification. The source checkout is a read-only handoff, and only the assigned
audit may be transferred back through the isolation helper.

## Deliverable

Replace the placeholder with the contract matrix, finding-by-finding disposition recommendation,
bounded repair plan, commands and evidence, residual uncertainty, and the mandatory isolation
attestation. Leave the independent sandbox intact until receipt is confirmed.

## Reviewer report

### Executive verdict

**Verdict:** Remediation design finalized with actionable adjustments. The first-round implementation review correctly identified critical boundary failures, but its initial remediation proposals contained several flaws:
1. **Parent H1 proposal was under-specified:** It identified the created-date fallback leak under `updated_at` in CSV/table, but did not show how `Column` and `SelectColumns` can resolve it cleanly without adding new abstractions or mutating renderer call signatures.
2. **Parent M1 proposal had an unexamined trade-off:** Preserving output key compatibility (`row["updated"]`) for legacy callers conflicts directly with the task's primary objective of eliminating non-wire keys from JSON machine output. Both sound solutions (Key-Preserving Compatibility vs. Canonical-Key View Enforcement) are fully analyzed below, with the concrete tradeoffs and exact code changes specified.
3. **Parent L1 proposal was broken for non-TTY:** Proposing `!useFang(os.Args[1:], false)` causes all non-TTY machine runs (CI, pipelines, redirects) to lose their JSON error envelope signal because `!stderrIsTTY` unconditionally forces `useFang` to return false. A pure argv bool scanner (`jsonFlagActive`) that evaluates last-flag precedence independently of TTY state is required.
4. **Parent L2, L3, and L4 were scope-inflated:**
   - **L2** correctly caught the omission of `CriterionJSON.state` in the regression test, but its demand to rewrite `FindingJSON.effort` and `FindingJSON.urgency` is rejected because no published schema registries exist for them.
   - **L3** caught the regex major-version limit `^// 1\.`, but its demand for strict `+1` version continuity is an invented requirement contradicted by repository history.
   - **L4** is rejected outright: no column registry in `taskflow` has parity with its full wire DTO; column registries are purposefully curated for compact list triage.

---

### Mandatory isolation and transfer attestation

All work for this review was performed exclusively within the independent sandbox workspace created by `scripts/isolated-review-workspace.sh`:

| Attestation Dimension | Sandbox Parameter Value |
| :--- | :--- |
| **Sandbox Path** | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.2fZ3j7` |
| **Resolved Git Directory** | `/private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.2fZ3j7/.git` |
| **Sandbox Baseline Commit** | `5691b6957023bf5090607ae083c61494361c253b` |
| **Captured Source Blob** | `3e96ee6a0a684b86702bcab990b1b3af6ccfe942` |
| **Captured Source Fingerprint** | `e4c2c3bc366d14368c6b0d6e4cae85eb6fc45591` |
| **Deliverable Relative Path** | `planning/audits/6g9sxk32b2er-2026-09-13-projected-json-and-cli-contract-remediation-design-antigravity.md` |

---

### Comprehensive observable dimensions model

To resolve the contract cleanly, every column is modeled across seven independently observable dimensions:

```
+---------------------------------------------------------------------------------------------------------+
|                                    OBSERVABLE COLUMN DIMENSIONS                                         |
+---------------------+-------------------+-------------------+-------------------+-----------------------+
| 1. Selector Spelling| 2. Advertised Name| 3. Table/CSV      | 4. Table/CSV      | 5. Projected-JSON     |
|    & Aliases        |    (Specs/Help)   |    Header         |    Value Semantics|    Key & Value        |
+---------------------+-------------------+-------------------+-------------------+-----------------------+
```

Below is the complete behavioral matrix across all affected registries (`TaskColumns`, `ResearchColumns`, `AuditColumns`) for a record created on `2026-01-01` that was **never edited** (`Created = "2026-01-01"`, `Updated = ""`, `OpenFindings = 3`):

| Registry | Invocation / Selector | Advertised Name | Output Header (Table/CSV) | Cell Value (Table/CSV) | Projected JSON Key | Projected JSON Value | Canonical Wire DTO Key & Value | Rationale & Behavioral Guarantee |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Task** | *(default table/csv)* | `updated_at` | `updated` | `"2026-01-01"` | N/A | N/A | `updated_at`: omitted | Byte-stable default human table. `updated` represents last touch (created fallback). |
| **Task** | `-c slug,updated` | `updated_at` | `updated` | `"2026-01-01"` | `"updated"` *(Option 1)* / `"updated_at"` *(Option 2)* | `""` | `updated_at`: omitted | Explicit legacy presentation selector. Preserves legacy header & fallback in table/CSV; JSON emits raw value. |
| **Task** | `-c slug,updated_at` | `updated_at` | `updated_at` | `""` (empty) | `"updated_at"` | `""` | `updated_at`: omitted | Explicit canonical wire selector. Header matches wire name; cell emits truthful raw timestamp (no lie). |
| **Research**| *(default table/csv)* | `updated_at` | `updated` | `"2026-01-01"` | N/A | N/A | `updated_at`: omitted | Identical to task: default table retains legacy header and created fallback. |
| **Research**| `-c slug,updated` | `updated_at` | `updated` | `"2026-01-01"` | `"updated"` *(Option 1)* / `"updated_at"` *(Option 2)* | `""` | `updated_at`: omitted | Legacy alias keeps created fallback in table/CSV; JSON emits raw empty string. |
| **Research**| `-c slug,updated_at` | `updated_at` | `updated_at` | `""` (empty) | `"updated_at"` | `""` | `updated_at`: omitted | Canonical selector emits raw empty value in all formats (table, CSV, JSON). |
| **Audit** | *(default table/csv)* | `open_findings` | `open` | `"3"` | N/A | N/A | `open_findings`: `3` | Default audit table retains legacy `open` header and count value. |
| **Audit** | `-c slug,open` | `open_findings` | `open` | `"3"` | `"open"` *(Option 1)* / `"open_findings"` *(Option 2)* | `"3"` | `open_findings`: `3` | Legacy alias keeps `open` header; JSON key matches selected option. |
| **Audit** | `-c slug,open_findings`| `open_findings`| `open_findings`| `"3"` | `"open_findings"` | `"3"` | `open_findings`: `3` | Canonical selector emits `open_findings` in table header, CSV header, and JSON key. |

---

### Challenge and contract analysis

#### 1. Table, CSV, and projected JSON for a never-edited task under schema 1.66

- **`-o table -c updated`**: Emits header `updated` and value `"2026-01-01"`. The caller selected the legacy human presentation alias, which historically means "last activity date" and falls back to `created`.
- **`-o csv -c updated`**: Emits header `updated` and value `"2026-01-01"`. Same as table.
- **`-o table -c updated_at`**: Emits header `updated_at` and value `""` (empty tab cell). Emitting a created date under an explicit `updated_at` header is false data; `updated_at` must represent the raw un-fallback value.
- **`-o csv -c updated_at`**: Emits header `updated_at` and value `""` (empty cell between commas). Downstream CSV consumers (Python pandas, R, SQL loaders) receive truthful timestamps.
- **`--json -c updated_at`**: Emits `{"updated_at": ""}`. Unambiguously matches the full wire DTO name and raw value semantics.
- **`--json -c updated`**: Emits empty string value `""`. Under Option 1 (key preservation), key is `"updated"`; under Option 2 (wire convergence), key is `"updated_at"`.

#### 2. SchemaVersion policy and projected JSON key renaming

`internal/wire/wire.go:21-25` states:
> "SchemaVersion is the semver of the --json payloads — ONE version for the whole CLI output schema, not per envelope (decided 2026-06-12). Adding a field bumps the minor; renaming/removing bumps the major."

Historical analysis of `wire.go` reveals:
- Versions 1.1 through 1.65 have never incremented the major version, even when dropping fields (e.g. `misfiled` in 1.26) or replacing vocabulary words (`landed` → `tracked` in 1.47).
- `schema.go:28-29` and `CLAUDE.md` explicitly document: *"A --json -c projection uses canonical wire keys with string values; only full --json validates against --json-schema."*
- Projected JSON is documented as a string-valued view, not a standalone schema-validated envelope.

There are two viable paths forward for the implementation owner:
- **Option 1 (Zero-Breakage Compatibility):** Have `ProjectedListJSON` emit the exact key requested by the caller (i.e. `-c updated` emits `"updated": ""` while `-c updated_at` emits `"updated_at": ""`). This completely avoids breaking existing callers reading `.tasks[].updated`.
- **Option 2 (Wire-Key View Convergence):** Retain `"updated_at"` unconditionally in JSON, but explicitly document in `wire.go` and `schema.go` that projection views always converge on canonical wire keys, and that callers querying `.updated` must update their jq expressions to `.updated_at`.
- **Recommendation:** **Option 1** is recommended because it requires only 2 lines in `SelectColumns` (`c.jsonName = n` when `n != canonical`), satisfies 100% backward compatibility for existing callers, and still guarantees that all newly authored scripts using `updated_at` receive canonical keys.

#### 3. Minimal abstraction for raw vs. fallback values in table/CSV

No new structs or renderer interface changes are needed. The entire capability is achieved inside `render.SelectColumns` in [`internal/cli/render/columns.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.2fZ3j7/internal/cli/render/columns.go#L85-L105):

```go
// SelectColumns resolves column names and rewires extractors:
for _, n := range names {
    c, ok := byName[n]
    if !ok {
        return nil, fmt.Errorf("%w: unknown column %q (available: %s)",
            domain.ErrValidation, n, columnNames(all))
    }
    canonical := c.selectorName()
    if prior, exists := selected[canonical]; exists {
        return nil, fmt.Errorf("%w: duplicate column %q (already selected as %q)",
            domain.ErrValidation, n, prior)
    }
    selected[canonical] = n
    if n == canonical {
        // Explicit canonical selector (e.g. "updated_at"):
        c.Name = canonical
        if c.jsonExtract != nil {
            c.Extract = c.jsonExtract // Table and CSV get raw un-fallback values!
        }
    } else {
        // Explicit legacy alias (e.g. "updated"):
        // Table/CSV retain legacy c.Name and fallback c.Extract.
        // Under Option 1, projected JSON echoes the requested key:
        c.jsonName = n
    }
    out = append(out, c)
}
```

This ensures:
1. `WriteTablePlain` and `WriteCSV` call `c.Extract(it)`, which evaluates to `c.jsonExtract` when `updated_at` was selected, outputting raw `""`.
2. When `updated` was selected, `c.Extract(it)` remains the display fallback, outputting `"2026-01-01"`.
3. Default tables (empty `names`) bypass the loop and return `all` untouched with legacy headers and fallbacks.

#### 4. Completion duplicate suppression across aliases

In [`internal/cli/listmode.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.2fZ3j7/internal/cli/listmode.go#L219-L241), `columnCompleter` only inspects literal strings against `used`.

To resolve this:
1. Extend `render.ColumnSpec` with `Aliases []string`.
2. In `render.Specs(cols)`, populate `Aliases` when `c.Name != c.selectorName()`.
3. In `columnCompleter`, build an alias-to-canonical index before filtering candidates:

```go
func columnCompleter(specs []render.ColumnSpec) completeFunc {
    aliasMap := make(map[string]string, len(specs)*2)
    for _, s := range specs {
        aliasMap[s.Name] = s.Name
        for _, a := range s.Aliases {
            aliasMap[a] = s.Name
        }
    }
    return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        parts := strings.Split(toComplete, ",")
        prefix := strings.Join(parts[:len(parts)-1], ",")
        last := parts[len(parts)-1]
        used := make(map[string]bool, len(parts)*2)
        for _, p := range parts[:len(parts)-1] {
            used[p] = true
            if canon, ok := aliasMap[p]; ok {
                used[canon] = true // Mark canonical name as used!
            }
        }
        var out []cobra.Completion
        for _, s := range specs {
            if used[s.Name] || !strings.HasPrefix(s.Name, last) {
                continue
            }
            cand := s.Name
            if prefix != "" {
                cand = prefix + "," + s.Name
            }
            out = append(out, cobra.CompletionWithDesc(cand, s.Desc))
        }
        return out, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
    }
}
```

This guarantees that typing `-c updated,` suppresses `updated_at` from completion suggestions, preventing duplicate validation errors.

#### 5. Root `--json` argv scanner specification

In [`cmd/tskflwctl/main.go`](file:///private/var/folders/16/5bk6wc255gn_1jpwz4qpyn_c0000gn/T/isolated-review.2fZ3j7/cmd/tskflwctl/main.go), do NOT call `useFang(os.Args[1:], false)` to format errors because `!stderrIsTTY` forces `useFang` to return false on all non-TTY runs, which would cause all piped/CI runs to emit prose errors.

Instead, introduce a pure scanner function `jsonFlagActive(args []string) bool` that mirrors pflag's boolean flag parsing rules:

```go
// jsonFlagActive reports whether root --json is effectively enabled in args.
// It scans raw argv with last-flag-wins precedence, matching pflag's behavior.
// `--` terminates flag scanning.
func jsonFlagActive(args []string) bool {
    active := false
    for _, a := range args {
        if a == "--" {
            break
        }
        if a == "--json" {
            active = true
            continue
        }
        if val, ok := strings.CutPrefix(a, "--json="); ok {
            if enabled, err := strconv.ParseBool(val); err == nil {
                active = enabled
            }
        }
    }
    return active
}
```

Then refactor `useFang` and `main()`:
```go
func useFang(args []string, stderrIsTTY bool) bool {
    if !stderrIsTTY {
        return false
    }
    return !jsonFlagActive(args)
}

func main() {
    root := cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr)
    if useFang(os.Args[1:], term.IsTerminal(int(os.Stderr.Fd()))) {
        // Human TTY path
        ...
        return
    }

    // Machine path
    if err := root.Execute(); err != nil {
        asJSON, _ := root.PersistentFlags().GetBool("json")
        if !asJSON {
            // If cobra halted early (e.g. unknown flag before --json), recover from argv
            asJSON = jsonFlagActive(os.Args[1:])
        }
        cli.WriteError(os.Stderr, err, asJSON)
        os.Exit(cli.ExitCode(err))
    }
}
```

Behavior verification:
- `tskflwctl task list --json --json=false` on TTY: `jsonFlagActive` returns `false`, so `useFang` returns `true` (uses fang human path).
- `tskflwctl task list --json=false --json` on TTY: `jsonFlagActive` returns `true`, so `useFang` returns `false` (bypasses fang).
- `tskflwctl --unknown --json`: Cobra halts at `--unknown`. `PersistentFlags().GetBool("json")` returns `false`, but `jsonFlagActive(os.Args[1:])` returns `true`, correctly emitting the JSON error envelope.
- `tskflwctl task new -- --json`: Literal `--json` after `--` is ignored.

#### 6. Executable schema vocabularies and L2 scope boundary

Inspection of `schema --json` confirms that the only published vocabulary registries are:
- `statuses`
- `epic_statuses`
- `thread_statuses`
- `audit_buckets`
- `finding_statuses`
- `criterion_states`

There are no published registries for `effort` or `urgency`. Demanding that `FindingJSON.effort` reference an un-published registry is scope inflation.
The only defect in L2 is that `schema_descriptions_test.go` tested `CriterionJSON.reason` while omitting `CriterionJSON.state`.
- **Repair:** Add `{"CriterionJSON", "state", "criterion_states"}` to the test table in `TestDispositionDescriptionsReferencePublishedVocabularies`.

#### 7. Schema changelog ascending guard (L3)

Repository history confirms that schema minor versions represent independent feature additions merged chronologically, but strict contiguous numbering (`+1`) has never been a documented requirement.
The test's purpose is to prevent concurrent branches from inserting entries out of ascending order.
- **Repair:** Keep `TestSchemaVersionChangelogIsAscending` focused on ascending monotonicity. Upgrading the regex to handle `2.0` is a follow-up hygiene task when a major version is actually scheduled.

#### 8. Column parity across entities (L4)

Repository-wide inspection proves that **no entity's column registry has parity with its full wire DTO**:
- `TaskColumns` omits 6 of its 14 DTO fields (`effort`, `autonomy_level`, `created`, `tags`, `depends_on`, `id`).
- `EpicColumns` omits `created`, `tags`, and `liveness`.
- `AuditColumns` omits 6 DTO fields.
Column registries are intentionally curated for compact terminal list views, not DTO serialization.
- **Verdict on L4:** **Rejected**. Adding `updated_at` to `AuditColumns` is an additive feature that can be considered in a separate task, not a defect in this PR.

---

### First-round findings disposition recommendation

| Finding Code | Parent Severity | Design Review Disposition | Core Evidence & Justification |
| :--- | :--- | :--- | :--- |
| **H1** | **High** | **Confirmed but remediation revised** | Emitting `created` fallback under canonical `updated_at` header in CSV/table is a data error. Resolved cleanly in `SelectColumns` by setting `c.Extract = c.jsonExtract` when `n == canonical`. |
| **M1** | **Medium** | **Confirmed but remediation revised** | Key renaming breaks `.tasks[].updated` jq callers. Resolved by setting `c.jsonName = n` when `n != canonical` (Option 1), preserving legacy keys while delivering raw value semantics. |
| **M2** | **Medium** | **Confirmed** | Shell completion suggests `updated_at` when `updated` is present. Resolved by adding `Aliases` to `ColumnSpec` and filtering canonical names in `columnCompleter`. |
| **L1** | **Low** | **Confirmed but remediation revised** | Parent proposal `!useFang(..., false)` fails on non-TTY. Resolved by decoupling pure scanner `jsonFlagActive` from TTY checks. |
| **L2** | **Low** | **Confirmed but remediation revised** | Add `{"CriterionJSON", "state", "criterion_states"}` to test table. Reject expanding to `effort`/`urgency` which lack published registries. |
| **L3** | **Low** | **Rejected as defect; follow-up hygiene** | Version continuity (`+1`) is not an established invariant. Existing monotonicity test is sufficient for 1.x. |
| **L4** | **Low** | **Rejected** | Parity between list columns and full wire DTOs is not promised anywhere in `taskflow` and contradicted by all other entity registries. |

---

### Bounded implementation plan

The entire remediation requires edits to exactly four production/test files, preserving clean separation of concerns:

#### Step 1: Fix table/CSV canonical extraction and JSON key preservation in `columns.go`
In `internal/cli/render/columns.go:99-105`:
```go
if n == canonical {
    c.Name = canonical
    if c.jsonExtract != nil {
        c.Extract = c.jsonExtract // Raw un-fallback values for table & CSV!
    }
} else {
    c.jsonName = n // Echo requested legacy selector in JSON!
}
```

#### Step 2: Prevent completion duplicate suggestions in `listmode.go`
1. In `internal/cli/render/columns.go`:
   Add `Aliases []string` to `ColumnSpec`.
   In `Specs(cols)`: populate `Aliases = []string{c.Name}` when `c.Name != c.selectorName()`.
2. In `internal/cli/listmode.go`:
   Update `columnCompleter` to index `aliasMap` and mark canonical names as used when aliases appear in `parts`.

#### Step 3: Fix argv scanner and error handling in `main.go`
In `cmd/tskflwctl/main.go`:
1. Add pure `jsonFlagActive(args []string) bool`.
2. Refactor `useFang` to return `!stderrIsTTY || !jsonFlagActive(args)`.
3. In `main()`, recover `asJSON = jsonFlagActive(os.Args[1:])` when `root.Execute()` returns an error.

#### Step 4: Add missing `CriterionJSON.state` check in `schema_descriptions_test.go`
In `internal/wire/schema_descriptions_test.go`:
Add `{"CriterionJSON", "state", "criterion_states"}` to `TestDispositionDescriptionsReferencePublishedVocabularies`.

#### Step 5: Update tests
1. In `internal/cli/render/columns_test.go`:
   - Assert that `SelectColumns(cols, []string{"updated_at"})` produces raw `""` for never-edited tasks in table and CSV.
   - Assert that `SelectColumns(cols, []string{"updated"})` produces `"2026-01-01"` for never-edited tasks in table and CSV.
   - Assert that `columnCompleter` suppresses `updated_at` when `updated` is in prefix.
2. In `cmd/tskflwctl/main_test.go`:
   - Add test cases for `jsonFlagActive` covering `--json --json=false`, `--json=false --json`, and preceding unknown flags.

---

### Residual uncertainty

1. **Option 1 vs. Option 2 for JSON Keys:** The implementation owner should explicitly choose between:
   - **Option 1 (Recommended):** Echoing the requested selector name in JSON (`c.jsonName = n`), preserving 100% backward compatibility for `.tasks[].updated` callers.
   - **Option 2:** Enforcing canonical wire keys (`updated_at`) across all JSON projections and documenting the breaking change in the schema 1.66 release notes.
2. **Timing of Release:** Because `6g9s401jtqb7` touches core machine contract surfaces, this remediation should be committed and verified with container release validation before cutting the final candidate.
