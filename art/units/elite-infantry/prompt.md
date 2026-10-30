---
id: units/elite-infantry
title: Elite infantry
kind: unit
status: approved
runtime_ids: CP_eliteInf, EC_eliteInf
masters: atlas.png, section.png, team-full.png
---

## Look
A machine-gun section: one gun and four men (gunner, assistant with ammunition box, two riflemen), drawn for a strength of 4 to 6 of 6; the two-man team is `units/elite-infantry-1`. It stands on a 96x96 logical sprite above the hex centre and must read at
48 pixels on any terrain, with the owner colour clearly on uniform tunics and trousers.

## Base prompt
Asset: six-direction tactical sprite atlas, 1024x1536 transparent RGBA, 2 columns by 3 rows of 512x512
cells, row-major headings N, NW, SW, S, SE, NE. Subject in every cell: one two-man machine-gun team: a kneeling gunner behind a tripod-mounted, belt-fed, water-cooled machine gun and a crouched assistant beside him with an ammunition box and belt; exactly two people and one gun, tripod feet and muzzle fully visible. In each cell the
front, where the gun barrel and both men point, points towards that cell's heading. Saturated, naturally shaded magenta
only on uniform tunics and trousers; everything else in natural materials. Whole subject centred on (256,261) within
the central 65 percent of every cell, all extremities inside, alpha zero outside the objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the gun barrel and both men point) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the gun barrel and both men point) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the gun barrel and both men point) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the gun barrel and both men point) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the gun barrel and both men point) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the gun barrel and both men point) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-elite-infantry-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/elite-infantry/<palette>/<facing>-<state>`. Promote into `assets/units/elite-infantry/` and re-index.

`team-full.png` is the earlier imagegen atlas (the base prompt above, two men); it is the strength 1-3 look, installed as
`units/elite-infantry-1`. `section.png` is imagegen output `exec-bfec7a7d` (2026-10-30, try 3 of 4): the same gun and
men with two kneeling riflemen added, four men in the N, NW and S views and three in the others. Its prompt was the
`art/style.md` Preamble and Units text, then the base prompt with "one WWI machine-gun SECTION of exactly FOUR men",
the two-man atlas attached as the look reference, and a bounding box of at most 330 px. The installed `atlas.png` is
`section.png` scaled to 0.58 about the anchor so it reads at the size of the sappers:
`go run ./tools/unit-squad -src art/units/elite-infantry/section.png -n 0 -scale 0.58 -out art/units/elite-infantry/atlas.png`.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- more than four men or a second gun
