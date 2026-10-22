---
id: portraits/static-aa
title: Static anti-aircraft emplacement
kind: portrait
status: revised
runtime_ids: CP_staticAA, EC_staticAA
masters: portrait.png
---

## Look
Owner-neutral static anti-aircraft emplacement in a quiet period scene, matching the existing portrait
paintings' subdued ochre and sepia palette, cloudy sky and connected brushwork.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, 105:94 landscape aspect ideally 1680x1504 or larger. Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914–1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: A water-cooled heavy machine gun on a tall improvised wooden anti-aircraft tripod, barrel elevated, with two khaki-clad crewmen in a sandbag emplacement.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_staticAA, EC_staticAA | Selected owner-neutral painting; shared by the listed gameplay classes. |

## Post
Selected source with a centred exact 105:94 crop, never stretched or enlarged.
`art-review -portrait art/pipeline/portrait-static-aa-v1.json` exports 840×752 (4×),
420×376 (2×) and 210×188 (1×) into `assets/portraits/static-aa/`. In-engine panel acceptance remains open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels, labels or UI
- photographic or 3D rendering
- action spectacle, muzzle flash or casualties
