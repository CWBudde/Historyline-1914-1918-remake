---
id: portraits/spad-xiii
title: SPAD XIII
kind: portrait
status: revised
runtime_ids: portrait/spad-xiii
masters: portrait.png
---

## Look
Owner-neutral spad xiii in a quiet period scene, matching the existing portrait
paintings' subdued ochre and sepia palette, cloudy sky and connected brushwork.

## Base prompt
Use case: precise-object-edit. Correct only the engine-front detail and safe framing in this SPAD XIII WWI unit portrait. Keep SAME airplane, exactlyTWO wing levels, ONE open cockpit, wood propeller, TWO parallel fuselage machine guns, struts, tail, wheels, pale cloudy sky and grassy airfield, natural unmarked tan/khaki fabric and sepia/ochre hand-painted brushwork. The FRONT engine opening directly BEHIND THE PROPELLER must be a dark ROUND HONEYCOMB WATER RADIATOR grille around a simple central propeller hub and small spinner, clearly fine dense mesh texture. Remove exposed radial/rotary cylinder shapes from that front opening; the water-cooled V8 is enclosed inside natural metal cowling, with short side exhausts. Do not change open cockpit into a radiator. Pull camera back about15% and extend the SAME sky and grass naturally so ALL wingtips, propeller and tail fit within central85% width, safe for centred exact105:94 crop. Full-bleed OPAQUE landscape105:94, ideally1680x1504. No markings, letters, numbers, flags, crosses, labels, frame, UI, watermark, photograph or3Drender.
Safe framing: aircraft wingspan stays less than image height and within centred exact105:94 crop.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| portrait | portrait/spad-xiii | Selected owner-neutral painting; gameplay class binding pending. |

## Post
Selected source with a centred exact 105:94 crop, never stretched or enlarged.
`art-review -portrait art/pipeline/portrait-spad-xiii-v1.json` exports 840×752 (4×),
420×376 (2×) and 210×188 (1×) into `assets/portraits/spad-xiii/`. Gameplay class binding
and in-engine panel acceptance remain open.

## Avoid
- national colours, flags, insignia or faction tint
- frames, bevels, labels or UI
- photographic or 3D rendering
- action spectacle, muzzle flash or casualties
