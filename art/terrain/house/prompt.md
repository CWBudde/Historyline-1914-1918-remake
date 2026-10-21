---
id: terrain/house
title: House
kind: terrain
status: candidate
runtime_ids: house, house_w
masters: hamlet.png, hamlet_w.png
---

## Look
A single rural building in open meadow. In the reference, houses are square buildings with a grey-brown
hipped roof of weathered shingles, each with a smaller shed beside it and a bush or small tree nearby,
about a quarter of a hex across.

## Base prompt
Transparent single-building cut-out, square RGBA canvas, strict nadir view: one compact rural house
with a square footprint and a weathered grey-brown shingle hipped roof (four sloped planes meeting at a
short ridge), a small chimney, a thin shadow-side eave, a smaller lean-to shed beside it, a few threshold
stones and a narrow grass fringe at the base. Centred, about half the canvas wide, generous transparent
margin. Soft grounded shadow to the lower right.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| house | house | Summer as above. |
| house_w | house_w | Same geometry; a thin uneven layer of settled snow on the roof planes with shingles showing along eaves and ridge; subdued snowy fringe. |

## Edges & joins
The building sits inside the hex with meadow around it; only the shared seasonal meadow reaches the
hex edges.

## Post
`art-review -compound art/pipeline/compound-hamlet-v1.json` places the cut-out on the seasonal meadow at a
fixed inset. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- orange clay or blue roofs
- gardens, fences or people
