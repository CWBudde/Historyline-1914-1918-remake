---
id: portraits/tunnellers
title: Tunnellers
kind: portrait
status: approved
runtime_ids: CP_tunnellers, EC_tunnellers
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: two tunnelling miners at a timbered shaft head with a windlass,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: a tunnelling party at a timbered shaft head behind the lines: a square frame of heavy timbers around a dark shaft opening with a wooden windlass and a rope going down, one miner in a cloth cap with rolled sleeves leaning on a pick, another carrying a sandbag on his shoulder, a pile of filled sandbags and spoil beside the shaft; the shaft head and miners in the lower middle; no explosion; broad, pale, cloudy ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork and selective detail. Fully opaque full-bleed LANDSCAPE illustration at 105:94 (slightly wider than tall), at least 1260x1128, preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene naturally to all edges. Owner-neutral: no magenta, no national colours, no faction tint, no insignia or flags.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_tunnellers, EC_tunnellers | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-tunnellers-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/tunnellers`. Promote into `assets/portraits/tunnellers/` and re-index.

The portrait is imagegen output `exec-666d34f5` (2026-11-10, try 1), 1326x1187 RGB. Its prompt was the `art/style.md` Preamble and Portraits text, then the base prompt above; no attachments.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
- explosions or mine craters
