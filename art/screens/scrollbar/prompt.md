---
id: screens/scrollbar
title: Scrollbar
kind: screen
status: candidate
runtime_ids: scrollbar/thumb, scrollbar/track
masters: 
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

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
