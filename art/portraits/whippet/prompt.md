---
id: portraits/whippet
title: Medium Mark A Whippet
kind: portrait
status: approved
runtime_ids: EC_whippet
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a Whippet medium tank, low hull with its tall machine-gun cab at the rear, on a muddy rural track in subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, LANDSCAPE 105:94 aspect, ideally 1680x1504 (at least 1260x1128). Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914-1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, magenta, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: A British Medium Mark A Whippet tank: a low, long hull with tracks running straight along the sides (not rhomboid), a long low engine deck at the front and a tall fixed boxy armoured fighting compartment at the rear with small machine-gun ports. No turret. Left-facing three-quarter view, the engine end leading, driving along a muddy rural track past a broken hedge.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | EC_whippet | The scene above. |

## Post
Selected source: pwhip-01 of gen-2026-11-10, exec-500b7f3e, 1326x1187. The prompt followed the `art/style.md` Preamble and Portraits text, with
`art/portraits/mark-iv/portrait.png` attached for palette, brushwork, sky and light only.
`art-review -portrait art/pipeline/portrait-whippet-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/whippet`. Promote into `assets/portraits/whippet/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
