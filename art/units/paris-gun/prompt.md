---
id: units/paris-gun
title: Paris Gun
kind: unit
status: approved
runtime_ids: CP_parisGun
masters: atlas.png
---

## Look
The Paris Gun: an extremely long, slender, braced barrel on a heavy multi-bogie railway carriage on a short rail stretch, drawn compact so that the whole gun stays inside the central 65 percent of the cell on the diagonals too. It stands on a 96x96 logical sprite above the hex centre, with the owner colour on the carriage paint. The master is a layout-corrected (20 % smaller) edit of the generated atlas.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas of six elevated overhead painted views of the same unit.
SUBJECT in every cell: the WWI Paris Gun: an extremely long, very slender steel gun barrel, braced along its upper side by a thin truss of tension rods, raised at a modest angle on a heavy box-girder railway carriage with two multi-axle rail bogies (many small steel rail wheels), the whole carriage standing on a short straight stretch of standard railway track (two steel rails on wooden sleepers) that is only as long as the carriage. FRONT = the muzzle end of the barrel; the breech sits over the carriage. The barrel is clearly far longer than the carriage, but the whole gun must stay compact: total length of barrel plus carriage plus rails at most 300 px measured along the gun, so it fits the 320x320 box on every heading including the diagonals NW, SW, SE, NE; on the diagonals the gun lies along the cell diagonal and never reaches closer than 80 px to any cell edge. Saturated, naturally shaded MAGENTA only on the carriage girders and gun cradle paint (owner tint mask); barrel natural dark blued steel, bracing rods dark steel, rail wheels and rails natural steel, sleepers natural brown wood. No crew, no smoke, no muzzle flash, no locomotive, no ground beyond the sleepers.
CORRECTION: the previous try had all six headings right and a fine design, but it painted the guns far too large: up to 453 px wide and 450 px tall, only 35 px from the cell edge. Paint every gun about 35 percent SMALLER around the pivot (256,261): each whole gun including the muzzle tip, railings and the rail stretch must fit inside a 300x300 box centred on (256,261), at least 100 px clear to every cell edge; in the N and S cells the gun is at most 300 px tall, on the diagonals at most 300 px wide and 300 px tall. Keep the barrel clearly about twice as long as the carriage. The carriage paint was a dull raspberry; make it saturated, naturally shaded pure MAGENTA (lit about RGB 215,40,200, shaded about RGB 120,10,115). Every pixel outside the gun exactly alpha 0, no faint veil (the previous try had alpha 1-31 haze over 33 percent of the canvas).

LAYOUT CORRECTION TO THE PROVIDED ATLAS: keep exactly the same gun design, the same six headings, the same painting and the same magenta carriage, but make ALL SIX guns uniformly 20 percent SMALLER around their cell pivots (256,261). Each whole gun including the muzzle tip, railings and the rail stretch must fit a 290x290 box centred on (256,261) of its own cell (the previous atlas measured up to 366 px wide in the SW cell and 362 px tall in the N cell), with at least 110 completely transparent pixels between every extremity and each cell edge. Output an EXACT 1024x1536 RGBA canvas with genuine alpha transparency: alpha exactly 0 outside the six guns, no ground beyond the sleepers, no shadow, haze, glow, backdrop, checkerboard, frame or visible grid. Exactly six clean, disconnected guns, one per cell; no text, numbers, insignia or flags.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (the muzzle end of the barrel) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (the muzzle end of the barrel) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (the muzzle end of the barrel) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (the muzzle end of the barrel) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (the muzzle end of the barrel) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (the muzzle end of the barrel) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-paris-gun-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/paris-gun/<palette>/<facing>-<state>`. Promote into `assets/units/paris-gun/` and re-index.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- crew, smoke, muzzle flash or a locomotive
- a barrel reaching beyond the central 65 percent of the cell
