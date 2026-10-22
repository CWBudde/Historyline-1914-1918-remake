---
id: units/mobile-aa
title: Mobile anti-aircraft emplacement
kind: unit
status: revised
runtime_ids: CP_mobileAA, EC_mobileAA
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Mobile anti-aircraft emplacement. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Newly authored hand-painted WWI tactical game sprite atlas. Identity reference: supplied portrait ONLY for subject/equipment; do NOT copy background, scene or camera. Exact1024x1536RGBA, two columns x three rows of512pxcells. Six independently painted elevated overhead three-quarter views, fixed camera and upper-left lighting; row-major FRONT headings UP, UPPERLEFT, LOWERLEFT, DOWN, LOWERRIGHT, UPPERRIGHT. Same equipment, crew count, proportions and scale in all6; never rotate or mirror a finished view. Centre each complete group at cell(256,261), fit whole group within central320x320, at least80px genuinely alpha-zero cell margins. Finely painted material detail, crisp silhouette readable48px. Ownership paint/fabric/uniforms shaded technicalMAGENTA; natural wood/steel/skin/leather/rubber. True transparent background including gaps. No ground, earth slab, castshadow, backdrop, halo, haze, writing, insignia, flags, grid, panels or labels. Subject: ONE early open-cab FOUR-wheeled lorry with single high-angle anti-aircraft CANNON mounted on pedestal on rear flatbed, exactlyTWO helmeted crew standing on rear platform beside gun controls. Long bonnet/radiator and two lamps at FRONT, open steering cab, no driver. ONE cannon with single barrel elevated10degrees toward vehicle FRONT; small gun shield. Magenta lorry paint/crewcloth, natural steel cannon/wooden bed. Whole vehicle, gun, four wheels and both crew visible.

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
Selected atlas. `art-review -unit art/pipeline/unit-mobile-aa-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/mobile-aa/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
