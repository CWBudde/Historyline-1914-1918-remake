# Distinct terrain production

The latest user brief requires new continuous high-resolution images for **every
tile**, excluding originals and close reconstructions from shipping. This supersedes
the earlier 90% original-likeness target. Keep familiar gameplay semantics and
connection topology; invent surface artwork, palette and detail.

Current pack: [terrain-distinct-v1.json](../../art/terrain-distinct-v1.json).
[Gallery](distinct-terrain-review.html) · [Batch 01 review](distinct-terrain-batch-01.md) ·
[Batch 02 review](distinct-terrain-batch-02.md) · [Batch 03 review](distinct-terrain-batch-03.md) ·
[Connected family review](connected-family-batch-01.md) ·
[Runtime coverage](runtime-tile-coverage.md).

## Delivery and coverage

Nineteen semantic/material candidates are delivered: meadow, dry meadow, sparse/dense forest,
unmarked road bend, diagonal river, open water, bare earth, snow, rocky ground,
hills, mountain, wet mud, water stones, winter forest/rough/stone/shelled ground,
and unmarked gravel. Selected sources are 1374×1145 with 38 reproducible Lanczos3
exports at 192×160 / 384×320. Independent review prompted dense forest 02,
hills 02 and mountain 02. Mountain remains provisional pending camera/material
assessment in an assembled terrain/unit preview.

[runtime-tile-coverage.csv](runtime-tile-coverage.csv) explicitly lists all 134 enabled
remake tile IDs (60 backgrounds, 74 infrastructure). Twelve backgrounds have candidate
source associations; grass and all 16 roads / 11 streams have built candidates.
That is 40 associated IDs, 94 still needing images, and zero accepted runtime replacements.
Wet mud has its own reviewed source. Opaque road/river studies remain separate from the
new transparent infrastructure overlays.

The source-pack exports are rectangular overscan images. The separate
[connected manifest](../../art/connected-terrain-v1.json) builds a grass base and
27 transparent overlays with shared world phase and explicit ports. Actual saved PNG
patches pass pixel-aligned 2×/4× seam checks at both parities; independent visual review
provisionally passes. Full-map repetition, tree scale, other families, seasons,
runtime integration and release acceptance remain open.

[terrain-production.csv](terrain-production.csv) retains the existing **350 native
library/resource pairs** (175 summer, 175 winter) as an inventory and historical
pilot record. Its old style-approved/candidate statuses describe the earlier studies,
which are excluded from the new manifest. Do not count those statuses as new-pack
acceptance. Keep the original reference hashes for private ID mapping; they are
neither source artwork nor release dependencies for the new exporter.

| Family | Summer IDs | Winter IDs |
| --- | ---: | ---: |
| SPLAI | 8 | 8 |
| SFORE | 9 | 9 |
| SCOAS | 23 | 23 |
| SSSEA | 1 | 1 |
| SMOUN | 20 | 20 |
| SBUIL | 24 | 24 |
| SSTRE | 60 | 60 |
| SRAIL | 20 | 20 |
| SWALL | 9 | 9 |
| VERBO | 1 | 1 |

Prefixes are inventory groups, not sufficient semantic mappings. The new pack's
semantic IDs must be mapped explicitly to all enabled remake/native resources.
Map every alias and required connection; do not mark a family complete based on one
representative image or a generated master merely resembling its name.

## Work order

1. Extend the reviewed road/stream compositor to paths, rail/trenches and bridge/crossing
   components using [verified runtime geometry](runtime-connection-geometry.md) and new
   materials. Continue houses, shores and remaining seasonal/building/mountain pieces. Retain camera,
   unit readability and material-quality caveats during integration.
2. Establish shared forest scale and assemble larger mixed-family previews using the
   delivered road/stream variants. Correct transitions and inspect three-tile vertices
   before accepting a family; small same-family patches alone do not close integration.
3. Expand all required coast, road/river/junction, bridge, rail, wall and multi-tile
   building/mountain combinations. Preserve topology without tracing original artwork.
4. Complete winter variants, remaining terrain, infrastructure and overlays; map
   every required ID. Periodically use an independent subagent art review.
5. Verify all tile images and their assembled joins. Complete the remaining sprite/
   portrait/UI scope, then continue Phase 2 as requested. Release requires a complete
   mapped asset set containing only the selected newly authored pack.

Terrain authoring continues while pilot and runtime gates are open. Do not let the
retired original-pixel alignment work delay new images. Do not infer seam acceptance
from a full opaque rectangle or matching metadata.

## Export

`go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-directory>`
validates source hashes and builds a portable gallery/report plus 2×/4× rectangles.
It uses only declared new source images and fixed alpha-aware Lanczos3:
centered exact 6:5 crop → 384×320 → 192×160. Current masters need no crop.
Preserve generated source pixels and exact prompts. No original footprint or
reference raster participates in this export mode.

`go run ./tools/art-review -connected art/connected-terrain-v1.json -out <fresh-directory>`
builds the connected base/overlays, actual-PNG patches and gallery. See the
[batch recipe and verification record](connected-family-batch-01.md) for protected
profiles, final-pixel ownership, 198-file repeat checks and zoom limitations.

Legacy `-reference/-candidate/-resampling lanczos` comparisons remain available
for historical research; their 90/100 fidelity gate does not govern the new pack.
