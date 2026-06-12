---
status: ready-to-start
epic: 17-pm-go-cli
description: 'DRAFT: goreleaser binaries on tag push + opt-in workflow_dispatch (no main-merge builds); darwin+linux, GitHub-only distribution; 3 questions open'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [ci, distribution, goreleaser, draft]
created: "2026-06-12"
updated_at: "2026-06-12"
---
# Binary releases via goreleaser

> 🚧 **DRAFT — not yet integrated into the overall plan.** Filed 2026-06-12
> from a feature request. The mechanics are well-trodden (goreleaser); the
> open questions are the trigger policy and the private-repo/licensing
> interaction with decision D11. Resolve those before starting.

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

## Open questions

- [x] **Trigger** — DECIDED 2026-06-12: no build on merge to main; release
      builds on tags/releases, plus an opt-in `workflow_dispatch`. (Micro-
      question folded in below: what dispatch produces.)
- [ ] **Dispatch output:** recommended = `--snapshot` → workflow artifacts
      only (never a Release); alternative = dispatch creates a prerelease.
      Awaiting sign-off on the recommendation.
- [ ] **Versioning scheme:** start at `v0.1.0`? Who/what bumps — manual
      tags, or release-please-style automation later?
- [ ] **D11 interaction** — NARROWED 2026-06-12: distribution is GitHub-
      repo-based only (no external channels), so private releases via gh
      auth raise no licensing issue. The question only revives if the repo
      ever goes public — leave open as that reminder.
- [x] **Platform matrix** — DECIDED 2026-06-12: darwin/linux × amd64/arm64;
      **Windows is out for now.**
- [ ] **Checksums only, or signing/attestation too** (cosign / GitHub
      artifact attestation)? Probably checksums-only at this scale.
- [x] **Homebrew tap** — DECIDED 2026-06-12: **no.** Everything stays
      GitHub-repo-based (`go install`, `gh release download`, source).

## ⚠️ Conflicts to resolve before starting

- **D11 (no license)** — largely defused by the GitHub-only decision (no
  Homebrew, no external channels); remains only as a tripwire if the repo
  ever goes public.
- `release.yml` must build on the **repaired CI gate**
  ([[repair-ci-lint-gate-and-local-test-parity]], in-progress — awaiting
  first green run); don't fork a second workflow until that's confirmed.
- The Justfile `version`/ldflags logic and goreleaser must agree on one
  stamping scheme (don't drift the local build's `git describe` from the
  released `{{.Version}}`).

## Acceptance criteria (draft)

- [ ] Open questions above resolved; task de-drafted.
- [ ] Then: a tagged (or decided-trigger) release produces darwin+linux
      binaries fetchable via `gh release download`, version-stamped, with
      checksums; README documents all install paths; `just
      release-snapshot` dry-runs locally.

## Related

- Epic [[17-pm-go-cli]] · [[2026-06-12-pending-decisions]] (D11) ·
  [[repair-ci-lint-gate-and-local-test-parity]] ·
  [[2026-06-11-critical-review-and-polish-research]] (Tier 3: distribution
  funnel).