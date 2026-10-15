# Distinct terrain batch 11 — new shoreline family

2026-10-15. Nine new summer shore candidates and a polygon-built open-water base
are retained at **2× and 4×**. They use the newly authored grass, water and earth
masters with new smooth coast geometry. Original-game pixels, traced shore
outlines and close-reconstruction masters are not inputs. Full-map and runtime
acceptance remain open.

## Topology and artwork

[The topology contract](shoreline-topology.md) records explicit land-corner states
for all nine stable `lakeside` IDs. Controller and independent audit verify 100/100
directed campaign endpoint checks. `lakeside23` requires effective land-side mask
234; literal filename interpretation would break its southwest contacts. The
unused `lakeside12` and four boundary corners of `123`/`126` remain explicitly
family-derived. Edge endpoints, rather than binary land/water ports, allow mixed
shore crossings to meet correctly.

Initial geometry passed edge checks but independent art review rejected sharp
shoulders on `23`, a flat cove wall on `3456`, and smaller knuckles on `56`/`126`.
The repaired renderer replaces the nearest-edge/interior blend with one smooth
quintic coast per profile. Straight approaches protect boundary crossings through
eight world pixels; collinear handles give zero curvature at the interior joins.
The broad interior variation also vanishes through its second derivative at each
join. A finely flattened curve supplies signed distance for the bank shading.

Independent actual 2×/4× re-review provisionally passes all reported repairs, both
parities and regression appearance. No new clipping, direction mismatch, visible
hex seams, halos or coarse pixelation. `56` is now a long, nearly straight bank;
later shape variety is a minor refinement. Grass/water material repetition remains
the existing wider quality gate, and the narrow earth bank is fairly uniform.
The rejected first bundle remains private; the retained exports come from the
repaired `shore-terrain-v1-shaped` build.

## Export and provenance

The bounded `-shore` mode requires the exact nine known corner profiles and three
SHA-pinned new materials. Strict canonical schema/path validation rejects unknown
or case-alias fields, duplicates, wrong types, stale pins, unsafe paths and original
or pilot sources. Manifest and material snapshots retain the exact bytes loaded.
Fresh output and failure rollback preserve existing files.

The grass/water/earth generation chains remain tracked in the existing terrain
recipes; no new source painting was needed. The terrain pack remains 27 masters
and 54 rectangle exports. Shared mirror-periodic material sampling preserves the
runtime lattice phase. Grass is byte-identical to its previous polygon export;
water uses the same adapter with its own authored material.

Mixed edges use signed projection from the edge midpoint toward the land endpoint.
Pure land/water edge guards use the corresponding base. Interior curves vary while
the crossing, tangent, bank width and material phase remain reciprocal. Restrained
shallow water and wet earth blend into meadow, without a white outline. Unmasked
4× rendering includes a four-world-pixel margin; whole-image Lanczos3 reduction
precedes 2× cropping and final complementary polygon ownership. Final tile sizes
are 192×160 / 384×320, with zero RGBA outside the new polygon.

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -shore art/shore-terrain-v1.json \
  -out build/art-review/shore-terrain-v1-shaped
```

Use a fresh directory. The bundle has 63 generated PNGs (22 tiles, 40 patch views,
one contact sheet), three material snapshots, manifest, report and gallery: 69 files.
Public retention is 20 runtime PNGs and 40 review PNGs. The existing grass files
are reused. [The gallery](shoreline-review.html) displays ten candidates and both
line parities; [the recipe](../../art/recipes/shore-terrain-v1.json) pins outputs,
profiles, material provenance and all patch cells.

## Verification and remaining work

Controller independently verifies all 11 fixed profiles and 22 saved tile PNGs:
1,249,600 opaque owned pixels and 440,000 zero-RGBA exterior pixels. Forty actual
PNG patches reassemble byte-for-byte, with 240 interior contacts including ring
joins and 120 consistent three-cell vertices. Independent enumeration covers 234
compatible ordered profile/direction pairs and 1,638 common-edge shader samples.

The separate edge-color oracle checks 12,276 actual 2× pixels and 49,016 actual
4× pixels, including 533 / 2,153 mixed-transition centers. It computes the analytic
protected bank/material profile independently and applies its own unmasked Lanczos
reference for 2×. Maximum channel error is one byte, within the declared one-byte
tolerance. It proves the sampled protected strips, not interior appearance or
whole-map/runtime acceptance; those need the separate art review and later gates.

Independent code review passes with no blocking findings. All 69 repeat files match;
all 769 earlier infrastructure/compound artifacts reproduce byte-for-byte in fresh
current-code builds. All source/input/output hashes, dimensions, pinned profiles,
scene cells and 134 runtime ID/path/campaign bindings pass controller checks.
Static image/link checks pass for all six retained galleries; the browser loads
all 40 shoreline gallery images with continuous rendering and no horizontal overflow.

Focused RED/GREEN tests reproduce the rejected shoulder and flat-cove defects and
pass on the repaired curves. Final `just ci` passes: zero lint issues, art-review
tests in 236.893 seconds, all remaining Go tests, desktop/WASM builds, seven inventory
tests and current inventory of 273 paths / 246 loader calls.

The ledger now has 13 source + 112 built = 125 associated IDs, nine backgrounds
still needing images and zero accepted runtime replacements. Water advances from
source-only to built; `waterStones` remains source-only. Next: eight seasonal
mountain pieces and fog, then geometry/material integration for the source-only
candidates. Keep the separate 350 native-resource pairs, other graphics categories
and complete runtime/full-map acceptance in scope. Continue Phase 2 after finishing
the complete graphics set.
