---
id: terrain/forest
title: Forest
kind: terrain
status: candidate
runtime_ids: forest, forest_w
masters: crowns.png, crowns_w.png
---

## Look
Dense woodland seen from above. In the reference, forest hexes are packed with rounded broadleaf and
conifer crowns in deep moss to dark teal green (average about RGB 59,72,25) with yellow-green sunlit
tops and soft shadows to the lower right; meadow shows only in small gaps. A single crown is about a
fifth of a hex across. Winter forest has bare grey-brown branch networks and snow-dusted conifers.

## Base prompt
Transparent RGBA atlas, square 1254x1254, 2x2 panels of 627x627, one complete isolated tree crown per
panel seen straight from above: top-left and top-right two different broadleaf crowns, bottom-left and
bottom-right two different conifers. Rounded, lumpy crowns built from many leaf clusters, deep moss and
dark teal green with warm yellow-green sunlit tops on the upper left and darker undersides to the lower
right. Each crown about 500 pixels across, centred, fully inside its panel with at least 50 pixels of
true alpha-zero margin. No trunks, ground, shadow mats, background or coloured alpha fringe.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| forest | forest | Summer crown atlas as above. |
| forest_w | forest_w | Same four crowns, same centres and footprints: broadleaf crowns become leafless radial grey-brown branch networks with a little snow; conifers keep their shape with patchy off-white snow on dark blue-green needles. |

## Edges & joins
The renderer scatters whole crowns over the seasonal meadow per context; crowns may overhang hex
edges into neighbouring forest, and towards non-forest neighbours the canopy thins so meadow shows at
the border. Scale and light must match the house, hq and mountain art.

## Post
`art-review -forest art/pipeline/forest-terrain-v1.json` places crowns by a fixed canopy profile into
context tiles keyed by season and neighbours (`context/forest/<key>`); winter uses the winter atlas
and `grass_w`. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- visible trunks or side views of trees
- blue or cyan edge mattes
- crowns larger than about a fifth of a hex
