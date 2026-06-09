---
status: completed
epic: 17-pm-go-cli
description: Acted on the easy+meaningful items (git conflict detection, slug dot-trim, fuzz tests); spun off --dry-run; skipped MCP/slog as off-scope
effort: Unknown
tier: 3
priority: low
autonomy_level: 3
tags: [pm-tooling, go, robustness]
created: "2026-06-09"
completed_at: "2026-06-09"
updated_at: "2026-06-09"
---

# Triage agent-readiness review: conflict markers, slug safety, parser fuzzing

## Objective

Triage a feature-request review (agent-readiness / sync-resilience / parser
robustness) by value × effort: do the easy+meaningful now, spin off the bigger
ones, skip the off-scope.

## Done now

- **Git conflict-marker detection** (`store/diagnose.go`): a synced repo
  (git/Dropbox/Syncthing) can leave `<<<<<<< / >>>>>>>` markers in a file.
  Instead of a confusing YAML error, `frontmatterError` now reports "git merge
  conflict markers detected — resolve the conflict." Demoed via `task list`.
- **Slug cross-platform safety** (`domain/slug.go`): the Windows-reserved
  characters `:\*?<>|"/` were *already* stripped (verified); the gap was
  **trailing/leading dots** (Windows strips trailing dots) — now trimmed. So
  `Done.` → `done`, `...x...` → `x`.
- **Fuzz tests for the hand-rolled byte parsers** (`store/fuzz_test.go`):
  `splitFrontmatter`, `updateFrontmatter`, `fixFrontmatterText`,
  `frontmatterError`. Ran ~12s each (millions of execs) — **no panics/crashes**;
  the seed corpus now runs under normal `go test`.

## Spun off (meaningful, more work)

- **Global `--dry-run` for mutating commands** →
  [[global-dry-run-for-mutating-commands]] (ready-to-start).

## Skipped (off-scope / low value — with rationale)

- **MCP server mode (`tskflwctl mcp start`)**: explicitly out by a long shot per
  the epic-17 scope fence (MCP/RAG/brain). Not a task — it contradicts the fence.
- **`log/slog` structured logs (`--log=json`)**: low value here — there's no
  real execution-log stream; the structured data agents need (problems/results)
  is already in the `--json` payloads (incl. the `unreadable` array). Skip.
- **Default output sorting**: already deterministic — `ListTasks` iterates
  `AllStatuses()` (fixed order) then `os.ReadDir` (sorted by filename). The
  agent-diffing concern is met. A `--sort` flag would be a minor nicety; not
  worth a task now.

## Acceptance

- [x] Conflict-marker message, slug dot-trim, and 4 fuzz targets land with
      tests; full suite + lint green; fuzzers find no crashes.
- [x] `--dry-run` captured as its own task; off-scope items recorded with reasons.

## Related

- Epic [[17-pm-go-cli]]; follows
  [[triage-go-idiom-review-typed-validators-status.isactive-date-format]].
