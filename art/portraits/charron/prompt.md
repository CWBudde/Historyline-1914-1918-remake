---
id: portraits/charron
title: Charron armoured car
kind: portrait
status: approved
runtime_ids: EC_charron
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a long early armoured car with a small round turret,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: an early armoured car in left-facing side view: a sloping rounded bonnet on the left, four spoked wheels, a tall riveted cabin on the right with a small round roof turret, filling the lower two thirds; broad, pale, cloudy
ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork
and selective detail. Fully opaque full-bleed landscape illustration at 105:94, at least 1260x1128,
preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene
naturally to all edges.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | EC_charron | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-charron-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/charron`. Promote into `assets/portraits/charron/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
