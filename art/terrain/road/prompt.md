---
id: terrain/road
title: Road
kind: terrain
status: candidate
runtime_ids: road124, road125, road13, road135, road14, road145, road15, road236, road24, road25, road26, road346, road35, road356, road36, road46
masters: asphalt.png
---

## Look
Paved country road. In the reference it is a medium-grey asphalt carriageway with a dashed off-white
centre line, thin pale edge lines and sandy gravel shoulders with grass tufts, running in smooth, even
curves (average including verges about RGB 124,118,79).

## Base prompt
Opaque square material swatch of a maintained paved road surface seen straight from above: modest
medium-dark neutral grey asphalt (average near RGB 91,94,93) with subtle fine aggregate, faint irregular
wear and low-contrast mottling. Smooth from a distance, crisp fine grain up close. Fill the canvas
uniformly. Do not draw a road strip, lane lines, edges, curbs, cracks, potholes, gravel or grass: the
renderer adds the road shape, markings and shoulders.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| road124 | road124 | Road T or Y junction to ports 1 NE, 2 SE, 4 SW. |
| road125 | road125 | Road T or Y junction to ports 1 NE, 2 SE, 5 NW. |
| road13 | road13 | Road wide bend between ports 1 NE, 3 S. |
| road135 | road135 | Road even three-way fork to ports 1 NE, 3 S, 5 NW. |
| road14 | road14 | Road straight run between ports 1 NE, 4 SW. |
| road145 | road145 | Road T or Y junction to ports 1 NE, 4 SW, 5 NW. |
| road15 | road15 | Road wide bend between ports 1 NE, 5 NW. |
| road236 | road236 | Road T or Y junction to ports 2 SE, 3 S, 6 N. |
| road24 | road24 | Road wide bend between ports 2 SE, 4 SW. |
| road25 | road25 | Road straight run between ports 2 SE, 5 NW. |
| road26 | road26 | Road wide bend between ports 2 SE, 6 N. |
| road346 | road346 | Road T or Y junction to ports 3 S, 4 SW, 6 N. |
| road35 | road35 | Road wide bend between ports 3 S, 5 NW. |
| road356 | road356 | Road T or Y junction to ports 3 S, 5 NW, 6 N. |
| road36 | road36 | Road straight run between ports 3 S, 6 N. |
| road46 | road46 | Road wide bend between ports 4 SW, 6 N. |

## Edges & joins
The road is 22 world pixels wide with a 2-pixel earth shoulder, centred on the shared port centres and
continuing straight through each port, so it joins the next road tile, bridges and compounds exactly.
Centre dashes are drawn by the renderer along the centre line.

## Post
`art-review -connected art/pipeline/connected-terrain-v1.json` cuts the asphalt to each port set, adds dashed
markings and earth shoulders (earth bank material) over meadow. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- loose gravel or cobble surfaces
- shiny black tarmac
- painted lane lines in the material
