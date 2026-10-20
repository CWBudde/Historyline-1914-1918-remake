# meadow-reference-v2

Recipe: `art/recipes/terrain/reference-v2/meadow-03.json`

- **cleanup**: Generated source preserved unchanged. No post-generation sharpening, procedural detail, original-pixel blending or color correction.
- **date**: 2026-10-15
- **id**: meadow-reference-v2
- **schema_version**: 1
- **source**: art/source/terrain/reference-v2/meadow-03.png
- **source_sha256**: 4de543f0906e1311eff5ae4f202eb8fa68665c67103039bee089e60e02c68eb4
- **status**: candidate
- **tool**: built-in image_gen.imagegen
- **transparent_background**: false

## original_image_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/hexTypes/hex_grass.png` | Only image input: composition, muted olive palette and low-frequency restoration reference; original file remains reference-only. | `67257c01c1e3e30ddc438e85e5e376cd1aa72f888cfb2851b288061e1eb7a0de` |

## Prompt

```text
Use case: identity-preserve image restoration / exact composition up-resolution. This is a FAITHFUL RESTORATION of the ONE attached96x80 grass tile, not a new grass texture design. Enlarge the reference to a crisp1374x1145-like high-resolution6:5 rectangular image while preserving its exact visible internal composition, average muted light olive color, soft blurred rounded grass cushions and their precise positions, outlines and shading. Every low-frequency reference shape stays where it is: DO NOT invent a different clump arrangement; do not substitute generic grass noise, directional swirls, tiny evenly repeating tufts, yellow-green tussocks or sharp dark pits. Reconstruct continuous organic detail smoothly within the same softly blended cushions, adding delicate short grass-blade details only on top of their existing light/dark areas. At96x80 Lanczos reduction the high-resolution restoration should visually return to the SAME original image, preserving its same muted light-olive green colors and low-contrast soft relief. Keep total high-frequency detail restrained so new blades disappear on reduction without changing original light/dark masses. The reference has a hex-shaped grass region and transparent corners: output OPAQUE full rectangular grass overscan at all four corners by continuing adjacent original cushions naturally, but keep the original hex-interior composition undistorted and aligned to the same full6:5 canvas. No visible hex outline, black corners, grid, pixel blocks, border, labels, text, transparency, earth, stones, flowers, trees, paths or other objects. No new design decisions. Camera stays directly overhead and original diffuse lighting stays unchanged. RESTORE AND UPRES THE SAME GRASS, preserve image composition as exactly as possible.
```
