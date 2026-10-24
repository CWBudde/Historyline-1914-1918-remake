---
id: terrain/stream
title: Stream
kind: terrain
status: revised
runtime_ids: stream13, stream135, stream14, stream15, stream24, stream246, stream25, stream26, stream35, stream36, stream46, stream13_w, stream135_w, stream14_w, stream15_w, stream24_w, stream246_w, stream25_w, stream26_w, stream35_w, stream36_w, stream46_w
masters: stream-bank-shared-v3.png, stream-bank-shared-v2.png, stream-bank-shared-v1.png
---

## Look
Winding blue-teal creek with broken white ripples, rounded brown-grey stony banks and dark wet lips.
Selected `stream-bank-shared-v3.png` is a byte-identical copy of the native 1254×1254 lakeside bank master.
All stream port sets share that material and `water-shared-v3.png` with the reusable shore tiles.

## Base prompt
Use the exact built-in generation prompt recorded in `art/terrain/lakeside/prompt.md`: the selected
stream bank shares that opaque rounded-rock material without cropping, rescaling or a separate
generation. The matching water material and its exact prompt are in `art/terrain/water/prompt.md`.

## Variants
| Variant | Runtime IDs | Prompt delta |
| --- | --- | --- |
| stream13 | stream13 | Stream wide bend between ports 1 NE, 3 S. |
| stream135 | stream135 | Stream even three-way fork to ports 1 NE, 3 S, 5 NW. |
| stream14 | stream14 | Stream straight run between ports 1 NE, 4 SW. |
| stream15 | stream15 | Stream wide bend between ports 1 NE, 5 NW. |
| stream24 | stream24 | Stream wide bend between ports 2 SE, 4 SW. |
| stream246 | stream246 | Stream even three-way fork to ports 2 SE, 4 SW, 6 N. |
| stream25 | stream25 | Stream straight run between ports 2 SE, 5 NW. |
| stream26 | stream26 | Stream wide bend between ports 2 SE, 6 N. |
| stream35 | stream35 | Stream wide bend between ports 3 S, 5 NW. |
| stream36 | stream36 | Stream straight run between ports 3 S, 6 N. |
| stream46 | stream46 | Stream wide bend between ports 4 SW, 6 N. |
| winter | stream13_w, stream135_w, stream14_w, stream15_w, stream24_w, stream246_w, stream25_w, stream26_w, stream35_w, stream36_w, stream46_w | The same courses in winter: open slate water in the centre, a pale ice rim along both banks, frosted grey stones with snow on their tops. |

## Edges & joins
The water channel is 18 world pixels wide with a 3.5-pixel stony bank, centred on the shared port
centres and continuing through each port, so creeks join each other, bridges and the lake shore.

## Post
The bank material retains the entire native lakeside master without enlargement.
`art-review -connected art/pipeline/connected-stream-v2.json` draws the water (shared water material) and banks
along each port set over meadow. Exports are fixed Lanczos3 at 2x (192x160) and 4x (384x320) with complementary half-open hex ownership; promote into `assets/terrain/<family>/` and re-index.

The selected `authored-rocks-v1` bank profile retains individual painted stone faces,
adds two logical pixels to the landward bank, and samples the authored water glints
without the old mirrored coset average or continuous generated foam line. Core ports stay unchanged.

`connected-stream-winter-v1.json` renders the `_w` forms with the same materials and geometry,
`season: winter` and `grass-winter-v1.png` as the ground: the dry bank stones turn cold grey and carry
snow on their light tops (the winter shore's treatment), and a pale ice rim 1.5–4 world px wide covers
the outer channel along both banks, its inner edge following lattice-periodic value noise so it joins
across every hex edge. The open water is the summer water; the pack grade of a `_w` tile turns it to
winter water. The winter stream mouths (`<shore>-outlet<n>_w`) carry the winter stream into the lake.
The tagged generator `TestGenerateWinterStreams` (`campaigngfx`) installs the streams and re-renders
the winter mouths. No new image: the measured criteria (bank < 1 % summer green, ice rim, open centre)
pass with the existing materials.

## Avoid
- visible tiling or repeated rounded "pillow" patterns
- perspective, horizon or tilted camera
- hex outlines, grid lines, labels or text
- pixel-art edges or nearest-neighbour enlargement
- muddy brown water
- regular cobble patterns
