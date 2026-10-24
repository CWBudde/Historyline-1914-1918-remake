---
id: terrain/cross
title: Level crossing
kind: terrain
status: candidate
runtime_ids: cross1, cross2, cross3, cross4, cross2_w, cross3_w
masters: 
---

## Look
Level crossing where a road or path crosses the railway on the flat: rails pass through the road
surface with a plain timber or asphalt infill.

## Base prompt
Top-down level crossing of a single-track railway and a country road or earth path: the rails run
straight through, the road surface continues between and over them with a short timber plank infill,
both at the same widths and colours as the road, path and railway materials. No gates, signs or
markings.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cross1 | cross1 | Crossing: path between ports 1 NE and 4 SW over track between 3 S and 6 N. |
| cross2 | cross2 | Crossing: road between ports 1 NE and 4 SW over track between 3 S and 6 N. |
| cross3 | cross3 | Crossing: road between ports 2 SE and 5 NW over track between 3 S and 6 N. |
| cross4 | cross4 | Crossing: path between ports 3 S and 6 N over track between 1 NE and 4 SW. |
| winter | cross2_w, cross3_w | cross2 and cross3 with the winter road (`road14_w`, `road25_w`); the track stays. |

## Edges & joins
Both ways keep their family widths and centre lines through the ports, so they join ordinary road,
path and rail tiles exactly.

## Post
Derived only: `art-review -mixed art/pipeline/connected-mixed-v1.json` layers the path or road variant over
the rail variant. The winter road crossings swap the road layer for the winter road of
`connected-road-winter-v1.json` (tagged generator `TestGenerateWinterRoads`). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- crossing gates, signals or signs
