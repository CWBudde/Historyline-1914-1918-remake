---
id: portraits/field-ambulance
title: Field ambulance
kind: portrait
status: approved
runtime_ids: CP_ambulance, EC_ambulance
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a canvas-covered motor ambulance with a red cross, bearers loading a covered stretcher,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Asset: owner-neutral WWI strategy unit portrait. Scene: a WWI motor field ambulance on a muddy country road seen from the front three-quarter: a tall boxy canvas-covered ambulance body at the centre with a small red cross on a white square on its side, the short bonnet and radiator to the right, spoked front wheel low right; two stretcher bearers carrying a covered stretcher towards the open rear at the lower left, a medical orderly in a white armband beside them, poplars behind; broad, pale, cloudy
ochre sky. Paint wool, leather, wood, metal and natural surfaces with broad, visible, connected oil-sketch brushwork (no photographic skin or fabric texture)
and selective detail. Fully opaque full-bleed landscape illustration at 105:94, at least 1260x1128,
preferably 1680x1504, the essential subject safe inside the centred 105:94 crop; extend the scene
naturally to all edges. No national colours, flags, insignia, magenta or faction tint; no frame, bevel, UI, text or numbers; no muzzle flash, explosions or casualties.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_ambulance, EC_ambulance | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-field-ambulance-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/field-ambulance`. Promote into `assets/portraits/field-ambulance/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
