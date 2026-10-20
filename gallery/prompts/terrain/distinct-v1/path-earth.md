# path-earth

Recipe: `art/recipes/terrain/distinct-v1/path-earth.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated master unchanged. Fixed Lanczos3 exports. No original-image input.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: path-earth
- **review**: docs/art/distinct-terrain-batch-04.md
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/path-earth.png
- **source_sha256**: eaae81c8dfea2ad506a02b80005d4ba94784b3827fa0ac6ee50d99e48519f04b
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## Prompt

```text
Use case: historical-scene. Generate ONE original continuous illustrated material for a 1914 European overhead strategy-game terrain set. Canvas6:5 opaque rectangle, filled edge to edge with warm muted tan compacted fine earth, tiny sandy grains and very occasional minute rounded pebbles. This is the dirt MATERIAL of a narrow rural footpath, not a road scene: NO path silhouette, verges, grass, tire tracks, ruts, lane paint, stone paving, large rocks or isolated objects. Even diffuse overcast light, no shadow bands or vignette. Quiet values and lowcontrast macro variation so it stays readable under map units at192x160, but finely painted crisp continuous organic texture atfullresolution. Natural brushwork, no pixelart, squarecells, outlines, text or UI. Fresh invented detail; no original game reference or reconstruction. Palette dusty ochre/greige earth lighter and warmer than dark wet mud, distinct from grey macadam.
```
