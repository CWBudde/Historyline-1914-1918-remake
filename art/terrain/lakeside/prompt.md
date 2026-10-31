---
id: terrain/lakeside
title: Lake shore
kind: terrain
status: approved
runtime_ids: lakeside12, lakeside123, lakeside126, lakeside156, lakeside16, lakeside23, lakeside34, lakeside3456, lakeside56
masters: lakeside-bank-shared-v3.png, lakeside-bank-shared-v2.png, lakeside-bank-shared-v1.png, lakeside-lake-summer-v3.png, lakeside-bank-autumn-v1.png
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

### Autumn bank material edit (2026-10-31)

`lakeside-bank-autumn-v1.png` (`exec-0bda9774`, native 1254×1254 RGB, unchanged) is the autumn edit of
`lakeside-bank-shared-v3.png`: every stone in place, the weeds and reeds dry golden straw, the moss
bronze, the earth damper, about 30 small leaves. Whole RGB 109,85,61 (L\* 36.9, hue 68; summer 117,97,61,
L\* 41.0). Selected: `-season autumn` exports take it wherever the bank material is sampled, the stream
banks (`tile/stream*_au`) and, once exported, the shore and lakeside banks. Attached:
`lakeside-bank-shared-v3.png`. The prompt was:

```text
Seasonal edit of the supplied stony river-bank material: change the season to a damp late-September/October day only, preserve every outline and position. Keep the square canvas exactly as supplied (1254x1254, opaque, full-bleed, seamless left-right and top-bottom, strict vertical plan view).
Every boulder, stone, pebble, crack and plant tuft stays exactly where it is, with its size and outline; the stones keep their colours and lightness. Change only:
- The green weeds, grass tufts and moss turn autumnal: dry golden straw and ochre reeds and grass (lit about RGB 185,150,75, shaded about 95,72,30), faded olive-bronze moss on the stones (about 110,95,40), a few rust-brown dead leaves of the small plants (about 125,70,35). No fresh green left except a little dull olive deep inside a few tufts.
- The earth between the stones is slightly damp and darker (about 6 L* darker than now), dark brown, not black.
- A few single fallen leaves (about 40 over the canvas, each 8 to 14 px long, ochre and rust brown) lying flat between the stones.
No snow, no water, no new stones, no orange glow, no text, no vignette, no visible tiling.
Corrections from the measured previous edit: the plants came out bright orange and rust (whole hue 67, a* +8), and its fallen leaves were about 2 times too large (20 to 30 px) and too many, so the bank read orange. Make the tufts and reeds pale golden straw and dull ochre (about RGB 175,145,80 lit, 95,75,35 shaded), never orange or red; keep only a few rust-brown accents. Paint only about 30 single leaves, each 8 to 14 px long. Whole average about RGB 115,92,60, hue about 75.

Image 1 is the selected summer bank material (lakeside-bank-shared-v3) to edit: keep its layout, every stone's position and size, and its scale; change only the season as described.
```

`art-review -shore art/pipeline/shore-terrain-v2.json` composes meadow, water and this bank material per
variant from the explicit land vertices. `shore-terrain-winter-v1.json` takes the snow meadow
(`grass-winter-v1.png`) and sets `season: winter`, which turns the dry stones cold grey with snow on their
light tops; the wet faces stay dark. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

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
winter water remain open. `shore-mouths-4x.png` presents the reusable mouth exports in the artbook.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- sand beaches or dunes
- surf or foam lines
