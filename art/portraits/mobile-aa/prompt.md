---
id: portraits/mobile-aa
title: Mobile anti-aircraft emplacement
kind: portrait
status: approved
runtime_ids: CP_mobileAA, EC_mobileAA
masters: portrait.png
---

## Look
Owner-neutral mobile anti-aircraft emplacement in a quiet period scene, matching the existing portrait
paintings' subdued ochre and sepia palette, cloudy sky and connected brushwork.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, 105:94 landscape aspect ideally 1680x1504 or larger. Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914–1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: An early open-cab military lorry carrying a high-angle anti-aircraft cannon on its rear flatbed, two crewmen beside it, parked on a rural track.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_mobileAA, EC_mobileAA | Selected owner-neutral painting; shared by the listed gameplay classes. |

## Post
Selected source with a centred exact 105:94 crop, never stretched or enlarged.
`art-review -portrait art/pipeline/portrait-mobile-aa-v1.json` exports 840×752 (4×),
420×376 (2×) and 210×188 (1×) into `assets/portraits/mobile-aa/`. In-engine panel acceptance remains open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels, labels or UI
- photographic or 3D rendering
- action spectacle, muzzle flash or casualties
