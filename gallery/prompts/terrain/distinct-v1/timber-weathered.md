# timber-weathered

Recipe: `art/recipes/terrain/distinct-v1/timber-weathered.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source unchanged; fixed Lanczos3 rectangle exports, separate controlled geometry creates runtime candidates. No original raster input.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: timber-weathered
- **review**: docs/art/distinct-terrain-batch-05.md
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/timber-weathered.png
- **source_sha256**: c4d42a70eee2b05cdcb5948b15598c4373104893187ece5202907d8f21048faa
- **status**: candidate-material
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## Prompt

```text
Use case: historical-scene. ONE original highresolution6:5 opaque continuous illustrated MATERIAL canvas, evenlyfilled edge-to-edge with dark weathered brown railway-sleeper TIMBER surface, freshly invented for a1914 European overhead strategygame. Extremeclose planar material ratherthan anobject: dense aged oak/pine endgrain-like microfibres, subtle subdued charcoal-brown grain hints, very small cracks and worn dust, matte irregulartexture. Quiet even diffuse overcast light, lowcontrast macrovariation. Brown muted umber, visibly warmer/darker than greyballast, but not pitchblack. NO plank or board outlines, repeated boardseams, nails, bolts, ties arranged inrows, tracks, rails, grass, borders, vignette, text or objects. Tiny grain detail should workinside thin sleeper silhouettes at2x/4x, no dramatic directionalstriping. Fine painted continuous brushwork, no pixelart or squarecells. Entirecanvas fullbleed, opaque, top-downorthographic, nooriginal-gamepixels or reconstruction.
```
