# earth-bank

Recipe: `art/recipes/terrain/refdesign-v1/earth-bank.json`

- **art_direction**: refdesign-v1
- **cleanup**: Generated master bytes retained unchanged. RefDesign.png is a visual reference only and remains untracked; no original/ or legacy raster pixels used.
- **date**: 2026-10-20
- **id**: earth-bank
- **integration**: Select as connected-trench earth material; existing lattice sampler and reciprocal numeric runtime ports retained.
- **schema_version**: 1
- **source**: art/source/terrain/refdesign-v1/earth-bank.png
- **source_sha256**: 6df0ffb17725f7c7ef81bbafd1c8537ad2a30b62beac21a4ce6ce8a4c522086d
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
Asset type: opaque continuous earth-bank material texture for WW1 overhead strategy trench edges.
Reference: user supplied RefDesign.png establishes fine realistic miniature painted detail and tan/umber soil of its circular earthwork, not a source to copy. Match that coarse bank soil closely. Fill full square canvas with rough dug tan-brown earth and small soil clods, shale crumbs, tiny broken roots and sparse gray/tan gravel stones. Absolutely no crater outline, trench outline, grass, road, timber, footprints, patterns, vignette or shadows of offscreen objects. Fine irregular earthy texture with some slightly larger clumps for readable trench bank detail at tiny scale, natural mixture of medium warm umber with lighter sandy clod faces and dark crevices. Diffuse upper-left light, medium contrast. Nadir orthographic overhead material. Painterly realistic 2D game map art, not cinematic and not pixel art. Opaque edge-to-edge square material.
```
