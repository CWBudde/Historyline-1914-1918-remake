# summer-atlas-refdesign-v1

Recipe: `art/recipes/forest/refdesign-v1/crowns-summer.json`

- **date**: 2026-10-20
- **id**: summer-atlas-refdesign-v1
- **schema_version**: 1
- **source**: art/source/forest/refdesign-v1/crowns-summer.png
- **source_sha256**: 5e2555f8508f4e75e715f719fe5eccb31c679ed1501449496ea5ff20516d97a5
- **status**: visually-reviewed-source-passes-current-atlas-guards-awaiting-root-integration
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true
- **use_case**: style-transfer
- **visual_review**: Warmer sunlit broadleaf crowns and deeper green conifers remove prominent blue-silver radial stars. Material palette is closer to RefDesign. Existing canopy template still has considerably more open grass between crowns than reference; source change alone does not resolve forest density.

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/forest/distinct-v1/crowns-summer.png` | new-authored-summer-four-crown-atlas-material-edit-target | `a1f12a95d12e48cefd79d1236efa0b30a93bb98e3e5a3af46da5cdf95e4e5d72` |

## Prompt

```text
Use case: style-transfer. Asset type: transparent summer FOUR-CROWN atlas for a fine-textured overhead European strategy map. Input image 1 is the edit target, our NEW authored existing atlas, NOT original game artwork. Conservative material-only edit toward the viewed RefDesign map's warm yellow-green and deep forest-green foliage. Keep EXACT 1254x1254 square canvas; same four separate trees, same 2x2 row-major species order: broadleaf-a upper left, broadleaf-b upper right, conifer-a lower left, conifer-b lower right. Preserve all four crown centres at the existing quadrant centres and their exact overall diameters/silhouettes and orthographic vertical nadir camera. Each complete crown and short local shadow remains safely inside the central 65 percent of its own 627-square panel. Leave genuine alpha-zero empty margins at least 17 percent of EACH panel along every side. Material change: upper broadleaf crowns should have naturally uneven rounded clumps of many tiny leaves, warm olive-green and yellow-green sunlit tips on upper left, deeper almost black green occlusion in interior gaps and lower right, with subtle irregular branch-level shapes. Increase realistic depth and fine clustered foliage, avoid waxy oversized individual leaves. Lower conifers stay radial overhead conifer crowns in same silhouettes, but change blue/silvery teal needles to deep natural forest-green with restrained warm olive-green tips and small brown twig details, fine dense branchlets, dark occlusion; remove conspicuous blue/cyan silver colour entirely. All four are the same woodland family, subdued and natural, warm diffuse upper-left daylight, short grounded soft shadows toward lower right, crisp natural foliage edges. Do NOT add a background, grass, trunks, roads, tree group, extra crown, horizon, visible side view, text, grid or border. No change to species order, anchors, crown sizes, footprint, canvas or layout. Clean antialiased alpha edges without coloured fringing or faint floating pixels in panel margins. Actual transparent background, no checkerboard.
```
