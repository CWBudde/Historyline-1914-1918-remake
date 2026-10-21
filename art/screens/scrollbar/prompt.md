---
id: screens/scrollbar
title: Scrollbar
kind: screen
status: candidate
runtime_ids: scrollbar/thumb, scrollbar/track
masters: track.png, thumb.png
---

## Look
Scrollbar of long text panels.

## Base prompt
Flat narrow vertical scrollbar: a dark slate track and a brass-edged thumb.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| track | scrollbar/track | Dark track. |
| thumb | scrollbar/thumb | Thumb with brass edge. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/screens-v2.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/screens/scrollbar/` and re-index.

The declared frame/control PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/screens/scrollbar/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

The generated track retains a soft halo; edge cleanup remains part of acceptance before export.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
