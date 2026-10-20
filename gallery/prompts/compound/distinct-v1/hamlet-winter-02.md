# hamlet-winter-02

Recipe: `art/recipes/compound/distinct-v1/hamlet-winter-02.json`

- **date**: 2026-10-15
- **id**: hamlet-winter-02
- **output_manifest**: art/compound-hamlet-v1.json
- **schema_version**: 1
- **source**: art/source/compound/distinct-v1/hamlet-winter.png
- **source_sha256**: cfde8d1a24a57fbeee54ebe7b61421f2f5bcb2bab6376af9f0e8e26ce1981257
- **status**: selected-provisional-source-minor-native-edge-flecks-retained-as-review-caveat
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true
- **use_case**: precise-object-edit

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/compound/distinct-v1/studies/hamlet-winter-01.png` | new-winter01-alpha-edge-cleanup-target | `35ca12846c88d40ac1cdd091d3c9f469698a2cfc7663dfa84f3929f0e2803bac` |

## Prompt

```text
Use case: precise-object-edit. Image1 is our newly illustrated transparent winter hamlet. Fix ONLY the saturated electric CYAN and BLUE color contamination along its alpha edges, especially around upperleft bare tree, upperright tree, right/lower fence and lower garden edge. Replace those unnatural colored fringe pixels with neutral pale grey-white snow, muted grey-brown branches, and properly antialiased neutral transparency. There must be NO saturated blue/cyan outline anywhere. Preserve EXACTLY all existing architecture/roof silhouettes, roof tile details, buildings positions/scale/orientation/chimneys, well, garden/fence shapes and locations, court, trees silhouettes, overhead camera, fine continuous illustration, 1374x1145 landscape6:5 canvas and alpha transparency; no movement or redesign. No enlargement or added ground: same generous transparent outer margin. Clean alpha-matted snow/wood edges without an outline or opaque backing, no checkerboard painted into the image. Existing restrained desaturated blue-grey snow shadows inside objects stay naturally subtle. Background must remain genuinely transparent.

```
