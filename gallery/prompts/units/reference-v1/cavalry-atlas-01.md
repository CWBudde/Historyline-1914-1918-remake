# unit-cavalry-01

Recipe: `art/recipes/units/reference-v1/cavalry-atlas-01.json`

- **acceptance**: Approximately 90% resemblance is a visual review target, not a measured score. Direction, physical pivot, native correspondence and runtime acceptance remain review gates.
- **art_direction**: reference-faithful-continuous-painted-v1
- **cleanup**: Unchanged generated PNG. Existing Go Lanczos3 exporter only; no pixel repair, alpha removal, rotations or original-pixel compositing.
- **date**: 2026-10-15
- **id**: unit-cavalry-01
- **schema_version**: 1
- **source**: art/source/units/reference-v1/cavalry.png
- **source_sha256**: b6132f86f8c8182f0710747fd2a6592a03b6351f923c72f930915206d60caba0
- **status**: candidate-source
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/units/distinct-v1/infantry.png` | <nil> | `a151ab9f110782ead6eba6bc4af3b4a17eb789204c5f19e4ca2e786614912290` |
| `legacy/python/gfx/units/icons/cavalry.png` | <nil> | `2eb5a1e776a75a335483d889b15c7d6c26bad3165a65814aedc5252c0d4dd4de` |
| `legacy/python/gfx/units/pictures/cavalry.png` | <nil> | `5b3241b5a7b83ffcb8139ee9cdd9da598544d0ef1187e9ab424ebfddc1f64358` |

## Prompt

```text
Use case: historical-scene. Asset: Cavalry six-direction tactical sprite atlas. Image1 supplies exact established painted atlas camera, layout, light and technical magenta recolour convention ONLY. Image2 is the tiny cavalry reference icon and Image3 its portrait: preserve recognizable subject identity and proportions. Subject in EACH cell: two mounted WWI cavalrymen on brown horses, riders with sabres and leather saddles, compact pair. EXACT1024x1536 transparent RGBA, 2 columns x3 rows of512square cells, no gutters. Row-major headings EXACT N(up),NW(up-left),SW(down-left),S(down),SE(down-right),NE(up-right). Front of subject points toward named screen heading. Six newly painted front/back/side views, not rotated or mirrored finished pixels. Fixed elevated three-quarter overhead strategy camera and screen upper-left lighting. Same object/person count, equipment, scale across views. Continuous detailed historical painting, not pixel art, cartoon or photo. Recolourable uniform cloth or vehicle armour paint ONLY saturated shaded MAGENTA; natural leather/skin/wood/horses/concrete/metal neutral. For bunker use subtle magenta painted identification strip only; keep concrete natural. Keep whole subject centred(256,261) within CENTRAL65% of EVERY cell; at least17% fully transparent margin on ALL FOUR cell edges including barrel/muzzle/feet/bow. Long guns/trains must fit compactly with all extremities. Real alpha0 outside physical objects and between limbs. No ground, track bed, water, shadow, haze, glow, backdrop, checkerboard, grid, labels, arrows, insignia or text. About90% likeness of recognizable subject after Lanczos reduction to tiny reference size is visual target.
```
