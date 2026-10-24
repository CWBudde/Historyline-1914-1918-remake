---
id: units/bunker
title: Bunker
kind: unit
status: candidate
runtime_ids: CP_bunker, EC_bunker
masters: top-view.png, top-view-cp.png, top-view-ec.png
---

## Look
A low round concrete machine-gun bunker seen from straight above, as in the original: a grey dome
with four embrasures and gun barrels, ringed by sandbags on brown earth, with the owner's national
insignia painted large on the roof so the owner reads at a glance. It fills its hex (78 logical
pixels across) and reads at 48 pixels on any terrain.

Exception to `art/style.md`, at the user's request: this unit carries national insignia (the
Entente's French roundel, the Central Powers' cross pattée) instead of a magenta owner strip, and it
is drawn from straight above instead of the three-quarter camera; the Units text block of
`art/style.md` does not apply. The masters are single top views, not six-facing atlases.

## Base prompt
```text
One cut-out game unit sprite master, not an atlas: a single transparent RGBA square canvas, 1024x1024
pixels. Strict top-down orthographic (nadir) view, looking straight down from directly above: no
horizon, no perspective tilt, no three-quarter angle and no visible wall faces. The object is a perfect
circle in plan, never an ellipse.

Subject: one low, round First World War reinforced concrete machine-gun bunker seen from directly
above. Centre it on (512,512). Its outer diameter including the rim is about 680 px (about two thirds
of the canvas width); keep at least 150 px of fully transparent margin on every side.
- Roof: a slightly domed, weathered grey concrete disc about 500 px across, with board-formwork marks,
  small chips, hairline cracks and faint rain streaks. Concrete is a cool natural grey: lit areas near
  RGB 160,160,154, shaded areas near RGB 104,104,100, mean near RGB 132,132,127. The dome is lit from
  the upper left and darker toward the lower right edge (form shading only).
- Centre: a FLAT, CLEAN, EVENLY LIT circular area about 260 px across in the middle of the roof, plain
  smooth concrete with only very faint texture: no hatch, vent, crack, stain or object there. It will
  later carry a painted marking, so leave it empty.
- Four short dark horizontal gun embrasure slits in the roof edge at the top, bottom, left and right
  (12, 3, 6 and 9 o'clock), each about 70 px long and 18 px deep, near-black RGB 30,30,30 inside, with a
  short dark steel machine-gun barrel stub poking out of each slit about 40 px beyond the rim.
- Rim: a ring about 80 px wide around the roof of packed brown earth (RGB near 112,96,72) with a
  single row of tan-khaki sandbags (RGB near 150,132,98) set into it, and a few tufts of dry grass on
  the outer edge. The outer edge of the rim is slightly irregular but stays round overall.
No people, no doors visible from above, no flags, no national insignia, no markings, no text, no
magenta. Alpha zero outside the object, no ground plate, no halo, no checkerboard, no cast shadow,
no glow or backdrop. Strong light/dark separation so it reads at 48 pixels.
```

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| neutral (`top-view.png`) | — | The base prompt above (after the `art/style.md` Preamble). |
| ec (`top-view-ec.png`) | — | Edit of the base, attached as image 1: keep the canvas, transparency and every outline, position, colour and detail; the only change: paint a large French military roundel (cocarde) onto the flat central disc as weathered paint on concrete, outer ring muted dark blue (RGB near 40,62,135), white middle ring (222,220,210), red centre (168,40,38), each ring a third of the radius, outer diameter about 85 percent of the central disc, slight wear and chips; no other marking. |
| cp (`top-view-cp.png`) | — | The same edit with a black cross pattée (the period Iron Cross: four arms narrow at the centre, flaring to wide straight ends, RGB near 24,24,24) with a clean white border (226,223,212) about 14 px wide, arms upright, overall about 85 percent of the central disc; no other insignia, swastika, eagle or text. |

## Post
`go run ./tools/unit-topview -manifest art/pipeline/unit-bunker-v2.json` clears alpha below 24,
centres the base's solid box on the anchor (48,49) and scales it to 78 logical pixels, and resamples
every owner to 2x (192x192) and 4x (384x384). An edit redraws the whole picture (its outline drifted
3-5 px and the roof 6 levels lighter), so only the disc of radius 15.8 logical pixels around the
centre is taken from `top-view-cp.png` and `top-view-ec.png`: registered on the base by the gradient
of the ring around it, matched to the base's concrete luma and blended over 10 px. All owners
therefore share one outline. The same image serves all six facings, so the insignia stays upright;
`spent` is the grey (R,R,R,A) of normal, as for every unit. It writes `assets/units/bunker/` and
re-pins those entries of `assets/index.json`; `-preview build/...` renders the bunker on the map.
Records: `build/review/gen-2026-10-24/bk/` (bk-01 base, bk-02 ec, bk-03 cp).

## Avoid
- perspective, a visible wall or an elliptical plan
- ground, shadows, haze or backdrop
- magenta or tinted concrete
- any marking other than the owner's insignia
- people or text
