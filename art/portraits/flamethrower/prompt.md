---
id: portraits/flamethrower
title: Flamethrower
kind: portrait
status: approved
runtime_ids: CP_flamethrower, EC_flamethrower
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a two-man flamethrower team with a backpack tank and an unlit lance in a trench,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: a two-man flamethrower team crouching in a shallow trench before an assault: the rear man in a steel helmet carries a large backpack tank of two round steel cylinders on a frame, a hose leads to the front man who holds a long straight metal lance pointing to the left over the trench lip; the lance is unlit, no flame, no fire, no smoke; the two men fill the lower middle; broad, pale, cloudy ochre sky. Paint wool, leather, wood, metal and natural surfaces with restrained connected brushwork and selective detail. Fully opaque full-bleed LANDSCAPE illustration at 105:94 (slightly wider than tall), at least 1260x1128, preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene naturally to all edges. Owner-neutral: no magenta, no national colours, no faction tint, no insignia or flags.
CORRECTION TO THE PREVIOUS TRY: it painted the lance as a scoped rifle with a stock and a muzzle brake, and the hose ended at the front man's belt. The flamethrower lance is a plain, long, thin steel tube about 1.2 m long with a small conical nozzle at its tip and a simple handle and valve lever at its rear end, held in both hands; NO stock, NO scope, NO muzzle brake, NO magazine, NO trigger guard of a rifle. The thick dark rubber hose runs visibly all the way from the bottom of the two backpack tanks to the rear end of the lance. Unlit: no flame, no pilot light.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_flamethrower, EC_flamethrower | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-flamethrower-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/flamethrower`. Promote into `assets/portraits/flamethrower/` and re-index.

The portrait is imagegen output `exec-24da3cc6` (2026-11-10, try 2), 1327x1186 RGB. Its prompt was the `art/style.md` Preamble and Portraits text, then the base prompt above; no attachments.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
- flame, fire or smoke
