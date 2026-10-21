---
id: interface/map
title: Map markers
kind: interface
status: candidate
runtime_ids: map/attack, map/attackable, map/camera, map/cursor, map/editor, map/from, map/invalid, map/path, map/reachable, map/selected
masters: attack.png, attackable.png, camera.png, cursor.png, editor.png, from.png, invalid.png, path.png, reachable.png, selected.png
---

## Look
Hex-shaped overlays on the map: double hex outlines in signal colours, some with a simple glyph,
drawn over terrain and units without hiding them.

## Base prompt
Flat double hex outline (96x80 logical) with a transparent interior, thin lines in one muted signal
colour, an optional small simple glyph in the centre.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| attack | map/attack | Salmon-red outline with a sword glyph. |
| attackable | map/attackable | Salmon-red outline with a small target ring. |
| camera | map/camera | Cream rectangular viewport frame. |
| cursor | map/cursor | Brass outline. |
| editor | map/editor | Cream outline. |
| from | map/from | Teal-green outline with an arrow. |
| invalid | map/invalid | Salmon-red outline with an X. |
| path | map/path | Teal-green outline with an arrow. |
| reachable | map/reachable | Teal-green outline with a centre dot. |
| selected | map/selected | Brass and cream double outline. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/map/` and re-index.

The declared interface PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/interface/map/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- filled hexes that hide terrain
