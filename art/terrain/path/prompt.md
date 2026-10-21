---
id: terrain/path
title: Path
kind: terrain
status: candidate
runtime_ids: path125, path13, path135, path136, path14, path15, path24, path25, path26, path346, path35, path356, path36, path46
masters: earth.png
---

## Look
Narrow earth track. In the reference it is a pale tan dirt track with two faint wheel ruts and a thin
darker edge, meadow growing right up to it.

## Base prompt
Opaque 6:5 material canvas, filled edge to edge with warm muted tan compacted fine earth: tiny sandy
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

## Edges & joins
The path is 7 world pixels wide with a 0.75-pixel earth border, centred on the shared port centres and
continuing through each port.

## Post
`art-review -connected art/pipeline/connected-paths-v1.json` cuts the earth material to each port set over
meadow. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- wheel ruts or paving in the material
