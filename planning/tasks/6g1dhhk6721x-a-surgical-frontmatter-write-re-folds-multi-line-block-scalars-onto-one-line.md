---
schema: 1
id: 6g1dhhk6721x
status: in-progress
epic: 21-code-quality-architecture-hardening
description: Preserve untouched YAML block scalars during frontmatter edits; more-indented folded values currently accumulate decoded newlines.
effort: 4-8 hours
tier: 2
priority: high
autonomy_level: 3
tags: [store, frontmatter]
created: "2026-08-18"
updated_at: "2026-10-06"
audited: "2026-09-27"
audit_sources: [2026-09-27-weekly-task-sweep, 2026-10-05-arch-data-model-and-storage]
started_at: "2026-10-06"
---

# Preserve untouched YAML block scalar values and formatting

## Objective

`updateFrontmatter` must preserve untouched scalar values, not just keys, comments, and
order. Uniformly indented folded scalars currently reflow; more-indented folded scalars
can silently accumulate decoded newlines. Fix both at the shared write boundary, with
decoded-value fidelity as the first safety invariant and wrapping preservation as the
git-native presentation invariant.

## Reproduction

Observed while backfilling research descriptions. A single
`research set <slug> --description "..."` on a doc carrying a folded scalar produced:

    purpose: >-
    -  Synthesize ADR best practices, cross-tool Project/initiative product research, and
    -  the two repos' house style into generic ADR + Project templates tskflwctl will
    -  scaffold, plus a cross-linking scheme. The decision record behind ADR-0001 (ADRs)
    -  and ADR-0002 (Projects).
    +  Synthesize ADR best practices, cross-tool Project/initiative product research, and the two repos' house style into generic ADR + Project templates tskflwctl will scaffold, plus a cross-linking scheme. The decision record behind ADR-0001 (ADRs) and ADR-0002 (Projects).

**This uniformly indented example loses no data** — a `>-` folded scalar joins its lines with spaces, so the value is
byte-identical (verified by round-tripping both versions through the YAML parser and
comparing). The cost is a noisy, misleading diff: a commit that claims to change one
field also rewrites an unrelated multi-line value, and the file gets less readable.

## Scope — NOT research-specific

`updateFrontmatter` is shared by `task set`, `epic set`, and `research set`, so any entity
with a folded/literal block scalar is affected. Research is only where it surfaced,
because the legacy corpus carries a `purpose: >-` block. Nothing the tool itself WRITES
uses a block scalar today, which is why this went unnoticed — it only bites hand-authored
frontmatter.

## Acceptance criteria

- [x] A surgical field write leaves an untouched multi-line block scalar byte-identical,
      including its wrap width and its chomping indicator (`>-` vs `>` vs `|`).
- [x] Test fixture covers `>-`, `>`, and `|` alongside a normal scalar edit.
- [x] If exact preservation isn't achievable through the current yaml.Node round-trip,
      the fallback is to leave a node untouched when its value is unchanged, rather than
      re-emitting it.
- [x] Repeated unrelated writes preserve decoded values of >, >-, and | scalars
  with blank and more-indented lines; regression assertions distinguish value
  corruption from formatting churn.

## Out of scope

- Re-wrapping values the tool itself writes (it emits single-line scalars by design).

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- Sibling frontmatter-fidelity guarantee: CRLF preservation (see `detectLineEnding` in
  `internal/store/frontmatter.go`), which IS handled — this is the same class of promise

## Sweep verification (2026-09-27)

Automated weekly sweep. The bug **still reproduces on HEAD** and no acceptance criterion
was ticked. Sections above are left intact; what follows records where the ground has
moved since 2026-08-18.

### Reproduced, in an isolated tree

Ran `research set --description` against a throwaway planning space holding a copy of the
very doc named in `## Reproduction`
(`planning/research/6fe4my001bdk-adrs-and-projects-format-design.md`). The four wrapped
lines of `purpose: >-` collapsed onto one, exactly as the diff above predicts. The
`## Objective` and `## Reproduction` sections are accurate as written.

One refinement to acceptance criterion 1, which asks for preservation *"including its wrap
width and its chomping indicator (`>-` vs `>` vs `|`)"*: the **chomping indicator is
already preserved** — the rewritten file still opens `purpose: >-`. Only the wrap is lost.
The criterion is left exactly as-is (a test asserting the indicator survives is still
worth having), but the implementer should expect the indicator half to pass on day one.

### `## Scope` understates the blast radius: 3 entry points → 13 call sites

The scope section says *"`updateFrontmatter` is shared by `task set`, `epic set`, and
`research set`"*. Today it has **13 non-test call sites across 10 files**, and the new ones
are not `set` verbs at all:

| Call site | Reaches the user as |
| --- | --- |
| `store/lifecyclemutation.go:209` | **every lifecycle verb** — `start`/`next`/`ready`/`complete`/`defer`/`deprecate` |
| `store/edit.go:109` | `task edit` stamping `updated_at` after a human's `$EDITOR` save |
| `store/fix.go:229`, `:316` | `lint --fix` id repair and backfill |
| `store/graphmutation.go:158`, `:166` | `task depend add`/`remove` |
| `store/auditstore.go:104` | `audit close`/`reopen`/`defer` (bucket moves) |
| `store/threadmutation.go:174` | Thread frontmatter writes |
| `store/fsstore.go:220`, `epicstore.go:109`, `:169`, `researchstore.go:144` | the three `set` verbs originally named |

Confirmed the widest of these by experiment, not inspection: `task start` on a probe task
carrying a `notes: >-` scalar reflowed it while moving `ready-to-start -> in-progress`. So
the promise being broken is not "a surgical *field* write" but **any frontmatter write at
all**, including two paths where the user never named a field: a lifecycle verb and a
human's own editor save. `description` has been tightened accordingly.

`lint --fix` was also checked and is clean — it reported "nothing to fix" and left the
scalar alone, because it only rewrites files it has something to repair.

### Counter-pressure on priority: the in-repo corpus is down to one file

Cutting the other way, and the reason `tier: 4` / `priority: low` are left untouched: a
scan of all frontmatter under `planning/` found **exactly one** file carrying a block
scalar — `6fe4my001bdk-adrs-and-projects-format-design.md`, the doc in `## Reproduction`.
`## Scope`'s observation that *"nothing the tool itself WRITES uses a block scalar today"*
still holds, so the corpus is not accumulating exposure. The widened call-site count raises
the *conditional* probability of hitting this; the shrinking corpus lowers the *base* rate.
That assessment was superseded by the decoded-value corruption reproduced below; this
is now high-priority tier 2 work, not formatting-only polish.

## Progress Log

- 2026-09-27: automated weekly sweep — reproduced the refold on HEAD in an isolated tree; found the blast radius has grown from 3 `set` verbs to 13 call sites including every lifecycle verb and `task edit`, confirmed via `task start` on a probe file; noted the chomping indicator is already preserved and that only one corpus file still carries a block scalar.

Reinforced by audit 2026-10-05-arch-data-model-and-storage: H1. That finding is a PARTIAL overlap, not a duplicate — it measures a case this task's `## Reproduction` does not cover. For a `>` folded scalar whose body carries more-indented lines, each surgical write appends one newline to the *decoded* value (41 → 46 bytes over five `task set` calls, read back through `go.yaml.in/yaml/v3`), so the "**No data is lost**" premise above holds only for the uniformly-indented `>-` case it reproduces. Anchored to the open upstream defect yaml/go-yaml#337. AC-1 as already written ("byte-identical, including its wrap width and its chomping indicator") would cover both halves; the audit's note is about the severity assessment, not the scope. Propose-only — no field on this task was changed beyond this annotation and `audit_sources`.

### Boundary closeout triage (2026-10-05)

Independently reproduced in a disposable clone through the actual `updateFrontmatter`
and YAML reader: five unrelated `tier` writes changed a more-indented `>` rationale
from 41 to 46 decoded bytes. This is an existing shared storage defect, not an
adapter-neutral boundary regression. H1 is tracked here; the task remains ready to
start in epic 21, outside the adapter-neutral Thread. Prefer this as a near-term
safety followup before the next release, without reopening the completed port migration.
No claim is made that the live planning corpus contains an affected value.

Require decoded-value equality across repeated unrelated writes for `>`, `>-`, and `|`,
including blank and more-indented lines, alongside byte/wrapping checks. If faithful
editing is infeasible, explicitly evaluate a fail-closed diagnostic rather than
silently corrupting intent; do not rely on a raw-string diff alone.

## Implementation evidence (2026-10-06)

Existing-document encoders now share source-preserving assembly for block-scalar
documents. Unchanged top-level entries retain original source (including nested
scalars, anchors, comments, indentation, chomping, and LF/CRLF). Edited output must
decode to the requested mapping; unsupported or unsafe layouts fail before writing.

Regression matrix: eight folded/literal/chomping/indent indicators, both line endings,
five unrelated writes, body timestamping, and dependency dedupe. Real filesystem
field writes preserve decoded values and a broken-anchor replacement leaves the
original file untouched. Sibling deletion/replacement also pins comment ownership.
The original encoder failed the preservation matrix; the new encoder passes.

Shared coverage: updateFrontmatter, updateDependencySourceEdits, and replaceBodyWith;
fresh file creation still uses the normal encoder. Race suite, lint, build,
module tidiness, and planning/audit lint passed. Await independent review before merge.

## External review handoff (2026-10-06)

Prepared two independent audits with no findings:
[Codex](../audits/6gh86jxj4sve-2026-10-06-shared-write-and-audit-safety-implementation-codex.md)
and [Antigravity](../audits/6gh86jxtyx5k-2026-10-06-shared-write-and-audit-safety-implementation-antigravity.md).

Codex leads init/audit/machine-contract checks; Antigravity leads YAML preservation.
Both cross-check the other lens and must use independent dirty-state-capturing sandboxes,
bounded compiler-valid mutation evidence, and guarded one-audit transfer. Review has
not run; implementation remains in-progress pending owner triage.

## Review triage (2026-10-07)

Codex completed an independent dirty-snapshot review, including repeated real filesystem
scalar writes and byte-identical refusal checks; it found no YAML preservation defect.
Its two init/audit findings are fixed with production-path regressions and recorded
resolutions. Final race suite, lint, build, generated-output comparisons, and planning
lint pass. Antigravity's preservation-led report is still pending, not a clean verdict;
the task remains in-progress while the reviewed implementation proceeds to a PR.
