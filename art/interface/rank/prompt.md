---
id: interface/rank
title: Ranks
kind: interface
status: candidate
runtime_ids: rank/00, rank/01, rank/02, rank/03, rank/04, rank/05, rank/06, rank/07, rank/08, rank/09, rank/10
masters: 
---

## Look
Experience ranks: zero to four brass stars, then the same inside a ring of brass laurel dashes, up to
one large star in a double ring.

## Base prompt
Small flat brass rank emblem on transparency, 88x88 logical, simple five-pointed stars, optional
ring of short brass laurel dashes.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| 00 | rank/00 | One dim slate star (no rank). |
| 01 | rank/01 | One brass star. |
| 02 | rank/02 | Two brass stars. |
| 03 | rank/03 | Three brass stars. |
| 04 | rank/04 | Four brass stars. |
| 05 | rank/05 | One large star in a laurel ring. |
| 06 | rank/06 | One star in a laurel ring. |
| 07 | rank/07 | Two stars in a laurel ring. |
| 08 | rank/08 | Three stars in a laurel ring. |
| 09 | rank/09 | Four stars in a laurel ring. |
| 10 | rank/10 | One large star in a double laurel ring. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/rank/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- military insignia of real armies
