---
id: portraits/schneider-ca1
title: Schneider CA1 tank
kind: portrait
status: approved
runtime_ids: EC_schneider
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a Schneider CA1 with its pointed wire-cutter prow and right-front 75 mm sponson, crossing broken wire in a field, in subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, LANDSCAPE 105:94 aspect, ideally 1680x1504 (at least 1260x1128). Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914-1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, magenta, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: A French Schneider CA1 tank: a tall box-like riveted armoured hull on a short Holt tractor track chassis, a pointed boat-like wire-cutter prow projecting at the front, a short stubby 75 mm gun in a small sponson on the hull's right front side. No turret. Three-quarter view from the vehicle's right front, the prow pointing to the right of the picture, so the gun sponson faces the viewer; crossing a field of broken barbed-wire entanglements, muted French horizon-blue-grey paint toned down to the sepia palette.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | EC_schneider | The scene above. |

## Post
Selected source: pca1-01 of gen-2026-11-10, exec-6f1ab71d, 1326x1187. The prompt followed the `art/style.md` Preamble and Portraits text, with
`art/portraits/mark-iv/portrait.png` attached for palette, brushwork, sky and light only.
`art-review -portrait art/pipeline/portrait-schneider-ca1-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/schneider-ca1`. Promote into `assets/portraits/schneider-ca1/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
