---
schema: 1
id: 6fjjpfg16ss5
status: completed
epic: 24-data-model-evolution-stable-key-storage-read-model-content-occ
description: 'Narrowed post-Phase-B: route fix.go writes through the version-CAS write-lock (lint --fix is an unguarded second writer). The misfiled-move dup-slug edges are mooted by the flatten.'
effort: Unknown
tier: 3
priority: medium
autonomy_level: 3
tags: [core, storage]
created: "2026-07-03"
updated_at: "2026-07-06"
completed_at: "2026-07-06"
---
# Harden lint --fix misfiled-move for dup-slug edge cases

> **Narrowed 2026-07-04.** Phase B moots the original scope — the misfiled-move and the
> dup-slug class both disappear once the layout is flat + id-led (no folder to be misfiled
> against; id-led filenames are unique). The ONE surviving, Phase-B-independent concern:
> **`fix.go`'s writes/relocations bypass the version-CAS write-lock** (flagged during the OCC
> work) — `lint --fix` is a second writer that can clobber a concurrent edit. That's the real
> remaining work here; the dup-slug edges below are historical context.

## Objective

The Phase-A `lint --fix` misfiled-move (store/fix.go — collect `plannedMove` during
the walk, apply after) has three low-severity edge cases, all gated behind an
already-dup-slug-broken tree (the same slug in two status dirs, itself flagged by
`duplicateSlugIssues`). Non-destructive: a second `lint --fix` converges. Surfaced by
the Phase-A adversarial review (2026-07-03); deferred from the flatten task to keep the
1–4 PR low-risk.

## The three edges (store/fix.go)

1. **Chained-move skip (order-dependent).** When file A's target dir is occupied by
   file B (same slug) that is itself scheduled to move away, applying A first sees B
   still on disk → skips A as "occupied," even though B will vacate. Whether the chain
   resolves depends on `AllStatuses()` walk order.
2. **Dry-run ≠ real run for chains.** The real run removes an applied move's source
   (freeing a slot a later move can claim); the dry-run removes nothing, so the later
   move sees the slot occupied and is skipped. `lint --fix --dry-run` thus previews a
   different move set than `lint --fix` performs — contradicting the `taken`-map
   comment's "so a dry-run preview matches the real one."
3. **Skipped move drops the file's text/id repairs.** Text-normalization + id-backfill
   for a misfiled file are folded into `plannedMove.content` and only written if the
   move applies. If the move is skipped (target occupied), those valid in-place repairs
   are discarded — the file stays both misfiled AND unrepaired until the collision is
   resolved.

## Recommended fix

- **Decouple in-place repairs from the relocation.** Write text/id repairs to the
  source *during* the walk (so they persist regardless of the move), and make the move
  a pure relocation applied after. Fixes #3 directly.
- **Compute occupancy against the planned final state, not live disk.** A target is
  "occupied" only if a file is there AND it is not itself a pending-move source; order
  the applies (or reserve targets) so a chain resolves deterministically and the
  dry-run matches. Fixes #1 and #2.
- Tests: chained same-slug moves converge in one pass, deterministically; dry-run
  preview == real run for a chain; a skipped move still persists the file's id/text
  repairs.

## Out of scope / notes

- All three require a same-slug-in-two-dirs collision, which `duplicateSlugIssues`
  already flags — the user is told to resolve the dup first. Non-destructive.
- Don't reintroduce the mutate-while-iterating hazard: any relocation still applies
  after the walk completes.

## Related

- Epic [24-data-model-evolution-stable-key-storage-read-model-content-occ](../epics/24-data-model-evolution-stable-key-storage-read-model-content-occ.md)
- [flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory](6fhnydm03edq-flatten-layout-status-bucket-to-frontmatter-retire-status-equals-directory.md) (Phase A) — where the misfiled-move was introduced.

## Re-confirmed by adversarial review (2026-07-05)

A whole-branch adversarial review independently flagged this as **CRITICAL**: `FixFrontmatter`
(`lint --fix`) writes via `writeFileAtomic` without `s.writeLock()` or the `verifyUnchanged`
OCC check, so a concurrent `task move`/`edit`/`set` and a `lint --fix` can silently
lost-update each other. Low real-world exposure (single-user local CLI) but real — the fix
is to route fix.go's writes through the same write-lock + CAS the other write paths use.

## Landed (2026-07-06)

`FixFrontmatter` (`lint --fix`) now takes the repo write-lock (`s.writeLock()`) for the whole
non-dry-run pass — reads happen inside the lock too, so each file's read→fix→write is atomic
against the other write paths (move/set/edit) and can't clobber a concurrent agent write.
Dry runs read only, so they stay lock-free. (No version-CAS needed: holding the lock across
the read makes the read→write window uninterruptible by another tool writer.)
