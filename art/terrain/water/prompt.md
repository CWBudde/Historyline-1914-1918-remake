---
id: terrain/water
title: Water
kind: terrain
status: candidate
runtime_ids: water, waterStones
masters: water-shared-v2.png, water-rocks-shared-v2.png, water-shared-v1.png, water-rocks-shared-v1.png
---

## Look
Open saturated blue-teal water with fine broken sun glints. Three isolated clusters of rounded,
weathered grey-brown rocks rise from it. Selected native masters: `water-shared-v2.png` (1254×1254 opaque)
and `water-rocks-shared-v2.png` (1374×1145 RGBA), generated with the built-in image tool. The requested
1536-pixel dimensions were not returned; the retained files are the actual native outputs.

## Base prompt
Use case: stylized-concept. Generate a NEW opaque square 1536x1536 master material texture for a hand-painted World War I tactical game, not an edit. Reference images attached by local path are STYLE AND QUALITY REFERENCES only: image 1 PULSE painting defines blue water color and natural aerial map style; image 2 crater defines fine native brush detail, never copy its crater subject. Strict vertical nadir orthographic surface of open rural European lake water filling the entire canvas with no shore or objects. Saturated clear deep blue/teal water close to RGB(25,90,117), darker blue troughs, fine irregular interconnected sun ripples, tiny broken white glints. Finely painted smooth natural curves and antialiased microdetail at native resolution, comparable finish to the crater reference. Very small low waves, calm readable surface, even detail distribution and brightness all the way to every edge, no large gradients. Diffuse upper-left daylight. Newly authored continuous painterly illustration, natural volume, no photographic noise, no pixel art, no enlarged pixels, no vector graphics, no grid or hex outline, no text, no boats, no islands, no underwater rocks, no landscape horizon, no foam lines, no tropical cyan, no checkerboard. Opaque full bleed square.

Water stones prompt (separate built-in call, transparent background):

Use case: stylized-concept. NEW transparent terrain object sprite, native high detail, wide 1536x1280 canvas intended for 6:5 crop. Reference 1 PULSE painting gives natural aerial map look and blue lake water; reference 2 crater is fine painted finish quality ONLY. Strict vertical top-down nadir orthographic, no perspective tilt. Three small separated clusters of low rounded warm grey-brown boulders emerging from water: one cluster at x38% y38%, one at x64% y40%, one at x52% y63%. In each cluster one or two substantial rounded boulders with a few smaller companions, subtle cracked weathered rock faces, tiny olive moss stains, damp dark bases. Natural varied shapes and gently domed flat tops, pale upper-left faces, short soft lower-right shadows. Only an extremely narrow broken translucent pale-blue ripple immediately hugging each boulder. All else completely alpha zero, including between the clusters, no opaque water areas, no ground, no grass, no haze. Objects together occupy central 50 percent canvas width and central 45 percent height leaving generous empty transparent margins. Continuous hand-painted tactical game illustration, same polished native stone microdetail as crater reference; antialiased natural edges, not pixel art, no enlarged pixels, no white foam rings, no checkerboard painted in, no text, no outlines, no square background. Genuine transparent background.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| water | water | The water material above. |
| waterStones | waterStones | Transparent 6:5 cut-out: three compact groups of low rounded tan-grey rocks with flat cracked tops and a few small companions, clustered in the central 55 percent; only a very narrow translucent ripple fringe at each rock; everything else alpha zero. |

## Edges & joins
The water material is periodic on the world lattice, so lake hexes form one continuous surface; shore
and stream tiles use the same material, so water matches across all of them.

## Post
`art-review -shore art/pipeline/shore-terrain-v2.json` samples the water material for `water`;
`art-review -compound art/pipeline/compound-water-stones-v2.json` places the rock cut-out over it for
`waterStones`. Summer only. Both referenced images were style references, not edit targets.

The selected v2 shore and water-stones manifests use the same `authored-rocks-v1` water profile:
uniform centre crop, 71×40 logical-pixel periods, 8 percent opposing-edge overlap, no mirrored coset
average. Open water, shores, streams and the water-stones base retain the same painted glint scale.
Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tropical turquoise or neon cyan
- white foam rings
- boats or bridges
