---
id: terrain/hq
title: Headquarters
kind: terrain
status: candidate
runtime_ids: hqC, hqC_w, hqN, hqN_w, hqNE, hqNE_w, hqNW, hqNW_w, hqS, hqS_w, hqSE, hqSE_w, hqSW, hqSW_w
masters: hq-summer-v3.png, hq-winter-v3.png, hq-summer-v2.png, hq-winter-v2.png, hq-summer-v1.png, hq-winter-v1.png
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
| winter | hqC_w, hqN_w, hqNE_w, hqNW_w, hqS_w, hqSE_w, hqSW_w | Same scene with patchy snow on roofs, walls and courtyard; the entrance road and bridge deck are packed snow with ruts like the winter road tile, without painted lines. |

## Edges & joins
The scene is sliced into seven hex crops with shared placements, so pieces join exactly; the north
road meets the regular road tile at the north port.

## Post
Selected sources are the native 1254x1254 `hq-summer-v3.png` and `hq-winter-v3.png`, generated with built-in
imagegen and true alpha. They are layout edits of `hq-summer-v2.png` (four round drum towers with low slate
roofs, red terracotta north gate and south inner hall, a textured dirt courtyard, narrow northern moat and
bridge; PULSE was the layout/style reference, `ground-crater-summer-v1.png` the material-detail reference):
the entrance road and the bridge deck are as wide as the road tile (22-23 world px), and in winter they
carry packed snow with ruts instead of asphalt with white lines. The candidate has visible short wall
faces; the requested camera remains nadir.

`art-review -compound art/pipeline/compound-hq-v3.json` places the summer and winter scenes on the seasonal
meadow and slices the seven cells; both seasons share exact placement. The scene fills the 238x240 box,
cropping the excess, and `shift` moves it 0.25 world px west so the north street lies on the port (centre within
0.1 world px); its first 10 world px blend into the season's road tile (`entry_road`). Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

### Wide entrance road summer (2026-10-24)

`hq-summer-v3.png` (`exec-aa88c376`, native 1254x1254 RGBA, unchanged) is `hq-summer-v2.png` with the
entrance road widened from 94 to 120 px (18 to 23 world px) and the bridge parapets moved out so the deck
is as wide; towers, walls, gatehouse and courtyard keep their places (alpha IoU 0.979 outside the gate
area). Cut: street 22-23.3 world px from the port to the bridge, 24-25 on the deck, narrowing only at the
gatehouse arch; asphalt mean within ΔE 4.2 of the road tile. Two earlier tries painted the road 27.5 and
26.3 world px wide. Selected. Attached: image 1 `hq-summer-v2.png`. The prompt was:

```text
Precise layout edit ONLY of image 1, a walled fortress seen from straight above on a 1254x1254 RGBA canvas (transparent outside the fortress). Widen the north entrance road and its moat bridge slightly; change nothing else.
Now: the grey asphalt road runs from the top edge (y 0) down to the bridge at about y 180. Including its thin pale edge lines it is 94 px wide (x 583 to 677), centred on x 630. The cobbled bridge deck between the two pale stone parapets is also 94 px wide (x 582 to 676); the parapets are about 22 px wide (west x 558-580, east x 678-700) and run from about y 175 to y 300, where the red-roofed gatehouse begins.
Change: make the road exactly 106 px wide including its pale edge lines, from x 577 to x 683, still centred on x 630 - that is one twelfth of the canvas width and only 12 px wider than now (6 px more on each side). Two previous attempts painted the road 144 px and 137 px wide, far too wide: the road must stay narrow, at most 110 px. Its pale edge lines stay right where the green verge begins. The road stays parallel-sided from the top edge all the way down to the bridge, with the same grey asphalt, the same white dashed centre line on x 630 and the same thin pale edge lines along both sides. Move both bridge parapets outward by 6 px (west parapet x 552-574, east parapet x 686-708) so the cobbled bridge deck is also 106 px wide and joins the road without a step; the deck stays cobbled up to the gatehouse arch. The grass verges, rocks and moat banks beside the road and the bridge give up the same 6 px on each side.
Keep everything else exactly as it is: canvas size, outline and transparent area, the four round towers, the curtain walls, the red gatehouse and its arch, the moat water, the courtyard and its props, the south hall, the bushes; same light from the upper left, same palette, same strict nadir view.
Avoid: a road wider than 110 px, moving or resizing any tower, wall or building, a wider gatehouse, a different road colour, a road that narrows before the bridge, new objects, perspective, a ground plate, a halo, a checkerboard.

Image 1 is the installed summer fortress (hq-summer-v2): edit it as described and preserve every other outline and position.
```

### Wide entrance road winter (2026-10-24)

`hq-winter-v3.png` (`exec-0b20ef0c`, native 1254x1254 RGBA, unchanged) is the season edit of
`hq-summer-v3.png`: snow as on `hq-winter-v2.png`, and the entrance road and bridge deck are packed grey
snow with faint ruts and snow rims, without painted lines. Cut (street mask from the summer tile): mean
L\* 58.8, C 1.3, ΔE 5.5 from the winter road surface at the port (L\* 53.4); no paint-white pixels in the
middle 16 world px, only the snow rims within 2 world px of the edges. Selected. Attached: image 1
`hq-summer-v3.png`, image 2 `hq-winter-v2.png` (snow look), image 3 `tile/road36_w` at 4x over snow (road
surface). The prompt was:

```text
Seasonal edit ONLY of image 1, a walled fortress seen from straight above on a 1254x1254 RGBA canvas (transparent outside the fortress): change the season to winter and nothing else. Keep the canvas size, the outline, the transparent area and every object exactly where it is: the north road, the moat bridge and its two parapets, the moat, the four round towers, the curtain walls and wall walks, the red gatehouse, the courtyard props, the well, the south hall and the bushes. Preserve every outline and position.
Winter, as on image 2 (the same fortress in its current winter version): patchy settled snow on the courtyard, the wall walks and the tower tops, a light snow crust on the roofs, parapets, crates and carts, frost on the stones; the moat stays dark open water with snow on its banks and boulders; the meadow tufts become bare frosted brown twigs.
The north entrance road becomes a packed-snow winter road like image 3: trodden grey snow and slush, cool neutral grey, mean about RGB 132,130,132 (lit grains up to 160, darker slush down to 100), with two slightly darker wheel ruts running straight along it and a soft rim of white snow along both edges. It has NO painted markings at all: no white dashed centre line, no white or pale edge lines, no black asphalt. Keep the road exactly where it is and exactly as wide as in image 1: 120 px wide, from x 568 to x 688, centred on x 628, parallel-sided from the top edge down to the bridge. The cobbled bridge deck between the parapets (x 565 to 688, y 180 to 310) gets the same trodden grey snow and slush with the ruts running on, no white markings, the cobbles only faintly showing through.
Snow white is a soft neutral grey-white with cool grey shading, as on image 2.
Outside the fortress outline: alpha zero, no ground plate, no halo, no checkerboard.
Avoid: moving, resizing or adding anything, a narrower or wider road, white lines or dashes on the road, dark asphalt, blue snow, ice on the road, snow outside the alpha footprint, perspective.

Image 1 is the selected summer fortress: change season only, preserve every outline and position. Image 2 is the installed winter fortress: take only its snow colour, frost and bare-twig look; its road still shows summer asphalt with white dashes, which must NOT be copied. Image 3 is the winter road tile of the map over snow: take only its road surface (packed grey snow, slush ruts, soft snow edges) for the entrance road and bridge deck; do not copy its width or position.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- castles with keeps or spires
- flags, people or vehicles
