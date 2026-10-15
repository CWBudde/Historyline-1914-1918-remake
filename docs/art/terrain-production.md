# Distinct terrain production

The latest user brief requires new continuous high-resolution images for **every
tile**, excluding originals and close reconstructions from shipping. This supersedes
the earlier 90% original-likeness target. Keep familiar gameplay semantics and
connection topology; invent surface artwork, palette and detail.

Current pack: [terrain-distinct-v1.json](../../art/terrain-distinct-v1.json).
[Gallery](distinct-terrain-review.html) · [Batch 01 review](distinct-terrain-batch-01.md) ·
[Batch 02 review](distinct-terrain-batch-02.md) · [Batch 03 review](distinct-terrain-batch-03.md) ·
[Batch 04 review](distinct-terrain-batch-04.md) ·
[Batch 05 review](distinct-terrain-batch-05.md) ·
[Batch 06 review](distinct-terrain-batch-06.md) ·
[Connected family review](connected-family-batch-01.md) ·
[Runtime coverage](runtime-tile-coverage.md).

## Delivery and coverage

Twenty-seven semantic/material candidates are delivered: meadow, dry meadow, sparse/dense forest,
unmarked road bend, diagonal river, open water, bare earth, snow, rocky ground,
hills, mountain, wet mud, water stones, winter forest/rough/stone/shelled ground,
unmarked gravel, path earth, summer/winter hamlets, winter mountain and new
ballast/timber/steel/trench-earth materials.
Selected sources are 1374×1145 with 54 reproducible Lanczos3
exports at 192×160 / 384×320. Independent review prompted dense forest 02,
hills 02 and mountain 02. Mountain remains provisional pending camera/material
assessment in an assembled terrain/unit preview.

[runtime-tile-coverage.csv](runtime-tile-coverage.csv) explicitly lists all 134 enabled
remake tile IDs (60 backgrounds, 74 infrastructure). Fourteen backgrounds have candidate
source associations; both ground seasons, eight depot pieces, fourteen HQ pieces, four factory pieces
and all 74 enabled infrastructure IDs have built candidates:
16 roads / 11 streams / 14 paths / 12 rails / 12 trenches / five bridges / four crossings.
That is 116 associated IDs, 18 still needing images, and zero accepted runtime replacements.
Wet mud has its own reviewed source. Opaque road/river studies remain separate from the
new transparent infrastructure overlays.

The source-pack exports are rectangular overscan images. The separate
[road/stream manifest](../../art/connected-terrain-v1.json) and
[path manifest](../../art/connected-paths-v1.json) and
[rail manifest](../../art/connected-rail-v1.json) and
[trench manifest](../../art/connected-trench-v1.json), plus the
[mixed manifest](../../art/connected-mixed-v1.json), build a grass base and
74 transparent overlays with shared world phase and explicit component semantics. Actual saved PNG
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

1. Extend the reviewed connected compositor to bridge/crossing
   components using [verified runtime geometry](runtime-connection-geometry.md) and new
   materials. Build/slice the five shared compound scenes for 26 building IDs using
   [numeric crop geometry](compound-building-geometry.md), then shores and seasonal/mountain pieces. Retain camera,
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
The path manifest uses the same command; [batch 04](distinct-terrain-batch-04.md)
records its 108-file builds and the new seasonal/hamlet sources.
The rail manifest also uses this command; [batch 05](distinct-terrain-batch-05.md)
records rail profiles, sleeper phase and the materials for subsequent trench work.
The [trench manifest](../../art/connected-trench-v1.json) uses the same command;
[batch 06](distinct-terrain-batch-06.md) records its distinct rooms, caps and 95-file build.

`go run ./tools/art-review -mixed art/connected-mixed-v1.json -out <fresh-directory>`
builds the nine bridge/crossing variants from three SHA-pinned connected manifests.
Finite road-bridge structures require the campaign's separate upper road/path/rail;
footbridges include path approaches. Level crossings preserve the component contacts
and steel gauge with fitted timber infill. [Batch 07](distinct-terrain-batch-07.md)
records the construction-detail repair, layer-order fixtures and 141-file bundle.
Every enabled infrastructure ID now has provisional 2×/4× artwork; this does not
close seasonal/full-map quality, runtime acceptance or native-resource mapping.

`go run ./tools/art-review -compound art/compound-depot-v1.json -out <fresh-directory>`
exports the two new depot scene masters into eight background pieces, plus shared
winter ground. The terrain source pack remains 27 masters / 54 rectangle exports;
these two transparent compound masters have their own manifest and provenance.
The whole artwork is sampled with uniform affine Lanczos3, then sliced after global
composition over actual exported ground cells. Fixed outer guard and explicit
crop origins preserve internal roofs and perimeter phase. [Batch 08](distinct-terrain-batch-08.md)
records independent scene/parity assembly, exact quiet-boundary samples and the
resource-budget repair. Browse the [compound gallery](compound-building-review.html).

The same command with `art/compound-hq-v1.json` builds two new HQ masters into
fourteen pieces. [Batch 09](distinct-terrain-batch-09.md) and the
[HQ gallery](compound-hq-review.html) record independent seasonal review and exact
seven-cell scene/parity assembly. HQ's manor sits in lower N with its entrance
in C; preserve the gameplay N ID and review its selection affordance in runtime.

The same command with `art/compound-factory-v1.json` exports the winter industrial
scene into four pieces, retaining E as the gameplay loading wing. [Batch 10](distinct-terrain-batch-10.md)
and the [factory gallery](compound-factory-review.html) record its uniform-framing
adjustment, actual PNG assembly and independent art review. Five transparent masters
now supply all 26 enabled compound-building IDs; full-map/runtime acceptance remains open.

Legacy `-reference/-candidate/-resampling lanczos` comparisons remain available
for historical research; their 90/100 fidelity gate does not govern the new pack.
