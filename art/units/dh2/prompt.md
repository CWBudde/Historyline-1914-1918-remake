---
id: units/dh2
title: Airco D.H.2
kind: unit
status: revised
runtime_ids: EC_dh2
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Airco D.H.2. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Use case: historical-scene. Create ONE game-ready TRANSPARENT WWI aircraft sprite atlas of the SAME Airco D.H.2. Exact canvas1024x1536RGBA. TWO columns by THREE rows of invisible512x512square cells, one COMPLETE small plane in each cell. Entire aircraft including ALL wings, propellers, wires, tail and wheels fits in central320x320box centred at(256,261) of each cell; at least80px alpha-zero margins on ALL cell edges. Keep six planes at same scale and prevent ALL boundary crossings.
HEADINGS row-major by AIRCRAFT FRONT NOSE direction: TOPLEFT N = NOSE UP, TAIL DOWN; TOPRIGHT NW = NOSE UPPERLEFT, TAIL LOWERRIGHT, rear quarter nearest viewer; MIDDLELEFT SW = NOSE LOWERLEFT, TAIL UPPERRIGHT, front quarter nearest viewer; MIDDLERIGHT S = NOSE DOWN, TAIL UP; BOTTOMLEFT SE = NOSE LOWERRIGHT, TAIL UPPERLEFT; BOTTOMRIGHT NE = NOSE UPPERRIGHT, TAIL LOWERLEFT, rear quarter nearest viewer. FRONT is the central crew/cockpit NOSE, NOT the rear tail. Pusher propellers sit BEHIND crew, so never define aircraft front by pusher propeller. Six genuinely different independently painted orientations, never rotate/mirror a finished image.
Subject: Airco D.H.2 single-seat PUSHER BIPLANE fighter. Exactly TWO main wing levels with simple wooden interplane struts, a short rounded open-cockpit nacelle in FRONT of the wings, ONE front-mounted forward machine gun, ONE small open pilot seat. ONE exposed natural dark steel ROTARY engine directly BEHIND the pilot, driving ONE two-bladed wooden PUSHER propeller at REAR of nacelle BETWEEN the tail booms. Four thin open tubular tail-boom members (upper/lower pairs on both sides) extend rearward to ONE horizontal tailplane and exactly ONE central oval vertical RUDDER at the far tail. NO twin fins and no box tail. Two spoked landing wheels under nacelle, short tail skid. Magenta fabric wings, nacelle and tail; metal engine and tubes, brown wood prop and struts, black tyres stay natural. No enclosed long fuselage between crew and tail. The FRONT is the blunt crew/gun nose, OPPOSITE the tail; rear propeller is NOT the front.
Fixed elevated overhead strategy camera, visible top surfaces with just enough side height to distinguish wing/tail levels. Aircraft level, same SCREENupperLEFT soft diffuse lighting in every cell. Newly authored finely hand-painted WWI game illustration, crisp natural antialiased edges, strong light/dark silhouette readable at48px, believable fabric ribs, wood and metal, no photo, pixel art or3Drender. Same aircraft identity, EXACT same wing/propeller/crew/equipment counts in all six views. TECHNICALMAGENTA shaded owner fabric paint, natural materials elsewhere. Open cockpits empty or recessed, no protruding pilots. TRUE RGBA alpha transparency outside aircraft AND through all boom/strut gaps. Padding RGBA(0,0,0,0). NO ground, sky, background, shadows, haze, glow, checkerboard, vignette, propeller blur, smoke, text, letters, numbers, labels, flags, crosses, national insignia, grid, arrows or frame. Exactly SIX complete aircraft and NO loose fragments.

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
Selected atlas. `art-review -unit art/pipeline/unit-dh2-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/dh2/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
