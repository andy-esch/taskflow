---
schema: 1
id: 6fq4mk2dhwdm
bucket: closed
area: data-pipeline
date: "2026-06-10"
---
# Audit: data-pipeline — 2026-06-10

> Edit findings in place and flip each `**Status:**` as you work it.

## Findings

#### H1. Backfill re-reads the full table per run  · **Status:** fixed (2026-06-11)

**File:** pipeline/backfill.go:64 | **Component:** backfill
**Effort:** M · **Urgency:** soon

Each backfill scans the whole events table instead of resuming from the last
watermark, so reruns are O(n) and slow.

**Recommendation:** persist a watermark and resume from it.

#### H2. Late events double-count in daily rollups  · **Status:** fixed (2026-06-11)

**File:** pipeline/rollup.go:40 | **Component:** rollup
**Effort:** M · **Urgency:** soon

Events arriving after the window closes are counted twice — once late, once on
recompute.

**Recommendation:** dedupe by event id within the rollup window.

#### H3. No compression on cold-storage writes  · **Status:** landed (2026-06-10)

**File:** pipeline/storage.go:88 | **Component:** storage
**Effort:** S · **Urgency:** eventually

Cold partitions are written uncompressed, tripling object-store cost.

**Recommendation:** zstd the cold-tier writes.

#### H4. Sampling drops error logs too  · **Status:** wontfix

**File:** pipeline/sample.go:22 | **Component:** sampling
**Effort:** S · **Urgency:** eventually

Uniform sampling also drops error-level logs — but error retention is handled by
the separate alerting path, so this is intentional here.

**Recommendation:** none; documented.

## Candidate tasks

- ✅ `tskflwctl task new "Backfill events table" --epic 03-data-pipeline --tags data` — resume from watermark
