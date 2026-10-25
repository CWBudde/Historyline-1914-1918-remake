---
id: screens/button
title: Menu button
kind: screen
status: approved
runtime_ids: button/disabled, button/focus, button/hover, button/normal, button/pressed
masters: normal.png, hover.png, focus.png, pressed.png, disabled.png
---

## Look
Menu button of the full-screen frames; live label on top.

## Base prompt
Flat rectangular menu button, dark slate fill, small rounded corners, thin edge; empty face.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| normal | button/normal | Resting look. |
| hover | button/hover | Slightly lighter slate fill and a brighter brass edge. |
| focus | button/focus | Double brass and cream outline around the resting look, for keyboard focus. |
| pressed | button/pressed | Darker, inset fill with a top brass edge, as if pushed in. |
| disabled | button/disabled | Flat dim slate with a muted grey-blue glyph or edge; no brass. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/screens/button/` and re-index.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/button/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
