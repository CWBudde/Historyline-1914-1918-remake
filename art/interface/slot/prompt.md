---
id: interface/slot
title: Slots
kind: interface
status: candidate
runtime_ids: slot/disabled, slot/focus, slot/hover, slot/normal, slot/pressed
masters: 
---

## Look
Square container and inventory slots: thin brass square frames.

## Base prompt
Flat square slot frame with a thin brass outline and a transparent or very dark centre.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| normal | slot/normal | Resting look. |
| hover | slot/hover | Slightly lighter slate fill and a brighter brass edge. |
| focus | slot/focus | Double brass and cream outline around the resting look, for keyboard focus. |
| pressed | slot/pressed | Darker, inset fill with a top brass edge, as if pushed in. |
| disabled | slot/disabled | Flat dim slate with a muted grey-blue glyph or edge; no brass. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/slot/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
