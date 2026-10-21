---
id: terrain/lakeside
title: Lake shore
kind: terrain
status: candidate
runtime_ids: lakeside12, lakeside123, lakeside126, lakeside156, lakeside16, lakeside23, lakeside34, lakeside3456, lakeside56
masters: bank.png
---

## Look
Shore between lake and meadow. In the reference a band of rounded pale boulders and stones (cream,
warm grey, sand) separates clear blue-teal water from grass, with small shrubs on the land side.

## Base prompt
Opaque square material swatch of a stony lake shore seen from above: tightly packed rounded and
flattened pale stones and pebbles in cream, warm grey and light sand, a few larger smooth boulders,
olive-grey dirt and moss between them, darker wet crevices; even scale and brightness to every edge.
No water, grass, horizon, paving or regular tessellation.

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
`art-review -shore art/pipeline/shore-terrain-v1.json` composes meadow, water and this bank material per
variant from the explicit land vertices. Summer only. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- sand beaches or dunes
- surf or foam lines
