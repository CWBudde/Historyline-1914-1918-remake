---
id: terrain/forest
title: Forest
kind: terrain
status: candidate
runtime_ids: forest, forest_w
masters: forest-crowns-summer-v3.png, forest-crowns-winter-v3.png, forest-crowns-inset-summer-v3.png, forest-crowns-inset-winter-v3.png, forest-crowns-summer-v2.png, forest-crowns-winter-v2.png, forest-crowns-inset-summer-v2.png, forest-crowns-inset-winter-v2.png, forest-crowns-spaced-summer-v2.png, forest-crowns-initial-winter-v2.png, forest-crowns-initial-summer-v2.png, forest-crowns-summer-v1.png, forest-crowns-winter-v1.png
---

## Look
Dense woodland seen from above. Selected v3 crowns are four irregular rounded deciduous canopies in
dark olive and deep moss green, with small leaf masses, muted upper-left highlights and deep internal
shade. A crown is about a fifth of a hex across. Winter keeps the crown arrangement as dense grey-brown
branch networks with thin settled snow. Earlier mixed broadleaf/conifer atlases remain as drafts.

## Base prompt
Transparent RGBA atlas, square 1254x1254, 2x2 panels of 627x627. Four distinct compact rounded oak/beech
crowns, strict nadir, built from small irregular leaf masses in dark olive/moss green with deep shade
and modest muted ochre upper-left highlights. No starburst needles, rosettes or oversized leaves.
Keep crowns isolated with transparent gutters. No trunks, ground or shadow mats. The exact native v3
generation prompts and deterministic runtime inset are recorded under Post.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| forest | forest | Four dark rounded deciduous summer crowns. |
| forest_w | forest_w | Same four crown centres and footprints: dense bare grey-brown branch networks with thin ivory snow. |

## Edges & joins
The renderer scatters whole crowns over the seasonal meadow per context; crowns may overhang hex
edges into neighbouring forest, and towards non-forest neighbours the canopy thins so meadow shows at
the border. Scale and light must match the house, hq and mountain art.

## Post
`art-review -forest art/pipeline/forest-terrain-v3.json` places crowns by a fixed canopy profile into
context tiles keyed by season and neighbours (`context/forest/<key>`); winter uses the winter atlas
and `grass_w`. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.


### Native v3 generation

The selected native summer and winter atlases are 1254×1254 RGBA, generated with the built-in imagegen
tool. RefDesign2 was the summer style reference; winter edits the summer arrangement. Runtime inset
atlases use the same deterministic alpha-aware Lanczos3 preparation as v2: every 627×627 cell becomes
547×547, centred with 40 extra transparent pixels per side in the original canvas. No artwork is repainted
by that export preparation. Runtime scatter, neighbour contexts and grass guards are unchanged.

#### Summer: forest-crowns-summer-v3.png

```text
Use case: stylized-concept. Asset type: transparent RGBA forest tree-crown atlas for the rural European World War I strategy map. Supplied RefDesign2 aerial map is a STYLE reference only: match its small dark olive rounded tree crowns with dense connected leaf masses and deep moss shadows. Create one square 1254x1254 atlas containing exactly four isolated crowns in an invisible 2x2 layout of 627x627 cells. Each crown is centred in its cell, round but asymmetrical, about 470 pixels across, fully inside its cell with at least 65 pixels clear alpha-zero margin on ALL sides. FOUR DIFFERENT dense DECIDUOUS European broadleaf crowns seen strictly vertically from above, no visible trunks. Build each crown from many SMALL irregular rounded masses of tiny leaves, full and bushy like compact oak/beech woodland. Overall canopy must be DARK olive/moss green, deep internal shade, modest muted ochre-green highlights on only the upper-left small leaf masses; never pale yellow-green speckling everywhere. Dense opaque interiors, natural dark irregular edges with a few small transparent notches. Fine believable hand-painted digital illustration, soft upper-left diffuse daylight and short interior shadows. The silhouette should read as a dense rounded forest tree at twenty world pixels, not a sprig, flower or fern. No conifer starburst, radial needles, rosette, oversized individual leaves, flat leafy discs, plastic 3D, photograph, pixel art, bright lime, ground, visible cast-shadow mat, haze, background color, checkerboard, panels, text or outlines. Genuine transparency outside all four crowns.
```

#### Winter: forest-crowns-winter-v3.png

```text
Use case: lighting-weather. Edit target is the supplied native1254x1254 transparent atlas of four dense rounded deciduous tree crowns in a2x2layout. Preserve exact canvas, four crown centres, full individual footprints, tree-scale, branch-mass arrangement and all transparent spaces between cells. Change only season: the four broadleaf crowns become four different dense grey-brown radial networks of bare small branches and fine twigs, viewed strictly vertically from above. Thin settled ivory snow rests selectively on upper branches, with cool dark internal shadows and sparse muted brown retained leaves. The compact rounded canopy silhouettes and small branch-mass structure must remain recognizable from summer, no new conifers or different trees. Keep the finely hand-painted aerial-map style, diffuse upper-left light, fine believable wood/snow detail. No ground, trunks seen from side, cast-shadow mats, haze, glow, frame, panels, checkerboard or text. Genuine alpha transparency outside every crown and between fine bare branches.
```


### Earlier v2 generation

Earlier outputs are native 1254×1254 RGBA. The final size correction and winter edit prompts were:

```text
Use case: precise-object-edit. Edit supplied four-tree atlas to correct tree size for game export. Preserve square canvas, 2x2 atlas geometry, same four crown centres, tree identities, foliage materials and upper-left lighting. Enlarge every crown by precisely about 25 percent of its current diameter. Final crown diameter about 75 percent of its own panel width, leaving at least 10 percent panel-width true alpha-zero empty guard on every side. No crown may enter that guard or touch another. Keep complete crowns and natural irregular outlines, no cropping. Strict nadir. Preserve fine hand-painted leaf/needle detail, genuine transparent background and gaps, no ground/shadows/text.
```

```text
Use case: lighting-weather. Edit target is supplied square four-tree atlas. Preserve exact1254x1254canvas, four crowncentres, strictnadir camera, crownscale and footprints, 2x2panel arrangement, alltransparentgutters. Changeonlyseason: top two broadleafcrowns become finelypainted leaflessgrey-brown radialbranchandtwig networks withdelicate settledivorysnow, transparentgaps. Bottomtwo conifers keepidentical radialbranchandneedlestructure withpatchy settledivorysnow ondarkcoolblue-greenneedles. Softupperleftlight andshortinternalshadow. Keep allbrancheswithinoriginalfootprints and at least8percentofeachpanel astransparentemptyguard. No ground, shadowmat, trunksvisible, glow, borders, checkerboard or text. Genuinealpha background.
```

Earlier native v2 atlases are `forest-crowns-summer-v2.png` and `forest-crowns-winter-v2.png`.
Runtime inputs `forest-crowns-inset-summer-v2.png` and `forest-crowns-inset-winter-v2.png` deterministically normalize each
627×627 panel to 547×547 with the existing alpha-aware Lanczos3 resampler, centred with 40 transparent
pixels of extra padding per side inside the unchanged 1254×1254 atlas. This satisfies the exporter's
8 percent alpha guard without painting or clipping crowns. Other v2 atlases are earlier spacing candidates.

Generated with the built-in imagegen tool. Dimensions below describe the actual output.

#### Summer master (1254×1254 RGBA)

```text
Use case: stylized-concept. Asset type: transparent RGBA tree-crown atlas for strict nadir European World War I tactical map. Create square 1536x1536 with exactly 2x2 invisible equal panels of 768x768 pixels. One separate complete isolated tree crown centered in each panel at (384,384),(1152,384),(384,1152),(1152,1152). Each crown 520 pixels in diameter maximum; fully inside its own panel leaving at least 110 pixels genuine alpha-zero transparent guard on all sides. Top two crowns: two differently shaped dense oak/beech broadleaf crowns with irregular leafy clustered outlines. Bottom two: two distinct spruce/fir conifer crowns from directly vertically above, layered radial branches, no visible trunk or side view. Finely detailed hand-painted digital illustration, realistic small leaf masses and needle sprays, rich deep moss and dark teal greens with warm olive yellow-green sunlit tops, light upper left and short soft interior shade lower right. Reference painting quality, lush natural aerial-map trees. No ground, tree shadows on ground, shadow mat, fog, glow, visible checkerboard, background, text, panels or outlines. Alpha transparent outside and between branches. No pixel art, vector, cartoon, photograph or plastic 3D rendering.
```

#### Summer gutter correction (1254×1254 RGBA)

```text
Use case: precise-object-edit. Edit the supplied tree-crown atlas only to add much larger transparent gutters. Keep square canvas, same 2x2 layout and same four tree identities, colors, light and camera. Shrink each crown to 62 percent of its existing diameter WITHOUT changing its centre within its own quarter of the image. Every crown must be fully isolated and surrounded by at least 18 percent of its panel width of true alpha-zero empty transparent margin on ALL FOUR sides, including near the centre cross of the full atlas. No crown touches another or a panel edge. Preserve the crisp hand-painted leaf and needle detail; transparent background, no shadows on ground, no panel lines or labels.
```

#### Winter master (1254×1254 RGBA)

```text
Use case: lighting-weather. Edit target: supplied square transparent four-tree atlas. Change season ONLY. Preserve exact canvas dimensions, 2x2 panel arrangement, four crown centres, crown scale, branch arrangement, maximum footprints and all transparent empty gutters. Top two broadleaf trees: replace leaves with finely detailed bare radial grey-brown twigs/branch networks seen strictly from above, with delicate settled ivory snow on upper branches, transparent visible gaps between branches. Bottom two conifers retain exact needle and radial branch structures, dark cool moss/blue-green needles with patchy settled ivory snow on upper-left-facing branches. Same top-down nadir camera and soft upper-left light, hand-painted aerial map material quality. Every tree remains completely inside its current footprint with at least 18 percent panel empty alpha-zero margin. No trunks visible, ground, shadow mat, text, framing or background. Genuine alpha transparency outside crowns and between bare twigs.
```

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- visible trunks or side views of trees
- blue or cyan edge mattes
- crowns larger than about a fifth of a hex
