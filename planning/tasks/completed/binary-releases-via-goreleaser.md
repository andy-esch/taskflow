---
status: completed
epic: 17-pm-go-cli
description: goreleaser darwin+linux x amd64/arm64 binaries on tag push + opt-in snapshot dispatch; GitHub-Releases-only; v0.1.0 manual tags, checksums-only
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [ci, distribution, goreleaser]
created: "2026-06-12"
updated_at: "2026-06-14"
started_at: "2026-06-14"
completed_at: "2026-06-14"
---
# Binary releases via goreleaser

> ✅ **De-drafted 2026-06-14.** All open questions resolved (see below); a
> first implementation has landed in the working tree (`.goreleaser.yml`,
> `.github/workflows/release.yml`, Justfile recipes, README install section).
> Remaining: validate config, push the first `v0.1.0` tag, confirm a green
> release run.

## Objective

Let a fresh machine get a working `tskflwctl` in one command, without
cloning or building. Two channels, both should work when this lands:

1. **`go install github.com/andy-esch/taskflow/cmd/tskflwctl@latest`** —
   already works (Go builds from source on the consumer's machine; no
   binaries involved). For the private repo this needs
   `GOPRIVATE=github.com/andy-esch/*` + git auth — document it in the
   README install section.
2. **Prebuilt binaries on GitHub Releases** — fetched with
   `gh release download -R andy-esch/taskflow -p "*darwin_arm64*"` (gh
   handles private-repo auth) or curl for public. This is the part to build.

## Implementation sketch

- **goreleaser** with a matrix of `darwin/linux × amd64/arm64`,
  `CGO_ENABLED=0` (the stack is pure Go — static single binaries, no
  cross-compile toolchains needed). Archives + `checksums.txt`.
- Version stamping: reuse the existing ldflags var
  (`internal/cli.version`) — goreleaser's `{{.Version}}` replaces the
  Justfile's `git describe` at release time; `tskflwctl version` and the
  smoke test already pin it.
- A `release.yml` workflow **separate from ci.yml**, triggered by
  `push: tags: ['v*']` AND `workflow_dispatch` (decided 2026-06-12 — no
  build on merge to main). Recommended dispatch behavior, pending sign-off:
  goreleaser `--snapshot` with binaries uploaded as **workflow artifacts**
  only — an on-demand build of any commit that never mints a Release or
  burns a version; real Releases stay tag-only. Plus `just release-snapshot`
  (goreleaser `--snapshot --clean`) for local dry-runs.
- README install section: the three paths (go install w/ GOPRIVATE note,
  gh release download, build from source).

## Open questions — all resolved

- [x] **Trigger** — DECIDED 2026-06-12: no build on merge to main; release
      builds on tags/releases, plus an opt-in `workflow_dispatch`.
- [x] **Dispatch output** — DECIDED 2026-06-14: `--snapshot` → binaries
      uploaded as **workflow artifacts only**, never a Release, never burns a
      version. (Implemented as a conditional `args` in `release.yml`.)
- [x] **Versioning scheme** — DECIDED 2026-06-14: start at **`v0.1.0`**,
      **manual git tags** for now. Release-please-style automation can come
      later if tagging by hand gets old.
- [x] **D11 interaction** — NARROWED 2026-06-12: distribution is GitHub-
      repo-based only (no external channels), so private releases via gh
      auth raise no licensing issue. Revives only if the repo goes public.
- [x] **Platform matrix** — DECIDED 2026-06-12: darwin/linux × amd64/arm64;
      **Windows is out for now.**
- [x] **Checksums vs signing/attestation** — DECIDED 2026-06-14:
      **checksums-only** (`checksums.txt`, sha256). Single-consumer tool to
      start; revisit signing if/when the audience grows.
- [x] **Homebrew tap** — DECIDED 2026-06-12: **no.** Everything stays
      GitHub-repo-based (`go install`, `gh release download`, source).

## Conflicts — resolved

- **D11 (no license)** — defused by the GitHub-only decision; remains only as
  a tripwire if the repo ever goes public.
- **Repaired CI gate** — [[repair-ci-lint-gate-and-local-test-parity]] is
  **completed**, so the dependency is cleared; `release.yml` is a second,
  independent workflow (separate from `ci.yml`).
- **One version-stamping scheme** — both the Justfile build (`git describe` →
  `internal/cli.version`) and goreleaser (`{{ .Version }}` → the same var via
  ldflags) stamp the identical variable. No drift.

## Acceptance criteria

- [x] Open questions resolved; task de-drafted.
- [x] `.goreleaser.yml` builds darwin/linux × amd64/arm64, `CGO_ENABLED=0`,
      archives + `checksums.txt`, version-stamped via ldflags.
- [x] `.github/workflows/release.yml`: tag `v*` → real Release; manual
      dispatch → snapshot artifacts only.
- [x] `just release-snapshot` / `just release-check` recipes; `dist/`
      gitignored.
- [x] README documents all three install paths.
- [x] **`goreleaser check` passes** on the config (validated 2026-06-14 with
      goreleaser v2.16.0; full `--snapshot` build produced all 4 archives +
      checksums, version-stamped binary verified).
- [x] **First `v0.1.0` tag pushed** and the release run is green — GitHub
      Release v0.1.0 published with `tskflwctl_0.1.0_{darwin,linux}_{amd64,arm64}.tar.gz`
      + `checksums.txt`.

## Downstream beneficiary

Once Releases exist, the **desirelines-planning Claude routines** (which run in
a sandbox without `tskflwctl` — they currently fall back to the committed
Python `./bin/pm`) can install the CLI with a `gh release download` step in
their bootloader, unblocking full `pm` deprecation there. Not urgent (routines
lean on the CLI lightly), but it's the logical follow-on.

## Related

- Epic [[17-pm-go-cli]] · [[2026-06-12-pending-decisions]] (D11) ·
  [[repair-ci-lint-gate-and-local-test-parity]] ·
  [[2026-06-11-critical-review-and-polish-research]] (Tier 3: distribution
  funnel).

## Progress Log

- **2026-06-14**: De-drafted — resolved the 3 remaining open questions
  (dispatch = snapshot-artifacts-only; start `v0.1.0` w/ manual tags;
  checksums-only). Implemented `.goreleaser.yml` (v2; darwin/linux ×
  amd64/arm64, `CGO_ENABLED=0`, `-trimpath`, `-s -w` + version ldflags,
  tar.gz archives, sha256 `checksums.txt`, snapshot template, GitHub-only
  release), `.github/workflows/release.yml` (tag `v*` → real release; manual
  dispatch → `--snapshot` → workflow artifacts), `just release-snapshot` /
  `release-check`, `dist/` gitignored, and the README **Install** section
  (gh release download / go install w/ GOPRIVATE / source). Validated
  locally: `goreleaser check` + full `--snapshot` build (4 archives,
  checksums, version-stamped binary).
- **2026-06-14**: Merged to main, tagged `v0.1.0`, release run green —
  **GitHub Release v0.1.0 published** with all 4 platform archives +
  `checksums.txt`. Distribution is live; `gh release download` /
  `go install @v0.1.0` both work. Task complete.