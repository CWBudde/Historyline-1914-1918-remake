---
id: interface/symbol
title: Symbols
kind: interface
status: candidate
runtime_ids: symbol/air, symbol/ammo, symbol/armour, symbol/capture, symbol/empty, symbol/fuel, symbol/ground, symbol/infinity, symbol/repair, symbol/sight, symbol/size, symbol/skill-unknown, symbol/turn, symbol/water, symbol/weight
masters: 
---

## Look
Small brass glyphs for unit values and terrain properties in the panel.

## Base prompt
Tiny flat solid brass glyph on transparency, simple silhouette, readable at 16 pixels.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| air | symbol/air | Glyph for air. |
| ammo | symbol/ammo | Glyph for ammo. |
| armour | symbol/armour | Glyph for armour. |
| capture | symbol/capture | Glyph for capture. |
| empty | symbol/empty | Glyph for empty. |
| fuel | symbol/fuel | Glyph for fuel. |
| ground | symbol/ground | Glyph for ground. |
| infinity | symbol/infinity | Glyph for infinity. |
| repair | symbol/repair | Glyph for repair. |
| sight | symbol/sight | Glyph for sight. |
| size | symbol/size | Glyph for size. |
| skill-unknown | symbol/skill-unknown | Glyph for skill unknown. |
| turn | symbol/turn | Glyph for turn. |
| water | symbol/water | Glyph for water. |
| weight | symbol/weight | Glyph for weight. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/symbol/` and re-index.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- outlines or multiple colours
