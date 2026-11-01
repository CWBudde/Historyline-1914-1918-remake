---
id: terrain/cmp
title: Depot
kind: terrain
status: approved
runtime_ids: cmpE, cmpE_w, cmpN, cmpN_w, cmpS, cmpS_w, cmpW, cmpW_w
masters: cmp-summer-v4.png, cmp-winter-v5.png, cmp-summer-v3.png, cmp-winter-v4.png, cmp-summer-v2.png, cmp-winter-v3.png, cmp-winter-v2.png, cmp-summer-v1.png, cmp-winter-v1.png, cmp-autumn-v1.png, cmp-spring-v1.png
---

## Look
A fenced supply depot over a four-hex diamond. In the reference it sits at the end of the road: a
barrier gate in the north, sheds and red vehicles to the west, white tents to the east and a large hall
to the south.

## Base prompt
Transparent whole-scene cut-out, 3:2, strict nadir roof plan: a low wide fenced military depot. A grey
road enters from the top centre through a pale barrier gate and runs down the middle. West: a squat
olive-brown barracks with a pale roof patch and two small red vehicles. East: three white canvas tents
in a column. South: one larger hall with a dark roof and a blank pale circular marker pad without a letter.
A low weathered timber-and-stone fence around the
yard; most of the yard floor transparent so meadow shows. Everything inside the four-hex diamond
outline (238x160 world pixels).

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cell N | cmpN, cmpN_w | Slice of the depot scene at the N cell of the diamond. |
| cell E | cmpE, cmpE_w | Slice of the depot scene at the E cell of the diamond. |
| cell S | cmpS, cmpS_w | Slice of the depot scene at the S cell of the diamond. |
| cell W | cmpW, cmpW_w | Slice of the depot scene at the W cell of the diamond. |
| winter | cmpE_w, cmpN_w, cmpS_w, cmpW_w | Same scene with patchy snow on roofs, tents and fence; the entry street is packed snow with ruts like the winter road. |

## Edges & joins
The scene is sliced into four hex crops with shared placements; the entrance road meets the regular
road tile at the north port.

## Post

### Spring edit (2026-11-01)
`cmp-spring-v1.png` (exec-38ae3193) is the spring form of `cmp-summer-v4.png` (native 1536x1024 RGBA): the imagegen season edit kept every
outline and position (registration shift 0,0, alpha IoU 0.793: the edit filled the soft halo and the notch), and was fitted in code onto the summer
master: the alpha is the summer's, and the edit's colours replace only the summer vegetation (a green-hue mask, dilated 3 px and softened
over 5 px); masonry, roofs, streets, water and metal keep the summer pixels. The vegetation is then graded in Lab, weighted by the same
mask, to L\* 39.1, hue 103 (dL\* +2.7, dh +6; summer vegetation L\* 39.4, hue 98). Lawn patches fresh late-April yellow-green, the star
shrubs in fresh leaf with the red berries mostly gone; the open lawns are transparent in the master and show the ground ring. A first try
that named an army depot was refused by the image tool; the prompt calls it a farm and goods yard. Mean RGB 116,105,77, L\* 44.1, hue 89
(summer 118,105,77, L\* 44.1, hue 86). `hl review art -season spring -compound art/pipeline/compound-depot-v6.json` swaps it in for
`cmp-summer-v4.png` and writes `cmpN_sp`, `cmpE_sp`, `cmpS_sp` and `cmpW_sp`; the ground ring is the spring meadow
(`grass-spring-v1.png`). Attachments: image 1 `cmp-summer-v4.png`, image 2 `grass-spring-v1.png` (grass palette only). The prompt was:

```text
Seasonal edit ONLY of image 1, a fenced supply depot (a farm and goods yard) seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha; transparent outside the depot): change the season from summer to a mild late-April day on the Western Front and nothing else. Keep the canvas size, the outline, the transparent area and every object exactly where it is, pixel for pixel: the grey asphalt street with its white dashed centre line, the barrier gate posts and the red-and-white barrier pole, the fence and its stone posts, the sheds, the two red lorries, the white tents, the crates and barrels, the dirt yards and the south hall with its circle marker. Preserve every outline and position.
Spring changes only the vegetation:
- The olive-green lawns between the buildings and the soft green meadow fringe at the outline become the same fresh spring meadow as image 2: young yellow-green grass with a little of last year's pale dead grass in the tuft bases; lit blade tips near RGB 160,170,85, mid grass near RGB 100,106,40, shaded grass near RGB 48,56,20. A few tiny dandelion-yellow and daisy-white specks in the lawns.
- The small green star-shaped shrubs turn to fresh spring leaf: light yellow-green new leaves (lit about RGB 150,165,70, shaded about RGB 40,50,18), some still sparse with grey-brown twigs showing inside; keep them darker than the grass, as now; the red berries become small dark buds.
- The dirt yards are slightly damp.
- Lightness of every part stays as in image 1; only the hue and freshness of the vegetation change.
Unchanged: every roof, plank, canvas, metal, the lorries' red paint, the fence, the street and its markings, the light from the upper left, the strict nadir view.
Outside the depot outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: moving, resizing or adding anything, blossom clouds, saturated emerald or neon green, a golden or straw cast, autumn colours, snow, frost, perspective.

Image 1 is the selected summer depot (cmp-summer-v4): change season only, preserve every outline and position. Image 2 is the selected spring meadow of the map: take only its grass colours for the lawns and fringe; do not copy its texture or layout.
```

### Autumn edit (2026-11-01)
`cmp-autumn-v1.png` (exec-eac1b9ba) is the autumn form of `cmp-summer-v4.png` (native 1536x1024 RGBA): the imagegen season edit kept every
outline and position (registration shift 0,0, alpha IoU 0.819: the edit made the soft halo and the south-west notch opaque), and was fitted in code onto the summer
master: the alpha is the summer's, and the edit's colours replace only the summer vegetation (green hue
92–175°, C\* ≥ 12, dilated 3 px and softened over 5 px); masonry, roofs, streets, water and metal keep
the summer pixels. Lawn patches straw-ochre, the star shrubs dull ochre and rust with dark berries. Mean RGB 122,100,76, L\* 43.2, a\* 5.8, hue 71. `-season autumn` swaps it in for `cmp-summer-v4.png`; the
ground ring is the autumn meadow. Attachments: image 1 `cmp-summer-v4.png`, image 2 `grass-autumn-v1.png` (palette
only). The prompt was:

```text
Seasonal edit ONLY of image 1, a fenced army supply depot seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha; transparent outside the depot): change the season from summer to early October on the Western Front and nothing else. Keep the canvas size, the outline, the transparent area and every object exactly where it is, pixel for pixel: the grey asphalt street with its white dashed centre line, the barrier gate posts and the red-and-white barrier pole, the fence and its stone posts, the sheds, the two red lorries, the white tents, the crates and barrels, the dirt yards and the south hall with its circle marker. Preserve every outline and position.
Autumn changes only the vegetation: the olive-green lawns between the buildings and the soft green meadow fringe at the outline become dry straw-gold and olive-ochre grass (lit near RGB 175,145,80, mid near RGB 120,95,40, shaded near RGB 65,52,20), with some faded olive green left in patches, matching image 2. The small green star-shaped shrubs with red berries become dry autumn shrubs, dull ochre and rust brown (near RGB 150,105,45), and the berries dark red-brown. A few tiny fallen leaves (2 to 4 px) on the dirt yards near the shrubs; the dirt yards slightly darker and damper.
Unchanged: every roof, plank, canvas, metal, the lorries' red paint, the fence, the street and its markings, the light from the upper left, the strict nadir view.
Outside the depot outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: moving, resizing or adding anything, bright orange or red foliage, snow, frost, a golden cast on roofs, tents or the street, perspective, green grass left anywhere.

Image 1 is the selected summer depot (cmp-summer-v4): change season only, preserve every outline and position. Image 2 is the selected autumn meadow of the map: take only its grass colours for the lawns and fringe; do not copy its texture or layout.
```

Selected sources are `cmp-summer-v4.png` and `cmp-winter-v5.png` (1536x1024, true alpha):
`cmp-summer-v2.png` and `cmp-winter-v3.png` with the south-west shed moved and the north-west shed
shortened inside the hex edge (below).
The summer master `cmp-summer-v2.png` is native built-in imagegen: the prompt requests fine timber, corrugated roof, canvas, stone and
crate detail, a north barrier road, olive sheds and two red period supply trucks west, three white
tents east and the south hall. PULSE supplies the layout/style; `ground-crater-summer-v1.png` supplies the
material-detail reference. The winter edit `cmp-winter-v2.png` changes only snow and dormant vegetation and
preserves the summer arrangement and transparent gaps. `cmp-winter-v3.png` is `cmp-winter-v2.png` with
the entry street (top edge to the hall roof, without the barrier pole) replaced by packed snow with two
wheel ruts and no centre line, so the gate continues the winter road tile.

`art-review -compound art/pipeline/compound-depot-v6.json` places both seasons on the seasonal meadow and
slices the four cells. Reference-layout framing scales the scene to fill the 238x160 box and crops the
excess, so the north street reaches the hex edge on the port; its first 10 world px blend into the season's
road tile (`entry_road`: `connected-terrain-v2.json`, `connected-road-winter-v1.json`). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

### Winter entry street (2026-10-24)

`cmp-winter-v3.png` takes the street pixels from `exec-1b14b37f` (native 1536×1024 RGBA, an edit of
`cmp-winter-v2.png` that kept every outline in place) and every other pixel unchanged from
`cmp-winter-v2.png`: the mask is x 696–839, y 0–499.5 master px without the barrier pole (y 106.5–119.5),
feathered over 2 px. Inside it the edit is graded to the port road's mean (L\* −4.4, a\* −0.6, b\* +0.7)
and its highlights pass a soft knee at L\* 70 (kerb snow stays below L\* 80). In the 4x winter scene: no
pixel with L\* > 85 and C\* < 8 inside the street, street mean L\* 53.6 against the winter road at the
port L\* 51.8 (ΔE 1.8), outside the street a mean absolute RGB difference to the v2 scene of 0.002.
Selected. Attached: image 1 `cmp-winter-v2.png`, image 2 `art/terrain/road/road-snow-winter-v1.png`
(packed-snow material), image 3 `assets/terrain/road/road36_w-4x.png` (ruts and value). The prompt was:

```text
Winter edit ONLY of image 1, a fenced army supply depot seen from straight above (strict nadir roof plan, 1536x1024 RGBA, true alpha). Change exactly one thing: the grey asphalt street with the white dashed centre line. Everything else stays pixel for pixel where and as it is: the canvas size, the transparent area, the snow, the fence and its posts, the barrier gate posts and the red-and-white barrier pole, the sheds, the two red lorries, the tents, the crates, barrels, bushes and the south hall with its circle marker.
The street: it runs from the top edge of the canvas straight down through the barrier gate to the hall roof, about 135 px wide (x about 700 to 835), and keeps exactly that position, width and straight edges. Repaint its whole surface, from the top edge through the gate down to the hall, as a packed-snow winter road like image 3: compacted grey snow and slush with fine grit, mean colour about RGB 128,126,127 (lit about 140,138,140, shaded about 105,103,106), neutral grey, no blue. Two darker wheel ruts run along the street, each about 20 px wide, their centres about 29 px left and right of the street centre line, a little darker than the snow between them (about RGB 105,103,105). Soft, slightly irregular rut edges, a few lighter snow streaks between them. The barrier pole keeps its soft shadow on the street.
No white centre line, no dashes, no lane paint, no asphalt black, no curbs. Do not move, add or remove anything else.

Image 1 is the installed winter depot: edit only its street surface, preserve every outline, position and the transparent area. Image 2 is the winter road material: take its packed-snow grain, patches and grey value for the street surface; do not copy its framing. Image 3 is the installed winter road tile at 4x (road 88 px wide on a 384 px hex, snow verges outside): take its ruts, their spacing relative to the road width and the overall value; do not copy its verges, its size or its hex shape.
```

### Soft outer edge (2026-10-24)

`compound-depot-v6.json` cuts the masters with `edge_feather` 9: the ground apron fades out over the
outer 2.5–7.75 world px along an irregular line (two octaves of value noise in world coordinates, the
same in both seasons) instead of a straight cut along the hex edges, as on the ruined depot; the entry
corridor stays open to the port. `edge_keep` polygons (world px of the scene box) keep what reaches or
nears the boundary exactly: the north-west shed from its shortened eave (55–81, 44–62), the fence rails left and right of the
gate (70–106 and 132–168, 32–40), the north-east tent with its crates and the end of the east fence
(148–195, 38–62), the south-west shed (57–75.6, 93.8–116), the crates west of it (53.2–57, 105.2–113)
and the crates and barrels east of it (through (76, 105.2), (83.3, 105.2), (83.3, 114.3),
(81.7, 114.3), (81.7, 119.7), (76, 119.7)); 2 world px outside a polygon the feather is back. In both
seasons no ground pixel within 2 world px of the boundary keeps alpha above 1 at 2x
(`TestCompoundFeatherIntactScenes`).

### South-west shed inside the hex edge (2026-10-24)

`cmp-summer-v3.png` and `cmp-winter-v4.png` are `cmp-summer-v2.png` and `cmp-winter-v3.png` with the
south-west shed moved 100 master px (15.6 world px) north and 8 px east onto the south-west fence line
(hiding a fence post; the fence runs into its walls), so it clears the boundary by 4 world px. The crates
west of it and the crates and barrels east of it follow by 60 px north and 8 px east; the crate in front
of its door would still reach the boundary and is dropped, the wall behind it repainted from the wall
beside it. No imagegen: the Go edit `compoundShedDepot` (`tools/art-review/compound_feather_shed_test.go`,
written by `TestGenerateCompoundShedMasters`, `-tags campaigngfx`) copies the pieces exactly. The ground
they leave is the apron that lay below the old shed, moved with it (in the west), the ground east of the
barrels mirrored (in the east) and the dirt strip below the fence repeated under the new shed, where the
feather takes it. Both seasons are edited identically; every other pixel equals the previous master
(`TestCompoundShedMasters`). In the 2x scenes no pixel of the old keep polygon (51–83, 106–131) within
2 world px of the boundary keeps alpha above 0, the 50 % contour lies 3.25–7.25 world px inside it
(`TestCompoundShedsInsideHexEdge`).

### North-west shed shortened (2026-10-25)

`cmp-summer-v4.png` and `cmp-winter-v5.png` are `cmp-summer-v3.png` and `cmp-winter-v4.png` with the
north-west shed shortened by 19 master px (3 world px), so its north eave (row 257, now 276) clears the
W cell's top edge by 3.1 world px; the crates beside it stay. Go edit `compoundShedDepotNorth`
(`compound_feather_shed_test.go`, written by `TestGenerateCompoundShedMasters`): over the gable the 19
corrugation rows below the eave drop out (the corrugation runs north-south, so the cut leaves no seam);
over the hip the corrugation north of the hip line slides with it and the hip's east face is pulled
affinely towards its fixed ridge end (452, 295) and south-east corner (510, 360), so the hip line runs
straight to the new north-east corner; ridge, south slope and wall stay. The ground strip between fence
and roof (rows 245–256, mirrored) fills the 19 rows the eave leaves, mostly under the feather. Both
seasons are edited identically: 9126 px per master equal the previous master under the slide or in
place, 3086 px of the east face are resampled, every pixel outside the edit is unchanged
(`TestCompoundShedMasters`). `compound-depot-v6.json` holds the shed only from 44 world px down (keep
polygon 55–81, 44–62). In the 2x scenes no pixel of the old keep polygon (55–81, 38–62) within 2 world
px of the boundary keeps alpha above 1 (before: 107 of 107 above 64, max 253/254), and the 50 %
contour lies 3.00 world px inside it (before 0.00, a hard cut).

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- manors or warehouses with red roofs
- people or flags
