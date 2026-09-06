---
schema: 1
id: 6g7ddfhh2jc2
status: next-up
epic: 30-threads-and-task-dependency-graphs
description: After the v0.20 preview soak, explicitly decide whether evidence supports removing the preview label or requires more named work.
effort: S
tier: 2
priority: medium
autonomy_level: 4
tags: [threads, release, compatibility, dogfood]
created: "2026-09-06"
depends_on: [6g7fhfpmy032]
updated_at: "2026-09-06"
---

# Graduate Threads from preview

## Objective

After the compatibility-hardened v0.20 checkpoint has shipped and accumulated real use, explicitly
open the evidence-based graduation decision from a clean release candidate. Remove the preview
label only when every required gate passes and the preview owner judges the soak evidence sufficient;
otherwise retain it, record the missing evidence, and sequence a named owner.

## Scope

- Verify gates G1–G6 in `docs/THREADS_COMPATIBILITY.md` against one clean candidate, recording the
  exact commit, binary version, commands, fixture results, and any exceptions here.
- Review findings from the v0.18.0, v0.19.0, and compatibility-hardened v0.20.0 previews. A clean
  gate run is necessary but does not force graduation when real-use evidence remains too thin.
- Repeat the throwaway-space workflow with shared membership, external gates, fan-out/fan-in,
  lifecycle changes, bulk-apply retry, deliberately broken and repaired dependency evidence, TUI
  stable-ID navigation, and watcher reload.
- Confirm the production implementation Thread and repository lint are healthy.
- Complete G7a: from the commit that passed G1–G6, align README, ADR, architecture, generated CLI
  docs/schema, and concise release notes; remove the preview notice; then run the release snapshot
  and full validation again on that exact candidate before tagging it.
- Complete G7b after tag publication: verify the tag, release workflow, binaries, and checksums all
  identify the G7a candidate. Retry a failed workflow for the same tag. If evidence proves the
  published candidate itself invalid, ship a patch release restoring the preview notice rather
  than rewriting the tag.
- If any pre-tag gate fails, do not tag the candidate; leave this task open and add or update its
  owning task and graph edge.

## Acceptance criteria

- [ ] `pin-thread-document-and-plan-backward-compatibility` is completed and its historical fixtures
      pass on the release candidate.
- [ ] The compatibility-hardened v0.20.0 preview is published and its recorded soak findings have
      either been resolved, explicitly accepted as non-blocking, or assigned to named follow-ups.
- [ ] G1–G5 have fresh automated evidence from the candidate, including the named hostile tests in
      the contract, full race tests, lint, generated docs/schema checks, and guarded-repair/mutation
      compatibility coverage.
- [ ] G6 dogfood passes from a clean binary and records the tested commit and commands here.
- [ ] G7a documentation and release checks pass on the preview-removal candidate; release notes
      state the compatibility boundary and keep optional spatial/ranking/portable-convenience-view
      work out of the gate.
- [ ] The preview-removal candidate descends directly from the recorded G1–G6 commit, is itself
      fully revalidated, and is the only commit tagged.
- [ ] G7b records that the immutable tag and published artifacts identify the G7a candidate with
      expected checksums, or records the named recovery action if publication/candidate evidence
      fails.

## Failure rule

Do not check off or complete this task after a partial pass. Before tagging, record the exact failed
gate, link its owner, add any real dependency edge, and retain or restore the preview notice until a
later clean candidate is re-run. After tagging, never rewrite the tag: retry publication for the
same candidate, or use a patch release to restore the preview notice if the candidate itself is
invalid. Passing the automated gates does not override an explicit decision to gather more preview
usage; in that case keep this task open and add another bounded preview checkpoint only when useful.

## Out of scope

- Requiring the two-dimensional graph experiment, frontier ranking metadata, a remote/database
  adapter, or portable board/status diagnostics merely to remove the local Thread preview label.
- Treating v0.20.0 as an automatic graduation deadline.
- Declaring all of taskflow 1.0-stable.

## Related

- Epic [30-threads-and-task-dependency-graphs](../epics/30-threads-and-task-dependency-graphs.md)
- Contract [Threads compatibility and preview graduation](../../docs/THREADS_COMPATIBILITY.md)
- ADR [0006 — Adopt Threads as task DAGs](../adrs/0006-adopt-threads-as-task-dags.md)
- Thread [Complete production Threads](../threads/6g503c6pfqeb-complete-production-threads.md)
