---
id: screens/loading-050
title: Loading 50%
kind: screen
status: approved
runtime_ids: screen/loading-050
masters: frame.png
---

## Look
Full-screen frame, 640x400 logical, drawn over the menu backdrop: a centred body panel with a half-filled progress bar. Translucent slate
surfaces with thin brass hairlines; titles, labels and controls are live and drawn on top.

## Base prompt
Flat full-screen interface frame, 640x400, for a WWI strategy game menu: a thin brass border inset 8
pixels, a dark slate title bar across the top with a brass rule under it, and a centred body panel with a half-filled progress bar as
dark, slightly translucent slate panels with small rounded corners, thin blue-grey edges and a short
brass tick at each panel's top-left corner. All panels empty; no text or controls.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| frame | screen/loading-050 | The layout above. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/screens-v2.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/screens/loading-050/` and re-index.

The shared painted candidate is `art/screens/mission-selection/illustration.png`; reuse it behind this frame
when a feature binding is added. It is not consumed by the current vector exporter.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/loading-050/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- painted scenery inside panels
