# grass-meadow-refdesign-v1

Recipe: `art/recipes/terrain/refdesign-v1/grass-meadow-01.json`

- **cleanup**: Selected built-in generated source copied unchanged. No post-generation retouching, tinting, sharpening or source-pixel compositing.
- **date**: 2026-10-20
- **id**: grass-meadow-refdesign-v1
- **schema_version**: 1
- **source**: art/source/terrain/refdesign-v1/grass-meadow-01.png
- **source_sha256**: d8e2f74e2a36cad07a0dc9b251a5bd3bcbbd9dae4d73341af72a7fb5349ecd60
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## original_image_inputs

| path | role | sha256 |
|---|---|---|
| `RefDesign.png` | User-supplied target: meadow palette, fine grass, local detail and sparse pebble material style reference only; not a crop or pixel source. | `3924e55c20ca63cbe31a7e4b3d469977f20c9ddf7802912e92782360b4d1aa69` |

## Prompt

```text
Use case: stylized-concept. Asset type: opaque grass terrain material master for a high detail overhead 2D strategy game. Input image 1 is RefDesign.png, a STYLE AND MATERIAL REFERENCE ONLY. Make a new 1536 by 1024 rectangular continuous close-up expanse of ONLY its open meadow surface: the muted warm olive, yellow-green and deep moss-green European grass visible in the reference between the road and tree groups. Match the reference's fine dense hand-painted realistic game texture, crisp tiny blades, short uneven grass tufts and little dry straw scratches, occasional tiny bare brown soil flecks, rare restrained grey-brown pebbles and sparse pale seed heads. At this closer material scale the short grass blades must be individually readable, not fuzzy noise. Fine marks should be quite small with lots of detailed surface between them. Create irregular natural local patch variation across the whole canvas without any repeating rounded cushion or pillow pattern. Fairly even ground overall; no large dark blobs, no large bright patches, no mounds. Light from upper left with very small shadows under a few raised tufts. Color must closely match reference open ground: subdued earthy olive green, not neon lime, not brown straw field, not saturated green. Orthographic directly overhead, full opaque edge-to-edge ground coverage. Do not include trees, bushes, roads, rivers, buildings, soldiers, craters, trenches, borders, hex outlines, grid, text, watermarks or transparency. Never imitate the old meadow-03 smooth puffed cushions. This is naturally textured flat ground, not moss pillows. Image 1's broad map composition is not the output: output only a new uninterrupted meadow material.
```
