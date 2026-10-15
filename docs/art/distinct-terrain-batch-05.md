# Distinct terrain batch 05: rail and trench materials

2026-10-15. [Source gallery](distinct-terrain-review.html) ·
[Connected gallery](connected-family-review.html) ·
[Rail manifest](../../art/connected-rail-v1.json) ·
[Recipe and hashes](../../art/recipes/connected-rail-v1.json).

Delivered all 12 loaded rail variants as transparent 192×160 / 384×320 candidates.
Four new 1374×1145 material masters provide ballast, weathered timber, worn steel
and trench-floor earth. The source pack now contains 27 selected masters and 54
Lanczos3 rectangle exports. All generation prompts, source/output hashes and
unchanged masters are retained under `art/`; no original raster or reconstruction
study is a source. Trench earth is material delivery only, not a completed trench.

## Track geometry

The shared builder adds a rail family and requires exactly its three additional
materials: `rail`, `timber`, `steel`. The chosen ballast core is 11 world pixels
wide with a one-pixel earth verge. Steel gauge is 4.5, rail width 0.8; sleepers are
7.5×1.05 with nominal pitch 4. Straight protected guards extend eight world pixels
inward from each port. Explicit sleepers sit at port distance 0/4/8; interior
arc-length spacing fits a whole number of intervals between guards. Reciprocal
half-sleepers therefore meet without duplicate or missing ties.

Separate offset rail paths follow bends; rails are not contours of merged ballast.
Turnouts keep their complete north/south spine and add a tangential branch. The
report/recipe pins cubic controls, subdivision count and every profile/phase value.
Lanczos3 overscan, final half-open polygon ownership and actual saved-PNG scene
assembly follow the earlier connected builder. Existing road/stream/path rendering
and reports remain unchanged.

## Review and checks

Independent visual review provisionally passes the sources and all rail variants.
Paired rails, sleepers and smooth bends remain readable at 2×/4×; inspected patches
have no visible boundary gap, clipped continuation or doubled end sleeper. Turnout
frogs have dense sleeper overlap and a compressed merge, acceptable provisionally
at map scale. Detailed railway engineering, full maps, actual unit readability,
seasons, fractional zoom and runtime acceptance remain open.

Code review caught a real validation defect: a declared narrow ballast bed could
clip the fixed-width rails/sleepers. Rail core widths below eight world pixels now
fail validation. Tests reject 2/7.5 and accept 8/11; independent re-review closes the
finding. Authored width 11 outputs are unchanged by that fix.

Controller checks cover 12 variants, 54 active-port/density pairs and 3,528 opaque
core seam pixels: zero deficient pixels, zero double owners, minimum alpha 255.
All 48 actual-PNG parity/density patches match independent source-over assembly.
Tests additionally check literal material colors, paired gauge, sleeper phase and
turnout spine/branch pixels. Independent code review has no remaining findings.
Repeated rail builds match all 96 files; repeated source-pack builds match all 83.
Controller verifies all 27 source/recipe/input hashes, 54 material exports, 87 rail
output hashes/dimensions, 24 retained overlays and 17 review images. All 198 earlier
road/stream and 108 path artifacts remain byte-identical. The 134-row runtime ledger
matches loader IDs/paths and campaign-use flags. Repository `just ci` passes.
Browser checks load all 120 connected-gallery and 81 source-gallery images, with
no broken images or horizontal overflow at the checked desktop viewport.

## Reproduce and remaining scope

```sh
go run ./tools/art-review -pack art/terrain-distinct-v1.json -out <fresh-material-dir>
go run ./tools/art-review -connected art/connected-rail-v1.json -out <fresh-rail-dir>
```

The 96-file rail build includes seven retained sources, two grass bases, 24 overlays,
12 grass composites, 48 patches, contact sheet, report and index. Tracked overlays
are in `art/exports/rail-v1/`, with selected review PNGs in `art/review/rail-v1/`.
The full local gallery stays under ignored build/. The source-pack build has 83 files.

Coverage: **15 background source candidates + 54 built candidates = 69 associated
IDs; 65 still need images; zero accepted runtime replacements**. The remaining 21
infrastructure IDs are trenches, bridges and crossings. The trench audit corrects
the earlier assumption that every branch touching an edge is a through-port; see
[geometry evidence](runtime-connection-geometry.md). Its new earth material is ready
for independent trench courses with enclosed caps and chambers. Native 350-resource
mapping, buildings/shores/mountain pieces, seasons, sprites and UI remain open.
Finish the graphics scope before continuing Phase 2 as requested.
