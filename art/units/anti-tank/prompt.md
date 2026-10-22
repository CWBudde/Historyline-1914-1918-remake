---
id: units/anti-tank
title: Anti-tank weapons
kind: unit
status: revised
runtime_ids: CP_antiTank, EC_antiTank
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Anti-tank weapons. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Transparent WWI strategy sprite atlas,1024x1536RGBA,2columns x3rows of512pxcells. SIX independently painted views of SAME subject, elevated overhead three-quarter camera and upper-left lighting, natural hand-painted detailed game illustration, strong48pxsilhouette. Same equipment,crew,scale. Fit COMPLETE group in central300x300 of everycell with100px transparent margins. ROW-MAJOR subject FRONT/aim directions, not camera directions: cell1 FRONT UP (back of figures nearestviewer); cell2 FRONT UPPERLEFT (back/rightquarter nearestviewer); cell3 FRONT LOWERLEFT (front/rightquarter nearestviewer); cell4 FRONT DOWN (front of figures nearestviewer); cell5 FRONT LOWERRIGHT (front/leftquarter nearestviewer); cell6 FRONT UPPERRIGHT (back/leftquarter nearestviewer). No mirror/rotate of finishedviews. Ownershipcloth or envelope technicalMAGENTA, natural metal/wood/skin/leather elsewhere. True alpha-zero exterior AND gaps. No background, ground, shadows, haze, glow, writing, flags or insignia. ExactlyTWO infantry: ONE prone rifleman aiming ONE unusually long, heavy WWI BOLT-ACTION ANTI-TANK RIFLE supported on bipod; ONE crouched spotter alongside using binoculars. Both wear magenta uniforms and plain natural steel helmets. One long slim steel barrel, wooden stock and small simple muzzle block, NO cannon, shield, extra weapons or tripod. Rifle points at heading, with muzzle at FRONT of group, stock/bothmen toward REAR. Cell1 muzzleTOP,menBOTTOM; cell4 muzzleBOTTOM,menTOP. Preserve clear continuous rifle and hands contacting woodstock.

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
Selected atlas. `art-review -unit art/pipeline/unit-anti-tank-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/anti-tank/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
