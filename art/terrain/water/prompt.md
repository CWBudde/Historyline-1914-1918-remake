---
id: terrain/water
title: Water
kind: terrain
status: candidate
runtime_ids: water, waterStones
masters: water.png, waterStones.png
---

## Look
Open lake water. In the reference it is a clear, fairly saturated blue-teal (average about RGB
32,89,107) with darker deep tones and small scattered white glints; waterStones shows a cluster of
three rounded tan-grey rocks rising from the water with pale ripple rings.

## Base prompt
Opaque square material of open lake water seen straight from above: clear blue-teal with darker deep
blue troughs and short broken pale ripple marks and small white glints scattered unevenly; average
close to RGB 32,89,107. Flat orthographic surface, diffuse light, even detail to all four edges for
periodic sampling. No shore, stones, reflections, gradients or foam.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| water | water | The water material above. |
| waterStones | waterStones | Transparent 6:5 cut-out: three compact groups of low rounded tan-grey rocks with flat cracked tops and a few small companions, clustered in the central 55 percent; only a very narrow translucent ripple fringe at each rock; everything else alpha zero. |

## Edges & joins
The water material is periodic on the world lattice, so lake hexes form one continuous surface; shore
and stream tiles use the same material, so water matches across all of them.

## Post
`art-review -shore` samples the water material for `water`; `art-review -compound
art/pipeline/compound-water-stones-v1.json` places the rock cut-out over it for `waterStones`. Summer only.
Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tropical turquoise or neon cyan
- white foam rings
- boats or bridges
