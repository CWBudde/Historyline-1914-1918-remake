---
id: terrain/cmp
title: Depot
kind: terrain
status: candidate
runtime_ids: cmpE, cmpE_w, cmpN, cmpN_w, cmpS, cmpS_w, cmpW, cmpW_w
masters: cmp-summer-v2.png, cmp-winter-v3.png, cmp-winter-v2.png, cmp-summer-v1.png, cmp-winter-v1.png
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
Selected sources are `cmp-summer-v2.png` and `cmp-winter-v3.png` (1536x1024, true alpha). The summer
master is native built-in imagegen: the prompt requests fine timber, corrugated roof, canvas, stone and
crate detail, a north barrier road, olive sheds and two red period supply trucks west, three white
tents east and the south hall. PULSE supplies the layout/style; `ground-crater-summer-v1.png` supplies the
material-detail reference. The winter edit `cmp-winter-v2.png` changes only snow and dormant vegetation and
preserves the summer arrangement and transparent gaps. `cmp-winter-v3.png` is `cmp-winter-v2.png` with
the entry street (top edge to the hall roof, without the barrier pole) replaced by packed snow with two
wheel ruts and no centre line, so the gate continues the winter road tile.

`art-review -compound art/pipeline/compound-depot-v3.json` places both seasons on the seasonal meadow and
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

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- manors or warehouses with red roofs
- people or flags
