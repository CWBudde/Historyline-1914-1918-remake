---
id: terrain/airfield
title: Airfield
kind: terrain
status: candidate
runtime_ids: runway36, runway36_w, hangar, hangar_w, apron-sw, apron-sw_w, apron-nw, apron-nw_w, standing/hangar, standing/hangar_w, standing/shadow/hangar, standing/shadow/hangar_w
masters: airfield-runway-summer-v1.png, airfield-runway-winter-v1.png, airfield-hangar-summer-v1.png, airfield-hangar-winter-v1.png
---

## Look
A 1916 field airfield: a straight north-south runway of rolled, packed earth and worn turf with two
broad wheel-track bands and small whitewashed boundary stones along both edges, and beside it a
Bessonneau-type canvas hangar, its barrel roof on timber ribs and its open doors facing the runway. In
winter the runway is trodden snow with the tracks churned to grey-brown ruts, and snow lies on the hangar
roof. It stands where the original maps have their airstrips (records 105 and 106) and the hangar beside
them (record 51), which the levels keep as a road and a house carrying the infrastructure runway36 and
hangar.

## Base prompt
Opaque horizontal runway strip, strict nadir, packed brown earth with two wheel-track bands and
whitewashed boundary stones on both edges, meadow beyond; and a transparent cut-out of one canvas hangar,
ridge east-west, open doors east, alpha zero outside the hangar, no cast shadow.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| runway | runway36, runway36_w | The strip turned to the vertical N/S axis, summer and winter. |
| apron | hangar, hangar_w, apron-sw, apron-sw_w, apron-nw, apron-nw_w | The apron in front of a hangar's doors, cut from the runway strip, on the hangar hex and the runway hexes at its NE and SE. |
| hangar | standing/hangar, standing/hangar_w, standing/shadow/hangar, standing/shadow/hangar_w | The hangar cut-out, summer and winter, and its cast shadow. |

## Edges & joins
`TestGenerateAirfieldTiles` (`tools/art-review`, tag `campaigngfx`) maps the strip at 750/80 source px per
world px with its centreline on row 502: two spacings of the painted boundary stones make the 80 world px
between two hex centres, so the periodic crossfade meets stone on stone and a chain of runway hexes joins
without a seam. The runway is 48 world px wide with its stones, its edges fading out over 3 world px into
the open ground beneath. The apron, 26 world px wide, runs from the hangar's doors east to the runway and is
split over the three hexes it touches. The hangar is scaled to 56 world px long, centred on its hex, with a
code-made shadow 2.5 world px toward the lower right.

## Post

### Look 8 airfield hangar, winter (2026-10-24)
`airfield-hangar-winter-v1.png` (exec-7ca3940d), native 1536×1024 RGBA: a seasonal edit of the summer
hangar, snow over the whole roof with the ribs, the gable and the dark door still showing. Opaque box about
1079×505 (aspect 2.14, about 1.16× the summer box despite the prompt); alpha haze 1.09 %. Selected: the
export scales each hangar from its own opaque box. Attachment: image 1 the selected summer hangar. The
prompt was:

```text
Seasonal edit ONLY of the attached hangar cut-out (image 1). Keep the exact canvas size 1536x1024, the hangar's position, outline, alpha, ribs and the open east door exactly as they are: the hangar must stay exactly 926 px long and 440 px wide, centred at (767,517) as in image 1 (the previous try enlarged it to 1088x519; do not scale it). Winter, overcast cold light: a continuous layer of settled snow covers the WHOLE canvas roof, both the upper and the lower slope, from eave to eave (near RGB 222,224,228 on the lit upper slope, 190,194,202 on the shaded lower slope), slightly thinner only in a narrow band along both eaves, with the frame ribs and canvas sags still showing through as faint lines; the canvas at the eaves, the gable and the door flaps stays visible, cold faded drab. The door interior stays dark. Alpha zero outside the object exactly as before: no ground plate, no snow on the ground, no halo, no cast shadow.
Avoid: any outline or geometry change, icicles, blue snow, a ground or background, text.

Image 1 is the selected summer hangar: change the season only, preserve every outline, position, size and the transparency.
```

### Look 8 airfield hangar, summer (2026-10-24)
`airfield-hangar-summer-v1.png` (exec-a1bcb53d), native 1536×1024 RGBA: one Bessonneau-type canvas hangar,
barrel roof with five thin ribs, blunt closed west end, open door with folded flaps east. Opaque box 926×440
centred at (767,517), aspect 2.10, about twice as long as the 4:3 footprint asked for (a second try fed the
size back and still painted a long tunnel); alpha haze 0.69 %, no ground plate, no shadow. Selected: the
export keeps its aspect. Attachment: image 1 a house cut-out (style and edge treatment only). The prompt was
the `art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: transparent RGBA cut-out, LANDSCAPE canvas exactly 1536x1024 px. Strict top-down orthographic nadir, painting scale 10 px per world pixel. Exactly ONE First World War Bessonneau-type field aircraft hangar, centred on the canvas: its footprint is exactly 480 px wide (east-west, left-right) by 360 px tall (north-south): nearly square, aspect 4:3, only 31 % of the canvas width and 35 % of its height, leaving wide empty transparent margins (about 528 px left and right, 332 px above and below). The previous try painted it 1162x543 px, 2.4 times too wide and far too long; paint it less than half that length. The roof ridge runs left-right along the middle; the roof is a tensioned canvas cover in faded khaki-drab (lit side near RGB 168,152,112, shaded side near RGB 118,106,78), a rounded, barrel-like arched roof that shades GRADUALLY from the lit upper slope to the shaded lower slope (no hard line at the ridge), divided by 5 thin timber frame ribs that show only as narrow slightly raised seams (about 8 px wide, canvas-coloured, slightly darker), evenly spaced, with the canvas sagging gently between them; guy ropes are not visible. The west end (left) is closed with a blunt, rounded canvas gable (not a pointed triangle); the EAST end (right edge) is the open door: show the dark interior shade in a narrow band at the right end and the two folded-back canvas door flaps as darker drab pleated strips at the upper-right and lower-right corners. Light from the upper left: the upper roof slope is lighter, the lower roof slope darker, form shading only inside the object. Clean straight outline. Alpha zero outside the object, no ground plate, no apron, no grass, no halo, no checkerboard, no cast shadow, no outline stroke.
Avoid: perspective, side view, isometric view, aircraft, people, vehicles, a second building, metal or corrugated roof, brick, windows, chimneys, flags, roundels, text, numbers, grid, frame, photograph, 3D render, pixel art.

Image 1 is a house cut-out painted for this map at the same scale: take only its painting style, detail level and edge treatment; do not copy its building, roof material or colours.
```

### Look 8 airfield runway, winter (2026-10-24)
`airfield-runway-winter-v1.png` (exec-c768c1b2), native 1536×1024 RGB: a seasonal edit of the summer
strip with unchanged geometry: trodden snow, the tracks churned to grey-brown ruts (about RGB 125,115,105),
crown about 201,201,201, meadow about 210,213,218 against the winter meadow's 217,221,225; snow runs across
the runway edge, so the stones mark it. Wrap step 11.4 against about 6 inside. Selected. Attachments: image
1 the selected summer strip, image 2 the winter meadow (snow palette only). The prompt was:

```text
Seasonal edit ONLY of the attached summer airstrip (image 1). Keep the exact canvas size 1536x1024, the centreline, the runway width and its edge positions, the two wheel-track bands, the boundary stones and every outline exactly where they are. Winter, overcast cold light: the runway surface is packed, trodden snow (ivory-grey, near RGB 214,212,204), with the two wheel-track bands churned to grey-brown slush and frozen ruts showing dark earth (near RGB 128,116,100) and faint long tyre marks; the lighter crown is smooth snow with a few dry grass stalks. The boundary stones keep their places and are capped with snow. The verges and the meadow beyond are under settled snow, matching image 2 (near RGB 222,224,226), with only a few dry grass stalks poking through along the verges. Keep it horizontally tileable.
Avoid: any geometry change, moved or added stones, ice sheen, footprints, aircraft, blue snow, green grass, the summer meadow colour, text.

Image 1 is the selected summer airstrip: change the season only, preserve every outline and position. Image 2 is the selected winter meadow: take only its snow palette and texture for the ground.
```

### Look 8 airfield runway, summer (2026-10-24)
`airfield-runway-summer-v1.png` (exec-cd14fd93), native 1536×1024 RGB: a straight earth airstrip across
the full width. Earth band rows 300–705 (405 px) centred on row 502, not 512; wheel tracks about RGB
108,85,54, crown 148,117,73; meadow 86,92,17, a little darker and more saturated than asked. Boundary stones
at x≈205, 580, 955, 1330 in rows 288 and 717, 375 px apart rather than 384, so the export picks its scale
from them. Wrap step 20.6 against 16–18 inside. Selected. Attachments: image 1 the selected road strip
(style, grain and scale only), image 2 the selected meadow (palette). The prompt was the `art/style.md`
Preamble and Terrain text, then:

```text
Opaque 1536x1024 canvas, strict nadir. One straight grass-and-earth military field airstrip of 1916 runs horizontally across the full width, centred on y=512, and the image tiles seamlessly left to right. Scale: the canvas is 144 by 96 world pixels (about 10.7 px per world px); the runway surface is exactly 427 px tall (40 world px, about 42 % of the canvas height), centred, from y=298 to y=725. Runway surface: rolled, packed brown earth and worn-out turf, warm mid brown (average near RGB 128,108,78), fine crumbly soil grain, low contrast; two broad compacted wheel-track bands, each about 110 px wide, centred 105 px above and below the centreline, slightly darker (RGB near 108,92,68) with faint long straight tyre marks along their length; a lighter dry crown about 90 px wide between them (RGB near 146,126,94); a few sparse thin grass tufts surviving on the crown. Edges: the earth ends in a soft, slightly ragged line, no kerb. On both edges sits a row of small whitewashed boundary stones, each about 22 px across, exactly every 384 px (x = 192, 576, 960, 1344), on the runway edge line. Beyond the edges: 50 px of worn short grass verge with bare patches, then warm olive meadow (RGB near 96,104,31) to the top and bottom edges. Diffuse light from the upper left; nothing casts long shadows.
Avoid: asphalt, concrete, tarmac, painted lines or dashes, kerbs, ditches, fences, aircraft, vehicles, people, buildings, trees, posts other than the boundary stones, puddles, craters, perspective, a left/right seam, repeated identical patches, text.

Image 1 is a road strip painted for this map at the same scale and orientation: take only its painting style, grain and scale; do not copy its macadam surface or ditches. Image 2 is the selected meadow material: take its palette and grass texture for the meadow beyond the verges.
```

## Avoid
- asphalt, concrete, painted lines or kerbs on the runway
- aircraft, vehicles, people on either painting
- a ground plate, halo or cast shadow on the hangar cut-out
- perspective, side or isometric views
