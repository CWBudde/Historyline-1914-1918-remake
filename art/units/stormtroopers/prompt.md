---
id: units/stormtroopers
title: Stormtroopers
kind: unit
status: approved
runtime_ids: CP_stormtroopers
masters: atlas.png
---

## Look
Three German assault troops in Stahlhelm with grenade bags, stick grenades, a carbine and a light machine gun, crouching forward to attack. The CP side only, but painted with the magenta owner colour like every unit. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on uniform tunics and trousers.

## Base prompt
LAYOUT: EXACT 1024x1536 PORTRAIT transparent RGBA canvas, TWO columns by THREE rows of 512x512 cells, invisible grid, no gutters. Six independently painted views of exactly the same group, one per cell, identical people, equipment and scale. Pivot at (256,261) in each cell; each whole group, including every tool, hose, helmet and boot, fits inside a maximum 270x270 box centred on the pivot, at least 120 completely transparent pixels between all extremities and every cell edge. Exactly six clean, disconnected groups, one per cell, nothing crossing into a neighbouring cell, no stray fragments.
CRITICAL DIRECTION ORDER (the FRONT is where the men face and work): top LEFT = NORTH, front points UP away from the viewer, backs of the men nearest the viewer; top RIGHT = NORTHWEST, front points to the upper LEFT, rear right quarter nearest the viewer; middle LEFT = SOUTHWEST, front points to the lower LEFT towards the viewer, front left quarter visible; middle RIGHT = SOUTH, front points DOWN towards the viewer, faces visible; bottom LEFT = SOUTHEAST, front points to the lower RIGHT towards the viewer, front right quarter visible; bottom RIGHT = NORTHEAST, front points to the upper RIGHT away, rear left quarter nearest the viewer. True separate orientations of one 3D group seen from one fixed elevated camera, never rotated or mirrored flat images.
COLOUR: saturated, naturally shaded MAGENTA (lit about RGB 225,40,185, shaded about 120,15,95) ONLY on the uniform tunics and trousers, as a technical owner colour; helmets, gas masks, skin, boots, puttees, belts, leather, wood, canvas and all metal stay natural dull colours (grey steel, brown leather, khaki canvas). Chunky, crisp, high-contrast silhouette that reads at 48 px. No national insignia, symbols, flags, letters or numbers.
Genuine alpha TRANSPARENCY outside the figures and between limbs. Absolutely NO ground, earth patch, grass, contact shadow, glow, haze, vignette, colour wash, checkerboard, frame or visible grid. Six isolated cut-out groups only; game-ready sprites, not a presentation illustration.

SUBJECT in every cell: a WWI German assault squad of exactly THREE men crouching forward together in an attack posture: coal-scuttle steel helmets (Stahlhelm), canvas grenade bags slung across the chest, stick grenades (wooden handle, metal head) tucked in their belts, one carries a short carbine, one holds a stick grenade raised ready to throw, one carries a light machine gun with a bipod and a drum or box magazine. All three lean towards the heading. No national insignia, no explosions, no smoke, no muzzle flash.
CORRECTIONS TO THE PREVIOUS TRY: keep the same three-man design and headings, but (1) the previous try drew every group far too large, 383-416 px wide and 383-407 px tall per cell: draw ALL SIX groups about 35 percent SMALLER, each group (all three men, the raised grenade, the carbine muzzle and the machine-gun bipod) at most 260 px wide and at most 260 px tall, so that about half of each cell stays empty around it; (2) the tunics were blotchy, mottled crimson-purple like camouflage: paint the tunics and trousers in clean, evenly and smoothly shaded saturated magenta with no blotches or stains; (3) in the SW cell the men must clearly turn to the lower LEFT (their guns and faces pointing to the lower left corner), distinct from the S cell where they face straight down.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the three men lean and their weapons point) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the three men lean and their weapons point) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the three men lean and their weapons point) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the three men lean and their weapons point) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the three men lean and their weapons point) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the three men lean and their weapons point) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-stormtroopers-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/stormtroopers/<palette>/<facing>-<state>`. Promote into `assets/units/stormtroopers/` and re-index.

The atlas is imagegen output `exec-59ba5ffa` (2026-11-10, try 2), 1024x1536 RGBA, per-cell boxes 286-352 x 302-341 px (about 1.2x the installed foot teams), alpha haze 10 %. Its prompt was the `art/style.md` Preamble and Units text, then the base prompt above; no attachments.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood, helmets or grenade bags
- national insignia, flags or text
- explosions, smoke or muzzle flash
