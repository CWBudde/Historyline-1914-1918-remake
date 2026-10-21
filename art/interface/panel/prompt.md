---
id: interface/panel
title: Panels
kind: interface
status: candidate
runtime_ids: panel/battle, panel/content, panel/terrain, panel/weapon
masters: 
---

## Look
Background panels of the battle and unit screens: dark slate surfaces with recessed slots and thin
brass dividers, empty of text.

## Base prompt
Flat dark slate interface panel with small rounded corners, a thin blue-grey edge, recessed darker
slots for portraits, values and weapon rows, thin brass divider lines; no text.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| battle | panel/battle | Wide 840x156 strip: header bar, four recessed portrait slots and a brass centre divider. |
| content | panel/content | Tall 527x670 unit information panel. |
| terrain | panel/terrain | Wide 654x110 terrain information strip. |
| weapon | panel/weapon | 670x286 weapon list panel with four row slots. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/panel/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- ornamental borders
