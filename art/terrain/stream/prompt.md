---
id: terrain/stream
title: Stream
kind: terrain
status: candidate
runtime_ids: stream13, stream135, stream14, stream15, stream24, stream246, stream25, stream26, stream35, stream36, stream46
masters: bank.png
---

## Look
Winding creek. In the reference it is clear blue-teal water with a bright network of white ripples,
lined by banks of rounded pale stones and pebbles, with grass and dark green bushes right beside it.

## Base prompt
Opaque square material swatch of a stony stream bank seen from above: tightly packed small and
medium rounded pale stones and pebbles in warm grey, cream and light sand, a few larger flat stones,
olive-grey dirt and moss in the gaps, darker wet crevices. Even scale and brightness to all edges; no
water, grass, boulders, paving or tessellation.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| stream13 | stream13 | Stream wide bend between ports 1 NE, 3 S. |
| stream135 | stream135 | Stream even three-way fork to ports 1 NE, 3 S, 5 NW. |
| stream14 | stream14 | Stream straight run between ports 1 NE, 4 SW. |
| stream15 | stream15 | Stream wide bend between ports 1 NE, 5 NW. |
| stream24 | stream24 | Stream wide bend between ports 2 SE, 4 SW. |
| stream246 | stream246 | Stream even three-way fork to ports 2 SE, 4 SW, 6 N. |
| stream25 | stream25 | Stream straight run between ports 2 SE, 5 NW. |
| stream26 | stream26 | Stream wide bend between ports 2 SE, 6 N. |
| stream35 | stream35 | Stream wide bend between ports 3 S, 5 NW. |
| stream36 | stream36 | Stream straight run between ports 3 S, 6 N. |
| stream46 | stream46 | Stream wide bend between ports 4 SW, 6 N. |

## Edges & joins
The water channel is 18 world pixels wide with a 3.5-pixel stony bank, centred on the shared port
centres and continuing through each port, so creeks join each other, bridges and the lake shore.

## Post
The bank material is an unchanged crop of the lakeside bank master (`riverbankgfx` generator).
`art-review -connected art/pipeline/connected-terrain-v1.json` draws the water (shared water material) and banks
along each port set over meadow. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- muddy brown water
- regular cobble patterns
