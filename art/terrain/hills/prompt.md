---
id: terrain/hills
title: Hills
kind: terrain
status: candidate
runtime_ids: hills
masters: hills-scrub-summer-v2.png, hills-height-shared-v1.png
---

## Look
Rolling hill country. The reference paints hills as meadow (average about RGB 103,109,39) with darker
scrub clumps, small bushes and only gentle, barely visible relief; rocky rises belong to mountain and
mount. The tile must read as hills through soft modelling and scrub, not through peaks.

## Base prompt
Opaque full-bleed scrub meadow, strict vertical plan view. Warm olive fine grass with many irregular
small dark moss-green bramble bushes, coarse weeds, dry twig flecks and scattered tiny pebbles.
Keep open meadow between scrub clumps and barely perceptible relief. No craters, holes, rings, bowls,
isolated mounds, mountain peaks, cliffs or bare rock faces. The exact native generation prompt follows below.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| hills | hills | Painted scrub meadow; the renderer blends into canonical meadow at grass-facing edges. |

## Edges & joins
The painted material uses the absolute grass lattice and blends through a 7..19 world-pixel distance
gate at grass-facing edges. The final four-world-pixel border exactly matches canonical meadow.

## Post
`art-review -hills art/pipeline/hills-terrain-v2.json` exports painted meadow into context tiles keyed by
neighbours and lattice phase (`context/hills/<key>`). Painted mode applies no height lighting or isolated
bump. Summer only. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open
hex ownership. The v1 manifest retains the older optional height workflow.


### Native v2 generation
The selected v2 master is painted scrub meadow. Height mode remains available through the v1 manifest; the v2 manifest blends the painted material into canonical grass, with no height-derived lighting or isolated bump.

Generated with the built-in imagegen tool. Dimensions below describe the actual output.

#### Painted scrub master (1536×1024 RGB)

```text
Use case: precise-object-edit. Asset type: opaque full-bleed 1536x1024 painted scrub-meadow material for gently rolling hill country. Edit the supplied summer meadow material, preserving its warm olive grass palette, nadir camera and fine blade scale. Add many irregular small clumps of dense low wild scrub, bramble bushes and coarse dark moss-green weeds covering about 25 percent of the canvas, distributed naturally with asymmetric gaps. Scrub is low, not tree crowns: fine leaves, little pale dry twig flecks, short soft lower-right shadows and modest upper-left olive highlights. Leave open meadow between them, tiny grey-beige pebbles and scattered fine straw. Very gentle terrain shape, barely perceptible relief. Finely hand-painted natural aerial map art matching the supplied meadow, no visible tiling. Absolutely no craters, holes, bowls, depressions, rings, isolated mounds, mountain peaks, cliffs, bare rock faces, paths, trees, buildings, water or text. No vignettes or grid.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- peaks, cliffs or bare rock
- directional lighting baked into the height map
