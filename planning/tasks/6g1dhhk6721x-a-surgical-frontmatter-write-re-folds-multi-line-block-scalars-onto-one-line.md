---
schema: 1
id: 6g1dhhk6721x
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: updateFrontmatter preserves keys/comments/order but not the wrapping of multi-line block scalars, so editing one field reflows unrelated values; affects task/epic/research set.
effort: Unknown
tier: 4
priority: low
autonomy_level: 3
tags: [store, frontmatter]
created: "2026-08-18"
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
