---
id: portraits/airship
title: Zeppelin airship
kind: portrait
status: approved
runtime_ids: CP_zeppelin
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a silver-grey rigid Zeppelin with two gondolas over a village,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: a WWI rigid Zeppelin airship seen from below and to the side, flying over a quiet countryside on a quiet day: the long silver-grey cigar-shaped hull crossing the upper middle from the rounded nose at the left to the cruciform tail fins at the right, two engine gondolas hanging beneath; small fields, a village church spire and hedgerows far below at the bottom; broad, pale, cloudy
ochre sky. Paint wool, leather, wood, metal and natural surfaces with broad, visible, connected oil-sketch brushwork (no photographic skin or fabric texture)
and selective detail. Fully opaque full-bleed landscape illustration at 105:94, at least 1260x1128,
preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene
naturally to all edges. No national colours, flags, insignia, magenta or faction tint; no frame, bevel, UI, text or numbers; no muzzle flash, explosions or casualties.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_zeppelin | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-airship-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/airship`. Promote into `assets/portraits/airship/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
