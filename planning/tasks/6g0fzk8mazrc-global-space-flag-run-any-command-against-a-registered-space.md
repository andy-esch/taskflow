---
schema: 1
id: 6g0fzk8mazrc
status: completed
epic: 29-multi-space-planning-a-home-registry-and-the-atlas
description: --space <id> (and TSKFLW_SPACE) resolves any command against a registered repo instead of the cwd. Small change to App.resolve(); also the cross-repo handle agents lack today.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, multi-repo]
created: "2026-08-15"
updated_at: "2026-08-21"
started_at: "2026-08-21"
completed_at: "2026-08-21"
---
# Global `--space` flag: run any command against a registered space

## Objective

Give every command a handle to another planning entry point without a `cd`:
`--space <id>` (plus `TSKFLW_SPACE`) looks the local label up in the registry and discovers from **there**
instead of the cwd. Small change to `App.resolve()`, disproportionate payoff — it is
also the cross-repo handle agents currently lack.

Plausibly **the most valuable single piece of epic 29**, and the reason the CLI slice
ships before any TUI work: if `--space` plus `status --all` turns out to be most of
the value, the board's cost can be re-decided honestly.

## Notes

- Precedence: `--space` > `TSKFLW_SPACE` > cwd discovery. `-C` and `--space` together
  should be a loud error, not a silent winner — they are two answers to one question.
- An unknown id errors with `ErrNotFound` (exit 10) and lists the known ids; a
  registered-but-broken space errors with the same diagnosis text the health task
  produces, not a raw `Discover` error.
- Shell completion for `--space` off the registry (the completion funcs already do
  their own forgiving discovery, so this must not break outside a planning repo).
- A registry label addresses one exact **entry point**, not the logical planning id. Two
  labels may share `planning_id` while intentionally selecting different checkouts; both
  verify and resolve to the same planning data. Receipts should retain the selected label
  or path so the execution context is not hidden by grouping.
- Must not perturb the advisory invariant: with no `--space` and no home config,
  resolution is byte-for-byte today's.

## Acceptance criteria

- [x] `--space <id>` resolves any command against the labeled registered entry point,
  `-C`-style, even when another label shares its planning identity
- [x] `TSKFLW_SPACE` honored at lower precedence than the flag
- [x] `--space` + `-C` together is a clear error
- [x] Unknown id → exit 10 with the known ids listed
- [x] Completion offers registered ids and stays silent outside a planning repo
- [x] `just test` + `just lint` green

## Out of scope

- Cross-space *aggregation* (`status --all`) — separate task
- Any TUI switching

## Related

- Epic [29-multi-space-planning-a-home-registry-and-the-atlas](../epics/29-multi-space-planning-a-home-registry-and-the-atlas.md)
- Sketch: [6g0ajre026c6-multi-space-home-registry-and-the-atlas](../research/6g0ajre026c6-multi-space-home-registry-and-the-atlas.md)

### 2026-08-18 — this flag is now the wrong-repo write guard

Audit 2026-07-24 H1 originally shipped a separate `--expect-root` precondition. It was
**reverted on ergonomic review**: it took the *resolved planning root* while `-C` takes a
*directory to stand in*, so `-C <impl> --expect-root <impl>` failed, relative values
resolved against process cwd rather than `-C`, and `--expect-root .` failed in this repo's
own `planning/` layout. Three flags in "which repo" space, and the assertion had the worst
ergonomics of the three. Full rationale in the audit's amendment.

**The design position that replaced it: explicit selection IS the assertion.** Naming a
tree by durable id cannot resolve to the wrong one, so the hazard is removed at its source
instead of checked afterwards. That is how every comparable CLI does it — `git -C`,
`kubectl --context`, `gh --repo`, `docker --context` — none carries an assert flag.

**So this task now carries a safety obligation, not just an ergonomic one:**

- `--space` and `-C` are **mutually exclusive** — two answers to one question. Erroring on
  both is not pedantry; it is what keeps "which tree" unambiguous.
- A `--space <id>` whose registry entry no longer resolves to a planning root must be a
  **loud error, never a silent fallback to cwd discovery**. Falling back would reintroduce
  exactly the hazard this replaces.
- Consider whether `--space` should also reject an id whose recorded path has drifted from
  where it now resolves (registry staleness), rather than following it blindly.

**Until this ships there is no pre-write precondition at all** — the `workspace` read and
the receipt field cover proof, not prevention. That gap is accepted and recorded on the
audit finding; it is the main reason this task is worth doing early in slice 2.

### 2026-08-19 — the guard is now mechanical, not conventional

Identity settled: `--space <label>` selects a registry entry, and the entry's `verify_id` is
checked against the resolved repo's committed id **after** resolving. So "naming a tree
cannot resolve to the wrong one" stops being a convention and becomes an enforced check.

Verification contract (opt-in per pointer, strict once opted in):

```
entry has verify_id + target matches   → proceed
entry has verify_id + target differs   → exit 14 (ErrConflict)
entry has verify_id + target has NONE  → exit 14
entry has NO verify_id (legacy)        → today's behavior
```

This supersedes the earlier note's worry: a `--space` whose entry has drifted no longer needs
a bespoke staleness rule — the id mismatch *is* the staleness detection, and it fails loudly.
The "never silently fall back to cwd discovery" requirement stands unchanged.

## Implementation notes (2026-08-21)

Shipped global `--space` plus `TSKFLW_SPACE` through the shared `App` start-directory
seam, so ordinary CLI reads and writes, configuration commands, TUI launch context, and
entity completion all resolve the exact registered entry point. Explicit `-C` wins over
an ambient environment selection; an explicit `--space` plus `-C` is a validation error.
Unknown labels enumerate known labels, broken entries reuse `spacehealth` diagnosis,
registry verify-id mismatch exits as conflict, and explicit selection never falls back to
cwd or built-in-only template behavior.

Workspace reads and every machine mutation receipt now carry the local selector as
`workspace.space`, distinct from durable `repo_id`; the additive wire contract is schema
version 1.41. Direct and pointer labels sharing one planning identity, environment and
flag precedence, duplicate labels, missing and mismatched entries, completion, and receipt
provenance are covered. Race-enabled full tests, golangci-lint, module tidiness, planning
lint, generated docs, schema comments, goldens, diff hygiene, and real-registry read-only
smoke checks are green.
