---
schema: 1
status: completed
epic: 23-point-an-impl-repo-at-an-external-planning-repo
description: Seed tracked_repos via --track; auto-append this repo to the planning repo's tracked_repos on init (--no-link-back to opt out); path-normalized.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [cli, config, init]
created: "2026-06-22"
updated_at: "2026-06-23"
started_at: "2026-06-23"
completed_at: "2026-06-23"
---
# `tracked_repos` seeding + auto-link-back

Wire the reverse direction: the planning repo records the impl repos that point
at it, and `init` keeps both sides in sync.

## Scope

1. Seed `tracked_repos` on the planning side: a repeatable `--track <path>` flag
   on `init`, and the interactive flow offers to add tracked repos.
2. Auto-link-back: when `init --planning-repo X` runs in an impl repo, also
   append this repo to **X's** `tracked_repos` — a surgical key append to a
   *second* repo's config. Opt out with `--no-link-back`. Skip silently (no
   error) if X's config doesn't exist yet.
3. **Physical-path normalization** (the correctness core): store/compare entries
   so `../planning`, an absolute path, and a symlinked checkout that resolve to
   the same place are treated as equal — never duplicated, never seen as a
   mismatch. Reuse the `evalOr` (`Abs` + `EvalSymlinks`) discipline already in
   `config.go`.

## Acceptance criteria

- [ ] `init --planning-repo X` appends this repo to X's `tracked_repos`;
      `--no-link-back` suppresses it; missing X config is a no-op, not an error.
- [ ] `init --track ../impl` seeds the planning side.
- [ ] Re-running never duplicates an entry that resolves to an existing one.
- [ ] Frontmatter/config edits are surgical (preserve key order, comments).
- [ ] Suite + lint green.

## Depends on

- `init` pointer mode (shares the write path).

## Related

- [[23-point-an-impl-repo-at-an-external-planning-repo]].

## Review hardening (2026-06-23)

Two adversarial reviewers. Link-back/CLI semantics verified clean (Rel across rel/abs/symlink/non-sibling, idempotency/repair, best-effort warn-not-fail keeps --json stdout pure, flag matrix, headless). Config-edit engine had a BLOCKER + 2 MAJORs, all from the array regex — fixed in one change:
- **BLOCKER**: a ']' in a tracked path truncated the '[^\\]]*' match, so the NEXT edit spliced invalid TOML and bricked discovery. Replaced the regex with a line-anchored, quote/bracket-aware single-assignment locator (trackedReposSpan/skipTOMLString).
- **MAJOR**: global ReplaceAll clobbered a bracketed COMMENT / a second assignment. Now edits exactly the first real (line-start) assignment.
Nits: reject --no-link-back in scaffold mode (symmetry with --track); print the link-back warning after the success line.
Deferred (MINOR): --json 'linked_back' field → routed to the JSON-envelopes task. Tests added (]-path round-trip, comment-not-clobbered, bracket span, scaffold guard).
