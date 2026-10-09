---
schema: 1
id: 6gj5zz9efdtt
status: in-progress
epic: 20-cli-ux-and-ergonomics
description: Qualify the adapter-neutral refactor, new CLI surfaces, and planning safety fixes; publish machine revision 1.84 with explicit upgrade guidance.
effort: 1 day
tier: 1
priority: high
autonomy_level: 2
tags: [release, contracts, architecture, safety, dogfood]
created: "2026-10-09"
started_at: "2026-10-09"
updated_at: "2026-10-09"
depends_on: [6ggdkzv2tnta, 6ghht05bzxkk]
---

# Cut v0.23.0 as an adapter-boundary and planning-safety release

## Objective

Qualify v0.23.0 as a checkpoint for the completed adapter-neutral planning refactor, useful new
CLI surfaces, and safer planning writes. Publish concise, planning-linked notes with explicit
upgrade guidance for machine consumers. Keep Threads in preview and qualify one immutable
candidate rather than treating earlier PR checks as release evidence.

## Version decision and release basis

Recommend **v0.23.0**, not v0.22.1. A refactor-only or backward-compatible bug-fix release would
fit a patch, but the changes since v0.22.0 also add public capabilities and deliberately tighten
previously accepted CLI operations. Before 1.0 this is a versioning signal, not a claim that
[SemVer](https://semver.org/) mandates this increment.

- New capabilities include `audit finding new`, Thread-list table/CSV/projected JSON, runnable
  command safety discovery, and the reusable `actions/lint` GitHub Action.
- `init` now honors target selectors and rejects conflicting selections rather than silently
  ignoring them. Audit close/defer requires complete, terminal finding evidence; finding-status
  writes cannot reactivate a closed/deferred audit without an explicit reopen.
- JSON machine revision moves from **1.68 to 1.84**. Its monotonic revision is separate from binary
  SemVer; additions and non-additive changes are classified in the
  [revision history](../../internal/wire/wire.go). Do not describe this entire upgrade as additive.
- The large Go refactor is internal, not a new public Go-library API. Persisted frontmatter remains
  `schema: 1`; this release does not require a new on-disk schema or graduate Threads from preview.

The completed [adapter-neutral planning Thread](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
is the architectural basis. The latest safety fixes are merged in PRs
[287](https://github.com/andy-esch/taskflow/pull/287),
[288](https://github.com/andy-esch/taskflow/pull/288), and
[289](https://github.com/andy-esch/taskflow/pull/289). Main at preparation start is
`fcc4af49e9b776a2a05a2962df47ba1f8db10e26`; this is **not yet a qualified release candidate**.

## Release execution playbook

Land this planning change, then freeze a clean `main` candidate. Run both gates on the same commit;
keep their logs outside the checkout. If candidate files change before tagging, qualify again.

```bash
git status --short
git rev-parse HEAD
just release-validate
just release-validate-container
```

Build that candidate and check the read surfaces without modifying this planning tree:

```bash
just build
bin/tskflwctl version
bin/tskflwctl schema --json
bin/tskflwctl schema --json-schema
bin/tskflwctl thread list --json -c id,slug,status
bin/tskflwctl audit list --all --json -c id,slug,bucket,open_findings
bin/tskflwctl lint --links
just run ui
```

Confirm machine revision 1.84 and its non-additive classification, runnable command capabilities,
the complete exit taxonomy, and ordered stable-ID projections. Retrieve the current binary's
JSON Schema for typed-envelope validation; projected column views are not typed envelopes, and
the v0.22.0 exact-revision schema is not a validator for revision 1.84.

In a throwaway planning space, exercise `audit finding new`, blocked close/defer with unresolved
or unparsed findings, settled closure, and explicit reopen before reactivation. Check conflicting
`init` selectors fail without writes, and unrelated task field edits preserve multiline YAML
values. Use the fresh candidate binary, not an installed older release. For the TUI, inspect
Thread list/detail, spatial/neighbor views, task navigation/back, and refresh after an external
throwaway-space task edit. Record any regression as a named task and assess whether it blocks release.

After successful validation and note review, the maintainer tags the exact qualified commit.
Verify tag-triggered CD, all four Darwin/Linux amd64/arm64 archives, their published checksums,
an extracted binary's embedded version, and the installed binary. Record publication evidence in
this task after tagging; do not claim a source-built install is byte-identical to an archive.

## Draft release notes

### v0.23.0 — safer planning writes and stronger boundaries

- **A stronger foundation:** CLI and TUI planning access now use explicit portable capabilities,
  with guarded writes and recovery behavior owned by the core.
  [Refactor and closeout](../threads/6gcwd78p9r04-make-planning-data-access-adapter-neutral.md)
- **More useful automation:** create findings with `audit finding new`, select Thread-list
  columns, discover command safety and exit codes through `schema`, and lint planning in CI with
  the reusable GitHub Action.
  [Finding creation](6gbpe6e8n87k-add-a-tool-owned-audit-finding-creation-verb.md),
  [Thread projections](6gbn4g1eypzs-make-thread-list-projectable-like-other-planning-entity-lists.md),
  [Command discovery](6g63hhk3eddf-make-the-command-safety-annotations-load-bearing.md),
  [CI action](6gdwqmzxp7ka-publish-a-reusable-github-action-for-planning-tree-linting.md)
- **Safer edits:** preserve multiline frontmatter values, reject ambiguous `init` targets, and
  prevent closing incomplete audits or silently reactivating settled ones.
  [Frontmatter fidelity](6g1dhhk6721x-a-surgical-frontmatter-write-re-folds-multi-line-block-scalars-onto-one-line.md),
  [Target selection](6ggjmtmdd54w-prevent-silently-ignored-target-selectors-during-init.md),
  [Incomplete evidence](6gh82rm9sf3b-carry-unparsed-finding-evidence-into-audit-tui-and-lifecycle-decisions.md),
  [Settlement policy](6ghht05bzxkk-align-audit-lifecycle-and-finding-writes-with-parsed-settlement-policy.md)

**Upgrade:** JSON revision is now **1.84**, with both additions and deliberate behavior changes.
Refresh exact-revision validators from `schema --json-schema`, review target-selector usage, and
reopen audits before resuming findings. Markdown frontmatter remains `schema: 1`.
[Machine-contract policy](../adrs/0008-use-monotonic-revisions-for-the-json-machine-contract.md)
Threads remain **preview**. [Compatibility boundary](../../docs/THREADS_COMPATIBILITY.md)

Use repository URLs for these planning links when publishing the GitHub release notes.

## Acceptance criteria

- [x] The version choice and public changes since v0.22.0 are documented, and
  the completed adapter-boundary and audit-safety work is merged on main.
- [ ] Both just release-validate and just release-validate-container pass on the
  same recorded clean candidate commit.
- [ ] A fresh candidate binary passes the bounded CLI contract and
  throwaway-space mutation checks, plus the Thread navigation and refresh TUI
  smoke pass.
- [ ] Concise release notes link planning evidence, explain the non-additive
  machine revision 1.84 upgrade, and retain the Threads preview notice.
- [ ] The v0.23.0 tag, successful release workflow, four archives, checksums,
  extracted version, and installed binary identify the qualified candidate.
- [x] Known audit-body write, Thread-only TUI recovery, and bounded-query
  followups remain explicitly tracked rather than claimed fixed or made
  accidental release gates.

## Out of scope

- Graduating Threads from preview, adopting a new persisted schema, or making graph/UI redesign
  and pagination accidental release gates.
- Claiming every JSON change since 1.68 is additive or that internal port types are a public Go API.
- Claiming the container snapshot is bit-for-bit reproducible or publishing/tagging without
  maintainer direction.

## Known followups, not release claims

- [Remaining audit body-write guard gaps](6g77rn6hvmh8-close-the-remaining-audit-body-write-guard-gaps.md):
  arbitrary body/editor writes are not covered by the new finding-status guard. Do not claim
  universal enforcement on every audit write path.
- [Thread-only lifecycle recovery guidance in the TUI](6ggkdbg0816h-preserve-thread-only-lifecycle-recovery-guidance-in-the-tui.md):
  preserve actionable guidance when no task transition receipt accompanies a Thread failure.
- [Bounded list queries](6gbn4g1j40pj-bound-and-page-agent-facing-list-queries.md):
  remains separate from the shipped Thread-list projection surface.

## Related

- Epic [20-cli-ux-and-ergonomics](../epics/20-cli-ux-and-ergonomics.md)
- Release guide [Releasing taskflow](../../docs/RELEASING.md)
- Previous checkpoint [v0.22.0](6gbmzgsnzhhp-cut-v0.22.0-as-a-machine-contract-hardening-release.md)
- Settlement reviews:
  [Codex](../audits/6gj5ct38xq4t-2026-10-09-audit-parsed-settlement-implementation-codex.md),
  [Antigravity](../audits/6gj5ct3hyzjs-2026-10-09-audit-parsed-settlement-implementation-antigravity.md)

## Validation and publication evidence

Pending: clean candidate commit, host/container gate results, fresh-binary CLI/TUI dogfood,
approved notes, tag/CD/assets/checksums, and installed release version. Earlier PR checks are
implementation evidence, not completion of these release gates.
