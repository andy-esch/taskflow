---
schema: 1
id: 6fq9h1et173b
bucket: closed
area: wheel-build-qc
date: "2026-06-10"
---
# Audit: wheel-build-qc — 2026-06-10

> Quality check on the freshly built touring rear wheel before it goes on the
> bike. Edit findings in place and flip each `**Status:**` as you work it.

## Findings

#### H1. Drive-side spoke tension uneven  · **Status:** fixed (2026-06-11)

**Component:** wheels
**Effort:** M · **Urgency:** soon

The tension meter shows a 25% spread across the drive side — a few loose spokes
that will detension and go pingy under a loaded rack.

**Recommendation:** even the tension to within 10%, stress-relieve, re-check.

#### H2. Rim not dished to center  · **Status:** fixed (2026-06-11)

**Component:** wheels
**Effort:** S · **Urgency:** soon

The rim sits ~2mm to the non-drive side in the dishing gauge, so the wheel won't
center in the frame.

**Recommendation:** add drive-side tension to pull the rim to true center.

#### H3. Nipples creak under first load  · **Status:** landed (2026-06-10)

**Component:** wheels
**Effort:** XS · **Urgency:** eventually

New nipples creak against the rim eyelets on the first hard pedal — dry seats,
not a build fault.

**Recommendation:** a drop of oil at each nipple/rim interface; re-stress.

#### H4. Slight radial hop at the valve hole  · **Status:** wontfix

**Component:** wheels
**Effort:** S · **Urgency:** eventually

~0.3mm of radial runout at the seam. It's within touring tolerance and chasing it
would unbalance the lateral true.

**Recommendation:** none; within spec.

## Candidate tasks

- ✅ `tskflwctl task new "True both wheels" --epic 01-touring-bike-repairs --tags wheels` — even tension + dish
