# rail-ballast

Recipe: `art/recipes/terrain/distinct-v1/rail-ballast.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source unchanged; fixed Lanczos3 rectangle exports, separate controlled geometry creates runtime candidates. No original raster input.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: rail-ballast
- **review**: docs/art/distinct-terrain-batch-05.md
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/rail-ballast.png
- **source_sha256**: 9464088e4a0ec25942b87eda3d229a07f3538d5ad575ef663848ba99fcb0a385
- **status**: candidate-material
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## Prompt

```text
Use case: historical-scene. ONE original highresolution6:5 opaque material canvas for an overhead illustrated1914 European strategy game. FilltheENTIRErectangle edge-to-edge with compacted railway ballast: small angular weathered grey limestone and greige stone chips closely interlocked, tiny earth/dust pockets between stones, no large boulders. Fine continuous handpainted mineraltexture, muted neutral grey/brown with quiet contrast and diffuse overcast upperleftlight. Nearly even values atedges, no scene silhouette or vignette. Ballast only: NO rails,sleepers,grass,verges,tracklines,buildings,text or objects. Small dense chips with enough shape/detail toremain legible under controlled trackgeometry at192x160, crisp illustrateddetail atfullresolution. No pixelart, squarecells, outlines or nearestneighbor enlargement. Fresh invented detail, nooriginal-game art inputs or reconstruction. Fullbleed continuous surface, top-down orthographic.
```
