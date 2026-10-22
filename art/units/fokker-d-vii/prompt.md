---
id: units/fokker-d-vii
title: Fokker D.VII
kind: unit
status: revised
runtime_ids: CP_fokkerDVII
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Fokker D.VII. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Use case: historical-scene. Produce ONE production transparent WWI tactical aircraft sprite ATLAS. Exact canvas 1024x1536 RGBA pixels. Invisible TWO columns by THREE rows of 512x512 square cells. Six SMALL, isolated independently painted views of the SAME Fokker D.VII, one aircraft per cell. No duplicated plane or fragments within a cell.
LAYOUT: fuselage centre pivot at each cell's (256,261). ENTIRE aircraft including wingtips, propeller, tail, wheels and wires fits inside the central 320x320 box of each512px cell. At least80px fully clear alpha margins on ALL four edges; wings NEVER cross a cell boundary. Same object scale throughout.
HEADING ORDER, row-major: top LEFT N = propeller/nose UP away from viewer and TAIL down nearest viewer; top RIGHT NW = NOSE upper LEFT, TAIL lower RIGHT; middle LEFT SW = NOSE lower LEFT, TAIL upper RIGHT; middle RIGHT S = NOSE DOWN towards viewer, TAIL UP; bottom LEFT SE = NOSE lower RIGHT, TAIL upper LEFT; bottom RIGHT NE = NOSE upper RIGHT, TAIL lower LEFT. Front is always the PROPELLER end. Draw six genuinely different 3D orientations, never rotate or mirror a finished image.
Subject: A Fokker D.VII WWI BIPLANE fighter, exactly TWO main wing levels: broad thick nearly straight upper wing and smaller lower wing, joined by N-shaped interplane struts. Short boxy fabric fuselage with rectangular inline-engine nose and top radiator, ONE small open cockpit aft of the top wing, two forward guns, broad rounded rudder, fixed two-wheel undercarriage with axle airfoil, ONE two-bladed natural wooden tractor propeller at the FRONT. Shaded technical magenta wings, fuselage and tail, natural steel engine panels, wooden propeller and black tyres. NO circular rotary cowling, no third main wing and no modern canopy.
Camera: fixed elevated overhead strategy camera, aircraft level and wings horizontal, visible top surfaces with just enough side height to distinguish wing levels. Light always from SCREEN upper LEFT, soft and diffuse. Newly authored finely hand-painted digital game illustration, crisp clean material edges and strong readable silhouette at48px; same painterly WWI tactical sprite vocabulary as existing units. Not pixel art, not photo, not a3D render. All six views have identical wing count, equipment and structural design. No pilot protruding above fuselage.
TRUE ALPHA transparency outside each aircraft and through open structural gaps. NO ground, sky, backdrop, clouds, shadows, haze, glow, checkerboard, propeller blur, smoke, reflections, labels, insignia, crosses, flags, letters, numbers, grid or frame. Exactly SIX complete aircraft, one per cell.

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
Selected atlas. `art-review -unit art/pipeline/unit-fokker-d-vii-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/fokker-d-vii/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
