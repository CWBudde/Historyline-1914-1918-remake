---
id: terrain/cmp
title: Depot
kind: terrain
status: candidate
runtime_ids: cmpE, cmpE_w, cmpN, cmpN_w, cmpS, cmpS_w, cmpW, cmpW_w
masters: cmp-summer-v2.png, cmp-winter-v2.png, cmp-summer-v1.png, cmp-winter-v1.png
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
| winter | cmpE_w, cmpN_w, cmpS_w, cmpW_w | Same scene with patchy snow on roofs, tents and fence; road stays open. |

## Edges & joins
The scene is sliced into four hex crops with shared placements; the entrance road meets the regular
road tile at the north port.

## Post
Selected sources are native 1536x1024 `cmp-summer-v2.png` and `cmp-winter-v2.png`, generated with built-in
imagegen and true alpha. The summer prompt requests fine timber, corrugated roof, canvas, stone and
crate detail, a north barrier road, olive sheds and two red period supply trucks west, three white
tents east and the south hall. PULSE supplies the layout/style; `ground-crater-summer-v1.png` supplies the
material-detail reference. The winter edit changes only snow and dormant vegetation and preserves
the summer arrangement and transparent gaps.

`art-review -compound art/pipeline/compound-depot-v2.json` places both seasons on the seasonal meadow and
slices the four cells. Reference-layout framing scales the scene to fill the 238x160 box and crops the
excess, so the north street reaches the hex edge on the port; its first 10 world px blend into the season's
road tile (`entry_road`: `connected-terrain-v2.json`, `connected-road-winter-v1.json`). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- manors or warehouses with red roofs
- people or flags
