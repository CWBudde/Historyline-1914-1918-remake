---
id: terrain/mount
title: Massif
kind: terrain
status: candidate
runtime_ids: mountE, mountE_w, mountN, mountN_w, mountS, mountS_w, mountW, mountW_w
masters: mountain.png, mountain_w.png
---

## Look
A grassy massif over a four-hex diamond. In the reference it is a broad ridge of long grassy slopes
with bands of dark scrub and bare grey-beige rock along the crest (average about RGB 94,98,45).

## Base prompt
Transparent terrain cut-out, square, strict nadir view: a broad, low, grassy massif with long rounded
slopes of olive meadow, darker green scrub bands in the folds and a crest of weathered warm grey-beige
rock slabs with olive moss in the cracks. Soft relief modelled by the upper-left light, short soft
shadows to the lower right. Irregular outline, generous transparent margin. No summit cone, snow,
trees, paths or buildings.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| cell N | mountN, mountN_w | Slice of the massif at the N cell of the diamond. |
| cell E | mountE, mountE_w | Slice of the massif at the E cell of the diamond. |
| cell S | mountS, mountS_w | Slice of the massif at the S cell of the diamond. |
| cell W | mountW, mountW_w | Slice of the massif at the W cell of the diamond. |
| winter | mountN_w, mountE_w, mountS_w, mountW_w | Same scene with thin settled snow on flat rock and slopes; rock and moss still visible. |

## Edges & joins
Sliced into four hex crops with shared placements; slopes fade into the seasonal meadow before the
outer edges of the diamond.

## Post
`art-review -compound art/pipeline/compound-mountain-v1.json` scales the mountain master into the four-cell
scene for both seasons and slices it. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- tall peaks, volcanic cones or ranges
- snow in summer
