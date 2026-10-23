---
id: units/light-artillery
title: Light artillery
kind: unit
status: candidate
runtime_ids: CP_lArtillery, EC_lArtillery
masters: atlas.png
---

## Look
A light field gun with two crewmen. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on crew uniforms and the gun shield paint.

## Base prompt
Asset: six-direction tactical sprite atlas, 1024x1536 transparent RGBA, 2 columns by 3 rows of 512x512
cells, row-major headings N, NW, SW, S, SE, NE. Subject in every cell: one light WWI field gun with two wooden spoked wheels on a single axle, a box trail and a short shield, with exactly two crewmen standing behind it. In each cell the
front, where the muzzle points, with the trail opposite, points towards that cell's heading. Saturated, naturally shaded magenta
only on crew uniforms and the gun shield paint; everything else in natural materials. Whole subject centred on (256,261) within
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
`art-review -unit art/pipeline/unit-light-artillery-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/light-artillery/<palette>/<facing>-<state>`. Promote into `assets/units/light-artillery/` and re-index.

The atlas is imagegen output `exec-2b7881ae`, preferred on the candidate desk over the earlier
atlas (2026-10-23).

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- limbers or horses
- crew scaled differently from the gun
