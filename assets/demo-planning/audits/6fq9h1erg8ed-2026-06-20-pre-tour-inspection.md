---
schema: 1
id: 6fq9h1erg8ed
bucket: open
area: pre-tour-inspection
date: "2026-06-20"
---
# Audit: pre-tour-inspection — 2026-06-20

> Pre-departure safety sweep of the touring bike. Edit findings in place and flip
> each `**Status:**` as you work it.

## Findings

#### H1. Rear brake pads worn past the wear line  · **Status:** fixed (2026-06-20)

**Component:** brakes
**Effort:** S · **Urgency:** acute

Both rear pads are down to the metal backing, glazed, and biting late — on a
loaded bike descending a pass that's a stopping-distance problem, not a comfort
one.

**Recommendation:** fit fresh cartridge pads, bed them in, and re-check toe-in.

#### H2. Chain stretched beyond 0.75%  · **Status:** fixed (2026-06-20)

**Component:** drivetrain
**Effort:** S · **Urgency:** soon

The wear gauge drops fully in at 0.75%. Left on, it will start skating over the
cassette under climbing load and chew the chainrings.

**Recommendation:** replace the chain; re-measure the cassette for hooking.

#### H3. Headset has a notchy index at center  · **Status:** landed (2026-06-19)

**Component:** headset
**Effort:** M · **Urgency:** soon

Bars settle to dead-center on their own — the bearing races are brinelled from
years of braking loads, so the steering catches straight ahead.

**Recommendation:** replace both cartridge bearings; grease and re-preload.

#### H4. Rear derailleur hanger is slightly bent inboard  · **Status:** in-progress

**Component:** drivetrain
**Effort:** S · **Urgency:** soon

Indexing won't hold across the two largest cogs; a hanger-alignment gauge shows
~4mm of inboard lean at the rim.

**Recommendation:** cold-set the hanger true with the gauge, then re-index.

#### H5. Front tire sidewall is cracking  · **Status:** open

**Component:** tires
**Effort:** S · **Urgency:** soon

Fine crazing runs the length of both sidewalls where the casing flexes. It holds
air today, but not a 60-mile day in the heat with a load.

**Recommendation:** replace with a touring-rated tire; keep the old one as a boot.

#### H6. Bottom bracket creaks under climbing load  · **Status:** open

**Component:** bottom-bracket
**Effort:** M · **Urgency:** soon

A sharp creak tracks pedal pressure, loudest out of the saddle. Could be the BB
cups, the crank interface, or the pedals — needs isolating.

**Recommendation:** pull the crank, degrease and re-torque the interfaces, then
retest before condemning the BB.

#### H7. Bar tape is frayed at the hoods  · **Status:** wontfix

**Component:** cockpit
**Effort:** XS · **Urgency:** eventually

Cosmetic wear where the hands rest. The rider likes the broken-in grip and wants
it left alone until after the trip.

**Recommendation:** none for now; re-wrap post-tour.

#### H8. Front fender rattles on washboard  · **Status:** deferred

**Component:** fenders
**Effort:** XS · **Urgency:** eventually

The fender stay buzzes on rough gravel. Annoying, not a safety issue; deferred
behind the drivetrain and brake work.

**Recommendation:** add a rubber grommet at the stay mount.

## Candidate tasks

- ✅ `tskflwctl task new "Replace chain and cassette" --epic 01-touring-bike-repairs --tags drivetrain` — chain past 0.75%
- ⏳ `tskflwctl task new "Overhaul bottom bracket" --epic 01-touring-bike-repairs --tags bearings` — isolate the climbing creak
