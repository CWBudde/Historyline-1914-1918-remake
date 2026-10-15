# Distinct terrain batch 03

2026-10-15. [Gallery](distinct-terrain-review.html) ·
[Runtime coverage](runtime-tile-coverage.md).

Six more new 1374×1145 sources bring the pack to **18 semantic candidates and
36 tracked 2×/4× exports**. No original image or raster mask is a generation/export
input. Recipes retain exact prompts, input/source/output hashes and rejected-attempt
provenance; selected sources are preserved unchanged.

| Subject | Runtime source candidate | Independent visual result |
| --- | --- | --- |
| Wet mud, revision 02 | mud | Larger restrained puddles now survive reduction |
| Water stones | waterStones | Low outcrops and submerged shelves; distinct from open water |
| Winter forest | forest_w | Radial bare crowns and overhead conifers; compatible snow |
| Winter rough ground | rough_w | Brown ground and scrub distinct from snow/stone |
| Winter stone ground | stoney_w | Exposed pale slabs remain visible at game size |
| Winter shelled ground, revision 02 | shelled_w | Varied broken/overlapping craters replace obvious four-bowl grid |

Reviewer `distinct_batch03_review` inspected all six full sources and 192×160
exports against new family references, then both corrected sources/reductions.
Initial mud lost wet reflections at small size; initial shelled ground repeated
four similarly sized round craters. Focused image-generator edits corrected those
issues. Both revisions passed source candidacy. Controller also inspected previews.
Unit overlays, clipping, shared texture phase and assembled seams remain untested.

Reproduce with `go run ./tools/art-review -pack art/terrain-distinct-v1.json -out
<fresh-directory>`. Fixed exact6:5 crop → 384×320 → 192×160 uses Lanczos3.
All selected masters are exact6:5 and retain their entire canvas. Two final builds
produce **56 byte-identical files**: 18 source copies, 36 exports, report and gallery.
Controller verification covers all source/input/recipe/export hashes, dimensions,
opacity, tracked bytes, crop geometry, gallery links and all 134 runtime ledger rows.
Repository `just ci` passes.

Coverage now records **13 candidate-source associations and 121 IDs needing images**,
zero accepted runtime replacements. The separate 350 native-resource inventory,
remaining buildings/mountains/seasonal pieces, all connected families, ownership
art, sprites and UI stay open. Historical batch01/02 counts describe their commits;
the active manifest builds all18 sources.

Next: shared port/path compositor and connected road/stream previews using new
materials; [runtime geometry evidence](runtime-connection-geometry.md) removes the
direction-numbering ambiguity. Then extend directional families and remaining
building/seasonal subjects. Full tile coverage and seams precede Phase2.
