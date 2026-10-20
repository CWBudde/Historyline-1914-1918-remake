# snow-ground

Recipe: `art/recipes/terrain/distinct-v1/snow-ground.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact centered 6:5 crop then Lanczos3 exports; no original reference or mask. Rectangular overscan; hex clipping, texture phase and connections remain unaccepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: snow-ground
- **review**: Batch02 review in docs/art/distinct-terrain-batch-02.md; source candidacy does not establish runtime replacement or seamlessness.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/snow-ground.png
- **source_sha256**: 127fd16b8ae4951273885c0863c14f1db355be774830ed557be99d4349a55ee3
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/meadow.png` | newly-authored-material-style-only | `8c8935efbbd463e39d49208e42ca0e87b140268a1b11941764dfa7c29c4fff85` |

## Prompt

```text
Create ONE original high-resolution terrain illustration for a European1914 strategy game. Image1 is solely the NEWLY AUTHORED meadow's fine continuous painted material/style reference, not an edit target. Strict vertically overhead orthographic terrain view, landscape canvas6:5, no horizon or perspective, consistent diffuse overcast light with subtle upperleft highlights. Entire opaque filled rectangle with overscan, quiet at192x160 but finely detailed at full resolution. No game UI, border, vignette, hex silhouette, labels, pixel art, blocky shapes or legacy art reconstruction. Subject WINTER SNOW GROUND filling all canvas edges: pale cool ivory fresh settled snow with very subtle pale blue-grey surface depressions and short fine wind-scoured soft texture. Sparse tiny faded straw grass tips under3percent cover, no footprints, no wheel tracks, no rocks, trees, objects, big drifts or compositional islands. Maintain almost even lowcontrast values for legibility, soft granular continuous detail rather than a flat white rectangle or repeated airbrush dots. Clean daylight snow, no yellow mud contamination or icy lake appearance, no sparkling glitter or isolated bright specks. Entire broad terrain snow surface at aerial map scale.
```
