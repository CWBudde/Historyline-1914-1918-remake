---
id: terrain/ground
title: Ground
kind: terrain
status: candidate
runtime_ids: mud, rough_w, shelled_w, stone, stoney_w
masters: earth-bank.png, mud.png, rough_w.png, shelled_w.png, stone.png, stoney_w.png
---

## Look
Rough ground variants of open land. In the reference, mud is a round shell crater with a raised,
cracked tan earth rim on churned stony ground, and stone is meadow strewn with rounded grey and
grey-brown boulders, pebble clusters and small bright shrubs. Winter variants are their snow-covered
counterparts. All stay meadow-coloured at the edges so they blend into surrounding grass.

## Base prompt
Opaque top-down terrain illustration, landscape 6:5 canvas filled edge to edge, strict nadir view.
The ground between features is the same warm olive meadow as the grass terrain (average about RGB
105,113,38) and continues to every edge. Features are painted at aerial map scale with fine natural
detail: rounded grey-beige boulders with soft lower-right shadows, small pebble clusters, tan and umber
earth, small bright green shrubs. Readable as distinct terrain at 192x160, finely detailed at full size.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| mud | mud | One centred round shell crater filling about 70 percent of the canvas: a raised rough rim of tan-brown clods and cracked earth, a dark scalloped inner cut bank, a flatter grassy and earthy floor; scattered stones around it; meadow outside. |
| stone | stone | Meadow strewn with rounded grey and grey-brown boulders of mixed size, small pebble clusters and a few bright green shrubs; stone covers about a quarter of the ground; no slabs, cliffs or paving. |
| rough_w | rough_w | Winter rough ground: about 40 percent snow, 35 percent frozen dark grey-brown clods, 20 percent dead straw tufts and low brush, a few stone fragments. |
| stoney_w | stoney_w | Winter stony ground: low grey stone slabs and rubble aprons interleaved with settled ivory snow, roughly half and half; snow fills hollows. |
| shelled_w | shelled_w | Winter shell-damaged ground: a broad, broken, partly collapsed crater basin with overlapping partial impacts, churned dark earth reaching the edges and snow in irregular patches over about a third; no neat rings. |
| earth bank | — | Shared border material of road, path, rail and trench. Close material swatch, square and opaque: rough dug tan-brown earth with small clods, shale crumbs, tiny roots and sparse grey gravel. |

## Edges & joins
Ground tiles are chosen by context: the renderer blends the material towards the seasonal meadow on
every side whose neighbour is not the same material, so a crater or stone patch never ends in a hard
edge. Features stay inside the hex; only meadow and fine scatter reach the border.

## Post
`art-review -ground art/pipeline/ground-terrain-v1.json` projects the five materials into context tiles
keyed by season, centre and the six neighbours (`context/ground/<key>`); the crater is placed as a
centred feature. Winter variants come from their own masters. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- black crevices or dramatic cast shadows
- flagstones, cobbles or gravel-road looks
- more than one crater per tile
