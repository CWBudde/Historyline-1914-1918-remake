---
id: portraits/transport-ship
title: Transport ship
kind: portrait
status: revised
runtime_ids: portrait/transport-ship
masters: portrait.png
---

## Look
Owner-neutral transport ship in a quiet period scene, matching the existing portrait
paintings' subdued ochre and sepia palette, cloudy sky and connected brushwork.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, 105:94 landscape aspect ideally 1680x1504 or larger. Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914–1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: An unarmed early steam cargo transport with a deep hull, a single central funnel, two cargo masts with booms and covered cargo holds, bow to the left, on quiet coastal water.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | portrait/transport-ship | Selected owner-neutral painting; gameplay class binding pending. |

## Post
Selected source with a centred exact 105:94 crop, never stretched or enlarged.
`art-review -portrait art/pipeline/portrait-transport-ship-v1.json` exports 840×752 (4×),
420×376 (2×) and 210×188 (1×) into `assets/portraits/transport-ship/`. Gameplay class binding
and in-engine panel acceptance remain open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels, labels or UI
- photographic or 3D rendering
- action spectacle, muzzle flash or casualties
