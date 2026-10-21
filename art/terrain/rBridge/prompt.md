---
id: terrain/rBridge
title: Road bridge
kind: terrain
status: candidate
runtime_ids: rBridge14, rBridge25, rBridge36, standing/rBridge14, standing/rBridge25, standing/rBridge36
masters: rbridge-14-shared-v1.png, rbridge-25-shared-v1.png, rbridge-36-shared-v1.png, rbridge-14-export-shared-v1.png, rbridge-25-export-shared-v1.png, rbridge-36-export-shared-v1.png
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
