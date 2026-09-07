---
schema: 1
id: 6g7r20ffjf2w
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Unify automated release checks behind one command and run the same contract in a pinned, headless Linux container.
effort: 1-2 days
tier: 2
priority: medium
autonomy_level: 4
tags: [release, automation, containers, developer-experience]
created: "2026-09-07"
---
# Make release validation a reproducible one-command gate

## Objective

Turn the automated portion of release qualification into one reusable command, then make that same contract runnable in a pinned headless Linux container. A release candidate should not depend on one developer's macOS setup, shell state, or globally installed tool versions.

## Scope

- Add `just release-validate` as the canonical automated gate, backed by one shared implementation rather than a second hand-maintained command list.
- Preserve the current release checks: focused and full race tests, lint, module tidiness, generated docs/schema checks, planning lint, GoReleaser validation, and a snapshot build.
- Fail early with actionable diagnostics for a dirty candidate, missing tools, or unsupported tool versions; leave the tracked worktree unchanged.
- Add a headless OCI-container entry point that runs the same gate with pinned toolchain versions, isolated home/config state, non-root execution, and reusable dependency/build caches outside the repository.
- Document the host and container invocations and the boundary between automated qualification, manual CLI/TUI dogfood, and tag/publish verification.
- Evaluate a devcontainer only as an optional thin wrapper around the same image/configuration. It must not become a second validation contract or the only portable entry point.

## Acceptance criteria

- [ ] One documented `just release-validate` command executes the repository's complete automated release gate and returns non-zero at the failing phase.
- [ ] The gate checks its candidate and prerequisites explicitly, prints concise phase-oriented progress, and leaves tracked files unchanged on both success and failure.
- [ ] One documented container command runs the same validation implementation in a pinned Linux environment without reading the developer's home-directory configuration or requiring host-installed Go, GoReleaser, or golangci-lint.
- [ ] Container execution works from a clean clone, uses non-root ownership safely, and keeps module/build caches and generated release artifacts from polluting tracked source.
- [ ] Host and container paths are covered by focused smoke tests or CI evidence, including at least a stale generated artifact, a failing test/check, a dirty candidate, and a missing prerequisite.
- [ ] The repository documents which environment is authoritative for release parity, any cross-platform limitations, and whether a reusable devcontainer adds value without duplicating the runner.
- [ ] The v0.20 release playbook remains historical evidence; future release tasks can link to the reusable gate instead of copying its automated command list.

## Design guidance

Prefer a purpose-built Containerfile and headless command as the portable baseline because they work locally and in CI without an editor. If a devcontainer is retained, have it consume the same image or setup rather than maintaining separate versions and steps. The GitHub Actions tag workflow remains the publication authority; the container is a repeatable preflight, not a claim of byte-for-byte reproducible cross-platform artifacts.

## Stress cases

- Apple Silicon and amd64 hosts invoking the Linux validation runner.
- Cold and warm dependency caches.
- A checkout with a dirty tracked file or stale generated documentation.
- Missing container engine, GoReleaser, or linter on the host path.
- A validation failure after snapshot output has been created.
- File ownership after a container run.

## Out of scope

- Replacing the existing GitHub Actions or GoReleaser publication workflow.
- Publishing taskflow as a container image.
- Bit-for-bit reproducible builds, signing, provenance, or SBOM policy.
- Automating manual Thread CLI/TUI dogfood, terminal GIF recording, tagging, or GitHub release publication.
- Making a specific editor or devcontainer implementation mandatory for contributors.

## Related

- Release checkpoint [cut-v0.20.0-as-a-compatibility-hardened-threads-preview](6g7fhfpmy032-cut-v0.20.0-as-a-compatibility-hardened-threads-preview.md)
- Distribution foundation [binary-releases-via-goreleaser](6fbj87000zs7-binary-releases-via-goreleaser.md)
- Epic [21-code-quality-architecture-hardening](../epics/21-code-quality-architecture-hardening.md)
