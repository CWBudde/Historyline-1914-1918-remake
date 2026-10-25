---
id: interface/ownership
title: Ownership marks
kind: interface
status: approved
runtime_ids: ownership/cp, ownership/cp-factory, ownership/depotOwnership, ownership/ec, ownership/ec-factory, ownership/neutral, ownership/neutral-factory
masters: cp.png, ec.png, neutral.png, cp-factory.png, ec-factory.png, neutral-factory.png, depot-atlas.png
---

## Look
Small owner strips on captured buildings and depots: the owner colour as a short bar.

## Base prompt
Small flat owner strip with fine diagonal hatching in the owner colour, thin dark outline.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cp | ownership/cp | Central Powers strip. |
| ec | ownership/ec | Entente strip. |
| neutral | ownership/neutral | Neutral strip. |
| cp-factory | ownership/cp-factory | Central Powers strip tilted 60 degrees for factory roofs. |
| ec-factory | ownership/ec-factory | Entente strip tilted 60 degrees. |
| neutral-factory | ownership/neutral-factory | Neutral strip tilted 60 degrees. |
| depot atlas | ownership/depotOwnership | Three stacked strips (neutral, cp, ec) for depots. |

## Post
Drawn as vectors by `tools/interface-art` from its Go catalogue (one path model to SVG and a 4x raster, reduced 4x to 2x to 1x with Lanczos3): `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`. A painted replacement would be exported the same way at 1x, 2x and 4x; promote into `assets/interface/ownership/` and re-index.

The declared interface PNG masters are separate imagegen candidates made directly from the Base
prompt and each variant delta, using the corresponding `assets/interface/ownership/` render as a
layout reference. The generated catalogue records their native dimensions. These candidate sources
await visual acceptance and export bindings; the current runtime vector exports remain separate.

## Avoid
- baked text, numbers or values
- gradients, glows or glossy bevels
- photographic or 3D-rendered surfaces
- flags or national emblems
