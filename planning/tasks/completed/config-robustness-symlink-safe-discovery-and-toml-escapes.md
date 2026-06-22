---
schema: 1
status: completed
epic: 21-code-quality-architecture-hardening
description: Discovery/containment are purely lexical (symlink-bypassable); the hand-rolled TOML reader mis-decodes basic-string escapes.
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [config, security]
created: "2026-06-22"
updated_at: "2026-06-22"
completed_at: "2026-06-22"
---
# Config robustness symlink-safe discovery and TOML escapes

## Objective

Config path handling is lexical. (1) configuredRoot's escape guard uses filepath.Rel
(no symlink resolution), so a taskflow_root pointing at a symlink can escape the repo
while passing the no-".." check — EvalSymlinks both dir and root before the check.
(2) Discover climbs logical ancestry; EvalSymlinks(start) once at the top. NOTE:
EvalSymlinks resolves temp-dir symlinks (macOS /var -> /private), so update path
comparisons in config tests accordingly. (3) tomlStringValue mis-reads basic-string
escapes; decode properly or reject backslash-bearing basic strings (don't mis-decode).

## Audit reference

planning/audits/open/2026-06-22-code-quality-architecture.md — **L8** (symlink escape, security), **L19** (symlink discovery), **L17** (TOML escapes).
Deferred from the quick-win tranche because of the temp-dir-symlink test subtlety.

## Acceptance criteria

- [ ] taskflow_root pointing at a symlink that escapes the repo is rejected (test).
- [ ] Discovery resolves symlinks; tests pass on macOS-style symlinked temp dirs.
- [ ] TOML basic strings with escapes are decoded or rejected, never mis-decoded.
- [ ] just test + just lint green.

## Implementation plan

**Approach.** Three localized hardenings in `internal/config/config.go`, all lexical→
physical path fixes plus one parser fix. (1) `configuredRoot`'s containment check uses
`filepath.Rel` (lexical) — resolve both `dir` and the candidate `root` with
`filepath.EvalSymlinks` before the `Rel`/no-`..` check so a `planning -> /etc` symlink
can't slip past containment. (2) `Discover` climbs logical ancestry — `EvalSymlinks`
the start dir once at the top before the walk-up. (3) `tomlStringValue` scans to the
first matching quote with no escape handling — for the narrow blast radius (the value
is a path), **reject** a basic (double-quoted) string containing a backslash rather
than mis-decode it (the recommendation's conservative option), and document that only
escape-free basic strings + literal single-quoted strings are supported. A full TOML
decoder is overkill for one key.

**Steps.**
1. **L8 (containment).** In `configuredRoot` (config.go:62–75), after computing
   `root := filepath.Join(dir, …rel)`, evaluate both with `EvalSymlinks` (falling back
   to the lexical path if the target doesn't exist yet — but the function already
   requires `root/tasks/` to exist, so the eval should succeed on the real case). Run
   the `filepath.Rel(evalDir, evalRoot)` no-`..` check on the *evaluated* paths; reject
   if the evaluated root escapes the evaluated dir. Keep the existing `tasks/`-exists
   check.
2. **L19 (discovery).** In `Discover` (config.go:25–55), `dir, err :=
   filepath.EvalSymlinks(absStart)` once after the `filepath.Abs(start)` (config.go:26),
   before the `for` climb — so the `.git`-boundary walk uses physical ancestry. Guard
   the case where `start` doesn't exist (EvalSymlinks errors): fall back to the abs path
   so a fresh dir still discovers/errs sensibly.
3. **L17 (TOML).** In `tomlStringValue` (config.go:105–119), in the double-quote branch,
   if the extracted-or-raw value contains a backslash, return "" (treat as unset,
   matching the unterminated-quote policy) — i.e. refuse to guess rather than pass
   `\"`/`\t` through literally. Update the doc comment on `taskflowRoot`/
   `tomlStringValue` to state the supported subset (escape-free basic strings, literal
   single-quoted strings). Single-quoted (literal) strings are already correct and
   unaffected.

**Tests (config package).** Extend `discover_test.go`:
- **L8:** create `dir/repo` with a `planning` symlink pointing outside the repo (to
  another temp dir), a `.tskflwctl.toml` with `taskflow_root = "planning"`, assert
  `Discover` rejects it (error mentions "escapes"). Mirror the existing
  `TestDiscover_RejectsBadConfiguredRoots` shape.
- **L19:** a symlinked-worktree discovery test: real planning tree at one path, a
  symlink dir pointing at it, `Discover(symlinkPath)` resolves to the real root.
- **L17:** add rows to `TestTaskflowRoot_TOMLValueForms` — `taskflow_root = "a\"b"`
  and `taskflow_root = "a\\b"` now return "." (rejected), while `'a\b'` (literal) and
  escape-free `"./planning"` still work.

**Risks / gotchas.** (a) **macOS-style symlinked temp dirs** — this is the reason the
task was deferred from the quick-win tranche: `t.TempDir()` on macOS lives under
`/var/folders/...` where `/var -> /private/var`, so `EvalSymlinks` changes the path.
Every existing config test that compares `cfg.Root` to a `t.TempDir()`-derived path
must compare against `EvalSymlinks(want)` (or wrap the expectation), or it will fail on
macOS even though discovery is correct. Audit `discover_test.go`/`init_test.go` for
direct path equality and update those comparisons. (b) Don't break the `.git`-file
boundary test (`TestDiscover_GitFileBoundary`) — EvalSymlinks on the start dir doesn't
change the boundary semantics, but re-run it. (c) The read path already rejects symlink
*dir-entries* (`markdownDoc` requires a regular file) — this task is specifically the
*root/config* path, a separate surface; don't conflate. (d) `EvalSymlinks` on a
not-yet-existing path errors — handle that fallback so `init` in a fresh dir still
works. (e) No `schema_version` / output change.

**Done when.** A `taskflow_root` symlink that escapes the repo is rejected, discovery
resolves symlinks (tests green on macOS-style symlinked temp dirs via EvalSymlinks-aware
expectations), backslash-bearing basic TOML strings are rejected not mis-decoded, and
`go build ./...`, `go test ./...`, `golangci-lint run ./...` are green.
