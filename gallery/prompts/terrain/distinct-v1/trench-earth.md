# trench-earth

Recipe: `art/recipes/terrain/distinct-v1/trench-earth.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source unchanged; fixed Lanczos3 rectangle exports, separate controlled geometry creates runtime candidates. No original raster input.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: trench-earth
- **review**: docs/art/distinct-terrain-batch-05.md
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/trench-earth.png
- **source_sha256**: fbe3e9e8e72d8d441dae8903d130e1edd1aef1f0803d152d91f6cd49d77f11ef
- **status**: candidate-material
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## Prompt

```text
Use case: historical-scene. ONE original highresolution6:5 opaque material for newly illustrated1914 European overhead strategy terrain. ENTIREcanvas filled edge-to-edge with compacted damp dark brown trench-floor earth: fine silty/clayey granular soil, tiny pebbles, occasional faint compacted footprints-like unevenness but NOdistinct literal footprint pattern, sparse tiny straw/root fragments, matte slightlydamp surface without reflective puddles. Quiet lowcontrast fine continuous painted texture, dark muted umber/sepia earth distinguishable from blackshadow and fromgreyrailballast. Even diffuse overcast lighting, no vignette, stripes or directional tracks. This is a material canvas ONLY, not a trenchscene: NO trench outline, banks, timber, sandbags, grassy edges, boards, holes, weapons, people, text or UI. Fine handpainted continuous detail, no pixelart or squarecells. Opaque fullbleed top-down orthographic, read cleanly within narrow trench shapes at192x160 and remain finely detailed athighresolution. Fresh invented artwork, nooriginal gamepixels or reconstruction.
```
