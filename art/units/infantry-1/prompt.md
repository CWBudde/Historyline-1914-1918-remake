---
id: units/infantry-1
title: Infantry, 1 of 6
kind: unit
status: approved
runtime_ids: unit/infantry-1
masters: atlas.png
---

## Look
The infantry group at a strength of 1: one of its six riflemen, in the formation and scale of the full group
(`units/infantry`), shown when the unit has 1 of 6 left. Same 96x96 logical sprite, anchor (48,49).

## Base prompt
No prompt: the atlas is composed from the riflemen of `units/infantry` (see Post), not generated.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the formation heads up. |
| NW (`nw`) | — | Top-right cell: the formation heads to the upper left. |
| SW (`sw`) | — | Middle-left cell: the formation heads to the lower left. |
| S (`s`) | — | Middle-right cell: the formation heads down. |
| SE (`se`) | — | Bottom-left cell: the formation heads to the lower right. |
| NE (`ne`) | — | Bottom-right cell: the formation heads to the upper right. |

## Post
Composed, not generated: `go run ./tools/unit-squad -src art/units/infantry/riflemen.png -n 1 -out
art/units/infantry-1/atlas.png` cuts the three riflemen out of each facing of the source atlas, scales them to
0.52 and sets 1 of them into a formation that turns with the facing. `art-review -unit
art/pipeline/unit-infantry-1-v1.json` exports it like any family; installed under `assets/units/infantry-1/`.
The game draws this family for an infantry unit of strength 1.

## Avoid
- a different look from `units/infantry`
- men overlapping so much that the count is unclear
