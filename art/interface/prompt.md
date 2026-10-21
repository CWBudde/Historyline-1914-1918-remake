---
id: interface
title: Interface metadata
kind: interface
status: candidate
runtime_ids: animations, bindings
masters: 
---

## Look
No image: animation timing and anchors (`animations.json`) and the binding of interface art to game
states (`bindings.json`) that `tools/interface-art` writes with the interface bundle.

## Base prompt
None. These files are generated data, not art; edit the interface catalogue instead.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| animations | animations | Frame lists, durations, reduced-motion frames and anchors. |
| bindings | bindings | Which interface art each game state uses. |

## Post
Written by `go run ./tools/interface-art -manifest art/pipeline/interface-v1.json -out <fresh dir>`;
promote into `assets/interface/` and re-index.

## Avoid
- hand edits that diverge from the catalogue
