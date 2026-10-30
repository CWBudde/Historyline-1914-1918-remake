---
id: units/elite-infantry-1
title: Elite infantry, strength 1-3
kind: unit
status: approved
runtime_ids: unit/elite-infantry-1
masters: atlas.png
---

## Look
The elite infantry at a strength of 1 to 3 of 6: the two-man machine-gun team (`team-full.png` of
`units/elite-infantry`), drawn smaller than the full section of `units/elite-infantry`. Same 96x96 logical sprite,
anchor (48,49).

## Base prompt
No prompt: the atlas is the two-man team atlas of `units/elite-infantry` scaled, not generated (see Post).

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| N (`n`) | — | Top-left cell: the team points up. |
| NW (`nw`) | — | Top-right cell: the team points to the upper left. |
| SW (`sw`) | — | Middle-left cell: the team points to the lower left. |
| S (`s`) | — | Middle-right cell: the team points down. |
| SE (`se`) | — | Bottom-left cell: the team points to the lower right. |
| NE (`ne`) | — | Bottom-right cell: the team points to the upper right. |

## Post
`go run ./tools/unit-squad -src art/units/elite-infantry/team-full.png -n 0 -scale 0.55 -out
art/units/elite-infantry-1/atlas.png`, then `art-review -unit art/pipeline/unit-elite-infantry-1-v1.json`;
installed under `assets/units/elite-infantry-1/`. The game draws this family for an elite infantry unit of
strength 1 to 3.

## Avoid
- a different look from `units/elite-infantry`
