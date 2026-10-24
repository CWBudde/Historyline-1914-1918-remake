---
id: terrain/path
title: Path
kind: terrain
status: candidate
runtime_ids: path125, path13, path135, path136, path14, path15, path24, path25, path26, path346, path35, path356, path36, path46, path125_w, path13_w, path135_w, path136_w, path14_w, path15_w, path24_w, path25_w, path26_w, path346_w, path35_w, path356_w, path36_w, path46_w
masters: path-earth-shared-v2.png, path-earth-shared-v1.png, path-track-winter-v1.png, path-track-summer-v1.png
---

## Look
Narrow earth track. In the reference it is a pale tan dirt track with two faint wheel ruts and a thin
darker edge, meadow growing right up to it.

## Base prompt
Opaque square material canvas, filled edge to edge with warm muted tan compacted fine earth: tiny sandy
grains, very occasional minute rounded pebbles, quiet low-contrast variation; dusty ochre-greige,
lighter and warmer than mud, distinct from grey asphalt. Do not draw a path silhouette, verges, ruts,
grass or objects: the renderer cuts the track.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| path125 | path125 | Path T or Y junction to ports 1 NE, 2 SE, 5 NW. |
| path13 | path13 | Path wide bend between ports 1 NE, 3 S. |
| path135 | path135 | Path even three-way fork to ports 1 NE, 3 S, 5 NW. |
| path136 | path136 | Path T or Y junction to ports 1 NE, 3 S, 6 N. |
| path14 | path14 | Path straight run between ports 1 NE, 4 SW. |
| path15 | path15 | Path wide bend between ports 1 NE, 5 NW. |
| path24 | path24 | Path wide bend between ports 2 SE, 4 SW. |
| path25 | path25 | Path straight run between ports 2 SE, 5 NW. |
| path26 | path26 | Path wide bend between ports 2 SE, 6 N. |
| path346 | path346 | Path T or Y junction to ports 3 S, 4 SW, 6 N. |
| path35 | path35 | Path wide bend between ports 3 S, 5 NW. |
| path356 | path356 | Path T or Y junction to ports 3 S, 5 NW, 6 N. |
| path36 | path36 | Path straight run between ports 3 S, 6 N. |
| path46 | path46 | Path wide bend between ports 4 SW, 6 N. |
| winter | path125_w, path13_w, path135_w, path136_w, path14_w, path15_w, path24_w, path25_w, path26_w, path346_w, path35_w, path356_w, path36_w, path46_w | The same courses in winter: a grey trodden track of packed snow and slush with two narrow cart ruts, its verge under snow. |

## Edges & joins
The path is 7 world pixels wide with a 0.75-pixel earth border, centred on the shared port centres and
continuing through each port. The winter forms keep the geometry; the renderer draws two cart ruts 1.6
world pixels either side of the centre line.

## Post

### Look 4 R2 path strip candidate (2026-10-23)

`path-track-summer-v1.png` (`exec-6bd6cf89`, native 1536×1024 RGB, unchanged) is the R2 cart track:
two broken bare-earth ruts with soft grooves, grass creeping in and frayed edges; rut colour 150,123,68
(target 160,124,78), meadow on target, no visible seam. Faults: the ruts are 98 px apart (spec 74), about
1.3× the requested scale in all four tries, and the track wobbles gently. Not selected. Attached: the
unused earlier painted track `exec-9ca6708e` (material only) and RefDesign2 crop r22c3. The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Opaque 1536x1024 canvas, strict nadir, horizontally tileable. One narrow farm cart track runs left to right centred on y=512; canvas = 144x96 world px. The track is 117 px wide (11 world px): two bare-earth wheel ruts, each 21 px wide, centred 37 px above and below the centreline, compacted warm tan-brown earth (RGB near 160,124,78) with a darker groove on their shadow side and a sunlit lip on the other. Between them a 32 px median of short worn grass and plantain. The outer edges fray into the meadow over 30-40 px with tufts and bare patches, never a straight border. Warm olive meadow fills the rest. Upper-left diffuse light.
Avoid: a uniform stripe, straight edges, gravel or asphalt, footprints, puddles with sky reflections, text.
As fractions of the 1024 px canvas height: the whole track is only 11% of the height (rows 454 to 570), the two ruts are centred on rows 475 and 549, and everything above row 420 and below row 605 is plain meadow, about 80% of the canvas. The track is a thin band, not the main subject.
Corrections from the measured previous tries, which painted the track too big and too high: they spread it over 165-175 px (17% of the height) with its centre at row 499 and the ruts 100-108 px apart; paint it narrower, about 117 px (11%), centred on row 512, with the ruts' centres only 74 px apart (rows 475 and 549) and each rut a narrow worn strip about 21 px wide. The grass median between them is a little wider than one rut. The last try made the ruts clean uniform stripes with straight edges; break them up instead: ruts of varying width, grass creeping in from the sides, bare patches and tufts fraying the outer edges irregularly. The left and right edges of the image must match exactly so it tiles without a seam.
Attached references: image 1 (an earlier painted track tile) for the earth material and painted finish only; image 2 (a reference map crop) for the colour of a narrow track in meadow; copy none of their layout or curves.
```

### Look 4 R2 winter edit candidate (2026-10-23)

`path-track-winter-v1.png` (`exec-2a95e46c`, native 1536×1024 RGB, unchanged) is the winter edit of
the R2 candidate: a calm neutral snowfield (196,193,192) with sparse stalks and dark mud ruts
(104,93,84). Faults: the upper rut sits 11 px higher than in summer, a faint seam shows when tiled, and a
few stalks are upright sprigs. Not selected. Attached: the R2 candidate. The edit prompt was:

```text
Same geometry; the ruts become dark trodden snow-and-mud tracks, the median and meadow are under snow with stalks showing, and the edges stay soft. Keep it tileable.
Corrections from the measured previous edit: its left and right edges did not match, leaving a visible vertical seam when tiled; make the leftmost and rightmost columns continue into each other seamlessly. Its meadow was crowded with large dark frosted clumps; show only sparse, small dry stalks and a few low tufts poking through smooth settled snow, so the snowfield stays calm. Its snow was faintly blue (average RGB 188,192,200); keep it a neutral soft white-grey under overcast light, about RGB 205,203,200.
Attached references: image 1 is the summer cart-track strip to edit; copy none of its colours, keep all of its layout.
```

`art-review -connected art/pipeline/connected-paths-v2.json` cuts the earth material to each port set over
meadow. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

The selected built-in imagegen master is native 1254×1254 RGB, with irregular fine grains and subdued
wear. The former earth source remains a retained draft. The path footprint and ports are unchanged.

`connected-paths-winter-v1.json` renders the `_w` forms from the winter road material
`art/terrain/road/road-snow-winter-v1.png` with `markings: ruts` and `verge: snow`: two narrow cart ruts,
and the earth verge blended 0.9 toward the winter ground (`grass-winter-v1.png`). The winter maps draw them
in place of the summer paths. The tagged generator `TestGenerateWinterPathsAndRails` (`campaigngfx`)
installs them with the winter rails, crossings and footbridge tiles; `TestWinterPathsAndRailsMeetCriterion`
measures them over the winter ground: verge ΔE 4.3–5.2 from the ground beside it, no summer green,
surface L\* 49–51 within ΔE 1.1–3.5 of the `road36_w` surface. The unpromoted strip `path-track-winter-v1.png`
is not used.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- wheel ruts or paving in the material
