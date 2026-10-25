---
id: terrain/hills
title: Hills
kind: terrain
status: approved
runtime_ids: hills
masters: hills-scrub-summer-v2.png, hills-height-shared-v1.png, hills-hummock-summer-v1.png
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

### Look 2 G3 hummock overlay candidate (2026-10-23)

`hills-hummock-summer-v1.png` (`exec-ad797ade`, native 1536×1024 RGB, unchanged) is the G3 hummock
and scrub overlay. It reads as hills at 1×: lumpy ground of lit hummock crescents, 6.3 L\* darker than
the meadow. Faults: crescents 40–70 world px lean into loose NE–SW rows, band ratio about 1.0 (target
≥ 1.4), and only about 4 scrub clumps per hex (target 6–10). Not selected. Attached:
`hills-scrub-summer-v2.png`, the G1 candidate result at 2 px per world px, and RefDesign2 crop r21c7.
The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material, 1536x1024, continuation of the supplied meadow: same palette, grass texture and scale (one 50 m hex is about 250 px across, so the canvas is about six hexes wide and five tall). Rough hill pasture seen strictly from above, clearly bumpy all over: the ground is a dense carpet of small rounded grass tussock hummocks, each only 30 to 50 px across (about 1/40 of the canvas width), packed close with narrow darker hollows between them, varying in size and outline and placed irregularly, never in rows, lines or a grid, never joined into ridges. Each hummock is clearly modelled by its own small form shading: its upper-left half lighter olive with dry straw tips, its lower-right half and the hollow beside it darker olive shade, a clear light-dark step inside every hummock; but there is no larger light gradient, sheen or band across the ground: at the scale of 100 px and more the tone has no light direction and is even. Dense, SMALL low scrub, bramble and gorse clumps scattered in the hollows, each clump only 8 to 14 px across (about 1/150 of the canvas width, never wider than 20 px; most clumps are single small bushes, not groups), very dark moss green (about RGB 38,50,16 in their shaded parts) with fine leaves and a short lower-right shadow at most 4 px long; about 12 clumps in every 250x200 px area, about 380 clumps over the whole canvas, spaced irregularly, none touching another. A few tiny grey pebbles. Overall slightly darker than the meadow: average colour close to RGB 88,96,34 (green channel clearly above red, blue near 34), not browner, not more yellow and not darker than that; clearly rougher, but still grass, not rock. Tileable left-right and top-bottom, even density in every direction.
Avoid: peaks, cliffs, bare rock faces, craters, rings, bowls, trees, bushes wider than 20 px, merged scrub groups, paths, contour-like stripes, ridges longer than 120 px, diagonal bands, aligned patches, parallel ridges, clumps lined up in diagonal rows, lit streaks, large pale straw patches, smooth flat lawn, regular polka-dot spacing, large dark blobs, neon green, perspective, horizon, vignette, text, grid, photograph, 3D render.

Attached images, in order: image 1 is the current hills scrub material (hills-scrub-summer-v2), for subject and scrub style only (its clumps are about three times too large; paint them at the sizes given above); image 2 is the supplied meadow (the G1 result), for palette, grass texture and scale; image 3 is a small crop of the target reference map at a hills cell (r21c7), for how hills read on the map (ignore the building and trees at its edges).
```

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
