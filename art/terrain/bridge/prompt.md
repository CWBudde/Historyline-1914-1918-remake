---
id: terrain/bridge
title: Footbridge
kind: terrain
status: candidate
runtime_ids: bridge25, bridge36, standing/bridge25, standing/bridge36
masters: bridge-25-shared-v1.png, bridge-36-shared-v1.png, bridge-25-export-shared-v1.png, bridge-36-export-shared-v1.png, bridge-plank-diagonal-shared-v1.png, bridge-plank-vertical-shared-v1.png
---

## Look
Narrow plank footbridge over a stream. In the reference it is a weathered grey timber deck with thin
side rails, crossing the blue creek between stony banks.

## Base prompt
Transparent cut-out of a narrow timber plank footbridge seen straight from above: weathered grey-brown
boards laid across the span with fine grain, nailheads and dark seams, thin wooden side rails and
posts, short abutments of pale stone at both ends. Strict nadir deck plan, soft upper-left light and
short lower-right shadow. The deck runs straight along its declared axis; no water, ground,
approach path, letters, flags or other objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| bridge25 | bridge25, standing/bridge25 | Separately painted SE/NW bridge, axis +29.396 degrees from horizontal. |
| bridge36 | bridge36, standing/bridge36 | Separately painted vertical N/S bridge. |

## Edges & joins
The finite timber deck is registered to 48 logical pixels along its port axis and 12 pixels across,
centred on `(48,40)` within a 96x80 canvas. Thin rails, posts and stone abutments extend beyond the deck.
The path continues beyond each deck end and the stream passes underneath.

## Post

### Look 3 B1 candidate (2026-10-23)

`bridge-plank-vertical-shared-v1.png` (`exec-4ac11a26`, native 1374×1145 RGBA, unchanged) is a
vertical plank footbridge: deck about 63.6 × 12 logical with round log handrails and small stone
blocks at the corners; the whole bridge is 66.1 × 20.8 logical. Faults: rails a little thick and orange
against the grey deck, abutments only corner blocks, faint shadow, alpha haze 3 % in the margins. Not
selected; the earlier `exec-62ac6b08` and `exec-3b558b9a` were re-registered for comparison
(review only). Attached: RefDesign2 crop r33c7+r37c7. The prompt also carried two stray lines of the
W1 water prompt after the Terrain text (a copy error; the result shows no water). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored TIMBER FOOTBRIDGE isolated terrain cutout on true alpha transparency, 6:5 canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir deck plan, no tilt. Orientation: vertical north-south, deck ends at (48,10) and (48,70).
Deck: weathered grey-brown planks laid across the span, 12 logical px wide and 60 to 64 logical px long, so it spans a 22 px creek plus both stony banks; thin wooden handrails and posts on both sides; short low stone or log abutments of 4 to 6 px at both ends. Soft upper-left light; a soft translucent dark shadow of the deck 3 to 4 px toward the lower right, kept as semi-transparent pixels.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall); the previous try came back portrait 1186x1326 with the deck touching the top and bottom edges: do not do that. In canvas pixels (14.3 canvas px per logical px): the deck is 172 px wide, the handrails add only 15 px on each side (whole width at most 205 px, 15 percent of the canvas width), centred on x=687; the deck runs from y=143 to y=1002 (75 percent of the canvas height), and the small stone abutments add at most 70 px at each end, so at least 70 px of empty alpha-zero canvas stays above and below the bridge. The deck shadow is a soft dark translucent band 45 px wide along its right side and lower end. Everything outside the bridge and its soft shadow is alpha exactly zero.
Avoid: water, banks, path, grass, arches, perspective sides, rope bridge, stone bridge, very short or square decks, railings thicker than 1.5 px, text, outline, checkerboard.
Attached references: image 1 is a crop of the reference painting with two timber footbridges over a blue creek; use it for the plank colour, the handrail weight and the painted finish only; copy none of its water, banks or layout.
```

### Look 3 B2 candidate (2026-10-23)

`bridge-plank-diagonal-shared-v1.png` (`exec-9c209483`, native 1374×1145 RGBA, unchanged) is the
SE/NW plank footbridge matching B1, painted at 36.8° instead of 29.4° (the first try came out at 45°);
registering it needs a 7.4° rotation, which turns the painted light by as much. Faults otherwise as B1.
Not selected. Attached: RefDesign2 crop r33c7+r37c7 and the B1 candidate. The prompt carried the same
two stray water lines as B1. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: newly authored TIMBER FOOTBRIDGE isolated terrain cutout on true alpha transparency, 6:5 canvas 1374x1145 = logical 96x80 (one hex about 50 m), centre (48,40). Strict nadir deck plan, no tilt. Orientation: SE/NW axis at 29.4 degrees from horizontal through (48,40): the bridge runs diagonally from the upper left to the lower right, deck length 64 logical.
Deck: weathered grey-brown planks laid across the span, 12 logical px wide and 60 to 64 logical px long, so it spans a 22 px creek plus both stony banks; thin wooden handrails and posts on both sides; short low stone or log abutments of 4 to 6 px at both ends. Soft upper-left light; a soft translucent dark shadow of the deck 3 to 4 px toward the lower right, kept as semi-transparent pixels.
LANDSCAPE canvas, wider than tall (1374 wide, 1145 tall). The previous try drew the bridge at 45 degrees, a square diagonal: that is wrong. This bridge is a SHALLOW diagonal, much closer to horizontal than to 45 degrees: its deck centreline goes from (287,348) to (1085,797), so the whole bridge spans 798 px across and only 449 px down, almost twice as wide as tall, a slope of 29.4 degrees through the canvas centre (687,572). The planks lie across the deck, perpendicular to that line. The deck is 172 px wide (narrower than the previous try); the handrails are thin, at most 20 px each. The small stone abutments add at most 70 px at each end. At least 120 px of empty alpha-zero canvas stays at the left and right edges. Everything outside the bridge and its soft shadow is alpha exactly zero.
Avoid: water, banks, path, grass, arches, perspective sides, rope bridge, stone bridge, very short or square decks, a vertical or horizontal bridge, a 45-degree diagonal, railings thicker than 1.5 px, text, outline, checkerboard.
Attached references: image 1 is a crop of the reference painting with two timber footbridges over a blue creek; image 2 is the vertical footbridge of this set: match its planks, handrails, posts, abutments, colour and painted finish exactly, only rotated to the diagonal; copy none of the water, banks or layout of image 1.
```

Built-in imagegen produced both independent true-alpha paintings at native 1374x1145. The previous
tiles supplied axis references and the selected crater supplied fine material detail. A follow-up
prompt corrected the narrow deck proportions; neither orientation is a rotated copy of the other.

The `-export-` masters apply bounded affine registration and transparent padding at 1152x960
(12 source pixels per logical pixel). `bridge-registration-v1.json` records native hashes, dimensions,
reviewed deck landmarks and every matrix. Native sampling density exceeds 16 pixels per logical pixel
on both axes, so no enlargement is needed. Small axis corrections retain the painted upper-left light.

`terrain-stamps -manifest art/pipeline/bridge-standing-v1.json` exports the prepared sources as
`standing/bridge*`, with 2x 192x160 and 4x 384x320 PNGs and anchor `(48,40)`. The v2 mixed manifest's
`bridge_artwork: standing` keeps only the original path approaches beyond the deck ends, removing old
finite timber, rails and earth abutments before density filtering. The default v1 mixed export
remains available with its derived bridge artwork.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- arched or perspective bridges
- masonry viaducts
