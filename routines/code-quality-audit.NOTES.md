# Code Quality Audit — Maintainer Notes

Human-facing reference for the `code-quality-audit` routine. None of this is read
or executed by the agent during a run — it lives here so the operational spec
(`code-quality-audit.md`) stays lean in agent context.

## Bootloader (set this in claude.ai/code/routines)

> You are running the code-quality-audit routine. Your full instructions live at
> `taskflow/routines/code-quality-audit.md` on `main`. Read that file first,
> then execute the workflow in its "Prompt" section. The schedule, lens
> rotation, lens scopes, audit-file template, and Slack format are all canonical
> there — this routine config is only a thin bootloader and may be stale
> relative to the markdown file.

That's the entire bootloader.

## Setup checklist (one-time)

- **Repos:** `andy-esch/taskflow` only, with write access.
- **Sandbox image:** must ship a **Go toolchain** (Step 0 builds the tool) and,
  for the `test-rigour` lens, enough headroom to run `go test -race ./...` —
  currently ~20s wall clock across all packages.
- **`golangci-lint` is NOT required.** The spec deliberately validates with
  `just build` + `just test` + the two `tskflwctl` linters, because
  `golangci-lint` needs a v2 binary built with a matching Go toolchain and that
  is a fragile thing to depend on in a fresh sandbox. If you later bake it into
  the image, add `golangci-lint run ./...` to step 13.
- **Network:** Trusted, plus WebSearch/WebFetch for the conditional research
  step.
- **Slack:** `#planning-updates`. Same caveat as the other routines — copied
  from the desirelines setup; give taskflow its own channel if the volume
  warrants.

## Operational notes

- **Schedule**: `0 10 * * 2,5` UTC = 6am EDT Tuesdays and Fridays. Deliberately
  off Monday (architecture audit) and Sunday (task sweep).
- **Model**: Opus-class. The lenses that pay off most — `concurrency`,
  `test-rigour`, `contract` — need whole-file reasoning across the hexagonal
  seam, which is exactly where cheaper models produce plausible-but-wrong
  findings.
- **Rotation index**: `(ISO week * 2 + slot) mod 6`, slot Tue=0 / Fri=1. Six
  lenses, two runs a week → a full cycle every three weeks, all six covered
  (verified for weeks 36–41). One quirk worth knowing: the index advances by 2
  per week and 6 is even, so **each slot is locked to one parity** — Tuesday
  always draws an even index (0/2/4: correctness, contract, adapter-hygiene) and
  Friday always an odd one (1/3/5: concurrency, test-rigour, simplification).
  Coverage is complete, but a lens can never move to the other day. If you add
  or remove a lens, re-verify: an *odd* lens count makes both slots visit every
  lens, an even count keeps the parity lock. Switch to a counter file in
  `planning/audits/` if strict round-robin ever matters.
- **Tuning knobs**: `max_open_audits` (10), the file budget (~4–5), the churn
  window (30 days), the research trigger (<3 Medium+), and LENS-SCOPES itself —
  that last one is where most of the value lives, and it should be edited as the
  codebase's risk surface moves.

## Why lenses, not directories

desirelines rotates `weekly-code-audit` by **area** (packages/web, packages/
dispatcher, …) because it is a genuine multi-service system where a service is a
meaningful unit of review.

taskflow is one binary with a hexagonal layering. Its interesting defects are
almost never contained in one directory — they are a `core` invariant, the
`store` guard that enforces it, and the `wire` type that publishes it,
disagreeing with each other. Rotating by directory would keep splitting those
three apart. Rotating by lens keeps them together and gives the run a question
rather than a location.

The cost: the same file gets read on multiple lenses. That's intended — a file
read for concurrency and a file read for contract are different reads.

## Guard-coverage tracing (test-rigour lens)

The `test-rigour` lens's technique is deliberately **read-only**: enumerate the
refusal predicates in a package, then grep the tests for a case that constructs
the state which trips each one and asserts the refusal.

An earlier draft used a *mutation probe* — temporarily invert a predicate, run
the suite, record which tests fail. That is genuinely the highest-signal
technique for this codebase (a green suite proves less than it looks like it
does, because most of taskflow's contract lives in tests rather than types), and
it is what a human reviewer should reach for. It was cut because it requires
writing to `internal/`, and this routine's one hard rule is that it writes only
to `planning/`. A scheduled job that edits source "just briefly" is one crash or
one confused turn away from committing a mutant.

The cost is real and the spec says so: a finding the agent cannot settle by
reading must be labelled "unverified — needs a probe" with the probe spelled out
for a human to run. Watch for those labels — if most test-rigour findings carry
one, the lens isn't earning its slot in read-only form and should be replaced by
a `just`-recipe probe harness that the human runs on demand instead.

## Relation to the other routines

- `weekly-architecture-audit` asks "is the system shaped right?"; this one asks
  "is this code right?". If both keep surfacing the same finding, the boundary
  is wrong — narrow this routine's lens or broaden the architecture lens.
- `weekly-task-sweep` verifies *tasks against code*; this verifies *code against
  itself*. Overlap shows up as the sweep's "candidate new tasks" duplicating an
  audit finding. That's benign; the cross-reference step catches it.
- The `simplification` lens folds in what desirelines runs as a separate
  `simplification-audit` routine. One lens in six is the right weight for a
  codebase this size; promote it to its own routine only if it consistently
  produces more than the other five combined.

## Open design questions

Revisit after ~6 runs (each lens has run once):

- **Lens taxonomy** — are six right? Candidates not included: security posture
  (thin surface: a local CLI with no network I/O and no credentials, which is
  why it was left out), performance (no known hot path), dependency hygiene
  (`go mod tidy -diff` + `govulncheck` already cover it in CI).
- **Backpressure threshold** — 10 was inherited from desirelines, which has a
  much larger planning tree. taskflow currently carries 1–3 open audits; 10 may
  be far too loose. Consider 5.
- **Random pick value** — measure whether the random file ever produces a
  finding. If it doesn't after six runs, drop it and read the adjacency hop more
  deeply instead.
- **Findings-per-run** — if runs consistently produce 0–1 Medium+, the lens
  scopes are too narrow or the codebase is genuinely clean. Either is worth
  knowing; the no-op log is the measurement.

## Changelog

- **v1 (2026-08-30)** — initial spec, adapted from
  `desirelines-planning/routines/weekly-code-audit.md` v7 with the
  `simplification-audit.md` v4 rules folded in as one lens. Changes for
  taskflow: single-repo (no install script); **lens rotation replaces area
  rotation**; explicit propose-only scope block; `tskflwctl audit finding
  --status "tracked by <task-id>"` replaces the hand-written `superseded by
  <link>` convention;
  validation extended to `just build` / `just test` / `audit lint`; and
  taskflow-native lens scopes (hexagonal seam leaks, atomic-write discipline,
  schema_version semantics, guard-coverage tracing). The SCOPE block reduces to
  one rule — write to `planning/` only — with a `git status --porcelain` bounds
  check before the PR, and the `test-rigour` lens uses read-only guard-coverage
  tracing rather than a source-mutating probe.
