# Distinct terrain batch 12 — overhead seasonal mountains

2026-10-15. Two new continuous mountain masters supply eight seasonal quarters
and two standalone tiles at **2× and 4×**. All are provisional graphics candidates.
Original pixels, traced mountain silhouettes and close-reconstruction studies are
excluded from generation and exports.

## Artwork and geometry

The first summer source and its camera edit failed independent review: upright
peaks and repeated frontal cliffs read as an oblique miniature. Both remain
private and are excluded from the selected generation chain. Summer revision 03
was generated afresh using only the new rocky-ground material as a palette and
texture reference. Its planimetric branching crests have slopes visible around
the ridge junctions. Winter revision 01 edits that selected master, preserving
the overhead camera, ridge network and footprint while adding snow to the slopes
and valleys. The exact prompts, input hashes and actual 1536×1024 dimensions are
retained in the source recipes. Generation used the built-in image generator.

Independent review provisionally passes both sources, all ten tiles, four complete
assemblies and their ground surrounds at both densities and line parities. No
remaining important camera, join, halo or framing defect was observed. Repeated
radial ridge motifs, the compact footprint and subdued summer-single readability
remain full-map/unit-overlay quality caveats. W/E quarters contain mostly ground;
these candidates are not runtime or release acceptance.

[The geometry audit](mountain-tile-geometry.md) accounts for all 49 campaign
quarter placements: twelve complete groups and one interior orphan in level 3.
Full scenes use N/W/E/S origins `(71,0)/(0,40)/(142,40)/(71,80)` in 238×160;
single scenes use C `(0,0)` in 96×80. Stable IDs are `mountN/W/E/S` with seasonal
suffixes, plus `mountain` and `mountain_w`. Terrain has no building-entry cell.

The level-3 orphan at zero-based line21/column6 is `mountW_w`, surrounded by grass
where the other three quarters would be. The new winter standalone tile supplies
a proposed conditional display fallback. Two retained PNG previews assemble eight
actual nearby campaign cells, including road14/road124 overlays. Independent art
review finds no visible cut or ground seam and confirms clearance from the road.
The recipe records this selection and every input hash. **Runtime fallback
selection remains unimplemented**; campaign data, IDs and gameplay rules are unchanged.

## Reproducible export

The bounded compound exporter adds four exact scene IDs: `mountain-summer`,
`mountain-winter`, `mountain-single-summer`, `mountain-single-winter`. The same two
masters are referenced by both layouts, without independent tile painting. The
strict manifest now permits up to nine known scenes; unsupported names and
duplicates still reject. Existing building layouts and profiles are unchanged.

One uniform affine transform fits each complete master. Opaque-art preflight
rejects unsafe outer framing rather than erasing solid ridges. Insets 0.9, 0.7
and 0.6 failed on the summer quad with 23,880 / 1,829 / 61 protected opaque samples;
0.55 passes for both layouts and seasons. The outer guard fades only the perimeter,
never internal quarter cuts. Exported seasonal ground supplies the shared phase.
Whole-scene Lanczos rendering/reduction precedes final complementary ownership.
Tiles are 192×160 and 384×320; quad assemblies are 476×320 and 952×640.

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -compound art/compound-mountain-v1.json \
  -out build/art-review/compound-mountain-v1-final
```

Use a fresh output directory. The build contains 64 generated PNGs, six source
snapshots, manifest, report and gallery: 73 files. Public retention is 20 runtime,
24 assembly/surround and two conditional-context PNGs. The existing ground exports
are reused. [The gallery](compound-mountain-review.html) shows all 46 retained
images; [the recipe](../../art/recipes/compound-mountain-v1.json) pins sources,
geometry, transforms, outputs and the conditional preview. The older 27-source
terrain study pack remains historical; the two standalone runtime bindings now
point to these new polygon-built mountains.

## Verification and remaining work

An independent actual-PNG verifier checks eight full scenes and sixteen surrounds
by ordinary source-over assembly, including absolute runtime lattice coordinates.
It counts 1,136,000 opaque owned pixels, 400,000 zero-RGBA exterior pixels,
73,432 exact quiet-boundary pixels and 6,615 artwork samples crossing internal cuts.
All match. Initial single-layout parity metadata inherited a quad-specific offset;
independent verification caught this. Formula-derived tests failed before repair
and pass with singles anchored at `(0,0)/(71,40)`. Old quad paths are unchanged.

Independent final code review finds no remaining blocking issue. Focused tests,
vet and lint pass. Fresh current-code builds reproduce all 838 earlier family
artifacts byte-for-byte after the parity repair. All 73 repeat files match, with
no additional files. Every source/snapshot/output hash and actual dimension,
nested report reference and retained copy passes verification; all 64 generated
PNGs match the independently reviewed preflight pixels. A separate reconstruction
of the two conditional-context previews matches eight actual campaign cells.
All 134 loader ID/path/campaign-use bindings pass. Eight static galleries have
417 images and valid local links; the browser loads all 46 new gallery images
with continuous rendering, ten tile cards and no horizontal overflow.
Final `just ci` passes: zero lint issues, art-review tests in 244.193 seconds,
remaining Go tests, desktop/WASM builds, seven inventory tests and the current
273 paths / 246 loader calls inventory.

The enabled ledger advances to **11 source + 122 built = 133 associated IDs**,
one background (`unseen` fog) needing an image, and zero accepted replacements.
Next deliver fog, then integrate geometry/materials for the eleven source-only
candidates. The separate 350 native-resource mapping obligation, units, UI and
other graphics, full maps, fractional zoom and runtime/release acceptance remain
open. Finish the graphics scope before continuing Phase 2, as requested.
