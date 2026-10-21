---
id: portraits/patrol-boat
title: Patrol boat
kind: portrait
status: candidate
runtime_ids: CP_patrolBoat, EC_patrolBoat
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a small armoured patrol boat,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: a small armoured patrol boat in side three-quarter view on rippling water: a low bow on the left, a boxy wheelhouse right of centre, a crewman on the left deck; the boat in the lower half; broad, pale, cloudy
ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork
and selective detail. Fully opaque full-bleed landscape illustration at 105:94, at least 1260x1128,
preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene
naturally to all edges.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_patrolBoat, EC_patrolBoat | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-patrol-boat-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/patrol-boat`. Promote into `assets/portraits/patrol-boat/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
