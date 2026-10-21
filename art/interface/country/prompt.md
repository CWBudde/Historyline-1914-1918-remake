---
id: interface/country
title: Country marks
kind: interface
status: candidate
runtime_ids: country/france, country/germany
masters: france.png, germany.png
---

## Look
Small square country swatches used beside player names.

## Base prompt
Small flat rectangular national colour swatch with three bands and no emblem.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| france | country/france | Vertical blue, cream and red bands. |
| germany | country/germany | Horizontal black, cream and red bands. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/country/` and re-index.

The declared interface PNG masters reuse the matching imagegen candidates in
`art/screens/country/`. These shared controls use the same design and variant prompts.
The catalogue records their native dimensions; candidate acceptance and runtime export bindings remain open.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- emblems, eagles or text in the swatch
