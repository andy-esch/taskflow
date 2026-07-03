---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: Add read-by-path to kill the O(N^2) audit rescan; split FixFrontmatter/WatchPaths off the use-case port; re-justify the doc boundary claims.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [architecture, performance]
created: "2026-06-22"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
id: 6fes83r01325
---
# Reshape core.Store port read-by-path and Fixer Layout split

## Objective

Three port/boundary items. (1) QueryFindings/LintAudits call GetAudit(slug) per
audit (re-resolve + re-read) — O(N^2) on the hottest read path; add a read-by-path
accessor and read a.Path directly (needs fakeStore rework). (2) FixFrontmatter and
WatchPaths are fs/text ops bloating the use-case core.Store port; split into narrow
Fixer/Layout interfaces cli wires to the FS. (3) Re-justify ARCHITECTURE.md: the
render->core diamond is ~5 types not "two", and the doc boundary claims are stale.

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **M16** (O(N^2), the reason this is a port change not a quick win),
**L12** (Fixer/Layout), **M8** (doc drift).

## Acceptance criteria

- [x] Findings queries read audit bodies in one scan / by path; no per-audit re-resolve.
- [x] core.Store carries only use-case methods; Fixer/Layout split out.
- [x] ARCHITECTURE.md boundary justifications match reality.
- [x] just test + just lint green.

## Implementation plan

**Approach.** Three port/boundary fixes, smallest-blast-radius first. (1) Kill the
O(N²) audit rescan with a **read-by-path** accessor on the audit port — preferred over
"have `ListAudits` return bodies" because list output (`audit list`) doesn't want
bodies, and a path-keyed read is a tiny, obviously-correct method the `fakeStore` can
satisfy. (2) Split `FixFrontmatter` (and `WatchPaths`) off the use-case `core.Store`
into narrow `Fixer`/`Layout` interfaces the CLI/TUI wire directly to the FS. (3)
Truth-up the doc. These are independent; land M16 first (clear perf + correctness win),
then L12, then M8.

**Steps — M16 (read-by-path).**
1. Add `GetAuditByPath(path string) (domain.Audit, string, error)` to the `AuditStore`
   interface in `internal/core/store.go`. Implement on `*FS` in
   `internal/store/auditstore.go`: `os.ReadFile(path)` → `parseAudit(content, path,
   bucketFromPath(path))`, deriving the bucket from the parent dir name (the
   `audits/<bucket>/` convention the store already owns) rather than re-resolving the
   slug. Reuse the existing `parseAudit`.
2. Rewrite the loops in `internal/core/finding.go` — `QueryFindings` (the
   `for _, a := range audits { s.store.GetAudit(a.Slug) }` at lines 61–68) and
   `LintAudits` (lines 98–105) — to call `GetAuditByPath(a.Path)` (`.Path` is already
   populated by `ListAudits`). That collapses 3(N+1) dir scans + 2N reads to one
   `ListAudits` scan + N direct reads, and closes the concurrent-edit re-resolve
   window. The single-audit branch (`f.Audit != ""`) still uses `GetAudit` (it must
   resolve a user-typed slug).
3. **fakeStore rework** (`internal/core/service_epic_test.go`): add a
   `GetAuditByPath` to `nopStore` (one line returning `ErrNotFound`) and to
   `fakeStore`. The fake currently keys `auditBodies` by slug for `GetAudit`; give it a
   path-keyed lookup too — simplest is to make `fakeStore.GetAuditByPath(path)` find
   the audit whose `.Path == path` and return `auditBodies[a.Slug]`, OR have the test
   audits carry `.Path = slug` so the two maps coincide. Update existing finding tests
   (`finding_test.go`) to populate `.Path` on their seed audits.

**Steps — L12 (Fixer/Layout split).**
4. Define in `internal/core/store.go` two narrow ports the *primary adapters* depend
   on, not `Service`: `type Fixer interface { FixFrontmatter(dryRun bool)
   ([]domain.FixResult, error) }` and `type Layout interface { WatchPaths() []string }`.
   Remove both methods from the `Store` interface (keep them on `*FS` — the compile
   assertion `var _ core.Store` shrinks; add `var _ core.Fixer`/`var _ core.Layout`).
5. `Service.LintFix` is a pure pass-through to `FixFrontmatter`; either keep
   `Service` holding a `Fixer` alongside its `Store` (constructor takes both, or `*FS`
   satisfies all three so `NewService(store)` still works), or move `LintFix` out of
   core entirely and have the cli's `runLintFix` call the `Fixer` directly. Prefer the
   latter (it's the L12 recommendation: "cli wires directly to the FS"), but it touches
   `internal/cli/lint.go` (`app.Svc.LintFix` → `app.Fixer.FixFrontmatter`) and the
   `App` struct (`root.go`) to carry a `Fixer`. `WatchPaths` similarly: the TUI reads
   it via `Service.WatchPaths` today — either keep that thin pass-through (it's
   harmless and the doc already justifies it) or expose `Layout` to the TUI. Lowest
   churn that still satisfies "Store carries only use-case methods": drop both from
   `Store`, keep the `Service.WatchPaths`/`LintFix` pass-throughs delegating to a
   `Fixer`/`Layout` the service also holds.
6. Update `nopStore`/`fakeStore` — they no longer need to implement
   `FixFrontmatter`/`WatchPaths` once those leave `Store` (delete those two methods
   from `nopStore`, removing two lines of fake burden — exactly the L12 win).

**Steps — M8 (doc).**
7. Edit `docs/ARCHITECTURE.md` §"Why these boundaries" render bullet: replace "imports
   `core` for two view-models" with the honest count — render imports ~5 core types
   (`Summary`, `StatusCount`, `EpicSummary`, `AuditFinding`, `LintResult`), growing
   per entity, and note stats/index/tags will deepen it; either re-justify (render is
   the isolation seam the TUI doesn't touch) or point at the `taskJSON`/`auditJSON`
   DTO-mapping pattern as the trend-reversal. Also update the §"Port purity leak (low)"
   theme to reflect the now-done Fixer/Layout split.

**Tests.** (a) M16: a `finding_test.go` case asserting `QueryFindings`/`LintAudits`
return identical results via the new path read (seed audits with `.Path` set);
optionally a store-level `TestFS_GetAuditByPath` round-trip against `t.TempDir()`. (b)
A store test that `GetAuditByPath` derives the right bucket from the parent dir. (c)
The `var _ core.Store/Fixer/Layout = (*FS)(nil)` assertions are compile-time tests.
Existing core/cli/golden suites must stay green (no output change).

**Risks / gotchas.** (a) **fakeStore rework is the crux** — the audit explicitly
deferred M16 for this reason; get the path↔slug mapping consistent across `GetAudit`
and `GetAuditByPath` or finding tests will diverge. (b) `bucketFromPath` must use the
parent-dir name, not re-parse frontmatter (the bucket==directory invariant); guard a
path outside `audits/<bucket>/`. (c) Splitting the port changes `NewService`'s
signature or `App`'s fields — touch every constructor call site (cmd/tskflwctl,
cli/root.go, tui.New via the service) and keep DI-via-one-App intact (no globals). (d)
No `schema_version` bump — these are internal, output is unchanged.

**Done when.** Findings/lint read audit bodies by path in one scan (no per-audit
re-resolve), `core.Store` carries only use-case methods with `Fixer`/`Layout` split
out and the fakes lighter by two methods, the doc's render→core and port-purity claims
match reality, and `go build ./...`, `go test ./...`, `golangci-lint run ./...` are
green.
