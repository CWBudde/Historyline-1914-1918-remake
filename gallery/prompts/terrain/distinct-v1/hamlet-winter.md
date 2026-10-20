# hamlet-winter

Recipe: `art/recipes/terrain/distinct-v1/hamlet-winter.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated master unchanged. Fixed Lanczos3 exports. No original-image input.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: hamlet-winter
- **review**: docs/art/distinct-terrain-batch-04.md
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/hamlet-winter.png
- **source_sha256**: ad974e3d68a990cf8505278d1aa067db8b85eaf7a5f1a711c2adf501d9181896
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/hamlet-02.png` | newly-authored-matching-season-edit-target | `5163ee929139be5bfd75c43d23ec5b709f11e58fcac1aea3f2818672bfc9cda7` |

## Prompt

```text
Use case: lighting-weather edit. Image1 is the selected newly authored overhead hamlet, the edit target. Make its matching WINTER terrain master. Preserve exactly the three cottage roof silhouettes/positions/orientations, courtyard footprint, well and garden plot boundaries, scale and near-vertical overhead orthographic map camera. Same6:5 opaque canvas and continuous fine illustrated brushwork. Change season only: subdued offwhite snow covering grass and gardenbeds, soft uneven snow on roof planes with enough reddish/slate edge exposed to keep the roofs recognizable, cleared/trampled grey-brown courtyard earth, dormant bare small tree crowns and shrubs, gardenplants dormant not lush green. Natural cold diffuse upperleft light, restrained blue-grey snow shadows, no dramatic cast shadows or vignette. Retain quiet snowy outerbuffer for joining future winter ground; all buildings centrally contained. No additional structures, roads, track exits, people, vehicles, text, hexoutline, grid or UI; no pixelart. Detailed fullresolution, clear at192x160. Do not alter the architecture or move any buildings.
```
