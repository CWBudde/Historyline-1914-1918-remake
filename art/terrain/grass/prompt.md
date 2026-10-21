---
id: terrain/grass
title: Meadow
kind: terrain
status: candidate
runtime_ids: grass, grass_w, grass-detail-00, grass-detail-01, grass-detail-02, grass-detail-03, grass-detail-04, grass-detail-05, grass-detail-06, grass-detail-07, grass-detail-08, grass-detail-09, grass-detail-10, grass-detail-11, grass-detail-12, grass-detail-13, grass-detail-14, grass-detail-15
masters: grass-meadow-summer-v2.png, grass-meadow-winter-v2.png, grass-flowers-painting-summer-v1.png, grass-flowers-material-summer-v1.png, grass-worn-painting-summer-v1.png, grass-worn-material-summer-v1.png, grass-summer-v1.png, grass-winter-v1.png, grass-meadow-summer-v1.png
---

## Look
Open summer meadow, the ground under every other tile and the most common hex on the map. In the
reference it is a warm olive to yellow-green grass (average about RGB 105,113,38) built from fine short
blades, darker moss-green clumps, small dry straw strokes and sparse yellow and white specks. It reads as
an even, quiet field at play size and shows crisp blades only up close.

## Base prompt
Opaque grass terrain material for a top-down strategy map, 1536x1024, filled edge to edge. Paint only
an open European meadow surface: short uneven grass with individually readable blades, darker moss-green
tufts, little dry straw scratches, occasional tiny bare brown soil flecks, rare small grey-brown pebbles
and sparse pale seed heads and yellow wildflower specks. Colour is a warm subdued olive and yellow-green
averaging about RGB 105,113,38, with deep moss-green shadows in the tufts; not lime, not saturated green,
not a brown straw field. Irregular natural patch variation across the whole canvas, fairly even overall:
no large dark blobs, bright patches, mounds or repeating cushion shapes. Tiny shadows under raised tufts
from the upper left. No trees, bushes, roads, water, buildings, craters or trenches.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| grass | grass | The base meadow material above. |
| grass_w | grass_w | Winter: pale cool ivory settled snow with faint blue-grey hollows and fine wind-scoured texture; sparse faded straw tips under 3 percent; no footprints, tracks, rocks or drifts. |
| base phases | grass-detail-00, grass-detail-05, grass-detail-10, grass-detail-15 | No new art: the base meadow sampled at four lattice phases. |
| flowers | grass-detail-01, grass-detail-03, grass-detail-06, grass-detail-08, grass-detail-11, grass-detail-13 | Edit of the base meadow: add about 40 tiny yellow and white wildflower specks, a few pale-grey gravel clusters and five small dark weed tufts; under 3 percent of the canvas, average colour unchanged. |
| worn | grass-detail-02, grass-detail-04, grass-detail-07, grass-detail-09, grass-detail-12, grass-detail-14 | Edit of the base meadow: about 15 small dry, thinner patches with earthy flecks and short dry blades plus four dark weed tufts; at least 95 percent unchanged, no paths, rings or bare islands. |

## Edges & joins
The meadow is a periodic material sampled on the world lattice, so neighbouring grass tiles share one
continuous field. Detail variants keep the canonical meadow unchanged within 8 world pixels of every
edge (smoothstep to the interior over 8..18), so any variant meets any other and every overlay tile.
Seams to forest, ground and hills are handled by those families' contexts, which use this meadow as
their base.

## Post
`art-review -ground`, `-forest`, `-hills`, `-shore`, `-fog` and the connected and compound manifests all
sample `grass-summer-v1.png` / `grass-winter-v1.png` as their base. The 16 detail variants are made by the `grassgfx`
generator (65 percent detail master, 35 percent phased base, edge guard). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.


### Native v2 generation
The native v2 summer and winter meadow masters are comparison candidates. Runtime `grass-summer-v1.png`, `grass-winter-v1.png` and the existing detail materials remain selected so neighbouring terrain keeps its established material and palette.

Generated with the built-in imagegen tool. Dimensions below describe the actual output.

#### Summer comparison candidate (1536×1024 RGB)

```text
Use case: stylized-concept. Asset type: opaque seamless terrain material master, landscape 1536x1024. Create newly authored very detailed hand-painted digital illustration for a World War I rural European tactical map. Strict top-down orthographic nadir view. Fill the entire image edge to edge with open summer meadow, no objects or terrain features: rich short grass formed of tiny individually painted fine blades and small natural irregular moss-green tufts, olive yellow-green sunlit tips, tiny straw scratches, rare small pale wildflower/seed-head dots, pinhead brown soil flecks and very sparse tiny grey pebbles. Fine believable material detail like the grass fringe in the supplied crater style reference, and the reference aerial map. Average warm olive around RGB 100,109,36; restrained natural palette with enough fine local contrast to read at 4x game zoom, but calm overall values for units. Soft diffuse light from upper left, tiny short soft shadows. Avoid uniform noise, repeated cushions, round large tufts, dark big blobs, beige bare islands, paths, tracks, trees, bushes, craters, holes, mountains, water, buildings, text, outlines, hex geometry, frame, vignette, pixels, vector art, 3D or photographic rendering. This is a material sheet only, not a map composition.
```

#### Winter comparison candidate (1536×1024 RGB)

```text
Use case: lighting-weather. Asset type: opaque 1536x1024 winter meadow material. Edit target is supplied summer meadow. Preserve exactly its nadir camera, texture scale, canvas and spatial arrangement, changing only season. Cover the meadow with settled cool ivory-white fine-grained snow, minute wind-scoured surface texture, subtle blue-grey creases corresponding to existing shallow grass texture. A few straw tips poke through in under 3 percent of surface. Retain the detailed hand-painted strategy-map material style. Calm fairly uniform overall values, no large drifts, no footprints or tracks, no new rocks/trees/objects, no crater holes, no vignette or text. Full opaque bleed, no borders.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- neon or saturated greens
- large flowers, bushes or boulders in the base material
