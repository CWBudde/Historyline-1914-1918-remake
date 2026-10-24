---
id: terrain/harbour
title: Harbour
kind: terrain
status: candidate
runtime_ids: harbourN, harbourN_w, harbourW, harbourW_w, harbourS, harbourS_w, standing/harbourN, standing/harbourN_w, standing/harbourW, standing/harbourW_w, standing/harbourS, standing/harbourS_w, standing/shadow/harbourN, standing/shadow/harbourN_w, standing/shadow/harbourW, standing/shadow/harbourW_w, standing/shadow/harbourS, standing/shadow/harbourS_w
masters: harbour-kit-summer-v1.png, harbour-kit-winter-v1.png, harbour-paving-summer-v1.png, harbour-paving-winter-v1.png
---

## Look
A small harbour of 1910 on three hexes paved with grey granite setts: a long brick quayside warehouse with
a slate roof and four hoist dormers, its ridge north-south; at its SW a small stone customs house with a red
clay tile hipped roof, two chimneys and an open timber coal bunker; at its S the quay yard with a hand crane
on a granite base, its jib reaching east over the water, stacked crates, barrels and jute sacks, where the
road comes in from the south. In winter snow lies on the roofs and the cargo, and old trodden snow with slush
patches on the setts. It stands where the original maps have their harbour (records 53, 52 and 120), which
the levels keep as two houses carrying harbourN and harbourW and a road carrying harbourS.

## Base prompt
A transparent cut-out sheet in three columns, strict nadir, no ground, no cast shadow: the warehouse, the
yard's crane and cargo, the customs house; and an opaque, even granite sett paving of a quay yard.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| paving | harbourN, harbourN_w, harbourW, harbourW_w, harbourS, harbourS_w | The paving under each of the three hexes, summer and winter. |
| buildings | standing/harbourN, standing/harbourN_w, standing/harbourW, standing/harbourW_w, standing/harbourS, standing/harbourS_w | The warehouse, the customs house and the yard's crane and cargo, summer and winter. |
| shadows | standing/shadow/harbourN, standing/shadow/harbourN_w, standing/shadow/harbourW, standing/shadow/harbourW_w, standing/shadow/harbourS, standing/shadow/harbourS_w | Their cast shadows, made in code. |

## Edges & joins
`TestGenerateHarbourTiles` (`tools/art-review`, tag `campaigngfx`) samples the paving at 6 source px per
world px in the group's frame (the warehouse's centre, the customs house 71 world px W and 40 S of it, the
yard 80 S), so the setts run on over the edges the three hexes share. Along the outer edges the paving fades
out over 5 world px behind a margin of 2 to 6 world px that wanders (smooth noise over 10 world px), with
the corners rounded; at the yard's S edge it leaves the road uncovered for 8 world px and fades in over it
slowly, so the road runs into the yard. The kit's pieces are the opaque shapes whose centre lies in a
column; each is scaled from its own box: the warehouse 24 world px wide (about 60 long) at 3 world px N of
its hex centre, the customs house 32 wide at its centre, the yard's cargo 40 wide at 17 E and 6 N of the
yard's centre, the crane jib reaching over the NE edge. Each casts a code-made shadow 2.5 world px toward
the lower right; a piece and its shadow are cut to the logical pixels they cover, with the anchor moved to
match, which keeps the standing pack within its decoded budget.

## Post

### Look 8 C4 harbour paving, winter (2026-10-24)
`harbour-paving-winter-v1.png` (exec-becfe6d5), native 1536×1024 RGB: a seasonal edit of the summer paving
with soft snow drifts 150–400 px across over many setts, snow in the joints and grey-brown slush patches;
mean RGB 172,172,177, L\* 70, even over the canvas. The setts came out redrawn smaller (about 30×22 px).
Selected; the first try (exec-0085cea7) whitened single setts into a checkerboard. Attachments: image 1 the
selected summer paving, image 2 `art/terrain/grass/grass-winter-v1.png` (snow palette only). The prompt was:

```text
Seasonal edit ONLY of the attached granite sett paving (image 1). Keep the exact canvas size 1536x1024, every sett, joint and course exactly where it is, and keep the image opaque and evenly dense. Winter, overcast cold light: old trodden snow lies over the paving. White snow (near RGB 214,216,220, matching image 2) fills every joint and lies in soft, irregular drifts and patches 150 to 400 px across that each cover many setts at once, their edges soft and feathered; between the patches the setts show through as cold wet grey stone (near RGB 128,126,124) dusted with snow. The previous try whitened single setts one by one into a checkerboard: do not do that; snow ignores the sett edges and lies in large soft shapes. A few scattered dark grey-brown slush patches and short wandering trodden trails of slush (near RGB 118,110,102), each no longer than 300 px and running in different directions. Even distribution over the whole canvas, no focal point, no bright or dark band along any edge.
Avoid: a checkerboard of white and grey setts, any change to the setts' size or layout, long straight tracks across the canvas, ice sheen, footprints, blue snow, people, objects, text.

Image 1 is the selected summer paving: change the season only, preserve every sett, joint and position. Image 2 is the selected winter meadow: take only its snow palette and texture for the snow.
```

### Look 8 C4 harbour paving, summer (2026-10-24)
`harbour-paving-summer-v1.png` (exec-5c7418a6), native 1536×1024 RGB: grey-beige granite setts in staggered
courses, mean RGB 143,129,110, L\* 54.5, σL\* 11.7, even over the canvas. Setts about 38×26 px (6×4 world px
at the export's 6 px per world px), still about 1.8× the 21×18 px asked for; four tries painted 88×70, 55×42,
38×26 and 45×38 px, so this third try is selected. Not tileable, which the export does not need. Attachment:
image 1 `art/terrain/quay/quay-pier-shared-v1.png` (sett colour and finish only). The prompt was the
`art/style.md` Preamble and Terrain text, then:

```text
Use case: stylized-concept. Asset type: opaque full-bleed terrain material, LANDSCAPE canvas exactly 1536x1024 px, strict top-down orthographic nadir. Scale: the canvas is 256 x 171 world px at 6 px per world px (one map hex is 96 world px across). The paved ground of a small harbour quay yard of 1910: grey granite setts laid in straight courses running left to right with staggered joints, each sett about 21 px long and 18 px wide (3.5 by 3 world px), so about 70 setts across the canvas width and 57 courses from top to bottom (the previous try still painted every sett about 55x42 px, 2.6 times too large, only 27 setts across: paint them less than half that size, 21x18 px, a fine dense texture of small setts like a mosaic seen from far above); dark gritty joints about 2 px wide; setts in mixed warm grey-beige tones (average near RGB 132,124,110, single setts between RGB 105,98,88 and 165,156,140), some sunk, chipped or worn smooth; faint scattered coal dust, a few rust and oil stains up to 60 px across, thin weeds in some joints. Even density and even brightness over the whole canvas, also along the edges (no darker bottom rows), no focal point, no larger feature. Flat overcast light, no light direction, whatever the style notes above say; no cast shadows.
Avoid: kerbs, quay edges, bollards, rails, cart tracks, painted lines, drains, puddles, water, grass areas, crates, people, buildings, vignette, a bright centre, perspective, photograph, text.

Image 1 is the selected quay pier of this map at a larger scale: take only the granite colour, grain and painted finish of its setts; do not copy the pier, its kerbs, bollards, timber, cart tracks or its sett size.
```

### Look 8 C4 harbour kit, winter (2026-10-24)
`harbour-kit-winter-v1.png` (exec-041e9e4e), native 1536×1024 RGBA: a seasonal edit of the summer kit with
settled snow on both roofs, snow caps on the dormer hoods, chimneys, crates, barrels, sacks and the crane
base; every piece's box within 4 px of the summer one. The crane's iron came out light grey rather than
dark. Selected. Attachment: image 1 the selected summer kit. The prompt was:

```text
Seasonal edit ONLY of the attached harbour cut-out sheet (image 1). Keep the exact canvas size 1536x1024 and every object exactly where it is, with the same outline, size and transparency: the long warehouse in the left column, the crane, crates, barrels and sacks in the middle column, the customs house with its coal bunker in the right column; do not move, scale or redraw any of them. Winter, overcast cold light: a continuous layer of settled snow covers both slopes of the warehouse's slate roof and all four planes of the customs house's tile roof (near RGB 222,224,228 on the lit left slopes, 190,194,202 on the shaded right slopes), thinner along the eaves and the ridges, so the slates and tiles and the dark lead ridge still show through as faint lines; snow caps on the hoist dormer hoods, the chimneys, the tops of the crates, barrels and sacks and on the crane's granite base; the coal heap in the bunker stays black with only a few white flecks; the iron crane jib stays dark. Alpha zero outside the objects exactly as before: no ground plate, no snow on the ground, no halo, no checkerboard, no cast shadow.
Avoid: any outline, position or size change, icicles, blue snow, a ground or background, people, text.

Image 1 is the selected summer harbour sheet: change the season only, preserve every outline, position, size and the transparency.
```

### Look 8 C4 harbour kit, summer (2026-10-24)
`harbour-kit-summer-v1.png` (exec-e798223b), native 1536×1024 RGBA: the warehouse (box 366×921, aspect 2.5
against the 2.14 asked for), the yard's crane, crates, barrels and sacks (570×645 with the jib, which
reaches 60 px into the third column) and the customs house with its coal bunker (431×491). Clean alpha,
38.4 % opaque, 4.4 % partial edges, no ground, no shadow. Selected: the export scales each piece from its own
box. Attachments: image 1 `art/terrain/house/house-summer-v2.png` (style and edges only), image 2
`art/terrain/quay/quay-pier-shared-v1.png` (granite only). The prompt was the `art/style.md` Preamble and
Terrain text, then:

```text
Use case: stylized-concept. Asset type: transparent RGBA cut-out sheet, LANDSCAPE canvas exactly 1536x1024 px, divided into three columns 512 px wide (x 0-511, 512-1023, 1024-1535), one object group centred in each column with wide empty transparent gutters between them. Strict top-down orthographic nadir: only roofs and tops are visible, no wall faces. Painting scale 10 px per world pixel (one map hex is 96 world px across).
Column 1, centred at (256,512): ONE long brick quayside warehouse of 1910, its roof ridge running vertically from top to bottom; footprint exactly 280 px wide by 600 px tall (28 by 60 world px), a narrow tall rectangle, only 55 % of the column width and 59 % of the canvas height. A grey-blue slate pitched roof (lit left slope near RGB 120,124,128, shaded right slope near RGB 88,92,98) with a straight ridge and dark lead flashing along it, a row of four small hoist dormers with dark tarred timber hoods on the right slope, two small skylights on the left slope, the brick-red gable copings just visible at the top and bottom ends.
Column 2, centred at (768,512): a quay-yard group of loose cargo and one hand crane that together fit a box of 400 x 440 px (40 x 44 world px): at the upper right a hand-operated quay crane seen from straight above, a round grey granite base about 90 px across with a dark iron post and a slim dark iron jib 170 px long pointing right; at the left a neat stack of six wooden crates (each about 60x60 px, pale weathered planks); below the crates a cluster of eight barrels (round tops about 36 px across, dark oak with iron hoops); at the bottom right a low pile of jute sacks; small gaps of transparency between the items.
Column 3, centred at (1280,512): ONE small customs house, a compact two-storey stone building with a red-brown clay tile hipped roof and two chimneys, footprint 340 x 280 px (34 x 28 world px); attached to its lower edge a small open timber coal bunker 140 x 100 px with a heap of black coal inside.
Weathered brick, slate, grey granite, dark tarred timber. Light from the upper left: form shading inside each object only. Clean crisp outlines. Alpha zero outside the objects: no ground plate, no cobbles, no apron, no grass, no halo, no checkerboard, no cast shadow.
Avoid: ships, boats, people, horses, carts, flags, signage, text, shipping containers, gantry cranes, rails, perspective, side view, isometric view, visible wall faces, a ground plane, water.

Image 1 is a house cut-out painted for this map at the same scale: take only its painting style, detail level and edge treatment; do not copy its building or roof. Image 2 is the selected quay pier of this map: take only the granite colour and painted finish for the crane base; do not copy the pier, its setts or its timber.
```

## Avoid
- ships, boats, people, flags, signage
- containers or gantry cranes, rails
- a ground plate, halo or cast shadow on the cut-outs
- perspective, side or isometric views, visible wall faces
