---
id: screens/new-map
title: New map
kind: screen
status: candidate
runtime_ids: screen/new-map
masters: 
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: one wide form panel. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and one wide form panel as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/new-map | The layout above. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/screens/new-map/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- painted scenery inside panels
