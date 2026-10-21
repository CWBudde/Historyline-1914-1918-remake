---
id: terrain/mountain
title: Rocky knoll
kind: terrain
status: candidate
runtime_ids: mountain, mountain_w
masters: 
---

## Look
A single small rocky knoll. In the reference it is a compact outcrop of cracked grey-beige rock with
grassy slopes, dark scrub and a few yellow flowers around it (average about RGB 80,87,38).

## Base prompt
Transparent terrain cut-out, square, strict nadir view: one low, compact, asymmetric knot of weathered
warm grey-beige rock with four or five broad cracked slabs, olive moss and small dark-green scrub in the
crevices, a short grassy apron fading out. About 60 percent of the canvas, generous transparent margin,
soft shadow to the lower right.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| mountain | mountain | Summer as above. |
| mountain_w | mountain_w | Same outcrop with thin settled snow on flat stone and in sheltered cracks; scrub muted and lightly snowy. |

## Edges & joins
The knoll sits inside the hex on the seasonal meadow; only meadow reaches the hex edges.

## Post
Derived from the mount masters (`art/terrain/mount/`): `art-review -compound
art/pipeline/compound-mountain-v1.json` places the single-cell scene for both seasons. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tall peaks or radial ridges
- grass islands or dirt bases
