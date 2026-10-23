---
id: terrain/rBridge
title: Road bridge
kind: terrain
status: candidate
runtime_ids: rBridge14, rBridge25, rBridge36, standing/rBridge14, standing/rBridge25, standing/rBridge36
masters: rbridge-14-shared-v1.png, rbridge-25-shared-v1.png, rbridge-36-shared-v1.png, rbridge-14-export-shared-v1.png, rbridge-25-export-shared-v1.png, rbridge-36-export-shared-v1.png, rbridge-stone-ne-shared-v1.png, rbridge-stone-se-shared-v1.png, rbridge-stone-vertical-shared-v1.png
---

## Look
Road bridge over a stream. In the reference the asphalt road with its dashed centre line crosses on a
concrete deck with low pale parapets, the creek passing underneath between stony banks.

## Base prompt
Transparent cut-out of a short concrete road bridge seen straight from above: grey asphalt carriageway
with three restrained off-white centre dashes, low pale concrete parapets on both sides, small stone
abutments, finely cracked and patched tarmac, masonry seams and sparse moss. Strict nadir deck plan,
soft upper-left light and short lower-right shadow. The deck runs straight along its declared axis;
no water, ground, approach road, text, letters, flags or other objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| rBridge14 | rBridge14, standing/rBridge14 | Separately painted NE/SW bridge, axis -29.396 degrees from horizontal. |
| rBridge25 | rBridge25, standing/rBridge25 | Separately painted SE/NW bridge, axis +29.396 degrees from horizontal. |
| rBridge36 | rBridge36, standing/rBridge36 | Separately painted vertical N/S bridge. |

## Edges & joins
The finite asphalt deck is registered to 48 logical pixels along its port axis and 22 pixels across,
centred on `(48,40)` within a 96x80 canvas. Parapets, posts, stone abutments and a short shadow extend
beyond the deck; parapets are irregular and slightly wider than the nominal three pixels per side.
An independent road layer provides the approaches and the stream passes underneath.

## Post

### Look 3 B3 candidate (2026-10-23)

`rbridge-stone-vertical-shared-v1.png` (`exec-ad14f8f9`, native 1373×1145 RGBA, unchanged) is a
vertical stone road bridge carrying the R1 macadam (two darker wheel tracks, no deck plate, no dashes)
between balustrade parapets with posts and short wing walls: 71.3 × 35.1 logical. Faults: 26.6
logical between the parapets (spec 22–24) but sandy verges take a third of it, so the grey carriageway
is about 14 logical against R1's 23; the generator keeps the length to width ratio, so scaling cannot
fix both. Not selected. Attached: RefDesign2 crop r4c1, the earlier `exec-07b389c2` (parapets only),
the R1 candidate and the first B3 try. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored STONE ROAD BRIDGE cutout on true alpha transparency, canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir plan. Orientation: vertical. The carriageway is 22 to 24 logical px wide and continues the attached macadam road exactly (same pale grey-ochre surface with two darker wheel tracks), bounded by two low pale stone parapets 2 to 3 px wide that run 64 to 70 logical px along the axis, with small end posts and short stone wing walls. No separate deck colour: the road simply crosses. A soft translucent shadow of each parapet falls 2 to 3 px toward the lower right.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). The previous try made the road between the parapets only 250 px wide, narrower than the road it carries: make it clearly wider, 330 px, and keep the bridge the same length. In canvas pixels (14.3 canvas px per logical px): the bridge runs vertically, centred on x=687. The carriageway between the parapets is 330 px wide (24 percent of the canvas width), x=522 to x=852. Each parapet is 36 px wide and 960 px long, from y=92 to y=1052, so at least 90 px of empty alpha-zero canvas stays above and below; the short wing walls angle outward at both ends by at most 60 px. The carriageway surface stops exactly at the parapet ends; nothing beyond them. Everything outside the bridge and the parapet shadows is alpha exactly zero.
Avoid: a box or plate wider than the road, a different deck colour, painted lines, centre dashes, asphalt, arches, towers, perspective sides, guard booths, barriers, water, banks, approach road beyond the parapets, text, outline, checkerboard.
Attached references: image 1 is a crop of the reference painting with a road bridge over a stream: take its parapet weight and painted finish only, not its asphalt or dashes; image 2 is an earlier bridge: take only its stone parapet and wing-wall detail, never its asphalt surface or white dashes; image 3 is the macadam road strip this bridge carries: its carriageway colour, grit and two darker wheel tracks must continue across the bridge unchanged. Image 4 is the previous try of this bridge: keep its parapets, posts, colours and surface, only widen the carriageway to 330 px between the parapets.
```

### Look 3 B4 candidate (2026-10-23)

`rbridge-stone-se-shared-v1.png` (`exec-a22c9b7a`, native 1453×1083 RGBA, unchanged; the request was
1374×1145) is the SE/NW stone road bridge, painted at 31.1°, matching B5. Faults: the parapets show a
little side face, the verges are wide as in B3, and the canvas aspect is off. The first try ran the road
past the parapets to the canvas edges. Not selected. Attached: the B5 candidate (to mirror) and the R1
candidate. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored STONE ROAD BRIDGE cutout on true alpha transparency, canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir plan. Orientation: SE/NW axis at 29.4 degrees from horizontal: the bridge runs diagonally from the upper left down to the lower right. The carriageway is 22 to 24 logical px wide and continues the attached macadam road exactly (same pale grey-ochre surface with two darker wheel tracks), bounded by two low pale stone parapets 2 to 3 px wide that run 64 to 70 logical px along the axis, with small end posts and short stone wing walls. No separate deck colour: the road simply crosses. A soft translucent shadow of each parapet falls 2 to 3 px toward the lower right.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). The previous try let the road run on past the parapets to the canvas edges and cut it off there: wrong. The road surface exists ONLY between the two parapets and ends in a straight line at their wing walls; the whole bridge, wing walls included, sits well inside the canvas. In canvas pixels (14.3 canvas px per logical px): the road centreline runs straight from (269,336) at the upper left to (1105,808) at the lower right, a slope of 29.4 degrees through the canvas centre (687,572). This is a SHALLOW diagonal, much closer to horizontal than to 45 degrees: the bridge spans about 840 px across and only about 470 px up or down, almost twice as wide as tall. The road between the parapets is 330 px wide, measured across the bridge; parapets 36 px wide; the short wing walls angle outward at both ends by at most 60 px. The carriageway surface stops at the parapet ends. At least 100 px of empty alpha-zero canvas stays at the left and right edges. Everything outside the bridge and the parapet shadows is alpha exactly zero.
Avoid: a 45-degree diagonal, a vertical or horizontal bridge, a box or plate wider than the road, a different deck colour, painted lines, centre dashes, asphalt, arches, towers, perspective sides, guard booths, barriers, water, banks, approach road beyond the parapets, text, outline, checkerboard.
Attached references: image 1 is the NE/SW stone road bridge of this set: match its parapets, posts, wing walls, road surface, colours and finish exactly, but mirror its direction so it runs from the upper left down to the lower right; keep the light from the upper left. Image 2 is the macadam road strip this bridge carries.
```

### Look 3 B5 candidate (2026-10-23)

`rbridge-stone-ne-shared-v1.png` (`exec-5c45b9e1`, native 1453×1083 RGBA, unchanged; the request was
1374×1145) is the NE/SW stone road bridge at 29.1°; the road ends at the wing walls and continues the
R1 macadam. Faults: the parapets show a little side face, wide verges, canvas aspect off. Not selected.
Attached: the B3 candidate and the R1 candidate. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored STONE ROAD BRIDGE cutout on true alpha transparency, canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir plan. Orientation: NE/SW axis at -29.4 degrees from horizontal: the bridge runs diagonally from the lower left up to the upper right. The carriageway is 22 to 24 logical px wide and continues the attached macadam road exactly (same pale grey-ochre surface with two darker wheel tracks), bounded by two low pale stone parapets 2 to 3 px wide that run 64 to 70 logical px along the axis, with small end posts and short stone wing walls. No separate deck colour: the road simply crosses. A soft translucent shadow of each parapet falls 2 to 3 px toward the lower right.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). In canvas pixels (14.3 canvas px per logical px): the road centreline runs straight from (269,808) at the lower left to (1105,336) at the upper right, a slope of 29.4 degrees through the canvas centre (687,572). This is a SHALLOW diagonal, much closer to horizontal than to 45 degrees: the bridge spans about 840 px across and only about 470 px up or down, almost twice as wide as tall. The road between the parapets is 330 px wide, measured across the bridge; parapets 36 px wide; the short wing walls angle outward at both ends by at most 60 px. The carriageway surface stops at the parapet ends. At least 100 px of empty alpha-zero canvas stays at the left and right edges. Everything outside the bridge and the parapet shadows is alpha exactly zero.
Avoid: a 45-degree diagonal, a vertical or horizontal bridge, a box or plate wider than the road, a different deck colour, painted lines, centre dashes, asphalt, arches, towers, perspective sides, guard booths, barriers, water, banks, approach road beyond the parapets, text, outline, checkerboard.
Attached references: image 1 is the vertical stone road bridge of this set: match its parapets, posts, wing walls, road surface, colours and painted finish exactly, only turned to the diagonal described above; image 2 is the macadam road strip this bridge carries: its carriageway colour, grit and two darker wheel tracks continue across the bridge unchanged.
```

Built-in imagegen produced three independent true-alpha paintings: native 1486x1058 (14),
1497x1051 (25) and 1374x1145 (36). The previous tiles supplied axis references and the selected crater
supplied fine material detail. Follow-up prompts widened the asphalt and restored the two diagonals;
no orientation is a rotated copy of another painting.

The `-export-` masters apply bounded affine registration and transparent padding at 1152x960
(12 source pixels per logical pixel). `rbridge-registration-v1.json` records native hashes, dimensions,
reviewed deck landmarks and every matrix. The smallest native sampling density is 15.09 pixels per
logical pixel, so both axes reduce. Small axis corrections retain the painted upper-left light.

`terrain-stamps -manifest art/pipeline/bridge-standing-v1.json` exports the prepared sources as
`standing/rBridge*`, with 2x 192x160 and 4x 384x320 PNGs and anchor `(48,40)`. The v2 mixed manifest's
`bridge_artwork: standing` emits transparent road-bridge tile overlays, avoiding the old timber
underpainting. The default v1 mixed export remains available with its derived bridge artwork.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- arches, towers or perspective
