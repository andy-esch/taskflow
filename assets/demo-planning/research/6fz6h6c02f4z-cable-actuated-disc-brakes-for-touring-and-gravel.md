---
schema: 1
id: 6fz6h6c02f4z
created: "2026-08-12"
description: Weighed mechanical vs hydraulic disc calipers for field repair on tour and gravel racing.
tags: [brakes, touring, gravel]
---
# Cable-actuated disc brakes for touring and gravel

## Question

Both the touring bike and the gravel bike are still on their original disc
calipers. Before committing either one to a brake swap: does a hydraulic
system actually earn its keep on a bike that spends weeks away from a shop, or
does cable-actuated (mechanical) win on the property that matters most out
there — nothing in the system that needs a bleed kit to fix?

## Findings

**Hydraulic, for the record.** Sealed fluid path gives lighter lever effort,
smoother modulation, and opposed pistons that self-center so both pads wear
evenly without adjustment. The cost is the failure mode: a cut hose or a
contaminated caliper needs the correct fluid (mineral oil and DOT are not
interchangeable, and mixing them wrecks seals), a syringe, and a still
work surface. None of that exists in a hardware store in a small town, and it
doesn't travel well strapped to a rear rack either.

**Mechanical, single-piston (Avid BB7 and clones).** Only the outer pad
moves; the rotor gets pushed sideways onto the fixed inner pad. Simple, cheap,
rebuildable with a 2.5mm hex, but it needs the inner pad dialed in by hand as
it wears — skip that and the rotor rubs constantly or the lever goes long.
Fine on a bike someone checks weekly; a nuisance on tour when everything else
also wants attention at the end of a riding day.

**Mechanical, dual-piston (TRP Spyre, Hy/Rd-style hybrids).** Both pads move
toward the rotor together, actuated by a cable pulling a cam that drives two
pistons. Closer to hydraulic pad wear behavior — no manual inner-pad
adjustment — while keeping an all-cable actuation path. This is the
meaningful upgrade over BB7-style calipers, not a switch to hydraulic.

**Housing quality matters more than the caliper.** Compressionless
(linear-strand) housing, not standard coil-wound derailleur housing, is what
keeps the lever firm under a long braking effort on a loaded descent —
coil-wound housing compresses slightly under sustained load and the lever
feels mushy exactly when it shouldn't. Full-length housing (no exposed cable
segments between stops) keeps grit and road spray out of the run, which
matters more on gravel than on pavement.

**Rotor sizing.** 160mm rear / 180mm front is the baseline for a loaded
tourer; 203mm front is worth it on a route with sustained descents, since a
bigger rotor buys leverage and heat capacity, which lowers required lever
force and narrows the modulation gap with hydraulic more than any caliper
choice does.

**Field repair, the actual argument.** A cable can be spliced or replaced
with a spare length of housing and a ferrule kit carried in a saddle bag,
using tools already in the tour kit (cable cutter, 2-3mm hex set). A caliper
that's dragging or dead can be diagnosed and adjusted trailside without
bleeding anything. That's the whole case for mechanical on a self-supported
trip — not raw stopping power, but that every failure mode is fixable with
what's already being carried.

**Power ceiling, honestly assessed.** A quality dual-piston mechanical
caliper on a 180-203mm rotor, with a matched drop-bar lever (Tektro RL340 for
flat bar, CR720/HyRd-compatible levers for drop bar), gets close enough to
hydraulic for loaded-touring and gravel-race speeds. It will not match a
4-piston hydraulic enduro caliper on a steep, fast descent — but that's not
the riding this bike does.

## Recommendation (as of 2026-08-12)

Spec dual-piston mechanical calipers (TRP Spyre or equivalent) over
single-piston BB7-style calipers for both the touring bike and the gravel
bike, paired with full-length compressionless housing and 180mm rotors front,
160mm rear. This closes most of the pad-wear and modulation gap with
hydraulic while keeping every failure mode fixable with a cable-cutter and a
hex key — no bleed kit joins the tour packing list. Revisit only if a future
build carries a compact hydraulic bleed kit as a matter of course, or if the
route changes to something with sustained enough descents that raw power
starts to matter more than field repair.

## Related

- Epic [01-touring-bike-repairs](../epics/01-touring-bike-repairs.md)
- Epic [02-gravel-bike-upgrades](../epics/02-gravel-bike-upgrades.md)
