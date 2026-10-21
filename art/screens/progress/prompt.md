---
id: screens/progress
title: Progress bar
kind: screen
status: candidate
runtime_ids: progress/fill, progress/track
masters: 
---

## Look
Loading and saving progress bar.

## Base prompt
Flat long progress bar: a dark recessed slate track and a brass fill segment.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| track | progress/track | Empty dark track. |
| fill | progress/fill | Brass fill segment. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/screens-v2.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/screens/progress/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
