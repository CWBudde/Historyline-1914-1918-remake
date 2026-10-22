---
id: units/torpedo-boat
title: Torpedo boat
kind: unit
status: revised
runtime_ids: CP_torpedoBoat, EC_torpedoBoat
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Torpedo boat. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Use case: historical-scene. Create one game-ready transparent tactical SPRITE ATLAS, showing the SAME Torpedo boat in SIX independently painted directions. Exact canvas 1024x1536 pixels, RGBA. TWO columns and THREE rows of invisible 512x512 square cells. Six small isolated ship icons, ONE per cell, absolutely no extra ship, fragment or diagram.
Strict layout: the entire ship including guns, mast and rigging fits within a 320x320 box centred at (256,261) of its cell, leaving AT LEAST 80px of genuinely clear alpha on ALL four sides. Do not enlarge the ships to fill their cells! Keep each silhouette away from the cell boundary. Same common scale within all views, no crop, no cutoff.
ORDER: TOP LEFT N = pointed BOW UP away from viewer, show STERN nearest viewer. TOP RIGHT NW = BOW UPPER LEFT away, show STERN right quarter nearest viewer. MIDDLE LEFT SW = BOW LOWER LEFT towards viewer, show FRONT left quarter. MIDDLE RIGHT S = BOW DOWN towards viewer, show FRONT nearest viewer. BOTTOM LEFT SE = BOW LOWER RIGHT towards viewer, show FRONT right quarter. BOTTOM RIGHT NE = BOW UPPER RIGHT away, show STERN left quarter nearest viewer. Bow and stern must not be confused. Rearrange the actual 3D ship orientation for each independently painted view, never rotate or mirror a finished image.
Subject: One slender low WWI torpedo boat with a very narrow pointed bow and rounded flat stern; two small slim funnels, a tiny raised forward wheelhouse, ONE light gun at the bow and two clearly visible deck torpedo tubes behind the funnels. No large turrets. Low freeboard and a long thin hull. Magenta hull sides and wheelhouse; muted natural wood/grey deck, dark steel funnels, tubes and gun.
Camera is elevated three-quarter overhead strategy view, looking mostly down, roofs and deck clearly visible, shallow side walls, NO underside or keel. Same camera and short diffuse daylight from screen upper LEFT in ALL six views. Newly authored finely painted digital illustration matching existing WWI tactical sprites, crisp natural antialiased edges, strong readable silhouette at 48px, neither pixel art, photograph nor 3D render. No tiny distracting excessive rigging.
TRUE alpha TRANSPARENCY outside the ship and through gaps. Waterline cutout only: no surrounding water, wake, foam, reflection, waves, ground, shadow, background, haze, glow, checkerboard or vignette. No words, letters, numbers, labels, flags, national insignia, grid, arrows or frame. Technical magenta is ONLY for owner-painted parts and preserves natural shading; other materials remain natural. No black background.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N | — | Independently painted N view; common camera, scale and lighting. |
| NW | — | Independently painted NW view; common camera, scale and lighting. |
| SW | — | Independently painted SW view; common camera, scale and lighting. |
| S | — | Independently painted S view; common camera, scale and lighting. |
| SE | — | Independently painted SE view; common camera, scale and lighting. |
| NE | — | Independently painted NE view; common camera, scale and lighting. |

## Post
Selected atlas. `art-review -unit art/pipeline/unit-torpedo-boat-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/torpedo-boat/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
