---
id: units/gas-pioneers
title: Gas pioneers
kind: unit
status: approved
runtime_ids: CP_gasPioneers, EC_gasPioneers
masters: atlas.png
---

## Look
Two gas pioneers in steel helmets and snout respirators kneeling behind a row of four upright gas cylinders with a hose; no rifles, no visible cloud. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on uniform tunics and trousers.

## Base prompt
LAYOUT: EXACT 1024x1536 PORTRAIT transparent RGBA canvas, TWO columns by THREE rows of 512x512 cells, invisible grid, no gutters. Six independently painted views of exactly the same group, one per cell, identical people, equipment and scale. Pivot at (256,261) in each cell; each whole group, including every tool, hose, helmet and boot, fits inside a maximum 260x260 box centred on the pivot, at least 126 completely transparent pixels between all extremities and every cell edge. Exactly six clean, disconnected groups, one per cell, nothing crossing into a neighbouring cell, no stray fragments.
CRITICAL DIRECTION ORDER (the FRONT is where the men face and work): top LEFT = NORTH, front points UP away from the viewer, backs of the men nearest the viewer; top RIGHT = NORTHWEST, front points to the upper LEFT, rear right quarter nearest the viewer; middle LEFT = SOUTHWEST, front points to the lower LEFT towards the viewer, front left quarter visible; middle RIGHT = SOUTH, front points DOWN towards the viewer, faces visible; bottom LEFT = SOUTHEAST, front points to the lower RIGHT towards the viewer, front right quarter visible; bottom RIGHT = NORTHEAST, front points to the upper RIGHT away, rear left quarter nearest the viewer. True separate orientations of one 3D group seen from one fixed elevated camera, never rotated or mirrored flat images.
COLOUR: saturated, naturally shaded MAGENTA (lit about RGB 225,40,185, shaded about 120,15,95) ONLY on the uniform tunics and trousers, as a technical owner colour; helmets, gas masks, skin, boots, puttees, belts, leather, wood, canvas and all metal stay natural dull colours (grey steel, brown leather, khaki canvas). Chunky, crisp, high-contrast silhouette that reads at 48 px. No national insignia, symbols, flags, letters or numbers.
Genuine alpha TRANSPARENCY outside the figures and between limbs. Absolutely NO ground, earth patch, grass, contact shadow, glow, haze, vignette, colour wash, checkerboard, frame or visible grid. Six isolated cut-out groups only; game-ready sprites, not a presentation illustration.

SUBJECT in every cell: a WWI gas-pioneer team: exactly TWO men in steel helmets wearing gas masks with round eyepieces and a short snout filter (respirator), kneeling side by side behind a short row of exactly FOUR upright tall steel gas cylinders (dark grey-green, rounded tops, valve caps) that stand close together in a line across the front of the men; a dark rubber hose runs from one cylinder valve forward over the front. One man works a cylinder valve with a spanner, the other holds the hose end. No rifles, no weapons, no visible gas cloud, no smoke, no sandbags, no trench.
SIZE: keep this exact design and these six headings, but the previous try still drew every group too large, 325-373 px wide and 337-366 px tall per cell. Draw ALL SIX groups about 30 percent SMALLER: each group (both men and all four cylinders) about 230-250 px wide and at most 250 px tall, so that well over half of each cell stays empty transparent space around it. The cylinders are about as tall as a kneeling man, not taller. Same scale in all six cells.
PLACEMENT RULE FOR THE HEADINGS: the row of cylinders always lies on the heading side of the two men, the men kneel behind it facing the heading. N cell: cylinders at the TOP, men below them showing their backs. NW cell: cylinder row at the UPPER LEFT running diagonally, men at the LOWER RIGHT of it, backs and right shoulders visible. SW cell: cylinder row at the LOWER LEFT, men above-right of it, masks turned to the lower left. S cell: cylinders at the BOTTOM, men behind them, masks facing the viewer. SE cell: cylinder row at the LOWER RIGHT, men above-left of it, masks turned to the lower right. NE cell: cylinder row at the UPPER RIGHT, men at the LOWER LEFT of it, backs and left shoulders visible. The previous try painted the NW cell as a near copy of the N cell; every cell must be a clearly different view.
ALPHA: an earlier try left a faint semi-transparent haze (about 21 percent of the canvas at alpha 1-31) around the figures; every pixel outside the figures must be exactly alpha zero.
Image 1 is the previous try (the same atlas layout): keep its design, its six headings and its rendering exactly, but paint all six groups about 30 percent smaller around their cell pivots; do not copy its size.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the two pioneers face and the cylinder row stands) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the two pioneers face and the cylinder row stands) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the two pioneers face and the cylinder row stands) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the two pioneers face and the cylinder row stands) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the two pioneers face and the cylinder row stands) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the two pioneers face and the cylinder row stands) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-gas-pioneers-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/gas-pioneers/<palette>/<facing>-<state>`. Promote into `assets/units/gas-pioneers/` and re-index.

The atlas is imagegen output `exec-688b6afa` (2026-11-10, try 4 of 4), 1024x1536 RGBA, per-cell boxes 252-287 x 255-275 px, alpha haze 5.5 % (threshold when cutting). Its prompt was the `art/style.md` Preamble and Units text, then the base prompt above, with try 3 (`exec-465901fb`, the same design drawn 1.35x too large) attached as image 1.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood, gas masks or cylinders
- national insignia, flags or text
- rifles, a visible gas cloud or smoke
