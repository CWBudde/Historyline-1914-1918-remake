---
id: terrain/mountain
title: Rocky knoll
kind: terrain
status: candidate
runtime_ids: mountain, mountain_w, stamp/knoll/1/1, standing/stamp/knoll/1/1, stamp/knoll/3_-1.1_-2.2/1, standing/stamp/knoll/3_-1.1_-2.2/1
masters: mountain-knoll-single-summer-v1.png, mountain-knoll-group-summer-v1.png
---

## Look
A single small rocky knoll. In the reference it is a compact outcrop of cracked grey-beige rock with
grassy slopes, dark scrub and a few yellow flowers around it (average about RGB 80,87,38).
The selected summer single-knoll cut-out has a grassy skirt extending beyond the hex.

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

## Edges & joins
The summer single-knoll standing cut-out extends across hex edges with an irregular grassy skirt and
transparent margins over the shared meadow. The eastern group uses its own continuous standing
painting. Both ground stamp layers are transparent. Winter knolls retain the compound exports below.

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

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tall peaks or radial ridges
- hard-edged grass islands or dirt bases
