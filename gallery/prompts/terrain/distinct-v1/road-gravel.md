# road-gravel

Recipe: `art/recipes/terrain/distinct-v1/road-gravel.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Source unchanged; fixed Lanczos3 pack export and separately declared connected compositor material sampling. No original raster input.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: road-gravel
- **review**: Material and derived road/stream family review: docs/art/connected-family-batch-01.md. Candidate, not release approval.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/road-gravel.png
- **source_sha256**: 93991a9f33f3d125f68f0ed2d14b329f1092428a0df59ba90f92c5b2bc85dd26
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/road-bend.png` | newly-authored-gravel-material-style-only | `4c347558277bfb36113dde0d9a904f8ee744ca8b06287a227a35f000a338ec6e` |

## Prompt

```text
Generate ONE NEW highresolution material illustration for a European1914 strategy-game terrain set. Image1 is ONLY the newly authored country-road's stone/earth MATERIAL and quiet continuouspaint STYLE reference, not a layout tocopy. Create a rectangular6:5 full opaque canvas ENTIRELY filled by fine compacted macadam/gravel surface. Even muted grey-brown/greige aggregate, tiny sharp irregular stone grains embedded in compacted earth, subtle broad lowcontrast wear variation but NO tire ruts, tracks or centralstreak. No grass, verge, bank, side boundary, lane paint, white lines, potholes, large loose stones, road silhouette or visible edge. This is the unmarked road material alone tobe used inside controlled newroad shapes. STRICT vertically overhead orthographic, diffuse overcast upperleftlight, quiet values underuniticons at192x160 but crisp organic grain atfullresolution. Approximately even edgecolors, no vignette, frame, pixelart, squarecells, outlines, text or UI. Entireoverscan filled, no transparency. Invent fresh detail, never tracelegacygamepixels. Fine illustrated continuous brushwork rather than enlargedpixelart or oversharpenedphoto.
```
