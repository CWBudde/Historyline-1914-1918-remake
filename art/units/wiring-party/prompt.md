---
id: units/wiring-party
title: Wiring party
kind: unit
status: approved
runtime_ids: CP_wiring, EC_wiring
masters: atlas.png
---

## Look
Two soldiers carrying a barbed-wire coil on a shoulder pole and a bundle of screw pickets. It stands on a 96x96
logical sprite above the hex centre and must read at 48 pixels on any terrain, with the owner colour clearly on
uniform tunics and trousers.

## Base prompt
Asset: six-direction tactical sprite atlas, high elevated strategy camera looking steeply down (about 55 degrees), so the men are foreshortened and their caps, shoulders, the pole and the top of the coil are seen from above, EXACT 1024x1536 transparent RGBA canvas, PORTRAIT orientation, 2 columns by 3 rows of 512x512 cells, invisible grid, row-major headings N, NW, SW, S, SE, NE. Subject in every cell: a two-man WWI WIRING PARTY walking in single file in the direction of the heading: the two soldiers carry between them, on their shoulders, one long wooden pole threaded through a big round coil of barbed wire (the coil hangs from the pole midway between them); the rear man also carries a bundle of four or five corkscrew iron screw pickets (twisted iron stakes) under his free arm. Both wear peaked caps or soft caps, thick leather gauntlets and puttees; NO rifles, no weapons. The two men, the pole and the coil form ONE compact group; same two men, same coil, same stakes and same scale in all six cells.
In each cell the front, where the leading man walks and both men face, points towards that cell's heading: the pole and the line of the two men run along the heading.
DIRECTIONS: top LEFT = N: the men walk UP and away from the viewer, their backs visible. Top RIGHT = NW: walking to the upper LEFT, backs and right flanks visible. Middle LEFT = SW: walking to the lower LEFT towards the viewer, faces and right flanks visible. Middle RIGHT = S: walking DOWN towards the viewer, faces visible. Bottom LEFT = SE: walking to the lower RIGHT towards the viewer, faces and left flanks visible. Bottom RIGHT = NE: walking to the upper RIGHT, backs and left flanks visible. Each cell is an independently painted pose; never a rotated or mirrored copy.
Saturated, naturally shaded MAGENTA only on the uniform tunics and trousers of both men; caps, skin, leather gauntlets, puttees, boots, the wooden pole, the grey steel barbed-wire coil and the rusty iron stakes stay in natural colours. Fine believable painted detail, strong light/dark separation, the group must read at 48 px.
Real alpha transparency outside the figures and between legs, pole and coil: no ground, shadow, haze, glow or backdrop. No insignia, flags, text or labels.
LAYOUT CORRECTION TO THE PROVIDED WIRING-PARTY ATLAS (image 1): keep exactly the same two men, uniforms, coil, pole, stakes, six headings, poses, size ratio and painting style as image 1, and only fix the placement: make every group about 12 percent smaller and centre it vertically on y=261 of its cell. Image 1 measured groups 126-230 px wide and 273-311 px tall; the top-left N group spans y=145-455 in its cell (centre 300, only 57 px above the cell's bottom edge) and the middle-right S group spans y=113-424. Now each whole group, including caps, boots, pole ends and stakes, must fit into a maximum 270 px tall and 240 px wide box whose centre is (256,261) of its own 512x512 cell: caps no higher than y=126 and boots no lower than y=396 within the cell, at least 110 completely transparent pixels to all four cell edges. Alpha exactly zero outside the figures, crisp antialiased edges only, no halo, no glow, no shadow.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the leading man walks and both men face) points up, away from the viewer; their backs are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the leading man walks and both men face) points to the upper left; backs and right flanks are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the leading man walks and both men face) points to the lower left; faces and right flanks are visible. |
| S (`s`) | — | Middle-right cell: the front (where the leading man walks and both men face) points down, towards the viewer; faces are visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the leading man walks and both men face) points to the lower right; faces and left flanks are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the leading man walks and both men face) points to the upper right; backs and left flanks are visible. |

## Post
Selected atlas: uwire-04 of gen-2026-11-10 (exec-a9baee8a, 1024x1536). The prompt followed the `art/style.md`
Preamble and Units text; image 1 was the previous try at full size. The N group sits about 41 px below the cell
centre. `art-review -unit art/pipeline/unit-wiring-party-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/wiring-party/<palette>/<facing>-<state>`. Promote into `assets/units/wiring-party/` and re-index.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood, the coil or the stakes
- national insignia, flags or text
- rifles, mines or explosions
