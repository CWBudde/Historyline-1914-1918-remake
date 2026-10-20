# shore-water-reference-v2

Recipe: `art/recipes/terrain/shore-reference-v2/water.json`

- **cleanup**: Generated master preserved unchanged. Shared lattice sampling, shore shading and Lanczos reduction are performed by Go tools/art-review.
- **date**: 2026-10-20
- **id**: shore-water-reference-v2
- **review**: DOS-inspired high-resolution redraw. Visual resemblance remains subject to user review; no numeric fidelity percentage claimed.
- **schema_version**: 1
- **source**: art/source/terrain/shore-reference-v2/water.png
- **source_sha256**: 007b9babe0e69db5e628352a38dae0f1502ac42513bbf4018ad3df82f76c4c31
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## original_image_inputs

| path | role | sha256 |
|---|---|---|
| `build/pulse-study/dos-04/PARTS/SSSEA083.png` | Original DOS reference only; ignored local export, never bundled as authored artwork. | `7e2ab820355398f16de88a88095eefe260a3345f8575dd4a3ca323a91b0bec74` |
| `build/pulse-study/dos-04/maps/00-PARTS.png` | Original DOS lake palette and shore character reference only; ignored local export. | `58952d5eb3dc629b1346225da3feaa3fb7a7781d13ad3dc2011e9e153bd4bb40` |

## Prompt

```text
Use case: stylized-concept. Asset type: faithful high-resolution restoration of the attached original DOS Historyline water tile. Input image 1 is the original DOS 24x24 water tile; input image 2 is the original DOS map for contextual reference to its lake. Paint a full square opaque field of exactly this original subdued dark grey-teal lake water with short broken pale grey-green and off-white ripples, top-down. CRITICAL COLOR MATCH: the original opaque water pixels average RGB(53,98,100), hex #356264. Match this mean color closely. Dark troughs around #224748, main grey-teal body around #356264 to #467477, a few thin bright glints around #9DB5AC and very sparse near-white tips. LOW SATURATION, subdued grey-green-blue, medium-dark overall. Do NOT make bright cyan or tropical turquoise water; no vivid aqua, no neon blue, no saturated white foamy ocean. Match the original's scattered small horizontally broken ripple marks, enlarging their hand-painted detail cleanly rather than enlarging pixel blocks. Retain clear small-scale contrast and uneven ripple spacing. Flat orthographic strategy-game surface with diffuse lighting, uniform overall color, no large gradients or reflected scenery. Even detail right to all four edges for periodic sampling. No shores, grass, stones, objects, horizon, hex outline, grid, labels, watermark or transparency. Return only the muted grey-teal water material. The #356264 overall average is more important than prettifying the water.
```
