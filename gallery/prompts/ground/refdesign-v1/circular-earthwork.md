# circular-earthwork

Recipe: `art/recipes/ground/refdesign-v1/circular-earthwork.json`

- **art_direction**: refdesign-v1
- **cleanup**: Generated master bytes retained unchanged. RefDesign.png is a visual reference only and remains untracked; no original/ or legacy raster pixels used.
- **date**: 2026-10-20
- **id**: circular-earthwork
- **integration**: Select as ground mud source and set optional mud_feature to circular-earthwork. Direct single-source bilinear feature composition, compact ellipse radius34x28 at48,40, zero opacity outside, existing overscan/Lanczos/ownership.
- **schema_version**: 1
- **source**: art/source/ground/refdesign-v1/circular-earthwork.png
- **source_sha256**: 4efb92fa6a89b1816063095a02621908814e9e3bae8a998cf24395465e6a0b02
- **status**: selected-candidate-pending-full-map-review
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `RefDesign.png` | user-supplied-visual-style-reference-only-not-edit-target-not-shipped | `3924e55c20ca63cbe31a7e4b3d469977f20c9ddf7802912e92782360b4d1aa69` |

## Prompt

```text
Use case: historical-scene
Asset type: one opaque top-down terrain feature master for a 2D strategy game; not a material swatch.
Input image: RefDesign.png is a user-supplied visual style reference, not an edit target. Match its circular brown earthwork beside the road around x350 y610, and its detailed miniature overhead painted realism.
Create ONE centered circular shell crater / abandoned circular field earthwork in an olive green European meadow. Straight overhead nadir view, near circular slightly oval footprint, no isometric walls. The complete rough earthen ring occupies x15%..85% and y14%..86%. Outer berm raised with coarse tan-brown clods and scattered stones; the INSIDE of ring has a deep dark brown cut bank, readable broken shadow scalloped around its inner edge. Inside center is mostly intact muted olive grass and bare compact earth with a few straw/root fragments, brighter than the dark cut-bank ring. Rough broken irregular lips, chunks of excavated earth and sparse gray-brown stone ejecta around the berm, no perfect vector circle. Ring soil tan/umber with small bright top edges under upper-left light, inner banks dark. Fine naturalistic painterly miniature texture with deliberate large forms, like the supplied reference. Grass backdrop should be calm olive ochre green. No roads, buildings, trees, units, text, UI, water, lens effects, border, vignette, giant boulders, trenches or other crater circles. Full opaque square canvas, isolated single feature with grass around all edges. Restrained realistic reference palette, never cartoon, pixel art, or photoreal photograph.
```
