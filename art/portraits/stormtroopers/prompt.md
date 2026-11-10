---
id: portraits/stormtroopers
title: Stormtroopers
kind: portrait
status: approved
runtime_ids: CP_stormtroopers
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: three German assault troops with stick grenades, a carbine and a light machine gun at a trench edge,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: three German assault troops in coal-scuttle steel helmets crouching ready to advance from a shell-scarred trench edge: grenade bags across their chests, stick grenades in their belts, the front man holding a stick grenade, one with a short carbine, one with a light machine gun on a bipod; men in the lower two thirds facing left; no explosions, no muzzle flash, no casualties; broad, pale, cloudy ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork and selective detail. Fully opaque full-bleed LANDSCAPE illustration at 105:94 (slightly wider than tall), at least 1260x1128, preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene naturally to all edges. Owner-neutral: no magenta, no national colours, no faction tint, no insignia or flags.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_stormtroopers | The scene above; the CP side only. |

## Post
`art-review -portrait art/pipeline/portrait-stormtroopers-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/stormtroopers`. Promote into `assets/portraits/stormtroopers/` and re-index.

The portrait is imagegen output `exec-c78d13e3` (2026-11-10, try 1), 1326x1187 RGB. Its prompt was the `art/style.md` Preamble and Portraits text, then the base prompt above; no attachments.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
- explosions
