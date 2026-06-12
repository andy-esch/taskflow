---
status: reference
created: 2026-06-12
tags: [decisions, review, reference]
---

# Pending decisions from the 2026-06-12 review

Every open question blocking a queued task, with options and a recommendation.
Fill in `Decision:` lines (or reply in chat with `D1: A, D2: B, …`). Each
decision unblocks the task named with it.

## D1 — `task new` ↔ `lint` tags contract
Task: [[align-task-new-scaffold-with-lint]] (H4). A fresh scaffold fails the
tool's own lint (`tags: missing`).
- **A (recommended):** require `--tags` at creation — loud, explicit, matches
  the repo's own convention that every task is tagged.

Decision:

## D2 — Epic status vocabulary
Task: [[reject-invalid-list-filters-and-epic-statuses]] (M9, the deferred
half). `epic new --status bananas` currently writes free text.
- **A (recommended):** closed enum from the values already in use —
  `planning | in-progress | completed | archived` — validated in `NewEpic`,
  linted on existing files.

Decision:

## D3 — Exit code 12 / transition rules
Task: [[implement-or-retire-exit-code-12-transition-rules]] (M16).
`ErrInvalidTransition` is documented but unreachable.
- **A (recommended):** retire it — delete the sentinel, the exit-code mapping,
  and the README/ARCHITECTURE rows (codes 13/14 keep their numbers). Current
  dogfooding moves tasks freely; reinstate later if a real rule emerges.

Decision:

## D4 — Unknown `--set` keys
Task: [[task-set-follow-ups-sentinels-unknown-keys-canonical-field-table]].
A typo'd key is silently written today (demonstrated live), lint never flags
it, and there's no way to remove it through the tool.
- **A (recommended):** error on keys outside the known field set, with
  `--force` to write a genuinely custom key — **plus add `--unset <key>`** so
  mistakes are recoverable.

Decision:

## D5 — Detaching a task from its epic
Same task as D4 (B1 residual). `task set <t> --epic ""` currently fails with
the confusing `unknown epic ""`.
- **A (recommended):** allow clearing — empty value (or `--unset epic` if D4
  adds it) removes the epic field.

Decision:

## D6 — Explicit `--set updated_at=...`
Same task as D4. The value validates, then is silently clobbered by the stamp.
- **A (recommended):** reject the key explicitly (like `status`) — it's a
  system-stamped field.

Decision:

## D7 — JSON schema versioning strategy
Task: [[json-and-output-contract-fidelity]]. One global `SchemaVersion`
("1.0") is stamped into ~10 different payload shapes.
- **A (recommended):** keep one global version, documented as "the CLI output
  schema"; bump the minor once when the round-trip fields land.

Decision:

## D8 — JSON key naming: `created` vs `updated_at`
Same task as D7.
- **A (recommended):** keep keys matching the frontmatter exactly
  (`created`, `updated_at`) and document that rule — no breakage.

Decision:

## D9 — Machine-readable errors under `--json`
Task: [[agent-facing-cli-ergonomics-batch]] (its headline item).
- **A (recommended):** on failure with `--json`, emit a JSON error envelope
  (`{"schema_version", "error": {"code", "message"}}`) to **stderr**; stdout
  stays empty on failure. Codes reuse the exit-code vocabulary
  (`not-found`, `validation`, …).

Decision:

## D10 — The retired Python prototype (`bin/pm`, `tests/test_pm.py`)
Task: [[repo-hygiene-batch]] (M19). Docs call `tests/test_pm.py` the kept
historical spec, but it isn't in git; `bin/pm` is.
- **A (recommended):** delete both relics and update README/CLAUDE.md — the Go
  suite is the spec now and the port epic is at 70%+ with parity shipped.

Decision:

## D11 — LICENSE
Task: [[repo-hygiene-batch]]. The module path is public-shaped
(`github.com/andy-esch/taskflow`); there's no license file.
- **C:** intentionally unlicensed for now (private project).

Decision:

## D12 — Close the two stale in-progress tasks?
Task: [[repo-hygiene-batch]]. `port-pm-to-go-cli…` and
`tui-sprint-3-fsnotify-live-reload` both look shipped.
- **A (recommended):** yes — verify their remaining checklist items and
  `task complete` both.
---

## Outcome (same day)

All twelve decisions were executed on 2026-06-12 — D1–D10/D12 as recommended,
D11 as chosen (no license for now). Implementation details live in the
closure notes of the respective task files (now under `tasks/completed/`);
the only decision-adjacent work still open is the remainder of
[[agent-facing-cli-ergonomics-batch]] (body-file/body-edit/create-envelope)
and confirming the repaired CI gate on the first push
([[repair-ci-lint-gate-and-local-test-parity]]).
