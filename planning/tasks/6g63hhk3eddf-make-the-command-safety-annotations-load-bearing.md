---
schema: 1
id: 6g63hhk3eddf
status: completed
epic: 21-code-quality-architecture-hardening
description: '77 commands carry safety annotations that nothing reads: expose them in schema --json and make them verify something'
effort: 3-5 hours
tier: 2
priority: medium
autonomy_level: 3
tags: [cli, agents, schema, architecture]
created: "2026-09-02"
audited: "2026-09-13"
updated_at: "2026-09-22"
started_at: "2026-09-21"
completed_at: "2026-09-22"
---
# Make the command safety annotations load-bearing

## Objective

Every CLI command carries `Annotations{"safety": "read-only"|"mutating"}` — 76 of
them today (41 read-only, 35 mutating) — and **nothing reads them**. There is no
consumer anywhere in `internal/`, no test asserting they exist, and `schema` does
not emit them.

`planning/research/6f9menr01t1n-tskflwctl-command-spec.md:278` marks the tagging
work complete, and line 248 states what it was for: an agent should introspect
safety tags rather than scrape `--help` prose. That consumer was never built.

A hand-maintained convention with no reader can only rot — a new command that
omits the tag, or tags itself wrong, is invisible today. Give the tag a consumer,
then let it enforce something. The end state worth aiming at is that the tag
*does* work rather than describing it: a `read-only` command that can reach a
mutating path should be a test failure, not a code-review catch.

## Acceptance criteria

- [x] `schema --json` emits the command surface with each command's safety tag, and a golden pins the output
- [x] A test asserts every registered command — including hidden and deprecated ones — carries a recognized `safety` value
- [x] The tag gates or verifies something rather than only describing it: a `read-only` command that reaches a mutating service path fails a test
- [x] Decide and record whether `--dry-run` applicability should derive from the tag rather than being restated per command
- [x] Agent-facing guidance (`schema`, and the CLAUDE.md triage section) tells an agent the tag exists and is machine-readable

## Out of scope

- Reclassifying any command's read-only/mutating value — audit values only where a new test disagrees with reality
- A general command-metadata registry: this is one field gaining one consumer, not a new descriptor layer
- Surfacing the tag in human `--help` prose

## Related

- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
- `planning/research/6f9menr01t1n-tskflwctl-command-spec.md` — command safety tags, the original intent
- Audit [AI-agent CLI ergonomics, M3](../audits/6fsa47r4f7es-2026-07-24-ai-agent-cli-ergonomics.md)
- Thread [Make CLI contracts self-describing and bounded](../threads/6gbn4v2jpf2m-make-cli-contracts-self-describing-and-bounded.md)

## Sweep audit 2026-09-13

Automated weekly sweep. The premise re-verified against `main` (`1ca31b9`).

**Still true, and the count has moved the way the task predicted.** There are now **77**
annotated commands (41 `read-only`, **36** `mutating`), up from the 76 (41/35) recorded
on 2026-09-02. A new mutating command arrived in the intervening eleven days and — as the
Objective argues — nothing noticed, because there is still nothing to notice with.
`grep -rn 'Annotations\[' internal/` returns **zero** hits: no consumer anywhere, no test
asserting the values, and `schema` still does not emit them (`internal/cli/schema.go:46`
is itself only another *definition* site). `description` updated 76 → 77.

That drift is the cheapest available evidence for acceptance criterion 2 — a convention
whose population changed silently in under two weeks is exactly the rot the task names.

**Verified accurate (2026-09-13):** both citations into
`planning/research/6f9menr01t1n-tskflwctl-command-spec.md` still land where the task says.
Line 248 is the `schema --type cli --json` intent ("emit the command tree … safety tags
… so an agent introspects syntax instead of scraping `--help`"); line 278 is the `✅
Command safety tagged via cobra Annotations` completion marker.

## Progress log

- 2026-09-13: automated weekly sweep — annotation count drifted 76 → 77 (41 read-only, 36 mutating) with still zero consumers, which is the rot the task predicts; `description` updated.

## Implementation closeout (2026-09-21)

Command safety is now an executable contract rather than a convention. The final runnable Cobra leaf is bound before discovery; filesystem, configuration, space-registry, and workspace-opening adapters receive a framework-neutral mutation authorizer and fail closed when a read-only invocation reaches either a real or dry-run mutation path. Direct init, TUI-launch, and durable Thread-plan writes enforce the same capability. The audit caught and corrected `lint`: because `--fix` can persist repairs, its command capability is mutating even when a particular invocation only reads.

`schema --json` revision 1.71 additively publishes 100 deterministic command records (49 read-only, 51 mutating) with hidden/deprecated markers. Coverage includes ordinary, hidden, deprecated, generated help/completion, and Cobra's runtime-only completion transport; a real CLI/core/filesystem probe proves a mislabeled read-only command cannot alter a task. Safety does not derive `--dry-run` support: it describes potential side effects, while previews remain command-specific because interactive mutating commands cannot preview meaningfully.

Validation: the full race-enabled Go suite passes, golangci-lint reports zero issues, generated CLI/schema artifacts and machine-contract goldens are current, planning lint passes, and `git diff --check` is clean.

## Adversarial review closeout (2026-09-22)

Claude found two medium and three low gaps; all five are fixed and both implementation audits are closed. The CLI now carries authorization through workspace-opened stores, Thread creation authorizes before touching the filesystem, and Thread compose authorizes previews as well as writes. Regression coverage kills removal of all 26 filesystem mutation entries, every CLI-composed persistence family, custom pre-run binding, unbound fail-closed behavior, and direct init/compose/UI guards. The command research spec now consistently classifies whole-command capability. Antigravity found no separate defects but independently identified the workspace boundary as residual risk, corroborating Claude M1.

Final validation: `go test -race ./...`, `golangci-lint run ./...`, generated CLI/schema artifacts, `go mod tidy -diff`, planning lint, audit lint, and `git diff --check` are clean.
