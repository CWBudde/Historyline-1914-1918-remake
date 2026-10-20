# forest-winter

Recipe: `art/recipes/terrain/distinct-v1/forest-winter.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Preserve generated source; fixed centered exact6:5 crop then Lanczos3 384x320->192x160. No original mask/pixels. Rectangular overscan; geometry, phase and seams pending.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: forest-winter
- **review**: Independent source/reduction review: docs/art/distinct-terrain-batch-03.md. Candidate source only; runtime acceptance separate.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/forest-winter.png
- **source_sha256**: 99d5d7316e0673e30c1c870002886c12f70b2eb421de42af6180e67e13bbb0e4
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/forest-sparse.png` | newly-authored-material-style-only | `3af519fe9bc698d1eec5068aaa6adb5213380ec6ac6568166a7ac58d056f1915` |
| `art/source/terrain/distinct-v1/snow-ground.png` | newly-authored-material-style-only | `127fd16b8ae4951273885c0863c14f1db355be774830ed557be99d4349a55ee3` |

## Prompt

```text
ONE new high resolution WINTER terrain illustration for a European1914 strategy game, landscape6:5 full opaque rectangle. Strict vertically overhead orthographic view, calm diffuse overcast upperleft light, continuous finely painted natural detail, quiet broad values. Legible192x160, fullsize crisp organic contours without squarepixels. No horizon, perspective, UI, lettering, hexoutline, frame, vignette or oldgameartimitation. All referenceimages NEWLY AUTHORED, not legacy assets. Image1 supplies newly authored sparse woodland scale and material illustration style; image2 supplies new pale bluegrey/ivorysnow. Paint winter aerial woodland with aboutseven mature trees at the same crown scale as image1, including four leafless deciduous crowns with fine branching radial grey/brown branch networks and three compact dark bluegreen conifer crowns seen straightdown. Snow rests in small broken patches on conifer branches, not icing blobs. Canopies occupy40to50percent, ample open snowground aroundtrees. No greenleafed broadleafcanopy in winter, no upright triangular side-view pine silhouettes, no visible tree trunks or brown vertical branches like standing miniature trees; each aerial crown branches radially in all directions from its center. Subtle soft short tree shadows, no dramatic longshadows. No tracks, logs, huts, paths or animals. Natural uneven cluster layout, not seven evenly spaced icons. Fill corners with snow and grass tips fromnewwintermaterial.
```
