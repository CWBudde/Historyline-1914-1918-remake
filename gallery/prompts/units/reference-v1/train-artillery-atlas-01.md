# unit-train-artillery-01

Recipe: `art/recipes/units/reference-v1/train-artillery-atlas-01.json`

- **acceptance**: Approximately 90% resemblance is a visual review target, not a measured score. Direction, physical pivot, native correspondence and runtime acceptance remain review gates.
- **art_direction**: reference-faithful-continuous-painted-v1
- **cleanup**: Unchanged generated PNG. Existing Go Lanczos3 exporter only; no pixel repair, alpha removal, rotations or original-pixel compositing.
- **date**: 2026-10-15
- **id**: unit-train-artillery-01
- **schema_version**: 1
- **source**: art/source/units/reference-v1/train-artillery.png
- **source_sha256**: e0dead7f6a77250c9306eec500536b029a602cc89dd8cb23c8d9e115e1c3d1ba
- **status**: candidate-source
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/units/distinct-v1/infantry.png` | <nil> | `a151ab9f110782ead6eba6bc4af3b4a17eb789204c5f19e4ca2e786614912290` |
| `legacy/python/gfx/units/icons/trainArtillery.png` | <nil> | `3143bf28e263a0d46412b7ea88656c7113089176213c4b21c480c2d145901650` |
| `legacy/python/gfx/units/pictures/trainArtillery.png` | <nil> | `7fb9f08b339e4b9ab827845d031a2104c60f372d39ad6955ed9e5f19eae8f3cc` |

## Prompt

```text
Use case: historical-scene. Asset: Artillery train six-direction tactical sprite atlas. Image1 supplies exact established painted atlas camera, layout, light and technical magenta recolour convention ONLY. Image2 is the tiny trainArtillery reference icon and Image3 its portrait: preserve recognizable subject identity and proportions. Subject in EACH cell: a single WWI railway artillery wagon with a large long cannon elevated slightly, platform and rail wheels; no tracks. EXACT1024x1536 transparent RGBA, 2 columns x3 rows of512square cells, no gutters. Row-major headings EXACT N(up),NW(up-left),SW(down-left),S(down),SE(down-right),NE(up-right). Front of subject points toward named screen heading. Six newly painted front/back/side views, not rotated or mirrored finished pixels. Fixed elevated three-quarter overhead strategy camera and screen upper-left lighting. Same object/person count, equipment, scale across views. Continuous detailed historical painting, not pixel art, cartoon or photo. Recolourable uniform cloth or vehicle armour paint ONLY saturated shaded MAGENTA; natural leather/skin/wood/horses/concrete/metal neutral. For bunker use subtle magenta painted identification strip only; keep concrete natural. Keep whole subject centred(256,261) within CENTRAL65% of EVERY cell; at least17% fully transparent margin on ALL FOUR cell edges including barrel/muzzle/feet/bow. Long guns/trains must fit compactly with all extremities. Real alpha0 outside physical objects and between limbs. No ground, track bed, water, shadow, haze, glow, backdrop, checkerboard, grid, labels, arrows, insignia or text. About90% likeness of recognizable subject after Lanczos reduction to tiny reference size is visual target.
CRITICAL camera-facing map: top-left N shows REAR of subject, its front/muzzle points UP/AWAY from viewer. Top-right NW shows REAR and RIGHT side, front/muzzle points UPPER LEFT. Middle-left SW shows FRONT and RIGHT side, front/muzzle points LOWER LEFT. Middle-right S shows FRONT, front/muzzle points DOWN/TOWARD viewer. Bottom-left SE shows FRONT and LEFT side, front/muzzle points LOWER RIGHT. Bottom-right NE shows REAR and LEFT side, front/muzzle points UPPER RIGHT. Do not confuse heading with camera location. No front view in N cell. All six occupy identical central65% cell rectangle; shrink to fit.
```
