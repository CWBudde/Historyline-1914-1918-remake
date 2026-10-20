# depot-summer-reference-02

Recipe: `art/recipes/compound/reference-v2/depot-summer.json`

- **acceptance**: User requests >90% original-size Lanczos resemblance; generation prompt target only, not measured acceptance. New master needs complete assembled comparison and edge access verification.
- **date**: 2026-10-15
- **geometry_reference**: docs/art/compound-building-geometry.md
- **id**: depot-summer-reference-02
- **original_pixels_in_output**: none intentionally copied; original assemblies are visual references only and remain ignored
- **schema_version**: 1
- **source**: art/source/compound/reference-v2/depot-summer.png
- **source_sha256**: 71c8e02269ed8b66f0c903abd541ee5ddab8a0c176f4b9afae8a9eb2affd56c5
- **status**: newly-painted-candidate-requires-assembled-review
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `build/compound-ref/depot-summer.png` | private assembled original visual composition reference, not edited or copied into output | `<nil>` |

## Prompt

```text
Use case: historical-scene. Asset type: newly painted high-resolution 1914 strategy-game depot/camp master, transparent overlay. Input image 1 is a COMPOSITION AND COLOR REFERENCE ONLY, a tiny original four-tile medical camp assembly enlarged with nearest-neighbor. Repaint it from scratch at 1536x1024, keeping its exact overhead silhouette, placements, broad color patches, and proportions. This must read as the same camp when reduced with Lanczos to 238x160, targeting greater than 90% structural resemblance. It must NOT become four red-roof warehouses or an elegant manor courtyard. Straight-down orthographic roof plan, no perspective tilt. Main camp occupies a low wide rectangular fenced enclosure from x21%-78%, y27%-73%. LEFT: one squat L-shaped olive-brown military medical barracks, vertical leg x21%-30% y36%-67%, upper leg reaches x39% at y35%-43%; a clearly visible RED CROSS on a pale roof patch of the vertical leg. RIGHT: three small white canvas tents in a vertical column at x64%-72%, at y38%,51%,64%, their roofs distinct and well shaded. Center remains mostly empty/transparent so shared meadow shows through. NORTH: narrow grey period road centered x50%, entering from top and continuing vertically down to y58%, with pale broken center markings; a horizontal pale gate boom at y10%. Low weathered stone-and-timber perimeter fence around the rectangle, finely spaced posts and restrained blue-grey caps. SOUTH: small separate fenced annex between x38%-58%, y76%-92%, with one low horizontal canvas supply shelter and stacked material. One small dense rounded olive bush beside the left medical hut. Materials have real depth: weathered canvas weave, wood slats, roof seams, tiny dark cast shadows, dirt marks beside objects. Muted original olive brown white grey palette, no added decorations. Quiet upper-left daylight. Everything else genuinely transparent, including most interior ground; do not paint a hex tile or grass backdrop, do not add guide lines, labels, lettering, people, vehicles, flags, vignette or watermark. Keep artwork inside the four-hex silhouette visible in reference, with restrained margin; road alone may touch top. High detail in microtexture while low-frequency color layout stays almost identical to original. Output exactly 3:2 landscape, entire camp visible without clipping.
```
