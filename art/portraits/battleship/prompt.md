---
id: portraits/battleship
title: Battleship
kind: portrait
status: revised
runtime_ids: portrait/battleship
masters: portrait.png
---

## Look
Owner-neutral battleship with two twin-gun turrets, two funnels and one mast, matching the
selected tactical sprite in a quiet coastal scene with subdued ochre and sepia brushwork.

## Base prompt
Use case: historical-scene. Create ONE newly authored opaque full-bleed WWI strategy game UNIT PORTRAIT, landscape aspect105:94, ideally1680x1504 or larger. Same restrained ochre, sepia, khaki, olive and sage palette and continuous hand-painted digital brushwork as reference portrait. Pale cloudy sky, diffuse light from upper left, quiet period scene, crisp readable subject at210x188 with selective mechanical detail. Complete principal subject fits inside safe centred105:94 crop. Reference atlas is EQUIPMENT/IDENTITY reference only: replace its technical magenta with natural unmarked khaki/grey paint and use a normal low three-quarter scene camera, NOT the six-view layout. No text, numbers, letters, flags, crosses, insignia, emblems, national colours, faction tint, frames, UI, watermarks, photo, pixel art or3D render. Exactly one complete main vehicle. Subject: early WWI battleship matching the reference sprite: broad armoured hull with pointed bow to LEFT, rounded stern to RIGHT, warm teak deck, exactly TWO main twin-gun turrets TOTAL: ONE low twin turret forward of bridge, ONE low twin turret aft of both funnels. Exactly FOUR large barrels TOTAL. No superimposed turrets, no extra turret behind the forward turret, no wing turrets. One compact forward bridge, exactly TWO thick black funnels in centre line behind bridge, and ONE simple short mast. Low three-quarter broadside view from off the port bow, entire ship including bow/stern and mast inside canvas on quiet coastal water. Quiet muted neutral grey/khaki hull, small restrained ripple, distant low rural coast. Preserve two-turret arrangement so both turrets are clearly visible.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | portrait/battleship | Selected owner-neutral painting; gameplay class binding pending. |

## Post
Selected source with a centred exact 105:94 crop, never stretched or enlarged.
`art-review -portrait art/pipeline/portrait-battleship-v1.json` exports 840×752 (4×),
420×376 (2×) and 210×188 (1×) into `assets/portraits/battleship/`. Gameplay class binding
and in-engine panel acceptance remain open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels, labels or UI
- photographic or 3D rendering
- action spectacle, muzzle flash or casualties
