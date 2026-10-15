# Distinct terrain batch 08 — shared depot scenes

2026-10-15. Two newly authored transparent depot masters now supply all eight
enabled summer/winter depot pieces at **2× and 4×**. Shared winter ground is also
exported as a polygon-built candidate. No original or close-reconstruction image
is an input. These are provisional graphics candidates, not accepted runtime replacements.

## Artwork and provenance

The complete compound is authored once per season, preserving warehouses, roofs,
chimneys, canopy, courtyard and prop placement. Winter01 was rejected for layout
drift; a new weather edit with explicit coordinate constraints repairs it. Independent
source review approves the selected summer02/winter02 pair. Actual 2×/4× scene and
tile review also passes: buildings remain readable, seasons align, roofs continue
through cuts and no clipped corners, grid lines or boundary halos are visible.
Minor caveats: dense roof/courtyard detail, small props becoming texture at 2×,
and subtle repeated ground texture in larger patches.

Selected masters and exact prompts/input hashes are retained in
`art/source/compound/distinct-v1/` and `art/recipes/compound/distinct-v1/`.
The new study01 and numeric guide are retained solely as generation inputs to
summer02, outside the runtime manifest. Rejected winter01 remains in ignored `build/`.
The terrain material pack stays at 27 sources / 54 rectangle exports; the two
compound masters have their own manifest and generation chain.

## Geometry and export

The four-cell diamond is bounded by 238×160 world pixels, with crop origins
N(71,0), W(0,40), E(142,40), S(71,80). Runtime IDs are `cmpN/W/E/S` and their `_w`
counterparts; the north piece remains the gameplay building cell. Full scenes are
476×320 / 952×640 and individual tiles 192×160 / 384×320.

Both 1536×1024 masters use the same uniform inset 0.9 transform:
`worldX=sourceX*0.139453125+11.9`, `worldY=sourceY*0.139453125+8.6`.
Premultiplied antialiased Lanczos3 samples the complete artwork to 4× without
aspect stretch; whole-scene reduction makes the 2× artwork. Tile crops are never
filtered independently. The nearest alpha≥250 samples lie 9.625 / 9.125 world
pixels inside the outer footprint. A quiet outer guard is zero through 4 and fades
to full at 8. Opaque architecture within 8 is rejected before export; internal tile
edges are excluded from this guard. Only alpha 1 boundary fringe is discarded:
50 summer pixels and 4 winter pixels, with no guarded opaque pixels.

The ground is assembled from actual exported/decoded cell PNGs at fixed origins.
Directly sampling at the scene's bounding-box origin would shift its lattice phase.
Artwork is composited globally, then sliced with final complementary polygon
ownership. Saved tile PNGs are decoded and assembled with ordinary source-over;
surrounding ground patches use the real runtime line-parity origins.

The bounded `-compound` mode supports the five known depot/HQ/factory scene IDs.
Only the depot pair is delivered here; synthetic HQ/factory tests do not supply
their artwork. SHA-pinned input snapshots, strict schema/path checks and fresh
output rollback are included. Code review caught excessive kernel allocation for
extreme source aspect ratios; a pre-allocation budget of four million aggregate
axis taps now rejects such inputs. The sampling arithmetic is unchanged for these
masters. Scoped re-review approves the repair and exact all-cell ID/origin tests.

## Reproduction and evidence

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -compound art/compound-depot-v1.json \
  -out build/art-review/compound-depot-v1-reviewed
```

Use a fresh output directory. The bundle contains 40 generated PNGs, four retained
source PNGs, the manifest snapshot, report and gallery: 47 files. Public retention
is 18 runtime exports (eight pieces plus winter ground at both densities) and 12
review images (whole scenes and both surround parities at both densities). Hashes,
dimensions, transforms and guard metadata are pinned in
[`art/recipes/compound-depot-v1.json`](../../art/recipes/compound-depot-v1.json).

Controller independently assembled four full-scene PNGs and eight actual parity
PNGs byte-for-byte. Checks cover 908,800 opaque owned tile pixels, 320,000 clear
outside pixels, 51,584 exact ground pixels in the quiet perimeter and 7,367 artwork
samples across internal cuts. All source and retained-output hashes and dimensions
pass verification; all 47 repeat-build files match exactly. All 638 earlier-family
artifacts remain unchanged. The browser loads all 24 compound-gallery images with
continuous rendering and no horizontal overflow; the source and connected galleries
also pass static image/link checks (81 and 177 images).

Final `just ci` passes: lint reports zero issues; the art-review tests complete in
193.930 seconds; all remaining Go tests, desktop/WASM builds and seven inventory
tests pass. The current inventory reports 273 paths and 246 loader calls.

Current coverage is 14 source + 84 built candidates = 98 associated IDs, 36 backgrounds
needing images, zero accepted runtime replacements. Next author HQ summer/winter
and winter factory scenes (18 pieces), then shores, mountain pieces and fog. A new
private HQ study prepares that work; its framing/seasonal/visual review remains open.
All 350 native-resource pairs, full-map/runtime acceptance and the remaining units/UI
graphics remain in scope. Continue Phase 2 after finishing the complete graphics set.
