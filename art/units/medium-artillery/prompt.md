---
id: units/medium-artillery
title: Medium artillery
kind: unit
status: approved
runtime_ids: CP_mArtillery, EC_mArtillery
masters: atlas.png
---

## Look
A medium field gun without crew. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on the carriage and shield paint.

## Base prompt
Asset: six-direction tactical sprite atlas, 1024x1536 transparent RGBA, 2 columns by 3 rows of 512x512
cells, row-major headings N, NW, SW, S, SE, NE. Subject in every cell: one unlimbered medium field gun with two large spoked wheels, a long slim barrel and a low carriage; no crew, no horses. In each cell the
front, where the muzzle points, with the trail opposite, points towards that cell's heading. Saturated, naturally shaded magenta
only on the carriage and shield paint; everything else in natural materials. Whole subject centred on (256,261) within
the central 65 percent of every cell, all extremities inside, alpha zero outside the objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the muzzle points, with the trail opposite) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the muzzle points, with the trail opposite) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the muzzle points, with the trail opposite) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the muzzle points, with the trail opposite) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the muzzle points, with the trail opposite) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the muzzle points, with the trail opposite) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-medium-artillery-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/medium-artillery/<palette>/<facing>-<state>`. Promote into `assets/units/medium-artillery/` and re-index.

The atlas is imagegen output `exec-8fd94b2d` (2026-10-30), each cell scaled to 0.85 about (256,261) so the three artillery classes keep a size ladder (light slim and small, medium mid, heavy bulky) inside the framing guard.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- crew, horses or riders
