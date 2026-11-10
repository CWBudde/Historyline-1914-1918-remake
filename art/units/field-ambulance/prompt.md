---
id: units/field-ambulance
title: Field ambulance
kind: unit
status: approved
runtime_ids: CP_ambulance, EC_ambulance
masters: atlas.png
---

## Look
A WWI motor field ambulance: a light truck chassis with an open cab and a tall canvas-covered body. It stands on a 96x96 logical sprite above the hex centre and must read at 48 pixels on any terrain, with the owner colour on the canvas body only; the small red crosses on white squares stay pure red. The master is a layout-corrected (25 % smaller) edit of the generated atlas.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas of six elevated overhead painted views of the same unit.
SUBJECT in every cell: one WWI motor field ambulance: a light truck chassis with an open-sided driver's cab under a small roof, a short bonnet with radiator at the front, and behind the cab a tall boxy canvas-covered ambulance body (a canvas tilt stretched over hoops, rear flaps closed), four thin spoked wheels with dark tyres. FRONT = the radiator and bonnet end. The vehicle is about 290 px long and fits the 320x320 box on every heading including the diagonals. Saturated, naturally shaded MAGENTA only on the canvas body/tilt (owner tint mask); bonnet, cab, chassis, mudguards and wheels natural dark olive-grey painted steel and wood, tyres dark. On each long side of the canvas body one small red cross on a small white square (pure red RGB about 200,30,30, never magenta), and the same on the rear flap if visible; nothing else on the canvas. No driver figure sticking out, no text, no numbers, no other markings.
SIZE AND ALPHA CORRECTION: two previous tries had all six headings right but the vehicles stayed too large: 415x426 px, then still 350 px wide and 364 px tall (only 64 px from the top edge in the S cell) although 310 was asked. Paint every ambulance at about 75 percent of that last size: each whole vehicle including mudguards, lamps, roof and wheels about 260 px wide and 270 px tall (never more than 290), centred on (256,261), with at least 110 px of empty transparent space to every cell edge. Keep the same design, the same small red crosses and the same six headings. Every pixel outside the vehicle exactly alpha 0, no faint veil or speckle (the last try had alpha 1-31 haze over 10 percent of the canvas).

LAYOUT CORRECTION TO THE PROVIDED ATLAS: keep exactly the same vehicle design, the same six headings, the same painting, the same magenta canvas and the same small red crosses, but make ALL SIX ambulances uniformly 25 percent SMALLER around their cell pivots (256,261). Each vehicle including mudguards, lamps, roof and wheels must fit a 280x280 box centred on (256,261) of its own cell (the previous atlas measured up to 359 px wide and 364 px tall), with at least 110 completely transparent pixels between every extremity and each cell edge. Output an EXACT 1024x1536 RGBA canvas with genuine alpha transparency: alpha exactly 0 outside the six vehicles, no ground, shadow, haze, glow, backdrop, checkerboard, frame or visible grid. Exactly six clean, disconnected vehicles, one per cell; no text, numbers or other markings.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (the radiator and bonnet end) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (the radiator and bonnet end) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (the radiator and bonnet end) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (the radiator and bonnet end) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (the radiator and bonnet end) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (the radiator and bonnet end) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-field-ambulance-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/field-ambulance/<palette>/<facing>-<state>`. Promote into `assets/units/field-ambulance/` and re-index.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- magenta crosses or magenta on the cab, bonnet or chassis
- text, numbers or other markings
- driver figures sticking out
