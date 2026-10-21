---
id: interface/button
title: Panel button
kind: interface
status: candidate
runtime_ids: button/disabled, button/focus, button/hover, button/normal, button/pressed
masters: normal.png, hover.png, focus.png, pressed.png, disabled.png
---

## Look
Generic rectangular panel button; live label drawn on top.

## Base prompt
Flat rectangular interface button, dark slate fill, small rounded corners, thin edge; empty face for a
live label.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| normal | button/normal | Resting look. |
| hover | button/hover | Slightly lighter slate fill and a brighter brass edge. |
| focus | button/focus | Double brass and cream outline around the resting look, for keyboard focus. |
| pressed | button/pressed | Darker, inset fill with a top brass edge, as if pushed in. |
| disabled | button/disabled | Flat dim slate with a muted grey-blue glyph or edge; no brass. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/button/` and re-index.

The declared interface PNG masters reuse the matching imagegen candidates in
`art/screens/button/`. These shared controls use the same design and variant prompts.
The catalogue records their native dimensions; candidate acceptance and runtime export bindings remain open.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
