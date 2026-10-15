# Seasonal hamlet geometry

The enabled background IDs `house` and `house_w` are single-cell terrain, not
depot/HQ/factory containers. Each uses one new overhead three-cottage master,
with its courtyard, well, gardens, walls and nearby trees retained together.
No special gameplay entrance or additional runtime ID is introduced.

`hamlet-summer` and `hamlet-winter` are bounded `-compound` scenes. Each has
a 96×80 world canvas, one C cell at (0,0), empty `entry_cell`, and the polygon
(25,0), (71,0), (96,40), (71,80), (25,80), (0,40). Runtime exports are
192×160 and 384×320. The [manifest](../../art/compound-hamlet-v1.json) pins
the two transparent 1374×1145 masters and existing new meadow/snow materials.

Both masters use uniform inset 0.75: scale 72/1374, offset (12,10) in world
pixels. The complete source receives a single premultiplied, signed Lanczos3
resampling into the 4× scene; 2× comes from the whole 4× artwork with Lanczos3.
The existing compound guard rejects opaque artwork within eight world pixels
of the outside boundary, leaves the first four clear, and fades alpha over 4–8.
No roof, fence or tree is erased to force a fit. Full 4× preflight reports zero
guarded opaque pixels and a 12.0177-world-pixel minimum opaque distance in both
seasons. See the exact guard/transform records in the
[export recipe](../../art/recipes/compound-hamlet-v1.json).

Ground is assembled from the existing actual seasonal ground PNGs. Artwork is
composed over that ground, then cropped with the existing complementary
half-open polygon ownership; no filtering follows ownership. Actual exported
tile PNGs are decoded and placed with ordinary source-over drawing for review.

Single-cell scene origins are (0,0) for absolute line parity 0 and (71,40) for
parity 1; surrounding-patch canvas origins are (-71,-80) and (0,-40).
The seven visible cells are C plus six neighbors at offsets (71,-40), (71,40),
(0,80), (-71,40), (-71,-40), (0,-80). These absolute coordinates matter even
when normalized patch PNGs look identical: the ground material must retain
its correct world phase. Tests include a witness that distinguishes the
incorrect quad-layout offset from these single-cell origins.

This contract covers pixel-aligned 2×/4× candidates. Forest/building world
scale, repeated ground motifs, fractional zoom, complete maps and runtime
selection/integration still require review. See the
[gallery](compound-hamlet-review.html) and [delivery record](distinct-terrain-batch-15.md).
