# art/recipes/units/reference-v2/infantry.json

Recipe: `art/recipes/units/reference-v2/infantry.json`

- **acceptance**: Original-size silhouette, facing, user visual likeness and runtime acceptance remain review gates. No computed greater-than-90-percent approval.
- **art_direction**: reference-faithful-painted-readability-v2
- **date**: 2026-10-16
- **schema_version**: 1
- **source**: art/source/units/reference-v2/infantry.png
- **source_sha256**: a81cbe6bc35d4e0709e9b1fdaf6690821f303e77d61b1717eb88369c333a8b80
- **status**: selected candidate
- **tool**: built-in image_gen
- **transparent_background**: true

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `legacy/python/gfx/units/icons/infantry.png` | unchanged original subject/proportions/camera reference only | `cadc0bd0092050d7a82cf2a471fc42f97b4bcf3088388f2f2ca7389383ee2a3d` |
| `art/source/units/distinct-v1/infantry.png` | previous authored paint/key convention reference | `a151ab9f110782ead6eba6bc4af3b4a17eb789204c5f19e4ca2e786614912290` |

## Prompt

```text
Use case: historical-scene / reference-faithful game sprite atlas. Image1 is the tiny ORIGINAL infantry icon: subject arrangement, overhead camera and readable silhouette reference. Image2 is the previous authored atlas: continuous painted style and MAGENTA recolouring convention only, do NOT copy its tall close-up soldiers, bulky packs or triangular spacing. Paint a NEW 1024x1536 RGBA atlas, exactly 2 columns x 3 rows of 512x512 square panels, no gutters. Six row-major headings N(up), NW(up-left), SW(down-left), S(down), SE(down-right), NE(up-right). Each panel has exactly THREE small WWI riflemen seen from a high overhead strategy camera, more overhead than image2, with visible helmet tops, shoulders, simple uniform bodies and distinct legs/rifles. They are the same unit in six individually painted views, light consistently upper-left. In NORTH/SOUTH panels place the three soldiers spread broadly side-by-side: left and right soldiers at almost the same height, central soldier only slightly forward, matching image1's wide shallow squad silhouette. North group opaque bounds about x=38..474, y=140..390 in its 512 panel; the formation must look WIDE, not tall or bunched into a triangle. Individual riflemen modest, no oversized backpacks, heads or toy-like thick 3D bodies. Rotated headings preserve clear three-person identity and the formation's actual ground orientation. Center the squad on panel (256,261). All six groups must stay inside x/y=38..474, including rifles; outside these bounds EXACT alpha0, no haze or ground shadows. Detailed hand-painted historical game illustration with crisp dark silhouette edges, warm light stone/khaki material highlights, restrained depth, NO photoreal/3D miniature rendering. Technical saturated MAGENTA ONLY on uniform cloth for later owner recolouring; muted steel helmets/rifle barrels, light tan leather equipment, natural skin and boots. Strong light/dark separation to remain recognizable at 48x48 Lanczos reduction. Background genuinely transparent, also between soldiers and legs. No ground, backdrop, colored glow, vignette, labels, grid, arrows, insignia or text. Match recognizable reference subject and proportion as closely as possible with newly authored high-resolution detail.
```
