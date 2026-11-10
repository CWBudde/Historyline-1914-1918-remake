---
id: portraits/mark-v
title: Mark V tank
kind: portrait
status: approved
runtime_ids: EC_markV
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a Mark V heavy tank with its raised rear cupola, unditching rails and timber beam, beside a ditch in a field, in subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, LANDSCAPE 105:94 aspect, ideally 1680x1504 (at least 1260x1128). Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914-1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, magenta, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: A British Mark V heavy tank with large rhomboid wraparound tracks, side gun sponsons, a forward driver cab, a raised boxy observation cupola on the rear of the roof and two long unditching-beam rails along the tops of the tracks carrying a heavy square timber beam at the rear. No turret, no trailing steering wheels. Left-facing three-quarter view crossing a shallow trench in a churned field, the cupola and beam clearly visible.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | EC_markV | The scene above. |

## Post
Selected source: pmk5-01 of gen-2026-11-10, exec-aa3bbeed, 1326x1187. The prompt followed the `art/style.md` Preamble and Portraits text, with
`art/portraits/renault-ft17/portrait.png` attached for palette, brushwork, sky and light only.
`art-review -portrait art/pipeline/portrait-mark-v-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/mark-v`. Promote into `assets/portraits/mark-v/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
