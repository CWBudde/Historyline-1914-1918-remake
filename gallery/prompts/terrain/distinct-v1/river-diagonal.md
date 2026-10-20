# river-diagonal

Recipe: `art/recipes/terrain/distinct-v1/river-diagonal.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact center crop then Lanczos3 exports, no original mask or pixels. Rectangular overscan; runtime hex clipping and connections not yet accepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: river-diagonal
- **review**: Independent batch review recorded in docs/art/distinct-terrain-batch-01.md. No original-fidelity acceptance sought under updated brief.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/river-diagonal.png
- **source_sha256**: de6237db22ba129e0825d233dd89964cba2e211f4db9f0e6f1d9b46bd5a2c6b4
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/meadow.png` | newly-authored-material-style-only | `8c8935efbbd463e39d49208e42ca0e87b140268a1b11941764dfa7c29c4fff85` |

## Prompt

```text
Create a NEW terrain tile illustration, image 1 only as newly authored meadow MATERIAL/STYLE reference. Match muted sage/moss greens, tiny calm continuous natural detail, diffuse overcast light and strict directly overhead orthographic viewpoint. Landscape6:5 full opaque canvas. Subject: one quiet flowing river crossing diagonally from LEFT canvas edge at25% height to RIGHT canvas edge at75% height. Smooth gently curving banks separated by about25% canvas height. Water dusty blue/teal, subtle currents, fine connected ripples with very few tiny reflected highlights, no white foam, no waves, no dark dramatic depths. Thin warm damp brown bank edges and restrained short grass. CRITICAL: the continuous channel remains full width at BOTH LEFT AND RIGHT image boundaries, both banks cross each canvas edge separately, water and banks extend beyond the frame, no grass dam, no river termination, no pond shape, no narrowing into a point. No islands, stones, bridges, buildings, boats, trees, additional channels. Grass fills upper and lower regions and corners, even quiet values under game units. No border, hex outline, grid, text, pixels or square blocks. Invent a fresh natural illustration with original composition; no reference to any legacy game's pixels or palette. Actual continuous high resolution detail, not enlarged pixel art.
```
