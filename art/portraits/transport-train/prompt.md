---
id: portraits/transport-train
title: Transport train
kind: portrait
status: revised
runtime_ids: CP_transTrain
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a single compact steam locomotive,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Use case: historical-scene. Create ONE newly authored opaque full-bleed WWI strategy game UNIT PORTRAIT, landscape aspect105:94, ideally1680x1504 or larger. Same restrained ochre, sepia, khaki, olive and sage palette and continuous hand-painted digital brushwork as reference portrait. Pale cloudy sky, diffuse light from upper left, quiet period scene, crisp readable subject at210x188 with selective mechanical detail. Complete principal subject fits inside safe centred105:94 crop. Reference atlas is EQUIPMENT/IDENTITY reference only: replace its technical magenta with natural unmarked khaki/grey paint and use a normal low three-quarter scene camera, NOT the six-view layout. No text, numbers, letters, flags, crosses, insignia, emblems, national colours, faction tint, frames, UI, watermarks, photo, pixel art or3D render. Exactly one complete main vehicle. Subject: the compact WWI transport STEAM LOCOMOTIVE from reference atlas. A single short black riveted steam tank locomotive, round horizontal boiler at FRONT, ONE tall narrow cylindrical chimney, ONE steam dome, a compact boxy enclosed rear cab with curved roof, coal bunker immediately behind cab, SIX large driving wheels TOTAL (three per side), connecting rods, buffer beams with two round buffers and hook coupling. Natural blackened steel boiler, muted olive/brown cab, brass valves; sacks and wooden crates beside track. Complete locomotive on rural European standard-gauge track, its front to LEFT, low three-quarter front-left view, grassy railway verge, distant small station under cloudy sky. Engine is entire main subject centred and fills lower two-thirds. No passenger coaches or people dominating foreground; not a workmen portrait. No locomotive cropped or hidden.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_transTrain | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-transport-train-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/transport-train`. Selected exports are installed under `assets/portraits/transport-train/`.
Gameplay binding remains `CP_transTrain`; in-engine panel acceptance is open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
