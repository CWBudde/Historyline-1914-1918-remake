---
id: screens/country
title: Side marks
kind: screen
status: candidate
runtime_ids: country/france, country/germany
masters: 
---

## Look
Country swatches on the side-selection cards.

## Base prompt
Flat rectangular national colour swatch with three bands and no emblem, 88x88 logical canvas.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| france | country/france | Vertical blue, cream and red bands. |
| germany | country/germany | Horizontal black, cream and red bands. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/screens-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/screens/country/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- emblems, eagles or text
