---
id: units/patrol-boat
title: Patrol boat
kind: unit
status: candidate
runtime_ids: CP_patrolBoat, EC_patrolBoat
masters: atlas.png
---

## Look
A small armoured patrol boat. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on the hull and wheelhouse paint.

## Base prompt
Asset: six-direction tactical sprite atlas, 1024x1536 transparent RGBA, 2 columns by 3 rows of 512x512
cells, row-major headings N, NW, SW, S, SE, NE. Subject in every cell: one small low WWI patrol boat with a boxy armoured wheelhouse, a short bow, an open stern and a small mounted gun; no water. In each cell the
front, where the bow points, points towards that cell's heading. Saturated, naturally shaded magenta
only on the hull and wheelhouse paint; everything else in natural materials. Whole subject centred on (256,261) within
the central 65 percent of every cell, all extremities inside, alpha zero outside the objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the bow points) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the bow points) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the bow points) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the bow points) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the bow points) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the bow points) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-patrol-boat-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/patrol-boat/<palette>/<facing>-<state>`. Promote into `assets/units/patrol-boat/` and re-index.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- water, wake or reflections
