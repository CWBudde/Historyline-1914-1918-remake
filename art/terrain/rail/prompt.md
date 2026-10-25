---
id: terrain/rail
title: Railway
kind: terrain
status: approved
runtime_ids: rail13, rail14, rail15, rail236, rail24, rail25, rail26, rail346, rail35, rail356, rail36, rail46, rail13_w, rail14_w, rail15_w, rail236_w, rail24_w, rail25_w, rail26_w, rail346_w, rail35_w, rail356_w, rail36_w, rail46_w
masters: rail-ballast-shared-v2.png, rail-steel-shared-v2.png, rail-ballast-shared-v1.png, rail-steel-shared-v1.png, rail-track-winter-v1.png, rail-track-summer-v1.png
---

## Look
Single-track railway: a narrow grey ballast bed with dark timber sleepers and two thin steel rails,
in the same muted palette and scale as the road.

## Base prompt
Opaque 6:5 material canvas filled edge to edge with compacted railway ballast: small angular weathered
grey limestone and greige stone chips, closely interlocked, tiny earth pockets between them, muted
neutral grey-brown, quiet contrast. No rails, sleepers, grass, verges or track lines.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| rail13 | rail13 | Track wide bend between ports 1 NE, 3 S. |
| rail14 | rail14 | Track straight run between ports 1 NE, 4 SW. |
| rail15 | rail15 | Track wide bend between ports 1 NE, 5 NW. |
| rail236 | rail236 | Track T or Y junction to ports 2 SE, 3 S, 6 N. |
| rail24 | rail24 | Track wide bend between ports 2 SE, 4 SW. |
| rail25 | rail25 | Track straight run between ports 2 SE, 5 NW. |
| rail26 | rail26 | Track wide bend between ports 2 SE, 6 N. |
| rail346 | rail346 | Track T or Y junction to ports 3 S, 4 SW, 6 N. |
| rail35 | rail35 | Track wide bend between ports 3 S, 5 NW. |
| rail356 | rail356 | Track T or Y junction to ports 3 S, 5 NW, 6 N. |
| rail36 | rail36 | Track straight run between ports 3 S, 6 N. |
| rail46 | rail46 | Track wide bend between ports 4 SW, 6 N. |
| winter | rail13_w, rail14_w, rail15_w, rail236_w, rail24_w, rail25_w, rail26_w, rail346_w, rail35_w, rail356_w, rail36_w, rail46_w | The same track in winter: snow deep on the ballast shoulders and verge, thinner between the sleepers, sleepers lightly dusted, rails dark cold steel. |

## Edges & joins
The ballast bed is 11 world pixels wide with a 1-pixel earth border, centred on the shared port centres
and continuing through each port; rails and sleepers run on the same centre line, so track joins
exactly and crossings line up.

## Post

### Look 4 R3 rail strip candidate (2026-10-23)

`rail-track-summer-v1.png` (`exec-fc8d50e4`, native 1536×1024 RGB, unchanged) is the R3 single
track: 32 sleepers at a 47.2 px pitch (spec 48), about 21 × 115 px, rails 64.5 px apart (spec 69) with
a bright running line, ballast about 165 px (spec 171), plain meadow, clean nadir. Faults: the sleeper
gap across the wrap is 74 px instead of 48, so the strip does not tile until it is respaced; ballast too
warm (141,127,112 against 150,146,138), sleepers too light. Not selected. Attached:
`rail-ballast-shared-v2.png` and `trench-timber-shared-v2.png` (materials). The prompt was the `art/style.md` Preamble and Terrain text, then:

```text
Opaque 1536x1024, strict nadir, horizontally tileable. One standard-gauge single railway track of 1915 runs left to right centred on y=512; canvas = 144x96 world px. Ballast bed 171 px wide (16 world px) of pale weathered grey limestone chippings (RGB near 150,146,138), slightly darker and oily between the sleepers. EXACTLY 32 evenly spaced timber sleepers across the full width (pitch 48 px), each 117 px long and 19 px wide, dark creosoted oak (RGB near 60,46,36) with worn grain and split ends. Two continuous steel rails 69 px apart (centred 35 px above and below the centreline), each 11 px wide: dark rust-brown web and a narrow bright running surface. Spikes and tie plates at every rail and sleeper crossing. The ballast shoulders taper into a 30 px grey-brown cinder verge with sparse weeds, then olive meadow. Upper-left light, short soft rail shadows on the sleepers.
Avoid: double track, concrete sleepers, third rail, wires or signals, uneven spacing, curves, chrome, perspective, text.

Size corrections (the first attempt drew everything about 1.6 times too large; keep the track SMALL and NARROW in the canvas):
- The track is a thin band whose centre is exactly y=512 (an earlier attempt sat 11 px too high).
- The two rails sit at y=477 and y=547, 70 px apart centre to centre (7% of the canvas height). The best earlier attempt had them 56 px apart, slightly too close; the gauge must be 70 px.
- Sleepers span only y=454 to y=570 (117 px, 11% of the canvas height), so each sleeper sticks out just 23 px beyond each rail. Each sleeper 19 px wide.
- Exactly 32 sleepers, one every 48 px, centred at x=24, 72, 120, ... 1464, 1512, so the spacing continues seamlessly across the left/right wrap (an earlier attempt left a 62 px gap at the wrap: the first sleeper must sit 24 px from the left edge and the last 24 px from the right edge).
- Ballast spans only y=427 to y=597 (171 px, 17% of the canvas height). The cinder verge is y=397-427 and y=597-627; everything above y=397 and below y=627 (about 75% of the canvas) is plain meadow.
- Rails are dark rust-brown steel (RGB near 70,50,38) with only a 3 px bright steel running line on top, not pale bars.
Colour corrections (measured on earlier attempts): ballast must be cool neutral light grey, blue channel almost as high as red (RGB near 150,146,138; earlier 142,125,105 was too warm and beige); sleepers darker brown-black creosote (RGB near 60,46,36; earlier 76,61,48); the meadow greener and less yellow (average near RGB 100,107,40, green clearly above red; earlier 113,104,46). The meadow is open short grass only: no bushes, shrubs, trees or rocks.
Attached references: image 1 (rail ballast material swatch) and image 2 (trench timber material swatch) set the material finish of the ballast chippings and the timber; copy none of their layout.
```

### Look 4 R3 winter edit candidate (2026-10-23)

`rail-track-winter-v1.png` (`exec-bb0898fe`, native 1536×1024 RGB, unchanged) is the winter edit of
the R3 candidate: rails and all 32 sleepers within 1 px of summer, neutral snow (204,203,202), dark
rails. Faults: the ballast edge blends into the meadow snow, the sleeper tops are fully bare (spec
partly clear), and it inherits the wrap gap. Not selected. Attached: the R3 candidate. The edit prompt
was:

```text
Same geometry; snow between the sleepers and on the shoulders, sleeper tops partly clear, dark rails with a bright running edge, snowy meadow.
Corrections to a previous attempt: the rails must stay dark rust-brown steel (RGB near 70,52,40) with only a thin bright running edge, not pale silver bars (previous rails averaged 134,128,127). Snow is neutral ivory-white in soft overcast light (lit RGB near 222,222,218, shaded near 190,192,194), not bluish (previous meadow snow averaged 193,197,207). Between the sleepers the ballast is filled with smooth settled snow with only a few dark chippings showing, not a field of individually frosted stones.
Attached references: image 1 (the summer rail strip to edit); copy none of their layout.
```

`art-review -connected art/pipeline/connected-rail-v2.json` cuts the ballast, lays timber sleepers (trench
timber material) and steel rails (steel material) along each port set. Season-neutral. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

Selected built-in imagegen masters: ballast is native 1536×1024 RGB; steel is native 1254×1254 RGB.
The steel prompt specifies matte weathered dark grey steel, fine scratches, slight oxide and diffuse
highlights; no chrome, rail silhouettes or objects. The former material sources remain retained drafts.
The ballast footprint, rails and registered ports are unchanged.

`connected-rail-winter-v1.json` renders the `_w` forms from the same materials with `verge: snow`: the
ballast blends toward the winter ground (`grass-winter-v1.png`) at 0.55 between the sleepers and 0.9 on the
shoulders and earth border, the sleepers at 0.12, and the steel darkens 0.5 toward RGB 42,42,46. The winter
maps draw them in place of the summer rails. The tagged generator `TestGenerateWinterPathsAndRails`
(`campaigngfx`) installs them; `TestWinterPathsAndRailsMeetCriterion` measures them over the winter ground:
verge ΔE 4.4–5.1 from the ground beside it, no summer green, rail centre lines L\* ≤ 40 along 99.4–99.8 %
(summer steel 77–80 %), sleeper autocorrelation peak at the 4-world-pixel pitch (r 0.80–0.82). The
unpromoted strip `rail-track-winter-v1.png` is not used.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- rails or sleepers in the material
- chrome highlights on steel
