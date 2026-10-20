# summer-canopy-refdesign-v2

Recipe: `art/recipes/forest/refdesign-v1/crowns-summer-canopy-v2.json`

- **date**: 2026-10-20
- **id**: summer-canopy-refdesign-v2
- **output_manifest**: art/forest-terrain-v1.json
- **parent_generation_recipe**: art/recipes/forest/refdesign-v1/crowns-summer-dense.json
- **parent_generation_recipe_sha256**: 5af496f092e0187d98216dcf67195cc7f245f4b569844eef02e9ec49b240bc57
- **schema_version**: 1
- **source**: art/source/forest/refdesign-v1/crowns-summer-canopy-v2.png
- **source_sha256**: 7a5fe25d5d41ad067e463f1938fac0dd280ed706022ce3f6433a10096f3732e5
- **status**: selected-source-reviewed-in-private-actual-export-proof
- **tool**: built-in image_gen.imagegen
- **transparent_background**: true
- **use_case**: style-transfer
- **visual_review**: Replaces bright golden rosette clusters with coherent dark forest-green masses, irregular broadleaf branch lobes and restrained green highlights. Lower conifers remain darker and less ornamental. Existing 24-world-pixel geometry retained after parent review; crown widths about72..74percent of627px panels. Material is substantially closer to user reference; reference individual crowns remain variable and somewhat larger in places. Selected optional canopy-26.5-30-v1 rendering enlarges the unchanged source panels modestly and uses the broadleaf-heavy fixed roster; parent approved the actual exported comparison.

## generation_inputs

| path | role | sha256 |
|---|---|---|
| `art/source/forest/refdesign-v1/crowns-summer-dense.png` | new-authored-four-crown-atlas-redesign-target-step-1 | `1f25c3bea12be1a0a43ec87e5366bb190a2c44bdfa96a2cbd6bd4f69a49a4755` |
| `RefDesign.png` | user-provided-reference-image-for-natural-dark-canopy-style-only-step-1; reference-only-not-shipped | `3924e55c20ca63cbe31a7e4b3d469977f20c9ddf7802912e92782360b4d1aa69` |
| `build/refdesign/forest-canopy-v2/study01.png` | new-authored-dark-canopy-framing-edit-target-step-2; ignored-private-intermediate | `5e0eaece03ae44e3b3c93e7e237e31dd3f7bfd61acf2e150dd7ffe34f2ad7950` |

## Prompt

```text
Use case: precise-object-edit. Keep EXACT1254x1254 transparent four-crown atlas and all the current DARK GREEN foliage art, camera, lighting and material. Correct ONLY framing in the four627x627 panels. Each crown is slightly too large: shrink EACH complete crown uniformly by5percent, and centre its complete outer silhouette in its own panel. Aim about500pixels diameter, safely inside local x60..567,y60..567 in every627-square panel. All leaf tips and shadows must fit inside this central area. Keep top-left/top-right broadleaf and bottom-left/bottom-right conifer order. All outer50pixel margins of each panel must be strictly alphaZERO, no noise or stray alpha. Do not enlarge, redraw, recolour or change the style of any tree. Only slight reduction and recentring. No background or checkerboard, no grid or text; preserve actual alpha transparency.
```

## full_prompt_history

Step 1: dark natural canopy redesign
Use case: style-transfer. Asset type: genuinely transparent summer four-tree atlas for overhead game-map woodland. Input1 is the NEW authored atlas to EDIT; image2 is user-provided RefDesign map, STYLE REFERENCE ONLY (do not copy image2 pixels or buildings). Redesign only the foliage materials of the four tree crowns in image1 to match the large dark irregular natural trees in image2. Current input1 is WRONG: its many tiny bright yellow rosettes look like golden flowers and a granular bush carpet when reduced. REMOVE that rosette/flower appearance completely. Each crown must read as one mature tree, not a flower bush. Build broadleaf crowns from 5 to9 broad irregular rounded branch-canopy masses with deep dark green gaps and chunky natural leafy edge lobes, only restrained larger medium-green leaf-cluster patches lit from upper left. Dark forest green should dominate at least75percent of crown area; deep lower-right and inter-clump shadows almost black green. Main foliage RGB impression roughly28,52,22 to44,76,30; small sunlit patches roughly70,105,35 to92,126,43. No yellow, gold, orange, lime-white, flower dots or tiny pale rosettes anywhere. Leaves should merge into coherent coarse aerial canopy forms, not hundreds of equally bright points. Lower conifers remain two irregular conifer crown silhouettes, but denser dark-green branch masses with a few coarse olive-green lit tips, NOT spiky ornamental radial starbursts and NO light yellow needle tufts. Match image2's grounded natural textured aerial 3D relief, dark short local shadows, realistic woodland, soft upper-left daylight. Camera strict vertical nadir orthographic. Preserve1254x1254 square canvas, 2x2 panel order broadleafA,B then coniferA,B, current tree centres and approximately current complete crown diameters, each entire crown safely within central80percent of its627-square quadrant. All four panel OUTER60pixel margins must be exactly alphaZERO, no faint stray alpha, floating noise or shadows. Keep separate crowns and empty gutters. No ground, lawn island, trunks, visible side views, buildings, roads, grid, text, checkerboard or opaque backdrop. Reference2 supplies only the dark natural tree style; final output is only FOUR trees with actual transparent alpha.

Step 2: transparent-panel framing correction
Use case: precise-object-edit. Keep EXACT1254x1254 transparent four-crown atlas and all the current DARK GREEN foliage art, camera, lighting and material. Correct ONLY framing in the four627x627 panels. Each crown is slightly too large: shrink EACH complete crown uniformly by5percent, and centre its complete outer silhouette in its own panel. Aim about500pixels diameter, safely inside local x60..567,y60..567 in every627-square panel. All leaf tips and shadows must fit inside this central area. Keep top-left/top-right broadleaf and bottom-left/bottom-right conifer order. All outer50pixel margins of each panel must be strictly alphaZERO, no noise or stray alpha. Do not enlarge, redraw, recolour or change the style of any tree. Only slight reduction and recentring. No background or checkerboard, no grid or text; preserve actual alpha transparency.
