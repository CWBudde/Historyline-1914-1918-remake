---
id: terrain/rBridge
title: Road bridge
kind: terrain
status: candidate
runtime_ids: rBridge14, rBridge25, rBridge36
masters: 
---

## Look
Road bridge over a stream. In the reference the asphalt road with its dashed centre line crosses on a
concrete deck with low pale parapets, the creek passing underneath between stony banks.

## Base prompt
Transparent cut-out of a short concrete road bridge seen straight from above: grey asphalt carriageway
with off-white edge lines, low pale concrete parapets on both sides, small abutments, soft shadow to the
lower right. The deck runs straight across the canvas; nothing else.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| rBridge14 | rBridge14 | Road bridge straight run between ports 1 NE, 4 SW, over the stream. |
| rBridge25 | rBridge25 | Road bridge straight run between ports 2 SE, 5 NW, over the stream. |
| rBridge36 | rBridge36 | Road bridge straight run between ports 3 S, 6 N, over the stream. |

## Edges & joins
Carriageway width and centre match the road family exactly, so the road continues through both ports;
the stream continues under the bridge through its own ports.

## Post
Derived only: `art-review -mixed art/pipeline/connected-mixed-v1.json` builds road bridges from the road and
stream materials. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- arches, towers or perspective
