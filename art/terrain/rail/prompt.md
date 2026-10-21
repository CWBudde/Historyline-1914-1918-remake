---
id: terrain/rail
title: Railway
kind: terrain
status: candidate
runtime_ids: rail13, rail14, rail15, rail236, rail24, rail25, rail26, rail346, rail35, rail356, rail36, rail46
masters: rail-ballast-shared-v2.png, rail-steel-shared-v2.png, rail-ballast-shared-v1.png, rail-steel-shared-v1.png
---

## Look
Single-track railway: a narrow grey ballast bed with dark timber sleepers and two thin steel rails,
in the same muted palette and scale as the road.

## Base prompt
Opaque 6:5 material canvas filled edge to edge with compacted railway ballast: small angular weathered
grey limestone and greige stone chips, closely interlocked, tiny earth pockets between them, muted
neutral grey-brown, quiet contrast. No rails, sleepers, grass, verges or track lines.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| rail13 | rail13 | Track wide bend between ports 1 NE, 3 S. |
| rail14 | rail14 | Track straight run between ports 1 NE, 4 SW. |
| rail15 | rail15 | Track wide bend between ports 1 NE, 5 NW. |
| rail236 | rail236 | Track T or Y junction to ports 2 SE, 3 S, 6 N. |
| rail24 | rail24 | Track wide bend between ports 2 SE, 4 SW. |
| rail25 | rail25 | Track straight run between ports 2 SE, 5 NW. |
| rail26 | rail26 | Track wide bend between ports 2 SE, 6 N. |
| rail346 | rail346 | Track T or Y junction to ports 3 S, 4 SW, 6 N. |
| rail35 | rail35 | Track wide bend between ports 3 S, 5 NW. |
| rail356 | rail356 | Track T or Y junction to ports 3 S, 5 NW, 6 N. |
| rail36 | rail36 | Track straight run between ports 3 S, 6 N. |
| rail46 | rail46 | Track wide bend between ports 4 SW, 6 N. |

## Edges & joins
The ballast bed is 11 world pixels wide with a 1-pixel earth border, centred on the shared port centres
and continuing through each port; rails and sleepers run on the same centre line, so track joins
exactly and crossings line up.

## Post
`art-review -connected art/pipeline/connected-rail-v2.json` cuts the ballast, lays timber sleepers (trench
timber material) and steel rails (steel material) along each port set. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

Selected built-in imagegen masters: ballast is native 1536×1024 RGB; steel is native 1254×1254 RGB.
The steel prompt specifies matte weathered dark grey steel, fine scratches, slight oxide and diffuse
highlights; no chrome, rail silhouettes or objects. The former material sources remain retained drafts.
The ballast footprint, rails and registered ports are unchanged.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- rails or sleepers in the material
- chrome highlights on steel
