---
id: interface/action
title: Action buttons
kind: interface
status: candidate
runtime_ids: action/attack-disabled, action/attack-focus, action/attack-hover, action/attack-normal, action/attack-pressed, action/contain-disabled, action/contain-focus, action/contain-hover, action/contain-normal, action/contain-pressed, action/exit-disabled, action/exit-focus, action/exit-hover, action/exit-normal, action/exit-pressed, action/move-disabled, action/move-focus, action/move-hover, action/move-normal, action/move-pressed
masters: 
---

## Look
Square command buttons in the unit panel: a dark slate tile with rounded corners and one simple brass
glyph per action, readable at 62x52.

## Base prompt
Flat square interface button, 62x52 logical, dark slate fill with small rounded corners and a thin
blue-grey edge, one bold simple brass glyph centred, no text.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| attack normal | action/attack-normal | A brass sword glyph. Resting look. |
| attack hover | action/attack-hover | A brass sword glyph. Slightly lighter slate fill and a brighter brass edge. |
| attack focus | action/attack-focus | A brass sword glyph. Double brass and cream outline around the resting look, for keyboard focus. |
| attack pressed | action/attack-pressed | A brass sword glyph. Darker, inset fill with a top brass edge, as if pushed in. |
| attack disabled | action/attack-disabled | A brass sword glyph. Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| contain normal | action/contain-normal | A brass crate glyph (board or unload a container). Resting look. |
| contain hover | action/contain-hover | A brass crate glyph (board or unload a container). Slightly lighter slate fill and a brighter brass edge. |
| contain focus | action/contain-focus | A brass crate glyph (board or unload a container). Double brass and cream outline around the resting look, for keyboard focus. |
| contain pressed | action/contain-pressed | A brass crate glyph (board or unload a container). Darker, inset fill with a top brass edge, as if pushed in. |
| contain disabled | action/contain-disabled | A brass crate glyph (board or unload a container). Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| exit normal | action/exit-normal | A brass X glyph (cancel). Resting look. |
| exit hover | action/exit-hover | A brass X glyph (cancel). Slightly lighter slate fill and a brighter brass edge. |
| exit focus | action/exit-focus | A brass X glyph (cancel). Double brass and cream outline around the resting look, for keyboard focus. |
| exit pressed | action/exit-pressed | A brass X glyph (cancel). Darker, inset fill with a top brass edge, as if pushed in. |
| exit disabled | action/exit-disabled | A brass X glyph (cancel). Flat dim slate with a muted grey-blue glyph or edge; no brass. |
| move normal | action/move-normal | A brass right-pointing arrow glyph. Resting look. |
| move hover | action/move-hover | A brass right-pointing arrow glyph. Slightly lighter slate fill and a brighter brass edge. |
| move focus | action/move-focus | A brass right-pointing arrow glyph. Double brass and cream outline around the resting look, for keyboard focus. |
| move pressed | action/move-pressed | A brass right-pointing arrow glyph. Darker, inset fill with a top brass edge, as if pushed in. |
| move disabled | action/move-disabled | A brass right-pointing arrow glyph. Flat dim slate with a muted grey-blue glyph or edge; no brass. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/action/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- more than one glyph per button
