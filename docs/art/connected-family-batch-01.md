# Connected family batch 01

2026-10-15. [Gallery](connected-family-review.html) ·
[Manifest](../../art/connected-terrain-v1.json) ·
[Recipe and hashes](../../art/recipes/connected-terrain-v1.json).

Delivered all **16 loaded road variants and 11 stream variants** as newly built
192×160 / 384×320 overlays, plus a matching grass base. The source pack now
contains 19 generated terrain/material masters, including a new unmarked gravel
material. Nothing in this builder reads original artwork or the old reconstruction
studies. Existing stable IDs and explicit declared ports connect the new artwork
to the runtime coverage table; no game rules are changed.

## Geometry and export

The fixed 96×80 world descriptor uses the new continuous polygon
(25,0),(71,0),(96,40),(71,80),(25,80),(0,40), area 5680.
It tiles the existing 142×80/odd-line71 layout. Ports use the shared positions in
[runtime geometry evidence](runtime-connection-geometry.md); each has a straight
protected section and extends beyond the cutout before interior curves join.

Roads have 14-world-pixel cores with 2-pixel earth verges and no painted lines.
Streams have 18-pixel cores and 2.5-pixel protected banks. Their interiors add
restrained course/width variation, rounded confluences, thinner banks and
translucent earth pockets, with no baked-in summer grass. The complete eight-pixel
edge zone remains unchanged. Recipe/report text pins all equations and constants.

New material sampling is mirror-periodic on (142,80), averaging both lattice
cosets separated by (71,40). Independently exported cells share the texture phase.
This produces more regular repetition than final broad-map art may need.

Render unmasked features with four world pixels of overscan at 4×. Lanczos3
reduces 416×352 to 208×176; crop the margins afterward to 384×320 /192×160.
Apply complementary half-open polygon ownership at the FINAL pixel centers.
Feature edges retain antialiasing; polygon cuts are binary so ordinary sprite
source-over draws have exactly one owner. Original raster masks are not used.

Tested scope is pixel-aligned 2×/4× integer blits. At other zoom, compose at asset
density and resample the completed scene once. Independently filtering/fractionally
placing each tile is not validated here; runtime/camera integration remains open.

## Review and verification

Independent code review caught a real first-build defect: separate polygon-edge
antialiasing reduced diagonal core coverage to about 75%. Direct geometry previews
hid it. That build is excluded. Overscan/filter/final ownership fixes the exports;
all review patches now decode and assemble saved PNGs using ordinary source-over.

Controller and independent reviewer verified all 27 variants, 126 active-port/density
pairs and 14,572 core seam pixels: zero coverage gaps, zero double owners, minimum
alpha 255/255. All 108 saved parity/density patch PNGs match independent assembly
byte-for-byte. Tests also cover actual filtered bank profiles, both line parities,
vertices, literal material colors, validation failures and overwrite/rollback behavior.
Repeated final builds match all 198 files; source-pack builds match 59 files.

Independent art reviewer `connected_family_art_review` inspected representative
straight/bend/junction tiles, their actual assembled patches at 2×/4×, and all 27
variants on the contact sheet. Roads provisionally pass. Initial streams looked
like engineered canals; the interior repair rounds their confluences and varies
banks/course without changing contacts. Revised streams provisionally pass too.
Remaining caveats: clean bank edges, broadly uniform water direction and visible
grass repetition. Full maps, unit overlays, seasons/transitions and release
acceptance remain open. Repository `just ci` passes. Controller hash checks cover
all 19 sources, 38 material exports, 192 connected PNG outputs and the 56 runtime /
34 retained review copies, including generation-input hashes and ledger bindings.
Browser checks load all 60 connected-gallery and 57 material-gallery images with
no broken images or horizontal overflow at the checked desktop viewport.

## Reproduce

```sh
GOCACHE=/tmp/historyline-gfx-go-cache GOFLAGS=-buildvcs=false \
  go run ./tools/art-review -connected art/connected-terrain-v1.json \
  -out build/art-review/connected-family-v1-final
```

Choose a fresh output directory when it exists. The 198-file gallery includes four
retained material sources, 56 base/overlay exports, 27 grass composites, 108 actual
PNG patches, contact sheet, report and index. Tracked runtime exports are under
`art/exports/connected-v1/`; selected durable review images under
`art/review/connected-v1/`. The full local gallery stays in ignored build/.

The ledger now has 12 background source candidates plus 28 built candidates
(grass base and 27 overlays): **40 associated IDs, 94 still needing images, zero
accepted runtime replacements**. Native 350-resource mapping, remaining terrain/
building/seasonal pieces, paths/rail/trenches/bridges/crossings, ownership art,
sprites and UI remain open. Next extend the shared compositor to paths and the
other infrastructure families, while generating the missing building subjects.
