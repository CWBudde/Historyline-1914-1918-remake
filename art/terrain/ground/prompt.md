---
id: terrain/ground
title: Ground
kind: terrain
status: candidate
runtime_ids: mud, rough_w, shelled_w, stone, stoney_w, stamp/crater/4_-1.1_1.1_0.2/1, stamp/boulders/1/1, stamp/boulders/4_2.0_1.1_0.2/1, stamp/boulders/5_-1.1_0.2_-1.3_1.3/1
masters: crater-stamp.png, boulders-single.png, boulders-four.png, boulders-five.png, earth-bank.png, mud.png, rough_w.png, shelled_w.png, stone.png, stoney_w.png
---

## Look
Rough ground variants of open land. In the reference, mud is a round shell crater with a raised,
cracked tan earth rim on churned stony ground, and stone is meadow strewn with rounded grey and
grey-brown boulders, pebble clusters and small bright shrubs. Winter variants are their snow-covered
counterparts. All stay meadow-coloured at the edges so they blend into surrounding grass.
The summer four-hex crater stamp is one round basin with a hand-painted cracked-earth rim and a low
olive floor, with transparent meadow edges.
Summer boulder stamps scatter rounded warm grey and beige stones, pebble clusters, sparse olive grass
tufts and moss over transparent ground; the shared meadow shows between the stones.

## Base prompt
Opaque top-down terrain illustration, landscape 6:5 canvas filled edge to edge, strict nadir view.
The ground between features is the same warm olive meadow as the grass terrain (average about RGB
105,113,38) and continues to every edge. Features are painted at aerial map scale with fine natural
detail: rounded grey-beige boulders with soft lower-right shadows, small pebble clusters, tan and umber
earth, small bright green shrubs. Readable as distinct terrain at 192x160, finely detailed at full size.

For the crater stamp, replace the opaque tile instructions with: one transparent landscape terrain
cut-out, strict nadir view, of a single circular shell crater with a raised scalloped rim of cracked tan
and umber earth, a dark inner cut bank, radial fissures and a shallow olive grassy floor. Warm grey
stones and clods form a sparse apron with irregular transparent margins. Continuous hand-painted
material detail, diffuse upper-left light and short soft lower-right shadows; no road, text, grid,
hex outline, people, checkerboard, photographic texture or 3D rendering.

For the boulder stamps, replace the opaque tile instructions with: one transparent landscape terrain
cut-out, strict nadir view, of scattered rounded warm grey and beige boulders of varied sizes and small
pebble clusters, covering about 18–20 percent of the scene. Sparse olive grass tufts and moss soften
the stones; real alpha transparency remains between them and around the irregular outer fringe.
Continuous hand-painted natural material detail, diffuse upper-left light and short soft lower-right
shadows. Keep a loose uneven scatter across the full scene, with no flat grey plate, paving, text,
grid, hex outline, people, checkerboard, photographic texture or 3D rendering.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| mud | mud | One centred round shell crater filling about 70 percent of the canvas: a raised rough rim of tan-brown clods and cracked earth, a dark scalloped inner cut bank, a flatter grassy and earthy floor; scattered stones around it; meadow outside. |
| stone | stone | Meadow strewn with rounded grey and grey-brown boulders of mixed size, small pebble clusters and a few bright green shrubs; stone covers about a quarter of the ground; no slabs, cliffs or paving. |
| rough_w | rough_w | Winter rough ground: about 40 percent snow, 35 percent frozen dark grey-brown clods, 20 percent dead straw tufts and low brush, a few stone fragments. |
| stoney_w | stoney_w | Winter stony ground: low grey stone slabs and rubble aprons interleaved with settled ivory snow, roughly half and half; snow fills hollows. |
| shelled_w | shelled_w | Winter shell-damaged ground: a broad, broken, partly collapsed crater basin with overlapping partial impacts, churned dark earth reaching the edges and snow in irregular patches over about a third; no neat rings. |
| earth bank | — | Shared border material of road, path, rail and trench. Close material swatch, square and opaque: rough dug tan-brown earth with small clods, shale crumbs, tiny roots and sparse grey gravel. |
| crater stamp | stamp/crater/4_-1.1_1.1_0.2/1 | One continuous round crater across a four-hex summer cluster, with a cracked earth rim, olive floor and transparent outer fringe. |
| single boulder stamp | stamp/boulders/1/1 | The transparent boulder scatter above in a 6:5 scene for one summer hex. |
| four-cell boulder stamp | stamp/boulders/4_2.0_1.1_0.2/1 | One continuous 3:2 summer scatter for the four-cell cluster; no cell boundaries or repeated tile motifs. |
| five-cell boulder stamp | stamp/boulders/5_-1.1_0.2_-1.3_1.3/1 | One continuous 6:5 summer scatter for the five-cell cluster; no cell boundaries or repeated tile motifs. |

## Edges & joins
Ground tiles are chosen by context: the renderer blends the material towards the seasonal meadow on
every side whose neighbour is not the same material, so a crater or stone patch never ends in a hard
edge. Context-tile features stay inside the hex; only meadow and fine scatter reach the border. The
summer crater stamp spans its four-cell cluster as one continuous scene, with transparent edges over
the shared meadow. Both summer four-hex clusters use it; the southern crater pair remains open.
Summer boulder stamps replace per-hex stone art across their complete clusters, preserving transparent
gaps over the shared meadow; winter stone art retains its context tiles.

## Post
`art-review -ground art/pipeline/ground-terrain-v1.json` projects the five materials into context tiles
keyed by season, centre and the six neighbours (`context/ground/<key>`); the crater is placed as a
centred feature. Winter variants come from their own masters. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

`tools/terrain-stamps` exports `crater-stamp.png` (native 1536x1024) through
`art/pipeline/feature-stamps-v1.json`. The full source preserves its aspect at 180x120 logical, anchor
`[90,60]`, centred on the cluster's northernmost hex: 2x 360x240 and 4x 720x480 under
`assets/terrain/ground/`. Still a candidate.
Generated with built-in imagegen; the local PULSE `RefDesign2.png` informed the style, without editing
the reference painting.

The same stamp manifest exports three summer boulder ground candidates from their full sources,
preserving aspect and alpha. `boulders-single.png` (native 1374x1145) uses 96x80 logical, anchor
`[48,40]`, exporting 2x 192x160 and 4x 384x320. `boulders-four.png` (native 1536x1024) uses 240x160
logical, anchor `[49,40]`, exporting 2x 480x320 and 4x 960x640. `boulders-five.png` (native 1374x1145)
uses 240x200 logical, anchor `[120,40]`, exporting 2x 480x400 and 4x 960x800. All exports live under
`assets/terrain/ground/`; anchors register each scene on its cluster's first cell. These masters were
generated with built-in imagegen using the local PULSE reference for style only and remain candidates.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- black crevices or dramatic cast shadows
- flagstones, cobbles or gravel-road looks
- more than one crater per tile
