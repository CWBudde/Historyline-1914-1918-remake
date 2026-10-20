# earth-bare

Recipe: `art/recipes/terrain/distinct-v1/earth-bare.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact centered 6:5 crop then Lanczos3 exports; no original reference or mask. Rectangular overscan; hex clipping, texture phase and connections remain unaccepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: earth-bare
- **review**: Batch02 review in docs/art/distinct-terrain-batch-02.md; source candidacy does not establish runtime replacement or seamlessness.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/earth-bare.png
- **source_sha256**: e7d3188bc9c734a1f45ad80ccce934c46d93cb4cf9843c079bbc6260ec848404
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/meadow.png` | newly-authored-material-style-only | `8c8935efbbd463e39d49208e42ca0e87b140268a1b11941764dfa7c29c4fff85` |

## Prompt

```text
Create ONE newly authored high-resolution terrain illustration for an early twentieth-century European strategy game. Landscape canvas aspect6:5. Strictly vertically overhead orthographic view, no horizon or perspective. Continuous finely painted natural surfaces with tiny crisp organic detail, calm lowcontrast large forms, diffuse overcast light with gentle upperleft illumination. No pixel art, blocky cells, traced grid, UI, labels, borders, hex outline, vignette, objects or original-game-art imitation. Entire full opaque rectangular canvas with overscan; fill all edges. Must read clearly reduced192x160 while retaining continuous detail at full source size. Image1 is MATERIAL/STYLE reference only, the new muted meadow. Paint bare compacted earth terrain: warm muted taupe/umber soil, very fine irregular grains, tiny occasional grey grit, shallow broken soil fissures and very sparse faded short grass tufts under5percent coverage. Calm broad small mottling, not a road, no tracks, craters or central bare island. Whole canvas soil material with compatible quiet values and same aerial material scale as reference meadow. No pebbles larger than tiny specks, no large rocks, no trees, no flowers or strong cast shadows. Invent a new material, no copy of meadow's surface patch pattern.
```
