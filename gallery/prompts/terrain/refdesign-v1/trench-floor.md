# trench-floor

Recipe: `art/recipes/terrain/refdesign-v1/trench-floor.json`

- **art_direction**: refdesign-v1
- **cleanup**: Generated master bytes retained unchanged. RefDesign.png is a visual reference only and remains untracked; no original/ or legacy raster pixels used.
- **date**: 2026-10-20
- **id**: trench-floor
- **integration**: Select as connected-trench trench material; existing lattice sampler and reciprocal numeric runtime ports retained.
- **schema_version**: 1
- **source**: art/source/terrain/refdesign-v1/trench-floor.png
- **source_sha256**: 118093f94da2873b4a66970a0864cd62ab10fd9af0c751c52b90e63f5a872062
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
Asset type: opaque compacted floor-earth material texture for narrow WW1 overhead trenches.
Style reference: RefDesign.png user supplied map, match realistic fine miniature ground detail and brown soil of the circular earthwork. Full square canvas closeup of dark compacted umber soil, earthy granular clay, tiny angular pebbles, restrained root splinters and straw, darker damp seams in irregular compacted earth. No water shine. Moderate contrast in texture, near black-brown crevices but enough ochre-brown detail visible in middle tones. Diffuse overhead light. This is ONLY an edge-to-edge continuous soil material, NO trench shape, no crater, no walls, no sandbags, no planks, no grass, no footprints, no stripes, no repetitive pattern, no text or border. Realistic painterly fine 2D strategy map art, not pixel art.
```
