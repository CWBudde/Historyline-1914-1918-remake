# hamlet-winter-refdesign-v1

Recipe: `art/recipes/compound/refdesign-v1/hamlet-winter.json`

- **alpha_validation**: Go image.Decode check confirmed RGBA, fully transparent pixels, antialiased partial alpha, zero saturated red opaque pixels. A few very low-alpha exterior pixels mean alpha>=128 bounds are more useful than alpha>0 bounds for visual framing.
- **date**: 2026-10-20
- **framing_note**: Building solid-alpha bounds occupy about x378..1067,y341..796 of 1448x1086; compact single hut, centered. Root may adjust affine framing for target scale.
- **id**: hamlet-winter-refdesign-v1
- **schema_version**: 1
- **source**: art/source/compound/refdesign-v1/hamlet-winter.png
- **source_sha256**: 7096508ef69bfd4bccf6105a275c1598e9b51a0b88f9a02ae06c16838fe9f762
- **status**: visually-reviewed-authored-master-awaiting-root-compound-integration
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true
- **use_case**: lighting-weather
- **visual_review**: Seasonal edit preserves summer composition and silhouette; snow remains patchy.

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/compound/refdesign-v1/hamlet-summer.png` | new-authored-season-edit-target | `4e20895a10f9563d9429cda3cbe8825a8525c37cab0b2e5be78c9254b6d216d4` |

## Prompt

```text
Use case: lighting-weather. Asset type: transparent winter single-building cutout for the same fine-textured overhead European strategy map. Input image 1 is the edit target: the newly authored one compact timber hut. Change ONLY season. Preserve precisely the roof geometry, hut silhouette, horizontal orientation, chimney position, shingle pattern, threshold stones, original canvas and genuine outside alpha transparency. Add a modest thin uneven layer of settled natural snow across the four roof planes, with dark weathered wooden shingles still clearly visible through patchier areas and along eaves and ridge. Snow lies flat on roof and tops of threshold stones, no tall snow caps, no icicles extending silhouette. Keep small narrow ground fringe subdued olive-brown with touches of snow, not a broad snow island. Same warm diffuse upper-left lighting and short soft grounded lower-right shadow. Camera strict vertical nadir orthographic; do not expose more side walls or shift the camera. All empty background stays truly transparent. No new object, person, structure, garden, road, tree, fence, text, border, checkerboard or backing.
```
