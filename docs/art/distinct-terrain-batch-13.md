# Distinct terrain batch 13 — opaque never-seen fog

2026-10-15. A newly authored dark slate master supplies the enabled `unseen` fog
background at **2× and 4×**. This completes candidate association for the 134
enabled background/infrastructure IDs. Eleven backgrounds remain source-only;
native-resource coverage, other graphics and runtime/release acceptance remain open.

## Artwork and visibility

The new source is a quiet continuous matte pigment texture with subtle cloudy
variation. It was generated through the built-in image generator without reference
images. No original raster, diagnostic mask or reconstruction study is an input.
The exact prompt requests 1024×1024; the unchanged generated file is actually
1254×1254 RGB. The source recipe records the real dimensions and hash. All
1,572,516 source pixels are opaque.

[The visibility contract](fog-visibility-contract.md) distinguishes never-seen fog
from remembered greyscale terrain. Owned fog pixels must be opaque rather than
merely dark: no terrain, road, river, ownership or unit information may leak through.
Visible-to-fog borders intentionally mark the knowledge boundary. Fog-to-fog joins
form a continuous surface without a hex outline or translucent fringe. Runtime
draw order and suppression of hidden entities are separate integration tasks.

Independent art review provisionally passes the actual 2×/4× tiles, all-hidden
assemblies, summer/winter frontiers at both line parities, and road/river before-and-
after views. No important seam, vertex pinhole, halo, outline or concealment defect
was observed. Broad texture repetition may appear over larger fields and remains
a full-map review caveat. Hidden route approaches terminate at the intentional
visibility frontier rather than being painted endpoints in visible ground.

## Reproducible export

The bounded `-fog` mode consumes exactly three SHA-pinned new opaque materials:
fog, summer meadow and winter snow ground. Canonical schema/path checks reject
duplicates, unknown/case-alias/null/wrong-type fields, extra documents, stale pins,
unsafe paths and original/pilot sources. Exact loaded bytes are retained in source
snapshots. A fresh output directory is required; failure removes only its owned
new output. Existing renderers, profiles and earlier exporter outputs are unchanged.

The material adapter shares the established runtime lattice phase. Unmasked 4×
overscan is reduced with Lanczos3 before complementary final polygon ownership.
Tile sizes are 192×160 / 384×320. Every owned pixel is opaque and every exterior
pixel zero RGBA. No reference mask, independent masked-tile filtering or edge fade
is used. Summer/winter ground PNGs reproduce the existing built candidates exactly.

```sh
env GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -fog art/fog-terrain-v1.json \
  -out build/art-review/fog-terrain-v1-final
```

Use a fresh directory. The 24-file build contains 18 rendered PNGs, three source
snapshots, manifest, report and gallery. Its six seven-cell cases are all-hidden,
summer-frontier and winter-frontier, each at both runtime line parities and
densities. Absolute center/cell/canvas coordinates are retained; normalized pictures
alone do not establish correct placement.

Public retention is two runtime PNGs, twelve frontier/all-hidden review PNGs and
eight concealment-context PNGs. Existing ground files are reused. The additional
before/after contexts use actual new road14 and stream36 overlays on the appropriate
seasonal ground, with explicit cells, visibility, density and input hashes in
[the recipe](../../art/recipes/fog-terrain-v1.json). [The gallery](fog-review.html)
shows all 22 retained images. Neither these fixtures nor the new asset changes game
vision rules or implements runtime artwork selection.

## Verification and remaining work

An independent actual-PNG verifier checks all six tile PNGs: 340,800 opaque owned
pixels and 120,000 zero-RGBA exterior pixels. Twelve scene PNGs match ordinary
source-over reassembly byte-for-byte, with 4,771,200 singly owned opaque pixels
and 2,083,200 clear exterior pixels. Two contrasting underlays verify 227,200
concealed samples and 80,000 unchanged exterior samples.

The independent material oracle checks all 90,880 owned 4× fog pixels and 1,394
boundary/interior 2× samples against its own scalar Lanczos reference. Maximum RGB
error is one byte, within the declared tolerance of one. Another 3,120 samples
check phase at absolute neighboring origins. This proves the bounded material and
pixel fixtures, not full maps or runtime draw order. The actual road/river contexts
add 908,800 hidden pixels equal to the opaque fog, with every other pixel unchanged.

Independent code review finds no remaining blocking issue. Focused tests pass
after the real CLI rejects the missing flag before implementation; scoped vet
and lint pass. Fresh current-code builds reproduce all 911 earlier family artifacts
byte-for-byte. All 24 repeat files match, with no additional files. Source recipes,
loaded snapshots, all rendered PNG hashes/dimensions, nested report references and
retained copies pass verification. A separate reconstruction of all eight context
PNGs matches their declared cells and public pinned inputs, including every hidden
and unchanged exterior pixel. All 134 loader ID/path/campaign-use bindings pass.
Nine static galleries have 439 images and valid local links; the browser loads
all 22 new gallery images with continuous rendering and no horizontal overflow.
Final `just ci` passes: zero lint issues, art-review tests in 259.924 seconds,
remaining Go tests, desktop/WASM builds, seven inventory tests and the current
273 paths / 246 loader calls inventory.

The ledger now has **11 source + 123 built = 134 associated IDs**, no enabled ID
without an image association, and zero accepted replacements. Next build the
eleven source-only terrain adapters and resolve their quality/scale caveats.
Do not equate this enabled-ID milestone with the separate 350 native-resource
mapping obligation or complete unit/UI/marker/effect coverage. Full maps, fractional
zoom, remembered states, runtime integration and release acceptance remain open.
Continue Phase 2 after finishing the graphics scope, as requested.
