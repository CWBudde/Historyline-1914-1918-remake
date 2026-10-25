---
id: portraits/armored-car
title: Armoured car
kind: portrait
status: approved
runtime_ids: CP_armCar, EC_armCar
masters: portrait.png
---

## Look
Owner-neutral picture of the unit shown in the unit panel at 210x188: a low riveted armoured car with a small round turret,
in a quiet period scene with subdued ochre and sepia values under a pale cloudy sky.

## Base prompt
Use case: historical-scene. Create ONE newly authored opaque full-bleed WWI strategy game UNIT PORTRAIT, landscape aspect105:94, ideally1680x1504 or larger. Same restrained ochre, sepia, khaki, olive and sage palette and continuous hand-painted digital brushwork as reference portrait. Pale cloudy sky, diffuse light from upper left, quiet period scene, crisp readable subject at210x188 with selective mechanical detail. Complete principal subject fits inside safe centred105:94 crop. Reference atlas is EQUIPMENT/IDENTITY reference only: replace its technical magenta with natural unmarked khaki/grey paint and use a normal low three-quarter scene camera, NOT the six-view layout. No text, numbers, letters, flags, crosses, insignia, emblems, national colours, faction tint, frames, UI, watermarks, photo, pixel art or3D render. Exactly one complete main vehicle. Subject: the WWI armoured car matching reference atlas: ONE low compact riveted rectangular armoured body with sloping long bonnet at FRONT, square radiator grille and TWO round brass headlights, FOUR narrow exposed spoked wheels TOTAL (two per side), ONE small LOW circular cylindrical turret near rear half with ONE short dark machine gun, rectangular rear compartment and riveted access door. No tall tower, stacked turret or high armoured lorry cab. Natural subdued khaki/olive armour, dark tyres, brass hardware. All panels entirely plain and unmarked; ABSOLUTELY NO crosses, flags, symbols, decals, badges, logos, letters or numbers anywhere. Parked on quiet rural dirt lane, low front-left three-quarter view, front pointing LEFT, complete vehicle visible including all visible wheels and gun, distant trees and pale cloudy sky.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | CP_armCar, EC_armCar | The scene above; one image serves both sides. |

## Post
`art-review -portrait art/pipeline/portrait-armored-car-v1.json` takes the largest exact centred 105:94
crop and reduces it with Lanczos3 to 4x (840x752), 2x (420x376) and 1x (210x188) as
`portrait/armored-car`. Selected exports are installed under `assets/portraits/armored-car/`.
Gameplay classes: `CP_armCar`, `EC_armCar`; in-engine panel acceptance is open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels or UI
- photographic skin and fabric texture
- action spectacle, muzzle flash or casualties
