---
id: terrain/house
title: House
kind: terrain
status: approved
runtime_ids: house, house_w
masters: house-summer-v2.png, house-winter-v2.png, house-summer-v1.png, house-winter-v1.png
---

## Look
A single rural building in open meadow. In the reference, houses are square buildings with a grey-brown
hipped roof of weathered shingles, each with a smaller shed beside it and a bush or small tree nearby,
about a quarter of a hex across.
The selected summer and winter candidates are `house-summer-v2.png` and `house-winter-v2.png`: one square
house with a hipped shingle roof and an adjacent woodshed; winter preserves the building geometry.

## Base prompt
Transparent single-building cut-out, strict nadir view: one compact square timber hut with a weathered
grey-brown shingle hipped roof, four sloped planes meeting at a very short ridge, a small chimney and a
thin shadow-side eave. A smaller woodshed stands adjacent on the lower right; a few threshold stones
and a minimal mossy fringe ground the buildings. Continuous hand-painted material detail, diffuse
upper-left light and short soft lower-right shadows, with generous transparent margins. No text,
grid, hex outline, people, checkerboard, photographic texture or 3D rendering.

For the winter edit, preserve the selected summer painting's camera, canvas dimensions, house
footprint, woodshed placement and roof geometry. Add thin patchy ivory snow on the four roof planes
and sheltered ground, keeping shingles visible along the ridge and eaves; mute vegetation to dormant
tones and cool the shadows. Retain the same hand-painted treatment, upper-left light, transparent
margins and exclusions.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| house | house | Summer as above. |
| house_w | house_w | Same geometry; a thin uneven layer of settled snow on the roof planes with shingles showing along eaves and ridge; subdued snowy fringe. |

## Edges & joins
The building sits inside the hex with meadow around it; only the shared seasonal meadow reaches the
hex edges.

## Post
`art-review -compound art/pipeline/compound-hamlet-v1.json` places the cut-out on the seasonal meadow at a
fixed inset. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.
The manifest selects `house-summer-v2.png` and `house-winter-v2.png`, each native 1315x1197, at inset 0.75.
The earlier `house-summer-v1.png` and `house-winter-v1.png` masters are retained. Both selected seasons remain candidates.
Generated with built-in imagegen, with winter edited from the selected summer painting. The local
PULSE `RefDesign2.png` informed the style; the reference painting itself was not edited.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- orange clay or blue roofs
- gardens, fences or people
