---
id: units/renault-ft17
title: Renault FT17 tank
kind: unit
status: revised
runtime_ids: CP_ft17, EC_ft17
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI tank, with a distinct period silhouette.
Technical magenta armour is recoloured to the three owner palettes during export.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas. EXACT 1024x1536 transparent RGBA canvas, TWO columns THREE rows of 512px square cells, invisible grid. Six independently painted elevated overhead three-quarter views of exactly the same Renault FT17 vehicle, one per cell, identical design and scale. Soft diffuse light from screen upper LEFT in all six. Pivot at (256,261) in each cell; whole subject inside central 65% with wide fully clear margins, no cutoffs. Strong roof visibility, shallow side walls, consistent camera.
Subject: Renault FT17 light tank: short narrow hull, large exposed tall triangular track assemblies along both sides, small round one-man turret near the FRONT carrying one short cannon, long spade-shaped rear trench-crossing tail. Rear tail ALWAYS opposite nose/cannon. Magenta hull and turret; natural dark steel tracks, gun and rear tail.
CRITICAL DIRECTION ORDER: top LEFT = NORTH, nose points UP and away, show REAR nearest viewer; top RIGHT = NORTHWEST, nose points upper LEFT and away, show REAR right quarter nearest viewer. Middle LEFT = SOUTHWEST, nose lower LEFT and towards viewer, show FRONT left quarter. Middle RIGHT = SOUTH, nose DOWN towards viewer, show FRONT nearest viewer. Bottom LEFT = SOUTHEAST, nose lower RIGHT towards viewer, show FRONT right quarter. Bottom RIGHT = NORTHEAST, nose upper RIGHT and away, show REAR left quarter nearest viewer. Draw true separate vehicle orientations, never rotated or mirrored flat images.
Painted digital illustration with fine restrained rivet/metal detail, believable WWI silhouette legible at 48px, same style as established game sprites. Saturated naturally shaded magenta ONLY on painted armour for owner tinting. No national insignia, symbols, labels, letters, numbers or flags.
Genuine alpha TRANSPARENCY outside objects and through gaps. Absolutely NO ground, backdrop, contact shadow, glow, vignette, haze, colour wash, checkerboard, frame or visible grid. Six isolated cut-out tanks only; game-ready assets, NOT a presentation illustration.

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
Selected six-view atlas. `art-review -unit art/pipeline/unit-renault-ft17-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/renault-ft17/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- terrain, ground shadows, backdrop, baked labels or national insignia
- clipped extremities, mirrored views or changed camera between facings
