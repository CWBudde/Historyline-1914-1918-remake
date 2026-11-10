---
id: units/schneider-ca1
title: Schneider CA1 tank
kind: unit
status: approved
runtime_ids: EC_schneider
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same French WWI tank: a tall box hull on short Holt tracks, a pointed wire-cutter prow leading and a short 75 mm gun in a sponson on the vehicle's right front.
Technical magenta armour is recoloured to the three owner palettes during export.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas. EXACT 1024x1536 transparent RGBA canvas, PORTRAIT orientation, TWO columns THREE rows of 512px square cells, invisible grid. Six independently painted elevated overhead three-quarter views of exactly the same Schneider CA1 vehicle, one per cell, identical design and scale. Soft diffuse light from screen upper LEFT in all six. Pivot at (256,261) in each cell; whole subject inside central 65% with wide fully clear margins, no cutoffs. Strong roof visibility, shallow side walls, consistent camera.
Subject: French Schneider CA1 WWI tank: a tall BOX-like riveted armoured hull sitting on a low Holt tractor track chassis whose short tracks are mostly hidden under the hull skirts; at the FRONT a POINTED, boat-like wire-cutter PROW (a sharp wedge with a steel cutting bar) projecting forward low beyond the tracks; a SHORT stubby 75 mm gun in a small rounded sponson on the hull's RIGHT front side (right as seen by the driver, so on the vehicle's starboard side); one small machine-gun mount on each side further back; a flat roof with a small hatch and a low rectangular commander's hump, two small rear doors and short rear tail hooks at the back. NO turret, NO rhomboid tracks. The pointed prow ALWAYS leads toward the heading; the 75 mm gun is ALWAYS on the vehicle's own right side next to the prow (in the N view on the RIGHT side of the hull, in the S view on the LEFT side of the image). Magenta hull, prow plating and sponson; natural dark steel tracks, cutter bar and gun barrel.
CRITICAL DIRECTION ORDER: top LEFT = NORTH, nose points UP and away, show REAR nearest viewer; top RIGHT = NORTHWEST, nose points upper LEFT and away, show REAR right quarter nearest viewer. Middle LEFT = SOUTHWEST, nose lower LEFT and towards viewer, show FRONT left quarter. Middle RIGHT = SOUTH, nose DOWN towards viewer, show FRONT nearest viewer. Bottom LEFT = SOUTHEAST, nose lower RIGHT towards viewer, show FRONT right quarter. Bottom RIGHT = NORTHEAST, nose upper RIGHT and away, show REAR left quarter nearest viewer. Draw true separate vehicle orientations, never rotated or mirrored flat images.
Painted digital illustration with fine restrained rivet/metal detail, believable WWI silhouette legible at 48px, same style as established game sprites. Saturated naturally shaded magenta ONLY on painted armour for owner tinting. No national insignia, symbols, labels, letters, numbers or flags.
Genuine alpha TRANSPARENCY outside objects and through gaps. Absolutely NO ground, backdrop, contact shadow, glow, vignette, haze, colour wash, checkerboard, frame or visible grid. Six isolated cut-out tanks only; game-ready assets, NOT a presentation illustration.
SIZE AND MARGINS: each individual tank including every track end and gun barrel must fit into a maximum 320x320 box centred on (256,261) in its own 512x512 cell. At least 80 completely transparent pixels between ALL extremities and each of the cell's four edges; nothing crosses x=512 or the row lines y=512 and y=1024. Exactly six clean disconnected tank silhouettes, one per cell, no extra fragments anywhere.
LAYOUT CORRECTION TO THE PROVIDED SCHNEIDER CA1 ATLAS (image 1): keep exactly the same design, colours, six headings, gun side, camera and rendering as image 1, but make ALL SIX tanks uniformly 25 percent SMALLER around their cell centres. Image 1 measured 236-380 px wide and 343-406 px tall, with only 51 px clear above the prow in the N cell; now every tank including the prow tip, gun barrel and tow hooks must fit into a maximum 320x320 box centred on (256,261) in its 512x512 cell, with at least 90 completely transparent pixels to all four cell edges. Exactly six clean disconnected silhouettes. Strictly transparent background.

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
Selected six-view atlas (uca1-02 of gen-2026-11-10, exec-1d330027, 1024x1536). The prompt followed the `art/style.md` Preamble and Units text; image 1 was the
previous try at full size. `art-review -unit art/pipeline/unit-schneider-ca1-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/schneider-ca1/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- terrain, ground shadows, backdrop, baked labels or national insignia
- a turret, rhomboid tracks or the 75 mm gun on the left side
- clipped extremities, mirrored views or changed camera between facings
