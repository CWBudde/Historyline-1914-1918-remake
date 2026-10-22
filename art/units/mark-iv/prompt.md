---
id: units/mark-iv
title: Mark IV tank
kind: unit
status: revised
runtime_ids: unit/mark-iv
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI tank, with a distinct period silhouette.
Technical magenta armour is recoloured to the three owner palettes during export.

## Base prompt
Use case: historical-scene. Production WWI tactical sprite atlas. EXACT 1024x1536 transparent RGBA canvas, TWO columns THREE rows of 512px square cells, invisible grid. Six independently painted elevated overhead three-quarter views of exactly the same Mark IV vehicle, one per cell, identical design and scale. Soft diffuse light from screen upper LEFT in all six. Pivot at (256,261) in each cell; whole subject inside central 65% with wide fully clear margins, no cutoffs. Strong roof visibility, shallow side walls, consistent camera.
Subject: British Mark IV male WWI tank: long RHOMBOID wraparound track assemblies on both sides, high pointed curved track nose, low central roof, two angular side sponsons each with short gun, NO top turret and NO rear steering wheels. Forward raised driver cab and two frontal slit windows identify nose; engine roof grille at rear. Magenta hull and sponsons; natural dark steel tracks and guns.
CRITICAL DIRECTION ORDER: top LEFT = NORTH, nose points UP and away, show REAR nearest viewer; top RIGHT = NORTHWEST, nose points upper LEFT and away, show REAR right quarter nearest viewer. Middle LEFT = SOUTHWEST, nose lower LEFT and towards viewer, show FRONT left quarter. Middle RIGHT = SOUTH, nose DOWN towards viewer, show FRONT nearest viewer. Bottom LEFT = SOUTHEAST, nose lower RIGHT towards viewer, show FRONT right quarter. Bottom RIGHT = NORTHEAST, nose upper RIGHT and away, show REAR left quarter nearest viewer. Draw true separate vehicle orientations, never rotated or mirrored flat images.
Painted digital illustration with fine restrained rivet/metal detail, believable WWI silhouette legible at 48px, same style as established game sprites. Saturated naturally shaded magenta ONLY on painted armour for owner tinting. No national insignia, symbols, labels, letters, numbers or flags.
Genuine alpha TRANSPARENCY outside objects and through gaps. Absolutely NO ground, backdrop, contact shadow, glow, vignette, haze, colour wash, checkerboard, frame or visible grid. Six isolated cut-out tanks only; game-ready assets, NOT a presentation illustration.
LAYOUT CORRECTION TO THE PROVIDED MARK IV ATLAS: keep the exact same design, six headings and rendering, but make ALL SIX tanks uniformly 25 percent SMALLER around their cell centres. Each individual tank including every tread and cannon must fit into a maximum 320x320 box centred in its own 512x512 cell. At least 80 completely transparent pixels between ALL extremities and each cell's four edges. The SW tank in middle-left must have its entire right track inside the LEFT HALF of the canvas: its rightmost opaque pixel must be before x=430, never crossing x=512. The S tank in middle-right must be entirely isolated, with NO tiny stray track fragment on its left. Exactly six clean disconnected tank silhouettes, one per cell. No extra fragments anywhere. Strictly transparent background.

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
Selected six-view atlas. `art-review -unit art/pipeline/unit-mark-iv-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/mark-iv/`; gameplay class binding remains open.

## Avoid
- terrain, ground shadows, backdrop, baked labels or national insignia
- clipped extremities, mirrored views or changed camera between facings
