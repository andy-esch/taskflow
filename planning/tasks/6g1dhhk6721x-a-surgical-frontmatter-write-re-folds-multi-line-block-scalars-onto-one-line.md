---
schema: 1
id: 6g1dhhk6721x
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: updateFrontmatter preserves keys/comments/order but not a multi-line block scalar's wrapping, so any frontmatter write reflows unrelated values.
effort: Unknown
tier: 4
priority: low
autonomy_level: 3
tags: [store, frontmatter]
created: "2026-08-18"
updated_at: "2026-10-05"
audited: "2026-09-27"
audit_sources: [2026-09-27-weekly-task-sweep, 2026-10-05-arch-data-model-and-storage]
---

# A surgical frontmatter write re-folds multi-line block scalars onto one line

## Objective

`updateFrontmatter` preserves unknown keys, comments, and key order — but it does NOT
preserve the LINE WRAPPING of a multi-line YAML block scalar. Editing any field in a file
that has one silently reflows that unrelated value onto a single long line.

## Reproduction

Observed while backfilling research descriptions. A single
`research set <slug> --description "..."` on a doc carrying a folded scalar produced:

    purpose: >-
    -  Synthesize ADR best practices, cross-tool Project/initiative product research, and
    -  the two repos' house style into generic ADR + Project templates tskflwctl will
    -  scaffold, plus a cross-linking scheme. The decision record behind ADR-0001 (ADRs)
    -  and ADR-0002 (Projects).
    +  Synthesize ADR best practices, cross-tool Project/initiative product research, and the two repos' house style into generic ADR + Project templates tskflwctl will scaffold, plus a cross-linking scheme. The decision record behind ADR-0001 (ADRs) and ADR-0002 (Projects).

**No data is lost** — a `>-` folded scalar joins its lines with spaces, so the value is
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

- [ ] A surgical field write leaves an untouched multi-line block scalar byte-identical,
      including its wrap width and its chomping indicator (`>-` vs `>` vs `|`).
- [ ] Test fixture covers `>-`, `>`, and `|` alongside a normal scalar edit.
- [ ] If exact preservation isn't achievable through the current yaml.Node round-trip,
      the fallback is to leave a node untouched when its value is unchanged, rather than
      re-emitting it.

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
Net priority judgment is deliberately left to a human.

## Progress Log

- 2026-09-27: automated weekly sweep — reproduced the refold on HEAD in an isolated tree; found the blast radius has grown from 3 `set` verbs to 13 call sites including every lifecycle verb and `task edit`, confirmed via `task start` on a probe file; noted the chomping indicator is already preserved and that only one corpus file still carries a block scalar.

Reinforced by audit 2026-10-05-arch-data-model-and-storage: H1. That finding is a PARTIAL overlap, not a duplicate — it measures a case this task's `## Reproduction` does not cover. For a `>` folded scalar whose body carries more-indented lines, each surgical write appends one newline to the *decoded* value (41 → 46 bytes over five `task set` calls, read back through `go.yaml.in/yaml/v3`), so the "**No data is lost**" premise above holds only for the uniformly-indented `>-` case it reproduces. Anchored to the open upstream defect yaml/go-yaml#337. AC-1 as already written ("byte-identical, including its wrap width and its chomping indicator") would cover both halves; the audit's note is about the severity assessment, not the scope. Propose-only — no field on this task was changed beyond this annotation and `audit_sources`.
