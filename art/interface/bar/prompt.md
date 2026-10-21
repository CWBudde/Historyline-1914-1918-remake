---
id: interface/bar
title: Value bars
kind: interface
status: candidate
runtime_ids: bar/negative, bar/neutral, bar/positive, bar/track, bar/unavailable
masters: positive.png, neutral.png, negative.png, track.png, unavailable.png
---

## Look
Small pips and tracks for unit values in the panel: green for good, brass for neutral, red for bad.

## Base prompt
Tiny flat vertical pip with a thin dark outline, filled with one muted signal colour; and a long dark
slate track it sits in.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| positive | bar/positive | Muted green fill. |
| neutral | bar/neutral | Brass fill. |
| negative | bar/negative | Muted brick-red fill. |
| track | bar/track | Long dark slate track, no pip. |
| unavailable | bar/unavailable | Small muted grey X. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/bar/` and re-index.

The declared interface PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/interface/bar/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- bright saturated colours
