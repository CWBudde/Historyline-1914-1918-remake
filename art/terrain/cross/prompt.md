---
id: terrain/cross
title: Level crossing
kind: terrain
status: approved
runtime_ids: cross1, cross2, cross3, cross4, cross1_w, cross2_w, cross3_w, cross4_w
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
| winter | cross1_w, cross2_w, cross3_w, cross4_w | The same crossings built from the winter path, road and rail (`path14_w`, `road14_w`, `road25_w`, `path36_w` over `rail36_w`, `rail14_w`); snow trodden into the plank infill. |

## Edges & joins
Road, path and rail are each 22 world pixels wide with a 2-pixel border and keep their centre lines
through the ports, so they join ordinary road, path and rail tiles exactly. The flush plank infill
covers the intersection only: the way's width plus 1 world pixel either side, and 9 world pixels either
side of the track's centre line (4.5 scaled with the rail family); the steel runs unbroken over it.

## Post
Derived only: `art-review -mixed art/pipeline/connected-mixed-v3.json` layers the rail variant over the
path or road variant. `connected-mixed-winter-v2.json` derives the winter crossings from
`connected-mixed-v3.json`, swapping every layer for its `_w` form of the same profile from
`connected-road-winter-v1.json`, `connected-paths-winter-v2.json` and `connected-rail-winter-v2.json`; the plank infill blends 0.85 toward the winter surface of the
crossing way, then 0.15 toward the rut slush, so the boards read as snowed over (infill within ΔE 3.7–4.3
of that surface, C\* ≤ 3.7, in `TestWinterPathsAndRailsMeetCriterion`). The tagged generator `TestGenerateWinterPathsAndRails` owns all winter crossings
(`TestGenerateWinterRoads` only the winter roads): rerun it after changing a winter road, path or rail. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- crossing gates, signals or signs
