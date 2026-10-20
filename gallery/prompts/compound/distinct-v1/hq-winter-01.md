# hq-winter-01

Recipe: `art/recipes/compound/distinct-v1/hq-winter-01.json`

- **date**: 2026-10-15
- **geometry_reference**: docs/art/compound-building-geometry.md
- **id**: hq-winter-01
- **output_manifest**: art/compound-hq-v1.json
- **schema_version**: 1
- **source**: art/source/compound/distinct-v1/hq-winter.png
- **source_sha256**: 064b38d83e33ddb61469cf2d16656070898bfaa12c9f8dd2c40c5a2b43da33f7
- **status**: reviewed-source-for-provisional-compound-exports
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/compound/distinct-v1/hq-summer.png` | new-authored-summer-architecture-locked-weather-edit | `8b26241b29cdf894265e2fd898de8f0d8fa2ba5fb7f104ba39baa3532b463f1c` |

## Prompt

```text
Use case: lighting-weather.
Asset type: transparent winter headquarters compound master for a 1914 strategy game.
Input image: EDIT TARGET, our newly authored summer HQ, 1280x1280 RGBA. It is NOT original-game artwork.
Primary request: change ONLY weather and season to restrained winter. Preserve the EXACT SAME 1280x1280 canvas, scale, registration, every building footprint, roof outline, chimney, cupola, canopy, courtyard shape, props and transparent outer silhouette. No cropping, reframing, zoom, rearrangement, enlargement or new structures. Copy architecture pixel positions from the target; use a weather layer.
Geometry anchors in the 1280x1280 image: main blue slate manor roof upper outline x345..910 y135..350, its four chimneys and central small green cupola unchanged; central projecting roof and north command entry at x550..705 y280..455. West long red auxiliary roof x102..300 y373..845; east roof x956..1152 y373..845. South red service building roof x345..909 y875..1062 with entry centered below at y1095. All anchors are preservation constraints, not new boxes.
Winter treatment: patchy thin pale snow on existing slate/clay roof planes, ridge edges and shrubs; roof construction and muted underlying blue-grey/red colors stay readable. Quiet diffuse cool daylight with short close shadows in exactly the same direction. Courtyard changes from pale earth to packed pale snow with a broad subtle worn-earth route, NO new geometric road and NO markings. Snow-tipped shrubs, lightly snow-dusted crates/barrels at their existing positions. Preserve clean roof silhouettes and architectural character.
Style: same continuous detailed hand-painted realistic miniature materials; coherent calm surfaces, readable at476x480; never pixel art or an upscaled pixel image. Straight-down orthographic roofplan, same camera as target.
Background: TRUE ALPHA TRANSPARENCY outside the existing compound overlay. Preserve all empty transparent padding. Do not fill grass, a rectangular snowfield, grey silhouette or backdrop.
No people, vehicles, flags, letters, labels, signs, logos, emblems, weapons, grid lines or tile outlines. Do not copy or reconstruct original-game art. This is a new seasonal edit of our new authored scene.
```
