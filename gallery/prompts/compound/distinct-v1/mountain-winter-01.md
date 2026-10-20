# mountain-winter-01

Recipe: `art/recipes/compound/distinct-v1/mountain-winter-01.json`

- **date**: 2026-10-15
- **geometry_reference**: docs/art/mountain-tile-geometry.md
- **id**: mountain-winter-01
- **method**: Seasonal edit of selected new summer03; same overhead ridge network.
- **output_manifest**: art/compound-mountain-v1.json
- **schema_version**: 1
- **source**: art/source/compound/distinct-v1/mountain-winter.png
- **source_sha256**: 2e90ccac63b02a4103ad48f015c3a9665e91ff1b19848500eb2da5b542a76225
- **status**: reviewed-source-for-provisional-compound-exports
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/compound/distinct-v1/mountain-summer.png` | new-summer-master-edit-target-geometry-camera | `620316294fc71cc75f37145f0516669ee140d73ff1d537cc01175c74d7a5ae66` |

## Prompt

```text
Use case: lighting-weather
Asset type: winter variant of our NEW overhead mountain terrain master.
Input image1 is the edit target and exact geometry/camera authority. Make ONLY a winter weather/material change. Keep the full 1536x1024 canvas, exact footprint, transparency, directly overhead nadir camera, every branching ridge crest/junction, valley and rock outcrop at exactly its current location and scale. No reframing, zooming, adding peaks or redesigning geometry.
Change olive grass to desaturated pale winter grass/earth with thin patchy snow across low slopes and valley ground; cool grey limestone remains readable. Snow settles lightly between branching rock ribs and on quieter terrain. Maintain a clear mountain silhouette at small tile scale, avoiding solid white fill or dense speckles. Cold soft diffuse daylight, restrained shadows. Preserve genuine transparent outer margins and softly irregular apron. No buildings, trees, roads, footprints, text, grid, hex guide, outline or watermark. Continuous high-resolution painted realism, no pixelation.

```
