# Current game graphics comparison

From the repository root:

```sh
go run ./cmd/hlgfxreview
# or: just gfx-game-review
```

Open `build/gfx-review/index.html` directly in a browser. The page needs no server,
network, JavaScript dependencies or running game. Its `media/` directory contains
the exact image snapshots it displays. Regenerate after updating graphics or
rearranging their paths in the ledgers/recipes. Rerunning refreshes the index and
content-addressed snapshots; it does not delete old snapshots.

Search by name, family or runtime unit ID, then select a collection or season.
Choose an asset in the left list; previous/next buttons and left/right arrow keys
step through the filtered results. Each entry has a stable URL fragment.

- **Side by side** shows both images at equal display scale.
- **Wipe** slides the boundary between the reference and current artwork.
- **Blend** varies the opacity of the candidate over the reference.
- **Difference** compares the displayed images over the same background. Black
  means equal display colors; there is no percentage or approval threshold.

Use **Lanczos / original size** to judge the small image, then zoom to inspect its
pixels. Use the retained **2×/4× exports** to inspect authored detail. These are
separate controls: selecting a denser source keeps the same comparison area.
Missing densities are disabled; a 4× painted backdrop or 4× composed screen fixture
is not synthesized. Checkerboard, slate, paper and grass-tone backgrounds help
inspect transparency and contrast. File details include source paths, exact
snapshot hashes, pixel dimensions and crop/reduction methods.

The selection currently contains **390 entries**:

| Collection | Entries | Selection source |
|---|---:|---|
| Terrain and infrastructure | 134 | `docs/art/runtime-tile-coverage.csv` |
| Unit sprite families | 15 | `docs/art/runtime-unit-coverage.csv` and its pinned recipes |
| Portrait families | 15 | Same ledger and portrait recipes |
| Interface, markers, weapons and effect frames | 117 | Interface ledger and `interface-v1/report.json` |
| Screen layouts | 36 | 16 `screens-v1` compositions and 20 additional `screens-v2` states |
| Screen components | 73 | Current `screens-v2` report, including 61 widget IDs |

All **27 enabled unit IDs** remain associated with their families. Each sprite
family exposes all **36 authored facing/owner/state variants**. Planes, atlases,
generation sources, contact sheets and historical alternatives are excluded.
V1 screen requirements and V2 additional states are complementary requirements.
Screen compositions contain example text/values for readability; they are review
fixtures for the intended separate artwork and live text, not game screenshots.

Reference images come from `legacy/python/gfx/`. Explicit reference atlas cells
are cropped without resizing for ranks, flags and explosion frames. Weapon
references retain their complete old rows, including old labels. Components with
no direct mapped reference show the candidate alone and disable overlay modes.
Current reference action icons stay unchanged when reviewing new interaction
states. No native DOS/Amiga correspondence is inferred.

The raw unit reference is 48×48; the current 96×96 logical sprite is reduced to
48×48 for original-size comparison. A 2× display zoom gives the 96×96 game
footprint. The raw reference pose/palette stays unchanged for all new variants;
the page invents no rotated, recolored or spent reference. Images with different
dimensions are centered on a common canvas, preserving their own aspect/size.
This inspection alignment does not verify physical game anchors or native facings.
Terrain uses the currently bound default context exports; full-map joins and
neighbor-aware composition remain separate work.

Missing original-size terrain/sprite/fixture images are reduced offline using the
same premultiplied byte-RGB Lanczos3 algorithm as `tools/art-review/image.go`:
signed sinc lobes, widened support, clipped/renormalized edges, float64 sums,
output-only half-up rounding and alpha/color clamping. Existing original-size
exports are used unchanged. Pinned recipe/report inputs are checked before
snapshotting, and PNG streams are fully decoded. Terrain ledger exports do not
provide output hash pins; their snapshot hashes record the bytes reviewed.

Run `go test ./cmd/hlgfxreview` for real-catalog coverage, all snapshot hashes and
dimensions, malformed-input/path/crop rejection and alpha-aware filtering checks.
For desktop/mobile interaction checks, set `HL_GFX_REVIEW_BROWSER` to an installed
Chromium executable and run the same test. It launches isolated headless profiles
with desktop and narrow window sizes, exercising filters, variants, source
density, wipe, opacity, difference, unpaired/empty states and image loading.
Chromium on this Windows installation enforces a minimum 500-CSS-pixel layout
width, including for the requested 460-pixel capture window.

All generated HTML/images stay under ignored `build/` because the comparisons
contain private reference images. The tool reads the user-owned graphics track
without editing `art/`, `docs/art/` or `tools/art-review/`. It does not approve,
install, embed or alter any game asset.
