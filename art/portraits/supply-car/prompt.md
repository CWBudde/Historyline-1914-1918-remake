---
id: portraits/supply-car
title: Supply car
kind: portrait
status: approved
runtime_ids: CP_supplyCar, EC_supplyCar
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a tall early supply truck with a canvas-covered bed,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: an early canvas-covered supply truck seen from the front three-quarter: a tall cargo canopy at the top centre, the bonnet to the right, the front wheel low right, soldiers standing along the lower left and one at the far right, trees behind; broad, pale, cloudy
ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork
and selective detail. Fully opaque full-bleed landscape illustration at 105:94, at least 1260x1128,
preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene
naturally to all edges.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_supplyCar, EC_supplyCar | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-supply-car-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/supply-car`. Promote into `assets/portraits/supply-car/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
