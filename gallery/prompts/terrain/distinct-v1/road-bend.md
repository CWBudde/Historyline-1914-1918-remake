# road-bend

Recipe: `art/recipes/terrain/distinct-v1/road-bend.json`

- **art_direction**: distinct-continuous-illustrated-v1
- **cleanup**: Generated source preserved unchanged. Exact center crop then Lanczos3 exports, no original mask or pixels. Rectangular overscan; runtime hex clipping and connections not yet accepted.
- **date**: 2026-10-15
- **export_manifest**: art/terrain-distinct-v1.json
- **export_tool**: go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>
- **id**: road-bend
- **review**: Independent batch review recorded in docs/art/distinct-terrain-batch-01.md. No original-fidelity acceptance sought under updated brief.
- **schema_version**: 1
- **source**: art/source/terrain/distinct-v1/road-bend.png
- **source_sha256**: 4c347558277bfb36113dde0d9a904f8ee744ca8b06287a227a35f000a338ec6e
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/terrain/distinct-v1/meadow.png` | newly-authored-material-style-only | `8c8935efbbd463e39d49208e42ca0e87b140268a1b11941764dfa7c29c4fff85` |

## Prompt

```text
Create a NEW terrain tile illustration. Image 1 is solely the newly authored meadow MATERIAL/STYLE reference: retain its muted sage/moss greens, delicate continuous natural paint and overhead scale. Do not copy its texture exactly. Strict directly overhead orthographic view, landscape 6:5 opaque canvas, no frame or hex outline. Early twentieth-century European strategy-game road: a quiet grey-brown compacted macadam/gravel country road, fine aggregate, restrained ruts, irregular thin earth verges. Absolutely NO painted road markings, no white or yellow lines, no lane stripes, no bollards. One smooth broad bend: road enters through the BOTTOM canvas edge centered at 50% width, runs upward then bends to exit through RIGHT canvas edge centered at 25% height. Road width about18% of canvas, consistent through both ends, tangent vertical at bottom and horizontal at right. Critically the road surface and both verges continue completely to and beyond BOTH image edges; neither end closes into grass, no cul-de-sac. Grass fills surrounding area, no trees, structures, stones, vehicles or text. No intersections or extra tracks. Flat overcast diffuse light, crisp continuous curves with tiny organic detail, no blocky contours or pixel art. Invent original artwork, no legacy game palette or pixel layout, no resemblance tracing. Calm values support legible game units. Fill corners and all overscan.
```
