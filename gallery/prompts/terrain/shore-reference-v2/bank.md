# shore-bank-reference-v2

Recipe: `art/recipes/terrain/shore-reference-v2/bank.json`

- **cleanup**: Generated master preserved unchanged. Shared lattice sampling, shore shading and Lanczos reduction are performed by Go tools/art-review.
- **date**: 2026-10-20
- **id**: shore-bank-reference-v2
- **review**: DOS-inspired high-resolution redraw. Visual resemblance remains subject to user review; no numeric fidelity percentage claimed.
- **schema_version**: 1
- **source**: art/source/terrain/shore-reference-v2/bank.png
- **source_sha256**: d43b40a2ed64e95cbb6afdacca4a384a39ec59c8a54d05bdcbe77a60e6aa6613
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## original_image_inputs

| path | role | sha256 |
|---|---|---|
| `build/pulse-study/dos-04/maps/00-PARTS.png` | Image 1: original DOS lake palette and shore character reference only; ignored local export. | `58952d5eb3dc629b1346225da3feaa3fb7a7781d13ad3dc2011e9e153bd4bb40` |
| `build/pulse-study/dos-04/PARTS/SCOAS060.png` | Image 2: original DOS shore reference only; ignored local export, never bundled as authored artwork. | `fb408d995aada77cff8bfa01bfc650f808ea5958394dba91c8a750774d9af24b` |

## Prompt

```text
Use case: stylized-concept. Asset type: high-resolution overhead stony shore material for a faithful DOS Historyline restoration. Input image 1 is the original DOS PULSE map; study ONLY the pale narrow banks around the blue lake at lower right. Input image 2 is an original DOS shore tile, for the same chalky stony palette. Paint a full square rectangular material of tightly packed small irregular pale limestone gravel and flat weathered stone chips in muted grey cream and light sand, with olive-grey dirt in the cracks and narrow darker grey-brown wet crevices. The material must resemble the original pale rough lake bank, reconstructed with real fine painted texture instead of enlarged pixels. Top-down orthographic terrain illustration, diffuse overhead light, shallow relief only. Dense irregular pebbles of varied small size with a few scattered larger flattened chips; avoid large boulders, round cobblestone paving, regular tessellation or beach dunes. This is a material swatch, NOT a complete shore tile. Cover every pixel with rock/gravel, opaque to all four edges, consistent brightness and scale, no water, no grass, no horizon, no labels, no border, no hex outline or watermark.
```
