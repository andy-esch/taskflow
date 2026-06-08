---
status: ready-to-start
epic: 17-pm-go-cli
description: Formal tskflwctl {noun} {verb} tree with per-command purpose, flags, and output — the build-to spec for the Go CLI.
effort: Unknown
tier: 2
priority: high
autonomy_level: 3
tags: [pm-tooling, go, cli, spec, architecture]
created: 2026-06-06
updated_at: 2026-06-06
---

# Phase 0.5: formal tskflwctl command hierarchy + purpose spec

## Objective

Produce the **formal `tskflwctl {noun} {verb}` command tree with
per-command purpose, flags, and output shape** — the spec the Go CLI is
built to from day one. Epic 17 mandates this before implementation: the
Python `pm` + `tests/test_pm.py` are a *starting point, not the whole
contract* (acknowledged incomplete).

## Status

First draft written: `research/2026-06-06-tskflwctl-command-spec.md` (full
noun-verb tree for task/audit/epic/adr + cross-cutting, global flags,
deliberate departures from pm, and open design questions). This task is
**iterate-to-done** on that doc.

## Scope

- [ ] Resolve the spec's open design questions: unify lifecycle verbs
      behind `task move`? keep `project` as a group or fold into filters?
      `--json` schema versioning? read-only-vs-mutating command tagging?
- [ ] Lock the global-flag set and `--json` schema conventions.
- [ ] Confirm the legacy-alias map (flat verb → `task <verb>`).
- [ ] Sanity-check against the audit surface (`bucket-audits…`) so
      `tskflwctl audit` matches what that task ships.

## Acceptance criteria

- [ ] `research/2026-06-06-tskflwctl-command-spec.md` is complete and flips
      to `status: completed` — every command has a purpose + flags + output;
      no open design questions left unresolved.
- [ ] The foundation + port tasks can build directly against it.

## Out of scope

- Implementing any commands (foundation + port tasks).
- The TUI surface (commands only; TUI is a later phase).

## Related

- Epic [[17-pm-go-cli]].
- `research/2026-06-06-tskflwctl-command-spec.md` — the spec.
- [[go-cli-foundation-layout-corestorecli-boundary-di-testlint-harness]] ·
  [[port-pm-to-go-cli-parity-with-python-prototype-test-suite-as-spec]].
