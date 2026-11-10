---
id: units/airship
title: Zeppelin airship
kind: unit
status: approved
runtime_ids: CP_zeppelin
masters: atlas.png
---

## Look
A WWI rigid Zeppelin airship flying high: a slender silver-grey doped-fabric hull with a rounded nose, cruciform tail fins and two engine gondolas. It flies on a 96x96 logical sprite above the hex centre with no ground; the owner colour is a broad band behind the nose and the four fins, the rest stays natural silver. The master is a layout-corrected (20 % smaller) edit of the generated atlas.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas of six elevated overhead painted views of the same unit.
SUBJECT in every cell: one WWI rigid Zeppelin airship flying high: a long cigar-shaped hull of silver-grey doped fabric with faint longitudinal girder lines and ring frames, a rounded blunt NOSE, a tapered tail with four cruciform tail fins (two horizontal, two vertical, each with a rudder or elevator), and two small engine gondolas slung under the hull, one forward and one aft, each with a small propeller. FRONT = the rounded nose; the cruciform fins are at the rear. The hull length is at most 290 px so that on the diagonals NW, SW, SE, NE the airship, lying along the cell diagonal, still fits the 320x320 box with 80 px clear to every cell edge; the hull is level in the world, never tilted up like a balloon. It flies: nothing below it, no ground, no shadow, no clouds. Owner tint: saturated, naturally shaded MAGENTA only on one broad painted band around the hull behind the nose and on the four tail fins; the rest of the hull natural doped silver-grey fabric, gondolas natural grey metal and dark wood propellers. No crew figures, no bombs falling, no searchlights.
CORRECTION: the previous try had all six headings right but the airships were too large (up to 424 px long, only 37 px from the cell edge), too fat (hull length only about 3 times its diameter, like a blimp) and the owner paint was a dull rose pink. Now: a SLENDER rigid Zeppelin, hull length about 7 times its diameter, with a long parallel-sided middle section; each whole airship including fins, gondolas and propellers at most 300 px long and inside a 300x300 box centred on (256,261), at least 100 px clear to every cell edge (N and S views at most 300 px tall). Owner band and fins in saturated, naturally shaded pure MAGENTA (lit about RGB 215,40,200, shaded about RGB 120,10,115), not pink or rose. Every pixel outside the airship exactly alpha 0, no faint veil (the previous try had alpha 1-31 haze over 37 percent of the canvas). Keep the same six headings.
SECOND CORRECTION: the second try fixed the shape and colour (keep exactly that slender hull, magenta band and fins) but the airships still measured up to 356 px wide and 349 px tall, only 66 px from the cell edge, and the empty canvas still carried faint alpha 1-31 haze over 27 percent of its area. Paint each airship at about 78 percent of that size: hull plus fins at most 270 px long, inside a 270x270 box centred on (256,261), at least 120 px of exactly alpha 0 space to every cell edge. The background must be perfectly clear: alpha exactly 0 everywhere outside the airship, no soft atmospheric veil, glow, cloud wisp or speckle.

LAYOUT CORRECTION TO THE PROVIDED ATLAS: keep exactly the same airship design, the same six headings, the same painting and the same magenta band and fins, but make ALL SIX airships uniformly 20 percent SMALLER around their cell pivots (256,261). Each whole airship including fins, gondolas and propellers must fit a 290x290 box centred on (256,261) of its own cell (the previous atlas measured 370 px tall in the N cell and up to 350 px wide on the diagonals), with at least 110 completely transparent pixels between every extremity and each cell edge. Output an EXACT 1024x1536 RGBA canvas with genuine alpha transparency: alpha exactly 0 outside the six airships, no ground, shadow, cloud, haze, glow, backdrop, checkerboard, frame or visible grid. Exactly six clean, disconnected airships, one per cell; no text, numbers, insignia or flags.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (the rounded nose, with the cruciform fins at the rear) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (the rounded nose, with the cruciform fins at the rear) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (the rounded nose, with the cruciform fins at the rear) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (the rounded nose, with the cruciform fins at the rear) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (the rounded nose, with the cruciform fins at the rear) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (the rounded nose, with the cruciform fins at the rear) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-airship-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/airship/<palette>/<facing>-<state>`. Promote into `assets/units/airship/` and re-index.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- a stubby blimp hull or a hull tilted up like a balloon
- clouds, searchlights or falling bombs
