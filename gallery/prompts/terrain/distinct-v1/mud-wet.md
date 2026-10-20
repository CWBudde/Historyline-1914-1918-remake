# mud-wet

Recipe: `art/recipes/terrain/distinct-v1/mud-wet.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Preserve generated source; fixed centered exact6:5 crop then Lanczos3 384x320->192x160. No original mask/pixels. Rectangular overscan; geometry, phase and seams pending.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: mud-wet
- **review**: Independent source/reduction review: docs/art/distinct-terrain-batch-03.md. Candidate source only; runtime acceptance separate.
- **revision_reason**: Focused image-generator edit following independent full-source/reduction review.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/mud-wet-02.png
- **source_sha256**: 93c62ec9452975a7ec4a725cf6b50ed0186d524cbb8d70943f799e68c348df8f
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `build/gfx-research/distinct-v1/mud-wet-01.png` | edit-target-newly-generated-first-attempt | `4f13d3b8992ef1cc580320c7bb4aa6dfeeb483ef6bb587dc5903601216f0effc` |

## Prompt

```text
Edit this NEWLY AUTHORED dark damp mud terrain image only to improve WET MUD recognition in a tiny192x160 game export. Preserve landscape6:5 canvas, full opaque overscan, strict overhead camera, earth hue/texture and diffuse upperleft light, no text/border. Enlarge a FEW existing shallow puddles into irregular connected wet patches that totalabout12percent surface. Some patches20to30percent of canvaswidth, skinny irregular contours with muddy small land tongues. Puddles have subdued cool bluegrey overcast-sky reflections distinctly lighter/cooler than soil by a moderate value step, softly wet boundaries. No brilliant white highlights, black pools, large lake or river, no circular dot pattern, tracks, stones, craters, plants or new props. Include several low smooth moistsoil areas between veryfineclods, reducebusygrain aroundpuddles. Keepoverall dark wetground identity andcalmvalues but wetpatches must remain visible after reduction. Originalnewillustration notlegacy.
```
