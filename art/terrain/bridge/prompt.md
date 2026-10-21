---
id: terrain/bridge
title: Footbridge
kind: terrain
status: candidate
runtime_ids: bridge25, bridge36, standing/bridge25, standing/bridge36
masters: bridge-25-shared-v1.png, bridge-36-shared-v1.png, bridge-25-export-shared-v1.png, bridge-36-export-shared-v1.png
---

## Look
Narrow plank footbridge over a stream. In the reference it is a weathered grey timber deck with thin
side rails, crossing the blue creek between stony banks.

## Base prompt
Transparent cut-out of a narrow timber plank footbridge seen straight from above: weathered grey-brown
boards laid across the span with fine grain, nailheads and dark seams, thin wooden side rails and
posts, short abutments of pale stone at both ends. Strict nadir deck plan, soft upper-left light and
short lower-right shadow. The deck runs straight along its declared axis; no water, ground,
approach path, letters, flags or other objects.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| bridge25 | bridge25, standing/bridge25 | Separately painted SE/NW bridge, axis +29.396 degrees from horizontal. |
| bridge36 | bridge36, standing/bridge36 | Separately painted vertical N/S bridge. |

## Edges & joins
The finite timber deck is registered to 48 logical pixels along its port axis and 12 pixels across,
centred on `(48,40)` within a 96x80 canvas. Thin rails, posts and stone abutments extend beyond the deck.
The path continues beyond each deck end and the stream passes underneath.

## Post
Built-in imagegen produced both independent true-alpha paintings at native 1374x1145. The previous
tiles supplied axis references and the selected crater supplied fine material detail. A follow-up
prompt corrected the narrow deck proportions; neither orientation is a rotated copy of the other.

The `-export-` masters apply bounded affine registration and transparent padding at 1152x960
(12 source pixels per logical pixel). `bridge-registration-v1.json` records native hashes, dimensions,
reviewed deck landmarks and every matrix. Native sampling density exceeds 16 pixels per logical pixel
on both axes, so no enlargement is needed. Small axis corrections retain the painted upper-left light.

`terrain-stamps -manifest art/pipeline/bridge-standing-v1.json` exports the prepared sources as
`standing/bridge*`, with 2x 192x160 and 4x 384x320 PNGs and anchor `(48,40)`. The v2 mixed manifest's
`bridge_artwork: standing` keeps only the original path approaches beyond the deck ends, removing old
finite timber, rails and earth abutments before density filtering. The default v1 mixed export
remains available with its derived bridge artwork.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- arched or perspective bridges
- masonry viaducts
