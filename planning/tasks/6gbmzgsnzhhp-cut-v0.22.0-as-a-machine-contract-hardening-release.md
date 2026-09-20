---
schema: 1
id: 6gbmzgsnzhhp
status: completed
epic: 20-cli-ux-and-ergonomics
description: Qualify and publish the stable-ID projections, registry guards, and explicit machine-contract revision policy as one coherent release.
effort: 1 day
tier: 1
priority: high
autonomy_level: 2
tags: [cli, json, contract, release, dogfood]
created: "2026-09-19"
updated_at: "2026-09-19"
depends_on: [6ga968z46e9y]
started_at: "2026-09-19"
completed_at: "2026-09-19"
---

# Cut v0.22.0 as a machine-contract hardening release

## Objective

Qualify and publish v0.22.0 as a coherent checkpoint for machine-facing CLI consumers. Center the
release on stable task and audit identities, truthful projected JSON, registry drift prevention,
and the explicit monotonic machine-contract revision policy. Keep Threads labelled as preview and
avoid pulling the next command-discovery phase into this bounded release.

## Scope

- Freeze one clean `main` candidate containing the completed
  [Harden CLI machine contracts](../threads/6g9txm0s3yyh-harden-cli-machine-contracts.md) work and
  record its commit and candidate version.
- Run both the host and pinned-container forms of the reusable release gate on that exact commit.
- Dogfood schema discovery plus task and audit projected JSON from a freshly built candidate,
  including stable IDs, selector order, raw registry values, and revision-policy metadata.
- Curate concise release notes that link the implementing tasks and audits, identify revision 1.68
  as additive, and retain the Threads preview boundary.
- Tag only the qualified candidate, then verify the GitHub workflow, four archives, checksums,
  embedded version, and an installed release binary before closing the task and Thread.

## Release execution playbook

After this release-planning change lands, freeze a clean `main` candidate and run both forms of the
same automated gate. The recorded commits must match.

```bash
git status --short
git rev-parse HEAD
just release-validate
just release-validate-container
```

Build from that candidate and exercise the contract surfaces directly:

```bash
just build
bin/tskflwctl version
bin/tskflwctl schema --json
bin/tskflwctl schema --json-schema
bin/tskflwctl task list --json -c id,slug,status
bin/tskflwctl audit list --all --json -c id,slug,bucket,open_findings
bin/tskflwctl lint
```

Confirm the schema response publishes revision 1.68, its monotonic/additive policy, and the
revision-qualified JSON Schema identity. Confirm both projected list surfaces return stable IDs in
the caller-selected order with their raw registry values rather than table display fallbacks.
Record any non-trivial finding as a named task rather than silently widening this release.

After release notes land, tag the exact qualified commit. Verify the tag and release workflow,
download `checksums.txt` and all four platform archives, validate every checksum, inspect an
extracted binary's version, and install the checksum-verified binary. Record immutable evidence
below before completing the publication criterion.

## Acceptance criteria

- [x] The four machine-contract Thread tasks are complete and merged into the clean candidate;
      projected JSON and schema revision 1.68 remain classified as additive.
- [x] `just release-validate` and `just release-validate-container` pass against the same recorded
      candidate without tracked worktree changes.
- [x] A fresh candidate binary passes the bounded schema/task/audit JSON dogfood; stable IDs,
      ordered selectors, raw registry values, and revision-policy metadata are verified.
- [x] Release notes concisely explain the machine-contract hardening, link planning and audit
      evidence, and retain the Threads preview classification.
- [x] The v0.22.0 tag, release workflow, four archives, checksums, extracted version, and installed
      binary all identify the recorded candidate.
- [x] Remaining exit-code, command-discovery, Thread-list projection, and bounded-query findings
      are tracked as a separate follow-on phase rather than made accidental release gates.

## Out of scope

- Graduating Threads from preview or changing the persisted Thread compatibility contract.
- Publishing the complete exit-code taxonomy or executable command manifest.
- Adding Thread-list projections, pagination, or changing compact mutation-receipt defaults.
- Claiming the container snapshot is a bit-for-bit reproducible release build.

## Related

- Thread [Harden CLI machine contracts](../threads/6g9txm0s3yyh-harden-cli-machine-contracts.md)
- ADR [Use monotonic revisions for the JSON machine contract](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
- Architecture audit [Machine contract](../audits/6g9zev1epg76-2026-09-14-arch-machine-contract.md)
- Release guide [Releasing taskflow](../../docs/RELEASING.md)
- Previous checkpoint [v0.21.0 spatial Threads preview](6g9mz2shwmb0-cut-v0.21.0-as-a-spatial-threads-preview.md)

## Release closeout (2026-09-19)

The immutable `v0.22.0` tag and candidate resolve to
`934e1cfc75b2f798bad6344498ca4e4b535bc8d6`. The maintainer reported both
`just release-validate` and `just release-validate-container` passing on the merged candidate, plus
the bounded schema/task/audit dogfood. Revision 1.68 remained additive, and the four prerequisite
machine-contract tasks were complete in the healthy Thread projection.

The [tag-triggered release workflow](https://github.com/andy-esch/taskflow/actions/runs/35455671413)
completed successfully. The [GitHub Release](https://github.com/andy-esch/taskflow/releases/tag/v0.22.0)
publishes `checksums.txt` and Darwin/Linux archives for amd64 and arm64. Fresh downloads of all four
archives passed the published SHA-256 checks; the extracted Darwin arm64 binary reports
`tskflwctl 0.22.0`. The locally installed binary, built from the tag, also reports `0.22.0`; it is
recorded as a source build rather than claimed byte-identical to the published archive.

The generated commit dump was replaced with concise planning-linked notes covering stable-ID
projections, structured JSON errors, registry invariants, revision 1.68, additive compatibility,
and the continuing Threads preview boundary.

Remaining findings are explicitly tracked in the unstarted
[Make CLI contracts self-describing and bounded](../threads/6gbn4v2jpf2m-make-cli-contracts-self-describing-and-bounded.md)
Thread: complete exit taxonomy (`6gbn4g1bb7wr`), command discovery (`6g63hhk3eddf`), Thread-list
projection (`6gbn4g1eypzs`), and bounded list queries (`6gbn4g1j40pj`).
