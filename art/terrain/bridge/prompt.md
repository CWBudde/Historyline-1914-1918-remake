---
id: terrain/bridge
title: Footbridge
kind: terrain
status: candidate
runtime_ids: bridge25, bridge36
masters: 
---

## Look
Narrow plank footbridge over a stream. In the reference it is a weathered grey timber deck with thin
side rails, crossing the blue creek between stony banks.

## Base prompt
Transparent cut-out of a narrow timber plank footbridge seen straight from above: weathered grey-brown
boards laid across the span, thin wooden side rails, short abutments of pale stone at both ends, soft
shadow on the water side to the lower right. The deck runs straight across the canvas; nothing else.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| bridge25 | bridge25 | Footbridge straight run between ports 2 SE, 5 NW, across the stream. |
| bridge36 | bridge36 | Footbridge straight run between ports 3 S, 6 N, across the stream. |

## Edges & joins
The deck runs straight between the two ports, centred on the port centres, and sits over a stream
that continues through the other ports of the crossing; the path continues beyond each end.

## Post
Derived only: `art-review -mixed art/pipeline/connected-mixed-v1.json` builds footbridges from the timber,
water, bank and meadow materials of the connected manifests. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- arched or perspective bridges
- masonry viaducts
