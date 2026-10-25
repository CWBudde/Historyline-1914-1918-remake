---
id: units/static-balloon
title: Static balloon
kind: unit
status: approved
runtime_ids: CP_balloon, EC_balloon
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Static balloon. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Transparent WWI strategy sprite atlas,1024x1536RGBA,2columns x3rows of512pxcells. SIX independently painted views of SAME subject, elevated overhead three-quarter camera and upper-left lighting, natural hand-painted detailed game illustration, strong48pxsilhouette. Same equipment,crew,scale. Fit COMPLETE group in central300x300 of everycell with100px transparent margins. ROW-MAJOR subject FRONT/aim directions, not camera directions: cell1 FRONT UP (back of figures nearestviewer); cell2 FRONT UPPERLEFT (back/rightquarter nearestviewer); cell3 FRONT LOWERLEFT (front/rightquarter nearestviewer); cell4 FRONT DOWN (front of figures nearestviewer); cell5 FRONT LOWERRIGHT (front/leftquarter nearestviewer); cell6 FRONT UPPERRIGHT (back/leftquarter nearestviewer). No mirror/rotate of finishedviews. Ownershipcloth or envelope technicalMAGENTA, natural metal/wood/skin/leather elsewhere. True alpha-zero exterior AND gaps. No background, ground, shadows, haze, glow, writing, flags or insignia. ONE elongated cigar-shaped tethered OBSERVATION BALLOON with rounded FRONT nose and three stabilizing fabric fins at rear, small wooden basket suspended beneath by fourriggingropes and exactlyONE observer in basket. ONE taut tether descends to ONE compact groundwinch on smalltwo-wheeled frame, exactlyTWO small groundcrew adjacent. No earthpatch. Balloon envelope magenta with seams, tailfinsmagenta, ropes/wood/naturalmetal/skin remainnatural; crewuniformsmagenta. The round NOSE points atheading and finnedTAIL opposite. Cell1 roundnoseUP, finnedtailDOWN. Cell4 roundnoseDOWN,finnedtailUP. Keep inflatedenvelope level horizontally in world,never uprightlikehotairballoon. Draw all6distinctviews, no duplicate direction.

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
Selected atlas. `art-review -unit art/pipeline/unit-static-balloon-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/static-balloon/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
