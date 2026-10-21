---
id: terrain/lakeside
title: Lake shore
kind: terrain
status: candidate
runtime_ids: lakeside12, lakeside123, lakeside126, lakeside156, lakeside16, lakeside23, lakeside34, lakeside3456, lakeside56
masters: lakeside-bank-shared-v2.png, lakeside-bank-shared-v1.png
---

## Look
Shore between lake and meadow with rounded warm grey and brown-grey boulders, damp dark crevices
and olive moss. Selected master `lakeside-bank-shared-v2.png` is a native 1254×1254 opaque built-in image-tool output;
the request for 1536×1536 did not determine the delivered dimensions.

## Base prompt
Use case: stylized-concept. Generate a NEW opaque square 1536x1536 material swatch for rural European stony lake shores in a WWI tactical game. The attached PULSE painting is a style/color reference, the attached crater painting is a reference for native detailed painted material finish only; do not copy either subject or scene. Strict vertical nadir, no tilt. The entire canvas is irregular naturally rounded and flattened boulders, cobbles and pebbles in weathered warm grey, earthy brown-grey and muted beige, with darker charcoal wet crevices, damp brown soil, olive moss in gaps and occasional small grains. Medium sized rounded rocks occupy about 6 to 12 percent of the canvas width each, with smaller companion stones, many irregular shapes, varied orientation. Enough rounded sculptural boulders to read clearly when this swatch becomes a narrow bank on a small map. Fine cracks, tiny mica and moss detail, restrained upper-left highlights, short soft lower-right shade. Natural continuous hand-painted digital game illustration, crisp antialiased edges and fine brushwork like the crater reference. Uniform brightness and density right to every edge. This is one continuous opaque material, no framing. Avoid white chalk/gravel dominance, flat white fringe, regular cobblestone paving, all-white stones, crystal forms, cartoony blobs, visible repeating pattern, grid, text, hex outlines, water, grass fields, horizon, gradient, vignette, pixel art, photoreal photo, 3D render. 1536x1536 square.

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

## Edges & joins
Land and water are defined on the shared hex vertices, so neighbouring shore, water and meadow tiles
agree at every edge and three-tile corner; the stone bank follows the boundary through the tile.

## Post
`art-review -shore art/pipeline/shore-terrain-v2.json` composes meadow, water and this bank material per
variant from the explicit land vertices. Summer only. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

The selected manifest opts into `authored-rocks-v1`: one uniformly cropped authored bank sample at
35.5×20 logical-pixel periods, with 8 percent seam overlap and no mirrored coset average. Rounded
rock faces remain about 2–5 logical pixels; the bank extends landward before blending into meadow.
Water uses the matching 71×40-period authored sampler. Existing corner masks remain unchanged.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- sand beaches or dunes
- surf or foam lines
