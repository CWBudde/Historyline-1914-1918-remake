---
id: units/flamethrower
title: Flamethrower
kind: unit
status: approved
runtime_ids: CP_flamethrower, EC_flamethrower
masters: atlas.png
---

## Look
A two-man flamethrower team: one carries a backpack of two round fuel cylinders, the other aims the long lance; unlit, no flame. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on uniform tunics and trousers.

## Base prompt
LAYOUT: EXACT 1024x1536 PORTRAIT transparent RGBA canvas, TWO columns by THREE rows of 512x512 cells, invisible grid, no gutters. Six independently painted views of exactly the same group, one per cell, identical people, equipment and scale. Pivot at (256,261) in each cell; each whole group, including every tool, hose, helmet and boot, fits inside a maximum 270x270 box centred on the pivot, at least 120 completely transparent pixels between all extremities and every cell edge. Exactly six clean, disconnected groups, one per cell, nothing crossing into a neighbouring cell, no stray fragments.
CRITICAL DIRECTION ORDER (the FRONT is where the men face and work): top LEFT = NORTH, front points UP away from the viewer, backs of the men nearest the viewer; top RIGHT = NORTHWEST, front points to the upper LEFT, rear right quarter nearest the viewer; middle LEFT = SOUTHWEST, front points to the lower LEFT towards the viewer, front left quarter visible; middle RIGHT = SOUTH, front points DOWN towards the viewer, faces visible; bottom LEFT = SOUTHEAST, front points to the lower RIGHT towards the viewer, front right quarter visible; bottom RIGHT = NORTHEAST, front points to the upper RIGHT away, rear left quarter nearest the viewer. True separate orientations of one 3D group seen from one fixed elevated camera, never rotated or mirrored flat images.
COLOUR: saturated, naturally shaded MAGENTA (lit about RGB 225,40,185, shaded about 120,15,95) ONLY on the uniform tunics and trousers, as a technical owner colour; helmets, gas masks, skin, boots, puttees, belts, leather, wood, canvas and all metal stay natural dull colours (grey steel, brown leather, khaki canvas). Chunky, crisp, high-contrast silhouette that reads at 48 px. No national insignia, symbols, flags, letters or numbers.
Genuine alpha TRANSPARENCY outside the figures and between limbs. Absolutely NO ground, earth patch, grass, contact shadow, glow, haze, vignette, colour wash, checkerboard, frame or visible grid. Six isolated cut-out groups only; game-ready sprites, not a presentation illustration.

SUBJECT in every cell: a WWI flamethrower team of exactly TWO men in steel helmets: the rear man carries on his back a large backpack fuel tank made of TWO round upright steel cylinders side by side on a carrying frame, a dark flexible hose runs from the tanks to the front man; the front man crouches forward aiming a long straight metal lance (about one man-height long) with both hands, its nozzle pointing forward towards the heading. The lance is UNLIT: no flame, no fire, no sparks, no smoke, no glow at the nozzle. No rifles.
SIZE CORRECTION: keep this exact design, but the previous try drew every group far too large: 347-376 px wide on the diagonal views and 415-446 px TALL on the N and S views (the lance and the two men stacked in a long vertical line). Draw ALL SIX groups about 30 percent SMALLER: each group (both men, the tanks and the whole lance) at most 270 px wide and at most 270 px tall. On the N and S views place the two men side by side, slightly offset, rather than one behind the other, so the group is not a tall column; keep the lance short enough to stay inside the 270 px box. Same scale in all six cells.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the lance nozzle and both men point) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the lance nozzle and both men point) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the lance nozzle and both men point) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the lance nozzle and both men point) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the lance nozzle and both men point) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the lance nozzle and both men point) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-flamethrower-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/flamethrower/<palette>/<facing>-<state>`. Promote into `assets/units/flamethrower/` and re-index.

The atlas is imagegen output `exec-0467f5d9` (2026-11-10, try 2), 1024x1536 RGBA, per-cell boxes 217-319 x 280-346 px (N 346 px tall with the raised lance), groups about 30 px below the pivot, alpha haze 6 %. Its prompt was the `art/style.md` Preamble and Units text, then the base prompt above; no attachments.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or the fuel tanks
- national insignia, flags or text
- flame, fire, smoke or a glowing nozzle; rifles
