# Compound tile geometry for new artwork

2026-10-15. Numeric production prerequisite, separate from original image pixels.
Five newly authored whole scenes can supply the 26 enabled compound background
images: summer/winter HQ, summer/winter depot, and winter factory. Generate each
scene once and slice its pieces; independent per-piece generation risks split roofs
and discontinuous yards. The depot pair is now delivered as provisional candidates;
HQ and factory scene artwork remain open.

The runtime places 96×80 tiles at `(142*column+71*(line%2),40*line)`. Every crop uses
the new polygon `(25,0),(71,0),(96,40),(71,80),(25,80),(0,40)` and complementary
final-pixel ownership, matching the connected terrain builder.

| Runtime pieces | Crop origin within scene | Crop center |
| --- | --- | --- |
| hqN, hqN_w | 71,0 | 119,40 |
| hqNW, hqNW_w | 0,40 | 48,80 |
| hqNE, hqNE_w | 142,40 | 190,80 |
| hqC, hqC_w | 71,80 | 119,120 |
| hqSW, hqSW_w | 0,120 | 48,160 |
| hqSE, hqSE_w | 142,120 | 190,160 |
| hqS, hqS_w | 71,160 | 119,200 |
| cmpN, cmpN_w, factoryN_w | 71,0 | 119,40 |
| cmpW, cmpW_w, factoryW_w | 0,40 | 48,80 |
| cmpE, cmpE_w, factoryE_w | 142,40 | 190,80 |
| cmpS, cmpS_w, factoryS_w | 71,80 | 119,120 |

HQ is a seven-cell rosette bounded by 238×240 world pixels, exporting a
476×480 / 952×960 scene. Depot/factory is a four-cell diamond bounded by 238×160,
exporting 476×320 / 952×640; it is not a rectangular 2×2 arrangement. Each crop
exports at 192×160 / 384×320. Seasons share exact placements, without rotation.

The quad union outline is `(96,0),(142,0),(167,40),(213,40),(238,80),(213,120),
(167,120),(142,160),(96,160),(71,120),(25,120),(0,80),(25,40),(71,40)`.
Major roofs, yards and shadows must stay inside this outline, not merely inside
the source rectangle. A numeric authoring guide and actual masked preview should
check that before source promotion. Matching transparent source overlays on shared
new ground can keep the compound perimeter compatible with surrounding terrain.

## Evidence and next implementation

Independent audit `compound_tile_layout_audit` checked loader IDs, all three shipped
campaigns and both row parities. Controller ran its Go probe in ignored
`build/gfx-research/_compound-geometry/`: every expected group is present. Level 1
HQ center `(line7,column0)` has scene origin `(0,200)` and the seven offsets above.
Its depot north `(18,5)` has scene origin `(639,720)`. Level 2 winter factory north
`(6,9)` has scene origin `(1207,240)`. The backup level repeats level 1 and is not an
additional campaign. All 26 loader images decode as 96×80.

Campaign group counts: summer HQ/depot 2/2; winter HQ/depot/factory 4/4/9. HQ and
depot gameplay building cells use the N piece, whereas factory uses E; retain this
distinction for visual entrances. Geometry proves placement, not final building
design or legacy alpha ownership. No reference raster is a new artwork input.

The shared-scene exporter now builds the new depot pair from
[`art/compound-depot-v1.json`](../../art/compound-depot-v1.json). Selected summer02
and winter02 masters, exact generation prompts, new study01 and numeric guide inputs are
retained under `art/source/compound/` and `art/recipes/compound/`. Study01 is only a
generation input; it is excluded from the runtime manifest. Rejected winter01 remains private.

Both selected 1536×1024 masters use uniform inset0.9 framing into 238×160:
`x=sourceX*0.139453125+11.9`, `y=sourceY*0.139453125+8.6`.
The actual Lanczos preflight finds the nearest alpha≥250 pixels 9.625 / 9.125 world
pixels inside the outer boundary. An outer-only guard is zero through4 and fades
smoothly to full alpha at8; opaque architecture inside that band is rejected before
export. No internal tile edge is feathered. The guard discards only low-alpha fringe
at the footprint boundary (50 summer /4 winter alpha1 pixels), without clipping roofs.

Ground cells are exported/decoded before assembling the common scene, preserving
the lattice phase; source art is sampled once over the complete scene. Crops at both
densities reassemble byte-for-byte with ordinary PNG source-over, and surrounding
ground matches at both line parities. See [batch08](distinct-terrain-batch-08.md).
The same bounded tool supports the remaining fixed HQ/factory geometry; synthetic
layout tests do not count as delivered artwork. A new private HQ study and numeric
guide prepare the next pair but still need framing, seasonal and visual review.
