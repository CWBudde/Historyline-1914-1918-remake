# Distinct terrain batch 10 — winter factory

2026-10-15. One newly authored transparent industrial scene supplies the four
enabled winter factory pieces at **2× and 4×**. This completes provisional scene
artwork for all 26 enabled depot/HQ/factory background IDs. Full graphics coverage
and runtime acceptance remain open. No original or close-reconstruction image is
an input.

## Artwork and provenance

The scene includes a northlight-roof workshop, northern boiler house and brick
stack, western coal/timber yard, southern service shed and eastern loading wing.
The east wing and outward loading canopy correspond to the factory gameplay
cell E. Short internal paths connect entrances without detached markings.

The 1536×1024 RGBA master was generated using a new numeric four-cell guide and
the newly authored winter HQ as a camera/material/season reference. Exact prompts,
input hashes, guide and selected master are retained in the compound source and
recipe directories. The existing HQ/depot generation-input chains remain tracked.
There are now five transparent compound masters, separate from the unchanged
27-source / 54-export opaque terrain pack.

Independent source and actual 2×/4× export review provisionally passes. Industrial
identity and main-roof construction remain readable at 2×; E is clearly associated
with its loading wing. No visible broken roof cuts, clipped architecture, tile
grid, dark halos or opaque vignette. Shared snow ground blends into both parity
surrounds. Facades are slightly more visible than HQ but remain compatible with
the near-overhead style. Small props become texture at 2×. Full maps, materials,
other zooms and runtime integration still need review.

## Geometry and framing

The four-cell diamond has world bounds 238×160 and crop origins N(71,0), W(0,40),
E(142,40), S(71,80). Runtime IDs remain `factoryN_w`, `factoryW_w`, `factoryE_w`,
`factoryS_w`; the entry remains E. Tile exports are 192×160 / 384×320; whole
scenes are 476×320 / 952×640.

An initial inset of 0.9 was rejected: 305 opaque sampled pixels entered the
eight-world-pixel outer guard. Uniform inset **0.85** preserves architecture and
passes, using scale `0.13170572916666665` and offset approximately
`(17.85,12.566666666666677)`. The nearest opaque sample lies 9.182 world pixels
inside the footprint. Only 98 alpha-1 pixels outside the union are discarded;
no guarded opaque architecture is clipped. Internal edges are excluded from
the outer guard.

The existing whole-scene premultiplied Lanczos3 exporter samples once, reduces
the whole artwork for 2×, composes over actual exported/decoded snow-ground tiles,
then slices with complementary polygon ownership. No production code changes
were needed. The retained ground output is reused without duplicate runtime files.

## Reproduction and evidence

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -compound art/compound-factory-v1.json \
  -out build/art-review/compound-factory-v1-final
```

Use a fresh output directory. The bundle contains 20 generated PNGs, two source
snapshots, manifest, report and gallery: 25 files. Public retention is eight tile
exports and six scene/parity review PNGs. See [the gallery](compound-factory-review.html)
and [output recipe](../../art/recipes/compound-factory-v1.json).

Independent actual-PNG assembly checks cover two density scenes and four parity
patches: 454,400 opaque owned pixels, 160,000 zero-RGBA outside pixels, 25,792
exact quiet-boundary pixels and 3,460 alpha≥128 artwork samples across all five
internal cuts. All 25 repeat-build files match exactly, and final-manifest PNGs
match the reviewed study. Source/input and output hashes and dimensions pass.
Existing depot, HQ, infrastructure and snow-ground artifacts remain unchanged.
All 134 ledger IDs, paths, campaign flags and candidate bindings pass verification.
Static image/link checks pass for the 81-image source, 177-image infrastructure,
24-image depot, 34-image HQ and 11-image factory galleries. Browser verification
loads all 11 factory images with continuous rendering and no horizontal overflow.

Final `just ci` passes: zero lint issues, all Go tests (cached; no production code
changes), desktop and WASM builds, seven inventory tests, and current inventory
of 273 paths / 246 loader calls.

Coverage is now 14 source + 102 built = 116 associated IDs, 18 backgrounds needing
images and zero accepted runtime replacements. Next: nine shore variants, eight
mountain pieces and fog, plus geometry for source-only candidates. The separate
350 native-resource pairs, remaining graphics categories and full-map/runtime
acceptance remain in scope. Continue Phase 2 after finishing the complete graphics set.
