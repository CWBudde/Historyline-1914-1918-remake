# water-open

Recipe: `art/recipes/terrain/distinct-v1/water-open.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact centered 6:5 crop then Lanczos3 exports; no original reference or mask. Rectangular overscan; hex clipping, texture phase and connections remain unaccepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: water-open
- **review**: Batch02 review in docs/art/distinct-terrain-batch-02.md; source candidacy does not establish runtime replacement or seamlessness.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/water-open.png
- **source_sha256**: 0b080163804881ac49bc4b3162da547abf443c9082d6e67065c119174d79b51d
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/river-diagonal.png` | newly-authored-material-style-only | `de6237db22ba129e0825d233dd89964cba2e211f4db9f0e6f1d9b46bd5a2c6b4` |

## Prompt

```text
Create ONE newly authored high-resolution terrain illustration for an early twentieth-century European strategy game. Landscape canvas aspect6:5. Strictly vertically overhead orthographic view, no horizon or perspective. Continuous finely painted natural surfaces with tiny crisp organic detail, calm lowcontrast large forms, diffuse overcast light with gentle upperleft illumination. No pixel art, blocky cells, traced grid, UI, labels, borders, hex outline, vignette, objects or original-game-art imitation. Entire full opaque rectangular canvas with overscan; fill all edges. Must read clearly reduced192x160 while retaining continuous detail at full source size. Image1 is only a material/style reference: the NEWLY AUTHORED muted dusty blue-teal river. Paint ONLY open water filling the entire image from edge to edge. Use similar quiet slate blue-teal hue, fine short gently irregular connected currents and very sparse small glints, evenly distributed low amplitude surface relief. No shore, grass, bank, land, wave breakers, foam bands, large white streaks, rocks, boats, islands, recognizable central focal feature or depth vignette. Water represents calm European lake/sea terrain at map scale rather than stormy closeupwaves. Never use the river's grass, diagonal bank or channel composition; recreate its water material on a fresh independent full-water canvas.
```
