---
id: interface/banner
title: Banners
kind: interface
status: candidate
runtime_ids: banner/capture, banner/round
masters: capture.png, round.png
---

## Look
Wide slate notice strips that announce a capture or a new round; a brass icon at the left, a thin
divider, and room for live text.

## Base prompt
Wide flat slate banner strip with small rounded corners, a brass icon at the left end, a thin
vertical brass divider after it and an empty area for live text.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| capture | banner/capture | Brass flag icon. |
| round | banner/round | Brass pair of opposed arrows (turn change). |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/banner/` and re-index.

The declared interface PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/interface/banner/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- decorative scrollwork
