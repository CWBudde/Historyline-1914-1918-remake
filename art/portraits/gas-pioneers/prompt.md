---
id: portraits/gas-pioneers
title: Gas pioneers
kind: portrait
status: approved
runtime_ids: CP_gasPioneers, EC_gasPioneers
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: two masked gas pioneers at a row of gas cylinders in a trench,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: a gas-pioneer team at a front-line trench parapet: two soldiers in steel helmets and gas masks with round eyepieces and snout filters kneel beside a row of four tall upright grey-green steel gas cylinders set into the sandbagged parapet, a rubber hose running from a cylinder valve over the parapet; one man turns a valve with a spanner; men and cylinders fill the lower middle of the picture; no visible gas cloud, calm; broad, pale, cloudy ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork and selective detail. Fully opaque full-bleed LANDSCAPE illustration at 105:94 (slightly wider than tall), at least 1260x1128, preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene naturally to all edges. Owner-neutral: no magenta, no national colours, no faction tint, no insignia or flags.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_gasPioneers, EC_gasPioneers | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-gas-pioneers-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/gas-pioneers`. Promote into `assets/portraits/gas-pioneers/` and re-index.

The portrait is imagegen output `exec-589fe31a` (2026-11-10, try 1), 1327x1186 RGB. Its prompt was the `art/style.md` Preamble and Portraits text, then the base prompt above; no attachments.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
- a visible gas cloud
