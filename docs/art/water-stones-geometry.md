# Water stones and shared-water geometry

`waterStones` is an existing shared-season single-cell background ID. Its new
`water-stones` compound scene is 96×80 world pixels, with C at (0,0), empty
entry cell and the polygon `(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)`.
The literal runtime spelling and shared season match the loader ledger; no
`waterStones_w`, `grass` alias or gameplay entrance is introduced.

The [manifest](../../art/compound-water-stones-v1.json) declares `grounds.shared`
using the existing newly authored `water-open.png`. The compound manifest now
permits this third ground slot exactly when a shared scene needs it. Seasonal
summer/winter scenes retain their previous grass/snow contracts. Ground exports
are named `ground/water-2x.png` and `ground/water-4x.png`; both are byte-identical
to the existing [shore water exports](../../art/exports/shore-v1/water-2x.png).
Report and per-scene profile text describe shared water explicitly. Historical
seasonal profiles and output bytes remain unchanged.

One new transparent 1374×1145 whole-scene master contains three irregular groups
of low limestone slabs and small partly submerged companions. Narrow local
shallows/ripples stay with the rocks. All broad frame water is transparent;
the shared water material supplies the actual tile and surrounding water.
Uniform inset 0.85 gives scale `68/1145`, offset (7.2,6), with no stretching.
The complete source is resampled into the 4× scene using the existing signed,
premultiplied Lanczos3 affine implementation. Its opaque-art preflight rejects
opaque samples within eight world pixels of the outer polygon. The guard clears
alpha through four pixels and fades over 4–8. No opaque rocks are erased to fit.

The 2× artwork is a whole-scene Lanczos3 reduction of 4×. Actual ground PNGs are
assembled at their tile origins, artwork is composed over them, then final
complementary half-open ownership produces 192×160 / 384×320 runtime PNGs.
No filtering follows ownership. Review scenes decode saved PNGs and use ordinary
source-over drawing rather than repainting the assembled view.

Single-cell absolute scene origins are (0,0) and (71,40) for line parities 0
and 1; surrounding-patch canvas origins are (-71,-80) and (0,-40). The six
neighbors use the same true `water` base at offsets (71,-40), (71,40), (0,80),
(-71,40), (-71,-40), (0,-80). Tests distinguish the correct absolute water
phase from the incorrect quad-layout translation, even where normalized patch
PNGs could hide that error.

Eight further context images retain actual seven-input PNG assemblies: two
densities × two parities × homogeneous stones or compatible mixed shore.
The shore case uses the validated `lakeside12` fixture and replaces one actual
water cell with `waterStones`; all pixels outside that cell remain unchanged.
Land, bank and other water cells are not repainted. This checks a water-facing
shore connection rather than inventing a new coast profile.

Dense coverage repeats the same three-group arrangement; shape/placement variants
are still needed for natural full maps. Shared-water texture repetition, fractional
zoom, native coverage, runtime integration and release acceptance remain open.
See the [gallery](water-stones-review.html), [delivery record](distinct-terrain-batch-17.md)
and [exact recipe](../../art/recipes/compound-water-stones-v1.json).
