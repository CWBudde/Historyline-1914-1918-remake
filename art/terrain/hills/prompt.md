---
id: terrain/hills
title: Hills
kind: terrain
status: candidate
runtime_ids: hills
masters: height.png
---

## Look
Rolling hill country. The reference paints hills as meadow (average about RGB 103,109,39) with darker
scrub clumps, small bushes and only gentle, barely visible relief; rocky rises belong to mountain and
mount. The tile must read as hills through soft modelling and scrub, not through peaks.

## Base prompt
Opaque square grayscale height-control image for continuous rolling terrain, not a lit artwork.
Strict vertical plan view. Values encode elevation only: dark low valleys, mid-grey slopes, pale high
ridges. Two broad asymmetric, gently winding, interconnected ridges separated by a shallow saddle; the
main ridge sweeps from upper left through the upper middle towards lower right, a smaller branch flows
to the lower left. Exceptionally smooth, broad, organic gradients, most values between 20 and 85
percent grey. No lighting, shadows, contour bands, noise, texture, objects or isolated cones.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| hills | hills | The height map above; the meadow, scrub and shading come from the renderer. |

## Edges & joins
The height field is sampled on the absolute map lattice (phase 568x320), so relief continues across
neighbouring hills hexes and fades to flat meadow at edges shared with other terrain.

## Post
`art-review -hills art/pipeline/hills-terrain-v1.json` lights the meadow with the height map into context
tiles keyed by neighbours and lattice phase (`context/hills/<key>`). Summer only. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- peaks, cliffs or bare rock
- directional lighting baked into the height map
