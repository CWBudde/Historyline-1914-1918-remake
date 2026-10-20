# steel-weathered

Recipe: `art/recipes/terrain/distinct-v1/steel-weathered.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source unchanged; fixed Lanczos3 rectangle exports, separate controlled geometry creates runtime candidates. No original raster input.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: steel-weathered
- **review**: docs/art/distinct-terrain-batch-05.md
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/steel-weathered.png
- **source_sha256**: a8cea787aeb3a319241245e2d439af4b536733b20634f713044ec36897e00551
- **status**: candidate-material
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## Prompt

```text
Use case: historical-scene. ONE original highresolution6:5 opaque fullbleed continuous illustrated material: the matte worn dark gunmetal STEEL surface used inside very thin railroad rail silhouettes for a1914 European overhead strategygame. Uniform mid-dark neutral grey with restrained tiny warm brown oxidation specks, cool muted worn grey flecks and fine mineral/metalgrain. No large patches or shiny hotspots. Diffuse overcast light, lowcontrast, nearly even values atall edges. This is a close flat surface MATERIAL, not anobject: NO railshapes, parallelstripes, planks, bolts, tracks, grass, objects, horizon, text, frame or vignette. No white/chrome highlights, mirrorreflection or glitter. Fine continuous painted illustration with coherent tiny detail, nopixelart or squarecells. Detail quiet enough for thin paired rails at192x160 yet highresolution atfullsize. Fresh invented texture, nooriginal gameinputs or reconstruction. Entirecanvasopaque.
```
