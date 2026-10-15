# Distinct terrain batch 09 — headquarters scenes

2026-10-15. Two newly authored transparent HQ masters supply all fourteen enabled
summer/winter headquarters pieces at **2× and 4×**. A slate-roof command manor,
cupola, auxiliary wings and courtyard distinguish headquarters from the depot.
No original or close-reconstruction image is an input. These are provisional
graphics candidates; runtime and release acceptance remain open.

## Artwork and provenance

The summer master uses a new numeric seven-cell layout guide and the new depot
master as a camera/material reference. Winter is a weather edit of that summer
master, retaining the manor, cupola, four chimneys, entrances and service wings.
Both actual PNG masters are 1254×1254 RGBA. Earlier private metadata described
summer as 1280 square, and the winter prompt repeats that mistake; the source
records now give the verified dimensions and retain the exact prompt verbatim.
No image resizing or registration repair was needed.

Independent source and actual-export review provisionally passes both seasons at
2×/4×: readable architecture, stable roof positions and scale, continuous cuts,
and no visible clipped corners, grid lines, perimeter halos or new material
texture defect. The manor occupies lower N and its entrance extends into C.
Its northern-sector association is clear, but the N gameplay cell's center lies
on grass above the roof. Keep this as an explicit runtime usability review item.
Large-map ground repetition and fractional zoom remain broader open gates.

Selected masters, numeric guide and exact generation recipes/input hashes are
retained under `art/source/compound/distinct-v1/` and
`art/recipes/compound/distinct-v1/`. The existing depot input chain remains
tracked. The opaque terrain pack stays at 27 sources / 54 rectangle exports;
there are now four additional transparent compound masters in separate manifests.

## Geometry and export

The seven-cell rosette has world bounds 238×240. Crop origins are
N(71,0), NW(0,40), NE(142,40), C(71,80), SW(0,120), SE(142,120), S(71,160).
Runtime IDs retain `hqN/NW/NE/C/SW/SE/S` and their `_w` counterparts; N remains
the gameplay building cell. Full scenes are 476×480 / 952×960; individual tiles
are 192×160 / 384×320.

Both masters use uniform inset 0.75, scale `0.1423444976076555` and world offset
`(29.75,30.75)`. The existing whole-scene premultiplied Lanczos3 exporter supplies
both densities without independently filtering crops. The nearest alpha≥250
samples are 13.529 / 13.025 world pixels from the outer footprint. No opaque
architecture enters the eight-pixel guard band. Only one alpha-1 boundary pixel
per source is discarded. Internal edges are excluded from the perimeter guard.
Actual ground tile PNGs are assembled before composition, preserving lattice
phase, and final complementary polygon ownership makes ordinary PNG assembly exact.

## Reproduction and evidence

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -compound art/compound-hq-v1.json \
  -out build/art-review/compound-hq-v1-final
```

Use a fresh output directory. The bundle contains 52 generated PNGs, four retained
source snapshots, the manifest, report and gallery: 59 files. Public retention is
28 tile exports and 12 review PNGs. Ground outputs match the previously delivered
summer/winter ground; they are not duplicated in this batch's runtime recipe.
See [the gallery](compound-hq-review.html) and
[output recipe](../../art/recipes/compound-hq-v1.json).

Controller independently reconstructed four density scenes and eight actual parity
PNGs byte-for-byte. Checks cover 1,590,400 opaque owned pixels, 560,000 zero-RGBA
outside pixels, 66,616 exact quiet-boundary pixels and 18,686 alpha≥128 artwork
samples across all twelve internal cuts. Every cut contains meaningful artwork
at both densities. All 59 repeat-build files match exactly. All source/input and
retained-output hashes and dimensions pass; reviewed study pixels remain unchanged
in the final manifest build. Both ground exports match the earlier candidates.
Existing depot and infrastructure artifacts remain unchanged. Static image/link
checks pass for the 81-image source, 177-image infrastructure, 24-image depot and
34-image HQ galleries. Browser verification loads all 34 HQ images with continuous
rendering and no horizontal overflow.

Final `just ci` passes: zero lint issues, all Go tests (cached with no production
code changes), desktop and WASM builds, seven inventory tests, and current inventory
of 273 paths / 246 loader calls. The controller also checks every one of the
134 runtime ledger IDs, paths, campaign flags and candidate bindings.

Coverage is now 14 source + 98 built = 112 associated IDs, 22 backgrounds still
needing images and zero accepted runtime replacements. Next: winter factory's
four pieces, then nine shores, eight mountain pieces and fog. Native-resource
mapping, source-candidate geometry, other graphics categories and runtime/full-map
acceptance remain open. Continue Phase 2 after completing the full graphics set.
