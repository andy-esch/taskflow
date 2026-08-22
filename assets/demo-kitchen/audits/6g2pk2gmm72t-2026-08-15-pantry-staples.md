---
schema: 1
id: 6g2pk2gmm72t
bucket: open
area: pantry-staples
date: "2026-08-15"
---

# Audit: pantry-staples — 2026-08-15

> A sweep of the pantry against what the dal, banh mi, and potato work
> actually needs on hand. Edit findings in place and flip each `**Status:**`
> as you work it.

## Findings

#### H1. Garam masala eight months past opening  · **Status:** fixed (2026-08-15)

**Component:** spices
**Effort:** XS · **Urgency:** soon

The jar's been open since December — ground spice loses most of its aromatic
oil within six months, and the tadka work depends on the spices actually
blooming with a real scent, not just coloring the oil.

**Recommendation:** toast and grind a fresh batch; date the jar going forward.

#### H2. Toor dal and masoor dal sharing one unlabeled bin  · **Status:** fixed (2026-08-15)

**Component:** legumes
**Effort:** XS · **Urgency:** soon

Grabbed toor dal by feel for a masoor dal recipe last week — same color range
in low kitchen light, very different cook times. Nothing caught it until the
lentils were still hard at the 20-minute mark.

**Recommendation:** separate airtight containers, labeled, one lentil per jar.

#### H3. Rice flour bag clumping from moisture  · **Status:** in-progress

**Component:** flour
**Effort:** S · **Urgency:** soon

The opened bag lives next to the stovetop, and the rice flour is picking up
enough ambient moisture to clump — a problem for the banh mi baguette work,
where the rice-flour ratio in the dough needs to be measured precisely, not
estimated around a clumped scoop.

**Recommendation:** move to an airtight container with a silica packet;
weigh a sample to confirm it's still dry enough to measure accurately.

#### H4. Frying oil reused past a safe number of fries  · **Status:** open

**Component:** oil
**Effort:** S · **Urgency:** acute

The same batch of oil has done three rounds of chip-frying without a full
strain and smoke-point check — it's gone darker and started smoking well
under its rated temperature, which will show up as bitterness in the next
chip batch.

**Recommendation:** strain and taste-test the current batch; replace outright
if the smoke point has dropped much below 400°F.

#### H5. Asafoetida tin off-gassing onto neighboring spices  · **Status:** wontfix

**Component:** spices
**Effort:** XS · **Urgency:** eventually

Hing's sulfurous smell has been noticeable on the cumin and coriander stored
next to it in the drawer, even double-sealed. Mildly annoying but not
functionally hurting anything cooked so far.

**Recommendation:** none — already double-sealed; not worth a dedicated
separate storage spot right now.

#### H6. Coconut oil stock too low for the duck-fat-style potato test  · **Status:** deferred

**Component:** oils
**Effort:** XS · **Urgency:** eventually

Only about a third of a jar of refined coconut oil left, not enough to steep
and roast a full batch for the vegetarian duck-fat-style roast potatoes task.
Deferred behind the frying-oil issue (H4) since that one's acute and this one
isn't blocking anything active.

**Recommendation:** pick up a fresh jar next grocery run; no rush.

## Candidate tasks

- ✅ Fresh garam masala ground and dated — no task needed, done in the moment.
- ✅ Lentil bins relabeled and separated — no task needed, done in the moment.
- ⚠️ `tskflwctl task new "Re-store rice flour airtight with a desiccant" --epic 02-banh-mi-completely-from-scratch --tags pantry` — protect the baguette rice-flour ratio from moisture
- ⏳ `tskflwctl task new "Strain and smoke-test the frying oil" --epic 03-potato-dominant-recipes --tags pantry,frying` — acute, blocks the next chip batch
- ⛔ hing storage — no task, tolerating the smell as-is
- ⏳ coconut oil restock — folded into the next grocery run, not a tracked task
