---
id: units/se5a
title: S.E.5a
kind: unit
status: revised
runtime_ids: unit/se5a
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI S.E.5a. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Use case: historical-scene. Create ONE newly authored transparent tactical SPRITE ATLAS of the SAME S.E.5a. Exact canvas1024x1536 RGBA, TWO columns by THREE rows of512x512 cells. SIX complete small isolated aircraft, one percell. Whole aircraft including wingtips, propeller and tail must fit inside central300x300box of each512pxcell, centred at(256,261), leaving at least90px genuinely alpha-zero padding on all sides. Do not fill cells or let wings cross cell boundaries. Same scale across six views.
HEADINGS row-major: TOPLEFT N = propeller/nose UP, tail DOWN; TOPRIGHT NW = propeller/nose UPPERLEFT, tail LOWERRIGHT; MIDDLELEFT SW = propeller/nose LOWERLEFT, tail UPPERRIGHT; MIDDLERIGHT S = propeller/nose DOWN, tail UP; BOTTOMLEFT SE = propeller/nose LOWERRIGHT, tail UPPERLEFT; BOTTOMRIGHT NE = propeller/nose UPPERRIGHT, tail LOWERLEFT. Aircraft FRONT is its PROPELLER end, never the tail. Six different independently painted 3Dorientations, never rotate or mirror a finished image.
Subject: Royal Aircraft Factory S.E.5a single-seat BIPLANE fighter. Exactly TWO main wing levels, straight staggered wings with squared clipped tips, long angular fabric fuselage, squared rectangular inline-engine nose with dark radiator grille, ONE open cockpit aft of wing with small headrest and tiny windscreen, ONE conspicuous Lewis gun with pan magazine above the upper wing on central curved Foster mounting, plus ONE small fuselage Vickers gun, long dark exhaust pipes on engine sides. ONE two-bladed natural wood propeller at FRONT, fixed two-wheel landing gear, angular fin. Shaded technical magenta fabric wings/fuselage/tail, natural grey metal nose, wooden propeller and struts, black tyres. No circular rotary cowling, no third wing, no enclosed canopy.
Fixed elevated overhead strategy camera looking mostly down, top surfaces visible with enough side height to distinguish each wing level, aircraft level. Same soft diffuse light from SCREENupperLEFT in every cell. Finely hand-painted WWI game illustration with strong readable48px silhouette, crisp natural antialiased edges and believable fabric, wood and metal. Same wing count, equipment, proportions, object and crew count in every view. No pilot protruding from cockpit. TECHNICALMAGENTA owner-painted fabric retains natural shaded values; all metal, wood and tyres stay natural. TRUE alpha transparency outside aircraft and through open gaps. NO ground, sky, backdrop, haze, glow, shadows, checkerboard, vignette, propeller blur, smoke, clouds, text, letters, numbers, labels, crosses, national insignia, flags, grid or frame. Transparent padding is RGBA(0,0,0,0). Exactly six complete planes and NO detached fragments.

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
Selected atlas. `art-review -unit art/pipeline/unit-se5a-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/se5a/`; gameplay class binding remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
