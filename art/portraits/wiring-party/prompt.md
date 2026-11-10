---
id: portraits/wiring-party
title: Wiring party
kind: portrait
status: approved
runtime_ids: CP_wiring, EC_wiring
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: two soldiers carrying a barbed-wire coil on a
pole and a bundle of screw pickets across a cratered field, in subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Newly authored opaque full-bleed WWI strategy-game unit portrait, LANDSCAPE 105:94 aspect, ideally 1680x1504 (at least 1260x1128). Match the reference portraits: subdued ochre, sepia, khaki, olive and sage; broad connected hand-painted digital brushwork with selective fine mechanical detail; pale cloudy sky and diffuse light from upper left. Rural Europe 1914-1918, calm period scene, owner-neutral. Central subject fully visible and readable at 210x188, safe centered crop. No text, letters, numbers, insignia, flags, faction tint, magenta, borders, frame, UI, watermark, pixel art, photograph or 3D render. Subject: A two-man WWI wiring party in pale daylight: two soldiers in drab greatcoats and soft caps, thick leather gauntlets, carrying a big round coil of barbed wire hung from a wooden pole on their shoulders between them, the rear man with a bundle of corkscrew iron screw pickets under his arm; no rifles. Three-quarter view, the men walking left across a cratered field towards a line of half-finished wire entanglements on screw pickets.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_wiring, EC_wiring | The scene above; one image serves both sides. |

## Post
Selected source: pwire-01 of gen-2026-11-10, exec-de765e9a, 1326x1187. The prompt followed the `art/style.md`
Preamble and Portraits text, with `art/portraits/sappers/portrait.png` attached for palette, brushwork, sky and
light only. `art-review -portrait art/pipeline/portrait-wiring-party-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/wiring-party`. Promote into `assets/portraits/wiring-party/` and re-index.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
- rifles or other weapons
