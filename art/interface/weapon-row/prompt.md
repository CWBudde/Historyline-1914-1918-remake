---
id: interface/weapon-row
title: Weapon row
kind: interface
status: candidate
runtime_ids: weapon-row/disabled, weapon-row/focus, weapon-row/hover, weapon-row/normal, weapon-row/pressed
masters: 
---

## Look
Row frame behind a weapon picture and its live values in the weapon panel.

## Base prompt
Long flat slate row frame, 654x50 logical, thin edge, slightly recessed picture area at the left and an
empty area for live values.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| normal | weapon-row/normal | Resting look. |
| hover | weapon-row/hover | Slightly lighter slate fill and a brighter brass edge. |
| focus | weapon-row/focus | Double brass and cream outline around the resting look, for keyboard focus. |
| pressed | weapon-row/pressed | Darker, inset fill with a top brass edge, as if pushed in. |
| disabled | weapon-row/disabled | Flat dim slate with a muted grey-blue glyph or edge; no brass. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/weapon-row/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
