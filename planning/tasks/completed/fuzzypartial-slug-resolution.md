---
status: completed
epic: 17-pm-go-cli
description: Resolve task/audit/epic slugs by unambiguous prefix or substring, listing candidates when ambiguous, to complement tab-completion
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: "2026-06-09"
updated_at: "2026-06-13"
started_at: "2026-06-13"
completed_at: "2026-06-13"
id: 6fakbec0252f
---

# Fuzzy/partial slug resolution

## Objective

Keyboard economy companion to tab-completion: when you half-remember a slug,
let an unambiguous prefix/substring resolve it. `task show retry` → finds
`add-retry-backoff`. Complete when you can, fuzzy-match when you can't.

## Implementation sketch

- [x] Extend `store.resolve` (and `resolveAudit`, epic `GetEpic`): on an exact
      miss, try (a) unique prefix match, then (b) unique substring match across
      the candidate set.
- [x] **Ambiguity is explicit:** >1 match → `ErrAmbiguous` listing the
      candidates (the error already names locations; extend to list slugs).
      Exact match always wins over fuzzy.
- [x] Apply consistently to every slug-taking command (show/set/move/verbs,
      audit show/close/…, epic show). Mutating commands resolve the same way.
- [x] Keep it deterministic (sorted candidates) and case-insensitive.
- [x] Tests: exact wins; unique prefix; unique substring; ambiguous → error with
      candidate list; no-match → not-found.

## Out of scope

- Typo/edit-distance fuzzing (Levenshtein) — prefix/substring only, to avoid
  surprising matches. Reconsider only if it feels needed in real use.

## Related

- Epic [[17-pm-go-cli]]; complements completion and the color work
  [[cli-color-glyphs-table-headers-render-styling]].

## Note (2026-06-12)

A draft task ([[interactive-prompt-layer-gh-style-pickers]]) proposes a
TTY-only picker over this task's ambiguous-candidate output. **This task owns
the resolution semantics** (exact > prefix > substring, explicit ErrAmbiguous)
— the draft is presentation only and defers here; sequence this first.

## Closure (2026-06-13)

Implemented as one shared matcher (`store/resolve.go: resolveID`) used by all
three resolvers (`resolve`, `resolveAudit`, `GetEpic`), so every slug-taking
command — show/set/move/verbs, audit show/close/…, epic show, and the TUI via
the same service — resolves identically: exact (case-insensitive) > unique
ci-prefix > unique ci-substring; ambiguity is ErrAmbiguous listing sorted,
located candidates ("slug (status)"); queries with path separators/.. are
ErrValidation (absorbing the slug-sanitization item from the discovery task).
One trap caught during design: Move/MoveAudit built the destination filename
from the QUERY — fixed to use the resolved canonical slug, with a test pinning
that `move backoff` cannot rename the file to backoff.md. Live-verified
against the planning repo (substring show, 16-way ambiguous listing, epic
fuzzy, traversal rejection → exit 11).
