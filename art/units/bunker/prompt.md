---
id: units/bunker
title: Bunker
kind: unit
status: approved
runtime_ids: CP_bunker, EC_bunker
masters: tilt-view.png, tilt-view-cp.png, tilt-view-ec.png
---

## Look
A low round concrete machine-gun bunker in the slightly tilted camera of the other units (about 55
degrees, the round plan reads as an ellipse): a grey dome with four embrasures and gun barrels, ringed
by sandbags on brown earth, with the owner's national insignia painted large on the flat roof disc so
the owner reads at a glance. It fills its hex (78 logical pixels across) and reads at 48 pixels on any
terrain.

Exception to `art/style.md`, at the user's request: this unit carries national insignia (the
Entente's French roundel, the Central Powers' cross pattée) instead of a magenta owner strip, and the
masters are single views, not six-facing atlases; the object is round and the same image serves every
facing, so the insignia stays upright.

## Base prompt
```text
One cut-out game unit sprite, not an atlas: a single transparent RGBA square canvas, 1024x1024 pixels. Fixed elevated three-quarter overhead strategy camera, about 55 degrees above the ground (a slight tilt: the round plan reads as an ellipse about 0.80 as tall as it is wide, and the front half of the bunker's side wall and sandbag ring shows a little below the roof). Light from the upper left, soft shadows to the lower right on the object itself only.

Subject: one low, round First World War reinforced concrete machine-gun bunker. Centre the whole object on (512,520). Its full width including the earth rim is about 700 px, its height about 600 px; keep at least 140 px of fully transparent margin on every side.
- Roof: a slightly domed, weathered cool-grey concrete cap (lit near RGB 160,160,154, shaded near 104,104,100) with board-formwork marks, small chips, hairline cracks and faint rain streaks, darker toward the lower right. The dome's lower front edge shows a short band of concrete wall above the earth.
- Centre of the roof: a FLAT, CLEAN, EVENLY LIT round disc of plain smooth concrete about 270 px wide (an ellipse in this view), no hatch, vent, crack, stain or object on it. It will later carry a painted marking, so leave it empty.
- Four short dark gun embrasure slits in the roof edge at the back, front, left and right (foreshortened by the camera), near-black inside, each with a short dark steel machine-gun barrel stub poking out about 40 px beyond the rim; the front one points toward the viewer and is shorter, the back one is partly hidden by the dome.
- Ring: packed brown earth (RGB near 112,96,72) about 80 px wide around the base with a single row of tan-khaki sandbags (RGB near 150,132,98) set into it, and a few tufts of dry grass on the outer edge. The ring is round in plan, an ellipse in this view.
No people, no doors, no flags, no national insignia, no markings, no text, no magenta. Alpha zero outside the object, no ground plate, no halo, no checkerboard, no cast shadow, no glow or backdrop. Strong light/dark separation so it reads at 48 pixels.
```
Attached as image 1: the earlier straight-from-above master of the same bunker (design only: "redraw it
from the tilted camera, do not copy its flat nadir view"). The prompt was the `art/style.md` Preamble
text, then this.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| neutral (`tilt-view.png`) | — | The base prompt above. |
| ec (`tilt-view-ec.png`) | — | Edit of the base, attached as image 1: keep the canvas, transparency, camera and every outline, position, colour and detail; the only change: paint a large French military roundel (cocarde) onto the flat central disc, in the disc's perspective, as weathered paint on concrete, outer ring muted dark blue (RGB near 40,62,135), white middle ring (222,220,210), red centre (168,40,38), each ring a third of the radius, overall about 85 percent of the disc, slight wear and chips; no other marking. |
| cp (`tilt-view-cp.png`) | — | The same edit with a black cross pattée (the period Iron Cross: four arms narrow at the centre, flaring to wide straight ends, RGB near 24,24,24) with a clean white border (226,223,212) about 14 px wide, arms pointing up, down, left and right of the screen; no swastika, eagle or other insignia. |

## Post
`go run ./tools/unit-topview -manifest art/pipeline/unit-bunker-v3.json` clears alpha below 24,
centres the base's solid box on the anchor (48,49) and scales it to 78 logical pixels, and resamples
every owner to 2x (192x192) and 4x (384x384). An edit redraws the whole picture (its outline drifted
by a few pixels), so only the roof disc is taken from `tilt-view-cp.png` and `tilt-view-ec.png`: an
ellipse of half-width 15.6 logical pixels, height 0.88 of its width, centred 6.6 logical pixels above
the box centre (`radius`, `squash`, `dy` in the manifest), registered on the base by the gradient of the
ring around it, matched to the base's concrete luma and blended over 10 px. All owners therefore share
one outline. The same image serves all six facings; `spent` is the grey (R,R,R,A) of normal, as for
every unit. It writes `assets/units/bunker/` and re-pins those entries of `assets/index.json`;
`-preview build/...` renders the bunker on the map.
Records: `build/review/gen-2026-10-30/bt/` (bt-01 base, bt-04 cp, bt-03 ec; bt-02, a try without the
attachment, tilted more and hid the back barrel).

## Avoid
- a straight-down view, or a camera tilted so far that the back barrel disappears
- ground, shadows, haze or backdrop
- magenta or tinted concrete
- any marking other than the owner's insignia
- people or text
