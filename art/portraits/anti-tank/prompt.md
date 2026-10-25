---
id: portraits/anti-tank
title: Anti-tank weapons
kind: portrait
status: approved
runtime_ids: CP_antiTank, EC_antiTank
masters: portrait.png
---

## Look
Owner-neutral anti-tank weapons in a quiet period scene, matching the existing portrait
paintings' subdued ochre and sepia palette, cloudy sky and connected brushwork.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, 105:94 landscape aspect ideally 1680x1504 or larger. Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914–1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: Two khaki-clad infantrymen with an unusually long, heavy bolt-action anti-tank rifle on a bipod, resting at the edge of a muddy field. No artillery cannon.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_antiTank, EC_antiTank | Selected owner-neutral painting; shared by the listed gameplay classes. |

## Post
Selected source with a centred exact 105:94 crop, never stretched or enlarged.
`art-review -portrait art/pipeline/portrait-anti-tank-v1.json` exports 840×752 (4×),
420×376 (2×) and 210×188 (1×) into `assets/portraits/anti-tank/`. In-engine panel acceptance remains open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels, labels or UI
- photographic or 3D rendering
- action spectacle, muzzle flash or casualties
