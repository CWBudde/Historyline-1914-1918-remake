---
id: terrain/hq
title: Headquarters
kind: terrain
status: candidate
runtime_ids: hqC, hqC_w, hqN, hqN_w, hqNE, hqNE_w, hqNW, hqNW_w, hqS, hqS_w, hqSE, hqSE_w, hqSW, hqSW_w
masters: hq-summer-v2.png, hq-winter-v2.png, hq-summer-v1.png, hq-winter-v1.png
---

## Look
A walled fortress filling a seven-hex rosette. In the reference it has four round white-grey corner
towers, pale stone curtain walls, an inner hall and a red-roofed gatehouse in the north wall; the road
enters through the north gate over a bridge.

## Base prompt
Transparent square whole-scene cut-out, strict nadir roof plan: a compact walled fortress with
four round pale-grey stone corner towers with dark conical slate tops, straight curtain walls with
walkways, a red tiled gatehouse centred in the north wall, a tan earth courtyard with an inner hall, a
narrow moat and bridge on the north side, and two small irregular bushes behind the south wall. The
north entrance road is a grey asphalt carriageway centred on the canvas and touching the top edge.
Everything stays inside the seven-hex rosette outline (238x240 world pixels); outside is alpha zero.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cell N | hqN, hqN_w | Slice of the fortress scene at the N cell of the rosette (gatehouse and north road). |
| cell NE | hqNE, hqNE_w | Slice of the fortress scene at the NE cell of the rosette. |
| cell SE | hqSE, hqSE_w | Slice of the fortress scene at the SE cell of the rosette. |
| cell S | hqS, hqS_w | Slice of the fortress scene at the S cell of the rosette. |
| cell SW | hqSW, hqSW_w | Slice of the fortress scene at the SW cell of the rosette. |
| cell NW | hqNW, hqNW_w | Slice of the fortress scene at the NW cell of the rosette. |
| cell C | hqC, hqC_w | Slice of the fortress scene at the C cell of the rosette (inner hall). |
| winter | hqC_w, hqN_w, hqNE_w, hqNW_w, hqS_w, hqSE_w, hqSW_w | Same scene with patchy snow on roofs, walls and courtyard; road stays open grey. |

## Edges & joins
The scene is sliced into seven hex crops with shared placements, so pieces join exactly; the north
road meets the regular road tile at the north port.

## Post
Selected sources are the native 1254x1254 `hq-summer-v2.png` and `hq-winter-v2.png`, generated with built-in
imagegen and true alpha. The summer prompt requests four round drum towers with low slate roofs,
red terracotta north gate and south inner hall, a textured dirt courtyard, narrow northern moat and
bridge, and a centred grey entrance road. PULSE is the layout/style reference; `ground-crater-summer-v1.png`
is the material-detail reference. The winter edit requests only patchy settled snow and dormant
vegetation, preserving every building, road, prop and silhouette. The current candidate has visible
short wall faces; the requested camera remains nadir.

`art-review -compound art/pipeline/compound-hq-v2.json` places the summer and winter scenes on the seasonal
meadow and slices the seven cells; both seasons share exact placement. The scene fills the 238x240 box,
cropping the excess, and `shift` moves it 0.77 world px west so the north street lies on the port; its first
10 world px blend into the season's road tile (`entry_road`). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- castles with keeps or spires
- flags, people or vehicles
