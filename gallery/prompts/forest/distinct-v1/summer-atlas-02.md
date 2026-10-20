# summer-atlas-02

Recipe: `art/recipes/forest/distinct-v1/summer-atlas-02.json`

- **date**: 2026-10-15
- **id**: summer-atlas-02
- **output_manifest**: art/forest-terrain-v1.json
- **schema_version**: 1
- **source**: art/source/forest/distinct-v1/crowns-summer.png
- **source_sha256**: a1f12a95d12e48cefd79d1236efa0b30a93bb98e3e5a3af46da5cdf95e4e5d72
- **status**: reviewed-source-for-provisional-forest-contexts
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true
- **use_case**: precise-object-edit

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/forest/distinct-v1/studies/crowns-summer-01.png` | new-atlas-margin-correction-target | `95eab31f8655a06f00fb6868f98821e34309338497da7834938613fbe014994a` |

## Prompt

```text
Use case: precise-object-edit. Image1 is the NEW transparent 2x2 FOUR-CROWN tree atlas. Change ONLY framing/scale within each quadrant. Keep exact square canvas, same four tree designs/species/material/color/lighting and strict vertical overhead camera, same2x2 order. SHRINK EACH TREE INCLUDING ALL ITS TWIGS AND SHADOWS to about75percent of its present diameter, keeping each center exactly at (25percent,25percent),(75percent,25percent),(25percent,75percent),(75percent,75percent) of canvas. Every complete crown must fit well inside its own quadrant's central65percent square. Leave at least17percent of EACH QUADRANT'S WIDTH as genuinely fully transparent clear margin along ALL FOUR sides of that quadrant. Thus no part of any crown, twig or shadow may approach the hidden quadrant boundaries, and the four trees remain cleanly separate with wide empty gaps. Background alpha exactlyZERO throughout these margins; remove faint loose alpha/noise pixels outside the complete crowns. Preserve fine smooth leaf/needle artwork and antialiased natural edges, no blue/cyan/colored fringe, no opaque ground bases or circles. No drawn grid/separators, words, extra objects or pixelation. Keep existing top-left and top-right broadleaf trees and bottom-left and bottom-right radial conifer crowns intact; no redesign. Real alpha transparency.

```
