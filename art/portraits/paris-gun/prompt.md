---
id: portraits/paris-gun
title: Paris Gun
kind: portrait
status: approved
runtime_ids: CP_parisGun
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: the Paris Gun's extremely long braced barrel on its railway carriage in a wooded clearing,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: the Paris Gun on its railway carriage in side view in a wooded clearing: an extremely long, slender gun barrel braced by a truss of rods on top, rising from the breech on the heavy multi-axle railway carriage at lower right to the muzzle far away at upper left, the carriage on rails along the bottom, a few small gunners beside the carriage for scale, camouflage netting poles and tall trees at the edges; broad, pale, cloudy
ochre sky. Paint wool, leather, wood, metal and natural surfaces with broad, visible, connected oil-sketch brushwork (no photographic skin or fabric texture)
and selective detail. Fully opaque full-bleed landscape illustration at 105:94, at least 1260x1128,
preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene
naturally to all edges. No national colours, flags, insignia, magenta or faction tint; no frame, bevel, UI, text or numbers; no muzzle flash, explosions or casualties.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_parisGun | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-paris-gun-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/paris-gun`. Promote into `assets/portraits/paris-gun/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
