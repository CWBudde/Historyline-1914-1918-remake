# Distinct terrain batch 01

Delivered 2026-10-15. [View sources and 2×/4× exports](distinct-terrain-review.html).

## Current brief

The user's latest goal supersedes the earlier approximately 90% original-likeness
target: create distinct continuous high-resolution artwork for every tile, without
shipping originals or close reconstructions. Preserve gameplay semantics, identifiers
and connectivity. Finish graphics coverage before continuing Phase 2.

The new [manifest](../../art/terrain-distinct-v1.json) is separate from the historical
pilot recipes. No original image, original pixel mask, or old reconstruction master is
an input to this batch. Meadow was generated from a text prompt; the other subjects
use only that newly generated meadow as a material/style reference. Dense forest 02
also uses its newly generated first attempt as the edit target. Generation provenance
and prompts are retained in [recipes](../../art/recipes/terrain/distinct-v1/).
This is evidence about the inputs; it is not a legal or exhaustive visual-similarity
assessment.

## Assets and independent review

All selected masters are opaque 1374×1145 PNGs. Rectangular exports are 192×160 and
384×320, respectively 2×/4× the remake's 96×80 terrain bounds. These are larger than
the earlier 48×48 / 96×96 intrinsic Amiga diagnostic exports.

Independent reviewer: subagent `distinct_batch_art_review`, direct inspection of
all six initial masters and their 192×160 exports, both forest 384×320 exports,
followed by inspection of the corrected forest master. Controller also inspected
selected sources and corrected forest/road/river reductions.

| Asset | Source-art result | Remaining work |
| --- | --- | --- |
| meadow | Candidate; continuous grass and quiet values | Shared ground phase; avoid conspicuous repeats |
| meadow-dry | Candidate; clear dry-ground distinction | Material transition edges |
| forest-sparse | Provisional candidate; overhead broadleaf crowns | Shared tree/ground scale |
| forest-dense, revision 02 | Provisional candidate after correction | Smaller typical crowns than sparse forest; a few tapered conifers remain |
| road-bend | Candidate; road and both shoulders reach bottom/right frame edges; no white paint | Declared hex ports, paired roads and width/phase matching |
| river-diagonal | Candidate; water and both banks reach left/right frame edges | Declared hex ports, paired banks, width/phase matching |

Dense forest 01 failed because its conifers read as upright triangular side
silhouettes. A focused image edit replaced them with radial crowns. The reviewer
confirmed that revision 02 resolves the main camera defect sufficiently to proceed
to a shared-scale preview. The rejected first source remains privately under
`build/gfx-research/distinct-v1/forest-dense-01.png`, with its hash and prompt in
the selected forest recipe.

No candidate is claimed seamless. Reaching a rectangular canvas boundary does not
prove correct intersections with an actual hex cutout or an adjacent variant.
Sources preserve overscan so features can be clipped through rather than end in grass.

## Reproduce without original installations

From the repository root, choose a fresh output directory:

```sh
GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -pack art/terrain-distinct-v1.json \
  -out build/art-review/terrain-distinct-v1
```

If that directory already exists, choose a new name. The command refuses overwrites.
Pack mode accepts only `-pack` and `-out`; its Lanczos3 filter is fixed.

The exporter validates the manifest, unique safe IDs, PNG dimensions/opacity and
source hashes before creating outputs. Paths resolving into an `original/`
directory are rejected. It copies each source verbatim for a portable local gallery,
then crops the largest exact 6:5 integer rectangle, centered with floor offsets.
All six current sources already have exact 6:5 proportions, so no source pixels
are cropped away. The fixed chain is crop → 384×320 → 192×160, using the existing
alpha-aware Lanczos3 implementation at each reduction. No original footprint is
applied. Rectangle corners remain opaque; renderer clipping is a later step.

Each build contains 20 files: six verbatim source previews, twelve exports,
`report.json` and `index.html`. The report includes manifest/source/output SHA256,
source dimensions, crop bounds, fixed filter chain and candidate status. Source
and selected export hashes are also retained in the tracked recipes. Tracked
exports omit duplicate source previews; the tracked gallery links to canonical
masters under `art/source/`.

## Verified delivery and open coverage

The controller checked source/recipe/export hashes and dimensions, full opacity,
unchanged crop geometry, gallery links, and repeated builds. Two fresh builds
produced all 20 identical files. Focused Go tests cover off-aspect crop registration
with actual reduction, hash mismatch, transparent/undersized sources, original
paths and resolving symlinks, unsafe/case-duplicate IDs, flag conflicts,
non-overwrite behavior and reproducible reports. Repository `just ci` passed. A Playwright
browser check loaded all 18 gallery images at their expected dimensions with smooth
display scaling and no horizontal overflow at the tested desktop viewport.

This delivers **six semantic candidates**, not six accepted native-ID replacements.
The old 350-entry summer/winter table remains an inventory and historical pilot
record. The new sources still need explicit runtime/native-ID mappings, family
variants, winter coverage and seam acceptance. Tile production is far from complete;
units, portraits and UI also remain open.

Next: generate missing water/earth/stone/snow and shore/hill/mountain images, build
a shared-scale/port preview for road and river variants, then expand buildings,
bridges, rail, walls and seasonal families. Independently review subsequent batches.
