---
id: units/tunnellers
title: Tunnellers
kind: unit
status: approved
runtime_ids: CP_tunnellers, EC_tunnellers
masters: atlas.png
---

## Look
Two tunnelling miners at a timbered shaft head with a windlass, one with a pick, one with a sandbag; no rifles, no explosion. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on uniform tunics and trousers.

## Base prompt
LAYOUT: EXACT 1024x1536 PORTRAIT transparent RGBA canvas, TWO columns by THREE rows of 512x512 cells, invisible grid, no gutters. Six independently painted views of exactly the same group, one per cell, identical people, equipment and scale. Pivot at (256,261) in each cell; each whole group, including every tool, hose, helmet and boot, fits inside a maximum 270x270 box centred on the pivot, at least 120 completely transparent pixels between all extremities and every cell edge. Exactly six clean, disconnected groups, one per cell, nothing crossing into a neighbouring cell, no stray fragments.
CRITICAL DIRECTION ORDER (the FRONT is where the men face and work): top LEFT = NORTH, front points UP away from the viewer, backs of the men nearest the viewer; top RIGHT = NORTHWEST, front points to the upper LEFT, rear right quarter nearest the viewer; middle LEFT = SOUTHWEST, front points to the lower LEFT towards the viewer, front left quarter visible; middle RIGHT = SOUTH, front points DOWN towards the viewer, faces visible; bottom LEFT = SOUTHEAST, front points to the lower RIGHT towards the viewer, front right quarter visible; bottom RIGHT = NORTHEAST, front points to the upper RIGHT away, rear left quarter nearest the viewer. True separate orientations of one 3D group seen from one fixed elevated camera, never rotated or mirrored flat images.
COLOUR: saturated, naturally shaded MAGENTA (lit about RGB 225,40,185, shaded about 120,15,95) ONLY on the uniform tunics and trousers, as a technical owner colour; helmets, gas masks, skin, boots, puttees, belts, leather, wood, canvas and all metal stay natural dull colours (grey steel, brown leather, khaki canvas). Chunky, crisp, high-contrast silhouette that reads at 48 px. No national insignia, symbols, flags, letters or numbers.
Genuine alpha TRANSPARENCY outside the figures and between limbs. Absolutely NO ground, earth patch, grass, contact shadow, glow, haze, vignette, colour wash, checkerboard, frame or visible grid. Six isolated cut-out groups only; game-ready sprites, not a presentation illustration.

SUBJECT in every cell: two WWI tunnelling miners at a timbered shaft head: a small SQUARE wooden frame of heavy timbers around a dark square shaft opening, with a wooden windlass (a horizontal log drum on two posts with a crank handle and a rope going down into the shaft) on the frame; one miner in a cloth cap and rolled sleeves swings a pick, the other carries a full sandbag on his shoulder. The two miners stand at the FRONT side of the shaft frame, facing the heading, the windlass behind them. Exactly two men, one frame, one windlass. No rifles, no explosion, no smoke, no ground patch around the shaft beyond the timber frame itself.
CORRECTIONS TO THE PREVIOUS TRY: (1) the previous try painted the shaft frame and windlass as the SAME unrotated copy in all six cells (windlass always at the top, drum always horizontal): the WHOLE group, frame and windlass included, must turn with the heading. Rule: the two miners stand side by side on the side of the shaft TOWARDS the heading and face the heading; the windlass stands on the opposite, rear side of the shaft, its drum lying crosswise to the heading. So the windlass is at the BOTTOM of the group (nearest the viewer, in front of the men's backs) in the N cell; at the lower RIGHT in the NW cell; at the upper RIGHT in the SW cell; at the TOP behind the men in the S cell; at the upper LEFT in the SE cell; at the lower LEFT in the NE cell. The square frame is seen from a different side in every cell. (2) the previous try drew every group far too large, 325-377 px wide and 380-398 px tall: draw ALL SIX groups about 30 percent SMALLER, each whole group (frame, windlass with crank, both miners, pick and sandbag) at most 260 px wide and at most 260 px tall, so that about half of each cell stays empty around it. Same scale in all six cells.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the two miners face, the side of the shaft away from the windlass) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the two miners face, the side of the shaft away from the windlass) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the two miners face, the side of the shaft away from the windlass) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the two miners face, the side of the shaft away from the windlass) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the two miners face, the side of the shaft away from the windlass) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the two miners face, the side of the shaft away from the windlass) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-tunnellers-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/tunnellers/<palette>/<facing>-<state>`. Promote into `assets/units/tunnellers/` and re-index.

The atlas is imagegen output `exec-3a77f0fc` (2026-11-10, try 2), 1024x1536 RGBA, per-cell boxes 232-308 x 285-331 px, alpha haze 25 % at alpha 1-31 (threshold when cutting). Its prompt was the `art/style.md` Preamble and Units text, then the base prompt above; no attachments.

## Avoid
- rotated or mirrored copies of one view; a shaft frame that does not turn with the heading
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or sacking
- national insignia, flags or text
- rifles, explosions or smoke
