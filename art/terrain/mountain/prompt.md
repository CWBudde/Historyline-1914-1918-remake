---
id: terrain/mountain
title: Rocky knoll
kind: terrain
status: approved
runtime_ids: mountain, mountain_w, stamp/knoll/1/1, standing/stamp/knoll/1/1, stamp/knoll/3_-1.1_-2.2/1, standing/stamp/knoll/3_-1.1_-2.2/1, stamp/knoll_w/1/1, standing/stamp/knoll_w/1/1, stamp/knoll_w/3_-1.1_-2.2/1, standing/stamp/knoll_w/3_-1.1_-2.2/1, stamp/knoll_w/3_2.0_1.1/1, standing/stamp/knoll_w/3_2.0_1.1/1, stamp/knoll_w/4_1.1_0.2_0.4/1, standing/stamp/knoll_w/4_1.1_0.2_0.4/1
masters: mountain-knoll-single-summer-v1.png, mountain-knoll-group-summer-v1.png, mountain-knoll-single-winter-v1.png, mountain-knoll-group-sw-three-winter-v1.png, mountain-knoll-group-triangle-three-winter-v1.png, mountain-knoll-group-vertical-four-winter-v1.png
---

## Look
A single small rocky knoll. In the reference it is a compact outcrop of cracked grey-beige rock with
grassy slopes, dark scrub and a few yellow flowers around it (average about RGB 80,87,38).
The selected summer single-knoll cut-out has a grassy skirt extending beyond the hex.

Winter standing stamps use the same low rocky crests with thin settled snow and sparse straw.
Single, southwest and triangular three-cell groups, and the vertical four-cell ridge have individual
paintings with transparent margins over the selected snow.

## Base prompt
Transparent terrain cut-out, strict nadir view: one low asymmetric grassy mound with broad warm olive
slopes and connected fractured grey-beige slabs covering about 35 percent of the feature. Olive moss,
small dark scrub and tiny yellow flowers occupy the cracks; an irregular grassy skirt fades into
transparent margins. Continuous hand-painted material detail, diffuse upper-left light and a short
soft lower-right shadow. No tall peaks, isolated island base, text, grid, hex outline, people,
checkerboard, photographic texture or 3D rendering.

For the eastern group, paint two distinct low grassy rocky knolls on one transparent 3:2 canvas,
one in the upper right and one in the lower left, connected loosely by scrub, meadow fringe and
small stones. Keep the same camera, light and rock scale; leave clear space at the other corners.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| mountain | mountain | Summer as above. |
| mountain_w | mountain_w | Same outcrop with thin settled snow on flat stone and in sheltered cracks; scrub muted and lightly snowy. |
| summer single-knoll stamp | stamp/knoll/1/1, standing/stamp/knoll/1/1 | Transparent ground plus the selected rocky standing cut-out and its irregular grassy skirt. |
| eastern knoll group | stamp/knoll/3_-1.1_-2.2/1, standing/stamp/knoll/3_-1.1_-2.2/1 | Two rocky crests across the three-cell group, with a connected grassy apron. |

| winter knoll-single | stamp/knoll_w/1/1, standing/stamp/knoll_w/1/1 | Snow-covered knoll_w; canvas 144×120, anchor (72,60), transparent margins. |
| winter knoll-group-sw-three | stamp/knoll_w/3_-1.1_-2.2/1, standing/stamp/knoll_w/3_-1.1_-2.2/1 | Snow-covered knoll_w; canvas 240×160, anchor (191,40), transparent margins. |
| winter knoll-group-triangle-three | stamp/knoll_w/3_2.0_1.1/1, standing/stamp/knoll_w/3_2.0_1.1/1 | Snow-covered knoll_w; canvas 240×160, anchor (48,40), transparent margins. |
| winter knoll-group-vertical-four | stamp/knoll_w/4_1.1_0.2_0.4/1, standing/stamp/knoll_w/4_1.1_0.2_0.4/1 | Snow-covered knoll_w; canvas 192×288, anchor (96,48), transparent margins. |

## Edges & joins
The summer single-knoll standing cut-out extends across hex edges with an irregular grassy skirt and
transparent margins over the shared meadow. The eastern group uses its own continuous standing
painting. Summer and winter ground stamp layers are transparent; winter groups use their own
standing paintings. Unknown winter shapes retain the compound exports below.

## Post
Derived from the mount masters (`art/terrain/mount/`): `art-review -compound
art/pipeline/compound-mountain-v1.json` places the single-cell scene for both seasons. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

For summer single knolls, `tools/terrain-stamps` exports the full native 1374x1145
`mountain-knoll-single-summer-v1.png` through `art/pipeline/feature-stamps-v1.json`, preserving aspect at 144x120 logical
and anchor `[72,60]`. Both the transparent ground and standing layers have 2x 288x240 and 4x 576x480
exports under `assets/terrain/mountain/`. The new source remains a candidate.
The eastern group uses native 1536x1024 `mountain-knoll-group-summer-v1.png`, canvas 240x160 logical and anchor `[191,40]`;
ground and standing exports are 480x320 at 2x and 960x640 at 4x. The anchor registers its upper-right
knoll to the group's first cell. This is also candidate art.
Generated with built-in imagegen; the local PULSE `RefDesign2.png` informed the style, without editing
the reference painting.

### Winter relief stamps
`tools/terrain-stamps -manifest art/pipeline/winter-relief-stamps-v1.json` exports the selected winter
relief paintings at 2×/4×. Sources, crop rectangles and hashes are pinned; aspect is preserved without
enlargement. Cluster anchors are relative to the first cell, independent of map coordinates. Boulder
scatters use the ground layer; knolls use a transparent ground stamp and an overhanging standing scene.

#### mountain-knoll-single-winter-v1.png
Native selected source, crop `[0 0 1374 1145]`; fixed 2×/4× export without enlargement.

```text
Use case: lighting-weather. Edit the FIRST reference from summer to WINTER ONLY. The SECOND reference is the selected snow material, colour only. Preserve the exact nadir camera, full composition, sizes, positions, silhouette and fine hand-painted tactical-map style. Thin settled ivory-white snow with subdued blue-grey creases rests on upward rock faces and among tufts, leaving warm grey/brown rock sides exposed. Summer grass becomes sparse pale straw; shrubs stay small and muted with light snow dusting. Soft diffuse upper-left light and short shadows. Keep original alpha-zero gaps and outer margins, genuine RGBA transparency; do not add a rectangular ground plate, haze, glow, broad drop shadow, solid white background or black mat. Keep all objects entirely in frame with clear padding. No extra objects, craters, roads, text, grid, footprints, photographic or3D style. One complete LOW grassy knoll with the SAME grey rocky crest, now softly snow-covered. Snow coats at least70percent of formerly grassy apron so it blends into selected winter meadow. Rock crest relief stays low and natural, no pointed Alpine mountain or tall cliff. Preserve original6:5 canvas at native1374x1145, no crop or stretching.
```

#### mountain-knoll-group-sw-three-winter-v1.png
Native selected source, crop `[0 0 1536 1024]`; fixed 2×/4× export without enlargement.

```text
Use case: lighting-weather. Edit the FIRST reference from summer to WINTER ONLY. The SECOND reference is the selected snow material, colour only. Preserve the exact nadir camera, full composition, sizes, positions, silhouette and fine hand-painted tactical-map style. Thin settled ivory-white snow with subdued blue-grey creases rests on upward rock faces and among tufts, leaving warm grey/brown rock sides exposed. Summer grass becomes sparse pale straw; shrubs stay small and muted with light snow dusting. Soft diffuse upper-left light and short shadows. Keep original alpha-zero gaps and outer margins, genuine RGBA transparency; do not add a rectangular ground plate, haze, glow, broad drop shadow, solid white background or black mat. Keep all objects entirely in frame with clear padding. No extra objects, craters, roads, text, grid, footprints, photographic or3D style. Preserve this broad elongated low ridge with upper-right and lower-left rocky knolls and the saddle between, now in winter. Snow coats at least70percent of formerly grassy apron, with bare rock and sparse faded straw. No new summit, huge mountain or cliffs. Preserve original3:2 canvas at native1536x1024, no crop or stretching.
```

#### mountain-knoll-group-triangle-three-winter-v1.png
Native selected source, crop `[0 0 1536 1024]`; fixed 2×/4× export without enlargement.

```text
Use case: stylized-concept. Newly authored WINTER terrain cutout in the exact fine hand-painted tactical-map material style of the supplied winter feature. Strict top-down orthographic nadir camera. Diffuse upper-left daylight, subtle short lower-right shadows. Thin cool ivory settled snow, subdued blue-grey snow creases, exposed warm grey-beige rock, sparse faded straw and tiny snow-dusted low shrubs. Genuine RGBA alpha-zero outside the irregular feature and in internal gaps, no rectangular snow mat, no brown or grey blurry backdrop, no glow or external shadow. Features entirely inside the image with clear outer margins. No text, hexes, borders, grids, buildings, roads, craters, people, footprints, photograph, pixels or3D look. One broad LOW snowy knoll GROUP with three modest rocky crests, across a shallow triangle. Crestcentres at20%width25%height (upperLEFT),79%width25%height (upperRIGHT),50%width50%height (lowerMIDDLE). Low rounded snowy mounds linked by a gentle snowy saddle; exposed cracked rocky crests no wider than17%ofcanvaswidth. Grassy-apron equivalent is mostly settled snow with sparse straw, its irregular skirt fades to clear alpha. Whole landscape3:2 canvas1536x1024; native game footprint240x160, anchor48,40. No tall peaks, Alpine mountains or enormous cliff blocks. Keep lowerquarter mostly transparent, allknolls safeinsideframe.
```

#### mountain-knoll-group-vertical-four-winter-v1.png
Native selected source, crop `[0 0 1024 1536]`; fixed 2×/4× export without enlargement.

```text
Use case: stylized-concept. Newly authored WINTER terrain cutout in the exact fine hand-painted tactical-map material style of the supplied winter feature. Strict top-down orthographic nadir camera. Diffuse upper-left daylight, subtle short lower-right shadows. Thin cool ivory settled snow, subdued blue-grey snow creases, exposed warm grey-beige rock, sparse faded straw and tiny snow-dusted low shrubs. Genuine RGBA alpha-zero outside the irregular feature and in internal gaps, no rectangular snow mat, no brown or grey blurry backdrop, no glow or external shadow. Features entirely inside the image with clear outer margins. No text, hexes, borders, grids, buildings, roads, craters, people, footprints, photograph, pixels or3D look. One elongated VERTICAL LOW rocky snowy ridge cutout, gently winding through four small shoulders: firstcentre33%width17%height; next70%width31%height; next33%width44%height; last33%width72%height. Thus the upper ridge bulgesRIGHT then bendsbackLEFT and continues down. The four shoulders are low natural outcrops connected by a subtle snowy saddle, not four tall mountains. Exposed rocky crests about22%canvaswidth each, with broader snowyclad slopes. Entire subject remains safelyinsidecentral90%width andheight. Native PORTRAIT2:3 canvas1024x1536 (vertical). Runtime footprint192x288; no stretching, clear empty lowerfifth and margins. No tall peaks, Alpine mountains, large cliffs or plateau.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tall peaks or radial ridges
- hard-edged grass islands or dirt bases
