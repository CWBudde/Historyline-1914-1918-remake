---
id: units/whippet
title: Medium Mark A Whippet
kind: unit
status: approved
runtime_ids: EC_whippet
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI medium tank: a low long hull with straight side tracks, the engine deck leading and a tall boxy machine-gun cab at the rear.
Technical magenta armour is recoloured to the three owner palettes during export.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas. EXACT 1024x1536 transparent RGBA canvas, PORTRAIT orientation, TWO columns THREE rows of 512px square cells, invisible grid. Six independently painted elevated overhead three-quarter views of exactly the same Medium Mark A Whippet vehicle, one per cell, identical design and scale. Soft diffuse light from screen upper LEFT in all six. Pivot at (256,261) in each cell; whole subject inside central 65% with wide fully clear margins, no cutoffs. Strong roof visibility, shallow side walls, consistent camera.
Subject: British Medium Mark A "Whippet" WWI tank: a LOW, LONG, narrow hull with low track frames running the full length along both sides (straight low tracks with a slightly raised front end, NOT rhomboid, NOT wraparound over the roof). At the FRONT (nose) a long low engine compartment with a louvred grille and a rounded armoured fuel tank at the very front tip. At the REAR a TALL fixed boxy fighting compartment (an angular armoured cab, clearly the highest part, rising well above the tracks) with small machine-gun ports and short MG barrels in its faces. NO turret, NO sponsons. The tall cab is ALWAYS at the rear, opposite the heading; the low engine deck always leads toward the heading. Magenta hull, engine covers and cab; natural dark steel tracks and MG barrels.
CRITICAL DIRECTION ORDER: top LEFT = NORTH, nose points UP and away, show REAR nearest viewer; top RIGHT = NORTHWEST, nose points upper LEFT and away, show REAR right quarter nearest viewer. Middle LEFT = SOUTHWEST, nose lower LEFT and towards viewer, show FRONT left quarter. Middle RIGHT = SOUTH, nose DOWN towards viewer, show FRONT nearest viewer. Bottom LEFT = SOUTHEAST, nose lower RIGHT towards viewer, show FRONT right quarter. Bottom RIGHT = NORTHEAST, nose upper RIGHT and away, show REAR left quarter nearest viewer. Draw true separate vehicle orientations, never rotated or mirrored flat images.
Painted digital illustration with fine restrained rivet/metal detail, believable WWI silhouette legible at 48px, same style as established game sprites. Saturated naturally shaded magenta ONLY on painted armour for owner tinting. No national insignia, symbols, labels, letters, numbers or flags.
Genuine alpha TRANSPARENCY outside objects and through gaps. Absolutely NO ground, backdrop, contact shadow, glow, vignette, haze, colour wash, checkerboard, frame or visible grid. Six isolated cut-out tanks only; game-ready assets, NOT a presentation illustration.
SIZE AND MARGINS: each individual tank including every track end and gun barrel must fit into a maximum 320x320 box centred on (256,261) in its own 512x512 cell. At least 80 completely transparent pixels between ALL extremities and each of the cell's four edges; nothing crosses x=512 or the row lines y=512 and y=1024. Exactly six clean disconnected tank silhouettes, one per cell, no extra fragments anywhere.
LAYOUT CORRECTION TO THE PROVIDED Whippet ATLAS (image 1): keep exactly the same design, colours, six headings, camera and rendering as image 1, but make ALL SIX tanks uniformly 25 percent SMALLER around their cell centres. Image 1 measured 352-408 px wide and 337-399 px tall, with only 41 px clear to the cell edge in the NW cell; now every tank including every track end and gun barrel must fit into a maximum 320x320 box centred on (256,261) in its 512x512 cell, with at least 90 completely transparent pixels to all four cell edges. Exactly six clean disconnected silhouettes. Strictly transparent background.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N | — | Independently painted N view; common camera and upper-left lighting. |
| NW | — | Independently painted NW view; common camera and upper-left lighting. |
| SW | — | Independently painted SW view; common camera and upper-left lighting. |
| S | — | Independently painted S view; common camera and upper-left lighting. |
| SE | — | Independently painted SE view; common camera and upper-left lighting. |
| NE | — | Independently painted NE view; common camera and upper-left lighting. |

## Post
Selected six-view atlas (uwhip-02 of gen-2026-11-10, exec-4d1f5e54, 1024x1536). The prompt followed the `art/style.md` Preamble and Units text; image 1 was the
previous try at full size. `art-review -unit art/pipeline/unit-whippet-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/whippet/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- terrain, ground shadows, backdrop, baked labels or national insignia
- clipped extremities, mirrored views or changed camera between facings
