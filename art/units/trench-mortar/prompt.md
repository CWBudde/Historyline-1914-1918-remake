---
id: units/trench-mortar
title: Trench mortar
kind: unit
status: approved
runtime_ids: CP_trenchMortar, EC_trenchMortar
masters: atlas.png
---

## Look
A light trench mortar team: a short stubby tube on a baseplate and bipod, two kneeling crewmen and a small stack of finned rounds. It stands on a 96x96 logical sprite above the hex centre and must read at 48 pixels on any terrain, with the owner colour clearly on the crew's tunics and trousers. The master is a layout-corrected (25 % smaller) edit of the generated atlas.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas of six elevated overhead painted views of the same unit.
SUBJECT in every cell: one WWI light trench mortar team: a short, stubby, wide-mouthed steel mortar tube (Stokes mortar or light Minenwerfer type) set steeply angled on a square steel baseplate and a thin bipod, with exactly TWO crewmen kneeling beside it, one on each side of the tube, and a small neat stack of four or five finned practice rounds on the ground just behind the baseplate. FRONT = the direction the mouth of the tilted tube leans and points (the bipod stands on the front side, the baseplate and round stack at the rear). The group is compact, roughly as wide as it is deep, about 240 to 300 px across inside the 320x320 box. Saturated, naturally shaded MAGENTA only on the two crewmen's uniform tunics and trousers (owner tint mask); steel tube, baseplate, bipod and rounds dark natural grey steel, helmets natural grey-green steel, skin, boots and belts natural. No rifles, no smoke, no firing, no sandbags, no ground patch.
SIZE, DIRECTION AND ALPHA CORRECTION: two previous tries painted each team 340 to 378 px wide and up to 338 px tall although 300 was asked; paint the figures and the mortar HALF their previous size relative to the cell, so each whole team including boots, bipod feet and round stack is about 250 px wide and 250 px tall (at most 270), centred on (256,261), with at least 120 px of empty transparent space to every cell edge. The tube is steeply tilted (about 45 degrees) and its MOUTH clearly leans toward the heading: in the SE cell the mouth and bipod point to the lower RIGHT, in the NE cell the mouth leans to the upper RIGHT and away (never vertical), in the N cell up and away. Uniforms in saturated, naturally shaded pure MAGENTA (lit about RGB 215,40,200, shaded about RGB 120,10,115), not raspberry. Every pixel outside the painted objects exactly alpha 0, no faint veil or speckle.

LAYOUT CORRECTION TO THE PROVIDED ATLAS: keep exactly the same design, poses, the six orientations, painting and magenta uniforms, but make ALL SIX groups uniformly 25 percent SMALLER around their cell pivots (256,261). Each whole group including boots, bipod feet and the canister stack must fit a 270x270 box centred on (256,261) of its own cell (the provided atlas measures up to 352 px wide and 340 px tall), with at least 115 completely transparent pixels between every extremity and each cell edge. Output an EXACT 1024x1536 RGBA canvas with genuine alpha transparency: alpha exactly 0 outside the six groups, no ground, shadow, haze, glow, backdrop, checkerboard, frame or visible grid. Exactly six clean, disconnected groups, one per cell; no text, numbers, insignia or flags.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the front (where the mouth of the tilted tube leans, with the bipod in front and the baseplate behind) points up, away from the viewer; the rear and top are visible. |
| NW (`nw`) | — | Top-right cell: the front (where the mouth of the tilted tube leans, with the bipod in front and the baseplate behind) points to the upper left; the rear quarter and the flank facing lower left are visible. |
| SW (`sw`) | — | Middle-left cell: the front (where the mouth of the tilted tube leans, with the bipod in front and the baseplate behind) points to the lower left; the front quarter and the flank facing lower right are visible. |
| S (`s`) | — | Middle-right cell: the front (where the mouth of the tilted tube leans, with the bipod in front and the baseplate behind) points down, towards the viewer; the front is visible. |
| SE (`se`) | — | Bottom-left cell: the front (where the mouth of the tilted tube leans, with the bipod in front and the baseplate behind) points to the lower right; the front quarter and the flank facing lower left are visible. |
| NE (`ne`) | — | Bottom-right cell: the front (where the mouth of the tilted tube leans, with the bipod in front and the baseplate behind) points to the upper right; the rear quarter and the flank facing lower right are visible. |

## Post
`art-review -unit art/pipeline/unit-trench-mortar-v1.json` slices the atlas into the six facings with a fixed
anchor and framing inset, splits the magenta into a tint mask and recolours it to the owner palettes
`cp`, `ec` and `neutral` from the manifest, and derives the `spent` state (greyscale) from each tinted
sprite. Exports 2x (192x192) and 4x (384x384) of the 96x96 logical sprite, anchor (48,49);
`unit/trench-mortar/<palette>/<facing>-<state>`. Promote into `assets/units/trench-mortar/` and re-index.

## Avoid
- rotated or mirrored copies of one view
- ground, shadows, haze or backdrop
- magenta on skin, leather, metal, wood or horses
- national insignia, flags or text
- rifles, smoke, firing or explosions
- an upright tube that hides the heading
