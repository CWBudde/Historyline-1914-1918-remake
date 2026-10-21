---
id: terrain/trench
title: Trench
kind: terrain
status: candidate
runtime_ids: trench0, trench1, trench12, trench124, trench125, trench14, trench15, trench24, trench25, trench4, trench5, trench5a
masters: floor.png, timber.png
---

## Look
Narrow front-line trench dug into meadow: a dark umber floor, weathered timber revetment and duckboards,
and tan spoil banks on both sides, following the hex ports in short straight runs.

## Base prompt
Two opaque square material swatches seen from above. Floor: close continuous dark compacted umber
soil and clay with tiny angular pebbles, a few root splinters and straw, darker damp seams, no water
shine. Timber: dry, worn brown-grey oak boards aligned vertically, subdued taupe grain, split fibres,
occasional nail heads, dark seams, muddy pale splinter edges. Neither swatch shows a trench shape,
walls, sandbags or grass.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| trench0 | trench0 | Isolated dug-in position with no exits. |
| trench1 | trench1 | Trench dead end opening to port 1 NE. |
| trench12 | trench12 | Trench sharp bend between ports 1 NE, 2 SE. |
| trench124 | trench124 | Trench T or Y junction to ports 1 NE, 2 SE, 4 SW. |
| trench125 | trench125 | Trench T or Y junction to ports 1 NE, 2 SE, 5 NW. |
| trench14 | trench14 | Trench straight run between ports 1 NE, 4 SW. |
| trench15 | trench15 | Trench wide bend between ports 1 NE, 5 NW. |
| trench24 | trench24 | Trench wide bend between ports 2 SE, 4 SW. |
| trench25 | trench25 | Trench straight run between ports 2 SE, 5 NW. |
| trench4 | trench4 | Trench dead end opening to port 4 SW. |
| trench5 | trench5 | Trench dead end opening to port 5 NW. |
| trench5a | trench5a | Trench dead end opening to port 5 NW. Second variant of trench5. |

## Edges & joins
The trench is 6 world pixels wide with a 3-pixel earth bank, centred on the shared port centres and
continuing through each port, so trench lines join across hexes.

## Post
`art-review -connected art/pipeline/connected-trench-v1.json` cuts floor, timber and earth bank (ground earth
bank material) along each port set over meadow. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- sandbags, barbed wire or bodies
- water in the trench floor
