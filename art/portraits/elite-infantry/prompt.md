---
id: portraits/elite-infantry
title: Elite infantry
kind: portrait
status: candidate
runtime_ids: CP_eliteInf, EC_eliteInf
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a compact two-man machine-gun team,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: a machine-gun team low in a grassy field: a long thin barrel pointing left from the lower middle, a dark receiver on a tripod, a helmeted gunner crouching on the right working the gun and a small assistant behind the receiver; gun and crew in the bottom half; broad, pale, cloudy
ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork
and selective detail. Fully opaque full-bleed landscape illustration at 105:94, at least 1260x1128,
preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene
naturally to all edges.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_eliteInf, EC_eliteInf | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-elite-infantry-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/elite-infantry`. Promote into `assets/portraits/elite-infantry/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
