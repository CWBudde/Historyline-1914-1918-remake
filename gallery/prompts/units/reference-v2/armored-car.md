# art/recipes/units/reference-v2/armored-car.json

Recipe: `art/recipes/units/reference-v2/armored-car.json`

- **acceptance**: Original-size silhouette, facing, user visual likeness and runtime acceptance remain review gates. No computed greater-than-90-percent approval.
- **art_direction**: reference-faithful-painted-readability-v2
- **date**: 2026-10-16
- **schema_version**: 1
- **source**: art/source/units/reference-v2/armored-car.png
- **source_sha256**: a76fa1ab724addcd63024acdc1cd54cc52d685724410d4bb93cb197a9a67a129
- **status**: selected candidate
- **tool**: built-in image_gen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/units/icons/armoredCar.png` | unchanged original subject/proportions/camera reference only | `f003d69247cebac62d953de4b638b309657608aa61ca21ae5fc0a49d5c363257` |

## Prompt

```text
Use case: historical-scene. Paint ONE NEW tactical unit atlas for a reference-faithful Historyline remake. The tiny input image is ONLY a subject, proportions, colours and overhead perspective reference; make newly authored high-resolution painted detail. EXACT 1024x1536 transparent PNG, 2 columns by 3 rows, six 512x512 square panels without gutters. Row-major headings N(up), NW(up-left), SW(down-left), S(down), SE(down-right), NE(up-right). Elevated near-overhead strategy view, substantially more overhead than a miniature product photograph: show the long top surface and small front/back faces. Fixed upper-left light, crisp dark contour, light tan/steel equipment highlights and readable broad shapes, not dark shiny green toys. EACH panel contains the SAME single subject facing its heading, independently painted without rotating completed pixels. Saturated technical MAGENTA only on the main painted armour/carriage area for later owner recolouring; wheels, exposed steel, wood and leather retain light natural colours. Bright body midtones, strong restrained dark edges. No decorative baggage absent from the reference. Center object at (256,261). Fit all nontransparent pixels within 38..474 in BOTH axes in every panel, including barrels and wheels; genuinely transparent outside, no ground/shadow/haze/glow/background. All panel scales consistent, no labels, text, grid, insignia or arrows. Result must read as the original subject after Lanczos reduction to 48x48, preserving its long/tall native silhouette rather than a foreshortened bulky toy. Subject: one tall narrow early WWI armoured car matching the icon's rectangular light tan riveted hull, tiny circular turret and four separate dark wheels. North panel subject bounds approximately x=145..367, y=58..458 (aspect 20:36). Make the RECTANGULAR HULL TOP and its length dominant; a very small top turret, modest projecting nose, short gun, four separated wheels. Keep icon-like slender proportions, not a short bulbous armoured jeep. No side sacks, extra gear, giant wheels or oversized barrel.
```
