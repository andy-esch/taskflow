---
status: ready-to-start
epic: 17-pm-go-cli
description: Resolve task/audit/epic slugs by unambiguous prefix or substring, listing candidates when ambiguous, to complement tab-completion
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [pm-tooling, go, cli, ergonomics]
created: "2026-06-09"
---

# Fuzzy/partial slug resolution

## Objective

Keyboard economy companion to tab-completion: when you half-remember a slug,
let an unambiguous prefix/substring resolve it. `task show retry` → finds
`add-retry-backoff`. Complete when you can, fuzzy-match when you can't.

## Implementation sketch

- [ ] Extend `store.resolve` (and `resolveAudit`, epic `GetEpic`): on an exact
      miss, try (a) unique prefix match, then (b) unique substring match across
      the candidate set.
- [ ] **Ambiguity is explicit:** >1 match → `ErrAmbiguous` listing the
      candidates (the error already names locations; extend to list slugs).
      Exact match always wins over fuzzy.
- [ ] Apply consistently to every slug-taking command (show/set/move/verbs,
      audit show/close/…, epic show). Mutating commands resolve the same way.
- [ ] Keep it deterministic (sorted candidates) and case-insensitive.
- [ ] Tests: exact wins; unique prefix; unique substring; ambiguous → error with
      candidate list; no-match → not-found.

## Out of scope

- Typo/edit-distance fuzzing (Levenshtein) — prefix/substring only, to avoid
  surprising matches. Reconsider only if it feels needed in real use.

## Related

- Epic [[17-pm-go-cli]]; complements completion and the color work
  [[cli-color-glyphs-table-headers-render-styling]].
