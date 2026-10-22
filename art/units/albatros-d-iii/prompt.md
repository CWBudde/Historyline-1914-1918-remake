---
id: units/albatros-d-iii
title: Albatros D.III
kind: unit
status: revised
runtime_ids: unit/albatros-d-iii
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Albatros D.III. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Use case: historical-scene. Produce ONE production transparent WWI tactical aircraft sprite ATLAS. Exact canvas 1024x1536 RGBA pixels. Invisible TWO columns by THREE rows of 512x512 square cells. Six SMALL, isolated independently painted views of the SAME Albatros D.III, one aircraft per cell. No duplicated plane or fragments within a cell.
LAYOUT: fuselage centre pivot at each cell's (256,261). ENTIRE aircraft including wingtips, propeller, tail, wheels and wires fits inside the central 320x320 box of each512px cell. At least80px fully clear alpha margins on ALL four edges; wings NEVER cross a cell boundary. Same object scale throughout.
HEADING ORDER, row-major: top LEFT N = propeller/nose UP away from viewer and TAIL down nearest viewer; top RIGHT NW = NOSE upper LEFT, TAIL lower RIGHT; middle LEFT SW = NOSE lower LEFT, TAIL upper RIGHT; middle RIGHT S = NOSE DOWN towards viewer, TAIL UP; bottom LEFT SE = NOSE lower RIGHT, TAIL upper LEFT; bottom RIGHT NE = NOSE upper RIGHT, TAIL lower LEFT. Front is always the PROPELLER end. Draw six genuinely different 3D orientations, never rotate or mirror a finished image.
Subject: An Albatros D.III single-seat WWI sesquiplane fighter: exactly TWO stacked wing levels, a broad straight UPPER wing and clearly NARROWER lower wing, one V-shaped interplane strut on each side, smooth streamlined natural varnished PLYWOOD fuselage, a pointed nose spinner with TWO natural wood propeller blades at the FRONT, visible inline engine cylinder tops, ONE open cockpit behind the upper wing, rounded fin and fixed two-wheel landing gear. Wings and tail fabric use shaded technical MAGENTA; plywood fuselage is warm natural brown, metal/struts/tyres natural. No third wing and no circular rotary-engine cowling.
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
Selected atlas. `art-review -unit art/pipeline/unit-albatros-d-iii-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/albatros-d-iii/`; gameplay class binding remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
