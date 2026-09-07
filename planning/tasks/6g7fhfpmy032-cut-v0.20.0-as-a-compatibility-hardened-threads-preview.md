---
schema: 1
id: 6g7fhfpmy032
status: in-progress
epic: 30-threads-and-task-dependency-graphs
description: Ship the compatibility and repair hardening as another explicit preview checkpoint before reconsidering graduation.
effort: 1 day
tier: 2
priority: medium
autonomy_level: 4
tags: [threads, release, compatibility, dogfood]
created: "2026-09-06"
depends_on: [6g7ddeyp773z, 6g7j2ebatzyt]
updated_at: "2026-09-07"
started_at: "2026-09-07"
---

# Cut v0.20.0 as a compatibility-hardened Threads preview

## Objective

Ship the historical-compatibility and guarded-repair foundation as a third deliberately preview-
labelled Thread release. Let real installed binaries exercise the hardened contracts before an
explicit graduation decision; passing the graduation gates makes graduation possible, not automatic.

## Scope

- Start from clean `main` after `pin-thread-document-and-plan-backward-compatibility` lands and
  record the exact candidate commit and binary version.
- Run the compatibility fixtures, focused hostile graph/repair/mutation tests, full race tests,
  lint, generated docs/schema checks, planning lint, and `just release-snapshot`.
- Repeat a bounded CLI/TUI dogfood pass with the production Thread, stable task navigation, live
  refresh, a retained apply plan, and graph diagnosis/repair. Record any behavioral or compatibility
  finding as a named task rather than silently accepting it for graduation.
- Curate concise release notes around compatibility-managed preview data, resumable apply, guarded
  recovery, shared core projections, and the current CLI/TUI capabilities.
- Keep the README preview notice in the tagged candidate and verify the immutable tag, release
  workflow, archives, checksums, and installed binary after publication.

## Release execution playbook

This task is the operational checklist and evidence log for v0.20.0. The authoritative reasons and
named G1–G6 tests remain in the
[Threads compatibility contract](../../docs/THREADS_COMPATIBILITY.md); the v0.18.0 task's
[`Clean-main dogfood`](6g5m69wpydzw-cut-v0.18.0-as-a-cli-threads-preview.md#clean-main-dogfood-2026-08-31)
is historical precedent, not a substitute for recording this run.

After release-planning edits have landed, freeze one clean candidate and record every output here:

```bash
git status --short
git rev-parse HEAD
just build
bin/tskflwctl version

go test ./internal/core ./internal/store ./internal/cli ./internal/tui ./internal/wire
go test -race ./...
just lint
just tidy-check
just docs-check
bin/tskflwctl lint
just release-check
just release-snapshot
git diff --check
git status --short
```

Run the manual CLI/TUI pass against copies, never the committed fixtures or installed binary:

```bash
V020_LAB=$(mktemp -d /tmp/taskflow-v020-dogfood.XXXXXX)
V020_CLI="$PWD/bin/tskflwctl"
cp -R assets/demo-threads "$V020_LAB/space"

"$V020_CLI" -C "$V020_LAB/space" lint
"$V020_CLI" -C "$V020_LAB/space" thread show touring-bike-departure
"$V020_CLI" -C "$V020_LAB/space" thread frontier touring-bike-departure
"$V020_CLI" -C "$V020_LAB/space" thread graph touring-bike-departure
"$V020_CLI" -C "$V020_LAB/space" thread plan touring-bike-departure
"$V020_CLI" -C "$V020_LAB/space" thread compose \
  --from "$PWD/assets/demo-threads/shared-safety-review.compose.yml" \
  --out "$V020_LAB/preview.apply.yml"
"$V020_CLI" -C "$V020_LAB/space" thread apply "$V020_LAB/preview.apply.yml" --dry-run
"$V020_CLI" -C "$V020_LAB/space" thread apply "$V020_LAB/preview.apply.yml"
"$V020_CLI" -C "$V020_LAB/space" thread apply "$V020_LAB/preview.apply.yml"
```

The retained apply plan creates a second Thread sharing three members and adds one dependency while
skipping two that already exist. Exercise guarded membership add/remove and Thread start/complete
refusal against those two Threads. In one terminal open `ui`, navigate Threads → topology → task →
`ctrl+o`; in another, start and complete the eligible wheel task and the external crown-mount task.
Confirm live reload, frontier, external-gate, and affected-Thread receipts.

Finally, raw-edit only the copied `space` to add a self-edge or duplicate dependency. Confirm
ordinary `lint` and frontier fail closed; run `task depend repair`, preview the exact `--auto` or
explicit selector, apply it, and prove lint plus Thread projections recover. Append the lab path,
commands, outcomes, and every filed follow-up below before checking the dogfood criterion.

After the candidate and notes are merged, tag that exact commit. Verify the release workflow,
downloaded `checksums.txt` and four archives, an extracted binary's `version`, and a clean install;
record the immutable tag/commit and links here before completing the publication criterion.

## Acceptance criteria

- [x] `pin-thread-document-and-plan-backward-compatibility` is complete and the v0.18.0/v0.19.0
      compatibility fixtures pass from the recorded clean candidate.
- [x] Focused hostile tests, full `go test -race ./...`, lint, generated docs/schema checks,
      planning lint, and `just release-snapshot` pass on that candidate.
- [x] A fresh candidate binary completes the bounded CLI/TUI dogfood pass; commands and outcomes,
      including any filed follow-ups, are recorded here.
- [x] README and release notes explicitly retain the Threads preview classification and explain
      the compatibility/recovery value of this release without implying graduation.
- [ ] The published v0.20.0 tag, release workflow, archives, checksums, and installed binary all
      identify the recorded candidate.
- [x] The graduation task remains a separate explicit decision after this checkpoint; the spatial
      graph experiment is neither silently promoted nor made a release gate.

## Out of scope

- Removing the Threads preview notice or declaring all graduation evidence current.
- Requiring the two-dimensional graph experiment, frontier ranking, or portable convenience-view
  diagnostics merely to cut this checkpoint.
- Folding an experimental spatial renderer into core Thread semantics.

## Preview visual evidence

The approved [touring-bike Thread topology recording](../../assets/threads.gif) demonstrates the
current terminal-native graph comprehension path: three explanatory waves, fan-out/fan-in,
eligible and in-flight work, an immediate external gate, and stable task navigation. It remains a
secondary preview demo; the optional two-dimensional spatial renderer is not a v0.20 release gate.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Contract [Threads compatibility and preview graduation](../../docs/THREADS_COMPATIBILITY.md)
- Compatibility fixtures
  [`6g7ddeyp773z`](6g7ddeyp773z-pin-thread-document-and-plan-backward-compatibility.md)
- Graduation decision [`6g7ddfhh2jc2`](6g7ddfhh2jc2-graduate-threads-from-preview.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
