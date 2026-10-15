# Distinct terrain batch 04 and connected paths

2026-10-15. [Source gallery](distinct-terrain-review.html) ·
[Connected gallery](connected-family-review.html) ·
[Source manifest](../../art/terrain-distinct-v1.json) ·
[Path manifest](../../art/connected-paths-v1.json).

Delivered four new 1374×1145 masters: fine compacted path earth, a three-cottage
hamlet, its matching winter version, and winter rocky mountain terrain. The pack
now has 23 selected masters and 46 Lanczos3 192×160 / 384×320 rectangle exports.
Generated images and exact prompts/input hashes are retained under `art/`.
No original-game image or close-reconstruction study is a generation/export input.

The first hamlet showed excessive wall faces. Retained `hamlet.png` is rejected
and excluded from the manifest; selected `hamlet-02.png` corrects the overhead
camera. Winter preserves its three roofs, courtyard, well and garden layout.
The mountain winter edit preserves the new summer source's branching rocky ridges.
Existing mountain camera/material caveats remain provisional; a winter edit does
not close the summer terrain review.

## Connected path delivery

All 14 loaded path IDs now have transparent 2×/4× candidates, separate from their
ground. The shared Go builder adds a path family and an optional `path` material
required exactly when that family exists. Existing road/stream declarations remain
valid. Paths use a nominal seven-world-pixel earth core and 0.75-pixel earth verge,
versus fourteen-pixel gravel roads. No lane paint or original footprint is used.

Independent art review rejected the initial sharp, uniform tan ribbons. The final
path-only interior has restrained course/width variation and edge wear, rounded
junctions and translucent earth fringes revealing the ground beneath. The complete
eight-world-pixel connection zone remains unchanged. Exact equations and constants
are pinned in the report and [path recipe](../../art/recipes/connected-paths-v1.json).
Sampling, Lanczos3 overscan and final complementary polygon ownership follow
[connected batch 01](connected-family-batch-01.md). Road/stream arithmetic is unchanged.

Independent re-review provisionally passes the corrected paths and all new subjects.
Remaining polish: paths still have tidy edges and long straight runs; grass fringes
and approaches to the hamlet need a mixed full-map preview. Buildings appear inside
the intended hex, but actual masked placement and winter vegetation clipping remain
unverified. Source associations are not accepted runtime bindings.

## Verification

Controller seam checks cover all 14 variants, 66 active-port/density pairs and
1,868 core seam pixels: zero deficient pixels, zero double owners, minimum alpha 255.
All 56 actual saved-PNG parity/density patches match independent source-over assembly.
Tests cover literal path material, exact width in the protected zone, organic
interiors, filtered/raw guards, lattice phase and strict material validation. Code
review found no material issues. Partial-alpha RGB comparisons allow one byte for
Go source-over conversion; alpha and opaque colors remain exact.

Repeated path builds match all 108 files; source-pack builds match all 71 files.
Controller checks verify all 23 source/recipe/input hashes, 46 material exports,
101 path PNG output hashes/dimensions, 28 retained overlays and 19 review copies.
All 198 historical road/stream artifacts remain byte-identical after the extension.
The 134-row ledger matches loader IDs/paths and campaign flags. Repository `just ci`
passes. Browser checks load all 92 connected-gallery and 69 source-gallery images
with no broken images or horizontal overflow at the checked desktop viewport.
Candidate appearance still needs runtime validation.

## Reproduce and remaining scope

```sh
go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-material-dir>
go run ./tools/art-review -connected art/connected-paths-v1.json -out <fresh-path-dir>
```

The 108-file path build includes five retained material sources, two grass bases,
28 overlays, 14 composites, 56 patches, contact sheet, report and index. Runtime
candidate overlays are retained under `art/exports/paths-v1/`; selected review PNGs
under `art/review/paths-v1/`. Other material inputs remain required for compatibility
with the existing connected builder, even in a path-only manifest.

Coverage is now 15 background source candidates and 42 built candidates (grass and
41 infrastructure overlays): **57 associated IDs, 77 needing images, zero accepted
runtime replacements**. Thirty-three infrastructure IDs remain: rail, trenches,
bridges and crossings. Buildings/shore/mountain pieces, seasons, ownership art,
350-native-resource mapping and the remaining sprite/UI scope are still open.
Complete the graphics scope, then continue Phase 2 as requested.
