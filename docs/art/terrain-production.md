# Distinct terrain production

The latest user brief requires new continuous high-resolution images for **every
tile**, excluding originals and close reconstructions from shipping. This supersedes
the earlier 90% original-likeness target. Keep familiar gameplay semantics and
connection topology; invent surface artwork, palette and detail.

Current pack: [terrain-distinct-v1.json](../../art/terrain-distinct-v1.json).
[Gallery](distinct-terrain-review.html) · [Batch 01 review](distinct-terrain-batch-01.md).

## Delivery and coverage

Six semantic candidates are delivered: meadow, dry meadow, sparse/dense forest,
unmarked road bend and diagonal river. Selected sources are 1374×1145; reproducible
Lanczos3 exports are 192×160 / 384×320. Dense forest 02 replaces the first attempt
after independent review found side-view conifers.

These are rectangular overscan images. Shared texture phase, tree scale, runtime
hex clipping, six-edge ports, family variants, seasonal coverage and runtime mapping
remain open. Candidate delivery is not accepted complete tile coverage.

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

1. Generate missing new water, earth, stone and snow materials, shores and hills/
   mountains. Maintain calm readable values and consistent overhead light.
2. Establish shared forest scale and a small assembled connected preview using road
   and river straights/bends, declared ports and protected material/bank profiles.
   Correct adjoining edges and three-tile vertices before accepting a family.
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

Legacy `-reference/-candidate/-resampling lanczos` comparisons remain available
for historical research; their 90/100 fidelity gate does not govern the new pack.
