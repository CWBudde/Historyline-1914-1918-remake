---
id: units/spad-xiii
title: SPAD XIII
kind: unit
status: revised
runtime_ids: EC_spad13
masters: atlas.png
---

## Look
Six elevated overhead painted views of the same WWI SPAD XIII. A distinct period silhouette,
fixed upper-left lighting and technical magenta owner paint; other materials stay natural.

## Base prompt
Use case: historical-scene. Create ONE newly authored transparent tactical SPRITE ATLAS of the SAME SPAD XIII. Exact canvas1024x1536 RGBA, TWO columns by THREE rows of512x512 cells. SIX complete small isolated aircraft, one percell. Whole aircraft including wingtips, propeller and tail must fit inside central300x300box of each512pxcell, centred at(256,261), leaving at least90px genuinely alpha-zero padding on all sides. Do not fill cells or let wings cross cell boundaries. Same scale across six views.
HEADINGS row-major: TOPLEFT N = propeller/nose UP, tail DOWN; TOPRIGHT NW = propeller/nose UPPERLEFT, tail LOWERRIGHT; MIDDLELEFT SW = propeller/nose LOWERLEFT, tail UPPERRIGHT; MIDDLERIGHT S = propeller/nose DOWN, tail UP; BOTTOMLEFT SE = propeller/nose LOWERRIGHT, tail UPPERLEFT; BOTTOMRIGHT NE = propeller/nose UPPERRIGHT, tail LOWERLEFT. Aircraft FRONT is its PROPELLER end, never the tail. Six different independently painted 3Dorientations, never rotate or mirror a finished image.
Subject: SPAD XIII single-seat BIPLANE fighter. Exactly TWO main wing levels, broad strong straight staggered wings and paired wooden interplane struts. Slightly larger and longer than SPADVII, sturdy tapered fuselage, ROUND honeycomb radiator face at FRONT with small conical central spinner/hub and TWO wood propeller blades, TWO parallel forward machine guns on the top of nose, ONE open cockpit aft of wing, long natural dark exhaust pipes along both sides of engine cowling, rounded tail fin and fixed spoked two-wheel gear. Water-cooled V8 engine enclosed INSIDE cowling; NO exposed radial/rotary cylinder ring. Shaded technical magenta wings/fuselage/tail, natural dull metal nose, natural wood propeller/struts and black tyres. No third wing, no second cockpit.
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
Selected atlas. `art-review -unit art/pipeline/unit-spad-xiii-v1.json` produces owner cp/ec/neutral,
normal/spent, and 2×/4× RGBA exports with logical anchor (48,49). No per-facing rotation, mirror or resize.
Installed under `assets/units/spad-xiii/`. Bound classes are listed above; final visual acceptance remains open.

## Avoid
- water, wakes, reflections, ground shadows, backdrop or detached fragments
- labels, national insignia, clipped extremities or inconsistent equipment between views
