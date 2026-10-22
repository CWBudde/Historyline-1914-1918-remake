---
id: units/morane-saulnier-n
title: Morane-Saulnier N
kind: unit
status: revised
runtime_ids: EC_moraneN
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI Morane-Saulnier N. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Use case: historical-scene. Create ONE newly authored transparent tactical SPRITE ATLAS of the SAME Morane-Saulnier N. Exact canvas1024x1536 RGBA, TWO columns by THREE rows of512x512 cells. SIX complete small isolated aircraft, one percell. Whole aircraft including wingtips, propeller and tail must fit inside central300x300box of each512pxcell, centred at(256,261), leaving at least90px genuinely alpha-zero padding on all sides. Do not fill cells or let wings cross cell boundaries. Same scale across six views.
HEADINGS row-major: TOPLEFT N = propeller/nose UP, tail DOWN; TOPRIGHT NW = propeller/nose UPPERLEFT, tail LOWERRIGHT; MIDDLELEFT SW = propeller/nose LOWERLEFT, tail UPPERRIGHT; MIDDLERIGHT S = propeller/nose DOWN, tail UP; BOTTOMLEFT SE = propeller/nose LOWERRIGHT, tail UPPERLEFT; BOTTOMRIGHT NE = propeller/nose UPPERRIGHT, tail LOWERLEFT. Aircraft FRONT is its PROPELLER end, never the tail. Six different independently painted 3Dorientations, never rotate or mirror a finished image.
Subject: Morane-Saulnier Type N single-seat shoulder-wing MONOPLANE. Exactly ONE main wing level, straight shoulder-mounted wings with rounded tips, compact smoothly rounded tapered fabric fuselage, a very large blunt bullet-shaped natural metal nose SPINNER covering most of rotary engine, TWO natural wood propeller blades at FRONT, ONE small open cockpit, small rounded rudder, spoked fixed two-wheel gear, short tailskid, thin kingpost above wings and fine wing-warping bracing wires. Technical shaded magenta fabric wings/fuselage/tail, natural silver nose spinner and brown wood propeller. No second main wing, no interplane struts, no engine sticking out of the nose.
Fixed elevated overhead strategy camera looking mostly down, top surfaces visible with enough side height to distinguish each wing level, aircraft level. Same soft diffuse light from SCREENupperLEFT in every cell. Finely hand-painted WWI game illustration with strong readable48px silhouette, crisp natural antialiased edges and believable fabric, wood and metal. Same wing count, equipment, proportions, object and crew count in every view. No pilot protruding from cockpit. TECHNICALMAGENTA owner-painted fabric retains natural shaded values; all metal, wood and tyres stay natural. TRUE alpha transparency outside aircraft and through open gaps. NO ground, sky, backdrop, haze, glow, shadows, checkerboard, vignette, propeller blur, smoke, clouds, text, letters, numbers, labels, crosses, national insignia, flags, grid or frame. Transparent padding is RGBA(0,0,0,0). Exactly six complete planes and NO detached fragments.
CRITICAL: natural silver ROUND rotary-engine cowling with NO long exposed inline cylinder bank on top. Exactly ONE main wing at shoulder height. Bracing consists of FINE WIRES ONLY and one short kingpost over fuselage; NO rigid V-shaped interplane struts dangling under wings. The big blunt bullet SPINNER encloses the front of the engine.

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
Selected atlas. `art-review -unit art/pipeline/unit-morane-saulnier-n-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/morane-saulnier-n/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
