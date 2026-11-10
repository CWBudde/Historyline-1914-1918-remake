---
id: units/mark-v
title: Mark V tank
kind: unit
status: approved
runtime_ids: EC_markV
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI heavy tank: Mark IV-like rhomboid tracks and sponsons, told apart by the raised rear cupola, the unditching-beam rails along the track tops and the timber beam at the rear.
Technical magenta armour is recoloured to the three owner palettes during export.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas. EXACT 1024x1536 transparent RGBA canvas, PORTRAIT orientation, TWO columns THREE rows of 512px square cells, invisible grid. Six independently painted elevated overhead three-quarter views of exactly the same Mark V vehicle, one per cell, identical design and scale. Soft diffuse light from screen upper LEFT in all six. Pivot at (256,261) in each cell; whole subject inside central 65% with wide fully clear margins, no cutoffs. Strong roof visibility, shallow side walls, consistent camera.
Subject: British Mark V male heavy WWI tank: long RHOMBOID wraparound track assemblies on both sides, high pointed curved track noses, two angular side sponsons each with a short 6-pounder gun, NO turret and NO rear steering wheels. Forward driver cab with two frontal slit windows identifies the nose. Distinct Mark V features: a RAISED boxy observation CUPOLA/cab on the roof near the REAR, clearly standing up above the roof line, and TWO long parallel UNDITCHING-BEAM RAILS running lengthwise along the top of both track frames from front to back, with a heavy dark wooden unditching beam (a square timber) lying across the rails at the rear. Rear: the raised cupola and beam always opposite the nose. Magenta hull, sponsons, cab and cupola; natural dark steel tracks, rails and guns, natural dark timber beam.
CRITICAL DIRECTION ORDER: top LEFT = NORTH, nose points UP and away, show REAR nearest viewer; top RIGHT = NORTHWEST, nose points upper LEFT and away, show REAR right quarter nearest viewer. Middle LEFT = SOUTHWEST, nose lower LEFT and towards viewer, show FRONT left quarter. Middle RIGHT = SOUTH, nose DOWN towards viewer, show FRONT nearest viewer. Bottom LEFT = SOUTHEAST, nose lower RIGHT towards viewer, show FRONT right quarter. Bottom RIGHT = NORTHEAST, nose upper RIGHT and away, show REAR left quarter nearest viewer. Draw true separate vehicle orientations, never rotated or mirrored flat images.
Painted digital illustration with fine restrained rivet/metal detail, believable WWI silhouette legible at 48px, same style as established game sprites. Saturated naturally shaded magenta ONLY on painted armour for owner tinting. No national insignia, symbols, labels, letters, numbers or flags.
Genuine alpha TRANSPARENCY outside objects and through gaps. Absolutely NO ground, backdrop, contact shadow, glow, vignette, haze, colour wash, checkerboard, frame or visible grid. Six isolated cut-out tanks only; game-ready assets, NOT a presentation illustration.
SIZE AND MARGINS: each individual tank including every track end and gun barrel must fit into a maximum 320x320 box centred on (256,261) in its own 512x512 cell. At least 80 completely transparent pixels between ALL extremities and each of the cell's four edges; nothing crosses x=512 or the row lines y=512 and y=1024. Exactly six clean disconnected tank silhouettes, one per cell, no extra fragments anywhere.
LAYOUT CORRECTION TO THE PROVIDED Mark V ATLAS (image 1): keep exactly the same design, colours, six headings, camera and rendering as image 1, but make ALL SIX tanks uniformly 25 percent SMALLER around their cell centres. Image 1 measured 394-413 px wide and 335-355 px tall, with only 46 px clear to the cell edge in the NW cell; now every tank including every track end and gun barrel must fit into a maximum 320x320 box centred on (256,261) in its 512x512 cell, with at least 90 completely transparent pixels to all four cell edges. Exactly six clean disconnected silhouettes. Strictly transparent background.

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
Selected six-view atlas (umk5-02 of gen-2026-11-10, exec-6bb3b847, 1024x1536). The prompt followed the `art/style.md` Preamble and Units text; image 1 was the
previous try at full size. `art-review -unit art/pipeline/unit-mark-v-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/mark-v/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- terrain, ground shadows, backdrop, baked labels or national insignia
- a plain Mark IV without the rear cupola, the rails or the beam
- clipped extremities, mirrored views or changed camera between facings
