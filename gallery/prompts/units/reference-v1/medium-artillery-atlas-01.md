# unit-medium-artillery-01

Recipe: `art/recipes/units/reference-v1/medium-artillery-atlas-01.json`

- **acceptance**: Approximately 90% resemblance is a visual review target, not a measured score. Direction, physical pivot, native correspondence and runtime acceptance remain review gates.
- **art_direction**: reference-faithful-continuous-painted-v1
- **cleanup**: Unchanged generated PNG. Existing Go Lanczos3 exporter only; no pixel repair, alpha removal, rotations or original-pixel compositing.
- **date**: 2026-10-15
- **id**: unit-medium-artillery-01
- **schema_version**: 1
- **source**: art/source/units/reference-v1/medium-artillery.png
- **source_sha256**: 16c5ed54199545002fffa7b533a9347ef87f708c026cb281b0dd1e1b799de251
- **status**: candidate-source
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/units/distinct-v1/infantry.png` | <nil> | `a151ab9f110782ead6eba6bc4af3b4a17eb789204c5f19e4ca2e786614912290` |
| `legacy/python/gfx/units/icons/mediumArtillery.png` | <nil> | `17147fe4cabee6dca25711357fef7a52d58ee7f736a3624daef5e300a5cf66df` |
| `legacy/python/gfx/units/pictures/mediumArtillery.png` | <nil> | `15291c1834bb86f23e87201fd0a3b691110375cbbd682071d94485c38b7e4ac0` |

## Prompt

```text
Use case: historical-scene. Asset: Medium artillery six-direction tactical sprite atlas. Image1 supplies exact established painted atlas camera, layout, light and technical magenta recolour convention ONLY. Image2 is the tiny mediumArtillery reference icon and Image3 its portrait: preserve recognizable subject identity and proportions. Subject in EACH cell: a medium field artillery gun with two large spoked wheels, long slim barrel and low carriage, no crew. EXACT1024x1536 transparent RGBA, 2 columns x3 rows of512square cells, no gutters. Row-major headings EXACT N(up),NW(up-left),SW(down-left),S(down),SE(down-right),NE(up-right). Front of subject points toward named screen heading. Six newly painted front/back/side views, not rotated or mirrored finished pixels. Fixed elevated three-quarter overhead strategy camera and screen upper-left lighting. Same object/person count, equipment, scale across views. Continuous detailed historical painting, not pixel art, cartoon or photo. Recolourable uniform cloth or vehicle armour paint ONLY saturated shaded MAGENTA; natural leather/skin/wood/horses/concrete/metal neutral. For bunker use subtle magenta painted identification strip only; keep concrete natural. Keep whole subject centred(256,261) within CENTRAL65% of EVERY cell; at least17% fully transparent margin on ALL FOUR cell edges including barrel/muzzle/feet/bow. Long guns/trains must fit compactly with all extremities. Real alpha0 outside physical objects and between limbs. No ground, track bed, water, shadow, haze, glow, backdrop, checkerboard, grid, labels, arrows, insignia or text. About90% likeness of recognizable subject after Lanczos reduction to tiny reference size is visual target. CRITICAL: standalone unlimbered field gun ONLY in all six panels. NO HORSES, NO RIDERS, NO PEOPLE. Muzzle points N,NW,SW,S,SE,NE respectively, carriage trails opposite muzzle.
```
