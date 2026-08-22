---
schema: 1
id: 6g1rxxc01qfb
created: "2026-08-20"
description: Weighed switching the BMX respray to powder coat mid-job vs finishing the wet-paint plan already underway.
tags: [paint, bmx]
---
# Powder coating a previously painted steel frame

## Question

The BMX respray is mid-stream — bare metal, etch-primed, wet-sanded — with
clear-coat-and-reassemble deferred to 2026-09-01. Before that task comes back
up: is switching the rest of the job over to powder coat worth the detour, or
does finishing the wet-paint plan already committed still make more sense?
Also worth settling for next time, on a frame that hasn't been touched yet.

## Findings

**Stripping the old finish.** Media blasting (glass bead or plastic bead,
not sand — sand cuts too aggressively for thin-wall tubing) strips fast and
won't touch bare steel if the operator doesn't linger in one spot, but it can
peen or warp thin seat stays and chain stays if it does. Chemical stripper
(aircraft-grade paint remover) is slower and more controllable right at
brazed joints and lug edges, but every trace has to be neutralized and rinsed
or it keeps working under the next coat and lifts it months later.

**This frame is already past that step.** `strip-old-paint` took it to bare
metal and `prime-and-sand-frame` already put an etch primer down. Powder
needs bare, clean, unprimed steel to bond properly — going to powder now
means stripping the etch primer back off, which throws away completed work
for no finish benefit at this stage.

**Masking.** Threads (BB shell, seatpost clamp) and bearing races (head tube
cups) need high-temp silicone plugs or oven-rated tape, not standard painter's
tape — it softens and off-gasses at cure temperature and leaves residue in
exactly the surfaces that need to stay precise. Powder adds real thickness;
unmasked bearing seats end up out of tolerance for a proper press fit.

**Degreasing and outgassing.** A steel frame with this much history — brazed
joints, years of prior paint — holds trapped oil and moisture in the tube
walls and lug junctions. A pre-bake around 200°C (400°F) for ten minutes
drives that out before powder goes on; skip it and the trapped gas escapes
through the curing coat as it flows, pinholing the finish.

**Cure temperature risk.** Thermoset polyester powder cures around
190-205°C (375-400°F) for 10-20 minutes. That's nowhere near the 620-800°C a
silver-braze joint flows at, so a cure cycle poses no risk to a brazed steel
frame's joints. The real risk is dwell time and even heat distribution — a
tube hung wrong on the rack, or held too close to an element, can warm
unevenly and warp a thin-wall section even at a "safe" bulk temperature.
(Heat-treated aluminum is the frame material where cure temp actually matters
against the material's temper — worth flagging for a future aluminum build,
not a concern on this steel one.)

**Primer choice.** Powder poly goes direct onto properly blasted, degreased
bare steel without a separate primer coat in most cases — a powder primer
mainly earns its keep on substrates that don't take powder well on their own
(cast parts, non-ferrous metal), not on clean strip steel.

**Single vs. multi-coat.** A base color coat under a dedicated clear top coat
— two full cure cycles — is what protects a neon color from UV chalking and
gives it depth, the same logic the current wet-paint plan is already
following with base coats then clear. Single-coat colored powder exists but
skips that UV protection layer.

**DIY vs. a shop.** A home setup needs an oven with full-frame clearance
(most household ovens are too short without tenting the dropout ends out the
door), a corona-gun powder sprayer ($200-600 to start), and a way to hold
cure temperature evenly across the whole part — a real investment for a
one-off frame. A local powder shop costs more per frame than the spray-gun
wet-paint job already budgeted, but returns a harder, more chip-resistant
finish and a same-week turnaround instead of the cure-and-sand-between-coats
timeline the wet job is running.

## Recommendation (as of 2026-08-20)

Finish the wet-paint plan already in motion. The primer coat is down; backing
out to strip for powder now spends real time to recover ground already
covered, for no gain on a frame that's midway through a job, not starting
one. When the next frame comes up bare from the start — no primer sunk into
it yet — default to powder coat instead: better chip resistance for a bike
that lives at a skatepark, provided oven access and a masking plan for the BB
shell and head tube are sorted before it goes in.

## Related

- Epic [03-bmx-neon-paint-job](../epics/03-bmx-neon-paint-job.md)
- Task [prime-and-sand-frame](../tasks/6fq9h1e0vpv5-prime-and-sand-frame.md)
- Task [clear-coat-and-reassemble](../tasks/6fq9h1e3qmyw-clear-coat-and-reassemble.md)
