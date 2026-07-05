---
status: completed
epic: 17-pm-go-cli
description: Mostly off-the-mark/too-late review of package boundaries; documented the rationale in ARCHITECTURE instead of churning a peer-reviewed design
effort: Unknown
tier: 4
priority: low
autonomy_level: 3
tags: [pm-tooling, architecture, docs]
created: "2026-06-09"
updated_at: "2026-06-09"
completed_at: "2026-06-09"
id: 6fakbec02ctt
---

# Triage architecture-collapse review (deferred: relitigates settled multi-adapter design)

## Objective

A review argued the package split is over-fragmented (collapse packages, drop
the `core.Store` interface, use concrete types). Verdict: **mostly off the mark
and too late** — it relitigates a deliberate, working, peer-reviewed design and
mis-frames it as "a CLI" when it's a **multi-adapter system (CLI now, TUI next)**.

## Disposition (no code change; rationale documented)

- **Factual errors (verified):** "tests don't use mocks" — false, `core/
  service_epic_test.go` tests against a `fakeStore`. "frontmatter logic scattered
  across packages" — false, `frontmatter.go`/`fix.go`/`diagnose.go` are all one
  package (`store`); `domain/validate.go` is semantic rules, correctly separate.
- **The TUI settles #3 and #5's premise:** a Bubble Tea TUI is a planned second
  primary adapter over the same `core`, so the `core.Store` seam and the
  presentation/logic split are *load-bearing*, not speculative.
- **One fair-but-lateral point (#5):** `cli/render` is genuinely cli-only and
  has a mild `cli→render→core` diamond. Folding it into `cli` is defensible but
  marginal; left split (the TUI replaces render, reuses core). Not worth churn.
- **Honest warts acknowledged:** `FixFrontmatter` sits awkwardly on the `Store`
  port (a `Fixer`-split candidate); the render→core import is a minor diamond.

Documented all of this in `docs/ARCHITECTURE.md` → "Why these boundaries (and
why not collapse them)" so the question stops reopening.

## Acceptance

- [x] Claims verified against the code; ARCHITECTURE decision note added.
- [x] No churn to a settled, working design.

## Related

- Epic [17-pm-go-cli](../epics/17-pm-go-cli.md); the prior idiom-review triage is
  [triage-go-idiom-review-typed-validators-status.isactive-date-format](6fakbec00ahf-triage-go-idiom-review-typed-validators-status.isactive-date-format.md).
