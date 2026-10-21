---
id: terrain/lakeside
title: Lake shore
kind: terrain
status: revised
runtime_ids: lakeside12, lakeside123, lakeside126, lakeside156, lakeside16, lakeside23, lakeside34, lakeside3456, lakeside56
masters: lakeside-bank-shared-v3.png, lakeside-bank-shared-v2.png, lakeside-bank-shared-v1.png, lakeside-lake-summer-v3.png
---

## Look
Reusable shore between lake and meadow with rounded warm grey and brown-grey boulders, damp dark
crevices, olive moss and small grass tufts. All nine shore shapes use the same bank, water and meadow
materials, including on PULSE. Selected `lakeside-bank-shared-v3.png` is a native 1254×1254 opaque
built-in imagegen output with larger rock faces and more vegetation between them. Requested 1536×1536
was not the delivered size. The whole-lake painting `lakeside-lake-summer-v3.png` is retained only as
a look reference; it is absent from the active recipe and runtime catalogue.
Inside each tile the coast bows asymmetrically and the rocky bank varies from a small lip to a broader
shelf. The shared crossing profile is confined to the border and its filter margin.
The water edge grades through partly submerged stones and damp blue-grey rock. On the land side,
olive moss and meadow colours enter painted crevices before the stones disappear into grass.

## Base prompt
Use case: stylized-concept. Asset type: GENERIC REUSABLE tactical game material, opaque seamless surface swatch, NOT a map scene. Reference image attached is ONLY the visual style/material reference for the pale rounded rocks lining the lake: do not copy lake silhouette, layout, water, bridges, or any landmark. Generate a new native high-resolution square 1536x1536 top-down orthographic material of mixed-size natural shoreline boulders: large smooth weathered flattened grey-beige and warm brown-grey rocks with dark damp mossy earth and small pebbles between. Roughly 5 to 7 major boulders across canvas width, varied sizes, natural incomplete distribution, some moss and tiny tufts of olive grass in crevices, shallow cracks, short soft shadows bottom-right, light upper-left. Rich painted microdetail at native resolution like the attached art. Boulders need distinct readable curved silhouettes, pale top faces and dark crevices; not a dense carpet of tiny gravel, not geometric cobblestone pavement. Entire canvas filled with generic rock/earth surface at consistent scale, density and lighting right to edges. Repeating swatch to be sampled into arbitrary shoreline curves, no bank direction, no focal point, no framing, no gradient, no water or lake, no complete coastline, no hex tile or grid, no text. Opaque full-bleed square. Newly authored hand-painted game art.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| lakeside12 | lakeside12 | Land at hex vertices (71,0), (96,40), (71,80); water elsewhere, stony bank along the boundary. |
| lakeside123 | lakeside123 | Land at hex vertices (71,0), (96,40), (71,80), (25,80); water elsewhere, stony bank along the boundary. |
| lakeside126 | lakeside126 | Land at hex vertices (25,0), (71,0), (96,40), (71,80); water elsewhere, stony bank along the boundary. |
| lakeside156 | lakeside156 | Land at hex vertices (25,0), (71,0), (96,40), (0,40); water elsewhere, stony bank along the boundary. |
| lakeside16 | lakeside16 | Land at hex vertices (25,0), (71,0), (96,40); water elsewhere, stony bank along the boundary. |
| lakeside23 | lakeside23 | Land at hex vertices (96,40), (71,80), (25,80), (0,40); water elsewhere, stony bank along the boundary. |
| lakeside34 | lakeside34 | Land at hex vertices (71,80), (25,80), (0,40); water elsewhere, stony bank along the boundary. |
| lakeside3456 | lakeside3456 | Land at hex vertices (25,0), (71,0), (71,80), (25,80), (0,40); water elsewhere, stony bank along the boundary. |
| lakeside56 | lakeside56 | Land at hex vertices (25,0), (71,0), (0,40); water elsewhere, stony bank along the boundary. |
| stream mouths | — | Derived outlet variants open a stream-width channel through a fully-land edge; the same bank and water materials serve every eligible shore and direction. |

## Edges & joins
Land and water are defined on the shared hex vertices, so neighbouring shore, water and meadow tiles
agree at every edge and three-tile corner; the stone bank follows the boundary through the tile.
When one neighbouring stream points into a fully-land edge, the compositor selects the corresponding
`tile/<shore>-outlet<port>` variant. The channel meets the stream at its declared width, leaving the
edge's land corners and all other edge guards intact. The same variant works at any map position.

## Post
`art-review -shore art/pipeline/shore-terrain-v2.json` composes meadow, water and this bank material per
variant from the explicit land vertices. Summer only. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

The selected manifest opts into `authored-rocks-v1`: one uniformly cropped authored bank sample at
35.5×20 logical-pixel periods, with 8 percent seam overlap and no mirrored coset average. The larger
authored rocks retain their scale across the nine masks; the bank extends landward before blending into meadow.
Water uses the matching 71×40-period authored sampler. Existing corner masks remain unchanged.
The coast's broad asymmetric interior bows preserve the original approach at each border. Inland
rock coverage varies through broad patches, with a 2.2–11.5 logical-pixel full-rock reach and a soft
irregular meadow fringe. Bank variation fades out through the filter margin, leaving the existing
eight-pixel geometry guards unchanged. The source paintings are retained without retouching or enlargement.

Material transitions follow the local RefDesign2 lake crop: water remains visible through a shallow
stone shelf, then wet faces become dry rock through wider colour ramps. The meadow fringe widens
around the same boundary, using painted dark crevices and grass marks for moss penetration. Stone
faces retain their native texture; no blur is applied. Shared border colours soften consistently
across compatible neighbours; shore curves, width fields, port states and alpha masks are retained.

The exporter includes all nine tile PNGs and six different compatible assemblies in
`shore-catalogue-2x.png` and `shore-catalogue-4x.png`, using the actual decoded exports without resizing.
The artbook embeds the 4× catalogue. All 234 compatible ordered tile pairs retain matching edge colours.
Reference: local RefDesign2 lake crop, used only for materials and lighting. Native generation used the
built-in image tool; no source pixels were enlarged. Universal bank silhouettes and vegetation remain
visual acceptance tasks.
The 24 single-mouth variants cover every fully-land edge of the nine supported masks. Missing art,
mixed/water edges and multiple facing streams retain the canonical tile; multi-mouth geometry and
winter acceptance remain open. `shore-mouths-4x.png` presents the reusable mouth exports in the artbook.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- sand beaches or dunes
- surf or foam lines
